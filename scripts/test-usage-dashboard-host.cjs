#!/usr/bin/env node
// Browser regression for running inside the host Management Center:
//  - the page is served with the same hash-based CSP as the plugin (dashboardCSP in plugin.go)
//    and must load and work with zero CSP violations;
//  - the page language follows a simulated host (parent <html lang> and cli-proxy-language) while the
//    shared preference is "Follow CLIProxyAPI" (auto); a language picked in the menu stays pinned;
//  - the language menu (popover) works with mouse and keyboard and stays inside the viewport;
//  - nothing sits above the console's floating buttons (top offset 88px) at 412/820/1280.
// Install Playwright or set PLAYWRIGHT_MODULE; BROWSER_EXECUTABLE selects a system browser;
// SCREENSHOT_DIR saves screenshots.
const assert = require('node:assert/strict');
const crypto = require('node:crypto');
const fs = require('node:fs');
const path = require('node:path');
const { chromium } = require(process.env.PLAYWRIGHT_MODULE || 'playwright');
const root = path.resolve(__dirname, '..');
const html = require('./lib/dashboard-html.cjs').readDashboardHTML(root);

// Same algorithm as dashboardCSP in internal/plugin/plugin.go (a Go test pins the server side).
function dashboardCSP(page) {
  const hashes = [...page.matchAll(/<script\b([^>]*)>([\s\S]*?)<\/script>/gi)]
    .map(m => "'sha256-" + crypto.createHash('sha256').update(m[2], 'utf8').digest('base64') + "'");
  return "default-src 'none'; script-src " + hashes.join(' ') + "; style-src 'unsafe-inline'; font-src 'self'; img-src 'self' data:; connect-src 'self'; base-uri 'none'; form-action 'none'; frame-ancestors 'self'";
}
const csp = dashboardCSP(html);
assert.ok(!/script-src[^;]*'unsafe-inline'/.test(csp), 'script-src must not allow unsafe-inline');

const now = new Date().toISOString();
const metric = (provider, model) => ({ provider, model, requests: 1234, failures: 3, generate_count: 100, input_tokens: 9000, fresh_input_tokens: 7000, output_tokens: 3000, cached_tokens: 2000, reasoning_tokens: 500, total_tokens: 1234567 });
const summary = {
  total_count: 1234,
  summary: [metric('anthropic', 'claude-opus'), metric('openai', 'gpt-5')],
  by_provider: [{ ...metric('anthropic', 'x'), key: 'anthropic' }], by_endpoint: [{ ...metric('a', 'x'), key: 'direct' }], by_model: [{ ...metric('a', 'x'), key: 'claude-opus' }],
  daily: [0, 1, 2].map(i => ({ key: new Date(Date.now() - i * 86400000).toISOString().slice(0, 10), requests: 10 + i, total_tokens: 100, input_tokens: 60, output_tokens: 40, failures: i })),
};
const records = { has_more: false, records: Array.from({ length: 3 }, (_, i) => ({ id: 'r' + i, requested_at: now, provider: 'anthropic', model: 'claude-opus', endpoint: 'direct', fresh_input: 10, cached: 5, output: 20, total: 35, input: 15, reasoning: 0, failed: i === 1, failure_status: i === 1 ? 500 : 0, latency: 1234, ttft: 210 })) };
const ORIGIN = 'http://host.fixture';
const DASH = ORIGIN + '/v0/resource/plugins/usage-audit/dashboard';
// The host page: an iframe filling the content area, with its own <html lang>.
const hostPage = lang => `<!doctype html><html lang="${lang}"><head><meta charset="utf-8"><title>Host</title><style>html,body{margin:0;height:100%}iframe{border:0;width:100%;height:100%}</style></head><body><iframe id="plugin" src="${DASH}"></iframe></body></html>`;

