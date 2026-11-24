package plugin_test

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/dohwi/cliproxyapi-plugin-usage-audit/internal/plugin"
	"github.com/dohwi/cliproxyapi-plugin-usage-audit/internal/store"
	"github.com/router-for-me/CLIProxyAPI/v8/sdk/pluginapi"
)

func TestRecordsPageByNumberAndTotalCount(t *testing.T) {
	dir := t.TempDir()
	st, err := store.Open(filepath.Join(dir, "usage.jsonl"), 0, 0)
	if err != nil {
		t.Fatalf("store.Open: %v", err)
	}
	base := time.Date(2026, 11, 24, 12, 0, 0, 0, time.UTC)
	// Write 52 records (3 pages of 20), including a timestamp tie and two providers.
	for i := 0; i < 52; i++ {
		ts := base.Add(time.Duration(i) * time.Second)
		if i == 10 || i == 11 {
			ts = base.Add(10 * time.Second) // tie to verify sequence tie-breaking
		}
		prov := "anthropic"
		if i%2 == 1 {
			prov = "openai"
		}
		if err := st.Append(store.LogRecord{
			T:            ts,
			RequestedAt:  ts,
			Provider:     prov,
			ExecutorType: "direct",
			Model:        fmt.Sprintf("m-%02d", i),
			TotalTokens:  100 + int64(i),
		}); err != nil {
			t.Fatalf("append %d: %v", i, err)
		}
	}
	_ = st.Close()

	t.Setenv("USAGE_AUDIT_DATA_DIR", dir)
	callRegister(t)

	type recItem struct {
		Model string `json:"model"`
	}
	type pageResp struct {
		Records    []recItem `json:"records"`
		Count      int       `json:"count"`
		TotalCount int       `json:"total_count"`
		Page       int       `json:"page"`
		HasMore    bool      `json:"has_more"`
		NextCursor string    `json:"next_cursor"`
		Code       string    `json:"code"`
	}

	call := func(path string, query map[string][]string) (int, pageResp) {
		t.Helper()
		resp := callManagement(t, pluginapi.ManagementRequest{
			Method: "GET",
			Path:   "/v0/management/plugins/usage-audit" + path,
			Query:  query,
		})
		var out pageResp
		if err := json.Unmarshal(resp.Body, &out); err != nil {
			t.Fatalf("unmarshal %s (%d): %v body=%s", path, resp.StatusCode, err, string(resp.Body))
		}
		return resp.StatusCode, out
	}

	// 1. Collect all 3 pages via cursor pagination as ground truth.
	var cursorModels [][]string
	cursor := ""
	for p := 1; p <= 3; p++ {
		q := map[string][]string{"limit": {"20"}}
		if cursor != "" {
			q["cursor"] = []string{cursor}
		}
		status, out := call("/data/records", q)
		if status != 200 {
			t.Fatalf("cursor page %d status=%d", p, status)
		}
		if out.TotalCount != 52 {
			t.Fatalf("cursor page %d total_count=%d want 52", p, out.TotalCount)
		}
		models := make([]string, len(out.Records))
		for i, r := range out.Records {
			models[i] = r.Model
		}
		cursorModels = append(cursorModels, models)
		cursor = out.NextCursor
	}
	if cursor != "" {
		t.Fatalf("expected terminal cursor after page 3, got %q", cursor)
	}

	// 2. Direct page=1, 2, 3 must match cursor pages identically on both /data/records and /records.
	for _, ep := range []string{"/data/records", "/records"} {
		for p := 1; p <= 3; p++ {
			status, out := call(ep, map[string][]string{"limit": {"20"}, "page": {fmt.Sprintf("%d", p)}})
			if status != 200 {
				t.Fatalf("%s page=%d status=%d", ep, p, status)
			}
			wantLen := 20
			wantMore := true
			if p == 3 {
				wantLen = 12
				wantMore = false
			}
			if out.Count != wantLen || len(out.Records) != wantLen || out.TotalCount != 52 || out.HasMore != wantMore || out.Page != p {
				t.Fatalf("%s page=%d got count=%d total=%d hasMore=%v page=%d", ep, p, out.Count, out.TotalCount, out.HasMore, out.Page)
			}
			for i, r := range out.Records {
				if r.Model != cursorModels[p-1][i] {
					t.Fatalf("%s page=%d idx=%d got model=%s want %s", ep, p, i, r.Model, cursorModels[p-1][i])
				}
			}
		}
	}

	// 3. Out-of-range page returns empty slice with accurate total_count.
	status, out := call("/data/records", map[string][]string{"limit": {"20"}, "page": {"99"}})
	if status != 200 || out.Count != 0 || len(out.Records) != 0 || out.TotalCount != 52 || out.HasMore {
		t.Fatalf("out-of-range page got status=%d %+v", status, out)
	}

	// 4. Filter + page works accurately (26 openai records -> 2 pages: 20 + 6).
	status, out = call("/data/records", map[string][]string{"limit": {"20"}, "provider": {"openai"}, "page": {"2"}})
	if status != 200 || out.TotalCount != 26 || out.Count != 6 || out.HasMore {
		t.Fatalf("filtered page 2 got status=%d %+v", status, out)
	}

	// 5. Invalid page / combined page+cursor return 400.
	for _, tc := range []struct {
		q        map[string][]string
		wantCode string
	}{
		{map[string][]string{"page": {"0"}}, "invalid_page"},
		{map[string][]string{"page": {"-3"}}, "invalid_page"},
		{map[string][]string{"page": {"abc"}}, "invalid_page"},
		{map[string][]string{"page": {"1"}, "cursor": {"abc.def"}}, "invalid_cursor"},
	} {
		stCode, res := call("/data/records", tc.q)
		if stCode != 400 || res.Code != tc.wantCode {
			t.Fatalf("query %v got status=%d code=%q want 400/%q", tc.q, stCode, res.Code, tc.wantCode)
		}
	}
}

