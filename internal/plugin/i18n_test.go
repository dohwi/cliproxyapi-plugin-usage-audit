package plugin_test

import (
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

// The dashboard must never persist credentials. The only permitted browser storage
// is the language preference stored under LANG_KEY.
func TestDashboardStorageLimitedToLanguagePreference(t *testing.T) {
	html := plugin.DashboardHTML
	if strings.Contains(html, "sessionStorage") {
		t.Fatal("dashboard must not use sessionStorage")
	}
	if !strings.Contains(html, "var LANG_KEY = 'cpa_plugin_lang';") {
		t.Fatal("shared language preference key is missing or renamed")
	}
	// The vendored React bundle is scanned separately: it must not touch storage at all.
	script := regexp.MustCompile(`(?s)<script>(.*?)</script>`).FindStringSubmatch(html)
	if script == nil {
		t.Fatal("main script not found")
	}
	total := strings.Count(script[1], "localStorage")
	allowed := regexp.MustCompile(`localStorage\.(?:getItem\(LANG_KEY\)|setItem\(LANG_KEY, lang\))`).FindAllString(script[1], -1)
	if total != len(allowed) || total != 2 {
		t.Fatalf("localStorage references = %d (allowed LANG_KEY accessors = %d, want exactly 2)", total, len(allowed))
	}
	if strings.Contains(html[strings.Index(html, `<script data-dohwi-ui="exported">`):], "localStorage") {
		t.Fatal("bundled UI code must not use localStorage")
	}
}

// Every translation key must exist in both languages with identical placeholders,
// and every key used by the page, markup or bundle must exist in the catalog.
func TestDashboardI18nCatalogsAreComplete(t *testing.T) {
	runDashboardNode(t, `
const main = html.match(/<script>([\s\S]*?)<\/script>/)[1];
const bundle = html.slice(html.indexOf('<script data-dohwi-ui="exported">'));
const start = main.indexOf('var I18N = {'), end = main.indexOf('var currentLang');
const I18N = new Function(main.slice(start, end) + '; return I18N;')();
const fail = [];
const ko = Object.keys(I18N.ko), en = Object.keys(I18N.en);
for (const k of ko) if (!(k in I18N.en)) fail.push('missing in en: ' + k);
for (const k of en) if (!(k in I18N.ko) && !(k.endsWith('.one') && k.slice(0, -4) in I18N.ko)) fail.push('missing in ko: ' + k);
const ph = s => (s.match(/\{\w+\}/g) || []).sort().join();
for (const k of ko) if (k in I18N.en && ph(I18N.ko[k]) !== ph(I18N.en[k])) fail.push('placeholder mismatch: ' + k);
for (const [k, v] of Object.entries(I18N.en)) if (/[가-힣]/.test(v)) fail.push('Korean in en catalog: ' + k);
const used = new Set();
for (const r of main.matchAll(/\b(?:t|tn|kvLabel)\('([\w.]+)'/g)) used.add(r[1]);
for (const r of bundle.matchAll(/\bt\)?\(["']([\w.]+)["']/g)) used.add(r[1]);
for (const r of html.matchAll(/data-i18n(?:-aria|-title|-label|-count)?="([\w.]+)"/g)) used.add(r[1]);
// Keys composed at runtime ('preset.' + value, 'period.' + preset) are validated expanded.
const dynamic = ['preset.', 'period.'];
for (const p of dynamic) { used.delete(p); for (const v of ['today','yesterday','7d','30d','month','custom']) used.add(p + v); }
for (const k of used) if (!(k in I18N.ko)) fail.push('unknown key used: ' + k);
if (fail.length) { console.error(fail.join('\n')); process.exit(1); }
console.log('OK');
`)
}

// Language switching rebuilds formatters and text, persists the choice, follows the
// shared storage key, and ignores unsupported values.
func TestDashboardLanguageSwitchNodeContract(t *testing.T) {
	runDashboardNode(t, `
const vm = require('vm');
const main = html.match(/<script>([\s\S]*?)<\/script>/)[1];
// The static page helpers only; the network bootstrap at the end is not under test.
const cut = main.indexOf('function eachNode(');
const core = main.slice(main.indexOf('var LANG_KEY'), cut);
function makeCtx(stored, navLangs) {
  const store = {}; if (stored) store.cpa_plugin_lang = stored;
  const ctx = { console, Math, Date, JSON, Object, String, Number, isNaN,
    navigator: navLangs ? { languages: navLangs, language: navLangs[0] } : undefined,
    localStorage: { getItem: k => (k in store ? store[k] : null), setItem: (k, v) => { store[k] = v; } } };
  vm.createContext(ctx); vm.runInContext(core, ctx);
  return { ctx, store };
}
function assert(c, m) { if (!c) throw new Error(m); }
assert(makeCtx(null, null).ctx.currentLang === 'ko', 'defaults to ko without navigator');
assert(makeCtx(null, ['en-US', 'ko']).ctx.currentLang === 'en', 'en-US selects en');
assert(makeCtx(null, ['ko-KR']).ctx.currentLang === 'ko', 'ko-KR selects ko');
assert(makeCtx(null, ['fr-FR']).ctx.currentLang === 'en', 'unsupported browser language falls back to en');
assert(makeCtx('ko', ['en-US']).ctx.currentLang === 'ko', 'stored preference wins');
assert(makeCtx('xx', ['ko-KR']).ctx.currentLang === 'ko', 'invalid stored value ignored');
const { ctx: c } = makeCtx('ko', null);
assert(c.t('chart.summary', { d: 3, r: '36', m: '12' }) === '3일 · 총 36건 · 최고 12건', 'ko chart summary');
assert(c.kvLabel('lbl.requests', '1,234') === '총 요청: 1,234', 'ko kv label');
assert(c.uiLocale() === 'ko-KR', 'ko locale');
c.currentLang = 'en';
assert(c.t('chart.summary', { d: 3, r: '36', m: '12' }) === '3 days · 36 requests total · peak 12', 'en chart summary: ' + c.t('chart.summary', { d: 3, r: '36', m: '12' }));
assert(c.kvLabel('lbl.requests', '1,234') === 'Total requests: 1,234', 'en kv label');
assert(c.tn('records.count', 1) === '1 record' && c.tn('records.count', 3) === '3 records', 'en plural');
assert(c.uiLocale() === 'en-US', 'en locale');
assert(c.t('no.such.key') === 'no.such.key', 'missing key degrades to the key itself');
console.log('OK');
`)
}