async function open(browser, { width, hostLang, storage, auth = true, summaryStatus = 200, summaryBody = summary }) {
  const ctx = await browser.newContext({ locale: 'en-US', reducedMotion: 'reduce', viewport: { width, height: 900 } });
  await ctx.addInitScript(([s, a]) => {
    try {
      if (location.pathname === '/host') for (const [k, v] of Object.entries(s || {})) localStorage.setItem(k, v);
      if (a) sessionStorage.setItem('cpa_usage_audit_mgmt_key', 'fixture-management-key');
    } catch (e) {}
  }, [storage, auth]);
  const page = await ctx.newPage();
  const violations = [], errors = [];
  page.on('console', m => { if (/Content Security Policy|Refused to/i.test(m.text())) violations.push(m.text()); });
  page.on('pageerror', e => errors.push(e.message));
  await ctx.route('**/*', async route => {
    const url = new URL(route.request().url());
    if (url.pathname === '/host') return route.fulfill({ contentType: 'text/html', body: hostPage(hostLang) });
    if (url.pathname.endsWith('/dashboard')) return route.fulfill({ contentType: 'text/html; charset=utf-8', headers: { 'Content-Security-Policy': csp }, body: html });
    if (url.pathname.endsWith('/data/summary')) return route.fulfill({ status: summaryStatus, contentType: 'application/json', body: JSON.stringify(summaryBody) });
    if (url.pathname.endsWith('/data/records')) return route.fulfill({ contentType: 'application/json', body: JSON.stringify(records) });
    const asset = path.join(root, 'internal/plugin/assets/fonts', path.basename(url.pathname));
    if (fs.existsSync(asset) && fs.statSync(asset).isFile()) return route.fulfill({ contentType: 'font/woff2', body: fs.readFileSync(asset) });
    return route.fulfill({ status: 404, body: '' });
  });
  // CSP violations inside the iframe are reported as securitypolicyviolation events there.
  await ctx.addInitScript(() => { document.addEventListener('securitypolicyviolation', e => { (window.__csp = window.__csp || []).push(e.violatedDirective + ' ' + e.blockedURI); }); });
  await page.goto(ORIGIN + '/host');
  const frame = await (await page.waitForSelector('#plugin')).contentFrame();
  return { ctx, page, frame, violations, errors };
}
const lang = frame => frame.evaluate(() => document.documentElement.lang);
const waitLang = (frame, l) => frame.waitForFunction(x => document.documentElement.lang === x, l, { timeout: 5000 });

// Every visible element's top edge must be at or below the console's floating-button line.
async function topCheck(frame) {
  return frame.evaluate(() => {
    const offset = parseFloat(getComputedStyle(document.documentElement).getPropertyValue('--dashboard-top-offset'));
    const bad = [];
    document.querySelectorAll('body *').forEach(el => {
      const cs = getComputedStyle(el), r = el.getBoundingClientRect();
      if (cs.display === 'none' || cs.visibility === 'hidden' || r.width === 0 || r.height === 0 || cs.position === 'fixed') return;
      if (el.closest('[hidden], .is-hidden, dialog:not([open])')) return;
      if (r.top + window.scrollY < 88 - 0.5) bad.push((el.id || el.className || el.tagName) + '@' + Math.round(r.top));
    });
    return { offset, bad: bad.slice(0, 5) };
  });
}

