#!/usr/bin/env node
// Management-key gate: the unauthenticated page must send no data request without a key,
// must not burn extra failed attempts on a wrong key, must keep a rejected key out of
// storage, and must send an accepted key only as a Bearer header.
const fs = require('node:fs'), path = require('node:path'), assert = require('node:assert/strict');
const { chromium } = require(process.env.PLAYWRIGHT_MODULE || 'playwright');
const root = path.resolve(__dirname, '..');
const html = require('./lib/dashboard-html.cjs').readDashboardHTML(root);
const GOOD = 'good-management-key', STORE = 'cpa_usage_audit_mgmt_key';
const summary = { total_count: 1, summary: [{ provider: 'codex', model: 'm', requests: 1, total_tokens: 4242 }], by_provider: [], by_endpoint: [], by_model: [], daily: [] };
const records = { records: [{ id: 'r1', requested_at: new Date().toISOString(), provider: 'codex', model: 'm', total_tokens: 4242 }], has_more: false };

(async () => {
  const browser = await chromium.launch({ headless: true, executablePath: process.env.BROWSER_EXECUTABLE || undefined, args: ['--no-sandbox'] });
  try {
    const page = await browser.newPage({ locale: 'en-US' });
    const dataCalls = [], urls = [];
    page.on('pageerror', e => assert.fail('page error: ' + e.message));
    await page.route('**/*', async route => {
      const req = route.request(), url = new URL(req.url());
      if (req.resourceType() === 'document') return route.fulfill({ contentType: 'text/html', body: html });
      if (url.pathname.includes('/assets/')) return route.fulfill({ contentType: 'font/woff2', body: fs.readFileSync(path.join(root, 'internal/plugin/assets/fonts', path.basename(url.pathname))) });
      if (!url.pathname.startsWith('/v0/management/plugins/usage-audit/data/')) return route.fulfill({ status: 404, body: '' });
      urls.push(req.url());
      const auth = req.headers()['authorization'] || '';
      dataCalls.push(auth);
      if (auth !== 'Bearer ' + GOOD) return route.fulfill({ status: 401, contentType: 'application/json', body: '{"error":"invalid management key"}' });
      return route.fulfill({ contentType: 'application/json', body: JSON.stringify(url.pathname.endsWith('/summary') ? summary : records) });
    });

    await page.goto('http://fixture.invalid/v0/resource/plugins/usage-audit/dashboard');
    await page.waitForSelector('#authGate:not(.is-hidden)');
    assert.equal(dataCalls.length, 0, 'no data request without a key (would count toward the host IP ban)');
    assert.equal(await page.locator('#dashboardContent').isHidden(), true, 'dashboard hidden behind the gate');
    assert.equal(await page.locator('#authKeyInput').getAttribute('type'), 'password');

    // The gate shows no dashboard title/subtitle, and offers its own language switch.
    assert.equal(await page.locator('#pageHeading').isHidden(), true, 'page title and subtitle hidden behind the gate');
    assert.equal(await page.locator('#authLanguageToggle').isVisible(), true, 'language toggle inside the gate');
    assert.match(await page.locator('#authGateTitle').innerText(), /Management key required/);
    await page.click('#authLanguageToggle');
    assert.match(await page.locator('#authGateTitle').innerText(), /관리 키/, 'gate switches to Korean');
    assert.match(await page.locator('#authForm button[type=submit]').innerText(), /확인/);
    assert.equal(await page.evaluate(() => localStorage.getItem('cpa_plugin_lang')), 'ko', 'choice persisted as the shared language preference');
    await page.click('#authLanguageToggle');
    assert.match(await page.locator('#authGateTitle').innerText(), /Management key required/, 'and back to English');
    assert.equal(await page.locator('#authLanguageToggle').getAttribute('aria-label'), 'Switch to Korean');

    // Empty submit never leaves the page.
    await page.click('#authForm button[type=submit]');
    assert.match(await page.locator('#authMessage').innerText(), /Enter the management key/);
    assert.equal(dataCalls.length, 0);

    // Wrong key: exactly one rejected request (summary and records are serialised until accepted).
    await page.fill('#authKeyInput', 'wrong-key');
    await page.click('#authForm button[type=submit]');
    await page.waitForFunction(() => /invalid or access was denied/.test(document.getElementById('authMessage').textContent));
    assert.equal(dataCalls.length, 1, 'a wrong key costs a single failed attempt, got ' + dataCalls.length);
    assert.equal(await page.evaluate(k => sessionStorage.getItem(k), STORE), null, 'rejected key is never stored');
    assert.equal(await page.locator('#authGate').isVisible(), true);
    assert.equal(await page.evaluate(() => document.getElementById('authKeyInput').value), '', 'rejected key is cleared from the field');

    // Right key: dashboard renders, key kept for this tab only, sent only as a header.
    await page.fill('#authKeyInput', GOOD);
    await page.click('#authForm button[type=submit]');
    await page.waitForFunction(() => document.querySelector('#kpiTotalTokens') && document.querySelector('#kpiTotalTokens').textContent === '4,242');
    assert.equal(await page.locator('#pageHeading').isVisible(), true, 'title returns once the key is accepted');
    assert.equal(await page.locator('#authGate').isHidden(), true);
    assert.equal(await page.evaluate(k => sessionStorage.getItem(k), STORE), GOOD);
    assert.equal(await page.evaluate(k => localStorage.getItem(k), STORE), null, 'key must never reach localStorage');
    assert(urls.every(u => !u.includes(GOOD) && !/[?&](key|token|password|management_key)=/i.test(u)), 'key must not appear in any URL');
    assert(await page.evaluate(() => !location.href.includes('good-management-key')));

    // Reload in the same tab: stored key is reused without prompting.
    const before = dataCalls.length;
    await page.reload();
    await page.waitForFunction(() => document.querySelector('#kpiTotalTokens') && document.querySelector('#kpiTotalTokens').textContent === '4,242');
    assert.equal(await page.locator('#authGate').isHidden(), true);
    assert(dataCalls.length > before && dataCalls.slice(before).every(a => a === 'Bearer ' + GOOD));

    // Key revoked while the tab is open: the next poll drops it and asks again.
    await page.evaluate(() => { mgmtKey = 'revoked'; });
    await page.evaluate(() => loadData(false, true));
    await page.waitForSelector('#authGate:not(.is-hidden)');
    assert.equal(await page.evaluate(k => sessionStorage.getItem(k), STORE), null, 'revoked key is dropped from storage');
    const pollsAfter = dataCalls.length;
    await page.waitForTimeout(400);
    assert.equal(dataCalls.length, pollsAfter, 'polling stops after a rejection');

    // New tab (new session): nothing is inherited, so no request and the gate shows.
    const ctx2 = await browser.newContext({ locale: 'ko-KR' });
    const page2 = await ctx2.newPage();
    let calls2 = 0;
    await page2.route('**/*', async route => {
      const req = route.request(), url = new URL(req.url());
      if (req.resourceType() === 'document') return route.fulfill({ contentType: 'text/html', body: html });
      if (url.pathname.includes('/data/')) calls2++;
      return route.fulfill({ status: 404, body: '' });
    });
    await page2.goto('http://fixture.invalid/v0/resource/plugins/usage-audit/dashboard');
    await page2.waitForSelector('#authGate:not(.is-hidden)');
    assert.match(await page2.locator('#authGateTitle').innerText(), /관리 키/);
    assert.equal(calls2, 0);
    console.log('PASS auth gate: no key => no request; wrong key => 1 attempt, not stored; good key => header only, session only; revoked key => dropped and polling stops');
  } finally {
    await browser.close();
  }
})().catch(err => { console.error(err); process.exit(1); });
