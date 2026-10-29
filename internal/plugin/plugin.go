// Package plugin implements the CLIProxyAPI plugin RPC dispatch for usage-audit.
// All exported symbols in this package are called from main.go.
package plugin

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/dohwi/cliproxyapi-plugin-usage-audit/internal/queue"
	"github.com/dohwi/cliproxyapi-plugin-usage-audit/internal/store"
	"github.com/router-for-me/CLIProxyAPI/v7/sdk/pluginabi"
	"github.com/router-for-me/CLIProxyAPI/v7/sdk/pluginapi"
)

// ----------------------------------------------------------------------------
// Plugin metadata
// ----------------------------------------------------------------------------

const (
	pluginName    = "usage-audit"
	pluginVersion = "0.2.5"
	pluginAuthor  = "dohwi"
	// pluginRepoURL is this plugin's own public repository, shown by the host and the plugin store.
	pluginRepoURL = "https://github.com/dohwi/cliproxyapi-plugin-usage-audit"
)

// defaultDataDir returns the per-user, OS-appropriate data directory used when
// USAGE_AUDIT_DATA_DIR is not set (for example ~/.config/cliproxyapi/usage-audit
// on Linux, ~/Library/Application Support/... on macOS, %AppData%\... on Windows).
// It returns "" when the platform has no usable per-user directory; the plugin
// then runs without storage instead of writing to a shared location.
func defaultDataDir() string {
	base, err := os.UserConfigDir()
	if err != nil || strings.TrimSpace(base) == "" {
		return ""
	}
	return filepath.Join(base, "cliproxyapi", "usage-audit")
}

// Font assets are embedded in the plugin binary and returned as raw bytes via
// the binary-safe ManagementResponse Body field. The host's resource protocol
// transports Body as base64 in the RPC envelope, then writes its original bytes
// to HTTP; this avoids a CDN and does not rely on a text response envelope.
//
//go:embed assets/fonts/manrope-variable.woff2
var manropeVariableWOFF2 []byte

//go:embed assets/fonts/suit-regular.woff2
var suitRegularWOFF2 []byte

//go:embed assets/fonts/suit-semibold.woff2
var suitSemiboldWOFF2 []byte

// ----------------------------------------------------------------------------
// Global state — replaced atomically on each plugin.register call.
// stateMu guards read/write of the state pointer.
// ----------------------------------------------------------------------------

var (
	stateMu sync.RWMutex
	state   *pluginState
)

type pluginState struct {
	q        *queue.Queue
	st       *store.Store
	dataDir  string
	logPath  string
	workerWg sync.WaitGroup
	// lifecycle prevents an in-flight observer from racing shutdown.
	lifecycle sync.RWMutex
	// shutdownOnce prevents double-close of q.shutdown.
	shutdownOnce sync.Once
	// epochMu serializes the consumer commit with DELETE, not queue consumption.
	epochMu sync.Mutex
	epoch   atomic.Uint64
}

// shutdown safely stops the consumer and flushes the store exactly once.
func (ps *pluginState) shutdown() {
	ps.shutdownOnce.Do(func() {
		// Exclude observers while stopping the queue and closing storage. Queue.Close
		// is itself idempotent, but this also protects the state transition around it.
		ps.lifecycle.Lock()
		defer ps.lifecycle.Unlock()
		ps.q.Close()
		done := make(chan struct{})
		go func() {
			ps.workerWg.Wait()
			close(done)
		}()
		select {
		case <-done:
			if err := ps.st.Close(); err != nil {
				log.Printf("usage-audit: shutdown store: %v", err)
			}
		case <-time.After(5 * time.Second):
			// Never let plugin shutdown block indefinitely. Leave the resources
			// open if a worker fails to stop; closing them would race the worker.
			log.Printf("usage-audit: shutdown timed out; storage left open")
		}
	})
}

// ----------------------------------------------------------------------------
// Envelope helpers
// ----------------------------------------------------------------------------

type envelope struct {
	OK     bool            `json:"ok"`
	Result json.RawMessage `json:"result,omitempty"`
	Error  *envelopeError  `json:"error,omitempty"`
}

type envelopeError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// OKEnvelope marshals result into a success envelope.
func OKEnvelope(result any) []byte {
	r, err := json.Marshal(result)
	if err != nil {
		return ErrorEnvelope("marshal_error", err.Error())
	}
	raw, _ := json.Marshal(envelope{OK: true, Result: json.RawMessage(r)})
	return raw
}

// ErrorEnvelope builds an error envelope.
func ErrorEnvelope(code, message string) []byte {
	raw, _ := json.Marshal(envelope{
		OK:    false,
		Error: &envelopeError{Code: code, Message: message},
	})
	return raw
}

// ----------------------------------------------------------------------------
// HandleMethod — main RPC dispatch (called from main.go CGo bridge)
// ----------------------------------------------------------------------------

// HandleMethod dispatches a host→plugin RPC call.
// Returns (envelope bytes, true) on success or (error envelope, false) for unknown methods.
func HandleMethod(method string, request []byte) ([]byte, bool) {
	switch method {
	case pluginabi.MethodPluginRegister, pluginabi.MethodPluginReconfigure:
		return handleRegister(request), true

	case pluginabi.MethodPluginQuiesce:
		return OKEnvelope(map[string]any{}), true

	case pluginabi.MethodPluginShutdown:
		Shutdown()
		return OKEnvelope(map[string]any{}), true

	case pluginabi.MethodUsageHandle:
		handleUsage(request)
		// usage.handle response: host ignores content, just needs ok=true.
		return OKEnvelope(map[string]any{}), true

	case pluginabi.MethodManagementRegister:
		return handleManagementRegister(request), true

	case pluginabi.MethodManagementHandle:
		return handleManagementHandle(request), true

	default:
		return ErrorEnvelope("unknown_method", fmt.Sprintf("unknown method: %s", method)), false
	}
}

