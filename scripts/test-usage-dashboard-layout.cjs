#!/usr/bin/env node
// Optional browser regression: install Playwright or set PLAYWRIGHT_MODULE.
// BROWSER_EXECUTABLE selects a system browser; SCREENSHOT_DIR captures full pages.
// Targets the mounted Radix dohwi-ui controls: #presetGroupControl (radiogroup inside #presetGroup),
// one range trigger #filterDateRangeTrigger (native #filterFromDate/#filterToDate stay hidden as the
// query source) and Select triggers #filter*Trigger (native <select> are hidden, so they have no geometry).
const assert = require('node:assert/strict');
const fs = require('node:fs');
const path = require('node:path');
const { chromium } = require(process.env.PLAYWRIGHT_MODULE || 'playwright');
const { seedKey } = require('./lib/seed-key.cjs');
const root = path.resolve(__dirname, '..');
const html = require('./lib/dashboard-html.cjs').readDashboardHTML(root);
const LONG_MODEL = 'claude-super-long-enterprise-model-identifier-'.repeat(5);
const LONG_ENDPOINT = '/enterprise/private/cluster/endpoint/resource/'.repeat(5);
const now = new Date().toISOString();
const metric = { provider: 'anthropic', model: LONG_MODEL, requests: 1234, failures: 3, input_tokens: 987654, fresh_input_tokens: 765432, output_tokens: 345678, cached_tokens: 222222, reasoning_tokens: 45678, total_tokens: 1333332 };
const full = {
  summary: { total_count: 1234, summary: [metric], by_provider: [{ ...metric, key: 'anthropic' }], by_endpoint: [{ ...metric, key: LONG_ENDPOINT }], by_model: Array.from({ length: 10 }, (_, i) => ({ ...metric, key: i ? 'model-' + i : LONG_MODEL })), daily: [{ ...metric, key: now.slice(0, 10) }] },
  records: { has_more: true, next_cursor: 'fixture-cursor', records: Array.from({ length: 20 }, (_, i) => ({ id: 'record-' + i, requested_at: now, provider: 'anthropic', model: LONG_MODEL, endpoint: LONG_ENDPOINT, fresh_input: 765432, cached: 222222, output: 345678, total: 1333332, latency_ms: 1234, failed: i === 2 })) },
};
const empty = { summary: { total_count: 0, summary: [] }, records: { has_more: false, records: [] } };
const SELECTS = [['filterProvider', 'anthropic'], ['filterEndpoint', LONG_ENDPOINT], ['filterModel', LONG_MODEL], ['filterStatus', 'failed']];
const selectors = ['.filter-card', '#presetGroup', '#presetGroupControl', '#customDateWrap', '#dateRangeMount', '#filterDateRangeTrigger', '.filter-fields-row', ...SELECTS.map(([id]) => '#' + id + 'Trigger'), '#resetFiltersBtn', '.kpi-groups', '.dashboard-main-grid', '#breakdownSection', '#chartCard', '#recordsPanel', '#recordsPanel .table-responsive', '#loadMoreBar', ...Array.from({ length: 10 }, (_, i) => '#recordsPanel th:nth-child(' + (i + 1) + ')')];
const widths = [1440, 1280, 1024, 900, 820, 768, 600, 560, 390, 360, 320];
const arg = process.argv.findIndex(x => x === '--width');
const selectedWidths = arg < 0 ? widths : [Number(process.argv[arg + 1])];

