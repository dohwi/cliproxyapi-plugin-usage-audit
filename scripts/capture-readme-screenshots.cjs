#!/usr/bin/env node
// Regenerates the README previews (docs/assets/usage-{desktop,mobile}.{en,ko}.png) from synthetic
// fixtures only. Install Playwright or set PLAYWRIGHT_MODULE; BROWSER_EXECUTABLE selects a browser.
const fs = require('node:fs');
const path = require('node:path');
const { chromium } = require(process.env.PLAYWRIGHT_MODULE || 'playwright');
const root = path.resolve(__dirname, '..');
const html = require('./lib/dashboard-html.cjs').readDashboardHTML(root);
const out = path.join(root, 'docs/assets');

const localKey = d => `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}-${String(d.getDate()).padStart(2, '0')}`;
const dailyRequests = [145, 180, 210, 165, 195, 220, 119];
const daily = dailyRequests.map((requests, i) => {
  const d = new Date(); d.setDate(d.getDate() - (6 - i));
  return { key: localKey(d), requests, failures: i % 3, total_tokens: requests * 1100, input_tokens: requests * 700, output_tokens: requests * 280 };
});
const metric = (provider, model, requests, failures, total) => ({ provider, model, requests, failures, generate_count: requests, input_tokens: Math.round(total * 0.7), fresh_input_tokens: Math.round(total * 0.55), output_tokens: Math.round(total * 0.25), cached_tokens: Math.round(total * 0.18), reasoning_tokens: Math.round(total * 0.04), total_tokens: total });
const summary = {
  total_count: 1234,
  summary: [metric('anthropic', 'claude-3-5-sonnet-20241022', 750, 1, 820000), metric('openai', 'gpt-4o', 400, 2, 450000), metric('google', 'gemini-1.5-pro', 84, 0, 63332)],
  by_provider: [{ ...metric('anthropic', 'x', 750, 1, 820000), key: 'anthropic' }, { ...metric('openai', 'x', 400, 2, 450000), key: 'openai' }, { ...metric('google', 'x', 84, 0, 63332), key: 'google' }],
  by_endpoint: [{ ...metric('a', 'x', 1100, 3, 1200000), key: 'direct' }, { ...metric('a', 'x', 134, 0, 133332), key: 'warp1' }],
  by_model: [{ ...metric('a', 'x', 750, 1, 820000), key: 'claude-3-5-sonnet-20241022' }, { ...metric('a', 'x', 400, 2, 450000), key: 'gpt-4o' }, { ...metric('a', 'x', 84, 0, 63332), key: 'gemini-1.5-pro' }],
  daily,
};
const at = m => new Date(Date.now() - m * 60000).toISOString();
const rec = (id, minutes, provider, model, fresh, cached, output, latency, failed) => ({ id, requested_at: at(minutes), provider, model, endpoint: 'direct', fresh_input: fresh, cached, output, total: fresh + cached + output, input: fresh + cached, reasoning: 0, latency_ms: latency, ttft_ms: Math.round(latency / 4), failed: !!failed, failure_status: failed ? '429' : '' });
const records = { has_more: true, next_cursor: 'fixture', records: [
  rec('r1', 2, 'anthropic', 'claude-3-5-sonnet-20241022', 4210, 18500, 1240, 2450),
  rec('r2', 6, 'openai', 'gpt-4o', 1820, 0, 640, 1120),
  rec('r3', 10, 'anthropic', 'claude-3-5-haiku', 3200, 8400, 450, 890),
  rec('r4', 15, 'openai', 'o3-mini', 12500, 0, 3200, 5400),
  rec('r5', 22, 'google', 'gemini-1.5-pro', 2400, 4100, 820, 1870),
  rec('r6', 31, 'openai', 'gpt-4o', 900, 0, 0, 310, true),
  rec('r7', 44, 'anthropic', 'claude-3-5-sonnet-20241022', 5100, 21000, 1500, 2900),
  rec('r8', 58, 'google', 'gemini-1.5-pro', 1800, 2200, 600, 1500),
] };

async function capture(browser, lang, kind) {
  const mobile = kind === 'mobile';
  const ctx = await browser.newContext({ locale: lang === 'en' ? 'en-US' : 'ko-KR', reducedMotion: 'reduce', viewport: mobile ? { width: 390, height: 900 } : { width: 1440, height: 1080 } });
  const page = await ctx.newPage();
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
  await page.click(`.lang-btn[data-lang="${lang}"]`);
  await page.click('#presetGroupControl [data-preset="7d"]');
  await page.waitForFunction(() => !state.inFlight);
  await page.waitForSelector('#dailyChartSummary');
  await page.evaluate(() => document.fonts.ready);
  await page.evaluate(() => new Promise(r => requestAnimationFrame(() => requestAnimationFrame(r))));
  const file = path.join(out, `usage-${kind}.${lang}.png`);
  await page.screenshot({ path: file, fullPage: mobile });
  await ctx.close();
  return file;
}

(async () => {
  fs.mkdirSync(out, { recursive: true });
  const browser = await chromium.launch({ headless: true, executablePath: process.env.BROWSER_EXECUTABLE || undefined, args: ['--no-sandbox'] });
  try {
    for (const lang of ['en', 'ko']) for (const kind of ['desktop', 'mobile']) console.log('wrote', path.relative(root, await capture(browser, lang, kind)));
  } finally { await browser.close(); }
})();