// ----------------------------------------------------------------------------
// plugin.register / plugin.reconfigure
// ----------------------------------------------------------------------------

type registrationResponse struct {
	SchemaVersion uint32             `json:"schema_version"`
	Metadata      pluginapi.Metadata `json:"metadata"`
	Capabilities  map[string]bool    `json:"capabilities"`
}

func handleRegister(request []byte) []byte {
	var req struct {
		SchemaVersion uint32 `json:"schema_version"`
	}
	_ = json.Unmarshal(request, &req)

	// Tear down any existing state (reconfigure path, or repeated register in tests).
	stateMu.Lock()
	old := state
	state = nil
	stateMu.Unlock()
	if old != nil {
		old.shutdown()
	}

	// Resolve data directory.
	dataDir := strings.TrimSpace(os.Getenv("USAGE_AUDIT_DATA_DIR"))
	if dataDir == "" {
		dataDir = defaultDataDir()
	}
	if dataDir == "" {
		log.Printf("usage-audit: no data directory; set USAGE_AUDIT_DATA_DIR")
		return buildRegistrationResponse(req.SchemaVersion)
	}
	if err := os.MkdirAll(dataDir, 0o750); err != nil {
		log.Printf("usage-audit: mkdir %s: %v", dataDir, err)
		// Return registration without storage — usage records will be silently dropped.
		return buildRegistrationResponse(req.SchemaVersion)
	}

	logPath := filepath.Join(dataDir, "usage.jsonl")
	st, err := store.OpenWithOptions(logPath, store.DefaultMaxBytes, store.DefaultMaxRotation, store.Options{
		StrictDecoding: strictDecodingEnabled(),
	})
	if err != nil {
		log.Printf("usage-audit: open store %s: %v", logPath, err)
		return buildRegistrationResponse(req.SchemaVersion)
	}

	q := queue.New(queue.DefaultCapacity)
	ps := &pluginState{
		q:       q,
		st:      st,
		dataDir: dataDir,
		logPath: logPath,
	}
	ps.workerWg.Add(1)
	go ps.consumeLoop()

	stateMu.Lock()
	state = ps
	stateMu.Unlock()

	return buildRegistrationResponse(req.SchemaVersion)
}

func strictDecodingEnabled() bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv("USAGE_AUDIT_STRICT_DECODING"))) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}

func buildRegistrationResponse(hostSchemaVersion uint32) []byte {
	sv := pluginabi.SchemaVersion
	if hostSchemaVersion > 0 && hostSchemaVersion < sv {
		sv = hostSchemaVersion
	}
	return OKEnvelope(registrationResponse{
		SchemaVersion: sv,
		Metadata: pluginapi.Metadata{
			Name:             pluginName,
			Version:          pluginVersion,
			Author:           pluginAuthor,
			GitHubRepository: pluginRepoURL,
			Logo:             pluginLogo,
		},
		Capabilities: map[string]bool{
			"usage_plugin":   true,
			"management_api": true,
		},
	})
}

// ----------------------------------------------------------------------------
// usage.handle
// ----------------------------------------------------------------------------

// usageHandleRequest mirrors pluginapi.UsageRecord JSON field names.
// Go struct field names match the JSON keys emitted by the host RPC serializer
// (rpc_client.go: callPlugin marshals pluginapi.UsageRecord directly).
type usageHandleRequest struct {
	Provider        string       `json:"Provider"`
	ExecutorType    string       `json:"ExecutorType"`
	Model           string       `json:"Model"`
	Alias           string       `json:"Alias"`
	APIKey          string       `json:"APIKey"`
	SessionID       string       `json:"SessionID"`
	ParentSessionID string       `json:"ParentSessionID"`
	AuthID          string       `json:"AuthID"`    // intentionally NOT stored
	AuthIndex       string       `json:"AuthIndex"` // stable index — stored
	AuthType        string       `json:"AuthType"`
	Source          string       `json:"Source"`
	ReasoningEffort string       `json:"ReasoningEffort"`
	ServiceTier     string       `json:"ServiceTier"`
	Generate        bool         `json:"Generate"`
	RequestedAt     time.Time    `json:"RequestedAt"`
	Latency         int64        `json:"Latency"` // nanoseconds
	TTFT            int64        `json:"TTFT"`    // nanoseconds
	Failed          bool         `json:"Failed"`
	Failure         usageFailure `json:"Failure"`
	Detail          usageDetail  `json:"Detail"`
	// Flat token fields are accepted for compatibility with older host/test
	// serializers; the canonical wire shape remains Detail.{...}.
	InputTokens         int64 `json:"InputTokens"`
	OutputTokens        int64 `json:"OutputTokens"`
	ReasoningTokens     int64 `json:"ReasoningTokens"`
	CachedTokens        int64 `json:"CachedTokens"`
	CacheReadTokens     int64 `json:"CacheReadTokens"`
	CacheCreationTokens int64 `json:"CacheCreationTokens"`
	TotalTokens         int64 `json:"TotalTokens"`
}

type usageFailure struct {
	StatusCode int    `json:"StatusCode"`
	Body       string `json:"Body"`
}

type usageDetail struct {
	InputTokens         int64 `json:"InputTokens"`
	OutputTokens        int64 `json:"OutputTokens"`
	ReasoningTokens     int64 `json:"ReasoningTokens"`
	CachedTokens        int64 `json:"CachedTokens"`
	CacheReadTokens     int64 `json:"CacheReadTokens"`
	CacheCreationTokens int64 `json:"CacheCreationTokens"`
	TotalTokens         int64 `json:"TotalTokens"`
}

