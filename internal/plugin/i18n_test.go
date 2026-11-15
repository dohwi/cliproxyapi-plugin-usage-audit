package plugin_test

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/dohwi/cliproxyapi-plugin-usage-audit/internal/plugin"
)

// runDashboardNode evaluates script with the dashboard HTML available as `html`.
func runDashboardNode(t *testing.T, script string) {
	t.Helper()
	nodePath, err := exec.LookPath("node")
	if err != nil {
		t.Skip("node not available")
	}
	dir := t.TempDir()
	file := filepath.Join(dir, "dashboard.html")
	if err := os.WriteFile(file, []byte(plugin.DashboardHTML), 0o600); err != nil {
		t.Fatal(err)
	}
	full := "const fs=require('fs');const html=fs.readFileSync(process.argv[1],'utf8');\n" + script
	if out, err := exec.Command(nodePath, "-e", full, file).CombinedOutput(); err != nil {
		t.Fatalf("node check failed: %v\n%s", err, out)
	}
}

// The dashboard persists the language preference (LANG_PREF_KEY, shared by every dohwi plugin page)
// and the theme mode (THEME_KEY) in localStorage and, for the current tab only, the management key
// (sessionStorage, KEY_STORAGE). It may read, never write, the host's own entries (auth key, theme,
// language). The management key must never be written to localStorage. The keys of the previous
// "last action wins" model (cpa_plugin_lang, cpa_plugin_lang_host) are gone.
func TestDashboardStorageContract(t *testing.T) {
	html := plugin.DashboardHTML
	if !strings.Contains(html, "var LANG_PREF_KEY = 'cpa_plugin_language';") || !strings.Contains(html, "var HOST_LANG_KEY = 'cli-proxy-language';") {
		t.Fatal("shared language preference key or host language key is missing or renamed")
	}
	if !strings.Contains(html, "var THEME_KEY = 'cpa_plugin_theme';") || !strings.Contains(html, "var HOST_THEME_KEY = 'cli-proxy-theme';") {
		t.Fatal("theme preference key or host theme key is missing or renamed")
	}
	if !strings.Contains(html, "var KEY_STORAGE = 'cpa_usage_audit_mgmt_key';") {
		t.Fatal("session key storage name is missing or renamed")
	}
	// The vendored React bundle is scanned separately: it must not touch storage at all.
	script := regexp.MustCompile(`(?s)<script>(.*?)</script>`).FindStringSubmatch(html)
	if script == nil {
		t.Fatal("main script not found")
	}
	for _, old := range []string{"'cpa_plugin_lang'", "'cpa_plugin_lang_host'", "LANG_HOST_KEY", "var LANG_KEY"} {
		if strings.Contains(script[1], old) {
			t.Fatalf("old language key %q must no longer be read or written", old)
		}
	}
	total := strings.Count(script[1], "localStorage")
	// Exactly five accessors: the language core's lsGet/lsSet wrappers, the theme mode read and
	// write, and readHostStore(name) for the host's auth/theme entries.
	allowed := regexp.MustCompile(`localStorage\.(?:getItem\((?:k|THEME_KEY|name)\)|setItem\((?:k, v|THEME_KEY, mode)\))`).FindAllString(script[1], -1)
	if total != len(allowed) || total != 5 {
		t.Fatalf("localStorage references = %d (allowed accessors = %d, want exactly 5)", total, len(allowed))
	}
	// lsSet is only ever called for the plugin's own preference; lsGet only for it and the host language.
	sets := regexp.MustCompile(`lsSet\(([A-Z_a-z]+)`).FindAllStringSubmatch(script[1], -1)
	if len(sets) != 2 || sets[0][1] != "k" || sets[1][1] != "LANG_PREF_KEY" {
		t.Fatalf("lsSet calls = %v, want only the definition and lsSet(LANG_PREF_KEY", sets)
	}
	for _, g := range regexp.MustCompile(`lsGet\(([A-Z_a-z]+)`).FindAllStringSubmatch(script[1], -1) {
		if g[1] != "k" && g[1] != "LANG_PREF_KEY" && g[1] != "HOST_LANG_KEY" {
			t.Fatalf("unexpected lsGet(%s)", g[1])
		}
	}
	if strings.Contains(script[1], "removeItem(HOST_LANG_KEY") || strings.Contains(script[1], "setItem(HOST_LANG_KEY") || strings.Contains(script[1], "'cli-proxy-language')") {
		t.Fatal("the host language entry must be read only")
	}
	sessionTotal := strings.Count(script[1], "sessionStorage")
	sessionAllowed := regexp.MustCompile(`sessionStorage\.(?:getItem\(KEY_STORAGE\)|setItem\(KEY_STORAGE, key\)|removeItem\(KEY_STORAGE\))`).FindAllString(script[1], -1)
	if sessionTotal != len(sessionAllowed) || sessionTotal != 3 {
		t.Fatalf("sessionStorage references = %d (allowed KEY_STORAGE accessors = %d, want exactly 3)", sessionTotal, len(sessionAllowed))
	}
	bundle := html[strings.Index(html, `<script data-dohwi-ui="exported">`):]
	if strings.Contains(bundle, "localStorage") || strings.Contains(bundle, "sessionStorage") {
		t.Fatal("bundled UI code must not use web storage")
	}
}

