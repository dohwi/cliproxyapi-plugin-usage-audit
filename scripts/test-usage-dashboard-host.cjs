#!/usr/bin/env node
// Browser regression for running inside the host Management Center:
//  - the page is served with the same hash-based CSP as the plugin (dashboardCSP in plugin.go)
//    and must load and work with zero CSP violations;
//  - the page language follows a simulated host (parent <html lang> and cli-proxy-language),
//    "last action wins" against the plugin's own toggle;
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

(async () => {
  const browser = await chromium.launch({ headless: true, executablePath: process.env.BROWSER_EXECUTABLE || undefined, args: ['--no-sandbox'] });
  const shots = process.env.SCREENSHOT_DIR;
  if (shots) fs.mkdirSync(shots, { recursive: true });
  try {
    // 1. Host ko, nothing saved → ko. Signed-in view at three widths, CSP clean, below 88px.
    for (const width of [412, 820, 1280]) {
      const s = await open(browser, { width, hostLang: 'ko', storage: { 'cli-proxy-language': JSON.stringify({ state: { language: 'ko' }, version: 0 }) } });
      await s.frame.waitForSelector('#recordsBody tr.ui-table-row');
      assert.equal(await lang(s.frame), 'ko', `${width}: follows host ko`);
      assert.equal(await s.frame.evaluate(() => document.title), '토큰 사용량 대시보드');
      // The interactions that used to be inline onclick handlers still work under the CSP.
      await s.frame.click('#tab-model-tab');
      assert.equal(await s.frame.getAttribute('#tab-model-tab', 'aria-selected'), 'true', 'tab click works');
      const label = await s.frame.evaluate(() => [...document.querySelectorAll('#chartArea svg text')].map(t => t.textContent).filter(t => /\d/.test(t)).pop() || '');
      assert.match(label, /^\d{1,2}\. \d{1,2}\.$/, `${width}: ko chart label uses the UI locale (${label})`);
      const top = await topCheck(s.frame);
      assert.equal(top.offset, 88, 'top offset variable');
      assert.deepEqual(top.bad, [], `${width}: elements above y=88`);
      const inner = await s.frame.evaluate(() => window.__csp || []);
      assert.deepEqual([...s.violations, ...inner], [], `${width}: CSP violations`);
      assert.deepEqual(s.errors, [], `${width}: page errors`);
      if (shots) await s.page.screenshot({ path: path.join(shots, `host-ko-signed-in-${width}.png`) });
      await s.ctx.close();
    }

    // 2. Auth gate (no key) at three widths: CSP clean, below 88px, host language en.
    for (const width of [412, 820, 1280]) {
      const s = await open(browser, { width, hostLang: 'en', auth: false });
      await s.frame.waitForSelector('#authGate:not(.is-hidden)');
      assert.equal(await lang(s.frame), 'en');
      const top = await topCheck(s.frame);
      assert.deepEqual(top.bad, [], `${width}: auth gate elements above y=88`);
      const inner = await s.frame.evaluate(() => window.__csp || []);
      assert.deepEqual([...s.violations, ...inner], [], `${width}: auth gate CSP violations`);
      if (shots) await s.page.screenshot({ path: path.join(shots, `host-en-auth-gate-${width}.png`) });
      await s.ctx.close();
    }

    // 3. Last action wins: toggle beats the host until the host changes again; live follow.
    {
      const s = await open(browser, { width: 1280, hostLang: 'ko' });
      await s.frame.waitForSelector('#recordsBody tr.ui-table-row');
      assert.equal(await lang(s.frame), 'ko');
      await s.frame.click('#languageToggle');
      await waitLang(s.frame, 'en');
      assert.deepEqual(await s.frame.evaluate(() => [localStorage.getItem('cpa_plugin_lang'), localStorage.getItem('cpa_plugin_lang_host'), localStorage.getItem('cli-proxy-language')]), ['en', 'ko', null], 'toggle stored with host language; host entry untouched');
      await s.page.reload();
      const frame = await (await s.page.waitForSelector('#plugin')).contentFrame();
      await frame.waitForSelector('#recordsBody tr.ui-table-row');
      assert.equal(await lang(frame), 'en', 'toggle survives reload while the host stays ko');
      // Host switches language live (parent <html lang> mutation) → follow.
      await s.page.evaluate(() => document.documentElement.setAttribute('lang', 'zh-CN'));
      await waitLang(frame, 'en');
      await s.page.evaluate(() => document.documentElement.setAttribute('lang', 'ko'));
      await waitLang(frame, 'ko');
      assert.equal(await frame.evaluate(() => document.title), '토큰 사용량 대시보드');
      assert.equal(await frame.evaluate(() => localStorage.getItem('cpa_plugin_lang')), 'ko', 'host change propagated to LANG_KEY');
      // Server error codes are shown in the UI language.
      await s.ctx.route(u => u.pathname.endsWith('/data/summary'), route => route.fulfill({ status: 400, contentType: 'application/json', body: JSON.stringify({ error: 'summary range must be ordered and at most 366 days', code: 'summary_range_too_large' }) }));
      await frame.waitForFunction(() => !state.inFlight);
      await frame.click('#refreshBtn');
      await frame.waitForSelector('#errorBox:not(.is-hidden)', { timeout: 10000 }).catch(async e => { throw new Error(e.message.split('\n')[0] + ' | ' + JSON.stringify(await frame.evaluate(() => ({ inFlight: state.inFlight, err: document.getElementById('errorMessage').textContent, cls: document.getElementById('errorBox').className })))); });
      assert.match(await frame.textContent('#errorMessage'), /조회 기간이 너무 깁니다/, 'error code mapped to Korean text');
      const inner = await frame.evaluate(() => window.__csp || []);
      assert.deepEqual([...s.violations, ...inner], [], 'CSP violations during interaction');
      if (shots) await s.page.screenshot({ path: path.join(shots, 'host-ko-error-1280.png') });
      await s.ctx.close();
    }
    console.log('PASS host: hash CSP with zero violations, host language follow (last action wins), localized errors/dates, nothing above 88px at 412/820/1280 (auth gate + signed in)');
  } catch (e) {
    console.error('FAIL', e.message);
    process.exitCode = 1;
  } finally {
    await browser.close();
  }
})();