func handleUsage(request []byte) {
	stateMu.RLock()
	ps := state
	stateMu.RUnlock()
	if ps == nil {
		return // store not ready; drop silently
	}

	// Capture before decoding: observers already in flight at DELETE are stale.
	epoch := ps.epoch.Load()
	var rec usageHandleRequest
	if err := json.Unmarshal(request, &rec); err != nil {
		log.Printf("usage-audit: unmarshal usage.handle: %v", err)
		return
	}
	// Some older callers emitted token fields at the top level. Preserve that
	// compatibility without changing the canonical nested Detail contract.
	if rec.Detail.InputTokens == 0 && rec.Detail.OutputTokens == 0 && rec.Detail.TotalTokens == 0 &&
		rec.Detail.ReasoningTokens == 0 && rec.Detail.CachedTokens == 0 && rec.Detail.CacheReadTokens == 0 && rec.Detail.CacheCreationTokens == 0 {
		rec.Detail = usageDetail{
			InputTokens: rec.InputTokens, OutputTokens: rec.OutputTokens,
			ReasoningTokens: rec.ReasoningTokens, CachedTokens: rec.CachedTokens,
			CacheReadTokens: rec.CacheReadTokens, CacheCreationTokens: rec.CacheCreationTokens,
			TotalTokens: rec.TotalTokens,
		}
	}

	lr := store.LogRecord{
		T:            time.Now().UTC(),
		Provider:     rec.Provider,
		ExecutorType: rec.ExecutorType,
		Model:        rec.Model,
		Alias:        rec.Alias,
		// APIKey is deliberately omitted: the host may supply the raw secret.
		SessionID:       rec.SessionID,
		ParentSessionID: rec.ParentSessionID,
		// AuthID is intentionally omitted — only the stable AuthIndex is stored.
		AuthIndex:       rec.AuthIndex,
		AuthType:        rec.AuthType,
		Source:          rec.Source,
		ReasoningEffort: rec.ReasoningEffort,
		ServiceTier:     rec.ServiceTier,
		Generate:        rec.Generate, // false ≠ missing; preserved explicitly
		RequestedAt:     rec.RequestedAt,
		LatencyMs:       float64(rec.Latency) / float64(time.Millisecond),
		TTFTMs:          float64(rec.TTFT) / float64(time.Millisecond),
		Failed:          rec.Failed,
		FailureStatus:   rec.Failure.StatusCode,
		// All token fields: zero is valid (token-count-only request or upstream omission).
		InputTokens:         rec.Detail.InputTokens,
		OutputTokens:        rec.Detail.OutputTokens,
		ReasoningTokens:     rec.Detail.ReasoningTokens,
		CachedTokens:        rec.Detail.CachedTokens,
		CacheReadTokens:     rec.Detail.CacheReadTokens,
		CacheCreationTokens: rec.Detail.CacheCreationTokens,
		TotalTokens:         rec.Detail.TotalTokens,
	}

	line, err := json.Marshal(queuedUsage{LogRecord: lr, Epoch: epoch})
	if err != nil {
		log.Printf("usage-audit: marshal log record: %v", err)
		return
	}
	ps.lifecycle.RLock()
	ps.q.Enqueue(queue.Record(line))
	ps.lifecycle.RUnlock()
}

// consumeLoop is the single background goroutine draining q → st.
// Errors are logged; they never propagate to the request path.
func (ps *pluginState) consumeLoop() {
	defer ps.workerWg.Done()
	for {
		select {
		case rec, ok := <-ps.q.Consume():
			if !ok {
				// Channel closed by shutdown — drain any remaining items.
				ps.drainRemaining()
				return
			}
			ps.appendRecord(rec)
		case <-ps.q.ShutdownCh():
			ps.drainRemaining()
			return
		}
	}
}

func (ps *pluginState) drainRemaining() {
	for {
		select {
		case rec, ok := <-ps.q.Consume():
			if !ok {
				return
			}
			ps.appendRecord(rec)
		default:
			if err := ps.st.Flush(); err != nil {
				log.Printf("usage-audit: drain flush: %v", err)
			}
			return
		}
	}
}

type queuedUsage struct {
	store.LogRecord
	Epoch uint64 `json:"queue_epoch"`
}

func (ps *pluginState) appendRecord(rec queue.Record) {
	var lr queuedUsage
	if err := json.Unmarshal(rec, &lr); err != nil {
		log.Printf("usage-audit: consumer unmarshal: %v", err)
		return
	}
	ps.epochMu.Lock()
	defer ps.epochMu.Unlock()
	if lr.Epoch != ps.epoch.Load() {
		return
	}
	if err := ps.st.Append(lr.LogRecord); err != nil {
		log.Printf("usage-audit: store append: %v", err)
	}
}

// ----------------------------------------------------------------------------
// management.register
// ----------------------------------------------------------------------------

type mgmtRoute struct {
	Method string `json:"Method"`
	Path   string `json:"Path"`
}

// resourceRoute mirrors pluginapi.ResourceRoute (JSON field names match the
// host RPC decoder, which uses default Go field names — no json tags).
type resourceRoute struct {
	Path        string `json:"Path"`
	Menu        string `json:"Menu"`
	Description string `json:"Description"`
}

type mgmtRegisterResponse struct {
	Routes    []mgmtRoute     `json:"routes"`
	Resources []resourceRoute `json:"resources"`
}

// mgmtBase is the prefix for all plugin management routes.
// Host constraint: must be under /v0/management/, no wildcards.
const mgmtBase = "/v0/management/plugins/usage-audit"

// resourceBase is the host prefix for browser-navigable plugin resources.
// Resource requests are dispatched to management.handle with the full URL path.
const resourceBase = "/v0/resource/plugins/usage-audit"

