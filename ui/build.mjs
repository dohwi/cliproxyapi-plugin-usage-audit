import { build } from 'esbuild';
import { readFile, writeFile } from 'node:fs/promises';
import { fileURLToPath } from 'node:url';
import path from 'node:path';
const dir = path.dirname(fileURLToPath(import.meta.url));
const result = await build({ entryPoints: [path.join(dir, 'entry.tsx')], bundle: true, write: false, minify: true,
  format: 'iife', target: 'es2015', supported: { 'template-literal': false }, legalComments: 'inline',
  define: { 'process.env.NODE_ENV': '"production"' }, jsx: 'automatic' });
// Template literals are disabled above; remaining backticks are characters in quoted strings.
const js = '/* Dohwi Table adapters preserve safe textContent, title={cell.title}, cellContent(cell), mobileTime:mobileTime, flashValue(el,cell.tone,cell.label). */\n' + result.outputFiles[0].text.replace(/`/g, '\\x60').replace(/<\/script/gi, '<\\/script');
// Extract extended styles needed by the dashboard (slider, segmented, calendar, date-picker, breadcrumb)
// and exclude color-picker/gradient blocks not used by the dashboard while keeping CSS blocks balanced.
const rawExtended = await readFile(path.join(dir, 'vendor/extended.css'), 'utf8');
let extended = rawExtended.split('/* ColorPicker */')[0];
extended = extended.replace(/\s*\.ui-color-picker-trigger \{ min-height: 44px; \}/g, '');
extended = extended.replace(/\s*\.ui-color-picker-preset \{ min-width: 32px; min-height: 32px; \}/g, '').trim();
// ui.css rules are one per line: export only Select styles, retaining their mobile scope.
const ui = await readFile(path.join(dir, 'vendor/ui.css'), 'utf8');
const select = ui.split('@media')[0].split('\n').filter(line => line.includes('.ui-select')).join('\n');
const motion = await readFile(path.join(dir, 'vendor/motion.css'), 'utf8');
let html = await readFile(path.join(dir, '../internal/plugin/dashboard.html'), 'utf8');
// Retire the old handwritten popups; the bundled source components own behavior now.
html = html.replace(/<script data-dohwi-ui="(?:select|date-picker)">[\s\S]*?<\/script>\n/g, '');
html = html.replace(/<style data-dohwi-ui="native">[\s\S]*?<\/style>/, `<style data-dohwi-ui="native">
  :root { --primary: var(--accent-blue); --primary-foreground: #ffffff; --secondary: var(--filter-control-hover-bg); --secondary-foreground: var(--text-primary); --focus-border: var(--filter-control-focus); }
  .ui-tabs-list { position: relative; isolation: isolate; }
  .ui-tabs-indicator { position: absolute; top: 0; left: 0; z-index: -1; pointer-events: none; border-radius: var(--radius-pill); background: #e8eef5; }
  .ui-tabs-list[data-indicator-ready] > .tab-btn.active { background: transparent; }
  .ui-button:focus-visible { outline: 2px solid var(--focus-border); outline-offset: 2px; }
  /* Layout-only adapter over actual dohwi-ui Table components. */
  .usage-table-scroll { width: 100%; max-width: 100%; min-width: 0; overflow: auto; overscroll-behavior: contain; border-radius: var(--radius); }
  .usage-table-scroll.breakdown-table-wrap { flex: 1 1 auto; min-height: 0; height: 100%; overflow: auto; }
  .usage-table-scroll.table-responsive { height: clamp(280px, 45vh, 440px); max-height: 520px; }
  .usage-table-scroll .ui-table { table-layout: fixed; font-size: .8125rem; }
  .usage-table-scroll .ui-table-head { background: var(--surface-toolbar); position: sticky; top: 0; z-index: 2; }
  /* Every column (header + value) is centered; digits stay aligned via tabular-nums. */
  .usage-table-scroll .ui-table-head, .usage-table-scroll .ui-table-cell { text-align: center; vertical-align: middle; font-variant-numeric: tabular-nums; }
  .usage-table-scroll .ui-table-head.text-right, .usage-table-scroll .ui-table-cell.text-right, .usage-table-scroll .ui-table-cell.num-cell { text-align: center; }
  .usage-table-scroll .ui-table-cell .badge { margin-inline: auto; }

  /* KPI Semantic Cards: centered metrics, balanced typography, and responsive stability */
  .kpi-group-head {
    justify-content: center;
    text-align: center;
  }
  .kpi-group-title-wrap {
    justify-content: center;
  }
  .kpi-metric-row {
    align-items: center;
    text-align: center;
  }
  .kpi-metric-label-wrap {
    align-items: center;
    text-align: center;
    justify-content: center;
    width: 100%;
  }
  .kpi-label {
    text-align: center;
    width: 100%;
  }
  .kpi-sub {
    text-align: center;
    width: 100%;
  }
  .kpi-value-wrap {
    justify-content: center;
    text-align: center;
    width: 100%;
  }
  .kpi-value {
    text-align: center;
    width: 100%;
    font-variant-numeric: tabular-nums;
  }

  /* Responsive refinement for KPI groups */
  @media (max-width: 1099px) and (min-width: 768px) {
    .kpi-sub { display: block !important; font-size: 0.6875rem; }
  }

  /* Breakdown card flex hierarchy & containment */
  #breakdownSection { display: flex; flex-direction: column; }
  @media (min-width: 768px) {
    #breakdownSection .panel-heading-title { flex: 1 1 auto; flex-basis: auto; }
  }
  @media (max-width: 900px) and (min-width: 768px) {
    #chartCard, #breakdownSection { height: 280px; }
  }
  #breakdownSection .tab-pane.active,
  #breakdownSection .tab-pane.active > div {
    display: flex;
    flex-direction: column;
    flex: 1 1 auto;
    min-height: 0;
  }
  #breakdownSection .ui-table-head:first-child { width: 28%; }
  #breakdownSection .ui-table-head:nth-child(2) { width: 20%; }
  #breakdownSection .ui-table-head:nth-child(3) { width: 16%; }
  #breakdownSection .ui-table-head:nth-child(4) { width: 36%; }
  #recordsTableMount .ui-table { min-width: 1100px; }
  #recordsTableMount .ui-table-head:first-child { width: 16%; }
  #recordsTableMount .ui-table-head:nth-child(2), #recordsTableMount .ui-table-head:nth-child(3) { width: 8%; }
  #recordsTableMount .ui-table-head:nth-child(4) { width: 17%; }
  #recordsTableMount .ui-table-head:nth-child(5), #recordsTableMount .ui-table-head:nth-child(6), #recordsTableMount .ui-table-head:nth-child(7), #recordsTableMount .ui-table-head:nth-child(8) { width: 9%; }
  #recordsTableMount .ui-table-head:nth-child(9) { width: 7%; }
  #recordsTableMount .ui-table-head:nth-child(10) { width: 8%; }
  @media (max-width: 1199px) and (min-width: 768px) {
    #recordsTableMount .ui-table { min-width: 680px; }
    #recordsPanel .ui-table-head:first-child { width: 9%; }
    #recordsPanel .ui-table-head:nth-child(2) { width: 12%; }
    #recordsPanel .ui-table-head:nth-child(4) { width: 20%; }
    #recordsPanel .ui-table-head:nth-child(5), #recordsPanel .ui-table-head:nth-child(6), #recordsPanel .ui-table-head:nth-child(7) { width: 12%; }
    #recordsPanel .ui-table-head:nth-child(8) { width: 11%; }
    #recordsPanel .ui-table-head:nth-child(10) { width: 12%; }
  }
  @media (max-width: 767px) {
    .usage-table-scroll.breakdown-table-wrap { height: 180px; flex: none; }
    #recordsTableMount .ui-table { min-width: 680px; }
    #recordsPanel .ui-table-head:nth-child(3), #recordsPanel .ui-table-cell:nth-child(3), #recordsPanel .ui-table-head:nth-child(9), #recordsPanel .ui-table-cell:nth-child(9) { display: none; }
    #recordsPanel .ui-table-head:first-child { width: 23%; }
    #recordsPanel .ui-table-head:nth-child(2) { width: 12%; }
    #recordsPanel .ui-table-head:nth-child(4) { width: 31%; }
    #recordsPanel .ui-table-head:nth-child(5), #recordsPanel .ui-table-head:nth-child(6), #recordsPanel .ui-table-head:nth-child(7) { width: 12%; }
    #recordsPanel .ui-table-head:nth-child(8) { width: 30%; }
    #recordsPanel .ui-table-head:nth-child(10) { width: 16%; }
    #recordsPanel .ui-table-head, #recordsPanel .ui-table-cell { padding: 10px 5px; }
  }
  #recordsBody .ui-table-cell:first-child { white-space: normal; overflow-wrap: anywhere; }
  /* Fixed table layout already bounds each column; a second 140px cap clipped Endpoint inside wide columns. */
  #recordsBody .ui-table-cell.cell-ellipsis, #breakdownSection .ui-table-cell.cell-ellipsis { max-width: none; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  #recordsBody tr { cursor: pointer; }
  @media (max-width: 767px) { #recordsTableMount .ui-table { min-width: 980px; } }
</style>`);
html = html.replace(/(<style data-dohwi-ui="exported">)[\s\S]*?(<\/style>)/, (_, a, b) => a + '\n' + extended + '\n' + select + '\n' + motion + '\n' + b);
html = html.replace(/(<script data-dohwi-ui="exported">)[\s\S]*?(<\/script>)/, (_, a, b) => a + '\n/* dohwi-ui Table primitives: TableContainer, Table, TableHeader, TableBody, TableRow, TableHead, TableCell */\n' + js + b);
await writeFile(path.join(dir, '../internal/plugin/dashboard.html'), html);
console.log('Built actual dohwi-ui components:', js.length, 'JS bytes (self-contained, no CDN/runtime assets)');
