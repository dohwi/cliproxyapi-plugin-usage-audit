#!/usr/bin/env node
// Browser E2E regression for request detail modal routing/auth-file mapping & edge cases:
// - OAuth file match (email + distinct account + label + account_type + disabled + numeric/string "0" auth_index)
// - Deleted / unmatched auth_index fallback
// - Runtime-only inline API key credential
// - Legacy record with no auth_index / auth_type
// - Prototype-pollution-safe auth_index ("__proto__", "constructor")
// - Bare-array /v0/management/auth-files payload compatibility
// - Non-fatal 404/500 auth-files failure (no error banner, clean fallback)
// - Delayed auth-files response updates open modal without losing focus, safe when closed
// - Never probes /v0/management/auth-files before mgmtKey is verified (IP-ban protection)
// - Clears cached auth metadata on 401/403 (dropKeyAndPrompt)
// - Never exposes raw host path or id_token fields from auth-files
const fs = require('fs');
const path = require('path');
const assert = require('node:assert/strict');
const { chromium } = require(process.env.PLAYWRIGHT_MODULE || 'playwright');
const { seedKey } = require('./lib/seed-key.cjs');

const root = path.resolve(__dirname, '..');
const html = require('./lib/dashboard-html.cjs').readDashboardHTML(root);
const now = new Date().toISOString();
const localDayKey = () => {
  const d = new Date();
  return d.getFullYear() + '-' + String(d.getMonth() + 1).padStart(2, '0') + '-' + String(d.getDate()).padStart(2, '0');
};
const summaryFixture = {
  total_count: 6,
  summary: [{ provider: 'codex', model: 'gpt-5', requests: 6, failures: 0, generate_count: 6, input_tokens: 600, fresh_input_tokens: 500, output_tokens: 300, cached_tokens: 100, reasoning_tokens: 50, total_tokens: 900 }],
  by_provider: [{ key: 'codex', requests: 6, failures: 0, input_tokens: 600, fresh_input_tokens: 500, output_tokens: 300, cached_tokens: 100, reasoning_tokens: 50, total_tokens: 900 }],
  by_endpoint: [{ key: 'codex', requests: 6, failures: 0, input_tokens: 600, fresh_input_tokens: 500, output_tokens: 300, cached_tokens: 100, reasoning_tokens: 50, total_tokens: 900 }],
  by_model: [{ key: 'gpt-5', requests: 6, failures: 0, input_tokens: 600, fresh_input_tokens: 500, output_tokens: 300, cached_tokens: 100, reasoning_tokens: 50, total_tokens: 900 }],
  daily: [{ key: localDayKey(), requests: 6, total_tokens: 900, input_tokens: 600, output_tokens: 300, failures: 0 }],
};
const recordsFixture = {
  has_more: false,
  total_count: 7,
  records: [
    // 0: falsy-looking "0" auth_index, default filename embedding email -> Account row auto-omitted (dedup), distinct account sub kept only if not in name
    { requested_at: now, provider: 'codex', model: 'gpt-5', executor_type: 'codex', auth_index: '0', auth_type: 'oauth', input: 100, fresh_input: 80, cached: 20, output: 50, reasoning: 10, total: 150, latency: 420, ttft: 110, failed: false },
    // 1: runtime-only inline API key
    { requested_at: now, provider: 'claude', model: 'claude-opus-5-5', executor_type: 'claude', auth_index: 'idx-inline-1', auth_type: 'api_key', input: 100, fresh_input: 100, cached: 0, output: 50, reasoning: 0, total: 150, latency: 380, ttft: 90, failed: false },
    // 2: deleted / unmatched authfile
    { requested_at: now, provider: 'antigravity', model: 'gemini-3', executor_type: 'antigravity', auth_index: 'idx-deleted-99', auth_type: 'oauth', input: 100, fresh_input: 100, cached: 0, output: 50, reasoning: 0, total: 150, latency: 300, ttft: 80, failed: false },
    // 3: legacy record with no auth_index or auth_type
    { requested_at: now, provider: 'openai', model: 'gpt-4.1', executor_type: 'openai', input: 100, fresh_input: 100, cached: 0, output: 50, reasoning: 0, total: 150, latency: 250, ttft: 70, failed: false },
    // 4: prototype property name as auth_index (custom filename NOT containing email -> Account row shown)
    { requested_at: now, provider: 'codex', model: 'gpt-5', executor_type: 'codex', auth_index: '__proto__', auth_type: 'oauth', input: 100, fresh_input: 100, cached: 0, output: 50, reasoning: 0, total: 150, latency: 200, ttft: 60, failed: false },
    // 5: "constructor" auth_index not in auth-files
    { requested_at: now, provider: 'codex', model: 'gpt-5', executor_type: 'codex', auth_index: 'constructor', auth_type: 'oauth', input: 100, fresh_input: 100, cached: 0, output: 50, reasoning: 0, total: 150, latency: 210, ttft: 65, failed: false },
    // 6: sanitized filename (xai-user_example_invalid-deadbeef.json) embedding email via underscores -> Account row auto-omitted
    { requested_at: now, provider: 'xai', model: 'grok-4', executor_type: 'xai', auth_index: 'idx-xai-sanitized', auth_type: 'oauth', input: 100, fresh_input: 100, cached: 0, output: 50, reasoning: 0, total: 150, latency: 220, ttft: 70, failed: false },
  ],
};