// Visible text that overflows its box (Chinese labels must fit buttons, filters, headers and cards).
async function clipCheck(frame) {
  return frame.evaluate(() => {
    const out = [];
    if (document.documentElement.scrollWidth > innerWidth + 1) out.push('page-h-scroll');
    document.querySelectorAll('.btn, .tab-btn, .preset-btn, .ui-segmented-item, .badge, .kpi-label, .kpi-sub, .card-header h2, .card-header h3, #recordsTable .ui-table-head, .breakdown-table-wrap .ui-table-head, .compact-select, .ui-select-trigger, .ui-date-picker-trigger, #activeFilterBadge, .auth-gate-head h2, .lang-menu-item').forEach(e => {
      if (!e.offsetWidth) return;
      const cs = getComputedStyle(e);
      if (e.scrollWidth > e.clientWidth + 1 && cs.textOverflow !== 'ellipsis' && cs.overflowX !== 'auto') out.push(e.tagName + '.' + String(e.className).slice(0, 24) + ' "' + e.textContent.trim().slice(0, 16) + '" ' + e.scrollWidth + '>' + e.clientWidth);
    });
    return out;
  });
}
// Opens the language menu from button and returns its geometry and items.
async function openMenu(frame, button, how = 'click') {
  if (how === 'click') await frame.click(button);
  else { await frame.focus(button); await frame.press(button, how); }
  await frame.waitForSelector('.lang-menu');
  return frame.evaluate(b => {
    const m = document.querySelector('.lang-menu'), r = m.getBoundingClientRect(), br = document.querySelector(b).getBoundingClientRect();
    const items = [...m.querySelectorAll('.lang-menu-item')].map(i => ({ pref: i.dataset.pref, text: i.querySelector('.lang-menu-label').textContent, sub: (i.querySelector('.lang-menu-sub') || {}).textContent || '', lang: i.getAttribute('lang'), checked: i.getAttribute('aria-checked') }));
    return { box: [r.left, r.top, r.right, r.bottom], vw: document.documentElement.clientWidth, vh: document.documentElement.clientHeight, buttonTop: br.top + scrollY, expanded: document.querySelector(b).getAttribute('aria-expanded'), haspopup: document.querySelector(b).getAttribute('aria-haspopup'), focused: document.activeElement && document.activeElement.dataset.pref, items };
  }, button);
}
function assertMenu(m, where) {
  assert.deepEqual(m.items.map(i => i.pref), ['auto', 'en', 'ko', 'zh-CN'], where + ': menu order');
  assert.deepEqual(m.items.slice(1).map(i => i.text), ['English', '한국어', '简体中文'], where + ': native names');
  assert.deepEqual(m.items.slice(1).map(i => i.lang), ['en', 'ko', 'zh-CN'], where + ': item lang attributes');
  assert.equal(m.expanded, 'true', where + ': aria-expanded'); assert.equal(m.haspopup, 'menu', where + ': aria-haspopup');
  assert.ok(m.box[0] >= 0 && m.box[1] >= 0 && m.box[2] <= m.vw && m.box[3] <= m.vh, where + ': menu inside the viewport ' + JSON.stringify(m.box) + ' vw=' + m.vw);
  assert.ok(m.buttonTop >= 88 - 0.5, where + ': language button below y=88 (' + m.buttonTop + ')');
}
const ZH_TITLE = 'Token 用量仪表盘', KO_TITLE = '토큰 사용량 대시보드';
const zustand = l => JSON.stringify({ state: { language: l }, version: 0 });
const cleanCSP = async (s, where) => { const inner = await s.frame.evaluate(() => window.__csp || []); assert.deepEqual([...s.violations, ...inner], [], where + ': CSP violations'); assert.deepEqual(s.errors, [], where + ': page errors'); };

