#!/usr/bin/env node
// Browser regression for the English/Korean switch across calendar, dialog, tables and filters,
// including switching while the dialog is open, shared-storage sync and persistence.
// Install Playwright or set PLAYWRIGHT_MODULE; BROWSER_EXECUTABLE selects a system browser.
const fs = require('fs'), path = require('path');
const switchLang = (async (page, lang) => { if ((await page.evaluate(() => document.documentElement.lang)) !== lang) await page.click('#languageToggle'); await page.waitForFunction(l => document.documentElement.lang === l, lang); });
const { chromium } = require(process.env.PLAYWRIGHT_MODULE || 'playwright');
const { seedKey } = require('./lib/seed-key.cjs');
const root = path.resolve(__dirname, '..');
const html = require('./lib/dashboard-html.cjs').readDashboardHTML(root);
const verifyLanguagePlacement = require('./lib/host-toolbar.cjs');
const now = new Date().toISOString();
const metric = (provider, model) => ({ provider, model, requests: 120, failures: 3, generate_count: 100, input_tokens: 9000, fresh_input_tokens: 7000, output_tokens: 3000, cached_tokens: 2000, reasoning_tokens: 500, total_tokens: 12500 });
const summary = {
  total_count: 240,
  summary: [metric('anthropic', 'claude-opus'), metric('openai', 'gpt-5')],
  by_provider: [{ ...metric('anthropic', 'x'), key: 'anthropic' }, { ...metric('openai', 'x'), key: 'openai' }],
  by_endpoint: [{ ...metric('a', 'x'), key: 'direct' }], by_model: [{ ...metric('a', 'x'), key: 'claude-opus' }],
  daily: [0, 1, 2].map(i => ({ key: new Date(Date.now() - i * 86400000).toISOString().slice(0, 10), requests: 10 + i, total_tokens: 100, input_tokens: 60, output_tokens: 40, failures: i })),
};
const records = { has_more: true, next_cursor: 'c1', records: Array.from({ length: 3 }, (_, i) => ({ id: 'r' + i, requested_at: now, provider: 'anthropic', model: 'claude-opus', endpoint: 'direct', fresh_input: 10, cached: 5, output: 20, total: 35, input: 15, reasoning: 1, latency_ms: 1200, ttft_ms: 300, failed: i === 1, failure_status: i === 1 ? '500' : '' })) };
const hangul = s => /[가-힣]/.test(s);
(async () => {
  const browser = await chromium.launch({ headless: true, executablePath: process.env.BROWSER_EXECUTABLE || undefined, args: ['--no-sandbox'] });
  const fail = [];
  try {
    for (const locale of ['en-US', 'ko-KR']) {
      const ctx = await browser.newContext({ locale, reducedMotion: 'reduce', viewport: { width: 1280, height: 900 } });
      await seedKey(ctx);
      const page = await ctx.newPage();
      const errors = [];
      page.on('pageerror', e => errors.push(e.message));
      await page.route('**/*', async route => {
        const url = new URL(route.request().url());
        if (route.request().resourceType() === 'document') return route.fulfill({ contentType: 'text/html', body: html });
        if (url.pathname.endsWith('/data/summary')) return route.fulfill({ contentType: 'application/json', body: JSON.stringify(summary) });
        if (url.pathname.endsWith('/data/records')) return route.fulfill({ contentType: 'application/json', body: JSON.stringify(records) });
        const asset = path.join(root, 'internal/plugin/assets/fonts', path.basename(url.pathname));
        if (fs.existsSync(asset) && fs.statSync(asset).isFile()) return route.fulfill({ contentType: 'font/woff2', body: fs.readFileSync(asset) });
        return route.fulfill({ status: 404, body: '' });
      });
      await page.goto('http://fixture.invalid/v0/resource/plugins/usage-audit/dashboard');
      await page.waitForSelector('#recordsBody tr.ui-table-row');
      const base = locale === 'en-US' ? 'en' : 'ko', other = base === 'en' ? 'ko' : 'en';
      const initial = await page.evaluate(() => document.documentElement.lang);
      if (initial !== base) fail.push(`${locale}: initial lang ${initial}`);
      const visible = () => page.evaluate(() => {
        const clone = document.body.cloneNode(true);
        clone.querySelectorAll('[aria-hidden="true"], script, style').forEach(n => n.remove());
        const bits = [document.title, clone.textContent.replace(/\s*\n\s*/g, '\n')];
        document.querySelectorAll('[aria-label],[title]').forEach(e => bits.push((e.getAttribute('aria-label') || '') + ' ' + (e.getAttribute('title') || '')));
        return bits.join('\n');
      });

      for (const lang of [other, base]) {
        await switchLang(page, lang);
        // open a few interactive surfaces so their text is covered too
        await page.click('#presetGroupControl [data-preset="custom"]');
        await page.click('#filterDateRangeTrigger');
        await page.waitForSelector('.ui-date-picker-content');
        const cal = await page.locator('.ui-date-picker-content').innerText();
        const calAria = await page.locator('.ui-calendar-nav').first().getAttribute('aria-label');
        await page.keyboard.press('Escape');
        await page.click('#presetGroupControl [data-preset="today"]');
        await page.waitForSelector('#recordsBody tr.ui-table-row');
        await page.click('#recordsBody tr.ui-table-row >> nth=1');
        await page.waitForSelector('#detailOverlay.open');
        const dialog = await page.evaluate(() => document.getElementById('recordDialog').textContent.replace(/\s+/g, ' '));
        // switching language while the dialog is open re-renders it
        await page.keyboard.press('Escape');
        await page.click('#filterStatusTrigger'); await page.click('[role=option] >> nth=2');
        await page.waitForFunction(() => !state.inFlight);
        const text = (await visible()) + '\n' + cal + '\n' + dialog + '\n' + calAria;
        const checks = {
          lang: await page.evaluate(() => document.documentElement.lang),
          heading: await page.locator('h1').innerText(),
          th: await page.locator('#recordsTable thead').innerText(),
          badge: await page.locator('#activeFilterText').innerText(),
          count: await page.locator('#recordsCountBadge').innerText(),
          fail: await page.locator('#kpiSubFail').innerText(),
          chart: await page.locator('#dailyChartSummary').innerText(),
          time: await page.locator('#recordsBody tr >> nth=0 >> td >> nth=0').innerText(),
          pressed: await page.locator('#languageToggle').getAttribute('data-current-lang'),
        };
        if (checks.lang !== lang) fail.push(`${locale}/${lang}: html lang ${checks.lang}`);
        if (checks.pressed !== lang) fail.push(`${locale}/${lang}: pressed ${checks.pressed}`);
        if (lang === 'en') {
          const stray = text.split('\n').filter(l => hangul(l) && !/한국어|^\s*가(\s+A)?\s*$/.test(l));
          if (stray.length) fail.push(`${locale}/en: Korean left: ${JSON.stringify(stray.slice(0, 8))} | parts: ${JSON.stringify({vis: (await visible()).split('\n').filter(l => hangul(l) && !/한국어|^\s*가(\s+A)?\s*$/.test(l)).slice(0,4), cal: cal.split('\n').filter(hangul).slice(0,4), dlg: dialog.split(' ').filter(hangul).slice(0,4), calAria: hangul(calAria), where: await page.evaluate(() => { const o = []; document.querySelectorAll('*').forEach(e => { const own = [...e.childNodes].filter(n => n.nodeType === 3).map(n => n.textContent).join(''); if (/전체/.test(own) && e.tagName !== 'SCRIPT') o.push(e.tagName + '#' + e.id + '.' + String(e.className).slice(0, 25) + ' vis=' + !!(e.offsetWidth || e.offsetHeight) + ' parent=' + (e.parentElement && e.parentElement.id)); }); return o; })})}`);
          if (checks.heading !== 'Token Usage') fail.push(`en heading ${checks.heading}`);
          if (!/New input/.test(checks.th) || !/Latency/.test(checks.th)) fail.push('en table headers: ' + checks.th);
          if (!/Failed only/.test(checks.badge)) fail.push('en badge: ' + checks.badge);
          if (!/records/.test(checks.count)) fail.push('en count: ' + checks.count);
          if (!/Failed: /.test(checks.fail)) fail.push('en fail kpi: ' + checks.fail);
        } else {
          if (checks.heading !== '토큰 사용량') fail.push(`ko heading ${checks.heading}`);
          if (!/신규 입력/.test(checks.th) || !/지연/.test(checks.th)) fail.push('ko headers: ' + checks.th);
          if (!/실패만/.test(checks.badge)) fail.push('ko badge: ' + checks.badge);
          const english = text.split('\n').filter(l => /\b(Total tokens|Request details|Load more|Previous month|Failed only)\b/.test(l));
          if (english.length) fail.push(`${locale}/ko: English left: ${JSON.stringify(english.slice(0, 5))}`);
        }
        for (const width of [1440, 768, 390, 320]) {
          await page.setViewportSize({ width, height: 900 });
          const over = await page.evaluate(() => {
            const out = [];
            if (document.documentElement.scrollWidth > innerWidth + 1) out.push('page-h-scroll');
            document.querySelectorAll('.badge, #recordsTable .ui-table-head, #recordsTable .ui-table-cell, .kpi-label, .kpi-sub, .tab-btn, .preset-btn, .btn').forEach(e => { if (e.offsetWidth && e.scrollWidth > e.clientWidth + 1 && getComputedStyle(e).textOverflow !== 'ellipsis') out.push('CLIPPED ' + e.tagName + '.' + String(e.className).slice(0, 24) + ' "' + e.textContent.trim().slice(0, 20) + '" ' + e.scrollWidth + '>' + e.clientWidth); });
            document.querySelectorAll('.page-heading, .lang-toggle, .filter-card, .kpi-group, #activeFilterBadge, .card-header').forEach(e => { if (e.scrollWidth > e.clientWidth + 1) out.push(e.className.toString().slice(0, 30) + ' ' + e.scrollWidth + '>' + e.clientWidth); });
            return out;
          });
          await verifyLanguagePlacement(page, '.filter-card');
          if (over.length) fail.push(`${locale}/${lang}/${width}px overflow: ${over.join(' | ')}`);
        }
        await page.setViewportSize({ width: 1280, height: 900 });
        if (process.env.SCREENSHOT_DIR) await page.screenshot({ path: path.join(process.env.SCREENSHOT_DIR, `lang-${locale}-${lang}.png`), fullPage: true });
        // reset the status filter for the next loop
        await page.click('#resetFiltersBtn');
        await page.waitForFunction(() => !state.inFlight);
      }

      // open dialog, switch language: dialog contents must follow
      await page.click('#recordsBody tr.ui-table-row >> nth=0');
      await page.waitForSelector('#detailOverlay.open');
      // the modal overlay covers the toggle, so drive the same API the button calls
      await page.evaluate(() => setLanguage('en'));
      const dlgEn = await page.evaluate(() => document.getElementById('recordDialog').textContent.replace(/\s+/g, ' '));
      const dbg = await page.evaluate(() => ({ idx: state.openRecordIdx, lang: currentLang, rows: document.querySelectorAll('#dialogBody .detail-row').length, open: document.getElementById('detailOverlay').className, recs: state.records.length, hasLoaded: state.hasLoadedData, inFlight: state.inFlight }));
      if (!/Request details/i.test(dlgEn) || !/Total tokens/i.test(dlgEn) || hangul(dlgEn)) fail.push(`${locale}: open dialog did not follow language: ${JSON.stringify(dbg)} text=${JSON.stringify(dlgEn.slice(0, 300))}`);
      await page.keyboard.press('Escape');

      // storage-event sync from another dashboard
      await page.evaluate(() => { localStorage.setItem('cpa_plugin_lang', 'ko'); window.dispatchEvent(new StorageEvent('storage', { key: 'cpa_plugin_lang', newValue: 'ko' })); });
      if ((await page.evaluate(() => document.documentElement.lang)) !== 'ko') fail.push(`${locale}: storage event not followed`);

      // persistence
      await switchLang(page, 'en');
      await page.reload();
      await page.waitForSelector('#recordsBody tr.ui-table-row');
      const after = await page.evaluate(() => [document.documentElement.lang, localStorage.getItem('cpa_plugin_lang')]);
      if (after[0] !== 'en' || after[1] !== 'en') fail.push(`${locale}: not persisted ${after}`);
      if (errors.length) fail.push(`${locale}: page errors ${errors.slice(0, 3)}`);
      await ctx.close();
    }
  } finally { await browser.close(); }
  if (fail.length) { console.log('FAIL\n' + fail.join('\n')); process.exit(1); }
  console.log('PASS: detect, toggle, calendar/dialog/table/filters localized, no stray text, dialog follows switch, storage sync, persistence, no overflow');
})();