func TestComputePaginationSlotsInvariants(t *testing.T) {
	nodePath, err := exec.LookPath("node")
	if err != nil || nodePath == "" {
		t.Skip("node not found")
	}
	html := plugin.DashboardHTML
	idx := strings.Index(html, "function computePaginationSlots(")
	if idx < 0 {
		t.Fatal("computePaginationSlots not found in DashboardHTML")
	}
	end := strings.Index(html[idx:], "\nfunction resolvedRecordsTotal(")
	if end < 0 {
		t.Fatal("end of computePaginationSlots not found")
	}
	fnSrc := html[idx : idx+end]

	script := fnSrc + `
function assert(cond, msg) { if (!cond) { console.error("FAIL:", msg); process.exit(1); } }
assert(JSON.stringify(computePaginationSlots(1, 0)) === '[]', '0 pages');
assert(JSON.stringify(computePaginationSlots(1, 1)) === '[]', '1 page');
assert(JSON.stringify(computePaginationSlots(1, 3)) === '[1,2,3]', '3 pages p1');
assert(JSON.stringify(computePaginationSlots(3, 3)) === '[1,2,3]', '3 pages p3');
assert(JSON.stringify(computePaginationSlots(4, 5)) === '[1,2,3,4,5]', '5 pages');
assert(JSON.stringify(computePaginationSlots(6, 6)) === '[2,3,4,5,6]', '6 pages tail fill');
assert(JSON.stringify(computePaginationSlots(5, 62)) === '[1,2,3,4,5]', '62 pages p5 stays in first block');
assert(JSON.stringify(computePaginationSlots(6, 62)) === '[6,7,8,9,10]', '62 pages p6 moves to second block');
assert(JSON.stringify(computePaginationSlots(62, 62)) === '[58,59,60,61,62]', '62 pages p62 keeps 5 slots');

for (const total of [6, 7, 10, 62]) {
  for (let cur = 1; cur <= total; cur++) {
    const slots = computePaginationSlots(cur, total);
    assert(slots.length === 5, 'fixed 5 slots for total=' + total + ' cur=' + cur + ' got ' + JSON.stringify(slots));
    assert(slots.includes(cur), 'includes cur=' + cur + ' in ' + JSON.stringify(slots));
    for (let i = 1; i < slots.length; i++) {
      assert(slots[i] === slots[i - 1] + 1, 'contiguous pages for total=' + total + ' cur=' + cur + ': ' + JSON.stringify(slots));
    }
  }
}
console.log('OK');
`
	out, err := exec.Command(nodePath, "-e", script).CombinedOutput()
	if err != nil {
		t.Fatalf("node invariant test failed: %v\n%s", err, string(out))
	}
}