// resourceDashboard is the single browser-navigable resource exposed by this
// plugin. Its Menu label makes it appear in the management console sidebar.
var resourceDashboard = resourceRoute{
	Path:        "/dashboard",
	Menu:        "Usage Audit", // matches the plugin store name; the host registers it once, so it cannot follow the page language
	Description: "토큰 사용량 및 요청 로그 대시보드",
}

func handleManagementRegister(_ []byte) []byte {
	return OKEnvelope(mgmtRegisterResponse{
		Routes: []mgmtRoute{
			{Method: "GET", Path: mgmtBase + "/dashboard"},
			{Method: "GET", Path: mgmtBase + "/records"},
			{Method: "GET", Path: mgmtBase + "/summary"},
			{Method: "DELETE", Path: mgmtBase + "/records"},
			// Dashboard-shaped (bounded, auth_index-free) data for the browser UI.
			{Method: "GET", Path: mgmtBase + "/data/summary"},
			{Method: "GET", Path: mgmtBase + "/data/records"},
		},
		// /v0/resource/... is unauthenticated, so only the static page shell and
		// fonts live here. All usage data is served from the authenticated
		// /v0/management/plugins/usage-audit/... routes above.
		Resources: []resourceRoute{
			resourceDashboard,
			{Path: "/assets/fonts/manrope-variable.woff2", Menu: "", Description: ""},
			{Path: "/assets/fonts/suit-regular.woff2", Menu: "", Description: ""},
			{Path: "/assets/fonts/suit-semibold.woff2", Menu: "", Description: ""},
		},
	})
}

// ----------------------------------------------------------------------------
// management.handle — sub-route dispatch
// ----------------------------------------------------------------------------

// managementHandleRequest mirrors pluginapi.ManagementRequest JSON fields.
type managementHandleRequest struct {
	Method string              `json:"Method"`
	Path   string              `json:"Path"`
	Query  map[string][]string `json:"Query"`
	Body   []byte              `json:"Body"`
}

type managementResponse struct {
	StatusCode int                 `json:"StatusCode"`
	Headers    map[string][]string `json:"Headers"`
	Body       []byte              `json:"Body"`
}

func handleManagementHandle(request []byte) []byte {
	var req managementHandleRequest
	if err := json.Unmarshal(request, &req); err != nil {
		return OKEnvelope(managementResponse{StatusCode: 400, Body: jsonStr(`{"error":"bad request"}`)})
	}
	method := strings.ToUpper(strings.TrimSpace(req.Method))
	path := strings.TrimRight(strings.TrimSpace(req.Path), "/")

	switch {
	// Dashboard is reachable both as a management API route and as a
	// browser-navigable resource under the host resource prefix. The bare
	// "/dashboard" form is accepted defensively for relative dispatches.
	case method == "GET" && (path == mgmtBase+"/dashboard" || path == resourceBase+"/dashboard" || path == "/dashboard"):
		return handleGetDashboard()
	case method == "GET" && path == mgmtBase+"/records":
		return handleGetRecords(req)
	case method == "GET" && path == mgmtBase+"/summary":
		return handleGetSummary(req)
	case method == "GET" && path == resourceBase+"/assets/fonts/manrope-variable.woff2":
		return fontResourceResp(manropeVariableWOFF2)
	case method == "GET" && path == resourceBase+"/assets/fonts/suit-regular.woff2":
		return fontResourceResp(suitRegularWOFF2)
	case method == "GET" && path == resourceBase+"/assets/fonts/suit-semibold.woff2":
		return fontResourceResp(suitSemiboldWOFF2)
	case method == "GET" && path == mgmtBase+"/data/records":
		return handleGetDashboardRecords(req)
	case method == "GET" && path == mgmtBase+"/data/summary":
		return handleGetDashboardSummary(req)
	case method == "DELETE" && path == mgmtBase+"/records":
		return handleDeleteRecords()
	default:
		return OKEnvelope(managementResponse{StatusCode: 404, Body: jsonStr(`{"error":"not found"}`)})
	}
}

// ----------------------------------------------------------------------------
// GET /dashboard — self-contained dashboard UI
// ----------------------------------------------------------------------------

// dashboardSecurityHeaders hardens the self-contained page. The page is a single
// HTML document with inline script/style, so script-src must allow 'unsafe-inline';
// everything else is locked down: no external origins, framing only by the same origin (the management console embeds this page), no sniffing,
// no referrer, no plugin content, no base-tag or form-target hijacking.
var dashboardSecurityHeaders = map[string][]string{
	"Content-Security-Policy": {"default-src 'none'; script-src 'unsafe-inline'; style-src 'unsafe-inline'; font-src 'self'; img-src 'self' data:; connect-src 'self'; base-uri 'none'; form-action 'none'; frame-ancestors 'self'"},
	"X-Content-Type-Options":  {"nosniff"},
	"Referrer-Policy":         {"no-referrer"},
	"X-Frame-Options":         {"SAMEORIGIN"},
}

var (
	dashboardRespOnce sync.Once
	dashboardRespEnv  []byte
)

// handleGetDashboard returns the fully encoded dashboard envelope. Status,
// headers (static CSP) and body (embedded assets) are constants that do not
// depend on the request, language, key or plugin state, so the envelope is built
// once. The result is read-only; the C bridge copies it and never mutates it.
func handleGetDashboard() []byte {
	dashboardRespOnce.Do(func() {
		headers := map[string][]string{
			"Content-Type":  {"text/html; charset=utf-8"},
			"Cache-Control": {"no-store"},
		}
		for k, v := range dashboardSecurityHeaders {
			headers[k] = v
		}
		dashboardRespEnv = OKEnvelope(managementResponse{
			StatusCode: 200,
			Headers:    headers,
			Body:       []byte(DashboardHTML),
		})
	})
	return dashboardRespEnv
}

func parseTimeQuery(val string, isTo bool) (*time.Time, error) {
	return parseTimeQueryInLocation(val, isTo, time.UTC)
}

