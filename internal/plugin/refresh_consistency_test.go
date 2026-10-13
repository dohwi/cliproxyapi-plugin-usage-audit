package plugin_test

import (
	"context"
	"os/exec"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/dohwi/cliproxyapi-usage-audit/internal/plugin"
)

func TestDashboardRefreshConsistencyNode(t *testing.T) {
	nodePath, err := exec.LookPath("node")
	if err != nil {
		t.Skip("node not found; skipping refresh consistency tests")
	}
	var functions []string
	for _, name := range []string{"buildQueryString", "makeUsageSnapshot", "loadData", "loadMoreRecords"} {
		functions = append(functions, mustExtractJSFunc(t, plugin.DashboardHTML, name))
	}
	for _, name := range []string{"snapshotFilterKey", "usageTone", "diffUsageSnapshots"} {
		fn := regexp.MustCompile(`(?m)^function ` + regexp.QuoteMeta(name) + `\([^\n]+`).FindString(plugin.DashboardHTML)
		if fn == "" {
			t.Fatalf("could not extract JS function %q", name)
		}
		functions = append(functions, fn)
	}
	setup := `
const assert = require('node:assert/strict');
var state = {currentPreset:'today',fromRFC3339:'',toRFC3339:'',provider:'old',endpoint:'',model:'',status:'',cursor:'old-cursor',hasMore:true,records:[{id:'old'}],recordsExpanded:false,requestID:0,inFlight:false,abortCtrl:null,hasLoadedData:true,pendingChanges:null};
const initialSummary = {total_count:4,summary:[{total_tokens:100}],by_provider:[{provider:'old',requests:4,total_tokens:100}]};
state.serverSnapshot = makeUsageSnapshot(initialSummary, {records:state.records});
state.snapshotFilterKey = snapshotFilterKey();
const initialSnapshot = state.serverSnapshot;
const requests = [], renders = [], errors = [], polls = [];
var UA_POLL_RETRY_MS = 123;
function fetchWithTimeout(url, signal) {
  return new Promise((resolve, reject) => requests.push({url,signal,resolve,reject}));
}
function renderRecords(data, append) {
  renders.push({kind:'records',append,changes:state.pendingChanges});
  state.records = append ? state.records.concat(data.records) : data.records;
  state.recordsExpanded = !!append;
  state.hasMore = !!data.has_more;
  state.cursor = data.next_cursor || null;
}
function renderSummary(data) { renders.push({kind:'summary',data,changes:state.pendingChanges}); }
function showError(message) { errors.push(message); }
function clearError() { errors.length = 0; }
function setDashboardBusy() {}
function renderSkeletons() {}
function hideSkeletons() {}
function extractFilterOptions() {}
function computeAdaptivePollMs() { return 456; }
function scheduleNextPoll(ms) { polls.push(ms); }
async function flush() { await new Promise(resolve => setImmediate(resolve)); }
`
	tests := []struct {
		name string
		body string
	}{
		{"failed_refresh_retains_pagination", `
loadData(false);
assert.equal(state.cursor, 'old-cursor', 'refresh must not clear retained cursor before success');
assert.equal(state.hasMore, true);
assert(!requests[1].url.includes('cursor='), 'replacement query must not use retained cursor');
requests[0].resolve(initialSummary);
requests[1].reject(new Error('offline'));
await flush();
assert.equal(state.cursor, 'old-cursor');
assert.equal(state.hasMore, true);
assert.deepEqual(state.records, [{id:'old'}]);
assert.equal(state.serverSnapshot, initialSnapshot);
assert.equal(state.inFlight, false);
assert(errors[0].includes('offline'), 'failed request remains visible without a header timestamp');
assert.equal(renders.length, 0);
loadMoreRecords();
assert.equal(requests.length, 3);
assert(requests[2].url.includes('cursor=old-cursor'), 'same-filter pagination still works after failure');
requests[2].resolve({records:[{id:'older'}],has_more:false});
await flush();
assert.deepEqual(state.records, [{id:'old'},{id:'older'}]);
`},
		{"stale_responses_and_errors_are_ignored", `
loadData(false);
const oldSignal = requests[0].signal;
state.provider = 'new';
loadData(false);
assert.equal(oldSignal.aborted, true);
requests[2].resolve({total_count:9,summary:[{total_tokens:900}]});
requests[3].resolve({records:[{id:'new'}],has_more:true,next_cursor:'new-cursor'});
await flush();
const latestSnapshot = state.serverSnapshot;
const latestKey = state.snapshotFilterKey;
const renderCount = renders.length, pollCount = polls.length;
requests[0].resolve(initialSummary);
requests[1].resolve({records:[{id:'stale'}],has_more:false});
await flush();
assert.deepEqual(state.records, [{id:'new'}]);
assert.equal(state.cursor, 'new-cursor');
assert.equal(state.hasMore, true);
assert.equal(state.serverSnapshot, latestSnapshot);
assert.equal(state.snapshotFilterKey, latestKey);
assert.equal(renders.length, renderCount);
assert.equal(polls.length, pollCount);
loadData(false);
loadData(false);
requests[4].reject(new Error('stale failure'));
requests[5].resolve({records:[]});
await flush();
assert.equal(state.inFlight, true, 'stale catch must not clear the active request');
assert.equal(errors.length, 0);
assert.equal(polls.length, pollCount, 'stale catch must not schedule a retry');
requests[6].resolve(initialSummary);
requests[7].resolve({records:[{id:'latest'}],has_more:false});
await flush();
assert.deepEqual(state.records, [{id:'latest'}]);
`},
		{"append_preserves_summary_diff_baseline", `
loadMoreRecords();
assert.equal(requests.length, 1, 'append must not request a summary');
requests[0].resolve({records:[{id:'older'}],has_more:true,next_cursor:'next-cursor'});
await flush();
assert.equal(state.serverSnapshot, initialSnapshot, 'append must keep the full summary snapshot');
assert.deepEqual(state.serverSnapshot.values, initialSnapshot.values);
assert.deepEqual(renders[0].changes, {values:{},rows:{},records:{}});
loadData(false, true);
requests[1].resolve({total_count:5,summary:[{total_tokens:150}],by_provider:[{provider:'old',requests:5,total_tokens:150}]});
requests[2].resolve({records:[{id:'new'},{id:'old'}],has_more:false});
await flush();
const summaryRender = renders.find(r => r.kind === 'summary');
assert.equal(summaryRender.changes.values.requests, 'flash-good');
assert.equal(summaryRender.changes.values.tokens, 'flash-good');
assert.equal(summaryRender.changes.rows['provider:old:tokens'], 'flash-good');
assert.deepEqual(summaryRender.changes.records, {}, 'new records are not assumed to be changed');
assert.deepEqual(state.records, [{id:'new'},{id:'old'}]);
`},
		{"failed_filter_change_cannot_append_old_records", `
state.recordsExpanded = true;
state.provider = 'new';
state.recordsExpanded = false; // Filter events reset this flag before loading.
loadData(false);
requests[0].reject(new Error('filter fetch failed'));
requests[1].resolve({records:[{id:'ignored'}]});
await flush();
assert.deepEqual(state.records, [{id:'old'}]);
assert.equal(state.cursor, 'old-cursor');
assert.equal(state.hasMore, true);
assert.notEqual(state.snapshotFilterKey, snapshotFilterKey());
state.recordsExpanded = true; // Force replacement must also bypass expanded-page gating.
loadMoreRecords();
assert.equal(requests.length, 4, 'filter mismatch must fetch summary and replacement records');
assert(requests[2].url.includes('provider=new'));
assert(requests[3].url.includes('provider=new'));
assert(!requests[3].url.includes('cursor='), 'old filter cursor must never reach the new query');
requests[2].resolve({total_count:1,summary:[{total_tokens:10}]});
requests[3].resolve({records:[{id:'new'}],has_more:true,next_cursor:'new-cursor'});
await flush();
assert.deepEqual(state.records, [{id:'new'}], 'new filter replaces rather than mixes records');
assert.equal(state.snapshotFilterKey, snapshotFilterKey());
assert.equal(state.recordsExpanded, false);
assert.deepEqual(renders[0].changes, {values:{},rows:{},records:{}});
loadMoreRecords();
assert.equal(requests.length, 5);
assert(requests[4].url.includes('cursor=new-cursor'));
requests[4].resolve({records:[{id:'new-older'}],has_more:false});
await flush();
assert.deepEqual(state.records, [{id:'new'},{id:'new-older'}]);
`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			script := strings.Join(functions, "\n") + setup + "\n(async function() {\n" + tt.body + "\n})().catch(err => { console.error(err); process.exitCode = 1; });"
			if out, err := exec.CommandContext(ctx, nodePath, "-e", script).CombinedOutput(); err != nil {
				t.Fatalf("refresh consistency contract failed: %v\n%s", err, out)
			}
		})
	}
}
