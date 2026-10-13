package plugin_test

import (
	"github.com/dohwi/cliproxyapi-plugin-usage-audit/internal/plugin"
	"github.com/router-for-me/CLIProxyAPI/v7/sdk/pluginapi"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"testing"
)

func TestDashboardHTMLIntegrityAndElements(t *testing.T) {
	t.Setenv("USAGE_AUDIT_DATA_DIR", t.TempDir())
	callRegister(t)

	resp := callManagement(t, pluginapi.ManagementRequest{
		Method: http.MethodGet,
		Path:   "/dashboard",
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	body := string(resp.Body)

	// 1. Title & Header verification
	if !strings.Contains(body, "<title>토큰 사용량 대시보드</title>") {
		t.Errorf("Dashboard HTML missing title '토큰 사용량 대시보드'")
	}
	if strings.Count(body, "<h1>토큰 사용량</h1>") != 1 {
		t.Errorf("Dashboard HTML must render one visible page heading")
	}
	if strings.Contains(body, "Usage Audit 대시보드") {
		t.Errorf("Dashboard HTML still contains 'Usage Audit 대시보드'")
	}

	// 2. Required DOM IDs (including individual KPI IDs)
	requiredIDs := []string{
		"errorBox",
		"dashboardContent",
		"usageSkeleton",
		"usageDataContent",
		"usageMainContent",
		"presetGroup",
		"customDateWrap",
		"filterFromDate",
		"filterToDate",
		"filterProvider",
		"filterEndpoint",
		"filterModel",
		"filterStatus",
		"activeFilterBadge",
		"kpiTotalCount",
		"kpiInputTokens",
		"kpiOutputTokens",
		"kpiTotalTokens",
		"kpiFailureRate",
		"kpiActiveEntities",
		"chartCard",
		"chartArea",
		"breakdownSection",
		"providerAggTableMount",
		"endpointAggTableMount",
		"modelAggTableMount",
		"providerAggBody",
		"endpointAggBody",
		"modelAggBody",
		"recordsCountBadge",
		"recordsBody",
		"recordsTableMount",
		"loadMoreBar",
		"loadMoreBtn",
		"detailOverlay",
		"recordDialog",
		"dialogBody",
		"recordDialogClose",
	}

	for _, id := range requiredIDs {
		expectedAttr := `id="` + id + `"`
		if !strings.Contains(body, expectedAttr) {
			t.Errorf("Dashboard HTML missing element with id %q", id)
		}
	}

	// 3. Date presets
	presets := []string{
		`data-preset="today"`,
		`data-preset="yesterday"`,
		`data-preset="7d"`,
		`data-preset="30d"`,
		`data-preset="month"`,
		`data-preset="custom"`,
	}
	for _, p := range presets {
		if !strings.Contains(body, p) {
			t.Errorf("Dashboard HTML missing preset %q", p)
		}
	}

	// 4. Manual refresh controls and persisted interval choices are removed.
	for _, forbidden := range []string{`id="autoRefreshSelect"`, `id="refreshBtn"`, "cpaAutoRefreshSec", "manualRefresh"} {
		if strings.Contains(body, forbidden) {
			t.Errorf("Dashboard HTML retains removed refresh UI/logic %q", forbidden)
		}
	}

	// 5. Verification of JS logic keywords and contracts
	expectedLogic := []string{
		"AbortController",
		"visibilitychange",
		"document.hidden",
		"computeAdaptivePollMs",
		"scheduleNextPoll",
		"UA_POLL_RECENT_MS",
		"UA_POLL_HISTORY_MS",
		"formatProvider",
		"openai compatible-b.ai",
		"compatible-b.ai",
		"openai-compatible-b.ai",
		"b.ai",
		// endpoint filter: UI sends the single "endpoint" query (backend matches
		// ExecutorType OR Source). Sending source= and executor_type= together
		// would create an AND filter that hides records.
		"endpoint=' + encodeURIComponent(state.endpoint)",
		"tz_offset=' + encodeURIComponent(String(new Date().getTimezoneOffset()))",
		"Intl.DateTimeFormat().resolvedOptions().timeZone",
		"time_zone=' + encodeURIComponent(timeZone)",
		"limit=20",
		"cursor=",
		"next_cursor",
		"has_more",
		// summary aggregate items carry the bucket name in "key"
		"item.key || item.date",
		"item[nameKey] || item.key",
		// record field names match store.LogRecord JSON tags
		"latency_ms",
		"ttft_ms",
		"failure_status",
		"by_provider",
		"by_endpoint",
		"by_model",
		"호스트 observer에서 제공하지 않음",
		"showRecordDetail",
		"closeDrawer",
		"beforeunload",
		"최근 요청 기록",
		"최신 20건 · 필터와 더보기로 전체 로그 추적",
	}

	for _, token := range expectedLogic {
		if !strings.Contains(body, token) {
			t.Errorf("Dashboard script missing logic keyword or string %q", token)
		}
	}

	// Filter controls use a semantic form, distinct visual sections, and touch-sized controls.
	for _, token := range []string{
		`<form id="filterForm" class="filter-panel" novalidate>`,
		`<fieldset class="filter-section filter-period-section">`,
		`<legend class="sr-only">기간 선택</legend>`,
		`<fieldset class="filter-section filter-detail-section">`,
		`<legend class="sr-only">세부 필터</legend>`,
		`class="filter-row filter-period-row"`,
		`class="filter-row filter-fields-row"`,
		`grid-template-columns: minmax(0, 1fr) minmax(280px, 360px);`,
		`grid-template-columns: repeat(6, minmax(0, 1fr));`,
		`grid-template-columns: repeat(3, minmax(0, 1fr));`,
		`grid-template-columns: repeat(2, minmax(0, 1fr));`,
		`--filter-control-height: 38px;`,
		`height: var(--filter-control-height);`,
		`appearance: none;`,
		`-webkit-appearance: none;`,
		`class="compact-select-wrap"`,
		`class="select compact-select"`,
		`class="custom-dates"`,
		`class="filter-reset-field"`,
		`id="resetFiltersBtn" aria-label="모든 세부 필터 초기화"`,
		`class="filter-topmeta"`,
		`id="activeFilterText"`,
		`id="activeFilterBadge" class="active-filter-badge" aria-label="적용 조건"`,
		`background-color: var(--surface-input);`,
		`border: 1px solid var(--border-color);`,
	} {
		if !strings.Contains(body, token) {
			t.Errorf("dashboard missing polished filter contract %q", token)
		}
	}
	for _, forbidden := range []string{`id="applyFiltersBtn"`, `id="applyCustomDateBtn"`, `class="filter-actions"`} {
		if strings.Contains(body, forbidden) {
			t.Errorf("dashboard must not retain a separate filter action control %q", forbidden)
		}
	}

	// 6. Visual contract: six KPI modifier classes and pastel semantic styling.
	// Flat group-panel design: metric rows carry kpi-card/kpi-accent/kpi-* as
	// compat modifier classes alongside the flat .kpi-metric-row class.
	// No card-inside-card: individual rows must NOT be the sole root element
	// (they are siblings inside .kpi-group-metrics, separated by 1px lines).
	kpiClasses := []string{
		`kpi-card kpi-accent kpi-request`,
		`kpi-card kpi-accent kpi-total`,
		`kpi-card kpi-accent kpi-input`,
		`kpi-card kpi-accent kpi-output`,
		`kpi-card kpi-accent kpi-reason`,
		`kpi-card kpi-accent kpi-cache`,
	}
	for _, class := range kpiClasses {
		if !strings.Contains(body, class) {
			t.Errorf("Dashboard HTML missing KPI modifier DOM class %q", class)
		}
	}
	// No nested card-inside-card: .kpi-card must NOT carry its own background/border/radius
	// via standalone rule — only flat row context is permitted.
	if strings.Contains(body, ".kpi-card {\n    background") || strings.Contains(body, ".kpi-card {\r\n    background") {
		t.Errorf("Dashboard must not apply card background on individual .kpi-card rows (no nested card appearance)")
	}
	pastelTokens := []string{"--surface-toolbar: #f7f8fc", "--surface-hover: #f3f7fc", "--surface-tab-active: #eef4fb", "--surface-loading: #f7f8fc", "--surface-error: #fdf1f1"}
	for _, token := range pastelTokens {
		if !strings.Contains(body, token) {
			t.Errorf("Dashboard CSS missing pastel semantic value %q", token)
		}
	}
	for _, class := range []string{"preset-btn:hover", "preset-btn.active", ".input:hover", ".tab-btn:hover", ".tab-btn.active", "position: sticky", "tbody tr:hover", "loading-state", "error-box", ".usage-skeleton", ".skeleton-line", "@keyframes skeletonPulse"} {
		if !strings.Contains(body, class) {
			t.Errorf("Dashboard CSS/DOM missing required state or interaction class %q", class)
		}
	}
	for _, token := range []string{"renderSkeletons", "hideSkeletons", "state.hasLoadedData", "cachedAccounts.length === 0"} {
		// usage-audit mirrors quota-monitor's initial-only skeleton contract;
		// cachedAccounts is the quota dashboard's equivalent and must not leak here.
		if token == "cachedAccounts.length === 0" {
			if strings.Contains(body, token) {
				t.Errorf("Dashboard skeleton logic leaked quota-monitor state %q", token)
			}
			continue
		}
		if !strings.Contains(body, token) {
			t.Errorf("Dashboard skeleton loading contract missing %q", token)
		}
	}
	if strings.Contains(strings.ToLower(body), "gradient") {
		t.Errorf("Dashboard must not use gradients")
	}
	externalAssetRe := regexp.MustCompile(`(?i)(?:src|href)=["']https?://`)
	if externalAssetRe.MatchString(body) {
		t.Errorf("Dashboard must not use external assets")
	}
	if !strings.Contains(body, "최신 20건") || !strings.Contains(body, "grid-template-columns: repeat(12, 1fr)") || !strings.Contains(body, "grid-column: span 4") || !strings.Contains(body, "grid-column: span 8") {
		t.Errorf("Dashboard must preserve latest-20 and 4:8 layout semantics")
	}

	// The custom date range is the remaining date input pair only; it has no
	// standalone action slot after auto-apply was introduced.
	if strings.Contains(body, "grid-template-columns: minmax(0, 1fr) auto minmax(0, 1fr) auto;") {
		t.Error("custom date layout must not retain an apply-button column")
	}

	// 7. JavaScript static syntax validation via node if installed
	scriptRe := regexp.MustCompile(`(?s)<script>(.*?)</script>`)
	matches := scriptRe.FindStringSubmatch(body)
	if len(matches) < 2 {
		t.Fatalf("Could not extract <script> from dashboard HTML")
	}
	jsCode := matches[1]

	nodePath, err := exec.LookPath("node")
	if err == nil && nodePath != "" {
		tmpFile, err := os.CreateTemp("", "dash_script_*.js")
		if err != nil {
			t.Fatalf("failed to create temp JS file: %v", err)
		}
		defer os.Remove(tmpFile.Name())

		if _, err := tmpFile.WriteString(jsCode); err != nil {
			t.Fatalf("failed to write JS to temp file: %v", err)
		}
		_ = tmpFile.Close()

		cmd := exec.Command(nodePath, "--check", tmpFile.Name())
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("JavaScript static syntax check failed: %v\nOutput: %s", err, string(out))
		}
	}
}

// Records pagination cursor must be reset on every replace-mode load (initial,
// manual refresh, filter/preset change). Otherwise a stale cursor from the
// previous filtered view could be reused by "기록 더보기" after a failed refresh,
// resuming pagination at a wrong offset.
func TestAutoApplyFilterControls(t *testing.T) {
	body := plugin.DashboardHTML

	for _, forbidden := range []string{`id="applyFiltersBtn"`, `id="applyCustomDateBtn"`, "필터 적용", "filter-actions"} {
		if strings.Contains(body, forbidden) {
			t.Errorf("separate apply action must be absent: %q", forbidden)
		}
	}
	for _, required := range []string{
		"field.addEventListener('change', applyFilters);",
		"function applyCustomDatesWhenReady()", "field.addEventListener('change', applyCustomDatesWhenReady);",
		"if (!from.value || !to.value)",
	} {
		if !strings.Contains(body, required) {
			t.Errorf("auto-apply filter contract missing %q", required)
		}
	}
}

func TestRequestDetailModalSheetContract(t *testing.T) {
	body := plugin.DashboardHTML
	for _, required := range []string{
		`id="detailOverlay" class="detail-overlay" aria-hidden="true"`,
		`id="recordDialog" class="record-dialog"`, `role="dialog"`, `aria-modal="true"`,
		`id="recordDialogClose"`, `class="dialog-drag-handle" aria-hidden="true"`,
		`width: min(680px, calc(100vw - 32px));`, `max-height: 82dvh;`,
		`box-shadow: 0 16px 40px`, `translateY(10px) scale(0.98)`,
		`@media (max-width: 767px)`, `max-height: 86dvh;`, `border-radius: 18px 18px 0 0;`,
		`env(safe-area-inset-bottom)`, `grid-template-columns: repeat(2, minmax(0, 1fr));`,
		`grid-template-columns: 1fr;`, `overscroll-behavior: contain;`,
		`function lockBodyScroll()`, `function unlockBodyScroll()`, `document.body.style.overflow='hidden'`,
		`function trapDrawerFocus`, `state.drawerOpener`, `if (e.target === DOM.detailOverlay) closeDrawer();`,
		`if (e.key === 'Escape') { closeDrawer();`, `prefers-reduced-motion: reduce`,
	} {
		if !strings.Contains(body, required) {
			t.Errorf("request-detail modal contract missing %q", required)
		}
	}
	for _, forbidden := range []string{`class="drawer-overlay"`, `class="drawer"`, `onclick="closeDrawer()`, `filterApplyStatus`, `filter-apply-status`, `announceFilterApply`} {
		if strings.Contains(body, forbidden) {
			t.Errorf("obsolete drawer/filter status artifact remains: %q", forbidden)
		}
	}
}

func TestRecordsCursorRetainedUntilReplaceSuccess(t *testing.T) {
	body := plugin.DashboardHTML

	resetBlock := "if (!appendRecords) {\n    state.cursor = null;\n    state.hasMore = false;\n  }"
	if strings.Contains(body, resetBlock) {
		t.Error("replace-mode loads must retain pagination until successful response")
	}
	if !strings.Contains(body, "state.hasMore=!!data.has_more;state.cursor=data.next_cursor||null;") {
		t.Error("successful record response must replace pagination metadata")
	}

	// The cursor parameter must only be sent when appending.
	if !strings.Contains(body, "buildQueryString(true, appendRecords ? state.cursor : null)") {
		t.Errorf("records request must pass cursor only in append mode")
	}

	// loadMoreRecords must guard on hasMore and cursor.
	if !strings.Contains(body, "if (!state.hasMore || !state.cursor) return;") {
		t.Errorf("loadMoreRecords must guard on hasMore and cursor")
	}
}

func TestFormatProviderBehavior(t *testing.T) {
	nodePath, err := exec.LookPath("node")
	if err != nil || nodePath == "" {
		t.Skip("node not found; skipping formatProvider test")
	}

	jsScript := `
function formatProvider(val) {
  if (!val) return '-';
  var s = String(val).trim();
  var low = s.toLowerCase();
  if (low === 'openai compatible-b.ai' || low === 'compatible-b.ai' || low === 'openai-compatible-b.ai') {
    return 'b.ai';
  }
  return s;
}

const tests = [
  { in: 'openai compatible-b.ai', want: 'b.ai' },
  { in: 'compatible-b.ai', want: 'b.ai' },
  { in: 'openai-compatible-b.ai', want: 'b.ai' },
  { in: ' OpenAI Compatible-B.ai ', want: 'b.ai' },
  { in: 'antigravity', want: 'antigravity' },
  { in: 'anthropic', want: 'anthropic' },
  { in: '', want: '-' },
  { in: null, want: '-' },
];

for (const tc of tests) {
  const got = formatProvider(tc.in);
  if (got !== tc.want) {
    console.error("Failed for " + tc.in + ": got " + got + ", want " + tc.want);
    process.exit(1);
  }
}
console.log("ALL_PASSED");
`
	cmd := exec.Command(nodePath, "-e", jsScript)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("formatProvider test failed: %v, out: %s", err, string(out))
	}
	if !strings.Contains(string(out), "ALL_PASSED") {
		t.Fatalf("formatProvider output unexpected: %s", string(out))
	}
}