// Every translation key must exist in en, ko and zh-CN with identical placeholders, every key used by
// the page, markup or bundle must exist in all three, and every server error code (apierror.go) has an
// err.<code> translation in all three.
func TestDashboardI18nCatalogsAreComplete(t *testing.T) {
	src, err := os.ReadFile("apierror.go")
	if err != nil {
		t.Fatal(err)
	}
	block := regexp.MustCompile(`(?s)// Stable error codes.*?\n\)`).FindString(string(src))
	var codes []string
	for _, m := range regexp.MustCompile(`=\s*"([a-z_]+)"`).FindAllStringSubmatch(block, -1) {
		codes = append(codes, m[1])
	}
	if len(codes) < 17 {
		t.Fatalf("error codes not found in apierror.go: %v", codes)
	}
	runDashboardNode(t, "const codes = "+jsonList(codes)+";\n"+`
const main = html.match(/<script>([\s\S]*?)<\/script>/)[1];
const bundle = html.slice(html.indexOf('<script data-dohwi-ui="exported">'));
const start = main.indexOf('var I18N = {'), end = main.indexOf('var lastSummaryData');
const I18N = new Function(main.slice(start, end) + '; return I18N;')();
const fail = [];
const langs = ['en', 'ko', 'zh-CN'];
if (Object.keys(I18N).sort().join() !== langs.slice().sort().join()) fail.push('catalogs: ' + Object.keys(I18N));
const en = Object.keys(I18N.en);
for (const k of en) {
  // ko may lack an English-only plural '.one' variant; zh-CN must have every en key.
  if (!(k in I18N.ko) && !(k.endsWith('.one') && k.slice(0, -4) in I18N.ko)) fail.push('missing in ko: ' + k);
  if (!(k in I18N['zh-CN'])) fail.push('missing in zh-CN: ' + k);
}
for (const l of ['ko', 'zh-CN']) for (const k of Object.keys(I18N[l])) if (!(k in I18N.en)) fail.push('extra in ' + l + ': ' + k);
const ph = s => (s.match(/\{\w+\}/g) || []).sort().join();
for (const l of ['ko', 'zh-CN']) for (const k of Object.keys(I18N[l])) if (k in I18N.en && !k.endsWith('.one') && ph(I18N[l][k]) !== ph(I18N.en[k])) fail.push('placeholder mismatch ' + l + ': ' + k);
for (const [k, v] of Object.entries(I18N.en)) if (/[가-힣]/.test(v)) fail.push('Korean in en catalog: ' + k);
for (const [k, v] of Object.entries(I18N['zh-CN'])) if (/[가-힣]/.test(v)) fail.push('Korean in zh-CN catalog: ' + k);
// Chinese sentences use full-width punctuation: no ASCII ':' '(' ')' ',' next to Han characters.
for (const [k, v] of Object.entries(I18N['zh-CN'])) if (/[\u4e00-\u9fff][:(),]|[(,][\u4e00-\u9fff]/.test(v)) fail.push('ASCII punctuation in zh-CN: ' + k + ' = ' + v);
const used = new Set();
for (const r of main.matchAll(/\b(?:t|tn|kvLabel)\('([\w.]+)'/g)) used.add(r[1]);
for (const r of bundle.matchAll(/\bt\)?\(["']([\w.]+)["']/g)) used.add(r[1]);
for (const r of html.matchAll(/data-i18n(?:-aria|-title|-label|-count)?="([\w.]+)"/g)) used.add(r[1]);
for (const r of main.matchAll(/opts\.t\('([\w.]+)'/g)) used.add(r[1]);
// Keys composed at runtime ('preset.' + value, 'period.' + preset) are validated expanded.
const dynamic = ['preset.', 'period.'];
for (const p of dynamic) { used.delete(p); for (const v of ['today','yesterday','7d','30d','month','custom']) used.add(p + v); }
for (const k of ['lang.menu', 'lang.auto']) used.add(k);
for (const c of codes) used.add('err.' + c);
for (const k of used) for (const l of langs) if (!(k in I18N[l])) fail.push('key used but missing in ' + l + ': ' + k);
// The old toggle keys are gone.
for (const l of langs) for (const k of ['lang.label', 'lang.toEn', 'lang.toKo']) if (k in I18N[l]) fail.push('stale key ' + l + ': ' + k);
if (I18N.en['lang.menu'] !== 'Language' || I18N.ko['lang.menu'] !== '언어' || I18N['zh-CN']['lang.menu'] !== '语言') fail.push('lang.menu labels');
if (I18N.en['lang.auto'] !== 'Follow CLIProxyAPI' || I18N.ko['lang.auto'] !== 'CLIProxyAPI 설정 따르기' || I18N['zh-CN']['lang.auto'] !== '跟随 CLIProxyAPI') fail.push('lang.auto labels');
if (fail.length) { console.error(fail.join('\n')); process.exit(1); }
console.log('OK');
`)
}

