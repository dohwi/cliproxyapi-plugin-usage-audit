// ── DOM cache: populated once at DOMContentLoaded, reused every refresh ──
var DOM = {};
var pendingDohwiTableUpdates = [];

// ── i18n ────────────────────────────────────────────────────────────────────
// The language preference is the only value this page persists (LANG_KEY). The key is
// shared with the other CLIProxyAPI plugin dashboards so one choice applies to all of them.
// Order: saved choice → browser language (ko* → ko, else en) → ko.
var LANG_KEY = 'cpa_plugin_lang';
var I18N = {
  ko: {
    'page.title': '토큰 사용량 대시보드',
    'page.heading': '토큰 사용량',
    'page.sub': '요청과 토큰 사용량을 한눈에 확인하세요.',
    'lang.label': '언어',
    'retry': '다시 시도',
    'period.aria': '기간 선택',
    'preset.today': '오늘', 'preset.yesterday': '어제', 'preset.7d': '7일', 'preset.30d': '30일', 'preset.month': '이번 달', 'preset.custom': '직접',
    'period.custom': '직접 기간', 'period.today': '오늘', 'period.yesterday': '어제', 'period.7d': '최근 7일', 'period.30d': '최근 30일', 'period.month': '이번 달',
    'range.aria': '조회 기간',
    'range.pick': '조회 기간 선택',
    'range.placeholder': '기간 선택',
    'range.title': '직접을 선택하면 날짜를 변경할 수 있습니다.',
    'date.from': '시작일',
    'date.to': '종료일',
    'cal.prev': '이전 달',
    'cal.next': '다음 달',
    'cal.hint': '종료일을 선택하세요',
    'filter.activeLabel': '조회 조건',
    'filter.activeAria': '적용 조건',
    'filter.activeAriaFull': '적용 조건: {d}',
    'filter.default': '오늘 · 전체 Provider · 전체 Model',
    'filter.detail': '세부 필터',
    'filter.providerAria': 'Provider 필터',
    'filter.endpointAria': 'Endpoint 필터',
    'filter.modelAria': 'Model 필터',
    'filter.status': '상태',
    'filter.statusAria': '상태 필터',
    'filter.reset': '초기화',
    'filter.resetAria': '모든 세부 필터 초기화',
    'all': '전체',
    'status.success': '성공만',
    'status.failed': '실패만',
    'validate.both': '시작일과 종료일을 모두 선택하세요.',
    'validate.order': '시작일은 종료일보다 늦을 수 없습니다.',
    'err.unavailable': '데이터 사용 불가',
    'err.http': '요청이 실패했습니다 (HTTP {s}).',
    'err.timeout': '요청 시간이 초과되었습니다. 다시 시도하세요.',
    'err.expanded': '더보기 기록을 보고 있습니다. 최신 데이터를 보려면 필터를 변경하거나 다시 시도하세요.',
    'err.load': '데이터를 불러오지 못했습니다: {e}',
    'flash.suffix': ' · 서버 업데이트로 변경됨',
    'kpi.overview': '개요',
    'kpi.overviewHelp': '공식 집계 기준 요청·전체 수치',
    'kpi.input': '입력',
    'kpi.inputHelp': 'Provider 보고 입력과 캐시 구성',
    'kpi.output': '출력',
    'kpi.outputHelp': '모델 생성 및 추론 토큰',
    'kpi.failPlaceholder': '실패: -건 (-)',
    'kpi.fail': '실패: {n}건 ({r})',
    'kpi.gen': '생성 요청: {n}건',
    'kpi.official': '공식 집계',
    'kpi.inputReported': '입력 (보고값)',
    'kpi.providerBasis': 'Provider 보고 기준',
    'kpi.providerBasisTip': 'Provider 보고값으로 집계 방식에 따라 캐시 포함 여부가 다릅니다.',
    'kpi.noCache': '캐시 제외',
    'kpi.cache': '캐시 입력',
    'kpi.reused': '재사용된 입력',
    'kpi.cacheTip': 'Provider 방식에 따라 캐시 입력은 입력에 포함되거나 별도로 집계됩니다.',
    'kpi.generated': '모델 생성',
    'kpi.reasoning': '추론',
    'kpi.noDouble': '중복 합산 제외',
    'kpi.reasoningTip': '추론은 출력 또는 전체 토큰에 포함될 수 있어 중복 합산하지 않습니다.',
    'kpi.entities': 'Provider {p}개 | Model {m}개',
    'lbl.requests': '총 요청',
    'lbl.failedRequests': '실패 요청',
    'lbl.tokens': '전체 토큰',
    'lbl.inputReported': '입력 보고값',
    'lbl.fresh': '신규 입력',
    'lbl.output': '출력 토큰',
    'lbl.reasoning': '추론 토큰',
    'lbl.cache': '캐시 재사용 토큰',
    'aria.kv': '{l}: {v}',
    'breakdown.title': '상세 집계',
    'breakdown.desc': '필터 기준별 사용량',
    'breakdown.aria': '상세 집계 기준',
    'chart.title': '사용량 추이',
    'chart.daily': '일별 요청량',
    'chart.summary': '{d}일 · 총 {r}건 · 최고 {m}건',
    'chart.empty': '선택 기간 요청 없음',
    'chart.aria': '날짜별 요청 추이 막대 차트',
    'chart.bar': '{date}, 요청 {n}, 상세 보기',
    'tip.requests': '요청: ',
    'tip.total': '전체 토큰 (공식): ',
    'tip.input': '입력 (보고값): ',
    'tip.output': '출력: ',
    'tip.failed': '실패: ',
    'tip.cache': '캐시 (보고값, 비가산): ',
    'tip.reasoning': '추론 (보고값, 비가산): ',
    'th.time': '시간',
    'th.requests': '요청',
    'th.failed': '실패',
    'th.tokens': '토큰',
    'th.fresh': '신규 입력',
    'th.cache': '캐시',
    'th.output': '출력',
    'th.total': '합계',
    'th.latency': '지연',
    'th.status': '상태',
    'tbl.provider': 'Provider 상세 집계',
    'tbl.endpoint': 'Endpoint 상세 집계',
    'tbl.model': 'Model 상세 집계',
    'tbl.records': '최근 요청 기록 테이블 스크롤 영역',
    'loading.agg': '불러오는 중...',
    'loading.records': '데이터를 불러오는 중입니다...',
    'empty.data': '데이터 없음',
    'empty.records': '표시할 기록이 없습니다.',
    'records.title': '최근 요청 기록',
    'records.count': '{n}건',
    'records.desc': '최신 20건 · 필터와 더보기로 전체 로그 추적',
    'records.hint': '기록을 누르면 전체 토큰·Provider·Endpoint를 확인할 수 있습니다.',
    'records.more': '기록 더보기',
    'rec.success': '성공',
    'rec.failed': '실패',
    'rec.status': '요청 상태',
    'dialog.title': '요청 상세 정보',
    'dialog.close': '닫기',
    'dlg.basic': '기본 요청 정보',
    'dlg.time': '요청 시각',
    'dlg.status': '상태',
    'dlg.failureStatus': '실패 상태',
    'dlg.tokens': '토큰 및 응답 속도',
    'dlg.total': '총 토큰 (Authoritative)',
    'dlg.input': '입력(캐시 포함) (Input)',
    'dlg.fresh': '신규 입력 (Fresh)',
    'dlg.output': '출력 토큰 (Output)',
    'dlg.reasoning': '추론 토큰 (Reasoning)',
    'dlg.cached': '캐시 입력 (Cached)',
    'dlg.latency': '총 Latency',
    'dlg.ttft': 'TTFT (첫 토큰 시간)'
  },
  en: {
    'page.title': 'Token Usage Dashboard',
    'page.heading': 'Token Usage',
    'page.sub': 'See requests and token usage at a glance.',
    'lang.label': 'Language',
    'retry': 'Retry',
    'period.aria': 'Select period',
    'preset.today': 'Today', 'preset.yesterday': 'Yesterday', 'preset.7d': '7 days', 'preset.30d': '30 days', 'preset.month': 'This month', 'preset.custom': 'Custom',
    'period.custom': 'Custom range', 'period.today': 'Today', 'period.yesterday': 'Yesterday', 'period.7d': 'Last 7 days', 'period.30d': 'Last 30 days', 'period.month': 'This month',
    'range.aria': 'Date range',
    'range.pick': 'Select date range',
    'range.placeholder': 'Select period',
    'range.title': 'Choose Custom to change the dates.',
    'date.from': 'Start date',
    'date.to': 'End date',
    'cal.prev': 'Previous month',
    'cal.next': 'Next month',
    'cal.hint': 'Select an end date',
    'filter.activeLabel': 'Filters',
    'filter.activeAria': 'Active filters',
    'filter.activeAriaFull': 'Active filters: {d}',
    'filter.default': 'Today \u00b7 All providers \u00b7 All models',
    'filter.detail': 'Detailed filters',
    'filter.providerAria': 'Provider filter',
    'filter.endpointAria': 'Endpoint filter',
    'filter.modelAria': 'Model filter',
    'filter.status': 'Status',
    'filter.statusAria': 'Status filter',
    'filter.reset': 'Reset',
    'filter.resetAria': 'Reset all detailed filters',
    'all': 'All',
    'status.success': 'Successful only',
    'status.failed': 'Failed only',
    'validate.both': 'Select both a start and an end date.',
    'validate.order': 'The start date cannot be after the end date.',
    'err.unavailable': 'Data unavailable',
    'err.http': 'Request failed (HTTP {s}).',
    'err.timeout': 'The request timed out. Please try again.',
    'err.expanded': 'You are viewing loaded records. Change a filter or retry to see the latest data.',
    'err.load': 'Failed to load data: {e}',
    'flash.suffix': ' \u00b7 Changed by server update',
    'kpi.overview': 'Overview',
    'kpi.overviewHelp': 'Requests and totals from official aggregation',
    'kpi.input': 'Input',
    'kpi.inputHelp': 'Provider-reported input and cache breakdown',
    'kpi.output': 'Output',
    'kpi.outputHelp': 'Model-generated and reasoning tokens',
    'kpi.failPlaceholder': 'Failed: - (-)',
    'kpi.fail': 'Failed: {n} ({r})',
    'kpi.gen': 'Generation requests: {n}',
    'kpi.official': 'Official count',
    'kpi.inputReported': 'Input (reported)',
    'kpi.providerBasis': 'Provider-reported',
    'kpi.providerBasisTip': 'Reported by the provider; whether cache is included depends on the provider.',
    'kpi.noCache': 'Excludes cache',
    'kpi.cache': 'Cached input',
    'kpi.reused': 'Reused input',
    'kpi.cacheTip': 'Depending on the provider, cached input is either included in input or counted separately.',
    'kpi.generated': 'Model generated',
    'kpi.reasoning': 'Reasoning',
    'kpi.noDouble': 'Not double-counted',
    'kpi.reasoningTip': 'Reasoning may already be included in output or total tokens, so it is not added again.',
    'kpi.entities': 'Providers {p} | Models {m}',
    'lbl.requests': 'Total requests',
    'lbl.failedRequests': 'Failed requests',
    'lbl.tokens': 'Total tokens',
    'lbl.inputReported': 'Reported input',
    'lbl.fresh': 'Fresh input',
    'lbl.output': 'Output tokens',
    'lbl.reasoning': 'Reasoning tokens',
    'lbl.cache': 'Cached tokens reused',
    'aria.kv': '{l}: {v}',
    'breakdown.title': 'Breakdown',
    'breakdown.desc': 'Usage by filter dimension',
    'breakdown.aria': 'Breakdown dimension',
    'chart.title': 'Usage trend',
    'chart.daily': 'Daily requests',
    'chart.summary': '{d} days \u00b7 {r} requests total \u00b7 peak {m}',
    'chart.empty': 'No requests in the selected period',
    'chart.aria': 'Bar chart of daily requests',
    'chart.bar': '{date}, {n} requests, view details',
    'tip.requests': 'Requests: ',
    'tip.total': 'Total tokens (official): ',
    'tip.input': 'Input (reported): ',
    'tip.output': 'Output: ',
    'tip.failed': 'Failed: ',
    'tip.cache': 'Cached (reported, non-additive): ',
    'tip.reasoning': 'Reasoning (reported, non-additive): ',
    'th.time': 'Time',
    'th.requests': 'Requests',
    'th.failed': 'Failed',
    'th.tokens': 'Tokens',
    'th.fresh': 'New input',
    'th.cache': 'Cache',
    'th.output': 'Output',
    'th.total': 'Total',
    'th.latency': 'Latency',
    'th.status': 'Status',
    'tbl.provider': 'Provider breakdown',
    'tbl.endpoint': 'Endpoint breakdown',
    'tbl.model': 'Model breakdown',
    'tbl.records': 'Recent requests table scroll area',
    'loading.agg': 'Loading...',
    'loading.records': 'Loading data...',
    'empty.data': 'No data',
    'empty.records': 'No records to display.',
    'records.title': 'Recent requests',
    'records.count': '{n} records',
    'records.count.one': '1 record',
    'records.desc': 'Latest 20 \u00b7 Trace the full log with filters and Load more',
    'records.hint': 'Tap a record to see full tokens, provider and endpoint.',
    'records.more': 'Load more',
    'rec.success': 'Success',
    'rec.failed': 'Failed',
    'rec.status': 'Request status',
    'dialog.title': 'Request details',
    'dialog.close': 'Close',
    'dlg.basic': 'Request info',
    'dlg.time': 'Requested at',
    'dlg.status': 'Status',
    'dlg.failureStatus': 'Failure status',
    'dlg.tokens': 'Tokens & latency',
    'dlg.total': 'Total tokens (authoritative)',
    'dlg.input': 'Input (incl. cache)',
    'dlg.fresh': 'Fresh input',
    'dlg.output': 'Output tokens',
    'dlg.reasoning': 'Reasoning tokens',
    'dlg.cached': 'Cached input',
    'dlg.latency': 'Total latency',
    'dlg.ttft': 'TTFT (time to first token)'
  }
};
var currentLang = 'ko';
var lastSummaryData = null;