// Root top spacing is driven by a single CSS custom property
// --dashboard-top-offset (24px desktop / 20px tablet / 16px mobile). The body
// padding-top must reference the var exactly once; the old hardcoded 32px/18px
// tops and the duplicate header margin must be gone, while horizontal/bottom
// padding is preserved.
func TestRootTopSpacingRequirements(t *testing.T) {
	body := plugin.DashboardHTML

	// Exact var declarations.
	if !strings.Contains(body, "--dashboard-top-offset: 24px;") {
		t.Errorf("dashboard :root must declare --dashboard-top-offset: 24px;")
	}
	if !strings.Contains(body, "--dashboard-top-offset: 20px;") {
		t.Errorf("dashboard must set --dashboard-top-offset: 20px at the <=1199px breakpoint")
	}
	if !strings.Contains(body, "--dashboard-top-offset: 16px;") {
		t.Errorf("dashboard must set --dashboard-top-offset: 16px at the <=767px breakpoint")
	}

	// 20px must live inside a max-width:1199px media query, 16px inside max-width:767px.
	mediaRe := regexp.MustCompile(`(?s)@media[^{]*max-width: 1199px[^{]*\{(.*?)\n  \}`)
	if m := mediaRe.FindStringSubmatch(body); len(m) < 2 || !strings.Contains(m[1], "--dashboard-top-offset: 20px;") {
		t.Errorf("20px var must be declared within the <=1199px media query")
	}
	media767Re := regexp.MustCompile(`(?s)@media[^{]*max-width: 767px[^{]*\{(.*?)\n  \}`)
	if m := media767Re.FindStringSubmatch(body); len(m) < 2 || !strings.Contains(m[1], "--dashboard-top-offset: 16px;") {
		t.Errorf("16px var must be declared within the <=767px media query")
	}

	// The var must be consumed exactly once, as the body padding-top.
	if uses := strings.Count(body, "var(--dashboard-top-offset)"); uses != 1 {
		t.Errorf("--dashboard-top-offset must be used exactly once, got %d", uses)
	}
	if !strings.Contains(body, "padding: var(--dashboard-top-offset) 14px 16px 14px;") {
		t.Errorf("body padding-top must use the var and preserve horizontal/bottom (14px 16px 14px)")
	}

	// Old hardcoded top values must be absent.
	if strings.Contains(body, "padding: 32px") {
		t.Errorf("old hardcoded desktop top padding (32px) must be removed")
	}
	if strings.Contains(body, "padding: 18px") {
		t.Errorf("old hardcoded mobile top padding (18px) must be removed")
	}

	// Mobile horizontal/bottom preserved via longhand without reusing the var for top.
	if !strings.Contains(body, "padding-right: 10px; padding-bottom: 12px; padding-left: 10px;") {
		t.Errorf("mobile body must use synchronized 10px/12px/10px padding")
	}

	// Duplicate header margin removed.
	if strings.Contains(body, "margin-bottom: 12px") {
		t.Errorf("duplicate header margin (margin-bottom: 12px) must be removed")
	}
}

// KPI numbers must use a responsive clamp and must not be truncated/ellipsized.
func TestKPIValueResponsiveClampNoTruncation(t *testing.T) {
	body := plugin.DashboardHTML

	kpiRe := regexp.MustCompile(`(?s)\.kpi-value\s*\{([^}]*)\}`)
	m := kpiRe.FindStringSubmatch(body)
	if len(m) < 2 {
		t.Fatalf("could not locate .kpi-value CSS rule")
	}
	rule := m[1]

	if !strings.Contains(rule, "clamp(") {
		t.Errorf(".kpi-value must use a responsive clamp() font-size, got: %s", rule)
	}
	if strings.Contains(rule, "text-overflow") || strings.Contains(rule, "overflow: hidden") {
		t.Errorf(".kpi-value must not truncate (no text-overflow/overflow:hidden), got: %s", rule)
	}
}

// Compact K/M/B formatting is forbidden across the whole dashboard; the only
// number formatter allowed is the exact comma-grouped integer formatter.
func TestNoCompactKMBSuffixFormatter(t *testing.T) {
	body := plugin.DashboardHTML

	if strings.Contains(body, "formatCompactNumber") {
		t.Errorf("dashboard must not reference formatCompactNumber (compact K/M/B is forbidden)")
	}
	if !strings.Contains(body, "function formatExactNumber") {
		t.Errorf("dashboard must define formatExactNumber")
	}
}