func parseTimeQueryInLocation(val string, isTo bool, loc *time.Location) (*time.Time, error) {
	val = strings.TrimSpace(val)
	if val == "" {
		return nil, nil
	}
	// Try RFC3339
	if t, err := time.Parse(time.RFC3339Nano, val); err == nil {
		return &t, nil
	}
	if t, err := time.Parse(time.RFC3339, val); err == nil {
		return &t, nil
	}
	// Try YYYY-MM-DD. Bare dates are calendar dates in the requested location;
	// RFC3339 values above remain absolute instants.
	if loc == nil {
		loc = time.UTC
	}
	if t, err := time.ParseInLocation("2006-01-02", val, loc); err == nil {
		if isTo {
			nextDay := t.AddDate(0, 0, 1)
			t = nextDay.Add(-time.Nanosecond)
		}
		return &t, nil
	}
	return nil, fmt.Errorf("invalid time format: %s", val)
}

func parseTimezoneLocation(query map[string][]string) (*time.Location, error) {
	value := strings.TrimSpace(queryFirst(query, "time_zone"))
	if value == "" {
		return nil, nil
	}
	if len(value) > 128 || strings.ContainsAny(value, "\x00\r\n") {
		return nil, fmt.Errorf("invalid time zone: %s", value)
	}
	loc, err := time.LoadLocation(value)
	if err != nil {
		return nil, fmt.Errorf("invalid time zone: %s", value)
	}
	return loc, nil
}

func parseTimezoneOffset(query map[string][]string) (time.Duration, error) {
	value := strings.TrimSpace(queryFirst(query, "tz_offset"))
	if value == "" {
		return 0, nil
	}
	offsetMinutes, err := strconv.Atoi(value)
	if err != nil || offsetMinutes < -14*60 || offsetMinutes > 14*60 {
		return 0, fmt.Errorf("invalid timezone offset: %s", value)
	}
	// JavaScript getTimezoneOffset() is UTC - local time, so negate it to
	// convert a UTC timestamp into the browser's local wall-clock date.
	return -time.Duration(offsetMinutes) * time.Minute, nil
}

func parseFilterOptions(query map[string][]string) (store.FilterOptions, error) {
	var opts store.FilterOptions
	var err error
	loc, err := parseTimezoneLocation(query)
	if err != nil {
		return opts, err
	}
	if loc == nil {
		loc = time.UTC
	}

	fromStr := queryFirst(query, "from")
	if fromStr == "" {
		fromStr = queryFirst(query, "since") // backwards compatibility
	}
	if fromStr != "" {
		opts.From, err = parseTimeQueryInLocation(fromStr, false, loc)
		if err != nil {
			return opts, err
		}
	}

	toStr := queryFirst(query, "to")
	if toStr != "" {
		opts.To, err = parseTimeQueryInLocation(toStr, true, loc)
		if err != nil {
			return opts, err
		}
	}

	if opts.From != nil && opts.To != nil && opts.To.Before(*opts.From) {
		return opts, fmt.Errorf("from must not be after to")
	}

	now := time.Now().Add(5 * time.Minute) // tolerate slight clock skew
	if opts.From != nil && opts.From.After(now) {
		return opts, fmt.Errorf("future timestamp not allowed for from")
	}

	opts.Provider = queryFirst(query, "provider")
	opts.Model = queryFirst(query, "model")
	opts.Source = queryFirst(query, "source")
	opts.ExecutorType = queryFirst(query, "executor_type")
	opts.AuthIndex = queryFirst(query, "auth_index")
	// "endpoint" is a single UI-facing filter matching ExecutorType OR Source.
	opts.Endpoint = queryFirst(query, "endpoint")

	status := strings.ToLower(queryFirst(query, "status"))
	if status != "" {
		if status != "success" && status != "failed" {
			return opts, fmt.Errorf("invalid status filter: %s", status)
		}
		opts.Status = status
	}

	return opts, nil
}

// ----------------------------------------------------------------------------
// GET /records — newest-first filtered tail of JSONL log
// ----------------------------------------------------------------------------

func handleGetRecords(req managementHandleRequest) []byte {
	stateMu.RLock()
	ps := state
	stateMu.RUnlock()
	if ps == nil {
		return jsonMgmtResp(503, map[string]string{"error": "store not ready"})
	}

	opts, err := parseFilterOptions(req.Query)
	if err != nil {
		return jsonMgmtResp(400, map[string]string{"error": err.Error()})
	}

	limit := 50
	if v := queryFirst(req.Query, "limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n <= 0 || n > 500 {
			return jsonMgmtResp(400, map[string]string{"error": "invalid limit; must be 1..500"})
		}
		limit = n
	}

	ps.lifecycle.RLock()
	defer ps.lifecycle.RUnlock()
	paged, nextCursor, err := recordsPage(ps.st, opts, "management", queryFirst(req.Query, "cursor"), limit)
	if err != nil {
		status, message := recordsPageFailure(err)
		if status == 500 {
			log.Printf("usage-audit: records read: %v", err)
		}
		return jsonMgmtResp(status, map[string]any{"error": message, "storage_health": ps.st.Health()})
	}
	resp := map[string]any{
		"records":        paged,
		"count":          len(paged),
		"has_more":       nextCursor != "",
		"storage_health": ps.st.Health(),
		"dropped":        ps.q.DroppedCount(),
	}
	if nextCursor != "" {
		resp["next_cursor"] = nextCursor
	}

	return jsonMgmtResp(200, resp)
}

// ----------------------------------------------------------------------------
// GET /summary — token aggregates per provider/model/auth_index
// ----------------------------------------------------------------------------

type summaryKey struct {
	Provider  string
	Model     string
	AuthIndex string
}