async function settled(page) {
  await page.waitForFunction(() => !state.inFlight);
  await page.evaluate(() => new Promise(resolve => requestAnimationFrame(() => requestAnimationFrame(resolve))));
}
async function geometry(page) {
  return page.evaluate(list => Object.fromEntries(list.map(selector => {
    const rect = document.querySelector(selector).getBoundingClientRect();
    if (!rect.width && !rect.height) return [selector, [0, 0, 0, 0]];
    return [selector, [rect.x + scrollX, rect.y + scrollY, rect.width, rect.height]];
  })), selectors);
}
async function stable(page, baseline, scenario) {
  const actual = await geometry(page);
  for (const selector of selectors) {
    actual[selector].forEach((v, i) => assert(Math.abs(v - baseline[selector][i]) <= 1, `${scenario}: ${selector} geometry[${i}] moved ${baseline[selector][i]} -> ${v}`));
  }
  const overflow = await page.evaluate(() => Math.max(document.documentElement.scrollWidth, document.body.scrollWidth) > innerWidth + 1);
  assert(!overflow, `${scenario}: horizontal page overflow`);
  assert(!(await page.evaluate(() => Boolean(document.getElementById('updateStatus')))), `${scenario}: last-updated block must stay removed`);
}
// The range trigger is the only visible date control; the native inputs are hidden query sources.
async function dateState(page, custom) {
  const info = await page.evaluate(() => {
    const trigger = document.getElementById('filterDateRangeTrigger');
    const natives = ['filterFromDate', 'filterToDate'].map(id => {
      const el = document.getElementById(id);
      return { hidden: el.hidden, disabled: el.disabled, width: el.getBoundingClientRect().width, value: el.value };
    });
    return { disabled: trigger.disabled, width: trigger.getBoundingClientRect().width, label: trigger.textContent.trim(), placeholder: trigger.hasAttribute('data-placeholder'), natives,
      stale: document.querySelectorAll('#filterFromDateTrigger, #filterToDateTrigger').length };
  });
  assert.equal(info.disabled, !custom, 'range trigger enabled only for custom');
  assert(info.width > 0 && info.label && !info.placeholder, 'range trigger stays visible and initialized');
  assert.equal(info.stale, 0, 'no per-date triggers');
  info.natives.forEach(v => {
    assert(v.hidden && v.width === 0, 'native date inputs stay hidden');
    assert.equal(v.disabled, !custom);
    assert(v.value, 'native date source remains initialized');
  });
  return info;
}
const checkedPreset = page => page.evaluate(() => document.querySelector('#presetGroupControl [role=radio][aria-checked=true]').dataset.preset);
const focusedPreset = page => page.evaluate(() => document.activeElement.dataset.preset || '');
const nativeDates = page => page.evaluate(() => ['filterFromDate', 'filterToDate'].map(id => document.getElementById(id).value));
// Radix Select owns a real listbox: options receive DOM focus (no aria-activedescendant) and live in a portal.
async function chooseOption(page, id, value) {
  const label = await page.evaluate(([id, value]) => [...document.getElementById(id).options].find(o => o.value === value).textContent, [id, value]);
  const trigger = page.locator('#' + id + 'Trigger');
  await trigger.click();
  assert.equal(await trigger.getAttribute('aria-expanded'), 'true');
  const option = page.getByRole('option', { name: label, exact: true });
  await option.click();
  await page.locator('[role=listbox]').waitFor({ state: 'detached' });
  assert.equal(await page.locator('#' + id).inputValue(), value, id + ' native source follows Radix Select');
  assert.equal((await trigger.innerText()).trim(), label, id + ' trigger mirrors selection');
  await page.waitForFunction(i => document.activeElement === document.getElementById(i), id + 'Trigger', { timeout: 2000 }).catch(() => { throw new Error(id + ' focus must return to trigger'); });
}
async function preset(page, value, baseline) {
  await page.locator(`[data-preset="${value}"]`).click();
  await settled(page);
  await dateState(page, value === 'custom');
  await stable(page, baseline, `preset ${value}`);
}