// formatExactNumber must return an em-dash for null/NaN, plain '0' for zero,
// signed comma-grouped integers for negatives, and never compact K/M/B units.
func TestFormatExactNumberBehavior(t *testing.T) {
	nodePath, err := exec.LookPath("node")
	if err != nil || nodePath == "" {
		t.Skip("node not found; skipping formatExactNumber test")
	}

	fnRe := regexp.MustCompile(`(?s)function formatExactNumber\(num\)\s*\{.*?\n\}`)
	match := fnRe.FindString(plugin.DashboardHTML)
	if match == "" {
		t.Fatalf("could not extract formatExactNumber implementation from dashboard")
	}

	jsScript := match + "\n" + `
const cases = [
  { in: null, want: '—' },
  { in: undefined, want: '—' },
  { in: 'abc', want: '—' },
  { in: 0, want: '0' },
  { in: 42, want: '42' },
  { in: 999, want: '999' },
  { in: 1000, want: '1,000' },
  { in: 1234567, want: '1,234,567' },
  { in: 9876543210, want: '9,876,543,210' },
  { in: -1234, want: '-1,234' },
  { in: -1000000, want: '-1,000,000' }
];
for (const tc of cases) {
  const got = formatExactNumber(tc.in);
  if (got !== tc.want) {
    console.error('formatExactNumber(' + String(tc.in) + ') = ' + got + ', want ' + tc.want);
    process.exit(1);
  }
  if (/[KMB]/.test(got)) {
    console.error('formatExactNumber produced a compact unit suffix: ' + got);
    process.exit(1);
  }
}
console.log('EXACT_FORMAT_TESTS_PASSED');
`

	cmd := exec.Command(nodePath, "-e", jsScript)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("formatExactNumber test failed: %v\nOutput: %s", err, string(out))
	}
	if !strings.Contains(string(out), "EXACT_FORMAT_TESTS_PASSED") {
		t.Fatalf("unexpected formatExactNumber output: %s", string(out))
	}
}

// mustExtractJSFunc pulls a top-level `function NAME(...) { ... }` block verbatim
// out of the dashboard HTML script. Each resolver function closes with a `}` at
// column 0, so the non-greedy `\n\}` stop matches only the real function end.
func mustExtractJSFunc(t *testing.T, body, name string) string {
	t.Helper()
	re := regexp.MustCompile(`(?s)function ` + regexp.QuoteMeta(name) + `\(.*?\n\}`)
	m := re.FindString(body)
	if m == "" {
		t.Fatalf("could not extract JS function %q from dashboard HTML", name)
	}
	return m
}

// Summary groups deliberately expose provider-reported fields without presenting
// cache/reasoning values as components of an additive total.
func TestSummaryGroupsAndNonAdditiveContract(t *testing.T) {
	body := plugin.DashboardHTML

	for _, token := range []string{
		`class="kpi-groups"`,
		`class="kpi-group kpi-group-overview"`,
		`class="kpi-group kpi-group-input"`,
		`class="kpi-group kpi-group-output"`,
		`id="kpiTotalCount"`, `id="kpiTotalTokens"`,
		`id="kpiInputTokens"`, `id="kpiFreshInputTokens"`, `id="kpiCachedTokens"`,
		`id="kpiOutputTokens"`, `id="kpiReasoningTokens"`,
		"Provider 보고값으로 집계 방식에 따라 캐시 포함 여부가 다릅니다.", "캐시 제외", "캐시 입력은 입력에 포함되거나 별도로 집계됩니다.", "추론은 출력 또는 전체 토큰에 포함될 수 있어 중복 합산하지 않습니다.",
	} {
		if !strings.Contains(body, token) {
			t.Errorf("dashboard missing summary group contract %q", token)
		}
	}

	// Total comes solely from stored total_tokens, never from a presentation
	// equation involving output/reasoning or fresh/cached fields.
	for _, forbidden := range []string{
		"totalTokens += inp + outp",
		"totalTokens += inputTokens + outputTokens",
		"totalTokens += freshInputTokens + cachedTokens",
		"outputTokens + reasoningTokens",
		"freshInputTokens + cachedTokens",
	} {
		if strings.Contains(body, forbidden) {
			t.Errorf("dashboard has misleading additive token equation %q", forbidden)
		}
	}
	if !strings.Contains(body, "inputTokens += inp;") || !strings.Contains(body, "var inp = item.input_tokens || 0;") {
		t.Error("renderSummary must aggregate reported summary.input_tokens for kpiInputTokens")
	}
}

func TestSummaryGroupResponsiveLayout(t *testing.T) {
	body := plugin.DashboardHTML
	for _, required := range []string{".kpi-groups {\n    display: grid;\n    grid-template-columns: repeat(3, minmax(0, 1fr));", ".kpi-groups, .skeleton-kpi-groups { grid-template-columns: 1fr;", ".kpi-group-overview { grid-column: auto; }", ".kpi-group-metrics { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr));", ".kpi-group-input .kpi-group-metrics { grid-template-columns: repeat(3, minmax(0, 1fr));"} {
		if !strings.Contains(body, required) {
			t.Errorf("missing current layout contract %q", required)
		}
	}
	for _, required := range []string{"@media (max-width: 1099px) and (min-width: 768px)", ".kpi-groups { grid-template-columns: repeat(2, minmax(0, 1fr)); }", ".kpi-group-overview { grid-column: 1 / -1; }"} {
		if !strings.Contains(body, required) {
			t.Errorf("tablet must give exact values enough width: %q", required)
		}
	}
}

func TestDashboardUsesUnauthenticatedResources(t *testing.T) {
	body := plugin.DashboardHTML
	for _, bad := range []string{
		"/v0/management/", "Authorization", "Bearer ", "management_key",
		"resolveManagementKey", "extractKeyFromParsed", "decodeAndParseStorageValue",
	} {
		if strings.Contains(body, bad) {
			t.Errorf("dashboard contains forbidden auth/management reference %q", bad)
		}
	}
	for _, good := range []string{
		"/v0/resource/plugins/usage-audit/data/summary",
		"/v0/resource/plugins/usage-audit/data/records",
		"endpoint=' + encodeURIComponent(state.endpoint)",
		"cursor=' + encodeURIComponent(cursor)",
		"limit=20",
	} {
		if !strings.Contains(body, good) {
			t.Errorf("dashboard missing resource/query contract %q", good)
		}
	}
}

// TestFlatGroupPanels: 개별 metric row에 카드 배경/그림자/radius를 부여하는
// 독립 CSS 규칙이 없어야 한다 (card-inside-card 제거).
// .kpi-metric-row 자체나 .kpi-card { background/box-shadow/border-radius } 단독 선언 금지.
func TestFlatGroupPanels(t *testing.T) {
	body := plugin.DashboardHTML

	// kpi-group-metrics 내부 metric rows는 .kpi-metric-row로 렌더링되어야 함
	if !strings.Contains(body, "kpi-metric-row") {
		t.Errorf("dashboard must use .kpi-metric-row for flat metric rows inside group panel")
	}
	// flat: 개별 row에 독립 카드 배경 부여 금지
	if strings.Contains(body, ".kpi-card {\n    background") ||
		strings.Contains(body, ".kpi-card{\n    background") {
		t.Errorf("individual .kpi-card rule must not set its own background (card-inside-card forbidden)")
	}
	// group body neutral: kpi-group-metrics 자체에 배경색 없어야 함
	// (컬러 배경은 .kpi-group-head 에만 허용)
	if strings.Contains(body, ".kpi-group-metrics {") {
		re := regexp.MustCompile(`(?s)\.kpi-group-metrics \{([^}]*)`)
		m := re.FindStringSubmatch(body)
		if len(m) >= 2 && strings.Contains(m[1], "background") {
			t.Errorf(".kpi-group-metrics must not have its own background (neutral body)")
		}
	}
	// 1px divider between rows: + 연결자로 border-top 구현
	if !strings.Contains(body, ".kpi-metric-row + .kpi-metric-row") {
		t.Errorf("metric rows must be separated by 1px line via .kpi-metric-row + .kpi-metric-row sibling selector")
	}
	// value right-aligned, tabular nums
	if !strings.Contains(body, "text-align: right;") {
		t.Errorf("dashboard must right-align values")
	}
	if !strings.Contains(body, "font-variant-numeric: tabular-nums") {
		t.Errorf("dashboard must use tabular-nums for numeric values")
	}
}

// TestVisualTokens: 공유 개념 토큰(page bg, surface, border, text, radius16, shadow) 검증.
// gradient/pattern 금지는 TestDashboardHTMLIntegrityAndElements에서 이미 검사.
func TestVisualTokens(t *testing.T) {
	body := plugin.DashboardHTML

	tokens := []string{
		"--bg-page: #f7f8fa",
		"--bg-card: #ffffff",
		"--border-color: #e1e6ed",
		"--text-primary: #182230",
		"--text-muted: #667085",
		"--radius-card: 12px",
		"--filter-control-height: 38px",
	}
	for _, tok := range tokens {
		if !strings.Contains(body, tok) {
			t.Errorf("dashboard missing shared visual token %q", tok)
		}
	}
	// Input group: blue-lavender subtle header
	if !strings.Contains(body, "kpi-group-input") || !strings.Contains(body, "kpi-group-output") {
		t.Errorf("dashboard must have kpi-group-input and kpi-group-output sections")
	}
}

// TestUpdateStatusRemoved: the page heading has no timestamp DOM or age ticker.
// Adaptive polling operates independently of the removed header-time helpers.
func TestUpdateStatusRemoved(t *testing.T) {
	body := plugin.DashboardHTML
	for _, gone := range []string{
		`class="update-status"`, `id="updateStatus"`, "lastUpdateTime", "lastUpdateRel",
		"update-status-label", "update-status-abs", "update-status-rel", "update-status-icon", "마지막 업데이트",
		"relativeTimeKo", "setFreshnessState", "formatDashboardTimestamp", "updateLastRefreshedTime",
		"startRelativeTicker", "tickRelative", "relativeTimer", "lastRefreshAt",
	} {
		if strings.Contains(body, gone) {
			t.Errorf("obsolete header-time DOM/logic must be removed; still present: %q", gone)
		}
	}
	for _, kept := range []string{
		"function computeAdaptivePollMs", "function scheduleNextPoll",
		"UA_POLL_RECENT_MS  = 30000", "UA_POLL_ACTIVE_MS  = 90000",
		"UA_POLL_QUIET_MS   = 300000", "UA_POLL_HISTORY_MS = 600000",
		"UA_POLL_RETRY_MS   = 120000", "UA_POLL_MAX_MS     = 600000",
		"if (delayMs <= 0) return;", "if (delayMs > UA_POLL_MAX_MS) delayMs = UA_POLL_MAX_MS;",
		"if (document.hidden) return;", "if (!state.inFlight) loadData(false);",
	} {
		if !strings.Contains(body, kept) {
			t.Errorf("adaptive polling contract %q must remain", kept)
		}
	}
}