(async () => {
  const browser = await chromium.launch({ headless: true, executablePath: process.env.BROWSER_EXECUTABLE || undefined, args: ['--no-sandbox'] });
  const shots = process.env.SCREENSHOT_DIR;
  if (shots) fs.mkdirSync(shots, { recursive: true });
  try {
    // 1. Console zh-CN, nothing saved → Chinese on load. Signed-in view at three widths: CSP clean,
    //    below 88px, Chinese text fits, menu inside the viewport, keyboard open/close.
    for (const width of [412, 820, 1280]) {
      const s = await open(browser, { width, hostLang: 'zh-CN', storage: { 'cli-proxy-language': zustand('zh-CN') } });
      await s.frame.waitForSelector('#recordsBody tr.ui-table-row');
      assert.equal(await lang(s.frame), 'zh-CN', `${width}: follows console zh-CN`);
      assert.equal(await s.frame.evaluate(() => document.title), ZH_TITLE);
      assert.equal(await s.frame.textContent('h1'), 'Token 用量');
      assert.equal(await s.frame.getAttribute('#languageToggle', 'aria-label'), '语言：简体中文');
      // The interactions that used to be inline onclick handlers still work under the CSP.
      await s.frame.click('#tab-model-tab');
      assert.equal(await s.frame.getAttribute('#tab-model-tab', 'aria-selected'), 'true', 'tab click works');
      await s.frame.click('#tab-provider-tab');
      const label = await s.frame.evaluate(() => [...document.querySelectorAll('#chartArea svg text')].map(t => t.textContent).filter(t => /\d/.test(t)).pop() || '');
      assert.match(label, /^\d{1,2}\/\d{1,2}$/, `${width}: zh chart label uses the zh-CN locale (${label})`);
      const top = await topCheck(s.frame);
      assert.equal(top.offset, 88, 'top offset variable');
      assert.deepEqual(top.bad, [], `${width}: elements above y=88`);
      assert.deepEqual(await clipCheck(s.frame), [], `${width}: Chinese text clipped`);
      await s.frame.evaluate(() => document.fonts.ready);
      if (shots) await s.page.screenshot({ path: path.join(shots, `host-zh-signed-in-${width}.png`), fullPage: true });
      // Menu by mouse: placement, items, check mark on auto, console language as hint.
      const m = await openMenu(s.frame, '#languageToggle');
      assertMenu(m, `${width} signed-in`);
      assert.equal(m.items[0].text, '跟随 CLIProxyAPI'); assert.equal(m.items[0].sub, '简体中文');
      assert.equal(m.items.find(i => i.checked === 'true').pref, 'auto', 'auto checked by default');
      assert.deepEqual(await clipCheck(s.frame), [], `${width}: menu text clipped`);
      if (shots && width !== 820) await s.page.screenshot({ path: path.join(shots, `host-zh-menu-${width}.png`) });
      // Outside click closes.
      await s.frame.click('h1');
      await s.frame.waitForSelector('.lang-menu', { state: 'detached' });
      assert.equal(await s.frame.getAttribute('#languageToggle', 'aria-expanded'), 'false');
      // Keyboard: ArrowDown opens on the checked item, arrows move, Escape closes and refocuses the button.
      const k = await openMenu(s.frame, '#languageToggle', 'ArrowDown');
      assert.equal(k.focused, 'auto', 'ArrowDown focuses the checked item');
      await s.frame.press('.lang-menu', 'ArrowDown');
      assert.equal(await s.frame.evaluate(() => document.activeElement.dataset.pref), 'en');
      await s.frame.press('.lang-menu', 'End');
      assert.equal(await s.frame.evaluate(() => document.activeElement.dataset.pref), 'zh-CN');
      await s.frame.press('.lang-menu', 'Escape');
      await s.frame.waitForSelector('.lang-menu', { state: 'detached' });
      assert.equal(await s.frame.evaluate(() => document.activeElement.id), 'languageToggle', 'Escape refocuses the button');
      const u = await openMenu(s.frame, '#languageToggle', 'ArrowUp');
      assert.equal(u.focused, 'zh-CN', 'ArrowUp opens on the last item');
      await s.frame.press('.lang-menu', 'Tab');
      await s.frame.waitForSelector('.lang-menu', { state: 'detached' });
      await cleanCSP(s, `${width}`);
      await s.ctx.close();
    }

    // 2. Auth gate (no key) at three widths, console zh-CN: Chinese, CSP clean, below 88px, menu fits.
    for (const width of [412, 820, 1280]) {
      const s = await open(browser, { width, hostLang: 'zh-CN', auth: false });
      await s.frame.waitForSelector('#authGate:not(.is-hidden)');
      assert.equal(await lang(s.frame), 'zh-CN');
      assert.equal(await s.frame.textContent('#authGateTitle'), '需要管理密钥');
      const top = await topCheck(s.frame);
      assert.deepEqual(top.bad, [], `${width}: auth gate elements above y=88`);
      assert.deepEqual(await clipCheck(s.frame), [], `${width}: auth gate Chinese text clipped`);
      if (shots) await s.page.screenshot({ path: path.join(shots, `host-zh-auth-gate-${width}.png`) });
      const m = await openMenu(s.frame, '#authLanguageToggle');
      assertMenu(m, `${width} auth gate`);
      if (shots && width === 412) await s.page.screenshot({ path: path.join(shots, `host-zh-auth-menu-${width}.png`) });
      await s.frame.click(`.lang-menu-item[data-pref="en"]`);
      await waitLang(s.frame, 'en');
      assert.equal(await s.frame.evaluate(() => document.activeElement.id), 'authLanguageToggle', 'selection refocuses the button');
      await cleanCSP(s, `${width} auth gate`);
      await s.ctx.close();
    }

    // 3. Live follow, pinning and returning to "Follow CLIProxyAPI".
    {
      const s = await open(browser, { width: 1280, hostLang: 'ko', storage: { cpa_plugin_lang: 'en', cpa_plugin_lang_host: 'ko' } });
      await s.frame.waitForSelector('#recordsBody tr.ui-table-row');
      assert.equal(await lang(s.frame), 'ko', 'old keys ignored: auto follows console ko');
      assert.equal(await s.frame.evaluate(() => document.title), KO_TITLE);
      const koLabel = await s.frame.evaluate(() => [...document.querySelectorAll('#chartArea svg text')].map(t => t.textContent).filter(t => /\d/.test(t)).pop() || '');
      assert.match(koLabel, /^\d{1,2}\. \d{1,2}\.$/, `ko chart label uses the UI locale (${koLabel})`);
      // Console changes live (parent <html lang> mutation) → follow.
      await s.page.evaluate(() => document.documentElement.setAttribute('lang', 'zh-CN'));
      await waitLang(s.frame, 'zh-CN');
      await s.page.evaluate(() => document.documentElement.setAttribute('lang', 'ko'));
      await waitLang(s.frame, 'ko');
      // Pick English: stays English after the console changes and after reload.
      await s.frame.click('#languageToggle');
      await s.frame.click('.lang-menu-item[data-pref="en"]');
      await waitLang(s.frame, 'en');
      assert.equal(await s.frame.getAttribute('#languageToggle', 'aria-label'), 'Language: English');
      await s.page.evaluate(() => document.documentElement.setAttribute('lang', 'zh-CN'));
      await s.frame.waitForTimeout(300);
      assert.equal(await lang(s.frame), 'en', 'pinned English ignores the console change');
      assert.deepEqual(await s.frame.evaluate(() => ['cpa_plugin_language', 'cli-proxy-language', 'cpa_plugin_lang', 'cpa_plugin_lang_host'].map(k => localStorage.getItem(k))), ['en', null, 'en', 'ko'], 'only the shared key written; console entry and old keys untouched');
      await s.page.reload();
      let frame = await (await s.page.waitForSelector('#plugin')).contentFrame();
      await frame.waitForSelector('#recordsBody tr.ui-table-row');
      assert.equal(await lang(frame), 'en', 'pinned English survives reload (console ko)');
      // Follow CLIProxyAPI again → console language (ko after reload), then live.
      const m = await openMenu(frame, '#languageToggle');
      assert.equal(m.items.find(i => i.checked === 'true').pref, 'en', 'en checked');
      assert.equal(m.items[0].text, 'Follow CLIProxyAPI'); assert.equal(m.items[0].sub, '한국어');
      await frame.click('.lang-menu-item[data-pref="auto"]');
      await waitLang(frame, 'ko');
      await s.page.evaluate(() => document.documentElement.setAttribute('lang', 'zh-CN'));
      await waitLang(frame, 'zh-CN');
      assert.equal(await frame.evaluate(() => document.title), ZH_TITLE);
      assert.equal(await frame.evaluate(() => localStorage.getItem('cpa_plugin_language')), 'auto');
      // Server error codes are shown in the UI language.
      await s.ctx.route(u => u.pathname.endsWith('/data/summary'), route => route.fulfill({ status: 400, contentType: 'application/json', body: JSON.stringify({ error: 'summary range must be ordered and at most 366 days', code: 'summary_range_too_large' }) }));
      await frame.waitForFunction(() => !state.inFlight);
      await frame.click('#refreshBtn');
      await frame.waitForSelector('#errorBox:not(.is-hidden)', { timeout: 10000 });
      assert.match(await frame.textContent('#errorMessage'), /时间范围过长/, 'error code mapped to Chinese text');
      const inner = await frame.evaluate(() => window.__csp || []);
      assert.deepEqual([...s.violations, ...inner], [], 'CSP violations during interaction');
      assert.deepEqual(s.errors, [], 'page errors during interaction');
      if (shots) await s.page.screenshot({ path: path.join(shots, 'host-zh-error-1280.png') });
      await s.ctx.close();
    }
    console.log('PASS host: hash CSP with zero violations, console language follow (auto) + pinned menu choice, language menu (mouse/keyboard, inside viewport), Chinese text fits, localized errors/dates, nothing above 88px at 412/820/1280 (auth gate + signed in)');
  } catch (e) {
    console.error('FAIL', e.message);
    process.exitCode = 1;
  } finally {
    await browser.close();
  }
})();