(async () => {
  const browser = await chromium.launch({ headless: true, executablePath: process.env.BROWSER_EXECUTABLE || undefined, args: ['--no-sandbox'] });
  try {
    for (const width of selectedWidths) {
      const page = await browser.newPage({ viewport: { width, height: 900 }, reducedMotion: 'reduce', locale: 'ko-KR' });
      await seedKey(page);
      const errors = [];
      page.on('pageerror', error => errors.push(error.message));
      let fixture = full, failed = false, summaryRequests = 0, release, gate = new Promise(resolve => { release = resolve; });
      await page.route('**/*', async route => {
        const url = new URL(route.request().url());
        if (route.request().resourceType() === 'document') return route.fulfill({ contentType: 'text/html', body: html });
        if (url.pathname.includes('/assets/fonts/')) return route.fulfill({ contentType: 'font/woff2', body: fs.readFileSync(path.join(root, 'internal/plugin/assets/fonts', path.basename(url.pathname))) });
        if (!url.pathname.includes('/data/')) return route.fulfill({ status: 404, body: '' });
        await gate;
        if (url.pathname.endsWith('/summary')) summaryRequests++;
        if (failed) return route.fulfill({ status: 500, contentType: 'application/json', body: JSON.stringify({ error: 'fixture error' }) });
        return route.fulfill({ contentType: 'application/json', body: JSON.stringify(url.pathname.endsWith('/summary') ? fixture.summary : fixture.records) });
      });
      await page.goto('http://dashboard.invalid/usage');
      await page.evaluate(() => document.fonts.ready);
      await page.waitForFunction(() => state.inFlight);
      await page.evaluate(() => new Promise(resolve => requestAnimationFrame(() => requestAnimationFrame(resolve))));
      const baseline = await geometry(page);
      release(); gate = Promise.resolve();
      await settled(page);
      await page.waitForFunction(() => state.hasLoadedData);
      await stable(page, baseline, 'initial loading -> full');
      await dateState(page, false);
      const visual = await page.evaluate(() => {
        const groups = [...document.querySelectorAll('.kpi-group')].map(el => el.getBoundingClientRect().height);
        const records = document.getElementById('recordsPanel').getBoundingClientRect();
        const main = document.getElementById('usageMainContent').getBoundingClientRect();
        const labels = [...document.querySelectorAll('.filter-field')].map(el => ({ label: el.querySelector('label').getBoundingClientRect().bottom, control: el.querySelector('.ui-select-trigger').getBoundingClientRect().top, forTarget: el.querySelector('label').htmlFor, trigger: el.querySelector('.ui-select-trigger').id, hiddenSelect: el.querySelector(':scope select[hidden]').getBoundingClientRect().width }));
        return { groups, recordsWidth: records.width, mainWidth: main.width, labels, heading: document.querySelector('h1').textContent };
      });
      assert.equal(visual.heading, '토큰 사용량');
      assert(Math.max(...visual.groups) - Math.min(...visual.groups) <= 1, 'uniform semantic summary cards');
      assert(Math.abs(visual.recordsWidth - visual.mainWidth) <= 1, 'request ledger gets full available width');
      visual.labels.forEach(pair => {
        assert(pair.label <= pair.control, 'filter labels sit above controls, never overlap values');
        assert.equal(pair.forTarget, pair.trigger, 'label targets the Radix trigger');
        assert.equal(pair.hiddenSelect, 0, 'native select stays hidden');
      });
      // Radix RadioGroup: #presetGroupControl is the radiogroup nested in the #presetGroup mount.
      const group = await page.evaluate(() => {
        const control = document.getElementById('presetGroupControl');
        return { role: control.getAttribute('role'), nested: control.parentElement.id, outerRole: document.getElementById('presetGroup').getAttribute('role'),
          radios: control.querySelectorAll('[role=radio]').length, checked: control.querySelectorAll('[aria-checked=true]').length, label: control.getAttribute('aria-label') };
      });
      assert.deepEqual(group, { role: 'radiogroup', nested: 'presetGroup', outerRole: null, radios: 6, checked: 1, label: '기간 선택' });
      assert.equal(await page.locator('#filterDateRangeTrigger').getAttribute('aria-haspopup'), 'dialog');
      // Radix keyboard model: arrow keys move focus and selection between items with roving tabindex.
      await page.locator('[data-preset="today"]').focus();
      await page.keyboard.down('ArrowRight');
      await page.waitForFunction(() => document.querySelector('#presetGroupControl [aria-checked=true]').dataset.preset === 'yesterday');
      await page.keyboard.up('ArrowRight');
      await settled(page);
      assert.equal(await checkedPreset(page), 'yesterday');
      await page.keyboard.down('ArrowLeft');
      await page.waitForFunction(() => document.querySelector('#presetGroupControl [aria-checked=true]').dataset.preset === 'today');
      await page.keyboard.up('ArrowLeft');
      await settled(page);
      assert.equal(await checkedPreset(page), 'today');
      await dateState(page, false);
      await stable(page, baseline, 'preset keyboard');
      for (const value of ['custom', 'today', 'yesterday', '7d', '30d', 'month', 'custom', 'today', 'custom']) await preset(page, value, baseline);
      // Range picker: only a complete, ordered range is committed (one atomic native update + one query).
      const trigger = page.locator('#filterDateRangeTrigger');
      const popup = page.locator('.ui-date-picker-content');
      const days = () => popup.locator('.ui-calendar-day:not([data-outside])');
      await preset(page, 'custom', baseline);
      const before = await nativeDates(page);
      let requests = summaryRequests;
      await trigger.click();
      await popup.waitFor();
      assert.equal(await popup.getAttribute('role'), 'dialog');
      assert.equal(await trigger.getAttribute('aria-expanded'), 'true');
      await stable(page, baseline, 'range popup open');
      await days().nth(14).click();
      assert.equal(await popup.count(), 1, 'start-only pick keeps the popup open');
      assert.equal(await popup.locator('.ui-calendar-status').innerText(), '종료일을 선택하세요');
      assert.deepEqual(await nativeDates(page), before, 'start-only pick must not touch the hidden dates');
      await page.keyboard.press('Escape');
      await popup.waitFor({ state: 'detached' });
      assert.deepEqual(await nativeDates(page), before, 'abandoned draft is discarded');
      assert.equal(summaryRequests, requests, 'incomplete range never queries');
      assert(await page.locator('#errorBox').evaluate(el => el.classList.contains('is-hidden')), 'unfinished range is not an error');
      // Radix FocusScope restores focus in a timer after unmount.
      await page.waitForFunction(() => document.activeElement === document.getElementById('filterDateRangeTrigger'), null, { timeout: 2000 }).catch(() => { throw new Error('Escape must return focus to the range trigger'); });
      // Picking the later day first still commits an ordered range.
      await trigger.click(); await popup.waitFor();
      const first = await days().nth(20).getAttribute('data-date'), second = await days().nth(4).getAttribute('data-date');
      await days().nth(20).click();
      await days().nth(4).click();
      await popup.waitFor({ state: 'detached' });
      await settled(page);
      assert.deepEqual(await nativeDates(page), [second, first], 'hidden dates receive the ordered range');
      assert.equal(summaryRequests, requests + 1, 'complete range queries exactly once');
      assert((await trigger.innerText()).trim().length > 4 && !await trigger.evaluate(el => el.hasAttribute('data-placeholder')), 'trigger mirrors committed range');
      await dateState(page, true);
      await stable(page, baseline, 'valid custom range');
      // Legacy guard behind the hidden source: an inverted or incomplete pair submitted explicitly is reported.
      await page.evaluate(() => { document.getElementById('filterFromDate').value = '2025-01-15'; document.getElementById('filterToDate').value = '2025-01-05'; });
      assert(await page.locator('#errorBox').evaluate(el => el.classList.contains('is-hidden')), 'programmatic draft should not interrupt editing');
      await page.locator('#filterForm').evaluate(form => form.requestSubmit());
      assert(!await page.locator('#errorBox').evaluate(el => el.classList.contains('is-hidden')), 'explicit invalid submission reports validation');
      assert(await page.locator('#retryLoadBtn').evaluate(el => el.classList.contains('is-hidden')), 'validation cannot retry a stale query');
      await stable(page, baseline, 'invalid dates validation');
      await preset(page, 'today', baseline);
      assert(await page.locator('#errorBox').evaluate(el => el.classList.contains('is-hidden')), 'preset clears stale validation');
      for (const [id, value] of SELECTS) {
        await chooseOption(page, id, value);
        await settled(page); await stable(page, baseline, 'filter ' + id);
      }
      // Radix Select: the listbox is a portal, options take real DOM focus (no aria-activedescendant) and Escape restores the trigger.
      const providerTrigger = page.locator('#filterProviderTrigger');
      await providerTrigger.focus();
      await page.keyboard.press('Enter');
      const listbox = page.getByRole('listbox');
      await listbox.waitFor();
      assert.equal(await providerTrigger.getAttribute('aria-activedescendant'), null);
      assert.equal(await listbox.getAttribute('aria-activedescendant'), null);
      assert.equal(await page.evaluate(() => document.activeElement.getAttribute('role')), 'option', 'focus lives on the option');
      assert.equal(await page.evaluate(() => document.activeElement.getAttribute('aria-selected')), 'true', 'focus starts on the selected option');
      await page.keyboard.press('Home');
      await page.waitForFunction(() => document.activeElement.getAttribute('role') === 'option' && document.activeElement.textContent === '전체', null, { timeout: 2000 }).catch(() => { throw new Error('Home must focus the first option'); });
      await page.keyboard.press('Escape');
      await listbox.waitFor({ state: 'detached' });
      await page.waitForFunction(() => document.activeElement === document.getElementById('filterProviderTrigger'), null, { timeout: 2000 }).catch(() => { throw new Error('Escape must restore trigger focus'); });
      assert.equal(await page.locator('#filterProvider').inputValue(), 'anthropic', 'Escape keeps the committed value');
      const badge = await page.locator('#activeFilterBadge').getAttribute('aria-label');
      assert(badge.includes(LONG_MODEL) && badge.includes('실패만'), 'accessible condition summary includes actual values');
      await page.locator('#resetFiltersBtn').click(); await settled(page); await stable(page, baseline, 'reset');
      for (const [id] of SELECTS) {
        assert.equal(await page.locator('#' + id).inputValue(), '', id + ' reset');
        assert.equal((await page.locator('#' + id + 'Trigger').innerText()).trim(), '전체', id + ' trigger synced after reset');
      }
      assert.equal(await checkedPreset(page), 'today', 'reset returns to today');
      for (const tab of ['endpoint', 'model', 'provider']) {
        await page.locator('#tab-' + tab + '-tab').click();
        await settled(page); await stable(page, baseline, 'aggregate tab ' + tab);
      }
      gate = new Promise(resolve => { release = resolve; });
      await page.locator('[data-preset="30d"]').click();
      await page.waitForFunction(() => state.inFlight);
      await stable(page, baseline, 'pending refresh');
      assert.equal(await page.locator('#dashboardContent').getAttribute('aria-busy'), 'true');
      release(); gate = Promise.resolve(); await settled(page); await stable(page, baseline, 'completed refresh');
      failed = true; await preset(page, 'month', baseline);
      assert(!await page.locator('#errorBox').evaluate(el => el.classList.contains('is-hidden')), 'network failure displayed');
      assert(!await page.locator('#retryLoadBtn').evaluate(el => el.classList.contains('is-hidden')), 'network failure can retry');
      failed = false; await page.locator('#retryLoadBtn').click(); await settled(page); await stable(page, baseline, 'retry');
      fixture = empty; await preset(page, 'today', baseline);
      assert.equal(await page.locator('#recordsBody tr').count(), 1);
      assert.equal(await page.locator('#recordsBody .table-empty').count(), 1);
      for (const tab of ['model', 'endpoint', 'provider']) {
        await page.locator('#tab-' + tab + '-tab').click(); await settled(page); await stable(page, baseline, 'empty tab ' + tab);
      }
      fixture = full; await preset(page, 'today', baseline);
      // Extreme exact values still fit the reserved KPI row geometry.
      await page.evaluate(() => {
        document.querySelectorAll('.kpi-value').forEach(el => { el.textContent = '9,007,199,254,740,991'; });
        document.getElementById('kpiSubFail').textContent = '실패: 9,007,199,254,740,991건 (100.0%)';
      });
      await stable(page, baseline, 'maximum exact safe integer');
      await preset(page, 'today', baseline);
      // The chart is always visible: a plain section, no disclosure widget, so nothing may move.
      const chart = await page.evaluate(() => ({ tag: document.getElementById('chartCard').tagName, details: document.querySelectorAll('#chartCard details, #chartCard summary, #chartCard .chart-toggle, details#chartCard').length,
        hidden: document.getElementById('chartCard').classList.contains('is-hidden'), height: document.getElementById('chartCard').getBoundingClientRect().height, buttons: document.querySelectorAll('#chartCard button').length }));
      assert.equal(chart.tag, 'SECTION');
      assert.equal(chart.details, 0, 'chart has no details/toggle');
      assert(!chart.hidden && chart.height > 0, 'chart is always shown');
      await stable(page, baseline, 'chart always visible');
      // Dialog scroll locking must not double-compensate a reserved scrollbar gutter.
      await page.locator('#recordsBody tr').first().click(); await settled(page); await stable(page, baseline, 'record dialog open');
      await page.keyboard.press('Escape'); await settled(page); await stable(page, baseline, 'record dialog closed');
      assert.deepEqual(errors, [], 'no runtime errors');
      if (process.env.SCREENSHOT_DIR) {
        fs.mkdirSync(process.env.SCREENSHOT_DIR, { recursive: true });
        // Snapshot hints are transient: capture the settled visual design, not pulses.
        await page.evaluate(() => {
          document.querySelectorAll('.value-flash').forEach(el => el.classList.remove('value-flash', 'flash-good', 'flash-bad', 'flash-danger', 'flash-neutral'));
          document.querySelectorAll('.panel-pulse').forEach(el => el.classList.remove('panel-pulse'));
        });
        await page.screenshot({ path: path.join(process.env.SCREENSHOT_DIR, 'usage-layout-' + width + '.png'), fullPage: true });
      }
      await page.close();
      console.log(`PASS ${width}px: Radix presets/range picker, validation, filters/reset, tabs, loading/error/retry, empty/full, exact values, always-visible chart and dialog`);
    }
  } finally { await browser.close(); }
})().catch(error => { console.error(error); process.exitCode = 1; });