// TestTopOffset24_20_16: dashboard-top-offset 정확히 128/96/16px.
func TestTopOffset24_20_16(t *testing.T) {
	body := plugin.DashboardHTML

	if !strings.Contains(body, "--dashboard-top-offset: 24px;") {
		t.Errorf("desktop top offset must be 24px")
	}
	mediaTablet := regexp.MustCompile(`(?s)@media[^{]*max-width: 1199px[^{]*\{(.*?)\n  \}`)
	if m := mediaTablet.FindStringSubmatch(body); len(m) < 2 || !strings.Contains(m[1], "--dashboard-top-offset: 20px;") {
		t.Errorf("tablet top offset must be 20px inside <=1199px media query")
	}
	mediaMobile := regexp.MustCompile(`(?s)@media[^{]*max-width: 767px[^{]*\{(.*?)\n  \}`)
	if m := mediaMobile.FindStringSubmatch(body); len(m) < 2 || !strings.Contains(m[1], "--dashboard-top-offset: 16px;") {
		t.Errorf("mobile top offset must be 16px inside <=767px media query")
	}
}

// TestResponsiveNoOverflow: overflow 방지 및 table internal-only 스크롤 검증.
func TestResponsiveNoOverflow(t *testing.T) {
	body := plugin.DashboardHTML
	for _, required := range []string{"min-width: 0;", "overflow-wrap: anywhere;", ".container { width: 100%; }", "#recordsPanel table { table-layout: fixed; min-width: 1100px; }", "#recordsPanel table { min-width: 680px; }", "#recordsPanel table { width: 100%; table-layout: fixed; }", ".num-cell {\n    white-space: nowrap;\n    overflow-wrap: normal;"} {
		if !strings.Contains(body, required) {
			t.Errorf("missing current layout contract %q", required)
		}
	}
}

func TestDOMCacheOptimization(t *testing.T) {
	body := plugin.DashboardHTML

	// DOM 객체가 선언되어야 함
	if !strings.Contains(body, "var DOM = {};") {
		t.Errorf("dashboard must declare DOM cache object: var DOM = {};")
	}
	// initDashboard에서 주요 요소를 캐시해야 함
	cacheKeys := []string{
		"DOM.kpiTotalCount",
		"DOM.kpiTotalTokens",
		"DOM.recordsBody",
	}
	for _, k := range cacheKeys {
		if !strings.Contains(body, k) {
			t.Errorf("DOM cache must include %s", k)
		}
	}
	// aggregate-once 패턴: summaryList.forEach 한 번에 모든 집계
	if !strings.Contains(body, "summaryList.forEach") {
		t.Errorf("renderSummary must aggregate once via summaryList.forEach")
	}
	// 프레임워크 금지: 실제 export된 dohwi-ui React 번들(인라인)만 허용하고 CDN/외부 JS는 금지한다.
	for _, fw := range []string{"Vue", "Angular", "import {", "require("} {
		if strings.Contains(body, fw) {
			t.Errorf("dashboard must not introduce framework: found %q", fw)
		}
	}
	for _, required := range []string{`<script data-dohwi-ui="exported">`, `Symbol.for("react.transitional.element")`} {
		if !strings.Contains(body, required) {
			t.Errorf("actual exported React dohwi-ui bundle must be inlined: missing %q", required)
		}
	}
	for _, external := range []*regexp.Regexp{
		regexp.MustCompile(`(?i)<script\b[^>]*\bsrc\s*=`),
		regexp.MustCompile(`(?i)<script\b[^>]*\btype\s*=\s*["']?module`),
		regexp.MustCompile(`(?i)<link\b[^>]*\bhref\s*=\s*["']?(?:https?:)?//`),
		regexp.MustCompile(`(?i)@import\s+(?:url\()?["']?(?:https?:)?//`),
		regexp.MustCompile(`(?i)(?:unpkg\.com|cdn\.jsdelivr|cdnjs\.|esm\.sh|skypack\.dev|googleapis\.com)`),
	} {
		if m := external.FindString(body); m != "" {
			t.Errorf("dashboard must not load CDN/external JS or assets: %q", m)
		}
	}
}

// V5.3 removes the former input/output composition UI entirely. Its presence
// would suggest an additive relationship that does not hold for cached and
// reasoning fields.
func TestV53NoTokenCompositionArtifacts(t *testing.T) {
	body := plugin.DashboardHTML
	for _, forbidden := range []string{
		"토큰 구성 비율", "composition-bar-wrap", "comp-seg-in", "comp-seg-out",
		"compSegIn", "compSegOut", "compRatioText", "compSemantics", "renderTokenComposition",
	} {
		if strings.Contains(body, forbidden) {
			t.Errorf("removed token composition artifact remains: %q", forbidden)
		}
	}
}

// The output section must have exactly its visible output/reasoning rows in
// its own metrics container. This protects against the V5.2 extra closing div
// that detached the rows at some layout widths (including zero-valued data).
func TestV53OutputGroupStructureAndResponsiveVisibility(t *testing.T) {
	body := plugin.DashboardHTML
	re := regexp.MustCompile(`(?s)<section class="kpi-group kpi-group-output".*?</section>`)
	group := re.FindString(body)
	if group == "" {
		t.Fatal("output KPI group not found")
	}
	for _, required := range []string{
		`class="kpi-group-metrics"`, `id="kpiOutputTokens"`, `id="kpiReasoningTokens"`,
		`class="kpi-value kpi-value-output"`, `class="kpi-value kpi-value-reason"`,
	} {
		if !strings.Contains(group, required) {
			t.Errorf("output group missing visible metric %q", required)
		}
	}
	if strings.Count(group, `class="kpi-group-metrics"`) != 1 {
		t.Errorf("output group needs one metrics container, got %d", strings.Count(group, `class="kpi-group-metrics"`))
	}
	// Verify that .kpi-group-head does not have a prematurely closed div that detaches .kpi-group-metrics from the section
	headRe := regexp.MustCompile(`(?s)<div class="kpi-group-head">.*?</div>\s*<div class="kpi-group-metrics">`)
	if !headRe.MatchString(group) {
		t.Errorf("output group head should be immediately followed by kpi-group-metrics without premature closing tags")
	}
	if strings.Contains(group, "display:none") || strings.Contains(group, `class="kpi-compat-hidden"`) {
		t.Error("output/reasoning values must remain visible at zero and all breakpoints")
	}
	for _, breakpoint := range []string{"@media (max-width: 1199px) and (min-width: 768px)", "@media (max-width: 767px)"} {
		if !strings.Contains(body, breakpoint) {
			t.Errorf("missing responsive breakpoint %q", breakpoint)
		}
	}
}

func TestV53DailyChartTooltipAccessibilityAndFields(t *testing.T) {
	body := plugin.DashboardHTML
	for _, required := range []string{
		`id="dailyChartTooltip"`, `role="tooltip"`, "daily-chart-tooltip",
		"showDailyChartTooltip", "closeDailyChartTooltip", "pointerenter", "pointerdown",
		"aria-describedby", "setAttribute('tabindex','0')", "daily-chart-bar", "e.key === 'Escape'",
		"요청: ", "전체 토큰 (공식): ", "입력 (보고값): ", "출력: ", "실패: ",
		"캐시 (보고값, 비가산): ", "추론 (보고값, 비가산): ",
		"Math.min(box.left, window.innerWidth - tooltip.offsetWidth - margin)",
		"prefers-reduced-motion: reduce",
	} {
		if !strings.Contains(body, required) {
			t.Errorf("daily tooltip contract missing %q", required)
		}
	}
	if !strings.Contains(body, "formatExactNumber(detail.requests)") || !strings.Contains(body, "formatExactNumber(detail.total)") {
		t.Error("tooltip numerical fields must use exact comma formatter")
	}
	for _, forbidden := range []string{"chart.js", "tippy", "bootstrap.Tooltip"} {
		if strings.Contains(strings.ToLower(body), strings.ToLower(forbidden)) {
			t.Errorf("tooltip must not use a library: found %q", forbidden)
		}
	}
}

func TestV53UnifiedControlPanelAndSelfHostedFontCSS(t *testing.T) {
	body := plugin.DashboardHTML
	for _, required := range []string{
		`id="activeFilterBadge"`,
		`font-family: "Manrope"`, `font-family: "SUIT"`, "manrope-variable.woff2", "font-weight: 400;", "font-weight: 600;", "font-display: swap;",
		`/v0/resource/plugins/usage-audit/assets/fonts/suit-regular.woff2`,
		`/v0/resource/plugins/usage-audit/assets/fonts/suit-semibold.woff2`,
		`font-family: "Manrope", "SUIT"`,
	} {
		if !strings.Contains(body, required) {
			t.Errorf("V5.3 control/font contract missing %q", required)
		}
	}
	if regexp.MustCompile(`(?i)(?:src|href)=["']https?://`).MatchString(body) {
		t.Error("dashboard must not load font or other external URL")
	}

	if strings.Contains(body, `class="dashboard-control-panel"`) {
		t.Error("usage dashboard must not introduce an extra control card")
	}

	// Removed controls stay absent.
	for _, forbidden := range []string{`id="autoRefreshSelect"`, `id="refreshBtn"`, "cpaAutoRefreshSec", "manualRefresh"} {
		if strings.Contains(body, forbidden) {
			t.Errorf("adaptive control panel retains %q", forbidden)
		}
	}
}

