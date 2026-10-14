// Package plugin_test verifies core plugin logic without the C ABI layer.
// Tests cover: registration, usage ingest, persistence, aggregation,
// queue-full overflow, and missing/zero field handling.
package plugin_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/dohwi/cliproxyapi-plugin-usage-audit/internal/plugin"
	"github.com/dohwi/cliproxyapi-plugin-usage-audit/internal/queue"
	"github.com/dohwi/cliproxyapi-plugin-usage-audit/internal/store"
	"github.com/router-for-me/CLIProxyAPI/v7/sdk/pluginapi"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func unmarshalOK(t *testing.T, raw []byte) map[string]any {
	t.Helper()
	var env map[string]any
	if err := json.Unmarshal(raw, &env); err != nil {
		t.Fatalf("unmarshal envelope: %v", err)
	}
	ok, _ := env["ok"].(bool)
	if !ok {
		t.Fatalf("expected ok=true, got: %s", raw)
	}
	result, _ := env["result"].(map[string]any)
	return result
}

func callRegister(t *testing.T) map[string]any {
	t.Helper()
	req, _ := json.Marshal(map[string]any{"schema_version": 6})
	raw, handled := plugin.HandleMethod("plugin.register", req)
	if !handled {
		t.Fatalf("plugin.register not handled")
	}
	return unmarshalOK(t, raw)
}

func callUsageHandle(t *testing.T, rec map[string]any) {
	t.Helper()
	req, _ := json.Marshal(rec)
	raw, handled := plugin.HandleMethod("usage.handle", req)
	if !handled {
		t.Fatalf("usage.handle not handled")
	}
	// response must be ok=true with empty result
	unmarshalOK(t, raw)
}

func callManagement(t *testing.T, req pluginapi.ManagementRequest) pluginapi.ManagementResponse {
	t.Helper()
	wire, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("marshal ManagementRequest: %v", err)
	}
	raw, handled := plugin.HandleMethod("management.handle", wire)
	if !handled {
		t.Fatal("management.handle not handled")
	}
	var env struct {
		OK     bool                         `json:"ok"`
		Result pluginapi.ManagementResponse `json:"result"`
	}
	if err := json.Unmarshal(raw, &env); err != nil {
		t.Fatalf("decode ManagementResponse envelope: %v", err)
	}
	if !env.OK {
		t.Fatalf("management envelope not ok: %s", raw)
	}
	return env.Result
}

func makeUsageRecord(provider, model, authIndex string, inputTokens, outputTokens int64, generate bool) map[string]any {
	return map[string]any{
		"Provider":    provider,
		"Model":       model,
		"Alias":       "",
		"APIKey":      "k-test",
		"AuthIndex":   authIndex,
		"AuthType":    "oauth",
		"Generate":    generate,
		"RequestedAt": time.Now().UTC().Format(time.RFC3339Nano),
		"Latency":     int64(500 * time.Millisecond),
		"TTFT":        int64(0),
		"Failed":      false,
		"Failure":     map[string]any{"StatusCode": 0, "Body": ""},
		"Detail": map[string]any{
			"InputTokens":         inputTokens,
			"OutputTokens":        outputTokens,
			"ReasoningTokens":     int64(0),
			"CachedTokens":        int64(0),
			"CacheReadTokens":     int64(0),
			"CacheCreationTokens": int64(0),
			"TotalTokens":         inputTokens + outputTokens,
		},
	}
}

// ---------------------------------------------------------------------------
// Test: plugin.register returns correct metadata and capabilities
// ---------------------------------------------------------------------------

