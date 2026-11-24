package plugin

import (
	"bytes"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/dohwi/cliproxyapi-plugin-usage-audit/internal/store"
)

// A process-local random HMAC key is sufficient: cursors deliberately expire on
// restart/reconfigure. There is no configured or persisted secret to leak.
var cursorSecret = func() []byte {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		panic("usage-audit: cursor entropy unavailable")
	}
	return key
}()

type recordsCursor struct {
	Version    int       `json:"v"`
	Filter     string    `json:"filter"`
	Generation string    `json:"generation"`
	Boundary   uint64    `json:"boundary"`
	Timestamp  time.Time `json:"timestamp"`
	Sequence   uint64    `json:"sequence"`
}

type pageError struct {
	status  int
	code    string
	message string
}

func (e *pageError) Error() string { return e.message }

func filterFingerprint(opts store.FilterOptions, scope string) string {
	if opts.From != nil {
		t := opts.From.UTC()
		opts.From = &t
	}
	if opts.To != nil {
		t := opts.To.UTC()
		opts.To = &t
	}
	b, _ := json.Marshal(struct {
		Scope   string
		Filters store.FilterOptions
	}{scope, opts})
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}
func encodeRecordsCursor(c recordsCursor) string {
	b, _ := json.Marshal(c)
	mac := hmac.New(sha256.New, cursorSecret)
	_, _ = mac.Write(b)
	return base64.RawURLEncoding.EncodeToString(b) + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
func decodeRecordsCursor(value, fingerprint string) (*recordsCursor, error) {
	if value == "" {
		return nil, nil
	}
	invalid := func(message string) (*recordsCursor, error) { return nil, &pageError{400, codeInvalidCursor, message} }
	if len(value) > 2048 {
		return invalid("invalid cursor: too long")
	}
	parts := strings.Split(value, ".")
	if len(parts) != 2 {
		return invalid("unsupported or legacy offset cursor; restart pagination without cursor")
	}
	b, err := base64.RawURLEncoding.Strict().DecodeString(parts[0])
	if err != nil {
		return invalid("invalid cursor encoding")
	}
	sig, err := base64.RawURLEncoding.Strict().DecodeString(parts[1])
	if err != nil {
		return invalid("invalid cursor signature")
	}
	mac := hmac.New(sha256.New, cursorSecret)
	_, _ = mac.Write(b)
	if !hmac.Equal(sig, mac.Sum(nil)) {
		return invalid("invalid cursor signature; restart pagination")
	}
	var c recordsCursor
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.DisallowUnknownFields()
	if dec.Decode(&c) != nil {
		return invalid("invalid cursor payload")
	}
	canonical, _ := json.Marshal(c)
	if !bytes.Equal(canonical, b) || c.Version != 1 || c.Filter != fingerprint || c.Generation == "" || len(c.Generation) > 128 || c.Boundary == 0 || c.Sequence == 0 || c.Sequence > c.Boundary {
		return invalid("invalid cursor or filters changed; restart pagination")
	}
	return &c, nil
}

// recordsPage uses a bounded snapshot frontier, not a shifting offset. Newly
// ingested records (even backdated records and timestamp ties) cannot enter a
// continuation page. Rotation/clear/external edits expire the generation rather
// than silently returning incomplete history.
func recordsPage(st *store.Store, opts store.FilterOptions, scope, value string, limit int) ([]store.LogRecord, string, error) {
	page, next, _, _, err := recordsPageQuery(st, opts, scope, value, 0, limit)
	return page, next, err
}

// recordsPageQuery supports both cursor continuation (pageNum == 0) and
// 1-indexed direct page jumps (pageNum > 0). Both modes return total matching
// records and hasMore so the UI can render fixed-slot numbered pagination.
func recordsPageQuery(st *store.Store, opts store.FilterOptions, scope, cursorValue string, pageNum, limit int) (page []store.LogRecord, next string, total int, hasMore bool, err error) {
	if pageNum > 0 && cursorValue != "" {
		return nil, "", 0, false, &pageError{400, codeInvalidCursor, "cannot combine page and cursor"}
	}
	if pageNum > 0 {
		snap, err := st.SnapshotView()
		if err != nil {
			return nil, "", 0, false, err
		}
		page, total, hasMore = selectRecordsPageByNumber(snap.Records[:int(snap.Boundary)], opts, pageNum, limit)
		return page, "", total, hasMore, nil
	}
	fingerprint := filterFingerprint(opts, scope)
	cursor, err := decodeRecordsCursor(cursorValue, fingerprint)
	if err != nil {
		return nil, "", 0, false, err
	}
	snap, err := st.SnapshotView() // read-only: only the selected page is copied
	if err != nil {
		return nil, "", 0, false, err
	}
	boundary := snap.Boundary
	if cursor != nil {
		if cursor.Generation != snap.Generation || cursor.Boundary > snap.Boundary {
			return nil, "", 0, false, &pageError{409, codeCursorExpired, "cursor snapshot expired; restart pagination"}
		}
		boundary = cursor.Boundary
	}
	// Scan only the original append frontier, including out-of-order timestamps.
	page, hasMore, anchorFound, total := selectRecordsPageWithTotal(snap.Records[:int(boundary)], opts, cursor, limit)
	if !anchorFound {
		return nil, "", 0, false, &pageError{400, codeInvalidCursor, "cursor anchor not found; restart pagination"}
	}
	if hasMore {
		r := page[len(page)-1]
		next = encodeRecordsCursor(recordsCursor{1, fingerprint, snap.Generation, boundary, r.EffectiveTimestamp().UTC(), r.Sequence})
	}
	return page, next, total, hasMore, nil
}

func recordsPageFailure(err error) (int, string, string) {
	var e *pageError
	if errors.As(err, &e) {
		return e.status, e.code, e.message
	}
	return 500, codeReadFailed, "read failed"
}

const (
	resourceSummaryDefaultDays = 30
	resourceSummaryMaxDays     = 366
	resourceSummaryMaxBuckets  = 4096
)

func boundResourceSummary(opts *store.FilterOptions) error {
	to := time.Now().UTC()
	if opts.To != nil {
		to = *opts.To
	} else {
		opts.To = &to
	}
	from := to.Add(-resourceSummaryDefaultDays * 24 * time.Hour)
	if opts.From != nil {
		from = *opts.From
	} else {
		opts.From = &from
	}
	if to.Before(from) || to.Sub(from) > resourceSummaryMaxDays*24*time.Hour {
		return newAPIError(codeSummaryRange, "summary range must be ordered and at most %d days", resourceSummaryMaxDays)
	}
	return nil
}