// Untrusted provider/endpoint/model and failure text must never be concatenated
// into HTML attributes. In particular, a quote breakout such as
// `x" onmouseover="...` must remain ordinary text/title content.
func TestDashboardDynamicDataUsesSafeDOMTextAndEscapesQuotes(t *testing.T) {
	body := plugin.DashboardHTML
	for _, forbidden := range []string{
		"tr.innerHTML", "tbody.innerHTML", "body.innerHTML", "tooltip.innerHTML",
	} {
		if strings.Contains(body, forbidden) {
			t.Errorf("dynamic dashboard renderer must not use %q", forbidden)
		}
	}
	for _, required := range []string{
		"TableCell", "textContent", "cell.title", "cellContent(cell)", "title={cell.title}",
		"function detailRow", "valueEl.textContent=value", "tooltip.replaceChildren()",
		"return String(str).replace(/[&<>\"']/g", "'&quot;'", "'&#39;'",
	} {
		if !strings.Contains(body, required) {
			t.Errorf("missing XSS-safe rendering/escaping contract %q", required)
		}
	}

	nodePath, err := exec.LookPath("node")
	if err != nil {
		t.Skip("node not found; skipping esc behavior test")
	}
	fn := mustExtractJSFunc(t, body, "esc")
	cmd := exec.Command(nodePath, "-e", fn+`
const got = esc('x" onmouseover="alert(1) & <tag> \'q\'');
if (got !== 'x&quot; onmouseover=&quot;alert(1) &amp; &lt;tag&gt; &#39;q&#39;') process.exit(1);`)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("esc quote-breakout test failed: %v: %s", err, out)
	}
}

func TestMobileFilterResponsiveAndNoInlineStyles(t *testing.T) {
	body := plugin.DashboardHTML
	if regexp.MustCompile(`(?:^|[^-])style="`).MatchString(body) {
		t.Errorf("dashboard filter page must not contain inline style attributes")
	}
	for _, token := range []string{
		`@media (max-width: 767px)`,
		`@media (max-width: 359px)`,
		`grid-template-columns: 1fr;`,
		`grid-template-columns: repeat(2, minmax(0, 1fr));`,
		`custom-dates { grid-column: 1;`,
		`filter-fields-row { grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 10px; }`,
		`filter-reset-field { grid-column: 2; justify-self: end; }`,
		`text-overflow: ellipsis;`,
		`filterForm.addEventListener('submit'`,
	} {
		if !strings.Contains(body, token) {
			t.Errorf("mobile filter contract missing %q", token)
		}
	}
}

func TestDashboardResilienceAndAccessibilityContracts(t *testing.T) {
	body := plugin.DashboardHTML
	for _, required := range []string{
		"UA_FETCH_TIMEOUT_MS = 20000", "function fetchWithTimeout", "safeErrorMessage",
		"retryLoadBtn", "요청 시간이 초과되었습니다", "hideSkeletons();",
		"recordsExpanded", "dateKeysForSelectedRange", "chart-wrapper svg",
		"aria-checked", `role:"radiogroup"`, "aria-controls", "aria-labelledby", "ArrowLeft", "ArrowRight",
		"trapDrawerFocus", "drawerOpener", "실패 상태", "failure_status",
		".custom-dates .input:disabled", "setCustomValidity", "시작일은 종료일보다 늦을 수 없습니다.",
	} {
		if !strings.Contains(body, required) {
			t.Errorf("missing dashboard resilience/accessibility contract %q", required)
		}
	}
	if strings.Contains(body, `customDateWrap" class="custom-dates" style=`) {
		t.Error("custom date visibility must use class toggling instead of inline display styles")
	}
}

func TestDashboardNativeSelectOptionsAndNoVisualShadows(t *testing.T) {
	body := plugin.DashboardHTML
	lower := strings.ToLower(body)
	for _, forbidden := range []string{"text-shadow", "drop-shadow", "filter:drop-shadow", "filter: drop-shadow", "<datalist", "suggestions", "filterdebouncetimer", "scheduletextfilterapply", "addeventlistener('input'"} {
		if strings.Contains(lower, forbidden) {
			t.Errorf("dashboard contains forbidden token %q", forbidden)
		}
	}
	for _, required := range []string{
		"--filter-control-height: 38px;", "--filter-control-radius: 8px;", "--filter-control-border: #c7ced8;",
		"--filter-control-focus: #2563eb;", "--filter-label-color: #667085;", "--filter-gap: 8px;",
		`<select id="filterProvider" class="select compact-select"`, `<select id="filterEndpoint" class="select compact-select"`, `<select id="filterModel" class="select compact-select"`, `<select id="filterStatus" class="select compact-select"`,
		"function populateSelectOptions", "option.textContent = label;", "option.dataset.transient = 'true';", "selectEl.value = selectedValue || '';",
		"Object.keys(knownMap).sort", "field.addEventListener('change', applyFilters);", "@media (max-width: 359px)",
	} {
		if !strings.Contains(body, required) {
			t.Errorf("dashboard missing native select contract %q", required)
		}
	}
}

func TestUsageSelectOptionPopulationNodeContract(t *testing.T) {
	nodePath, err := exec.LookPath("node")
	if err != nil || nodePath == "" {
		t.Skip("node not found")
	}
	scriptRe := regexp.MustCompile(`(?s)<script>(.*?)</script>`)
	matches := scriptRe.FindStringSubmatch(plugin.DashboardHTML)
	if len(matches) < 2 {
		t.Fatal("dashboard script not found")
	}
	nodeScript := `
const vm=require('vm');
const script=process.argv[1];
function node(tag){const n={tagName:tag.toUpperCase(),children:[],style:{},dataset:{},setCustomValidity(){},classList:{add(){},remove(){},toggle(){}},appendChild(c){this.children.push(c);this.firstChild=this.children[0]||null;return c;},replaceChildren(){this.children=[];this.firstChild=null;},addEventListener(){},setAttribute(){},getAttribute(){return null;}};Object.defineProperty(n,'selectedOptions',{get(){return this.children.filter(o=>o.value===(this.value||''))}});Object.defineProperty(n,'textContent',{get(){return this._text||''},set(v){this._text=String(v)}});return n;}
const ids={filterProvider:node('select'),filterEndpoint:node('select'),filterModel:node('select'),filterStatus:node('select')};
const document={hidden:false,activeElement:null,addEventListener(){},createElement:node,getElementById(id){return ids[id]||node('div')},querySelectorAll(){return[]}};
const ctx={document,window:{addEventListener(){}},console,Math,Number,String,Array,Object,Date,Intl,isNaN,setTimeout(){return 1},clearTimeout(){},AbortController:function(){this.signal={addEventListener(){},removeEventListener(){}};this.abort=function(){};}};
// Exercise option/query helpers without starting live network initialization in this DOM mock.
vm.createContext(ctx);vm.runInContext(script.replace(/^(?:initDashboard|bindDashboardEvents)\(\);$/gm, ''),ctx);
ctx.state.provider='<img src=x onerror=alert(1)>';
ctx.extractFilterOptions({by_provider:[{key:'zeta'},{key:'Alpha'},{key:'zeta'}],by_endpoint:[{key:'warp2'},{key:'direct'}],by_model:[{key:'m2'},{key:'m1'}],summary:[]},{records:[]});
function assert(ok,msg){if(!ok)throw new Error(msg)}
assert(ids.filterProvider.children[0].value==='', 'all option first');
assert(ids.filterProvider.children[0].textContent==='전체','all label');
assert(ids.filterProvider.children[1].value===ctx.state.provider,'transient value preserved');
assert(ids.filterProvider.children[1].dataset.transient==='true','transient marked');
assert(ids.filterProvider.value===ctx.state.provider,'selection preserved');
assert(ids.filterProvider.title===ctx.state.provider,'tooltip uses full selected text safely');
assert(ids.filterEndpoint.title==='전체' && ids.filterModel.title==='전체','short default tooltip');
assert(ids.filterProvider.children.filter(o=>o.value==='zeta').length===1,'dedupe');
assert(ids.filterProvider.children[1].textContent==='<img src=x onerror=alert(1)>','XSS stays text');
assert(ids.filterEndpoint.children.map(o=>o.value).join(',')===',direct,warp2','endpoint sorted');
assert(ids.filterModel.children.map(o=>o.value).join(',')===',m1,m2','model sorted');
let changes=0;ctx.applyFilters=function(){changes++};const field=node('select');field.addEventListener=function(type,fn){if(type==='change')fn()};field.addEventListener('change',ctx.applyFilters);assert(changes===1,'change applies immediately');
ids.filterModel.value='m2';ctx.updateSelectTitle(ids.filterModel);assert(ids.filterModel.title==='m2','changed selection tooltip');
ids.filterStatus.children=[{value:'',textContent:'전체'},{value:'failed',textContent:'실패만'}];ids.filterStatus.value='failed';ctx.updateSelectTitle(ids.filterStatus);assert(ids.filterStatus.title==='실패만','status uses label not value');
let resetPreset='';ctx.setPreset=function(p){resetPreset=p};ctx.resetFilters();
for(const field of Object.values(ids)){assert(field.value==='','reset value');assert(field.title==='전체','reset tooltip')}
assert(ctx.state.provider==='' && ctx.state.endpoint==='' && ctx.state.model==='' && ctx.state.status==='','reset state');assert(resetPreset==='today','reset date preset unchanged');
console.log('OK_NATIVE_SELECTS');`
	cmd := exec.Command(nodePath, "-e", nodeScript, "--", matches[1])
	out, err := cmd.CombinedOutput()
	if err != nil || !strings.Contains(string(out), "OK_NATIVE_SELECTS") {
		t.Fatalf("native select node contract failed: %v\n%s", err, out)
	}
}

func TestDashboardSnapshotChangeCueContract(t *testing.T) {
	body := plugin.DashboardHTML
	if strings.Contains(body, ":{values:{},rows:{}}; state.serverSnapshot") {
		t.Fatal("snapshot baseline reset must initialize records map before record rendering")
	}
	for _, tok := range []string{
		"function makeUsageSnapshot", "function diffUsageSnapshots", "snapshotFilterKey",
		"state.pendingChanges=", "setDashboardBusy(true)", "setDashboardBusy(false)",
		"if (!appendRecords) scheduleNextPoll(computeAdaptivePollMs(state.records, state.currentPreset));",
		"scheduleNextPoll(UA_POLL_RETRY_MS);",
		"flashValue(elTotalCount", "flashValue(DOM.kpiSubFail", "cell.tone,cell.label)", "flashValue(el,cell.tone,cell.label)",
		".value-flash", "flash-good", "flash-danger", ".value-flash { animation: none; }",
		"function applyPanelPulse", "function pulseKPIGroup", "#breakdownSection.panel-pulse", "#recordsPanel.panel-pulse",
		"id=\"recordsPanel\"", "state.pendingChanges.records", "_panelPulseTimer", "},1000);",
	} {
		if !strings.Contains(body, tok) {
			t.Errorf("dashboard missing snapshot change cue contract %q", tok)
		}
	}
}