function normalizeLang(value) {
  var v = String(value || '').toLowerCase();
  if (v.indexOf('ko') === 0) return 'ko';
  if (v.indexOf('en') === 0) return 'en';
  return null;
}
function readStoredLang() {
  try { return normalizeLang(localStorage.getItem(LANG_KEY)); } catch (e) { return null; }
}
function writeStoredLang(lang) {
  try { localStorage.setItem(LANG_KEY, lang); } catch (e) {}
}
function detectLang() {
  var stored = readStoredLang();
  if (stored) return stored;
  try {
    var nav = typeof navigator !== 'undefined' ? navigator : null;
    var list = nav && (nav.languages && nav.languages.length ? nav.languages : [nav.language]);
    if (list && list[0]) return normalizeLang(list[0]) || 'en';
  } catch (e) {}
  return 'ko';
}
currentLang = detectLang();
function uiLocale() { return currentLang === 'en' ? 'en-US' : 'ko-KR'; }

function t(key, params) {
  var table = I18N[currentLang] || I18N.ko;
  var s = table[key];
  if (s == null) s = I18N.ko[key];
  if (s == null) return key;
  if (params) s = s.replace(/\{(\w+)\}/g, function(m, k) { return params[k] != null ? String(params[k]) : m; });
  return s;
}
// Count-aware lookup: English has a '.one' variant where grammar needs it.
function tn(key, n, params) {
  var p = params || {};
  p.n = n;
  if (currentLang === 'en' && Number(n) === 1 && Object.prototype.hasOwnProperty.call(I18N.en, key + '.one')) key += '.one';
  return t(key, p);
}
function kvLabel(labelKey, value) { return t('aria.kv', { l: t(labelKey), v: value }); }

function eachNode(selector, fn) {
  if (typeof document === 'undefined' || !document.querySelectorAll) return;
  var list = document.querySelectorAll(selector);
  if (list && list.forEach) list.forEach(fn);
}
function applyStaticI18n() {
  if (typeof document === 'undefined') return;
  var root = document.documentElement;
  if (root && root.setAttribute) root.setAttribute('lang', currentLang);
  try { document.title = t('page.title'); } catch (e) {}
  eachNode('[data-i18n]', function(el) { el.textContent = t(el.getAttribute('data-i18n')); });
  eachNode('[data-i18n-aria]', function(el) { el.setAttribute('aria-label', t(el.getAttribute('data-i18n-aria'))); });
  eachNode('[data-i18n-title]', function(el) { el.setAttribute('title', t(el.getAttribute('data-i18n-title'))); });
  // KPI values keep their numeric aria text; only the label part follows the language.
  eachNode('[data-i18n-label]', function(el) {
    var shown = String(el.textContent || '');
    el.setAttribute('aria-label', kvLabel(el.getAttribute('data-i18n-label'), /\d/.test(shown) ? shown : '0'));
  });
  eachNode('[data-i18n-count]', function(el) { el.textContent = tn(el.getAttribute('data-i18n-count'), (state.records && state.records.length) || 0); });
  eachNode('[data-lang]', function(btn) {
    btn.setAttribute('aria-pressed', btn.getAttribute('data-lang') === currentLang ? 'true' : 'false');
  });
}
function setLanguage(lang, fromStorage) {
  var next = normalizeLang(lang);
  if (!next || next === currentLang) return;
  currentLang = next;
  if (!fromStorage) writeStoredLang(next);
  applyStaticI18n();
  if (typeof window !== 'undefined' && window.renderDashboardLanguage) window.renderDashboardLanguage();
  // Re-derive everything that was built from translated text.
  extractFilterOptions(null, null);
  [DOM.filterProvider, DOM.filterEndpoint, DOM.filterModel, DOM.filterStatus].forEach(function(field) { if (field) updateSelectTitle(field); });
  updateActiveFilterBadge();
  if (lastSummaryData) renderSummary(lastSummaryData);
  if (state.hasLoadedData) paintRecords();
  var overlay = DOM.detailOverlay || document.getElementById('detailOverlay');
  if (overlay && overlay.classList && overlay.classList.contains('open') && state.openRecordIdx != null) showRecordDetail(state.openRecordIdx, state.drawerOpener);
}
function initLanguageControls() {
  eachNode('[data-lang]', function(btn) {
    btn.addEventListener('click', function() { setLanguage(btn.getAttribute('data-lang')); });
  });
  if (typeof window !== 'undefined' && window.addEventListener) {
    window.addEventListener('storage', function(e) {
      if (e && e.key === LANG_KEY) setLanguage(e.newValue, true);
    });
  }
  applyStaticI18n();
}


// State management
var state = {
  currentPreset: 'today',
  fromRFC3339: '',
  toRFC3339: '',
  provider: '',
  endpoint: '',
  model: '',
  status: '',
  cursor: null,
  hasMore: false,
  records: [],
  recordsExpanded: false,
  drawerOpener: null,
  requestID: 0,
  pollTimer: null,
  inFlight: false,
  abortCtrl: null,
  hasLoadedData: false,
  serverSnapshot: null,
  snapshotFilterKey: '',
  pendingChanges: null
};

// Adaptive polling intervals (ms) for usage-audit dashboard
// today with records within 2m => 30s; activity today but older => 90s;
// quiet today => 5m; historical presets => 10m.
var UA_POLL_RECENT_MS  = 30000;   // today, latest record within 2 min
var UA_POLL_ACTIVE_MS  = 90000;   // today, activity but older
var UA_POLL_QUIET_MS   = 300000;  // today, no records
var UA_POLL_HISTORY_MS = 600000;  // historical preset
var UA_POLL_RETRY_MS   = 120000;  // retry after failure
var UA_POLL_MAX_MS     = 600000;  // 10m cap
var UA_FETCH_TIMEOUT_MS = 20000;   // requests must not leave the UI loading indefinitely

