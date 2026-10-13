package plugin_test

import (
	"os/exec"
	"strings"
	"testing"

	"github.com/dohwi/cliproxyapi-usage-audit/internal/plugin"
)

func TestDashboardStableInteractionLayout(t *testing.T) {
	body := plugin.DashboardHTML
	for _, required := range []string{
		"html { scrollbar-gutter: stable; }",
		".filter-topline { display: contents; }",
		`id="customDateWrap" class="custom-dates" aria-label="조회 기간"`,
		`<div id="dateRangeMount"></div>`,
		`aria-label="시작일" hidden disabled`, `aria-label="종료일" hidden disabled`,
		"var description = conditions.length ? conditions.join(' · ') : '오늘 · 전체 Provider · 전체 Model';",
		"#errorBox { position: fixed;", "#errorBox.is-hidden { display: none; }",
		"#loadMoreBar.is-hidden { display: flex !important; visibility: hidden; }",
		"height: 178px;", "height: clamp(280px, 45vh, 440px);",
		"height: 151px;", "min-height: 116px;", "min-height: 138px;",
		`.breakdown-table-wrap table { table-layout: fixed; }`,
		`#recordsPanel table { table-layout: fixed; min-width: 1100px; }`,
		`<div id="usageDataContent" class="is-loading">`,
		`#usageDataContent.is-loading .kpi-value-wrap > span`,
		`<div class="card" id="breakdownSection">`,
		`<section id="chartCard" class="card chart-card-compact" aria-labelledby="chartTitle">`,
		`gutter.indexOf('stable')!==-1?0:`,
		`badge.setAttribute('aria-label', '적용 조건: ' + description)`,
		`if (!field) return;`,
		`if (message) showError(message, false);`,
		`retry.classList.toggle('is-hidden', retryable === false)`,
		`if (!from.value || !to.value || from.value > to.value)`,
		`<div id="dailyChartSummary" class="card-desc" tabindex="0">`,
	} {
		if !strings.Contains(body, required) {
			t.Errorf("missing selection-stable layout %q", required)
		}
	}
	for _, forbidden := range []string{
		`<details id="chartCard"`, ".custom-dates.is-hidden", "customWrap.classList.toggle('is-hidden'",
		".active-filter-badge:not(.is-visible) { display: none; }",
	} {
		if strings.Contains(body, forbidden) {
			t.Errorf("selection must not insert/remove layout tracks: %q", forbidden)
		}
	}
	if strings.Contains(mustExtractJSFunc(t, body, "renderSkeletons"), "classList.add('is-hidden')") {
		t.Error("loading must keep the real canvas visible, not replace its geometry")
	}
	if strings.Contains(body, "syncUsagePanelHeight") || strings.Contains(body, "--usage-panel-height") {
		t.Error("record counts and aggregate tabs must not resize the desktop canvas")
	}
}

func TestDashboardReadableLatencyNode(t *testing.T) {
	node, err := exec.LookPath("node")
	if err != nil {
		t.Skip("node not found")
	}
	script := mustExtractJSFunc(t, plugin.DashboardHTML, "formatLatency") + `
for (const [input, expected] of [[933.238822,'933ms'],[0.12,'0ms'],[1234,'1.23s'],[null,'-'],[Infinity,'-'],[-1,'-']]) {
  if (formatLatency(input) !== expected) throw new Error('readable latency: '+input);
}
console.log('OK_LATENCY');`
	out, err := exec.Command(node, "-e", script).CombinedOutput()
	if err != nil || !strings.Contains(string(out), "OK_LATENCY") {
		t.Fatalf("latency contract: %v\n%s", err, out)
	}
}

func TestDashboardPresetDatesRemainVisibleAndConsistentNode(t *testing.T) {
	node, err := exec.LookPath("node")
	if err != nil {
		t.Skip("node not found")
	}
	body := plugin.DashboardHTML
	script := mustExtractJSFunc(t, body, "computePresetDates") + "\n" +
		mustExtractJSFunc(t, body, "formatLocalDate") + "\n" +
		mustExtractJSFunc(t, body, "syncPeriodInputs") + "\n" +
		mustExtractJSFunc(t, body, "setPreset") + `
const assert = (v, m) => { if (!v) throw new Error(m); };
const from = {value:'', disabled:true}, to = {value:'', disabled:true};
var DOM = {filterFromDate:from, filterToDate:to};
var state = {currentPreset:'today', fromRFC3339:'', toRFC3339:'', recordsExpanded:true};
let loads = 0, cleared = 0, badges = 0;
const buttons = ['today','yesterday','7d','30d','month','custom'].map(preset => ({
  preset, classList:{toggle(){}}, getAttribute(){return this.preset;},
  setAttribute(name, value){assert(name==='aria-checked'||name==='data-state','radio semantics');if(name==='aria-checked')this.pressed=value;}
}));
let presetRenders = 0;
var window = {renderDashboardPresets(){presetRenders++;}};
var document = {querySelectorAll(){return buttons;}};
function setFilterValidation(message){assert(message==='','preset clears stale validation');cleared++;}
function updateActiveFilterBadge(){badges++;}
function loadData(append){assert(!append,'replace mode');loads++;}
for (const preset of ['today','yesterday','7d','30d','month']) {
  setPreset(preset);
  const bounds = computePresetDates(preset);
  assert(from.disabled && to.disabled,'preset dates are disabled, not removed');
  assert(from.value===formatLocalDate(bounds.from) && to.value===formatLocalDate(bounds.to),'display reflects query');
  const before = [from.value,to.value], beforeLoads = loads;
  setPreset('custom');
  assert(!from.disabled && !to.disabled,'custom enables same date inputs');
  assert(from.value===before[0] && to.value===before[1],'custom preserves displayed period');
  assert(loads===beforeLoads,'custom editing must not send a premature request');
  assert(presetRenders > 0,'selection rendered by React, not direct DOM mutation');
  from.value='2000-01-01';to.value='2000-02-01';
}
from.value='';to.value='';syncPeriodInputs();
assert(from.value && to.value,'missing custom dates initialize from active query');
assert(presetRenders===10,'exported React SegmentedControl re-rendered for every selection');
assert(cleared===10 && badges===10,'all selections clear validation and update summary');
console.log('OK_PRESET_STABILITY');`
	out, err := exec.Command(node, "-e", script).CombinedOutput()
	if err != nil || !strings.Contains(string(out), "OK_PRESET_STABILITY") {
		t.Fatalf("preset date interaction contract: %v\n%s", err, out)
	}
}