func TestRequestDetailModalScrollLockAndFocusRestoreNode(t *testing.T) {
	nodePath, err := exec.LookPath("node")
	if err != nil || nodePath == "" {
		t.Skip("node not found")
	}
	body := plugin.DashboardHTML
	lineFunc := func(name string) string {
		t.Helper()
		match := regexp.MustCompile(`function ` + regexp.QuoteMeta(name) + `\(\) \{[^\n]*`).FindString(body)
		if match == "" {
			t.Fatalf("could not extract %s", name)
		}
		return match
	}
	lock := lineFunc("lockBodyScroll")
	unlock := lineFunc("unlockBodyScroll")
	close := lineFunc("closeDrawer")
	script := `
var state={dialogScrollLock:null,drawerOpener:null};
var DOM={};
var documentElement={clientWidth:980};
var body={style:{overflow:'auto',paddingRight:'3px'}};
var document={body:body,documentElement:documentElement,contains:function(node){return node===opener;},getElementById:function(){return null;}};
var window={innerWidth:1000};
var opener={focus:function(){this.focused=true;}};
var dialog={setAttribute:function(k,v){this[k]=v;}};
var overlay={open:true,classList:{contains:function(c){return c==='open'&&overlay.open;},remove:function(c){if(c==='open')overlay.open=false;}},setAttribute:function(k,v){this[k]=v;}};
DOM.recordDialog=dialog; DOM.detailOverlay=overlay;
` + lock + "\n" + unlock + "\n" + close + `
lockBodyScroll();
if(body.style.overflow!=='hidden'||body.style.paddingRight!=='20px')throw new Error('lock failed');
state.drawerOpener=opener;
closeDrawer();
if(body.style.overflow!=='auto'||body.style.paddingRight!=='3px')throw new Error('unlock failed');
if(!opener.focused)throw new Error('focus was not restored');
var getComputedStyle=function(){return {scrollbarGutter:'stable'};};
lockBodyScroll();
if(body.style.paddingRight!=='3px')throw new Error('stable gutter must not be compensated twice');
unlockBodyScroll();
console.log('DETAIL_MODAL_A11Y_OK');`
	cmd := exec.Command(nodePath, "-e", script)
	out, err := cmd.CombinedOutput()
	if err != nil || !strings.Contains(string(out), "DETAIL_MODAL_A11Y_OK") {
		t.Fatalf("detail modal node contract failed: %v\n%s", err, out)
	}
}

func TestDashboardCenteredControlsAndNumericAlignment(t *testing.T) {
	body := plugin.DashboardHTML
	style := regexp.MustCompile(`(?s)<style>(.*?)</style>`).FindStringSubmatch(body)
	if len(style) != 2 {
		t.Fatal("dashboard stylesheet not found")
	}
	if regexp.MustCompile(`font\s*:[^;{}]*\binherit\s*;`).MatchString(style[1]) {
		t.Error("font shorthand must not mix inherit with other font values")
	}
	for _, tc := range []struct {
		selector string
		required []string
	}{
		{".compact-select", []string{"padding: 0 28px 0 11px;", "text-align: left;", "text-align-last: left;", "font-family: inherit;", "font-size: 0.875rem;", "font-weight: 500;", "line-height: 1.2;"}},
		{".btn", []string{"align-items: center;", "justify-content: center;", "font-family: inherit;", "font-size: 0.875rem;", "font-weight: 600;", "line-height: 1.2;"}},
		{".preset-btn", []string{"display: inline-flex;", "align-items: center;", "justify-content: center;", "min-height: var(--filter-control-height);"}},
		{"th, td", []string{"text-align: left;", "vertical-align: middle;"}},
		{".num-cell", []string{"text-align: right;"}},
		{".text-right", []string{"text-align: right;"}},
	} {
		rule := regexp.MustCompile(`(?m)^  ` + regexp.QuoteMeta(tc.selector) + `\s*\{([^}]*)\}`).FindStringSubmatch(style[1])
		if len(rule) != 2 {
			t.Errorf("missing CSS rule %s", tc.selector)
			continue
		}
		for _, required := range tc.required {
			if !strings.Contains(rule[1], required) {
				t.Errorf("%s missing %q", tc.selector, required)
			}
		}
	}
}

// Inspect actual HTML ancestry/order, not merely the presence of CSS or IDs.
func TestDashboardChartAndFreshnessPlacement(t *testing.T) {
	body := strings.Split(plugin.DashboardHTML, "<script>")[0]
	tags := regexp.MustCompile(`<(/?)([a-zA-Z][a-zA-Z0-9]*)([^>]*)>`).FindAllStringSubmatchIndex(body, -1)
	var stack []string
	parents := map[string]string{}
	positions := map[string]int{}
	attribute := regexp.MustCompile(`(?:id|class)="([^"]+)"`)
	void := map[string]bool{"meta": true, "link": true, "input": true, "br": true, "hr": true, "img": true}
	for _, tag := range tags {
		name := body[tag[4]:tag[5]]
		if body[tag[2]:tag[3]] == "/" {
			if len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}
			continue
		}
		attrs := body[tag[6]:tag[7]]
		identity := name
		for _, attr := range attribute.FindAllStringSubmatch(attrs, -1) {
			if strings.HasPrefix(attr[0], "id=") {
				identity = attr[1]
				break
			}
			identity = attr[1]
		}
		if len(stack) > 0 {
			parents[identity] = stack[len(stack)-1]
		}
		positions[identity] = tag[0]
		if !void[name] && !strings.HasSuffix(strings.TrimSpace(attrs), "/") {
			stack = append(stack, identity)
		}
	}
	for child, parent := range map[string]string{
		"chartCard": "col-left-5", "col-left-5": "usageMainContent", "col-right-7": "usageMainContent",
		"breakdownSection": "col-left-5", "recordsPanel": "col-right-7", "activeFilterBadge": "filter-topmeta", "filter-topmeta": "filter-topline", "filter-topline": "filterForm",
	} {
		if parents[child] != parent {
			t.Errorf("%s parent = %q, want %q", child, parents[child], parent)
		}
	}
	if !(positions["col-left-5"] < positions["breakdownSection"] && positions["breakdownSection"] < positions["chartCard"] && positions["chartCard"] < positions["col-right-7"]) {
		t.Error("left stack must contain aggregate then chart before the records column")
	}
}

func TestDashboardReadableSharedHeadingsAndSemanticContrast(t *testing.T) {
	body := plugin.DashboardHTML
	for _, panel := range []string{"breakdownSection", "recordsPanel"} {
		heading := regexp.MustCompile(`(?s)id="` + panel + `">\s*<div class="card-header panel-heading">\s*<div class="panel-heading-title">\s*<h2 class="card-title">`)
		if !heading.MatchString(body) {
			t.Errorf("%s must use the shared title baseline/heading layout", panel)
		}
	}
	for _, required := range []string{
		".panel-heading { min-height: 48px; flex: 0 0 auto; align-items: center; flex-wrap: wrap; }",
		".breakdown-table-wrap {\n    height: 178px;\n    overflow: auto;\n    border: 0;\n    border-radius: var(--radius-card);\n    background: transparent;",
		"#chartCard, #breakdownSection { height: 260px; min-height: 0; }", ".panel-heading .card-desc { white-space: normal; overflow-wrap: anywhere; }",
		"#breakdownSection .breakdown-table-wrap { flex: 1; min-height: 0; }",
		".table-responsive { height: clamp(280px, 45vh, 440px); max-height: 520px;",
		".chart-card-compact {\n    width: 100%;",
		".chart-wrapper svg { display: block; min-width: 0; margin-inline: auto; }",
		".dashboard-main-grid > .card, .col-left-5 > .card, .col-right-7 > .card { margin-bottom: 0; }",
		"--accent-blue: #3b6593;", "--tok-in: #7c3aed;", "--tok-out: #059669;", "--text-error: #b91c1c;",
		".preset-btn.active { color: #294e76; background: #e8eef5;",
		".tab-btn.active {\n    color: #294e76;\n    background-color: #e8eef5;",

		"font-size: 1rem;\n    font-weight: 700;", "font-size: 0.9rem;", "font-size: 0.84rem;",
		"font-size: clamp(0.95rem, 0.8rem + 0.35vw, 1.125rem);", "font-size: 0.85rem;", "font-size: 0.8rem;",
		"label.setAttribute('font-size','10')",
	} {
		if !strings.Contains(body, required) {
			t.Errorf("missing readability/layout contract %q", required)
		}
	}
	mobile := body[strings.Index(body, "@media (max-width: 767px)"):strings.Index(body, "</style>")]
	// Compact secondary failure metadata may be smaller; primary labels/values remain readable.
	if regexp.MustCompile(`font-size:\s*0\.[0-5]`).MatchString(mobile) {
		t.Error("mobile must not shrink text below the readable base styles")
	}
	if strings.Contains(body, "border-top: 3px solid") {
		t.Error("KPI cards must not have heavy colored top lines")
	}
}

func TestDashboardSimplifiedFilterSurfaces(t *testing.T) {
	body := plugin.DashboardHTML
	for _, required := range []string{"#dashboardContent { min-width: 0; }", ".filter-panel { display: grid; gap: 12px; }", ".filter-topline { display: contents; }", ".filter-detail-section { min-width: 0; }", ".filter-section {\n    min-width: 0;\n    border: 0;\n    border-radius: 0;\n    padding: 0;\n    background: transparent;"} {
		if !strings.Contains(body, required) {
			t.Errorf("missing current layout contract %q", required)
		}
	}
}

func TestDashboardIndependentBoundedPanelHeight(t *testing.T) {
	body := plugin.DashboardHTML
	for _, required := range []string{"#chartCard, #breakdownSection { height: 260px; min-height: 0; }", ".breakdown-table-wrap {\n    height: 178px;\n    overflow: auto;", ".table-responsive { height: clamp(280px, 45vh, 440px); max-height: 520px;"} {
		if !strings.Contains(body, required) {
			t.Errorf("missing current layout contract %q", required)
		}
	}
	for _, forbidden := range []string{"syncUsagePanelHeight", "--usage-panel-height", "#chartCard:not([open])"} {
		if strings.Contains(body, forbidden) {
			t.Errorf("obsolete layout contract remains %q", forbidden)
		}
	}
}