type summaryBucket struct {
	Provider            string `json:"provider"`
	Model               string `json:"model"`
	AuthIndex           string `json:"auth_index"`
	Requests            int64  `json:"requests"`
	Failures            int64  `json:"failures"`
	InputTokens         int64  `json:"input_tokens"`
	FreshInputTokens    int64  `json:"fresh_input_tokens"`
	OutputTokens        int64  `json:"output_tokens"`
	ReasoningTokens     int64  `json:"reasoning_tokens"`
	CachedTokens        int64  `json:"cached_tokens"`
	CacheReadTokens     int64  `json:"cache_read_tokens"`
	CacheCreationTokens int64  `json:"cache_creation_tokens"`
	TotalTokens         int64  `json:"total_tokens"`
	GenerateCount       int64  `json:"generate_count"` // Generate=true records
}

type aggregateItem struct {
	Key                 string `json:"key"`
	Requests            int64  `json:"requests"`
	Failures            int64  `json:"failures"`
	InputTokens         int64  `json:"input_tokens"`
	FreshInputTokens    int64  `json:"fresh_input_tokens"`
	OutputTokens        int64  `json:"output_tokens"`
	ReasoningTokens     int64  `json:"reasoning_tokens"`
	CachedTokens        int64  `json:"cached_tokens"`
	CacheReadTokens     int64  `json:"cache_read_tokens"`
	CacheCreationTokens int64  `json:"cache_creation_tokens"`
	TotalTokens         int64  `json:"total_tokens"`
	GenerateCount       int64  `json:"generate_count"`
}

type tokenSemantics struct {
	Definition        string `json:"definition"`
	TotalSource       string `json:"total_source"`
	CachedRelation    string `json:"cached_relation"`
	ReasoningRelation string `json:"reasoning_relation"`
}

func handleGetSummary(req managementHandleRequest) []byte {
	status, body := summaryResult(req, false)
	return jsonMgmtResp(status, body)
}

// resourceSummaryBucket is summaryBucket without auth_index, which the
// unauthenticated resource routes must never expose.
type resourceSummaryBucket struct {
	Provider            string `json:"provider"`
	Model               string `json:"model"`
	Requests            int64  `json:"requests"`
	Failures            int64  `json:"failures"`
	InputTokens         int64  `json:"input_tokens"`
	FreshInputTokens    int64  `json:"fresh_input_tokens"`
	OutputTokens        int64  `json:"output_tokens"`
	ReasoningTokens     int64  `json:"reasoning_tokens"`
	CachedTokens        int64  `json:"cached_tokens"`
	CacheReadTokens     int64  `json:"cache_read_tokens"`
	CacheCreationTokens int64  `json:"cache_creation_tokens"`
	TotalTokens         int64  `json:"total_tokens"`
	GenerateCount       int64  `json:"generate_count"`
}

func toResourceSummaryBuckets(in []*summaryBucket) []resourceSummaryBucket {
	out := make([]resourceSummaryBucket, len(in))
	for i, b := range in {
		out[i] = resourceSummaryBucket{b.Provider, b.Model, b.Requests, b.Failures, b.InputTokens, b.FreshInputTokens,
			b.OutputTokens, b.ReasoningTokens, b.CachedTokens, b.CacheReadTokens, b.CacheCreationTokens, b.TotalTokens, b.GenerateCount}
	}
	return out
}