function serveFont(route, url) {
  const asset = path.join(root, 'internal/plugin/assets/fonts', path.basename(url.pathname));
  if (fs.existsSync(asset) && fs.statSync(asset).isFile()) {
    return route.fulfill({ contentType: 'font/woff2', body: fs.readFileSync(asset) });
  }
  return null;
}

(async () => {
  const browser = await chromium.launch({ headless: true, executablePath: process.env.BROWSER_EXECUTABLE || undefined, args: ['--no-sandbox'] });
  try {
    // --- Scenario 1: Full mapping, numeric "0", runtime_only, deleted, legacy, prototype safety, secret stripping, delayed refresh ---
    {
      const ctx = await browser.newContext({ locale: 'ko-KR', reducedMotion: 'reduce', viewport: { width: 1280, height: 900 } });
      await seedKey(ctx);
      const page = await ctx.newPage();
      const errors = [];
      page.on('pageerror', e => errors.push(e.message));

      let releaseAuthFiles;
      const authFilesGate = new Promise(resolve => { releaseAuthFiles = resolve; });
      let authFilesCalls = 0;

      await page.route('**/*', async route => {
        const url = new URL(route.request().url());
        if (route.request().resourceType() === 'document') return route.fulfill({ contentType: 'text/html', body: html });
        const font = serveFont(route, url);
        if (font) return font;
        if (url.pathname.endsWith('/data/summary')) return route.fulfill({ contentType: 'application/json', body: JSON.stringify(summaryFixture) });
        if (url.pathname.endsWith('/data/records')) return route.fulfill({ contentType: 'application/json', body: JSON.stringify(recordsFixture) });
        if (url.pathname.endsWith('/management/auth-files')) {
          authFilesCalls++;
          assert.equal(route.request().headers()['authorization'], 'Bearer fixture-management-key');
          await authFilesGate;
          return route.fulfill({
            contentType: 'application/json',
            body: JSON.stringify({
              files: [
                {
                  auth_index: 0, // numeric 0 must match string "0"
                  name: 'codex-zero@example.invalid.json',
                  email: 'zero@example.invalid',
                  account: 'zero@example.invalid',
                  account_type: 'oauth',
                  label: 'codex', // duplicates provider -> must be suppressed
                  provider: 'codex',
                  disabled: true,
                  runtime_only: false,
                  path: '/secret/host/path/never-show.json',
                  id_token: { secret_claim: 'never-show-token' },
                },
                {
                  auth_index: 'idx-inline-1',
                  name: '',
                  account: 'sk-ant-raw-api-key-must-never-leak',
                  account_type: 'api_key',
                  label: 'Team-Key',
                  provider: 'claude',
                  disabled: false,
                  runtime_only: true,
                },
                {
                  auth_index: '__proto__',
                  name: 'work-main.json',
                  email: 'proto@example.invalid',
                  account: 'org-proto',
                  label: 'Work-Primary',
                  provider: 'codex',
                },
                {
                  auth_index: 'idx-xai-sanitized',
                  name: 'xai-sanitized_user_example_invalid-deadbeef.json',
                  email: 'sanitized.user@example.invalid',
                  account: 'sanitized.user@example.invalid',
                  account_type: 'oauth',
                  label: 'xai',
                  provider: 'xai',
                },
              ],
            }),
          });
        }
        return route.fulfill({ status: 404, body: '' });
      });

      await page.goto('http://fixture.invalid/v0/resource/plugins/usage-audit/dashboard');
      await page.waitForSelector('#recordsBody tr.ui-table-row');

      // Open row 0 BEFORE auth-files resolves: should show loading state, then auto-update without stealing focus.
      await page.click('#recordsBody tr.ui-table-row >> nth=0');
      await page.waitForSelector('#detailOverlay.open');
      const loadingText = await page.locator('#recordDialog').innerText();
      assert(loadingText.includes('인증 정보를 불러오는 중...'), 'expected loading text while auth-files in flight: ' + loadingText);
      assert(loadingText.includes('0'), 'auth_index "0" should appear even while loading');

      // Release auth-files response and wait for live modal refresh.
      releaseAuthFiles();
      await page.waitForFunction(() => document.getElementById('recordDialog').textContent.includes('codex-zero@example.invalid.json'));
      const row0Text = await page.locator('#recordDialog').innerText();
      assert(row0Text.includes('라우팅 및 인증'), 'routing section heading missing');
      assert(row0Text.includes('codex-zero@example.invalid.json'), 'auth file name missing: ' + row0Text);
      assert(row0Text.includes('현재 비활성화됨'), 'disabled sub missing on auth file row: ' + row0Text);
      assert(row0Text.includes('OAuth'), 'OAuth auth type missing: ' + row0Text);
      assert(!row0Text.includes('/secret/host/path'), 'host filesystem path must never be exposed');
      assert(!row0Text.includes('never-show-token'), 'id_token must never be exposed');
      // Default filename already embeds email & label duplicates provider -> Account and Label rows must be omitted!
      const row0Keys = await page.locator('#recordDialog .detail-section-full .detail-key').allInnerTexts();
      assert.deepEqual(row0Keys, ['인증 파일', '인증 방식', '인증 인덱스'], 'redundant Account/Label rows should be omitted for default filename: ' + JSON.stringify(row0Keys));

      // Verify whitelisted cache keys only (no path / id_token retained in memory).
      const leakedInState = await page.evaluate(() => {
        const entry = state.authFilesByIndex['0'];
        return entry ? ('path' in entry || 'id_token' in entry) : true;
      });
      assert.equal(leakedInState, false, 'state.authFilesByIndex must only store whitelisted non-secret fields');

      // Row 1: runtime-only inline API key (must NEVER leak raw `account` API key from CLIProxyAPI Auth.AccountInfo)
      await page.keyboard.press('Escape');
      await page.click('#recordsBody tr.ui-table-row >> nth=1');
      await page.waitForSelector('#detailOverlay.open');
      const row1Text = await page.locator('#recordDialog').innerText();
      assert(row1Text.includes('런타임 전용(파일 없음)'), 'runtime_only text missing: ' + row1Text);
      assert(row1Text.includes('Team-Key'), 'inline label missing: ' + row1Text);
      assert(row1Text.includes('API 키'), 'API key label missing: ' + row1Text);
      assert(row1Text.includes('idx-inline-1'), 'auth_index missing: ' + row1Text);
      assert(!row1Text.includes('sk-ant-raw-api-key-must-never-leak'), 'raw api_key account must NEVER appear in DOM');
      const cachedInlineAccount = await page.evaluate(() => state.authFilesByIndex['idx-inline-1'] ? state.authFilesByIndex['idx-inline-1'].account : 'MISSING');
      assert.equal(cachedInlineAccount, '', 'raw api_key account must NEVER be cached in state.authFilesByIndex');

      // Background state.records reorder while modal is open must NOT swap open modal to another record on refresh.
      await page.evaluate(() => {
        state.records = [state.records[3], state.records[0]];
        refreshOpenRecordDetail();
      });
      const stillRow1Text = await page.locator('#recordDialog').innerText();
      assert(stillRow1Text.includes('claude-opus-5-5') && stillRow1Text.includes('idx-inline-1'), 'open modal must keep original record when state.records changes: ' + stillRow1Text);
      await page.evaluate(recs => { state.records = recs; }, recordsFixture.records);

      // Row 2: deleted / unmatched authfile
      await page.keyboard.press('Escape');
      await page.click('#recordsBody tr.ui-table-row >> nth=2');
      await page.waitForSelector('#detailOverlay.open');
      const row2Text = await page.locator('#recordDialog').innerText();
      assert(row2Text.includes('현재 활성 인증 파일에 없음'), 'unmatched fallback missing: ' + row2Text);
      assert(row2Text.includes('idx-deleted-99'), 'deleted auth_index fallback missing: ' + row2Text);

      // Row 3: legacy record with no auth_index / auth_type
      await page.keyboard.press('Escape');
      await page.click('#recordsBody tr.ui-table-row >> nth=3');
      await page.waitForSelector('#detailOverlay.open');
      const row3Text = await page.locator('#recordDialog').innerText();
      assert(row3Text.includes('기록된 인증 라우팅 정보가 없습니다.'), 'legacy unavailable notice missing: ' + row3Text);

      // Row 4 & 5 & 6: prototype safety + custom filename shows Account + sanitized filename omits Account
      await page.keyboard.press('Escape');
      await page.click('#recordsBody tr.ui-table-row >> nth=4');
      await page.waitForSelector('#detailOverlay.open');
      const row4Text = await page.locator('#recordDialog').innerText();
      const row4Keys = await page.locator('#recordDialog .detail-section-full .detail-key').allInnerTexts();
      assert(row4Text.includes('work-main.json') && row4Text.includes('proto@example.invalid') && row4Text.includes('org-proto') && row4Text.includes('Work-Primary'), 'custom filename should show Account & Label: ' + row4Text);
      assert.deepEqual(row4Keys, ['인증 파일', '계정', '라벨', '인증 방식', '인증 인덱스'], 'custom filename row keys mismatch: ' + JSON.stringify(row4Keys));

      await page.keyboard.press('Escape');
      await page.click('#recordsBody tr.ui-table-row >> nth=5');
      await page.waitForSelector('#detailOverlay.open');
      const row5Text = await page.locator('#recordDialog').innerText();
      const row5Keys = await page.locator('#recordDialog .detail-section-full .detail-key').allInnerTexts();
      assert(row5Text.includes('현재 활성 인증 파일에 없음'), 'constructor must not collide with Object.prototype: ' + row5Text);
      assert.deepEqual(row5Keys, ['인증 파일', '인증 방식', '인증 인덱스'], 'unmatched fallback must not repeat status twice: ' + JSON.stringify(row5Keys));

      await page.keyboard.press('Escape');
      await page.click('#recordsBody tr.ui-table-row >> nth=6');
      await page.waitForSelector('#detailOverlay.open');
      const row6Keys = await page.locator('#recordDialog .detail-section-full .detail-key').allInnerTexts();
      assert.deepEqual(row6Keys, ['인증 파일', '인증 방식', '인증 인덱스'], 'sanitized filename embedding email should omit Account row: ' + JSON.stringify(row6Keys));

      assert.equal(authFilesCalls, 1, 'auth-files should be fetched once within TTL');
      assert.deepEqual(errors, [], 'no page errors expected');
      await ctx.close();
    }

    // --- Scenario 2: Bare-array auth-files payload & 401 cache wipe ---
    {
      const ctx = await browser.newContext({ locale: 'en-US', reducedMotion: 'reduce', viewport: { width: 1280, height: 900 } });
      await seedKey(ctx);
      const page = await ctx.newPage();
      let failAuthFiles401 = false;

      await page.route('**/*', async route => {
        const url = new URL(route.request().url());
        if (route.request().resourceType() === 'document') return route.fulfill({ contentType: 'text/html', body: html });
        const font = serveFont(route, url);
        if (font) return font;
        if (url.pathname.endsWith('/data/summary')) return route.fulfill({ contentType: 'application/json', body: JSON.stringify(summaryFixture) });
        if (url.pathname.endsWith('/data/records')) return route.fulfill({ contentType: 'application/json', body: JSON.stringify(recordsFixture) });
        if (url.pathname.endsWith('/management/auth-files')) {
          if (failAuthFiles401) return route.fulfill({ status: 401, contentType: 'application/json', body: '{"error":"unauthorized"}' });
          return route.fulfill({
            contentType: 'application/json',
            body: JSON.stringify([{ auth_index: '0', name: 'bare-array.json', account: 'bare@example.invalid' }]),
          });
        }
        return route.fulfill({ status: 404, body: '' });
      });

      await page.goto('http://fixture.invalid/v0/resource/plugins/usage-audit/dashboard');
      await page.waitForFunction(() => !!state.authFilesByIndex['0']);
      await page.click('#recordsBody tr.ui-table-row >> nth=0');
      await page.waitForSelector('#detailOverlay.open');
      const dlgText = await page.locator('#recordDialog').innerText();
      assert(dlgText.includes('bare-array.json') && dlgText.includes('bare@example.invalid'), 'bare array auth-files should parse: ' + dlgText);

      // Force refresh where auth-files returns 401 -> must prompt auth gate and wipe auth cache.
      failAuthFiles401 = true;
      await page.keyboard.press('Escape');
      await page.click('#refreshBtn');
      await page.waitForSelector('#authGate:not(.is-hidden)');
      const cachedCount = await page.evaluate(() => Object.keys(state.authFilesByIndex).length);
      assert.equal(cachedCount, 0, '401 must wipe cached auth-files metadata');
      await ctx.close();
    }

    // --- Scenario 3: Non-fatal 404 on auth-files (older host / missing route) ---
    {
      const ctx = await browser.newContext({ locale: 'en-US', reducedMotion: 'reduce', viewport: { width: 1280, height: 900 } });
      await seedKey(ctx);
      const page = await ctx.newPage();
      const errors = [];
      page.on('pageerror', e => errors.push(e.message));

      await page.route('**/*', async route => {
        const url = new URL(route.request().url());
        if (route.request().resourceType() === 'document') return route.fulfill({ contentType: 'text/html', body: html });
        const font = serveFont(route, url);
        if (font) return font;
        if (url.pathname.endsWith('/data/summary')) return route.fulfill({ contentType: 'application/json', body: JSON.stringify(summaryFixture) });
        if (url.pathname.endsWith('/data/records')) return route.fulfill({ contentType: 'application/json', body: JSON.stringify(recordsFixture) });
        if (url.pathname.endsWith('/management/auth-files')) return route.fulfill({ status: 404, body: 'not found' });
        return route.fulfill({ status: 404, body: '' });
      });

      await page.goto('http://fixture.invalid/v0/resource/plugins/usage-audit/dashboard');
      await page.waitForFunction(() => state.hasLoadedData && !state.authFilesInFlight);
      const errHidden = await page.locator('#errorBox').evaluate(el => el.classList.contains('is-hidden'));
      assert.equal(errHidden, true, 'auth-files 404 must never surface an error banner');
      await page.click('#recordsBody tr.ui-table-row >> nth=0');
      await page.waitForSelector('#detailOverlay.open');
      const dlgText = await page.locator('#recordDialog').innerText();
      assert(dlgText.includes('Could not load auth files') && dlgText.includes('OAuth'), '404 fallback should distinguish fetch failure + auth_type: ' + dlgText);
      assert.deepEqual(errors, []);
      await ctx.close();
    }

    // --- Scenario 4: Wrong key must NEVER call /v0/management/auth-files (5-strike IP ban protection) ---
    {
      const ctx = await browser.newContext({ locale: 'en-US', reducedMotion: 'reduce', viewport: { width: 1280, height: 900 } });
      const page = await ctx.newPage();
      let authFilesCalls = 0;
      let summaryCalls = 0;

      await page.route('**/*', async route => {
        const url = new URL(route.request().url());
        if (route.request().resourceType() === 'document') return route.fulfill({ contentType: 'text/html', body: html });
        const font = serveFont(route, url);
        if (font) return font;
        if (url.pathname.endsWith('/data/summary')) {
          summaryCalls++;
          return route.fulfill({ status: 401, contentType: 'application/json', body: '{"error":"invalid key"}' });
        }
        if (url.pathname.endsWith('/management/auth-files')) {
          authFilesCalls++;
          return route.fulfill({ status: 401, contentType: 'application/json', body: '{"error":"invalid key"}' });
        }
        return route.fulfill({ status: 404, body: '' });
      });

      await page.goto('http://fixture.invalid/v0/resource/plugins/usage-audit/dashboard');
      await page.waitForSelector('#authGate:not(.is-hidden)');
      await page.fill('#authKeyInput', 'wrong-key');
      await page.click('#authForm button[type="submit"]');
      await page.waitForSelector('#authMessage:not(.is-hidden)');
      assert.equal(summaryCalls, 1, 'wrong key should cost only 1 request');
      assert.equal(authFilesCalls, 0, 'unverified key must NEVER probe /v0/management/auth-files');
      await ctx.close();
    }

    // --- Scenario 5: Live language switching (ko / en / zh-CN) inside open dialog + mobile 412px bottom sheet geometry ---
    {
      const ctx = await browser.newContext({ locale: 'ko-KR', reducedMotion: 'reduce', viewport: { width: 412, height: 900 } });
      await seedKey(ctx);
      const page = await ctx.newPage();
      const errors = [];
      page.on('pageerror', e => errors.push(e.message));

      await page.route('**/*', async route => {
        const url = new URL(route.request().url());
        if (route.request().resourceType() === 'document') return route.fulfill({ contentType: 'text/html', body: html });
        const font = serveFont(route, url);
        if (font) return font;
        if (url.pathname.endsWith('/data/summary')) return route.fulfill({ contentType: 'application/json', body: JSON.stringify(summaryFixture) });
        if (url.pathname.endsWith('/data/records')) return route.fulfill({ contentType: 'application/json', body: JSON.stringify(recordsFixture) });
        if (url.pathname.endsWith('/management/auth-files')) {
          return route.fulfill({
            contentType: 'application/json',
            body: JSON.stringify({
              files: [{ auth_index: '0', name: 'codex-zero@example.invalid.json', email: 'zero@example.invalid', account_type: 'pro', label: 'Primary', disabled: true }],
            }),
          });
        }
        return route.fulfill({ status: 404, body: '' });
      });

      await page.goto('http://fixture.invalid/v0/resource/plugins/usage-audit/dashboard');
      await page.waitForFunction(() => !!state.authFilesByIndex['0']);
      await page.click('#recordsBody tr.ui-table-row >> nth=0');
      await page.waitForSelector('#detailOverlay.open');

      // Verify mobile bottom sheet geometry (top rounded 16px, bottom 0px, flush full width, no horizontal overflow).
      const sheet = await page.evaluate(() => {
        const overlay = document.getElementById('detailOverlay');
        const dlg = document.getElementById('recordDialog');
        const cs = getComputedStyle(dlg);
        const rect = dlg.getBoundingClientRect();
        const ovRect = overlay.getBoundingClientRect();
        return {
          width: rect.width,
          overlayWidth: ovRect.width,
          bottomGap: Math.abs(ovRect.bottom - rect.bottom),
          tl: cs.borderTopLeftRadius,
          tr: cs.borderTopRightRadius,
          bl: cs.borderBottomLeftRadius,
          br: cs.borderBottomRightRadius,
          overflowX: dlg.scrollWidth > dlg.clientWidth + 1,
        };
      });
      assert(Math.abs(sheet.width - sheet.overlayWidth) <= 1, 'mobile sheet should span overlay width: ' + JSON.stringify(sheet));
      assert(sheet.bottomGap <= 1, 'mobile sheet should sit flush at bottom: ' + JSON.stringify(sheet));
      assert.equal(sheet.tl, '16px', 'mobile sheet top-left radius');
      assert.equal(sheet.tr, '16px', 'mobile sheet top-right radius');
      assert.equal(sheet.bl, '0px', 'mobile sheet bottom-left radius');
      assert.equal(sheet.br, '0px', 'mobile sheet bottom-right radius');
      assert.equal(sheet.overflowX, false, 'mobile sheet must not overflow horizontally');

      // Live language switch while dialog is open: ko -> en -> zh-CN -> ko
      const hangul = s => /[가-힣]/.test(s);
      const han = s => /[\u4e00-\u9fff]/.test(s);

      await page.evaluate(() => selectLanguagePref('en'));
      const enDlg = await page.locator('#recordDialog').innerText();
      assert(enDlg.includes('Routing & auth') && enDlg.includes('Auth file') && enDlg.includes('Currently disabled'), 'en routing labels missing: ' + enDlg);
      assert(!hangul(enDlg) && !han(enDlg), 'en dialog must not contain Korean/Chinese: ' + enDlg);

      await page.evaluate(() => selectLanguagePref('zh-CN'));
      const zhDlg = await page.locator('#recordDialog').innerText();
      assert(zhDlg.includes('路由与认证') && zhDlg.includes('认证文件') && zhDlg.includes('当前已禁用'), 'zh-CN routing labels missing: ' + zhDlg);
      assert(!hangul(zhDlg), 'zh-CN dialog must not contain Korean: ' + zhDlg);

      await page.evaluate(() => selectLanguagePref('ko'));
      const koDlg = await page.locator('#recordDialog').innerText();
      assert(koDlg.includes('라우팅 및 인증') && koDlg.includes('인증 파일') && koDlg.includes('현재 비활성화됨'), 'ko routing labels missing: ' + koDlg);
      assert(!han(koDlg), 'ko dialog must not contain Chinese: ' + koDlg);

      assert.deepEqual(errors, []);
      await ctx.close();
    }

    console.log('PASS auth-detail E2E (5 scenarios)');
  } finally {
    await browser.close();
  }
})().catch(err => { console.error(err); process.exit(1); });