func TestDashboardMobileRecordProgressiveDetail(t *testing.T) {
	body := plugin.DashboardHTML
	for _, token := range []string{"#recordsTableMount .ui-table { min-width: 1100px; }", "table-layout: fixed;", ".record-time-full { display: none; }", ".record-time-mobile { display: inline; }", "mobileTime:mobileTime", "전체 토큰·Provider·Endpoint를 확인", "full:formatDateTime(rawT)"} {
		if !strings.Contains(body, token) {
			t.Errorf("missing mobile progressive records %q", token)
		}
	}
}

func TestDashboardChartSectionAlwaysVisible(t *testing.T) {
	body := plugin.DashboardHTML
	if !strings.Contains(body, `<section id="chartCard" class="card chart-card-compact" aria-labelledby="chartTitle">`) || !strings.Contains(body, `<header class="card-header chart-heading">`) {
		t.Error("daily chart must be an always-visible section with a static heading")
	}
	for _, forbidden := range []string{`<details id="chartCard"`, `<summary class="card-header chart-toggle">`, ".chart-toggle", "#chartCard[open]", "#chartCard:not([open])", "chartCard.addEventListener('toggle'", "chartCard.open"} {
		if strings.Contains(body, forbidden) {
			t.Errorf("chart must not be a collapsible disclosure: %q", forbidden)
		}
	}
	for _, token := range []string{"#chartCard { flex: 0 0 auto; }", "overflow-wrap: anywhere; margin: 4px 0 0; }", ".daily-chart-bar:focus-visible { stroke: #0f172a;", `id="chartArea"`, `id="dailyChartTooltip"`, `id="dailyChartSummary"`} {
		if !strings.Contains(body, token) {
			t.Errorf("missing accessible chart section %q", token)
		}
	}
}

// The filter surface uses actual exported React dohwi-ui components. Native controls stay
// in the DOM as hidden state holders for the legacy query code.
func TestDashboardUsesExportedDohwiUIFilterControls(t *testing.T) {
	body := plugin.DashboardHTML
	for _, required := range []string{
		`<script data-dohwi-ui="exported">`,
		`id="presetGroup" class="dashboard-preset-mount"`, "renderDashboardPresets", "presetGroupControl",
		`role:"radiogroup"`, "ui-segmented-item", "ui-segmented-indicator",
		`<div id="dateRangeMount"></div>`, "filterDateRangeTrigger", `mode:"range"`, "ui-date-picker-trigger", "ui-date-picker-content", "ui-calendar",
		"dashboard-select-mount", "ui-select-trigger", "ui-select-content",
		`<input type="date" id="filterFromDate" class="input" aria-label="시작일" hidden disabled>`,
		`<input type="date" id="filterToDate" class="input" aria-label="종료일" hidden disabled>`,
		"#customDateWrap > input[hidden], .compact-select-wrap > select[hidden] { display: none; }",
	} {
		if !strings.Contains(body, required) {
			t.Errorf("exported dohwi-ui filter control contract missing %q", required)
		}
	}
	// Legacy native selects remain as the state source for filter backend queries.
	for _, id := range []string{"filterProvider", "filterEndpoint", "filterModel", "filterStatus"} {
		if !regexp.MustCompile(`<select id="` + id + `" class="select compact-select"`).MatchString(body) {
			t.Errorf("native select %s must remain as hidden state holder", id)
		}
	}
	if strings.Contains(body, `id="customDateWrap" class="custom-dates is-hidden"`) {
		t.Error("date range wrapper must remain in layout")
	}
}

func TestDashboardCompactPanelsAndAccessibleKPIHelp(t *testing.T) {
	body := plugin.DashboardHTML
	for _, required := range []string{
		"#chartCard, #breakdownSection { height: 260px; min-height: 0; }",
		".panel-heading { min-height: 48px; flex: 0 0 auto;", ".panel-heading .card-title { min-width: 0; flex-wrap: wrap; overflow-wrap: anywhere; }",
		".chart-card-compact {\n    width: 100%;",
		"#breakdownSection { display: flex; flex-direction: column; }",
		".breakdown-table-wrap {\n    height: 178px;\n    overflow: auto;",
		".chart-wrapper {\n    width: 100%;\n    overflow-x: auto;\n    height: 151px;\n    max-height: 160px;",
		".kpi-group-head {\n    padding: 10px 14px;", "padding: 8px 6px;",
		".sr-only { position: absolute; width: 1px; height: 1px;", "clip-path: inset(50%);",
	} {
		if !strings.Contains(body, required) {
			t.Errorf("compact panel contract missing %q", required)
		}
	}
	for _, id := range []string{"grpOverviewHelp", "grpInputHelp", "grpOutputHelp"} {
		if !strings.Contains(body, `aria-describedby="`+id+`"`) || !strings.Contains(body, `class="kpi-group-help sr-only" id="`+id+`"`) {
			t.Errorf("KPI description %s must remain available to assistive technology without occupying visual space", id)
		}
	}
	for _, forbidden := range []string{"clamp(460px, 60vh, 620px)", "min-height: 56px;", "flex: 0 0 56px;"} {
		if strings.Contains(body, forbidden) {
			t.Errorf("oversized panel contract remains: %q", forbidden)
		}
	}
}

func TestDashboardResponsiveDensity(t *testing.T) {
	body := plugin.DashboardHTML
	for _, required := range []string{".filter-period-row { grid-template-columns: minmax(0, 1fr) minmax(280px, 360px);", ".preset-group { grid-column: 1; grid-template-columns: repeat(3, minmax(0, 1fr)); }", ".custom-dates { grid-column: 1; grid-template-columns: minmax(0, 1fr) auto minmax(0, 1fr); }", ".filter-fields-row { grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 10px; }", "min-height: 116px;", ".kpi-metric-row { min-height: 116px;", ".kpi-metric-row { min-height: 138px; }", ".kpi-value-wrap { min-height: 84px; }", ".kpi-metric-row + .kpi-metric-row { border-left: 1px solid var(--border-subtle); }", ".breakdown-table-wrap table { min-width: 0; }"} {
		if !strings.Contains(body, required) {
			t.Errorf("missing current layout contract %q", required)
		}
	}
}

func TestDashboardFluidNativeFilters(t *testing.T) {
	body := plugin.DashboardHTML
	for _, required := range []string{".filter-fields-row { grid-template-columns: minmax(0, 1fr) minmax(0, 1fr) minmax(0, 1.5fr) minmax(0, 0.8fr) 86px;", ".filter-field { position: relative; width: 100%; min-width: 0; display: grid; gap: 5px; }", ".compact-select { width: 100%; min-width: 0;", "text-align: left; text-align-last: left;", ".active-filter-badge {\n    width: 100%;\n    max-width: 100%;\n    height: 24px;\n    overflow: auto;"} {
		if !strings.Contains(body, required) {
			t.Errorf("missing current layout contract %q", required)
		}
	}
	for _, forbidden := range []string{"#dashboardContent { background: var(--bg-card);", "padding: 0 24px 0 68px;"} {
		if strings.Contains(body, forbidden) {
			t.Errorf("obsolete layout contract remains %q", forbidden)
		}
	}
}

func TestDashboardAboveFilterLabelsAndMobileNativeControls(t *testing.T) {
	body := plugin.DashboardHTML
	cssRule := func(selector string) string {
		t.Helper()
		match := regexp.MustCompile(`(?m)^  ` + regexp.QuoteMeta(selector) + ` \{([^}]+)\}`).FindStringSubmatch(body)
		if len(match) != 2 {
			t.Fatalf("missing desktop rule %s", selector)
		}
		return match[1]
	}
	for selector, properties := range map[string][]string{
		".filter-field":       {"position: relative;", "width: 100%;", "min-width: 0;", "display: grid;", "gap: 5px;"},
		".filter-field label": {"font-size: 0.75rem;", "font-weight: 600;", "line-height: 1.3;"},
		".compact-select":     {"padding: 0 28px 0 11px;", "text-align: left;", "text-align-last: left;"},
	} {
		rule := cssRule(selector)
		for _, property := range properties {
			if !strings.Contains(rule, property) {
				t.Errorf("%s missing %q", selector, property)
			}
		}
	}
	if strings.Contains(cssRule(".filter-field label"), "position: absolute;") {
		t.Error("associated labels must occupy their own row above the native select")
	}
	mobileMatch := regexp.MustCompile(`(?s)@media \(max-width: 767px\) \{(.*?)\n  \}`).FindStringSubmatch(body)
	if len(mobileMatch) != 2 {
		t.Fatal("missing mobile rules")
	}
	mobile := mobileMatch[1]
	for _, token := range []string{
		"#filterForm { --filter-control-height: 44px; }",
		".filter-field label { font-size: 0.75rem; }",
		".filter-field .compact-select { padding: 0 28px 0 10px; }",
		".filter-field .input, .filter-field .select, .custom-dates .input { height: 44px; min-height: 44px; font-size: 16px; }",
		".filter-reset-button { height: 44px; min-height: 44px; }",
		".preset-btn { min-height: 44px;",
	} {
		if !strings.Contains(mobile, token) {
			t.Errorf("mobile native control contract missing %q", token)
		}
	}
	for _, field := range []string{"Provider", "Endpoint", "Model", "Status"} {
		if !regexp.MustCompile(`<label for="filter` + field + `">[^<]+</label>\s*<div class="compact-select-wrap">\s*<select id="filter` + field + `"[^>]+aria-label="[^"]+"`).MatchString(body) {
			t.Errorf("%s must retain associated label and accessible native select", field)
		}
	}
}