// summaryResult computes the aggregate response body. Records are scanned in
// place (no copy, no sort): aggregation is order-independent.
func summaryResult(req managementHandleRequest, resource bool) (int, any) {
	stateMu.RLock()
	ps := state
	stateMu.RUnlock()
	if ps == nil {
		return 503, map[string]string{"error": "store not ready"}
	}

	opts, err := parseFilterOptions(req.Query)
	if err != nil {
		return 400, map[string]string{"error": err.Error()}
	}
	if resource {
		if err := boundResourceSummary(&opts); err != nil {
			return 400, map[string]string{"error": err.Error()}
		}
	}
	dailyTimezoneLocation, err := parseTimezoneLocation(req.Query)
	if err != nil {
		return 400, map[string]string{"error": err.Error()}
	}
	dailyTimezoneOffset, err := parseTimezoneOffset(req.Query)
	if err != nil {
		return 400, map[string]string{"error": err.Error()}
	}

	ps.lifecycle.RLock()
	defer ps.lifecycle.RUnlock()
	recs, err := ps.st.ReadAllView() // read-only: aggregation never mutates records
	if err != nil {
		log.Printf("usage-audit: summary read: %v", err)
		return 500, map[string]any{"error": "read failed", "storage_health": ps.st.Health()}
	}

	buckets := make(map[summaryKey]*summaryBucket)
	dailyMap := make(map[string]*aggregateItem)
	byProviderMap := make(map[string]*aggregateItem)
	byEndpointMap := make(map[string]*aggregateItem)
	byModelMap := make(map[string]*aggregateItem)

	matched := 0
	dayCache := make(map[int64]string) // unix-hour -> formatted day (zone offsets are whole minutes; DST changes on hour edges)
	for i := range recs {
		r := &recs[i]
		if !store.MatchRecord(r, &opts) {
			continue
		}
		matched++
		fresh := freshInputTokens(r)
		// Existing summary bucket
		k := summaryKey{Provider: r.Provider, Model: r.Model, AuthIndex: r.AuthIndex}
		b := buckets[k]
		if b == nil {
			b = &summaryBucket{Provider: r.Provider, Model: r.Model, AuthIndex: r.AuthIndex}
			buckets[k] = b
		}
		b.Requests++
		if r.Failed {
			b.Failures++
		}
		b.InputTokens += r.InputTokens
		b.FreshInputTokens += fresh
		b.OutputTokens += r.OutputTokens
		b.ReasoningTokens += r.ReasoningTokens
		b.CachedTokens += r.CachedTokens
		b.CacheReadTokens += r.CacheReadTokens
		b.CacheCreationTokens += r.CacheCreationTokens
		b.TotalTokens += r.TotalTokens
		if r.Generate {
			b.GenerateCount++
		}

		// Aggregate helpers
		addAgg := func(m map[string]*aggregateItem, key string) {
			if key == "" {
				key = "unknown"
			}
			agg := m[key]
			if agg == nil {
				agg = &aggregateItem{Key: key}
				m[key] = agg
			}
			agg.Requests++
			if r.Failed {
				agg.Failures++
			}
			agg.InputTokens += r.InputTokens
			agg.FreshInputTokens += fresh
			agg.OutputTokens += r.OutputTokens
			agg.ReasoningTokens += r.ReasoningTokens
			agg.CachedTokens += r.CachedTokens
			agg.CacheReadTokens += r.CacheReadTokens
			agg.CacheCreationTokens += r.CacheCreationTokens
			agg.TotalTokens += r.TotalTokens
			if r.Generate {
				agg.GenerateCount++
			}
		}

		// Daily key (YYYY-MM-DD)
		ts := r.EffectiveTimestamp()
		hour := ts.Unix() / 3600
		if ts.Unix() < 0 {
			hour = -1 << 62 // pre-1970 timestamps: bypass cache
		}
		dayKey, ok := dayCache[hour]
		if !ok {
			if dailyTimezoneLocation != nil {
				dayKey = ts.In(dailyTimezoneLocation).Format("2006-01-02")
			} else {
				dayKey = ts.UTC().Add(dailyTimezoneOffset).Format("2006-01-02")
			}
			if hour != -1<<62 {
				dayCache[hour] = dayKey
			}
		}
		addAgg(dailyMap, dayKey)

		// Provider
		addAgg(byProviderMap, r.Provider)

		// Model
		addAgg(byModelMap, r.Model)

		// Endpoint: executor_type 우선, 없으면 source, 둘 다 없으면 unknown
		endpoint := r.ExecutorType
		if endpoint == "" {
			endpoint = r.Source
		}
		if endpoint == "" {
			endpoint = "unknown"
		}
		addAgg(byEndpointMap, endpoint)
		if resource && len(buckets)+len(dailyMap)+len(byProviderMap)+len(byEndpointMap)+len(byModelMap) > resourceSummaryMaxBuckets {
			return 400, map[string]string{"error": "summary bucket limit exceeded; narrow range or filters"}
		}
	}

	result := make([]*summaryBucket, 0, len(buckets))
	for _, b := range buckets {
		result = append(result, b)
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Provider != result[j].Provider {
			return result[i].Provider < result[j].Provider
		}
		if result[i].Model != result[j].Model {
			return result[i].Model < result[j].Model
		}
		return result[i].AuthIndex < result[j].AuthIndex
	})

	toAggList := func(m map[string]*aggregateItem) []*aggregateItem {
		list := make([]*aggregateItem, 0, len(m))
		for _, item := range m {
			list = append(list, item)
		}
		sort.Slice(list, func(i, j int) bool {
			return list[i].Key < list[j].Key
		})
		return list
	}

	var summaryOut any = result
	if resource {
		summaryOut = toResourceSummaryBuckets(result)
	}
	return 200, map[string]any{
		"summary":     summaryOut,
		"daily":       toAggList(dailyMap),
		"by_provider": toAggList(byProviderMap),
		"by_endpoint": toAggList(byEndpointMap),
		"by_model":    toAggList(byModelMap),
		"token_semantics": tokenSemantics{
			Definition:        "fresh_input_tokens is provider-dependent: Claude/Anthropic input; otherwise max(0,input-cached_tokens), falling back to cache_read+cache_creation.",
			TotalSource:       "record.total_tokens",
			CachedRelation:    "subset_of_input",
			ReasoningRelation: "provider_reported_in_total_when_applicable",
		},
		"total_count":    matched,
		"storage_health": ps.st.Health(),
		"dropped":        ps.q.DroppedCount(),
	}
}

// freshInputTokens is a documented heuristic: Claude/Anthropic reports cache
// independently, so all input is fresh; other providers report cache as an
// input subset. When cached_tokens is absent, cache_read+cache_creation is used.
func freshInputTokens(r *store.LogRecord) int64 {
	if isClaudeFamily(r) {
		return maxNonNegative(r.InputTokens)
	}
	cached := r.CachedTokens
	if cached == 0 {
		cached = r.CacheReadTokens + r.CacheCreationTokens
	}
	fresh := r.InputTokens - cached
	return maxNonNegative(fresh)
}

func isClaudeFamily(r *store.LogRecord) bool {
	for _, f := range [...]string{r.Provider, r.AuthType, r.ExecutorType, r.Source, r.Model} {
		if containsFold(f, "claude") || containsFold(f, "anthropic") {
			return true
		}
	}
	return false
}

// containsFold reports whether s contains the lowercase ASCII needle,
// ignoring ASCII case, without allocating.
func containsFold(s, needle string) bool {
	n := len(needle)
	for i := 0; i+n <= len(s); i++ {
		j := 0
		for ; j < n; j++ {
			c := s[i+j]
			if 'A' <= c && c <= 'Z' {
				c += 'a' - 'A'
			}
			if c != needle[j] {
				break
			}
		}
		if j == n {
			return true
		}
	}
	return false
}

func maxNonNegative(v int64) int64 {
	if v < 0 {
		return 0
	}
	return v
}