func jsonList(items []string) string {
	return "['" + strings.Join(items, "','") + "']"
}

// coreScript returns the shared language core (between the @lang-core markers) for Node tests.
func coreScript(t *testing.T) string {
	t.Helper()
	html := plugin.DashboardHTML
	a, b := strings.Index(html, "// @lang-core-begin"), strings.Index(html, "// @lang-core-end")
	if a < 0 || b < a {
		t.Fatal("@lang-core markers not found")
	}
	return html[a:b]
}

// The shared language core: tag mapping, auto (console → browser → en), pinned preferences,
// invalid stored values, old keys ignored, and the console's own entry is never written.
func TestDashboardLanguageCoreNode(t *testing.T) {
	core := coreScript(t)
	runDashboardNode(t, "const core = "+strconvQuote(core)+";\n"+`
const vm = require('vm');
function setNav(l) { Object.defineProperty(globalThis, 'navigator', { value: l ? { language: l, languages: [l] } : {}, configurable: true }); }
function make(opts) {
  const store = Object.assign({}, opts.store || {});
  const writes = [];
  const parentRoot = { lang: opts.parentLang || '', getAttribute(n) { return n === 'lang' ? this.lang : null; } };
  const win = {};
  win.parent = opts.parentLang !== undefined ? { document: { documentElement: parentRoot } } : win;
  setNav(opts.nav);
  const ctx = { console, JSON, String, window: win, navigator: globalThis.navigator,
    localStorage: { getItem: k => (k in store ? store[k] : null), setItem: (k, v) => { writes.push(k); store[k] = String(v); }, removeItem: k => { writes.push(k); delete store[k]; } } };
  vm.createContext(ctx); vm.runInContext(core, ctx);
  return { c: ctx, store, writes, parentRoot };
}
function assert(c, m) { if (!c) throw new Error(m); }
const zustand = l => JSON.stringify({ state: { language: l }, version: 0 });
let { c } = make({});
const table = { ko: 'ko', 'ko-KR': 'ko', zh: 'zh-CN', 'zh-CN': 'zh-CN', 'zh-TW': 'zh-CN', 'zh-Hant': 'zh-CN', zh_HK: 'zh-CN', 'en-US': 'en', ru: 'en', tr: 'en', vi: 'en', '': '' };
for (const [k, v] of Object.entries(table)) assert(c.normLang(k) === v, 'normLang(' + k + ') = ' + c.normLang(k));
assert(c.intlLocale('ko') === 'ko-KR' && c.intlLocale('zh-CN') === 'zh-CN' && c.intlLocale('en') === 'en-US', 'intlLocale');
// auto: host JSON, {language}, bare string, parent <html lang>, no host → browser.
assert(make({ store: { 'cli-proxy-language': zustand('zh-CN') }, nav: 'en-US' }).c.resolveLang('auto') === 'zh-CN', 'zustand JSON');
assert(make({ store: { 'cli-proxy-language': '{"language":"ko"}' }, nav: 'en-US' }).c.resolveLang('auto') === 'ko', '{language}');
assert(make({ store: { 'cli-proxy-language': 'zh-TW' }, nav: 'en-US' }).c.resolveLang('auto') === 'zh-CN', 'bare string');
assert(make({ store: { 'cli-proxy-language': 'ru' }, nav: 'ko-KR' }).c.resolveLang('auto') === 'en', 'other console language → en');
assert(make({ parentLang: 'zh-CN', store: { 'cli-proxy-language': 'ko' }, nav: 'en-US' }).c.resolveLang('auto') === 'zh-CN', 'parent <html lang> first');
assert(make({ nav: 'zh-CN' }).c.resolveLang('auto') === 'zh-CN', 'no host → browser zh');
assert(make({ nav: 'ko-KR' }).c.resolveLang('auto') === 'ko', 'no host → browser ko');
assert(make({}).c.resolveLang('auto') === 'en', 'nothing → en');
// Pinned preferences win over the host.
let r = make({ parentLang: 'ko', store: { cpa_plugin_language: 'zh-CN' } });
assert(r.c.readLangPref() === 'zh-CN' && r.c.resolveLang(r.c.readLangPref()) === 'zh-CN', 'pinned zh-CN wins');
r = make({ store: { 'cli-proxy-language': zustand('zh-CN'), cpa_plugin_language: 'en' } });
assert(r.c.resolveLang(r.c.readLangPref()) === 'en', 'pinned en wins');
// Invalid stored value → auto; old keys ignored.
r = make({ store: { cpa_plugin_language: 'fr', 'cli-proxy-language': 'ko' } });
assert(r.c.readLangPref() === 'auto' && r.c.resolveLang(r.c.readLangPref()) === 'ko', 'invalid → auto');
r = make({ store: { cpa_plugin_lang: 'en', cpa_plugin_lang_host: 'ko', 'cli-proxy-language': 'ko' }, nav: 'en-US' });
assert(r.c.readLangPref() === 'auto' && r.c.resolveLang(r.c.readLangPref()) === 'ko', 'old keys ignored');
// Writing: only the shared key, valid values only, never the console's entry.
r = make({ parentLang: 'ko' });
r.c.writeLangPref('zh-CN'); assert(r.store.cpa_plugin_language === 'zh-CN', 'write zh-CN');
r.c.writeLangPref('bogus'); assert(r.store.cpa_plugin_language === 'auto', 'invalid write → auto');
r.c.writeLangPref('auto'); r.c.readLangPref(); r.c.hostLangRaw();
assert(r.writes.every(k => k === 'cpa_plugin_language'), 'writes: ' + r.writes);
console.log('OK');
`)
}