func TestDashboardAppliedConditionsInFilterTopline(t *testing.T) {
	body := plugin.DashboardHTML
	topline := regexp.MustCompile(`(?s)<div class="filter-topline">.*?<div class="filter-topmeta">\s*<div id="activeFilterBadge"[^>]*role="status"[^>]*aria-live="polite"[^>]*>\s*<span class="active-filter-label">조회 조건</span>\s*<span id="activeFilterText" class="active-filter-text">오늘 · 전체 Provider · 전체 Model</span>\s*</div>\s*</div>\s*</div>\s*<fieldset class="filter-section filter-detail-section">`)
	if !topline.MatchString(body) {
		t.Error("conditions must remain live metadata inside the filter topline, before detail controls")
	}
	if strings.Contains(body, "filter-summary-row") || strings.Count(body, `id="activeFilterBadge"`) != 1 {
		t.Error("separate summary row must be removed without duplicating the badge")
	}
	badge := regexp.MustCompile(`(?s)\.active-filter-badge \{([^}]+)\}`).FindStringSubmatch(body)
	if len(badge) != 2 {
		t.Fatal("missing neutral badge style")
	}
	for _, property := range []string{"margin: 0;", "padding: 0;", "border: 0;", "background: transparent;", "color: var(--text-secondary);"} {
		if !strings.Contains(badge[1], property) {
			t.Errorf("neutral condition summary missing %q", property)
		}
	}
	text := regexp.MustCompile(`(?s)\.active-filter-text \{([^}]+)\}`).FindStringSubmatch(body)
	if len(text) != 2 {
		t.Fatal("missing applied condition text style")
	}
	for _, property := range []string{"overflow-wrap: anywhere;", "white-space: normal;"} {
		if !strings.Contains(text[1], property) {
			t.Errorf("condition summary must visibly wrap in full: missing %q", property)
		}
	}
}

func TestDashboardDailyChartGeometryNode(t *testing.T) {
	nodePath, err := exec.LookPath("node")
	if err != nil {
		t.Skip("node not found")
	}
	body := plugin.DashboardHTML
	script := mustExtractJSFunc(t, body, "getDailyChartGeometry") + `
const assert = (ok, message) => { if (!ok) throw new Error(message); };
for (const width of [280, 1400]) {
  for (const days of [1, 7, 31, 90]) {
    const g = getDailyChartGeometry(days, width);
    assert([g.width,g.slot,g.barWidth,g.labelStride,...g.centers].every(v=>Number.isFinite(v)&&v>0), 'finite positive geometry');
    assert(g.barWidth>=16&&g.barWidth<=36, 'bounded bar width');
    assert(g.centers.length===days, 'one slot per calendar day');
    assert(Math.abs(g.centers[0]+g.centers[days-1]-g.width)<1e-6, 'symmetric full-width extent');
    assert(g.labelStride===Math.max(1,Math.ceil(days/(g.width/60))), 'date label stride');
    assert(g.labelIndices[0]===0&&g.labelIndices.at(-1)===days-1, 'first and last dates visible');
    assert(g.centers[0]-g.barWidth/2>=0&&g.centers.at(-1)+g.barWidth/2<=g.width, 'bars inside SVG');
    if(days===1) assert(g.centers[0]===width/2, 'single day centered');
    if(days<=7) assert(g.width===width, 'short periods fit without horizontal scroll');
    else assert(g.slot>=40&&g.width>=width, 'long period uses readable slots and optional internal scroll');
    assert(g.slot>=g.barWidth, 'bars never overlap');
    for(let k=1;k<days;k++) assert(Math.abs(g.centers[k]-g.centers[k-1]-g.slot)<1e-6, 'even calendar gap');
    for(let k=1;k<g.labelIndices.length;k++) {
      const gap=g.centers[g.labelIndices[k]]-g.centers[g.labelIndices[k-1]];
      assert(gap>=35, 'date labels do not overlap');
    }
    if(days>1) assert(g.centers.at(-1)-g.centers[0]>g.width*.4, 'no left-aligned sparse extent');
  }
}
for(const width of [0,NaN,Infinity,-1]) assert(getDailyChartGeometry(7,width).width===280,'hidden width fallback');
` + mustExtractJSFunc(t, body, "dateKeysForSelectedRange") + `
var state={fromRFC3339:'2025-03-01T00:00:00Z',toRFC3339:'2025-03-07T23:59:59Z'};
const sparse={'2025-03-01':{date:'2025-03-01',requests:4},'2025-03-07':{date:'2025-03-07',requests:2}};
const days=dateKeysForSelectedRange(sparse), g=getDailyChartGeometry(days.length,280);
assert(days.length===7&&sparse[days[3]].requests===0,'sparse dates retain zero calendar slots');
assert(Math.abs(g.centers[6]-g.centers[0]-6*g.slot)<1e-6,'sparse endpoints preserve calendar extent');
`
	cmd := exec.Command(nodePath, "-e", script)
	cmd.Env = append(os.Environ(), "TZ=UTC")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("daily chart geometry failed: %v\n%s", err, out)
	}
}

func TestDashboardDailyChartResponsiveAndEmptyContract(t *testing.T) {
	body := plugin.DashboardHTML
	for _, required := range []string{
		`id="dailyChartSummary" class="card-desc"`, "일 · 총 ", "건 · 최고 ", "선택 기간 요청 없음",
		"if (!maxVal) { appendText(chartArea, 'chart-empty-state'", "chartCard.clientWidth - 24 : 280",
		"geometry.slot >= 64", "var(--accent-blue, #3b6593)", "'#cbd5e1'", "barH = Math.max(6,",
		"setAttribute('role','button')", "setAttribute('data-day',day)", "chartArea.contains(document.activeElement)",
		"focusedBar.focus({ preventScroll: true })", "cancelAnimationFrame(dailyChartRedrawFrame)",
		"new ResizeObserver", "chartResizeObserver.observe(chartCard)",
	} {
		if !strings.Contains(body, required) {
			t.Errorf("responsive chart contract missing %q", required)
		}
	}
	chart := mustExtractJSFunc(t, body, "renderChartFromDaily")
	for _, forbidden := range []string{"innerHTML", "setInterval", "setTimeout", "fetch(", "chartCard.open =", "'#2563eb'", "Math.max(340"} {
		if strings.Contains(chart, forbidden) {
			t.Errorf("chart must not contain %q", forbidden)
		}
	}
	if strings.Contains(body, ".chart-wrapper svg { display: block; min-width: 340px;") {
		t.Error("chart SVG must not force overflow on mobile")
	}
}

func TestDashboardMainGridIndependentRows(t *testing.T) {
	body := plugin.DashboardHTML
	for _, required := range []string{".dashboard-main-grid { display: grid; grid-template-columns: minmax(0, 1fr) minmax(0, 1fr);", "#chartCard { grid-column: 1; grid-row: 1; }", "#breakdownSection { grid-column: 2; grid-row: 1; }", "#recordsPanel { grid-column: 1 / -1; grid-row: 2; }", "#chartCard { grid-column: 1; grid-row: 1; }", "#breakdownSection { grid-column: 1; grid-row: 2; }", "#recordsPanel { grid-column: 1; grid-row: 3; }"} {
		if !strings.Contains(body, required) {
			t.Errorf("missing current layout contract %q", required)
		}
	}
	for _, forbidden := range []string{"syncUsagePanelHeight", "--usage-panel-height"} {
		if strings.Contains(body, forbidden) {
			t.Errorf("obsolete layout contract remains %q", forbidden)
		}
	}
}

func TestDashboardHeadingsAndMetadataNeverClip(t *testing.T) {
	body := plugin.DashboardHTML
	for _, selector := range []string{
		".panel-heading", ".panel-heading-title", ".panel-heading .card-title",
		".panel-heading .card-desc", ".chart-heading", ".chart-heading .card-title",
		".chart-card-compact > .card-desc", ".active-filter-text",
	} {
		rules := regexp.MustCompile(`(?m)^  +`+regexp.QuoteMeta(selector)+`\s*\{([^}]*)\}`).FindAllStringSubmatch(body, -1)
		if len(rules) == 0 {
			t.Fatalf("missing visible heading/metadata rule %s", selector)
		}
		for _, rule := range rules {
			if regexp.MustCompile(`(?:^|;)\s*(?:max-)?height\s*:`).MatchString(rule[1]) {
				if selector != ".chart-card-compact > .card-desc" && selector != ".active-filter-text" {
					t.Errorf("%s must grow to show headings", selector)
				} else if !strings.Contains(rule[1], "overflow: auto;") {
					t.Errorf("%s must make bounded metadata scrollable", selector)
				}
			}
			for _, forbidden := range []string{
				`overflow\s*:\s*hidden`,
				`text-overflow\s*:`, `white-space\s*:\s*nowrap`, `line-clamp\s*:`,
				`flex\s*:\s*0\s+0\s+\d+(?:px|rem|em)`,
			} {
				if regexp.MustCompile(forbidden).MatchString(rule[1]) {
					t.Errorf("%s must grow to show its full contents, got %s", selector, rule[1])
				}
			}
		}
	}
	if !strings.Contains(body, "document.fonts.ready.then(scheduleDailyChartRedraw)") {
		t.Error("shared panel measurements must refresh when web fonts finish loading")
	}
}

func TestDashboardKPISharedContentDrivenRows(t *testing.T) {
	body := plugin.DashboardHTML
	rules := regexp.MustCompile(`(?m)^  +\.kpi-groups(?:, \.skeleton-kpi-groups)?\s*\{([^}]*)\}`).FindAllStringSubmatch(body, -1)
	if len(rules) != 3 {
		t.Fatalf("expected desktop, tablet and mobile grid rules, got %d", len(rules))
	}
	for _, property := range []string{"display: grid;", "grid-auto-rows: 1fr;"} {
		if !strings.Contains(rules[0][1], property) {
			t.Errorf("KPI grid must share tallest intrinsic row size: missing %q", property)
		}
	}
	for _, rule := range rules[1:] {
		if strings.Contains(rule[1], "grid-auto-rows") {
			t.Error("tablet/mobile must inherit shared content-driven KPI row sizing")
		}
	}
	if !strings.Contains(body, "color: var(--text-primary);\n    line-height: 1.35;\n    font-variant-numeric: tabular-nums;") {
		t.Error("KPI value line height must leave room for loaded font glyphs")
	}
	for _, selector := range []string{".kpi-groups", ".kpi-group", ".kpi-group-head", ".kpi-group-metrics", ".kpi-metric-row"} {
		matches := regexp.MustCompile(`(?m)^  +`+regexp.QuoteMeta(selector)+`\s*\{([^}]*)\}`).FindAllStringSubmatch(body, -1)
		for _, match := range matches {
			if regexp.MustCompile(`(?:^|;)\s*(?:max-)?height\s*:`).MatchString(match[1]) {
				t.Errorf("%s must not use fixed KPI heights: %s", selector, match[1])
			}
		}
	}
}