type resourceRecord struct {
	RequestedAt   time.Time `json:"requested_at,omitempty"`
	Provider      string    `json:"provider,omitempty"`
	Model         string    `json:"model,omitempty"`
	Alias         string    `json:"alias,omitempty"`
	ExecutorType  string    `json:"executor_type,omitempty"`
	Source        string    `json:"source,omitempty"`
	Input         int64     `json:"input"`
	FreshInput    int64     `json:"fresh_input"`
	Cached        int64     `json:"cached"`
	Output        int64     `json:"output"`
	Reasoning     int64     `json:"reasoning"`
	Total         int64     `json:"total"`
	Latency       float64   `json:"latency"`
	TTFT          float64   `json:"ttft"`
	Failed        bool      `json:"failed"`
	FailureStatus int       `json:"failure_status,omitempty"`
}

func toResourceRecord(r store.LogRecord) resourceRecord {
	return resourceRecord{RequestedAt: r.EffectiveTimestamp(), Provider: r.Provider, Model: r.Model, Alias: r.Alias,
		ExecutorType: r.ExecutorType, Source: r.Source, Input: r.InputTokens, FreshInput: freshInputTokens(&r),
		Cached: r.CachedTokens, Output: r.OutputTokens, Reasoning: r.ReasoningTokens, Total: r.TotalTokens,
		Latency: r.LatencyMs, TTFT: r.TTFTMs, Failed: r.Failed, FailureStatus: r.FailureStatus}
}

func handleGetDashboardRecords(req managementHandleRequest) []byte {
	stateMu.RLock()
	ps := state
	stateMu.RUnlock()
	if ps == nil {
		return jsonResourceResp(503, map[string]string{"error": "store not ready"})
	}
	opts, err := parseFilterOptions(req.Query)
	if err != nil {
		return jsonResourceResp(400, map[string]string{"error": err.Error()})
	}
	limit := 20
	if v := queryFirst(req.Query, "limit"); v != "" {
		n, e := strconv.Atoi(v)
		if e != nil || n < 1 || n > 100 {
			return jsonResourceResp(400, map[string]string{"error": "invalid limit; must be 1..100"})
		}
		limit = n
	}
	ps.lifecycle.RLock()
	defer ps.lifecycle.RUnlock()
	page, nextCursor, err := recordsPage(ps.st, opts, "resource", queryFirst(req.Query, "cursor"), limit)
	if err != nil {
		status, message := recordsPageFailure(err)
		if status == 500 {
			log.Printf("usage-audit: resource records read: %v", err)
		}
		return jsonResourceResp(status, map[string]any{"error": message, "storage_health": ps.st.Health()})
	}
	out := make([]resourceRecord, 0, len(page))
	for _, r := range page {
		out = append(out, toResourceRecord(r))
	}
	resp := map[string]any{"records": out, "count": len(out), "has_more": nextCursor != "", "dropped": ps.q.DroppedCount(), "storage_health": ps.st.Health()}
	if nextCursor != "" {
		resp["next_cursor"] = nextCursor
	}
	return jsonResourceResp(200, resp)
}

func handleGetDashboardSummary(req managementHandleRequest) []byte {
	status, body := summaryResult(req, true)
	return jsonResourceResp(status, body)
}
func fontResourceResp(font []byte) []byte {
	return OKEnvelope(managementResponse{
		StatusCode: 200,
		Headers: map[string][]string{
			"Content-Type":           {"font/woff2"},
			"Cache-Control":          {"public, max-age=31536000, immutable"},
			"X-Content-Type-Options": {"nosniff"},
		},
		Body: font,
	})
}

func jsonResourceResp(status int, body any) []byte {
	var b []byte
	var e error
	if raw, ok := body.([]byte); ok {
		b = raw
	} else {
		b, e = json.Marshal(body)
	}
	if e != nil {
		b = []byte(`{"error":"marshal failed"}`)
	}
	return OKEnvelope(managementResponse{StatusCode: status, Headers: map[string][]string{"Content-Type": {"application/json"}, "Cache-Control": {"no-store"}, "X-Content-Type-Options": {"nosniff"}}, Body: b})
}

// ----------------------------------------------------------------------------
// DELETE /records — clear active and rotated logs under the Store writer lock
// ----------------------------------------------------------------------------

func handleDeleteRecords() []byte {
	stateMu.RLock()
	ps := state
	stateMu.RUnlock()
	if ps == nil {
		return jsonMgmtResp(503, map[string]string{"error": "store not ready"})
	}
	ps.lifecycle.RLock()
	defer ps.lifecycle.RUnlock()
	ps.epochMu.Lock()
	defer ps.epochMu.Unlock()
	// Advance even on filesystem failure; old queued data must never be replayed.
	ps.epoch.Add(1)
	if err := ps.st.Clear(); err != nil {
		log.Printf("usage-audit: delete records: %v", err)
		return jsonMgmtResp(500, map[string]string{"error": "truncate failed"})
	}
	return jsonMgmtResp(200, map[string]string{"status": "ok"})
}

// ----------------------------------------------------------------------------
// Shutdown — called from cliproxyPluginShutdown in main.go
// ----------------------------------------------------------------------------

// Shutdown drains the queue, flushes, and closes the store. Safe to call multiple times.
func Shutdown() {
	stateMu.Lock()
	ps := state
	state = nil
	stateMu.Unlock()
	if ps != nil {
		ps.shutdown()
	}
}

// ----------------------------------------------------------------------------
// Internal helpers
// ----------------------------------------------------------------------------

func jsonMgmtResp(status int, body any) []byte {
	b, err := json.Marshal(body)
	if err != nil {
		b = []byte(`{"error":"marshal failed"}`)
	}
	return OKEnvelope(managementResponse{
		StatusCode: status,
		Headers:    map[string][]string{"Content-Type": {"application/json"}},
		Body:       b,
	})
}

func jsonStr(s string) []byte { return []byte(s) }

func queryFirst(query map[string][]string, key string) string {
	if vals, ok := query[key]; ok && len(vals) > 0 {
		return strings.TrimSpace(vals[0])
	}
	return ""
}