func strconvQuote(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}

// Page glue: formatters and text follow the shown language; the menu choice is stored as the
// shared preference; storage events and the console's <html lang> are followed live (auto only).
func TestDashboardLanguageSwitchNodeContract(t *testing.T) {
	runDashboardNode(t, `
const vm = require('vm');
const main = html.match(/<script>([\s\S]*?)<\/script>/)[1];
const core = main.slice(main.indexOf('var I18N = {'), main.indexOf('// Theme: system'));
function makeCtx(opts) {
  const store = Object.assign({}, opts.store || {});
  const writes = [], listeners = {}, observers = [];
  const parentRoot = { lang: opts.parentLang || '', getAttribute(n) { return n === 'lang' ? this.lang : null; } };
  const win = { addEventListener: (type, fn) => { (listeners[type] = listeners[type] || []).push(fn); } };
  win.parent = opts.parentLang !== undefined ? { document: { documentElement: parentRoot } } : win;
  win.MutationObserver = function(fn) { this.observe = () => observers.push(fn); };
  const nav = opts.nav ? { languages: opts.nav, language: opts.nav[0] } : {};
  const ctx = { console, Math, Date, JSON, Object, String, Number, isNaN, Intl, window: win, navigator: nav, MutationObserver: win.MutationObserver,
    document: { documentElement: { attrs: {}, setAttribute(k, v) { this.attrs[k] = v; } }, title: '', querySelectorAll: () => [], getElementById: () => null },
    DOM: {}, state: {}, extractFilterOptions() {}, updateSelectTitle() {}, updateActiveFilterBadge() {},
    renderSummary() {}, paintRecords() {}, showRecordDetail() {},
    localStorage: { getItem: k => (k in store ? store[k] : null), setItem: (k, v) => { writes.push(k); store[k] = String(v); } } };
  vm.createContext(ctx); vm.runInContext(core, ctx); vm.runInContext('initLanguageControls()', ctx);
  return { ctx, store, writes, parentRoot, fire: (key) => (listeners.storage || []).forEach(fn => fn({ key })), mutate: () => observers.forEach(fn => fn()) };
}
function assert(c, m) { if (!c) throw new Error(m); }
assert(makeCtx({}).ctx.currentLang === 'en', 'defaults to en without host or navigator');
assert(makeCtx({ nav: ['ko-KR'] }).ctx.currentLang === 'ko', 'ko-KR browser selects ko');
assert(makeCtx({ nav: ['zh-TW'] }).ctx.currentLang === 'zh-CN', 'zh-TW browser selects zh-CN');
assert(makeCtx({ nav: ['fr-FR'] }).ctx.currentLang === 'en', 'unsupported browser language falls back to en');
let r = makeCtx({ parentLang: 'zh-CN' });
let c = r.ctx;
assert(c.currentLang === 'zh-CN' && c.document.documentElement.attrs.lang === 'zh-CN' && c.document.title === 'Token 用量仪表盘', 'zh-CN on load: lang + title');
assert(c.uiLocale() === 'zh-CN', 'zh locale');
assert(c.t('chart.summary', { d: 3, r: '36', m: '12' }) === '3 天 · 共 36 次请求 · 峰值 12 次', 'zh chart summary: ' + c.t('chart.summary', { d: 3, r: '36', m: '12' }));
assert(c.kvLabel('lbl.requests', '1,234') === '请求总数：1,234', 'zh kv label');
assert(c.tn('records.count', 1) === '1 条' && c.tn('records.count', 3) === '3 条', 'zh count');
// Console switches to ko live (parent <html lang> mutation) → follows.
r.parentRoot.lang = 'ko'; r.mutate();
assert(c.currentLang === 'ko', 'follows console ko');
assert(c.t('chart.summary', { d: 3, r: '36', m: '12' }) === '3일 · 총 36건 · 최고 12건', 'ko chart summary');
assert(c.uiLocale() === 'ko-KR', 'ko locale');
// Pin English via the menu path: stays English after the console changes.
c.selectLanguagePref('en');
assert(c.currentLang === 'en' && r.store.cpa_plugin_language === 'en', 'pinned en stored');
assert(c.tn('records.count', 1) === '1 record' && c.tn('records.count', 3) === '3 records', 'en plural');
r.parentRoot.lang = 'zh-CN'; r.mutate();
assert(c.currentLang === 'en', 'pinned en ignores console change');
// Another plugin page chose "Follow CLIProxyAPI" → storage event → follow the console again.
r.store.cpa_plugin_language = 'auto'; r.fire('cpa_plugin_language');
assert(c.currentLang === 'zh-CN', 'auto via storage event follows console');
// Host key storage event (no same-origin parent).
r = makeCtx({ store: { 'cli-proxy-language': JSON.stringify({ state: { language: 'en' } }) } });
assert(r.ctx.currentLang === 'en', 'start en');
r.store['cli-proxy-language'] = JSON.stringify({ state: { language: 'ko' } }); r.fire('cli-proxy-language');
assert(r.ctx.currentLang === 'ko', 'host key storage event followed');
// Old keys ignored, never written; the console entry is never written.
r = makeCtx({ store: { cpa_plugin_lang: 'en', cpa_plugin_lang_host: 'ko' }, parentLang: 'ko' });
assert(r.ctx.currentLang === 'ko', 'old keys ignored');
r.ctx.selectLanguagePref('zh-CN'); r.ctx.selectLanguagePref('auto');
assert(r.writes.every(k => k === 'cpa_plugin_language'), 'writes only the shared key: ' + r.writes);
assert(r.ctx.t('no.such.key') === 'no.such.key', 'missing key degrades to the key itself');
console.log('OK');
`)
}