func TestRegisterDecodesWithHostSDKMetadata(t *testing.T) {
	t.Setenv("USAGE_AUDIT_DATA_DIR", t.TempDir())
	req := []byte(`{"schema_version":6}`)
	raw, handled := plugin.HandleMethod("plugin.register", req)
	if !handled {
		t.Fatal("plugin.register not handled")
	}
	var envelope struct {
		OK     bool `json:"ok"`
		Result struct {
			SchemaVersion uint32             `json:"schema_version"`
			Metadata      pluginapi.Metadata `json:"metadata"`
			Capabilities  map[string]bool    `json:"capabilities"`
		} `json:"result"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		t.Fatalf("decode host registration envelope: %v", err)
	}
	if !envelope.OK || envelope.Result.SchemaVersion == 0 {
		t.Fatalf("invalid registration envelope: %s", raw)
	}
	if envelope.Result.Metadata.Name != "usage-audit" || envelope.Result.Metadata.Version == "" || envelope.Result.Metadata.Author == "" || envelope.Result.Metadata.GitHubRepository == "" {
		t.Fatalf("SDK metadata decode = %#v", envelope.Result.Metadata)
	}
	if !envelope.Result.Capabilities["usage_plugin"] || !envelope.Result.Capabilities["management_api"] {
		t.Fatalf("capability decode = %#v", envelope.Result.Capabilities)
	}
}

func TestRegister(t *testing.T) {
	t.Setenv("USAGE_AUDIT_DATA_DIR", t.TempDir())
	result := callRegister(t)

	sv, _ := result["schema_version"].(float64)
	if sv == 0 {
		t.Error("schema_version must be non-zero")
	}
	meta, _ := result["metadata"].(map[string]any)
	if name, _ := meta["Name"].(string); name != "usage-audit" {
		t.Errorf("unexpected plugin name: %q", name)
	}
	caps, _ := result["capabilities"].(map[string]any)
	if !caps["usage_plugin"].(bool) {
		t.Error("usage_plugin capability must be true")
	}
	if !caps["management_api"].(bool) {
		t.Error("management_api capability must be true")
	}
}

// ---------------------------------------------------------------------------
// Test: management.register returns expected routes
// ---------------------------------------------------------------------------

func TestManagementRequestAndResponseBodyBase64RoundTrip(t *testing.T) {
	t.Setenv("USAGE_AUDIT_DATA_DIR", t.TempDir())
	callRegister(t)

	requestBody := []byte(`{"probe":"request bytes"}`)
	req := pluginapi.ManagementRequest{
		Method:  http.MethodGet,
		Path:    "/v0/management/plugins/usage-audit/records",
		Headers: http.Header{"X-Test": []string{"roundtrip"}},
		Query:   url.Values{"limit": []string{"1"}},
		Body:    requestBody,
	}
	wire, err := json.Marshal(req)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(wire, []byte(base64.StdEncoding.EncodeToString(requestBody))) {
		t.Fatalf("request Body is not SDK []byte/base64 on wire: %s", wire)
	}
	var decodedReq pluginapi.ManagementRequest
	if err := json.Unmarshal(wire, &decodedReq); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(decodedReq.Body, requestBody) {
		t.Fatalf("request Body roundtrip: want %q got %q", requestBody, decodedReq.Body)
	}

	raw, handled := plugin.HandleMethod("management.handle", wire)
	if !handled {
		t.Fatal("management.handle not handled")
	}
	var rawEnv struct {
		OK     bool            `json:"ok"`
		Result json.RawMessage `json:"result"`
	}
	if err := json.Unmarshal(raw, &rawEnv); err != nil || !rawEnv.OK {
		t.Fatalf("bad response envelope: err=%v body=%s", err, raw)
	}
	var wireResp struct {
		StatusCode int    `json:"StatusCode"`
		Body       string `json:"Body"`
	}
	if err := json.Unmarshal(rawEnv.Result, &wireResp); err != nil {
		t.Fatal(err)
	}
	responseBody, err := base64.StdEncoding.DecodeString(wireResp.Body)
	if err != nil {
		t.Fatalf("response Body is not SDK []byte/base64: %v (%q)", err, wireResp.Body)
	}
	if wireResp.StatusCode != http.StatusOK || !json.Valid(responseBody) {
		t.Fatalf("unexpected response: status=%d body=%q", wireResp.StatusCode, responseBody)
	}
}

func TestEmptyRecordsIsJSONArray(t *testing.T) {
	t.Setenv("USAGE_AUDIT_DATA_DIR", t.TempDir())
	callRegister(t)

	resp := callManagement(t, pluginapi.ManagementRequest{
		Method: http.MethodGet,
		Path:   "/v0/management/plugins/usage-audit/records",
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status=%d body=%s", resp.StatusCode, resp.Body)
	}
	var body struct {
		Records []store.LogRecord `json:"records"`
		Count   int               `json:"count"`
	}
	if err := json.Unmarshal(resp.Body, &body); err != nil {
		t.Fatal(err)
	}
	if body.Records == nil || body.Count != 0 {
		t.Fatalf("empty records must be [] with count 0: %s", resp.Body)
	}
	if !bytes.Contains(resp.Body, []byte(`"records":[]`)) {
		t.Fatalf("records encoded as null: %s", resp.Body)
	}
}

func TestManagementRegister(t *testing.T) {
	t.Setenv("USAGE_AUDIT_DATA_DIR", t.TempDir())
	callRegister(t)

	raw, handled := plugin.HandleMethod("management.register", []byte(`{}`))
	if !handled {
		t.Fatal("management.register not handled")
	}
	result := unmarshalOK(t, raw)
	routes, _ := result["routes"].([]any)
	if len(routes) < 3 {
		t.Errorf("expected >=3 routes, got %d", len(routes))
	}
	methods := map[string]int{}
	for _, r := range routes {
		rm, _ := r.(map[string]any)
		methods[rm["Method"].(string)]++
	}
	if methods["GET"] < 2 {
		t.Errorf("expected >=2 GET routes")
	}
	if methods["DELETE"] < 1 {
		t.Errorf("expected >=1 DELETE route")
	}

	// Resource with Menu label must be declared so the management console
	// renders it in the left sidebar via GET /v0/management/plugins.
	resources, _ := result["resources"].([]any)
	if len(resources) != 4 {
		t.Fatalf("expected exactly 4 resources (dashboard, self-hosted fonts), got %d: %s", len(resources), raw)
	}
	// /v0/resource/... is unauthenticated: it may only carry the static shell.
	for _, r := range resources {
		rm, _ := r.(map[string]any)
		p, _ := rm["Path"].(string)
		if strings.Contains(p, "/data") {
			t.Errorf("unauthenticated resource route %q must not serve data", p)
		}
	}
	// The data routes the dashboard polls must be management-authenticated.
	wantMgmt := map[string]bool{
		"GET /v0/management/plugins/usage-audit/data/summary": false,
		"GET /v0/management/plugins/usage-audit/data/records": false,
	}
	for _, r := range routes {
		rm, _ := r.(map[string]any)
		key := rm["Method"].(string) + " " + rm["Path"].(string)
		if _, ok := wantMgmt[key]; ok {
			wantMgmt[key] = true
		}
	}
	for key, seen := range wantMgmt {
		if !seen {
			t.Errorf("missing management route %s", key)
		}
	}
	res, _ := resources[0].(map[string]any)
	if res["Path"] != "/dashboard" {
		t.Errorf("resource Path = %v, want /dashboard", res["Path"])
	}
	if res["Menu"] != "토큰 사용량" {
		t.Errorf("resource Menu = %v, want 토큰 사용량", res["Menu"])
	}
	if desc, _ := res["Description"].(string); desc == "" {
		t.Error("resource Description must be non-empty")
	}
}

// ---------------------------------------------------------------------------
// Test: browser-navigable resource request serves the dashboard HTML
// ---------------------------------------------------------------------------

func TestResourceDashboardServedViaHostResourcePath(t *testing.T) {
	t.Setenv("USAGE_AUDIT_DATA_DIR", t.TempDir())
	callRegister(t)

	// Host dispatches resource requests to management.handle with the full
	// browser URL path /v0/resource/plugins/<pluginID>/dashboard.
	for _, path := range []string{
		"/v0/resource/plugins/usage-audit/dashboard",
		"/dashboard",
	} {
		resp := callManagement(t, pluginapi.ManagementRequest{
			Method: http.MethodGet,
			Path:   path,
		})
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("%s: status=%d", path, resp.StatusCode)
		}
		if ct := resp.Headers.Get("Content-Type"); !strings.HasPrefix(ct, "text/html") {
			t.Errorf("%s: Content-Type = %q, want text/html", path, ct)
		}
		if !bytes.Contains(resp.Body, []byte("<html")) {
			t.Errorf("%s: body is not the dashboard HTML", path)
		}
	}
}

// ---------------------------------------------------------------------------
// Test: usage.handle ingest → persistence → /records query
// ---------------------------------------------------------------------------

func TestIngestAndPersist(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("USAGE_AUDIT_DATA_DIR", dir)
	callRegister(t)

	callUsageHandle(t, makeUsageRecord("antigravity", "gemini-3.8-flash-high", "idx-001", 100, 50, true))
	callUsageHandle(t, makeUsageRecord("codex", "gpt-6-astra", "idx-002", 200, 80, true))

	// Allow async consumer to flush.
	time.Sleep(200 * time.Millisecond)
	plugin.Shutdown()

	// Read JSONL directly.
	recs, err := store.ReadAll(filepath.Join(dir, "usage.jsonl"))
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if len(recs) != 2 {
		t.Fatalf("expected 2 records, got %d", len(recs))
	}
	if recs[0].Provider != "antigravity" {
		t.Errorf("provider mismatch: %s", recs[0].Provider)
	}
	if recs[0].InputTokens != 100 {
		t.Errorf("input_tokens mismatch: %d", recs[0].InputTokens)
	}
	if recs[0].AuthIndex != "idx-001" {
		t.Errorf("auth_index mismatch: %s", recs[0].AuthIndex)
	}
}

// ---------------------------------------------------------------------------
// Test: zero-token Detail is stored as-is (zero ≠ missing)
// ---------------------------------------------------------------------------

func TestZeroTokensStoredExplicitly(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("USAGE_AUDIT_DATA_DIR", dir)
	callRegister(t)

	callUsageHandle(t, makeUsageRecord("nim", "gpt-oss-20b", "idx-003", 0, 0, false))

	time.Sleep(200 * time.Millisecond)
	plugin.Shutdown()

	recs, err := store.ReadAll(filepath.Join(dir, "usage.jsonl"))
	if err != nil || len(recs) == 0 {
		t.Fatalf("expected 1 record, got %d, err=%v", len(recs), err)
	}
	r := recs[0]
	// Zero tokens must be stored, not missing.
	if r.InputTokens != 0 || r.OutputTokens != 0 {
		t.Errorf("zero tokens mutated: in=%d out=%d", r.InputTokens, r.OutputTokens)
	}
	// Generate=false must be preserved.
	if r.Generate {
		t.Error("Generate=false was not preserved (stored as true)")
	}
}

// ---------------------------------------------------------------------------
// Test: queue overflow drops records and increments counter
// ---------------------------------------------------------------------------

func TestQueueOverflow(t *testing.T) {
	q := queue.New(4)
	defer q.Close()
	for i := 0; i < 20; i++ {
		q.Enqueue(queue.Record(strconv.Itoa(i)))
	}
	if got := q.DroppedCount(); got != 16 {
		t.Fatalf("drops=%d, want 16", got)
	}
	for i := 0; i < 4; i++ {
		select {
		case got := <-q.Consume():
			if string(got) != strconv.Itoa(i) {
				t.Fatalf("record=%q want=%d", got, i)
			}
		case <-time.After(time.Second):
			t.Fatal("missing queued record")
		}
	}
}

// ---------------------------------------------------------------------------
// Test: store rotation
// ---------------------------------------------------------------------------

func TestStoreRotation(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "usage.jsonl")
	// maxBytes=100 forces rotation after a few records.
	st, err := store.Open(logPath, 100, 2)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer st.Close()

	for i := 0; i < 10; i++ {
		err := st.Append(store.LogRecord{
			T:        time.Now().UTC(),
			Provider: "test",
			Model:    "model",
		})
		if err != nil {
			t.Fatalf("Append i=%d: %v", i, err)
		}
	}
	// After rotation, original and .1 must exist.
	if _, err := os.Stat(logPath); err != nil {
		t.Errorf("current log missing: %v", err)
	}
	if _, err := os.Stat(logPath + ".1"); err != nil {
		t.Errorf("rotated log .1 missing: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Test: /summary aggregation correctness
// ---------------------------------------------------------------------------

func TestSummaryRequestsCountAllEvents(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("USAGE_AUDIT_DATA_DIR", dir)
	callRegister(t)

	callUsageHandle(t, makeUsageRecord("provider", "model", "idx", 10, 5, true))
	callUsageHandle(t, makeUsageRecord("provider", "model", "idx", 0, 0, false))
	callUsageHandle(t, makeUsageRecord("provider", "model", "idx", 20, 10, true))
	time.Sleep(200 * time.Millisecond)

	resp := callManagement(t, pluginapi.ManagementRequest{
		Method: http.MethodGet,
		Path:   "/v0/management/plugins/usage-audit/summary",
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status=%d body=%s", resp.StatusCode, resp.Body)
	}
	var body struct {
		Summary []struct {
			Requests      int64 `json:"requests"`
			GenerateCount int64 `json:"generate_count"`
		} `json:"summary"`
	}
	if err := json.Unmarshal(resp.Body, &body); err != nil {
		t.Fatal(err)
	}
	if len(body.Summary) != 1 || body.Summary[0].Requests != 3 || body.Summary[0].GenerateCount != 2 {
		t.Fatalf("requests must count all events while generate_count counts generation: %s", resp.Body)
	}
}

func TestSummaryAggregation(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "usage.jsonl")
	st, _ := store.Open(logPath, store.DefaultMaxBytes, store.DefaultMaxRotation)

	now := time.Now().UTC()
	records := []store.LogRecord{
		{T: now, Provider: "antigravity", Model: "m1", AuthIndex: "a1", Generate: true, InputTokens: 100, OutputTokens: 50, TotalTokens: 150},
		{T: now, Provider: "antigravity", Model: "m1", AuthIndex: "a1", Generate: true, InputTokens: 200, OutputTokens: 80, TotalTokens: 280},
		{T: now, Provider: "codex", Model: "m2", AuthIndex: "a2", Generate: false, InputTokens: 0, OutputTokens: 0, TotalTokens: 0},
	}
	for _, r := range records {
		if err := st.Append(r); err != nil {
			t.Fatalf("Append: %v", err)
		}
	}
	_ = st.Close()

	all, err := store.ReadAll(logPath)
	if err != nil {
		t.Fatalf("ReadAll: %v", err)
	}
	if len(all) != 3 {
		t.Fatalf("expected 3, got %d", len(all))
	}

	// Manually aggregate (mirrors plugin.go summary logic).
	totalInput := int64(0)
	for _, r := range all {
		if r.Provider == "antigravity" {
			totalInput += r.InputTokens
		}
	}
	if totalInput != 300 {
		t.Errorf("aggregated input_tokens: want 300 got %d", totalInput)
	}

	// zero-token record must be present.
	found := false
	for _, r := range all {
		if r.Provider == "codex" && !r.Generate {
			found = true
			if r.InputTokens != 0 {
				t.Errorf("codex zero tokens changed: %d", r.InputTokens)
			}
		}
	}
	if !found {
		t.Error("codex Generate=false zero-token record not found")
	}
}

// ---------------------------------------------------------------------------
// Test: unknown method returns error envelope
// ---------------------------------------------------------------------------

func TestUnknownMethod(t *testing.T) {
	raw, handled := plugin.HandleMethod("does.not.exist", nil)
	if handled {
		t.Error("unknown method should return handled=false")
	}
	var env map[string]any
	_ = json.Unmarshal(raw, &env)
	if env["ok"].(bool) {
		t.Error("unknown method should return ok=false")
	}
}

// ---------------------------------------------------------------------------
// Test: no raw AuthID stored (only AuthIndex)
// ---------------------------------------------------------------------------

func TestAuthIDNotStored(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("USAGE_AUDIT_DATA_DIR", dir)
	callRegister(t)

	rec := makeUsageRecord("antigravity", "m1", "stable-idx", 10, 5, true)
	rec["AuthID"] = "raw-internal-uuid-must-not-appear"
	callUsageHandle(t, rec)

	time.Sleep(200 * time.Millisecond)
	plugin.Shutdown()

	raw, _ := os.ReadFile(filepath.Join(dir, "usage.jsonl"))
	if len(raw) == 0 {
		t.Fatal("log file empty")
	}
	if contains(raw, "raw-internal-uuid-must-not-appear") {
		t.Error("raw AuthID must not be persisted in log")
	}
	if !contains(raw, "stable-idx") {
		t.Error("AuthIndex must be persisted in log")
	}
}

func contains(haystack []byte, needle string) bool {
	return len(haystack) > 0 && string(haystack) != "" &&
		len(needle) > 0 &&
		indexOf(haystack, []byte(needle)) >= 0
}

func indexOf(haystack, needle []byte) int {
	for i := 0; i <= len(haystack)-len(needle); i++ {
		if string(haystack[i:i+len(needle)]) == string(needle) {
			return i
		}
	}
	return -1
}

// ---------------------------------------------------------------------------
// Test: records pagination & filtering
// ---------------------------------------------------------------------------

func TestRecordsPaginationAndFiltering(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("USAGE_AUDIT_DATA_DIR", dir)
	callRegister(t)

	for i := 1; i <= 5; i++ {
		rec := makeUsageRecord("prov1", fmt.Sprintf("m%d", i), "auth1", int64(i*10), 5, true)
		if i == 3 {
			rec["Failed"] = true
			rec["Provider"] = "prov2"
		}
		callUsageHandle(t, rec)
	}

	time.Sleep(200 * time.Millisecond)

	// Fetch with limit 2
	resp := callManagement(t, pluginapi.ManagementRequest{
		Method: "GET",
		Path:   "/v0/management/plugins/usage-audit/records",
		Query:  map[string][]string{"limit": {"2"}},
	})
	if resp.StatusCode != 200 {
		t.Fatalf("records status: %d", resp.StatusCode)
	}
	var data struct {
		Records    []store.LogRecord `json:"records"`
		Count      int               `json:"count"`
		HasMore    bool              `json:"has_more"`
		NextCursor string            `json:"next_cursor"`
	}
	if err := json.Unmarshal(resp.Body, &data); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if data.Count != 2 || len(data.Records) != 2 || !data.HasMore || data.NextCursor == "" {
		t.Fatalf("unexpected pagination: %+v", data)
	}

	// Fetch next page
	resp2 := callManagement(t, pluginapi.ManagementRequest{
		Method: "GET",
		Path:   "/v0/management/plugins/usage-audit/records",
		Query:  map[string][]string{"limit": {"2"}, "cursor": {data.NextCursor}},
	})
	var data2 struct {
		Records []store.LogRecord `json:"records"`
		Count   int               `json:"count"`
		HasMore bool              `json:"has_more"`
	}
	if err := json.Unmarshal(resp2.Body, &data2); err != nil {
		t.Fatalf("unmarshal page 2: %v", err)
	}
	if data2.Count != 2 || !data2.HasMore {
		t.Fatalf("unexpected page 2: %+v", data2)
	}

	// Verify status filter
	respFailed := callManagement(t, pluginapi.ManagementRequest{
		Method: "GET",
		Path:   "/v0/management/plugins/usage-audit/records",
		Query:  map[string][]string{"status": {"failed"}},
	})
	var dataFailed struct {
		Records []store.LogRecord `json:"records"`
		Count   int               `json:"count"`
	}
	_ = json.Unmarshal(respFailed.Body, &dataFailed)
	if dataFailed.Count != 1 || dataFailed.Records[0].Provider != "prov2" {
		t.Fatalf("status=failed filter failed: %+v", dataFailed)
	}

	// Verify load-more pages do not duplicate records (keyset cursor).
	seen := map[string]bool{}
	for _, r := range append(data.Records, data2.Records...) {
		id := r.T.Format(time.RFC3339Nano) + "|" + r.Model
		if seen[id] {
			t.Errorf("duplicate record across pages: %+v", r)
		}
		seen[id] = true
	}
	if len(seen) != 4 {
		t.Errorf("expected 4 unique records across pages 1+2, got %d", len(seen))
	}

	// Verify 400 on invalid query
	respBad := callManagement(t, pluginapi.ManagementRequest{
		Method: "GET",
		Path:   "/v0/management/plugins/usage-audit/records",
		Query:  map[string][]string{"status": {"invalid_status"}},
	})
	if respBad.StatusCode != 400 {
		t.Fatalf("expected 400 for bad status, got %d", respBad.StatusCode)
	}

	respBadDate := callManagement(t, pluginapi.ManagementRequest{
		Method: "GET",
		Path:   "/v0/management/plugins/usage-audit/records",
		Query:  map[string][]string{"from": {"not-a-date"}},
	})
	if respBadDate.StatusCode != 400 {
		t.Fatalf("expected 400 for bad date, got %d", respBadDate.StatusCode)
	}
}

// ---------------------------------------------------------------------------
// Test: summary aggregation with breakdown
// ---------------------------------------------------------------------------

func TestSummaryBreakdownAggregation(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("USAGE_AUDIT_DATA_DIR", dir)
	callRegister(t)

	r1 := makeUsageRecord("p1", "m1", "a1", 100, 50, true)
	r1["ExecutorType"] = "exec1"
	r1["Source"] = "src1"
	callUsageHandle(t, r1)

	r2 := makeUsageRecord("p2", "m2", "a2", 200, 100, false)
	r2["ExecutorType"] = ""
	r2["Source"] = "src2"
	callUsageHandle(t, r2)

	r3 := makeUsageRecord("p3", "m3", "a3", 300, 150, true)
	r3["ExecutorType"] = ""
	r3["Source"] = ""
	callUsageHandle(t, r3)

	time.Sleep(200 * time.Millisecond)

	resp := callManagement(t, pluginapi.ManagementRequest{
		Method: "GET",
		Path:   "/v0/management/plugins/usage-audit/summary",
	})
	if resp.StatusCode != 200 {
		t.Fatalf("summary status: %d", resp.StatusCode)
	}

	var sumResp struct {
		TotalCount int `json:"total_count"`
		Summary    []struct {
			Provider string `json:"provider"`
		} `json:"summary"`
		Daily []struct {
			Key      string `json:"key"`
			Requests int64  `json:"requests"`
		} `json:"daily"`
		ByProvider []struct {
			Key string `json:"key"`
		} `json:"by_provider"`
		ByEndpoint []struct {
			Key string `json:"key"`
		} `json:"by_endpoint"`
		ByModel []struct {
			Key string `json:"key"`
		} `json:"by_model"`
	}
	if err := json.Unmarshal(resp.Body, &sumResp); err != nil {
		t.Fatalf("unmarshal summary: %v", err)
	}

	if sumResp.TotalCount != 3 {
		t.Errorf("total_count want 3 got %d", sumResp.TotalCount)
	}
	if len(sumResp.Daily) == 0 {
		t.Errorf("daily empty")
	}
	if len(sumResp.ByProvider) != 3 {
		t.Errorf("by_provider want 3 got %d", len(sumResp.ByProvider))
	}
	if len(sumResp.ByModel) != 3 {
		t.Errorf("by_model want 3 got %d", len(sumResp.ByModel))
	}

	// Check endpoint derivation: exec1, src2, unknown
	endpoints := make(map[string]bool)
	for _, ep := range sumResp.ByEndpoint {
		endpoints[ep.Key] = true
	}
	if !endpoints["exec1"] || !endpoints["src2"] || !endpoints["unknown"] {
		t.Errorf("expected endpoints [exec1, src2, unknown], got %+v", endpoints)
	}
}

func TestSummaryDailyUsesIANAZoneAcrossDST(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("USAGE_AUDIT_DATA_DIR", dir)
	callRegister(t)

	record := makeUsageRecord("p1", "m1", "a1", 10, 5, true)
	// During New York daylight time this instant is 2026-07-01 00:30 EDT,
	// while a fixed winter UTC-5 offset would incorrectly place it on June 30.
	record["RequestedAt"] = "2026-07-01T04:30:00Z"
	callUsageHandle(t, record)
	time.Sleep(200 * time.Millisecond)

	resp := callManagement(t, pluginapi.ManagementRequest{
		Method: "GET",
		Path:   "/v0/management/plugins/usage-audit/summary",
		Query: map[string][]string{
			"time_zone": {"America/New_York"},
			"tz_offset": {"300"},
		},
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("summary status: %d body=%s", resp.StatusCode, resp.Body)
	}
	var sumResp struct {
		Daily []struct {
			Key string `json:"key"`
		} `json:"daily"`
	}
	if err := json.Unmarshal(resp.Body, &sumResp); err != nil {
		t.Fatalf("unmarshal summary: %v", err)
	}
	if len(sumResp.Daily) != 1 || sumResp.Daily[0].Key != "2026-07-01" {
		t.Fatalf("daily key = %+v, want DST-aware local date 2026-07-01", sumResp.Daily)
	}

	bad := callManagement(t, pluginapi.ManagementRequest{
		Method: "GET",
		Path:   "/v0/management/plugins/usage-audit/summary",
		Query:  map[string][]string{"time_zone": {"Not/A_Real_Zone"}},
	})
	if bad.StatusCode != http.StatusBadRequest {
		t.Fatalf("invalid time zone status = %d, want 400", bad.StatusCode)
	}
}

func TestSummaryDailyUsesBrowserTimezoneOffset(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("USAGE_AUDIT_DATA_DIR", dir)
	callRegister(t)

	record := makeUsageRecord("p1", "m1", "a1", 10, 5, true)
	// 00:30 KST is still the previous day in UTC. The browser offset must place
	// this request in the user's local calendar day.
	record["RequestedAt"] = "2026-03-12T00:30:00+09:00"
	callUsageHandle(t, record)
	time.Sleep(200 * time.Millisecond)

	resp := callManagement(t, pluginapi.ManagementRequest{
		Method: "GET",
		Path:   "/v0/management/plugins/usage-audit/summary",
		Query:  map[string][]string{"tz_offset": {"-540"}},
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("summary status: %d body=%s", resp.StatusCode, resp.Body)
	}

	var sumResp struct {
		Daily []struct {
			Key string `json:"key"`
		} `json:"daily"`
	}
	if err := json.Unmarshal(resp.Body, &sumResp); err != nil {
		t.Fatalf("unmarshal summary: %v", err)
	}
	if len(sumResp.Daily) != 1 || sumResp.Daily[0].Key != "2026-03-12" {
		t.Fatalf("daily key = %+v, want local date 2026-03-12", sumResp.Daily)
	}

	bad := callManagement(t, pluginapi.ManagementRequest{
		Method: "GET",
		Path:   "/v0/management/plugins/usage-audit/summary",
		Query:  map[string][]string{"tz_offset": {"not-a-number"}},
	})
	if bad.StatusCode != http.StatusBadRequest {
		t.Fatalf("invalid timezone offset status = %d, want 400", bad.StatusCode)
	}
}

// ---------------------------------------------------------------------------
// Test: single "endpoint" query matches ExecutorType OR Source (not AND)
// ---------------------------------------------------------------------------

func TestEndpointFilterMatchesExecutorTypeOrSource(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("USAGE_AUDIT_DATA_DIR", dir)
	callRegister(t)

	// r1: only ExecutorType set (Source empty)
	r1 := makeUsageRecord("p1", "m1", "a1", 10, 5, true)
	r1["ExecutorType"] = "exec-a"
	r1["Source"] = ""
	callUsageHandle(t, r1)

	// r2: only Source set (ExecutorType empty)
	r2 := makeUsageRecord("p2", "m2", "a2", 10, 5, true)
	r2["ExecutorType"] = ""
	r2["Source"] = "src-b"
	callUsageHandle(t, r2)

	time.Sleep(200 * time.Millisecond)

	fetchCount := func(query map[string][]string) int {
		t.Helper()
		resp := callManagement(t, pluginapi.ManagementRequest{
			Method: "GET",
			Path:   "/v0/management/plugins/usage-audit/records",
			Query:  query,
		})
		if resp.StatusCode != 200 {
			t.Fatalf("records status: %d body=%s", resp.StatusCode, resp.Body)
		}
		var data struct {
			Count int `json:"count"`
		}
		if err := json.Unmarshal(resp.Body, &data); err != nil {
			t.Fatal(err)
		}
		return data.Count
	}

	// OR semantics: a single "endpoint" value must match records where only
	// one of ExecutorType/Source carries that value. The previous UI sent both
	// source= and executor_type= together, which AND-ed them to zero results.
	if got := fetchCount(map[string][]string{"endpoint": {"exec-a"}}); got != 1 {
		t.Errorf("endpoint=exec-a want 1 record, got %d", got)
	}
	if got := fetchCount(map[string][]string{"endpoint": {"src-b"}}); got != 1 {
		t.Errorf("endpoint=src-b want 1 record, got %d", got)
	}
	if got := fetchCount(map[string][]string{"endpoint": {"no-such"}}); got != 0 {
		t.Errorf("endpoint=no-such want 0 records, got %d", got)
	}

	// Summary must honor the same endpoint filter.
	resp := callManagement(t, pluginapi.ManagementRequest{
		Method: "GET",
		Path:   "/v0/management/plugins/usage-audit/summary",
		Query:  map[string][]string{"endpoint": {"src-b"}},
	})
	var sum struct {
		TotalCount int `json:"total_count"`
	}
	if err := json.Unmarshal(resp.Body, &sum); err != nil {
		t.Fatal(err)
	}
	if sum.TotalCount != 1 {
		t.Errorf("summary endpoint filter want total_count=1, got %d", sum.TotalCount)
	}
}

// ---------------------------------------------------------------------------
// Test: to=YYYY-MM-DD includes the entire target day (23:59:59.999...)
// ---------------------------------------------------------------------------

func TestToDateIncludesFullDay(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("USAGE_AUDIT_DATA_DIR", dir)
	callRegister(t)

	// Record with RequestedAt at 23:59:59 UTC on the target day.
	day := time.Date(2025, 6, 15, 23, 59, 59, int(500*time.Millisecond), time.UTC)
	rec := makeUsageRecord("p1", "m1", "a1", 10, 5, true)
	rec["RequestedAt"] = day.Format(time.RFC3339Nano)
	callUsageHandle(t, rec)
	time.Sleep(200 * time.Millisecond)

	fetchCount := func(query map[string][]string) int {
		t.Helper()
		resp := callManagement(t, pluginapi.ManagementRequest{
			Method: "GET",
			Path:   "/v0/management/plugins/usage-audit/records",
			Query:  query,
		})
		if resp.StatusCode != 200 {
			t.Fatalf("records status: %d body=%s", resp.StatusCode, resp.Body)
		}
		var data struct {
			Count int `json:"count"`
		}
		if err := json.Unmarshal(resp.Body, &data); err != nil {
			t.Fatal(err)
		}
		return data.Count
	}

	// Bare-date "to" must be inclusive of the whole day.
	if got := fetchCount(map[string][]string{"to": {"2025-06-15"}}); got != 1 {
		t.Errorf("to=2025-06-15 should include 23:59:59 record, got %d", got)
	}
	// Previous day must exclude it.
	if got := fetchCount(map[string][]string{"to": {"2025-06-14"}}); got != 0 {
		t.Errorf("to=2025-06-14 should exclude the record, got %d", got)
	}
	// Full-day from+to range.
	if got := fetchCount(map[string][]string{"from": {"2025-06-15"}, "to": {"2025-06-15"}}); got != 1 {
		t.Errorf("from=to=2025-06-15 want 1 record, got %d", got)
	}
	// RFC3339 from the frontend (ISO with ms, local-tz based) must parse too.
	if got := fetchCount(map[string][]string{
		"from": {"2025-06-15T00:00:00.000Z"},
		"to":   {"2025-06-15T23:59:59.999Z"},
	}); got != 1 {
		t.Errorf("RFC3339 from/to range want 1 record, got %d", got)
	}
}

// ---------------------------------------------------------------------------
// Test: /summary token semantics and cached/reasoning token sums
// ---------------------------------------------------------------------------

func TestSummaryTokenSemanticsAndAggregation(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("USAGE_AUDIT_DATA_DIR", dir)
	callRegister(t)

	// r1: total is authoritative. cached is subset of input. reasoning is present.
	// total = 1000 + 200 + 300 = 1500 (or reported as 1500 directly).
	// cached = 400 (subset of 1000 input). cache_creation = 50. cache_read = 400.
	// TotalTokens must remain 1500, never recomputed to 1500 + 400 or 1500 + 300.
	r1 := map[string]any{
		"Provider":            "p1",
		"Model":               "m1",
		"AuthIndex":           "a1",
		"Generate":            true,
		"InputTokens":         int64(1000),
		"OutputTokens":        int64(200),
		"ReasoningTokens":     int64(300),
		"CachedTokens":        int64(400),
		"CacheReadTokens":     int64(400),
		"CacheCreationTokens": int64(50),
		"TotalTokens":         int64(1500),
		"ExecutorType":        "exec-p1",
	}
	callUsageHandle(t, r1)

	// r2: zero/null values
	r2 := map[string]any{
		"Provider":  "p2",
		"Model":     "m2",
		"AuthIndex": "a2",
		"Generate":  false,
	}
	callUsageHandle(t, r2)

	time.Sleep(200 * time.Millisecond)

	resp := callManagement(t, pluginapi.ManagementRequest{
		Method: http.MethodGet,
		Path:   "/v0/management/plugins/usage-audit/summary",
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status=%d body=%s", resp.StatusCode, resp.Body)
	}

	var data struct {
		TokenSemantics struct {
			TotalSource       string `json:"total_source"`
			CachedRelation    string `json:"cached_relation"`
			ReasoningRelation string `json:"reasoning_relation"`
		} `json:"token_semantics"`
		Summary []struct {
			Provider            string `json:"provider"`
			Model               string `json:"model"`
			InputTokens         int64  `json:"input_tokens"`
			OutputTokens        int64  `json:"output_tokens"`
			ReasoningTokens     int64  `json:"reasoning_tokens"`
			CachedTokens        int64  `json:"cached_tokens"`
			CacheReadTokens     int64  `json:"cache_read_tokens"`
			CacheCreationTokens int64  `json:"cache_creation_tokens"`
			TotalTokens         int64  `json:"total_tokens"`
		} `json:"summary"`
		Daily []struct {
			CachedTokens        int64 `json:"cached_tokens"`
			CacheReadTokens     int64 `json:"cache_read_tokens"`
			CacheCreationTokens int64 `json:"cache_creation_tokens"`
			ReasoningTokens     int64 `json:"reasoning_tokens"`
			TotalTokens         int64 `json:"total_tokens"`
		} `json:"daily"`
		ByProvider []struct {
			Key                 string `json:"key"`
			CachedTokens        int64  `json:"cached_tokens"`
			CacheReadTokens     int64  `json:"cache_read_tokens"`
			CacheCreationTokens int64  `json:"cache_creation_tokens"`
			ReasoningTokens     int64  `json:"reasoning_tokens"`
			TotalTokens         int64  `json:"total_tokens"`
		} `json:"by_provider"`
		ByEndpoint []struct {
			Key                 string `json:"key"`
			CachedTokens        int64  `json:"cached_tokens"`
			CacheReadTokens     int64  `json:"cache_read_tokens"`
			CacheCreationTokens int64  `json:"cache_creation_tokens"`
			ReasoningTokens     int64  `json:"reasoning_tokens"`
			TotalTokens         int64  `json:"total_tokens"`
		} `json:"by_endpoint"`
		ByModel []struct {
			Key                 string `json:"key"`
			CachedTokens        int64  `json:"cached_tokens"`
			CacheReadTokens     int64  `json:"cache_read_tokens"`
			CacheCreationTokens int64  `json:"cache_creation_tokens"`
			ReasoningTokens     int64  `json:"reasoning_tokens"`
			TotalTokens         int64  `json:"total_tokens"`
		} `json:"by_model"`
	}

	if err := json.Unmarshal(resp.Body, &data); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	// Verify token semantics metadata
	if data.TokenSemantics.TotalSource != "record.total_tokens" {
		t.Errorf("expected total_source 'record.total_tokens', got %q", data.TokenSemantics.TotalSource)
	}
	if !strings.Contains(data.TokenSemantics.CachedRelation, "subset_of_input") {
		t.Errorf("expected cached_relation 'subset_of_input', got %q", data.TokenSemantics.CachedRelation)
	}
	if data.TokenSemantics.ReasoningRelation != "provider_reported_in_total_when_applicable" {
		t.Errorf("expected reasoning_relation 'provider_reported_in_total_when_applicable', got %q", data.TokenSemantics.ReasoningRelation)
	}

	// Find p1 bucket
	var p1Summary *struct {
		Provider            string `json:"provider"`
		Model               string `json:"model"`
		InputTokens         int64  `json:"input_tokens"`
		OutputTokens        int64  `json:"output_tokens"`
		ReasoningTokens     int64  `json:"reasoning_tokens"`
		CachedTokens        int64  `json:"cached_tokens"`
		CacheReadTokens     int64  `json:"cache_read_tokens"`
		CacheCreationTokens int64  `json:"cache_creation_tokens"`
		TotalTokens         int64  `json:"total_tokens"`
	}
	var p2Summary *struct {
		Provider            string `json:"provider"`
		Model               string `json:"model"`
		InputTokens         int64  `json:"input_tokens"`
		OutputTokens        int64  `json:"output_tokens"`
		ReasoningTokens     int64  `json:"reasoning_tokens"`
		CachedTokens        int64  `json:"cached_tokens"`
		CacheReadTokens     int64  `json:"cache_read_tokens"`
		CacheCreationTokens int64  `json:"cache_creation_tokens"`
		TotalTokens         int64  `json:"total_tokens"`
	}
	for i := range data.Summary {
		if data.Summary[i].Provider == "p1" {
			p1Summary = &data.Summary[i]
		}
		if data.Summary[i].Provider == "p2" {
			p2Summary = &data.Summary[i]
		}
	}

	if p1Summary == nil {
		t.Fatal("p1 summary not found")
	}
	// Crucial check: TotalTokens is exactly 1500 (stored authoritative total), NOT 1500 + 400 or 1500 + 300
	if p1Summary.TotalTokens != 1500 {
		t.Errorf("p1 TotalTokens want 1500, got %d", p1Summary.TotalTokens)
	}
	if p1Summary.CachedTokens != 400 {
		t.Errorf("p1 CachedTokens want 400, got %d", p1Summary.CachedTokens)
	}
	if p1Summary.CacheReadTokens != 400 {
		t.Errorf("p1 CacheReadTokens want 400, got %d", p1Summary.CacheReadTokens)
	}
	if p1Summary.CacheCreationTokens != 50 {
		t.Errorf("p1 CacheCreationTokens want 50, got %d", p1Summary.CacheCreationTokens)
	}
	if p1Summary.ReasoningTokens != 300 {
		t.Errorf("p1 ReasoningTokens want 300, got %d", p1Summary.ReasoningTokens)
	}

	// Check zero/null record p2
	if p2Summary == nil {
		t.Fatal("p2 summary not found")
	}
	if p2Summary.TotalTokens != 0 || p2Summary.CachedTokens != 0 || p2Summary.CacheReadTokens != 0 || p2Summary.CacheCreationTokens != 0 || p2Summary.ReasoningTokens != 0 {
		t.Errorf("p2 zero fields mismatch: %+v", p2Summary)
	}

	// Verify breakdown lists have cache fields populated
	foundP1Provider := false
	for _, bp := range data.ByProvider {
		if bp.Key == "p1" {
			foundP1Provider = true
			if bp.TotalTokens != 1500 || bp.CachedTokens != 400 || bp.CacheReadTokens != 400 || bp.CacheCreationTokens != 50 || bp.ReasoningTokens != 300 {
				t.Errorf("by_provider p1 mismatch: %+v", bp)
			}
		}
	}
	if !foundP1Provider {
		t.Error("by_provider p1 not found")
	}

	foundP1Endpoint := false
	for _, be := range data.ByEndpoint {
		if be.Key == "exec-p1" {
			foundP1Endpoint = true
			if be.TotalTokens != 1500 || be.CachedTokens != 400 || be.CacheReadTokens != 400 || be.CacheCreationTokens != 50 || be.ReasoningTokens != 300 {
				t.Errorf("by_endpoint exec-p1 mismatch: %+v", be)
			}
		}
	}
	if !foundP1Endpoint {
		t.Error("by_endpoint exec-p1 not found")
	}

	foundP1Model := false
	for _, bm := range data.ByModel {
		if bm.Key == "m1" {
			foundP1Model = true
			if bm.TotalTokens != 1500 || bm.CachedTokens != 400 || bm.CacheReadTokens != 400 || bm.CacheCreationTokens != 50 || bm.ReasoningTokens != 300 {
				t.Errorf("by_model m1 mismatch: %+v", bm)
			}
		}
	}
	if !foundP1Model {
		t.Error("by_model m1 not found")
	}

	// Test filter preserves cache fields
	filterResp := callManagement(t, pluginapi.ManagementRequest{
		Method: http.MethodGet,
		Path:   "/v0/management/plugins/usage-audit/summary",
		Query:  map[string][]string{"provider": {"p1"}},
	})
	if filterResp.StatusCode != http.StatusOK {
		t.Fatalf("filter status=%d body=%s", filterResp.StatusCode, filterResp.Body)
	}
	var filterData struct {
		Summary []struct {
			Provider     string `json:"provider"`
			TotalTokens  int64  `json:"total_tokens"`
			CachedTokens int64  `json:"cached_tokens"`
		} `json:"summary"`
	}
	if err := json.Unmarshal(filterResp.Body, &filterData); err != nil {
		t.Fatalf("unmarshal filter: %v", err)
	}
	if len(filterData.Summary) != 1 || filterData.Summary[0].Provider != "p1" || filterData.Summary[0].CachedTokens != 400 || filterData.Summary[0].TotalTokens != 1500 {
		t.Errorf("filtered summary mismatch: %+v", filterData.Summary)
	}
}

// Resource responses carry []byte through the host RPC schema (base64 only in
// its JSON transport), so self-hosted WOFF2 is binary safe rather than rendered
// through the JSON response envelope. Immutable caching is safe because assets
// are compiled into this plugin version.
func TestSelfHostedSUITFontResourcesAreBinarySafeAndImmutable(t *testing.T) {
	t.Setenv("USAGE_AUDIT_DATA_DIR", t.TempDir())
	callRegister(t)
	fonts := map[string]struct {
		size int
		hash string
	}{
		"/v0/resource/plugins/usage-audit/assets/fonts/manrope-variable.woff2": {24836, "a30ddcd349703aff7464c34bef3fffdff405ee50c113440d7c8693c02d210972"},
		"/v0/resource/plugins/usage-audit/assets/fonts/suit-regular.woff2":     {167672, "7a1971b4e54d6d797f105f2d8ad2fc5f5d6fd532195f582d74f4403177e83185"},
		"/v0/resource/plugins/usage-audit/assets/fonts/suit-semibold.woff2":    {171164, "a0ea79e549e2ef42930d16c9c5b40600132a7f3661fe0e26a84a20c2c8d182a0"},
	}
	for path, want := range fonts {
		resp := callManagement(t, pluginapi.ManagementRequest{Method: http.MethodGet, Path: path})
		if resp.StatusCode != http.StatusOK {
			t.Errorf("%s: status=%d", path, resp.StatusCode)
			continue
		}
		if got := resp.Headers.Get("Content-Type"); got != "font/woff2" {
			t.Errorf("%s: Content-Type=%q", path, got)
		}
		if got := resp.Headers.Get("Cache-Control"); got != "public, max-age=31536000, immutable" {
			t.Errorf("%s: Cache-Control=%q", path, got)
		}
		if len(resp.Body) != want.size {
			t.Errorf("%s: size=%d want=%d", path, len(resp.Body), want.size)
		}
		gotHash := fmt.Sprintf("%x", sha256.Sum256(resp.Body))
		if gotHash != want.hash {
			t.Errorf("%s: sha256=%s want=%s", path, gotHash, want.hash)
		}
		if len(resp.Body) < 4 || string(resp.Body[:4]) != "wOF2" {
			t.Errorf("%s: not WOFF2 bytes", path)
		}
	}
}

func TestAPIKeySecretNeverPersisted(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("USAGE_AUDIT_DATA_DIR", dir)
	callRegister(t)
	t.Cleanup(plugin.Shutdown)
	rec := makeUsageRecord("provider", "model", "stable-index", 10, 5, true)
	rec["APIKey"] = "sk-raw-client-secret-never-persist"
	callUsageHandle(t, rec)
	plugin.Shutdown() // deterministic drain, no sleep
	path := filepath.Join(dir, "usage.jsonl")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(data, []byte("sk-raw-client-secret")) || bytes.Contains(data, []byte("api_key_id")) {
		t.Fatalf("secret stored: %s", data)
	}
	recs, err := store.ReadAll(path)
	if err != nil || len(recs) != 1 || recs[0].APIKeyID != "" || recs[0].AuthIndex != "stable-index" {
		t.Fatalf("persisted record: %+v, %v", recs, err)
	}
	callRegister(t)
	resp := callManagement(t, pluginapi.ManagementRequest{Method: "GET", Path: "/v0/management/plugins/usage-audit/records"})
	if resp.StatusCode != 200 || bytes.Contains(resp.Body, []byte("sk-raw-client-secret")) {
		t.Fatalf("records: %d %s", resp.StatusCode, resp.Body)
	}
}

func TestRotatedHistoryAPIsAndDeleteWriterResume(t *testing.T) {
	plugin.Shutdown()
	dir := t.TempDir()
	path := filepath.Join(dir, "usage.jsonl")
	st, err := store.Open(path, 1, 4)
	if err != nil {
		t.Fatal(err)
	}
	base := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	// Completion order differs from request order; legacy record omits RequestedAt.
	for i, hour := range []int{4, 1, 3, 2} {
		r := store.LogRecord{T: base.Add(time.Duration(hour) * time.Hour), Model: fmt.Sprintf("m%d", hour), Provider: "p", TotalTokens: int64(hour)}
		if i != 0 {
			r.RequestedAt = r.T
			r.T = base.Add(10 * time.Hour)
		}
		if err := st.Append(r); err != nil {
			t.Fatal(err)
		}
	}
	if err := st.Close(); err != nil {
		t.Fatal(err)
	}
	t.Setenv("USAGE_AUDIT_DATA_DIR", dir)
	callRegister(t)
	t.Cleanup(plugin.Shutdown)
	for _, prefix := range []string{"/v0/management/plugins/usage-audit", "/v0/management/plugins/usage-audit/data"} {
		resp := callManagement(t, pluginapi.ManagementRequest{Method: "GET", Path: prefix + "/records"})
		if resp.StatusCode != 200 {
			t.Fatalf("records: %d %s", resp.StatusCode, resp.Body)
		}
		var body struct {
			Records []store.LogRecord `json:"records"`
			Count   int               `json:"count"`
		}
		if err := json.Unmarshal(resp.Body, &body); err != nil {
			t.Fatal(err)
		}
		if body.Count != 4 || len(body.Records) != 4 {
			t.Fatalf("all rotations not queried: %s", resp.Body)
		}
		for i, r := range body.Records {
			if r.Model != fmt.Sprintf("m%d", 4-i) {
				t.Errorf("request order %d: %+v", i, r)
			}
		}
		if strings.Contains(prefix, "/resource/") && !body.Records[0].RequestedAt.Equal(base.Add(4*time.Hour)) {
			t.Fatalf("resource fallback missing: %s", resp.Body)
		}
		// Resource summary defaults to 30 days; explicitly select this legacy fixture.
		resp = callManagement(t, pluginapi.ManagementRequest{Method: "GET", Path: prefix + "/summary", Query: url.Values{"from": {"2025-01-01"}, "to": {"2025-01-01"}}})
		var summary struct {
			TotalCount int `json:"total_count"`
			Daily      []struct {
				TotalTokens int64 `json:"total_tokens"`
			} `json:"daily"`
		}
		if err := json.Unmarshal(resp.Body, &summary); err != nil {
			t.Fatal(err)
		}
		if resp.StatusCode != 200 || summary.TotalCount != 4 || len(summary.Daily) != 1 || summary.Daily[0].TotalTokens != 10 {
			t.Fatalf("retained summary: %d %s", resp.StatusCode, resp.Body)
		}
		// Legacy offset cursors are explicitly rejected, never silently reinterpreted.
		cursor := base64.URLEncoding.EncodeToString([]byte(strconv.Itoa(int(^uint(0) >> 1))))
		resp = callManagement(t, pluginapi.ManagementRequest{Method: "GET", Path: prefix + "/records", Query: url.Values{"cursor": {cursor}}})
		if resp.StatusCode != 400 || !bytes.Contains(resp.Body, []byte("legacy offset cursor")) {
			t.Fatalf("large legacy cursor: %d %s", resp.StatusCode, resp.Body)
		}
	}
	resp := callManagement(t, pluginapi.ManagementRequest{Method: "DELETE", Path: "/v0/management/plugins/usage-audit/records"})
	if resp.StatusCode != 200 {
		t.Fatalf("delete: %d %s", resp.StatusCode, resp.Body)
	}
	for i := 1; i <= 4; i++ {
		if _, err := os.Stat(fmt.Sprintf("%s.%d", path, i)); !os.IsNotExist(err) {
			t.Fatalf("rotation %d remains: %v", i, err)
		}
	}
	resp = callManagement(t, pluginapi.ManagementRequest{Method: "GET", Path: "/v0/management/plugins/usage-audit/summary"})
	if resp.StatusCode != 200 || !bytes.Contains(resp.Body, []byte(`"total_count":0`)) {
		t.Fatalf("summary after delete: %s", resp.Body)
	}
	callUsageHandle(t, makeUsageRecord("after", "delete", "idx", 1, 2, true))
	plugin.Shutdown()
	recs, err := store.ReadAll(path)
	if err != nil || len(recs) != 1 || recs[0].Provider != "after" {
		t.Fatalf("writer resume: %+v, %v", recs, err)
	}
}

func TestQuerySkipsCorruptedRetainedHistoryAndReportsDegradedHealth(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "usage.jsonl")
	secret := "sk-corrupt-secret-must-never-surface"
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	line := func(model string, hour int) string {
		b, err := json.Marshal(store.LogRecord{T: base.Add(time.Duration(hour) * time.Hour), RequestedAt: base.Add(time.Duration(hour) * time.Hour), Provider: "p", Model: model, TotalTokens: int64(hour)})
		if err != nil {
			t.Fatal(err)
		}
		return string(b)
	}
	// Oldest rotation and active file both contain valid records before and
	// after a malformed line. The active tail is truncated and also skipped.
	rotated := line("rot-before", 1) + "\n" + secret + "\n" + line("rot-after", 2) + "\n"
	active := line("active-before", 3) + "\n{" + secret + "\n" + line("active-after", 4) + "\n" + `{"model":"truncated-tail`
	if err := os.WriteFile(path+".1", []byte(rotated), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(active), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("USAGE_AUDIT_DATA_DIR", dir)
	callRegister(t)
	t.Cleanup(plugin.Shutdown)

	for _, prefix := range []string{"/v0/management/plugins/usage-audit", "/v0/management/plugins/usage-audit/data"} {
		recordsResp := callManagement(t, pluginapi.ManagementRequest{Method: "GET", Path: prefix + "/records", Query: url.Values{"limit": {"2"}}})
		if recordsResp.StatusCode != http.StatusOK || bytes.Contains(recordsResp.Body, []byte(secret)) || bytes.Contains(recordsResp.Body, []byte(path)) {
			t.Fatalf("%s records unsafe/failing: %d %s", prefix, recordsResp.StatusCode, recordsResp.Body)
		}
		var recordsBody struct {
			Records    []store.LogRecord `json:"records"`
			NextCursor string            `json:"next_cursor"`
			Health     store.Health      `json:"storage_health"`
		}
		if err := json.Unmarshal(recordsResp.Body, &recordsBody); err != nil {
			t.Fatal(err)
		}
		if len(recordsBody.Records) != 2 || recordsBody.Records[0].Model != "active-after" || recordsBody.Records[1].Model != "active-before" || recordsBody.NextCursor == "" {
			t.Fatalf("%s first page: %s", prefix, recordsResp.Body)
		}
		if !recordsBody.Health.Degraded || recordsBody.Health.DecodeFailures != 1 || recordsBody.Health.CorruptLines != 3 || recordsBody.Health.TruncatedLines != 1 || recordsBody.Health.LastError != "corrupt_jsonl" {
			t.Fatalf("%s health: %+v", prefix, recordsBody.Health)
		}

		secondResp := callManagement(t, pluginapi.ManagementRequest{Method: "GET", Path: prefix + "/records", Query: url.Values{"limit": {"2"}, "cursor": {recordsBody.NextCursor}}})
		var second struct {
			Records []store.LogRecord `json:"records"`
			Health  store.Health      `json:"storage_health"`
		}
		if err := json.Unmarshal(secondResp.Body, &second); err != nil {
			t.Fatal(err)
		}
		if secondResp.StatusCode != http.StatusOK || len(second.Records) != 2 || second.Records[0].Model != "rot-after" || second.Records[1].Model != "rot-before" {
			t.Fatalf("%s second page: %d %s", prefix, secondResp.StatusCode, secondResp.Body)
		}
		if second.Health.DecodeFailures != 1 || second.Health.CorruptLines != 3 {
			t.Fatalf("cache/cursor recounted corruption: %+v", second.Health)
		}

		summaryResp := callManagement(t, pluginapi.ManagementRequest{Method: "GET", Path: prefix + "/summary", Query: url.Values{"from": {"2026-01-01"}, "to": {"2026-01-01"}}})
		if summaryResp.StatusCode != http.StatusOK || bytes.Contains(summaryResp.Body, []byte(secret)) || bytes.Contains(summaryResp.Body, []byte(path)) {
			t.Fatalf("%s summary unsafe/failing: %d %s", prefix, summaryResp.StatusCode, summaryResp.Body)
		}
		var summary struct {
			TotalCount int          `json:"total_count"`
			Health     store.Health `json:"storage_health"`
		}
		if err := json.Unmarshal(summaryResp.Body, &summary); err != nil {
			t.Fatal(err)
		}
		if summary.TotalCount != 4 || !summary.Health.Degraded || summary.Health.CorruptLines != 3 {
			t.Fatalf("%s summary: %s", prefix, summaryResp.Body)
		}
	}
}

func TestStrictDecoderAdministrativeOption(t *testing.T) {
	dir := t.TempDir()
	secret := "strict-secret-must-not-surface"
	if err := os.WriteFile(filepath.Join(dir, "usage.jsonl"), []byte(secret+"\n{}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("USAGE_AUDIT_DATA_DIR", dir)
	t.Setenv("USAGE_AUDIT_STRICT_DECODING", "true")
	callRegister(t)
	t.Cleanup(plugin.Shutdown)
	resp := callManagement(t, pluginapi.ManagementRequest{Method: "GET", Path: "/v0/management/plugins/usage-audit/records"})
	if resp.StatusCode != http.StatusInternalServerError || bytes.Contains(resp.Body, []byte(secret)) || bytes.Contains(resp.Body, []byte(dir)) {
		t.Fatalf("strict response unsafe/unexpected: %d %s", resp.StatusCode, resp.Body)
	}
	var body struct {
		Health store.Health `json:"storage_health"`
	}
	if err := json.Unmarshal(resp.Body, &body); err != nil {
		t.Fatal(err)
	}
	if body.Health.DecodeMode != "strict" || body.Health.LastError != "corrupt_jsonl" || body.Health.CorruptLines != 1 {
		t.Fatalf("strict health: %+v", body.Health)
	}
}

func TestOperationalTwoCorruptedLinesPreservationAndCursorStability(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "usage.jsonl")

	// Approved corrupted lines that must be preserved verbatim in production logs.
	corruptedLine1 := `{"t":"2026-09-12T10:00:00Z", truncated json line`
	corruptedLine2 := `THIS IS NOT JSON AT ALL`
	multiObjectLine := `{"model":"multi-a"}{"model":"multi-b"}`
	truncatedTail := `{"model":"unterminated-corrupt-tail`
	secretSentinel := "sk-approved-corrupt-marker-must-not-leak"

	makeRec := func(model string, hour int, tokens int64) string {
		ts := time.Date(2026, 9, 12, hour, 0, 0, 0, time.UTC)
		rec := store.LogRecord{
			T:           ts,
			RequestedAt: ts,
			Provider:    "test-prov",
			Model:       model,
			TotalTokens: tokens,
			Generate:    true,
		}
		b, err := json.Marshal(rec)
		if err != nil {
			t.Fatal(err)
		}
		return string(b)
	}

	// Rotation log (.1): contains rot-1, corruptedLine1 (with secret sentinel), rot-2
	rotLineCorrupt := corruptedLine1 + " " + secretSentinel
	rot1Data := makeRec("rot-1", 1, 10) + "\n" +
		rotLineCorrupt + "\n" +
		makeRec("rot-2", 2, 20) + "\n"

	// Active log: contains active-1, corruptedLine2 (with secret sentinel), multiObjectLine, active-2, and unterminated truncatedTail
	activeLineCorrupt := corruptedLine2 + " " + secretSentinel
	activeData := makeRec("active-1", 3, 30) + "\n" +
		activeLineCorrupt + "\n" +
		multiObjectLine + "\n" +
		makeRec("active-2", 4, 40) + "\n" +
		truncatedTail

	if err := os.WriteFile(logPath+".1", []byte(rot1Data), 0o640); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(logPath, []byte(activeData), 0o640); err != nil {
		t.Fatal(err)
	}

	t.Setenv("USAGE_AUDIT_DATA_DIR", dir)
	t.Setenv("USAGE_AUDIT_STRICT_DECODING", "false") // default resilient
	callRegister(t)
	t.Cleanup(plugin.Shutdown)

	// 1. Summary endpoint: all valid records across rotated and active files must be aggregated,
	// skipping all intermediate malformed/non-object/multi-object lines and final truncated tail.
	summaryResp := callManagement(t, pluginapi.ManagementRequest{
		Method: "GET",
		Path:   "/v0/management/plugins/usage-audit/summary",
		Query:  url.Values{"from": {"2026-09-12"}, "to": {"2026-09-12"}},
	})
	if summaryResp.StatusCode != http.StatusOK {
		t.Fatalf("summary returned status %d: %s", summaryResp.StatusCode, summaryResp.Body)
	}
	if bytes.Contains(summaryResp.Body, []byte(secretSentinel)) ||
		bytes.Contains(summaryResp.Body, []byte(corruptedLine1)) ||
		bytes.Contains(summaryResp.Body, []byte(corruptedLine2)) ||
		bytes.Contains(summaryResp.Body, []byte(dir)) {
		t.Fatalf("summary leaked corrupt payload or file path: %s", summaryResp.Body)
	}

	var summaryBody struct {
		TotalCount int `json:"total_count"`
		Summary    []struct {
			Model       string `json:"model"`
			TotalTokens int64  `json:"total_tokens"`
		} `json:"summary"`
		Health store.Health `json:"storage_health"`
	}
	if err := json.Unmarshal(summaryResp.Body, &summaryBody); err != nil {
		t.Fatalf("unmarshaling summary response: %v", err)
	}
	// 4 valid records: rot-1 (10) + rot-2 (20) + active-1 (30) + active-2 (40) = 100
	if summaryBody.TotalCount != 4 {
		t.Errorf("summary TotalCount = %d, want 4", summaryBody.TotalCount)
	}
	var totalTokens int64
	for _, b := range summaryBody.Summary {
		totalTokens += b.TotalTokens
	}
	if totalTokens != 100 {
		t.Errorf("summary totalTokens = %d, want 100", totalTokens)
	}

	// Health counters:
	// - 4 corrupt lines: corruptedLine1, corruptedLine2, multiObjectLine, truncatedTail
	// - 1 truncated line: truncatedTail only
	// - DecodeFailures: 1
	// - Degraded: true
	// - DecodeMode: "resilient"
	// - LastError: "corrupt_jsonl"
	h := summaryBody.Health
	if h.DecodeMode != "resilient" || !h.Degraded || h.DecodeFailures != 1 || h.CorruptLines != 4 || h.TruncatedLines != 1 || h.LastError != "corrupt_jsonl" {
		t.Fatalf("summary storage_health mismatch: %+v", h)
	}

	// 2. Records pagination (keyset cursor stability test):
	// Page 1: limit=2. Expect active-2, active-1 (newest first).
	p1Resp := callManagement(t, pluginapi.ManagementRequest{
		Method: "GET",
		Path:   "/v0/management/plugins/usage-audit/records",
		Query:  url.Values{"limit": {"2"}},
	})
	if p1Resp.StatusCode != http.StatusOK {
		t.Fatalf("records p1 returned %d: %s", p1Resp.StatusCode, p1Resp.Body)
	}
	if bytes.Contains(p1Resp.Body, []byte(secretSentinel)) || bytes.Contains(p1Resp.Body, []byte(dir)) {
		t.Fatalf("records p1 leaked corrupt content or path: %s", p1Resp.Body)
	}

	var p1Body struct {
		Records    []store.LogRecord `json:"records"`
		Count      int               `json:"count"`
		HasMore    bool              `json:"has_more"`
		NextCursor string            `json:"next_cursor"`
		Health     store.Health      `json:"storage_health"`
	}
	if err := json.Unmarshal(p1Resp.Body, &p1Body); err != nil {
		t.Fatalf("unmarshaling p1: %v", err)
	}
	if len(p1Body.Records) != 2 || p1Body.Records[0].Model != "active-2" || p1Body.Records[1].Model != "active-1" {
		t.Fatalf("records p1 unexpected models: %+v", p1Body.Records)
	}
	if !p1Body.HasMore || p1Body.NextCursor == "" {
		t.Fatal("expected p1 to have more records and valid next_cursor")
	}
	// Cache reuse: corruption must not be recounted.
	if p1Body.Health.DecodeFailures != 1 || p1Body.Health.CorruptLines != 4 || p1Body.Health.CacheHits < 1 {
		t.Fatalf("p1 health recounted or cache miss: %+v", p1Body.Health)
	}

	// Page 2: limit=2 with cursor. Expect rot-2, rot-1.
	p2Resp := callManagement(t, pluginapi.ManagementRequest{
		Method: "GET",
		Path:   "/v0/management/plugins/usage-audit/records",
		Query:  url.Values{"limit": {"2"}, "cursor": {p1Body.NextCursor}},
	})
	if p2Resp.StatusCode != http.StatusOK {
		t.Fatalf("records p2 returned %d: %s", p2Resp.StatusCode, p2Resp.Body)
	}
	var p2Body struct {
		Records    []store.LogRecord `json:"records"`
		Count      int               `json:"count"`
		HasMore    bool              `json:"has_more"`
		NextCursor string            `json:"next_cursor"`
		Health     store.Health      `json:"storage_health"`
	}
	if err := json.Unmarshal(p2Resp.Body, &p2Body); err != nil {
		t.Fatalf("unmarshaling p2: %v", err)
	}
	if len(p2Body.Records) != 2 || p2Body.Records[0].Model != "rot-2" || p2Body.Records[1].Model != "rot-1" {
		t.Fatalf("records p2 unexpected models: %+v", p2Body.Records)
	}
	if p2Body.HasMore || p2Body.NextCursor != "" {
		t.Errorf("expected p2 to be terminal, got has_more=%v next_cursor=%q", p2Body.HasMore, p2Body.NextCursor)
	}
	if p2Body.Health.DecodeFailures != 1 || p2Body.Health.CorruptLines != 4 {
		t.Fatalf("p2 health recounted corruption: %+v", p2Body.Health)
	}

	// 3. Dashboard data endpoint (management-authenticated) returns the same page.
	resResp := callManagement(t, pluginapi.ManagementRequest{
		Method: "GET",
		Path:   "/v0/management/plugins/usage-audit/data/records",
		Query:  url.Values{"limit": {"2"}},
	})
	if resResp.StatusCode != http.StatusOK {
		t.Fatalf("resource records returned %d: %s", resResp.StatusCode, resResp.Body)
	}
	var resBody struct {
		Records []store.LogRecord `json:"records"`
		Health  store.Health      `json:"storage_health"`
	}
	if err := json.Unmarshal(resResp.Body, &resBody); err != nil {
		t.Fatalf("unmarshaling resource response: %v", err)
	}
	if len(resBody.Records) != 2 || resBody.Records[0].Model != "active-2" {
		t.Fatalf("resource records unexpected: %+v", resBody.Records)
	}
}
