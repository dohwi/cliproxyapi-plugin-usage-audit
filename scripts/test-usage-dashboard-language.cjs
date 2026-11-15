#!/usr/bin/env node
// Browser regression for the English/Korean/Simplified Chinese menu across calendar, dialog, tables and
// filters, including switching while the dialog is open, shared-storage sync, persistence and no clipping.
// Install Playwright or set PLAYWRIGHT_MODULE; BROWSER_EXECUTABLE selects a system browser.
const fs = require('fs'), path = require('path');
const { chromium } = require(process.env.PLAYWRIGHT_MODULE || 'playwright');
const { seedKey } = require('./lib/seed-key.cjs');
const root = path.resolve(__dirname, '..');
const html = require('./lib/dashboard-html.cjs').readDashboardHTML(root);
const verifyLanguagePlacement = require('./lib/host-toolbar.cjs');
const { switchTo: switchLang, REFRESH } = verifyLanguagePlacement;
const now = new Date().toISOString();
// The server buckets days in the browser's time zone (time_zone query), so the fixture must
// too: UTC keys put no bar on "today" whenever the local date is ahead of UTC.
const localDayKey = i => { const d = new Date(); d.setDate(d.getDate() - i); return d.getFullYear() + '-' + String(d.getMonth() + 1).padStart(2, '0') + '-' + String(d.getDate()).padStart(2, '0'); };
const metric = (provider, model) => ({ provider, model, requests: 120, failures: 3, generate_count: 100, input_tokens: 9000, fresh_input_tokens: 7000, output_tokens: 3000, cached_tokens: 2000, reasoning_tokens: 500, total_tokens: 12500 });
const summary = {
  total_count: 240,
  summary: [metric('anthropic', 'claude-opus'), metric('openai', 'gpt-5')],
  by_provider: [{ ...metric('anthropic', 'x'), key: 'anthropic' }, { ...metric('openai', 'x'), key: 'openai' }],
  by_endpoint: [{ ...metric('a', 'x'), key: 'direct' }], by_model: [{ ...metric('a', 'x'), key: 'claude-opus' }],
  daily: [0, 1, 2].map(i => ({ key: localDayKey(i), requests: 10 + i, total_tokens: 100, input_tokens: 60, output_tokens: 40, failures: i })),
};
const records = { has_more: true, next_cursor: 'c1', records: Array.from({ length: 3 }, (_, i) => ({ id: 'r' + i, requested_at: now, provider: 'anthropic', model: 'claude-opus', endpoint: 'direct', fresh_input: 10, cached: 5, output: 20, total: 35, input: 15, reasoning: 1, latency_ms: 1200, ttft_ms: 300, failed: i === 1, failure_status: i === 1 ? '500' : '' })) };
const hangul = s => /[가-힣]/.test(s);
const han = s => /[\u4e00-\u9fff]/.test(s);
// Native language names in the menu/labels are never translated.
const NATIVE = /한국어|简体中文|^\s*가(\s+A)?\s*$/;
(async () => {
  const browser = await chromium.launch({ headless: true, executablePath: process.env.BROWSER_EXECUTABLE || undefined, args: ['--no-sandbox'] });
  const fail = [];
  try {
    for (const locale of ['en-US', 'ko-KR', 'zh-CN']) {
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
      const base = { 'en-US': 'en', 'ko-KR': 'ko', 'zh-CN': 'zh-CN' }[locale], others = ['en', 'ko', 'zh-CN'].filter(l => l !== base), other = others[0];
      const initial = await page.evaluate(() => document.documentElement.lang);
      if (initial !== base) fail.push(`${locale}: initial lang ${initial}`);
      const visible = () => page.evaluate(() => {
        const clone = document.body.cloneNode(true);
        clone.querySelectorAll('[aria-hidden="true"], script, style').forEach(n => n.remove());
        const bits = [document.title, clone.textContent.replace(/\s*\n\s*/g, '\n')];
        document.querySelectorAll('[aria-label],[title]').forEach(e => bits.push((e.getAttribute('aria-label') || '') + ' ' + (e.getAttribute('title') || '')));
        return bits.join('\n');
      });

      for (const lang of [...others, base]) {
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
          label: await page.locator('#languageToggle').getAttribute('aria-label'),
        };
        if (checks.lang !== lang) fail.push(`${locale}/${lang}: html lang ${checks.lang}`);
        if (checks.pressed !== lang) fail.push(`${locale}/${lang}: pressed ${checks.pressed}`);
        if (checks.label !== verifyLanguagePlacement.MENU_LABEL[lang]) fail.push(`${locale}/${lang}: button label ${checks.label}`);
        if (lang === 'en') {
          const stray = text.split('\n').filter(l => (hangul(l) || han(l)) && !NATIVE.test(l));
          if (stray.length) fail.push(`${locale}/en: Korean left: ${JSON.stringify(stray.slice(0, 8))} | parts: ${JSON.stringify({vis: (await visible()).split('\n').filter(l => hangul(l) && !/한국어|^\s*가(\s+A)?\s*$/.test(l)).slice(0,4), cal: cal.split('\n').filter(hangul).slice(0,4), dlg: dialog.split(' ').filter(hangul).slice(0,4), calAria: hangul(calAria), where: await page.evaluate(() => { const o = []; document.querySelectorAll('*').forEach(e => { const own = [...e.childNodes].filter(n => n.nodeType === 3).map(n => n.textContent).join(''); if (/전체/.test(own) && e.tagName !== 'SCRIPT') o.push(e.tagName + '#' + e.id + '.' + String(e.className).slice(0, 25) + ' vis=' + !!(e.offsetWidth || e.offsetHeight) + ' parent=' + (e.parentElement && e.parentElement.id)); }); return o; })})}`);
          if (checks.heading !== 'Token Usage') fail.push(`en heading ${checks.heading}`);
          if (!/New input/.test(checks.th) || !/Latency/.test(checks.th)) fail.push('en table headers: ' + checks.th);
          if (!/Failed only/.test(checks.badge)) fail.push('en badge: ' + checks.badge);
          if (!/records/.test(checks.count)) fail.push('en count: ' + checks.count);
          if (!/Failed: /.test(checks.fail)) fail.push('en fail kpi: ' + checks.fail);
        } else if (lang === 'zh-CN') {
          if (checks.heading !== 'Token 用量') fail.push(`zh heading ${checks.heading}`);
          if (!/新输入/.test(checks.th) || !/延迟/.test(checks.th)) fail.push('zh headers: ' + checks.th);
          if (!/仅失败/.test(checks.badge)) fail.push('zh badge: ' + checks.badge);
          if (!/条/.test(checks.count)) fail.push('zh count: ' + checks.count);
          if (!/^失败：/.test(checks.fail)) fail.push('zh fail kpi: ' + checks.fail);
          if (!/天/.test(checks.chart)) fail.push('zh chart: ' + checks.chart);
          const stray = text.split('\n').filter(l => hangul(l) && !NATIVE.test(l));
          if (stray.length) fail.push(`${locale}/zh-CN: Korean left: ${JSON.stringify(stray.slice(0, 6))}`);
          const english = text.split('\n').filter(l => /\b(Total tokens|Request details|Load more|Previous month|Failed only|Recent requests|Overview)\b/.test(l));
          if (english.length) fail.push(`${locale}/zh-CN: English left: ${JSON.stringify(english.slice(0, 5))}`);
        } else {
          if (checks.heading !== '토큰 사용량') fail.push(`ko heading ${checks.heading}`);
          if (!/신규 입력/.test(checks.th) || !/지연/.test(checks.th)) fail.push('ko headers: ' + checks.th);
          if (!/실패만/.test(checks.badge)) fail.push('ko badge: ' + checks.badge);
          const english = text.split('\n').filter(l => /\b(Total tokens|Request details|Load more|Previous month|Failed only)\b/.test(l) || (han(l) && !NATIVE.test(l)));
          if (english.length) fail.push(`${locale}/ko: English/Chinese left: ${JSON.stringify(english.slice(0, 5))}`);
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
      await page.evaluate(() => selectLanguagePref('en'));
      const dlgEn = await page.evaluate(() => document.getElementById('recordDialog').textContent.replace(/\s+/g, ' '));
      const dbg = await page.evaluate(() => ({ idx: state.openRecordIdx, lang: currentLang, rows: document.querySelectorAll('#dialogBody .detail-row').length, open: document.getElementById('detailOverlay').className, recs: state.records.length, hasLoaded: state.hasLoadedData, inFlight: state.inFlight }));
      if (!/Request details/i.test(dlgEn) || !/Total tokens/i.test(dlgEn) || hangul(dlgEn)) fail.push(`${locale}: open dialog did not follow language: ${JSON.stringify(dbg)} text=${JSON.stringify(dlgEn.slice(0, 300))}`);
      await page.keyboard.press('Escape');

      // storage-event sync from another dashboard
      for (const l of ['ko', 'zh-CN']) {
        await page.evaluate(v => { localStorage.setItem('cpa_plugin_language', v); window.dispatchEvent(new StorageEvent('storage', { key: 'cpa_plugin_language', newValue: v })); }, l);
        if ((await page.evaluate(() => document.documentElement.lang)) !== l) fail.push(`${locale}: storage event ${l} not followed`);
      }

      // The header refresh icon reloads the data without changing the language or the label.
      for (const lang of [base, ...others]) {
        await switchLang(page, lang);
        const hits = []; const onReq = r => { if (/\/data\/(summary|records)/.test(r.url())) hits.push(r.url()); };
        page.on('request', onReq);
        await page.click('#refreshBtn');
        await page.waitForTimeout(400);
        page.off('request', onReq);
        const st = await page.evaluate(() => ({ lang: document.documentElement.lang, stored: localStorage.getItem('cpa_plugin_language'), label: document.getElementById('refreshBtn').getAttribute('aria-label'), disabled: document.getElementById('refreshBtn').disabled }));
        if (hits.length < 2) fail.push(`${locale}/${lang}: refresh click did not reload summary+records (${hits.length})`);
        if (st.lang !== lang || st.stored !== lang) fail.push(`${locale}/${lang}: refresh click changed language -> ${st.lang}/${st.stored}`);
        if (st.label !== REFRESH[lang]) fail.push(`${locale}/${lang}: refresh label ${st.label}`);
        if (st.disabled) fail.push(`${locale}/${lang}: refresh button stayed disabled after the load finished`);
      }
      // persistence
      await switchLang(page, 'en');
      await page.reload();
      await page.waitForSelector('#recordsBody tr.ui-table-row');
      const after = await page.evaluate(() => [document.documentElement.lang, localStorage.getItem('cpa_plugin_language')]);
      if (after[0] !== 'en' || after[1] !== 'en') fail.push(`${locale}: not persisted ${after}`);
      if (errors.length) fail.push(`${locale}: page errors ${errors.slice(0, 3)}`);
      await ctx.close();
    }
  } finally { await browser.close(); }
  if (fail.length) { console.log('FAIL\n' + fail.join('\n')); process.exit(1); }
  console.log('PASS: detect (en/ko/zh-CN), menu, calendar/dialog/table/filters localized, no stray text, dialog follows switch, storage sync, persistence, no overflow');
})();