// Escapes text for the rare static HTML-string fallback. Dynamic API data is
// rendered with textContent/title instead, so quote characters never cross an
// HTML attribute boundary.
function esc(str) {
  if (str == null) return '';
  return String(str).replace(/[&<>"']/g, function(ch) {
    return { '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }[ch];
  });
}

// Exact number formatter: comma-grouped integer. KPI/records/breakdowns must
// NEVER use compact K/M/B notation; every figure is shown in full precision.
// null/undefined/NaN => '—'
// 0 => '0'
// negative => '-1,234'
// integer, grouped with commas (en-US) regardless of host locale
function formatExactNumber(num) {
  if (num === null || num === undefined) return '—';
  var n = Number(num);
  if (isNaN(n)) return '—';
  return Math.round(n).toLocaleString('en-US');
}

// Provider display mapper: trims/lowers and turns compatible-b.ai aliases into "b.ai"
function formatProvider(val) {
  if (!val) return '-';
  var s = String(val).trim();
  var low = s.toLowerCase();
  if (low === 'openai compatible-b.ai' || low === 'compatible-b.ai' || low === 'openai-compatible-b.ai') {
    return 'b.ai';
  }
  return s;
}

function formatLatency(ms) {
  if (ms === null || ms === undefined || ms === '') return '-';
  var n = Number(ms);
  if (!Number.isFinite(n) || n < 0) return '-';
  if (n > 1000000) { // nanoseconds
    n = Math.round(n / 1000000);
  }
  if (n >= 1000) {
    return (n / 1000).toFixed(2) + 's';
  }
  return Math.round(n) + 'ms';
}

function formatDateTime(val) {
  if (!val) return '-';
  try {
    var d = new Date(val);
    if (!isNaN(d.getTime())) {
      return d.toLocaleString(uiLocale());
    }
  } catch (e) {}
  return String(val);
}

// Tabs switching in left panel
function switchTab(targetId, btn) {
  var panes = document.querySelectorAll('.tab-pane');
  var tabs = Array.prototype.slice.call(document.querySelectorAll('.tab-btn'));
  var section = document.getElementById('breakdownSection');
  var from = -1, to = tabs.indexOf(btn);
  tabs.forEach(function(t, i) { if (t.classList.contains('active')) from = i; });
  // Same contract as dohwi-ui Tabs: the new panel enters from the side it sits on.
  if (section && from >= 0 && to >= 0 && from !== to) {
    section.style.setProperty('--tabs-enter-x', (to > from ? 18 : -18) + 'px');
    section.setAttribute('data-tabs-direction', to > from ? 'forward' : 'backward');
  }
  panes.forEach(function(pane) { pane.classList.toggle('active', pane.id === targetId); });
  tabs.forEach(function(t) { var selected = t === btn; t.classList.toggle('active', selected); t.setAttribute('aria-selected', selected ? 'true' : 'false'); t.setAttribute('data-state', selected ? 'active' : 'inactive'); t.tabIndex = selected ? 0 : -1; });
  if (btn) btn.focus();
}
function computePresetDates(preset) {
  var now = new Date(), start = new Date(now.getFullYear(), now.getMonth(), now.getDate()), end = new Date(now.getFullYear(), now.getMonth(), now.getDate(), 23, 59, 59, 999);
  if (preset === 'yesterday') { start.setDate(start.getDate() - 1); end = new Date(start.getFullYear(), start.getMonth(), start.getDate(), 23, 59, 59, 999); }
  else if (preset === '7d') start.setDate(start.getDate() - 6); else if (preset === '30d') start.setDate(start.getDate() - 29);
  else if (preset === 'month') start = new Date(now.getFullYear(), now.getMonth(), 1); else if (preset !== 'today') return null;
  return { from: start.toISOString(), to: end.toISOString() };
}
function setPreset(preset) {
  state.currentPreset = preset;
  if (window.renderDashboardPresets) window.renderDashboardPresets();
  else document.querySelectorAll('#presetGroup .preset-btn').forEach(function(btn) { var active = btn.getAttribute('data-preset') === preset; btn.classList.toggle('active', active); btn.setAttribute('aria-checked', active ? 'true' : 'false'); btn.setAttribute('data-state', active ? 'checked' : 'unchecked'); btn.tabIndex = active ? 0 : -1; });
  setFilterValidation('');
  syncPeriodInputs();
  updateActiveFilterBadge();
  if (preset !== 'custom') { state.recordsExpanded = false; loadData(false); }
}
function formatLocalDate(value) {
  var date = new Date(value);
  return date.getFullYear() + '-' + String(date.getMonth() + 1).padStart(2, '0') + '-' + String(date.getDate()).padStart(2, '0');
}
function syncPeriodInputs() {
  var custom = state.currentPreset === 'custom';
  var bounds = custom ? null : computePresetDates(state.currentPreset);
  if (bounds) { state.fromRFC3339 = bounds.from; state.toRFC3339 = bounds.to; }
  var from = DOM.filterFromDate || document.getElementById('filterFromDate'), to = DOM.filterToDate || document.getElementById('filterToDate');
  [from, to].forEach(function(input, index) {
    if (!input) return;
    input.disabled = !custom;
    // Starting custom editing preserves the displayed query, not stale hidden dates.
    if (!custom || !input.value) {
      var value = index === 0 ? state.fromRFC3339 : state.toRFC3339;
      if (value) input.value = formatLocalDate(value);
    }
  });
}
function setFilterValidation(message) {
  var from = DOM.filterFromDate || document.getElementById('filterFromDate'), to = DOM.filterToDate || document.getElementById('filterToDate');
  [from, to].forEach(function(input) { if (input) input.setCustomValidity(message || ''); });
  if (message) showError(message, false); else clearError();
}
var knownProviders = {};
var knownEndpoints = {};
var knownModels = {};

function addKnownOption(knownMap, value) {
  if (value === null || value === undefined) return;
  var normalized = String(value).trim();
  if (normalized && normalized !== '-' && normalized !== 'undefined') knownMap[normalized] = true;
}

function updateSelectTitle(selectEl) {
  if (!selectEl) return;
  var selected = selectEl.selectedOptions && selectEl.selectedOptions[0];
  selectEl.title = selected ? selected.textContent : '';
}

function populateSelectOptions(selectEl, knownMap, allLabel, selectedValue, formatter) {
  if (!selectEl) return;
  var keys = Object.keys(knownMap).sort(function(a, b) { return a.localeCompare(b, currentLang, { sensitivity: 'base' }); });
  selectEl.replaceChildren();
  function appendOption(value, label, transient) {
    var option = document.createElement('option');
    option.value = value;
    option.textContent = label;
    if (transient) option.dataset.transient = 'true';
    selectEl.appendChild(option);
  }
  appendOption('', allLabel || t('all'), false);
  if (selectedValue && keys.indexOf(selectedValue) === -1) appendOption(selectedValue, formatter ? formatter(selectedValue) : selectedValue, true);
  keys.forEach(function(key) { appendOption(key, formatter ? formatter(key) : key, false); });
  selectEl.value = selectedValue || '';
  updateSelectTitle(selectEl);
}

function extractFilterOptions(summaryData, recordsData) {
  if (summaryData) {
    [['by_provider', knownProviders], ['by_endpoint', knownEndpoints], ['by_model', knownModels]].forEach(function(pair) {
      var list = summaryData[pair[0]];
      if (Array.isArray(list)) list.forEach(function(item) { addKnownOption(pair[1], item && (item.key || item.name)); });
    });
    if (Array.isArray(summaryData.summary)) summaryData.summary.forEach(function(item) {
      if (!item) return;
      addKnownOption(knownProviders, item.provider);
      addKnownOption(knownModels, item.model);
    });
    [['providers', knownProviders, ['provider','key','name']], ['endpoints', knownEndpoints, ['endpoint','key','name','executor_type']], ['models', knownModels, ['model','key','name']]].forEach(function(spec) {
      var list = summaryData[spec[0]];
      if (!Array.isArray(list)) return;
      list.forEach(function(item) {
        if (!item) return;
        for (var i = 0; i < spec[2].length; i++) if (item[spec[2][i]]) { addKnownOption(spec[1], item[spec[2][i]]); break; }
      });
    });
  }
  var recs = Array.isArray(recordsData) ? recordsData : (recordsData && Array.isArray(recordsData.records) ? recordsData.records : []);
  recs.forEach(function(record) {
    if (!record) return;
    addKnownOption(knownProviders, record.provider || record.Provider);
    addKnownOption(knownEndpoints, record.endpoint || record.executor_type || record.ExecutorType || record.source || record.Source);
    addKnownOption(knownModels, record.model || record.Model || record.alias || record.Alias);
  });
  populateSelectOptions(DOM.filterProvider || document.getElementById('filterProvider'), knownProviders, t('all'), state.provider, formatProvider);
  populateSelectOptions(DOM.filterEndpoint || document.getElementById('filterEndpoint'), knownEndpoints, t('all'), state.endpoint);
  populateSelectOptions(DOM.filterModel || document.getElementById('filterModel'), knownModels, t('all'), state.model);
}

function updateActiveFilterBadge() {
  var badge = DOM.activeFilterBadge || document.getElementById('activeFilterBadge');
  var textEl = DOM.activeFilterText || document.getElementById('activeFilterText');
  var conditions = [];
  if (state.currentPreset !== 'today') conditions.push(t('period.' + state.currentPreset));
  if (state.provider) conditions.push('Provider: ' + state.provider);
  if (state.endpoint) conditions.push('Endpoint: ' + state.endpoint);
  if (state.model) conditions.push('Model: ' + state.model);
  if (state.status) conditions.push(state.status === 'failed' ? t('status.failed') : t('status.success'));
  var active = conditions.length > 0;
  var description = conditions.length ? conditions.join(' · ') : t('filter.default');
  if (textEl) textEl.textContent = description;
  if (badge) { badge.classList.toggle('is-visible', active); badge.tabIndex = 0; badge.setAttribute('aria-label', t('filter.activeAriaFull', { d: description })); }
}

function applyFilters() {
  var from = DOM.filterFromDate || document.getElementById('filterFromDate'), to = DOM.filterToDate || document.getElementById('filterToDate');
  if (state.currentPreset === 'custom') {
    if (!from.value || !to.value) { setFilterValidation(t('validate.both')); (from.value ? to : from).reportValidity(); return; }
    if (from.value > to.value) { setFilterValidation(t('validate.order')); from.reportValidity(); return; }
    setFilterValidation(''); var sParts = from.value.split('-'), eParts = to.value.split('-');
    state.fromRFC3339 = new Date(Number(sParts[0]), Number(sParts[1])-1, Number(sParts[2])).toISOString();
    state.toRFC3339 = new Date(Number(eParts[0]), Number(eParts[1])-1, Number(eParts[2]), 23, 59, 59, 999).toISOString();
  }
  state.provider = (DOM.filterProvider || document.getElementById('filterProvider')).value; state.endpoint = (DOM.filterEndpoint || document.getElementById('filterEndpoint')).value; state.model = (DOM.filterModel || document.getElementById('filterModel')).value; state.status = (DOM.filterStatus || document.getElementById('filterStatus')).value;
  state.recordsExpanded = false; updateActiveFilterBadge(); loadData(false);
}
function resetFilters() {
  [DOM.filterProvider || document.getElementById('filterProvider'), DOM.filterEndpoint || document.getElementById('filterEndpoint'), DOM.filterModel || document.getElementById('filterModel'), DOM.filterStatus || document.getElementById('filterStatus'), DOM.filterFromDate || document.getElementById('filterFromDate'), DOM.filterToDate || document.getElementById('filterToDate')].forEach(function(input) { input.value = ''; if (input.tagName === 'SELECT') updateSelectTitle(input); });
  state.provider = ''; state.endpoint = ''; state.model = ''; state.status = ''; setFilterValidation(''); updateActiveFilterBadge(); setPreset('today');
}
function buildQueryString(isRecords, cursor) {
  var params = [];
  if (state.fromRFC3339) params.push('from=' + encodeURIComponent(state.fromRFC3339));
  if (state.toRFC3339) params.push('to=' + encodeURIComponent(state.toRFC3339));
  if (state.provider) params.push('provider=' + encodeURIComponent(state.provider));
  if (state.model) params.push('model=' + encodeURIComponent(state.model));
  if (state.status) params.push('status=' + encodeURIComponent(state.status));
  if (state.endpoint) params.push('endpoint=' + encodeURIComponent(state.endpoint));
  if (!isRecords) {
    params.push('tz_offset=' + encodeURIComponent(String(new Date().getTimezoneOffset())));
    try {
      var timeZone = Intl.DateTimeFormat().resolvedOptions().timeZone;
      if (timeZone) params.push('time_zone=' + encodeURIComponent(timeZone));
    } catch (_) {}
  }
  if (isRecords) {
    params.push('limit=20');
    if (cursor) {
      params.push('cursor=' + encodeURIComponent(cursor));
    }
  }
  return params.length > 0 ? ('?' + params.join('&')) : '';
}

function showError(msg, retryable) { var box = DOM.errorBox || document.getElementById('errorBox'), text = DOM.errorMessage || document.getElementById('errorMessage'), retry = DOM.retryLoadBtn || document.getElementById('retryLoadBtn'); if (text) text.textContent = msg || t('err.unavailable'); if (retry) retry.classList.toggle('is-hidden', retryable === false); if (box) box.classList.remove('is-hidden'); }
function clearError() { var box = DOM.errorBox || document.getElementById('errorBox'), text = DOM.errorMessage || document.getElementById('errorMessage'); if (box) box.classList.add('is-hidden'); if (text) text.textContent = ''; }
function safeErrorMessage(response) { return response.text().then(function(text) { var fallback = t('err.http', { s: response.status }); if (!text) return fallback; try { var parsed = JSON.parse(text); text = parsed.message || parsed.error || fallback; } catch (_) {} return String(text).replace(/\s+/g, ' ').slice(0, 300) || fallback; }); }
// Snapshot comparison is only performed for a repeated server query. Filter-driven
// queries reset the baseline, so a UI rerender never looks like a data change.
function snapshotFilterKey() { return [state.currentPreset,state.fromRFC3339,state.toRFC3339,state.provider,state.endpoint,state.model,state.status].join('|'); }
function makeUsageSnapshot(summary, records) {
  var snap={ values:{}, rows:{} };
  if (summary) {
    var total=typeof summary.total_count==='number'?summary.total_count:0, failures=0, tokens=0, input=0, fresh=0, output=0, reasoning=0, cached=0;
    (Array.isArray(summary.summary)?summary.summary:[]).forEach(function(item){item=item||{}; failures+=Number(item.failures)||0;tokens+=Number(item.total_tokens)||0;input+=Number(item.input_tokens)||0;fresh+=Number(item.fresh_input_tokens)||0;output+=Number(item.output_tokens)||0;reasoning+=Number(item.reasoning_tokens)||0;cached+=Number(item.cached_tokens)||0; if(!total)total+=Number(item.requests)||0;});
    snap.values={ requests:total, failures:failures, tokens:tokens, input:input, fresh:fresh, output:output, reasoning:reasoning, cached:cached };
    [['provider',summary.by_provider],['endpoint',summary.by_endpoint],['model',summary.by_model]].forEach(function(pair){(Array.isArray(pair[1])?pair[1]:[]).forEach(function(row){row=row||{};var rawKey=String(row.key||row[pair[0]]||row.name||row.endpoint||row.source||row.executor_type||''); if(!rawKey) return; var key=pair[0]+':'+rawKey;snap.rows[key]={requests:Number(row.requests||row.count)||0,failures:Number(row.failures)||0,tokens:Number(row.total_tokens)||0};});});
  }
  (Array.isArray(records)?records:(records&&records.records)||[]).forEach(function(row,index){row=row||{};var key=String(row.id||row.request_id||row.requested_at||row.t||row.Timestamp||index);snap.rows['record:'+key]=!!(row.failed||row.Failed);}); return snap;
}
function usageTone(before, after, badWhenUp) { if (before === after) return ''; if (typeof before==='number' && typeof after==='number') return ((after>before) !== !!badWhenUp)?'flash-good':'flash-bad'; return 'flash-neutral'; }
function diffUsageSnapshots(previous, next) { if(!previous)return {values:{},rows:{},records:{}}; var changes={values:{},rows:{},records:{}};Object.keys(next.values).forEach(function(key){if(previous.values[key]!==undefined&&previous.values[key]!==next.values[key])changes.values[key]=usageTone(previous.values[key],next.values[key],key==='failures');});Object.keys(next.rows).forEach(function(key){var before=previous.rows[key], after=next.rows[key];if(before===undefined||after===undefined)return;if(key.indexOf('record:')===0){if(before!==after)changes.records[key]='flash-neutral';return;}['requests','failures','tokens'].forEach(function(field){if(before[field]!==after[field])changes.rows[key+':'+field]=usageTone(before[field],after[field],field==='failures');});});return changes; }
function applyPanelPulse(el) { if(!el||!el.classList)return;if(el._panelPulseTimer)clearTimeout(el._panelPulseTimer);el.classList.remove('panel-pulse');void el.offsetWidth;el.classList.add('panel-pulse');el._panelPulseTimer=setTimeout(function(){el.classList.remove('panel-pulse');el._panelPulseTimer=null;},1000); }
function pulseKPIGroup(el, tone) { if(tone&&el&&el.closest)applyPanelPulse(el.closest('.kpi-group')); }
function flashValue(el, tone, label) { if(!el||!tone)return; var tones=['flash-good','flash-bad','flash-danger','flash-neutral']; if(el._flashTimer)clearTimeout(el._flashTimer); tones.forEach(function(c){el.classList.remove(c);});el.classList.remove('value-flash'); var originalAria=el.getAttribute('aria-label'), originalTitle=el.getAttribute('title'); el.classList.add('value-flash',tone);el.title=label+t('flash.suffix');el.setAttribute('aria-label',(originalAria||el.textContent||label)+t('flash.suffix'));el._flashTimer=setTimeout(function(){el.classList.remove('value-flash');tones.forEach(function(c){el.classList.remove(c);});if(originalTitle===null)el.removeAttribute('title');else el.setAttribute('title',originalTitle);if(originalAria===null)el.removeAttribute('aria-label');else el.setAttribute('aria-label',originalAria);el._flashTimer=null;},1000); }
// computeAdaptivePollMs: determines the next poll delay based on the current
// preset and the freshness of the most recent record.
// - today with record within 2m => 30s
// - today with activity but older => 90s
// - quiet today (no records) => 5m
// - historical presets (yesterday/7d/30d/month/custom) => 10m
function computeAdaptivePollMs(records, preset) {
  var now = Date.now();
  if (state.toRFC3339 && new Date(state.toRFC3339).getTime() < now - 60000) return 0;
  if (preset !== 'today') return UA_POLL_HISTORY_MS;
  if (!records || records.length === 0) return UA_POLL_QUIET_MS;
  var latestMs = 0;
  records.forEach(function(rec) { var rawT = rec.requested_at || rec.t || rec.RequestedAt; var t = rawT ? new Date(rawT).getTime() : 0; if (!isNaN(t) && t > latestMs) latestMs = t; });
  if (!latestMs) return UA_POLL_QUIET_MS;
  return Date.now() - latestMs <= 2 * 60 * 1000 ? UA_POLL_RECENT_MS : UA_POLL_ACTIVE_MS;
}
function scheduleNextPoll(delayMs) {
  if (state.pollTimer) { clearTimeout(state.pollTimer); state.pollTimer = null; }
  if (delayMs <= 0) return; // completed historical range: do not poll
  if (delayMs > UA_POLL_MAX_MS) delayMs = UA_POLL_MAX_MS;
  state.pollTimer = setTimeout(function() {
    state.pollTimer = null;
    if (document.hidden) return; // paused; will reload on visibilitychange
    if (!state.inFlight) loadData(false);
  }, delayMs);
}

function setDashboardBusy(busy) {
  var content = DOM.dashboardContent || document.getElementById('dashboardContent');
  if (content) content.setAttribute('aria-busy', busy ? 'true' : 'false');
}

function renderSkeletons() {
  var skeleton = DOM.usageSkeleton || document.getElementById('usageSkeleton');
  var dataContent = DOM.usageDataContent || document.getElementById('usageDataContent');
  if (skeleton) {
    skeleton.classList.add('is-visible');
    skeleton.setAttribute('aria-hidden', 'false');
  }
  if (dataContent) dataContent.classList.add('is-loading');
  scheduleDailyChartRedraw();
  setDashboardBusy(true);
}

function hideSkeletons() {
  var skeleton = DOM.usageSkeleton || document.getElementById('usageSkeleton');
  var dataContent = DOM.usageDataContent || document.getElementById('usageDataContent');
  if (skeleton) {
    skeleton.classList.remove('is-visible');
    skeleton.setAttribute('aria-hidden', 'true');
  }
  if (dataContent) dataContent.classList.remove('is-loading');
}

function fetchWithTimeout(url, signal) {
  var timedOut = false, timeoutCtrl = new AbortController(), timeout = setTimeout(function() { timedOut = true; timeoutCtrl.abort(); }, UA_FETCH_TIMEOUT_MS), relayAbort = function() { timeoutCtrl.abort(); };
  signal.addEventListener('abort', relayAbort, { once: true });
  return fetch(url, { signal: timeoutCtrl.signal }).then(function(response) { if (!response.ok) return safeErrorMessage(response).then(function(message) { throw new Error(message); }); return response.json(); }).catch(function(err) { if (timedOut) throw new Error(t('err.timeout')); throw err; }).finally(function() { clearTimeout(timeout); signal.removeEventListener('abort', relayAbort); });
}
function loadData(appendRecords, forceReplace) {
  if (!appendRecords && state.recordsExpanded && !forceReplace && state.hasLoadedData && state.snapshotFilterKey === snapshotFilterKey()) { showError(t('err.expanded')); return; }
  if (state.inFlight && state.abortCtrl) state.abortCtrl.abort();
  var requestID = ++state.requestID; state.abortCtrl = new AbortController(); state.inFlight = true;
  if (!appendRecords && !state.hasLoadedData) renderSkeletons(); else setDashboardBusy(true);
  var signal = state.abortCtrl.signal, filterKey = snapshotFilterKey();
  var previous = state.snapshotFilterKey === filterKey ? state.serverSnapshot : null;
  var progressive = !state.hasLoadedData && !appendRecords;
  var summaryUrl = '/v0/resource/plugins/usage-audit/data/summary' + buildQueryString(false);
  var recordsUrl = '/v0/resource/plugins/usage-audit/data/records' + buildQueryString(true, appendRecords ? state.cursor : null);
  function currentRequest() { return requestID === state.requestID && !signal.aborted; }
  // Independent panels paint as soon as their own response arrives: a slow records
  // request must not hold the token KPIs/chart behind Promise.all (or vice versa).
  var summaryReq = appendRecords ? Promise.resolve(null) : fetchWithTimeout(summaryUrl, signal).then(function(summary) {
    if (!currentRequest() || !progressive) return summary;
    state.pendingChanges=diffUsageSnapshots(previous, makeUsageSnapshot(summary, null));
    renderSummary(summary); state.pendingChanges=null; hideSkeletons();
    return summary;
  });
  var recordsReq = fetchWithTimeout(recordsUrl, signal).then(function(records) {
    if (!currentRequest() || !progressive) return records;
    state.pendingChanges=appendRecords ? {values:{},rows:{},records:{}} : diffUsageSnapshots(previous, makeUsageSnapshot(null, records));
    renderRecords(records, appendRecords); state.pendingChanges=null; hideSkeletons();
    return records;
  });
  Promise.all([summaryReq, recordsReq]).then(function(data) {
    if (!currentRequest()) return;
    state.inFlight = false; clearError();
    if (!progressive) {
      state.pendingChanges=appendRecords ? {values:{},rows:{},records:{}} : diffUsageSnapshots(previous, makeUsageSnapshot(data[0], data[1]));
      if (data[0]) renderSummary(data[0]); renderRecords(data[1], appendRecords); state.pendingChanges=null;
    }
    if (!appendRecords) { state.serverSnapshot=makeUsageSnapshot(data[0], data[1]); state.snapshotFilterKey=filterKey; }
    extractFilterOptions(data[0], data[1]); state.hasLoadedData = true; hideSkeletons(); setDashboardBusy(false);
    if (!appendRecords) scheduleNextPoll(computeAdaptivePollMs(state.records, state.currentPreset));
  }).catch(function(err) {
    if (requestID !== state.requestID) return;
    // Prevent the surviving sibling response from painting after this request failed.
    if (state.abortCtrl) state.abortCtrl.abort(); state.pendingChanges=null; state.inFlight = false; setDashboardBusy(false);
    if (err.name !== 'AbortError') { hideSkeletons(); showError(err.message === 'The user aborted a request.' ? t('err.timeout') : t('err.load', { e: err.message })); scheduleNextPoll(UA_POLL_RETRY_MS); }
  });
}
// Render Summary and Aggregations
// NOTE: total_tokens from backend is authoritative.
// cached_tokens may be included in input_tokens or reported separately by a provider.
// reasoning_tokens is provider-reported and may already be in total_tokens.
// NEVER calculate total_tokens from the displayed component fields.
function renderSummary(data) {
  if (!data) return;
  lastSummaryData = data;

  var totalCount = (typeof data.total_count === 'number') ? data.total_count : 0;
  var summaryList = Array.isArray(data.summary) ? data.summary : [];

  // Aggregate once across all summary buckets (avoids repeated reduce)
  var generateCount = 0;
  var inputTokens = 0;
  var freshInputTokens = 0;
  var outputTokens = 0;
  // KPI token values are derived from the summary buckets. The backend does
  // not expose additive top-level token fields; total_tokens is authoritative
  // per bucket and is never reconstructed from input/output/cache/reasoning.
  var totalTokens = 0;
  var reasoningTokens = 0;
  var cachedTokens = 0;
  var failures = 0;

  var providerMap = {};
  var modelMap = {};
  var endpointMap = {};
  var dailyMap = {};

  var hasDaily = Array.isArray(data.daily) && data.daily.length > 0;
  var hasByProvider = Array.isArray(data.by_provider) && data.by_provider.length > 0;
  var hasByEndpoint = Array.isArray(data.by_endpoint) && data.by_endpoint.length > 0;
  var hasByModel = Array.isArray(data.by_model) && data.by_model.length > 0;

  if (hasDaily) {
    data.daily.forEach(function(item) {
      var k = item.key || item.date || item.day || item.time || item.t;
      if (!k) return;
      var dateKey = String(k).slice(0, 10);
      dailyMap[dateKey] = {
        date: dateKey,
        requests: Number(item.requests || item.count || 0),
        total: (typeof item.total_tokens === 'number' ? item.total_tokens : 0),
        input: (typeof item.input_tokens === 'number' ? item.input_tokens : 0),
        output: (typeof item.output_tokens === 'number' ? item.output_tokens : 0),
        failures: (typeof item.failures === 'number' ? item.failures : 0),
        cached: (typeof item.cached_tokens === 'number' ? item.cached_tokens : 0),
        reasoning: (typeof item.reasoning_tokens === 'number' ? item.reasoning_tokens : 0)
      };
    });
  }

  summaryList.forEach(function(item) {
    if (!item) return;
    if (typeof item.generate_count === 'number') generateCount += item.generate_count;
    if (typeof item.requests === 'number' && !totalCount) totalCount += item.requests;
    if (typeof item.failures === 'number') failures += item.failures;
    var inp = item.input_tokens || 0;
    var outp = item.output_tokens || 0;
    inputTokens += inp;
    freshInputTokens += (typeof item.fresh_input_tokens === 'number' ? item.fresh_input_tokens : 0);
    outputTokens += outp;
    // Authoritative total token item fallback
    // Missing stored totals are treated as zero; never infer a total.
    totalTokens += (typeof item.total_tokens === 'number' ? item.total_tokens : 0);
    reasoningTokens += (typeof item.reasoning_tokens === 'number' ? item.reasoning_tokens : 0);
    cachedTokens += (typeof item.cached_tokens === 'number' ? item.cached_tokens : 0);

    var p = item.provider || 'unknown';
    providerMap[p] = (providerMap[p] || { requests: 0, failures: 0, tokens: 0 });
    providerMap[p].requests += (item.requests || 0);
    providerMap[p].failures += (item.failures || 0);
    // total_tokens is authoritative; absent totals are not inferred from input/output.
    providerMap[p].tokens += (typeof item.total_tokens === 'number' ? item.total_tokens : 0);

    var m = item.model || 'unknown';
    modelMap[m] = (modelMap[m] || { requests: 0, failures: 0, tokens: 0 });
    modelMap[m].requests += (item.requests || 0);
    modelMap[m].failures += (item.failures || 0);
    // total_tokens is authoritative; absent totals are not inferred from input/output.
    modelMap[m].tokens += (typeof item.total_tokens === 'number' ? item.total_tokens : 0);

    var timeVal = item.time || item.date || item.day || item.t || item.timestamp;
    if (!hasDaily && timeVal) {
      var dk = String(timeVal).slice(0, 10);
      var daily = dailyMap[dk] || { date: dk, requests: 0, total: 0, input: 0, output: 0, failures: 0, cached: 0, reasoning: 0 };
      daily.requests += (item.requests || 1);
      daily.total += (typeof item.total_tokens === 'number' ? item.total_tokens : 0);
      daily.input += inp;
      daily.output += outp;
      daily.failures += (item.failures || 0);
      daily.cached += (typeof item.cached_tokens === 'number' ? item.cached_tokens : 0);
      daily.reasoning += (typeof item.reasoning_tokens === 'number' ? item.reasoning_tokens : 0);
      dailyMap[dk] = daily;
    }
  });

  // Use cached DOM refs for all KPI updates
  var elTotalCount = DOM.kpiTotalCount;
  elTotalCount.textContent = formatExactNumber(totalCount);
  elTotalCount.setAttribute('title', formatExactNumber(totalCount));
  elTotalCount.setAttribute('aria-label', kvLabel('lbl.requests', formatExactNumber(totalCount)));
  flashValue(elTotalCount, state.pendingChanges && state.pendingChanges.values.requests, t('lbl.requests')); pulseKPIGroup(elTotalCount, state.pendingChanges && state.pendingChanges.values.requests);

  var failRate = (totalCount > 0) ? ((failures / totalCount) * 100).toFixed(1) + '%' : '0.0%';
  DOM.kpiSubFail.textContent = t('kpi.fail', { n: formatExactNumber(failures), r: failRate });
  flashValue(DOM.kpiSubFail, state.pendingChanges && state.pendingChanges.values.failures, t('lbl.failedRequests')); pulseKPIGroup(DOM.kpiSubFail, state.pendingChanges && state.pendingChanges.values.failures);
  DOM.kpiFailureRate.textContent = failRate;
  DOM.kpiSubReq.textContent = t('kpi.gen', { n: formatExactNumber(generateCount) });

  var elTotalTokens = DOM.kpiTotalTokens;
  elTotalTokens.textContent = formatExactNumber(totalTokens);
  elTotalTokens.setAttribute('title', formatExactNumber(totalTokens));
  elTotalTokens.setAttribute('aria-label', kvLabel('lbl.tokens', formatExactNumber(totalTokens)));
  flashValue(elTotalTokens, state.pendingChanges && state.pendingChanges.values.tokens, t('lbl.tokens')); pulseKPIGroup(elTotalTokens, state.pendingChanges && state.pendingChanges.values.tokens);
  if (DOM.kpiSubTokens) DOM.kpiSubTokens.textContent = t('kpi.official');

  var elInputReported = DOM.kpiInputTokens;
  if (elInputReported) {
    elInputReported.textContent = formatExactNumber(inputTokens);
    elInputReported.setAttribute('title', formatExactNumber(inputTokens));
    elInputReported.setAttribute('aria-label', kvLabel('lbl.inputReported', formatExactNumber(inputTokens)));
    flashValue(elInputReported, state.pendingChanges && state.pendingChanges.values.input, t('lbl.inputReported')); pulseKPIGroup(elInputReported, state.pendingChanges && state.pendingChanges.values.input);
  }

  var elInput = DOM.kpiFreshInputTokens;
  if (elInput) {
    elInput.textContent = formatExactNumber(freshInputTokens);
    elInput.setAttribute('title', formatExactNumber(freshInputTokens));
    elInput.setAttribute('aria-label', kvLabel('lbl.fresh', formatExactNumber(freshInputTokens)));
    flashValue(elInput, state.pendingChanges && state.pendingChanges.values.fresh, t('lbl.fresh')); pulseKPIGroup(elInput, state.pendingChanges && state.pendingChanges.values.fresh);
  }

  var elOutput = DOM.kpiOutputTokens;
  if (elOutput) {
    elOutput.textContent = formatExactNumber(outputTokens);
    elOutput.setAttribute('title', formatExactNumber(outputTokens));
    elOutput.setAttribute('aria-label', kvLabel('lbl.output', formatExactNumber(outputTokens)));
    flashValue(elOutput, state.pendingChanges && state.pendingChanges.values.output, t('lbl.output')); pulseKPIGroup(elOutput, state.pendingChanges && state.pendingChanges.values.output);
  }

  var elReason = DOM.kpiReasoningTokens;
  if (elReason) {
    elReason.textContent = formatExactNumber(reasoningTokens);
    elReason.setAttribute('title', formatExactNumber(reasoningTokens));
    elReason.setAttribute('aria-label', kvLabel('lbl.reasoning', formatExactNumber(reasoningTokens)));
    flashValue(elReason, state.pendingChanges && state.pendingChanges.values.reasoning, t('lbl.reasoning')); pulseKPIGroup(elReason, state.pendingChanges && state.pendingChanges.values.reasoning);
  }

  var elCache = DOM.kpiCachedTokens;
  if (elCache) {
    elCache.textContent = formatExactNumber(cachedTokens);
    elCache.setAttribute('title', formatExactNumber(cachedTokens));
    elCache.setAttribute('aria-label', kvLabel('lbl.cache', formatExactNumber(cachedTokens)));
    flashValue(elCache, state.pendingChanges && state.pendingChanges.values.cached, t('lbl.cache')); pulseKPIGroup(elCache, state.pendingChanges && state.pendingChanges.values.cached);
  }

  // Entities compat
  var pCount = hasByProvider ? data.by_provider.length : Object.keys(providerMap).length;
  var mCount = hasByModel ? data.by_model.length : Object.keys(modelMap).length;
  if (DOM.kpiActiveEntities) DOM.kpiActiveEntities.textContent = pCount + ' / ' + mCount;
  if (DOM.kpiSubEntities) DOM.kpiSubEntities.textContent = t('kpi.entities', { p: pCount, m: mCount });

  // Render Daily Trend Chart
  renderChartFromDaily(dailyMap);

  // Render Breakdown Section
  renderBreakdowns({
    hasByProvider: hasByProvider,
    providers: hasByProvider ? data.by_provider : providerMap,
    hasByEndpoint: hasByEndpoint,
    endpoints: hasByEndpoint ? data.by_endpoint : endpointMap,
    hasByModel: hasByModel,
    models: hasByModel ? data.by_model : modelMap
  });
}

function appendText(parent, className, text) { var el = document.createElement('div'); if (className) el.className = className; el.textContent = text; parent.appendChild(el); return el; }
function closeDailyChartTooltip() { var tooltip = DOM.dailyChartTooltip || document.getElementById('dailyChartTooltip'); if (tooltip) tooltip.hidden = true; }
function showDailyChartTooltip(detail, anchor) {
  var tooltip = DOM.dailyChartTooltip || document.getElementById('dailyChartTooltip'); if (!tooltip || !anchor) return;
  tooltip.replaceChildren();
  appendText(tooltip, 'daily-chart-tooltip-title', detail.date);
  appendText(tooltip, '', t('tip.requests') + formatExactNumber(detail.requests)); appendText(tooltip, '', t('tip.total') + formatExactNumber(detail.total));
  appendText(tooltip, '', t('tip.input') + formatExactNumber(detail.input)); appendText(tooltip, '', t('tip.output') + formatExactNumber(detail.output)); appendText(tooltip, '', t('tip.failed') + formatExactNumber(detail.failures));
  appendText(tooltip, 'daily-chart-tooltip-note', t('tip.cache') + formatExactNumber(detail.cached) + ' · ' + t('tip.reasoning') + formatExactNumber(detail.reasoning));
  tooltip.hidden = false; var box = anchor.getBoundingClientRect(), margin = 8;
  var left = Math.max(margin, Math.min(box.left, window.innerWidth - tooltip.offsetWidth - margin)), top = box.top - tooltip.offsetHeight - margin;
  if (top < margin) top = Math.min(window.innerHeight - tooltip.offsetHeight - margin, box.bottom + margin);
  tooltip.style.left = Math.max(margin, left) + 'px'; tooltip.style.top = Math.max(margin, top) + 'px';
}
function dateKeysForSelectedRange(dailyMap) {
  var keys = Object.keys(dailyMap), from = state.fromRFC3339 ? new Date(state.fromRFC3339) : null, to = state.toRFC3339 ? new Date(state.toRFC3339) : null;
  if (!from || !to || isNaN(from) || isNaN(to)) return keys.sort();
  // RFC3339 bounds represent local-day selections. Use local calendar fields so
  // positive/negative timezone offsets do not add an adjacent empty UTC day.
  var current = new Date(from.getFullYear(), from.getMonth(), from.getDate()), last = new Date(to.getFullYear(), to.getMonth(), to.getDate()), result = [];
  while (current <= last) {
    var key = current.getFullYear() + '-' + String(current.getMonth() + 1).padStart(2, '0') + '-' + String(current.getDate()).padStart(2, '0');
    result.push(key);
    if (!dailyMap[key]) dailyMap[key] = { date:key, requests:0, total:0, input:0, output:0, failures:0, cached:0, reasoning:0 };
    current.setDate(current.getDate() + 1);
  }
  return result;
}
var lastDailyChartData = null;
var dailyChartRedrawFrame = null;
function getDailyChartGeometry(days, width) {
  var count = Math.max(1, Math.floor(Number(days) || 1));
  var availableWidth = Number(width);
  if (!Number.isFinite(availableWidth) || availableWidth <= 0) availableWidth = 280;
  var svgWidth = count <= 7 ? availableWidth : Math.max(availableWidth, count * 40 + 36);
  var inset = Math.min(18, svgWidth / 4), slot = (svgWidth - inset * 2) / count;
  var barWidth = Math.min(36, Math.max(16, slot * 0.55));
  var labelStride = Math.max(1, Math.ceil(count / (svgWidth / 60)));
  var centers = [], labelIndices = [];
  for (var k = 0; k < count; k++) {
    centers.push(inset + (k + 0.5) * slot);
    // Keep endpoint labels, but omit the last intermediate label if it would collide.
    if (k === 0 || k === count - 1 || (k % labelStride === 0 && (count - 1 - k) * slot >= 60)) labelIndices.push(k);
  }
  return { width: svgWidth, slot: slot, barWidth: barWidth, labelStride: labelStride, centers: centers, labelIndices: labelIndices };
}
function scheduleDailyChartRedraw() {
  if (dailyChartRedrawFrame !== null) cancelAnimationFrame(dailyChartRedrawFrame);
  dailyChartRedrawFrame = requestAnimationFrame(function() {
    dailyChartRedrawFrame = null;
    var card = DOM.chartCard || document.getElementById('chartCard');
    if (card && card.open && lastDailyChartData) renderChartFromDaily(lastDailyChartData);
  });
}
function renderChartFromDaily(dailyMap) {
  var chartCard = DOM.chartCard || document.getElementById('chartCard'), chartArea = DOM.chartArea || document.getElementById('chartArea'), tooltip = DOM.dailyChartTooltip || document.getElementById('dailyChartTooltip');
  lastDailyChartData = dailyMap;
  if (dailyChartRedrawFrame !== null) { cancelAnimationFrame(dailyChartRedrawFrame); dailyChartRedrawFrame = null; }
  var focusedDay = chartArea.contains(document.activeElement) ? document.activeElement.getAttribute('data-day') : null, focusedBar = null;
  chartArea.replaceChildren(); closeDailyChartTooltip(); var days = dateKeysForSelectedRange(dailyMap);
  var summary = DOM.dailyChartSummary || document.getElementById('dailyChartSummary');
  var maxVal = 0, totalRequests = 0;
  days.forEach(function(day) { var requests = Number(dailyMap[day].requests) || 0; maxVal = Math.max(maxVal, requests); totalRequests += requests; });
  if (summary) summary.textContent = totalRequests ? t('chart.summary', { d: days.length, r: formatExactNumber(totalRequests), m: formatExactNumber(maxVal) }) : t('chart.empty');
  chartCard.classList.remove('is-hidden');
  if (!maxVal) { appendText(chartArea, 'chart-empty-state', t('chart.empty')); return; }
  var geometry = getDailyChartGeometry(days.length, chartCard.clientWidth > 24 ? chartCard.clientWidth - 24 : 280);
  var svgWidth = geometry.width, svgHeight = 135, barWidth = geometry.barWidth, chartBottom = 110, chartTop = 15;
  var svg = document.createElementNS('http://www.w3.org/2000/svg', 'svg'); svg.setAttribute('width', String(svgWidth)); svg.setAttribute('height', String(svgHeight)); svg.setAttribute('viewBox', '0 0 ' + svgWidth + ' ' + svgHeight); svg.setAttribute('role', 'img'); svg.setAttribute('aria-label', t('chart.aria')); svg.setAttribute('aria-describedby', 'dailyChartTooltip');
  var line = document.createElementNS('http://www.w3.org/2000/svg', 'line'); line.setAttribute('x1', '10'); line.setAttribute('x2', String(svgWidth - 10)); line.setAttribute('y1', String(chartBottom)); line.setAttribute('y2', String(chartBottom)); line.setAttribute('stroke', '#e5e7eb'); line.setAttribute('stroke-width', '1'); svg.appendChild(line);
  days.forEach(function(day, k) {
    var detail = dailyMap[day], barH = Math.max(6, Math.round(((Number(detail.requests) || 0) / maxVal) * (chartBottom - chartTop)));
    var x = geometry.centers[k] - barWidth / 2, y = chartBottom - barH, rect = document.createElementNS('http://www.w3.org/2000/svg', 'rect');
    rect.setAttribute('x',String(x));rect.setAttribute('y',String(y));rect.setAttribute('width',String(barWidth));rect.setAttribute('height',String(barH));rect.setAttribute('fill',detail.requests ? 'var(--accent-blue, #3b6593)' : '#cbd5e1');rect.setAttribute('rx','3');rect.setAttribute('class','daily-chart-bar');rect.setAttribute('tabindex','0');rect.setAttribute('role','button');rect.setAttribute('data-day',day);rect.setAttribute('aria-describedby','dailyChartTooltip');rect.setAttribute('aria-label',t('chart.bar',{date:detail.date,n:formatExactNumber(detail.requests)}));
    if (day === focusedDay) focusedBar = rect;
    rect.addEventListener('pointerenter',function(){showDailyChartTooltip(detail,rect);});rect.addEventListener('pointerleave',closeDailyChartTooltip);rect.addEventListener('focus',function(){showDailyChartTooltip(detail,rect);});rect.addEventListener('click',function(e){e.stopPropagation();showDailyChartTooltip(detail,rect);});rect.addEventListener('blur',closeDailyChartTooltip);rect.addEventListener('keydown',function(e){if(e.key === 'Enter'||e.key === ' '){e.preventDefault();showDailyChartTooltip(detail,rect);}if(e.key === 'Escape'){e.preventDefault();closeDailyChartTooltip();rect.blur();}}); svg.appendChild(rect);
    if (geometry.slot >= 64) {
      var value=document.createElementNS('http://www.w3.org/2000/svg','text');value.setAttribute('x',String(geometry.centers[k]));value.setAttribute('y',String(y-4));value.setAttribute('text-anchor','middle');value.setAttribute('font-size','10');value.setAttribute('font-variant-numeric','tabular-nums');value.setAttribute('fill','#6b7280');value.textContent=formatExactNumber(detail.requests);svg.appendChild(value);
    }
    if (geometry.labelIndices.indexOf(k) !== -1) {
      var label=document.createElementNS('http://www.w3.org/2000/svg','text');label.setAttribute('x',String(geometry.centers[k]));label.setAttribute('y',String(chartBottom+14));label.setAttribute('text-anchor','middle');label.setAttribute('font-size','10');label.setAttribute('fill','#6b7280');label.textContent=detail.date.length>=5?detail.date.slice(5):detail.date;svg.appendChild(label);
    }
  });
  chartArea.appendChild(svg);
  if (focusedBar) focusedBar.focus({ preventScroll: true });
  closeDailyChartTooltip(); if (tooltip) tooltip.hidden = true;
}
function renderBreakdowns(data) {
  var section = DOM.breakdownSection || document.getElementById('breakdownSection'); section.classList.remove('is-hidden');
  function buildRows(source, isArray, nameKey, category) {
    var list = []; if (isArray) list = source.map(function(item) { item=item||{}; var rawName = String(item[nameKey] || item.key || item.name || item.endpoint || item.source || item.executor_type || '-'); return { rawKey:category+':'+rawName, name:category==='provider' ? formatProvider(rawName) : rawName, requests:item.requests || item.count || 0, failures:item.failures || 0, tokens:typeof item.total_tokens === 'number' ? item.total_tokens : 0 }; });
    else Object.keys(source).forEach(function(key) { list.push({ rawKey:category+':'+String(key), name:category==='provider' ? formatProvider(key) : key, requests:source[key].requests, failures:source[key].failures, tokens:source[key].tokens }); });
    return list.sort(function(a,b) { return b.requests-a.requests; });
  }
  function toTableRows(rows) {
    return rows.slice(0,10).map(function(r) {
      var changes=state.pendingChanges&&state.pendingChanges.rows;
      function cell(text,className,field,label) { var key=r.rawKey+':'+field;return {text:text,className:className,title:field?null:String(r.name),changeKey:field?key:null,tone:field&&changes&&changes[key]||null,label:label}; }
      return {key:r.rawKey,cells:[cell(String(r.name),'cell-ellipsis',null),cell(formatExactNumber(r.requests),'text-right','requests',t('th.requests')),cell(formatExactNumber(r.failures),'text-right','failures',t('th.failed')),cell(formatExactNumber(r.tokens),'text-right','tokens',t('th.tokens'))]};
    });
  }
  var tableUpdates=[['providerAggTableMount',toTableRows(buildRows(data.providers,data.hasByProvider,'provider','provider')),t('empty.data')],['endpointAggTableMount',toTableRows(buildRows(data.endpoints,data.hasByEndpoint,'endpoint','endpoint')),t('empty.data')],['modelAggTableMount',toTableRows(buildRows(data.models,data.hasByModel,'model','model')),t('empty.data')]];
  if(window.renderDohwiTable)tableUpdates.forEach(function(update){window.renderDohwiTable(update[0],update[1],update[2]);});else window.pendingDohwiTableUpdates=(window.pendingDohwiTableUpdates||[]).concat(tableUpdates);
  if(state.pendingChanges&&Object.keys(state.pendingChanges.rows||{}).length)applyPanelPulse(section);
}
// Render Records & Pagination (20 records limit, dohwi-ui Table primitives)
// legacy safe text contract: function appendRecordCell; cell.textContent=text; cell.title=title
function renderRecords(data, append) {
  var countBadge=DOM.recordsCountBadge || document.getElementById('recordsCountBadge'), loadMoreBar=DOM.loadMoreBar || document.getElementById('loadMoreBar'), recordsList=[];
  if(Array.isArray(data)){recordsList=data;state.hasMore=false;state.cursor=null;}else if(data&&Array.isArray(data.records)){recordsList=data.records;state.hasMore=!!data.has_more;state.cursor=data.next_cursor||null;}
  if(append){state.records=state.records.concat(recordsList);state.recordsExpanded=true;}else{state.records=recordsList;state.recordsExpanded=false;}
  paintRecords();
}
// Builds the table from state.records; reused when the language changes.
function paintRecords() {
  var countBadge=DOM.recordsCountBadge || document.getElementById('recordsCountBadge'), loadMoreBar=DOM.loadMoreBar || document.getElementById('loadMoreBar');
  countBadge.textContent=tn('records.count', state.records.length);
  var rows=state.records.map(function(rec,idx) {
    var rawT=rec.requested_at||rec.t||rec.RequestedAt||rec.Timestamp, rawProv=rec.provider||rec.Provider||'-', displayProv=formatProvider(rawProv), model=rec.model||rec.Model||rec.alias||rec.Alias||'-', endpoint=rec.endpoint||rec.executor_type||rec.ExecutorType||rec.source||rec.Source||'-', failed=!!(rec.failed||rec.Failed), recordKey='record:'+String(rec.id||rec.request_id||rawT||idx), changes=state.pendingChanges&&state.pendingChanges.records;
    var parsedTime=new Date(rawT), mobileTime=Number.isNaN(parsedTime.getTime())?'—':parsedTime.toLocaleTimeString(uiLocale(),{hour:'2-digit',minute:'2-digit',hour12:false}); // mobileTime.textContent, timeCell.title=formatDateTime(rawT)
    // Compose semantic UI Table cells without HTML string concatenation (legacy: mobileTime.textContent). // mobileTime.textContent, timeCell.title=formatDateTime(rawT)
    var statusCell={badge:{text:failed?t('rec.failed'):t('rec.success'),className:'badge '+(failed?'badge-fail':'badge-success')},tone:changes&&changes[recordKey]||null,label:t('rec.status')};
    return {key:String(rec.id||rec.request_id||rawT||idx),dataIndex:idx,tabIndex:0,onClick:function(e){showRecordDetail(idx,e.currentTarget);},onKeyDown:function(e){if(e.key==='Enter'||e.key===' '){e.preventDefault();showRecordDetail(idx,e.currentTarget);}},cells:[
      {time:{full:formatDateTime(rawT),mobile:mobileTime},title:formatDateTime(rawT)},
      {text:displayProv,className:'cell-ellipsis',title:displayProv},
      {text:String(endpoint),className:'cell-ellipsis',title:String(endpoint)},
      {text:String(model),className:'cell-ellipsis',title:String(model)},
      {text:formatExactNumber(rec.fresh_input != null ? rec.fresh_input : null),className:'num-cell num-input'},
      {text:formatExactNumber(rec.cached),className:'num-cell num-input'},
      {text:formatExactNumber(rec.output != null ? rec.output : null),className:'num-cell num-output'},
      {text:formatExactNumber(rec.total != null ? rec.total : null),className:'num-cell num-total'},
      {text:formatLatency(rec.latency_ms != null ? rec.latency_ms : (rec.latency != null ? rec.latency : rec.Latency)),className:'num-cell'},
      statusCell
    ]};
  });
  if(window.renderDohwiTable)window.renderDohwiTable('recordsTableMount',rows,t('empty.records'));else window.pendingDohwiTableUpdates=(window.pendingDohwiTableUpdates||[]).concat([['recordsTableMount',rows,t('empty.records')]]);
  var renderedRows=document.querySelectorAll('#recordsBody tr.ui-table-row');
  renderedRows.forEach(function(row,index){row.dataset.index=String(index);row.dataset.rowKey=rows[index]?rows[index].key:'';var rec=state.records[index];if(rec&&state.pendingChanges&&state.pendingChanges.records['record:'+String(rec.id||rec.request_id||rec.requested_at||index)])applyPanelPulse(document.getElementById('recordsPanel'));});
  loadMoreBar.classList.toggle('is-hidden', !state.hasMore);
}
function loadMoreRecords() {
  if (state.inFlight) return;
  if (state.snapshotFilterKey !== snapshotFilterKey()) { loadData(false, true); return; }
  if (!state.hasMore || !state.cursor) return;
  loadData(true);
}
// Record detail modal: focus-safe dialog/sheet with reversible body scroll lock.
function detailRow(key, value, className) { var row=document.createElement('div'), keyEl=document.createElement('div'), valueEl=document.createElement('div');row.className='detail-row';keyEl.className='detail-key';keyEl.textContent=key;valueEl.className='detail-val'+(className?' '+className:'');valueEl.textContent=value;row.appendChild(keyEl);row.appendChild(valueEl);return row; }
function detailSection(title, rows) { var section=document.createElement('section'), heading=document.createElement('div');section.className='detail-section';heading.className='detail-section-title';heading.textContent=title;section.appendChild(heading);rows.forEach(function(row){section.appendChild(row);});return section; }
function lockBodyScroll() { if (state.dialogScrollLock) return; state.dialogScrollLock={overflow:document.body.style.overflow,paddingRight:document.body.style.paddingRight}; var gutter=typeof getComputedStyle==='function'?getComputedStyle(document.documentElement).scrollbarGutter:'';var width=gutter&&gutter.indexOf('stable')!==-1?0:window.innerWidth-document.documentElement.clientWidth; document.body.style.overflow='hidden'; if(width>0)document.body.style.paddingRight=width+'px'; }
function unlockBodyScroll() { if(!state.dialogScrollLock)return; document.body.style.overflow=state.dialogScrollLock.overflow;document.body.style.paddingRight=state.dialogScrollLock.paddingRight;state.dialogScrollLock=null; }
function showRecordDetail(idx, opener) {
  var rec=state.records[idx];if(!rec)return;state.openRecordIdx=idx;var dialog=DOM.recordDialog||document.getElementById('recordDialog'),overlay=DOM.detailOverlay||document.getElementById('detailOverlay'),body=DOM.dialogBody||document.getElementById('dialogBody'),rawT=rec.requested_at||rec.t||rec.RequestedAt||rec.Timestamp,rawProv=rec.provider||rec.Provider||'-',displayProv=formatProvider(rawProv),model=rec.model||rec.Model||'-',alias=rec.alias||rec.Alias||'-',endpoint=rec.endpoint||rec.executor_type||rec.ExecutorType||rec.source||rec.Source||'-',failed=!!(rec.failed||rec.Failed),latency=formatLatency(rec.latency_ms != null?rec.latency_ms:(rec.latency != null?rec.latency:rec.Latency)),ttft=formatLatency(rec.ttft_ms != null?rec.ttft_ms:(rec.ttft != null?rec.ttft:rec.TTFT));
  if(!dialog||!overlay||!body)return;state.drawerOpener=opener||document.activeElement;body.replaceChildren();var grid=document.createElement('div');grid.className='detail-grid';grid.appendChild(detailSection(t('dlg.basic'),[detailRow(t('dlg.time'),formatDateTime(rawT)+' ('+String(rawT||'-')+')'),detailRow(t('dlg.status'),failed?t('rec.failed'):t('rec.success')),detailRow(t('dlg.failureStatus'),String(rec.failure_status||'-')),detailRow('Provider',displayProv+(rawProv!==displayProv?' ('+String(rawProv)+')':'')),detailRow('Model',String(model)),detailRow('Alias',String(alias)),detailRow('Endpoint / Source',String(endpoint))]));grid.appendChild(detailSection(t('dlg.tokens'),[detailRow(t('dlg.total'),formatExactNumber(rec.total != null?rec.total:null)),detailRow(t('dlg.input'),formatExactNumber(rec.input != null?rec.input:null),'num-input'),detailRow(t('dlg.fresh'),formatExactNumber(rec.fresh_input != null?rec.fresh_input:null),'num-input'),detailRow(t('dlg.output'),formatExactNumber(rec.output != null?rec.output:null),'num-output'),detailRow(t('dlg.reasoning'),formatExactNumber(rec.reasoning != null?rec.reasoning:0)),detailRow(t('dlg.cached'),formatExactNumber(rec.cached != null?rec.cached:0)),detailRow(t('dlg.latency'),latency),detailRow(t('dlg.ttft'),ttft)]));body.appendChild(grid);lockBodyScroll();overlay.classList.add('open');overlay.setAttribute('aria-hidden','false');dialog.setAttribute('aria-hidden','false');var close=DOM.recordDialogClose||dialog.querySelector('.dialog-close');if(close)close.focus();
}
function closeDrawer() { var dialog=DOM.recordDialog||document.getElementById('recordDialog'),overlay=DOM.detailOverlay||document.getElementById('detailOverlay');var wasOpen=!!(overlay&&overlay.classList.contains('open'));if(dialog)dialog.setAttribute('aria-hidden','true');if(overlay){overlay.classList.remove('open');overlay.setAttribute('aria-hidden','true');}unlockBodyScroll();var opener=state.drawerOpener;state.drawerOpener=null;state.openRecordIdx=null;if(wasOpen&&opener&&document.contains(opener))opener.focus(); }
function trapDrawerFocus(e) { var dialog=DOM.recordDialog||document.getElementById('recordDialog'),overlay=DOM.detailOverlay||document.getElementById('detailOverlay');if(e.key!=='Tab'||!dialog||!overlay||!overlay.classList.contains('open'))return;var items=dialog.querySelectorAll('button,[href],input,select,textarea,[tabindex]:not([tabindex="-1"])');if(!items.length){e.preventDefault();return;}var first=items[0],last=items[items.length-1];if(e.shiftKey&&document.activeElement===first){e.preventDefault();last.focus();}else if(!e.shiftKey&&document.activeElement===last){e.preventDefault();first.focus();} }

// Visibility change handler: pause on hidden, immediate refresh + resume on visible
document.addEventListener('visibilitychange', function() {
  if (document.hidden) {
    // Cancel pending poll timer; will refresh on visible.
    if (state.pollTimer) { clearTimeout(state.pollTimer); state.pollTimer = null; }
  } else {
    // Refresh immediately on becoming visible.
    if (!state.inFlight) loadData(false);
  }
});

// Window unload cleanup
window.addEventListener('beforeunload', function() {
  if (state.pollTimer) { clearTimeout(state.pollTimer); state.pollTimer = null; }
  if (state.abortCtrl) {
    state.abortCtrl.abort();
    state.abortCtrl = null;
  }
});

function initDashboard() {
  // Cache frequently accessed DOM refs once (reused every refresh)
  DOM.errorBox = document.getElementById('errorBox');
  DOM.errorMessage = document.getElementById('errorMessage');
  DOM.retryLoadBtn = document.getElementById('retryLoadBtn');
  DOM.dashboardContent = document.getElementById('dashboardContent');
  DOM.usageSkeleton = document.getElementById('usageSkeleton');
  DOM.usageDataContent = document.getElementById('usageDataContent');
  DOM.usageMainContent = document.getElementById('usageMainContent');
  DOM.activeFilterBadge = document.getElementById('activeFilterBadge');
  DOM.activeFilterText = document.getElementById('activeFilterText');
  DOM.customDateWrap = document.getElementById('customDateWrap');
  DOM.filterFromDate = document.getElementById('filterFromDate'); DOM.filterToDate = document.getElementById('filterToDate');
  DOM.filterProvider = document.getElementById('filterProvider'); DOM.filterEndpoint = document.getElementById('filterEndpoint'); DOM.filterModel = document.getElementById('filterModel'); DOM.filterStatus = document.getElementById('filterStatus');
  DOM.kpiTotalCount = document.getElementById('kpiTotalCount');
  DOM.kpiSubFail = document.getElementById('kpiSubFail');
  DOM.kpiFailureRate = document.getElementById('kpiFailureRate');
  DOM.kpiSubReq = document.getElementById('kpiSubReq');
  DOM.kpiTotalTokens = document.getElementById('kpiTotalTokens');
  DOM.kpiSubTokens = document.getElementById('kpiSubTokens');
  DOM.kpiInputTokens = document.getElementById('kpiInputTokens');
  DOM.kpiFreshInputTokens = document.getElementById('kpiFreshInputTokens');
  DOM.kpiCachedTokens = document.getElementById('kpiCachedTokens');
  DOM.kpiOutputTokens = document.getElementById('kpiOutputTokens');
  DOM.kpiReasoningTokens = document.getElementById('kpiReasoningTokens');
  DOM.kpiActiveEntities = document.getElementById('kpiActiveEntities');
  DOM.kpiSubEntities = document.getElementById('kpiSubEntities');
  DOM.chartCard = document.getElementById('chartCard');
  DOM.chartArea = document.getElementById('chartArea');
  DOM.dailyChartTooltip = document.getElementById('dailyChartTooltip');
  DOM.dailyChartSummary = document.getElementById('dailyChartSummary');
  DOM.breakdownSection = document.getElementById('breakdownSection');
  DOM.recordsBody = document.getElementById('recordsBody');
  DOM.recordsCountBadge = document.getElementById('recordsCountBadge');
  DOM.loadMoreBar = document.getElementById('loadMoreBar');
  DOM.recordDialog = document.getElementById('recordDialog'); DOM.detailOverlay = document.getElementById('detailOverlay'); DOM.dialogBody = document.getElementById('dialogBody'); DOM.recordDialogClose = document.getElementById('recordDialogClose');
  if (DOM.retryLoadBtn) DOM.retryLoadBtn.addEventListener('click', function() { state.recordsExpanded = false; loadData(false, true); });
  if (DOM.recordDialogClose) DOM.recordDialogClose.addEventListener('click', closeDrawer);
  if (DOM.detailOverlay) DOM.detailOverlay.addEventListener('click', function(e) { if (e.target === DOM.detailOverlay) closeDrawer(); });

  closeDrawer();
  var bounds = computePresetDates('today');
  if (bounds) {
    state.fromRFC3339 = bounds.from;
    state.toRFC3339 = bounds.to;
  }
  syncPeriodInputs();
  updateActiveFilterBadge();
  loadData(false);
}

function bindDashboardEvents() {
  if (typeof window !== 'undefined' && window.addEventListener) window.addEventListener('resize', scheduleDailyChartRedraw);
  if (document.fonts && document.fonts.ready) document.fonts.ready.then(scheduleDailyChartRedraw);
  var chartCard = document.getElementById('chartCard');
  if (chartCard) {
    if (typeof ResizeObserver !== 'undefined') {
      var lastChartWidth = chartCard.clientWidth;
      var chartResizeObserver = new ResizeObserver(function() {
        var width = chartCard.clientWidth;
        if (width !== lastChartWidth) { lastChartWidth = width; scheduleDailyChartRedraw(); }
      });
      chartResizeObserver.observe(chartCard);
    }
  }
  // Fallback buttons are used only if the exported React components did not mount.
  var pBtns = window.renderDashboardPresets ? [] : document.querySelectorAll('#presetGroup .preset-btn');
  pBtns.forEach(function(btn, index) {
    btn.addEventListener('click', function() {
      setPreset(btn.getAttribute('data-preset'));
    });
    // Radix RadioGroup keyboard model: arrows move focus and selection, Home/End jump.
    btn.addEventListener('keydown', function(e) {
      var list = Array.prototype.slice.call(pBtns), at = list.indexOf(btn), next = -1;
      if (e.key === 'ArrowRight' || e.key === 'ArrowDown') next = (at + 1) % list.length;
      else if (e.key === 'ArrowLeft' || e.key === 'ArrowUp') next = (at + list.length - 1) % list.length;
      else if (e.key === 'Home') next = 0; else if (e.key === 'End') next = list.length - 1;
      if (next < 0) return;
      e.preventDefault(); list[next].focus(); setPreset(list[next].getAttribute('data-preset'));
    });
  });

  // Escape and outside tap close transient overlays.
  document.addEventListener('keydown', function(e) {
    if (e.key === 'Escape') { closeDrawer(); closeDailyChartTooltip(); } trapDrawerFocus(e);
  });
  document.addEventListener('pointerdown', function(e) {
    var area = DOM.chartArea || document.getElementById('chartArea');
    if (!area || !area.contains(e.target)) closeDailyChartTooltip();
  });

  document.querySelectorAll('.tab-btn').forEach(function(tab) { tab.addEventListener('keydown', function(e) { if (e.key !== 'ArrowLeft' && e.key !== 'ArrowRight' && e.key !== 'Home' && e.key !== 'End') return; e.preventDefault(); var tabs=Array.prototype.slice.call(document.querySelectorAll('.tab-btn')), index=tabs.indexOf(tab); if(e.key==='ArrowLeft')index=(index+tabs.length-1)%tabs.length; else if(e.key==='ArrowRight')index=(index+1)%tabs.length; else if(e.key==='Home')index=0; else index=tabs.length-1; switchTab(tabs[index].getAttribute('aria-controls'),tabs[index]); }); });
  var filterForm = document.getElementById('filterForm');
  if (filterForm) filterForm.addEventListener('submit', function(e) { e.preventDefault(); applyFilters(); });
  var resetFiltersBtn = document.getElementById('resetFiltersBtn');
  if (resetFiltersBtn) resetFiltersBtn.addEventListener('click', resetFilters);

  [DOM.filterProvider || document.getElementById('filterProvider'), DOM.filterEndpoint || document.getElementById('filterEndpoint'), DOM.filterModel || document.getElementById('filterModel'), DOM.filterStatus || document.getElementById('filterStatus')].forEach(function(field) {
    if (field) {
      updateSelectTitle(field);
      field.addEventListener('change', function() { updateSelectTitle(field); });
      field.addEventListener('change', applyFilters);
    }
  });

  function applyCustomDatesWhenReady() {
    if (state.currentPreset !== 'custom') return;
    var from = DOM.filterFromDate || document.getElementById('filterFromDate');
    var to = DOM.filterToDate || document.getElementById('filterToDate');
    if (!from.value || !to.value || from.value > to.value) {
      // Editing one endpoint is an unfinished draft, not an error to interrupt.
      setFilterValidation('');
      return;
    }
    applyFilters();
  }
  [DOM.filterFromDate || document.getElementById('filterFromDate'), DOM.filterToDate || document.getElementById('filterToDate')].forEach(function(field) {
    if (!field) return;
    field.addEventListener('change', applyCustomDatesWhenReady);
    field.addEventListener('keydown', function(e) { if (e.key === 'Enter') { e.preventDefault(); applyFilters(); } });
  });
}
// The static DOM is already present: bind interactions before starting requests too.
initLanguageControls();
bindDashboardEvents();
// Start API requests before parsing/mounting the embedded React bundle. The DOM above is
// already present; waiting for DOMContentLoaded would serialize network behind that work.
initDashboard();
