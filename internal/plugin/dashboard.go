package plugin

// DashboardHTML is the self-contained dashboard page for usage-audit.
const DashboardHTML = `<!DOCTYPE html>
<html lang="ko" data-motion="on" data-motion-style="material">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>토큰 사용량 대시보드</title>
<style>
  :root {
    /* Shared conceptual tokens (aligned with quota-monitor) */
    --bg-page: #f7f8fa;
    --bg-card: #ffffff;
    --border-color: #e1e6ed;
    --border-subtle: #e4e7ec;
    --text-primary: #182230;
    --text-secondary: #667085;
    --text-muted: #667085;
    --accent-blue: #3b6593;
    --accent-blue-hover: #294e76;
    --accent-blue-light: #eef4fb;
    --radius-card: 12px;
    --radius-sm: 6px;
    --radius-pill: 9999px;
    --filter-control-height: 38px;
    --filter-control-radius: 8px;
    --filter-control-border: #c7ced8;
    --filter-control-bg: #ffffff;
    --filter-control-hover-bg: #f3f7fc;
    --filter-control-focus: #2563eb;
    --filter-label-color: #667085;
    --filter-gap: 8px;

    /* Surface tokens */
    --surface-toolbar: #f7f8fc;
    --surface-hover: #f3f7fc;
    --surface-active: #ffffff;
    --surface-input: #ffffff;
    --surface-tab-active: #eef4fb;
    --surface-loading: #f7f8fc;
    --surface-error: #fdf1f1;
    --border-focus: #9fc1e3;
    --border-error: #e8b7b7;
    --text-error: #b91c1c;
    --text-success: #047857;

    /* Semantic token colors (AA-safe for text/legend/table numerals) */
    --tok-in: #7c3aed;
    --tok-out: #059669;
    --tok-reason: #a75d24;
    --tok-cache: #a24d78;

    /* Semantic KPI accents on near-white surfaces. */
    --in-fg: #7c3aed;   --in-tint: #f4f2fc;   --in-strip: #7c3aed;
    --out-fg: #059669;  --out-tint: #eef8f2;  --out-strip: #059669;
    --total-fg: #2563eb; --total-tint: #eef4fb; --total-strip: #2563eb;
    --reason-fg: #a75d24; --reason-tint: #fdf4ec; --reason-strip: #f2c6a0;
    --cache-fg: #a24d78; --cache-tint: #fbf0f5; --cache-strip: #e6b8cf;
    --req-fg: #3f5f80;   --req-tint: #f0f4f8;   --req-strip: #b9cfe4;

    /* Single source of truth for the app's top offset */
    --dashboard-top-offset: 24px;
  }
  /* kpi-card / kpi-accent compat classes — preserved for test and tool contracts;
     visual rendering is via .kpi-metric-row (flat, no card-inside-card). */
  .kpi-card { display: contents; }
  .kpi-accent { display: contents; }
  .kpi-request, .kpi-total, .kpi-input-report,
  .kpi-input, .kpi-cache, .kpi-output, .kpi-reason { display: contents; }
  * { box-sizing: border-box; margin: 0; padding: 0; }
  @font-face {
    font-family: "Manrope";
    src: url("/v0/resource/plugins/usage-audit/assets/fonts/manrope-variable.woff2") format("woff2");
    font-style: normal;
    font-weight: 200 800;
    font-display: swap;
    unicode-range: U+0000-024F, U+1E00-1EFF, U+2000-206F, U+20A0-20CF;
  }
  @font-face {
    font-family: "SUIT";
    src: url("/v0/resource/plugins/usage-audit/assets/fonts/suit-regular.woff2") format("woff2");
    font-style: normal;
    font-weight: 400;
    font-display: swap;
  }
  @font-face {
    font-family: "SUIT";
    src: url("/v0/resource/plugins/usage-audit/assets/fonts/suit-semibold.woff2") format("woff2");
    font-style: normal;
    font-weight: 600;
    font-display: swap;
  }
  html { scrollbar-gutter: stable; }
  body {
    background-color: var(--bg-page);
    color: var(--text-primary);
    font-family: "Manrope", "SUIT", -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif;
    line-height: 1.45;
    padding: var(--dashboard-top-offset) 14px 16px 14px;
    font-size: 0.85rem;
    min-width: 0;
    overflow-wrap: anywhere;
  }
  .container {
    max-width: 1480px;
    margin: 0 auto;
    min-width: 0;
  }
  /* One page, not a card containing every other card. */
  #dashboardContent { min-width: 0; }
  /* Shared header: description/time share a baseline; freshness is one compact group. */
  .page-heading { display: grid; grid-template-columns: minmax(0, 1fr); align-items: end; gap: 16px; margin-bottom: 18px; min-width: 0; }
  .page-heading-copy { min-width: 0; }
  .page-heading h1 { font-size: 1.5rem; font-weight: 800; letter-spacing: -0.04em; line-height: 1.3; }
  .page-heading p { margin-top: 4px; color: var(--text-secondary); font-size: 0.8125rem; line-height: 20px; }
  /* Quota-aligned unified freshness and refresh controls. */
  .dashboard-control-panel {
    background: var(--bg-card);
    border: 1px solid var(--border-color);
    border-radius: var(--radius-card);
    padding: 10px 14px;
    display: flex;
    align-items: center;
    gap: 10px 18px;
    flex-wrap: wrap;
    margin-bottom: 14px;
    min-width: 0;
  }

  .card { background: var(--bg-card); border: 1px solid var(--border-color); border-radius: var(--radius-card); padding: 12px 14px; margin-bottom: 10px; min-width: 0; }
  .filter-card { padding: 16px; border-radius: 12px; }
  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 8px;
    flex-wrap: wrap;
    gap: 6px;
  }
  .card-title {
    font-size: 1rem;
    font-weight: 700;
    color: var(--text-primary);
    display: flex;
    align-items: center;
    gap: 6px;
  }
  .card-desc {
    font-size: 0.8125rem;
    color: var(--text-secondary);
    font-weight: 400;
  }

  /* Semantic filter sections share one surface, with no nested boxes. */
  .sr-only { position: absolute; width: 1px; height: 1px; padding: 0; margin: -1px; overflow: hidden; clip-path: inset(50%); white-space: nowrap; border: 0; }
  .filter-card { width: 100%; margin-inline: 0; }
  .filter-panel { display: grid; gap: 12px; }
  .filter-section {
    min-width: 0;
    border: 0;
    border-radius: 0;
    padding: 0;
    background: transparent;
  }
  .filter-section legend {
    padding: 0;
    color: var(--text-primary);
    font-size: 0.875rem;
    font-weight: 750;
    letter-spacing: -0.01em;
  }
  /* Fixed tracks: dates stay put when presets/custom or query content change. */
  .filter-topline { display: contents; }
  .filter-topmeta { grid-row: 3; min-width: 0; height: 24px; }
  .filter-period-section { width: 100%; min-width: 0; }
  .filter-detail-section { min-width: 0; }
  .filter-row { min-width: 0; display: grid; align-items: center; gap: 12px; }
  .filter-period-row { grid-template-columns: minmax(0, 1fr) minmax(280px, 360px); width: 100%; }
  .filter-fields-row { grid-template-columns: minmax(0, 1fr) minmax(0, 1fr) minmax(0, 1.5fr) minmax(0, 0.8fr) 86px; align-items: end; width: 100%; }
  .filter-reset-field { min-width: 0; align-self: end; }
  .filter-reset-field .btn { width: 100%; }
  .preset-group {
    width: 100%;
    max-width: none;
    display: grid;
    grid-template-columns: repeat(6, minmax(0, 1fr));
    gap: 3px;
    min-width: 0;
    padding: 3px;
    border: 1px solid var(--border-color);
    border-radius: 10px;
    background: var(--surface-toolbar);
  }
  .preset-btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    min-width: 0;
    min-height: var(--filter-control-height);
    padding: 0 8px;
    overflow: hidden;
    border: 1px solid transparent;
    border-radius: 7px;
    background: transparent;
    color: var(--text-secondary);
    cursor: pointer;
    font-family: inherit;
    font-size: 0.875rem;
    font-weight: 600;
    line-height: 1.2;
    text-overflow: ellipsis;
    white-space: nowrap;
    transition: background-color 0.15s ease, border-color 0.15s ease, color 0.15s ease;
  }
  .preset-btn:hover { color: var(--text-primary); background: var(--surface-hover); }
  .preset-btn.active { color: #294e76; background: #e8eef5; border-color: #9eafc3; font-weight: 750; }
  .preset-btn:disabled { opacity: 0.5; cursor: not-allowed; }
  .custom-dates {
    width: 100%;
    grid-column: 2;
    display: grid;
    grid-template-columns: minmax(0, 1fr) auto minmax(0, 1fr);
    align-items: center;
    gap: 8px;
    min-width: 0;
  }
  .custom-dates .input { width: 100%; min-width: 0; padding: 0 6px; font-size: 0.8125rem; }
  /* Presets keep the date pair visible and disabled; custom only enables editing. */
  .custom-dates .input:disabled { background: #f2f4f7; color: var(--text-secondary); cursor: not-allowed; opacity: 1; }
  .date-range-separator { color: var(--text-muted); font-weight: 700; text-align: center; }
  .filter-field { position: relative; width: 100%; min-width: 0; display: grid; gap: 5px; }
  .filter-field label { color: var(--filter-label-color); font-size: 0.75rem; font-weight: 600; line-height: 1.3; }
  .filter-field .input, .filter-field .select { width: 100%; }
  .compact-select-wrap { position: relative; min-width: 0; }
  .compact-select-wrap::after {
    position: absolute;
    top: 50%;
    right: 10px;
    width: 7px;
    height: 7px;
    border-right: 1.5px solid var(--filter-label-color);
    border-bottom: 1.5px solid var(--filter-label-color);
    content: "";
    pointer-events: none;
    opacity: 0.75;
    transform: translateY(-65%) rotate(45deg);
  }
  .active-filter-badge {
    width: 100%;
    max-width: 100%;
    height: 24px;
    overflow: auto;
    display: inline-flex;
    align-items: flex-start;
    justify-content: flex-start;
    font-family: inherit;
    line-height: 1.5;
    gap: 6px;
    min-width: 0;
    margin: 0;
    padding: 0;
    border: 0;
    background: transparent;
    color: var(--text-secondary);
    font-size: 0.8125rem;
    font-weight: 500;
  }
  .active-filter-badge:not(.is-visible) { color: var(--text-muted); }
  .active-filter-label { flex: 0 0 auto; }
  .active-filter-text { min-width: 0; overflow-wrap: anywhere; white-space: normal; }

  /* Buttons & Inputs */
  .btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 5px;
    min-height: var(--filter-control-height);
    padding: 0 13px;
    border: 1px solid transparent;
    border-radius: var(--filter-control-radius);
    cursor: pointer;
    font-family: inherit;
    font-size: 0.875rem;
    font-weight: 600;
    line-height: 1.2;
    white-space: nowrap;
    transition: background-color 0.15s ease, border-color 0.15s ease, color 0.15s ease;
  }
  .btn:hover:not(:disabled) { background-color: var(--accent-blue-hover); border-color: var(--accent-blue-hover); }
  .btn:disabled { opacity: 0.5; cursor: not-allowed; }
  .btn-secondary { background-color: var(--surface-input); color: var(--text-primary); border-color: var(--border-color); }
  .btn-secondary:hover:not(:disabled) { background-color: var(--surface-hover); color: var(--text-primary); }
  .btn-sm { min-height: var(--filter-control-height); height: var(--filter-control-height); padding: 0 13px; font-size: 0.875rem; border-radius: var(--filter-control-radius); }
  .filter-reset-button { min-height: var(--filter-control-height); height: var(--filter-control-height); padding: 0 13px; border: 1px solid var(--filter-control-border); border-radius: var(--filter-control-radius); background: var(--filter-control-bg); color: #182230; }
  .filter-reset-button:hover:not(:disabled) { background: var(--filter-control-hover-bg); border-color: #98a2b3; }
  .filter-reset-button:focus-visible { outline: 2px solid var(--filter-control-focus); outline-offset: 2px; border-color: var(--filter-control-focus); }
  .input, .select {
    width: 100%;
    min-width: 0;
    height: var(--filter-control-height);
    min-height: var(--filter-control-height);
    padding: 0 11px;
    border: 1px solid var(--filter-control-border);
    border-radius: var(--filter-control-radius);
    outline: none;
    background-color: var(--filter-control-bg);
    color: var(--text-primary);
    font-family: inherit; font-size: 0.875rem; font-weight: 500; line-height: 1.2;
    text-overflow: ellipsis;
    white-space: nowrap;
    transition: border-color 0.15s ease, background-color 0.15s ease;
  }
  .select { padding: 0 28px 0 11px; text-align: left; text-align-last: left; appearance: none; -webkit-appearance: none; }
  .compact-select { width: 100%; min-width: 0; height: var(--filter-control-height); min-height: var(--filter-control-height); padding: 0 28px 0 11px; text-align: left; text-align-last: left; border: 1px solid var(--filter-control-border); border-radius: var(--filter-control-radius); outline: none; appearance: none; -webkit-appearance: none; background: var(--filter-control-bg); color: #182230; font-family: inherit; font-size: 0.875rem; font-weight: 500; line-height: 1.2; text-overflow: ellipsis; white-space: nowrap; transition: border-color 0.15s ease, background-color 0.15s ease; }
  .compact-select:hover:not(:disabled) { background: var(--filter-control-hover-bg); border-color: #98a2b3; }
  .compact-select:focus-visible { outline: 2px solid var(--filter-control-focus); outline-offset: 2px; border-color: var(--filter-control-focus); }
  .compact-select:disabled { background: #f2f4f7; color: var(--filter-label-color); cursor: not-allowed; opacity: 0.75; }
  .input::placeholder { color: #98a2b3; opacity: 1; }
  .input:hover { background-color: var(--filter-control-hover-bg); border-color: #98a2b3; }
  .input:focus { border-color: var(--filter-control-focus); background-color: var(--surface-input); outline: 2px solid var(--filter-control-focus); outline-offset: 2px; }
  .input:disabled { background: #f2f4f7; color: var(--text-muted); cursor: not-allowed; opacity: 0.75; }
  .is-hidden { display: none !important; }

  /* Focus visible accessibility */
  button:focus-visible, input:focus-visible, select:focus-visible, tr[tabindex="0"]:focus-visible { outline: 2px solid var(--accent-blue); outline-offset: 1px; }

  /* Initial loading skeleton — mirrors the account usage dashboard pattern. */
  .usage-skeleton { display: none; }
  /* Initial loading uses the real card geometry, not a differently sized replacement. */
  .usage-skeleton.is-visible { display: block; position: absolute; width: 1px; height: 1px; overflow: hidden; clip-path: inset(50%); }
  #usageDataContent.is-loading .kpi-value-wrap > span { color: transparent; background: #e2e8f0; border-radius: 5px; animation: skeletonPulse 1.5s infinite ease-in-out; }
  #usageDataContent.is-loading .loading-state { height: 135px; }
  #usageDataContent.is-loading { pointer-events: none; }
  .skeleton-line {
    display: block;
    background: #e2e8f0;
    border-radius: 5px;
    animation: skeletonPulse 1.5s infinite ease-in-out;
  }
  @keyframes skeletonPulse {
    0%, 100% { opacity: 0.62; }
    50% { opacity: 0.28; }
  }
  .skeleton-kpi-groups {
    display: grid;
    grid-template-columns: repeat(3, minmax(0, 1fr));
    gap: 10px;
    margin-bottom: 10px;
  }
  .skeleton-kpi-group {
    height: 164px;
    padding: 12px;
    background: var(--bg-card);
    border: 1px solid var(--border-subtle);
    border-radius: var(--radius-card);
  }
  .skeleton-kpi-head {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding-bottom: 11px;
    margin-bottom: 11px;
    border-bottom: 1px solid var(--border-subtle);
  }
  .skeleton-kpi-rows { display: grid; gap: 13px; }
  .skeleton-kpi-row { display: flex; justify-content: space-between; align-items: center; gap: 18px; }
  .skeleton-main-grid {
    display: grid;
    grid-template-columns: repeat(12, 1fr);
    gap: 10px;
    align-items: stretch;
  }
  .skeleton-panel {
    min-height: 310px;
    padding: 14px;
    background: var(--bg-card);
    border: 1px solid var(--border-color);
    border-radius: var(--radius-card);
  }
  .skeleton-panel-left { grid-column: span 4; }
  .skeleton-panel-right { grid-column: span 8; }
  .skeleton-panel-head { display: flex; justify-content: space-between; align-items: center; margin-bottom: 18px; }
  .skeleton-chart-bars { height: 108px; display: flex; align-items: flex-end; gap: 7px; padding: 12px 2px 0; }
  .skeleton-chart-bars .skeleton-line { flex: 1; }
  .skeleton-table { display: grid; gap: 12px; margin-top: 20px; }
  .skeleton-table-row { display: grid; grid-template-columns: 1.1fr 0.8fr 1fr 0.7fr; gap: 10px; }

  /* ── KPI Group Panels: flat card, NO card-inside-card ── */
  .kpi-groups {
    display: grid;
    grid-template-columns: repeat(3, minmax(0, 1fr));
    /* Share the tallest content-driven row size, including separate tablet/mobile rows. */
    grid-auto-rows: 1fr;
    gap: 10px;
    margin-bottom: 10px;
  }

  /* Group panel: one flat white/subtle surface, colored header, neutral body */
  .kpi-group { background: var(--bg-card); border: 1px solid var(--border-color); border-radius: var(--radius-card); display: flex; flex-direction: column; min-width: 0; overflow: hidden; }

  /* Colored header strip — subtle tint; group body is neutral, no decoration */
  .kpi-group-head {
    padding: 10px 14px;
    border-bottom: 1px solid var(--border-subtle);
    display: flex;
    flex-direction: column;
    gap: 2px;
  }
  .kpi-group-overview .kpi-group-head { background: #f8faff; }
  .kpi-group-input    .kpi-group-head { background: #faf9fe; }
  .kpi-group-output   .kpi-group-head { background: #f6fbf8; }

  .kpi-group-title-wrap {
    display: flex;
    align-items: center;
    gap: 6px;
  }
  .kpi-group-icon {
    flex: 0 0 auto;
    width: 14px;
    height: 14px;
    opacity: 0.8;
  }
  .kpi-group-title {
    font-size: 0.9rem;
    font-weight: 700;
    margin: 0;
    line-height: 1.25;
  }
  .kpi-group-overview .kpi-group-title { color: #3b6593; }
  .kpi-group-input    .kpi-group-title { color: #7c3aed; }
  .kpi-group-output   .kpi-group-title { color: #059669; }

  .kpi-group-help {
    font-size: 0.75rem;
    color: var(--text-muted);
    line-height: 1.25;
  }

  /* Metric rows: neutral body, separated by 1px divider lines; no background decoration/radius */
  .kpi-group-metrics { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); flex: 1; padding: 0 12px; }
  .kpi-group-input .kpi-group-metrics { grid-template-columns: repeat(3, minmax(0, 1fr)); }
  .kpi-metric-row { display: flex; flex-direction: column; justify-content: center; align-items: stretch; padding: 8px 6px; gap: 4px; min-width: 0; min-height: 126px; }
  .kpi-metric-row + .kpi-metric-row { border-left: 1px solid var(--border-subtle); }
  .kpi-metric-label-wrap {
    display: flex;
    flex-direction: column;
    gap: 1px;
    min-width: 0;
    flex: 0 0 auto;
    height: 38px;
    overflow: auto;
  }
  .kpi-label {
    font-size: 0.84rem;
    color: var(--text-secondary);
    font-weight: 500;
    white-space: normal;
    overflow-wrap: anywhere;
  }
  .kpi-sub {
    font-size: 0.6875rem;
    color: var(--text-muted);
    line-height: 1.4;
    white-space: normal;
    overflow-wrap: anywhere;
  }
  .kpi-value-wrap {
    text-align: left;
    flex: 0 0 auto;
    min-width: 0;
    max-width: 100%;
    min-height: 62px;
    display: flex;
    align-items: center;
  }
  .kpi-value {
    font-size: clamp(0.95rem, 0.8rem + 0.35vw, 1.125rem);
    font-weight: 800;
    color: var(--text-primary);
    line-height: 1.35;
    font-variant-numeric: tabular-nums;
    cursor: default;
    white-space: normal;
    overflow-wrap: anywhere;
    word-break: break-word;
    display: block;
    text-align: left;
  }
  /* Semantic values keep neutral metric row backgrounds. */
  .kpi-value-request      { color: var(--accent-blue); }
  .kpi-value-input-report { color: var(--tok-in); }
  .kpi-value-input        { color: var(--tok-in); }
  .kpi-value-cache        { color: var(--tok-cache); }
  .kpi-value-output       { color: var(--tok-out); }
  .kpi-value-reason       { color: var(--tok-reason); }
  .kpi-value-total        { color: var(--accent-blue); }

  /* Compat hidden elements (tests reference these IDs) */
  .kpi-compat-hidden { display: none; }

  /* Read summary -> trend/breakdown -> a full-width, readable request ledger. */
  .dashboard-main-grid { display: grid; grid-template-columns: minmax(0, 1fr) minmax(0, 1fr); gap: 14px; align-items: start; }
  .col-left-5, .col-right-7 { display: contents; }
  #chartCard { grid-column: 1; grid-row: 1; }
  #breakdownSection { grid-column: 2; grid-row: 1; }
  #recordsPanel { grid-column: 1 / -1; grid-row: 2; }

  .dashboard-main-grid > .card, .col-left-5 > .card, .col-right-7 > .card { margin-bottom: 0; }
  .panel-heading { min-height: 48px; flex: 0 0 auto; align-items: center; flex-wrap: wrap; }
  .panel-heading-title { min-width: 0; flex: 1 1 160px; display: flex; flex-direction: column; justify-content: center; gap: 2px; }
  .panel-heading .card-title { min-width: 0; flex-wrap: wrap; overflow-wrap: anywhere; }
  .panel-heading .card-desc { white-space: normal; overflow-wrap: anywhere; }
  .panel-heading .tab-nav { align-self: center; max-width: 100%; flex-wrap: wrap; }

  /* Daily trend is optional: native details retains keyboard behavior and open state. */
  #chartCard { flex: 0 0 auto; }
  #chartCard .chart-wrapper { margin-top: 6px; }
  .chart-heading { min-height: 38px; margin-bottom: 0 !important; }
  .chart-heading .card-title { min-width: 0; font-size: 0.9rem; overflow-wrap: anywhere; }
  .chart-card-compact > .card-desc { min-width: 0; height: 36px; overflow: auto; white-space: normal; overflow-wrap: anywhere; margin: 4px 0 0; }
  /* Daily trend chart compact */
  .chart-card-compact {
    width: 100%;
    min-width: 0;
    padding: 14px 16px;
  }
  .chart-wrapper {
    width: 100%;
    overflow-x: auto;
    height: 151px;
    max-height: 160px;
    -webkit-overflow-scrolling: touch;
  }
  .chart-wrapper svg { display: block; min-width: 0; margin-inline: auto; }
  .chart-empty-state { min-height: 135px; display: flex; align-items: center; justify-content: center; padding: 12px 0; color: var(--text-secondary); font-size: 0.8125rem; }
  .chart-wrapper { position: relative; }
  .daily-chart-bar { cursor: pointer; }
  .daily-chart-bar:focus-visible { stroke: #0f172a; stroke-width: 2; outline: none; }
  .daily-chart-tooltip {
    position: fixed;
    z-index: 110;
    max-width: min(300px, calc(100vw - 16px));
    padding: 8px 10px;
    background: #182230;
    color: #ffffff;
    border-radius: var(--radius-sm);
    font-size: 0.84rem;
    line-height: 1.4;
    pointer-events: none;
  }
  .daily-chart-tooltip[hidden] { display: none; }
  .daily-chart-tooltip-title { font-weight: 600; margin-bottom: 3px; }
  .daily-chart-tooltip-note { color: #d0d5dd; margin-top: 4px; }

  /* Aggregations Tabs (Compact Pill Navigation) */
  .tab-nav {
    display: inline-flex;
    background: var(--surface-toolbar);
    border-radius: var(--radius-pill);
    padding: 2px;
    gap: 2px;
    border: 1px solid var(--border-color);
  }
  .tab-btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    min-height: 30px;
    font-family: inherit;
    line-height: 1.2;
    background: transparent;
    border: none;
    color: var(--text-secondary);
    padding: 3px 11px;
    border-radius: var(--radius-pill);
    font-size: 0.875rem;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.15s ease;
  }
  .tab-btn:hover {
    color: var(--accent-blue-hover);
    background-color: var(--surface-hover);
  }
  .tab-btn.active {
    color: #294e76;
    background-color: #e8eef5;
    font-weight: 600;
  }
  .tab-pane {
    display: none;
  }
  .tab-pane.active {
    display: block;
  }
  .breakdown-table-wrap {
    height: 178px;
    overflow: auto;
    border: 0;
    border-radius: var(--radius-card);
    background: transparent;
  }

  /* Focus visible accessibility */
  button:focus-visible,
  input:focus-visible,
  select:focus-visible,
  tr[tabindex="0"]:focus-visible {
    outline: 2px solid var(--accent-blue);
    outline-offset: 1px;
  }

  /* Query results never renegotiate column widths; overflow belongs to the table. */
  .breakdown-table-wrap table { table-layout: fixed; }
  .breakdown-table-wrap th:first-child { width: 30%; }
  .breakdown-table-wrap th:nth-child(2) { width: 18%; }
  .breakdown-table-wrap th:nth-child(3) { width: 12%; }
  .breakdown-table-wrap th:nth-child(4) { width: 40%; }
  .breakdown-table-wrap th, .breakdown-table-wrap td { padding-inline: 6px; }
  .breakdown-table-wrap td { white-space: normal; overflow-wrap: anywhere; }
  .breakdown-table-wrap td.text-right { white-space: nowrap; }
  .breakdown-table-wrap table { min-width: 0; }
  .breakdown-table-wrap td.cell-ellipsis { white-space: nowrap; }
  @media (min-width: 768px) {
    #recordsPanel table { table-layout: fixed; min-width: 1100px; }
    #recordsPanel th:first-child { width: 16%; }
    #recordsPanel th:nth-child(2), #recordsPanel th:nth-child(3) { width: 8%; }
    #recordsPanel th:nth-child(4) { width: 20%; }
    #recordsPanel th:nth-child(5), #recordsPanel th:nth-child(6), #recordsPanel th:nth-child(7), #recordsPanel th:nth-child(8) { width: 9%; }
    #recordsPanel th:nth-child(9) { width: 7%; }
    #recordsPanel th:nth-child(10) { width: 5%; }
  }
  .table-empty { text-align: center; }
  .message-box {
    padding: 8px 12px;
    border-radius: var(--radius-sm);
    background-color: var(--surface-loading);
    border: 1px solid var(--border-focus);
    color: var(--accent-blue-hover);
    font-size: 0.8rem;
    margin-bottom: 10px;
  }
  .loading-state { background-color: var(--surface-loading); color: var(--text-secondary); }
  .error-box {
    background-color: var(--surface-error);
    border-color: var(--border-error);
    color: var(--text-error);
    align-items: center;
    justify-content: space-between;
    gap: 10px;
  }
  /* Validation/network feedback never pushes the dashboard or steals its space. */
  #errorBox { position: fixed; z-index: 120; right: 14px; bottom: max(14px, env(safe-area-inset-bottom)); width: min(560px, calc(100% - 28px)); max-height: min(200px, 40dvh); overflow: auto; display: flex; margin: 0; box-shadow: 0 4px 16px rgba(24, 34, 48, 0.12); }
  #errorBox.is-hidden { display: none; }
  #errorMessage { min-width: 0; }
  .error-box .btn { flex: 0 0 auto; }

  /* Recent Records Table: dense 36-40px, sticky header, internal scroll ~520px */
  .record-time-mobile, .records-mobile-hint { display: none; }
  .table-responsive { height: clamp(280px, 45vh, 440px); overflow: auto; max-height: 520px; border: 0; border-radius: var(--radius-card); background: transparent; -webkit-overflow-scrolling: touch; }
  table {
    width: 100%;
    border-collapse: collapse;
    font-size: 0.85rem;
  }
  th, td {
    text-align: left;
    vertical-align: middle;
    padding: 6px 10px;
    border-bottom: 1px solid var(--border-subtle);
    white-space: nowrap;
    height: 38px;
  }
  th {
    background-color: var(--surface-toolbar);
    color: var(--text-secondary);
    font-weight: 600;
    font-size: 0.8rem;
    position: sticky;
    top: 0;
    z-index: 2;
    border-bottom: 1px solid var(--border-subtle);
  }
  #recordsBody tr { cursor: pointer; }
  #recordsBody tr:hover, #breakdownSection tbody tr:hover {
    background-color: var(--surface-hover);
  }
  .cell-ellipsis {
    max-width: 140px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  #recordsPanel td:first-child { white-space: normal; overflow-wrap: anywhere; }
  .num-cell {
    white-space: nowrap;
    overflow-wrap: normal;
    text-align: right;
    font-variant-numeric: tabular-nums;
  }
  .num-input { color: var(--tok-in); }
  .num-output { color: var(--tok-out); }
  .num-total { font-weight: 600; color: var(--text-primary); }
  .badge {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    min-height: 22px;
    font-family: inherit;
    padding: 2px 6px;
    border-radius: var(--radius-pill);
    font-size: 0.75rem;
    font-weight: 600;
    line-height: 1.2;
  }
  .badge-success {
    background-color: #ecfdf5;
    color: #047857;
    border: 1px solid #a7f3d0;
  }
  .badge-fail {
    background-color: #fef2f2;
    color: #b91c1c;
    border: 1px solid #fecaca;
  }
  .badge-tag {
    background-color: #f3f4f6;
    color: var(--text-secondary);
  }
  .text-right {
    text-align: right;
    font-variant-numeric: tabular-nums;
  }
  .text-muted {
    color: var(--text-muted);
  }
  .pagination-bar {
    display: flex;
    justify-content: center;
    align-items: center;
    padding: 10px 0 4px 0;
    gap: 8px;
  }

  /* Reserve footer space even when pagination has nothing to load. */
  #loadMoreBar { min-height: calc(var(--filter-control-height) + 14px); }
  #loadMoreBar.is-hidden { display: flex !important; visibility: hidden; }

  /* Request detail modal: centered dialog on desktop, bottom sheet on mobile. */
  .detail-overlay { position: fixed; inset: 0; z-index: 1000; display: grid; place-items: center; padding: 24px; background: rgba(24, 34, 48, 0.38); opacity: 0; visibility: hidden; pointer-events: none; transition: opacity 0.2s ease, visibility 0s linear 0.2s; }
  .detail-overlay.open { opacity: 1; visibility: visible; pointer-events: auto; transition-delay: 0s; }
  .record-dialog { width: min(680px, calc(100vw - 32px)); max-height: 82vh; max-height: 82dvh; display: flex; flex-direction: column; overflow: hidden; border: 1px solid var(--border-color); border-radius: 16px; background: var(--bg-card); box-shadow: 0 16px 40px rgba(24, 34, 48, 0.16); opacity: 0; transform: translateY(10px) scale(0.98); transition: opacity 0.2s ease, transform 0.2s cubic-bezier(0.16, 1, 0.3, 1); }
  .detail-overlay.open .record-dialog { opacity: 1; transform: translateY(0) scale(1); }
  .dialog-drag-handle { display: none; }
  .dialog-header { position: sticky; top: 0; z-index: 1; display: flex; align-items: center; justify-content: space-between; gap: 12px; padding: 14px 18px; border-bottom: 1px solid var(--border-color); background: #fcfcfd; }
  .dialog-title { font-size: 0.95rem; font-weight: 700; color: var(--text-primary); }
  .dialog-close { flex: 0 0 auto; display: inline-flex; align-items: center; justify-content: center; min-width: 32px; min-height: 32px; padding: 4px 8px; border: 1px solid transparent; border-radius: 6px; background: transparent; color: var(--text-secondary); font-size: 1.25rem; line-height: 1; cursor: pointer; }
  .dialog-close:hover { color: var(--text-primary); background: var(--surface-hover); border-color: var(--border-color); }
  .dialog-body { min-width: 0; overflow-y: auto; overscroll-behavior: contain; padding: 14px 16px; }
  .detail-grid { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 12px; }
  .detail-section { min-width: 0; border: 1px solid var(--border-color); border-radius: 12px; padding: 10px 12px; background: #ffffff; }
  .detail-section-title { font-size: 0.875rem; font-weight: 600; color: var(--text-secondary); text-transform: uppercase; letter-spacing: 0.03em; margin-bottom: 8px; border-bottom: 1px solid #f3f4f6; padding-bottom: 3px; }
  .detail-row { display: flex; justify-content: space-between; gap: 8px; font-size: 0.8rem; padding: 3px 0; border-bottom: 1px dashed #f3f4f6; }
  .detail-row:last-child { border-bottom: none; }
  .detail-key { min-width: 0; color: var(--text-secondary); font-weight: 500; }
  .detail-val { min-width: 0; max-width: 65%; color: var(--text-primary); text-align: right; overflow-wrap: anywhere; font-variant-numeric: tabular-nums; }

  /* Responsive layout breakpoints */
  @media (min-width: 768px) {
    #chartCard, #breakdownSection { height: 260px; min-height: 0; }
    #breakdownSection { display: flex; flex-direction: column; }
    #breakdownSection .tab-pane.active { display: flex; flex-direction: column; flex: 1; min-height: 0; }
    #breakdownSection .breakdown-table-wrap { flex: 1; min-height: 0; }
  }
  @media (max-width: 1199px) and (min-width: 768px) {
    :root { --dashboard-top-offset: 20px; }
  }
  @media (max-width: 1199px) and (min-width: 768px) {
    #recordsPanel table { min-width: 680px; }
    #recordsPanel th:nth-child(3), #recordsPanel td:nth-child(3), #recordsPanel th:nth-child(9), #recordsPanel td:nth-child(9) { display: none; }
    #recordsPanel th:first-child { width: 9%; }
    #recordsPanel th:nth-child(2) { width: 12%; }
    #recordsPanel th:nth-child(4) { width: 20%; }
    #recordsPanel th:nth-child(5), #recordsPanel th:nth-child(6), #recordsPanel th:nth-child(7) { width: 12%; }
    #recordsPanel th:nth-child(8) { width: 15%; }
    #recordsPanel th:nth-child(10) { width: 8%; }
    .record-time-full { display: none; }
    .record-time-mobile { display: inline; }
  }
  @media (max-width: 1099px) and (min-width: 768px) {
    .kpi-groups { grid-template-columns: repeat(2, minmax(0, 1fr)); }
    .kpi-group-overview { grid-column: 1 / -1; }
    .kpi-value { font-size: 0.875rem; }
    .kpi-group-input .kpi-value { font-size: 0.8125rem; }
    .kpi-metric-row { min-height: 138px; }
    .kpi-value-wrap { min-height: 84px; }
    .kpi-label { font-size: 0.75rem; }
    .kpi-sub { display: none; }
    .kpi-group-overview .kpi-sub { display: block; }
    .panel-heading-title { flex-basis: 100%; }
    .filter-period-row { grid-template-columns: minmax(0, 1fr) minmax(230px, 280px); }
    .filter-fields-row { grid-template-columns: repeat(4, minmax(0, 1fr)) 76px; gap: 8px; }
  }
  @media (max-width: 767px) {
    :root { --dashboard-top-offset: 16px; }
    .page-heading { grid-template-columns: minmax(0, 1fr); gap: 12px; margin-bottom: 14px; }
    .page-heading h1 { font-size: 1.3rem; }
    body { padding-right: 10px; padding-bottom: 12px; padding-left: 10px; }
    .container { width: 100%; }
    #dashboardContent { padding: 0; }
    .card { padding: 12px; border-radius: 12px; }
    .filter-card { padding: 12px; }
    .kpi-groups, .skeleton-kpi-groups { grid-template-columns: 1fr; gap: 8px; }
    .skeleton-kpi-group:first-child { grid-column: auto; }
    .skeleton-kpi-group { height: 150px; }
    .skeleton-main-grid { display: flex; flex-direction: column; }
    .skeleton-panel { min-height: 250px; }
    .skeleton-panel-left, .skeleton-panel-right { width: 100%; }
    .kpi-group-overview { grid-column: auto; }
    .kpi-group-head { padding: 8px 12px; }
    .kpi-group-metrics { display: grid; grid-template-columns: repeat(2, minmax(0, 1fr)); padding: 0 12px; }
    .kpi-group-input .kpi-group-metrics { grid-template-columns: repeat(3, minmax(0, 1fr)); }
    .kpi-metric-row { min-height: 116px; padding: 8px 6px; flex-direction: column; align-items: stretch; justify-content: center; gap: 4px; }
    .kpi-metric-row + .kpi-metric-row { border-top: 0; border-left: 1px solid var(--border-subtle); }
    .kpi-metric-label-wrap { flex: 0 0 auto; height: 38px; align-items: flex-start; text-align: left; overflow: auto; }
    .kpi-label { font-size: 0.75rem; }
    .kpi-sub { display: none; }
    .kpi-group-overview .kpi-sub { display: block; font-size: 0.6875rem; }
    .kpi-value-wrap { max-width: 100%; flex: 0 0 auto; text-align: left; }
    .kpi-value { font-size: 0.95rem; text-align: left; }
    .kpi-value-wrap { min-height: 62px; display: flex; align-items: center; }
    .kpi-group-input .kpi-value { font-size: 0.9rem; }
    .dashboard-main-grid { grid-template-columns: minmax(0, 1fr); }
    #chartCard { grid-column: 1; grid-row: 1; }
    #breakdownSection { grid-column: 1; grid-row: 2; }
    #recordsPanel { grid-column: 1; grid-row: 3; }
    .breakdown-table-wrap { height: 180px; }
    .breakdown-table-wrap table { font-size: 0.75rem; }
    .breakdown-table-wrap th, .breakdown-table-wrap td { padding-inline: 4px; }
    .breakdown-table-wrap th { font-size: 0.75rem; }
    .panel-heading { flex-wrap: wrap; }
    .panel-heading-title { flex-basis: 100%; }
    #filterForm { --filter-control-height: 44px; }
    .filter-period-section { display: block; }
    .filter-row { gap: 8px; }
    .filter-period-row { grid-column: auto; grid-template-columns: minmax(0, 1fr); }
    .preset-group { grid-column: 1; grid-template-columns: repeat(3, minmax(0, 1fr)); }
    .preset-btn { min-height: 44px; padding: 0 4px; }
    .custom-dates { grid-column: 1; grid-template-columns: minmax(0, 1fr) auto minmax(0, 1fr); }
    .filter-fields-row { grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 10px; }
    .filter-reset-field { grid-column: 2; justify-self: end; }
    .filter-field { width: 100%; }
    .filter-field label { font-size: 0.75rem; }
    .filter-field .compact-select { padding: 0 28px 0 10px; }
    .filter-topmeta { height: 24px; }
    .filter-reset-field .btn { width: auto; }
    .filter-field .input, .filter-field .select, .custom-dates .input { height: 44px; min-height: 44px; font-size: 16px; }
    .filter-reset-button { height: 44px; min-height: 44px; }
    .detail-overlay { place-items: end center; padding: 0; }
    .record-dialog { width: 100%; max-height: 86vh; max-height: 86dvh; border-right: 0; border-bottom: 0; border-left: 0; border-radius: 18px 18px 0 0; transform: translateY(28px); }
    .dialog-drag-handle { display: block; width: 36px; height: 4px; margin: 9px auto 0; border-radius: var(--radius-pill); background: #c7ced8; }
    .dialog-header { padding: 10px 14px 12px; }
    .dialog-body { padding: 12px 14px calc(12px + env(safe-area-inset-bottom)); }
    .detail-grid { grid-template-columns: 1fr; gap: 10px; }
    .btn-sm { min-height: var(--filter-control-height); }
    .table-responsive { height: clamp(280px, 45vh, 440px); max-height: 58vh; }
    #recordsPanel table { width: 100%; table-layout: fixed; }
    #recordsPanel th:nth-child(2), #recordsPanel td:nth-child(2),
    #recordsPanel th:nth-child(3), #recordsPanel td:nth-child(3),
    #recordsPanel th:nth-child(5), #recordsPanel td:nth-child(5),
    #recordsPanel th:nth-child(6), #recordsPanel td:nth-child(6),
    #recordsPanel th:nth-child(7), #recordsPanel td:nth-child(7),
    #recordsPanel th:nth-child(9), #recordsPanel td:nth-child(9) { display: none; }
    #recordsPanel th, #recordsPanel td { padding: 10px 5px; }
    #recordsPanel th:first-child { width: 23%; }
    #recordsPanel th:nth-child(4) { width: 31%; }
    #recordsPanel th:nth-child(8) { width: 30%; }
    #recordsPanel th:nth-child(10) { width: 16%; }
    .record-time-full { display: none; }
    .record-time-mobile { display: inline; }
    #recordsPanel tbody tr { min-height: 44px; }
    .records-mobile-hint { display: block; margin: 4px 0 0; color: var(--text-secondary); font-size: 0.75rem; }
    th, td { padding: 7px 9px; }
  }
  @media (max-width: 359px) {
    .custom-dates { gap: 4px; }
    .custom-dates .input { padding: 0 2px; font-size: 0.8125rem; }
    .preset-btn { font-size: 0.8125rem; }
  }
  /* Server-snapshot change cue. It is deliberately scoped to a value/badge, never a card or page. */
  .value-flash { animation: valueFlash 1s ease-out both; border-radius: 3px; }
  .value-flash.flash-good { background-color: #ecfdf5; color: #166534 !important; }
  .value-flash.flash-bad { background-color: #fffbeb; color: #92400e !important; }
  .value-flash.flash-danger { background-color: #fef2f2; color: #991b1b !important; }
  .value-flash.flash-neutral { background-color: #f1f5f9; }
  @keyframes valueFlash { 0% { opacity: .5; } 100% { opacity: 1; } }
  /* Snapshot pulses stay on the smallest meaningful visual group, never the dashboard. */
  .kpi-group.panel-pulse, #breakdownSection.panel-pulse, #recordsPanel.panel-pulse { animation: panelPulse 1s ease-out both; }
  @keyframes panelPulse { 0%, 100% { border-color: var(--border-color); } 45% { border-color: #94a3b8; } }
  @media (prefers-reduced-motion: reduce) {
    *, *::before, *::after { transition-duration: 0.01ms !important; animation-duration: 0.01ms !important; }
    .skeleton-line { animation: none; opacity: 0.45; }
    .value-flash { animation: none; }
    #usageDataContent.is-loading .kpi-value-wrap > span { animation: none; }
    .kpi-group.panel-pulse, #breakdownSection.panel-pulse, #recordsPanel.panel-pulse { animation: none; }
  }

</style>
<style data-dohwi-ui="exported">
/* Host tokens and opt-in motion are supplied by the application. */
:where(.ui-slider, .ui-slider *, .ui-radio-group, .ui-radio-group *, .ui-tooltip-content, .ui-popover-content, .ui-popover-content *, .ui-dropdown-menu-content, .ui-dropdown-menu-content *, .ui-progress, .ui-progress *, .ui-avatar, .ui-avatar *, .ui-separator, .ui-skeleton, .ui-table-container, .ui-table, .ui-table *, .ui-pagination, .ui-pagination *, .ui-segmented, .ui-segmented *, .ui-calendar, .ui-calendar *, .ui-date-picker-trigger, .ui-date-picker-content, .ui-breadcrumb, .ui-breadcrumb *) { box-sizing: border-box; }
:where(.ui-radio-group-item, .ui-dropdown-menu-item, .ui-pagination-link, .ui-segmented-item, .ui-calendar-day, .ui-calendar-nav, .ui-date-picker-trigger) { font: inherit; }
:where(.ui-slider-thumb, .ui-radio-group-item, .ui-popover-content, .ui-table-container, .ui-pagination-link, .ui-segmented-item, .ui-calendar-day, .ui-calendar-nav, .ui-date-picker-trigger, .ui-breadcrumb-link):focus-visible { outline: 2px solid var(--focus-border); outline-offset: 2px; }
:where(.ui-slider, .ui-radio-group-item, .ui-dropdown-menu-item, .ui-segmented-item)[data-disabled], .ui-pagination-link[aria-disabled='true'], .ui-date-picker-trigger:disabled, .ui-calendar-nav:disabled { opacity: .5; cursor: not-allowed; }
.ui-slider { position: relative; display: flex; align-items: center; width: 100%; min-height: 1.5rem; touch-action: none; user-select: none; }
.ui-slider-track { position: relative; flex-grow: 1; height: .375rem; overflow: hidden; border-radius: 999px; background: var(--secondary); }
.ui-slider-range { position: absolute; height: 100%; background: var(--primary); }
.ui-slider-thumb { position: relative; display: block; width: 1.125rem; height: 1.125rem; border: 2px solid var(--primary); border-radius: 50%; background: var(--background); }
.ui-slider[data-orientation='vertical'] { flex-direction: column; width: 1.5rem; min-height: 8rem; }
.ui-slider-track[data-orientation='vertical'] { width: .375rem; height: auto; }
.ui-slider-range[data-orientation='vertical'] { width: 100%; height: auto; }
.ui-radio-group { display: grid; gap: .75rem; }
.ui-radio-group-item { position: relative; display: inline-flex; align-items: center; justify-content: center; width: 1.125rem; height: 1.125rem; padding: 0; border: 1px solid var(--border); border-radius: 50%; background: var(--background); color: var(--primary); cursor: pointer; }
.ui-radio-group-item[data-state='checked'] { border-color: var(--primary); }
.ui-radio-group-indicator { width: .5rem; height: .5rem; border-radius: 50%; background: currentColor; }
.ui-tooltip-content { min-width: 0; z-index: 70; max-width: min(20rem, calc(100vw - 2rem)); padding: .375rem .625rem; border-radius: var(--radius); background: var(--primary); color: var(--primary-foreground); font-size: .875rem; line-height: 1.5; overflow-wrap: anywhere; }
.ui-popover-content, .ui-dropdown-menu-content { min-width: 0; overscroll-behavior: contain; overflow-wrap: anywhere; z-index: 65; max-width: calc(100vw - 2rem); border: 1px solid var(--border); border-radius: var(--radius); background: var(--surface); color: var(--foreground); box-shadow: var(--shadow-popover); overflow-y: auto; }
.ui-popover-content { width: 20rem; padding: 1rem; max-height: var(--radix-popover-content-available-height); }
.ui-dropdown-menu-content { min-width: min(11rem, calc(100vw - 2rem)); padding: .25rem; max-height: var(--radix-dropdown-menu-content-available-height); }
.ui-dropdown-menu-item { min-width: 0; position: relative; display: flex; align-items: center; gap: .5rem; min-height: 2rem; padding: .375rem .625rem; border-radius: calc(var(--radius) * .6); font-size: .9375rem; cursor: default; user-select: none; outline: none; }
.ui-dropdown-menu-item[data-highlighted] { background: var(--secondary); color: var(--secondary-foreground); }
.ui-dropdown-menu-checkbox-item { padding-inline-start: 2rem; }
.ui-dropdown-menu-indicator { position: absolute; inset-inline-start: .625rem; display: flex; align-items: center; }
.ui-dropdown-menu-label { padding: .375rem .625rem; font-size: .875rem; font-weight: 600; }
.ui-dropdown-menu-separator { height: 1px; margin: .25rem -.25rem; background: var(--border); }
.ui-progress { position: relative; width: 100%; height: .5rem; overflow: hidden; border-radius: 999px; background: var(--secondary); }
.ui-progress-indicator { height: 100%; background: var(--primary); }
.ui-avatar { position: relative; display: inline-flex; flex-shrink: 0; width: 2.5rem; height: 2.5rem; overflow: hidden; border-radius: 50%; }
.ui-avatar-image { width: 100%; height: 100%; object-fit: cover; }
.ui-avatar-fallback { display: flex; align-items: center; justify-content: center; width: 100%; height: 100%; background: var(--secondary); color: var(--secondary-foreground); font-size: .9375rem; font-weight: 500; }
.ui-separator { flex-shrink: 0; background: var(--border); }
.ui-separator[data-orientation='horizontal'] { width: 100%; height: 1px; }
.ui-separator[data-orientation='vertical'] { width: 1px; height: 100%; align-self: stretch; }
.ui-skeleton { border-radius: var(--radius); background: var(--muted); }
.ui-table-container { border-radius: var(--radius); width: 100%; max-width: 100%; min-width: 0; overflow: auto; overscroll-behavior: contain; }
.ui-table { width: 100%; overflow: hidden; border: 1px solid var(--border); border-radius: var(--radius); border-collapse: separate; border-spacing: 0; color: var(--foreground); font-size: .9375rem; text-align: start; }
.ui-table-row > :where(.ui-table-head, .ui-table-cell) { border-bottom: 1px solid var(--border); }
.ui-table-body .ui-table-row:last-child > .ui-table-cell { border-bottom: 0; }
.ui-table-body .ui-table-row:hover { background: var(--muted); }
.ui-table-head { color: var(--muted-foreground); font-weight: 500; text-align: start; }
.ui-table-head, .ui-table-cell { padding: .75rem 1rem; vertical-align: middle; }
.ui-table-caption { caption-side: bottom; margin-top: .75rem; color: var(--muted-foreground); text-align: start; }
.ui-pagination { min-width: 0; max-width: 100%; display: flex; justify-content: center; }
.ui-pagination-content { display: flex; flex-wrap: wrap; justify-content: center; align-items: center; gap: .25rem; margin: 0; padding: 0; list-style: none; }
.ui-pagination-link, .ui-pagination-ellipsis { max-width: 100%; overflow-wrap: anywhere; display: inline-flex; align-items: center; justify-content: center; gap: .375rem; min-width: 2.25rem; min-height: 2.25rem; padding: .375rem .625rem; border: 1px solid transparent; border-radius: var(--radius); color: var(--foreground); text-decoration: none; }
.ui-pagination-link:hover:not([aria-disabled='true']) { background: var(--secondary); color: var(--secondary-foreground); }
.ui-pagination-link[data-active] { border-color: var(--border); background: var(--surface); }
.ui-pagination-direction { padding-inline: .75rem; }

/* SegmentedControl: one measured box slides between single-choice items. */
.ui-segmented { position: relative; isolation: isolate; display: inline-flex; max-width: 100%; overflow-x: auto; overscroll-behavior-x: contain; scrollbar-width: none; gap: .25rem; padding: .25rem; border: 1px solid var(--border); border-radius: var(--radius); background: var(--muted); }
.ui-segmented--full { display: flex; width: 100%; }
.ui-segmented-indicator { position: absolute; top: 0; left: 0; z-index: -1; pointer-events: none; border: 1px solid var(--border); border-radius: calc(var(--radius) * .7); background: var(--background); }
.ui-segmented::-webkit-scrollbar { display: none; }
.ui-segmented-item { display: inline-flex; align-items: center; justify-content: center; gap: .375rem; flex: 0 0 auto; min-width: 0; white-space: nowrap; min-height: 2.25rem; padding: .375rem .875rem; border: 1px solid transparent; border-radius: calc(var(--radius) * .7); background: transparent; color: var(--muted-foreground); font-size: .9375rem; font-weight: 500; line-height: 1.4; text-align: center; cursor: pointer; user-select: none; }
.ui-segmented-item > svg { flex: none; width: 1.125em; height: 1.125em; }
.ui-segmented--full .ui-segmented-item { flex: 1 0 auto; }
.ui-segmented-item[data-state='checked'] { color: var(--foreground); }
.ui-segmented:not([data-indicator-ready]) .ui-segmented-item[data-state='checked'] { border-color: var(--border); background: var(--background); }
.ui-segmented[data-indicator-ready] .ui-segmented-item { background: transparent; border-color: transparent; }
.ui-segmented-item:hover:not([data-disabled], [data-state='checked']) { color: var(--foreground); }
.ui-segmented--sm .ui-segmented-item { min-height: 1.875rem; padding: .25rem .625rem; font-size: .875rem; }
.ui-segmented--lg .ui-segmented-item { min-height: 2.75rem; padding: .5rem 1.125rem; font-size: 1rem; }

/* Calendar and DatePicker */
.ui-calendar { display: inline-block; max-width: 100%; color: var(--foreground); }
.ui-calendar-header { display: flex; align-items: center; justify-content: space-between; gap: .5rem; margin-bottom: .25rem; }
.ui-calendar-caption { min-width: 0; font-size: .9375rem; font-weight: 600; text-align: center; }
.ui-calendar-nav { display: inline-flex; align-items: center; justify-content: center; flex-shrink: 0; width: 2rem; height: 2rem; padding: 0; border: 1px solid transparent; border-radius: var(--radius); background: transparent; color: var(--foreground); cursor: pointer; }
.ui-calendar-nav:hover:not(:disabled) { background: var(--secondary); }
.ui-calendar-grid { width: 100%; border-collapse: separate; border-spacing: 0; table-layout: fixed; }
.ui-calendar-grid tr { background: transparent; }
.ui-calendar-grid tr:hover { background: transparent; }
.ui-calendar-grid td, .ui-calendar-grid th { border: 0; }
.ui-calendar-weekday { padding: .25rem 0; color: var(--muted-foreground); font-size: .75rem; font-weight: 500; text-align: center; }
.ui-calendar-cell { padding: 0; text-align: center; vertical-align: middle; }
.ui-calendar-day { display: inline-flex; align-items: center; justify-content: center; width: 100%; min-width: 2.25rem; height: 2.25rem; padding: 0; border: 1px solid transparent; border-radius: var(--radius); background: transparent; color: var(--foreground); font-size: .875rem; font-variant-numeric: tabular-nums; cursor: pointer; }
.ui-calendar-day:hover:not([aria-disabled='true'], [data-selected]) { background: var(--secondary); color: var(--secondary-foreground); }
.ui-calendar-day[data-outside] { color: var(--muted-foreground); }
.ui-calendar-day[data-today] { border-color: var(--border); font-weight: 700; }
.ui-calendar-day[data-selected] { border-color: transparent; background: var(--primary); color: var(--primary-foreground); font-weight: 600; }
.ui-calendar-day[aria-disabled='true'] { opacity: .4; cursor: not-allowed; }
/* Range: the band is painted on the cell so adjacent days connect without gaps. */
.ui-calendar-cell[data-range-band] { background: var(--secondary); }
.ui-calendar-cell[data-range-preview] { background: color-mix(in srgb, var(--secondary) 55%, transparent); }
.ui-calendar-cell[data-range-band]:first-child, .ui-calendar-cell[data-range-start] { border-start-start-radius: var(--radius); border-end-start-radius: var(--radius); }
.ui-calendar-cell[data-range-band]:last-child, .ui-calendar-cell[data-range-end] { border-start-end-radius: var(--radius); border-end-end-radius: var(--radius); }
.ui-calendar-cell[data-range-band] > .ui-calendar-day:not([data-selected]) { border-radius: 0; color: var(--secondary-foreground); }
.ui-calendar-cell[data-range-band] > .ui-calendar-day:not([data-selected], [aria-disabled='true']):hover { background: color-mix(in srgb, var(--foreground) 10%, transparent); }
.ui-calendar-status { min-height: 1.25rem; margin-top: .25rem; color: var(--muted-foreground); font-size: .75rem; text-align: center; }
.ui-calendar[data-disabled] { opacity: .5; }
.ui-date-picker-trigger { display: inline-flex; align-items: center; justify-content: space-between; gap: .75rem; width: 12rem; max-width: 100%; min-width: 0; min-height: 2.25rem; padding: .375rem .75rem; border: 1px solid var(--border); border-radius: var(--radius); background: var(--background); color: var(--foreground); font-size: .9375rem; text-align: start; cursor: pointer; }
.ui-date-picker-trigger > span { min-width: 0; overflow-wrap: anywhere; }
.ui-date-picker-trigger[data-mode='range'] { width: 18rem; }
.ui-date-picker-trigger > svg { flex-shrink: 0; }
.ui-date-picker-trigger[data-placeholder] { color: var(--muted-foreground); }
.ui-date-picker-trigger:focus-visible, .ui-date-picker-trigger[data-state='open'] { outline: none; border-color: var(--focus-border); }
.ui-date-picker-content { width: min(18rem, calc(100vw - 1rem)); padding: .5rem; }
.ui-date-picker-content .ui-calendar { display: block; width: 100%; }

/* Breadcrumb */
.ui-breadcrumb { min-width: 0; max-width: 100%; }
.ui-breadcrumb-list { display: flex; flex-wrap: wrap; align-items: center; gap: .375rem; margin: 0; padding: 0; list-style: none; color: var(--muted-foreground); font-size: .9375rem; overflow-wrap: anywhere; }
.ui-breadcrumb-item { display: inline-flex; align-items: center; gap: .375rem; min-width: 0; }
.ui-breadcrumb-link { color: inherit; text-decoration: none; border-radius: calc(var(--radius) * .5); }
.ui-breadcrumb-link:hover { color: var(--foreground); text-decoration: underline; text-underline-offset: 3px; }
.ui-breadcrumb-page { color: var(--foreground); font-weight: 500; }
.ui-breadcrumb-separator, .ui-breadcrumb-ellipsis { display: inline-flex; align-items: center; }

@media (max-width: 600px), (pointer: coarse) {
  .ui-dropdown-menu-item, .ui-pagination-link, .ui-pagination-ellipsis { min-width: 44px; min-height: 44px; }
  .ui-segmented-item, .ui-date-picker-trigger { min-height: 44px; }
  .ui-calendar-day { min-width: 0; height: 44px; }
  .ui-calendar-nav { width: 44px; height: 44px; }
  .ui-date-picker-content { width: min(20rem, calc(100vw - .5rem)); padding: .25rem; }
  .ui-date-picker-trigger { font-size: max(1rem, 16px); }
  .ui-breadcrumb-link { display: inline-flex; align-items: center; min-height: 44px; }
  .ui-slider { min-height: 44px; }
  .ui-slider[data-orientation='vertical'] { width: 44px; }
  .ui-radio-group-item::after, .ui-slider-thumb::after { content: ''; position: absolute; top: calc(50% - 22px); inset-inline-start: calc(50% - 22px); width: 44px; height: 44px; }
  .ui-radio-group-item:is(:disabled, [data-disabled])::after, .ui-slider[data-disabled] .ui-slider-thumb::after { pointer-events: none; }
}

/* Navbar — responds to its own width (container query), not the viewport. */
:where(.ui-navbar, .ui-navbar *, .ui-notice-bar, .ui-notice-bar *) { box-sizing: border-box; }
:where(.ui-navbar-link, .ui-navbar-brand, .ui-notice-bar-link):focus-visible { outline: 2px solid var(--focus-border); outline-offset: -2px; }
.ui-navbar { container: ui-navbar / inline-size; position: relative; width: 100%; min-width: 0; background: var(--background); color: var(--foreground); border-bottom: 1px solid var(--border); --navbar-marker: var(--muted); }
.ui-navbar[data-variant='muted'] { background: var(--muted); --navbar-marker: var(--background); }
.ui-navbar[data-variant='transparent'] { background: transparent; border-bottom-color: transparent; }
.ui-navbar[data-sticky='true'] { position: sticky; top: 0; z-index: 20; }
.ui-navbar-inner { display: flex; align-items: center; gap: .5rem 1rem; min-width: 0; min-height: 3.5rem; padding-inline: 1rem; }
.ui-navbar-brand { display: inline-flex; align-items: center; flex: 0 0 auto; min-height: 2.75rem; border-radius: calc(var(--radius) * .7); color: var(--foreground); font-size: var(--text-lg, 1.125rem); font-weight: 700; letter-spacing: -.01em; text-decoration: none; }
.ui-navbar-content { position: relative; isolation: isolate; display: flex; align-items: stretch; flex: 1 1 auto; align-self: stretch; min-width: 0; overflow-x: auto; overscroll-behavior-x: contain; scrollbar-width: none; }
.ui-navbar-content::-webkit-scrollbar { display: none; }
.ui-navbar-list { display: flex; align-items: stretch; gap: .25rem; min-width: 0; margin: 0; padding: 0; list-style: none; }
.ui-navbar-content[data-align='center'] .ui-navbar-list { margin-inline: auto; }
.ui-navbar-content[data-align='end'] .ui-navbar-list { margin-inline-start: auto; }
.ui-navbar-item { display: flex; }
.ui-navbar-link { display: inline-flex; align-items: center; min-height: 2.5rem; padding: .5rem .75rem; border: 0; border-radius: calc(var(--radius) * .7); color: var(--muted-foreground); font-size: .9375rem; font-weight: 500; text-decoration: none; white-space: nowrap; }
.ui-navbar-link:hover, .ui-navbar-link[aria-current='page'] { color: var(--foreground); }
.ui-navbar-indicator { position: absolute; top: 0; left: 0; z-index: -1; pointer-events: none; box-sizing: border-box; }
/* underline: the marker is a 2px line at the bottom edge of the bar */
.ui-navbar[data-indicator='underline'] .ui-navbar-link { min-height: 3.5rem; border-radius: 0; border-bottom: 2px solid transparent; }
.ui-navbar[data-indicator='underline'] .ui-navbar-indicator { border-bottom: 2px solid var(--primary); }
.ui-navbar[data-indicator='underline'] .ui-navbar-content:not([data-indicator-ready]) .ui-navbar-link[aria-current='page'] { border-bottom-color: var(--primary); }
/* pill: the marker is a filled, bordered box behind the current link */
.ui-navbar[data-indicator='pill'] .ui-navbar-content, .ui-navbar[data-indicator='pill'] .ui-navbar-list { align-items: center; }
.ui-navbar[data-indicator='pill'] .ui-navbar-link { min-height: 2.25rem; }
.ui-navbar[data-indicator='pill'] .ui-navbar-indicator { border: 1px solid var(--border); border-radius: calc(var(--radius) * .7); background: var(--navbar-marker); }
.ui-navbar[data-indicator='pill'] .ui-navbar-content:not([data-indicator-ready]) .ui-navbar-link[aria-current='page'] { background: var(--navbar-marker); }
.ui-navbar-actions { display: flex; align-items: center; flex: 0 0 auto; gap: .5rem; }
.ui-navbar .ui-navbar-toggle { display: none; margin-inline-end: -.5rem; }
@container ui-navbar (max-width: 40rem) {
  .ui-navbar-inner { flex-wrap: wrap; padding-inline: 1rem .75rem; }
  .ui-navbar-brand { margin-inline-end: auto; }
  .ui-navbar .ui-navbar-toggle { display: inline-flex; }
  .ui-navbar-content { display: none; flex: 0 0 100%; align-self: auto; overflow: visible; padding-block-end: .5rem; }
  .ui-navbar-content[data-state='open'] { display: block; }
  .ui-navbar-content .ui-navbar-indicator { display: none; }
  .ui-navbar-content[data-align] .ui-navbar-list { flex-direction: column; align-items: stretch; gap: .125rem; margin: 0; }
  .ui-navbar[data-indicator] .ui-navbar-link { width: 100%; min-height: 2.75rem; border-radius: calc(var(--radius) * .7); border-bottom: 0; border-inline-start: 2px solid transparent; }
  .ui-navbar[data-indicator] .ui-navbar-content .ui-navbar-link[aria-current='page'] { border-inline-start-color: var(--primary); background: var(--navbar-marker); }
  .ui-navbar-actions { display: none; flex: 0 0 100%; flex-wrap: wrap; padding-block-end: .75rem; }
  .ui-navbar[data-open='true'] .ui-navbar-actions { display: flex; }
}

/* NoticeBar — a one-line strip for announcements. */
.ui-notice-bar { --notice-tone: var(--foreground); display: flex; align-items: center; gap: .5rem; min-width: 0; max-width: 100%; min-height: 3rem; padding-inline: 1rem; border: 1px solid var(--notice-border, var(--border)); border-radius: var(--radius); background: var(--notice-background, var(--muted)); color: var(--notice-foreground, var(--foreground)); }
.ui-notice-bar:is([data-accent], [data-variant='info'], [data-variant='success'], [data-variant='warning'], [data-variant='destructive']) { border-color: var(--notice-border, var(--notice-tone)); background: var(--notice-background, color-mix(in srgb, var(--notice-tone) 9%, var(--surface))); color: var(--notice-foreground, color-mix(in srgb, var(--notice-tone) 82%, var(--foreground))); }
.ui-notice-bar[data-variant='info'] { --notice-tone: var(--info); }
.ui-notice-bar[data-variant='success'] { --notice-tone: var(--success); }
.ui-notice-bar[data-variant='warning'] { --notice-tone: var(--warning); }
.ui-notice-bar[data-variant='destructive'] { --notice-tone: var(--danger); }
/* flush: sits edge to edge directly under a Navbar, which already draws the top line */
.ui-notice-bar[data-flush='true'] { border-inline-width: 0; border-block-start-width: 0; border-radius: 0; }
.ui-notice-bar-content { flex: 1 1 0; min-width: 0; display: flex; align-items: center; justify-content: flex-start; gap: .75rem; }
.ui-notice-bar[data-align='center'] .ui-notice-bar-content { justify-content: center; }
.ui-notice-bar[data-align='end'] .ui-notice-bar-content { justify-content: flex-end; }
.ui-notice-bar-icon { display: inline-flex; flex: 0 0 auto; color: var(--notice-icon-color, inherit); }
.ui-notice-bar-icon > svg { width: 18px; height: 18px; }
.ui-notice-bar-actions .ui-icon-button { color: inherit; }
.ui-notice-bar-actions .ui-icon-button:hover { background: color-mix(in srgb, currentColor 10%, transparent); }
.ui-notice-bar-viewport { flex: 0 1 auto; min-width: 0; overflow: hidden; }
.ui-notice-bar[data-animated][data-effect='marquee'] .ui-notice-bar-viewport { flex: 1 1 0; }
.ui-notice-bar-list { display: flex; align-items: center; min-width: 0; margin: 0; padding: 0; list-style: none; }
.ui-notice-bar-item { min-width: 0; padding-block: .5rem; font-size: .9375rem; line-height: 1.5; }
.ui-notice-bar-link { color: inherit; font-weight: 600; text-decoration: underline; text-underline-offset: 3px; border-radius: calc(var(--radius) * .5); }
.ui-notice-bar-link:hover { text-decoration-thickness: 2px; }
.ui-notice-bar-actions { display: flex; align-items: center; flex: 0 0 auto; margin-inline-end: -.5rem; }
/* Without motion every notice is a readable, wrapping list. */
.ui-notice-bar:not([data-animated]) .ui-notice-bar-list { flex-direction: column; align-items: stretch; }
.ui-notice-bar:not([data-animated]) .ui-notice-bar-item { white-space: normal; overflow-wrap: anywhere; }
/* align: items fit their text, so content justify-content moves both icon and text. */
.ui-notice-bar-item { text-align: start; }
.ui-notice-bar[data-animated][data-effect='marquee'] .ui-notice-bar-track { display: flex; width: max-content; }
.ui-notice-bar[data-animated][data-effect='marquee'] .ui-notice-bar-list { flex: 0 0 auto; }
.ui-notice-bar[data-animated][data-effect='marquee'] .ui-notice-bar-item { padding-inline-end: 3rem; white-space: nowrap; }
/* True vertical marquee, sized to one complete notice (long text wraps). */
.ui-notice-bar[data-animated][data-effect='marquee']:is([data-direction='up'], [data-direction='down']) .ui-notice-bar-viewport { height: var(--notice-height, 2.5rem); }
.ui-notice-bar[data-animated][data-effect='marquee']:is([data-direction='up'], [data-direction='down']) .ui-notice-bar-track { flex-direction: column; width: 100%; }
.ui-notice-bar[data-animated][data-effect='marquee']:is([data-direction='up'], [data-direction='down']) .ui-notice-bar-list { flex-direction: column; align-items: stretch; }
.ui-notice-bar[data-animated][data-effect='marquee']:is([data-direction='up'], [data-direction='down']) .ui-notice-bar-item { flex: 0 0 auto; padding-inline-end: 0; white-space: normal; overflow-wrap: anywhere; }
/* slide / fade: notices share one grid cell; only the current one is visible */
.ui-notice-bar[data-animated]:is([data-effect='slide'], [data-effect='fade']) .ui-notice-bar-list { display: grid; align-items: center; }
.ui-notice-bar[data-animated]:is([data-effect='slide'], [data-effect='fade']) .ui-notice-bar-item { grid-area: 1 / 1; visibility: hidden; opacity: 0; overflow-wrap: anywhere; }
.ui-notice-bar[data-animated]:is([data-effect='slide'], [data-effect='fade']) .ui-notice-bar-item[data-state='active'] { visibility: visible; opacity: 1; }
.ui-notice-bar[data-direction='left'] { --notice-enter-x: 100%; --notice-leave-x: -100%; --notice-enter-y: 0px; --notice-leave-y: 0px; }
.ui-notice-bar[data-direction='right'] { --notice-enter-x: -100%; --notice-leave-x: 100%; --notice-enter-y: 0px; --notice-leave-y: 0px; }
.ui-notice-bar[data-direction='up'] { --notice-enter-x: 0px; --notice-leave-x: 0px; --notice-enter-y: 100%; --notice-leave-y: -100%; }
.ui-notice-bar[data-direction='down'] { --notice-enter-x: 0px; --notice-leave-x: 0px; --notice-enter-y: -100%; --notice-leave-y: 100%; }
.ui-notice-bar[data-effect='fade'] { --notice-enter-x: 0px; --notice-leave-x: 0px; --notice-enter-y: 0px; --notice-leave-y: 0px; }

@media (max-width: 600px), (pointer: coarse) {
  .ui-navbar .ui-navbar-link { min-height: 44px; }
  .ui-navbar-brand { min-height: 44px; }
  .ui-notice-bar-link { display: inline-flex; align-items: center; min-height: 44px; }
}
:where(.ui-button, .ui-input, .ui-textarea, .ui-badge, .ui-card, .ui-card *, .ui-checkbox, .ui-switch, .ui-switch-thumb, .ui-select-trigger, .ui-select-content, .ui-select-content *, .ui-tabs, .ui-tabs *, .ui-accordion, .ui-accordion *, .ui-dialog-overlay, .ui-dialog-content, .ui-dialog-content *, .ui-alert, .ui-alert *) { box-sizing: border-box; }
:where(.ui-button, .ui-input, .ui-textarea, .ui-checkbox, .ui-switch, .ui-select-trigger, .ui-select-item, .ui-tabs-trigger, .ui-accordion-trigger, .ui-dialog-close) { font: inherit; }
:where(.ui-button, .ui-input, .ui-textarea, .ui-checkbox, .ui-switch, .ui-select-trigger, .ui-tabs-trigger, .ui-tabs-content, .ui-accordion-trigger, .ui-dialog-close):focus-visible { outline: 2px solid var(--focus-border); outline-offset: 2px; }
:where(.ui-button, .ui-input, .ui-textarea, .ui-checkbox, .ui-switch, .ui-select-trigger, .ui-tabs-trigger, .ui-accordion-trigger, .ui-dialog-close):disabled,
:where(.ui-checkbox, .ui-switch, .ui-select-trigger, .ui-select-item, .ui-tabs-trigger, .ui-accordion-trigger)[data-disabled] { opacity: .5; cursor: not-allowed; }
.ui-input, .ui-textarea, .ui-select-trigger { width: 100%; min-width: 0; border: 1px solid var(--border); border-radius: var(--radius); background: var(--background); color: var(--foreground); font-size: 1rem; padding: .5rem .75rem; }
.ui-input, .ui-select-trigger { min-height: 2.5rem; }
.ui-input::placeholder, .ui-textarea::placeholder, .ui-select-trigger[data-placeholder] { color: var(--muted-foreground); }
.ui-input:focus, .ui-textarea:focus, .ui-select-trigger:focus-visible { outline: none; box-shadow: none; border-color: var(--focus-border); }
.ui-input[aria-invalid='true'], .ui-textarea[aria-invalid='true'], .ui-select-trigger[aria-invalid='true'] { border-color: var(--danger); }
.ui-select-trigger { display: flex; align-items: center; justify-content: space-between; gap: .5rem; cursor: pointer; text-align: start; }
.ui-select-trigger > span { min-width: 0; overflow-wrap: anywhere; }
.ui-select-trigger > svg { flex-shrink: 0; }
.ui-select-content { z-index: 60; overflow: hidden; overscroll-behavior: contain; overflow-wrap: anywhere; min-width: min(var(--radix-select-trigger-width), calc(100vw - 2rem)); max-width: calc(100vw - 2rem); max-height: var(--radix-select-content-available-height); border: 1px solid var(--border); border-radius: var(--radius); background: var(--surface); color: var(--foreground); box-shadow: var(--shadow-popover); }
.ui-select-viewport { min-width: 0; overscroll-behavior: contain; padding: .25rem; }
.ui-select-item { position: relative; display: flex; align-items: center; min-height: 2rem; padding-block: .375rem; padding-inline: .625rem 2rem; border-radius: calc(var(--radius) * .6); font-size: .9375rem; cursor: default; user-select: none; outline: none; }
.ui-select-item[data-highlighted] { background: var(--secondary); color: var(--secondary-foreground); }
.ui-select-item-indicator { position: absolute; inset-inline-end: .625rem; display: flex; }
.ui-select-scroll { display: flex; justify-content: center; padding: .25rem; }
/* Library motion is opt-in. Scoped appearance also reaches Radix portals.
   The documentation site enables its own motion independently of demos.
   Theme tokens MUST NEVER transition: all surfaces, text and borders change
   together in one paint. Animate interaction geometry and opacity only. */
:root {
  --motion-fast: 140ms;
  --motion-normal: 220ms;
  --motion-slow: 300ms;
  --motion-ease: cubic-bezier(.2, .8, .2, 1);
  --motion-ripple: 620ms;
  --motion-dialog-shift: 8px;
  --motion-dialog-scale: .98;
  --motion-popup-scale: .97;
  --motion-spring: cubic-bezier(.2, .8, .2, 1);
}
:where([data-motion-style='material']) {
  --motion-fast: 180ms;
  --motion-normal: 360ms;
  --motion-slow: 460ms;
  --motion-ease: cubic-bezier(.22, 1, .36, 1);
  --motion-dialog-shift: 18px;
  --motion-dialog-scale: .94;
  --motion-popup-scale: .9;
  --motion-spring: cubic-bezier(.34, 1.56, .64, 1);
}
:where([data-motion-style='standard']) {
  --motion-fast: 140ms;
  --motion-normal: 220ms;
  --motion-slow: 300ms;
  --motion-ease: cubic-bezier(.2, .8, .2, 1);
  --motion-dialog-shift: 8px;
  --motion-dialog-scale: .98;
  --motion-popup-scale: .97;
  --motion-spring: cubic-bezier(.2, .8, .2, 1);
}

@media (prefers-reduced-motion: no-preference) {
  /* Feedback stays small and never changes layout or scroll position. */
  :where([data-motion='on']) .ui-button {
    transition: transform var(--motion-fast) var(--motion-ease), opacity var(--motion-fast) ease;
  }
  /* Hover never moves a control or its icon; only press feedback scales in place. */
  @media (hover: hover) {
    :where([data-motion='on']) .ui-slider:not([data-disabled]) .ui-slider-thumb:hover { transform: scale(1.12); }
  }
  :where([data-motion='on']) .ui-button:active:not(:disabled) { transform: scale(.97); }
  :where([data-motion='on']) :where(.ui-input, .ui-textarea, .ui-select-trigger, .ui-checkbox, .ui-switch, .ui-radio-group-item, .ui-tabs-trigger, .ui-segmented-item, .ui-calendar-day, .ui-pagination-link, .ui-dropdown-menu-item, .ui-select-item, .ui-dialog-close) {
    transition: opacity var(--motion-fast) ease;
  }
  :where([data-motion='on']) .ui-segmented[data-indicator-animate] > .ui-segmented-indicator { transition: transform var(--motion-normal) var(--motion-ease), width var(--motion-normal) var(--motion-ease), height var(--motion-normal) var(--motion-ease); }
  :where([data-motion='on']) .ui-tabs-list[data-indicator-animate] > .ui-tabs-indicator { transition: transform var(--motion-normal) var(--motion-ease), width var(--motion-normal) var(--motion-ease), height var(--motion-normal) var(--motion-ease); }
  :where([data-motion='on']) .ui-switch-thumb { transition: transform var(--motion-normal) var(--motion-spring); }
  :where([data-motion='on'][data-motion-style='material']) .ui-button:active:not(:disabled) { transform: scale(.94); }
  :where([data-motion='on'][data-motion-style='material']) .ui-button-ripple { animation: ui-material-ripple var(--motion-ripple) var(--motion-ease); }
  [data-motion='on'][data-motion-style='material'] .ui-checkbox-indicator[data-state='checked'],
  [data-motion='on'][data-motion-style='material'] .ui-radio-group-indicator[data-state='checked'] { animation-name: ui-material-indicator; animation-duration: var(--motion-normal); }
  :where([data-motion='on']) .ui-slider-thumb { transition: transform var(--motion-fast) var(--motion-spring); }
  :where([data-motion='on']) .ui-progress-indicator { transition: transform var(--motion-slow) var(--motion-ease); }
  :where([data-motion='on']) .ui-accordion-chevron { transition: transform var(--motion-normal) var(--motion-ease); }

  /* Radix Presence keeps closing content mounted until animationend. */
  :is(:where([data-motion='on']) .ui-dialog-overlay, .ui-dialog-overlay:where([data-motion='on']))[data-state='open'] { animation: ui-fade-in var(--motion-normal) ease; }
  :is(:where([data-motion='on']) .ui-dialog-overlay, .ui-dialog-overlay:where([data-motion='on']))[data-state='closed'] { animation: ui-fade-out var(--motion-fast) ease; }
  :is(:where([data-motion='on']) .ui-dialog-content, .ui-dialog-content:where([data-motion='on']))[data-state='open'] { animation: ui-dialog-in var(--motion-normal) var(--motion-ease); }
  :is(:where([data-motion='on']) .ui-dialog-content, .ui-dialog-content:where([data-motion='on']))[data-state='closed'] { animation: ui-dialog-out var(--motion-fast) ease; }
  :where([data-motion='on'] :is(.ui-select-content, .ui-popover-content, .ui-dropdown-menu-content, .ui-tooltip-content), :is(.ui-select-content, .ui-popover-content, .ui-dropdown-menu-content, .ui-tooltip-content)[data-motion='on']) {
    --popup-x: 0px;
    --popup-y: 0px;
    transform-origin: var(--radix-select-content-transform-origin, var(--radix-popover-content-transform-origin, var(--radix-dropdown-menu-content-transform-origin, var(--radix-tooltip-content-transform-origin, center))));
  }
  :where([data-motion='on'] :is(.ui-select-content, .ui-popover-content, .ui-dropdown-menu-content, .ui-tooltip-content), :is(.ui-select-content, .ui-popover-content, .ui-dropdown-menu-content, .ui-tooltip-content)[data-motion='on'])[data-side='bottom'] { --popup-y: -4px; }
  :where([data-motion='on'] :is(.ui-select-content, .ui-popover-content, .ui-dropdown-menu-content, .ui-tooltip-content), :is(.ui-select-content, .ui-popover-content, .ui-dropdown-menu-content, .ui-tooltip-content)[data-motion='on'])[data-side='top'] { --popup-y: 4px; }
  :where([data-motion='on'] :is(.ui-select-content, .ui-popover-content, .ui-dropdown-menu-content, .ui-tooltip-content), :is(.ui-select-content, .ui-popover-content, .ui-dropdown-menu-content, .ui-tooltip-content)[data-motion='on'])[data-side='left'] { --popup-x: 4px; }
  :where([data-motion='on'] :is(.ui-select-content, .ui-popover-content, .ui-dropdown-menu-content, .ui-tooltip-content), :is(.ui-select-content, .ui-popover-content, .ui-dropdown-menu-content, .ui-tooltip-content)[data-motion='on'])[data-side='right'] { --popup-x: -4px; }
  :where([data-motion='on'] :is(.ui-select-content, .ui-popover-content, .ui-dropdown-menu-content, .ui-tooltip-content), :is(.ui-select-content, .ui-popover-content, .ui-dropdown-menu-content, .ui-tooltip-content)[data-motion='on']):is([data-state='open'], [data-state='delayed-open'], [data-state='instant-open']) { animation: ui-popup-in var(--motion-normal) var(--motion-ease); }
  :where([data-motion='on'] :is(.ui-select-content, .ui-popover-content, .ui-dropdown-menu-content, .ui-tooltip-content), :is(.ui-select-content, .ui-popover-content, .ui-dropdown-menu-content, .ui-tooltip-content)[data-motion='on'])[data-state='closed'] { animation: ui-popup-out var(--motion-fast) ease; }
  :where([data-motion='on']) .ui-alert .ui-alert-dismiss > svg:last-child { transition: opacity var(--motion-fast) ease; }
  :where([data-motion='on']) .ui-alert-stack-entry { transition: transform var(--motion-normal) var(--motion-ease); }
  :where([data-motion='on']) .ui-alert[data-state='open'] { animation: ui-alert-in var(--motion-normal) var(--motion-ease); }
  :where([data-motion='on']) .ui-alert[data-state='closed'] { animation: ui-alert-out var(--motion-fast) ease; }
  :where([data-motion='on']) :is(.ui-alert-viewport[data-placement='bottom'], .ui-alert-viewport[data-placement='bottom-start'], .ui-alert-viewport[data-placement='bottom-end']) .ui-alert[data-state='open'] { animation-name: ui-alert-in-up; }
  :where([data-motion='on']) :is(.ui-alert-viewport[data-placement='bottom'], .ui-alert-viewport[data-placement='bottom-start'], .ui-alert-viewport[data-placement='bottom-end']) .ui-alert[data-state='closed'] { animation-name: ui-alert-out-down; }
  /* Dismissed alerts sweep off sideways toward the edge they are anchored to; centered placements keep the vertical exit. */
  :where([data-motion='on']) :is(.ui-alert-viewport[data-placement='inline'], .ui-alert-viewport[data-placement='top-end'], .ui-alert-viewport[data-placement='bottom-end']) .ui-alert[data-state='closed'] { animation: ui-alert-out-end var(--motion-normal) cubic-bezier(.4, 0, 1, 1) forwards; }
  :where([data-motion='on']) :is(.ui-alert-viewport[data-placement='top-start'], .ui-alert-viewport[data-placement='bottom-start']) .ui-alert[data-state='closed'] { animation: ui-alert-out-start var(--motion-normal) cubic-bezier(.4, 0, 1, 1) forwards; }
  /* Scroll containers clip on both axes; release them while a card sweeps out so it can leave through the screen edge. */
  :where([data-motion='on']) :is(.ui-alert-viewport[data-placement='inline'], .ui-alert-viewport[data-placement$='-end'], .ui-alert-viewport[data-placement$='-start']):has(.ui-alert[data-state='closed']),
  :where([data-motion='on']) :is(.ui-alert-viewport[data-placement$='-end'], .ui-alert-viewport[data-placement$='-start']):has(.ui-alert[data-state='closed']) .ui-alert-stack-list { overflow: visible; }
  :where([data-motion='on']) .ui-accordion-content[data-state='open'] { animation: ui-accordion-open var(--motion-normal) var(--motion-ease); }
  :where([data-motion='on']) .ui-accordion-content[data-state='closed'] { animation: ui-accordion-close var(--motion-normal) var(--motion-ease); }
  :where([data-motion='on']) .ui-checkbox-indicator[data-state='checked'],
  :where([data-motion='on']) .ui-radio-group-indicator[data-state='checked'] { animation: ui-indicator-in var(--motion-fast) var(--motion-ease); }
  :where([data-motion='on']) .ui-checkbox-indicator[data-state='unchecked'],
  :where([data-motion='on']) .ui-radio-group-indicator[data-state='unchecked'] { animation: ui-indicator-out var(--motion-fast) ease; }
  :where([data-motion='on']) .ui-tabs-content[data-state='active'] { animation: ui-content-in var(--motion-normal) var(--motion-ease); }
  :where([data-motion='on']) .ui-skeleton { animation: ui-pulse 1.8s ease-in-out infinite; }
  :where([data-motion='on']) .ui-progress[data-state='indeterminate'] .ui-progress-indicator { animation: ui-progress-loading 1.5s ease-in-out infinite; }
  :where([data-motion='on']) .ui-button .lucide-loader-circle { animation: ui-spin .9s linear infinite; }
}

/* OFF and accessibility preferences also stop animations already in flight.
   A nested OFF region intentionally takes precedence over an outer ON region. */
:where([data-motion='off']):where([class^='ui-'], [class*=' ui-']),
:where([data-motion='off']) :where([class^='ui-'], [class*=' ui-'], .ui-button svg) {
  animation: none !important;
  transition: none !important;
}
:where([data-motion='off']) :where(.ui-button, .ui-button .lucide-arrow-right, .ui-slider-thumb) { transform: none !important; }
@media (prefers-reduced-motion: reduce) {
  :where([class^='ui-'], [class*=' ui-'], .ui-button svg) { animation: none !important; transition: none !important; }
}
@keyframes ui-fade-in { from { opacity: 0; } to { opacity: 1; } }
@keyframes ui-fade-out { from { opacity: 1; } to { opacity: 0; } }
@keyframes ui-dialog-in { from { opacity: 0; transform: translate(-50%, calc(-50% + var(--motion-dialog-shift))) scale(var(--motion-dialog-scale)); } to { opacity: 1; transform: translate(-50%, -50%) scale(1); } }
@keyframes ui-dialog-out { from { opacity: 1; transform: translate(-50%, -50%) scale(1); } to { opacity: 0; transform: translate(-50%, calc(-50% + 4px)) scale(.98); } }
@keyframes ui-popup-in { from { opacity: 0; transform: translate(var(--popup-x), var(--popup-y)) scale(var(--motion-popup-scale)); } to { opacity: 1; transform: translate(0, 0) scale(1); } }
@keyframes ui-popup-out { from { opacity: 1; transform: translate(0, 0) scale(1); } to { opacity: 0; transform: translate(var(--popup-x), var(--popup-y)) scale(.98); } }
@keyframes ui-accordion-open { from { height: 0; } to { height: var(--radix-accordion-content-height); } }
@keyframes ui-accordion-close { from { height: var(--radix-accordion-content-height); } to { height: 0; } }
@keyframes ui-indicator-in { from { opacity: 0; transform: scale(.65); } to { opacity: 1; transform: scale(1); } }
@keyframes ui-indicator-out { from { opacity: 1; transform: scale(1); } to { opacity: 0; transform: scale(.65); } }
@keyframes ui-alert-in { from { opacity: 0; transform: translateY(calc(-1 * var(--motion-dialog-shift))); } to { opacity: 1; transform: translateY(0); } }
@keyframes ui-alert-out { from { opacity: 1; transform: translateY(0); } to { opacity: 0; transform: translateY(-4px); } }
/* Bottom placements rise from below instead of dropping from above. */
@keyframes ui-alert-in-up { from { opacity: 0; transform: translateY(var(--motion-dialog-shift)); } to { opacity: 1; transform: translateY(0); } }
@keyframes ui-alert-out-down { from { opacity: 1; transform: translateY(0); } to { opacity: 0; transform: translateY(4px); } }
@keyframes ui-alert-out-end { from { opacity: 1; transform: translateX(0); } to { opacity: 0; transform: translateX(calc(100% + 24px)); } }
@keyframes ui-alert-out-start { from { opacity: 1; transform: translateX(0); } to { opacity: 0; transform: translateX(calc(-100% - 24px)); } }
@keyframes ui-content-in { from { opacity: 0; transform: translate(var(--tabs-enter-x, 8px), var(--tabs-enter-y, 0px)); } to { opacity: 1; transform: translate(0, 0); } }
@keyframes ui-material-ripple { 0% { opacity: .24; transform: translate(-50%, -50%) scale(0); } 45% { opacity: .16; } 100% { opacity: 0; transform: translate(-50%, -50%) scale(1); } }
@keyframes ui-material-indicator { 0% { opacity: 0; transform: scale(.4); } 65% { opacity: 1; transform: scale(1.15); } 100% { opacity: 1; transform: scale(1); } }
@keyframes ui-pulse { 50% { opacity: .45; } }
@keyframes ui-progress-loading { 0% { transform: translateX(-100%) scaleX(.35); } 100% { transform: translateX(100%) scaleX(.35); } }
@keyframes ui-spin { to { transform: rotate(360deg); } }

/* Navbar indicator + menu, NoticeBar flow. NoticeBar also gates itself in script (data-animated). */
@media (prefers-reduced-motion: no-preference) {
  :where([data-motion='on']) .ui-navbar-content[data-indicator-animate] > .ui-navbar-indicator { transition: transform var(--motion-normal) var(--motion-ease), width var(--motion-normal) var(--motion-ease), height var(--motion-normal) var(--motion-ease); }
  :where([data-motion='on']) .ui-navbar-content[data-state='open'] { animation: ui-content-in var(--motion-normal) var(--motion-ease); --tabs-enter-x: 0px; --tabs-enter-y: -6px; }
  :where([data-motion='on']) .ui-notice-bar[data-animated][data-effect='marquee'] .ui-notice-bar-track { animation: ui-notice-marquee var(--notice-duration, 20s) linear infinite; }
  :where([data-motion='on']) .ui-notice-bar[data-animated][data-effect='marquee']:is([data-direction='right'], [data-direction='down']) .ui-notice-bar-track { animation-direction: reverse; }
  :where([data-motion='on']) .ui-notice-bar[data-animated][data-effect='marquee'] .ui-notice-bar-track:dir(rtl) { animation-name: ui-notice-marquee-rtl; }
  :where([data-motion='on']) .ui-notice-bar[data-animated][data-effect='marquee']:is([data-direction='up'], [data-direction='down']) .ui-notice-bar-track { animation-name: ui-notice-marquee-vertical; }
  :where([data-motion='on']) .ui-notice-bar[data-animated][data-effect='marquee'][data-holding] .ui-notice-bar-track { animation-play-state: paused; }
  :where([data-motion='on']) .ui-notice-bar-list[data-rotated] > .ui-notice-bar-item[data-state='active'] { animation: ui-notice-in var(--motion-slow) var(--motion-ease); }
  :where([data-motion='on']) .ui-notice-bar-list[data-rotated] > .ui-notice-bar-item[data-state='leaving'] { animation: ui-notice-out var(--motion-slow) var(--motion-ease) forwards; }
}
@keyframes ui-notice-marquee-vertical { from { transform: translateY(0); } to { transform: translateY(-50%); } }
@keyframes ui-notice-marquee { from { transform: translateX(0); } to { transform: translateX(-50%); } }
@keyframes ui-notice-marquee-rtl { from { transform: translateX(0); } to { transform: translateX(50%); } }
@keyframes ui-notice-in { from { opacity: 0; transform: translate(var(--notice-enter-x, 0px), var(--notice-enter-y, 0px)); } to { opacity: 1; transform: translate(0, 0); } }
@keyframes ui-notice-out { from { opacity: 1; visibility: visible; transform: translate(0, 0); } to { opacity: 0; visibility: hidden; transform: translate(var(--notice-leave-x, 0px), var(--notice-leave-y, 0px)); } }

</style>
<style data-dohwi-ui="native">
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
  #recordsTableMount .ui-table-head:nth-child(4) { width: 20%; }
  #recordsTableMount .ui-table-head:nth-child(5), #recordsTableMount .ui-table-head:nth-child(6), #recordsTableMount .ui-table-head:nth-child(7), #recordsTableMount .ui-table-head:nth-child(8) { width: 9%; }
  #recordsTableMount .ui-table-head:nth-child(9) { width: 7%; }
  #recordsTableMount .ui-table-head:nth-child(10) { width: 5%; }
  @media (max-width: 1199px) and (min-width: 768px) {
    #recordsTableMount .ui-table { min-width: 680px; }
    #recordsPanel .ui-table-head:first-child { width: 9%; }
    #recordsPanel .ui-table-head:nth-child(2) { width: 12%; }
    #recordsPanel .ui-table-head:nth-child(4) { width: 20%; }
    #recordsPanel .ui-table-head:nth-child(5), #recordsPanel .ui-table-head:nth-child(6), #recordsPanel .ui-table-head:nth-child(7) { width: 12%; }
    #recordsPanel .ui-table-head:nth-child(8) { width: 15%; }
    #recordsPanel .ui-table-head:nth-child(10) { width: 8%; }
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
</style>
<style data-dohwi-motion="material">
  /* Material motion tokens copied from dohwi-ui/src/components/ui/motion.css.
     Only transform and opacity are animated; colors, borders and sizes never change. */
  :root {
    --motion-fast: 180ms; --motion-normal: 360ms; --motion-slow: 460ms; --motion-ripple: 620ms;
    --motion-ease: cubic-bezier(.22, 1, .36, 1); --motion-spring: cubic-bezier(.34, 1.56, .64, 1);
    --motion-dialog-shift: 18px; --motion-dialog-scale: .94; --motion-popup-scale: .9;
  }
  .ui-button-ripple-layer { position: absolute; inset: 0; border-radius: inherit; overflow: hidden; pointer-events: none; }
  .ui-button-ripple { position: absolute; border-radius: 50%; background: currentColor; opacity: 0; pointer-events: none; transform: translate(-50%, -50%) scale(0); }
  .ui-ripple-host { position: relative; overflow: hidden; }
  @media (prefers-reduced-motion: no-preference) {
    [data-indicator-animate] > .ui-segmented-indicator, [data-indicator-animate] > .ui-tabs-indicator { transition: transform var(--motion-normal) var(--motion-ease), width var(--motion-normal) var(--motion-ease), height var(--motion-normal) var(--motion-ease); }
    #breakdownSection[data-tabs-direction] .tab-pane.active { animation: ui-content-in var(--motion-normal) var(--motion-ease); }
    .ui-ripple-host:active:not(:disabled) { transform: scale(.96); }
    .ui-ripple-host { transition: transform var(--motion-fast) var(--motion-ease); }
    .ui-ripple-host .ui-button-ripple { animation: ui-material-ripple var(--motion-ripple) var(--motion-ease); }
    .is-entering { animation: ui-enter-up var(--motion-slow) var(--motion-ease) both; animation-delay: var(--enter-delay, 0ms); }
    .ui-select-content[data-state='open'], .ui-date-picker-content[data-state='open'] { animation: ui-popup-in var(--motion-normal) var(--motion-ease); }
    .ui-select-item-indicator { animation: ui-material-indicator var(--motion-normal) var(--motion-ease); }
    .compact-select-wrap::after { transition: transform var(--motion-normal) var(--motion-spring); }
    @keyframes ui-enter-up { from { opacity: 0; transform: translate(0, 14px); } to { opacity: 1; transform: translate(0, 0); } }
    @keyframes ui-content-in { from { opacity: 0; transform: translate(var(--tabs-enter-x, 18px), 0); } to { opacity: 1; transform: translate(0, 0); } }
    @keyframes ui-popup-in { from { opacity: 0; transform: translate(0, var(--popup-y, -4px)) scale(var(--motion-popup-scale)); } to { opacity: 1; transform: translate(0, 0) scale(1); } }
    @keyframes ui-material-ripple { 0% { opacity: .24; transform: translate(-50%, -50%) scale(0); } 45% { opacity: .16; } 100% { opacity: 0; transform: translate(-50%, -50%) scale(1); } }
    @keyframes ui-material-indicator { 0% { opacity: 0; transform: scale(.4); } 65% { opacity: 1; transform: scale(1.15); } 100% { opacity: 1; transform: scale(1); } }
    #usageDataContent .kpi-group.is-entering, #usageDataContent #breakdownSection.is-entering, #usageDataContent #recordsPanel.is-entering, #usageDataContent #chartCard.is-entering { will-change: transform, opacity; }
    .record-dialog { transition: opacity var(--motion-fast) ease, transform var(--motion-normal) var(--motion-spring); transform: translateY(var(--motion-dialog-shift)) scale(var(--motion-dialog-scale)); }
    .detail-overlay.open .record-dialog { transform: translateY(0) scale(1); }
  }
</style>
<style data-dashboard-ui="adapter">
  :root { --border: var(--filter-control-border); --radius: var(--filter-control-radius); --background: var(--bg-card); --surface: var(--bg-card); --foreground: var(--text-primary); --muted: var(--surface-toolbar); --muted-foreground: var(--text-secondary); --shadow-popover: 0 4px 16px #18223014; }
  /* Completely neutralize Radix react-remove-scroll so it cannot hide the scrollbar or shift the layout */
  html body[data-scroll-locked] {
    overflow: visible !important;
    overflow-y: visible !important;
    position: static !important;
    margin: 0 !important;
    margin-right: 0 !important;
    padding-top: 24px !important;
    padding-right: 14px !important;
    padding-bottom: 16px !important;
    padding-left: 14px !important;
    --removed-body-scroll-bar-size: 0px !important;
  }
  @media (max-width: 1199px) {
    html body[data-scroll-locked] { padding-top: 20px !important; }
  }
  @media (max-width: 767px) {
    html body[data-scroll-locked] {
      padding-top: 16px !important;
      padding-right: 10px !important;
      padding-bottom: 12px !important;
      padding-left: 10px !important;
    }
  }
  html .with-scroll-bars-hidden {
    overflow: visible !important;
    padding-right: 0 !important;
  }
  .dashboard-preset-mount { min-width: 0; }
  #presetGroupControl { width: 100%; }
  #presetGroupControl .ui-segmented-item, #presetGroupControl .preset-btn {
    min-height: var(--filter-control-height);
    background: transparent !important;
    border-color: transparent !important;
    box-shadow: none !important;
  }
  #presetGroupControl .ui-segmented-item:hover, #presetGroupControl .preset-btn:hover {
    background: transparent !important;
  }
  #presetGroupControl .ui-segmented-item[data-state='checked'], #presetGroupControl .preset-btn.active {
    background: transparent !important;
    border-color: transparent !important;
  }
  #customDateWrap { position: relative; display: block; min-width: 0; }
  #dateRangeMount, #dateRangeMount .ui-date-picker-trigger { width: 100%; }
  #dateRangeMount .ui-date-picker-trigger { height: var(--filter-control-height); font-size: .875rem; }
  .ui-calendar { font-family: inherit; }
  .ui-calendar-grid { min-width: 0; }
  .ui-calendar-weekday { border-bottom: 0; background: transparent; }
  .ui-calendar-cell { vertical-align: middle; }
  #customDateWrap > input[hidden], .compact-select-wrap > select[hidden] { display: none; }
  .compact-select-wrap::after { display: none; }
  .ui-select-trigger { display: flex; padding-right: 10px; }
  .ui-select-trigger > span { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  .ui-select-content { max-height: var(--radix-select-content-available-height); }
  .ui-select-viewport { max-height: min(320px, var(--radix-select-content-available-height)); }
  .ui-popover-content { box-sizing: border-box; }
  .ui-calendar-cell { border-bottom: 0; }
  @media (max-width: 600px), (pointer: coarse) {
    #presetGroupControl .ui-segmented-item, #dateRangeMount .ui-date-picker-trigger, .ui-select-item { min-height: 44px; }
  }
</style>
</head>
<body>
<div class="container">
  <div id="errorBox" class="message-box error-box is-hidden" role="alert" aria-live="assertive">
    <span id="errorMessage"></span>
    <button id="retryLoadBtn" type="button" class="btn btn-sm btn-secondary">다시 시도</button>
  </div>

  <header class="page-heading">
    <div class="page-heading-copy"><h1>토큰 사용량</h1><p>요청과 토큰 사용량을 한눈에 확인하세요.</p></div>
  </header>
  <div id="dashboardContent" aria-busy="true">
    <!-- Filter Panel -->
    <div class="card filter-card">
      <form id="filterForm" class="filter-panel" novalidate>
        <div class="filter-topline">
          <fieldset class="filter-section filter-period-section">
            <legend class="sr-only">기간 선택</legend>
            <div class="filter-row filter-period-row">
              <div id="presetGroup" class="dashboard-preset-mount" aria-label="기간 선택">
                <span class="ui-segmented-indicator" aria-hidden="true"></span>
                <button type="button" class="preset-btn ui-segmented-item active" role="radio" aria-checked="true" data-state="checked" tabindex="0" data-preset="today">오늘</button>
                <button type="button" class="preset-btn ui-segmented-item" role="radio" aria-checked="false" data-state="unchecked" tabindex="-1" data-preset="yesterday">어제</button>
                <button type="button" class="preset-btn ui-segmented-item" role="radio" aria-checked="false" data-state="unchecked" tabindex="-1" data-preset="7d">7일</button>
                <button type="button" class="preset-btn ui-segmented-item" role="radio" aria-checked="false" data-state="unchecked" tabindex="-1" data-preset="30d">30일</button>
                <button type="button" class="preset-btn ui-segmented-item" role="radio" aria-checked="false" data-state="unchecked" tabindex="-1" data-preset="month">이번 달</button>
                <button type="button" class="preset-btn ui-segmented-item" role="radio" aria-checked="false" data-state="unchecked" tabindex="-1" data-preset="custom" aria-controls="customDateWrap">직접</button>
              </div>
              <div id="customDateWrap" class="custom-dates" aria-label="조회 기간" title="직접을 선택하면 날짜를 변경할 수 있습니다.">
                <div id="dateRangeMount"></div>
                <input type="date" id="filterFromDate" class="input" aria-label="시작일" hidden disabled>
                <input type="date" id="filterToDate" class="input" aria-label="종료일" hidden disabled>
              </div>
            </div>
          </fieldset>
          <div class="filter-topmeta">
            <div id="activeFilterBadge" class="active-filter-badge" aria-label="적용 조건" role="status" aria-live="polite" tabindex="0">
              <span class="active-filter-label">조회 조건</span>
              <span id="activeFilterText" class="active-filter-text">오늘 · 전체 Provider · 전체 Model</span>
            </div>
          </div>
        </div>

        <fieldset class="filter-section filter-detail-section">
          <legend class="sr-only">세부 필터</legend>
          <div class="filter-row filter-fields-row">
            <div class="filter-field">
              <label for="filterProvider">Provider</label>
              <div class="compact-select-wrap">
                <select id="filterProvider" class="select compact-select" aria-label="Provider 필터"><option value="">전체</option></select>
              </div>
            </div>
            <div class="filter-field filter-endpoint-field">
              <label for="filterEndpoint">Endpoint</label>
              <div class="compact-select-wrap">
                <select id="filterEndpoint" class="select compact-select" aria-label="Endpoint 필터"><option value="">전체</option></select>
              </div>
            </div>
            <div class="filter-field filter-model-field">
              <label for="filterModel">Model</label>
              <div class="compact-select-wrap">
                <select id="filterModel" class="select compact-select" aria-label="Model 필터"><option value="">전체</option></select>
              </div>
            </div>
            <div class="filter-field filter-status-field">
              <label for="filterStatus">상태</label>
              <div class="compact-select-wrap">
                <select id="filterStatus" class="select compact-select" aria-label="상태 필터">
                  <option value="">전체</option>
                  <option value="success">성공만</option>
                  <option value="failed">실패만</option>
                </select>
              </div>
            </div>
            <div class="filter-reset-field">
              <button type="button" class="btn btn-sm btn-secondary btn-reset filter-reset-button" id="resetFiltersBtn" aria-label="모든 세부 필터 초기화">초기화</button>
            </div>
          </div>
        </fieldset>
      </form>
    </div>

    <div id="usageSkeleton" class="usage-skeleton is-visible" aria-hidden="true">
      <div class="skeleton-kpi-groups">
        <div class="skeleton-kpi-group"><div class="skeleton-kpi-head"><span class="skeleton-line"></span><span class="skeleton-line"></span></div><div class="skeleton-kpi-rows"><div class="skeleton-kpi-row"><span class="skeleton-line"></span><span class="skeleton-line"></span></div><div class="skeleton-kpi-row"><span class="skeleton-line"></span><span class="skeleton-line"></span></div></div></div>
        <div class="skeleton-kpi-group"><div class="skeleton-kpi-head"><span class="skeleton-line"></span><span class="skeleton-line"></span></div><div class="skeleton-kpi-rows"><div class="skeleton-kpi-row"><span class="skeleton-line"></span><span class="skeleton-line"></span></div><div class="skeleton-kpi-row"><span class="skeleton-line"></span><span class="skeleton-line"></span></div><div class="skeleton-kpi-row"><span class="skeleton-line"></span><span class="skeleton-line"></span></div></div></div>
        <div class="skeleton-kpi-group"><div class="skeleton-kpi-head"><span class="skeleton-line"></span><span class="skeleton-line"></span></div><div class="skeleton-kpi-rows"><div class="skeleton-kpi-row"><span class="skeleton-line"></span><span class="skeleton-line"></span></div><div class="skeleton-kpi-row"><span class="skeleton-line"></span><span class="skeleton-line"></span></div></div></div>
      </div>
      <div class="skeleton-main-grid">
        <div class="skeleton-panel skeleton-panel-left"><div class="skeleton-panel-head"><span class="skeleton-line"></span><span class="skeleton-line"></span></div><div class="skeleton-chart-bars"><span class="skeleton-line"></span><span class="skeleton-line"></span><span class="skeleton-line"></span><span class="skeleton-line"></span><span class="skeleton-line"></span><span class="skeleton-line"></span><span class="skeleton-line"></span></div><div class="skeleton-table"><span class="skeleton-line"></span><div class="skeleton-table-row"><span class="skeleton-line"></span><span class="skeleton-line"></span><span class="skeleton-line"></span><span class="skeleton-line"></span></div><div class="skeleton-table-row"><span class="skeleton-line"></span><span class="skeleton-line"></span><span class="skeleton-line"></span><span class="skeleton-line"></span></div></div></div>
        <div class="skeleton-panel skeleton-panel-right"><div class="skeleton-panel-head"><span class="skeleton-line"></span><span class="skeleton-line"></span></div><div class="skeleton-table"><div class="skeleton-table-row"><span class="skeleton-line"></span><span class="skeleton-line"></span><span class="skeleton-line"></span><span class="skeleton-line"></span></div><span class="skeleton-line"></span><div class="skeleton-table-row"><span class="skeleton-line"></span><span class="skeleton-line"></span><span class="skeleton-line"></span><span class="skeleton-line"></span></div><div class="skeleton-table-row"><span class="skeleton-line"></span><span class="skeleton-line"></span><span class="skeleton-line"></span><span class="skeleton-line"></span></div><div class="skeleton-table-row"><span class="skeleton-line"></span><span class="skeleton-line"></span><span class="skeleton-line"></span><span class="skeleton-line"></span></div><div class="skeleton-table-row"><span class="skeleton-line"></span><span class="skeleton-line"></span><span class="skeleton-line"></span><span class="skeleton-line"></span></div></div></div>
      </div>
    </div>

    <div id="usageDataContent" class="is-loading">
    <!-- KPI semantic groups: Overview / Input / Output (flat panels, no nested cards) -->
    <div class="kpi-groups">
      <!-- Overview group -->
      <section class="kpi-group kpi-group-overview" role="group" aria-labelledby="grpOverviewTitle" aria-describedby="grpOverviewHelp">
        <div class="kpi-group-head">
          <div class="kpi-group-title-wrap">
            <svg class="kpi-group-icon" viewBox="0 0 16 16" fill="none" xmlns="http://www.w3.org/2000/svg" aria-hidden="true">
              <rect x="1.5" y="1.5" width="13" height="13" rx="2" stroke="currentColor" stroke-width="1.25"/>
              <path d="M4 8h8M4 5.5h5M4 10.5h6" stroke="currentColor" stroke-width="1.25" stroke-linecap="round"/>
            </svg>
            <h2 class="kpi-group-title" id="grpOverviewTitle">개요</h2>
          </div>
          <div class="kpi-group-help sr-only" id="grpOverviewHelp">공식 집계 기준 요청·전체 수치</div>
        </div>
        <div class="kpi-group-metrics">
          <!-- 총 요청 row -->
          <div class="kpi-metric-row kpi-card kpi-accent kpi-request">
            <div class="kpi-metric-label-wrap">
              <span class="kpi-label">총 요청</span>
              <span class="kpi-sub" id="kpiSubFail" tabindex="0">실패: -건 (-)</span>
            </div>
            <div class="kpi-value-wrap">
              <span class="kpi-value kpi-value-request" id="kpiTotalCount" title="0" aria-label="총 요청: 0">-</span>
            </div>
          </div>
          <!-- 전체 토큰 row -->
          <div class="kpi-metric-row kpi-card kpi-accent kpi-total">
            <div class="kpi-metric-label-wrap">
              <span class="kpi-label">전체 토큰</span>
              <span class="kpi-sub" id="kpiSubTokens">공식 집계</span>
            </div>
            <div class="kpi-value-wrap">
              <span class="kpi-value kpi-value-total" id="kpiTotalTokens" title="0" aria-label="전체 토큰: 0">-</span>
            </div>
          </div>
        </div>
        <!-- Hidden/compat IDs preserved for tests -->
        <span class="kpi-compat-hidden" id="kpiFailureRate">-</span>
        <span class="kpi-compat-hidden" id="kpiSubReq">-</span>
        <span class="kpi-compat-hidden" id="kpiCardEntities">
          <span id="kpiActiveEntities">-</span>
          <span id="kpiSubEntities">-</span>
        </span>
      </section>

      <!-- Input group -->
      <section class="kpi-group kpi-group-input" role="group" aria-labelledby="grpInputTitle" aria-describedby="grpInputHelp">
        <div class="kpi-group-head">
          <div class="kpi-group-title-wrap">
            <svg class="kpi-group-icon" viewBox="0 0 16 16" fill="none" xmlns="http://www.w3.org/2000/svg" aria-hidden="true">
              <path d="M8 3v10M4 9l4 4 4-4" stroke="currentColor" stroke-width="1.25" stroke-linecap="round" stroke-linejoin="round"/>
            </svg>
            <h2 class="kpi-group-title" id="grpInputTitle">입력</h2>
          </div>
          <div class="kpi-group-help sr-only" id="grpInputHelp">Provider 보고 입력과 캐시 구성</div>
        </div>
        <div class="kpi-group-metrics">
          <!-- 입력 (Provider 보고값) row -->
          <div class="kpi-metric-row kpi-card kpi-accent kpi-input-report">
            <div class="kpi-metric-label-wrap">
              <span class="kpi-label">입력 (보고값)</span>
              <span class="kpi-sub" title="Provider 보고값으로 집계 방식에 따라 캐시 포함 여부가 다릅니다.">Provider 보고 기준</span>
            </div>
            <div class="kpi-value-wrap">
              <span class="kpi-value kpi-value-input-report" id="kpiInputTokens" title="0" aria-label="입력 보고값: 0">-</span>
            </div>
          </div>
          <!-- 신규 입력 row -->
          <div class="kpi-metric-row kpi-card kpi-accent kpi-input">
            <div class="kpi-metric-label-wrap">
              <span class="kpi-label">신규 입력</span>
              <span class="kpi-sub">캐시 제외</span>
            </div>
            <div class="kpi-value-wrap">
              <span class="kpi-value kpi-value-input" id="kpiFreshInputTokens" title="0" aria-label="신규 입력: 0">-</span>
            </div>
          </div>
          <!-- 캐시 입력 row -->
          <div class="kpi-metric-row kpi-card kpi-accent kpi-cache">
            <div class="kpi-metric-label-wrap">
              <span class="kpi-label">캐시 입력</span>
              <span class="kpi-sub" title="Provider 방식에 따라 캐시 입력은 입력에 포함되거나 별도로 집계됩니다.">재사용된 입력</span>
            </div>
            <div class="kpi-value-wrap">
              <span class="kpi-value kpi-value-cache" id="kpiCachedTokens" title="0" aria-label="캐시 입력: 0">-</span>
            </div>
          </div>
        </div>
      </section>

      <!-- Output group -->
      <section class="kpi-group kpi-group-output" role="group" aria-labelledby="grpOutputTitle" aria-describedby="grpOutputHelp">
        <div class="kpi-group-head">
          <div class="kpi-group-title-wrap">
            <svg class="kpi-group-icon" viewBox="0 0 16 16" fill="none" xmlns="http://www.w3.org/2000/svg" aria-hidden="true">
              <path d="M8 13V3M4 7l4-4 4 4" stroke="currentColor" stroke-width="1.25" stroke-linecap="round" stroke-linejoin="round"/>
            </svg>
            <h2 class="kpi-group-title" id="grpOutputTitle">출력</h2>
          </div>
          <div class="kpi-group-help sr-only" id="grpOutputHelp">모델 생성 및 추론 토큰</div>
        </div>
        <div class="kpi-group-metrics">
          <!-- 출력 row -->
          <div class="kpi-metric-row kpi-card kpi-accent kpi-output">
            <div class="kpi-metric-label-wrap">
              <span class="kpi-label">출력</span>
              <span class="kpi-sub">모델 생성</span>
            </div>
            <div class="kpi-value-wrap">
              <span class="kpi-value kpi-value-output" id="kpiOutputTokens" title="0" aria-label="출력 토큰: 0">-</span>
            </div>
          </div>
          <!-- 추론 row -->
          <div class="kpi-metric-row kpi-card kpi-accent kpi-reason">
            <div class="kpi-metric-label-wrap">
              <span class="kpi-label">추론</span>
              <span class="kpi-sub" title="추론은 출력 또는 전체 토큰에 포함될 수 있어 중복 합산하지 않습니다.">중복 합산 제외</span>
            </div>
            <div class="kpi-value-wrap">
              <span class="kpi-value kpi-value-reason" id="kpiReasoningTokens" title="0" aria-label="추론 토큰: 0">-</span>
            </div>
          </div>
        </div>
      </section>
    </div>

    <!-- Desktop Main Layout (12 cols: left 4 / right 8) -->
    <div id="usageMainContent" class="dashboard-main-grid">

      <!-- Left 4 Columns: Breakdown Tabs and Daily Trend -->
      <div class="col-left-5">
        <!-- Aggregations breakdown tabs (Provider / Endpoint / Model) -->
        <div class="card" id="breakdownSection">
          <div class="card-header panel-heading">
            <div class="panel-heading-title">
              <h2 class="card-title">상세 집계</h2>
              <div class="card-desc">필터 기준별 사용량</div>
            </div>
            <div class="tab-nav ui-tabs-list" role="tablist" aria-label="상세 집계 기준">
              <span class="ui-tabs-indicator" aria-hidden="true"></span>
              <button id="tab-provider-tab" type="button" class="tab-btn ui-tabs-trigger active" data-state="active" onclick="switchTab('tab-provider', this)" role="tab" aria-selected="true" aria-controls="tab-provider" tabindex="0">Provider</button>
              <button id="tab-endpoint-tab" type="button" class="tab-btn ui-tabs-trigger" data-state="inactive" onclick="switchTab('tab-endpoint', this)" role="tab" aria-selected="false" aria-controls="tab-endpoint" tabindex="-1">Endpoint</button>
              <button id="tab-model-tab" type="button" class="tab-btn ui-tabs-trigger" data-state="inactive" onclick="switchTab('tab-model', this)" role="tab" aria-selected="false" aria-controls="tab-model" tabindex="-1">Model</button>
            </div>
          </div>

          <!-- Tab Provider -->
          <div id="tab-provider" class="tab-pane active" role="tabpanel" aria-labelledby="tab-provider-tab">
            <div id="providerAggTableMount"></div><!-- dohwi-ui TableBody: <tbody id="providerAggBody"></tbody> -->
          </div>

          <!-- Tab Endpoint -->
          <div id="tab-endpoint" class="tab-pane" role="tabpanel" aria-labelledby="tab-endpoint-tab">
            <div id="endpointAggTableMount"></div><!-- dohwi-ui TableBody: <tbody id="endpointAggBody"></tbody> -->
          </div>

          <!-- Tab Model -->
          <div id="tab-model" class="tab-pane" role="tabpanel" aria-labelledby="tab-model-tab">
            <div id="modelAggTableMount"></div><!-- dohwi-ui TableBody: <tbody id="modelAggBody"></tbody> -->
          </div>
        </div>

        <section id="chartCard" class="card chart-card-compact" aria-labelledby="chartTitle">
          <header class="card-header chart-heading">
            <h2 class="card-title" id="chartTitle">사용량 추이</h2>
          </header>
          <div id="dailyChartSummary" class="card-desc" tabindex="0">일별 요청량</div>
          <div class="chart-wrapper" id="chartArea"></div>
          <div id="dailyChartTooltip" class="daily-chart-tooltip" role="tooltip" hidden></div>
        </section>
      </div>

      <!-- Right 8 Columns: Latest 20 Table with Sticky Header and Internal Scroll -->
      <div class="col-right-7">
        <div class="card" id="recordsPanel">
          <div class="card-header panel-heading">
            <div class="panel-heading-title">
              <h2 class="card-title">
                <span>최근 요청 기록</span>
                <span id="recordsCountBadge" class="badge badge-tag">0건</span>
              </h2>
              <div class="card-desc">최신 20건 · 필터와 더보기로 전체 로그 추적</div>
              <p class="records-mobile-hint">기록을 누르면 전체 토큰·Provider·Endpoint를 확인할 수 있습니다.</p>
            </div>
          </div>
          <div id="recordsTableMount"></div><!-- dohwi-ui TableBody: <tbody id="recordsBody"></tbody> -->
          <div class="pagination-bar is-hidden" id="loadMoreBar">
            <button id="loadMoreBtn" class="btn btn-secondary" onclick="loadMoreRecords()">기록 더보기</button>
          </div>
        </div>
      </div>
    </div>
    </div>
  </div>
</div>

<!-- Request detail modal -->
<div id="detailOverlay" class="detail-overlay" aria-hidden="true">
  <section id="recordDialog" class="record-dialog" aria-hidden="true" role="dialog" aria-modal="true" aria-labelledby="recordDialogTitle">
    <div class="dialog-drag-handle" aria-hidden="true"></div>
    <div class="dialog-header">
      <div id="recordDialogTitle" class="dialog-title">요청 상세 정보</div>
      <button id="recordDialogClose" type="button" class="dialog-close" title="닫기" aria-label="닫기">&times;</button>
    </div>
    <div class="dialog-body" id="dialogBody">
      <!-- Populated dynamically; 호스트 observer에서 제공하지 않음 -->
    </div>
  </section>
</div>

<script>
// ── DOM cache: populated once at DOMContentLoaded, reused every refresh ──
var DOM = {};
var pendingDohwiTableUpdates = [];

// State management
var state = {
  currentPreset: 'today',
  fromRFC3339: '',
  toRFC3339: '',
  provider: '',
  endpoint: '',
  model: '',
  status: '',
  cursor: null,
  hasMore: false,
  records: [],
  recordsExpanded: false,
  drawerOpener: null,
  requestID: 0,
  pollTimer: null,
  inFlight: false,
  abortCtrl: null,
  hasLoadedData: false,
  serverSnapshot: null,
  snapshotFilterKey: '',
  pendingChanges: null
};

// Adaptive polling intervals (ms) for usage-audit dashboard
// today with records within 2m => 30s; activity today but older => 90s;
// quiet today => 5m; historical presets => 10m.
var UA_POLL_RECENT_MS  = 30000;   // today, latest record within 2 min
var UA_POLL_ACTIVE_MS  = 90000;   // today, activity but older
var UA_POLL_QUIET_MS   = 300000;  // today, no records
var UA_POLL_HISTORY_MS = 600000;  // historical preset
var UA_POLL_RETRY_MS   = 120000;  // retry after failure
var UA_POLL_MAX_MS     = 600000;  // 10m cap
var UA_FETCH_TIMEOUT_MS = 20000;   // requests must not leave the UI loading indefinitely

// Escapes text for the rare static HTML-string fallback. Dynamic API data is
// rendered with textContent/title instead, so quote characters never cross an
// HTML attribute boundary.
function esc(str) {
  if (str == null) return '';
  return String(str).replace(/[&<>"']/g, function(ch) {
    return { '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }[ch];
  });
}

// Exact number formatter: comma-grouped integer. KPI/records/breakdowns must
// NEVER use compact K/M/B notation; every figure is shown in full precision.
// null/undefined/NaN => '—'
// 0 => '0'
// negative => '-1,234'
// integer, grouped with commas (en-US) regardless of host locale
function formatExactNumber(num) {
  if (num === null || num === undefined) return '—';
  var n = Number(num);
  if (isNaN(n)) return '—';
  return Math.round(n).toLocaleString('en-US');
}

// Provider display mapper: trims/lowers and turns compatible-b.ai aliases into "b.ai"
function formatProvider(val) {
  if (!val) return '-';
  var s = String(val).trim();
  var low = s.toLowerCase();
  if (low === 'openai compatible-b.ai' || low === 'compatible-b.ai' || low === 'openai-compatible-b.ai') {
    return 'b.ai';
  }
  return s;
}

function formatLatency(ms) {
  if (ms === null || ms === undefined || ms === '') return '-';
  var n = Number(ms);
  if (!Number.isFinite(n) || n < 0) return '-';
  if (n > 1000000) { // nanoseconds
    n = Math.round(n / 1000000);
  }
  if (n >= 1000) {
    return (n / 1000).toFixed(2) + 's';
  }
  return Math.round(n) + 'ms';
}

function formatDateTime(val) {
  if (!val) return '-';
  try {
    var d = new Date(val);
    if (!isNaN(d.getTime())) {
      return d.toLocaleString('ko-KR');
    }
  } catch (e) {}
  return String(val);
}

// Tabs switching in left panel
function switchTab(targetId, btn) {
  var panes = document.querySelectorAll('.tab-pane');
  var tabs = Array.prototype.slice.call(document.querySelectorAll('.tab-btn'));
  var section = document.getElementById('breakdownSection');
  var from = -1, to = tabs.indexOf(btn);
  tabs.forEach(function(t, i) { if (t.classList.contains('active')) from = i; });
  // Same contract as dohwi-ui Tabs: the new panel enters from the side it sits on.
  if (section && from >= 0 && to >= 0 && from !== to) {
    section.style.setProperty('--tabs-enter-x', (to > from ? 18 : -18) + 'px');
    section.setAttribute('data-tabs-direction', to > from ? 'forward' : 'backward');
  }
  panes.forEach(function(pane) { pane.classList.toggle('active', pane.id === targetId); });
  tabs.forEach(function(t) { var selected = t === btn; t.classList.toggle('active', selected); t.setAttribute('aria-selected', selected ? 'true' : 'false'); t.setAttribute('data-state', selected ? 'active' : 'inactive'); t.tabIndex = selected ? 0 : -1; });
  if (btn) btn.focus();
}
function computePresetDates(preset) {
  var now = new Date(), start = new Date(now.getFullYear(), now.getMonth(), now.getDate()), end = new Date(now.getFullYear(), now.getMonth(), now.getDate(), 23, 59, 59, 999);
  if (preset === 'yesterday') { start.setDate(start.getDate() - 1); end = new Date(start.getFullYear(), start.getMonth(), start.getDate(), 23, 59, 59, 999); }
  else if (preset === '7d') start.setDate(start.getDate() - 6); else if (preset === '30d') start.setDate(start.getDate() - 29);
  else if (preset === 'month') start = new Date(now.getFullYear(), now.getMonth(), 1); else if (preset !== 'today') return null;
  return { from: start.toISOString(), to: end.toISOString() };
}
function setPreset(preset) {
  state.currentPreset = preset;
  if (window.renderDashboardPresets) window.renderDashboardPresets();
  else document.querySelectorAll('#presetGroup .preset-btn').forEach(function(btn) { var active = btn.getAttribute('data-preset') === preset; btn.classList.toggle('active', active); btn.setAttribute('aria-checked', active ? 'true' : 'false'); btn.setAttribute('data-state', active ? 'checked' : 'unchecked'); btn.tabIndex = active ? 0 : -1; });
  setFilterValidation('');
  syncPeriodInputs();
  updateActiveFilterBadge();
  if (preset !== 'custom') { state.recordsExpanded = false; loadData(false); }
}
function formatLocalDate(value) {
  var date = new Date(value);
  return date.getFullYear() + '-' + String(date.getMonth() + 1).padStart(2, '0') + '-' + String(date.getDate()).padStart(2, '0');
}
function syncPeriodInputs() {
  var custom = state.currentPreset === 'custom';
  var bounds = custom ? null : computePresetDates(state.currentPreset);
  if (bounds) { state.fromRFC3339 = bounds.from; state.toRFC3339 = bounds.to; }
  var from = DOM.filterFromDate || document.getElementById('filterFromDate'), to = DOM.filterToDate || document.getElementById('filterToDate');
  [from, to].forEach(function(input, index) {
    if (!input) return;
    input.disabled = !custom;
    // Starting custom editing preserves the displayed query, not stale hidden dates.
    if (!custom || !input.value) {
      var value = index === 0 ? state.fromRFC3339 : state.toRFC3339;
      if (value) input.value = formatLocalDate(value);
    }
  });
}
function setFilterValidation(message) {
  var from = DOM.filterFromDate || document.getElementById('filterFromDate'), to = DOM.filterToDate || document.getElementById('filterToDate');
  [from, to].forEach(function(input) { if (input) input.setCustomValidity(message || ''); });
  if (message) showError(message, false); else clearError();
}
var knownProviders = {};
var knownEndpoints = {};
var knownModels = {};

function addKnownOption(knownMap, value) {
  if (value === null || value === undefined) return;
  var normalized = String(value).trim();
  if (normalized && normalized !== '-' && normalized !== 'undefined') knownMap[normalized] = true;
}

function updateSelectTitle(selectEl) {
  if (!selectEl) return;
  var selected = selectEl.selectedOptions && selectEl.selectedOptions[0];
  selectEl.title = selected ? selected.textContent : '';
}

function populateSelectOptions(selectEl, knownMap, allLabel, selectedValue, formatter) {
  if (!selectEl) return;
  var keys = Object.keys(knownMap).sort(function(a, b) { return a.localeCompare(b, 'ko', { sensitivity: 'base' }); });
  selectEl.replaceChildren();
  function appendOption(value, label, transient) {
    var option = document.createElement('option');
    option.value = value;
    option.textContent = label;
    if (transient) option.dataset.transient = 'true';
    selectEl.appendChild(option);
  }
  appendOption('', allLabel || '전체', false);
  if (selectedValue && keys.indexOf(selectedValue) === -1) appendOption(selectedValue, formatter ? formatter(selectedValue) : selectedValue, true);
  keys.forEach(function(key) { appendOption(key, formatter ? formatter(key) : key, false); });
  selectEl.value = selectedValue || '';
  updateSelectTitle(selectEl);
}

function extractFilterOptions(summaryData, recordsData) {
  if (summaryData) {
    [['by_provider', knownProviders], ['by_endpoint', knownEndpoints], ['by_model', knownModels]].forEach(function(pair) {
      var list = summaryData[pair[0]];
      if (Array.isArray(list)) list.forEach(function(item) { addKnownOption(pair[1], item && (item.key || item.name)); });
    });
    if (Array.isArray(summaryData.summary)) summaryData.summary.forEach(function(item) {
      if (!item) return;
      addKnownOption(knownProviders, item.provider);
      addKnownOption(knownModels, item.model);
    });
    [['providers', knownProviders, ['provider','key','name']], ['endpoints', knownEndpoints, ['endpoint','key','name','executor_type']], ['models', knownModels, ['model','key','name']]].forEach(function(spec) {
      var list = summaryData[spec[0]];
      if (!Array.isArray(list)) return;
      list.forEach(function(item) {
        if (!item) return;
        for (var i = 0; i < spec[2].length; i++) if (item[spec[2][i]]) { addKnownOption(spec[1], item[spec[2][i]]); break; }
      });
    });
  }
  var recs = Array.isArray(recordsData) ? recordsData : (recordsData && Array.isArray(recordsData.records) ? recordsData.records : []);
  recs.forEach(function(record) {
    if (!record) return;
    addKnownOption(knownProviders, record.provider || record.Provider);
    addKnownOption(knownEndpoints, record.endpoint || record.executor_type || record.ExecutorType || record.source || record.Source);
    addKnownOption(knownModels, record.model || record.Model || record.alias || record.Alias);
  });
  populateSelectOptions(DOM.filterProvider || document.getElementById('filterProvider'), knownProviders, '전체', state.provider, formatProvider);
  populateSelectOptions(DOM.filterEndpoint || document.getElementById('filterEndpoint'), knownEndpoints, '전체', state.endpoint);
  populateSelectOptions(DOM.filterModel || document.getElementById('filterModel'), knownModels, '전체', state.model);
}

function updateActiveFilterBadge() {
  var badge = DOM.activeFilterBadge || document.getElementById('activeFilterBadge');
  var textEl = DOM.activeFilterText || document.getElementById('activeFilterText');
  var conditions = [];
  if (state.currentPreset !== 'today') conditions.push(state.currentPreset === 'custom' ? '직접 기간' : ({yesterday:'어제', '7d':'최근 7일', '30d':'최근 30일', month:'이번 달'}[state.currentPreset] || state.currentPreset));
  if (state.provider) conditions.push('Provider: ' + state.provider);
  if (state.endpoint) conditions.push('Endpoint: ' + state.endpoint);
  if (state.model) conditions.push('Model: ' + state.model);
  if (state.status) conditions.push(state.status === 'failed' ? '실패만' : '성공만');
  var active = conditions.length > 0;
  var description = conditions.length ? conditions.join(' · ') : '오늘 · 전체 Provider · 전체 Model';
  if (textEl) textEl.textContent = description;
  if (badge) { badge.classList.toggle('is-visible', active); badge.tabIndex = 0; badge.setAttribute('aria-label', '적용 조건: ' + description); }
}

function applyFilters() {
  var from = DOM.filterFromDate || document.getElementById('filterFromDate'), to = DOM.filterToDate || document.getElementById('filterToDate');
  if (state.currentPreset === 'custom') {
    if (!from.value || !to.value) { setFilterValidation('시작일과 종료일을 모두 선택하세요.'); (from.value ? to : from).reportValidity(); return; }
    if (from.value > to.value) { setFilterValidation('시작일은 종료일보다 늦을 수 없습니다.'); from.reportValidity(); return; }
    setFilterValidation(''); var sParts = from.value.split('-'), eParts = to.value.split('-');
    state.fromRFC3339 = new Date(Number(sParts[0]), Number(sParts[1])-1, Number(sParts[2])).toISOString();
    state.toRFC3339 = new Date(Number(eParts[0]), Number(eParts[1])-1, Number(eParts[2]), 23, 59, 59, 999).toISOString();
  }
  state.provider = (DOM.filterProvider || document.getElementById('filterProvider')).value; state.endpoint = (DOM.filterEndpoint || document.getElementById('filterEndpoint')).value; state.model = (DOM.filterModel || document.getElementById('filterModel')).value; state.status = (DOM.filterStatus || document.getElementById('filterStatus')).value;
  state.recordsExpanded = false; updateActiveFilterBadge(); loadData(false);
}
function resetFilters() {
  [DOM.filterProvider || document.getElementById('filterProvider'), DOM.filterEndpoint || document.getElementById('filterEndpoint'), DOM.filterModel || document.getElementById('filterModel'), DOM.filterStatus || document.getElementById('filterStatus'), DOM.filterFromDate || document.getElementById('filterFromDate'), DOM.filterToDate || document.getElementById('filterToDate')].forEach(function(input) { input.value = ''; if (input.tagName === 'SELECT') updateSelectTitle(input); });
  state.provider = ''; state.endpoint = ''; state.model = ''; state.status = ''; setFilterValidation(''); updateActiveFilterBadge(); setPreset('today');
}
function buildQueryString(isRecords, cursor) {
  var params = [];
  if (state.fromRFC3339) params.push('from=' + encodeURIComponent(state.fromRFC3339));
  if (state.toRFC3339) params.push('to=' + encodeURIComponent(state.toRFC3339));
  if (state.provider) params.push('provider=' + encodeURIComponent(state.provider));
  if (state.model) params.push('model=' + encodeURIComponent(state.model));
  if (state.status) params.push('status=' + encodeURIComponent(state.status));
  if (state.endpoint) params.push('endpoint=' + encodeURIComponent(state.endpoint));
  if (!isRecords) {
    params.push('tz_offset=' + encodeURIComponent(String(new Date().getTimezoneOffset())));
    try {
      var timeZone = Intl.DateTimeFormat().resolvedOptions().timeZone;
      if (timeZone) params.push('time_zone=' + encodeURIComponent(timeZone));
    } catch (_) {}
  }
  if (isRecords) {
    params.push('limit=20');
    if (cursor) {
      params.push('cursor=' + encodeURIComponent(cursor));
    }
  }
  return params.length > 0 ? ('?' + params.join('&')) : '';
}

function showError(msg, retryable) { var box = DOM.errorBox || document.getElementById('errorBox'), text = DOM.errorMessage || document.getElementById('errorMessage'), retry = DOM.retryLoadBtn || document.getElementById('retryLoadBtn'); if (text) text.textContent = msg || '데이터 사용 불가'; if (retry) retry.classList.toggle('is-hidden', retryable === false); if (box) box.classList.remove('is-hidden'); }
function clearError() { var box = DOM.errorBox || document.getElementById('errorBox'), text = DOM.errorMessage || document.getElementById('errorMessage'); if (box) box.classList.add('is-hidden'); if (text) text.textContent = ''; }
function safeErrorMessage(response) { return response.text().then(function(text) { var fallback = '요청이 실패했습니다 (HTTP ' + response.status + ').'; if (!text) return fallback; try { var parsed = JSON.parse(text); text = parsed.message || parsed.error || fallback; } catch (_) {} return String(text).replace(/\s+/g, ' ').slice(0, 300) || fallback; }); }
// Snapshot comparison is only performed for a repeated server query. Filter-driven
// queries reset the baseline, so a UI rerender never looks like a data change.
function snapshotFilterKey() { return [state.currentPreset,state.fromRFC3339,state.toRFC3339,state.provider,state.endpoint,state.model,state.status].join('|'); }
function makeUsageSnapshot(summary, records) {
  var snap={ values:{}, rows:{} };
  if (summary) {
    var total=typeof summary.total_count==='number'?summary.total_count:0, failures=0, tokens=0, input=0, fresh=0, output=0, reasoning=0, cached=0;
    (Array.isArray(summary.summary)?summary.summary:[]).forEach(function(item){item=item||{}; failures+=Number(item.failures)||0;tokens+=Number(item.total_tokens)||0;input+=Number(item.input_tokens)||0;fresh+=Number(item.fresh_input_tokens)||0;output+=Number(item.output_tokens)||0;reasoning+=Number(item.reasoning_tokens)||0;cached+=Number(item.cached_tokens)||0; if(!total)total+=Number(item.requests)||0;});
    snap.values={ requests:total, failures:failures, tokens:tokens, input:input, fresh:fresh, output:output, reasoning:reasoning, cached:cached };
    [['provider',summary.by_provider],['endpoint',summary.by_endpoint],['model',summary.by_model]].forEach(function(pair){(Array.isArray(pair[1])?pair[1]:[]).forEach(function(row){row=row||{};var rawKey=String(row.key||row[pair[0]]||row.name||row.endpoint||row.source||row.executor_type||''); if(!rawKey) return; var key=pair[0]+':'+rawKey;snap.rows[key]={requests:Number(row.requests||row.count)||0,failures:Number(row.failures)||0,tokens:Number(row.total_tokens)||0};});});
  }
  (Array.isArray(records)?records:(records&&records.records)||[]).forEach(function(row,index){row=row||{};var key=String(row.id||row.request_id||row.requested_at||row.t||row.Timestamp||index);snap.rows['record:'+key]=!!(row.failed||row.Failed);}); return snap;
}
function usageTone(before, after, badWhenUp) { if (before === after) return ''; if (typeof before==='number' && typeof after==='number') return ((after>before) !== !!badWhenUp)?'flash-good':'flash-bad'; return 'flash-neutral'; }
function diffUsageSnapshots(previous, next) { if(!previous)return {values:{},rows:{},records:{}}; var changes={values:{},rows:{},records:{}};Object.keys(next.values).forEach(function(key){if(previous.values[key]!==undefined&&previous.values[key]!==next.values[key])changes.values[key]=usageTone(previous.values[key],next.values[key],key==='failures');});Object.keys(next.rows).forEach(function(key){var before=previous.rows[key], after=next.rows[key];if(before===undefined||after===undefined)return;if(key.indexOf('record:')===0){if(before!==after)changes.records[key]='flash-neutral';return;}['requests','failures','tokens'].forEach(function(field){if(before[field]!==after[field])changes.rows[key+':'+field]=usageTone(before[field],after[field],field==='failures');});});return changes; }
function applyPanelPulse(el) { if(!el||!el.classList)return;if(el._panelPulseTimer)clearTimeout(el._panelPulseTimer);el.classList.remove('panel-pulse');void el.offsetWidth;el.classList.add('panel-pulse');el._panelPulseTimer=setTimeout(function(){el.classList.remove('panel-pulse');el._panelPulseTimer=null;},1000); }
function pulseKPIGroup(el, tone) { if(tone&&el&&el.closest)applyPanelPulse(el.closest('.kpi-group')); }
function flashValue(el, tone, label) { if(!el||!tone)return; var tones=['flash-good','flash-bad','flash-danger','flash-neutral']; if(el._flashTimer)clearTimeout(el._flashTimer); tones.forEach(function(c){el.classList.remove(c);});el.classList.remove('value-flash'); var originalAria=el.getAttribute('aria-label'), originalTitle=el.getAttribute('title'); el.classList.add('value-flash',tone);el.title=label+' · 서버 업데이트로 변경됨';el.setAttribute('aria-label',(originalAria||el.textContent||label)+' · 서버 업데이트로 변경됨');el._flashTimer=setTimeout(function(){el.classList.remove('value-flash');tones.forEach(function(c){el.classList.remove(c);});if(originalTitle===null)el.removeAttribute('title');else el.setAttribute('title',originalTitle);if(originalAria===null)el.removeAttribute('aria-label');else el.setAttribute('aria-label',originalAria);el._flashTimer=null;},1000); }
// computeAdaptivePollMs: determines the next poll delay based on the current
// preset and the freshness of the most recent record.
// - today with record within 2m => 30s
// - today with activity but older => 90s
// - quiet today (no records) => 5m
// - historical presets (yesterday/7d/30d/month/custom) => 10m
function computeAdaptivePollMs(records, preset) {
  var now = Date.now();
  if (state.toRFC3339 && new Date(state.toRFC3339).getTime() < now - 60000) return 0;
  if (preset !== 'today') return UA_POLL_HISTORY_MS;
  if (!records || records.length === 0) return UA_POLL_QUIET_MS;
  var latestMs = 0;
  records.forEach(function(rec) { var rawT = rec.requested_at || rec.t || rec.RequestedAt; var t = rawT ? new Date(rawT).getTime() : 0; if (!isNaN(t) && t > latestMs) latestMs = t; });
  if (!latestMs) return UA_POLL_QUIET_MS;
  return Date.now() - latestMs <= 2 * 60 * 1000 ? UA_POLL_RECENT_MS : UA_POLL_ACTIVE_MS;
}
function scheduleNextPoll(delayMs) {
  if (state.pollTimer) { clearTimeout(state.pollTimer); state.pollTimer = null; }
  if (delayMs <= 0) return; // completed historical range: do not poll
  if (delayMs > UA_POLL_MAX_MS) delayMs = UA_POLL_MAX_MS;
  state.pollTimer = setTimeout(function() {
    state.pollTimer = null;
    if (document.hidden) return; // paused; will reload on visibilitychange
    if (!state.inFlight) loadData(false);
  }, delayMs);
}

function setDashboardBusy(busy) {
  var content = DOM.dashboardContent || document.getElementById('dashboardContent');
  if (content) content.setAttribute('aria-busy', busy ? 'true' : 'false');
}

function renderSkeletons() {
  var skeleton = DOM.usageSkeleton || document.getElementById('usageSkeleton');
  var dataContent = DOM.usageDataContent || document.getElementById('usageDataContent');
  if (skeleton) {
    skeleton.classList.add('is-visible');
    skeleton.setAttribute('aria-hidden', 'false');
  }
  if (dataContent) dataContent.classList.add('is-loading');
  scheduleDailyChartRedraw();
  setDashboardBusy(true);
}

function hideSkeletons() {
  var skeleton = DOM.usageSkeleton || document.getElementById('usageSkeleton');
  var dataContent = DOM.usageDataContent || document.getElementById('usageDataContent');
  if (skeleton) {
    skeleton.classList.remove('is-visible');
    skeleton.setAttribute('aria-hidden', 'true');
  }
  if (dataContent) dataContent.classList.remove('is-loading');
}

function fetchWithTimeout(url, signal) {
  var timedOut = false, timeoutCtrl = new AbortController(), timeout = setTimeout(function() { timedOut = true; timeoutCtrl.abort(); }, UA_FETCH_TIMEOUT_MS), relayAbort = function() { timeoutCtrl.abort(); };
  signal.addEventListener('abort', relayAbort, { once: true });
  return fetch(url, { signal: timeoutCtrl.signal }).then(function(response) { if (!response.ok) return safeErrorMessage(response).then(function(message) { throw new Error(message); }); return response.json(); }).catch(function(err) { if (timedOut) throw new Error('요청 시간이 초과되었습니다. 다시 시도하세요.'); throw err; }).finally(function() { clearTimeout(timeout); signal.removeEventListener('abort', relayAbort); });
}
function loadData(appendRecords, forceReplace) {
  if (!appendRecords && state.recordsExpanded && !forceReplace && state.hasLoadedData && state.snapshotFilterKey === snapshotFilterKey()) { showError('더보기 기록을 보고 있습니다. 최신 데이터를 보려면 필터를 변경하거나 다시 시도하세요.'); return; }
  if (state.inFlight && state.abortCtrl) state.abortCtrl.abort();
  var requestID = ++state.requestID; state.abortCtrl = new AbortController(); state.inFlight = true;
  if (!appendRecords && !state.hasLoadedData) renderSkeletons(); else setDashboardBusy(true);
  var signal = state.abortCtrl.signal, filterKey = snapshotFilterKey();
  var previous = state.snapshotFilterKey === filterKey ? state.serverSnapshot : null;
  var progressive = !state.hasLoadedData && !appendRecords;
  var summaryUrl = '/v0/resource/plugins/usage-audit/data/summary' + buildQueryString(false);
  var recordsUrl = '/v0/resource/plugins/usage-audit/data/records' + buildQueryString(true, appendRecords ? state.cursor : null);
  function currentRequest() { return requestID === state.requestID && !signal.aborted; }
  // Independent panels paint as soon as their own response arrives: a slow records
  // request must not hold the token KPIs/chart behind Promise.all (or vice versa).
  var summaryReq = appendRecords ? Promise.resolve(null) : fetchWithTimeout(summaryUrl, signal).then(function(summary) {
    if (!currentRequest() || !progressive) return summary;
    state.pendingChanges=diffUsageSnapshots(previous, makeUsageSnapshot(summary, null));
    renderSummary(summary); state.pendingChanges=null; hideSkeletons();
    return summary;
  });
  var recordsReq = fetchWithTimeout(recordsUrl, signal).then(function(records) {
    if (!currentRequest() || !progressive) return records;
    state.pendingChanges=appendRecords ? {values:{},rows:{},records:{}} : diffUsageSnapshots(previous, makeUsageSnapshot(null, records));
    renderRecords(records, appendRecords); state.pendingChanges=null; hideSkeletons();
    return records;
  });
  Promise.all([summaryReq, recordsReq]).then(function(data) {
    if (!currentRequest()) return;
    state.inFlight = false; clearError();
    if (!progressive) {
      state.pendingChanges=appendRecords ? {values:{},rows:{},records:{}} : diffUsageSnapshots(previous, makeUsageSnapshot(data[0], data[1]));
      if (data[0]) renderSummary(data[0]); renderRecords(data[1], appendRecords); state.pendingChanges=null;
    }
    if (!appendRecords) { state.serverSnapshot=makeUsageSnapshot(data[0], data[1]); state.snapshotFilterKey=filterKey; }
    extractFilterOptions(data[0], data[1]); state.hasLoadedData = true; hideSkeletons(); setDashboardBusy(false);
    if (!appendRecords) scheduleNextPoll(computeAdaptivePollMs(state.records, state.currentPreset));
  }).catch(function(err) {
    if (requestID !== state.requestID) return;
    // Prevent the surviving sibling response from painting after this request failed.
    state.abortCtrl.abort(); state.pendingChanges=null; state.inFlight = false; setDashboardBusy(false);
    if (err.name !== 'AbortError') { hideSkeletons(); showError(err.message === 'The user aborted a request.' ? '요청 시간이 초과되었습니다. 다시 시도하세요.' : '데이터를 불러오지 못했습니다: ' + err.message); scheduleNextPoll(UA_POLL_RETRY_MS); }
  });
}
// Render Summary and Aggregations
// NOTE: total_tokens from backend is authoritative.
// cached_tokens may be included in input_tokens or reported separately by a provider.
// reasoning_tokens is provider-reported and may already be in total_tokens.
// NEVER calculate total_tokens from the displayed component fields.
function renderSummary(data) {
  if (!data) return;

  var totalCount = (typeof data.total_count === 'number') ? data.total_count : 0;
  var summaryList = Array.isArray(data.summary) ? data.summary : [];

  // Aggregate once across all summary buckets (avoids repeated reduce)
  var generateCount = 0;
  var inputTokens = 0;
  var freshInputTokens = 0;
  var outputTokens = 0;
  // KPI token values are derived from the summary buckets. The backend does
  // not expose additive top-level token fields; total_tokens is authoritative
  // per bucket and is never reconstructed from input/output/cache/reasoning.
  var totalTokens = 0;
  var reasoningTokens = 0;
  var cachedTokens = 0;
  var failures = 0;

  var providerMap = {};
  var modelMap = {};
  var endpointMap = {};
  var dailyMap = {};

  var hasDaily = Array.isArray(data.daily) && data.daily.length > 0;
  var hasByProvider = Array.isArray(data.by_provider) && data.by_provider.length > 0;
  var hasByEndpoint = Array.isArray(data.by_endpoint) && data.by_endpoint.length > 0;
  var hasByModel = Array.isArray(data.by_model) && data.by_model.length > 0;

  if (hasDaily) {
    data.daily.forEach(function(item) {
      var k = item.key || item.date || item.day || item.time || item.t;
      if (!k) return;
      var dateKey = String(k).slice(0, 10);
      dailyMap[dateKey] = {
        date: dateKey,
        requests: Number(item.requests || item.count || 0),
        total: (typeof item.total_tokens === 'number' ? item.total_tokens : 0),
        input: (typeof item.input_tokens === 'number' ? item.input_tokens : 0),
        output: (typeof item.output_tokens === 'number' ? item.output_tokens : 0),
        failures: (typeof item.failures === 'number' ? item.failures : 0),
        cached: (typeof item.cached_tokens === 'number' ? item.cached_tokens : 0),
        reasoning: (typeof item.reasoning_tokens === 'number' ? item.reasoning_tokens : 0)
      };
    });
  }

  summaryList.forEach(function(item) {
    if (!item) return;
    if (typeof item.generate_count === 'number') generateCount += item.generate_count;
    if (typeof item.requests === 'number' && !totalCount) totalCount += item.requests;
    if (typeof item.failures === 'number') failures += item.failures;
    var inp = item.input_tokens || 0;
    var outp = item.output_tokens || 0;
    inputTokens += inp;
    freshInputTokens += (typeof item.fresh_input_tokens === 'number' ? item.fresh_input_tokens : 0);
    outputTokens += outp;
    // Authoritative total token item fallback
    // Missing stored totals are treated as zero; never infer a total.
    totalTokens += (typeof item.total_tokens === 'number' ? item.total_tokens : 0);
    reasoningTokens += (typeof item.reasoning_tokens === 'number' ? item.reasoning_tokens : 0);
    cachedTokens += (typeof item.cached_tokens === 'number' ? item.cached_tokens : 0);

    var p = item.provider || 'unknown';
    providerMap[p] = (providerMap[p] || { requests: 0, failures: 0, tokens: 0 });
    providerMap[p].requests += (item.requests || 0);
    providerMap[p].failures += (item.failures || 0);
    // total_tokens is authoritative; absent totals are not inferred from input/output.
    providerMap[p].tokens += (typeof item.total_tokens === 'number' ? item.total_tokens : 0);

    var m = item.model || 'unknown';
    modelMap[m] = (modelMap[m] || { requests: 0, failures: 0, tokens: 0 });
    modelMap[m].requests += (item.requests || 0);
    modelMap[m].failures += (item.failures || 0);
    // total_tokens is authoritative; absent totals are not inferred from input/output.
    modelMap[m].tokens += (typeof item.total_tokens === 'number' ? item.total_tokens : 0);

    var timeVal = item.time || item.date || item.day || item.t || item.timestamp;
    if (!hasDaily && timeVal) {
      var dk = String(timeVal).slice(0, 10);
      var daily = dailyMap[dk] || { date: dk, requests: 0, total: 0, input: 0, output: 0, failures: 0, cached: 0, reasoning: 0 };
      daily.requests += (item.requests || 1);
      daily.total += (typeof item.total_tokens === 'number' ? item.total_tokens : 0);
      daily.input += inp;
      daily.output += outp;
      daily.failures += (item.failures || 0);
      daily.cached += (typeof item.cached_tokens === 'number' ? item.cached_tokens : 0);
      daily.reasoning += (typeof item.reasoning_tokens === 'number' ? item.reasoning_tokens : 0);
      dailyMap[dk] = daily;
    }
  });

  // Use cached DOM refs for all KPI updates
  var elTotalCount = DOM.kpiTotalCount;
  elTotalCount.textContent = formatExactNumber(totalCount);
  elTotalCount.setAttribute('title', formatExactNumber(totalCount));
  elTotalCount.setAttribute('aria-label', '총 요청: ' + formatExactNumber(totalCount));
  flashValue(elTotalCount, state.pendingChanges && state.pendingChanges.values.requests, '총 요청'); pulseKPIGroup(elTotalCount, state.pendingChanges && state.pendingChanges.values.requests);

  var failRate = (totalCount > 0) ? ((failures / totalCount) * 100).toFixed(1) + '%' : '0.0%';
  DOM.kpiSubFail.textContent = '실패: ' + formatExactNumber(failures) + '건 (' + failRate + ')';
  flashValue(DOM.kpiSubFail, state.pendingChanges && state.pendingChanges.values.failures, '실패 요청'); pulseKPIGroup(DOM.kpiSubFail, state.pendingChanges && state.pendingChanges.values.failures);
  DOM.kpiFailureRate.textContent = failRate;
  DOM.kpiSubReq.textContent = '생성 요청: ' + formatExactNumber(generateCount) + '건';

  var elTotalTokens = DOM.kpiTotalTokens;
  elTotalTokens.textContent = formatExactNumber(totalTokens);
  elTotalTokens.setAttribute('title', formatExactNumber(totalTokens));
  elTotalTokens.setAttribute('aria-label', '전체 토큰: ' + formatExactNumber(totalTokens));
  flashValue(elTotalTokens, state.pendingChanges && state.pendingChanges.values.tokens, '전체 토큰'); pulseKPIGroup(elTotalTokens, state.pendingChanges && state.pendingChanges.values.tokens);
  if (DOM.kpiSubTokens) DOM.kpiSubTokens.textContent = '공식 집계';

  var elInputReported = DOM.kpiInputTokens;
  if (elInputReported) {
    elInputReported.textContent = formatExactNumber(inputTokens);
    elInputReported.setAttribute('title', formatExactNumber(inputTokens));
    elInputReported.setAttribute('aria-label', '입력 보고값: ' + formatExactNumber(inputTokens));
    flashValue(elInputReported, state.pendingChanges && state.pendingChanges.values.input, '입력 보고값'); pulseKPIGroup(elInputReported, state.pendingChanges && state.pendingChanges.values.input);
  }

  var elInput = DOM.kpiFreshInputTokens;
  if (elInput) {
    elInput.textContent = formatExactNumber(freshInputTokens);
    elInput.setAttribute('title', formatExactNumber(freshInputTokens));
    elInput.setAttribute('aria-label', '신규 입력: ' + formatExactNumber(freshInputTokens));
    flashValue(elInput, state.pendingChanges && state.pendingChanges.values.fresh, '신규 입력'); pulseKPIGroup(elInput, state.pendingChanges && state.pendingChanges.values.fresh);
  }

  var elOutput = DOM.kpiOutputTokens;
  if (elOutput) {
    elOutput.textContent = formatExactNumber(outputTokens);
    elOutput.setAttribute('title', formatExactNumber(outputTokens));
    elOutput.setAttribute('aria-label', '출력 토큰: ' + formatExactNumber(outputTokens));
    flashValue(elOutput, state.pendingChanges && state.pendingChanges.values.output, '출력 토큰'); pulseKPIGroup(elOutput, state.pendingChanges && state.pendingChanges.values.output);
  }

  var elReason = DOM.kpiReasoningTokens;
  if (elReason) {
    elReason.textContent = formatExactNumber(reasoningTokens);
    elReason.setAttribute('title', formatExactNumber(reasoningTokens));
    elReason.setAttribute('aria-label', '추론 토큰: ' + formatExactNumber(reasoningTokens));
    flashValue(elReason, state.pendingChanges && state.pendingChanges.values.reasoning, '추론 토큰'); pulseKPIGroup(elReason, state.pendingChanges && state.pendingChanges.values.reasoning);
  }

  var elCache = DOM.kpiCachedTokens;
  if (elCache) {
    elCache.textContent = formatExactNumber(cachedTokens);
    elCache.setAttribute('title', formatExactNumber(cachedTokens));
    elCache.setAttribute('aria-label', '캐시 재사용 토큰: ' + formatExactNumber(cachedTokens));
    flashValue(elCache, state.pendingChanges && state.pendingChanges.values.cached, '캐시 재사용 토큰'); pulseKPIGroup(elCache, state.pendingChanges && state.pendingChanges.values.cached);
  }

  // Entities compat
  var pCount = hasByProvider ? data.by_provider.length : Object.keys(providerMap).length;
  var mCount = hasByModel ? data.by_model.length : Object.keys(modelMap).length;
  if (DOM.kpiActiveEntities) DOM.kpiActiveEntities.textContent = pCount + ' / ' + mCount;
  if (DOM.kpiSubEntities) DOM.kpiSubEntities.textContent = 'Provider ' + pCount + '개 | Model ' + mCount + '개';

  // Render Daily Trend Chart
  renderChartFromDaily(dailyMap);

  // Render Breakdown Section
  renderBreakdowns({
    hasByProvider: hasByProvider,
    providers: hasByProvider ? data.by_provider : providerMap,
    hasByEndpoint: hasByEndpoint,
    endpoints: hasByEndpoint ? data.by_endpoint : endpointMap,
    hasByModel: hasByModel,
    models: hasByModel ? data.by_model : modelMap
  });
}

function appendText(parent, className, text) { var el = document.createElement('div'); if (className) el.className = className; el.textContent = text; parent.appendChild(el); return el; }
function closeDailyChartTooltip() { var tooltip = DOM.dailyChartTooltip || document.getElementById('dailyChartTooltip'); if (tooltip) tooltip.hidden = true; }
function showDailyChartTooltip(detail, anchor) {
  var tooltip = DOM.dailyChartTooltip || document.getElementById('dailyChartTooltip'); if (!tooltip || !anchor) return;
  tooltip.replaceChildren();
  appendText(tooltip, 'daily-chart-tooltip-title', detail.date);
  appendText(tooltip, '', '요청: ' + formatExactNumber(detail.requests)); appendText(tooltip, '', '전체 토큰 (공식): ' + formatExactNumber(detail.total));
  appendText(tooltip, '', '입력 (보고값): ' + formatExactNumber(detail.input)); appendText(tooltip, '', '출력: ' + formatExactNumber(detail.output)); appendText(tooltip, '', '실패: ' + formatExactNumber(detail.failures));
  appendText(tooltip, 'daily-chart-tooltip-note', '캐시 (보고값, 비가산): ' + formatExactNumber(detail.cached) + ' · 추론 (보고값, 비가산): ' + formatExactNumber(detail.reasoning));
  tooltip.hidden = false; var box = anchor.getBoundingClientRect(), margin = 8;
  var left = Math.max(margin, Math.min(box.left, window.innerWidth - tooltip.offsetWidth - margin)), top = box.top - tooltip.offsetHeight - margin;
  if (top < margin) top = Math.min(window.innerHeight - tooltip.offsetHeight - margin, box.bottom + margin);
  tooltip.style.left = Math.max(margin, left) + 'px'; tooltip.style.top = Math.max(margin, top) + 'px';
}
function dateKeysForSelectedRange(dailyMap) {
  var keys = Object.keys(dailyMap), from = state.fromRFC3339 ? new Date(state.fromRFC3339) : null, to = state.toRFC3339 ? new Date(state.toRFC3339) : null;
  if (!from || !to || isNaN(from) || isNaN(to)) return keys.sort();
  // RFC3339 bounds represent local-day selections. Use local calendar fields so
  // positive/negative timezone offsets do not add an adjacent empty UTC day.
  var current = new Date(from.getFullYear(), from.getMonth(), from.getDate()), last = new Date(to.getFullYear(), to.getMonth(), to.getDate()), result = [];
  while (current <= last) {
    var key = current.getFullYear() + '-' + String(current.getMonth() + 1).padStart(2, '0') + '-' + String(current.getDate()).padStart(2, '0');
    result.push(key);
    if (!dailyMap[key]) dailyMap[key] = { date:key, requests:0, total:0, input:0, output:0, failures:0, cached:0, reasoning:0 };
    current.setDate(current.getDate() + 1);
  }
  return result;
}
var lastDailyChartData = null;
var dailyChartRedrawFrame = null;
function getDailyChartGeometry(days, width) {
  var count = Math.max(1, Math.floor(Number(days) || 1));
  var availableWidth = Number(width);
  if (!Number.isFinite(availableWidth) || availableWidth <= 0) availableWidth = 280;
  var svgWidth = count <= 7 ? availableWidth : Math.max(availableWidth, count * 40 + 36);
  var inset = Math.min(18, svgWidth / 4), slot = (svgWidth - inset * 2) / count;
  var barWidth = Math.min(36, Math.max(16, slot * 0.55));
  var labelStride = Math.max(1, Math.ceil(count / (svgWidth / 60)));
  var centers = [], labelIndices = [];
  for (var k = 0; k < count; k++) {
    centers.push(inset + (k + 0.5) * slot);
    // Keep endpoint labels, but omit the last intermediate label if it would collide.
    if (k === 0 || k === count - 1 || (k % labelStride === 0 && (count - 1 - k) * slot >= 60)) labelIndices.push(k);
  }
  return { width: svgWidth, slot: slot, barWidth: barWidth, labelStride: labelStride, centers: centers, labelIndices: labelIndices };
}
function scheduleDailyChartRedraw() {
  if (dailyChartRedrawFrame !== null) cancelAnimationFrame(dailyChartRedrawFrame);
  dailyChartRedrawFrame = requestAnimationFrame(function() {
    dailyChartRedrawFrame = null;
    var card = DOM.chartCard || document.getElementById('chartCard');
    if (card && card.open && lastDailyChartData) renderChartFromDaily(lastDailyChartData);
  });
}
function renderChartFromDaily(dailyMap) {
  var chartCard = DOM.chartCard || document.getElementById('chartCard'), chartArea = DOM.chartArea || document.getElementById('chartArea'), tooltip = DOM.dailyChartTooltip || document.getElementById('dailyChartTooltip');
  lastDailyChartData = dailyMap;
  if (dailyChartRedrawFrame !== null) { cancelAnimationFrame(dailyChartRedrawFrame); dailyChartRedrawFrame = null; }
  var focusedDay = chartArea.contains(document.activeElement) ? document.activeElement.getAttribute('data-day') : null, focusedBar = null;
  chartArea.replaceChildren(); closeDailyChartTooltip(); var days = dateKeysForSelectedRange(dailyMap);
  var summary = DOM.dailyChartSummary || document.getElementById('dailyChartSummary');
  var maxVal = 0, totalRequests = 0;
  days.forEach(function(day) { var requests = Number(dailyMap[day].requests) || 0; maxVal = Math.max(maxVal, requests); totalRequests += requests; });
  if (summary) summary.textContent = totalRequests ? days.length + '일 · 총 ' + formatExactNumber(totalRequests) + '건 · 최고 ' + formatExactNumber(maxVal) + '건' : '선택 기간 요청 없음';
  chartCard.classList.remove('is-hidden');
  if (!maxVal) { appendText(chartArea, 'chart-empty-state', '선택 기간 요청 없음'); return; }
  var geometry = getDailyChartGeometry(days.length, chartCard.clientWidth > 24 ? chartCard.clientWidth - 24 : 280);
  var svgWidth = geometry.width, svgHeight = 135, barWidth = geometry.barWidth, chartBottom = 110, chartTop = 15;
  var svg = document.createElementNS('http://www.w3.org/2000/svg', 'svg'); svg.setAttribute('width', String(svgWidth)); svg.setAttribute('height', String(svgHeight)); svg.setAttribute('viewBox', '0 0 ' + svgWidth + ' ' + svgHeight); svg.setAttribute('role', 'img'); svg.setAttribute('aria-label', '날짜별 요청 추이 막대 차트'); svg.setAttribute('aria-describedby', 'dailyChartTooltip');
  var line = document.createElementNS('http://www.w3.org/2000/svg', 'line'); line.setAttribute('x1', '10'); line.setAttribute('x2', String(svgWidth - 10)); line.setAttribute('y1', String(chartBottom)); line.setAttribute('y2', String(chartBottom)); line.setAttribute('stroke', '#e5e7eb'); line.setAttribute('stroke-width', '1'); svg.appendChild(line);
  days.forEach(function(day, k) {
    var detail = dailyMap[day], barH = Math.max(6, Math.round(((Number(detail.requests) || 0) / maxVal) * (chartBottom - chartTop)));
    var x = geometry.centers[k] - barWidth / 2, y = chartBottom - barH, rect = document.createElementNS('http://www.w3.org/2000/svg', 'rect');
    rect.setAttribute('x',String(x));rect.setAttribute('y',String(y));rect.setAttribute('width',String(barWidth));rect.setAttribute('height',String(barH));rect.setAttribute('fill',detail.requests ? 'var(--accent-blue, #3b6593)' : '#cbd5e1');rect.setAttribute('rx','3');rect.setAttribute('class','daily-chart-bar');rect.setAttribute('tabindex','0');rect.setAttribute('role','button');rect.setAttribute('data-day',day);rect.setAttribute('aria-describedby','dailyChartTooltip');rect.setAttribute('aria-label',detail.date+', 요청 '+formatExactNumber(detail.requests)+', 상세 보기');
    if (day === focusedDay) focusedBar = rect;
    rect.addEventListener('pointerenter',function(){showDailyChartTooltip(detail,rect);});rect.addEventListener('pointerleave',closeDailyChartTooltip);rect.addEventListener('focus',function(){showDailyChartTooltip(detail,rect);});rect.addEventListener('click',function(e){e.stopPropagation();showDailyChartTooltip(detail,rect);});rect.addEventListener('blur',closeDailyChartTooltip);rect.addEventListener('keydown',function(e){if(e.key === 'Enter'||e.key === ' '){e.preventDefault();showDailyChartTooltip(detail,rect);}if(e.key === 'Escape'){e.preventDefault();closeDailyChartTooltip();rect.blur();}}); svg.appendChild(rect);
    if (geometry.slot >= 64) {
      var value=document.createElementNS('http://www.w3.org/2000/svg','text');value.setAttribute('x',String(geometry.centers[k]));value.setAttribute('y',String(y-4));value.setAttribute('text-anchor','middle');value.setAttribute('font-size','10');value.setAttribute('font-variant-numeric','tabular-nums');value.setAttribute('fill','#6b7280');value.textContent=formatExactNumber(detail.requests);svg.appendChild(value);
    }
    if (geometry.labelIndices.indexOf(k) !== -1) {
      var label=document.createElementNS('http://www.w3.org/2000/svg','text');label.setAttribute('x',String(geometry.centers[k]));label.setAttribute('y',String(chartBottom+14));label.setAttribute('text-anchor','middle');label.setAttribute('font-size','10');label.setAttribute('fill','#6b7280');label.textContent=detail.date.length>=5?detail.date.slice(5):detail.date;svg.appendChild(label);
    }
  });
  chartArea.appendChild(svg);
  if (focusedBar) focusedBar.focus({ preventScroll: true });
  closeDailyChartTooltip(); if (tooltip) tooltip.hidden = true;
}
function renderBreakdowns(data) {
  var section = DOM.breakdownSection || document.getElementById('breakdownSection'); section.classList.remove('is-hidden');
  function buildRows(source, isArray, nameKey, category) {
    var list = []; if (isArray) list = source.map(function(item) { item=item||{}; var rawName = String(item[nameKey] || item.key || item.name || item.endpoint || item.source || item.executor_type || '-'); return { rawKey:category+':'+rawName, name:category==='provider' ? formatProvider(rawName) : rawName, requests:item.requests || item.count || 0, failures:item.failures || 0, tokens:typeof item.total_tokens === 'number' ? item.total_tokens : 0 }; });
    else Object.keys(source).forEach(function(key) { list.push({ rawKey:category+':'+String(key), name:category==='provider' ? formatProvider(key) : key, requests:source[key].requests, failures:source[key].failures, tokens:source[key].tokens }); });
    return list.sort(function(a,b) { return b.requests-a.requests; });
  }
  function toTableRows(rows) {
    return rows.slice(0,10).map(function(r) {
      var changes=state.pendingChanges&&state.pendingChanges.rows;
      function cell(text,className,field,label) { var key=r.rawKey+':'+field;return {text:text,className:className,title:field?null:String(r.name),changeKey:field?key:null,tone:field&&changes&&changes[key]||null,label:label}; }
      return {key:r.rawKey,cells:[cell(String(r.name),'cell-ellipsis',null),cell(formatExactNumber(r.requests),'text-right','requests','요청'),cell(formatExactNumber(r.failures),'text-right','failures','실패'),cell(formatExactNumber(r.tokens),'text-right','tokens','토큰')]};
    });
  }
  var tableUpdates=[['providerAggTableMount',toTableRows(buildRows(data.providers,data.hasByProvider,'provider','provider')),'데이터 없음'],['endpointAggTableMount',toTableRows(buildRows(data.endpoints,data.hasByEndpoint,'endpoint','endpoint')),'데이터 없음'],['modelAggTableMount',toTableRows(buildRows(data.models,data.hasByModel,'model','model')),'데이터 없음']];
  if(window.renderDohwiTable)tableUpdates.forEach(function(update){window.renderDohwiTable(update[0],update[1],update[2]);});else window.pendingDohwiTableUpdates=(window.pendingDohwiTableUpdates||[]).concat(tableUpdates);
  if(state.pendingChanges&&Object.keys(state.pendingChanges.rows||{}).length)applyPanelPulse(section);
}
// Render Records & Pagination (20 records limit, dohwi-ui Table primitives)
// legacy safe text contract: function appendRecordCell; cell.textContent=text; cell.title=title
function renderRecords(data, append) {
  var countBadge=DOM.recordsCountBadge || document.getElementById('recordsCountBadge'), loadMoreBar=DOM.loadMoreBar || document.getElementById('loadMoreBar'), recordsList=[];
  if(Array.isArray(data)){recordsList=data;state.hasMore=false;state.cursor=null;}else if(data&&Array.isArray(data.records)){recordsList=data.records;state.hasMore=!!data.has_more;state.cursor=data.next_cursor||null;}
  if(append){state.records=state.records.concat(recordsList);state.recordsExpanded=true;}else{state.records=recordsList;state.recordsExpanded=false;}
  countBadge.textContent=state.records.length+'건';
  var rows=state.records.map(function(rec,idx) {
    var rawT=rec.requested_at||rec.t||rec.RequestedAt||rec.Timestamp, rawProv=rec.provider||rec.Provider||'-', displayProv=formatProvider(rawProv), model=rec.model||rec.Model||rec.alias||rec.Alias||'-', endpoint=rec.endpoint||rec.executor_type||rec.ExecutorType||rec.source||rec.Source||'-', failed=!!(rec.failed||rec.Failed), recordKey='record:'+String(rec.id||rec.request_id||rawT||idx), changes=state.pendingChanges&&state.pendingChanges.records;
    var parsedTime=new Date(rawT), mobileTime=Number.isNaN(parsedTime.getTime())?'—':parsedTime.toLocaleTimeString('ko-KR',{timeZone:'Asia/Seoul',hour:'2-digit',minute:'2-digit',hour12:false}); // mobileTime.textContent, timeCell.title=formatDateTime(rawT)
    // Compose semantic UI Table cells without HTML string concatenation (legacy: mobileTime.textContent). // mobileTime.textContent, timeCell.title=formatDateTime(rawT)
    var statusCell={badge:{text:failed?'실패':'성공',className:'badge '+(failed?'badge-fail':'badge-success')},tone:changes&&changes[recordKey]||null,label:'요청 상태'};
    return {key:String(rec.id||rec.request_id||rawT||idx),dataIndex:idx,tabIndex:0,onClick:function(e){showRecordDetail(idx,e.currentTarget);},onKeyDown:function(e){if(e.key==='Enter'||e.key===' '){e.preventDefault();showRecordDetail(idx,e.currentTarget);}},cells:[
      {time:{full:formatDateTime(rawT),mobile:mobileTime},title:formatDateTime(rawT)},
      {text:displayProv,className:'cell-ellipsis',title:displayProv},
      {text:String(endpoint),className:'cell-ellipsis',title:String(endpoint)},
      {text:String(model),className:'cell-ellipsis',title:String(model)},
      {text:formatExactNumber(rec.fresh_input != null ? rec.fresh_input : null),className:'num-cell num-input'},
      {text:formatExactNumber(rec.cached),className:'num-cell num-input'},
      {text:formatExactNumber(rec.output != null ? rec.output : null),className:'num-cell num-output'},
      {text:formatExactNumber(rec.total != null ? rec.total : null),className:'num-cell num-total'},
      {text:formatLatency(rec.latency_ms != null ? rec.latency_ms : (rec.latency != null ? rec.latency : rec.Latency)),className:'num-cell'},
      statusCell
    ]};
  });
  if(window.renderDohwiTable)window.renderDohwiTable('recordsTableMount',rows,'표시할 기록이 없습니다.');else window.pendingDohwiTableUpdates=(window.pendingDohwiTableUpdates||[]).concat([['recordsTableMount',rows,'표시할 기록이 없습니다.']]);
  var renderedRows=document.querySelectorAll('#recordsBody tr.ui-table-row');
  renderedRows.forEach(function(row,index){row.dataset.index=String(index);row.dataset.rowKey=rows[index]?rows[index].key:'';var rec=state.records[index];if(rec&&state.pendingChanges&&state.pendingChanges.records['record:'+String(rec.id||rec.request_id||rec.requested_at||index)])applyPanelPulse(document.getElementById('recordsPanel'));});
  loadMoreBar.classList.toggle('is-hidden', !state.hasMore);
}
function loadMoreRecords() {
  if (state.inFlight) return;
  if (state.snapshotFilterKey !== snapshotFilterKey()) { loadData(false, true); return; }
  if (!state.hasMore || !state.cursor) return;
  loadData(true);
}
// Record detail modal: focus-safe dialog/sheet with reversible body scroll lock.
function detailRow(key, value, className) { var row=document.createElement('div'), keyEl=document.createElement('div'), valueEl=document.createElement('div');row.className='detail-row';keyEl.className='detail-key';keyEl.textContent=key;valueEl.className='detail-val'+(className?' '+className:'');valueEl.textContent=value;row.appendChild(keyEl);row.appendChild(valueEl);return row; }
function detailSection(title, rows) { var section=document.createElement('section'), heading=document.createElement('div');section.className='detail-section';heading.className='detail-section-title';heading.textContent=title;section.appendChild(heading);rows.forEach(function(row){section.appendChild(row);});return section; }
function lockBodyScroll() { if (state.dialogScrollLock) return; state.dialogScrollLock={overflow:document.body.style.overflow,paddingRight:document.body.style.paddingRight}; var gutter=typeof getComputedStyle==='function'?getComputedStyle(document.documentElement).scrollbarGutter:'';var width=gutter&&gutter.indexOf('stable')!==-1?0:window.innerWidth-document.documentElement.clientWidth; document.body.style.overflow='hidden'; if(width>0)document.body.style.paddingRight=width+'px'; }
function unlockBodyScroll() { if(!state.dialogScrollLock)return; document.body.style.overflow=state.dialogScrollLock.overflow;document.body.style.paddingRight=state.dialogScrollLock.paddingRight;state.dialogScrollLock=null; }
function showRecordDetail(idx, opener) {
  var rec=state.records[idx];if(!rec)return;var dialog=DOM.recordDialog||document.getElementById('recordDialog'),overlay=DOM.detailOverlay||document.getElementById('detailOverlay'),body=DOM.dialogBody||document.getElementById('dialogBody'),rawT=rec.requested_at||rec.t||rec.RequestedAt||rec.Timestamp,rawProv=rec.provider||rec.Provider||'-',displayProv=formatProvider(rawProv),model=rec.model||rec.Model||'-',alias=rec.alias||rec.Alias||'-',endpoint=rec.endpoint||rec.executor_type||rec.ExecutorType||rec.source||rec.Source||'-',failed=!!(rec.failed||rec.Failed),latency=formatLatency(rec.latency_ms != null?rec.latency_ms:(rec.latency != null?rec.latency:rec.Latency)),ttft=formatLatency(rec.ttft_ms != null?rec.ttft_ms:(rec.ttft != null?rec.ttft:rec.TTFT));
  if(!dialog||!overlay||!body)return;state.drawerOpener=opener||document.activeElement;body.replaceChildren();var grid=document.createElement('div');grid.className='detail-grid';grid.appendChild(detailSection('기본 요청 정보',[detailRow('요청 시각',formatDateTime(rawT)+' ('+String(rawT||'-')+')'),detailRow('상태',failed?'실패':'성공'),detailRow('실패 상태',String(rec.failure_status||'-')),detailRow('Provider',displayProv+(rawProv!==displayProv?' ('+String(rawProv)+')':'')),detailRow('Model',String(model)),detailRow('Alias',String(alias)),detailRow('Endpoint / Source',String(endpoint))]));grid.appendChild(detailSection('토큰 및 응답 속도',[detailRow('총 토큰 (Authoritative)',formatExactNumber(rec.total != null?rec.total:null)),detailRow('입력(캐시 포함) (Input)',formatExactNumber(rec.input != null?rec.input:null),'num-input'),detailRow('신규 입력 (Fresh)',formatExactNumber(rec.fresh_input != null?rec.fresh_input:null),'num-input'),detailRow('출력 토큰 (Output)',formatExactNumber(rec.output != null?rec.output:null),'num-output'),detailRow('추론 토큰 (Reasoning)',formatExactNumber(rec.reasoning != null?rec.reasoning:0)),detailRow('캐시 입력 (Cached)',formatExactNumber(rec.cached != null?rec.cached:0)),detailRow('총 Latency',latency),detailRow('TTFT (첫 토큰 시간)',ttft)]));body.appendChild(grid);lockBodyScroll();overlay.classList.add('open');overlay.setAttribute('aria-hidden','false');dialog.setAttribute('aria-hidden','false');var close=DOM.recordDialogClose||dialog.querySelector('.dialog-close');if(close)close.focus();
}
function closeDrawer() { var dialog=DOM.recordDialog||document.getElementById('recordDialog'),overlay=DOM.detailOverlay||document.getElementById('detailOverlay');var wasOpen=!!(overlay&&overlay.classList.contains('open'));if(dialog)dialog.setAttribute('aria-hidden','true');if(overlay){overlay.classList.remove('open');overlay.setAttribute('aria-hidden','true');}unlockBodyScroll();var opener=state.drawerOpener;state.drawerOpener=null;if(wasOpen&&opener&&document.contains(opener))opener.focus(); }
function trapDrawerFocus(e) { var dialog=DOM.recordDialog||document.getElementById('recordDialog'),overlay=DOM.detailOverlay||document.getElementById('detailOverlay');if(e.key!=='Tab'||!dialog||!overlay||!overlay.classList.contains('open'))return;var items=dialog.querySelectorAll('button,[href],input,select,textarea,[tabindex]:not([tabindex="-1"])');if(!items.length){e.preventDefault();return;}var first=items[0],last=items[items.length-1];if(e.shiftKey&&document.activeElement===first){e.preventDefault();last.focus();}else if(!e.shiftKey&&document.activeElement===last){e.preventDefault();first.focus();} }

// Visibility change handler: pause on hidden, immediate refresh + resume on visible
document.addEventListener('visibilitychange', function() {
  if (document.hidden) {
    // Cancel pending poll timer; will refresh on visible.
    if (state.pollTimer) { clearTimeout(state.pollTimer); state.pollTimer = null; }
  } else {
    // Refresh immediately on becoming visible.
    if (!state.inFlight) loadData(false);
  }
});

// Window unload cleanup
window.addEventListener('beforeunload', function() {
  if (state.pollTimer) { clearTimeout(state.pollTimer); state.pollTimer = null; }
  if (state.abortCtrl) {
    state.abortCtrl.abort();
    state.abortCtrl = null;
  }
});

function initDashboard() {
  // Cache frequently accessed DOM refs once (reused every refresh)
  DOM.errorBox = document.getElementById('errorBox');
  DOM.errorMessage = document.getElementById('errorMessage');
  DOM.retryLoadBtn = document.getElementById('retryLoadBtn');
  DOM.dashboardContent = document.getElementById('dashboardContent');
  DOM.usageSkeleton = document.getElementById('usageSkeleton');
  DOM.usageDataContent = document.getElementById('usageDataContent');
  DOM.usageMainContent = document.getElementById('usageMainContent');
  DOM.activeFilterBadge = document.getElementById('activeFilterBadge');
  DOM.activeFilterText = document.getElementById('activeFilterText');
  DOM.customDateWrap = document.getElementById('customDateWrap');
  DOM.filterFromDate = document.getElementById('filterFromDate'); DOM.filterToDate = document.getElementById('filterToDate');
  DOM.filterProvider = document.getElementById('filterProvider'); DOM.filterEndpoint = document.getElementById('filterEndpoint'); DOM.filterModel = document.getElementById('filterModel'); DOM.filterStatus = document.getElementById('filterStatus');
  DOM.kpiTotalCount = document.getElementById('kpiTotalCount');
  DOM.kpiSubFail = document.getElementById('kpiSubFail');
  DOM.kpiFailureRate = document.getElementById('kpiFailureRate');
  DOM.kpiSubReq = document.getElementById('kpiSubReq');
  DOM.kpiTotalTokens = document.getElementById('kpiTotalTokens');
  DOM.kpiSubTokens = document.getElementById('kpiSubTokens');
  DOM.kpiInputTokens = document.getElementById('kpiInputTokens');
  DOM.kpiFreshInputTokens = document.getElementById('kpiFreshInputTokens');
  DOM.kpiCachedTokens = document.getElementById('kpiCachedTokens');
  DOM.kpiOutputTokens = document.getElementById('kpiOutputTokens');
  DOM.kpiReasoningTokens = document.getElementById('kpiReasoningTokens');
  DOM.kpiActiveEntities = document.getElementById('kpiActiveEntities');
  DOM.kpiSubEntities = document.getElementById('kpiSubEntities');
  DOM.chartCard = document.getElementById('chartCard');
  DOM.chartArea = document.getElementById('chartArea');
  DOM.dailyChartTooltip = document.getElementById('dailyChartTooltip');
  DOM.dailyChartSummary = document.getElementById('dailyChartSummary');
  DOM.breakdownSection = document.getElementById('breakdownSection');
  DOM.recordsBody = document.getElementById('recordsBody');
  DOM.recordsCountBadge = document.getElementById('recordsCountBadge');
  DOM.loadMoreBar = document.getElementById('loadMoreBar');
  DOM.recordDialog = document.getElementById('recordDialog'); DOM.detailOverlay = document.getElementById('detailOverlay'); DOM.dialogBody = document.getElementById('dialogBody'); DOM.recordDialogClose = document.getElementById('recordDialogClose');
  if (DOM.retryLoadBtn) DOM.retryLoadBtn.addEventListener('click', function() { state.recordsExpanded = false; loadData(false, true); });
  if (DOM.recordDialogClose) DOM.recordDialogClose.addEventListener('click', closeDrawer);
  if (DOM.detailOverlay) DOM.detailOverlay.addEventListener('click', function(e) { if (e.target === DOM.detailOverlay) closeDrawer(); });

  closeDrawer();
  var bounds = computePresetDates('today');
  if (bounds) {
    state.fromRFC3339 = bounds.from;
    state.toRFC3339 = bounds.to;
  }
  syncPeriodInputs();
  updateActiveFilterBadge();
  loadData(false);
}

function bindDashboardEvents() {
  if (typeof window !== 'undefined' && window.addEventListener) window.addEventListener('resize', scheduleDailyChartRedraw);
  if (document.fonts && document.fonts.ready) document.fonts.ready.then(scheduleDailyChartRedraw);
  var chartCard = document.getElementById('chartCard');
  if (chartCard) {
    if (typeof ResizeObserver !== 'undefined') {
      var lastChartWidth = chartCard.clientWidth;
      var chartResizeObserver = new ResizeObserver(function() {
        var width = chartCard.clientWidth;
        if (width !== lastChartWidth) { lastChartWidth = width; scheduleDailyChartRedraw(); }
      });
      chartResizeObserver.observe(chartCard);
    }
  }
  // Fallback buttons are used only if the exported React components did not mount.
  var pBtns = window.renderDashboardPresets ? [] : document.querySelectorAll('#presetGroup .preset-btn');
  pBtns.forEach(function(btn, index) {
    btn.addEventListener('click', function() {
      setPreset(btn.getAttribute('data-preset'));
    });
    // Radix RadioGroup keyboard model: arrows move focus and selection, Home/End jump.
    btn.addEventListener('keydown', function(e) {
      var list = Array.prototype.slice.call(pBtns), at = list.indexOf(btn), next = -1;
      if (e.key === 'ArrowRight' || e.key === 'ArrowDown') next = (at + 1) % list.length;
      else if (e.key === 'ArrowLeft' || e.key === 'ArrowUp') next = (at + list.length - 1) % list.length;
      else if (e.key === 'Home') next = 0; else if (e.key === 'End') next = list.length - 1;
      if (next < 0) return;
      e.preventDefault(); list[next].focus(); setPreset(list[next].getAttribute('data-preset'));
    });
  });

  // Escape and outside tap close transient overlays.
  document.addEventListener('keydown', function(e) {
    if (e.key === 'Escape') { closeDrawer(); closeDailyChartTooltip(); } trapDrawerFocus(e);
  });
  document.addEventListener('pointerdown', function(e) {
    var area = DOM.chartArea || document.getElementById('chartArea');
    if (!area || !area.contains(e.target)) closeDailyChartTooltip();
  });

  document.querySelectorAll('.tab-btn').forEach(function(tab) { tab.addEventListener('keydown', function(e) { if (e.key !== 'ArrowLeft' && e.key !== 'ArrowRight' && e.key !== 'Home' && e.key !== 'End') return; e.preventDefault(); var tabs=Array.prototype.slice.call(document.querySelectorAll('.tab-btn')), index=tabs.indexOf(tab); if(e.key==='ArrowLeft')index=(index+tabs.length-1)%tabs.length; else if(e.key==='ArrowRight')index=(index+1)%tabs.length; else if(e.key==='Home')index=0; else index=tabs.length-1; switchTab(tabs[index].getAttribute('aria-controls'),tabs[index]); }); });
  var filterForm = document.getElementById('filterForm');
  if (filterForm) filterForm.addEventListener('submit', function(e) { e.preventDefault(); applyFilters(); });
  var resetFiltersBtn = document.getElementById('resetFiltersBtn');
  if (resetFiltersBtn) resetFiltersBtn.addEventListener('click', resetFilters);

  [DOM.filterProvider || document.getElementById('filterProvider'), DOM.filterEndpoint || document.getElementById('filterEndpoint'), DOM.filterModel || document.getElementById('filterModel'), DOM.filterStatus || document.getElementById('filterStatus')].forEach(function(field) {
    if (field) {
      updateSelectTitle(field);
      field.addEventListener('change', function() { updateSelectTitle(field); });
      field.addEventListener('change', applyFilters);
    }
  });

  function applyCustomDatesWhenReady() {
    if (state.currentPreset !== 'custom') return;
    var from = DOM.filterFromDate || document.getElementById('filterFromDate');
    var to = DOM.filterToDate || document.getElementById('filterToDate');
    if (!from.value || !to.value || from.value > to.value) {
      // Editing one endpoint is an unfinished draft, not an error to interrupt.
      setFilterValidation('');
      return;
    }
    applyFilters();
  }
  [DOM.filterFromDate || document.getElementById('filterFromDate'), DOM.filterToDate || document.getElementById('filterToDate')].forEach(function(field) {
    if (!field) return;
    field.addEventListener('change', applyCustomDatesWhenReady);
    field.addEventListener('keydown', function(e) { if (e.key === 'Enter') { e.preventDefault(); applyFilters(); } });
  });
}
// The static DOM is already present: bind interactions before starting requests too.
bindDashboardEvents();
// Start API requests before parsing/mounting the embedded React bundle. The DOM above is
// already present; waiting for DOMContentLoaded would serialize network behind that work.
initDashboard();
</script>
<script data-dohwi-ui="exported">
/* dohwi-ui Table primitives: TableContainer, Table, TableHeader, TableBody, TableRow, TableHead, TableCell */
/* Dohwi Table adapters preserve safe textContent, title={cell.title}, cellContent(cell), mobileTime:mobileTime, flashValue(el,cell.tone,cell.label). */
(()=>{var vb=Object.create;var _f=Object.defineProperty,Sb=Object.defineProperties,xb=Object.getOwnPropertyDescriptor,Lb=Object.getOwnPropertyDescriptors,yb=Object.getOwnPropertyNames,Vr=Object.getOwnPropertySymbols,Cb=Object.getPrototypeOf,Uf=Object.prototype.hasOwnProperty,Ch=Object.prototype.propertyIsEnumerable;var Rh=e=>{throw TypeError(e)};var yh=(e,t,a)=>t in e?_f(e,t,{enumerable:!0,configurable:!0,writable:!0,value:a}):e[t]=a,y=(e,t)=>{for(var a in t||(t={}))Uf.call(t,a)&&yh(e,a,t[a]);if(Vr)for(var a of Vr(t))Ch.call(t,a)&&yh(e,a,t[a]);return e},b=(e,t)=>Sb(e,Lb(t));var D=(e,t)=>{var a={};for(var o in e)Uf.call(e,o)&&t.indexOf(o)<0&&(a[o]=e[o]);if(e!=null&&Vr)for(var o of Vr(e))t.indexOf(o)<0&&Ch.call(e,o)&&(a[o]=e[o]);return a};var qa=(e,t)=>()=>{try{return t||e((t={exports:{}}).exports,t),t.exports}catch(a){throw t=0,a}};var Rb=(e,t,a,o)=>{if(t&&typeof t=="object"||typeof t=="function")for(let l of yb(t))!Uf.call(e,l)&&l!==a&&_f(e,l,{get:()=>t[l],enumerable:!(o=xb(t,l))||o.enumerable});return e};var T=(e,t,a)=>(a=e!=null?vb(Cb(e)):{},Rb(t||!e||!e.__esModule?_f(a,"default",{value:e,enumerable:!0}):a,e));var Ih=(e,t,a)=>t.has(e)||Rh("Cannot "+a);var dt=(e,t,a)=>(Ih(e,t,"read from private field"),a?a.call(e):t.get(e)),Nf=(e,t,a)=>t.has(e)?Rh("Cannot add the same private member more than once"):t instanceof WeakSet?t.add(e):t.set(e,a),Xr=(e,t,a,o)=>(Ih(e,t,"write to private field"),o?o.call(e,a):t.set(e,a),a);var ua=(e,t,a)=>new Promise((o,l)=>{var n=i=>{try{r(a.next(i))}catch(s){l(s)}},u=i=>{try{r(a.throw(i))}catch(s){l(s)}},r=i=>i.done?o(i.value):Promise.resolve(i.value).then(n,u);r((a=a.apply(e,t)).next())});var Bh=qa(Ae=>{"use strict";/**
 * @license React
 * scheduler.production.js
 *
 * Copyright (c) Meta Platforms, Inc. and affiliates.
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */function Ff(e,t){var a=e.length;e.push(t);e:for(;0<a;){var o=a-1>>>1,l=e[o];if(0<jr(l,t))e[o]=t,e[a]=l,a=o;else break e}}function Ga(e){return e.length===0?null:e[0]}function Kr(e){if(e.length===0)return null;var t=e[0],a=e.pop();if(a!==t){e[0]=a;e:for(var o=0,l=e.length,n=l>>>1;o<n;){var u=2*(o+1)-1,r=e[u],i=u+1,s=e[i];if(0>jr(r,a))i<l&&0>jr(s,r)?(e[o]=s,e[i]=a,o=i):(e[o]=r,e[u]=a,o=u);else if(i<l&&0>jr(s,a))e[o]=s,e[i]=a,o=i;else break e}}return t}function jr(e,t){var a=e.sortIndex-t.sortIndex;return a!==0?a:e.id-t.id}Ae.unstable_now=void 0;typeof performance=="object"&&typeof performance.now=="function"?(bh=performance,Ae.unstable_now=function(){return bh.now()}):(kf=Date,wh=kf.now(),Ae.unstable_now=function(){return kf.now()-wh});var bh,kf,wh,lo=[],Uo=[],Ib=1,ra=null,Lt=3,qf=!1,gu=!1,vu=!1,Gf=!1,Eh=typeof setTimeout=="function"?setTimeout:null,Dh=typeof clearTimeout=="function"?clearTimeout:null,Ah=typeof setImmediate!="undefined"?setImmediate:null;function Yr(e){for(var t=Ga(Uo);t!==null;){if(t.callback===null)Kr(Uo);else if(t.startTime<=e)Kr(Uo),t.sortIndex=t.expirationTime,Ff(lo,t);else break;t=Ga(Uo)}}function Vf(e){if(vu=!1,Yr(e),!gu)if(Ga(lo)!==null)gu=!0,ln||(ln=!0,on());else{var t=Ga(Uo);t!==null&&Xf(Vf,t.startTime-e)}}var ln=!1,Su=-1,Mh=5,Oh=-1;function Ph(){return Gf?!0:!(Ae.unstable_now()-Oh<Mh)}function Hf(){if(Gf=!1,ln){var e=Ae.unstable_now();Oh=e;var t=!0;try{e:{gu=!1,vu&&(vu=!1,Dh(Su),Su=-1),qf=!0;var a=Lt;try{t:{for(Yr(e),ra=Ga(lo);ra!==null&&!(ra.expirationTime>e&&Ph());){var o=ra.callback;if(typeof o=="function"){ra.callback=null,Lt=ra.priorityLevel;var l=o(ra.expirationTime<=e);if(e=Ae.unstable_now(),typeof l=="function"){ra.callback=l,Yr(e),t=!0;break t}ra===Ga(lo)&&Kr(lo),Yr(e)}else Kr(lo);ra=Ga(lo)}if(ra!==null)t=!0;else{var n=Ga(Uo);n!==null&&Xf(Vf,n.startTime-e),t=!1}}break e}finally{ra=null,Lt=a,qf=!1}t=void 0}}finally{t?on():ln=!1}}}var on;typeof Ah=="function"?on=function(){Ah(Hf)}:typeof MessageChannel!="undefined"?(zf=new MessageChannel,Th=zf.port2,zf.port1.onmessage=Hf,on=function(){Th.postMessage(null)}):on=function(){Eh(Hf,0)};var zf,Th;function Xf(e,t){Su=Eh(function(){e(Ae.unstable_now())},t)}Ae.unstable_IdlePriority=5;Ae.unstable_ImmediatePriority=1;Ae.unstable_LowPriority=4;Ae.unstable_NormalPriority=3;Ae.unstable_Profiling=null;Ae.unstable_UserBlockingPriority=2;Ae.unstable_cancelCallback=function(e){e.callback=null};Ae.unstable_forceFrameRate=function(e){0>e||125<e?console.error("forceFrameRate takes a positive int between 0 and 125, forcing frame rates higher than 125 fps is not supported"):Mh=0<e?Math.floor(1e3/e):5};Ae.unstable_getCurrentPriorityLevel=function(){return Lt};Ae.unstable_next=function(e){switch(Lt){case 1:case 2:case 3:var t=3;break;default:t=Lt}var a=Lt;Lt=t;try{return e()}finally{Lt=a}};Ae.unstable_requestPaint=function(){Gf=!0};Ae.unstable_runWithPriority=function(e,t){switch(e){case 1:case 2:case 3:case 4:case 5:break;default:e=3}var a=Lt;Lt=e;try{return t()}finally{Lt=a}};Ae.unstable_scheduleCallback=function(e,t,a){var o=Ae.unstable_now();switch(typeof a=="object"&&a!==null?(a=a.delay,a=typeof a=="number"&&0<a?o+a:o):a=o,e){case 1:var l=-1;break;case 2:l=250;break;case 5:l=1073741823;break;case 4:l=1e4;break;default:l=5e3}return l=a+l,e={id:Ib++,callback:t,priorityLevel:e,startTime:a,expirationTime:l,sortIndex:-1},a>o?(e.sortIndex=a,Ff(Uo,e),Ga(lo)===null&&e===Ga(Uo)&&(vu?(Dh(Su),Su=-1):vu=!0,Xf(Vf,a-o))):(e.sortIndex=l,Ff(lo,e),gu||qf||(gu=!0,ln||(ln=!0,on()))),e};Ae.unstable_shouldYield=Ph;Ae.unstable_wrapCallback=function(e){var t=Lt;return function(){var a=Lt;Lt=t;try{return e.apply(this,arguments)}finally{Lt=a}}}});var Uh=qa((T1,_h)=>{"use strict";_h.exports=Bh()});var Yh=qa(Y=>{"use strict";/**
 * @license React
 * react.production.js
 *
 * Copyright (c) Meta Platforms, Inc. and affiliates.
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */var Kf=Symbol.for("react.transitional.element"),bb=Symbol.for("react.portal"),wb=Symbol.for("react.fragment"),Ab=Symbol.for("react.strict_mode"),Tb=Symbol.for("react.profiler"),Eb=Symbol.for("react.consumer"),Db=Symbol.for("react.context"),Mb=Symbol.for("react.forward_ref"),Ob=Symbol.for("react.suspense"),Pb=Symbol.for("react.memo"),Fh=Symbol.for("react.lazy"),Bb=Symbol.for("react.activity"),Nh=Symbol.iterator;function _b(e){return e===null||typeof e!="object"?null:(e=Nh&&e[Nh]||e["@@iterator"],typeof e=="function"?e:null)}var qh={isMounted:function(){return!1},enqueueForceUpdate:function(){},enqueueReplaceState:function(){},enqueueSetState:function(){}},Gh=Object.assign,Vh={};function un(e,t,a){this.props=e,this.context=t,this.refs=Vh,this.updater=a||qh}un.prototype.isReactComponent={};un.prototype.setState=function(e,t){if(typeof e!="object"&&typeof e!="function"&&e!=null)throw Error("takes an object of state variables to update or a function which returns an object of state variables.");this.updater.enqueueSetState(this,e,t,"setState")};un.prototype.forceUpdate=function(e){this.updater.enqueueForceUpdate(this,e,"forceUpdate")};function Xh(){}Xh.prototype=un.prototype;function Zf(e,t,a){this.props=e,this.context=t,this.refs=Vh,this.updater=a||qh}var Qf=Zf.prototype=new Xh;Qf.constructor=Zf;Gh(Qf,un.prototype);Qf.isPureReactComponent=!0;var kh=Array.isArray;function Yf(){}var ye={H:null,A:null,T:null,S:null},jh=Object.prototype.hasOwnProperty;function Wf(e,t,a){var o=a.ref;return{$$typeof:Kf,type:e,key:t,ref:o!==void 0?o:null,props:a}}function Ub(e,t){return Wf(e.type,t,e.props)}function Jf(e){return typeof e=="object"&&e!==null&&e.$$typeof===Kf}function Nb(e){var t={"=":"=0",":":"=2"};return"$"+e.replace(/[=:]/g,function(a){return t[a]})}var Hh=/\/+/g;function jf(e,t){return typeof e=="object"&&e!==null&&e.key!=null?Nb(""+e.key):t.toString(36)}function kb(e){switch(e.status){case"fulfilled":return e.value;case"rejected":throw e.reason;default:switch(typeof e.status=="string"?e.then(Yf,Yf):(e.status="pending",e.then(function(t){e.status==="pending"&&(e.status="fulfilled",e.value=t)},function(t){e.status==="pending"&&(e.status="rejected",e.reason=t)})),e.status){case"fulfilled":return e.value;case"rejected":throw e.reason}}throw e}function nn(e,t,a,o,l){var n=typeof e;(n==="undefined"||n==="boolean")&&(e=null);var u=!1;if(e===null)u=!0;else switch(n){case"bigint":case"string":case"number":u=!0;break;case"object":switch(e.$$typeof){case Kf:case bb:u=!0;break;case Fh:return u=e._init,nn(u(e._payload),t,a,o,l)}}if(u)return l=l(e),u=o===""?"."+jf(e,0):o,kh(l)?(a="",u!=null&&(a=u.replace(Hh,"$&/")+"/"),nn(l,t,a,"",function(s){return s})):l!=null&&(Jf(l)&&(l=Ub(l,a+(l.key==null||e&&e.key===l.key?"":(""+l.key).replace(Hh,"$&/")+"/")+u)),t.push(l)),1;u=0;var r=o===""?".":o+":";if(kh(e))for(var i=0;i<e.length;i++)o=e[i],n=r+jf(o,i),u+=nn(o,t,a,n,l);else if(i=_b(e),typeof i=="function")for(e=i.call(e),i=0;!(o=e.next()).done;)o=o.value,n=r+jf(o,i++),u+=nn(o,t,a,n,l);else if(n==="object"){if(typeof e.then=="function")return nn(kb(e),t,a,o,l);throw t=String(e),Error("Objects are not valid as a React child (found: "+(t==="[object Object]"?"object with keys {"+Object.keys(e).join(", ")+"}":t)+"). If you meant to render a collection of children, use an array instead.")}return u}function Zr(e,t,a){if(e==null)return e;var o=[],l=0;return nn(e,o,"","",function(n){return t.call(a,n,l++)}),o}function Hb(e){if(e._status===-1){var t=e._result;t=t(),t.then(function(a){(e._status===0||e._status===-1)&&(e._status=1,e._result=a)},function(a){(e._status===0||e._status===-1)&&(e._status=2,e._result=a)}),e._status===-1&&(e._status=0,e._result=t)}if(e._status===1)return e._result.default;throw e._result}var zh=typeof reportError=="function"?reportError:function(e){if(typeof window=="object"&&typeof window.ErrorEvent=="function"){var t=new window.ErrorEvent("error",{bubbles:!0,cancelable:!0,message:typeof e=="object"&&e!==null&&typeof e.message=="string"?String(e.message):String(e),error:e});if(!window.dispatchEvent(t))return}else if(typeof process=="object"&&typeof process.emit=="function"){process.emit("uncaughtException",e);return}console.error(e)},zb={map:Zr,forEach:function(e,t,a){Zr(e,function(){t.apply(this,arguments)},a)},count:function(e){var t=0;return Zr(e,function(){t++}),t},toArray:function(e){return Zr(e,function(t){return t})||[]},only:function(e){if(!Jf(e))throw Error("React.Children.only expected to receive a single React element child.");return e}};Y.Activity=Bb;Y.Children=zb;Y.Component=un;Y.Fragment=wb;Y.Profiler=Tb;Y.PureComponent=Zf;Y.StrictMode=Ab;Y.Suspense=Ob;Y.__CLIENT_INTERNALS_DO_NOT_USE_OR_WARN_USERS_THEY_CANNOT_UPGRADE=ye;Y.__COMPILER_RUNTIME={__proto__:null,c:function(e){return ye.H.useMemoCache(e)}};Y.cache=function(e){return function(){return e.apply(null,arguments)}};Y.cacheSignal=function(){return null};Y.cloneElement=function(e,t,a){if(e==null)throw Error("The argument must be a React element, but you passed "+e+".");var o=Gh({},e.props),l=e.key;if(t!=null)for(n in t.key!==void 0&&(l=""+t.key),t)!jh.call(t,n)||n==="key"||n==="__self"||n==="__source"||n==="ref"&&t.ref===void 0||(o[n]=t[n]);var n=arguments.length-2;if(n===1)o.children=a;else if(1<n){for(var u=Array(n),r=0;r<n;r++)u[r]=arguments[r+2];o.children=u}return Wf(e.type,l,o)};Y.createContext=function(e){return e={$$typeof:Db,_currentValue:e,_currentValue2:e,_threadCount:0,Provider:null,Consumer:null},e.Provider=e,e.Consumer={$$typeof:Eb,_context:e},e};Y.createElement=function(e,t,a){var o,l={},n=null;if(t!=null)for(o in t.key!==void 0&&(n=""+t.key),t)jh.call(t,o)&&o!=="key"&&o!=="__self"&&o!=="__source"&&(l[o]=t[o]);var u=arguments.length-2;if(u===1)l.children=a;else if(1<u){for(var r=Array(u),i=0;i<u;i++)r[i]=arguments[i+2];l.children=r}if(e&&e.defaultProps)for(o in u=e.defaultProps,u)l[o]===void 0&&(l[o]=u[o]);return Wf(e,n,l)};Y.createRef=function(){return{current:null}};Y.forwardRef=function(e){return{$$typeof:Mb,render:e}};Y.isValidElement=Jf;Y.lazy=function(e){return{$$typeof:Fh,_payload:{_status:-1,_result:e},_init:Hb}};Y.memo=function(e,t){return{$$typeof:Pb,type:e,compare:t===void 0?null:t}};Y.startTransition=function(e){var t=ye.T,a={};ye.T=a;try{var o=e(),l=ye.S;l!==null&&l(a,o),typeof o=="object"&&o!==null&&typeof o.then=="function"&&o.then(Yf,zh)}catch(n){zh(n)}finally{t!==null&&a.types!==null&&(t.types=a.types),ye.T=t}};Y.unstable_useCacheRefresh=function(){return ye.H.useCacheRefresh()};Y.use=function(e){return ye.H.use(e)};Y.useActionState=function(e,t,a){return ye.H.useActionState(e,t,a)};Y.useCallback=function(e,t){return ye.H.useCallback(e,t)};Y.useContext=function(e){return ye.H.useContext(e)};Y.useDebugValue=function(){};Y.useDeferredValue=function(e,t){return ye.H.useDeferredValue(e,t)};Y.useEffect=function(e,t){return ye.H.useEffect(e,t)};Y.useEffectEvent=function(e){return ye.H.useEffectEvent(e)};Y.useId=function(){return ye.H.useId()};Y.useImperativeHandle=function(e,t,a){return ye.H.useImperativeHandle(e,t,a)};Y.useInsertionEffect=function(e,t){return ye.H.useInsertionEffect(e,t)};Y.useLayoutEffect=function(e,t){return ye.H.useLayoutEffect(e,t)};Y.useMemo=function(e,t){return ye.H.useMemo(e,t)};Y.useOptimistic=function(e,t){return ye.H.useOptimistic(e,t)};Y.useReducer=function(e,t,a){return ye.H.useReducer(e,t,a)};Y.useRef=function(e){return ye.H.useRef(e)};Y.useState=function(e){return ye.H.useState(e)};Y.useSyncExternalStore=function(e,t,a){return ye.H.useSyncExternalStore(e,t,a)};Y.useTransition=function(){return ye.H.useTransition()};Y.version="19.2.0"});var B=qa((D1,Kh)=>{"use strict";Kh.exports=Yh()});var Qh=qa(At=>{"use strict";/**
 * @license React
 * react-dom.production.js
 *
 * Copyright (c) Meta Platforms, Inc. and affiliates.
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */var Fb=B();function Zh(e){var t="https://react.dev/errors/"+e;if(1<arguments.length){t+="?args[]="+encodeURIComponent(arguments[1]);for(var a=2;a<arguments.length;a++)t+="&args[]="+encodeURIComponent(arguments[a])}return"Minified React error #"+e+"; visit "+t+" for the full message or use the non-minified dev environment for full errors and additional helpful warnings."}function No(){}var wt={d:{f:No,r:function(){throw Error(Zh(522))},D:No,C:No,L:No,m:No,X:No,S:No,M:No},p:0,findDOMNode:null},qb=Symbol.for("react.portal");function Gb(e,t,a){var o=3<arguments.length&&arguments[3]!==void 0?arguments[3]:null;return{$$typeof:qb,key:o==null?null:""+o,children:e,containerInfo:t,implementation:a}}var xu=Fb.__CLIENT_INTERNALS_DO_NOT_USE_OR_WARN_USERS_THEY_CANNOT_UPGRADE;function Qr(e,t){if(e==="font")return"";if(typeof t=="string")return t==="use-credentials"?t:""}At.__DOM_INTERNALS_DO_NOT_USE_OR_WARN_USERS_THEY_CANNOT_UPGRADE=wt;At.createPortal=function(e,t){var a=2<arguments.length&&arguments[2]!==void 0?arguments[2]:null;if(!t||t.nodeType!==1&&t.nodeType!==9&&t.nodeType!==11)throw Error(Zh(299));return Gb(e,t,null,a)};At.flushSync=function(e){var t=xu.T,a=wt.p;try{if(xu.T=null,wt.p=2,e)return e()}finally{xu.T=t,wt.p=a,wt.d.f()}};At.preconnect=function(e,t){typeof e=="string"&&(t?(t=t.crossOrigin,t=typeof t=="string"?t==="use-credentials"?t:"":void 0):t=null,wt.d.C(e,t))};At.prefetchDNS=function(e){typeof e=="string"&&wt.d.D(e)};At.preinit=function(e,t){if(typeof e=="string"&&t&&typeof t.as=="string"){var a=t.as,o=Qr(a,t.crossOrigin),l=typeof t.integrity=="string"?t.integrity:void 0,n=typeof t.fetchPriority=="string"?t.fetchPriority:void 0;a==="style"?wt.d.S(e,typeof t.precedence=="string"?t.precedence:void 0,{crossOrigin:o,integrity:l,fetchPriority:n}):a==="script"&&wt.d.X(e,{crossOrigin:o,integrity:l,fetchPriority:n,nonce:typeof t.nonce=="string"?t.nonce:void 0})}};At.preinitModule=function(e,t){if(typeof e=="string")if(typeof t=="object"&&t!==null){if(t.as==null||t.as==="script"){var a=Qr(t.as,t.crossOrigin);wt.d.M(e,{crossOrigin:a,integrity:typeof t.integrity=="string"?t.integrity:void 0,nonce:typeof t.nonce=="string"?t.nonce:void 0})}}else t==null&&wt.d.M(e)};At.preload=function(e,t){if(typeof e=="string"&&typeof t=="object"&&t!==null&&typeof t.as=="string"){var a=t.as,o=Qr(a,t.crossOrigin);wt.d.L(e,a,{crossOrigin:o,integrity:typeof t.integrity=="string"?t.integrity:void 0,nonce:typeof t.nonce=="string"?t.nonce:void 0,type:typeof t.type=="string"?t.type:void 0,fetchPriority:typeof t.fetchPriority=="string"?t.fetchPriority:void 0,referrerPolicy:typeof t.referrerPolicy=="string"?t.referrerPolicy:void 0,imageSrcSet:typeof t.imageSrcSet=="string"?t.imageSrcSet:void 0,imageSizes:typeof t.imageSizes=="string"?t.imageSizes:void 0,media:typeof t.media=="string"?t.media:void 0})}};At.preloadModule=function(e,t){if(typeof e=="string")if(t){var a=Qr(t.as,t.crossOrigin);wt.d.m(e,{as:typeof t.as=="string"&&t.as!=="script"?t.as:void 0,crossOrigin:a,integrity:typeof t.integrity=="string"?t.integrity:void 0})}else wt.d.m(e)};At.requestFormReset=function(e){wt.d.r(e)};At.unstable_batchedUpdates=function(e,t){return e(t)};At.useFormState=function(e,t,a){return xu.H.useFormState(e,t,a)};At.useFormStatus=function(){return xu.H.useHostTransitionStatus()};At.version="19.2.0"});var Va=qa((O1,Jh)=>{"use strict";function Wh(){if(!(typeof __REACT_DEVTOOLS_GLOBAL_HOOK__=="undefined"||typeof __REACT_DEVTOOLS_GLOBAL_HOOK__.checkDCE!="function"))try{__REACT_DEVTOOLS_GLOBAL_HOOK__.checkDCE(Wh)}catch(e){console.error(e)}}Wh(),Jh.exports=Qh()});var fL=qa(ys=>{"use strict";/**
 * @license React
 * react-dom-client.production.js
 *
 * Copyright (c) Meta Platforms, Inc. and affiliates.
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */var Ze=Uh(),Iv=B(),Vb=Va();function E(e){var t="https://react.dev/errors/"+e;if(1<arguments.length){t+="?args[]="+encodeURIComponent(arguments[1]);for(var a=2;a<arguments.length;a++)t+="&args[]="+encodeURIComponent(arguments[a])}return"Minified React error #"+e+"; visit "+t+" for the full message or use the non-minified dev environment for full errors and additional helpful warnings."}function bv(e){return!(!e||e.nodeType!==1&&e.nodeType!==9&&e.nodeType!==11)}function nr(e){var t=e,a=e;if(e.alternate)for(;t.return;)t=t.return;else{e=t;do t=e,(t.flags&4098)!==0&&(a=t.return),e=t.return;while(e)}return t.tag===3?a:null}function wv(e){if(e.tag===13){var t=e.memoizedState;if(t===null&&(e=e.alternate,e!==null&&(t=e.memoizedState)),t!==null)return t.dehydrated}return null}function Av(e){if(e.tag===31){var t=e.memoizedState;if(t===null&&(e=e.alternate,e!==null&&(t=e.memoizedState)),t!==null)return t.dehydrated}return null}function $h(e){if(nr(e)!==e)throw Error(E(188))}function Xb(e){var t=e.alternate;if(!t){if(t=nr(e),t===null)throw Error(E(188));return t!==e?null:e}for(var a=e,o=t;;){var l=a.return;if(l===null)break;var n=l.alternate;if(n===null){if(o=l.return,o!==null){a=o;continue}break}if(l.child===n.child){for(n=l.child;n;){if(n===a)return $h(l),e;if(n===o)return $h(l),t;n=n.sibling}throw Error(E(188))}if(a.return!==o.return)a=l,o=n;else{for(var u=!1,r=l.child;r;){if(r===a){u=!0,a=l,o=n;break}if(r===o){u=!0,o=l,a=n;break}r=r.sibling}if(!u){for(r=n.child;r;){if(r===a){u=!0,a=n,o=l;break}if(r===o){u=!0,o=n,a=l;break}r=r.sibling}if(!u)throw Error(E(189))}}if(a.alternate!==o)throw Error(E(190))}if(a.tag!==3)throw Error(E(188));return a.stateNode.current===a?e:t}function Tv(e){var t=e.tag;if(t===5||t===26||t===27||t===6)return e;for(e=e.child;e!==null;){if(t=Tv(e),t!==null)return t;e=e.sibling}return null}var Ie=Object.assign,jb=Symbol.for("react.element"),Wr=Symbol.for("react.transitional.element"),Au=Symbol.for("react.portal"),pn=Symbol.for("react.fragment"),Ev=Symbol.for("react.strict_mode"),Oc=Symbol.for("react.profiler"),Dv=Symbol.for("react.consumer"),po=Symbol.for("react.context"),Ad=Symbol.for("react.forward_ref"),Pc=Symbol.for("react.suspense"),Bc=Symbol.for("react.suspense_list"),Td=Symbol.for("react.memo"),ko=Symbol.for("react.lazy"),_c=Symbol.for("react.activity"),Yb=Symbol.for("react.memo_cache_sentinel"),eg=Symbol.iterator;function Lu(e){return e===null||typeof e!="object"?null:(e=eg&&e[eg]||e["@@iterator"],typeof e=="function"?e:null)}var Kb=Symbol.for("react.client.reference");function Uc(e){if(e==null)return null;if(typeof e=="function")return e.$$typeof===Kb?null:e.displayName||e.name||null;if(typeof e=="string")return e;switch(e){case pn:return"Fragment";case Oc:return"Profiler";case Ev:return"StrictMode";case Pc:return"Suspense";case Bc:return"SuspenseList";case _c:return"Activity"}if(typeof e=="object")switch(e.$$typeof){case Au:return"Portal";case po:return e.displayName||"Context";case Dv:return(e._context.displayName||"Context")+".Consumer";case Ad:var t=e.render;return e=e.displayName,e||(e=t.displayName||t.name||"",e=e!==""?"ForwardRef("+e+")":"ForwardRef"),e;case Td:return t=e.displayName||null,t!==null?t:Uc(e.type)||"Memo";case ko:t=e._payload,e=e._init;try{return Uc(e(t))}catch(a){}}return null}var Tu=Array.isArray,V=Iv.__CLIENT_INTERNALS_DO_NOT_USE_OR_WARN_USERS_THEY_CANNOT_UPGRADE,ie=Vb.__DOM_INTERNALS_DO_NOT_USE_OR_WARN_USERS_THEY_CANNOT_UPGRADE,bl={pending:!1,data:null,method:null,action:null},Nc=[],mn=-1;function Za(e){return{current:e}}function nt(e){0>mn||(e.current=Nc[mn],Nc[mn]=null,mn--)}function xe(e,t){mn++,Nc[mn]=e.current,e.current=t}var Ka=Za(null),Xu=Za(null),Zo=Za(null),Di=Za(null);function Mi(e,t){switch(xe(Zo,t),xe(Xu,e),xe(Ka,null),t.nodeType){case 9:case 11:e=(e=t.documentElement)&&(e=e.namespaceURI)?rv(e):0;break;default:if(e=t.tagName,t=t.namespaceURI)t=rv(t),e=Qx(t,e);else switch(e){case"svg":e=1;break;case"math":e=2;break;default:e=0}}nt(Ka),xe(Ka,e)}function On(){nt(Ka),nt(Xu),nt(Zo)}function kc(e){e.memoizedState!==null&&xe(Di,e);var t=Ka.current,a=Qx(t,e.type);t!==a&&(xe(Xu,e),xe(Ka,a))}function Oi(e){Xu.current===e&&(nt(Ka),nt(Xu)),Di.current===e&&(nt(Di),ar._currentValue=bl)}var $f,tg;function yl(e){if($f===void 0)try{throw Error()}catch(a){var t=a.stack.trim().match(/\n( *(at )?)/);$f=t&&t[1]||"",tg=-1<a.stack.indexOf("\n    at")?" (<anonymous>)":-1<a.stack.indexOf("@")?"@unknown:0:0":""}return"\n"+$f+e+tg}var ec=!1;function tc(e,t){if(!e||ec)return"";ec=!0;var a=Error.prepareStackTrace;Error.prepareStackTrace=void 0;try{var o={DetermineComponentFrameRoot:function(){try{if(t){var d=function(){throw Error()};if(Object.defineProperty(d.prototype,"props",{set:function(){throw Error()}}),typeof Reflect=="object"&&Reflect.construct){try{Reflect.construct(d,[])}catch(g){var c=g}Reflect.construct(e,[],d)}else{try{d.call()}catch(g){c=g}e.call(d.prototype)}}else{try{throw Error()}catch(g){c=g}(d=e())&&typeof d.catch=="function"&&d.catch(function(){})}}catch(g){if(g&&c&&typeof g.stack=="string")return[g.stack,c.stack]}return[null,null]}};o.DetermineComponentFrameRoot.displayName="DetermineComponentFrameRoot";var l=Object.getOwnPropertyDescriptor(o.DetermineComponentFrameRoot,"name");l&&l.configurable&&Object.defineProperty(o.DetermineComponentFrameRoot,"name",{value:"DetermineComponentFrameRoot"});var n=o.DetermineComponentFrameRoot(),u=n[0],r=n[1];if(u&&r){var i=u.split("\n"),s=r.split("\n");for(l=o=0;o<i.length&&!i[o].includes("DetermineComponentFrameRoot");)o++;for(;l<s.length&&!s[l].includes("DetermineComponentFrameRoot");)l++;if(o===i.length||l===s.length)for(o=i.length-1,l=s.length-1;1<=o&&0<=l&&i[o]!==s[l];)l--;for(;1<=o&&0<=l;o--,l--)if(i[o]!==s[l]){if(o!==1||l!==1)do if(o--,l--,0>l||i[o]!==s[l]){var f="\n"+i[o].replace(" at new "," at ");return e.displayName&&f.includes("<anonymous>")&&(f=f.replace("<anonymous>",e.displayName)),f}while(1<=o&&0<=l);break}}}finally{ec=!1,Error.prepareStackTrace=a}return(a=e?e.displayName||e.name:"")?yl(a):""}function Zb(e,t){switch(e.tag){case 26:case 27:case 5:return yl(e.type);case 16:return yl("Lazy");case 13:return e.child!==t&&t!==null?yl("Suspense Fallback"):yl("Suspense");case 19:return yl("SuspenseList");case 0:case 15:return tc(e.type,!1);case 11:return tc(e.type.render,!1);case 1:return tc(e.type,!0);case 31:return yl("Activity");default:return""}}function ag(e){try{var t="",a=null;do t+=Zb(e,a),a=e,e=e.return;while(e);return t}catch(o){return"\nError generating stack: "+o.message+"\n"+o.stack}}var Hc=Object.prototype.hasOwnProperty,Ed=Ze.unstable_scheduleCallback,ac=Ze.unstable_cancelCallback,Qb=Ze.unstable_shouldYield,Wb=Ze.unstable_requestPaint,Wt=Ze.unstable_now,Jb=Ze.unstable_getCurrentPriorityLevel,Mv=Ze.unstable_ImmediatePriority,Ov=Ze.unstable_UserBlockingPriority,Pi=Ze.unstable_NormalPriority,$b=Ze.unstable_LowPriority,Pv=Ze.unstable_IdlePriority,e0=Ze.log,t0=Ze.unstable_setDisableYieldValue,ur=null,Jt=null;function Vo(e){if(typeof e0=="function"&&t0(e),Jt&&typeof Jt.setStrictMode=="function")try{Jt.setStrictMode(ur,e)}catch(t){}}var $t=Math.clz32?Math.clz32:l0,a0=Math.log,o0=Math.LN2;function l0(e){return e>>>=0,e===0?32:31-(a0(e)/o0|0)|0}var Jr=256,$r=262144,ei=4194304;function Cl(e){var t=e&42;if(t!==0)return t;switch(e&-e){case 1:return 1;case 2:return 2;case 4:return 4;case 8:return 8;case 16:return 16;case 32:return 32;case 64:return 64;case 128:return 128;case 256:case 512:case 1024:case 2048:case 4096:case 8192:case 16384:case 32768:case 65536:case 131072:return e&261888;case 262144:case 524288:case 1048576:case 2097152:return e&3932160;case 4194304:case 8388608:case 16777216:case 33554432:return e&62914560;case 67108864:return 67108864;case 134217728:return 134217728;case 268435456:return 268435456;case 536870912:return 536870912;case 1073741824:return 0;default:return e}}function ns(e,t,a){var o=e.pendingLanes;if(o===0)return 0;var l=0,n=e.suspendedLanes,u=e.pingedLanes;e=e.warmLanes;var r=o&134217727;return r!==0?(o=r&~n,o!==0?l=Cl(o):(u&=r,u!==0?l=Cl(u):a||(a=r&~e,a!==0&&(l=Cl(a))))):(r=o&~n,r!==0?l=Cl(r):u!==0?l=Cl(u):a||(a=o&~e,a!==0&&(l=Cl(a)))),l===0?0:t!==0&&t!==l&&(t&n)===0&&(n=l&-l,a=t&-t,n>=a||n===32&&(a&4194048)!==0)?t:l}function rr(e,t){return(e.pendingLanes&~(e.suspendedLanes&~e.pingedLanes)&t)===0}function n0(e,t){switch(e){case 1:case 2:case 4:case 8:case 64:return t+250;case 16:case 32:case 128:case 256:case 512:case 1024:case 2048:case 4096:case 8192:case 16384:case 32768:case 65536:case 131072:case 262144:case 524288:case 1048576:case 2097152:return t+5e3;case 4194304:case 8388608:case 16777216:case 33554432:return-1;case 67108864:case 134217728:case 268435456:case 536870912:case 1073741824:return-1;default:return-1}}function Bv(){var e=ei;return ei<<=1,(ei&62914560)===0&&(ei=4194304),e}function oc(e){for(var t=[],a=0;31>a;a++)t.push(e);return t}function ir(e,t){e.pendingLanes|=t,t!==268435456&&(e.suspendedLanes=0,e.pingedLanes=0,e.warmLanes=0)}function u0(e,t,a,o,l,n){var u=e.pendingLanes;e.pendingLanes=a,e.suspendedLanes=0,e.pingedLanes=0,e.warmLanes=0,e.expiredLanes&=a,e.entangledLanes&=a,e.errorRecoveryDisabledLanes&=a,e.shellSuspendCounter=0;var r=e.entanglements,i=e.expirationTimes,s=e.hiddenUpdates;for(a=u&~a;0<a;){var f=31-$t(a),d=1<<f;r[f]=0,i[f]=-1;var c=s[f];if(c!==null)for(s[f]=null,f=0;f<c.length;f++){var g=c[f];g!==null&&(g.lane&=-536870913)}a&=~d}o!==0&&_v(e,o,0),n!==0&&l===0&&e.tag!==0&&(e.suspendedLanes|=n&~(u&~t))}function _v(e,t,a){e.pendingLanes|=t,e.suspendedLanes&=~t;var o=31-$t(t);e.entangledLanes|=t,e.entanglements[o]=e.entanglements[o]|1073741824|a&261930}function Uv(e,t){var a=e.entangledLanes|=t;for(e=e.entanglements;a;){var o=31-$t(a),l=1<<o;l&t|e[o]&t&&(e[o]|=t),a&=~l}}function Nv(e,t){var a=t&-t;return a=(a&42)!==0?1:Dd(a),(a&(e.suspendedLanes|t))!==0?0:a}function Dd(e){switch(e){case 2:e=1;break;case 8:e=4;break;case 32:e=16;break;case 256:case 512:case 1024:case 2048:case 4096:case 8192:case 16384:case 32768:case 65536:case 131072:case 262144:case 524288:case 1048576:case 2097152:case 4194304:case 8388608:case 16777216:case 33554432:e=128;break;case 268435456:e=134217728;break;default:e=0}return e}function Md(e){return e&=-e,2<e?8<e?(e&134217727)!==0?32:268435456:8:2}function kv(){var e=ie.p;return e!==0?e:(e=window.event,e===void 0?32:rL(e.type))}function og(e,t){var a=ie.p;try{return ie.p=e,t()}finally{ie.p=a}}var il=Math.random().toString(36).slice(2),mt="__reactFiber$"+il,Ht="__reactProps$"+il,Gn="__reactContainer$"+il,zc="__reactEvents$"+il,r0="__reactListeners$"+il,i0="__reactHandles$"+il,lg="__reactResources$"+il,sr="__reactMarker$"+il;function Od(e){delete e[mt],delete e[Ht],delete e[zc],delete e[r0],delete e[i0]}function hn(e){var t=e[mt];if(t)return t;for(var a=e.parentNode;a;){if(t=a[Gn]||a[mt]){if(a=t.alternate,t.child!==null||a!==null&&a.child!==null)for(e=dv(e);e!==null;){if(a=e[mt])return a;e=dv(e)}return t}e=a,a=e.parentNode}return null}function Vn(e){if(e=e[mt]||e[Gn]){var t=e.tag;if(t===5||t===6||t===13||t===31||t===26||t===27||t===3)return e}return null}function Eu(e){var t=e.tag;if(t===5||t===26||t===27||t===6)return e.stateNode;throw Error(E(33))}function bn(e){var t=e[lg];return t||(t=e[lg]={hoistableStyles:new Map,hoistableScripts:new Map}),t}function lt(e){e[sr]=!0}var Hv=new Set,zv={};function _l(e,t){Pn(e,t),Pn(e+"Capture",t)}function Pn(e,t){for(zv[e]=t,e=0;e<t.length;e++)Hv.add(t[e])}var s0=RegExp("^[:A-Z_a-z\\u00C0-\\u00D6\\u00D8-\\u00F6\\u00F8-\\u02FF\\u0370-\\u037D\\u037F-\\u1FFF\\u200C-\\u200D\\u2070-\\u218F\\u2C00-\\u2FEF\\u3001-\\uD7FF\\uF900-\\uFDCF\\uFDF0-\\uFFFD][:A-Z_a-z\\u00C0-\\u00D6\\u00D8-\\u00F6\\u00F8-\\u02FF\\u0370-\\u037D\\u037F-\\u1FFF\\u200C-\\u200D\\u2070-\\u218F\\u2C00-\\u2FEF\\u3001-\\uD7FF\\uF900-\\uFDCF\\uFDF0-\\uFFFD\\-.0-9\\u00B7\\u0300-\\u036F\\u203F-\\u2040]*$"),ng={},ug={};function f0(e){return Hc.call(ug,e)?!0:Hc.call(ng,e)?!1:s0.test(e)?ug[e]=!0:(ng[e]=!0,!1)}function hi(e,t,a){if(f0(t))if(a===null)e.removeAttribute(t);else{switch(typeof a){case"undefined":case"function":case"symbol":e.removeAttribute(t);return;case"boolean":var o=t.toLowerCase().slice(0,5);if(o!=="data-"&&o!=="aria-"){e.removeAttribute(t);return}}e.setAttribute(t,""+a)}}function ti(e,t,a){if(a===null)e.removeAttribute(t);else{switch(typeof a){case"undefined":case"function":case"symbol":case"boolean":e.removeAttribute(t);return}e.setAttribute(t,""+a)}}function no(e,t,a,o){if(o===null)e.removeAttribute(a);else{switch(typeof o){case"undefined":case"function":case"symbol":case"boolean":e.removeAttribute(a);return}e.setAttributeNS(t,a,""+o)}}function sa(e){switch(typeof e){case"bigint":case"boolean":case"number":case"string":case"undefined":return e;case"object":return e;default:return""}}function Fv(e){var t=e.type;return(e=e.nodeName)&&e.toLowerCase()==="input"&&(t==="checkbox"||t==="radio")}function c0(e,t,a){var o=Object.getOwnPropertyDescriptor(e.constructor.prototype,t);if(!e.hasOwnProperty(t)&&typeof o!="undefined"&&typeof o.get=="function"&&typeof o.set=="function"){var l=o.get,n=o.set;return Object.defineProperty(e,t,{configurable:!0,get:function(){return l.call(this)},set:function(u){a=""+u,n.call(this,u)}}),Object.defineProperty(e,t,{enumerable:o.enumerable}),{getValue:function(){return a},setValue:function(u){a=""+u},stopTracking:function(){e._valueTracker=null,delete e[t]}}}}function Fc(e){if(!e._valueTracker){var t=Fv(e)?"checked":"value";e._valueTracker=c0(e,t,""+e[t])}}function qv(e){if(!e)return!1;var t=e._valueTracker;if(!t)return!0;var a=t.getValue(),o="";return e&&(o=Fv(e)?e.checked?"true":"false":e.value),e=o,e!==a?(t.setValue(e),!0):!1}function Bi(e){if(e=e||(typeof document!="undefined"?document:void 0),typeof e=="undefined")return null;try{return e.activeElement||e.body}catch(t){return e.body}}var d0=/[\n"\\]/g;function da(e){return e.replace(d0,function(t){return"\\"+t.charCodeAt(0).toString(16)+" "})}function qc(e,t,a,o,l,n,u,r){e.name="",u!=null&&typeof u!="function"&&typeof u!="symbol"&&typeof u!="boolean"?e.type=u:e.removeAttribute("type"),t!=null?u==="number"?(t===0&&e.value===""||e.value!=t)&&(e.value=""+sa(t)):e.value!==""+sa(t)&&(e.value=""+sa(t)):u!=="submit"&&u!=="reset"||e.removeAttribute("value"),t!=null?Gc(e,u,sa(t)):a!=null?Gc(e,u,sa(a)):o!=null&&e.removeAttribute("value"),l==null&&n!=null&&(e.defaultChecked=!!n),l!=null&&(e.checked=l&&typeof l!="function"&&typeof l!="symbol"),r!=null&&typeof r!="function"&&typeof r!="symbol"&&typeof r!="boolean"?e.name=""+sa(r):e.removeAttribute("name")}function Gv(e,t,a,o,l,n,u,r){if(n!=null&&typeof n!="function"&&typeof n!="symbol"&&typeof n!="boolean"&&(e.type=n),t!=null||a!=null){if(!(n!=="submit"&&n!=="reset"||t!=null)){Fc(e);return}a=a!=null?""+sa(a):"",t=t!=null?""+sa(t):a,r||t===e.value||(e.value=t),e.defaultValue=t}o=o!=null?o:l,o=typeof o!="function"&&typeof o!="symbol"&&!!o,e.checked=r?e.checked:!!o,e.defaultChecked=!!o,u!=null&&typeof u!="function"&&typeof u!="symbol"&&typeof u!="boolean"&&(e.name=u),Fc(e)}function Gc(e,t,a){t==="number"&&Bi(e.ownerDocument)===e||e.defaultValue===""+a||(e.defaultValue=""+a)}function wn(e,t,a,o){if(e=e.options,t){t={};for(var l=0;l<a.length;l++)t["$"+a[l]]=!0;for(a=0;a<e.length;a++)l=t.hasOwnProperty("$"+e[a].value),e[a].selected!==l&&(e[a].selected=l),l&&o&&(e[a].defaultSelected=!0)}else{for(a=""+sa(a),t=null,l=0;l<e.length;l++){if(e[l].value===a){e[l].selected=!0,o&&(e[l].defaultSelected=!0);return}t!==null||e[l].disabled||(t=e[l])}t!==null&&(t.selected=!0)}}function Vv(e,t,a){if(t!=null&&(t=""+sa(t),t!==e.value&&(e.value=t),a==null)){e.defaultValue!==t&&(e.defaultValue=t);return}e.defaultValue=a!=null?""+sa(a):""}function Xv(e,t,a,o){if(t==null){if(o!=null){if(a!=null)throw Error(E(92));if(Tu(o)){if(1<o.length)throw Error(E(93));o=o[0]}a=o}a==null&&(a=""),t=a}a=sa(t),e.defaultValue=a,o=e.textContent,o===a&&o!==""&&o!==null&&(e.value=o),Fc(e)}function Bn(e,t){if(t){var a=e.firstChild;if(a&&a===e.lastChild&&a.nodeType===3){a.nodeValue=t;return}}e.textContent=t}var p0=new Set("animationIterationCount aspectRatio borderImageOutset borderImageSlice borderImageWidth boxFlex boxFlexGroup boxOrdinalGroup columnCount columns flex flexGrow flexPositive flexShrink flexNegative flexOrder gridArea gridRow gridRowEnd gridRowSpan gridRowStart gridColumn gridColumnEnd gridColumnSpan gridColumnStart fontWeight lineClamp lineHeight opacity order orphans scale tabSize widows zIndex zoom fillOpacity floodOpacity stopOpacity strokeDasharray strokeDashoffset strokeMiterlimit strokeOpacity strokeWidth MozAnimationIterationCount MozBoxFlex MozBoxFlexGroup MozLineClamp msAnimationIterationCount msFlex msZoom msFlexGrow msFlexNegative msFlexOrder msFlexPositive msFlexShrink msGridColumn msGridColumnSpan msGridRow msGridRowSpan WebkitAnimationIterationCount WebkitBoxFlex WebKitBoxFlexGroup WebkitBoxOrdinalGroup WebkitColumnCount WebkitColumns WebkitFlex WebkitFlexGrow WebkitFlexPositive WebkitFlexShrink WebkitLineClamp".split(" "));function rg(e,t,a){var o=t.indexOf("--")===0;a==null||typeof a=="boolean"||a===""?o?e.setProperty(t,""):t==="float"?e.cssFloat="":e[t]="":o?e.setProperty(t,a):typeof a!="number"||a===0||p0.has(t)?t==="float"?e.cssFloat=a:e[t]=(""+a).trim():e[t]=a+"px"}function jv(e,t,a){if(t!=null&&typeof t!="object")throw Error(E(62));if(e=e.style,a!=null){for(var o in a)!a.hasOwnProperty(o)||t!=null&&t.hasOwnProperty(o)||(o.indexOf("--")===0?e.setProperty(o,""):o==="float"?e.cssFloat="":e[o]="");for(var l in t)o=t[l],t.hasOwnProperty(l)&&a[l]!==o&&rg(e,l,o)}else for(var n in t)t.hasOwnProperty(n)&&rg(e,n,t[n])}function Pd(e){if(e.indexOf("-")===-1)return!1;switch(e){case"annotation-xml":case"color-profile":case"font-face":case"font-face-src":case"font-face-uri":case"font-face-format":case"font-face-name":case"missing-glyph":return!1;default:return!0}}var m0=new Map([["acceptCharset","accept-charset"],["htmlFor","for"],["httpEquiv","http-equiv"],["crossOrigin","crossorigin"],["accentHeight","accent-height"],["alignmentBaseline","alignment-baseline"],["arabicForm","arabic-form"],["baselineShift","baseline-shift"],["capHeight","cap-height"],["clipPath","clip-path"],["clipRule","clip-rule"],["colorInterpolation","color-interpolation"],["colorInterpolationFilters","color-interpolation-filters"],["colorProfile","color-profile"],["colorRendering","color-rendering"],["dominantBaseline","dominant-baseline"],["enableBackground","enable-background"],["fillOpacity","fill-opacity"],["fillRule","fill-rule"],["floodColor","flood-color"],["floodOpacity","flood-opacity"],["fontFamily","font-family"],["fontSize","font-size"],["fontSizeAdjust","font-size-adjust"],["fontStretch","font-stretch"],["fontStyle","font-style"],["fontVariant","font-variant"],["fontWeight","font-weight"],["glyphName","glyph-name"],["glyphOrientationHorizontal","glyph-orientation-horizontal"],["glyphOrientationVertical","glyph-orientation-vertical"],["horizAdvX","horiz-adv-x"],["horizOriginX","horiz-origin-x"],["imageRendering","image-rendering"],["letterSpacing","letter-spacing"],["lightingColor","lighting-color"],["markerEnd","marker-end"],["markerMid","marker-mid"],["markerStart","marker-start"],["overlinePosition","overline-position"],["overlineThickness","overline-thickness"],["paintOrder","paint-order"],["panose-1","panose-1"],["pointerEvents","pointer-events"],["renderingIntent","rendering-intent"],["shapeRendering","shape-rendering"],["stopColor","stop-color"],["stopOpacity","stop-opacity"],["strikethroughPosition","strikethrough-position"],["strikethroughThickness","strikethrough-thickness"],["strokeDasharray","stroke-dasharray"],["strokeDashoffset","stroke-dashoffset"],["strokeLinecap","stroke-linecap"],["strokeLinejoin","stroke-linejoin"],["strokeMiterlimit","stroke-miterlimit"],["strokeOpacity","stroke-opacity"],["strokeWidth","stroke-width"],["textAnchor","text-anchor"],["textDecoration","text-decoration"],["textRendering","text-rendering"],["transformOrigin","transform-origin"],["underlinePosition","underline-position"],["underlineThickness","underline-thickness"],["unicodeBidi","unicode-bidi"],["unicodeRange","unicode-range"],["unitsPerEm","units-per-em"],["vAlphabetic","v-alphabetic"],["vHanging","v-hanging"],["vIdeographic","v-ideographic"],["vMathematical","v-mathematical"],["vectorEffect","vector-effect"],["vertAdvY","vert-adv-y"],["vertOriginX","vert-origin-x"],["vertOriginY","vert-origin-y"],["wordSpacing","word-spacing"],["writingMode","writing-mode"],["xmlnsXlink","xmlns:xlink"],["xHeight","x-height"]]),h0=/^[\u0000-\u001F ]*j[\r\n\t]*a[\r\n\t]*v[\r\n\t]*a[\r\n\t]*s[\r\n\t]*c[\r\n\t]*r[\r\n\t]*i[\r\n\t]*p[\r\n\t]*t[\r\n\t]*:/i;function gi(e){return h0.test(""+e)?"javascript:throw new Error('React has blocked a javascript: URL as a security precaution.')":e}function mo(){}var Vc=null;function Bd(e){return e=e.target||e.srcElement||window,e.correspondingUseElement&&(e=e.correspondingUseElement),e.nodeType===3?e.parentNode:e}var gn=null,An=null;function ig(e){var t=Vn(e);if(t&&(e=t.stateNode)){var a=e[Ht]||null;e:switch(e=t.stateNode,t.type){case"input":if(qc(e,a.value,a.defaultValue,a.defaultValue,a.checked,a.defaultChecked,a.type,a.name),t=a.name,a.type==="radio"&&t!=null){for(a=e;a.parentNode;)a=a.parentNode;for(a=a.querySelectorAll('input[name="'+da(""+t)+'"][type="radio"]'),t=0;t<a.length;t++){var o=a[t];if(o!==e&&o.form===e.form){var l=o[Ht]||null;if(!l)throw Error(E(90));qc(o,l.value,l.defaultValue,l.defaultValue,l.checked,l.defaultChecked,l.type,l.name)}}for(t=0;t<a.length;t++)o=a[t],o.form===e.form&&qv(o)}break e;case"textarea":Vv(e,a.value,a.defaultValue);break e;case"select":t=a.value,t!=null&&wn(e,!!a.multiple,t,!1)}}}var lc=!1;function Yv(e,t,a){if(lc)return e(t,a);lc=!0;try{var o=e(t);return o}finally{if(lc=!1,(gn!==null||An!==null)&&(vs(),gn&&(t=gn,e=An,An=gn=null,ig(t),e)))for(t=0;t<e.length;t++)ig(e[t])}}function ju(e,t){var a=e.stateNode;if(a===null)return null;var o=a[Ht]||null;if(o===null)return null;a=o[t];e:switch(t){case"onClick":case"onClickCapture":case"onDoubleClick":case"onDoubleClickCapture":case"onMouseDown":case"onMouseDownCapture":case"onMouseMove":case"onMouseMoveCapture":case"onMouseUp":case"onMouseUpCapture":case"onMouseEnter":(o=!o.disabled)||(e=e.type,o=!(e==="button"||e==="input"||e==="select"||e==="textarea")),e=!o;break e;default:e=!1}if(e)return null;if(a&&typeof a!="function")throw Error(E(231,t,typeof a));return a}var xo=!(typeof window=="undefined"||typeof window.document=="undefined"||typeof window.document.createElement=="undefined"),Xc=!1;if(xo)try{rn={},Object.defineProperty(rn,"passive",{get:function(){Xc=!0}}),window.addEventListener("test",rn,rn),window.removeEventListener("test",rn,rn)}catch(e){Xc=!1}var rn,Xo=null,_d=null,vi=null;function Kv(){if(vi)return vi;var e,t=_d,a=t.length,o,l="value"in Xo?Xo.value:Xo.textContent,n=l.length;for(e=0;e<a&&t[e]===l[e];e++);var u=a-e;for(o=1;o<=u&&t[a-o]===l[n-o];o++);return vi=l.slice(e,1<o?1-o:void 0)}function Si(e){var t=e.keyCode;return"charCode"in e?(e=e.charCode,e===0&&t===13&&(e=13)):e=t,e===10&&(e=13),32<=e||e===13?e:0}function ai(){return!0}function sg(){return!1}function zt(e){function t(a,o,l,n,u){this._reactName=a,this._targetInst=l,this.type=o,this.nativeEvent=n,this.target=u,this.currentTarget=null;for(var r in e)e.hasOwnProperty(r)&&(a=e[r],this[r]=a?a(n):n[r]);return this.isDefaultPrevented=(n.defaultPrevented!=null?n.defaultPrevented:n.returnValue===!1)?ai:sg,this.isPropagationStopped=sg,this}return Ie(t.prototype,{preventDefault:function(){this.defaultPrevented=!0;var a=this.nativeEvent;a&&(a.preventDefault?a.preventDefault():typeof a.returnValue!="unknown"&&(a.returnValue=!1),this.isDefaultPrevented=ai)},stopPropagation:function(){var a=this.nativeEvent;a&&(a.stopPropagation?a.stopPropagation():typeof a.cancelBubble!="unknown"&&(a.cancelBubble=!0),this.isPropagationStopped=ai)},persist:function(){},isPersistent:ai}),t}var Ul={eventPhase:0,bubbles:0,cancelable:0,timeStamp:function(e){return e.timeStamp||Date.now()},defaultPrevented:0,isTrusted:0},us=zt(Ul),fr=Ie({},Ul,{view:0,detail:0}),g0=zt(fr),nc,uc,yu,rs=Ie({},fr,{screenX:0,screenY:0,clientX:0,clientY:0,pageX:0,pageY:0,ctrlKey:0,shiftKey:0,altKey:0,metaKey:0,getModifierState:Ud,button:0,buttons:0,relatedTarget:function(e){return e.relatedTarget===void 0?e.fromElement===e.srcElement?e.toElement:e.fromElement:e.relatedTarget},movementX:function(e){return"movementX"in e?e.movementX:(e!==yu&&(yu&&e.type==="mousemove"?(nc=e.screenX-yu.screenX,uc=e.screenY-yu.screenY):uc=nc=0,yu=e),nc)},movementY:function(e){return"movementY"in e?e.movementY:uc}}),fg=zt(rs),v0=Ie({},rs,{dataTransfer:0}),S0=zt(v0),x0=Ie({},fr,{relatedTarget:0}),rc=zt(x0),L0=Ie({},Ul,{animationName:0,elapsedTime:0,pseudoElement:0}),y0=zt(L0),C0=Ie({},Ul,{clipboardData:function(e){return"clipboardData"in e?e.clipboardData:window.clipboardData}}),R0=zt(C0),I0=Ie({},Ul,{data:0}),cg=zt(I0),b0={Esc:"Escape",Spacebar:" ",Left:"ArrowLeft",Up:"ArrowUp",Right:"ArrowRight",Down:"ArrowDown",Del:"Delete",Win:"OS",Menu:"ContextMenu",Apps:"ContextMenu",Scroll:"ScrollLock",MozPrintableKey:"Unidentified"},w0={8:"Backspace",9:"Tab",12:"Clear",13:"Enter",16:"Shift",17:"Control",18:"Alt",19:"Pause",20:"CapsLock",27:"Escape",32:" ",33:"PageUp",34:"PageDown",35:"End",36:"Home",37:"ArrowLeft",38:"ArrowUp",39:"ArrowRight",40:"ArrowDown",45:"Insert",46:"Delete",112:"F1",113:"F2",114:"F3",115:"F4",116:"F5",117:"F6",118:"F7",119:"F8",120:"F9",121:"F10",122:"F11",123:"F12",144:"NumLock",145:"ScrollLock",224:"Meta"},A0={Alt:"altKey",Control:"ctrlKey",Meta:"metaKey",Shift:"shiftKey"};function T0(e){var t=this.nativeEvent;return t.getModifierState?t.getModifierState(e):(e=A0[e])?!!t[e]:!1}function Ud(){return T0}var E0=Ie({},fr,{key:function(e){if(e.key){var t=b0[e.key]||e.key;if(t!=="Unidentified")return t}return e.type==="keypress"?(e=Si(e),e===13?"Enter":String.fromCharCode(e)):e.type==="keydown"||e.type==="keyup"?w0[e.keyCode]||"Unidentified":""},code:0,location:0,ctrlKey:0,shiftKey:0,altKey:0,metaKey:0,repeat:0,locale:0,getModifierState:Ud,charCode:function(e){return e.type==="keypress"?Si(e):0},keyCode:function(e){return e.type==="keydown"||e.type==="keyup"?e.keyCode:0},which:function(e){return e.type==="keypress"?Si(e):e.type==="keydown"||e.type==="keyup"?e.keyCode:0}}),D0=zt(E0),M0=Ie({},rs,{pointerId:0,width:0,height:0,pressure:0,tangentialPressure:0,tiltX:0,tiltY:0,twist:0,pointerType:0,isPrimary:0}),dg=zt(M0),O0=Ie({},fr,{touches:0,targetTouches:0,changedTouches:0,altKey:0,metaKey:0,ctrlKey:0,shiftKey:0,getModifierState:Ud}),P0=zt(O0),B0=Ie({},Ul,{propertyName:0,elapsedTime:0,pseudoElement:0}),_0=zt(B0),U0=Ie({},rs,{deltaX:function(e){return"deltaX"in e?e.deltaX:"wheelDeltaX"in e?-e.wheelDeltaX:0},deltaY:function(e){return"deltaY"in e?e.deltaY:"wheelDeltaY"in e?-e.wheelDeltaY:"wheelDelta"in e?-e.wheelDelta:0},deltaZ:0,deltaMode:0}),N0=zt(U0),k0=Ie({},Ul,{newState:0,oldState:0}),H0=zt(k0),z0=[9,13,27,32],Nd=xo&&"CompositionEvent"in window,Ou=null;xo&&"documentMode"in document&&(Ou=document.documentMode);var F0=xo&&"TextEvent"in window&&!Ou,Zv=xo&&(!Nd||Ou&&8<Ou&&11>=Ou),pg=" ",mg=!1;function Qv(e,t){switch(e){case"keyup":return z0.indexOf(t.keyCode)!==-1;case"keydown":return t.keyCode!==229;case"keypress":case"mousedown":case"focusout":return!0;default:return!1}}function Wv(e){return e=e.detail,typeof e=="object"&&"data"in e?e.data:null}var vn=!1;function q0(e,t){switch(e){case"compositionend":return Wv(t);case"keypress":return t.which!==32?null:(mg=!0,pg);case"textInput":return e=t.data,e===pg&&mg?null:e;default:return null}}function G0(e,t){if(vn)return e==="compositionend"||!Nd&&Qv(e,t)?(e=Kv(),vi=_d=Xo=null,vn=!1,e):null;switch(e){case"paste":return null;case"keypress":if(!(t.ctrlKey||t.altKey||t.metaKey)||t.ctrlKey&&t.altKey){if(t.char&&1<t.char.length)return t.char;if(t.which)return String.fromCharCode(t.which)}return null;case"compositionend":return Zv&&t.locale!=="ko"?null:t.data;default:return null}}var V0={color:!0,date:!0,datetime:!0,"datetime-local":!0,email:!0,month:!0,number:!0,password:!0,range:!0,search:!0,tel:!0,text:!0,time:!0,url:!0,week:!0};function hg(e){var t=e&&e.nodeName&&e.nodeName.toLowerCase();return t==="input"?!!V0[e.type]:t==="textarea"}function Jv(e,t,a,o){gn?An?An.push(o):An=[o]:gn=o,t=Ji(t,"onChange"),0<t.length&&(a=new us("onChange","change",null,a,o),e.push({event:a,listeners:t}))}var Pu=null,Yu=null;function X0(e){Yx(e,0)}function is(e){var t=Eu(e);if(qv(t))return e}function gg(e,t){if(e==="change")return t}var $v=!1;xo&&(xo?(li="oninput"in document,li||(ic=document.createElement("div"),ic.setAttribute("oninput","return;"),li=typeof ic.oninput=="function"),oi=li):oi=!1,$v=oi&&(!document.documentMode||9<document.documentMode));var oi,li,ic;function vg(){Pu&&(Pu.detachEvent("onpropertychange",eS),Yu=Pu=null)}function eS(e){if(e.propertyName==="value"&&is(Yu)){var t=[];Jv(t,Yu,e,Bd(e)),Yv(X0,t)}}function j0(e,t,a){e==="focusin"?(vg(),Pu=t,Yu=a,Pu.attachEvent("onpropertychange",eS)):e==="focusout"&&vg()}function Y0(e){if(e==="selectionchange"||e==="keyup"||e==="keydown")return is(Yu)}function K0(e,t){if(e==="click")return is(t)}function Z0(e,t){if(e==="input"||e==="change")return is(t)}function Q0(e,t){return e===t&&(e!==0||1/e===1/t)||e!==e&&t!==t}var ta=typeof Object.is=="function"?Object.is:Q0;function Ku(e,t){if(ta(e,t))return!0;if(typeof e!="object"||e===null||typeof t!="object"||t===null)return!1;var a=Object.keys(e),o=Object.keys(t);if(a.length!==o.length)return!1;for(o=0;o<a.length;o++){var l=a[o];if(!Hc.call(t,l)||!ta(e[l],t[l]))return!1}return!0}function Sg(e){for(;e&&e.firstChild;)e=e.firstChild;return e}function xg(e,t){var a=Sg(e);e=0;for(var o;a;){if(a.nodeType===3){if(o=e+a.textContent.length,e<=t&&o>=t)return{node:a,offset:t-e};e=o}e:{for(;a;){if(a.nextSibling){a=a.nextSibling;break e}a=a.parentNode}a=void 0}a=Sg(a)}}function tS(e,t){return e&&t?e===t?!0:e&&e.nodeType===3?!1:t&&t.nodeType===3?tS(e,t.parentNode):"contains"in e?e.contains(t):e.compareDocumentPosition?!!(e.compareDocumentPosition(t)&16):!1:!1}function aS(e){e=e!=null&&e.ownerDocument!=null&&e.ownerDocument.defaultView!=null?e.ownerDocument.defaultView:window;for(var t=Bi(e.document);t instanceof e.HTMLIFrameElement;){try{var a=typeof t.contentWindow.location.href=="string"}catch(o){a=!1}if(a)e=t.contentWindow;else break;t=Bi(e.document)}return t}function kd(e){var t=e&&e.nodeName&&e.nodeName.toLowerCase();return t&&(t==="input"&&(e.type==="text"||e.type==="search"||e.type==="tel"||e.type==="url"||e.type==="password")||t==="textarea"||e.contentEditable==="true")}var W0=xo&&"documentMode"in document&&11>=document.documentMode,Sn=null,jc=null,Bu=null,Yc=!1;function Lg(e,t,a){var o=a.window===a?a.document:a.nodeType===9?a:a.ownerDocument;Yc||Sn==null||Sn!==Bi(o)||(o=Sn,"selectionStart"in o&&kd(o)?o={start:o.selectionStart,end:o.selectionEnd}:(o=(o.ownerDocument&&o.ownerDocument.defaultView||window).getSelection(),o={anchorNode:o.anchorNode,anchorOffset:o.anchorOffset,focusNode:o.focusNode,focusOffset:o.focusOffset}),Bu&&Ku(Bu,o)||(Bu=o,o=Ji(jc,"onSelect"),0<o.length&&(t=new us("onSelect","select",null,t,a),e.push({event:t,listeners:o}),t.target=Sn)))}function Ll(e,t){var a={};return a[e.toLowerCase()]=t.toLowerCase(),a["Webkit"+e]="webkit"+t,a["Moz"+e]="moz"+t,a}var xn={animationend:Ll("Animation","AnimationEnd"),animationiteration:Ll("Animation","AnimationIteration"),animationstart:Ll("Animation","AnimationStart"),transitionrun:Ll("Transition","TransitionRun"),transitionstart:Ll("Transition","TransitionStart"),transitioncancel:Ll("Transition","TransitionCancel"),transitionend:Ll("Transition","TransitionEnd")},sc={},oS={};xo&&(oS=document.createElement("div").style,"AnimationEvent"in window||(delete xn.animationend.animation,delete xn.animationiteration.animation,delete xn.animationstart.animation),"TransitionEvent"in window||delete xn.transitionend.transition);function Nl(e){if(sc[e])return sc[e];if(!xn[e])return e;var t=xn[e],a;for(a in t)if(t.hasOwnProperty(a)&&a in oS)return sc[e]=t[a];return e}var lS=Nl("animationend"),nS=Nl("animationiteration"),uS=Nl("animationstart"),J0=Nl("transitionrun"),$0=Nl("transitionstart"),ew=Nl("transitioncancel"),rS=Nl("transitionend"),iS=new Map,Kc="abort auxClick beforeToggle cancel canPlay canPlayThrough click close contextMenu copy cut drag dragEnd dragEnter dragExit dragLeave dragOver dragStart drop durationChange emptied encrypted ended error gotPointerCapture input invalid keyDown keyPress keyUp load loadedData loadedMetadata loadStart lostPointerCapture mouseDown mouseMove mouseOut mouseOver mouseUp paste pause play playing pointerCancel pointerDown pointerMove pointerOut pointerOver pointerUp progress rateChange reset resize seeked seeking stalled submit suspend timeUpdate touchCancel touchEnd touchStart volumeChange scroll toggle touchMove waiting wheel".split(" ");Kc.push("scrollEnd");function Ea(e,t){iS.set(e,t),_l(t,[e])}var _i=typeof reportError=="function"?reportError:function(e){if(typeof window=="object"&&typeof window.ErrorEvent=="function"){var t=new window.ErrorEvent("error",{bubbles:!0,cancelable:!0,message:typeof e=="object"&&e!==null&&typeof e.message=="string"?String(e.message):String(e),error:e});if(!window.dispatchEvent(t))return}else if(typeof process=="object"&&typeof process.emit=="function"){process.emit("uncaughtException",e);return}console.error(e)},ia=[],Ln=0,Hd=0;function ss(){for(var e=Ln,t=Hd=Ln=0;t<e;){var a=ia[t];ia[t++]=null;var o=ia[t];ia[t++]=null;var l=ia[t];ia[t++]=null;var n=ia[t];if(ia[t++]=null,o!==null&&l!==null){var u=o.pending;u===null?l.next=l:(l.next=u.next,u.next=l),o.pending=l}n!==0&&sS(a,l,n)}}function fs(e,t,a,o){ia[Ln++]=e,ia[Ln++]=t,ia[Ln++]=a,ia[Ln++]=o,Hd|=o,e.lanes|=o,e=e.alternate,e!==null&&(e.lanes|=o)}function zd(e,t,a,o){return fs(e,t,a,o),Ui(e)}function kl(e,t){return fs(e,null,null,t),Ui(e)}function sS(e,t,a){e.lanes|=a;var o=e.alternate;o!==null&&(o.lanes|=a);for(var l=!1,n=e.return;n!==null;)n.childLanes|=a,o=n.alternate,o!==null&&(o.childLanes|=a),n.tag===22&&(e=n.stateNode,e===null||e._visibility&1||(l=!0)),e=n,n=n.return;return e.tag===3?(n=e.stateNode,l&&t!==null&&(l=31-$t(a),e=n.hiddenUpdates,o=e[l],o===null?e[l]=[t]:o.push(t),t.lane=a|536870912),n):null}function Ui(e){if(50<Gu)throw Gu=0,hd=null,Error(E(185));for(var t=e.return;t!==null;)e=t,t=e.return;return e.tag===3?e.stateNode:null}var yn={};function tw(e,t,a,o){this.tag=e,this.key=a,this.sibling=this.child=this.return=this.stateNode=this.type=this.elementType=null,this.index=0,this.refCleanup=this.ref=null,this.pendingProps=t,this.dependencies=this.memoizedState=this.updateQueue=this.memoizedProps=null,this.mode=o,this.subtreeFlags=this.flags=0,this.deletions=null,this.childLanes=this.lanes=0,this.alternate=null}function Zt(e,t,a,o){return new tw(e,t,a,o)}function Fd(e){return e=e.prototype,!(!e||!e.isReactComponent)}function go(e,t){var a=e.alternate;return a===null?(a=Zt(e.tag,t,e.key,e.mode),a.elementType=e.elementType,a.type=e.type,a.stateNode=e.stateNode,a.alternate=e,e.alternate=a):(a.pendingProps=t,a.type=e.type,a.flags=0,a.subtreeFlags=0,a.deletions=null),a.flags=e.flags&65011712,a.childLanes=e.childLanes,a.lanes=e.lanes,a.child=e.child,a.memoizedProps=e.memoizedProps,a.memoizedState=e.memoizedState,a.updateQueue=e.updateQueue,t=e.dependencies,a.dependencies=t===null?null:{lanes:t.lanes,firstContext:t.firstContext},a.sibling=e.sibling,a.index=e.index,a.ref=e.ref,a.refCleanup=e.refCleanup,a}function fS(e,t){e.flags&=65011714;var a=e.alternate;return a===null?(e.childLanes=0,e.lanes=t,e.child=null,e.subtreeFlags=0,e.memoizedProps=null,e.memoizedState=null,e.updateQueue=null,e.dependencies=null,e.stateNode=null):(e.childLanes=a.childLanes,e.lanes=a.lanes,e.child=a.child,e.subtreeFlags=0,e.deletions=null,e.memoizedProps=a.memoizedProps,e.memoizedState=a.memoizedState,e.updateQueue=a.updateQueue,e.type=a.type,t=a.dependencies,e.dependencies=t===null?null:{lanes:t.lanes,firstContext:t.firstContext}),e}function xi(e,t,a,o,l,n){var u=0;if(o=e,typeof e=="function")Fd(e)&&(u=1);else if(typeof e=="string")u=lA(e,a,Ka.current)?26:e==="html"||e==="head"||e==="body"?27:5;else e:switch(e){case _c:return e=Zt(31,a,t,l),e.elementType=_c,e.lanes=n,e;case pn:return wl(a.children,l,n,t);case Ev:u=8,l|=24;break;case Oc:return e=Zt(12,a,t,l|2),e.elementType=Oc,e.lanes=n,e;case Pc:return e=Zt(13,a,t,l),e.elementType=Pc,e.lanes=n,e;case Bc:return e=Zt(19,a,t,l),e.elementType=Bc,e.lanes=n,e;default:if(typeof e=="object"&&e!==null)switch(e.$$typeof){case po:u=10;break e;case Dv:u=9;break e;case Ad:u=11;break e;case Td:u=14;break e;case ko:u=16,o=null;break e}u=29,a=Error(E(130,e===null?"null":typeof e,"")),o=null}return t=Zt(u,a,t,l),t.elementType=e,t.type=o,t.lanes=n,t}function wl(e,t,a,o){return e=Zt(7,e,o,t),e.lanes=a,e}function fc(e,t,a){return e=Zt(6,e,null,t),e.lanes=a,e}function cS(e){var t=Zt(18,null,null,0);return t.stateNode=e,t}function cc(e,t,a){return t=Zt(4,e.children!==null?e.children:[],e.key,t),t.lanes=a,t.stateNode={containerInfo:e.containerInfo,pendingChildren:null,implementation:e.implementation},t}var yg=new WeakMap;function pa(e,t){if(typeof e=="object"&&e!==null){var a=yg.get(e);return a!==void 0?a:(t={value:e,source:t,stack:ag(t)},yg.set(e,t),t)}return{value:e,source:t,stack:ag(t)}}var Cn=[],Rn=0,Ni=null,Zu=0,fa=[],ca=0,ll=null,Xa=1,ja="";function fo(e,t){Cn[Rn++]=Zu,Cn[Rn++]=Ni,Ni=e,Zu=t}function dS(e,t,a){fa[ca++]=Xa,fa[ca++]=ja,fa[ca++]=ll,ll=e;var o=Xa;e=ja;var l=32-$t(o)-1;o&=~(1<<l),a+=1;var n=32-$t(t)+l;if(30<n){var u=l-l%5;n=(o&(1<<u)-1).toString(32),o>>=u,l-=u,Xa=1<<32-$t(t)+l|a<<l|o,ja=n+e}else Xa=1<<n|a<<l|o,ja=e}function qd(e){e.return!==null&&(fo(e,1),dS(e,1,0))}function Gd(e){for(;e===Ni;)Ni=Cn[--Rn],Cn[Rn]=null,Zu=Cn[--Rn],Cn[Rn]=null;for(;e===ll;)ll=fa[--ca],fa[ca]=null,ja=fa[--ca],fa[ca]=null,Xa=fa[--ca],fa[ca]=null}function pS(e,t){fa[ca++]=Xa,fa[ca++]=ja,fa[ca++]=ll,Xa=t.id,ja=t.overflow,ll=e}var ht=null,Re=null,oe=!1,Qo=null,ma=!1,Zc=Error(E(519));function nl(e){var t=Error(E(418,1<arguments.length&&arguments[1]!==void 0&&arguments[1]?"text":"HTML",""));throw Qu(pa(t,e)),Zc}function Cg(e){var t=e.stateNode,a=e.type,o=e.memoizedProps;switch(t[mt]=e,t[Ht]=o,a){case"dialog":ee("cancel",t),ee("close",t);break;case"iframe":case"object":case"embed":ee("load",t);break;case"video":case"audio":for(a=0;a<er.length;a++)ee(er[a],t);break;case"source":ee("error",t);break;case"img":case"image":case"link":ee("error",t),ee("load",t);break;case"details":ee("toggle",t);break;case"input":ee("invalid",t),Gv(t,o.value,o.defaultValue,o.checked,o.defaultChecked,o.type,o.name,!0);break;case"select":ee("invalid",t);break;case"textarea":ee("invalid",t),Xv(t,o.value,o.defaultValue,o.children)}a=o.children,typeof a!="string"&&typeof a!="number"&&typeof a!="bigint"||t.textContent===""+a||o.suppressHydrationWarning===!0||Zx(t.textContent,a)?(o.popover!=null&&(ee("beforetoggle",t),ee("toggle",t)),o.onScroll!=null&&ee("scroll",t),o.onScrollEnd!=null&&ee("scrollend",t),o.onClick!=null&&(t.onclick=mo),t=!0):t=!1,t||nl(e,!0)}function Rg(e){for(ht=e.return;ht;)switch(ht.tag){case 5:case 31:case 13:ma=!1;return;case 27:case 3:ma=!0;return;default:ht=ht.return}}function sn(e){if(e!==ht)return!1;if(!oe)return Rg(e),oe=!0,!1;var t=e.tag,a;if((a=t!==3&&t!==27)&&((a=t===5)&&(a=e.type,a=!(a!=="form"&&a!=="button")||Ld(e.type,e.memoizedProps)),a=!a),a&&Re&&nl(e),Rg(e),t===13){if(e=e.memoizedState,e=e!==null?e.dehydrated:null,!e)throw Error(E(317));Re=cv(e)}else if(t===31){if(e=e.memoizedState,e=e!==null?e.dehydrated:null,!e)throw Error(E(317));Re=cv(e)}else t===27?(t=Re,sl(e.type)?(e=Id,Id=null,Re=e):Re=t):Re=ht?ga(e.stateNode.nextSibling):null;return!0}function Dl(){Re=ht=null,oe=!1}function dc(){var e=Qo;return e!==null&&(Nt===null?Nt=e:Nt.push.apply(Nt,e),Qo=null),e}function Qu(e){Qo===null?Qo=[e]:Qo.push(e)}var Qc=Za(null),Hl=null,ho=null;function zo(e,t,a){xe(Qc,t._currentValue),t._currentValue=a}function vo(e){e._currentValue=Qc.current,nt(Qc)}function Wc(e,t,a){for(;e!==null;){var o=e.alternate;if((e.childLanes&t)!==t?(e.childLanes|=t,o!==null&&(o.childLanes|=t)):o!==null&&(o.childLanes&t)!==t&&(o.childLanes|=t),e===a)break;e=e.return}}function Jc(e,t,a,o){var l=e.child;for(l!==null&&(l.return=e);l!==null;){var n=l.dependencies;if(n!==null){var u=l.child;n=n.firstContext;e:for(;n!==null;){var r=n;n=l;for(var i=0;i<t.length;i++)if(r.context===t[i]){n.lanes|=a,r=n.alternate,r!==null&&(r.lanes|=a),Wc(n.return,a,e),o||(u=null);break e}n=r.next}}else if(l.tag===18){if(u=l.return,u===null)throw Error(E(341));u.lanes|=a,n=u.alternate,n!==null&&(n.lanes|=a),Wc(u,a,e),u=null}else u=l.child;if(u!==null)u.return=l;else for(u=l;u!==null;){if(u===e){u=null;break}if(l=u.sibling,l!==null){l.return=u.return,u=l;break}u=u.return}l=u}}function Xn(e,t,a,o){e=null;for(var l=t,n=!1;l!==null;){if(!n){if((l.flags&524288)!==0)n=!0;else if((l.flags&262144)!==0)break}if(l.tag===10){var u=l.alternate;if(u===null)throw Error(E(387));if(u=u.memoizedProps,u!==null){var r=l.type;ta(l.pendingProps.value,u.value)||(e!==null?e.push(r):e=[r])}}else if(l===Di.current){if(u=l.alternate,u===null)throw Error(E(387));u.memoizedState.memoizedState!==l.memoizedState.memoizedState&&(e!==null?e.push(ar):e=[ar])}l=l.return}e!==null&&Jc(t,e,a,o),t.flags|=262144}function ki(e){for(e=e.firstContext;e!==null;){if(!ta(e.context._currentValue,e.memoizedValue))return!0;e=e.next}return!1}function Ml(e){Hl=e,ho=null,e=e.dependencies,e!==null&&(e.firstContext=null)}function gt(e){return mS(Hl,e)}function ni(e,t){return Hl===null&&Ml(e),mS(e,t)}function mS(e,t){var a=t._currentValue;if(t={context:t,memoizedValue:a,next:null},ho===null){if(e===null)throw Error(E(308));ho=t,e.dependencies={lanes:0,firstContext:t},e.flags|=524288}else ho=ho.next=t;return a}var aw=typeof AbortController!="undefined"?AbortController:function(){var e=[],t=this.signal={aborted:!1,addEventListener:function(a,o){e.push(o)}};this.abort=function(){t.aborted=!0,e.forEach(function(a){return a()})}},ow=Ze.unstable_scheduleCallback,lw=Ze.unstable_NormalPriority,Fe={$$typeof:po,Consumer:null,Provider:null,_currentValue:null,_currentValue2:null,_threadCount:0};function Vd(){return{controller:new aw,data:new Map,refCount:0}}function cr(e){e.refCount--,e.refCount===0&&ow(lw,function(){e.controller.abort()})}var _u=null,$c=0,_n=0,Tn=null;function nw(e,t){if(_u===null){var a=_u=[];$c=0,_n=hp(),Tn={status:"pending",value:void 0,then:function(o){a.push(o)}}}return $c++,t.then(Ig,Ig),t}function Ig(){if(--$c===0&&_u!==null){Tn!==null&&(Tn.status="fulfilled");var e=_u;_u=null,_n=0,Tn=null;for(var t=0;t<e.length;t++)(0,e[t])()}}function uw(e,t){var a=[],o={status:"pending",value:null,reason:null,then:function(l){a.push(l)}};return e.then(function(){o.status="fulfilled",o.value=t;for(var l=0;l<a.length;l++)(0,a[l])(t)},function(l){for(o.status="rejected",o.reason=l,l=0;l<a.length;l++)(0,a[l])(void 0)}),o}var bg=V.S;V.S=function(e,t){Tx=Wt(),typeof t=="object"&&t!==null&&typeof t.then=="function"&&nw(e,t),bg!==null&&bg(e,t)};var Al=Za(null);function Xd(){var e=Al.current;return e!==null?e:Se.pooledCache}function Li(e,t){t===null?xe(Al,Al.current):xe(Al,t.pool)}function hS(){var e=Xd();return e===null?null:{parent:Fe._currentValue,pool:e}}var jn=Error(E(460)),jd=Error(E(474)),cs=Error(E(542)),Hi={then:function(){}};function wg(e){return e=e.status,e==="fulfilled"||e==="rejected"}function gS(e,t,a){switch(a=e[a],a===void 0?e.push(t):a!==t&&(t.then(mo,mo),t=a),t.status){case"fulfilled":return t.value;case"rejected":throw e=t.reason,Tg(e),e;default:if(typeof t.status=="string")t.then(mo,mo);else{if(e=Se,e!==null&&100<e.shellSuspendCounter)throw Error(E(482));e=t,e.status="pending",e.then(function(o){if(t.status==="pending"){var l=t;l.status="fulfilled",l.value=o}},function(o){if(t.status==="pending"){var l=t;l.status="rejected",l.reason=o}})}switch(t.status){case"fulfilled":return t.value;case"rejected":throw e=t.reason,Tg(e),e}throw Tl=t,jn}}function Rl(e){try{var t=e._init;return t(e._payload)}catch(a){throw a!==null&&typeof a=="object"&&typeof a.then=="function"?(Tl=a,jn):a}}var Tl=null;function Ag(){if(Tl===null)throw Error(E(459));var e=Tl;return Tl=null,e}function Tg(e){if(e===jn||e===cs)throw Error(E(483))}var En=null,Wu=0;function ui(e){var t=Wu;return Wu+=1,En===null&&(En=[]),gS(En,e,t)}function Cu(e,t){t=t.props.ref,e.ref=t!==void 0?t:null}function ri(e,t){throw t.$$typeof===jb?Error(E(525)):(e=Object.prototype.toString.call(t),Error(E(31,e==="[object Object]"?"object with keys {"+Object.keys(t).join(", ")+"}":e)))}function vS(e){function t(h,p){if(e){var m=h.deletions;m===null?(h.deletions=[p],h.flags|=16):m.push(p)}}function a(h,p){if(!e)return null;for(;p!==null;)t(h,p),p=p.sibling;return null}function o(h){for(var p=new Map;h!==null;)h.key!==null?p.set(h.key,h):p.set(h.index,h),h=h.sibling;return p}function l(h,p){return h=go(h,p),h.index=0,h.sibling=null,h}function n(h,p,m){return h.index=m,e?(m=h.alternate,m!==null?(m=m.index,m<p?(h.flags|=67108866,p):m):(h.flags|=67108866,p)):(h.flags|=1048576,p)}function u(h){return e&&h.alternate===null&&(h.flags|=67108866),h}function r(h,p,m,v){return p===null||p.tag!==6?(p=fc(m,h.mode,v),p.return=h,p):(p=l(p,m),p.return=h,p)}function i(h,p,m,v){var C=m.type;return C===pn?f(h,p,m.props.children,v,m.key):p!==null&&(p.elementType===C||typeof C=="object"&&C!==null&&C.$$typeof===ko&&Rl(C)===p.type)?(p=l(p,m.props),Cu(p,m),p.return=h,p):(p=xi(m.type,m.key,m.props,null,h.mode,v),Cu(p,m),p.return=h,p)}function s(h,p,m,v){return p===null||p.tag!==4||p.stateNode.containerInfo!==m.containerInfo||p.stateNode.implementation!==m.implementation?(p=cc(m,h.mode,v),p.return=h,p):(p=l(p,m.children||[]),p.return=h,p)}function f(h,p,m,v,C){return p===null||p.tag!==7?(p=wl(m,h.mode,v,C),p.return=h,p):(p=l(p,m),p.return=h,p)}function d(h,p,m){if(typeof p=="string"&&p!==""||typeof p=="number"||typeof p=="bigint")return p=fc(""+p,h.mode,m),p.return=h,p;if(typeof p=="object"&&p!==null){switch(p.$$typeof){case Wr:return m=xi(p.type,p.key,p.props,null,h.mode,m),Cu(m,p),m.return=h,m;case Au:return p=cc(p,h.mode,m),p.return=h,p;case ko:return p=Rl(p),d(h,p,m)}if(Tu(p)||Lu(p))return p=wl(p,h.mode,m,null),p.return=h,p;if(typeof p.then=="function")return d(h,ui(p),m);if(p.$$typeof===po)return d(h,ni(h,p),m);ri(h,p)}return null}function c(h,p,m,v){var C=p!==null?p.key:null;if(typeof m=="string"&&m!==""||typeof m=="number"||typeof m=="bigint")return C!==null?null:r(h,p,""+m,v);if(typeof m=="object"&&m!==null){switch(m.$$typeof){case Wr:return m.key===C?i(h,p,m,v):null;case Au:return m.key===C?s(h,p,m,v):null;case ko:return m=Rl(m),c(h,p,m,v)}if(Tu(m)||Lu(m))return C!==null?null:f(h,p,m,v,null);if(typeof m.then=="function")return c(h,p,ui(m),v);if(m.$$typeof===po)return c(h,p,ni(h,m),v);ri(h,m)}return null}function g(h,p,m,v,C){if(typeof v=="string"&&v!==""||typeof v=="number"||typeof v=="bigint")return h=h.get(m)||null,r(p,h,""+v,C);if(typeof v=="object"&&v!==null){switch(v.$$typeof){case Wr:return h=h.get(v.key===null?m:v.key)||null,i(p,h,v,C);case Au:return h=h.get(v.key===null?m:v.key)||null,s(p,h,v,C);case ko:return v=Rl(v),g(h,p,m,v,C)}if(Tu(v)||Lu(v))return h=h.get(m)||null,f(p,h,v,C,null);if(typeof v.then=="function")return g(h,p,m,ui(v),C);if(v.$$typeof===po)return g(h,p,m,ni(p,v),C);ri(p,v)}return null}function L(h,p,m,v){for(var C=null,I=null,R=p,w=p=0,A=null;R!==null&&w<m.length;w++){R.index>w?(A=R,R=null):A=R.sibling;var M=c(h,R,m[w],v);if(M===null){R===null&&(R=A);break}e&&R&&M.alternate===null&&t(h,R),p=n(M,p,w),I===null?C=M:I.sibling=M,I=M,R=A}if(w===m.length)return a(h,R),oe&&fo(h,w),C;if(R===null){for(;w<m.length;w++)R=d(h,m[w],v),R!==null&&(p=n(R,p,w),I===null?C=R:I.sibling=R,I=R);return oe&&fo(h,w),C}for(R=o(R);w<m.length;w++)A=g(R,h,w,m[w],v),A!==null&&(e&&A.alternate!==null&&R.delete(A.key===null?w:A.key),p=n(A,p,w),I===null?C=A:I.sibling=A,I=A);return e&&R.forEach(function(U){return t(h,U)}),oe&&fo(h,w),C}function x(h,p,m,v){if(m==null)throw Error(E(151));for(var C=null,I=null,R=p,w=p=0,A=null,M=m.next();R!==null&&!M.done;w++,M=m.next()){R.index>w?(A=R,R=null):A=R.sibling;var U=c(h,R,M.value,v);if(U===null){R===null&&(R=A);break}e&&R&&U.alternate===null&&t(h,R),p=n(U,p,w),I===null?C=U:I.sibling=U,I=U,R=A}if(M.done)return a(h,R),oe&&fo(h,w),C;if(R===null){for(;!M.done;w++,M=m.next())M=d(h,M.value,v),M!==null&&(p=n(M,p,w),I===null?C=M:I.sibling=M,I=M);return oe&&fo(h,w),C}for(R=o(R);!M.done;w++,M=m.next())M=g(R,h,w,M.value,v),M!==null&&(e&&M.alternate!==null&&R.delete(M.key===null?w:M.key),p=n(M,p,w),I===null?C=M:I.sibling=M,I=M);return e&&R.forEach(function(N){return t(h,N)}),oe&&fo(h,w),C}function S(h,p,m,v){if(typeof m=="object"&&m!==null&&m.type===pn&&m.key===null&&(m=m.props.children),typeof m=="object"&&m!==null){switch(m.$$typeof){case Wr:e:{for(var C=m.key;p!==null;){if(p.key===C){if(C=m.type,C===pn){if(p.tag===7){a(h,p.sibling),v=l(p,m.props.children),v.return=h,h=v;break e}}else if(p.elementType===C||typeof C=="object"&&C!==null&&C.$$typeof===ko&&Rl(C)===p.type){a(h,p.sibling),v=l(p,m.props),Cu(v,m),v.return=h,h=v;break e}a(h,p);break}else t(h,p);p=p.sibling}m.type===pn?(v=wl(m.props.children,h.mode,v,m.key),v.return=h,h=v):(v=xi(m.type,m.key,m.props,null,h.mode,v),Cu(v,m),v.return=h,h=v)}return u(h);case Au:e:{for(C=m.key;p!==null;){if(p.key===C)if(p.tag===4&&p.stateNode.containerInfo===m.containerInfo&&p.stateNode.implementation===m.implementation){a(h,p.sibling),v=l(p,m.children||[]),v.return=h,h=v;break e}else{a(h,p);break}else t(h,p);p=p.sibling}v=cc(m,h.mode,v),v.return=h,h=v}return u(h);case ko:return m=Rl(m),S(h,p,m,v)}if(Tu(m))return L(h,p,m,v);if(Lu(m)){if(C=Lu(m),typeof C!="function")throw Error(E(150));return m=C.call(m),x(h,p,m,v)}if(typeof m.then=="function")return S(h,p,ui(m),v);if(m.$$typeof===po)return S(h,p,ni(h,m),v);ri(h,m)}return typeof m=="string"&&m!==""||typeof m=="number"||typeof m=="bigint"?(m=""+m,p!==null&&p.tag===6?(a(h,p.sibling),v=l(p,m),v.return=h,h=v):(a(h,p),v=fc(m,h.mode,v),v.return=h,h=v),u(h)):a(h,p)}return function(h,p,m,v){try{Wu=0;var C=S(h,p,m,v);return En=null,C}catch(R){if(R===jn||R===cs)throw R;var I=Zt(29,R,null,h.mode);return I.lanes=v,I.return=h,I}}}var Ol=vS(!0),SS=vS(!1),Ho=!1;function Yd(e){e.updateQueue={baseState:e.memoizedState,firstBaseUpdate:null,lastBaseUpdate:null,shared:{pending:null,lanes:0,hiddenCallbacks:null},callbacks:null}}function ed(e,t){e=e.updateQueue,t.updateQueue===e&&(t.updateQueue={baseState:e.baseState,firstBaseUpdate:e.firstBaseUpdate,lastBaseUpdate:e.lastBaseUpdate,shared:e.shared,callbacks:null})}function Wo(e){return{lane:e,tag:0,payload:null,callback:null,next:null}}function Jo(e,t,a){var o=e.updateQueue;if(o===null)return null;if(o=o.shared,(re&2)!==0){var l=o.pending;return l===null?t.next=t:(t.next=l.next,l.next=t),o.pending=t,t=Ui(e),sS(e,null,a),t}return fs(e,o,t,a),Ui(e)}function Uu(e,t,a){if(t=t.updateQueue,t!==null&&(t=t.shared,(a&4194048)!==0)){var o=t.lanes;o&=e.pendingLanes,a|=o,t.lanes=a,Uv(e,a)}}function pc(e,t){var a=e.updateQueue,o=e.alternate;if(o!==null&&(o=o.updateQueue,a===o)){var l=null,n=null;if(a=a.firstBaseUpdate,a!==null){do{var u={lane:a.lane,tag:a.tag,payload:a.payload,callback:null,next:null};n===null?l=n=u:n=n.next=u,a=a.next}while(a!==null);n===null?l=n=t:n=n.next=t}else l=n=t;a={baseState:o.baseState,firstBaseUpdate:l,lastBaseUpdate:n,shared:o.shared,callbacks:o.callbacks},e.updateQueue=a;return}e=a.lastBaseUpdate,e===null?a.firstBaseUpdate=t:e.next=t,a.lastBaseUpdate=t}var td=!1;function Nu(){if(td){var e=Tn;if(e!==null)throw e}}function ku(e,t,a,o){td=!1;var l=e.updateQueue;Ho=!1;var n=l.firstBaseUpdate,u=l.lastBaseUpdate,r=l.shared.pending;if(r!==null){l.shared.pending=null;var i=r,s=i.next;i.next=null,u===null?n=s:u.next=s,u=i;var f=e.alternate;f!==null&&(f=f.updateQueue,r=f.lastBaseUpdate,r!==u&&(r===null?f.firstBaseUpdate=s:r.next=s,f.lastBaseUpdate=i))}if(n!==null){var d=l.baseState;u=0,f=s=i=null,r=n;do{var c=r.lane&-536870913,g=c!==r.lane;if(g?(ae&c)===c:(o&c)===c){c!==0&&c===_n&&(td=!0),f!==null&&(f=f.next={lane:0,tag:r.tag,payload:r.payload,callback:null,next:null});e:{var L=e,x=r;c=t;var S=a;switch(x.tag){case 1:if(L=x.payload,typeof L=="function"){d=L.call(S,d,c);break e}d=L;break e;case 3:L.flags=L.flags&-65537|128;case 0:if(L=x.payload,c=typeof L=="function"?L.call(S,d,c):L,c==null)break e;d=Ie({},d,c);break e;case 2:Ho=!0}}c=r.callback,c!==null&&(e.flags|=64,g&&(e.flags|=8192),g=l.callbacks,g===null?l.callbacks=[c]:g.push(c))}else g={lane:c,tag:r.tag,payload:r.payload,callback:r.callback,next:null},f===null?(s=f=g,i=d):f=f.next=g,u|=c;if(r=r.next,r===null){if(r=l.shared.pending,r===null)break;g=r,r=g.next,g.next=null,l.lastBaseUpdate=g,l.shared.pending=null}}while(!0);f===null&&(i=d),l.baseState=i,l.firstBaseUpdate=s,l.lastBaseUpdate=f,n===null&&(l.shared.lanes=0),rl|=u,e.lanes=u,e.memoizedState=d}}function xS(e,t){if(typeof e!="function")throw Error(E(191,e));e.call(t)}function LS(e,t){var a=e.callbacks;if(a!==null)for(e.callbacks=null,e=0;e<a.length;e++)xS(a[e],t)}var Un=Za(null),zi=Za(0);function Eg(e,t){e=Ro,xe(zi,e),xe(Un,t),Ro=e|t.baseLanes}function ad(){xe(zi,Ro),xe(Un,Un.current)}function Kd(){Ro=zi.current,nt(Un),nt(zi)}var aa=Za(null),ha=null;function Fo(e){var t=e.alternate;xe(Ne,Ne.current&1),xe(aa,e),ha===null&&(t===null||Un.current!==null||t.memoizedState!==null)&&(ha=e)}function od(e){xe(Ne,Ne.current),xe(aa,e),ha===null&&(ha=e)}function yS(e){e.tag===22?(xe(Ne,Ne.current),xe(aa,e),ha===null&&(ha=e)):qo(e)}function qo(){xe(Ne,Ne.current),xe(aa,aa.current)}function Kt(e){nt(aa),ha===e&&(ha=null),nt(Ne)}var Ne=Za(0);function Fi(e){for(var t=e;t!==null;){if(t.tag===13){var a=t.memoizedState;if(a!==null&&(a=a.dehydrated,a===null||Cd(a)||Rd(a)))return t}else if(t.tag===19&&(t.memoizedProps.revealOrder==="forwards"||t.memoizedProps.revealOrder==="backwards"||t.memoizedProps.revealOrder==="unstable_legacy-backwards"||t.memoizedProps.revealOrder==="together")){if((t.flags&128)!==0)return t}else if(t.child!==null){t.child.return=t,t=t.child;continue}if(t===e)break;for(;t.sibling===null;){if(t.return===null||t.return===e)return null;t=t.return}t.sibling.return=t.return,t=t.sibling}return null}var Lo=0,K=null,pe=null,He=null,qi=!1,Dn=!1,Pl=!1,Gi=0,Ju=0,Mn=null,rw=0;function Oe(){throw Error(E(321))}function Zd(e,t){if(t===null)return!1;for(var a=0;a<t.length&&a<e.length;a++)if(!ta(e[a],t[a]))return!1;return!0}function Qd(e,t,a,o,l,n){return Lo=n,K=t,t.memoizedState=null,t.updateQueue=null,t.lanes=0,V.H=e===null||e.memoizedState===null?JS:rp,Pl=!1,n=a(o,l),Pl=!1,Dn&&(n=RS(t,a,o,l)),CS(e),n}function CS(e){V.H=$u;var t=pe!==null&&pe.next!==null;if(Lo=0,He=pe=K=null,qi=!1,Ju=0,Mn=null,t)throw Error(E(300));e===null||qe||(e=e.dependencies,e!==null&&ki(e)&&(qe=!0))}function RS(e,t,a,o){K=e;var l=0;do{if(Dn&&(Mn=null),Ju=0,Dn=!1,25<=l)throw Error(E(301));if(l+=1,He=pe=null,e.updateQueue!=null){var n=e.updateQueue;n.lastEffect=null,n.events=null,n.stores=null,n.memoCache!=null&&(n.memoCache.index=0)}V.H=$S,n=t(a,o)}while(Dn);return n}function iw(){var e=V.H,t=e.useState()[0];return t=typeof t.then=="function"?dr(t):t,e=e.useState()[0],(pe!==null?pe.memoizedState:null)!==e&&(K.flags|=1024),t}function Wd(){var e=Gi!==0;return Gi=0,e}function Jd(e,t,a){t.updateQueue=e.updateQueue,t.flags&=-2053,e.lanes&=~a}function $d(e){if(qi){for(e=e.memoizedState;e!==null;){var t=e.queue;t!==null&&(t.pending=null),e=e.next}qi=!1}Lo=0,He=pe=K=null,Dn=!1,Ju=Gi=0,Mn=null}function Tt(){var e={memoizedState:null,baseState:null,baseQueue:null,queue:null,next:null};return He===null?K.memoizedState=He=e:He=He.next=e,He}function ke(){if(pe===null){var e=K.alternate;e=e!==null?e.memoizedState:null}else e=pe.next;var t=He===null?K.memoizedState:He.next;if(t!==null)He=t,pe=e;else{if(e===null)throw K.alternate===null?Error(E(467)):Error(E(310));pe=e,e={memoizedState:pe.memoizedState,baseState:pe.baseState,baseQueue:pe.baseQueue,queue:pe.queue,next:null},He===null?K.memoizedState=He=e:He=He.next=e}return He}function ds(){return{lastEffect:null,events:null,stores:null,memoCache:null}}function dr(e){var t=Ju;return Ju+=1,Mn===null&&(Mn=[]),e=gS(Mn,e,t),t=K,(He===null?t.memoizedState:He.next)===null&&(t=t.alternate,V.H=t===null||t.memoizedState===null?JS:rp),e}function ps(e){if(e!==null&&typeof e=="object"){if(typeof e.then=="function")return dr(e);if(e.$$typeof===po)return gt(e)}throw Error(E(438,String(e)))}function ep(e){var t=null,a=K.updateQueue;if(a!==null&&(t=a.memoCache),t==null){var o=K.alternate;o!==null&&(o=o.updateQueue,o!==null&&(o=o.memoCache,o!=null&&(t={data:o.data.map(function(l){return l.slice()}),index:0})))}if(t==null&&(t={data:[],index:0}),a===null&&(a=ds(),K.updateQueue=a),a.memoCache=t,a=t.data[t.index],a===void 0)for(a=t.data[t.index]=Array(e),o=0;o<e;o++)a[o]=Yb;return t.index++,a}function yo(e,t){return typeof t=="function"?t(e):t}function yi(e){var t=ke();return tp(t,pe,e)}function tp(e,t,a){var o=e.queue;if(o===null)throw Error(E(311));o.lastRenderedReducer=a;var l=e.baseQueue,n=o.pending;if(n!==null){if(l!==null){var u=l.next;l.next=n.next,n.next=u}t.baseQueue=l=n,o.pending=null}if(n=e.baseState,l===null)e.memoizedState=n;else{t=l.next;var r=u=null,i=null,s=t,f=!1;do{var d=s.lane&-536870913;if(d!==s.lane?(ae&d)===d:(Lo&d)===d){var c=s.revertLane;if(c===0)i!==null&&(i=i.next={lane:0,revertLane:0,gesture:null,action:s.action,hasEagerState:s.hasEagerState,eagerState:s.eagerState,next:null}),d===_n&&(f=!0);else if((Lo&c)===c){s=s.next,c===_n&&(f=!0);continue}else d={lane:0,revertLane:s.revertLane,gesture:null,action:s.action,hasEagerState:s.hasEagerState,eagerState:s.eagerState,next:null},i===null?(r=i=d,u=n):i=i.next=d,K.lanes|=c,rl|=c;d=s.action,Pl&&a(n,d),n=s.hasEagerState?s.eagerState:a(n,d)}else c={lane:d,revertLane:s.revertLane,gesture:s.gesture,action:s.action,hasEagerState:s.hasEagerState,eagerState:s.eagerState,next:null},i===null?(r=i=c,u=n):i=i.next=c,K.lanes|=d,rl|=d;s=s.next}while(s!==null&&s!==t);if(i===null?u=n:i.next=r,!ta(n,e.memoizedState)&&(qe=!0,f&&(a=Tn,a!==null)))throw a;e.memoizedState=n,e.baseState=u,e.baseQueue=i,o.lastRenderedState=n}return l===null&&(o.lanes=0),[e.memoizedState,o.dispatch]}function mc(e){var t=ke(),a=t.queue;if(a===null)throw Error(E(311));a.lastRenderedReducer=e;var o=a.dispatch,l=a.pending,n=t.memoizedState;if(l!==null){a.pending=null;var u=l=l.next;do n=e(n,u.action),u=u.next;while(u!==l);ta(n,t.memoizedState)||(qe=!0),t.memoizedState=n,t.baseQueue===null&&(t.baseState=n),a.lastRenderedState=n}return[n,o]}function IS(e,t,a){var o=K,l=ke(),n=oe;if(n){if(a===void 0)throw Error(E(407));a=a()}else a=t();var u=!ta((pe||l).memoizedState,a);if(u&&(l.memoizedState=a,qe=!0),l=l.queue,ap(AS.bind(null,o,l,e),[e]),l.getSnapshot!==t||u||He!==null&&He.memoizedState.tag&1){if(o.flags|=2048,Nn(9,{destroy:void 0},wS.bind(null,o,l,a,t),null),Se===null)throw Error(E(349));n||(Lo&127)!==0||bS(o,t,a)}return a}function bS(e,t,a){e.flags|=16384,e={getSnapshot:t,value:a},t=K.updateQueue,t===null?(t=ds(),K.updateQueue=t,t.stores=[e]):(a=t.stores,a===null?t.stores=[e]:a.push(e))}function wS(e,t,a,o){t.value=a,t.getSnapshot=o,TS(t)&&ES(e)}function AS(e,t,a){return a(function(){TS(t)&&ES(e)})}function TS(e){var t=e.getSnapshot;e=e.value;try{var a=t();return!ta(e,a)}catch(o){return!0}}function ES(e){var t=kl(e,2);t!==null&&kt(t,e,2)}function ld(e){var t=Tt();if(typeof e=="function"){var a=e;if(e=a(),Pl){Vo(!0);try{a()}finally{Vo(!1)}}}return t.memoizedState=t.baseState=e,t.queue={pending:null,lanes:0,dispatch:null,lastRenderedReducer:yo,lastRenderedState:e},t}function DS(e,t,a,o){return e.baseState=a,tp(e,pe,typeof o=="function"?o:yo)}function sw(e,t,a,o,l){if(hs(e))throw Error(E(485));if(e=t.action,e!==null){var n={payload:l,action:e,next:null,isTransition:!0,status:"pending",value:null,reason:null,listeners:[],then:function(u){n.listeners.push(u)}};V.T!==null?a(!0):n.isTransition=!1,o(n),a=t.pending,a===null?(n.next=t.pending=n,MS(t,n)):(n.next=a.next,t.pending=a.next=n)}}function MS(e,t){var a=t.action,o=t.payload,l=e.state;if(t.isTransition){var n=V.T,u={};V.T=u;try{var r=a(l,o),i=V.S;i!==null&&i(u,r),Dg(e,t,r)}catch(s){nd(e,t,s)}finally{n!==null&&u.types!==null&&(n.types=u.types),V.T=n}}else try{n=a(l,o),Dg(e,t,n)}catch(s){nd(e,t,s)}}function Dg(e,t,a){a!==null&&typeof a=="object"&&typeof a.then=="function"?a.then(function(o){Mg(e,t,o)},function(o){return nd(e,t,o)}):Mg(e,t,a)}function Mg(e,t,a){t.status="fulfilled",t.value=a,OS(t),e.state=a,t=e.pending,t!==null&&(a=t.next,a===t?e.pending=null:(a=a.next,t.next=a,MS(e,a)))}function nd(e,t,a){var o=e.pending;if(e.pending=null,o!==null){o=o.next;do t.status="rejected",t.reason=a,OS(t),t=t.next;while(t!==o)}e.action=null}function OS(e){e=e.listeners;for(var t=0;t<e.length;t++)(0,e[t])()}function PS(e,t){return t}function Og(e,t){if(oe){var a=Se.formState;if(a!==null){e:{var o=K;if(oe){if(Re){t:{for(var l=Re,n=ma;l.nodeType!==8;){if(!n){l=null;break t}if(l=ga(l.nextSibling),l===null){l=null;break t}}n=l.data,l=n==="F!"||n==="F"?l:null}if(l){Re=ga(l.nextSibling),o=l.data==="F!";break e}}nl(o)}o=!1}o&&(t=a[0])}}return a=Tt(),a.memoizedState=a.baseState=t,o={pending:null,lanes:0,dispatch:null,lastRenderedReducer:PS,lastRenderedState:t},a.queue=o,a=ZS.bind(null,K,o),o.dispatch=a,o=ld(!1),n=up.bind(null,K,!1,o.queue),o=Tt(),l={state:t,dispatch:null,action:e,pending:null},o.queue=l,a=sw.bind(null,K,l,n,a),l.dispatch=a,o.memoizedState=e,[t,a,!1]}function Pg(e){var t=ke();return BS(t,pe,e)}function BS(e,t,a){if(t=tp(e,t,PS)[0],e=yi(yo)[0],typeof t=="object"&&t!==null&&typeof t.then=="function")try{var o=dr(t)}catch(u){throw u===jn?cs:u}else o=t;t=ke();var l=t.queue,n=l.dispatch;return a!==t.memoizedState&&(K.flags|=2048,Nn(9,{destroy:void 0},fw.bind(null,l,a),null)),[o,n,e]}function fw(e,t){e.action=t}function Bg(e){var t=ke(),a=pe;if(a!==null)return BS(t,a,e);ke(),t=t.memoizedState,a=ke();var o=a.queue.dispatch;return a.memoizedState=e,[t,o,!1]}function Nn(e,t,a,o){return e={tag:e,create:a,deps:o,inst:t,next:null},t=K.updateQueue,t===null&&(t=ds(),K.updateQueue=t),a=t.lastEffect,a===null?t.lastEffect=e.next=e:(o=a.next,a.next=e,e.next=o,t.lastEffect=e),e}function _S(){return ke().memoizedState}function Ci(e,t,a,o){var l=Tt();K.flags|=e,l.memoizedState=Nn(1|t,{destroy:void 0},a,o===void 0?null:o)}function ms(e,t,a,o){var l=ke();o=o===void 0?null:o;var n=l.memoizedState.inst;pe!==null&&o!==null&&Zd(o,pe.memoizedState.deps)?l.memoizedState=Nn(t,n,a,o):(K.flags|=e,l.memoizedState=Nn(1|t,n,a,o))}function _g(e,t){Ci(8390656,8,e,t)}function ap(e,t){ms(2048,8,e,t)}function cw(e){K.flags|=4;var t=K.updateQueue;if(t===null)t=ds(),K.updateQueue=t,t.events=[e];else{var a=t.events;a===null?t.events=[e]:a.push(e)}}function US(e){var t=ke().memoizedState;return cw({ref:t,nextImpl:e}),function(){if((re&2)!==0)throw Error(E(440));return t.impl.apply(void 0,arguments)}}function NS(e,t){return ms(4,2,e,t)}function kS(e,t){return ms(4,4,e,t)}function HS(e,t){if(typeof t=="function"){e=e();var a=t(e);return function(){typeof a=="function"?a():t(null)}}if(t!=null)return e=e(),t.current=e,function(){t.current=null}}function zS(e,t,a){a=a!=null?a.concat([e]):null,ms(4,4,HS.bind(null,t,e),a)}function op(){}function FS(e,t){var a=ke();t=t===void 0?null:t;var o=a.memoizedState;return t!==null&&Zd(t,o[1])?o[0]:(a.memoizedState=[e,t],e)}function qS(e,t){var a=ke();t=t===void 0?null:t;var o=a.memoizedState;if(t!==null&&Zd(t,o[1]))return o[0];if(o=e(),Pl){Vo(!0);try{e()}finally{Vo(!1)}}return a.memoizedState=[o,t],o}function lp(e,t,a){return a===void 0||(Lo&1073741824)!==0&&(ae&261930)===0?e.memoizedState=t:(e.memoizedState=a,e=Dx(),K.lanes|=e,rl|=e,a)}function GS(e,t,a,o){return ta(a,t)?a:Un.current!==null?(e=lp(e,a,o),ta(e,t)||(qe=!0),e):(Lo&42)===0||(Lo&1073741824)!==0&&(ae&261930)===0?(qe=!0,e.memoizedState=a):(e=Dx(),K.lanes|=e,rl|=e,t)}function VS(e,t,a,o,l){var n=ie.p;ie.p=n!==0&&8>n?n:8;var u=V.T,r={};V.T=r,up(e,!1,t,a);try{var i=l(),s=V.S;if(s!==null&&s(r,i),i!==null&&typeof i=="object"&&typeof i.then=="function"){var f=uw(i,o);Hu(e,t,f,ea(e))}else Hu(e,t,o,ea(e))}catch(d){Hu(e,t,{then:function(){},status:"rejected",reason:d},ea())}finally{ie.p=n,u!==null&&r.types!==null&&(u.types=r.types),V.T=u}}function dw(){}function ud(e,t,a,o){if(e.tag!==5)throw Error(E(476));var l=XS(e).queue;VS(e,l,t,bl,a===null?dw:function(){return jS(e),a(o)})}function XS(e){var t=e.memoizedState;if(t!==null)return t;t={memoizedState:bl,baseState:bl,baseQueue:null,queue:{pending:null,lanes:0,dispatch:null,lastRenderedReducer:yo,lastRenderedState:bl},next:null};var a={};return t.next={memoizedState:a,baseState:a,baseQueue:null,queue:{pending:null,lanes:0,dispatch:null,lastRenderedReducer:yo,lastRenderedState:a},next:null},e.memoizedState=t,e=e.alternate,e!==null&&(e.memoizedState=t),t}function jS(e){var t=XS(e);t.next===null&&(t=e.alternate.memoizedState),Hu(e,t.next.queue,{},ea())}function np(){return gt(ar)}function YS(){return ke().memoizedState}function KS(){return ke().memoizedState}function pw(e){for(var t=e.return;t!==null;){switch(t.tag){case 24:case 3:var a=ea();e=Wo(a);var o=Jo(t,e,a);o!==null&&(kt(o,t,a),Uu(o,t,a)),t={cache:Vd()},e.payload=t;return}t=t.return}}function mw(e,t,a){var o=ea();a={lane:o,revertLane:0,gesture:null,action:a,hasEagerState:!1,eagerState:null,next:null},hs(e)?QS(t,a):(a=zd(e,t,a,o),a!==null&&(kt(a,e,o),WS(a,t,o)))}function ZS(e,t,a){var o=ea();Hu(e,t,a,o)}function Hu(e,t,a,o){var l={lane:o,revertLane:0,gesture:null,action:a,hasEagerState:!1,eagerState:null,next:null};if(hs(e))QS(t,l);else{var n=e.alternate;if(e.lanes===0&&(n===null||n.lanes===0)&&(n=t.lastRenderedReducer,n!==null))try{var u=t.lastRenderedState,r=n(u,a);if(l.hasEagerState=!0,l.eagerState=r,ta(r,u))return fs(e,t,l,0),Se===null&&ss(),!1}catch(i){}if(a=zd(e,t,l,o),a!==null)return kt(a,e,o),WS(a,t,o),!0}return!1}function up(e,t,a,o){if(o={lane:2,revertLane:hp(),gesture:null,action:o,hasEagerState:!1,eagerState:null,next:null},hs(e)){if(t)throw Error(E(479))}else t=zd(e,a,o,2),t!==null&&kt(t,e,2)}function hs(e){var t=e.alternate;return e===K||t!==null&&t===K}function QS(e,t){Dn=qi=!0;var a=e.pending;a===null?t.next=t:(t.next=a.next,a.next=t),e.pending=t}function WS(e,t,a){if((a&4194048)!==0){var o=t.lanes;o&=e.pendingLanes,a|=o,t.lanes=a,Uv(e,a)}}var $u={readContext:gt,use:ps,useCallback:Oe,useContext:Oe,useEffect:Oe,useImperativeHandle:Oe,useLayoutEffect:Oe,useInsertionEffect:Oe,useMemo:Oe,useReducer:Oe,useRef:Oe,useState:Oe,useDebugValue:Oe,useDeferredValue:Oe,useTransition:Oe,useSyncExternalStore:Oe,useId:Oe,useHostTransitionStatus:Oe,useFormState:Oe,useActionState:Oe,useOptimistic:Oe,useMemoCache:Oe,useCacheRefresh:Oe};$u.useEffectEvent=Oe;var JS={readContext:gt,use:ps,useCallback:function(e,t){return Tt().memoizedState=[e,t===void 0?null:t],e},useContext:gt,useEffect:_g,useImperativeHandle:function(e,t,a){a=a!=null?a.concat([e]):null,Ci(4194308,4,HS.bind(null,t,e),a)},useLayoutEffect:function(e,t){return Ci(4194308,4,e,t)},useInsertionEffect:function(e,t){Ci(4,2,e,t)},useMemo:function(e,t){var a=Tt();t=t===void 0?null:t;var o=e();if(Pl){Vo(!0);try{e()}finally{Vo(!1)}}return a.memoizedState=[o,t],o},useReducer:function(e,t,a){var o=Tt();if(a!==void 0){var l=a(t);if(Pl){Vo(!0);try{a(t)}finally{Vo(!1)}}}else l=t;return o.memoizedState=o.baseState=l,e={pending:null,lanes:0,dispatch:null,lastRenderedReducer:e,lastRenderedState:l},o.queue=e,e=e.dispatch=mw.bind(null,K,e),[o.memoizedState,e]},useRef:function(e){var t=Tt();return e={current:e},t.memoizedState=e},useState:function(e){e=ld(e);var t=e.queue,a=ZS.bind(null,K,t);return t.dispatch=a,[e.memoizedState,a]},useDebugValue:op,useDeferredValue:function(e,t){var a=Tt();return lp(a,e,t)},useTransition:function(){var e=ld(!1);return e=VS.bind(null,K,e.queue,!0,!1),Tt().memoizedState=e,[!1,e]},useSyncExternalStore:function(e,t,a){var o=K,l=Tt();if(oe){if(a===void 0)throw Error(E(407));a=a()}else{if(a=t(),Se===null)throw Error(E(349));(ae&127)!==0||bS(o,t,a)}l.memoizedState=a;var n={value:a,getSnapshot:t};return l.queue=n,_g(AS.bind(null,o,n,e),[e]),o.flags|=2048,Nn(9,{destroy:void 0},wS.bind(null,o,n,a,t),null),a},useId:function(){var e=Tt(),t=Se.identifierPrefix;if(oe){var a=ja,o=Xa;a=(o&~(1<<32-$t(o)-1)).toString(32)+a,t="_"+t+"R_"+a,a=Gi++,0<a&&(t+="H"+a.toString(32)),t+="_"}else a=rw++,t="_"+t+"r_"+a.toString(32)+"_";return e.memoizedState=t},useHostTransitionStatus:np,useFormState:Og,useActionState:Og,useOptimistic:function(e){var t=Tt();t.memoizedState=t.baseState=e;var a={pending:null,lanes:0,dispatch:null,lastRenderedReducer:null,lastRenderedState:null};return t.queue=a,t=up.bind(null,K,!0,a),a.dispatch=t,[e,t]},useMemoCache:ep,useCacheRefresh:function(){return Tt().memoizedState=pw.bind(null,K)},useEffectEvent:function(e){var t=Tt(),a={impl:e};return t.memoizedState=a,function(){if((re&2)!==0)throw Error(E(440));return a.impl.apply(void 0,arguments)}}},rp={readContext:gt,use:ps,useCallback:FS,useContext:gt,useEffect:ap,useImperativeHandle:zS,useInsertionEffect:NS,useLayoutEffect:kS,useMemo:qS,useReducer:yi,useRef:_S,useState:function(){return yi(yo)},useDebugValue:op,useDeferredValue:function(e,t){var a=ke();return GS(a,pe.memoizedState,e,t)},useTransition:function(){var e=yi(yo)[0],t=ke().memoizedState;return[typeof e=="boolean"?e:dr(e),t]},useSyncExternalStore:IS,useId:YS,useHostTransitionStatus:np,useFormState:Pg,useActionState:Pg,useOptimistic:function(e,t){var a=ke();return DS(a,pe,e,t)},useMemoCache:ep,useCacheRefresh:KS};rp.useEffectEvent=US;var $S={readContext:gt,use:ps,useCallback:FS,useContext:gt,useEffect:ap,useImperativeHandle:zS,useInsertionEffect:NS,useLayoutEffect:kS,useMemo:qS,useReducer:mc,useRef:_S,useState:function(){return mc(yo)},useDebugValue:op,useDeferredValue:function(e,t){var a=ke();return pe===null?lp(a,e,t):GS(a,pe.memoizedState,e,t)},useTransition:function(){var e=mc(yo)[0],t=ke().memoizedState;return[typeof e=="boolean"?e:dr(e),t]},useSyncExternalStore:IS,useId:YS,useHostTransitionStatus:np,useFormState:Bg,useActionState:Bg,useOptimistic:function(e,t){var a=ke();return pe!==null?DS(a,pe,e,t):(a.baseState=e,[e,a.queue.dispatch])},useMemoCache:ep,useCacheRefresh:KS};$S.useEffectEvent=US;function hc(e,t,a,o){t=e.memoizedState,a=a(o,t),a=a==null?t:Ie({},t,a),e.memoizedState=a,e.lanes===0&&(e.updateQueue.baseState=a)}var rd={enqueueSetState:function(e,t,a){e=e._reactInternals;var o=ea(),l=Wo(o);l.payload=t,a!=null&&(l.callback=a),t=Jo(e,l,o),t!==null&&(kt(t,e,o),Uu(t,e,o))},enqueueReplaceState:function(e,t,a){e=e._reactInternals;var o=ea(),l=Wo(o);l.tag=1,l.payload=t,a!=null&&(l.callback=a),t=Jo(e,l,o),t!==null&&(kt(t,e,o),Uu(t,e,o))},enqueueForceUpdate:function(e,t){e=e._reactInternals;var a=ea(),o=Wo(a);o.tag=2,t!=null&&(o.callback=t),t=Jo(e,o,a),t!==null&&(kt(t,e,a),Uu(t,e,a))}};function Ug(e,t,a,o,l,n,u){return e=e.stateNode,typeof e.shouldComponentUpdate=="function"?e.shouldComponentUpdate(o,n,u):t.prototype&&t.prototype.isPureReactComponent?!Ku(a,o)||!Ku(l,n):!0}function Ng(e,t,a,o){e=t.state,typeof t.componentWillReceiveProps=="function"&&t.componentWillReceiveProps(a,o),typeof t.UNSAFE_componentWillReceiveProps=="function"&&t.UNSAFE_componentWillReceiveProps(a,o),t.state!==e&&rd.enqueueReplaceState(t,t.state,null)}function Bl(e,t){var a=t;if("ref"in t){a={};for(var o in t)o!=="ref"&&(a[o]=t[o])}if(e=e.defaultProps){a===t&&(a=Ie({},a));for(var l in e)a[l]===void 0&&(a[l]=e[l])}return a}function ex(e){_i(e)}function tx(e){console.error(e)}function ax(e){_i(e)}function Vi(e,t){try{var a=e.onUncaughtError;a(t.value,{componentStack:t.stack})}catch(o){setTimeout(function(){throw o})}}function kg(e,t,a){try{var o=e.onCaughtError;o(a.value,{componentStack:a.stack,errorBoundary:t.tag===1?t.stateNode:null})}catch(l){setTimeout(function(){throw l})}}function id(e,t,a){return a=Wo(a),a.tag=3,a.payload={element:null},a.callback=function(){Vi(e,t)},a}function ox(e){return e=Wo(e),e.tag=3,e}function lx(e,t,a,o){var l=a.type.getDerivedStateFromError;if(typeof l=="function"){var n=o.value;e.payload=function(){return l(n)},e.callback=function(){kg(t,a,o)}}var u=a.stateNode;u!==null&&typeof u.componentDidCatch=="function"&&(e.callback=function(){kg(t,a,o),typeof l!="function"&&($o===null?$o=new Set([this]):$o.add(this));var r=o.stack;this.componentDidCatch(o.value,{componentStack:r!==null?r:""})})}function hw(e,t,a,o,l){if(a.flags|=32768,o!==null&&typeof o=="object"&&typeof o.then=="function"){if(t=a.alternate,t!==null&&Xn(t,a,l,!0),a=aa.current,a!==null){switch(a.tag){case 31:case 13:return ha===null?Zi():a.alternate===null&&Pe===0&&(Pe=3),a.flags&=-257,a.flags|=65536,a.lanes=l,o===Hi?a.flags|=16384:(t=a.updateQueue,t===null?a.updateQueue=new Set([o]):t.add(o),wc(e,o,l)),!1;case 22:return a.flags|=65536,o===Hi?a.flags|=16384:(t=a.updateQueue,t===null?(t={transitions:null,markerInstances:null,retryQueue:new Set([o])},a.updateQueue=t):(a=t.retryQueue,a===null?t.retryQueue=new Set([o]):a.add(o)),wc(e,o,l)),!1}throw Error(E(435,a.tag))}return wc(e,o,l),Zi(),!1}if(oe)return t=aa.current,t!==null?((t.flags&65536)===0&&(t.flags|=256),t.flags|=65536,t.lanes=l,o!==Zc&&(e=Error(E(422),{cause:o}),Qu(pa(e,a)))):(o!==Zc&&(t=Error(E(423),{cause:o}),Qu(pa(t,a))),e=e.current.alternate,e.flags|=65536,l&=-l,e.lanes|=l,o=pa(o,a),l=id(e.stateNode,o,l),pc(e,l),Pe!==4&&(Pe=2)),!1;var n=Error(E(520),{cause:o});if(n=pa(n,a),qu===null?qu=[n]:qu.push(n),Pe!==4&&(Pe=2),t===null)return!0;o=pa(o,a),a=t;do{switch(a.tag){case 3:return a.flags|=65536,e=l&-l,a.lanes|=e,e=id(a.stateNode,o,e),pc(a,e),!1;case 1:if(t=a.type,n=a.stateNode,(a.flags&128)===0&&(typeof t.getDerivedStateFromError=="function"||n!==null&&typeof n.componentDidCatch=="function"&&($o===null||!$o.has(n))))return a.flags|=65536,l&=-l,a.lanes|=l,l=ox(l),lx(l,e,a,o),pc(a,l),!1}a=a.return}while(a!==null);return!1}var ip=Error(E(461)),qe=!1;function pt(e,t,a,o){t.child=e===null?SS(t,null,a,o):Ol(t,e.child,a,o)}function Hg(e,t,a,o,l){a=a.render;var n=t.ref;if("ref"in o){var u={};for(var r in o)r!=="ref"&&(u[r]=o[r])}else u=o;return Ml(t),o=Qd(e,t,a,u,n,l),r=Wd(),e!==null&&!qe?(Jd(e,t,l),Co(e,t,l)):(oe&&r&&qd(t),t.flags|=1,pt(e,t,o,l),t.child)}function zg(e,t,a,o,l){if(e===null){var n=a.type;return typeof n=="function"&&!Fd(n)&&n.defaultProps===void 0&&a.compare===null?(t.tag=15,t.type=n,nx(e,t,n,o,l)):(e=xi(a.type,null,o,t,t.mode,l),e.ref=t.ref,e.return=t,t.child=e)}if(n=e.child,!sp(e,l)){var u=n.memoizedProps;if(a=a.compare,a=a!==null?a:Ku,a(u,o)&&e.ref===t.ref)return Co(e,t,l)}return t.flags|=1,e=go(n,o),e.ref=t.ref,e.return=t,t.child=e}function nx(e,t,a,o,l){if(e!==null){var n=e.memoizedProps;if(Ku(n,o)&&e.ref===t.ref)if(qe=!1,t.pendingProps=o=n,sp(e,l))(e.flags&131072)!==0&&(qe=!0);else return t.lanes=e.lanes,Co(e,t,l)}return sd(e,t,a,o,l)}function ux(e,t,a,o){var l=o.children,n=e!==null?e.memoizedState:null;if(e===null&&t.stateNode===null&&(t.stateNode={_visibility:1,_pendingMarkers:null,_retryCache:null,_transitions:null}),o.mode==="hidden"){if((t.flags&128)!==0){if(n=n!==null?n.baseLanes|a:a,e!==null){for(o=t.child=e.child,l=0;o!==null;)l=l|o.lanes|o.childLanes,o=o.sibling;o=l&~n}else o=0,t.child=null;return Fg(e,t,n,a,o)}if((a&536870912)!==0)t.memoizedState={baseLanes:0,cachePool:null},e!==null&&Li(t,n!==null?n.cachePool:null),n!==null?Eg(t,n):ad(),yS(t);else return o=t.lanes=536870912,Fg(e,t,n!==null?n.baseLanes|a:a,a,o)}else n!==null?(Li(t,n.cachePool),Eg(t,n),qo(t),t.memoizedState=null):(e!==null&&Li(t,null),ad(),qo(t));return pt(e,t,l,a),t.child}function Du(e,t){return e!==null&&e.tag===22||t.stateNode!==null||(t.stateNode={_visibility:1,_pendingMarkers:null,_retryCache:null,_transitions:null}),t.sibling}function Fg(e,t,a,o,l){var n=Xd();return n=n===null?null:{parent:Fe._currentValue,pool:n},t.memoizedState={baseLanes:a,cachePool:n},e!==null&&Li(t,null),ad(),yS(t),e!==null&&Xn(e,t,o,!0),t.childLanes=l,null}function Ri(e,t){return t=Xi({mode:t.mode,children:t.children},e.mode),t.ref=e.ref,e.child=t,t.return=e,t}function qg(e,t,a){return Ol(t,e.child,null,a),e=Ri(t,t.pendingProps),e.flags|=2,Kt(t),t.memoizedState=null,e}function gw(e,t,a){var o=t.pendingProps,l=(t.flags&128)!==0;if(t.flags&=-129,e===null){if(oe){if(o.mode==="hidden")return e=Ri(t,o),t.lanes=536870912,Du(null,e);if(od(t),(e=Re)?(e=Jx(e,ma),e=e!==null&&e.data==="&"?e:null,e!==null&&(t.memoizedState={dehydrated:e,treeContext:ll!==null?{id:Xa,overflow:ja}:null,retryLane:536870912,hydrationErrors:null},a=cS(e),a.return=t,t.child=a,ht=t,Re=null)):e=null,e===null)throw nl(t);return t.lanes=536870912,null}return Ri(t,o)}var n=e.memoizedState;if(n!==null){var u=n.dehydrated;if(od(t),l)if(t.flags&256)t.flags&=-257,t=qg(e,t,a);else if(t.memoizedState!==null)t.child=e.child,t.flags|=128,t=null;else throw Error(E(558));else if(qe||Xn(e,t,a,!1),l=(a&e.childLanes)!==0,qe||l){if(o=Se,o!==null&&(u=Nv(o,a),u!==0&&u!==n.retryLane))throw n.retryLane=u,kl(e,u),kt(o,e,u),ip;Zi(),t=qg(e,t,a)}else e=n.treeContext,Re=ga(u.nextSibling),ht=t,oe=!0,Qo=null,ma=!1,e!==null&&pS(t,e),t=Ri(t,o),t.flags|=4096;return t}return e=go(e.child,{mode:o.mode,children:o.children}),e.ref=t.ref,t.child=e,e.return=t,e}function Ii(e,t){var a=t.ref;if(a===null)e!==null&&e.ref!==null&&(t.flags|=4194816);else{if(typeof a!="function"&&typeof a!="object")throw Error(E(284));(e===null||e.ref!==a)&&(t.flags|=4194816)}}function sd(e,t,a,o,l){return Ml(t),a=Qd(e,t,a,o,void 0,l),o=Wd(),e!==null&&!qe?(Jd(e,t,l),Co(e,t,l)):(oe&&o&&qd(t),t.flags|=1,pt(e,t,a,l),t.child)}function Gg(e,t,a,o,l,n){return Ml(t),t.updateQueue=null,a=RS(t,o,a,l),CS(e),o=Wd(),e!==null&&!qe?(Jd(e,t,n),Co(e,t,n)):(oe&&o&&qd(t),t.flags|=1,pt(e,t,a,n),t.child)}function Vg(e,t,a,o,l){if(Ml(t),t.stateNode===null){var n=yn,u=a.contextType;typeof u=="object"&&u!==null&&(n=gt(u)),n=new a(o,n),t.memoizedState=n.state!==null&&n.state!==void 0?n.state:null,n.updater=rd,t.stateNode=n,n._reactInternals=t,n=t.stateNode,n.props=o,n.state=t.memoizedState,n.refs={},Yd(t),u=a.contextType,n.context=typeof u=="object"&&u!==null?gt(u):yn,n.state=t.memoizedState,u=a.getDerivedStateFromProps,typeof u=="function"&&(hc(t,a,u,o),n.state=t.memoizedState),typeof a.getDerivedStateFromProps=="function"||typeof n.getSnapshotBeforeUpdate=="function"||typeof n.UNSAFE_componentWillMount!="function"&&typeof n.componentWillMount!="function"||(u=n.state,typeof n.componentWillMount=="function"&&n.componentWillMount(),typeof n.UNSAFE_componentWillMount=="function"&&n.UNSAFE_componentWillMount(),u!==n.state&&rd.enqueueReplaceState(n,n.state,null),ku(t,o,n,l),Nu(),n.state=t.memoizedState),typeof n.componentDidMount=="function"&&(t.flags|=4194308),o=!0}else if(e===null){n=t.stateNode;var r=t.memoizedProps,i=Bl(a,r);n.props=i;var s=n.context,f=a.contextType;u=yn,typeof f=="object"&&f!==null&&(u=gt(f));var d=a.getDerivedStateFromProps;f=typeof d=="function"||typeof n.getSnapshotBeforeUpdate=="function",r=t.pendingProps!==r,f||typeof n.UNSAFE_componentWillReceiveProps!="function"&&typeof n.componentWillReceiveProps!="function"||(r||s!==u)&&Ng(t,n,o,u),Ho=!1;var c=t.memoizedState;n.state=c,ku(t,o,n,l),Nu(),s=t.memoizedState,r||c!==s||Ho?(typeof d=="function"&&(hc(t,a,d,o),s=t.memoizedState),(i=Ho||Ug(t,a,i,o,c,s,u))?(f||typeof n.UNSAFE_componentWillMount!="function"&&typeof n.componentWillMount!="function"||(typeof n.componentWillMount=="function"&&n.componentWillMount(),typeof n.UNSAFE_componentWillMount=="function"&&n.UNSAFE_componentWillMount()),typeof n.componentDidMount=="function"&&(t.flags|=4194308)):(typeof n.componentDidMount=="function"&&(t.flags|=4194308),t.memoizedProps=o,t.memoizedState=s),n.props=o,n.state=s,n.context=u,o=i):(typeof n.componentDidMount=="function"&&(t.flags|=4194308),o=!1)}else{n=t.stateNode,ed(e,t),u=t.memoizedProps,f=Bl(a,u),n.props=f,d=t.pendingProps,c=n.context,s=a.contextType,i=yn,typeof s=="object"&&s!==null&&(i=gt(s)),r=a.getDerivedStateFromProps,(s=typeof r=="function"||typeof n.getSnapshotBeforeUpdate=="function")||typeof n.UNSAFE_componentWillReceiveProps!="function"&&typeof n.componentWillReceiveProps!="function"||(u!==d||c!==i)&&Ng(t,n,o,i),Ho=!1,c=t.memoizedState,n.state=c,ku(t,o,n,l),Nu();var g=t.memoizedState;u!==d||c!==g||Ho||e!==null&&e.dependencies!==null&&ki(e.dependencies)?(typeof r=="function"&&(hc(t,a,r,o),g=t.memoizedState),(f=Ho||Ug(t,a,f,o,c,g,i)||e!==null&&e.dependencies!==null&&ki(e.dependencies))?(s||typeof n.UNSAFE_componentWillUpdate!="function"&&typeof n.componentWillUpdate!="function"||(typeof n.componentWillUpdate=="function"&&n.componentWillUpdate(o,g,i),typeof n.UNSAFE_componentWillUpdate=="function"&&n.UNSAFE_componentWillUpdate(o,g,i)),typeof n.componentDidUpdate=="function"&&(t.flags|=4),typeof n.getSnapshotBeforeUpdate=="function"&&(t.flags|=1024)):(typeof n.componentDidUpdate!="function"||u===e.memoizedProps&&c===e.memoizedState||(t.flags|=4),typeof n.getSnapshotBeforeUpdate!="function"||u===e.memoizedProps&&c===e.memoizedState||(t.flags|=1024),t.memoizedProps=o,t.memoizedState=g),n.props=o,n.state=g,n.context=i,o=f):(typeof n.componentDidUpdate!="function"||u===e.memoizedProps&&c===e.memoizedState||(t.flags|=4),typeof n.getSnapshotBeforeUpdate!="function"||u===e.memoizedProps&&c===e.memoizedState||(t.flags|=1024),o=!1)}return n=o,Ii(e,t),o=(t.flags&128)!==0,n||o?(n=t.stateNode,a=o&&typeof a.getDerivedStateFromError!="function"?null:n.render(),t.flags|=1,e!==null&&o?(t.child=Ol(t,e.child,null,l),t.child=Ol(t,null,a,l)):pt(e,t,a,l),t.memoizedState=n.state,e=t.child):e=Co(e,t,l),e}function Xg(e,t,a,o){return Dl(),t.flags|=256,pt(e,t,a,o),t.child}var gc={dehydrated:null,treeContext:null,retryLane:0,hydrationErrors:null};function vc(e){return{baseLanes:e,cachePool:hS()}}function Sc(e,t,a){return e=e!==null?e.childLanes&~a:0,t&&(e|=Qt),e}function rx(e,t,a){var o=t.pendingProps,l=!1,n=(t.flags&128)!==0,u;if((u=n)||(u=e!==null&&e.memoizedState===null?!1:(Ne.current&2)!==0),u&&(l=!0,t.flags&=-129),u=(t.flags&32)!==0,t.flags&=-33,e===null){if(oe){if(l?Fo(t):qo(t),(e=Re)?(e=Jx(e,ma),e=e!==null&&e.data!=="&"?e:null,e!==null&&(t.memoizedState={dehydrated:e,treeContext:ll!==null?{id:Xa,overflow:ja}:null,retryLane:536870912,hydrationErrors:null},a=cS(e),a.return=t,t.child=a,ht=t,Re=null)):e=null,e===null)throw nl(t);return Rd(e)?t.lanes=32:t.lanes=536870912,null}var r=o.children;return o=o.fallback,l?(qo(t),l=t.mode,r=Xi({mode:"hidden",children:r},l),o=wl(o,l,a,null),r.return=t,o.return=t,r.sibling=o,t.child=r,o=t.child,o.memoizedState=vc(a),o.childLanes=Sc(e,u,a),t.memoizedState=gc,Du(null,o)):(Fo(t),fd(t,r))}var i=e.memoizedState;if(i!==null&&(r=i.dehydrated,r!==null)){if(n)t.flags&256?(Fo(t),t.flags&=-257,t=xc(e,t,a)):t.memoizedState!==null?(qo(t),t.child=e.child,t.flags|=128,t=null):(qo(t),r=o.fallback,l=t.mode,o=Xi({mode:"visible",children:o.children},l),r=wl(r,l,a,null),r.flags|=2,o.return=t,r.return=t,o.sibling=r,t.child=o,Ol(t,e.child,null,a),o=t.child,o.memoizedState=vc(a),o.childLanes=Sc(e,u,a),t.memoizedState=gc,t=Du(null,o));else if(Fo(t),Rd(r)){if(u=r.nextSibling&&r.nextSibling.dataset,u)var s=u.dgst;u=s,o=Error(E(419)),o.stack="",o.digest=u,Qu({value:o,source:null,stack:null}),t=xc(e,t,a)}else if(qe||Xn(e,t,a,!1),u=(a&e.childLanes)!==0,qe||u){if(u=Se,u!==null&&(o=Nv(u,a),o!==0&&o!==i.retryLane))throw i.retryLane=o,kl(e,o),kt(u,e,o),ip;Cd(r)||Zi(),t=xc(e,t,a)}else Cd(r)?(t.flags|=192,t.child=e.child,t=null):(e=i.treeContext,Re=ga(r.nextSibling),ht=t,oe=!0,Qo=null,ma=!1,e!==null&&pS(t,e),t=fd(t,o.children),t.flags|=4096);return t}return l?(qo(t),r=o.fallback,l=t.mode,i=e.child,s=i.sibling,o=go(i,{mode:"hidden",children:o.children}),o.subtreeFlags=i.subtreeFlags&65011712,s!==null?r=go(s,r):(r=wl(r,l,a,null),r.flags|=2),r.return=t,o.return=t,o.sibling=r,t.child=o,Du(null,o),o=t.child,r=e.child.memoizedState,r===null?r=vc(a):(l=r.cachePool,l!==null?(i=Fe._currentValue,l=l.parent!==i?{parent:i,pool:i}:l):l=hS(),r={baseLanes:r.baseLanes|a,cachePool:l}),o.memoizedState=r,o.childLanes=Sc(e,u,a),t.memoizedState=gc,Du(e.child,o)):(Fo(t),a=e.child,e=a.sibling,a=go(a,{mode:"visible",children:o.children}),a.return=t,a.sibling=null,e!==null&&(u=t.deletions,u===null?(t.deletions=[e],t.flags|=16):u.push(e)),t.child=a,t.memoizedState=null,a)}function fd(e,t){return t=Xi({mode:"visible",children:t},e.mode),t.return=e,e.child=t}function Xi(e,t){return e=Zt(22,e,null,t),e.lanes=0,e}function xc(e,t,a){return Ol(t,e.child,null,a),e=fd(t,t.pendingProps.children),e.flags|=2,t.memoizedState=null,e}function jg(e,t,a){e.lanes|=t;var o=e.alternate;o!==null&&(o.lanes|=t),Wc(e.return,t,a)}function Lc(e,t,a,o,l,n){var u=e.memoizedState;u===null?e.memoizedState={isBackwards:t,rendering:null,renderingStartTime:0,last:o,tail:a,tailMode:l,treeForkCount:n}:(u.isBackwards=t,u.rendering=null,u.renderingStartTime=0,u.last=o,u.tail=a,u.tailMode=l,u.treeForkCount=n)}function ix(e,t,a){var o=t.pendingProps,l=o.revealOrder,n=o.tail;o=o.children;var u=Ne.current,r=(u&2)!==0;if(r?(u=u&1|2,t.flags|=128):u&=1,xe(Ne,u),pt(e,t,o,a),o=oe?Zu:0,!r&&e!==null&&(e.flags&128)!==0)e:for(e=t.child;e!==null;){if(e.tag===13)e.memoizedState!==null&&jg(e,a,t);else if(e.tag===19)jg(e,a,t);else if(e.child!==null){e.child.return=e,e=e.child;continue}if(e===t)break e;for(;e.sibling===null;){if(e.return===null||e.return===t)break e;e=e.return}e.sibling.return=e.return,e=e.sibling}switch(l){case"forwards":for(a=t.child,l=null;a!==null;)e=a.alternate,e!==null&&Fi(e)===null&&(l=a),a=a.sibling;a=l,a===null?(l=t.child,t.child=null):(l=a.sibling,a.sibling=null),Lc(t,!1,l,a,n,o);break;case"backwards":case"unstable_legacy-backwards":for(a=null,l=t.child,t.child=null;l!==null;){if(e=l.alternate,e!==null&&Fi(e)===null){t.child=l;break}e=l.sibling,l.sibling=a,a=l,l=e}Lc(t,!0,a,null,n,o);break;case"together":Lc(t,!1,null,null,void 0,o);break;default:t.memoizedState=null}return t.child}function Co(e,t,a){if(e!==null&&(t.dependencies=e.dependencies),rl|=t.lanes,(a&t.childLanes)===0)if(e!==null){if(Xn(e,t,a,!1),(a&t.childLanes)===0)return null}else return null;if(e!==null&&t.child!==e.child)throw Error(E(153));if(t.child!==null){for(e=t.child,a=go(e,e.pendingProps),t.child=a,a.return=t;e.sibling!==null;)e=e.sibling,a=a.sibling=go(e,e.pendingProps),a.return=t;a.sibling=null}return t.child}function sp(e,t){return(e.lanes&t)!==0?!0:(e=e.dependencies,!!(e!==null&&ki(e)))}function vw(e,t,a){switch(t.tag){case 3:Mi(t,t.stateNode.containerInfo),zo(t,Fe,e.memoizedState.cache),Dl();break;case 27:case 5:kc(t);break;case 4:Mi(t,t.stateNode.containerInfo);break;case 10:zo(t,t.type,t.memoizedProps.value);break;case 31:if(t.memoizedState!==null)return t.flags|=128,od(t),null;break;case 13:var o=t.memoizedState;if(o!==null)return o.dehydrated!==null?(Fo(t),t.flags|=128,null):(a&t.child.childLanes)!==0?rx(e,t,a):(Fo(t),e=Co(e,t,a),e!==null?e.sibling:null);Fo(t);break;case 19:var l=(e.flags&128)!==0;if(o=(a&t.childLanes)!==0,o||(Xn(e,t,a,!1),o=(a&t.childLanes)!==0),l){if(o)return ix(e,t,a);t.flags|=128}if(l=t.memoizedState,l!==null&&(l.rendering=null,l.tail=null,l.lastEffect=null),xe(Ne,Ne.current),o)break;return null;case 22:return t.lanes=0,ux(e,t,a,t.pendingProps);case 24:zo(t,Fe,e.memoizedState.cache)}return Co(e,t,a)}function sx(e,t,a){if(e!==null)if(e.memoizedProps!==t.pendingProps)qe=!0;else{if(!sp(e,a)&&(t.flags&128)===0)return qe=!1,vw(e,t,a);qe=(e.flags&131072)!==0}else qe=!1,oe&&(t.flags&1048576)!==0&&dS(t,Zu,t.index);switch(t.lanes=0,t.tag){case 16:e:{var o=t.pendingProps;if(e=Rl(t.elementType),t.type=e,typeof e=="function")Fd(e)?(o=Bl(e,o),t.tag=1,t=Vg(null,t,e,o,a)):(t.tag=0,t=sd(null,t,e,o,a));else{if(e!=null){var l=e.$$typeof;if(l===Ad){t.tag=11,t=Hg(null,t,e,o,a);break e}else if(l===Td){t.tag=14,t=zg(null,t,e,o,a);break e}}throw t=Uc(e)||e,Error(E(306,t,""))}}return t;case 0:return sd(e,t,t.type,t.pendingProps,a);case 1:return o=t.type,l=Bl(o,t.pendingProps),Vg(e,t,o,l,a);case 3:e:{if(Mi(t,t.stateNode.containerInfo),e===null)throw Error(E(387));o=t.pendingProps;var n=t.memoizedState;l=n.element,ed(e,t),ku(t,o,null,a);var u=t.memoizedState;if(o=u.cache,zo(t,Fe,o),o!==n.cache&&Jc(t,[Fe],a,!0),Nu(),o=u.element,n.isDehydrated)if(n={element:o,isDehydrated:!1,cache:u.cache},t.updateQueue.baseState=n,t.memoizedState=n,t.flags&256){t=Xg(e,t,o,a);break e}else if(o!==l){l=pa(Error(E(424)),t),Qu(l),t=Xg(e,t,o,a);break e}else for(e=t.stateNode.containerInfo,e.nodeType===9?e=e.body:e=e.nodeName==="HTML"?e.ownerDocument.body:e,Re=ga(e.firstChild),ht=t,oe=!0,Qo=null,ma=!0,a=SS(t,null,o,a),t.child=a;a;)a.flags=a.flags&-3|4096,a=a.sibling;else{if(Dl(),o===l){t=Co(e,t,a);break e}pt(e,t,o,a)}t=t.child}return t;case 26:return Ii(e,t),e===null?(a=mv(t.type,null,t.pendingProps,null))?t.memoizedState=a:oe||(a=t.type,e=t.pendingProps,o=$i(Zo.current).createElement(a),o[mt]=t,o[Ht]=e,vt(o,a,e),lt(o),t.stateNode=o):t.memoizedState=mv(t.type,e.memoizedProps,t.pendingProps,e.memoizedState),null;case 27:return kc(t),e===null&&oe&&(o=t.stateNode=$x(t.type,t.pendingProps,Zo.current),ht=t,ma=!0,l=Re,sl(t.type)?(Id=l,Re=ga(o.firstChild)):Re=l),pt(e,t,t.pendingProps.children,a),Ii(e,t),e===null&&(t.flags|=4194304),t.child;case 5:return e===null&&oe&&((l=o=Re)&&(o=Xw(o,t.type,t.pendingProps,ma),o!==null?(t.stateNode=o,ht=t,Re=ga(o.firstChild),ma=!1,l=!0):l=!1),l||nl(t)),kc(t),l=t.type,n=t.pendingProps,u=e!==null?e.memoizedProps:null,o=n.children,Ld(l,n)?o=null:u!==null&&Ld(l,u)&&(t.flags|=32),t.memoizedState!==null&&(l=Qd(e,t,iw,null,null,a),ar._currentValue=l),Ii(e,t),pt(e,t,o,a),t.child;case 6:return e===null&&oe&&((e=a=Re)&&(a=jw(a,t.pendingProps,ma),a!==null?(t.stateNode=a,ht=t,Re=null,e=!0):e=!1),e||nl(t)),null;case 13:return rx(e,t,a);case 4:return Mi(t,t.stateNode.containerInfo),o=t.pendingProps,e===null?t.child=Ol(t,null,o,a):pt(e,t,o,a),t.child;case 11:return Hg(e,t,t.type,t.pendingProps,a);case 7:return pt(e,t,t.pendingProps,a),t.child;case 8:return pt(e,t,t.pendingProps.children,a),t.child;case 12:return pt(e,t,t.pendingProps.children,a),t.child;case 10:return o=t.pendingProps,zo(t,t.type,o.value),pt(e,t,o.children,a),t.child;case 9:return l=t.type._context,o=t.pendingProps.children,Ml(t),l=gt(l),o=o(l),t.flags|=1,pt(e,t,o,a),t.child;case 14:return zg(e,t,t.type,t.pendingProps,a);case 15:return nx(e,t,t.type,t.pendingProps,a);case 19:return ix(e,t,a);case 31:return gw(e,t,a);case 22:return ux(e,t,a,t.pendingProps);case 24:return Ml(t),o=gt(Fe),e===null?(l=Xd(),l===null&&(l=Se,n=Vd(),l.pooledCache=n,n.refCount++,n!==null&&(l.pooledCacheLanes|=a),l=n),t.memoizedState={parent:o,cache:l},Yd(t),zo(t,Fe,l)):((e.lanes&a)!==0&&(ed(e,t),ku(t,null,null,a),Nu()),l=e.memoizedState,n=t.memoizedState,l.parent!==o?(l={parent:o,cache:o},t.memoizedState=l,t.lanes===0&&(t.memoizedState=t.updateQueue.baseState=l),zo(t,Fe,o)):(o=n.cache,zo(t,Fe,o),o!==l.cache&&Jc(t,[Fe],a,!0))),pt(e,t,t.pendingProps.children,a),t.child;case 29:throw t.pendingProps}throw Error(E(156,t.tag))}function uo(e){e.flags|=4}function yc(e,t,a,o,l){if((t=(e.mode&32)!==0)&&(t=!1),t){if(e.flags|=16777216,(l&335544128)===l)if(e.stateNode.complete)e.flags|=8192;else if(Px())e.flags|=8192;else throw Tl=Hi,jd}else e.flags&=-16777217}function Yg(e,t){if(t.type!=="stylesheet"||(t.state.loading&4)!==0)e.flags&=-16777217;else if(e.flags|=16777216,!aL(t))if(Px())e.flags|=8192;else throw Tl=Hi,jd}function ii(e,t){t!==null&&(e.flags|=4),e.flags&16384&&(t=e.tag!==22?Bv():536870912,e.lanes|=t,kn|=t)}function Ru(e,t){if(!oe)switch(e.tailMode){case"hidden":t=e.tail;for(var a=null;t!==null;)t.alternate!==null&&(a=t),t=t.sibling;a===null?e.tail=null:a.sibling=null;break;case"collapsed":a=e.tail;for(var o=null;a!==null;)a.alternate!==null&&(o=a),a=a.sibling;o===null?t||e.tail===null?e.tail=null:e.tail.sibling=null:o.sibling=null}}function Ce(e){var t=e.alternate!==null&&e.alternate.child===e.child,a=0,o=0;if(t)for(var l=e.child;l!==null;)a|=l.lanes|l.childLanes,o|=l.subtreeFlags&65011712,o|=l.flags&65011712,l.return=e,l=l.sibling;else for(l=e.child;l!==null;)a|=l.lanes|l.childLanes,o|=l.subtreeFlags,o|=l.flags,l.return=e,l=l.sibling;return e.subtreeFlags|=o,e.childLanes=a,t}function Sw(e,t,a){var o=t.pendingProps;switch(Gd(t),t.tag){case 16:case 15:case 0:case 11:case 7:case 8:case 12:case 9:case 14:return Ce(t),null;case 1:return Ce(t),null;case 3:return a=t.stateNode,o=null,e!==null&&(o=e.memoizedState.cache),t.memoizedState.cache!==o&&(t.flags|=2048),vo(Fe),On(),a.pendingContext&&(a.context=a.pendingContext,a.pendingContext=null),(e===null||e.child===null)&&(sn(t)?uo(t):e===null||e.memoizedState.isDehydrated&&(t.flags&256)===0||(t.flags|=1024,dc())),Ce(t),null;case 26:var l=t.type,n=t.memoizedState;return e===null?(uo(t),n!==null?(Ce(t),Yg(t,n)):(Ce(t),yc(t,l,null,o,a))):n?n!==e.memoizedState?(uo(t),Ce(t),Yg(t,n)):(Ce(t),t.flags&=-16777217):(e=e.memoizedProps,e!==o&&uo(t),Ce(t),yc(t,l,e,o,a)),null;case 27:if(Oi(t),a=Zo.current,l=t.type,e!==null&&t.stateNode!=null)e.memoizedProps!==o&&uo(t);else{if(!o){if(t.stateNode===null)throw Error(E(166));return Ce(t),null}e=Ka.current,sn(t)?Cg(t,e):(e=$x(l,o,a),t.stateNode=e,uo(t))}return Ce(t),null;case 5:if(Oi(t),l=t.type,e!==null&&t.stateNode!=null)e.memoizedProps!==o&&uo(t);else{if(!o){if(t.stateNode===null)throw Error(E(166));return Ce(t),null}if(n=Ka.current,sn(t))Cg(t,n);else{var u=$i(Zo.current);switch(n){case 1:n=u.createElementNS("http://www.w3.org/2000/svg",l);break;case 2:n=u.createElementNS("http://www.w3.org/1998/Math/MathML",l);break;default:switch(l){case"svg":n=u.createElementNS("http://www.w3.org/2000/svg",l);break;case"math":n=u.createElementNS("http://www.w3.org/1998/Math/MathML",l);break;case"script":n=u.createElement("div"),n.innerHTML="<script><\/script>",n=n.removeChild(n.firstChild);break;case"select":n=typeof o.is=="string"?u.createElement("select",{is:o.is}):u.createElement("select"),o.multiple?n.multiple=!0:o.size&&(n.size=o.size);break;default:n=typeof o.is=="string"?u.createElement(l,{is:o.is}):u.createElement(l)}}n[mt]=t,n[Ht]=o;e:for(u=t.child;u!==null;){if(u.tag===5||u.tag===6)n.appendChild(u.stateNode);else if(u.tag!==4&&u.tag!==27&&u.child!==null){u.child.return=u,u=u.child;continue}if(u===t)break e;for(;u.sibling===null;){if(u.return===null||u.return===t)break e;u=u.return}u.sibling.return=u.return,u=u.sibling}t.stateNode=n;e:switch(vt(n,l,o),l){case"button":case"input":case"select":case"textarea":o=!!o.autoFocus;break e;case"img":o=!0;break e;default:o=!1}o&&uo(t)}}return Ce(t),yc(t,t.type,e===null?null:e.memoizedProps,t.pendingProps,a),null;case 6:if(e&&t.stateNode!=null)e.memoizedProps!==o&&uo(t);else{if(typeof o!="string"&&t.stateNode===null)throw Error(E(166));if(e=Zo.current,sn(t)){if(e=t.stateNode,a=t.memoizedProps,o=null,l=ht,l!==null)switch(l.tag){case 27:case 5:o=l.memoizedProps}e[mt]=t,e=!!(e.nodeValue===a||o!==null&&o.suppressHydrationWarning===!0||Zx(e.nodeValue,a)),e||nl(t,!0)}else e=$i(e).createTextNode(o),e[mt]=t,t.stateNode=e}return Ce(t),null;case 31:if(a=t.memoizedState,e===null||e.memoizedState!==null){if(o=sn(t),a!==null){if(e===null){if(!o)throw Error(E(318));if(e=t.memoizedState,e=e!==null?e.dehydrated:null,!e)throw Error(E(557));e[mt]=t}else Dl(),(t.flags&128)===0&&(t.memoizedState=null),t.flags|=4;Ce(t),e=!1}else a=dc(),e!==null&&e.memoizedState!==null&&(e.memoizedState.hydrationErrors=a),e=!0;if(!e)return t.flags&256?(Kt(t),t):(Kt(t),null);if((t.flags&128)!==0)throw Error(E(558))}return Ce(t),null;case 13:if(o=t.memoizedState,e===null||e.memoizedState!==null&&e.memoizedState.dehydrated!==null){if(l=sn(t),o!==null&&o.dehydrated!==null){if(e===null){if(!l)throw Error(E(318));if(l=t.memoizedState,l=l!==null?l.dehydrated:null,!l)throw Error(E(317));l[mt]=t}else Dl(),(t.flags&128)===0&&(t.memoizedState=null),t.flags|=4;Ce(t),l=!1}else l=dc(),e!==null&&e.memoizedState!==null&&(e.memoizedState.hydrationErrors=l),l=!0;if(!l)return t.flags&256?(Kt(t),t):(Kt(t),null)}return Kt(t),(t.flags&128)!==0?(t.lanes=a,t):(a=o!==null,e=e!==null&&e.memoizedState!==null,a&&(o=t.child,l=null,o.alternate!==null&&o.alternate.memoizedState!==null&&o.alternate.memoizedState.cachePool!==null&&(l=o.alternate.memoizedState.cachePool.pool),n=null,o.memoizedState!==null&&o.memoizedState.cachePool!==null&&(n=o.memoizedState.cachePool.pool),n!==l&&(o.flags|=2048)),a!==e&&a&&(t.child.flags|=8192),ii(t,t.updateQueue),Ce(t),null);case 4:return On(),e===null&&gp(t.stateNode.containerInfo),Ce(t),null;case 10:return vo(t.type),Ce(t),null;case 19:if(nt(Ne),o=t.memoizedState,o===null)return Ce(t),null;if(l=(t.flags&128)!==0,n=o.rendering,n===null)if(l)Ru(o,!1);else{if(Pe!==0||e!==null&&(e.flags&128)!==0)for(e=t.child;e!==null;){if(n=Fi(e),n!==null){for(t.flags|=128,Ru(o,!1),e=n.updateQueue,t.updateQueue=e,ii(t,e),t.subtreeFlags=0,e=a,a=t.child;a!==null;)fS(a,e),a=a.sibling;return xe(Ne,Ne.current&1|2),oe&&fo(t,o.treeForkCount),t.child}e=e.sibling}o.tail!==null&&Wt()>Yi&&(t.flags|=128,l=!0,Ru(o,!1),t.lanes=4194304)}else{if(!l)if(e=Fi(n),e!==null){if(t.flags|=128,l=!0,e=e.updateQueue,t.updateQueue=e,ii(t,e),Ru(o,!0),o.tail===null&&o.tailMode==="hidden"&&!n.alternate&&!oe)return Ce(t),null}else 2*Wt()-o.renderingStartTime>Yi&&a!==536870912&&(t.flags|=128,l=!0,Ru(o,!1),t.lanes=4194304);o.isBackwards?(n.sibling=t.child,t.child=n):(e=o.last,e!==null?e.sibling=n:t.child=n,o.last=n)}return o.tail!==null?(e=o.tail,o.rendering=e,o.tail=e.sibling,o.renderingStartTime=Wt(),e.sibling=null,a=Ne.current,xe(Ne,l?a&1|2:a&1),oe&&fo(t,o.treeForkCount),e):(Ce(t),null);case 22:case 23:return Kt(t),Kd(),o=t.memoizedState!==null,e!==null?e.memoizedState!==null!==o&&(t.flags|=8192):o&&(t.flags|=8192),o?(a&536870912)!==0&&(t.flags&128)===0&&(Ce(t),t.subtreeFlags&6&&(t.flags|=8192)):Ce(t),a=t.updateQueue,a!==null&&ii(t,a.retryQueue),a=null,e!==null&&e.memoizedState!==null&&e.memoizedState.cachePool!==null&&(a=e.memoizedState.cachePool.pool),o=null,t.memoizedState!==null&&t.memoizedState.cachePool!==null&&(o=t.memoizedState.cachePool.pool),o!==a&&(t.flags|=2048),e!==null&&nt(Al),null;case 24:return a=null,e!==null&&(a=e.memoizedState.cache),t.memoizedState.cache!==a&&(t.flags|=2048),vo(Fe),Ce(t),null;case 25:return null;case 30:return null}throw Error(E(156,t.tag))}function xw(e,t){switch(Gd(t),t.tag){case 1:return e=t.flags,e&65536?(t.flags=e&-65537|128,t):null;case 3:return vo(Fe),On(),e=t.flags,(e&65536)!==0&&(e&128)===0?(t.flags=e&-65537|128,t):null;case 26:case 27:case 5:return Oi(t),null;case 31:if(t.memoizedState!==null){if(Kt(t),t.alternate===null)throw Error(E(340));Dl()}return e=t.flags,e&65536?(t.flags=e&-65537|128,t):null;case 13:if(Kt(t),e=t.memoizedState,e!==null&&e.dehydrated!==null){if(t.alternate===null)throw Error(E(340));Dl()}return e=t.flags,e&65536?(t.flags=e&-65537|128,t):null;case 19:return nt(Ne),null;case 4:return On(),null;case 10:return vo(t.type),null;case 22:case 23:return Kt(t),Kd(),e!==null&&nt(Al),e=t.flags,e&65536?(t.flags=e&-65537|128,t):null;case 24:return vo(Fe),null;case 25:return null;default:return null}}function fx(e,t){switch(Gd(t),t.tag){case 3:vo(Fe),On();break;case 26:case 27:case 5:Oi(t);break;case 4:On();break;case 31:t.memoizedState!==null&&Kt(t);break;case 13:Kt(t);break;case 19:nt(Ne);break;case 10:vo(t.type);break;case 22:case 23:Kt(t),Kd(),e!==null&&nt(Al);break;case 24:vo(Fe)}}function pr(e,t){try{var a=t.updateQueue,o=a!==null?a.lastEffect:null;if(o!==null){var l=o.next;a=l;do{if((a.tag&e)===e){o=void 0;var n=a.create,u=a.inst;o=n(),u.destroy=o}a=a.next}while(a!==l)}}catch(r){fe(t,t.return,r)}}function ul(e,t,a){try{var o=t.updateQueue,l=o!==null?o.lastEffect:null;if(l!==null){var n=l.next;o=n;do{if((o.tag&e)===e){var u=o.inst,r=u.destroy;if(r!==void 0){u.destroy=void 0,l=t;var i=a,s=r;try{s()}catch(f){fe(l,i,f)}}}o=o.next}while(o!==n)}}catch(f){fe(t,t.return,f)}}function cx(e){var t=e.updateQueue;if(t!==null){var a=e.stateNode;try{LS(t,a)}catch(o){fe(e,e.return,o)}}}function dx(e,t,a){a.props=Bl(e.type,e.memoizedProps),a.state=e.memoizedState;try{a.componentWillUnmount()}catch(o){fe(e,t,o)}}function zu(e,t){try{var a=e.ref;if(a!==null){switch(e.tag){case 26:case 27:case 5:var o=e.stateNode;break;case 30:o=e.stateNode;break;default:o=e.stateNode}typeof a=="function"?e.refCleanup=a(o):a.current=o}}catch(l){fe(e,t,l)}}function Ya(e,t){var a=e.ref,o=e.refCleanup;if(a!==null)if(typeof o=="function")try{o()}catch(l){fe(e,t,l)}finally{e.refCleanup=null,e=e.alternate,e!=null&&(e.refCleanup=null)}else if(typeof a=="function")try{a(null)}catch(l){fe(e,t,l)}else a.current=null}function px(e){var t=e.type,a=e.memoizedProps,o=e.stateNode;try{e:switch(t){case"button":case"input":case"select":case"textarea":a.autoFocus&&o.focus();break e;case"img":a.src?o.src=a.src:a.srcSet&&(o.srcset=a.srcSet)}}catch(l){fe(e,e.return,l)}}function Cc(e,t,a){try{var o=e.stateNode;Hw(o,e.type,a,t),o[Ht]=t}catch(l){fe(e,e.return,l)}}function mx(e){return e.tag===5||e.tag===3||e.tag===26||e.tag===27&&sl(e.type)||e.tag===4}function Rc(e){e:for(;;){for(;e.sibling===null;){if(e.return===null||mx(e.return))return null;e=e.return}for(e.sibling.return=e.return,e=e.sibling;e.tag!==5&&e.tag!==6&&e.tag!==18;){if(e.tag===27&&sl(e.type)||e.flags&2||e.child===null||e.tag===4)continue e;e.child.return=e,e=e.child}if(!(e.flags&2))return e.stateNode}}function cd(e,t,a){var o=e.tag;if(o===5||o===6)e=e.stateNode,t?(a.nodeType===9?a.body:a.nodeName==="HTML"?a.ownerDocument.body:a).insertBefore(e,t):(t=a.nodeType===9?a.body:a.nodeName==="HTML"?a.ownerDocument.body:a,t.appendChild(e),a=a._reactRootContainer,a!=null||t.onclick!==null||(t.onclick=mo));else if(o!==4&&(o===27&&sl(e.type)&&(a=e.stateNode,t=null),e=e.child,e!==null))for(cd(e,t,a),e=e.sibling;e!==null;)cd(e,t,a),e=e.sibling}function ji(e,t,a){var o=e.tag;if(o===5||o===6)e=e.stateNode,t?a.insertBefore(e,t):a.appendChild(e);else if(o!==4&&(o===27&&sl(e.type)&&(a=e.stateNode),e=e.child,e!==null))for(ji(e,t,a),e=e.sibling;e!==null;)ji(e,t,a),e=e.sibling}function hx(e){var t=e.stateNode,a=e.memoizedProps;try{for(var o=e.type,l=t.attributes;l.length;)t.removeAttributeNode(l[0]);vt(t,o,a),t[mt]=e,t[Ht]=a}catch(n){fe(e,e.return,n)}}var co=!1,ze=!1,Ic=!1,Kg=typeof WeakSet=="function"?WeakSet:Set,ot=null;function Lw(e,t){if(e=e.containerInfo,Sd=os,e=aS(e),kd(e)){if("selectionStart"in e)var a={start:e.selectionStart,end:e.selectionEnd};else e:{a=(a=e.ownerDocument)&&a.defaultView||window;var o=a.getSelection&&a.getSelection();if(o&&o.rangeCount!==0){a=o.anchorNode;var l=o.anchorOffset,n=o.focusNode;o=o.focusOffset;try{a.nodeType,n.nodeType}catch(x){a=null;break e}var u=0,r=-1,i=-1,s=0,f=0,d=e,c=null;t:for(;;){for(var g;d!==a||l!==0&&d.nodeType!==3||(r=u+l),d!==n||o!==0&&d.nodeType!==3||(i=u+o),d.nodeType===3&&(u+=d.nodeValue.length),(g=d.firstChild)!==null;)c=d,d=g;for(;;){if(d===e)break t;if(c===a&&++s===l&&(r=u),c===n&&++f===o&&(i=u),(g=d.nextSibling)!==null)break;d=c,c=d.parentNode}d=g}a=r===-1||i===-1?null:{start:r,end:i}}else a=null}a=a||{start:0,end:0}}else a=null;for(xd={focusedElem:e,selectionRange:a},os=!1,ot=t;ot!==null;)if(t=ot,e=t.child,(t.subtreeFlags&1028)!==0&&e!==null)e.return=t,ot=e;else for(;ot!==null;){switch(t=ot,n=t.alternate,e=t.flags,t.tag){case 0:if((e&4)!==0&&(e=t.updateQueue,e=e!==null?e.events:null,e!==null))for(a=0;a<e.length;a++)l=e[a],l.ref.impl=l.nextImpl;break;case 11:case 15:break;case 1:if((e&1024)!==0&&n!==null){e=void 0,a=t,l=n.memoizedProps,n=n.memoizedState,o=a.stateNode;try{var L=Bl(a.type,l);e=o.getSnapshotBeforeUpdate(L,n),o.__reactInternalSnapshotBeforeUpdate=e}catch(x){fe(a,a.return,x)}}break;case 3:if((e&1024)!==0){if(e=t.stateNode.containerInfo,a=e.nodeType,a===9)yd(e);else if(a===1)switch(e.nodeName){case"HEAD":case"HTML":case"BODY":yd(e);break;default:e.textContent=""}}break;case 5:case 26:case 27:case 6:case 4:case 17:break;default:if((e&1024)!==0)throw Error(E(163))}if(e=t.sibling,e!==null){e.return=t.return,ot=e;break}ot=t.return}}function gx(e,t,a){var o=a.flags;switch(a.tag){case 0:case 11:case 15:io(e,a),o&4&&pr(5,a);break;case 1:if(io(e,a),o&4)if(e=a.stateNode,t===null)try{e.componentDidMount()}catch(u){fe(a,a.return,u)}else{var l=Bl(a.type,t.memoizedProps);t=t.memoizedState;try{e.componentDidUpdate(l,t,e.__reactInternalSnapshotBeforeUpdate)}catch(u){fe(a,a.return,u)}}o&64&&cx(a),o&512&&zu(a,a.return);break;case 3:if(io(e,a),o&64&&(e=a.updateQueue,e!==null)){if(t=null,a.child!==null)switch(a.child.tag){case 27:case 5:t=a.child.stateNode;break;case 1:t=a.child.stateNode}try{LS(e,t)}catch(u){fe(a,a.return,u)}}break;case 27:t===null&&o&4&&hx(a);case 26:case 5:io(e,a),t===null&&o&4&&px(a),o&512&&zu(a,a.return);break;case 12:io(e,a);break;case 31:io(e,a),o&4&&xx(e,a);break;case 13:io(e,a),o&4&&Lx(e,a),o&64&&(e=a.memoizedState,e!==null&&(e=e.dehydrated,e!==null&&(a=Ew.bind(null,a),Yw(e,a))));break;case 22:if(o=a.memoizedState!==null||co,!o){t=t!==null&&t.memoizedState!==null||ze,l=co;var n=ze;co=o,(ze=t)&&!n?so(e,a,(a.subtreeFlags&8772)!==0):io(e,a),co=l,ze=n}break;case 30:break;default:io(e,a)}}function vx(e){var t=e.alternate;t!==null&&(e.alternate=null,vx(t)),e.child=null,e.deletions=null,e.sibling=null,e.tag===5&&(t=e.stateNode,t!==null&&Od(t)),e.stateNode=null,e.return=null,e.dependencies=null,e.memoizedProps=null,e.memoizedState=null,e.pendingProps=null,e.stateNode=null,e.updateQueue=null}var Te=null,Ut=!1;function ro(e,t,a){for(a=a.child;a!==null;)Sx(e,t,a),a=a.sibling}function Sx(e,t,a){if(Jt&&typeof Jt.onCommitFiberUnmount=="function")try{Jt.onCommitFiberUnmount(ur,a)}catch(n){}switch(a.tag){case 26:ze||Ya(a,t),ro(e,t,a),a.memoizedState?a.memoizedState.count--:a.stateNode&&(a=a.stateNode,a.parentNode.removeChild(a));break;case 27:ze||Ya(a,t);var o=Te,l=Ut;sl(a.type)&&(Te=a.stateNode,Ut=!1),ro(e,t,a),Vu(a.stateNode),Te=o,Ut=l;break;case 5:ze||Ya(a,t);case 6:if(o=Te,l=Ut,Te=null,ro(e,t,a),Te=o,Ut=l,Te!==null)if(Ut)try{(Te.nodeType===9?Te.body:Te.nodeName==="HTML"?Te.ownerDocument.body:Te).removeChild(a.stateNode)}catch(n){fe(a,t,n)}else try{Te.removeChild(a.stateNode)}catch(n){fe(a,t,n)}break;case 18:Te!==null&&(Ut?(e=Te,sv(e.nodeType===9?e.body:e.nodeName==="HTML"?e.ownerDocument.body:e,a.stateNode),qn(e)):sv(Te,a.stateNode));break;case 4:o=Te,l=Ut,Te=a.stateNode.containerInfo,Ut=!0,ro(e,t,a),Te=o,Ut=l;break;case 0:case 11:case 14:case 15:ul(2,a,t),ze||ul(4,a,t),ro(e,t,a);break;case 1:ze||(Ya(a,t),o=a.stateNode,typeof o.componentWillUnmount=="function"&&dx(a,t,o)),ro(e,t,a);break;case 21:ro(e,t,a);break;case 22:ze=(o=ze)||a.memoizedState!==null,ro(e,t,a),ze=o;break;default:ro(e,t,a)}}function xx(e,t){if(t.memoizedState===null&&(e=t.alternate,e!==null&&(e=e.memoizedState,e!==null))){e=e.dehydrated;try{qn(e)}catch(a){fe(t,t.return,a)}}}function Lx(e,t){if(t.memoizedState===null&&(e=t.alternate,e!==null&&(e=e.memoizedState,e!==null&&(e=e.dehydrated,e!==null))))try{qn(e)}catch(a){fe(t,t.return,a)}}function yw(e){switch(e.tag){case 31:case 13:case 19:var t=e.stateNode;return t===null&&(t=e.stateNode=new Kg),t;case 22:return e=e.stateNode,t=e._retryCache,t===null&&(t=e._retryCache=new Kg),t;default:throw Error(E(435,e.tag))}}function si(e,t){var a=yw(e);t.forEach(function(o){if(!a.has(o)){a.add(o);var l=Dw.bind(null,e,o);o.then(l,l)}})}function Bt(e,t){var a=t.deletions;if(a!==null)for(var o=0;o<a.length;o++){var l=a[o],n=e,u=t,r=u;e:for(;r!==null;){switch(r.tag){case 27:if(sl(r.type)){Te=r.stateNode,Ut=!1;break e}break;case 5:Te=r.stateNode,Ut=!1;break e;case 3:case 4:Te=r.stateNode.containerInfo,Ut=!0;break e}r=r.return}if(Te===null)throw Error(E(160));Sx(n,u,l),Te=null,Ut=!1,n=l.alternate,n!==null&&(n.return=null),l.return=null}if(t.subtreeFlags&13886)for(t=t.child;t!==null;)yx(t,e),t=t.sibling}var Ta=null;function yx(e,t){var a=e.alternate,o=e.flags;switch(e.tag){case 0:case 11:case 14:case 15:Bt(t,e),_t(e),o&4&&(ul(3,e,e.return),pr(3,e),ul(5,e,e.return));break;case 1:Bt(t,e),_t(e),o&512&&(ze||a===null||Ya(a,a.return)),o&64&&co&&(e=e.updateQueue,e!==null&&(o=e.callbacks,o!==null&&(a=e.shared.hiddenCallbacks,e.shared.hiddenCallbacks=a===null?o:a.concat(o))));break;case 26:var l=Ta;if(Bt(t,e),_t(e),o&512&&(ze||a===null||Ya(a,a.return)),o&4){var n=a!==null?a.memoizedState:null;if(o=e.memoizedState,a===null)if(o===null)if(e.stateNode===null){e:{o=e.type,a=e.memoizedProps,l=l.ownerDocument||l;t:switch(o){case"title":n=l.getElementsByTagName("title")[0],(!n||n[sr]||n[mt]||n.namespaceURI==="http://www.w3.org/2000/svg"||n.hasAttribute("itemprop"))&&(n=l.createElement(o),l.head.insertBefore(n,l.querySelector("head > title"))),vt(n,o,a),n[mt]=e,lt(n),o=n;break e;case"link":var u=gv("link","href",l).get(o+(a.href||""));if(u){for(var r=0;r<u.length;r++)if(n=u[r],n.getAttribute("href")===(a.href==null||a.href===""?null:a.href)&&n.getAttribute("rel")===(a.rel==null?null:a.rel)&&n.getAttribute("title")===(a.title==null?null:a.title)&&n.getAttribute("crossorigin")===(a.crossOrigin==null?null:a.crossOrigin)){u.splice(r,1);break t}}n=l.createElement(o),vt(n,o,a),l.head.appendChild(n);break;case"meta":if(u=gv("meta","content",l).get(o+(a.content||""))){for(r=0;r<u.length;r++)if(n=u[r],n.getAttribute("content")===(a.content==null?null:""+a.content)&&n.getAttribute("name")===(a.name==null?null:a.name)&&n.getAttribute("property")===(a.property==null?null:a.property)&&n.getAttribute("http-equiv")===(a.httpEquiv==null?null:a.httpEquiv)&&n.getAttribute("charset")===(a.charSet==null?null:a.charSet)){u.splice(r,1);break t}}n=l.createElement(o),vt(n,o,a),l.head.appendChild(n);break;default:throw Error(E(468,o))}n[mt]=e,lt(n),o=n}e.stateNode=o}else vv(l,e.type,e.stateNode);else e.stateNode=hv(l,o,e.memoizedProps);else n!==o?(n===null?a.stateNode!==null&&(a=a.stateNode,a.parentNode.removeChild(a)):n.count--,o===null?vv(l,e.type,e.stateNode):hv(l,o,e.memoizedProps)):o===null&&e.stateNode!==null&&Cc(e,e.memoizedProps,a.memoizedProps)}break;case 27:Bt(t,e),_t(e),o&512&&(ze||a===null||Ya(a,a.return)),a!==null&&o&4&&Cc(e,e.memoizedProps,a.memoizedProps);break;case 5:if(Bt(t,e),_t(e),o&512&&(ze||a===null||Ya(a,a.return)),e.flags&32){l=e.stateNode;try{Bn(l,"")}catch(L){fe(e,e.return,L)}}o&4&&e.stateNode!=null&&(l=e.memoizedProps,Cc(e,l,a!==null?a.memoizedProps:l)),o&1024&&(Ic=!0);break;case 6:if(Bt(t,e),_t(e),o&4){if(e.stateNode===null)throw Error(E(162));o=e.memoizedProps,a=e.stateNode;try{a.nodeValue=o}catch(L){fe(e,e.return,L)}}break;case 3:if(Ai=null,l=Ta,Ta=es(t.containerInfo),Bt(t,e),Ta=l,_t(e),o&4&&a!==null&&a.memoizedState.isDehydrated)try{qn(t.containerInfo)}catch(L){fe(e,e.return,L)}Ic&&(Ic=!1,Cx(e));break;case 4:o=Ta,Ta=es(e.stateNode.containerInfo),Bt(t,e),_t(e),Ta=o;break;case 12:Bt(t,e),_t(e);break;case 31:Bt(t,e),_t(e),o&4&&(o=e.updateQueue,o!==null&&(e.updateQueue=null,si(e,o)));break;case 13:Bt(t,e),_t(e),e.child.flags&8192&&e.memoizedState!==null!=(a!==null&&a.memoizedState!==null)&&(gs=Wt()),o&4&&(o=e.updateQueue,o!==null&&(e.updateQueue=null,si(e,o)));break;case 22:l=e.memoizedState!==null;var i=a!==null&&a.memoizedState!==null,s=co,f=ze;if(co=s||l,ze=f||i,Bt(t,e),ze=f,co=s,_t(e),o&8192)e:for(t=e.stateNode,t._visibility=l?t._visibility&-2:t._visibility|1,l&&(a===null||i||co||ze||Il(e)),a=null,t=e;;){if(t.tag===5||t.tag===26){if(a===null){i=a=t;try{if(n=i.stateNode,l)u=n.style,typeof u.setProperty=="function"?u.setProperty("display","none","important"):u.display="none";else{r=i.stateNode;var d=i.memoizedProps.style,c=d!=null&&d.hasOwnProperty("display")?d.display:null;r.style.display=c==null||typeof c=="boolean"?"":(""+c).trim()}}catch(L){fe(i,i.return,L)}}}else if(t.tag===6){if(a===null){i=t;try{i.stateNode.nodeValue=l?"":i.memoizedProps}catch(L){fe(i,i.return,L)}}}else if(t.tag===18){if(a===null){i=t;try{var g=i.stateNode;l?fv(g,!0):fv(i.stateNode,!1)}catch(L){fe(i,i.return,L)}}}else if((t.tag!==22&&t.tag!==23||t.memoizedState===null||t===e)&&t.child!==null){t.child.return=t,t=t.child;continue}if(t===e)break e;for(;t.sibling===null;){if(t.return===null||t.return===e)break e;a===t&&(a=null),t=t.return}a===t&&(a=null),t.sibling.return=t.return,t=t.sibling}o&4&&(o=e.updateQueue,o!==null&&(a=o.retryQueue,a!==null&&(o.retryQueue=null,si(e,a))));break;case 19:Bt(t,e),_t(e),o&4&&(o=e.updateQueue,o!==null&&(e.updateQueue=null,si(e,o)));break;case 30:break;case 21:break;default:Bt(t,e),_t(e)}}function _t(e){var t=e.flags;if(t&2){try{for(var a,o=e.return;o!==null;){if(mx(o)){a=o;break}o=o.return}if(a==null)throw Error(E(160));switch(a.tag){case 27:var l=a.stateNode,n=Rc(e);ji(e,n,l);break;case 5:var u=a.stateNode;a.flags&32&&(Bn(u,""),a.flags&=-33);var r=Rc(e);ji(e,r,u);break;case 3:case 4:var i=a.stateNode.containerInfo,s=Rc(e);cd(e,s,i);break;default:throw Error(E(161))}}catch(f){fe(e,e.return,f)}e.flags&=-3}t&4096&&(e.flags&=-4097)}function Cx(e){if(e.subtreeFlags&1024)for(e=e.child;e!==null;){var t=e;Cx(t),t.tag===5&&t.flags&1024&&t.stateNode.reset(),e=e.sibling}}function io(e,t){if(t.subtreeFlags&8772)for(t=t.child;t!==null;)gx(e,t.alternate,t),t=t.sibling}function Il(e){for(e=e.child;e!==null;){var t=e;switch(t.tag){case 0:case 11:case 14:case 15:ul(4,t,t.return),Il(t);break;case 1:Ya(t,t.return);var a=t.stateNode;typeof a.componentWillUnmount=="function"&&dx(t,t.return,a),Il(t);break;case 27:Vu(t.stateNode);case 26:case 5:Ya(t,t.return),Il(t);break;case 22:t.memoizedState===null&&Il(t);break;case 30:Il(t);break;default:Il(t)}e=e.sibling}}function so(e,t,a){for(a=a&&(t.subtreeFlags&8772)!==0,t=t.child;t!==null;){var o=t.alternate,l=e,n=t,u=n.flags;switch(n.tag){case 0:case 11:case 15:so(l,n,a),pr(4,n);break;case 1:if(so(l,n,a),o=n,l=o.stateNode,typeof l.componentDidMount=="function")try{l.componentDidMount()}catch(s){fe(o,o.return,s)}if(o=n,l=o.updateQueue,l!==null){var r=o.stateNode;try{var i=l.shared.hiddenCallbacks;if(i!==null)for(l.shared.hiddenCallbacks=null,l=0;l<i.length;l++)xS(i[l],r)}catch(s){fe(o,o.return,s)}}a&&u&64&&cx(n),zu(n,n.return);break;case 27:hx(n);case 26:case 5:so(l,n,a),a&&o===null&&u&4&&px(n),zu(n,n.return);break;case 12:so(l,n,a);break;case 31:so(l,n,a),a&&u&4&&xx(l,n);break;case 13:so(l,n,a),a&&u&4&&Lx(l,n);break;case 22:n.memoizedState===null&&so(l,n,a),zu(n,n.return);break;case 30:break;default:so(l,n,a)}t=t.sibling}}function fp(e,t){var a=null;e!==null&&e.memoizedState!==null&&e.memoizedState.cachePool!==null&&(a=e.memoizedState.cachePool.pool),e=null,t.memoizedState!==null&&t.memoizedState.cachePool!==null&&(e=t.memoizedState.cachePool.pool),e!==a&&(e!=null&&e.refCount++,a!=null&&cr(a))}function cp(e,t){e=null,t.alternate!==null&&(e=t.alternate.memoizedState.cache),t=t.memoizedState.cache,t!==e&&(t.refCount++,e!=null&&cr(e))}function Aa(e,t,a,o){if(t.subtreeFlags&10256)for(t=t.child;t!==null;)Rx(e,t,a,o),t=t.sibling}function Rx(e,t,a,o){var l=t.flags;switch(t.tag){case 0:case 11:case 15:Aa(e,t,a,o),l&2048&&pr(9,t);break;case 1:Aa(e,t,a,o);break;case 3:Aa(e,t,a,o),l&2048&&(e=null,t.alternate!==null&&(e=t.alternate.memoizedState.cache),t=t.memoizedState.cache,t!==e&&(t.refCount++,e!=null&&cr(e)));break;case 12:if(l&2048){Aa(e,t,a,o),e=t.stateNode;try{var n=t.memoizedProps,u=n.id,r=n.onPostCommit;typeof r=="function"&&r(u,t.alternate===null?"mount":"update",e.passiveEffectDuration,-0)}catch(i){fe(t,t.return,i)}}else Aa(e,t,a,o);break;case 31:Aa(e,t,a,o);break;case 13:Aa(e,t,a,o);break;case 23:break;case 22:n=t.stateNode,u=t.alternate,t.memoizedState!==null?n._visibility&2?Aa(e,t,a,o):Fu(e,t):n._visibility&2?Aa(e,t,a,o):(n._visibility|=2,cn(e,t,a,o,(t.subtreeFlags&10256)!==0||!1)),l&2048&&fp(u,t);break;case 24:Aa(e,t,a,o),l&2048&&cp(t.alternate,t);break;default:Aa(e,t,a,o)}}function cn(e,t,a,o,l){for(l=l&&((t.subtreeFlags&10256)!==0||!1),t=t.child;t!==null;){var n=e,u=t,r=a,i=o,s=u.flags;switch(u.tag){case 0:case 11:case 15:cn(n,u,r,i,l),pr(8,u);break;case 23:break;case 22:var f=u.stateNode;u.memoizedState!==null?f._visibility&2?cn(n,u,r,i,l):Fu(n,u):(f._visibility|=2,cn(n,u,r,i,l)),l&&s&2048&&fp(u.alternate,u);break;case 24:cn(n,u,r,i,l),l&&s&2048&&cp(u.alternate,u);break;default:cn(n,u,r,i,l)}t=t.sibling}}function Fu(e,t){if(t.subtreeFlags&10256)for(t=t.child;t!==null;){var a=e,o=t,l=o.flags;switch(o.tag){case 22:Fu(a,o),l&2048&&fp(o.alternate,o);break;case 24:Fu(a,o),l&2048&&cp(o.alternate,o);break;default:Fu(a,o)}t=t.sibling}}var Mu=8192;function fn(e,t,a){if(e.subtreeFlags&Mu)for(e=e.child;e!==null;)Ix(e,t,a),e=e.sibling}function Ix(e,t,a){switch(e.tag){case 26:fn(e,t,a),e.flags&Mu&&e.memoizedState!==null&&nA(a,Ta,e.memoizedState,e.memoizedProps);break;case 5:fn(e,t,a);break;case 3:case 4:var o=Ta;Ta=es(e.stateNode.containerInfo),fn(e,t,a),Ta=o;break;case 22:e.memoizedState===null&&(o=e.alternate,o!==null&&o.memoizedState!==null?(o=Mu,Mu=16777216,fn(e,t,a),Mu=o):fn(e,t,a));break;default:fn(e,t,a)}}function bx(e){var t=e.alternate;if(t!==null&&(e=t.child,e!==null)){t.child=null;do t=e.sibling,e.sibling=null,e=t;while(e!==null)}}function Iu(e){var t=e.deletions;if((e.flags&16)!==0){if(t!==null)for(var a=0;a<t.length;a++){var o=t[a];ot=o,Ax(o,e)}bx(e)}if(e.subtreeFlags&10256)for(e=e.child;e!==null;)wx(e),e=e.sibling}function wx(e){switch(e.tag){case 0:case 11:case 15:Iu(e),e.flags&2048&&ul(9,e,e.return);break;case 3:Iu(e);break;case 12:Iu(e);break;case 22:var t=e.stateNode;e.memoizedState!==null&&t._visibility&2&&(e.return===null||e.return.tag!==13)?(t._visibility&=-3,bi(e)):Iu(e);break;default:Iu(e)}}function bi(e){var t=e.deletions;if((e.flags&16)!==0){if(t!==null)for(var a=0;a<t.length;a++){var o=t[a];ot=o,Ax(o,e)}bx(e)}for(e=e.child;e!==null;){switch(t=e,t.tag){case 0:case 11:case 15:ul(8,t,t.return),bi(t);break;case 22:a=t.stateNode,a._visibility&2&&(a._visibility&=-3,bi(t));break;default:bi(t)}e=e.sibling}}function Ax(e,t){for(;ot!==null;){var a=ot;switch(a.tag){case 0:case 11:case 15:ul(8,a,t);break;case 23:case 22:if(a.memoizedState!==null&&a.memoizedState.cachePool!==null){var o=a.memoizedState.cachePool.pool;o!=null&&o.refCount++}break;case 24:cr(a.memoizedState.cache)}if(o=a.child,o!==null)o.return=a,ot=o;else e:for(a=e;ot!==null;){o=ot;var l=o.sibling,n=o.return;if(vx(o),o===a){ot=null;break e}if(l!==null){l.return=n,ot=l;break e}ot=n}}}var Cw={getCacheForType:function(e){var t=gt(Fe),a=t.data.get(e);return a===void 0&&(a=e(),t.data.set(e,a)),a},cacheSignal:function(){return gt(Fe).controller.signal}},Rw=typeof WeakMap=="function"?WeakMap:Map,re=0,Se=null,te=null,ae=0,se=0,Yt=null,jo=!1,Yn=!1,dp=!1,Ro=0,Pe=0,rl=0,El=0,pp=0,Qt=0,kn=0,qu=null,Nt=null,dd=!1,gs=0,Tx=0,Yi=1/0,Ki=null,$o=null,Ke=0,el=null,Hn=null,So=0,pd=0,md=null,Ex=null,Gu=0,hd=null;function ea(){return(re&2)!==0&&ae!==0?ae&-ae:V.T!==null?hp():kv()}function Dx(){if(Qt===0)if((ae&536870912)===0||oe){var e=$r;$r<<=1,($r&3932160)===0&&($r=262144),Qt=e}else Qt=536870912;return e=aa.current,e!==null&&(e.flags|=32),Qt}function kt(e,t,a){(e===Se&&(se===2||se===9)||e.cancelPendingCommit!==null)&&(zn(e,0),Yo(e,ae,Qt,!1)),ir(e,a),((re&2)===0||e!==Se)&&(e===Se&&((re&2)===0&&(El|=a),Pe===4&&Yo(e,ae,Qt,!1)),Qa(e))}function Mx(e,t,a){if((re&6)!==0)throw Error(E(327));var o=!a&&(t&127)===0&&(t&e.expiredLanes)===0||rr(e,t),l=o?ww(e,t):bc(e,t,!0),n=o;do{if(l===0){Yn&&!o&&Yo(e,t,0,!1);break}else{if(a=e.current.alternate,n&&!Iw(a)){l=bc(e,t,!1),n=!1;continue}if(l===2){if(n=t,e.errorRecoveryDisabledLanes&n)var u=0;else u=e.pendingLanes&-536870913,u=u!==0?u:u&536870912?536870912:0;if(u!==0){t=u;e:{var r=e;l=qu;var i=r.current.memoizedState.isDehydrated;if(i&&(zn(r,u).flags|=256),u=bc(r,u,!1),u!==2){if(dp&&!i){r.errorRecoveryDisabledLanes|=n,El|=n,l=4;break e}n=Nt,Nt=l,n!==null&&(Nt===null?Nt=n:Nt.push.apply(Nt,n))}l=u}if(n=!1,l!==2)continue}}if(l===1){zn(e,0),Yo(e,t,0,!0);break}e:{switch(o=e,n=l,n){case 0:case 1:throw Error(E(345));case 4:if((t&4194048)!==t)break;case 6:Yo(o,t,Qt,!jo);break e;case 2:Nt=null;break;case 3:case 5:break;default:throw Error(E(329))}if((t&62914560)===t&&(l=gs+300-Wt(),10<l)){if(Yo(o,t,Qt,!jo),ns(o,0,!0)!==0)break e;So=t,o.timeoutHandle=Wx(Zg.bind(null,o,a,Nt,Ki,dd,t,Qt,El,kn,jo,n,"Throttled",-0,0),l);break e}Zg(o,a,Nt,Ki,dd,t,Qt,El,kn,jo,n,null,-0,0)}}break}while(!0);Qa(e)}function Zg(e,t,a,o,l,n,u,r,i,s,f,d,c,g){if(e.timeoutHandle=-1,d=t.subtreeFlags,d&8192||(d&16785408)===16785408){d={stylesheets:null,count:0,imgCount:0,imgBytes:0,suspenseyImages:[],waitingForImages:!0,waitingForViewTransition:!1,unsuspend:mo},Ix(t,n,d);var L=(n&62914560)===n?gs-Wt():(n&4194048)===n?Tx-Wt():0;if(L=uA(d,L),L!==null){So=n,e.cancelPendingCommit=L(Wg.bind(null,e,t,n,a,o,l,u,r,i,f,d,null,c,g)),Yo(e,n,u,!s);return}}Wg(e,t,n,a,o,l,u,r,i)}function Iw(e){for(var t=e;;){var a=t.tag;if((a===0||a===11||a===15)&&t.flags&16384&&(a=t.updateQueue,a!==null&&(a=a.stores,a!==null)))for(var o=0;o<a.length;o++){var l=a[o],n=l.getSnapshot;l=l.value;try{if(!ta(n(),l))return!1}catch(u){return!1}}if(a=t.child,t.subtreeFlags&16384&&a!==null)a.return=t,t=a;else{if(t===e)break;for(;t.sibling===null;){if(t.return===null||t.return===e)return!0;t=t.return}t.sibling.return=t.return,t=t.sibling}}return!0}function Yo(e,t,a,o){t&=~pp,t&=~El,e.suspendedLanes|=t,e.pingedLanes&=~t,o&&(e.warmLanes|=t),o=e.expirationTimes;for(var l=t;0<l;){var n=31-$t(l),u=1<<n;o[n]=-1,l&=~u}a!==0&&_v(e,a,t)}function vs(){return(re&6)===0?(mr(0,!1),!1):!0}function mp(){if(te!==null){if(se===0)var e=te.return;else e=te,ho=Hl=null,$d(e),En=null,Wu=0,e=te;for(;e!==null;)fx(e.alternate,e),e=e.return;te=null}}function zn(e,t){var a=e.timeoutHandle;a!==-1&&(e.timeoutHandle=-1,qw(a)),a=e.cancelPendingCommit,a!==null&&(e.cancelPendingCommit=null,a()),So=0,mp(),Se=e,te=a=go(e.current,null),ae=t,se=0,Yt=null,jo=!1,Yn=rr(e,t),dp=!1,kn=Qt=pp=El=rl=Pe=0,Nt=qu=null,dd=!1,(t&8)!==0&&(t|=t&32);var o=e.entangledLanes;if(o!==0)for(e=e.entanglements,o&=t;0<o;){var l=31-$t(o),n=1<<l;t|=e[l],o&=~n}return Ro=t,ss(),a}function Ox(e,t){K=null,V.H=$u,t===jn||t===cs?(t=Ag(),se=3):t===jd?(t=Ag(),se=4):se=t===ip?8:t!==null&&typeof t=="object"&&typeof t.then=="function"?6:1,Yt=t,te===null&&(Pe=1,Vi(e,pa(t,e.current)))}function Px(){var e=aa.current;return e===null?!0:(ae&4194048)===ae?ha===null:(ae&62914560)===ae||(ae&536870912)!==0?e===ha:!1}function Bx(){var e=V.H;return V.H=$u,e===null?$u:e}function _x(){var e=V.A;return V.A=Cw,e}function Zi(){Pe=4,jo||(ae&4194048)!==ae&&aa.current!==null||(Yn=!0),(rl&134217727)===0&&(El&134217727)===0||Se===null||Yo(Se,ae,Qt,!1)}function bc(e,t,a){var o=re;re|=2;var l=Bx(),n=_x();(Se!==e||ae!==t)&&(Ki=null,zn(e,t)),t=!1;var u=Pe;e:do try{if(se!==0&&te!==null){var r=te,i=Yt;switch(se){case 8:mp(),u=6;break e;case 3:case 2:case 9:case 6:aa.current===null&&(t=!0);var s=se;if(se=0,Yt=null,In(e,r,i,s),a&&Yn){u=0;break e}break;default:s=se,se=0,Yt=null,In(e,r,i,s)}}bw(),u=Pe;break}catch(f){Ox(e,f)}while(!0);return t&&e.shellSuspendCounter++,ho=Hl=null,re=o,V.H=l,V.A=n,te===null&&(Se=null,ae=0,ss()),u}function bw(){for(;te!==null;)Ux(te)}function ww(e,t){var a=re;re|=2;var o=Bx(),l=_x();Se!==e||ae!==t?(Ki=null,Yi=Wt()+500,zn(e,t)):Yn=rr(e,t);e:do try{if(se!==0&&te!==null){t=te;var n=Yt;t:switch(se){case 1:se=0,Yt=null,In(e,t,n,1);break;case 2:case 9:if(wg(n)){se=0,Yt=null,Qg(t);break}t=function(){se!==2&&se!==9||Se!==e||(se=7),Qa(e)},n.then(t,t);break e;case 3:se=7;break e;case 4:se=5;break e;case 7:wg(n)?(se=0,Yt=null,Qg(t)):(se=0,Yt=null,In(e,t,n,7));break;case 5:var u=null;switch(te.tag){case 26:u=te.memoizedState;case 5:case 27:var r=te;if(u?aL(u):r.stateNode.complete){se=0,Yt=null;var i=r.sibling;if(i!==null)te=i;else{var s=r.return;s!==null?(te=s,Ss(s)):te=null}break t}}se=0,Yt=null,In(e,t,n,5);break;case 6:se=0,Yt=null,In(e,t,n,6);break;case 8:mp(),Pe=6;break e;default:throw Error(E(462))}}Aw();break}catch(f){Ox(e,f)}while(!0);return ho=Hl=null,V.H=o,V.A=l,re=a,te!==null?0:(Se=null,ae=0,ss(),Pe)}function Aw(){for(;te!==null&&!Qb();)Ux(te)}function Ux(e){var t=sx(e.alternate,e,Ro);e.memoizedProps=e.pendingProps,t===null?Ss(e):te=t}function Qg(e){var t=e,a=t.alternate;switch(t.tag){case 15:case 0:t=Gg(a,t,t.pendingProps,t.type,void 0,ae);break;case 11:t=Gg(a,t,t.pendingProps,t.type.render,t.ref,ae);break;case 5:$d(t);default:fx(a,t),t=te=fS(t,Ro),t=sx(a,t,Ro)}e.memoizedProps=e.pendingProps,t===null?Ss(e):te=t}function In(e,t,a,o){ho=Hl=null,$d(t),En=null,Wu=0;var l=t.return;try{if(hw(e,l,t,a,ae)){Pe=1,Vi(e,pa(a,e.current)),te=null;return}}catch(n){if(l!==null)throw te=l,n;Pe=1,Vi(e,pa(a,e.current)),te=null;return}t.flags&32768?(oe||o===1?e=!0:Yn||(ae&536870912)!==0?e=!1:(jo=e=!0,(o===2||o===9||o===3||o===6)&&(o=aa.current,o!==null&&o.tag===13&&(o.flags|=16384))),Nx(t,e)):Ss(t)}function Ss(e){var t=e;do{if((t.flags&32768)!==0){Nx(t,jo);return}e=t.return;var a=Sw(t.alternate,t,Ro);if(a!==null){te=a;return}if(t=t.sibling,t!==null){te=t;return}te=t=e}while(t!==null);Pe===0&&(Pe=5)}function Nx(e,t){do{var a=xw(e.alternate,e);if(a!==null){a.flags&=32767,te=a;return}if(a=e.return,a!==null&&(a.flags|=32768,a.subtreeFlags=0,a.deletions=null),!t&&(e=e.sibling,e!==null)){te=e;return}te=e=a}while(e!==null);Pe=6,te=null}function Wg(e,t,a,o,l,n,u,r,i){e.cancelPendingCommit=null;do xs();while(Ke!==0);if((re&6)!==0)throw Error(E(327));if(t!==null){if(t===e.current)throw Error(E(177));if(n=t.lanes|t.childLanes,n|=Hd,u0(e,a,n,u,r,i),e===Se&&(te=Se=null,ae=0),Hn=t,el=e,So=a,pd=n,md=l,Ex=o,(t.subtreeFlags&10256)!==0||(t.flags&10256)!==0?(e.callbackNode=null,e.callbackPriority=0,Mw(Pi,function(){return qx(),null})):(e.callbackNode=null,e.callbackPriority=0),o=(t.flags&13878)!==0,(t.subtreeFlags&13878)!==0||o){o=V.T,V.T=null,l=ie.p,ie.p=2,u=re,re|=4;try{Lw(e,t,a)}finally{re=u,ie.p=l,V.T=o}}Ke=1,kx(),Hx(),zx()}}function kx(){if(Ke===1){Ke=0;var e=el,t=Hn,a=(t.flags&13878)!==0;if((t.subtreeFlags&13878)!==0||a){a=V.T,V.T=null;var o=ie.p;ie.p=2;var l=re;re|=4;try{yx(t,e);var n=xd,u=aS(e.containerInfo),r=n.focusedElem,i=n.selectionRange;if(u!==r&&r&&r.ownerDocument&&tS(r.ownerDocument.documentElement,r)){if(i!==null&&kd(r)){var s=i.start,f=i.end;if(f===void 0&&(f=s),"selectionStart"in r)r.selectionStart=s,r.selectionEnd=Math.min(f,r.value.length);else{var d=r.ownerDocument||document,c=d&&d.defaultView||window;if(c.getSelection){var g=c.getSelection(),L=r.textContent.length,x=Math.min(i.start,L),S=i.end===void 0?x:Math.min(i.end,L);!g.extend&&x>S&&(u=S,S=x,x=u);var h=xg(r,x),p=xg(r,S);if(h&&p&&(g.rangeCount!==1||g.anchorNode!==h.node||g.anchorOffset!==h.offset||g.focusNode!==p.node||g.focusOffset!==p.offset)){var m=d.createRange();m.setStart(h.node,h.offset),g.removeAllRanges(),x>S?(g.addRange(m),g.extend(p.node,p.offset)):(m.setEnd(p.node,p.offset),g.addRange(m))}}}}for(d=[],g=r;g=g.parentNode;)g.nodeType===1&&d.push({element:g,left:g.scrollLeft,top:g.scrollTop});for(typeof r.focus=="function"&&r.focus(),r=0;r<d.length;r++){var v=d[r];v.element.scrollLeft=v.left,v.element.scrollTop=v.top}}os=!!Sd,xd=Sd=null}finally{re=l,ie.p=o,V.T=a}}e.current=t,Ke=2}}function Hx(){if(Ke===2){Ke=0;var e=el,t=Hn,a=(t.flags&8772)!==0;if((t.subtreeFlags&8772)!==0||a){a=V.T,V.T=null;var o=ie.p;ie.p=2;var l=re;re|=4;try{gx(e,t.alternate,t)}finally{re=l,ie.p=o,V.T=a}}Ke=3}}function zx(){if(Ke===4||Ke===3){Ke=0,Wb();var e=el,t=Hn,a=So,o=Ex;(t.subtreeFlags&10256)!==0||(t.flags&10256)!==0?Ke=5:(Ke=0,Hn=el=null,Fx(e,e.pendingLanes));var l=e.pendingLanes;if(l===0&&($o=null),Md(a),t=t.stateNode,Jt&&typeof Jt.onCommitFiberRoot=="function")try{Jt.onCommitFiberRoot(ur,t,void 0,(t.current.flags&128)===128)}catch(i){}if(o!==null){t=V.T,l=ie.p,ie.p=2,V.T=null;try{for(var n=e.onRecoverableError,u=0;u<o.length;u++){var r=o[u];n(r.value,{componentStack:r.stack})}}finally{V.T=t,ie.p=l}}(So&3)!==0&&xs(),Qa(e),l=e.pendingLanes,(a&261930)!==0&&(l&42)!==0?e===hd?Gu++:(Gu=0,hd=e):Gu=0,mr(0,!1)}}function Fx(e,t){(e.pooledCacheLanes&=t)===0&&(t=e.pooledCache,t!=null&&(e.pooledCache=null,cr(t)))}function xs(){return kx(),Hx(),zx(),qx()}function qx(){if(Ke!==5)return!1;var e=el,t=pd;pd=0;var a=Md(So),o=V.T,l=ie.p;try{ie.p=32>a?32:a,V.T=null,a=md,md=null;var n=el,u=So;if(Ke=0,Hn=el=null,So=0,(re&6)!==0)throw Error(E(331));var r=re;if(re|=4,wx(n.current),Rx(n,n.current,u,a),re=r,mr(0,!1),Jt&&typeof Jt.onPostCommitFiberRoot=="function")try{Jt.onPostCommitFiberRoot(ur,n)}catch(i){}return!0}finally{ie.p=l,V.T=o,Fx(e,t)}}function Jg(e,t,a){t=pa(a,t),t=id(e.stateNode,t,2),e=Jo(e,t,2),e!==null&&(ir(e,2),Qa(e))}function fe(e,t,a){if(e.tag===3)Jg(e,e,a);else for(;t!==null;){if(t.tag===3){Jg(t,e,a);break}else if(t.tag===1){var o=t.stateNode;if(typeof t.type.getDerivedStateFromError=="function"||typeof o.componentDidCatch=="function"&&($o===null||!$o.has(o))){e=pa(a,e),a=ox(2),o=Jo(t,a,2),o!==null&&(lx(a,o,t,e),ir(o,2),Qa(o));break}}t=t.return}}function wc(e,t,a){var o=e.pingCache;if(o===null){o=e.pingCache=new Rw;var l=new Set;o.set(t,l)}else l=o.get(t),l===void 0&&(l=new Set,o.set(t,l));l.has(a)||(dp=!0,l.add(a),e=Tw.bind(null,e,t,a),t.then(e,e))}function Tw(e,t,a){var o=e.pingCache;o!==null&&o.delete(t),e.pingedLanes|=e.suspendedLanes&a,e.warmLanes&=~a,Se===e&&(ae&a)===a&&(Pe===4||Pe===3&&(ae&62914560)===ae&&300>Wt()-gs?(re&2)===0&&zn(e,0):pp|=a,kn===ae&&(kn=0)),Qa(e)}function Gx(e,t){t===0&&(t=Bv()),e=kl(e,t),e!==null&&(ir(e,t),Qa(e))}function Ew(e){var t=e.memoizedState,a=0;t!==null&&(a=t.retryLane),Gx(e,a)}function Dw(e,t){var a=0;switch(e.tag){case 31:case 13:var o=e.stateNode,l=e.memoizedState;l!==null&&(a=l.retryLane);break;case 19:o=e.stateNode;break;case 22:o=e.stateNode._retryCache;break;default:throw Error(E(314))}o!==null&&o.delete(t),Gx(e,a)}function Mw(e,t){return Ed(e,t)}var Qi=null,dn=null,gd=!1,Wi=!1,Ac=!1,Ko=0;function Qa(e){e!==dn&&e.next===null&&(dn===null?Qi=dn=e:dn=dn.next=e),Wi=!0,gd||(gd=!0,Pw())}function mr(e,t){if(!Ac&&Wi){Ac=!0;do for(var a=!1,o=Qi;o!==null;){if(!t)if(e!==0){var l=o.pendingLanes;if(l===0)var n=0;else{var u=o.suspendedLanes,r=o.pingedLanes;n=(1<<31-$t(42|e)+1)-1,n&=l&~(u&~r),n=n&201326741?n&201326741|1:n?n|2:0}n!==0&&(a=!0,$g(o,n))}else n=ae,n=ns(o,o===Se?n:0,o.cancelPendingCommit!==null||o.timeoutHandle!==-1),(n&3)===0||rr(o,n)||(a=!0,$g(o,n));o=o.next}while(a);Ac=!1}}function Ow(){Vx()}function Vx(){Wi=gd=!1;var e=0;Ko!==0&&Fw()&&(e=Ko);for(var t=Wt(),a=null,o=Qi;o!==null;){var l=o.next,n=Xx(o,t);n===0?(o.next=null,a===null?Qi=l:a.next=l,l===null&&(dn=a)):(a=o,(e!==0||(n&3)!==0)&&(Wi=!0)),o=l}Ke!==0&&Ke!==5||mr(e,!1),Ko!==0&&(Ko=0)}function Xx(e,t){for(var a=e.suspendedLanes,o=e.pingedLanes,l=e.expirationTimes,n=e.pendingLanes&-62914561;0<n;){var u=31-$t(n),r=1<<u,i=l[u];i===-1?((r&a)===0||(r&o)!==0)&&(l[u]=n0(r,t)):i<=t&&(e.expiredLanes|=r),n&=~r}if(t=Se,a=ae,a=ns(e,e===t?a:0,e.cancelPendingCommit!==null||e.timeoutHandle!==-1),o=e.callbackNode,a===0||e===t&&(se===2||se===9)||e.cancelPendingCommit!==null)return o!==null&&o!==null&&ac(o),e.callbackNode=null,e.callbackPriority=0;if((a&3)===0||rr(e,a)){if(t=a&-a,t===e.callbackPriority)return t;switch(o!==null&&ac(o),Md(a)){case 2:case 8:a=Ov;break;case 32:a=Pi;break;case 268435456:a=Pv;break;default:a=Pi}return o=jx.bind(null,e),a=Ed(a,o),e.callbackPriority=t,e.callbackNode=a,t}return o!==null&&o!==null&&ac(o),e.callbackPriority=2,e.callbackNode=null,2}function jx(e,t){if(Ke!==0&&Ke!==5)return e.callbackNode=null,e.callbackPriority=0,null;var a=e.callbackNode;if(xs()&&e.callbackNode!==a)return null;var o=ae;return o=ns(e,e===Se?o:0,e.cancelPendingCommit!==null||e.timeoutHandle!==-1),o===0?null:(Mx(e,o,t),Xx(e,Wt()),e.callbackNode!=null&&e.callbackNode===a?jx.bind(null,e):null)}function $g(e,t){if(xs())return null;Mx(e,t,!0)}function Pw(){Gw(function(){(re&6)!==0?Ed(Mv,Ow):Vx()})}function hp(){if(Ko===0){var e=_n;e===0&&(e=Jr,Jr<<=1,(Jr&261888)===0&&(Jr=256)),Ko=e}return Ko}function ev(e){return e==null||typeof e=="symbol"||typeof e=="boolean"?null:typeof e=="function"?e:gi(""+e)}function tv(e,t){var a=t.ownerDocument.createElement("input");return a.name=t.name,a.value=t.value,e.id&&a.setAttribute("form",e.id),t.parentNode.insertBefore(a,t),e=new FormData(e),a.parentNode.removeChild(a),e}function Bw(e,t,a,o,l){if(t==="submit"&&a&&a.stateNode===l){var n=ev((l[Ht]||null).action),u=o.submitter;u&&(t=(t=u[Ht]||null)?ev(t.formAction):u.getAttribute("formAction"),t!==null&&(n=t,u=null));var r=new us("action","action",null,o,l);e.push({event:r,listeners:[{instance:null,listener:function(){if(o.defaultPrevented){if(Ko!==0){var i=u?tv(l,u):new FormData(l);ud(a,{pending:!0,data:i,method:l.method,action:n},null,i)}}else typeof n=="function"&&(r.preventDefault(),i=u?tv(l,u):new FormData(l),ud(a,{pending:!0,data:i,method:l.method,action:n},n,i))},currentTarget:l}]})}}for(fi=0;fi<Kc.length;fi++)ci=Kc[fi],av=ci.toLowerCase(),ov=ci[0].toUpperCase()+ci.slice(1),Ea(av,"on"+ov);var ci,av,ov,fi;Ea(lS,"onAnimationEnd");Ea(nS,"onAnimationIteration");Ea(uS,"onAnimationStart");Ea("dblclick","onDoubleClick");Ea("focusin","onFocus");Ea("focusout","onBlur");Ea(J0,"onTransitionRun");Ea($0,"onTransitionStart");Ea(ew,"onTransitionCancel");Ea(rS,"onTransitionEnd");Pn("onMouseEnter",["mouseout","mouseover"]);Pn("onMouseLeave",["mouseout","mouseover"]);Pn("onPointerEnter",["pointerout","pointerover"]);Pn("onPointerLeave",["pointerout","pointerover"]);_l("onChange","change click focusin focusout input keydown keyup selectionchange".split(" "));_l("onSelect","focusout contextmenu dragend focusin keydown keyup mousedown mouseup selectionchange".split(" "));_l("onBeforeInput",["compositionend","keypress","textInput","paste"]);_l("onCompositionEnd","compositionend focusout keydown keypress keyup mousedown".split(" "));_l("onCompositionStart","compositionstart focusout keydown keypress keyup mousedown".split(" "));_l("onCompositionUpdate","compositionupdate focusout keydown keypress keyup mousedown".split(" "));var er="abort canplay canplaythrough durationchange emptied encrypted ended error loadeddata loadedmetadata loadstart pause play playing progress ratechange resize seeked seeking stalled suspend timeupdate volumechange waiting".split(" "),_w=new Set("beforetoggle cancel close invalid load scroll scrollend toggle".split(" ").concat(er));function Yx(e,t){t=(t&4)!==0;for(var a=0;a<e.length;a++){var o=e[a],l=o.event;o=o.listeners;e:{var n=void 0;if(t)for(var u=o.length-1;0<=u;u--){var r=o[u],i=r.instance,s=r.currentTarget;if(r=r.listener,i!==n&&l.isPropagationStopped())break e;n=r,l.currentTarget=s;try{n(l)}catch(f){_i(f)}l.currentTarget=null,n=i}else for(u=0;u<o.length;u++){if(r=o[u],i=r.instance,s=r.currentTarget,r=r.listener,i!==n&&l.isPropagationStopped())break e;n=r,l.currentTarget=s;try{n(l)}catch(f){_i(f)}l.currentTarget=null,n=i}}}}function ee(e,t){var a=t[zc];a===void 0&&(a=t[zc]=new Set);var o=e+"__bubble";a.has(o)||(Kx(t,e,2,!1),a.add(o))}function Tc(e,t,a){var o=0;t&&(o|=4),Kx(a,e,o,t)}var di="_reactListening"+Math.random().toString(36).slice(2);function gp(e){if(!e[di]){e[di]=!0,Hv.forEach(function(a){a!=="selectionchange"&&(_w.has(a)||Tc(a,!1,e),Tc(a,!0,e))});var t=e.nodeType===9?e:e.ownerDocument;t===null||t[di]||(t[di]=!0,Tc("selectionchange",!1,t))}}function Kx(e,t,a,o){switch(rL(t)){case 2:var l=sA;break;case 8:l=fA;break;default:l=Lp}a=l.bind(null,t,a,e),l=void 0,!Xc||t!=="touchstart"&&t!=="touchmove"&&t!=="wheel"||(l=!0),o?l!==void 0?e.addEventListener(t,a,{capture:!0,passive:l}):e.addEventListener(t,a,!0):l!==void 0?e.addEventListener(t,a,{passive:l}):e.addEventListener(t,a,!1)}function Ec(e,t,a,o,l){var n=o;if((t&1)===0&&(t&2)===0&&o!==null)e:for(;;){if(o===null)return;var u=o.tag;if(u===3||u===4){var r=o.stateNode.containerInfo;if(r===l)break;if(u===4)for(u=o.return;u!==null;){var i=u.tag;if((i===3||i===4)&&u.stateNode.containerInfo===l)return;u=u.return}for(;r!==null;){if(u=hn(r),u===null)return;if(i=u.tag,i===5||i===6||i===26||i===27){o=n=u;continue e}r=r.parentNode}}o=o.return}Yv(function(){var s=n,f=Bd(a),d=[];e:{var c=iS.get(e);if(c!==void 0){var g=us,L=e;switch(e){case"keypress":if(Si(a)===0)break e;case"keydown":case"keyup":g=D0;break;case"focusin":L="focus",g=rc;break;case"focusout":L="blur",g=rc;break;case"beforeblur":case"afterblur":g=rc;break;case"click":if(a.button===2)break e;case"auxclick":case"dblclick":case"mousedown":case"mousemove":case"mouseup":case"mouseout":case"mouseover":case"contextmenu":g=fg;break;case"drag":case"dragend":case"dragenter":case"dragexit":case"dragleave":case"dragover":case"dragstart":case"drop":g=S0;break;case"touchcancel":case"touchend":case"touchmove":case"touchstart":g=P0;break;case lS:case nS:case uS:g=y0;break;case rS:g=_0;break;case"scroll":case"scrollend":g=g0;break;case"wheel":g=N0;break;case"copy":case"cut":case"paste":g=R0;break;case"gotpointercapture":case"lostpointercapture":case"pointercancel":case"pointerdown":case"pointermove":case"pointerout":case"pointerover":case"pointerup":g=dg;break;case"toggle":case"beforetoggle":g=H0}var x=(t&4)!==0,S=!x&&(e==="scroll"||e==="scrollend"),h=x?c!==null?c+"Capture":null:c;x=[];for(var p=s,m;p!==null;){var v=p;if(m=v.stateNode,v=v.tag,v!==5&&v!==26&&v!==27||m===null||h===null||(v=ju(p,h),v!=null&&x.push(tr(p,v,m))),S)break;p=p.return}0<x.length&&(c=new g(c,L,null,a,f),d.push({event:c,listeners:x}))}}if((t&7)===0){e:{if(c=e==="mouseover"||e==="pointerover",g=e==="mouseout"||e==="pointerout",c&&a!==Vc&&(L=a.relatedTarget||a.fromElement)&&(hn(L)||L[Gn]))break e;if((g||c)&&(c=f.window===f?f:(c=f.ownerDocument)?c.defaultView||c.parentWindow:window,g?(L=a.relatedTarget||a.toElement,g=s,L=L?hn(L):null,L!==null&&(S=nr(L),x=L.tag,L!==S||x!==5&&x!==27&&x!==6)&&(L=null)):(g=null,L=s),g!==L)){if(x=fg,v="onMouseLeave",h="onMouseEnter",p="mouse",(e==="pointerout"||e==="pointerover")&&(x=dg,v="onPointerLeave",h="onPointerEnter",p="pointer"),S=g==null?c:Eu(g),m=L==null?c:Eu(L),c=new x(v,p+"leave",g,a,f),c.target=S,c.relatedTarget=m,v=null,hn(f)===s&&(x=new x(h,p+"enter",L,a,f),x.target=m,x.relatedTarget=S,v=x),S=v,g&&L)t:{for(x=Uw,h=g,p=L,m=0,v=h;v;v=x(v))m++;v=0;for(var C=p;C;C=x(C))v++;for(;0<m-v;)h=x(h),m--;for(;0<v-m;)p=x(p),v--;for(;m--;){if(h===p||p!==null&&h===p.alternate){x=h;break t}h=x(h),p=x(p)}x=null}else x=null;g!==null&&lv(d,c,g,x,!1),L!==null&&S!==null&&lv(d,S,L,x,!0)}}e:{if(c=s?Eu(s):window,g=c.nodeName&&c.nodeName.toLowerCase(),g==="select"||g==="input"&&c.type==="file")var I=gg;else if(hg(c))if($v)I=Z0;else{I=Y0;var R=j0}else g=c.nodeName,!g||g.toLowerCase()!=="input"||c.type!=="checkbox"&&c.type!=="radio"?s&&Pd(s.elementType)&&(I=gg):I=K0;if(I&&(I=I(e,s))){Jv(d,I,a,f);break e}R&&R(e,c,s),e==="focusout"&&s&&c.type==="number"&&s.memoizedProps.value!=null&&Gc(c,"number",c.value)}switch(R=s?Eu(s):window,e){case"focusin":(hg(R)||R.contentEditable==="true")&&(Sn=R,jc=s,Bu=null);break;case"focusout":Bu=jc=Sn=null;break;case"mousedown":Yc=!0;break;case"contextmenu":case"mouseup":case"dragend":Yc=!1,Lg(d,a,f);break;case"selectionchange":if(W0)break;case"keydown":case"keyup":Lg(d,a,f)}var w;if(Nd)e:{switch(e){case"compositionstart":var A="onCompositionStart";break e;case"compositionend":A="onCompositionEnd";break e;case"compositionupdate":A="onCompositionUpdate";break e}A=void 0}else vn?Qv(e,a)&&(A="onCompositionEnd"):e==="keydown"&&a.keyCode===229&&(A="onCompositionStart");A&&(Zv&&a.locale!=="ko"&&(vn||A!=="onCompositionStart"?A==="onCompositionEnd"&&vn&&(w=Kv()):(Xo=f,_d="value"in Xo?Xo.value:Xo.textContent,vn=!0)),R=Ji(s,A),0<R.length&&(A=new cg(A,e,null,a,f),d.push({event:A,listeners:R}),w?A.data=w:(w=Wv(a),w!==null&&(A.data=w)))),(w=F0?q0(e,a):G0(e,a))&&(A=Ji(s,"onBeforeInput"),0<A.length&&(R=new cg("onBeforeInput","beforeinput",null,a,f),d.push({event:R,listeners:A}),R.data=w)),Bw(d,e,s,a,f)}Yx(d,t)})}function tr(e,t,a){return{instance:e,listener:t,currentTarget:a}}function Ji(e,t){for(var a=t+"Capture",o=[];e!==null;){var l=e,n=l.stateNode;if(l=l.tag,l!==5&&l!==26&&l!==27||n===null||(l=ju(e,a),l!=null&&o.unshift(tr(e,l,n)),l=ju(e,t),l!=null&&o.push(tr(e,l,n))),e.tag===3)return o;e=e.return}return[]}function Uw(e){if(e===null)return null;do e=e.return;while(e&&e.tag!==5&&e.tag!==27);return e||null}function lv(e,t,a,o,l){for(var n=t._reactName,u=[];a!==null&&a!==o;){var r=a,i=r.alternate,s=r.stateNode;if(r=r.tag,i!==null&&i===o)break;r!==5&&r!==26&&r!==27||s===null||(i=s,l?(s=ju(a,n),s!=null&&u.unshift(tr(a,s,i))):l||(s=ju(a,n),s!=null&&u.push(tr(a,s,i)))),a=a.return}u.length!==0&&e.push({event:t,listeners:u})}var Nw=/\r\n?/g,kw=/\u0000|\uFFFD/g;function nv(e){return(typeof e=="string"?e:""+e).replace(Nw,"\n").replace(kw,"")}function Zx(e,t){return t=nv(t),nv(e)===t}function de(e,t,a,o,l,n){switch(a){case"children":typeof o=="string"?t==="body"||t==="textarea"&&o===""||Bn(e,o):(typeof o=="number"||typeof o=="bigint")&&t!=="body"&&Bn(e,""+o);break;case"className":ti(e,"class",o);break;case"tabIndex":ti(e,"tabindex",o);break;case"dir":case"role":case"viewBox":case"width":case"height":ti(e,a,o);break;case"style":jv(e,o,n);break;case"data":if(t!=="object"){ti(e,"data",o);break}case"src":case"href":if(o===""&&(t!=="a"||a!=="href")){e.removeAttribute(a);break}if(o==null||typeof o=="function"||typeof o=="symbol"||typeof o=="boolean"){e.removeAttribute(a);break}o=gi(""+o),e.setAttribute(a,o);break;case"action":case"formAction":if(typeof o=="function"){e.setAttribute(a,"javascript:throw new Error('A React form was unexpectedly submitted. If you called form.submit() manually, consider using form.requestSubmit() instead. If you\\'re trying to use event.stopPropagation() in a submit event handler, consider also calling event.preventDefault().')");break}else typeof n=="function"&&(a==="formAction"?(t!=="input"&&de(e,t,"name",l.name,l,null),de(e,t,"formEncType",l.formEncType,l,null),de(e,t,"formMethod",l.formMethod,l,null),de(e,t,"formTarget",l.formTarget,l,null)):(de(e,t,"encType",l.encType,l,null),de(e,t,"method",l.method,l,null),de(e,t,"target",l.target,l,null)));if(o==null||typeof o=="symbol"||typeof o=="boolean"){e.removeAttribute(a);break}o=gi(""+o),e.setAttribute(a,o);break;case"onClick":o!=null&&(e.onclick=mo);break;case"onScroll":o!=null&&ee("scroll",e);break;case"onScrollEnd":o!=null&&ee("scrollend",e);break;case"dangerouslySetInnerHTML":if(o!=null){if(typeof o!="object"||!("__html"in o))throw Error(E(61));if(a=o.__html,a!=null){if(l.children!=null)throw Error(E(60));e.innerHTML=a}}break;case"multiple":e.multiple=o&&typeof o!="function"&&typeof o!="symbol";break;case"muted":e.muted=o&&typeof o!="function"&&typeof o!="symbol";break;case"suppressContentEditableWarning":case"suppressHydrationWarning":case"defaultValue":case"defaultChecked":case"innerHTML":case"ref":break;case"autoFocus":break;case"xlinkHref":if(o==null||typeof o=="function"||typeof o=="boolean"||typeof o=="symbol"){e.removeAttribute("xlink:href");break}a=gi(""+o),e.setAttributeNS("http://www.w3.org/1999/xlink","xlink:href",a);break;case"contentEditable":case"spellCheck":case"draggable":case"value":case"autoReverse":case"externalResourcesRequired":case"focusable":case"preserveAlpha":o!=null&&typeof o!="function"&&typeof o!="symbol"?e.setAttribute(a,""+o):e.removeAttribute(a);break;case"inert":case"allowFullScreen":case"async":case"autoPlay":case"controls":case"default":case"defer":case"disabled":case"disablePictureInPicture":case"disableRemotePlayback":case"formNoValidate":case"hidden":case"loop":case"noModule":case"noValidate":case"open":case"playsInline":case"readOnly":case"required":case"reversed":case"scoped":case"seamless":case"itemScope":o&&typeof o!="function"&&typeof o!="symbol"?e.setAttribute(a,""):e.removeAttribute(a);break;case"capture":case"download":o===!0?e.setAttribute(a,""):o!==!1&&o!=null&&typeof o!="function"&&typeof o!="symbol"?e.setAttribute(a,o):e.removeAttribute(a);break;case"cols":case"rows":case"size":case"span":o!=null&&typeof o!="function"&&typeof o!="symbol"&&!isNaN(o)&&1<=o?e.setAttribute(a,o):e.removeAttribute(a);break;case"rowSpan":case"start":o==null||typeof o=="function"||typeof o=="symbol"||isNaN(o)?e.removeAttribute(a):e.setAttribute(a,o);break;case"popover":ee("beforetoggle",e),ee("toggle",e),hi(e,"popover",o);break;case"xlinkActuate":no(e,"http://www.w3.org/1999/xlink","xlink:actuate",o);break;case"xlinkArcrole":no(e,"http://www.w3.org/1999/xlink","xlink:arcrole",o);break;case"xlinkRole":no(e,"http://www.w3.org/1999/xlink","xlink:role",o);break;case"xlinkShow":no(e,"http://www.w3.org/1999/xlink","xlink:show",o);break;case"xlinkTitle":no(e,"http://www.w3.org/1999/xlink","xlink:title",o);break;case"xlinkType":no(e,"http://www.w3.org/1999/xlink","xlink:type",o);break;case"xmlBase":no(e,"http://www.w3.org/XML/1998/namespace","xml:base",o);break;case"xmlLang":no(e,"http://www.w3.org/XML/1998/namespace","xml:lang",o);break;case"xmlSpace":no(e,"http://www.w3.org/XML/1998/namespace","xml:space",o);break;case"is":hi(e,"is",o);break;case"innerText":case"textContent":break;default:(!(2<a.length)||a[0]!=="o"&&a[0]!=="O"||a[1]!=="n"&&a[1]!=="N")&&(a=m0.get(a)||a,hi(e,a,o))}}function vd(e,t,a,o,l,n){switch(a){case"style":jv(e,o,n);break;case"dangerouslySetInnerHTML":if(o!=null){if(typeof o!="object"||!("__html"in o))throw Error(E(61));if(a=o.__html,a!=null){if(l.children!=null)throw Error(E(60));e.innerHTML=a}}break;case"children":typeof o=="string"?Bn(e,o):(typeof o=="number"||typeof o=="bigint")&&Bn(e,""+o);break;case"onScroll":o!=null&&ee("scroll",e);break;case"onScrollEnd":o!=null&&ee("scrollend",e);break;case"onClick":o!=null&&(e.onclick=mo);break;case"suppressContentEditableWarning":case"suppressHydrationWarning":case"innerHTML":case"ref":break;case"innerText":case"textContent":break;default:if(!zv.hasOwnProperty(a))e:{if(a[0]==="o"&&a[1]==="n"&&(l=a.endsWith("Capture"),t=a.slice(2,l?a.length-7:void 0),n=e[Ht]||null,n=n!=null?n[a]:null,typeof n=="function"&&e.removeEventListener(t,n,l),typeof o=="function")){typeof n!="function"&&n!==null&&(a in e?e[a]=null:e.hasAttribute(a)&&e.removeAttribute(a)),e.addEventListener(t,o,l);break e}a in e?e[a]=o:o===!0?e.setAttribute(a,""):hi(e,a,o)}}}function vt(e,t,a){switch(t){case"div":case"span":case"svg":case"path":case"a":case"g":case"p":case"li":break;case"img":ee("error",e),ee("load",e);var o=!1,l=!1,n;for(n in a)if(a.hasOwnProperty(n)){var u=a[n];if(u!=null)switch(n){case"src":o=!0;break;case"srcSet":l=!0;break;case"children":case"dangerouslySetInnerHTML":throw Error(E(137,t));default:de(e,t,n,u,a,null)}}l&&de(e,t,"srcSet",a.srcSet,a,null),o&&de(e,t,"src",a.src,a,null);return;case"input":ee("invalid",e);var r=n=u=l=null,i=null,s=null;for(o in a)if(a.hasOwnProperty(o)){var f=a[o];if(f!=null)switch(o){case"name":l=f;break;case"type":u=f;break;case"checked":i=f;break;case"defaultChecked":s=f;break;case"value":n=f;break;case"defaultValue":r=f;break;case"children":case"dangerouslySetInnerHTML":if(f!=null)throw Error(E(137,t));break;default:de(e,t,o,f,a,null)}}Gv(e,n,r,i,s,u,l,!1);return;case"select":ee("invalid",e),o=u=n=null;for(l in a)if(a.hasOwnProperty(l)&&(r=a[l],r!=null))switch(l){case"value":n=r;break;case"defaultValue":u=r;break;case"multiple":o=r;default:de(e,t,l,r,a,null)}t=n,a=u,e.multiple=!!o,t!=null?wn(e,!!o,t,!1):a!=null&&wn(e,!!o,a,!0);return;case"textarea":ee("invalid",e),n=l=o=null;for(u in a)if(a.hasOwnProperty(u)&&(r=a[u],r!=null))switch(u){case"value":o=r;break;case"defaultValue":l=r;break;case"children":n=r;break;case"dangerouslySetInnerHTML":if(r!=null)throw Error(E(91));break;default:de(e,t,u,r,a,null)}Xv(e,o,l,n);return;case"option":for(i in a)a.hasOwnProperty(i)&&(o=a[i],o!=null)&&(i==="selected"?e.selected=o&&typeof o!="function"&&typeof o!="symbol":de(e,t,i,o,a,null));return;case"dialog":ee("beforetoggle",e),ee("toggle",e),ee("cancel",e),ee("close",e);break;case"iframe":case"object":ee("load",e);break;case"video":case"audio":for(o=0;o<er.length;o++)ee(er[o],e);break;case"image":ee("error",e),ee("load",e);break;case"details":ee("toggle",e);break;case"embed":case"source":case"link":ee("error",e),ee("load",e);case"area":case"base":case"br":case"col":case"hr":case"keygen":case"meta":case"param":case"track":case"wbr":case"menuitem":for(s in a)if(a.hasOwnProperty(s)&&(o=a[s],o!=null))switch(s){case"children":case"dangerouslySetInnerHTML":throw Error(E(137,t));default:de(e,t,s,o,a,null)}return;default:if(Pd(t)){for(f in a)a.hasOwnProperty(f)&&(o=a[f],o!==void 0&&vd(e,t,f,o,a,void 0));return}}for(r in a)a.hasOwnProperty(r)&&(o=a[r],o!=null&&de(e,t,r,o,a,null))}function Hw(e,t,a,o){switch(t){case"div":case"span":case"svg":case"path":case"a":case"g":case"p":case"li":break;case"input":var l=null,n=null,u=null,r=null,i=null,s=null,f=null;for(g in a){var d=a[g];if(a.hasOwnProperty(g)&&d!=null)switch(g){case"checked":break;case"value":break;case"defaultValue":i=d;default:o.hasOwnProperty(g)||de(e,t,g,null,o,d)}}for(var c in o){var g=o[c];if(d=a[c],o.hasOwnProperty(c)&&(g!=null||d!=null))switch(c){case"type":n=g;break;case"name":l=g;break;case"checked":s=g;break;case"defaultChecked":f=g;break;case"value":u=g;break;case"defaultValue":r=g;break;case"children":case"dangerouslySetInnerHTML":if(g!=null)throw Error(E(137,t));break;default:g!==d&&de(e,t,c,g,o,d)}}qc(e,u,r,i,s,f,n,l);return;case"select":g=u=r=c=null;for(n in a)if(i=a[n],a.hasOwnProperty(n)&&i!=null)switch(n){case"value":break;case"multiple":g=i;default:o.hasOwnProperty(n)||de(e,t,n,null,o,i)}for(l in o)if(n=o[l],i=a[l],o.hasOwnProperty(l)&&(n!=null||i!=null))switch(l){case"value":c=n;break;case"defaultValue":r=n;break;case"multiple":u=n;default:n!==i&&de(e,t,l,n,o,i)}t=r,a=u,o=g,c!=null?wn(e,!!a,c,!1):!!o!=!!a&&(t!=null?wn(e,!!a,t,!0):wn(e,!!a,a?[]:"",!1));return;case"textarea":g=c=null;for(r in a)if(l=a[r],a.hasOwnProperty(r)&&l!=null&&!o.hasOwnProperty(r))switch(r){case"value":break;case"children":break;default:de(e,t,r,null,o,l)}for(u in o)if(l=o[u],n=a[u],o.hasOwnProperty(u)&&(l!=null||n!=null))switch(u){case"value":c=l;break;case"defaultValue":g=l;break;case"children":break;case"dangerouslySetInnerHTML":if(l!=null)throw Error(E(91));break;default:l!==n&&de(e,t,u,l,o,n)}Vv(e,c,g);return;case"option":for(var L in a)c=a[L],a.hasOwnProperty(L)&&c!=null&&!o.hasOwnProperty(L)&&(L==="selected"?e.selected=!1:de(e,t,L,null,o,c));for(i in o)c=o[i],g=a[i],o.hasOwnProperty(i)&&c!==g&&(c!=null||g!=null)&&(i==="selected"?e.selected=c&&typeof c!="function"&&typeof c!="symbol":de(e,t,i,c,o,g));return;case"img":case"link":case"area":case"base":case"br":case"col":case"embed":case"hr":case"keygen":case"meta":case"param":case"source":case"track":case"wbr":case"menuitem":for(var x in a)c=a[x],a.hasOwnProperty(x)&&c!=null&&!o.hasOwnProperty(x)&&de(e,t,x,null,o,c);for(s in o)if(c=o[s],g=a[s],o.hasOwnProperty(s)&&c!==g&&(c!=null||g!=null))switch(s){case"children":case"dangerouslySetInnerHTML":if(c!=null)throw Error(E(137,t));break;default:de(e,t,s,c,o,g)}return;default:if(Pd(t)){for(var S in a)c=a[S],a.hasOwnProperty(S)&&c!==void 0&&!o.hasOwnProperty(S)&&vd(e,t,S,void 0,o,c);for(f in o)c=o[f],g=a[f],!o.hasOwnProperty(f)||c===g||c===void 0&&g===void 0||vd(e,t,f,c,o,g);return}}for(var h in a)c=a[h],a.hasOwnProperty(h)&&c!=null&&!o.hasOwnProperty(h)&&de(e,t,h,null,o,c);for(d in o)c=o[d],g=a[d],!o.hasOwnProperty(d)||c===g||c==null&&g==null||de(e,t,d,c,o,g)}function uv(e){switch(e){case"css":case"script":case"font":case"img":case"image":case"input":case"link":return!0;default:return!1}}function zw(){if(typeof performance.getEntriesByType=="function"){for(var e=0,t=0,a=performance.getEntriesByType("resource"),o=0;o<a.length;o++){var l=a[o],n=l.transferSize,u=l.initiatorType,r=l.duration;if(n&&r&&uv(u)){for(u=0,r=l.responseEnd,o+=1;o<a.length;o++){var i=a[o],s=i.startTime;if(s>r)break;var f=i.transferSize,d=i.initiatorType;f&&uv(d)&&(i=i.responseEnd,u+=f*(i<r?1:(r-s)/(i-s)))}if(--o,t+=8*(n+u)/(l.duration/1e3),e++,10<e)break}}if(0<e)return t/e/1e6}return navigator.connection&&(e=navigator.connection.downlink,typeof e=="number")?e:5}var Sd=null,xd=null;function $i(e){return e.nodeType===9?e:e.ownerDocument}function rv(e){switch(e){case"http://www.w3.org/2000/svg":return 1;case"http://www.w3.org/1998/Math/MathML":return 2;default:return 0}}function Qx(e,t){if(e===0)switch(t){case"svg":return 1;case"math":return 2;default:return 0}return e===1&&t==="foreignObject"?0:e}function Ld(e,t){return e==="textarea"||e==="noscript"||typeof t.children=="string"||typeof t.children=="number"||typeof t.children=="bigint"||typeof t.dangerouslySetInnerHTML=="object"&&t.dangerouslySetInnerHTML!==null&&t.dangerouslySetInnerHTML.__html!=null}var Dc=null;function Fw(){var e=window.event;return e&&e.type==="popstate"?e===Dc?!1:(Dc=e,!0):(Dc=null,!1)}var Wx=typeof setTimeout=="function"?setTimeout:void 0,qw=typeof clearTimeout=="function"?clearTimeout:void 0,iv=typeof Promise=="function"?Promise:void 0,Gw=typeof queueMicrotask=="function"?queueMicrotask:typeof iv!="undefined"?function(e){return iv.resolve(null).then(e).catch(Vw)}:Wx;function Vw(e){setTimeout(function(){throw e})}function sl(e){return e==="head"}function sv(e,t){var a=t,o=0;do{var l=a.nextSibling;if(e.removeChild(a),l&&l.nodeType===8)if(a=l.data,a==="/$"||a==="/&"){if(o===0){e.removeChild(l),qn(t);return}o--}else if(a==="$"||a==="$?"||a==="$~"||a==="$!"||a==="&")o++;else if(a==="html")Vu(e.ownerDocument.documentElement);else if(a==="head"){a=e.ownerDocument.head,Vu(a);for(var n=a.firstChild;n;){var u=n.nextSibling,r=n.nodeName;n[sr]||r==="SCRIPT"||r==="STYLE"||r==="LINK"&&n.rel.toLowerCase()==="stylesheet"||a.removeChild(n),n=u}}else a==="body"&&Vu(e.ownerDocument.body);a=l}while(a);qn(t)}function fv(e,t){var a=e;e=0;do{var o=a.nextSibling;if(a.nodeType===1?t?(a._stashedDisplay=a.style.display,a.style.display="none"):(a.style.display=a._stashedDisplay||"",a.getAttribute("style")===""&&a.removeAttribute("style")):a.nodeType===3&&(t?(a._stashedText=a.nodeValue,a.nodeValue=""):a.nodeValue=a._stashedText||""),o&&o.nodeType===8)if(a=o.data,a==="/$"){if(e===0)break;e--}else a!=="$"&&a!=="$?"&&a!=="$~"&&a!=="$!"||e++;a=o}while(a)}function yd(e){var t=e.firstChild;for(t&&t.nodeType===10&&(t=t.nextSibling);t;){var a=t;switch(t=t.nextSibling,a.nodeName){case"HTML":case"HEAD":case"BODY":yd(a),Od(a);continue;case"SCRIPT":case"STYLE":continue;case"LINK":if(a.rel.toLowerCase()==="stylesheet")continue}e.removeChild(a)}}function Xw(e,t,a,o){for(;e.nodeType===1;){var l=a;if(e.nodeName.toLowerCase()!==t.toLowerCase()){if(!o&&(e.nodeName!=="INPUT"||e.type!=="hidden"))break}else if(o){if(!e[sr])switch(t){case"meta":if(!e.hasAttribute("itemprop"))break;return e;case"link":if(n=e.getAttribute("rel"),n==="stylesheet"&&e.hasAttribute("data-precedence"))break;if(n!==l.rel||e.getAttribute("href")!==(l.href==null||l.href===""?null:l.href)||e.getAttribute("crossorigin")!==(l.crossOrigin==null?null:l.crossOrigin)||e.getAttribute("title")!==(l.title==null?null:l.title))break;return e;case"style":if(e.hasAttribute("data-precedence"))break;return e;case"script":if(n=e.getAttribute("src"),(n!==(l.src==null?null:l.src)||e.getAttribute("type")!==(l.type==null?null:l.type)||e.getAttribute("crossorigin")!==(l.crossOrigin==null?null:l.crossOrigin))&&n&&e.hasAttribute("async")&&!e.hasAttribute("itemprop"))break;return e;default:return e}}else if(t==="input"&&e.type==="hidden"){var n=l.name==null?null:""+l.name;if(l.type==="hidden"&&e.getAttribute("name")===n)return e}else return e;if(e=ga(e.nextSibling),e===null)break}return null}function jw(e,t,a){if(t==="")return null;for(;e.nodeType!==3;)if((e.nodeType!==1||e.nodeName!=="INPUT"||e.type!=="hidden")&&!a||(e=ga(e.nextSibling),e===null))return null;return e}function Jx(e,t){for(;e.nodeType!==8;)if((e.nodeType!==1||e.nodeName!=="INPUT"||e.type!=="hidden")&&!t||(e=ga(e.nextSibling),e===null))return null;return e}function Cd(e){return e.data==="$?"||e.data==="$~"}function Rd(e){return e.data==="$!"||e.data==="$?"&&e.ownerDocument.readyState!=="loading"}function Yw(e,t){var a=e.ownerDocument;if(e.data==="$~")e._reactRetry=t;else if(e.data!=="$?"||a.readyState!=="loading")t();else{var o=function(){t(),a.removeEventListener("DOMContentLoaded",o)};a.addEventListener("DOMContentLoaded",o),e._reactRetry=o}}function ga(e){for(;e!=null;e=e.nextSibling){var t=e.nodeType;if(t===1||t===3)break;if(t===8){if(t=e.data,t==="$"||t==="$!"||t==="$?"||t==="$~"||t==="&"||t==="F!"||t==="F")break;if(t==="/$"||t==="/&")return null}}return e}var Id=null;function cv(e){e=e.nextSibling;for(var t=0;e;){if(e.nodeType===8){var a=e.data;if(a==="/$"||a==="/&"){if(t===0)return ga(e.nextSibling);t--}else a!=="$"&&a!=="$!"&&a!=="$?"&&a!=="$~"&&a!=="&"||t++}e=e.nextSibling}return null}function dv(e){e=e.previousSibling;for(var t=0;e;){if(e.nodeType===8){var a=e.data;if(a==="$"||a==="$!"||a==="$?"||a==="$~"||a==="&"){if(t===0)return e;t--}else a!=="/$"&&a!=="/&"||t++}e=e.previousSibling}return null}function $x(e,t,a){switch(t=$i(a),e){case"html":if(e=t.documentElement,!e)throw Error(E(452));return e;case"head":if(e=t.head,!e)throw Error(E(453));return e;case"body":if(e=t.body,!e)throw Error(E(454));return e;default:throw Error(E(451))}}function Vu(e){for(var t=e.attributes;t.length;)e.removeAttributeNode(t[0]);Od(e)}var va=new Map,pv=new Set;function es(e){return typeof e.getRootNode=="function"?e.getRootNode():e.nodeType===9?e:e.ownerDocument}var Io=ie.d;ie.d={f:Kw,r:Zw,D:Qw,C:Ww,L:Jw,m:$w,X:tA,S:eA,M:aA};function Kw(){var e=Io.f(),t=vs();return e||t}function Zw(e){var t=Vn(e);t!==null&&t.tag===5&&t.type==="form"?jS(t):Io.r(e)}var Kn=typeof document=="undefined"?null:document;function eL(e,t,a){var o=Kn;if(o&&typeof t=="string"&&t){var l=da(t);l='link[rel="'+e+'"][href="'+l+'"]',typeof a=="string"&&(l+='[crossorigin="'+a+'"]'),pv.has(l)||(pv.add(l),e={rel:e,crossOrigin:a,href:t},o.querySelector(l)===null&&(t=o.createElement("link"),vt(t,"link",e),lt(t),o.head.appendChild(t)))}}function Qw(e){Io.D(e),eL("dns-prefetch",e,null)}function Ww(e,t){Io.C(e,t),eL("preconnect",e,t)}function Jw(e,t,a){Io.L(e,t,a);var o=Kn;if(o&&e&&t){var l='link[rel="preload"][as="'+da(t)+'"]';t==="image"&&a&&a.imageSrcSet?(l+='[imagesrcset="'+da(a.imageSrcSet)+'"]',typeof a.imageSizes=="string"&&(l+='[imagesizes="'+da(a.imageSizes)+'"]')):l+='[href="'+da(e)+'"]';var n=l;switch(t){case"style":n=Fn(e);break;case"script":n=Zn(e)}va.has(n)||(e=Ie({rel:"preload",href:t==="image"&&a&&a.imageSrcSet?void 0:e,as:t},a),va.set(n,e),o.querySelector(l)!==null||t==="style"&&o.querySelector(hr(n))||t==="script"&&o.querySelector(gr(n))||(t=o.createElement("link"),vt(t,"link",e),lt(t),o.head.appendChild(t)))}}function $w(e,t){Io.m(e,t);var a=Kn;if(a&&e){var o=t&&typeof t.as=="string"?t.as:"script",l='link[rel="modulepreload"][as="'+da(o)+'"][href="'+da(e)+'"]',n=l;switch(o){case"audioworklet":case"paintworklet":case"serviceworker":case"sharedworker":case"worker":case"script":n=Zn(e)}if(!va.has(n)&&(e=Ie({rel:"modulepreload",href:e},t),va.set(n,e),a.querySelector(l)===null)){switch(o){case"audioworklet":case"paintworklet":case"serviceworker":case"sharedworker":case"worker":case"script":if(a.querySelector(gr(n)))return}o=a.createElement("link"),vt(o,"link",e),lt(o),a.head.appendChild(o)}}}function eA(e,t,a){Io.S(e,t,a);var o=Kn;if(o&&e){var l=bn(o).hoistableStyles,n=Fn(e);t=t||"default";var u=l.get(n);if(!u){var r={loading:0,preload:null};if(u=o.querySelector(hr(n)))r.loading=5;else{e=Ie({rel:"stylesheet",href:e,"data-precedence":t},a),(a=va.get(n))&&vp(e,a);var i=u=o.createElement("link");lt(i),vt(i,"link",e),i._p=new Promise(function(s,f){i.onload=s,i.onerror=f}),i.addEventListener("load",function(){r.loading|=1}),i.addEventListener("error",function(){r.loading|=2}),r.loading|=4,wi(u,t,o)}u={type:"stylesheet",instance:u,count:1,state:r},l.set(n,u)}}}function tA(e,t){Io.X(e,t);var a=Kn;if(a&&e){var o=bn(a).hoistableScripts,l=Zn(e),n=o.get(l);n||(n=a.querySelector(gr(l)),n||(e=Ie({src:e,async:!0},t),(t=va.get(l))&&Sp(e,t),n=a.createElement("script"),lt(n),vt(n,"link",e),a.head.appendChild(n)),n={type:"script",instance:n,count:1,state:null},o.set(l,n))}}function aA(e,t){Io.M(e,t);var a=Kn;if(a&&e){var o=bn(a).hoistableScripts,l=Zn(e),n=o.get(l);n||(n=a.querySelector(gr(l)),n||(e=Ie({src:e,async:!0,type:"module"},t),(t=va.get(l))&&Sp(e,t),n=a.createElement("script"),lt(n),vt(n,"link",e),a.head.appendChild(n)),n={type:"script",instance:n,count:1,state:null},o.set(l,n))}}function mv(e,t,a,o){var l=(l=Zo.current)?es(l):null;if(!l)throw Error(E(446));switch(e){case"meta":case"title":return null;case"style":return typeof a.precedence=="string"&&typeof a.href=="string"?(t=Fn(a.href),a=bn(l).hoistableStyles,o=a.get(t),o||(o={type:"style",instance:null,count:0,state:null},a.set(t,o)),o):{type:"void",instance:null,count:0,state:null};case"link":if(a.rel==="stylesheet"&&typeof a.href=="string"&&typeof a.precedence=="string"){e=Fn(a.href);var n=bn(l).hoistableStyles,u=n.get(e);if(u||(l=l.ownerDocument||l,u={type:"stylesheet",instance:null,count:0,state:{loading:0,preload:null}},n.set(e,u),(n=l.querySelector(hr(e)))&&!n._p&&(u.instance=n,u.state.loading=5),va.has(e)||(a={rel:"preload",as:"style",href:a.href,crossOrigin:a.crossOrigin,integrity:a.integrity,media:a.media,hrefLang:a.hrefLang,referrerPolicy:a.referrerPolicy},va.set(e,a),n||oA(l,e,a,u.state))),t&&o===null)throw Error(E(528,""));return u}if(t&&o!==null)throw Error(E(529,""));return null;case"script":return t=a.async,a=a.src,typeof a=="string"&&t&&typeof t!="function"&&typeof t!="symbol"?(t=Zn(a),a=bn(l).hoistableScripts,o=a.get(t),o||(o={type:"script",instance:null,count:0,state:null},a.set(t,o)),o):{type:"void",instance:null,count:0,state:null};default:throw Error(E(444,e))}}function Fn(e){return'href="'+da(e)+'"'}function hr(e){return'link[rel="stylesheet"]['+e+"]"}function tL(e){return Ie({},e,{"data-precedence":e.precedence,precedence:null})}function oA(e,t,a,o){e.querySelector('link[rel="preload"][as="style"]['+t+"]")?o.loading=1:(t=e.createElement("link"),o.preload=t,t.addEventListener("load",function(){return o.loading|=1}),t.addEventListener("error",function(){return o.loading|=2}),vt(t,"link",a),lt(t),e.head.appendChild(t))}function Zn(e){return'[src="'+da(e)+'"]'}function gr(e){return"script[async]"+e}function hv(e,t,a){if(t.count++,t.instance===null)switch(t.type){case"style":var o=e.querySelector('style[data-href~="'+da(a.href)+'"]');if(o)return t.instance=o,lt(o),o;var l=Ie({},a,{"data-href":a.href,"data-precedence":a.precedence,href:null,precedence:null});return o=(e.ownerDocument||e).createElement("style"),lt(o),vt(o,"style",l),wi(o,a.precedence,e),t.instance=o;case"stylesheet":l=Fn(a.href);var n=e.querySelector(hr(l));if(n)return t.state.loading|=4,t.instance=n,lt(n),n;o=tL(a),(l=va.get(l))&&vp(o,l),n=(e.ownerDocument||e).createElement("link"),lt(n);var u=n;return u._p=new Promise(function(r,i){u.onload=r,u.onerror=i}),vt(n,"link",o),t.state.loading|=4,wi(n,a.precedence,e),t.instance=n;case"script":return n=Zn(a.src),(l=e.querySelector(gr(n)))?(t.instance=l,lt(l),l):(o=a,(l=va.get(n))&&(o=Ie({},a),Sp(o,l)),e=e.ownerDocument||e,l=e.createElement("script"),lt(l),vt(l,"link",o),e.head.appendChild(l),t.instance=l);case"void":return null;default:throw Error(E(443,t.type))}else t.type==="stylesheet"&&(t.state.loading&4)===0&&(o=t.instance,t.state.loading|=4,wi(o,a.precedence,e));return t.instance}function wi(e,t,a){for(var o=a.querySelectorAll('link[rel="stylesheet"][data-precedence],style[data-precedence]'),l=o.length?o[o.length-1]:null,n=l,u=0;u<o.length;u++){var r=o[u];if(r.dataset.precedence===t)n=r;else if(n!==l)break}n?n.parentNode.insertBefore(e,n.nextSibling):(t=a.nodeType===9?a.head:a,t.insertBefore(e,t.firstChild))}function vp(e,t){e.crossOrigin==null&&(e.crossOrigin=t.crossOrigin),e.referrerPolicy==null&&(e.referrerPolicy=t.referrerPolicy),e.title==null&&(e.title=t.title)}function Sp(e,t){e.crossOrigin==null&&(e.crossOrigin=t.crossOrigin),e.referrerPolicy==null&&(e.referrerPolicy=t.referrerPolicy),e.integrity==null&&(e.integrity=t.integrity)}var Ai=null;function gv(e,t,a){if(Ai===null){var o=new Map,l=Ai=new Map;l.set(a,o)}else l=Ai,o=l.get(a),o||(o=new Map,l.set(a,o));if(o.has(e))return o;for(o.set(e,null),a=a.getElementsByTagName(e),l=0;l<a.length;l++){var n=a[l];if(!(n[sr]||n[mt]||e==="link"&&n.getAttribute("rel")==="stylesheet")&&n.namespaceURI!=="http://www.w3.org/2000/svg"){var u=n.getAttribute(t)||"";u=e+u;var r=o.get(u);r?r.push(n):o.set(u,[n])}}return o}function vv(e,t,a){e=e.ownerDocument||e,e.head.insertBefore(a,t==="title"?e.querySelector("head > title"):null)}function lA(e,t,a){if(a===1||t.itemProp!=null)return!1;switch(e){case"meta":case"title":return!0;case"style":if(typeof t.precedence!="string"||typeof t.href!="string"||t.href==="")break;return!0;case"link":if(typeof t.rel!="string"||typeof t.href!="string"||t.href===""||t.onLoad||t.onError)break;return t.rel==="stylesheet"?(e=t.disabled,typeof t.precedence=="string"&&e==null):!0;case"script":if(t.async&&typeof t.async!="function"&&typeof t.async!="symbol"&&!t.onLoad&&!t.onError&&t.src&&typeof t.src=="string")return!0}return!1}function aL(e){return!(e.type==="stylesheet"&&(e.state.loading&3)===0)}function nA(e,t,a,o){if(a.type==="stylesheet"&&(typeof o.media!="string"||matchMedia(o.media).matches!==!1)&&(a.state.loading&4)===0){if(a.instance===null){var l=Fn(o.href),n=t.querySelector(hr(l));if(n){t=n._p,t!==null&&typeof t=="object"&&typeof t.then=="function"&&(e.count++,e=ts.bind(e),t.then(e,e)),a.state.loading|=4,a.instance=n,lt(n);return}n=t.ownerDocument||t,o=tL(o),(l=va.get(l))&&vp(o,l),n=n.createElement("link"),lt(n);var u=n;u._p=new Promise(function(r,i){u.onload=r,u.onerror=i}),vt(n,"link",o),a.instance=n}e.stylesheets===null&&(e.stylesheets=new Map),e.stylesheets.set(a,t),(t=a.state.preload)&&(a.state.loading&3)===0&&(e.count++,a=ts.bind(e),t.addEventListener("load",a),t.addEventListener("error",a))}}var Mc=0;function uA(e,t){return e.stylesheets&&e.count===0&&Ti(e,e.stylesheets),0<e.count||0<e.imgCount?function(a){var o=setTimeout(function(){if(e.stylesheets&&Ti(e,e.stylesheets),e.unsuspend){var n=e.unsuspend;e.unsuspend=null,n()}},6e4+t);0<e.imgBytes&&Mc===0&&(Mc=62500*zw());var l=setTimeout(function(){if(e.waitingForImages=!1,e.count===0&&(e.stylesheets&&Ti(e,e.stylesheets),e.unsuspend)){var n=e.unsuspend;e.unsuspend=null,n()}},(e.imgBytes>Mc?50:800)+t);return e.unsuspend=a,function(){e.unsuspend=null,clearTimeout(o),clearTimeout(l)}}:null}function ts(){if(this.count--,this.count===0&&(this.imgCount===0||!this.waitingForImages)){if(this.stylesheets)Ti(this,this.stylesheets);else if(this.unsuspend){var e=this.unsuspend;this.unsuspend=null,e()}}}var as=null;function Ti(e,t){e.stylesheets=null,e.unsuspend!==null&&(e.count++,as=new Map,t.forEach(rA,e),as=null,ts.call(e))}function rA(e,t){if(!(t.state.loading&4)){var a=as.get(e);if(a)var o=a.get(null);else{a=new Map,as.set(e,a);for(var l=e.querySelectorAll("link[data-precedence],style[data-precedence]"),n=0;n<l.length;n++){var u=l[n];(u.nodeName==="LINK"||u.getAttribute("media")!=="not all")&&(a.set(u.dataset.precedence,u),o=u)}o&&a.set(null,o)}l=t.instance,u=l.getAttribute("data-precedence"),n=a.get(u)||o,n===o&&a.set(null,l),a.set(u,l),this.count++,o=ts.bind(this),l.addEventListener("load",o),l.addEventListener("error",o),n?n.parentNode.insertBefore(l,n.nextSibling):(e=e.nodeType===9?e.head:e,e.insertBefore(l,e.firstChild)),t.state.loading|=4}}var ar={$$typeof:po,Provider:null,Consumer:null,_currentValue:bl,_currentValue2:bl,_threadCount:0};function iA(e,t,a,o,l,n,u,r,i){this.tag=1,this.containerInfo=e,this.pingCache=this.current=this.pendingChildren=null,this.timeoutHandle=-1,this.callbackNode=this.next=this.pendingContext=this.context=this.cancelPendingCommit=null,this.callbackPriority=0,this.expirationTimes=oc(-1),this.entangledLanes=this.shellSuspendCounter=this.errorRecoveryDisabledLanes=this.expiredLanes=this.warmLanes=this.pingedLanes=this.suspendedLanes=this.pendingLanes=0,this.entanglements=oc(0),this.hiddenUpdates=oc(null),this.identifierPrefix=o,this.onUncaughtError=l,this.onCaughtError=n,this.onRecoverableError=u,this.pooledCache=null,this.pooledCacheLanes=0,this.formState=i,this.incompleteTransitions=new Map}function oL(e,t,a,o,l,n,u,r,i,s,f,d){return e=new iA(e,t,a,u,i,s,f,d,r),t=1,n===!0&&(t|=24),n=Zt(3,null,null,t),e.current=n,n.stateNode=e,t=Vd(),t.refCount++,e.pooledCache=t,t.refCount++,n.memoizedState={element:o,isDehydrated:a,cache:t},Yd(n),e}function lL(e){return e?(e=yn,e):yn}function nL(e,t,a,o,l,n){l=lL(l),o.context===null?o.context=l:o.pendingContext=l,o=Wo(t),o.payload={element:a},n=n===void 0?null:n,n!==null&&(o.callback=n),a=Jo(e,o,t),a!==null&&(kt(a,e,t),Uu(a,e,t))}function Sv(e,t){if(e=e.memoizedState,e!==null&&e.dehydrated!==null){var a=e.retryLane;e.retryLane=a!==0&&a<t?a:t}}function xp(e,t){Sv(e,t),(e=e.alternate)&&Sv(e,t)}function uL(e){if(e.tag===13||e.tag===31){var t=kl(e,67108864);t!==null&&kt(t,e,67108864),xp(e,67108864)}}function xv(e){if(e.tag===13||e.tag===31){var t=ea();t=Dd(t);var a=kl(e,t);a!==null&&kt(a,e,t),xp(e,t)}}var os=!0;function sA(e,t,a,o){var l=V.T;V.T=null;var n=ie.p;try{ie.p=2,Lp(e,t,a,o)}finally{ie.p=n,V.T=l}}function fA(e,t,a,o){var l=V.T;V.T=null;var n=ie.p;try{ie.p=8,Lp(e,t,a,o)}finally{ie.p=n,V.T=l}}function Lp(e,t,a,o){if(os){var l=bd(o);if(l===null)Ec(e,t,o,ls,a),Lv(e,o);else if(dA(l,e,t,a,o))o.stopPropagation();else if(Lv(e,o),t&4&&-1<cA.indexOf(e)){for(;l!==null;){var n=Vn(l);if(n!==null)switch(n.tag){case 3:if(n=n.stateNode,n.current.memoizedState.isDehydrated){var u=Cl(n.pendingLanes);if(u!==0){var r=n;for(r.pendingLanes|=2,r.entangledLanes|=2;u;){var i=1<<31-$t(u);r.entanglements[1]|=i,u&=~i}Qa(n),(re&6)===0&&(Yi=Wt()+500,mr(0,!1))}}break;case 31:case 13:r=kl(n,2),r!==null&&kt(r,n,2),vs(),xp(n,2)}if(n=bd(o),n===null&&Ec(e,t,o,ls,a),n===l)break;l=n}l!==null&&o.stopPropagation()}else Ec(e,t,o,null,a)}}function bd(e){return e=Bd(e),yp(e)}var ls=null;function yp(e){if(ls=null,e=hn(e),e!==null){var t=nr(e);if(t===null)e=null;else{var a=t.tag;if(a===13){if(e=wv(t),e!==null)return e;e=null}else if(a===31){if(e=Av(t),e!==null)return e;e=null}else if(a===3){if(t.stateNode.current.memoizedState.isDehydrated)return t.tag===3?t.stateNode.containerInfo:null;e=null}else t!==e&&(e=null)}}return ls=e,null}function rL(e){switch(e){case"beforetoggle":case"cancel":case"click":case"close":case"contextmenu":case"copy":case"cut":case"auxclick":case"dblclick":case"dragend":case"dragstart":case"drop":case"focusin":case"focusout":case"input":case"invalid":case"keydown":case"keypress":case"keyup":case"mousedown":case"mouseup":case"paste":case"pause":case"play":case"pointercancel":case"pointerdown":case"pointerup":case"ratechange":case"reset":case"resize":case"seeked":case"submit":case"toggle":case"touchcancel":case"touchend":case"touchstart":case"volumechange":case"change":case"selectionchange":case"textInput":case"compositionstart":case"compositionend":case"compositionupdate":case"beforeblur":case"afterblur":case"beforeinput":case"blur":case"fullscreenchange":case"focus":case"hashchange":case"popstate":case"select":case"selectstart":return 2;case"drag":case"dragenter":case"dragexit":case"dragleave":case"dragover":case"mousemove":case"mouseout":case"mouseover":case"pointermove":case"pointerout":case"pointerover":case"scroll":case"touchmove":case"wheel":case"mouseenter":case"mouseleave":case"pointerenter":case"pointerleave":return 8;case"message":switch(Jb()){case Mv:return 2;case Ov:return 8;case Pi:case $b:return 32;case Pv:return 268435456;default:return 32}default:return 32}}var wd=!1,tl=null,al=null,ol=null,or=new Map,lr=new Map,Go=[],cA="mousedown mouseup touchcancel touchend touchstart auxclick dblclick pointercancel pointerdown pointerup dragend dragstart drop compositionend compositionstart keydown keypress keyup input textInput copy cut paste click change contextmenu reset".split(" ");function Lv(e,t){switch(e){case"focusin":case"focusout":tl=null;break;case"dragenter":case"dragleave":al=null;break;case"mouseover":case"mouseout":ol=null;break;case"pointerover":case"pointerout":or.delete(t.pointerId);break;case"gotpointercapture":case"lostpointercapture":lr.delete(t.pointerId)}}function bu(e,t,a,o,l,n){return e===null||e.nativeEvent!==n?(e={blockedOn:t,domEventName:a,eventSystemFlags:o,nativeEvent:n,targetContainers:[l]},t!==null&&(t=Vn(t),t!==null&&uL(t)),e):(e.eventSystemFlags|=o,t=e.targetContainers,l!==null&&t.indexOf(l)===-1&&t.push(l),e)}function dA(e,t,a,o,l){switch(t){case"focusin":return tl=bu(tl,e,t,a,o,l),!0;case"dragenter":return al=bu(al,e,t,a,o,l),!0;case"mouseover":return ol=bu(ol,e,t,a,o,l),!0;case"pointerover":var n=l.pointerId;return or.set(n,bu(or.get(n)||null,e,t,a,o,l)),!0;case"gotpointercapture":return n=l.pointerId,lr.set(n,bu(lr.get(n)||null,e,t,a,o,l)),!0}return!1}function iL(e){var t=hn(e.target);if(t!==null){var a=nr(t);if(a!==null){if(t=a.tag,t===13){if(t=wv(a),t!==null){e.blockedOn=t,og(e.priority,function(){xv(a)});return}}else if(t===31){if(t=Av(a),t!==null){e.blockedOn=t,og(e.priority,function(){xv(a)});return}}else if(t===3&&a.stateNode.current.memoizedState.isDehydrated){e.blockedOn=a.tag===3?a.stateNode.containerInfo:null;return}}}e.blockedOn=null}function Ei(e){if(e.blockedOn!==null)return!1;for(var t=e.targetContainers;0<t.length;){var a=bd(e.nativeEvent);if(a===null){a=e.nativeEvent;var o=new a.constructor(a.type,a);Vc=o,a.target.dispatchEvent(o),Vc=null}else return t=Vn(a),t!==null&&uL(t),e.blockedOn=a,!1;t.shift()}return!0}function yv(e,t,a){Ei(e)&&a.delete(t)}function pA(){wd=!1,tl!==null&&Ei(tl)&&(tl=null),al!==null&&Ei(al)&&(al=null),ol!==null&&Ei(ol)&&(ol=null),or.forEach(yv),lr.forEach(yv)}function pi(e,t){e.blockedOn===t&&(e.blockedOn=null,wd||(wd=!0,Ze.unstable_scheduleCallback(Ze.unstable_NormalPriority,pA)))}var mi=null;function Cv(e){mi!==e&&(mi=e,Ze.unstable_scheduleCallback(Ze.unstable_NormalPriority,function(){mi===e&&(mi=null);for(var t=0;t<e.length;t+=3){var a=e[t],o=e[t+1],l=e[t+2];if(typeof o!="function"){if(yp(o||a)===null)continue;break}var n=Vn(a);n!==null&&(e.splice(t,3),t-=3,ud(n,{pending:!0,data:l,method:a.method,action:o},o,l))}}))}function qn(e){function t(i){return pi(i,e)}tl!==null&&pi(tl,e),al!==null&&pi(al,e),ol!==null&&pi(ol,e),or.forEach(t),lr.forEach(t);for(var a=0;a<Go.length;a++){var o=Go[a];o.blockedOn===e&&(o.blockedOn=null)}for(;0<Go.length&&(a=Go[0],a.blockedOn===null);)iL(a),a.blockedOn===null&&Go.shift();if(a=(e.ownerDocument||e).$$reactFormReplay,a!=null)for(o=0;o<a.length;o+=3){var l=a[o],n=a[o+1],u=l[Ht]||null;if(typeof n=="function")u||Cv(a);else if(u){var r=null;if(n&&n.hasAttribute("formAction")){if(l=n,u=n[Ht]||null)r=u.formAction;else if(yp(l)!==null)continue}else r=u.action;typeof r=="function"?a[o+1]=r:(a.splice(o,3),o-=3),Cv(a)}}}function sL(){function e(n){n.canIntercept&&n.info==="react-transition"&&n.intercept({handler:function(){return new Promise(function(u){return l=u})},focusReset:"manual",scroll:"manual"})}function t(){l!==null&&(l(),l=null),o||setTimeout(a,20)}function a(){if(!o&&!navigation.transition){var n=navigation.currentEntry;n&&n.url!=null&&navigation.navigate(n.url,{state:n.getState(),info:"react-transition",history:"replace"})}}if(typeof navigation=="object"){var o=!1,l=null;return navigation.addEventListener("navigate",e),navigation.addEventListener("navigatesuccess",t),navigation.addEventListener("navigateerror",t),setTimeout(a,100),function(){o=!0,navigation.removeEventListener("navigate",e),navigation.removeEventListener("navigatesuccess",t),navigation.removeEventListener("navigateerror",t),l!==null&&(l(),l=null)}}}function Cp(e){this._internalRoot=e}Ls.prototype.render=Cp.prototype.render=function(e){var t=this._internalRoot;if(t===null)throw Error(E(409));var a=t.current,o=ea();nL(a,o,e,t,null,null)};Ls.prototype.unmount=Cp.prototype.unmount=function(){var e=this._internalRoot;if(e!==null){this._internalRoot=null;var t=e.containerInfo;nL(e.current,2,null,e,null,null),vs(),t[Gn]=null}};function Ls(e){this._internalRoot=e}Ls.prototype.unstable_scheduleHydration=function(e){if(e){var t=kv();e={blockedOn:null,target:e,priority:t};for(var a=0;a<Go.length&&t!==0&&t<Go[a].priority;a++);Go.splice(a,0,e),a===0&&iL(e)}};var Rv=Iv.version;if(Rv!=="19.2.0")throw Error(E(527,Rv,"19.2.0"));ie.findDOMNode=function(e){var t=e._reactInternals;if(t===void 0)throw typeof e.render=="function"?Error(E(188)):(e=Object.keys(e).join(","),Error(E(268,e)));return e=Xb(t),e=e!==null?Tv(e):null,e=e===null?null:e.stateNode,e};var mA={bundleType:0,version:"19.2.0",rendererPackageName:"react-dom",currentDispatcherRef:V,reconcilerVersion:"19.2.0"};if(typeof __REACT_DEVTOOLS_GLOBAL_HOOK__!="undefined"&&(wu=__REACT_DEVTOOLS_GLOBAL_HOOK__,!wu.isDisabled&&wu.supportsFiber))try{ur=wu.inject(mA),Jt=wu}catch(e){}var wu;ys.createRoot=function(e,t){if(!bv(e))throw Error(E(299));var a=!1,o="",l=ex,n=tx,u=ax;return t!=null&&(t.unstable_strictMode===!0&&(a=!0),t.identifierPrefix!==void 0&&(o=t.identifierPrefix),t.onUncaughtError!==void 0&&(l=t.onUncaughtError),t.onCaughtError!==void 0&&(n=t.onCaughtError),t.onRecoverableError!==void 0&&(u=t.onRecoverableError)),t=oL(e,1,!1,null,null,a,o,null,l,n,u,sL),e[Gn]=t.current,gp(e),new Cp(t)};ys.hydrateRoot=function(e,t,a){if(!bv(e))throw Error(E(299));var o=!1,l="",n=ex,u=tx,r=ax,i=null;return a!=null&&(a.unstable_strictMode===!0&&(o=!0),a.identifierPrefix!==void 0&&(l=a.identifierPrefix),a.onUncaughtError!==void 0&&(n=a.onUncaughtError),a.onCaughtError!==void 0&&(u=a.onCaughtError),a.onRecoverableError!==void 0&&(r=a.onRecoverableError),a.formState!==void 0&&(i=a.formState)),t=oL(e,1,!0,t,a!=null?a:null,o,l,i,n,u,r,sL),t.context=lL(null),a=t.current,o=ea(),o=Dd(o),l=Wo(o),l.callback=null,Jo(a,l,o),a=o,t.current.lanes=a,ir(t,a),Qa(t),e[Gn]=t.current,gp(e),new Ls(t)};ys.version="19.2.0"});var pL=qa((B1,dL)=>{"use strict";function cL(){if(!(typeof __REACT_DEVTOOLS_GLOBAL_HOOK__=="undefined"||typeof __REACT_DEVTOOLS_GLOBAL_HOOK__.checkDCE!="function"))try{__REACT_DEVTOOLS_GLOBAL_HOOK__.checkDCE(cL)}catch(e){console.error(e)}}cL(),dL.exports=fL()});var LL=qa(Cs=>{"use strict";/**
 * @license React
 * react-jsx-runtime.production.js
 *
 * Copyright (c) Meta Platforms, Inc. and affiliates.
 *
 * This source code is licensed under the MIT license found in the
 * LICENSE file in the root directory of this source tree.
 */var SA=Symbol.for("react.transitional.element"),xA=Symbol.for("react.fragment");function xL(e,t,a){var o=null;if(a!==void 0&&(o=""+a),t.key!==void 0&&(o=""+t.key),"key"in t){a={};for(var l in t)l!=="key"&&(a[l]=t[l])}else a=t;return t=a.ref,{$$typeof:SA,type:e,key:o,ref:t!==void 0?t:null,props:a}}Cs.Fragment=xA;Cs.jsx=xL;Cs.jsxs=xL});var X=qa((k1,yL)=>{"use strict";yL.exports=LL()});var Gr=T(pL()),mu=T(Va());var Pa=T(B());var Gt=T(B(),1);var hA=Object.defineProperty,Qn=(e,t)=>hA(e,"name",{value:t,configurable:!0}),mL=!!(typeof window!="undefined"&&window.document&&window.document.createElement);function Ee(e,t,{checkForDefaultPrevented:a=!0}={}){return Qn(function(l){if(e==null||e(l),a===!1||!l||!l.defaultPrevented)return t==null?void 0:t(l)},"handleEvent")}Qn(Ee,"composeEventHandlers");function gA(e){var t,a;if(!mL)throw new Error("Cannot access window outside of the DOM");return(a=(t=e==null?void 0:e.ownerDocument)==null?void 0:t.defaultView)!=null?a:window}Qn(gA,"getOwnerWindow");function Rp(e){var t;if(!mL)throw new Error("Cannot access document outside of the DOM");return(t=e==null?void 0:e.ownerDocument)!=null?t:document}Qn(Rp,"getOwnerDocument");function hL(e,t=!1){let{activeElement:a}=Rp(e);if(!(a!=null&&a.nodeName))return null;if(gL(a)&&a.contentDocument)return hL(a.contentDocument.body,t);if(t){let o=a.getAttribute("aria-activedescendant");if(o){let l=Rp(a).getElementById(o);if(l)return l}}return a}Qn(hL,"getActiveElement");function gL(e){return e.tagName==="IFRAME"}Qn(gL,"isFrame");var vL=T(B(),1),vA=Object.defineProperty,bp=(e,t)=>vA(e,"name",{value:t,configurable:!0});function Ip(e,t){if(typeof e=="function")return e(t);e!=null&&(e.current=t)}bp(Ip,"setRef");function SL(...e){return t=>{let a=!1,o=e.map(l=>{let n=Ip(l,t);return!a&&typeof n=="function"&&(a=!0),n});if(a)return()=>{for(let l=0;l<o.length;l++){let n=o[l];typeof n=="function"?n():Ip(e[l],null)}}}}bp(SL,"composeRefs");function ce(...e){return vL.useCallback(SL(...e),e)}bp(ce,"useComposedRefs");var xa=T(B(),1),wp=T(X(),1);var LA=Object.defineProperty,Sa=(e,t)=>LA(e,"name",{value:t,configurable:!0});function yA(e,t){let a=xa.createContext(t);a.displayName=e+"Context";let o=Sa(n=>{let s=n,{children:u}=s,r=D(s,["children"]),i=xa.useMemo(()=>r,Object.values(r));return(0,wp.jsx)(a.Provider,{value:i,children:u})},"Provider");o.displayName=e+"Provider";function l(n,u={}){let{optional:r=!1}=u,i=xa.useContext(a);if(i)return i;if(t!==void 0)return t;if(!r)throw new Error("\x60".concat(n,"\x60 must be used within \x60").concat(e,"\x60"))}return Sa(l,"useContext"),[o,l]}Sa(yA,"createContext");function Et(e,t=[]){let a=[];function o(n,u){let r=xa.createContext(u);r.displayName=n+"Context";let i=a.length;a=[...a,u];let s=Sa(d=>{var p;let h=d,{scope:c,children:g}=h,L=D(h,["scope","children"]),x=((p=c==null?void 0:c[e])==null?void 0:p[i])||r,S=xa.useMemo(()=>L,Object.values(L));return(0,wp.jsx)(x.Provider,{value:S,children:g})},"Provider");s.displayName=n+"Provider";function f(d,c,g={}){var h;let{optional:L=!1}=g,x=((h=c==null?void 0:c[e])==null?void 0:h[i])||r,S=xa.useContext(x);if(S)return S;if(u!==void 0)return u;if(!L)throw new Error("\x60".concat(d,"\x60 must be used within \x60").concat(n,"\x60"))}return Sa(f,"useContext"),[s,f]}Sa(o,"createContext");let l=Sa(()=>{let n=a.map(u=>xa.createContext(u));return Sa(function(r){let i=(r==null?void 0:r[e])||n;return xa.useMemo(()=>({["__scope".concat(e)]:b(y({},r),{[e]:i})}),[r,i])},"useScope")},"createScope");return l.scopeName=e,[o,CL(l,...t)]}Sa(Et,"createContextScope");function CL(...e){let t=e[0];if(e.length===1)return t;let a=Sa(()=>{let o=e.map(l=>({useScope:l(),scopeName:l.scopeName}));return Sa(function(n){let u=o.reduce((r,{useScope:i,scopeName:s})=>{let d=i(n)["__scope".concat(s)];return y(y({},r),d)},{});return xa.useMemo(()=>({["__scope".concat(t.scopeName)]:u}),[u])},"useComposedScopes")},"createScope");return a.scopeName=t.scopeName,a}Sa(CL,"composeContextScopes");var EL=T(B(),1),DL=T(Va(),1);var St=T(B(),1);var CA=Object.defineProperty,La=(e,t)=>CA(e,"name",{value:t,configurable:!0});function Da(e){let t=St.forwardRef((a,o)=>{var g;let c=a,{children:l}=c,n=D(c,["children"]),u=null,r=!1,i=[];Ap(l)&&typeof Rs=="function"&&(l=Rs(l._payload)),St.Children.forEach(l,L=>{var x;if(wL(L)){r=!0;let S=L,h="child"in S.props?S.props.child:S.props.children;Ap(h)&&typeof Rs=="function"&&(h=Rs(h._payload)),u=IA(S,h),i.push((x=u==null?void 0:u.props)==null?void 0:x.children)}else i.push(L)}),u?u=St.cloneElement(u,void 0,i):!r&&St.Children.count(l)===1&&St.isValidElement(l)&&(u=l);let s=u?bL(u):void 0,f=ce(o,s);if(!u){if(l||l===0)throw new Error(r?AA(e):wA(e));return l}let d=IL(n,(g=u.props)!=null?g:{});return u.type!==St.Fragment&&(d.ref=o?f:s),St.cloneElement(u,d)});return t.displayName="".concat(e,".Slot"),t}La(Da,"createSlot");var RL=Symbol.for("radix.slottable");function RA(e){let t=La(a=>"child"in a?a.children(a.child):a.children,"Slottable");return t.displayName="".concat(e,".Slottable"),t.__radixId=RL,t}La(RA,"createSlottable");var IA=La((e,t)=>{if("child"in e.props){let a=e.props.child;return St.isValidElement(a)?St.cloneElement(a,void 0,e.props.children(a.props.children)):null}return St.isValidElement(t)?t:null},"getSlottableElementFromSlottable");function IL(e,t){let a=y({},t);for(let o in t){let l=e[o],n=t[o];/^on[A-Z]/.test(o)?l&&n?a[o]=(...r)=>{let i=n(...r);return l(...r),i}:l&&(a[o]=l):o==="style"?a[o]=y(y({},l),n):o==="className"?a[o]=[l,n].filter(Boolean).join(" "):o==="aria-describedby"&&(a[o]=TL(n,l))}return y(y({},e),a)}La(IL,"mergeProps");function bL(e){var o,l;let t=(o=Object.getOwnPropertyDescriptor(e.props,"ref"))==null?void 0:o.get,a=t&&"isReactWarning"in t&&t.isReactWarning;return a?e.ref:(t=(l=Object.getOwnPropertyDescriptor(e,"ref"))==null?void 0:l.get,a=t&&"isReactWarning"in t&&t.isReactWarning,a?e.props.ref:e.props.ref||e.ref)}La(bL,"getElementRef");function wL(e){return St.isValidElement(e)&&typeof e.type=="function"&&"__radixId"in e.type&&e.type.__radixId===RL}La(wL,"isSlottable");var bA=Symbol.for("react.lazy");function Ap(e){return e!=null&&typeof e=="object"&&"$$typeof"in e&&e.$$typeof===bA&&"_payload"in e&&AL(e._payload)}La(Ap,"isLazyComponent");function AL(e){return typeof e=="object"&&e!==null&&"then"in e}La(AL,"isPromiseLike");function TL(...e){let t=new Set;for(let a of e)if(typeof a=="string")for(let o of String(a).trim().split(/\s+/))o&&t.add(o);return t.size>0?Array.from(t).join(" "):void 0}La(TL,"concatAriaDescribedby");var wA=La(e=>"".concat(e," failed to slot onto its children. Expected a single React element child or \x60Slottable\x60."),"createSlotError"),AA=La(e=>"".concat(e," failed to slot onto its \x60Slottable\x60. Expected \x60Slottable\x60 to receive a single React element child."),"createSlottableError"),Rs=St[" use ".trim().toString()];var ML=T(X(),1);var TA=Object.defineProperty,EA=(e,t)=>TA(e,"name",{value:t,configurable:!0}),DA=["a","button","div","form","h2","h3","img","input","label","li","nav","ol","p","select","span","svg","ul"],Qe=DA.reduce((e,t)=>{let a=Da("Primitive.".concat(t)),o=EL.forwardRef((l,n)=>{let s=l,{asChild:u}=s,r=D(s,["asChild"]),i=u?a:t;return typeof window!="undefined"&&(window[Symbol.for("radix-ui")]=!0),(0,ML.jsx)(i,b(y({},r),{ref:n}))});return o.displayName="Primitive.".concat(t),b(y({},e),{[t]:o})},{});function Tp(e,t){e&&DL.flushSync(()=>e.dispatchEvent(t))}EA(Tp,"dispatchDiscreteCustomEvent");var rt=T(B(),1);var Ma=T(B(),1);var Is=T(X(),1),Ft=T(B(),1);var zl=T(X(),1);var MA=Object.defineProperty,ut=(e,t)=>MA(e,"name",{value:t,configurable:!0});function Mp(e){let t=e+"CollectionProvider",[a,o]=Et(t),[l,n]=a(t,{collectionRef:{current:null},itemMap:new Map}),u=ut(x=>{let{scope:S,children:h}=x,p=Ma.useRef(null),m=Ma.useRef(new Map).current;return(0,Is.jsx)(l,{scope:S,itemMap:m,collectionRef:p,children:h})},"CollectionProvider");u.displayName=t;let r=e+"CollectionSlot",i=Da(r),s=Ma.forwardRef((x,S)=>{let{scope:h,children:p}=x,m=n(r,h),v=ce(S,m.collectionRef);return(0,Is.jsx)(i,{ref:v,children:p})});s.displayName=r;let f=e+"CollectionItemSlot",d="data-radix-collection-item",c=Da(f),g=Ma.forwardRef((x,S)=>{let R=x,{scope:h,children:p}=R,m=D(R,["scope","children"]),v=Ma.useRef(null),C=ce(S,v),I=n(f,h);return Ma.useEffect(()=>(I.itemMap.set(v,y({ref:v},m)),()=>{I.itemMap.delete(v)})),(0,Is.jsx)(c,{[d]:"",ref:C,children:p})});g.displayName=f;function L(x){let S=n(e+"CollectionConsumer",x);return Ma.useCallback(()=>{let p=S.collectionRef.current;if(!p)return[];let m=Array.from(p.querySelectorAll("[".concat(d,"]")));return Array.from(S.itemMap.values()).sort((I,R)=>m.indexOf(I.ref.current)-m.indexOf(R.ref.current))},[S.collectionRef,S.itemMap])}return ut(L,"useCollection"),[{Provider:u,Slot:s,ItemSlot:g},L,o]}ut(Mp,"createCollection");var OL=new WeakMap,Ge,oa,Ep=(oa=class extends Map{constructor(a){super(a);Nf(this,Ge);Xr(this,Ge,[...super.keys()]),OL.set(this,!0)}set(a,o){return OL.get(this)&&(this.has(a)?dt(this,Ge)[dt(this,Ge).indexOf(a)]=a:dt(this,Ge).push(a)),super.set(a,o),this}insert(a,o,l){let n=this.has(o),u=dt(this,Ge).length,r=Op(a),i=r>=0?r:u+r,s=i<0||i>=u?-1:i;if(s===this.size||n&&s===this.size-1||s===-1)return this.set(o,l),this;let f=this.size+(n?0:1);r<0&&i++;let d=[...dt(this,Ge)],c,g=!1;for(let L=i;L<f;L++)if(i===L){let x=d[L];d[L]===o&&(x=d[L+1]),n&&this.delete(o),c=this.get(x),this.set(o,l)}else{!g&&d[L-1]===o&&(g=!0);let x=d[g?L:L-1],S=c;c=this.get(x),this.delete(x),this.set(x,S)}return this}with(a,o,l){let n=new oa(this);return n.insert(a,o,l),n}before(a){let o=dt(this,Ge).indexOf(a)-1;if(!(o<0))return this.entryAt(o)}setBefore(a,o,l){let n=dt(this,Ge).indexOf(a);return n===-1?this:this.insert(n,o,l)}after(a){let o=dt(this,Ge).indexOf(a);if(o=o===-1||o===this.size-1?-1:o+1,o!==-1)return this.entryAt(o)}setAfter(a,o,l){let n=dt(this,Ge).indexOf(a);return n===-1?this:this.insert(n+1,o,l)}first(){return this.entryAt(0)}last(){return this.entryAt(-1)}clear(){return Xr(this,Ge,[]),super.clear()}delete(a){let o=super.delete(a);return o&&dt(this,Ge).splice(dt(this,Ge).indexOf(a),1),o}deleteAt(a){let o=this.keyAt(a);return o!==void 0?this.delete(o):!1}at(a){let o=bs(dt(this,Ge),a);if(o!==void 0)return this.get(o)}entryAt(a){let o=bs(dt(this,Ge),a);if(o!==void 0)return[o,this.get(o)]}indexOf(a){return dt(this,Ge).indexOf(a)}keyAt(a){return bs(dt(this,Ge),a)}from(a,o){let l=this.indexOf(a);if(l===-1)return;let n=l+o;return n<0&&(n=0),n>=this.size&&(n=this.size-1),this.at(n)}keyFrom(a,o){let l=this.indexOf(a);if(l===-1)return;let n=l+o;return n<0&&(n=0),n>=this.size&&(n=this.size-1),this.keyAt(n)}find(a,o){let l=0;for(let n of this){if(Reflect.apply(a,o,[n,l,this]))return n;l++}}findIndex(a,o){let l=0;for(let n of this){if(Reflect.apply(a,o,[n,l,this]))return l;l++}return-1}filter(a,o){let l=[],n=0;for(let u of this)Reflect.apply(a,o,[u,n,this])&&l.push(u),n++;return new oa(l)}map(a,o){let l=[],n=0;for(let u of this)l.push([u[0],Reflect.apply(a,o,[u,n,this])]),n++;return new oa(l)}reduce(...a){let[o,l]=a,n=0,u=l!=null?l:this.at(0);for(let r of this)n===0&&a.length===1?u=r:u=Reflect.apply(o,this,[u,r,n,this]),n++;return u}reduceRight(...a){let[o,l]=a,n=l!=null?l:this.at(-1);for(let u=this.size-1;u>=0;u--){let r=this.at(u);u===this.size-1&&a.length===1?n=r:n=Reflect.apply(o,this,[n,r,u,this])}return n}toSorted(a){let o=[...this.entries()].sort(a);return new oa(o)}toReversed(){let a=new oa;for(let o=this.size-1;o>=0;o--){let l=this.keyAt(o),n=this.get(l);a.set(l,n)}return a}toSpliced(...a){let o=[...this.entries()];return o.splice(...a),new oa(o)}slice(a,o){let l=new oa,n=this.size-1;if(a===void 0)return l;a<0&&(a=a+this.size),o!==void 0&&o>0&&(n=o-1);for(let u=a;u<=n;u++){let r=this.keyAt(u),i=this.get(r);l.set(r,i)}return l}every(a,o){let l=0;for(let n of this){if(!Reflect.apply(a,o,[n,l,this]))return!1;l++}return!0}some(a,o){let l=0;for(let n of this){if(Reflect.apply(a,o,[n,l,this]))return!0;l++}return!1}},Ge=new WeakMap,ut(oa,"OrderedDict"),oa);function bs(e,t){if("at"in Array.prototype)return Array.prototype.at.call(e,t);let a=PL(e,t);return a===-1?void 0:e[a]}ut(bs,"at");function PL(e,t){let a=e.length,o=Op(t),l=o>=0?o:a+o;return l<0||l>=a?-1:l}ut(PL,"toSafeIndex");function Op(e){return e!==e||e===0?0:Math.trunc(e)}ut(Op,"toSafeInteger");function OA(e){let t=e+"CollectionProvider",[a,o]=Et(t),[l,n]=a(t,{collectionElement:null,collectionRef:{current:null},collectionRefObject:{current:null},itemMap:new Ep,setItemMap:ut(()=>{},"setItemMap")}),u=ut(C=>{var I=C,{state:m}=I,v=D(I,["state"]);return m?(0,zl.jsx)(i,b(y({},v),{state:m})):(0,zl.jsx)(r,y({},v))},"CollectionProvider");u.displayName=t;let r=ut(m=>{let v=S();return(0,zl.jsx)(i,b(y({},m),{state:v}))},"CollectionInit");r.displayName=t+"Init";let i=ut(m=>{let{scope:v,children:C,state:I}=m,R=Ft.useRef(null),[w,A]=Ft.useState(null),M=ce(R,A),[U,N]=I;return Ft.useEffect(()=>{if(!w)return;let H=UL(()=>{});return H.observe(w,{childList:!0,subtree:!0}),()=>{H.disconnect()}},[w]),(0,zl.jsx)(l,{scope:v,itemMap:U,setItemMap:N,collectionRef:M,collectionRefObject:R,collectionElement:w,children:C})},"CollectionProviderImpl");i.displayName=t+"Impl";let s=e+"CollectionSlot",f=Da(s),d=Ft.forwardRef((m,v)=>{let{scope:C,children:I}=m,R=n(s,C),w=ce(v,R.collectionRef);return(0,zl.jsx)(f,{ref:w,children:I})});d.displayName=s;let c=e+"CollectionItemSlot",g="data-radix-collection-item",L=Da(c),x=Ft.forwardRef((m,v)=>{let k=m,{scope:C,children:I}=k,R=D(k,["scope","children"]),w=Ft.useRef(null),[A,M]=Ft.useState(null),U=ce(v,w,M),N=n(c,C),{setItemMap:H}=N,O=Ft.useRef(R);BL(O.current,R)||(O.current=R);let G=O.current;return Ft.useEffect(()=>{let j=G;return H(P=>A?P.has(A)?P.set(A,b(y({},j),{element:A})).toSorted(Dp):(P.set(A,b(y({},j),{element:A})),P.toSorted(Dp)):P),()=>{H(P=>!A||!P.has(A)?P:(P.delete(A),new Ep(P)))}},[A,G,H]),(0,zl.jsx)(L,{[g]:"",ref:U,children:I})});x.displayName=c;function S(){return Ft.useState(new Ep)}ut(S,"useInitCollection");function h(m){let{itemMap:v}=n(e+"CollectionConsumer",m);return v}return ut(h,"useCollection"),[{Provider:u,Slot:d,ItemSlot:x},{createCollectionScope:o,useCollection:h,useInitCollection:S}]}ut(OA,"createCollection");function BL(e,t){if(e===t)return!0;if(typeof e!="object"||typeof t!="object"||e==null||t==null)return!1;let a=Object.keys(e),o=Object.keys(t);if(a.length!==o.length)return!1;for(let l of a)if(!Object.prototype.hasOwnProperty.call(t,l)||e[l]!==t[l])return!1;return!0}ut(BL,"shallowEqual");function _L(e,t){return!!(t.compareDocumentPosition(e)&Node.DOCUMENT_POSITION_PRECEDING)}ut(_L,"isElementPreceding");function Dp(e,t){return!e[1].element||!t[1].element?0:_L(e[1].element,t[1].element)?-1:1}ut(Dp,"sortByDocumentPosition");function UL(e){return new MutationObserver(a=>{for(let o of a)if(o.type==="childList"){e();return}})}ut(UL,"getChildListObserver");var Pp=T(B(),1);var NL=T(B(),1),We=globalThis!=null&&globalThis.document?NL.useLayoutEffect:()=>{};var PA=Object.defineProperty,BA=(e,t)=>PA(e,"name",{value:t,configurable:!0}),_A=Pp[" useId ".trim().toString()]||(()=>{}),UA=0;function Fl(e){let[t,a]=Pp.useState(_A());return We(()=>{e||a(o=>o!=null?o:String(UA++))},[e]),e||(t?"radix-".concat(t):"")}BA(Fl,"useId");var Wn=T(B(),1),NA=Object.defineProperty,kA=(e,t)=>NA(e,"name",{value:t,configurable:!0});function ya(e){let t=Wn.useRef(e);return Wn.useEffect(()=>{t.current=e}),Wn.useMemo(()=>((...a)=>{var o;return(o=t.current)==null?void 0:o.call(t,...a)}),[])}kA(ya,"useCallbackRef");var Ca=T(B(),1);var ws=!1;var Oa=T(B(),1);var Jn=T(B(),1),HA=Object.defineProperty,zA=(e,t)=>HA(e,"name",{value:t,configurable:!0}),kL=Jn[" useEffectEvent ".trim().toString()],HL=Jn[" useInsertionEffect ".trim().toString()];function Bp(e){if(typeof kL=="function")return kL(e);let t=Jn.useRef(()=>{throw new Error("Cannot call an event handler while rendering.")});return typeof HL=="function"?HL(()=>{t.current=e}):We(()=>{t.current=e}),Jn.useMemo(()=>((...a)=>{var o;return(o=t.current)==null?void 0:o.call(t,...a)}),[])}zA(Bp,"useEffectEvent");var FA=Object.defineProperty,vr=(e,t)=>FA(e,"name",{value:t,configurable:!0}),qA=Ca[" useInsertionEffect ".trim().toString()]||We;function ql({prop:e,defaultProp:t,onChange:a=vr(()=>{},"onChange"),caller:o}){let[l,n,u]=FL({defaultProp:t,onChange:a}),r=e!==void 0,i=r?e:l;if(ws){let f=Ca.useRef(e!==void 0);Ca.useEffect(()=>{let d=f.current;if(d!==r){let c=d?"controlled":"uncontrolled",g=r?"controlled":"uncontrolled";console.warn("".concat(o," is changing from ").concat(c," to ").concat(g,". Components should not switch from controlled to uncontrolled (or vice versa). Decide between using a controlled or uncontrolled value for the lifetime of the component."))}f.current=r},[r,o])}let s=Ca.useCallback(f=>{var d;if(r){let c=qL(f)?f(e):f;c!==e&&((d=u.current)==null||d.call(u,c))}else n(f)},[r,e,n,u]);return[i,s]}vr(ql,"useControllableState");function FL({defaultProp:e,onChange:t}){let[a,o]=Ca.useState(e),l=Ca.useRef(a),n=Ca.useRef(t);return qA(()=>{n.current=t},[t]),Ca.useEffect(()=>{var u;l.current!==a&&((u=n.current)==null||u.call(n,a),l.current=a)},[a,l]),[a,o,n]}vr(FL,"useUncontrolledState");function qL(e){return typeof e=="function"}vr(qL,"isFunction");var zL=Symbol("RADIX:SYNC_STATE");function GA(e,t,a,o){let{prop:l,defaultProp:n,onChange:u,caller:r}=t,i=l!==void 0,s=Bp(u);if(ws){let S=Oa.useRef(l!==void 0);Oa.useEffect(()=>{let h=S.current;if(h!==i){let p=h?"controlled":"uncontrolled",m=i?"controlled":"uncontrolled";console.warn("".concat(r," is changing from ").concat(p," to ").concat(m,". Components should not switch from controlled to uncontrolled (or vice versa). Decide between using a controlled or uncontrolled value for the lifetime of the component."))}S.current=i},[i,r])}let f=[b(y({},a),{state:n})];o&&f.push(o);let[d,c]=Oa.useReducer((S,h)=>{if(h.type===zL)return b(y({},S),{state:h.state});let p=e(S,h);return i&&!Object.is(p.state,S.state)&&s(p.state),p},...f),g=d.state,L=Oa.useRef(g);Oa.useEffect(()=>{L.current!==g&&(L.current=g,i||s(g))},[g,L,i]);let x=Oa.useMemo(()=>l!==void 0?b(y({},d),{state:l}):d,[d,l]);return Oa.useEffect(()=>{i&&!Object.is(l,d.state)&&c({type:zL,state:l})},[l,d.state,i]),[x,c]}vr(GA,"useControllableStateReducer");var As=T(B(),1),jA=T(X(),1),VA=Object.defineProperty,XA=(e,t)=>VA(e,"name",{value:t,configurable:!0}),YA={LTR:"ltr",RTL:"rtl"},KA=As.createContext(void 0);function Sr(e){let t=As.useContext(KA);return e||t||YA.LTR}XA(Sr,"useDirection");var QA=T(B(),1),Ts=T(B(),1),ZA=Object.defineProperty,Up=(e,t)=>ZA(e,"name",{value:t,configurable:!0}),_p=!1;function GL(){let[e,t]=Ts.useState(_p);return Ts.useEffect(()=>{_p||(_p=!0,t(!0))},[]),e}Up(GL,"useIsHydrated");var VL=QA[" useSyncExternalStore ".trim().toString()];function XL(){return()=>{}}Up(XL,"subscribe");function jL(){return VL(XL,()=>!0,()=>!1)}Up(jL,"useIsHydratedModern");var YL=typeof VL=="function"?jL:GL;var fl=T(X(),1);var WA=Object.defineProperty,Gl=(e,t)=>WA(e,"name",{value:t,configurable:!0}),Np="rovingFocusGroup.onEntryFocus",JA={bubbles:!1,cancelable:!0},Es="RovingFocusGroup",[kp,ZL,$A]=Mp(Es),[eT,Hp]=Et(Es,[$A]),KL={Vertical:"vertical",Horizontal:"horizontal"},tT={LTR:"ltr",RTL:"rtl"},[aT,oT]=eT(Es),QL=rt.forwardRef(Gl(function(t,a){return(0,fl.jsx)(kp.Provider,{scope:t.__scopeRovingFocusGroup,children:(0,fl.jsx)(kp.Slot,{scope:t.__scopeRovingFocusGroup,children:(0,fl.jsx)(lT,b(y({},t),{ref:a}))})})},"RovingFocusGroup")),lT=rt.forwardRef(Gl(function(t,a){let A=t,{__scopeRovingFocusGroup:o,orientation:l,loop:n=!1,dir:u,currentTabStopId:r,defaultCurrentTabStopId:i,onCurrentTabStopIdChange:s,onEntryFocus:f,preventScrollOnEntryFocus:d=!1}=A,c=D(A,["__scopeRovingFocusGroup","orientation","loop","dir","currentTabStopId","defaultCurrentTabStopId","onCurrentTabStopIdChange","onEntryFocus","preventScrollOnEntryFocus"]),g=rt.useRef(null),L=ce(a,g),x=Sr(u),[S,h]=ql({prop:r,defaultProp:i!=null?i:null,onChange:s,caller:Es}),[p,m]=rt.useState(!1),v=ya(f),C=ZL(o),I=rt.useRef(!1),[R,w]=rt.useState(0);return rt.useEffect(()=>{let M=g.current;if(M)return M.addEventListener(Np,v),()=>M.removeEventListener(Np,v)},[v]),(0,fl.jsx)(aT,{scope:o,orientation:l,dir:x,loop:n,currentTabStopId:S,onItemFocus:rt.useCallback(M=>h(M),[h]),onItemShiftTab:rt.useCallback(()=>m(!0),[]),onFocusableItemAdd:rt.useCallback(()=>w(M=>M+1),[]),onFocusableItemRemove:rt.useCallback(()=>w(M=>M-1),[]),children:(0,fl.jsx)(Qe.div,b(y({tabIndex:p||R===0?-1:0,"data-orientation":l},c),{ref:L,style:y({outline:"none"},t.style),onMouseDown:Ee(t.onMouseDown,()=>{I.current=!0}),onFocus:Ee(t.onFocus,M=>{let U=!I.current;if(M.target===M.currentTarget&&U&&!p){let N=new CustomEvent(Np,JA);if(M.currentTarget.dispatchEvent(N),!N.defaultPrevented){let H=C().filter(P=>P.focusable),O=H.find(P=>P.active),G=H.find(P=>P.id===S),j=[O,G,...H].filter(Boolean).map(P=>P.ref.current);zp(j,d)}}I.current=!1}),onBlur:Ee(t.onBlur,()=>m(!1))}))})},"RovingFocusGroupImpl")),nT="RovingFocusGroupItem",WL=rt.forwardRef(Gl(function(t,a){let p=t,{__scopeRovingFocusGroup:o,focusable:l=!0,active:n=!1,tabStopId:u,children:r}=p,i=D(p,["__scopeRovingFocusGroup","focusable","active","tabStopId","children"]),s=Fl(),f=u||s,d=oT(nT,o),c=d.currentTabStopId===f,g=ZL(o),{onFocusableItemAdd:L,onFocusableItemRemove:x,currentTabStopId:S}=d,h=YL();return We(()=>{if(!(!h||!l))return L(),()=>x()},[h,l,L,x]),rt.useEffect(()=>{if(!(h||!l))return L(),()=>x()},[h,l,L,x]),(0,fl.jsx)(kp.ItemSlot,{scope:o,id:f,focusable:l,active:n,children:(0,fl.jsx)(Qe.span,b(y({tabIndex:c?0:-1,"data-orientation":d.orientation},i),{ref:a,onMouseDown:Ee(t.onMouseDown,m=>{l?d.onItemFocus(f):m.preventDefault()}),onFocus:Ee(t.onFocus,()=>d.onItemFocus(f)),onKeyDown:Ee(t.onKeyDown,m=>{if(m.key==="Tab"&&m.shiftKey){d.onItemShiftTab();return}if(m.target!==m.currentTarget)return;let v=$L(m,d.orientation,d.dir);if(v!==void 0){if(m.metaKey||m.ctrlKey||m.altKey||m.shiftKey)return;m.preventDefault();let I=g().filter(R=>R.focusable).map(R=>R.ref.current);if(v==="last")I.reverse();else if(v==="prev"||v==="next"){v==="prev"&&I.reverse();let R=I.indexOf(m.currentTarget);I=d.loop?ey(I,R+1):I.slice(R+1)}setTimeout(()=>zp(I))}}),children:typeof r=="function"?r({isCurrentTabStop:c,hasTabStop:S!=null}):r}))})},"RovingFocusGroupItem")),uT={ArrowLeft:"prev",ArrowUp:"prev",ArrowRight:"next",ArrowDown:"next",PageUp:"first",Home:"first",PageDown:"last",End:"last"};function JL(e,t){return t!==tT.RTL?e:e==="ArrowLeft"?"ArrowRight":e==="ArrowRight"?"ArrowLeft":e}Gl(JL,"getDirectionAwareKey");function $L(e,t,a){let o=JL(e.key,a);if(!(t===KL.Vertical&&["ArrowLeft","ArrowRight"].includes(o))&&!(t===KL.Horizontal&&["ArrowUp","ArrowDown"].includes(o)))return uT[o]}Gl($L,"getFocusIntent");function zp(e,t=!1){let a=document.activeElement;for(let o of e)if(o===a||(o.focus({preventScroll:t}),document.activeElement!==a))return}Gl(zp,"focusFirst");function ey(e,t){return e.map((a,o)=>e[(t+o)%e.length])}Gl(ey,"wrapArray");var yt=T(B(),1);var ty=T(B(),1);var iT=Object.defineProperty,sT=(e,t)=>iT(e,"name",{value:t,configurable:!0});function xr(e){let[t,a]=ty.useState(void 0);return We(()=>{if(e){a({width:e.offsetWidth,height:e.offsetHeight});let o=0,l=new ResizeObserver(n=>{if(!Array.isArray(n)||!n.length)return;let u=n[0];window.cancelAnimationFrame(o),o=window.requestAnimationFrame(()=>{let r,i;if("borderBoxSize"in u){let s=u.borderBoxSize,f=Array.isArray(s)?s[0]:s;r=f.inlineSize,i=f.blockSize}else r=e.offsetWidth,i=e.offsetHeight;a({width:r,height:i})})});return l.observe(e,{box:"border-box"}),()=>{window.cancelAnimationFrame(o),l.unobserve(e)}}else a(void 0)},[e]),t}sT(xr,"useSize");var xt=T(B(),1);var ay=T(B(),1),fT=Object.defineProperty,bo=(e,t)=>fT(e,"name",{value:t,configurable:!0});function oy(e,t){return ay.useReducer((a,o)=>{let l=t[a][o];return l!=null?l:a},e)}bo(oy,"useStateMachine");var qp=bo(e=>{let{present:t,children:a}=e,o=ly(t),l=typeof a=="function"?a({present:o.isPresent}):xt.Children.only(a),n=ny(o.ref,uy(l));return typeof a=="function"||o.isPresent?xt.cloneElement(l,{ref:n}):null},"Presence");function ly(e){let[t,a]=xt.useState(),o=xt.useRef(null),l=xt.useRef(e),n=xt.useRef("none"),u=xt.useRef(void 0),r=e?"mounted":"unmounted",[i,s]=oy(r,{mounted:{UNMOUNT:"unmounted",ANIMATION_OUT:"unmountSuspended"},unmountSuspended:{MOUNT:"mounted",ANIMATION_END:"unmounted"},unmounted:{MOUNT:"mounted"}});return xt.useEffect(()=>{var f;i==="mounted"?(n.current=(f=u.current)!=null?f:$n(o.current),u.current=void 0):n.current="none"},[i]),We(()=>{let f=o.current,d=l.current;if(d!==e){let g=n.current,L=$n(f);e?(u.current=L,s("MOUNT")):L==="none"||(f==null?void 0:f.display)==="none"?s("UNMOUNT"):s(d&&g!==L?"ANIMATION_OUT":"UNMOUNT"),l.current=e}},[e,s]),We(()=>{var f;if(t){let d,c=(f=t.ownerDocument.defaultView)!=null?f:window,g=bo(x=>{let h=$n(o.current).includes(CSS.escape(x.animationName));if(x.target===t&&h&&(s("ANIMATION_END"),!l.current)){let p=t.style.animationFillMode;t.style.animationFillMode="forwards",d=c.setTimeout(()=>{t.style.animationFillMode==="forwards"&&(t.style.animationFillMode=p)})}},"handleAnimationEnd"),L=bo(x=>{x.target===t&&(n.current=$n(o.current))},"handleAnimationStart");return t.addEventListener("animationstart",L),t.addEventListener("animationcancel",g),t.addEventListener("animationend",g),()=>{c.clearTimeout(d),t.removeEventListener("animationstart",L),t.removeEventListener("animationcancel",g),t.removeEventListener("animationend",g)}}else s("ANIMATION_END")},[t,s]),{isPresent:["mounted","unmountSuspended"].includes(i),ref:xt.useCallback(f=>{if(f){let d=getComputedStyle(f);o.current=d,u.current=$n(d)}else o.current=null;a(f)},[])}}bo(ly,"usePresence");function Fp(e,t){if(typeof e=="function")return e(t);e!=null&&(e.current=t)}bo(Fp,"setRef");function ny(...e){let t=xt.useRef(e);return t.current=e,xt.useCallback(a=>{let o=t.current,l=!1,n=o.map(u=>{let r=Fp(u,a);return!l&&typeof r=="function"&&(l=!0),r});if(l)return()=>{for(let u=0;u<n.length;u++){let r=n[u];typeof r=="function"?r():Fp(o[u],null)}}},[])}bo(ny,"useStableComposedRefs");function $n(e){return(e==null?void 0:e.animationName)||"none"}bo($n,"getAnimationName");function uy(e){var o,l;let t=(o=Object.getOwnPropertyDescriptor(e.props,"ref"))==null?void 0:o.get,a=t&&"isReactWarning"in t&&t.isReactWarning;return a?e.ref:(t=(l=Object.getOwnPropertyDescriptor(e,"ref"))==null?void 0:l.get,a=t&&"isReactWarning"in t&&t.isReactWarning,a?e.props.ref:e.props.ref||e.ref)}bo(uy,"getElementRef");var Ds=T(X(),1),Dt=T(X(),1);var cT=Object.defineProperty,qt=(e,t)=>cT(e,"name",{value:t,configurable:!0}),ry="Radio",[dT,iy]=Et(ry),[pT,Gp]=dT(ry);function sy(e){let{__scopeRadio:t,checked:a=!1,children:o,disabled:l,form:n,name:u,onCheck:r,required:i,value:s="on",internal_do_not_use_render:f}=e,[d,c]=yt.useState(null),[g,L]=yt.useState(null),x=yt.useRef(!1),[S,h]=yt.useReducer(v=>v+1,0),p=d?!!n||!!d.closest("form"):!0,m={checked:a,disabled:l,required:i,name:u,form:n,value:s,control:d,setControl:c,hasConsumerStoppedPropagationRef:x,userInteractionCount:S,onUserInteraction:h,isFormControl:p,bubbleInput:g,setBubbleInput:L,onCheck:qt(()=>r==null?void 0:r(),"onCheck")};return(0,Ds.jsx)(pT,b(y({scope:t},m),{children:fy(f)?f(m):o}))}qt(sy,"RadioProvider");var mT="RadioTrigger",hT=yt.forwardRef(qt(function(n,l){var u=n,{__scopeRadio:t,onClick:a}=u,o=D(u,["__scopeRadio","onClick"]);let{checked:r,disabled:i,value:s,setControl:f,onCheck:d,hasConsumerStoppedPropagationRef:c,onUserInteraction:g,isFormControl:L,bubbleInput:x}=Gp(mT,t),S=ce(l,f);return(0,Ds.jsx)(Qe.button,b(y({type:"button",role:"radio","aria-checked":r,"data-state":cy(r),"data-disabled":i?"":void 0,disabled:i,value:s},o),{ref:S,onClick:Ee(a,h=>{r||(g(),d()),x&&L&&(c.current=h.isPropagationStopped(),c.current||h.stopPropagation())})}))},"RadioTrigger"));var gT="RadioBubbleInput",vT=yt.forwardRef(qt(function(n,l){var u=n,{__scopeRadio:t,onClick:a}=u,o=D(u,["__scopeRadio","onClick"]);let{control:r,checked:i,required:s,disabled:f,name:d,value:c,form:g,bubbleInput:L,setBubbleInput:x,hasConsumerStoppedPropagationRef:S,userInteractionCount:h}=Gp(gT,t),p=ce(l,x),m=xr(r),v=yt.useRef(!1),C=yt.useRef(i),I=yt.useRef(h);yt.useEffect(()=>{let w=L;if(!w)return;let A=window.HTMLInputElement.prototype,U=Object.getOwnPropertyDescriptor(A,"checked").set,N=h!==I.current;I.current=h;let H=C.current!==i;C.current=i;let O=!(N&&S.current);if(H&&U){v.current=!N;let G=new Event("click",{bubbles:O});U.call(w,i),w.dispatchEvent(G),v.current=!1}},[L,i,S,h]);let R=yt.useRef(i);return(0,Ds.jsx)(Qe.input,b(y({type:"radio","aria-hidden":!0,defaultChecked:R.current,required:s,disabled:f,name:d,value:c,form:g},o),{tabIndex:-1,ref:p,onClick:Ee(a,w=>{v.current&&w.stopPropagation()}),style:b(y(y({},o.style),m),{position:"absolute",pointerEvents:"none",opacity:0,margin:0,transform:"translateX(-100%)"})}))},"RadioBubbleInput"));function fy(e){return typeof e=="function"}qt(fy,"isFunction");function cy(e){return e?"checked":"unchecked"}qt(cy,"getState");var ST=["ArrowUp","ArrowDown","ArrowLeft","ArrowRight"],Vp="RadioGroup",[xT,jM]=Et(Vp,[Hp,iy]),dy=Hp(),Xp=iy(),[LT,yT]=xT(Vp),py=Gt.forwardRef(qt(function(t,a){let R=t,{__scopeRadioGroup:o,name:l,form:n,defaultValue:u,value:r,required:i=!1,disabled:s=!1,orientation:f,dir:d,loop:c=!0,onValueChange:g}=R,L=D(R,["__scopeRadioGroup","name","form","defaultValue","value","required","disabled","orientation","dir","loop","onValueChange"]),x=dy(o),S=Sr(d),[h,p]=ql({prop:r,defaultProp:u!=null?u:null,onChange:g,caller:Vp}),[m,v]=Gt.useState(null),C=ce(a,v),I=Gt.useRef(h);return Gt.useEffect(()=>{let w=n?m==null?void 0:m.ownerDocument.getElementById(n):m==null?void 0:m.closest("form");if(w instanceof HTMLFormElement){let A=qt(()=>p(I.current),"reset");return w.addEventListener("reset",A),()=>w.removeEventListener("reset",A)}},[m,n,p]),(0,Dt.jsx)(LT,{scope:o,name:l,form:n,required:i,disabled:s,value:h,onValueChange:p,children:(0,Dt.jsx)(QL,b(y({asChild:!0},x),{orientation:f,dir:S,loop:c,children:(0,Dt.jsx)(Qe.div,b(y({role:"radiogroup","aria-required":i,"aria-orientation":f,"data-disabled":s?"":void 0,dir:S},L),{ref:C}))}))})},"RadioGroup")),CT="RadioGroupItemProvider",RT="RadioGroupItemTrigger";function my(e){let{__scopeRadioGroup:t,value:a,disabled:o,children:l,internal_do_not_use_render:n}=e,u=yT(CT,t),r=Xp(t),i=u.disabled||o;return(0,Dt.jsx)(sy,b(y({},r),{checked:u.value===a,disabled:i,required:u.required,name:u.name,form:u.form,value:a,onCheck:()=>u.onValueChange(a),internal_do_not_use_render:n,children:l}))}qt(my,"RadioGroupItemProvider");var IT=Gt.forwardRef(qt(function(t,a){let c=t,{__scopeRadioGroup:o}=c,l=D(c,["__scopeRadioGroup"]),n=dy(o),u=Xp(o),{checked:r,disabled:i}=Gp(RT,u.__scopeRadio),s=Gt.useRef(null),f=ce(a,s),d=Gt.useRef(!1);return Gt.useEffect(()=>{let g=qt(x=>{ST.includes(x.key)&&(d.current=!0)},"handleKeyDown"),L=qt(()=>d.current=!1,"handleKeyUp");return document.addEventListener("keydown",g),document.addEventListener("keyup",L),()=>{document.removeEventListener("keydown",g),document.removeEventListener("keyup",L)}},[]),(0,Dt.jsx)(WL,b(y({asChild:!0},n),{focusable:!i,active:r,children:(0,Dt.jsx)(hT,b(y(y({},u),l),{ref:f,onKeyDown:Ee(l.onKeyDown,g=>{g.key==="Enter"&&g.preventDefault()}),onFocus:Ee(l.onFocus,()=>{var g;d.current&&((g=s.current)==null||g.click())})}))}))},"RadioGroupItemTrigger")),hy=Gt.forwardRef(qt(function(t,a){let r=t,{__scopeRadioGroup:o,value:l,disabled:n}=r,u=D(r,["__scopeRadioGroup","value","disabled"]);return(0,Dt.jsx)(my,{__scopeRadioGroup:o,value:l,disabled:n,internal_do_not_use_render:({isFormControl:i})=>(0,Dt.jsxs)(Dt.Fragment,{children:[(0,Dt.jsx)(IT,b(y({},u),{ref:a,__scopeRadioGroup:o})),i&&(0,Dt.jsx)(bT,{__scopeRadioGroup:o})]})})},"RadioGroupItem")),bT=Gt.forwardRef(qt(function(t,a){let u=t,{__scopeRadioGroup:o}=u,l=D(u,["__scopeRadioGroup"]),n=Xp(o);return(0,Dt.jsx)(vT,b(y(y({},n),l),{ref:a}))},"RadioGroupItemBubbleInput"));function Be(...e){return e.filter(Boolean).join(" ")}var Ms=T(B());function gy(e,t,a){let[o,l]=Ms.useState();return Ms.useLayoutEffect(()=>{var h;let n=e.current;if(!n){l(void 0);return}let u=n,r=u.ownerDocument.defaultView;if(!r)return;let i=!1,s=0,f=null,d=new Set,c=()=>{if(i)return;f=u.querySelector(t),x();let p=u.getBoundingClientRect(),m=f==null?void 0:f.getBoundingClientRect();if(!u.isConnected||!p.width||!p.height||!(m!=null&&m.width)||!m.height||!(f!=null&&f.getClientRects().length)){l(C=>C===void 0?C:void 0);return}let v="translate(".concat(m.left-p.left+u.scrollLeft-u.clientLeft,"px, ").concat(m.top-p.top+u.scrollTop-u.clientTop,"px)");l(C=>(C==null?void 0:C.width)===m.width&&C.height===m.height&&C.transform===v?C:{position:"absolute",top:0,left:0,width:m.width,height:m.height,transform:v})},g=()=>{i||s||(s=r.requestAnimationFrame(()=>{s=0,c()}))},L=typeof ResizeObserver=="undefined"?void 0:new ResizeObserver(g);function x(){let p=new Set([u,...Array.from(u.children).filter(m=>!m.classList.contains("ui-tabs-indicator")&&!m.classList.contains("ui-segmented-indicator")&&!m.classList.contains("navigation-indicator"))]);f&&p.add(f);for(let m of d)p.has(m)||(L==null||L.unobserve(m),d.delete(m));for(let m of p)d.has(m)||(L==null||L.observe(m),d.add(m))}let S=new MutationObserver(g);return S.observe(u,{subtree:!0,childList:!0,attributes:!0,attributeFilter:["class","data-state","aria-current"]}),r.addEventListener("resize",g),u.addEventListener("scroll",g,{passive:!0}),(h=u.ownerDocument.fonts)==null||h.ready.then(g),c(),()=>{i=!0,r.cancelAnimationFrame(s),L==null||L.disconnect(),S.disconnect(),r.removeEventListener("resize",g),u.removeEventListener("scroll",g)}},[e,t,a]),{style:o,ready:o!==void 0}}var Lr=T(X()),jp=Pa.forwardRef((u,n)=>{var r=u,{className:e,children:t,size:a="default",fullWidth:o=!1}=r,l=D(r,["className","children","size","fullWidth"]);let i=Pa.useRef(null);Pa.useImperativeHandle(n,()=>i.current,[]);let{style:s,ready:f}=gy(i,'.ui-segmented-item[data-state="checked"]',l.value),[d,c]=Pa.useState(!1);return Pa.useLayoutEffect(()=>{var S;if(!f){c(!1);return}let g=(S=i.current)==null?void 0:S.ownerDocument.defaultView;if(!g)return;let L=0,x=g.requestAnimationFrame(()=>{L=g.requestAnimationFrame(()=>c(!0))});return()=>{g.cancelAnimationFrame(x),g.cancelAnimationFrame(L)}},[f]),(0,Lr.jsxs)(py,b(y({},l),{ref:i,className:Be("ui-segmented","ui-segmented--".concat(a),o&&"ui-segmented--full",e),"data-indicator-ready":f?"true":void 0,"data-indicator-animate":f&&d?"true":void 0,children:[(0,Lr.jsx)("span",{className:"ui-segmented-indicator","aria-hidden":"true",style:s}),t]}))});jp.displayName="SegmentedControl";var Yp=Pa.forwardRef((o,a)=>{var l=o,{className:e}=l,t=D(l,["className"]);return(0,Lr.jsx)(hy,y({ref:a,className:Be("ui-segmented-item",e)},t))});Yp.displayName="SegmentedControlItem";var su=T(B());var Ps=T(B());/**
 * @license lucide-react v0.468.0 - ISC
 *
 * This source code is licensed under the ISC license.
 * See the LICENSE file in the root directory of this source tree.
 */var vy=e=>e.replace(/([a-z0-9])([A-Z])/g,"$1-$2").toLowerCase(),Os=(...e)=>e.filter((t,a,o)=>!!t&&t.trim()!==""&&o.indexOf(t)===a).join(" ").trim();var yr=T(B());/**
 * @license lucide-react v0.468.0 - ISC
 *
 * This source code is licensed under the ISC license.
 * See the LICENSE file in the root directory of this source tree.
 */var Sy={xmlns:"http://www.w3.org/2000/svg",width:24,height:24,viewBox:"0 0 24 24",fill:"none",stroke:"currentColor",strokeWidth:2,strokeLinecap:"round",strokeLinejoin:"round"};/**
 * @license lucide-react v0.468.0 - ISC
 *
 * This source code is licensed under the ISC license.
 * See the LICENSE file in the root directory of this source tree.
 */var xy=(0,yr.forwardRef)((s,i)=>{var f=s,{color:e="currentColor",size:t=24,strokeWidth:a=2,absoluteStrokeWidth:o,className:l="",children:n,iconNode:u}=f,r=D(f,["color","size","strokeWidth","absoluteStrokeWidth","className","children","iconNode"]);return(0,yr.createElement)("svg",y(b(y({ref:i},Sy),{width:t,height:t,stroke:e,strokeWidth:o?Number(a)*24/Number(t):a,className:Os("lucide",l)}),r),[...u.map(([d,c])=>(0,yr.createElement)(d,c)),...Array.isArray(n)?n:[n]])});/**
 * @license lucide-react v0.468.0 - ISC
 *
 * This source code is licensed under the ISC license.
 * See the LICENSE file in the root directory of this source tree.
 */var Ra=(e,t)=>{let a=(0,Ps.forwardRef)((u,n)=>{var r=u,{className:o}=r,l=D(r,["className"]);return(0,Ps.createElement)(xy,y({ref:n,iconNode:t,className:Os("lucide-".concat(vy(e)),o)},l))});return a.displayName="".concat(e),a};/**
 * @license lucide-react v0.468.0 - ISC
 *
 * This source code is licensed under the ISC license.
 * See the LICENSE file in the root directory of this source tree.
 */var Cr=Ra("CalendarDays",[["path",{d:"M8 2v4",key:"1cmpym"}],["path",{d:"M16 2v4",key:"4m81vk"}],["rect",{width:"18",height:"18",x:"3",y:"4",rx:"2",key:"1hopcy"}],["path",{d:"M3 10h18",key:"8toen8"}],["path",{d:"M8 14h.01",key:"6423bh"}],["path",{d:"M12 14h.01",key:"1etili"}],["path",{d:"M16 14h.01",key:"1gbofw"}],["path",{d:"M8 18h.01",key:"lrp35t"}],["path",{d:"M12 18h.01",key:"mhygvu"}],["path",{d:"M16 18h.01",key:"kzsmim"}]]);/**
 * @license lucide-react v0.468.0 - ISC
 *
 * This source code is licensed under the ISC license.
 * See the LICENSE file in the root directory of this source tree.
 */var Rr=Ra("Check",[["path",{d:"M20 6 9 17l-5-5",key:"1gmf2c"}]]);/**
 * @license lucide-react v0.468.0 - ISC
 *
 * This source code is licensed under the ISC license.
 * See the LICENSE file in the root directory of this source tree.
 */var eu=Ra("ChevronDown",[["path",{d:"m6 9 6 6 6-6",key:"qrunsl"}]]);/**
 * @license lucide-react v0.468.0 - ISC
 *
 * This source code is licensed under the ISC license.
 * See the LICENSE file in the root directory of this source tree.
 */var Ir=Ra("ChevronLeft",[["path",{d:"m15 18-6-6 6-6",key:"1wnfg3"}]]);/**
 * @license lucide-react v0.468.0 - ISC
 *
 * This source code is licensed under the ISC license.
 * See the LICENSE file in the root directory of this source tree.
 */var br=Ra("ChevronRight",[["path",{d:"m9 18 6-6-6-6",key:"mthhwq"}]]);/**
 * @license lucide-react v0.468.0 - ISC
 *
 * This source code is licensed under the ISC license.
 * See the LICENSE file in the root directory of this source tree.
 */var wr=Ra("ChevronUp",[["path",{d:"m18 15-6-6-6 6",key:"153udz"}]]);/**
 * @license lucide-react v0.468.0 - ISC
 *
 * This source code is licensed under the ISC license.
 * See the LICENSE file in the root directory of this source tree.
 */var dR=T(B());var we=T(B(),1);var me=T(B(),1);var yy=T(X(),1);var AT=Object.defineProperty,it=(e,t)=>AT(e,"name",{value:t,configurable:!0}),Kp="dismissableLayer.update",TT="dismissableLayer.pointerDownOutside",ET="dismissableLayer.focusOutside",Ly,Cy=me.createContext({layers:new Set,layersWithOutsidePointerEventsDisabled:new Set,branches:new Set,dismissableSurfaces:new Set}),Ry=me.forwardRef(it(function(t,a){var H;let N=t,{disableOutsidePointerEvents:o=!1,deferPointerDownOutside:l=!1,onEscapeKeyDown:n,onPointerDownOutside:u,onFocusOutside:r,onInteractOutside:i,onDismiss:s}=N,f=D(N,["disableOutsidePointerEvents","deferPointerDownOutside","onEscapeKeyDown","onPointerDownOutside","onFocusOutside","onInteractOutside","onDismiss"]),d=me.useContext(Cy),[c,g]=me.useState(null),L=(H=c==null?void 0:c.ownerDocument)!=null?H:globalThis==null?void 0:globalThis.document,[,x]=me.useState({}),S=ce(a,g),h=Array.from(d.layers),[p]=[...d.layersWithOutsidePointerEventsDisabled].slice(-1),m=p?h.indexOf(p):-1,v=c?h.indexOf(c):-1,C=d.layersWithOutsidePointerEventsDisabled.size>0,I=v>=m,R=me.useRef(!1),w=Iy(O=>{u==null||u(O),i==null||i(O),O.defaultPrevented||s==null||s()},{ownerDocument:L,deferPointerDownOutside:l,isDeferredPointerDownOutsideRef:R,dismissableSurfaces:d.dismissableSurfaces,shouldHandlePointerDownOutside:me.useCallback(O=>{if(!(O instanceof Node))return!1;let G=[...d.branches].some(k=>k.contains(O));return I&&!G},[d.branches,I])}),A=by(O=>{if(l&&R.current)return;let G=O.target;[...d.branches].some(j=>j.contains(G))||(r==null||r(O),i==null||i(O),O.defaultPrevented||s==null||s())},L),M=c?v===h.length-1:!1,U=ya(O=>{O.key==="Escape"&&(n==null||n(O),!O.defaultPrevented&&s&&(O.preventDefault(),s()))});return me.useEffect(()=>{if(M)return L.addEventListener("keydown",U,{capture:!0}),()=>L.removeEventListener("keydown",U,{capture:!0})},[L,M,U]),me.useEffect(()=>{if(c)return o&&(d.layersWithOutsidePointerEventsDisabled.size===0&&(Ly=L.body.style.pointerEvents,L.body.style.pointerEvents="none"),d.layersWithOutsidePointerEventsDisabled.add(c)),d.layers.add(c),Zp(),()=>{o&&(d.layersWithOutsidePointerEventsDisabled.delete(c),d.layersWithOutsidePointerEventsDisabled.size===0&&(L.body.style.pointerEvents=Ly))}},[c,L,o,d]),me.useEffect(()=>()=>{c&&(d.layers.delete(c),d.layersWithOutsidePointerEventsDisabled.delete(c),Zp())},[c,d]),me.useEffect(()=>{let O=it(()=>x({}),"handleUpdate");return document.addEventListener(Kp,O),()=>document.removeEventListener(Kp,O)},[]),(0,yy.jsx)(Qe.div,b(y({},f),{ref:S,style:y({pointerEvents:C?I?"auto":"none":void 0},t.style),onFocusCapture:Ee(t.onFocusCapture,A.onFocusCapture),onBlurCapture:Ee(t.onBlurCapture,A.onBlurCapture),onPointerDownCapture:Ee(t.onPointerDownCapture,w.onPointerDownCapture)}))},"DismissableLayer"));function DT(){let e=me.useContext(Cy),[t,a]=me.useState(null);return me.useEffect(()=>{if(t)return e.dismissableSurfaces.add(t),()=>{e.dismissableSurfaces.delete(t)}},[t,e.dismissableSurfaces]),a}it(DT,"useDismissableLayerSurface");var MT=it(()=>!0,"IS_TRUE");function Iy(e,t){let{ownerDocument:a=globalThis==null?void 0:globalThis.document,deferPointerDownOutside:o=!1,isDeferredPointerDownOutsideRef:l,dismissableSurfaces:n,shouldHandlePointerDownOutside:u=MT}=t,r=ya(e),i=me.useRef(!1),s=me.useRef(!1),f=me.useRef(new Map),d=me.useRef(()=>{});return me.useEffect(()=>{function c(){s.current=!1,l.current=!1,f.current.clear()}it(c,"resetOutsideInteraction");function g(){return Array.from(f.current.values()).some(Boolean)}it(g,"isOutsideInteractionIntercepted");function L(m){if(!s.current)return;let v=m.target;v instanceof Node&&[...n].some(I=>I.contains(v))||f.current.set(m.type,!0),m.type==="click"&&window.setTimeout(()=>{s.current&&d.current()},0)}it(L,"handleInteractionCapture");function x(m){s.current&&f.current.set(m.type,!1)}it(x,"handleInteractionBubble");let S=it(m=>{if(m.target&&!i.current){let C=function(){a.removeEventListener("click",d.current);let R=g();c(),R||Qp(TT,r,I,{discrete:!0})};var v=C;if(it(C,"handleAndDispatchPointerDownOutsideEvent"),!u(m.target)){a.removeEventListener("click",d.current),c(),i.current=!1;return}let I={originalEvent:m};s.current=!0,l.current=o&&m.button===0,f.current.clear(),!o||m.button!==0?C():(a.removeEventListener("click",d.current),d.current=C,a.addEventListener("click",d.current,{once:!0}))}else a.removeEventListener("click",d.current),c();i.current=!1},"handlePointerDown"),h=["pointerup","mousedown","mouseup","touchstart","touchend","click"];for(let m of h)a.addEventListener(m,L,!0),a.addEventListener(m,x);let p=window.setTimeout(()=>{a.addEventListener("pointerdown",S)},0);return()=>{window.clearTimeout(p),a.removeEventListener("pointerdown",S),a.removeEventListener("click",d.current);for(let m of h)a.removeEventListener(m,L,!0),a.removeEventListener(m,x)}},[a,r,o,l,n,u]),{onPointerDownCapture:it(()=>i.current=!0,"onPointerDownCapture")}}it(Iy,"usePointerDownOutside");function by(e,t=globalThis==null?void 0:globalThis.document){let a=ya(e),o=me.useRef(!1);return me.useEffect(()=>{let l=it(n=>{n.target&&!o.current&&Qp(ET,a,{originalEvent:n},{discrete:!1})},"handleFocus");return t.addEventListener("focusin",l),()=>t.removeEventListener("focusin",l)},[t,a]),{onFocusCapture:it(()=>o.current=!0,"onFocusCapture"),onBlurCapture:it(()=>o.current=!1,"onBlurCapture")}}it(by,"useFocusOutside");function Zp(){let e=new CustomEvent(Kp);document.dispatchEvent(e)}it(Zp,"dispatchUpdate");function Qp(e,t,a,{discrete:o}){let l=a.originalEvent.target,n=new CustomEvent(e,{bubbles:!1,cancelable:!0,detail:a});t&&l.addEventListener(e,t,{once:!0}),o?Tp(l,n):l.dispatchEvent(n)}it(Qp,"handleAndDispatchCustomEvent");var wy=T(B(),1),OT=Object.defineProperty,Jp=(e,t)=>OT(e,"name",{value:t,configurable:!0}),Bs=0,Wa=null;function PT(e){return _s(),e.children}Jp(PT,"FocusGuards");function _s(){wy.useEffect(()=>{Wa||(Wa={start:Wp(),end:Wp()});let{start:e,end:t}=Wa;return document.body.firstElementChild!==e&&document.body.insertAdjacentElement("afterbegin",e),document.body.lastElementChild!==t&&document.body.insertAdjacentElement("beforeend",t),Bs++,()=>{Bs===1&&(Wa==null||Wa.start.remove(),Wa==null||Wa.end.remove(),Wa=null),Bs=Math.max(0,Bs-1)}},[])}Jp(_s,"useFocusGuards");function Wp(){let e=document.createElement("span");return e.setAttribute("data-radix-focus-guard",""),e.tabIndex=0,e.style.outline="none",e.style.opacity="0",e.style.position="fixed",e.style.pointerEvents="none",e}Jp(Wp,"createFocusGuard");var De=T(B(),1);var om=T(X(),1);var BT=Object.defineProperty,Ve=(e,t)=>BT(e,"name",{value:t,configurable:!0}),$p="focusScope.autoFocusOnMount",em="focusScope.autoFocusOnUnmount",Ay={bubbles:!1,cancelable:!0},Ey=De.forwardRef(Ve(function(t,a){let m=t,{loop:o=!1,trapped:l=!1,branches:n,onMountAutoFocus:u,onUnmountAutoFocus:r}=m,i=D(m,["loop","trapped","branches","onMountAutoFocus","onUnmountAutoFocus"]),[s,f]=De.useState(null),d=ya(u),c=ya(r),g=De.useRef(null),L=ce(a,f),x=De.useRef(n);De.useEffect(()=>{x.current=n});let S=De.useCallback(v=>{var C;return v?s!=null&&s.contains(v)?!0:!!((C=x.current)!=null&&C.some(I=>I.contains(v))):!1},[s]),h=De.useRef({paused:!1,pause(){this.paused=!0},resume(){this.paused=!1}}).current;De.useEffect(()=>{if(l){let R=function(U){if(h.paused||!s)return;let N=U.target;S(N)?g.current=N:wo(g.current,{select:!0})},w=function(U){if(h.paused||!s)return;let N=U.relatedTarget;N!==null&&(S(N)||wo(g.current,{select:!0}))},A=function(U){if(document.activeElement===document.body)for(let H of U)H.removedNodes.length>0&&wo(s)};var v=R,C=w,I=A;Ve(R,"handleFocusIn"),Ve(w,"handleFocusOut"),Ve(A,"handleMutations"),document.addEventListener("focusin",R),document.addEventListener("focusout",w);let M=new MutationObserver(A);return s&&M.observe(s,{childList:!0,subtree:!0}),()=>{document.removeEventListener("focusin",R),document.removeEventListener("focusout",w),M.disconnect()}}},[l,s,h.paused,S]),De.useEffect(()=>{if(s){Ty.add(h);let v=document.activeElement;if(!s.contains(v)){let I=new CustomEvent($p,Ay);s.addEventListener($p,d),s.dispatchEvent(I),I.defaultPrevented||(My(Uy(rm(s)),{select:!0}),document.activeElement===v&&wo(s))}return()=>{s.removeEventListener($p,d),setTimeout(()=>{let I=new CustomEvent(em,Ay);s.addEventListener(em,c),s.dispatchEvent(I),I.defaultPrevented||wo(v!=null?v:document.body,{select:!0}),s.removeEventListener(em,c),Ty.remove(h)},0)}}},[s,d,c,h]);let p=De.useCallback(v=>{if(!o&&!l||h.paused)return;let C=v.key==="Tab"&&!v.altKey&&!v.ctrlKey&&!v.metaKey,I=document.activeElement;if(C&&I){let R=v.currentTarget,[w,A]=Oy(R);w&&A?!v.shiftKey&&I===A?(v.preventDefault(),o&&wo(w,{select:!0})):v.shiftKey&&I===w&&(v.preventDefault(),o&&wo(A,{select:!0})):I===R&&v.preventDefault()}},[o,l,h.paused]);return(0,om.jsx)(Qe.div,b(y({tabIndex:-1},i),{ref:L,onKeyDown:p}))},"FocusScope")),Dy=De.createContext(null);function lm({registry:e,children:t}){return e?(0,om.jsx)(Dy.Provider,{value:e,children:t}):t}Ve(lm,"FocusScopeBranchProvider");function nm(){let[e,t]=De.useState([]),a=De.useMemo(()=>({add:Ve(o=>t(l=>l.includes(o)?l:[...l,o]),"add"),remove:Ve(o=>t(l=>l.filter(n=>n!==o)),"remove")}),[]);return{nodes:e,registry:a}}Ve(nm,"useFocusScopeBranchRegistry");function um(e){let t=De.useContext(Dy);De.useEffect(()=>{if(!(!e||!t))return t.add(e),()=>t.remove(e)},[e,t])}Ve(um,"useFocusScopeBranch");function My(e,{select:t=!1}={}){let a=document.activeElement;for(let o of e)if(wo(o,{select:t}),document.activeElement!==a)return}Ve(My,"focusFirst");function Oy(e){let t=rm(e),a=tm(t,e),o=tm(t.reverse(),e);return[a,o]}Ve(Oy,"getTabbableEdges");function rm(e){let t=[],a=document.createTreeWalker(e,NodeFilter.SHOW_ELEMENT,{acceptNode:Ve(o=>{let l=o.tagName==="INPUT"&&o.type==="hidden";return o.disabled||o.hidden||l?NodeFilter.FILTER_SKIP:o.tabIndex>=0?NodeFilter.FILTER_ACCEPT:NodeFilter.FILTER_SKIP},"acceptNode")});for(;a.nextNode();)t.push(a.currentNode);return t}Ve(rm,"getTabbableCandidates");function tm(e,t){let a=typeof t.checkVisibility=="function"&&t.checkVisibility({checkVisibilityCSS:!0});for(let o of e)if(!(a?!o.checkVisibility({checkVisibilityCSS:!0}):Py(o,{upTo:t})))return o}Ve(tm,"findVisible");function Py(e,{upTo:t}){if(getComputedStyle(e).visibility==="hidden")return!0;for(;e;){if(t!==void 0&&e===t)return!1;if(getComputedStyle(e).display==="none")return!0;e=e.parentElement}return!1}Ve(Py,"isHidden");function By(e){return e instanceof HTMLInputElement&&"select"in e}Ve(By,"isSelectableInput");function wo(e,{select:t=!1}={}){if(e&&e.focus){let a=document.activeElement;e.focus({preventScroll:!0}),e!==a&&By(e)&&t&&e.select()}}Ve(wo,"focus");var Ty=_y();function _y(){let e=[];return{add(t){let a=e[0];t!==a&&(a==null||a.pause()),e=am(e,t),e.unshift(t)},remove(t){var a;e=am(e,t),(a=e[0])==null||a.resume()}}}Ve(_y,"createFocusScopesStack");function am(e,t){let a=[...e],o=a.indexOf(t);return o!==-1&&a.splice(o,1),a}Ve(am,"arrayRemove");function Uy(e){return e.filter(t=>t.tagName!=="A")}Ve(Uy,"removeLinks");var Ct=T(B(),1);var Hy=["top","right","bottom","left"];var Ja=Math.min,Ba=Math.max,Tr=Math.round,Er=Math.floor,$a=e=>({x:e,y:e}),_T={left:"right",right:"left",bottom:"top",top:"bottom"};function im(e,t,a){return Ba(e,Ja(t,a))}function eo(e,t){return typeof e=="function"?e(t):e}function Ao(e){return e.split("-")[0]}function Vl(e){return e.split("-")[1]}function Ns(e){return e==="x"?"y":"x"}function ks(e){return e==="y"?"height":"width"}function _a(e){let t=e[0];return t==="t"||t==="b"?"y":"x"}function Hs(e){return Ns(_a(e))}function zy(e,t,a){a===void 0&&(a=!1);let o=Vl(e),l=Hs(e),n=ks(l),u=l==="x"?o===(a?"end":"start")?"right":"left":o==="start"?"bottom":"top";return t.reference[n]>t.floating[n]&&(u=Ar(u)),[u,Ar(u)]}function Fy(e){let t=Ar(e);return[Us(e),t,Us(t)]}function Us(e){return e.includes("start")?e.replace("start","end"):e.replace("end","start")}var Ny=["left","right"],ky=["right","left"],UT=["top","bottom"],NT=["bottom","top"];function kT(e,t,a){switch(e){case"top":case"bottom":return a?t?ky:Ny:t?Ny:ky;case"left":case"right":return t?UT:NT;default:return[]}}function qy(e,t,a,o){let l=Vl(e),n=kT(Ao(e),a==="start",o);return l&&(n=n.map(u=>u+"-"+l),t&&(n=n.concat(n.map(Us)))),n}function Ar(e){let t=Ao(e);return _T[t]+e.slice(t.length)}function HT(e){var t,a,o,l;return{top:(t=e.top)!=null?t:0,right:(a=e.right)!=null?a:0,bottom:(o=e.bottom)!=null?o:0,left:(l=e.left)!=null?l:0}}function sm(e){return typeof e!="number"?HT(e):{top:e,right:e,bottom:e,left:e}}function Xl(e){let{x:t,y:a,width:o,height:l}=e;return{width:o,height:l,top:a,left:t,right:t+o,bottom:a+l,x:t,y:a}}function Gy(e,t,a){let{reference:o,floating:l}=e,n=_a(t),u=Hs(t),r=ks(u),i=Ao(t),s=n==="y",f=o.x+o.width/2-l.width/2,d=o.y+o.height/2-l.height/2,c=o[r]/2-l[r]/2,g;switch(i){case"top":g={x:f,y:o.y-l.height};break;case"bottom":g={x:f,y:o.y+o.height};break;case"right":g={x:o.x+o.width,y:d};break;case"left":g={x:o.x-l.width,y:d};break;default:g={x:o.x,y:o.y}}let L=Vl(t);return L&&(g[u]+=c*(L==="end"?1:-1)*(a&&s?-1:1)),g}function jy(e,t){return ua(this,null,function*(){var a;t===void 0&&(t={});let{x:o,y:l,platform:n,rects:u,elements:r,strategy:i}=e,{boundary:s="clippingAncestors",rootBoundary:f="viewport",elementContext:d="floating",altBoundary:c=!1,padding:g=0}=eo(t,e),L=sm(g),S=r[c?d==="floating"?"reference":"floating":d],h=Xl(yield n.getClippingRect({element:(a=yield n.isElement==null?void 0:n.isElement(S))==null||a?S:S.contextElement||(yield n.getDocumentElement==null?void 0:n.getDocumentElement(r.floating)),boundary:s,rootBoundary:f,strategy:i})),p=d==="floating"?{x:o,y:l,width:u.floating.width,height:u.floating.height}:u.reference,m=yield n.getOffsetParent==null?void 0:n.getOffsetParent(r.floating),v=(yield n.isElement==null?void 0:n.isElement(m))&&(yield n.getScale==null?void 0:n.getScale(m))||{x:1,y:1},C=Xl(n.convertOffsetParentRelativeRectToViewportRelativeRect?yield n.convertOffsetParentRelativeRectToViewportRelativeRect({elements:r,rect:p,offsetParent:m,strategy:i}):p);return{top:(h.top-C.top+L.top)/v.y,bottom:(C.bottom-h.bottom+L.bottom)/v.y,left:(h.left-C.left+L.left)/v.x,right:(C.right-h.right+L.right)/v.x}})}var zT=50,Yy=(e,t,a)=>ua(null,null,function*(){let{placement:o="bottom",strategy:l="absolute",middleware:n=[],platform:u}=a,r=u.detectOverflow?u:b(y({},u),{detectOverflow:jy}),i=yield u.isRTL==null?void 0:u.isRTL(t),s=yield u.getElementRects({reference:e,floating:t,strategy:l}),{x:f,y:d}=Gy(s,o,i),c=o,g=0,L={};for(let x=0;x<n.length;x++){let S=n[x];if(!S)continue;let{name:h,fn:p}=S,{x:m,y:v,data:C,reset:I}=yield p({x:f,y:d,initialPlacement:o,placement:c,strategy:l,middlewareData:L,rects:s,platform:r,elements:{reference:e,floating:t}});f=m!=null?m:f,d=v!=null?v:d,L[h]=y(y({},L[h]),C),I&&g<zT&&(g++,typeof I=="object"&&(I.placement&&(c=I.placement),I.rects&&(s=I.rects===!0?yield u.getElementRects({reference:e,floating:t,strategy:l}):I.rects),{x:f,y:d}=Gy(s,c,i)),x=-1)}return{x:f,y:d,placement:c,strategy:l,middlewareData:L}}),Ky=e=>({name:"arrow",options:e,fn(a){return ua(this,null,function*(){let{x:o,y:l,placement:n,rects:u,platform:r,elements:i,middlewareData:s}=a,{element:f,padding:d=0}=eo(e,a)||{};if(f==null)return{};let c=sm(d),g={x:o,y:l},L=Hs(n),x=ks(L),S=yield r.getDimensions(f),h=L==="y",p=h?"top":"left",m=h?"bottom":"right",v=h?"clientHeight":"clientWidth",C=u.reference[x]+u.reference[L]-g[L]-u.floating[x],I=g[L]-u.reference[L],R=yield r.getOffsetParent==null?void 0:r.getOffsetParent(f),w=R?R[v]:0;(!w||!(yield r.isElement==null?void 0:r.isElement(R)))&&(w=i.floating[v]||u.floating[x]);let A=C/2-I/2,M=w/2-S[x]/2-1,U=Ja(c[p],M),N=Ja(c[m],M),H=w-S[x]-N,O=w/2-S[x]/2+A,G=im(U,O,H),k=!s.arrow&&Vl(n)!=null&&O!==G&&u.reference[x]/2-(O<U?U:N)-S[x]/2<0,j=k?O<U?O-U:O-H:0;return{[L]:g[L]+j,data:y({[L]:G,centerOffset:O-G-j},k&&{alignmentOffset:j}),reset:k}})}});var Zy=function(e){return e===void 0&&(e={}),{name:"flip",options:e,fn(a){return ua(this,null,function*(){var o,l;let{placement:n,middlewareData:u,rects:r,initialPlacement:i,platform:s,elements:f}=a,G=eo(e,a),{mainAxis:d=!0,crossAxis:c=!0,fallbackPlacements:g,fallbackStrategy:L="bestFit",fallbackAxisSideDirection:x="none",flipAlignment:S=!0}=G,h=D(G,["mainAxis","crossAxis","fallbackPlacements","fallbackStrategy","fallbackAxisSideDirection","flipAlignment"]);if((o=u.arrow)!=null&&o.alignmentOffset)return{};let p=Ao(n),m=_a(i),v=Ao(i)===i,C=yield s.isRTL==null?void 0:s.isRTL(f.floating),I=g||(v||!S?[Ar(i)]:Fy(i)),R=x!=="none";!g&&R&&I.push(...qy(i,S,x,C));let w=[i,...I],A=yield s.detectOverflow(a,h),M=[],U=((l=u.flip)==null?void 0:l.overflows)||[];if(d&&M.push(A[p]),c){let k=zy(n,r,C);M.push(A[k[0]],A[k[1]])}if(U=[...U,{placement:n,overflows:M}],!M.every(k=>k<=0)){var N,H;let k=(((N=u.flip)==null?void 0:N.index)||0)+1,j=w[k];if(j&&(!(c==="alignment"?m!==_a(j):!1)||U.every(J=>_a(J.placement)===m?J.overflows[0]>0:!0)))return{data:{index:k,overflows:U},reset:{placement:j}};let P=(H=U.filter(Z=>Z.overflows[0]<=0).sort((Z,J)=>Z.overflows[1]-J.overflows[1])[0])==null?void 0:H.placement;if(!P)switch(L){case"bestFit":{var O;let Z=(O=U.filter(J=>{if(R){let ge=_a(J.placement);return ge===m||ge==="y"}return!0}).map(J=>[J.placement,J.overflows.filter(ge=>ge>0).reduce((ge,ue)=>ge+ue,0)]).sort((J,ge)=>J[1]-ge[1])[0])==null?void 0:O[0];Z&&(P=Z);break}case"initialPlacement":P=i;break}if(n!==P)return{reset:{placement:P}}}return{}})}}};function Vy(e,t){return{top:e.top-t.height,right:e.right-t.width,bottom:e.bottom-t.height,left:e.left-t.width}}function Xy(e){return Hy.some(t=>e[t]>=0)}var Qy=function(e){return e===void 0&&(e={}),{name:"hide",options:e,fn(a){return ua(this,null,function*(){let{rects:o,platform:l}=a,r=eo(e,a),{strategy:n="referenceHidden"}=r,u=D(r,["strategy"]);switch(n){case"referenceHidden":{let i=yield l.detectOverflow(a,b(y({},u),{elementContext:"reference"})),s=Vy(i,o.reference);return{data:{referenceHiddenOffsets:s,referenceHidden:Xy(s)}}}case"escaped":{let i=yield l.detectOverflow(a,b(y({},u),{altBoundary:!0})),s=Vy(i,o.floating);return{data:{escapedOffsets:s,escaped:Xy(s)}}}default:return{}}})}}};var Wy=new Set(["left","top"]);function FT(e,t){return ua(this,null,function*(){let{placement:a,platform:o,elements:l}=e,n=yield o.isRTL==null?void 0:o.isRTL(l.floating),u=Ao(a),r=Vl(a),i=_a(a)==="y",s=Wy.has(u)?-1:1,f=n&&i?-1:1,d=eo(t,e),{mainAxis:c,crossAxis:g,alignmentAxis:L}=typeof d=="number"?{mainAxis:d,crossAxis:0,alignmentAxis:null}:{mainAxis:d.mainAxis||0,crossAxis:d.crossAxis||0,alignmentAxis:d.alignmentAxis};return r&&typeof L=="number"&&(g=r==="end"?L*-1:L),i?{x:g*f,y:c*s}:{x:c*s,y:g*f}})}var Jy=function(e){return e===void 0&&(e=0),{name:"offset",options:e,fn(a){return ua(this,null,function*(){var o,l;let{x:n,y:u,placement:r,middlewareData:i}=a,s=yield FT(a,e);return r===((o=i.offset)==null?void 0:o.placement)&&(l=i.arrow)!=null&&l.alignmentOffset?{}:{x:n+s.x,y:u+s.y,data:b(y({},s),{placement:r})}})}}},$y=function(e){return e===void 0&&(e={}),{name:"shift",options:e,fn(a){return ua(this,null,function*(){let{x:o,y:l,placement:n,platform:u}=a,m=eo(e,a),{mainAxis:r=!0,crossAxis:i=!1,limiter:s={fn:v=>{let{x:C,y:I}=v;return{x:C,y:I}}}}=m,f=D(m,["mainAxis","crossAxis","limiter"]),d={x:o,y:l},c=yield u.detectOverflow(a,f),g=_a(n),L=Ns(g),x=d[L],S=d[g],h=(v,C)=>im(C+c[v==="y"?"top":"left"],C,C-c[v==="y"?"bottom":"right"]);r&&(x=h(L,x)),i&&(S=h(g,S));let p=s.fn(b(y({},a),{[L]:x,[g]:S}));return b(y({},p),{data:{x:p.x-o,y:p.y-l,enabled:{[L]:r,[g]:i}}})})}}},eC=function(e){return e===void 0&&(e={}),{options:e,fn(t){var a,o;let{x:l,y:n,placement:u,rects:r,middlewareData:i}=t,{offset:s=0,mainAxis:f=!0,crossAxis:d=!0}=eo(e,t),c={x:l,y:n},g=_a(u),L=Ns(g),x=c[L],S=c[g],h=eo(s,t),p=typeof h=="number"?{mainAxis:h,crossAxis:0}:{mainAxis:(a=h.mainAxis)!=null?a:0,crossAxis:(o=h.crossAxis)!=null?o:0};if(f){let C=L==="y"?"height":"width",I=r.reference[L]-r.floating[C]+p.mainAxis,R=r.reference[L]+r.reference[C]-p.mainAxis;x<I?x=I:x>R&&(x=R)}if(d){var m,v;let C=L==="y"?"width":"height",I=Wy.has(Ao(u)),R=r.reference[g]-r.floating[C]+(I&&((m=i.offset)==null?void 0:m[g])||0)+(I?0:p.crossAxis),w=r.reference[g]+r.reference[C]+(I?0:((v=i.offset)==null?void 0:v[g])||0)-(I?p.crossAxis:0);S<R?S=R:S>w&&(S=w)}return{[L]:x,[g]:S}}}},tC=function(e){return e===void 0&&(e={}),{name:"size",options:e,fn(a){return ua(this,null,function*(){let{placement:o,rects:l,platform:n,elements:u}=a,M=eo(e,a),{apply:r=()=>{}}=M,i=D(M,["apply"]),s=yield n.detectOverflow(a,i),f=Ao(o),d=Vl(o),c=_a(o)==="y",{width:g,height:L}=l.floating,x,S;f==="top"||f==="bottom"?(x=f,S=d===((yield n.isRTL==null?void 0:n.isRTL(u.floating))?"start":"end")?"left":"right"):(S=f,x=d==="end"?"top":"bottom");let h=L-s.top-s.bottom,p=g-s.left-s.right,m=Ja(L-s[x],h),v=Ja(g-s[S],p),C=a.middlewareData.shift,I=!C,R=m,w=v;C!=null&&C.enabled.x&&(w=p),C!=null&&C.enabled.y&&(R=h),I&&!d&&(c?w=g-2*Ba(s.left,s.right):R=L-2*Ba(s.top,s.bottom)),yield r(b(y({},a),{availableWidth:w,availableHeight:R}));let A=yield n.getDimensions(u.floating);return g!==A.width||L!==A.height?{reset:{rects:!0}}:{}})}}};function zs(){return typeof window!="undefined"}function Kl(e){return oC(e)?(e.nodeName||"").toLowerCase():"#document"}function Mt(e){var t;return(e==null||(t=e.ownerDocument)==null?void 0:t.defaultView)||window}function to(e){var t;return(t=(oC(e)?e.ownerDocument:e.document)||window.document)==null?void 0:t.documentElement}function oC(e){return zs()?e instanceof Node||e instanceof Mt(e).Node:!1}function Ua(e){return zs()?e instanceof Element||e instanceof Mt(e).Element:!1}function To(e){return zs()?e instanceof HTMLElement||e instanceof Mt(e).HTMLElement:!1}function aC(e){return!zs()||typeof ShadowRoot=="undefined"?!1:e instanceof ShadowRoot||e instanceof Mt(e).ShadowRoot}function Dr(e){let{overflow:t,overflowX:a,overflowY:o,display:l}=Na(e);return/auto|scroll|overlay|hidden|clip/.test(t+o+a)&&l!=="inline"&&l!=="contents"}function lC(e){return/^(table|td|th)$/.test(Kl(e))}function Mr(e){try{if(e.matches(":popover-open"))return!0}catch(t){}try{return e.matches(":modal")}catch(t){return!1}}var qT=/transform|translate|scale|rotate|perspective|filter/,GT=/paint|layout|strict|content/,jl=e=>!!e&&e!=="none",fm;function Fs(e){let t=Ua(e)?Na(e):e;return jl(t.transform)||jl(t.translate)||jl(t.scale)||jl(t.rotate)||jl(t.perspective)||!qs()&&(jl(t.backdropFilter)||jl(t.filter))||qT.test(t.willChange||"")||GT.test(t.contain||"")}function nC(e){let t=cl(e);for(;To(t)&&!tu(t);){if(Fs(t))return t;if(Mr(t))return null;t=cl(t)}return null}function qs(){return fm==null&&(fm=typeof CSS!="undefined"&&CSS.supports&&CSS.supports("-webkit-backdrop-filter","none")),fm}function tu(e){return/^(html|body|#document)$/.test(Kl(e))}function Na(e){return Mt(e).getComputedStyle(e)}function Or(e){return Ua(e)?{scrollLeft:e.scrollLeft,scrollTop:e.scrollTop}:{scrollLeft:e.scrollX,scrollTop:e.scrollY}}function cl(e){if(Kl(e)==="html")return e;let t=e.assignedSlot||e.parentNode||aC(e)&&e.host||to(e);return aC(t)?t.host:t}function uC(e){let t=cl(e);return tu(t)?(e.ownerDocument||e).body:To(t)&&Dr(t)?t:uC(t)}function Yl(e,t,a){var o;t===void 0&&(t=[]),a===void 0&&(a=!0);let l=uC(e),n=l===((o=e.ownerDocument)==null?void 0:o.body),u=Mt(l);if(n){let r=Gs(u);return t.concat(u,u.visualViewport||[],Dr(l)?l:[],r&&a?Yl(r):[])}else return t.concat(l,Yl(l,[],a))}function Gs(e){return e.parent&&Object.getPrototypeOf(e.parent)?e.frameElement:null}function sC(e){let t=Na(e),a=parseFloat(t.width)||0,o=parseFloat(t.height)||0,l=To(e),n=l?e.offsetWidth:a,u=l?e.offsetHeight:o,r=Tr(a)!==n||Tr(o)!==u;return r&&(a=n,o=u),{width:a,height:o,$:r}}function dm(e){return Ua(e)?e:e.contextElement}function au(e){let t=dm(e);if(!To(t))return $a(1);let a=t.getBoundingClientRect(),{width:o,height:l,$:n}=sC(t),u=(n?Tr(a.width):a.width)/o,r=(n?Tr(a.height):a.height)/l;return(!u||!Number.isFinite(u))&&(u=1),(!r||!Number.isFinite(r))&&(r=1),{x:u,y:r}}var VT=$a(0);function fC(e){let t=Mt(e);return!qs()||!t.visualViewport?VT:{x:t.visualViewport.offsetLeft,y:t.visualViewport.offsetTop}}function XT(e,t,a){return t===void 0&&(t=!1),!!a&&t&&a===Mt(e)}function Zl(e,t,a,o){t===void 0&&(t=!1),a===void 0&&(a=!1);let l=e.getBoundingClientRect(),n=dm(e),u=$a(1);t&&(o?Ua(o)&&(u=au(o)):u=au(e));let r=XT(n,a,o)?fC(n):$a(0),i=(l.left+r.x)/u.x,s=(l.top+r.y)/u.y,f=l.width/u.x,d=l.height/u.y;if(n&&o){let c=Mt(n),g=Ua(o)?Mt(o):o,L=c,x=Gs(L);for(;x&&g!==L;){let S=au(x),h=x.getBoundingClientRect(),p=Na(x),m=h.left+(x.clientLeft+parseFloat(p.paddingLeft))*S.x,v=h.top+(x.clientTop+parseFloat(p.paddingTop))*S.y;i*=S.x,s*=S.y,f*=S.x,d*=S.y,i+=m,s+=v,L=Mt(x),x=Gs(L)}}return Xl({width:f,height:d,x:i,y:s})}function Vs(e,t){let a=Or(e).scrollLeft;return t?t.left+a:Zl(to(e)).left+a}function cC(e,t){let a=e.getBoundingClientRect(),o=a.left+t.scrollLeft-Vs(e,a),l=a.top+t.scrollTop;return{x:o,y:l}}function jT(e){let{elements:t,rect:a,offsetParent:o,strategy:l}=e,n=l==="fixed",u=to(o),r=t?Mr(t.floating):!1;if(o===u||r&&n)return a;let i={scrollLeft:0,scrollTop:0},s=$a(1),f=$a(0),d=To(o);if((d||!n)&&((Kl(o)!=="body"||Dr(u))&&(i=Or(o)),d)){let g=Zl(o);s=au(o),f.x=g.x+o.clientLeft,f.y=g.y+o.clientTop}let c=u&&!d&&!n?cC(u,i):$a(0);return{width:a.width*s.x,height:a.height*s.y,x:a.x*s.x-i.scrollLeft*s.x+f.x+c.x,y:a.y*s.y-i.scrollTop*s.y+f.y+c.y}}function YT(e){return e.getClientRects?Array.from(e.getClientRects()):[]}function KT(e){let t=Or(e),a=e.ownerDocument.body,o=Ba(e.scrollWidth,e.clientWidth,a.scrollWidth,a.clientWidth),l=Ba(e.scrollHeight,e.clientHeight,a.scrollHeight,a.clientHeight),n=-t.scrollLeft+Vs(e),u=-t.scrollTop;return Na(a).direction==="rtl"&&(n+=Ba(e.clientWidth,a.clientWidth)-o),{width:o,height:l,x:n,y:u}}var ZT=25;function QT(e,t,a){a===void 0&&(a="viewport");let o=a==="layoutViewport",l=Mt(e),n=to(e),u=l.visualViewport,r=n.clientWidth,i=n.clientHeight,s=0,f=0;if(u){let c=!qs()||t==="fixed";o?c||(s=-u.offsetLeft,f=-u.offsetTop):(r=u.width,i=u.height,c&&(s=u.offsetLeft,f=u.offsetTop))}if(Vs(n)<=0){let c=n.ownerDocument,g=c.body,L=getComputedStyle(g),x=c.compatMode==="CSS1Compat"&&parseFloat(L.marginLeft)+parseFloat(L.marginRight)||0,S=Math.abs(n.clientWidth-g.clientWidth-x),h=getComputedStyle(n).scrollbarGutter==="stable both-edges"?S/2:S;h<=ZT&&(r-=h)}return{width:r,height:i,x:s,y:f}}function WT(e,t){let a=Zl(e,!0,t==="fixed"),o=a.top+e.clientTop,l=a.left+e.clientLeft,n=au(e),u=e.clientWidth*n.x,r=e.clientHeight*n.y,i=l*n.x,s=o*n.y;return{width:u,height:r,x:i,y:s}}function rC(e,t,a){let o;if(t==="viewport"||t==="layoutViewport")o=QT(e,a,t);else if(t==="document")o=KT(to(e));else if(Ua(t))o=WT(t,a);else{let l=fC(e);o={x:t.x-l.x,y:t.y-l.y,width:t.width,height:t.height}}return Xl(o)}function JT(e,t){let a=t.get(e);if(a)return a;let o=Yl(e,[],!1).filter(r=>Ua(r)&&Kl(r)!=="body"),l=null,n=Na(e).position==="fixed",u=n?cl(e):e;for(;Ua(u)&&!tu(u);){let r=Na(u),i=Fs(u),s=l?l.position:n?"fixed":"";!i&&(s==="fixed"||s==="absolute"&&r.position==="static")?o=o.filter(d=>d!==u):l=r,u=cl(u)}return t.set(e,o),o}function $T(e){let{element:t,boundary:a,rootBoundary:o,strategy:l}=e,u=[...a==="clippingAncestors"?Mr(t)?[]:JT(t,this._c):[].concat(a),o],r=rC(t,u[0],l),i=r.top,s=r.right,f=r.bottom,d=r.left;for(let c=1;c<u.length;c++){let g=rC(t,u[c],l);i=Ba(g.top,i),s=Ja(g.right,s),f=Ja(g.bottom,f),d=Ba(g.left,d)}return{width:s-d,height:f-i,x:d,y:i}}function eE(e){let{width:t,height:a}=sC(e);return{width:t,height:a}}function tE(e,t,a){let o=To(t),l=to(t),n=a==="fixed",u=Zl(e,!0,n,t),r={scrollLeft:0,scrollTop:0},i=$a(0);if((o||!n)&&((Kl(t)!=="body"||Dr(l))&&(r=Or(t)),o)){let c=Zl(t,!0,n,t);i.x=c.x+t.clientLeft,i.y=c.y+t.clientTop}!o&&l&&(i.x=Vs(l));let s=l&&!o&&!n?cC(l,r):$a(0),f=u.left+r.scrollLeft-i.x-s.x,d=u.top+r.scrollTop-i.y-s.y;return{x:f,y:d,width:u.width,height:u.height}}function cm(e){return Na(e).position==="static"}function iC(e,t){if(!To(e)||Na(e).position==="fixed")return null;if(t)return t(e);let a=e.offsetParent;return to(e)===a&&(a=a.ownerDocument.body),a}function dC(e,t){let a=Mt(e);if(Mr(e))return a;if(!To(e)){let l=cl(e);for(;l&&!tu(l);){if(Ua(l)&&!cm(l))return l;l=cl(l)}return a}let o=iC(e,t);for(;o&&lC(o)&&cm(o);)o=iC(o,t);return o&&tu(o)&&cm(o)&&!Fs(o)?a:o||nC(e)||a}var aE=function(e){return ua(this,null,function*(){let t=this.getOffsetParent||dC,a=this.getDimensions,o=yield a(e.floating);return{reference:tE(e.reference,yield t(e.floating),e.strategy),floating:{x:0,y:0,width:o.width,height:o.height}}})};function oE(e){return Na(e).direction==="rtl"}var pC={convertOffsetParentRelativeRectToViewportRelativeRect:jT,getDocumentElement:to,getClippingRect:$T,getOffsetParent:dC,getElementRects:aE,getClientRects:YT,getDimensions:eE,getScale:au,isElement:Ua,isRTL:oE};function mC(e,t){return e.x===t.x&&e.y===t.y&&e.width===t.width&&e.height===t.height}function lE(e,t,a){let o=null,l,n=to(e);function u(){var f;clearTimeout(l),(f=o)==null||f.disconnect(),o=null}function r(f,d){f===void 0&&(f=!1),d===void 0&&(d=1),u();let c=e.getBoundingClientRect(),{left:g,top:L,width:x,height:S}=c;if(f||t(),!x||!S)return;let h=Er(L),p=Er(n.clientWidth-(g+x)),m=Er(n.clientHeight-(L+S)),v=Er(g),I={rootMargin:-h+"px "+-p+"px "+-m+"px "+-v+"px",threshold:Ba(0,Ja(1,d))||1},R=!0;function w(A){let M=A[0].intersectionRatio;if(!mC(c,e.getBoundingClientRect()))return r();if(M!==d){if(!R)return r();M?r(!1,M):l=setTimeout(()=>{r(!1,1e-7)},1e3)}R=!1}try{o=new IntersectionObserver(w,b(y({},I),{root:n.ownerDocument}))}catch(A){o=new IntersectionObserver(w,I)}o.observe(e)}let i=Mt(e),s=()=>r(a);return i.addEventListener("resize",s),r(!0),()=>{i.removeEventListener("resize",s),u()}}function Pr(e,t,a,o){o===void 0&&(o={});let{ancestorScroll:l=!0,ancestorResize:n=!0,elementResize:u=typeof ResizeObserver=="function",layoutShift:r=typeof IntersectionObserver=="function",animationFrame:i=!1}=o,s=dm(e),f=l||n?[...s?Yl(s):[],...t?Yl(t):[]]:[];f.forEach(h=>{l&&h.addEventListener("scroll",a),n&&h.addEventListener("resize",a)});let d=s&&r?lE(s,a,n):null,c=-1,g=null;u&&(g=new ResizeObserver(h=>{let[p]=h;p&&p.target===s&&g&&t&&(g.unobserve(t),cancelAnimationFrame(c),c=requestAnimationFrame(()=>{var m;(m=g)==null||m.observe(t)})),a()}),s&&!i&&g.observe(s),t&&g.observe(t));let L,x=i?Zl(e):null;i&&S();function S(){let h=Zl(e);x&&!mC(x,h)&&a(),x=h,L=requestAnimationFrame(S)}return a(),()=>{var h;f.forEach(p=>{l&&p.removeEventListener("scroll",a),n&&p.removeEventListener("resize",a)}),d==null||d(),(h=g)==null||h.disconnect(),g=null,i&&cancelAnimationFrame(L)}}var hC=Jy;var gC=$y,vC=Zy,SC=tC,xC=Qy,pm=Ky;var LC=eC,mm=(e,t,a)=>{let o=new Map,l=a!=null?a:{},n=b(y(y({},pC),l.platform),{_c:o});return Yy(e,t,b(y({},l),{platform:n}))};var Xe=T(B(),1),CC=T(B(),1),RC=T(Va(),1);var nE=typeof document!="undefined",uE=function(){},Xs=nE?CC.useLayoutEffect:uE;function js(e,t){if(e===t)return!0;if(typeof e!=typeof t)return!1;if(typeof e=="function"&&e.toString()===t.toString())return!0;let a,o,l;if(e&&t&&typeof e=="object"){if(Array.isArray(e)){if(a=e.length,a!==t.length)return!1;for(o=a;o--!==0;)if(!js(e[o],t[o]))return!1;return!0}if(l=Object.keys(e),a=l.length,a!==Object.keys(t).length)return!1;for(o=a;o--!==0;)if(!{}.hasOwnProperty.call(t,l[o]))return!1;for(o=a;o--!==0;){let n=l[o];if(!(n==="_owner"&&e.$$typeof)&&!js(e[n],t[n]))return!1}return!0}return e!==e&&t!==t}function IC(e){return typeof window=="undefined"?1:(e.ownerDocument.defaultView||window).devicePixelRatio||1}function yC(e,t){let a=IC(e);return Math.round(t*a)/a}function hm(e){let t=Xe.useRef(e);return Xs(()=>{t.current=e}),t}function Ys(e){e===void 0&&(e={});let{placement:t="bottom",strategy:a="absolute",middleware:o=[],platform:l,elements:{reference:n,floating:u}={},transform:r=!0,whileElementsMounted:i,open:s}=e,[f,d]=Xe.useState({x:0,y:0,strategy:a,placement:t,middlewareData:{},isPositioned:!1}),[c,g]=Xe.useState(o);js(c,o)||g(o);let[L,x]=Xe.useState(null),[S,h]=Xe.useState(null),p=Xe.useCallback(P=>{P!==I.current&&(I.current=P,x(P))},[]),m=Xe.useCallback(P=>{P!==R.current&&(R.current=P,h(P))},[]),v=n||L,C=u||S,I=Xe.useRef(null),R=Xe.useRef(null),w=Xe.useRef(f),A=i!=null,M=hm(i),U=hm(l),N=hm(s),H=Xe.useCallback(()=>{if(!I.current||!R.current)return;let P={placement:t,strategy:a,middleware:c};U.current&&(P.platform=U.current),mm(I.current,R.current,P).then(Z=>{let J=b(y({},Z),{isPositioned:N.current!==!1});O.current&&!js(w.current,J)&&(w.current=J,RC.flushSync(()=>{d(J)}))})},[c,t,a,U,N]);Xs(()=>{s===!1&&w.current.isPositioned&&(w.current.isPositioned=!1,d(P=>b(y({},P),{isPositioned:!1})))},[s]);let O=Xe.useRef(!1);Xs(()=>(O.current=!0,()=>{O.current=!1}),[]),Xs(()=>{if(v&&(I.current=v),C&&(R.current=C),v&&C){if(M.current)return M.current(v,C,H);H()}},[v,C,H,M,A]);let G=Xe.useMemo(()=>({reference:I,floating:R,setReference:p,setFloating:m}),[p,m]),k=Xe.useMemo(()=>({reference:v,floating:C}),[v,C]),j=Xe.useMemo(()=>{let P={position:a,left:0,top:0};if(!k.floating)return P;let Z=yC(k.floating,f.x),J=yC(k.floating,f.y);return r?y(b(y({},P),{transform:"translate("+Z+"px, "+J+"px)"}),IC(k.floating)>=1.5&&{willChange:"transform"}):{position:a,left:Z,top:J}},[a,r,k.floating,f.x,f.y]);return Xe.useMemo(()=>b(y({},f),{update:H,refs:G,elements:k,floatingStyles:j}),[f,H,G,k,j])}var rE=e=>{function t(a){return{}.hasOwnProperty.call(a,"current")}return{name:"arrow",options:e,fn(a){let{element:o,padding:l}=typeof e=="function"?e(a):e;return o&&t(o)?o.current!=null?pm({element:o.current,padding:l}).fn(a):{}:o?pm({element:o,padding:l}).fn(a):{}}}},Ks=(e,t)=>{let a=hC(e);return{name:a.name,fn:a.fn,options:[e,t]}},Zs=(e,t)=>{let a=gC(e);return{name:a.name,fn:a.fn,options:[e,t]}},Qs=(e,t)=>({fn:LC(e).fn,options:[e,t]}),Ws=(e,t)=>{let a=vC(e);return{name:a.name,fn:a.fn,options:[e,t]}},Js=(e,t)=>{let a=SC(e);return{name:a.name,fn:a.fn,options:[e,t]}};var $s=(e,t)=>{let a=xC(e);return{name:a.name,fn:a.fn,options:[e,t]}};var ef=(e,t)=>{let a=rE(e);return{name:a.name,fn:a.fn,options:[e,t]}};var ou=T(X(),1);var iE=Object.defineProperty,dl=(e,t)=>iE(e,"name",{value:t,configurable:!0});var bC={Partial:"partial",Always:"always"},wC={Optimized:"optimized",Always:"always"},AC="Popper",[TC,gm]=Et(AC),[sE,EC]=TC(AC),DC=dl(e=>{let{__scopePopper:t,children:a}=e,[o,l]=Ct.useState(null),[n,u]=Ct.useState(void 0);return(0,ou.jsx)(sE,{scope:t,anchor:o,onAnchorChange:l,placementState:n,setPlacementState:u,children:a})},"Popper"),fE="PopperAnchor",MC=Ct.forwardRef(dl(function(t,a){let x=t,{__scopePopper:o,virtualRef:l}=x,n=D(x,["__scopePopper","virtualRef"]),u=EC(fE,o),r=Ct.useRef(null),i=u.onAnchorChange,s=Ct.useCallback(S=>{r.current=S,S&&i(S)},[i]),f=ce(a,s),d=Ct.useRef(null);Ct.useEffect(()=>{if(!l)return;let S=d.current;d.current=l.current,S!==d.current&&i(d.current)});let c=u.placementState&&tf(u.placementState),g=c==null?void 0:c[0],L=c==null?void 0:c[1];return l?null:(0,ou.jsx)(Qe.div,b(y({"data-radix-popper-side":g,"data-radix-popper-align":L},n),{ref:f}))},"PopperAnchor")),OC="PopperContent",[cE,h2]=TC(OC),PC=Ct.forwardRef(dl(function(t,a){var ve,$,ne,W,at,Ye,Pt,wa,za;let le=t,{__scopePopper:o,side:l="bottom",sideOffset:n=0,align:u="center",alignOffset:r=0,arrowPadding:i=0,avoidCollisions:s=!0,collisionBoundary:f=[],collisionPadding:d=0,sticky:c=bC.Partial,hideWhenDetached:g=!1,updatePositionStrategy:L=wC.Optimized,onPlaced:x}=le,S=D(le,["__scopePopper","side","sideOffset","align","alignOffset","arrowPadding","avoidCollisions","collisionBoundary","collisionPadding","sticky","hideWhenDetached","updatePositionStrategy","onPlaced"]),h=EC(OC,o),[p,m]=Ct.useState(null),v=ce(a,m),[C,I]=Ct.useState(null),R=xr(C),w=(ve=R==null?void 0:R.width)!=null?ve:0,A=($=R==null?void 0:R.height)!=null?$:0,M=l+(u!=="center"?"-"+u:""),U=typeof d=="number"?d:y({top:0,right:0,bottom:0,left:0},d),N=Array.isArray(f)?f:[f],H=N.length>0,O={padding:U,boundary:N.filter(BC),altBoundary:H},{refs:G,floatingStyles:k,placement:j,isPositioned:P,middlewareData:Z}=Ys({strategy:"fixed",placement:M,whileElementsMounted:dl((...xl)=>Pr(...xl,{animationFrame:L===wC.Always}),"whileElementsMounted"),elements:{reference:h.anchor},middleware:[Ks({mainAxis:n+A,alignmentAxis:r}),s&&Zs(y({mainAxis:!0,crossAxis:!1,limiter:c===bC.Partial?Qs():void 0},O)),s&&Ws(y({},O)),Js(b(y({},O),{apply:dl(({elements:xl,rects:oo,availableWidth:hu,availableHeight:Bo})=>{let{width:Pf,height:Bf}=oo.reference,an=xl.floating.style;an.setProperty("--radix-popper-available-width","".concat(hu,"px")),an.setProperty("--radix-popper-available-height","".concat(Bo,"px")),an.setProperty("--radix-popper-anchor-width","".concat(Pf,"px")),an.setProperty("--radix-popper-anchor-height","".concat(Bf,"px"))},"apply")})),C&&ef({element:C,padding:i}),dE({arrowWidth:w,arrowHeight:A}),g&&$s(b(y({strategy:"referenceHidden"},O),{boundary:H?O.boundary:void 0}))]}),J=h.setPlacementState;We(()=>(J(j),()=>{J(void 0)}),[j,J]);let[ge,ue]=tf(j),ct=ya(x);We(()=>{P&&(ct==null||ct())},[P,ct]);let la=(ne=Z.arrow)==null?void 0:ne.x,jt=(W=Z.arrow)==null?void 0:W.y,Ot=((at=Z.arrow)==null?void 0:at.centerOffset)!==0,[tt,F]=Ct.useState();return We(()=>{p&&F(window.getComputedStyle(p).zIndex)},[p]),(0,ou.jsx)("div",{ref:G.setFloating,"data-radix-popper-content-wrapper":"",style:y(b(y({},k),{transform:P?k.transform:"translate(0, -200%)",minWidth:"max-content",zIndex:tt,"--radix-popper-transform-origin":[(Ye=Z.transformOrigin)==null?void 0:Ye.x,(Pt=Z.transformOrigin)==null?void 0:Pt.y].join(" ")}),((wa=Z.hide)==null?void 0:wa.referenceHidden)&&{visibility:"hidden",pointerEvents:"none"}),dir:t.dir,children:(0,ou.jsx)(cE,{scope:o,placedSide:ge,placedAlign:ue,onArrowChange:I,arrowX:la,arrowY:jt,shouldHideArrow:Ot,children:(0,ou.jsx)(Qe.div,b(y({"data-side":ge,"data-align":ue},S),{ref:v,style:b(y({},S.style),{animation:P?(za=S.style)==null?void 0:za.animation:"none"})}))})})},"PopperContent"));function BC(e){return e!==null}dl(BC,"isNotNull");var dE=dl(e=>({name:"transformOrigin",options:e,fn(t){var S,h,p,m,v;let{placement:a,rects:o,middlewareData:l}=t,u=((S=l.arrow)==null?void 0:S.centerOffset)!==0,r=u?0:e.arrowWidth,i=u?0:e.arrowHeight,[s,f]=tf(a),d={start:"0%",center:"50%",end:"100%"}[f],c=((p=(h=l.arrow)==null?void 0:h.x)!=null?p:0)+r/2,g=((v=(m=l.arrow)==null?void 0:m.y)!=null?v:0)+i/2,L="",x="";return s==="bottom"?(L=u?d:"".concat(c,"px"),x="".concat(-i,"px")):s==="top"?(L=u?d:"".concat(c,"px"),x="".concat(o.floating.height+i,"px")):s==="right"?(L="".concat(-i,"px"),x=u?d:"".concat(g,"px")):s==="left"&&(L="".concat(o.floating.width+i,"px"),x=u?d:"".concat(g,"px")),{data:{x:L,y:x}}}}),"transformOrigin");function tf(e){let[t,a="center"]=e.split("-");return[t,a]}dl(tf,"getSideAndAlignFromPlacement");var af=T(B(),1),_C=T(Va(),1);var UC=T(X(),1);var mE=Object.defineProperty,hE=(e,t)=>mE(e,"name",{value:t,configurable:!0}),NC=af.forwardRef(hE(function(t,a){var s;let i=t,{container:o}=i,l=D(i,["container"]),[n,u]=af.useState(!1);We(()=>u(!0),[]);let r=o||n&&((s=globalThis==null?void 0:globalThis.document)==null?void 0:s.body);return r?_C.createPortal((0,UC.jsx)(Qe.div,b(y({},l),{ref:a})),r):null},"Portal"));var gE=function(e){if(typeof document=="undefined")return null;var t=Array.isArray(e)?e[0]:e;return t.ownerDocument.body},lu=new WeakMap,of=new WeakMap,lf={},vm=0,kC=function(e){return e&&(e.host||kC(e.parentNode))},vE=function(e,t){return t.map(function(a){if(e.contains(a))return a;var o=kC(a);return o&&e.contains(o)?o:(console.error("aria-hidden",a,"in not contained inside",e,". Doing nothing"),null)}).filter(function(a){return!!a})},SE=function(e,t,a,o){var l=vE(t,Array.isArray(e)?e:[e]);lf[a]||(lf[a]=new WeakMap);var n=lf[a],u=[],r=new Set,i=new Set(l),s=function(d){!d||r.has(d)||(r.add(d),s(d.parentNode))};l.forEach(s);var f=function(d){!d||i.has(d)||Array.prototype.forEach.call(d.children,function(c){if(r.has(c))f(c);else try{var g=c.getAttribute(o),L=g!==null&&g!=="false",x=(lu.get(c)||0)+1,S=(n.get(c)||0)+1;lu.set(c,x),n.set(c,S),u.push(c),x===1&&L&&of.set(c,!0),S===1&&c.setAttribute(a,"true"),L||c.setAttribute(o,"true")}catch(h){console.error("aria-hidden: cannot operate on ",c,h)}})};return f(t),r.clear(),vm++,function(){u.forEach(function(d){var c=lu.get(d)-1,g=n.get(d)-1;lu.set(d,c),n.set(d,g),c||(of.has(d)||d.removeAttribute(o),of.delete(d)),g||d.removeAttribute(a)}),vm--,vm||(lu=new WeakMap,lu=new WeakMap,of=new WeakMap,lf={})}},nf=function(e,t,a){a===void 0&&(a="data-aria-hidden");var o=Array.from(Array.isArray(e)?e:[e]),l=t||gE(e);return l?(o.push.apply(o,Array.from(l.querySelectorAll("[aria-live], script"))),SE(o,l,a,"aria-hidden")):function(){return null}};var Vt=function(){return Vt=Object.assign||function(t){for(var a,o=1,l=arguments.length;o<l;o++){a=arguments[o];for(var n in a)Object.prototype.hasOwnProperty.call(a,n)&&(t[n]=a[n])}return t},Vt.apply(this,arguments)};function uf(e,t){var a={};for(var o in e)Object.prototype.hasOwnProperty.call(e,o)&&t.indexOf(o)<0&&(a[o]=e[o]);if(e!=null&&typeof Object.getOwnPropertySymbols=="function")for(var l=0,o=Object.getOwnPropertySymbols(e);l<o.length;l++)t.indexOf(o[l])<0&&Object.prototype.propertyIsEnumerable.call(e,o[l])&&(a[o[l]]=e[o[l]]);return a}function HC(e,t,a){if(a||arguments.length===2)for(var o=0,l=t.length,n;o<l;o++)(n||!(o in t))&&(n||(n=Array.prototype.slice.call(t,0,o)),n[o]=t[o]);return e.concat(n||Array.prototype.slice.call(t))}var df=T(B());var Rt=T(B());var Ql="right-scroll-bar-position",Wl="width-before-scroll-bar",Sm="with-scroll-bars-hidden",xm="--removed-body-scroll-bar-size";function rf(e,t){return typeof e=="function"?e(t):e&&(e.current=t),e}var zC=T(B());function FC(e,t){var a=(0,zC.useState)(function(){return{value:e,callback:t,facade:{get current(){return a.value},set current(o){var l=a.value;l!==o&&(a.value=o,a.callback(o,l))}}}})[0];return a.callback=t,a.facade}var sf=T(B());var xE=typeof window!="undefined"?sf.useLayoutEffect:sf.useEffect,qC=new WeakMap;function Lm(e,t){var a=FC(t||null,function(o){return e.forEach(function(l){return rf(l,o)})});return xE(function(){var o=qC.get(a);if(o){var l=new Set(o),n=new Set(e),u=a.current;l.forEach(function(r){n.has(r)||rf(r,null)}),n.forEach(function(r){l.has(r)||rf(r,u)})}qC.set(a,e)},[e]),a}function LE(e){return e}function yE(e,t){t===void 0&&(t=LE);var a=[],o=!1,l={read:function(){if(o)throw new Error("Sidecar: could not \x60read\x60 from an \x60assigned\x60 medium. \x60read\x60 could be used only with \x60useMedium\x60.");return a.length?a[a.length-1]:e},useMedium:function(n){var u=t(n,o);return a.push(u),function(){a=a.filter(function(r){return r!==u})}},assignSyncMedium:function(n){for(o=!0;a.length;){var u=a;a=[],u.forEach(n)}a={push:function(r){return n(r)},filter:function(){return a}}},assignMedium:function(n){o=!0;var u=[];if(a.length){var r=a;a=[],r.forEach(n),u=a}var i=function(){var f=u;u=[],f.forEach(n)},s=function(){return Promise.resolve().then(i)};s(),a={push:function(f){u.push(f),s()},filter:function(f){return u=u.filter(f),a}}}};return l}function ym(e){e===void 0&&(e={});var t=yE(null);return t.options=Vt({async:!0,ssr:!1},e),t}var GC=T(B()),VC=function(e){var t=e.sideCar,a=uf(e,["sideCar"]);if(!t)throw new Error("Sidecar: please provide \x60sideCar\x60 property to import the right car");var o=t.read();if(!o)throw new Error("Sidecar medium not found");return GC.createElement(o,Vt({},a))};VC.isSideCarExport=!0;function Cm(e,t){return e.useMedium(t),VC}var ff=ym();var Rm=function(){},Br=Rt.forwardRef(function(e,t){var a=Rt.useRef(null),o=Rt.useState({onScrollCapture:Rm,onWheelCapture:Rm,onTouchMoveCapture:Rm}),l=o[0],n=o[1],u=e.forwardProps,r=e.children,i=e.className,s=e.removeScrollBar,f=e.enabled,d=e.shards,c=e.sideCar,g=e.noRelative,L=e.noIsolation,x=e.inert,S=e.allowPinchZoom,h=e.as,p=h===void 0?"div":h,m=e.gapMode,v=uf(e,["forwardProps","children","className","removeScrollBar","enabled","shards","sideCar","noRelative","noIsolation","inert","allowPinchZoom","as","gapMode"]),C=c,I=Lm([a,t]),R=Vt(Vt({},v),l);return Rt.createElement(Rt.Fragment,null,f&&Rt.createElement(C,{sideCar:ff,removeScrollBar:s,shards:d,noRelative:g,noIsolation:L,inert:x,setCallbacks:n,allowPinchZoom:!!S,lockRef:a,gapMode:m}),u?Rt.cloneElement(Rt.Children.only(r),Vt(Vt({},R),{ref:I})):Rt.createElement(p,Vt({},R,{className:i,ref:I}),r))});Br.defaultProps={enabled:!0,removeScrollBar:!0,inert:!1};Br.classNames={fullWidth:Wl,zeroRight:Ql};var be=T(B());var uu=T(B());var YC=T(B());var XC;var jC=function(){if(XC)return XC;if(typeof __webpack_nonce__!="undefined")return __webpack_nonce__};function CE(){if(!document)return null;var e=document.createElement("style");e.type="text/css";var t=jC();return t&&e.setAttribute("nonce",t),e}function RE(e,t){e.styleSheet?e.styleSheet.cssText=t:e.appendChild(document.createTextNode(t))}function IE(e){var t=document.head||document.getElementsByTagName("head")[0];t.appendChild(e)}var Im=function(){var e=0,t=null;return{add:function(a){e==0&&(t=CE())&&(RE(t,a),IE(t)),e++},remove:function(){e--,!e&&t&&(t.parentNode&&t.parentNode.removeChild(t),t=null)}}};var bm=function(){var e=Im();return function(t,a){YC.useEffect(function(){return e.add(t),function(){e.remove()}},[t&&a])}};var _r=function(){var e=bm(),t=function(a){var o=a.styles,l=a.dynamic;return e(o,l),null};return t};var bE={left:0,top:0,right:0,gap:0},wm=function(e){return parseInt(e||"",10)||0},wE=function(e){var t=window.getComputedStyle(document.body),a=t[e==="padding"?"paddingLeft":"marginLeft"],o=t[e==="padding"?"paddingTop":"marginTop"],l=t[e==="padding"?"paddingRight":"marginRight"];return[wm(a),wm(o),wm(l)]},Am=function(e){if(e===void 0&&(e="margin"),typeof window=="undefined")return bE;var t=wE(e),a=document.documentElement.clientWidth,o=window.innerWidth;return{left:t[0],top:t[1],right:t[2],gap:Math.max(0,o-a+t[2]-t[0])}};var AE=_r(),nu="data-scroll-locked",TE=function(e,t,a,o){var l=e.left,n=e.top,u=e.right,r=e.gap;return a===void 0&&(a="margin"),"\n  .".concat(Sm," {\n   overflow: hidden ").concat(o,";\n   padding-right: ").concat(r,"px ").concat(o,";\n  }\n  body[").concat(nu,"] {\n    overflow: hidden ").concat(o,";\n    overscroll-behavior: contain;\n    ").concat([t&&"position: relative ".concat(o,";"),a==="margin"&&"\n    padding-left: ".concat(l,"px;\n    padding-top: ").concat(n,"px;\n    padding-right: ").concat(u,"px;\n    margin-left:0;\n    margin-top:0;\n    margin-right: ").concat(r,"px ").concat(o,";\n    "),a==="padding"&&"padding-right: ".concat(r,"px ").concat(o,";")].filter(Boolean).join(""),"\n  }\n  \n  .").concat(Ql," {\n    right: ").concat(r,"px ").concat(o,";\n  }\n  \n  .").concat(Wl," {\n    margin-right: ").concat(r,"px ").concat(o,";\n  }\n  \n  .").concat(Ql," .").concat(Ql," {\n    right: 0 ").concat(o,";\n  }\n  \n  .").concat(Wl," .").concat(Wl," {\n    margin-right: 0 ").concat(o,";\n  }\n  \n  body[").concat(nu,"] {\n    ").concat(xm,": ").concat(r,"px;\n  }\n")},KC=function(){var e=parseInt(document.body.getAttribute(nu)||"0",10);return isFinite(e)?e:0},EE=function(){uu.useEffect(function(){return document.body.setAttribute(nu,(KC()+1).toString()),function(){var e=KC()-1;e<=0?document.body.removeAttribute(nu):document.body.setAttribute(nu,e.toString())}},[])},Tm=function(e){var t=e.noRelative,a=e.noImportant,o=e.gapMode,l=o===void 0?"margin":o;EE();var n=uu.useMemo(function(){return Am(l)},[l]);return uu.createElement(AE,{styles:TE(n,!t,l,a?"":"!important")})};var Em=!1;if(typeof window!="undefined")try{Ur=Object.defineProperty({},"passive",{get:function(){return Em=!0,!0}}),window.addEventListener("test",Ur,Ur),window.removeEventListener("test",Ur,Ur)}catch(e){Em=!1}var Ur,Jl=Em?{passive:!1}:!1;var DE=function(e){return e.tagName==="TEXTAREA"},ZC=function(e,t){if(!(e instanceof Element))return!1;var a=window.getComputedStyle(e);return a[t]!=="hidden"&&!(a.overflowY===a.overflowX&&!DE(e)&&a[t]==="visible")},ME=function(e){return ZC(e,"overflowY")},OE=function(e){return ZC(e,"overflowX")},Dm=function(e,t){var a=t.ownerDocument,o=t;do{typeof ShadowRoot!="undefined"&&o instanceof ShadowRoot&&(o=o.host);var l=QC(e,o);if(l){var n=WC(e,o),u=n[1],r=n[2];if(u>r)return!0}o=o.parentNode}while(o&&o!==a.body);return!1},PE=function(e){var t=e.scrollTop,a=e.scrollHeight,o=e.clientHeight;return[t,a,o]},BE=function(e){var t=e.scrollLeft,a=e.scrollWidth,o=e.clientWidth;return[t,a,o]},QC=function(e,t){return e==="v"?ME(t):OE(t)},WC=function(e,t){return e==="v"?PE(t):BE(t)},_E=function(e,t){return e==="h"&&t==="rtl"?-1:1},JC=function(e,t,a,o,l){var n=_E(e,window.getComputedStyle(t).direction),u=n*o,r=a.target,i=t.contains(r),s=!1,f=u>0,d=0,c=0;do{if(!r)break;var g=WC(e,r),L=g[0],x=g[1],S=g[2],h=x-S-n*L;(L||h)&&QC(e,r)&&(d+=h,c+=L);var p=r.parentNode;r=p&&p.nodeType===Node.DOCUMENT_FRAGMENT_NODE?p.host:p}while(!i&&r!==document.body||i&&(t.contains(r)||t===r));return(f&&(l&&Math.abs(d)<1||!l&&u>d)||!f&&(l&&Math.abs(c)<1||!l&&-u>c))&&(s=!0),s};var cf=function(e){return"changedTouches"in e?[e.changedTouches[0].clientX,e.changedTouches[0].clientY]:[0,0]},$C=function(e){return[e.deltaX,e.deltaY]},eR=function(e){return e&&"current"in e?e.current:e},UE=function(e,t){return e[0]===t[0]&&e[1]===t[1]},NE=function(e){return"\n  .block-interactivity-".concat(e," {pointer-events: none;}\n  .allow-interactivity-").concat(e," {pointer-events: all;}\n")},kE=0,ru=[];function tR(e){var t=be.useRef([]),a=be.useRef([0,0]),o=be.useRef(),l=be.useState(kE++)[0],n=be.useState(_r)[0],u=be.useRef(e);be.useEffect(function(){u.current=e},[e]),be.useEffect(function(){if(e.inert){document.body.classList.add("block-interactivity-".concat(l));var x=HC([e.lockRef.current],(e.shards||[]).map(eR),!0).filter(Boolean);return x.forEach(function(S){return S.classList.add("allow-interactivity-".concat(l))}),function(){document.body.classList.remove("block-interactivity-".concat(l)),x.forEach(function(S){return S.classList.remove("allow-interactivity-".concat(l))})}}},[e.inert,e.lockRef.current,e.shards]);var r=be.useCallback(function(x,S){if("touches"in x&&x.touches.length===2||x.type==="wheel"&&x.ctrlKey)return!u.current.allowPinchZoom;var h=cf(x),p=a.current,m="deltaX"in x?x.deltaX:p[0]-h[0],v="deltaY"in x?x.deltaY:p[1]-h[1],C,I=x.target,R=Math.abs(m)>Math.abs(v)?"h":"v";if("touches"in x&&R==="h"&&I.type==="range")return!1;var w=window.getSelection(),A=w&&w.anchorNode,M=A?A===I||A.contains(I):!1;if(M)return!1;var U=Dm(R,I);if(!U)return!0;if(U?C=R:(C=R==="v"?"h":"v",U=Dm(R,I)),!U)return!1;if(!o.current&&"changedTouches"in x&&(m||v)&&(o.current=C),!C)return!0;var N=o.current||C;return JC(N,S,x,N==="h"?m:v,!0)},[]),i=be.useCallback(function(x){var S=x;if(!(!ru.length||ru[ru.length-1]!==n)){var h="deltaY"in S?$C(S):cf(S),p=t.current.filter(function(C){return C.name===S.type&&(C.target===S.target||S.target===C.shadowParent)&&UE(C.delta,h)})[0];if(p&&p.should){S.cancelable&&S.preventDefault();return}if(!p){var m=(u.current.shards||[]).map(eR).filter(Boolean).filter(function(C){return C.contains(S.target)}),v=m.length>0?r(S,m[0]):!u.current.noIsolation;v&&S.cancelable&&S.preventDefault()}}},[]),s=be.useCallback(function(x,S,h,p){var m={name:x,delta:S,target:h,should:p,shadowParent:HE(h)};t.current.push(m),setTimeout(function(){t.current=t.current.filter(function(v){return v!==m})},1)},[]),f=be.useCallback(function(x){a.current=cf(x),o.current=void 0},[]),d=be.useCallback(function(x){s(x.type,$C(x),x.target,r(x,e.lockRef.current))},[]),c=be.useCallback(function(x){s(x.type,cf(x),x.target,r(x,e.lockRef.current))},[]);be.useEffect(function(){return ru.push(n),e.setCallbacks({onScrollCapture:d,onWheelCapture:d,onTouchMoveCapture:c}),document.addEventListener("wheel",i,Jl),document.addEventListener("touchmove",i,Jl),document.addEventListener("touchstart",f,Jl),function(){ru=ru.filter(function(x){return x!==n}),document.removeEventListener("wheel",i,Jl),document.removeEventListener("touchmove",i,Jl),document.removeEventListener("touchstart",f,Jl)}},[]);var g=e.removeScrollBar,L=e.inert;return be.createElement(be.Fragment,null,L?be.createElement(n,{styles:NE(l)}):null,g?be.createElement(Tm,{noRelative:e.noRelative,gapMode:e.gapMode}):null)}function HE(e){for(var t=null;e!==null;)e instanceof ShadowRoot&&(t=e.host,e=e.host),e=e.parentNode;return t}var aR=Cm(ff,tR);var oR=df.forwardRef(function(e,t){return df.createElement(Br,Vt({},e,{ref:t,sideCar:aR}))});oR.classNames=Br.classNames;var Nr=oR;var st=T(X(),1);var zE=Object.defineProperty,Eo=(e,t)=>zE(e,"name",{value:t,configurable:!0}),Mm="Popover",[lR,jP]=Et(Mm,[gm]),Om=gm(),[FE,iu]=lR(Mm),nR=Eo(e=>{let{__scopePopover:t,children:a,open:o,defaultOpen:l,onOpenChange:n,modal:u=!1}=e,r=Om(t),i=we.useRef(null),[s,f]=we.useState(!1),[d,c]=we.useState(0),[g,L]=we.useState(0),[x,S]=ql({prop:o,defaultProp:l!=null?l:!1,onChange:n,caller:Mm});return(0,st.jsx)(DC,b(y({},r),{children:(0,st.jsx)(FE,{scope:t,contentId:Fl(),titleId:Fl(),descriptionId:Fl(),titlePresent:d>0,descriptionPresent:g>0,setTitleCount:c,setDescriptionCount:L,triggerRef:i,open:x,onOpenChange:S,onOpenToggle:we.useCallback(()=>S(h=>!h),[S]),hasCustomAnchor:s,onCustomAnchorAdd:we.useCallback(()=>f(!0),[]),onCustomAnchorRemove:we.useCallback(()=>f(!1),[]),modal:u,children:a})}))},"Popover");var qE="PopoverTrigger",uR=we.forwardRef(Eo(function(t,a){let s=t,{__scopePopover:o}=s,l=D(s,["__scopePopover"]),n=iu(qE,o),u=Om(o),r=ce(a,n.triggerRef),i=(0,st.jsx)(Qe.button,b(y({type:"button","aria-haspopup":"dialog","aria-expanded":n.open,"aria-controls":n.open?n.contentId:void 0,"data-state":Pm(n.open)},l),{ref:r,onClick:Ee(t.onClick,n.onOpenToggle)}));return n.hasCustomAnchor?i:(0,st.jsx)(MC,b(y({asChild:!0},u),{children:i}))},"PopoverTrigger")),rR="PopoverPortal",[GE,VE]=lR(rR,{forceMount:void 0}),iR=Eo(e=>{let{__scopePopover:t,forceMount:a,children:o,container:l}=e,n=iu(rR,t);return(0,st.jsx)(GE,{scope:t,forceMount:a,children:(0,st.jsx)(qp,{present:a||n.open,children:(0,st.jsx)(NC,{asChild:!0,container:l,children:o})})})},"PopoverPortal"),kr="PopoverContent",sR=we.forwardRef(Eo(function(t,a){let o=VE(kr,t.__scopePopover),r=t,{forceMount:l=o.forceMount}=r,n=D(r,["forceMount"]),u=iu(kr,t.__scopePopover);return(0,st.jsx)(qp,{present:l||u.open,children:u.modal?(0,st.jsx)(jE,b(y({},n),{ref:a})):(0,st.jsx)(YE,b(y({},n),{ref:a}))})},"PopoverContent")),XE=Da("PopoverContent.RemoveScroll"),jE=we.forwardRef(Eo(function(t,a){let o=iu(kr,t.__scopePopover),l=we.useRef(null),n=ce(a,l),u=we.useRef(!1),{nodes:r,registry:i}=nm();return we.useEffect(()=>{let s=l.current;if(s)return nf(s)},[]),(0,st.jsx)(Nr,{as:XE,allowPinchZoom:!0,shards:we.useMemo(()=>[l,...r.map(s=>({current:s}))],[l,r]),children:(0,st.jsx)(fR,b(y({},t),{ref:n,branchNodes:r,branchRegistry:i,trapFocus:o.open,disableOutsidePointerEvents:!0,onCloseAutoFocus:Ee(t.onCloseAutoFocus,s=>{var f;s.preventDefault(),u.current||(f=o.triggerRef.current)==null||f.focus()}),onPointerDownOutside:Ee(t.onPointerDownOutside,s=>{let f=s.detail.originalEvent,d=f.button===0&&f.ctrlKey===!0,c=f.button===2||d;u.current=c},{checkForDefaultPrevented:!1}),onFocusOutside:Ee(t.onFocusOutside,s=>s.preventDefault(),{checkForDefaultPrevented:!1})}))})},"PopoverContentModal")),YE=we.forwardRef(Eo(function(t,a){let o=iu(kr,t.__scopePopover),l=we.useRef(!1),n=we.useRef(!1);return(0,st.jsx)(fR,b(y({},t),{ref:a,trapFocus:!1,disableOutsidePointerEvents:!1,onCloseAutoFocus:u=>{var r,i;(r=t.onCloseAutoFocus)==null||r.call(t,u),u.defaultPrevented||(l.current||(i=o.triggerRef.current)==null||i.focus(),u.preventDefault()),l.current=!1,n.current=!1},onInteractOutside:u=>{var s,f;(s=t.onInteractOutside)==null||s.call(t,u),u.defaultPrevented||(l.current=!0,u.detail.originalEvent.type==="pointerdown"&&(n.current=!0));let r=u.target;(f=o.triggerRef.current)!=null&&f.contains(r)&&u.preventDefault(),u.detail.originalEvent.type==="focusin"&&n.current&&u.preventDefault()}}))},"PopoverContentNonModal")),fR=we.forwardRef(Eo(function(t,a){let C=t,{__scopePopover:o,trapFocus:l,onOpenAutoFocus:n,onCloseAutoFocus:u,disableOutsidePointerEvents:r,onEscapeKeyDown:i,onPointerDownOutside:s,onFocusOutside:f,onInteractOutside:d,"aria-describedby":c,branchNodes:g,branchRegistry:L}=C,x=D(C,["__scopePopover","trapFocus","onOpenAutoFocus","onCloseAutoFocus","disableOutsidePointerEvents","onEscapeKeyDown","onPointerDownOutside","onFocusOutside","onInteractOutside","aria-describedby","branchNodes","branchRegistry"]),S=iu(kr,o),h=Om(o),[p,m]=we.useState(null),v=ce(a,m);return um(p),_s(),(0,st.jsx)(lm,{registry:L,children:(0,st.jsx)(Ey,{asChild:!0,loop:!0,trapped:l,branches:g,onMountAutoFocus:n,onUnmountAutoFocus:u,children:(0,st.jsx)(Ry,{asChild:!0,disableOutsidePointerEvents:r,onInteractOutside:d,onEscapeKeyDown:i,onPointerDownOutside:s,onFocusOutside:f,onDismiss:()=>S.onOpenChange(!1),deferPointerDownOutside:!0,children:(0,st.jsx)(PC,b(y(y({"data-state":Pm(S.open),role:"dialog",id:S.contentId,"aria-labelledby":S.titlePresent?S.titleId:void 0,"aria-describedby":S.descriptionPresent?cR(c,S.descriptionId):c},h),x),{ref:v,style:b(y({},x.style),{"--radix-popover-content-transform-origin":"var(--radix-popper-transform-origin)","--radix-popover-content-available-width":"var(--radix-popper-available-width)","--radix-popover-content-available-height":"var(--radix-popper-available-height)","--radix-popover-trigger-width":"var(--radix-popper-anchor-width)","--radix-popover-trigger-height":"var(--radix-popper-anchor-height)"})}))})})})},"PopoverContentImpl"));function Pm(e){return e?"open":"closed"}Eo(Pm,"getState");function cR(...e){let t=new Set;for(let a of e)if(typeof a=="string")for(let o of String(a).trim().split(/\s+/))o&&t.add(o);return t.size>0?Array.from(t).join(" "):void 0}Eo(cR,"concatAriaDescribedby");var pf=T(B());var QE=T(X()),ZE=(0,pf.createContext)(void 0);function mf(){var t;let e=(0,pf.useContext)(ZE);return e?{"data-ui-appearance":"","data-theme":e.theme,"data-motion":e.motion,"data-motion-style":(t=e.motionStyle)!=null?t:"standard",style:e.style}:void 0}var Bm=T(X()),pR=nR,mR=uR;var _m=dR.forwardRef((r,u)=>{var i=r,{className:e,sideOffset:t=6,align:a="start",collisionPadding:o=8,style:l}=i,n=D(i,["className","sideOffset","align","collisionPadding","style"]);let s=mf();return(0,Bm.jsx)(iR,{children:(0,Bm.jsx)(sR,b(y(b(y({},s),{ref:u,sideOffset:t,align:a,collisionPadding:o,className:Be("ui-popover-content",e)}),n),{style:y(y({},s==null?void 0:s.style),l)}))})});_m.displayName="PopoverContent";var Je=T(B());var ft=T(X());function Hr(e){return new Date(e.getFullYear(),e.getMonth(),e.getDate())}function Um(e){return new Date(e.getFullYear(),e.getMonth(),1)}function Do(e,t){return new Date(e.getFullYear(),e.getMonth(),e.getDate()+t)}function zr(e,t){let a=new Date(e.getFullYear(),e.getMonth()+t,1),o=new Date(a.getFullYear(),a.getMonth()+1,0).getDate();return new Date(a.getFullYear(),a.getMonth(),Math.min(e.getDate(),o))}function pl(e,t){return e.getFullYear()===t.getFullYear()&&e.getMonth()===t.getMonth()&&e.getDate()===t.getDate()}function ka(e){let t=a=>String(a).padStart(2,"0");return"".concat(e.getFullYear(),"-").concat(t(e.getMonth()+1),"-").concat(t(e.getDate()))}var WE={from:null,to:null};function hR(e,t){return e<=t?[e,t]:[t,e]}var gR=Je.forwardRef((v,m)=>{var C=v,{className:e,mode:t="single",value:a,defaultValue:o,onValueChange:l,rangeHint:n="\uC885\uB8CC\uC77C\uC744 \uC120\uD0DD\uD558\uC138\uC694",month:u,defaultMonth:r,onMonthChange:i,min:s,max:f,isDateDisabled:d,weekStartsOn:c=0,locale:g="ko-KR",disabled:L=!1,previousMonthLabel:x="\uC774\uC804 \uB2EC",nextMonthLabel:S="\uB2E4\uC74C \uB2EC","aria-label":h="\uB2EC\uB825"}=C,p=D(C,["className","mode","value","defaultValue","onValueChange","rangeHint","month","defaultMonth","onMonthChange","min","max","isDateDisabled","weekStartsOn","locale","disabled","previousMonthLabel","nextMonthLabel","aria-label"]);var xh,Lh;let I=Je.useRef(null);Je.useImperativeHandle(m,()=>I.current,[]);let R=Je.useId(),w=Je.useMemo(()=>Hr(new Date),[]),A=t==="range",[M,U]=Je.useState(o!=null?o:A?WE:null),N=a!==void 0?a:M,H=A?N:null,O={from:H!=null&&H.from?Hr(H.from):null,to:H!=null&&H.to?Hr(H.to):null},G=A?O.from:N,[k,j]=Je.useState(()=>{var Q;return Um((Q=r!=null?r:G)!=null?Q:w)}),P=u!==void 0?Um(u):k,[Z,J]=Je.useState(null),[ge,ue]=Je.useState(null),ct=Je.useRef(!1),la=A&&O.from!==null&&O.to===null,jt=la?ge!=null?ge:Z:null,[Ot,tt]=A&&O.from?hR(O.from,(Lh=(xh=O.to)!=null?xh:jt)!=null?Lh:O.from):[null,null],F=Ot!==null&&tt!==null&&!pl(Ot,tt),le=Je.useMemo(()=>({caption:new Intl.DateTimeFormat(g,{year:"numeric",month:"long"}),short:new Intl.DateTimeFormat(g,{weekday:"short"}),long:new Intl.DateTimeFormat(g,{weekday:"long"}),full:new Intl.DateTimeFormat(g,{dateStyle:"full"}),day:new Intl.NumberFormat(g)}),[g]),ve=s?Hr(s):void 0,$=f?Hr(f):void 0,ne=Q=>L||!!(ve&&Q<ve)||!!($&&Q>$)||!!(d!=null&&d(Q)),W=Q=>Q.getFullYear()===P.getFullYear()&&Q.getMonth()===P.getMonth(),at=Z&&W(Z)?Z:G&&W(G)?G:W(w)?w:P,Ye=(P.getDay()-c+7)%7,Pt=new Date(P.getFullYear(),P.getMonth()+1,0).getDate(),wa=Do(P,-Ye),za=Array.from({length:Math.ceil((Ye+Pt)/7)},(Q,z)=>Array.from({length:7},(_o,Fa)=>Do(wa,z*7+Fa))),xl=Array.from({length:7},(Q,z)=>new Date(2023,0,1+(c+z)%7));function oo(Q){let z=Um(Q);u===void 0&&j(z),i==null||i(z)}function hu(Q){let z=Q;ve&&z<ve&&(z=ve),$&&z>$&&(z=$),W(z)||oo(z),ct.current=!0,ue(null),J(z)}function Bo(Q){if(ne(Q))return;let z=Q;if(A)if(O.from&&!O.to){let[_o,Fa]=hR(O.from,Q);z={from:_o,to:Fa}}else z={from:Q,to:null};a===void 0&&U(z),W(Q)||oo(Q),J(Q),l==null||l(z)}function Pf(Q,z){let _o=getComputedStyle(Q.currentTarget).direction==="rtl",Fa=(z.getDay()-c+7)%7,na;switch(Q.key){case"ArrowLeft":na=Do(z,_o?1:-1);break;case"ArrowRight":na=Do(z,_o?-1:1);break;case"ArrowUp":na=Do(z,-7);break;case"ArrowDown":na=Do(z,7);break;case"Home":na=Do(z,-Fa);break;case"End":na=Do(z,6-Fa);break;case"PageUp":na=zr(z,Q.shiftKey?-12:-1);break;case"PageDown":na=zr(z,Q.shiftKey?12:1);break;default:return}Q.preventDefault(),hu(na)}Je.useLayoutEffect(()=>{var Q,z;!ct.current||!Z||(ct.current=!1,(z=(Q=I.current)==null?void 0:Q.querySelector('[data-date="'.concat(ka(Z),'"]')))==null||z.focus())});let Bf=L||!!(ve&&Do(P,-1)<ve),an=L||!!($&&zr(P,1)>$);return(0,ft.jsxs)("div",b(y({},p),{ref:I,role:"group","aria-label":h,className:Be("ui-calendar",e),"data-disabled":L?"":void 0,children:[(0,ft.jsxs)("div",{className:"ui-calendar-header",children:[(0,ft.jsx)("button",{type:"button",className:"ui-calendar-nav","aria-label":x,disabled:Bf,onClick:()=>oo(zr(P,-1)),children:(0,ft.jsx)(Ir,{size:16,"aria-hidden":"true"})}),(0,ft.jsx)("div",{id:R,className:"ui-calendar-caption","aria-live":"polite",children:le.caption.format(P)}),(0,ft.jsx)("button",{type:"button",className:"ui-calendar-nav","aria-label":S,disabled:an,onClick:()=>oo(zr(P,1)),children:(0,ft.jsx)(br,{size:16,"aria-hidden":"true"})})]}),(0,ft.jsxs)("table",{className:"ui-calendar-grid",role:"grid","aria-labelledby":R,"aria-multiselectable":A||void 0,onPointerLeave:()=>ue(null),children:[(0,ft.jsx)("thead",{children:(0,ft.jsx)("tr",{children:xl.map(Q=>(0,ft.jsx)("th",{scope:"col",abbr:le.long.format(Q),className:"ui-calendar-weekday",children:le.short.format(Q)},Q.getDay()))})}),(0,ft.jsx)("tbody",{children:za.map(Q=>(0,ft.jsx)("tr",{children:Q.map(z=>{let _o=A?!!(O.from&&pl(z,O.from)||O.to&&pl(z,O.to)):!!(G&&pl(z,G)),Fa=Ot!==null&&tt!==null&&z>=Ot&&z<=tt,na=pl(z,w),hb=ne(z);return(0,ft.jsx)("td",{role:"gridcell","aria-selected":A&&O.to?Fa:_o,className:"ui-calendar-cell","data-range-band":F&&Fa||void 0,"data-range-start":F&&Ot!==null&&pl(z,Ot)||void 0,"data-range-end":F&&tt!==null&&pl(z,tt)||void 0,"data-range-preview":F&&Fa&&la||void 0,children:(0,ft.jsx)("button",{type:"button",className:"ui-calendar-day","data-date":ka(z),"data-selected":_o||void 0,"data-today":na||void 0,"data-outside":!W(z)||void 0,"aria-label":le.full.format(z),"aria-current":na?"date":void 0,"aria-disabled":hb||void 0,tabIndex:pl(z,at)?0:-1,onClick:()=>Bo(z),onKeyDown:gb=>Pf(gb,z),onFocus:()=>J(z),onPointerEnter:la?()=>ue(z):void 0,children:le.day.format(z.getDate())})},ka(z))})},ka(Q[0])))})]}),A&&(0,ft.jsx)("div",{role:"status",className:"ui-calendar-status",children:la?n:""})]}))});gR.displayName="Calendar";var Nm=gR;var It=T(X());function JE(e,t){if(!e.from||!e.to)return;let a=new Intl.DateTimeFormat(t,{dateStyle:"long"});return typeof a.formatRange=="function"?a.formatRange(e.from,e.to):"".concat(a.format(e.from)," \u2013 ").concat(a.format(e.to))}var vR=su.forwardRef((m,p)=>{var v=m,{className:e,mode:t="single",value:a,defaultValue:o,onValueChange:l,open:n,defaultOpen:u=!1,onOpenChange:r,placeholder:i,calendarLabel:s,name:f,min:d,max:c,isDateDisabled:g,weekStartsOn:L,locale:x="ko-KR",disabled:S}=v,h=D(v,["className","mode","value","defaultValue","onValueChange","open","defaultOpen","onOpenChange","placeholder","calendarLabel","name","min","max","isDateDisabled","weekStartsOn","locale","disabled"]);let C=t==="range",[I,R]=su.useState(o!=null?o:null),[w,A]=su.useState(u),[M,U]=su.useState(null),N=a!==void 0?a:I,H=n!==void 0?n:w,O=l,G=C?null:N,k=C?N:null,j=C?k?JE(k,x):void 0:G?new Intl.DateTimeFormat(x,{dateStyle:"long"}).format(G):void 0,P=i!=null?i:C?"\uAE30\uAC04 \uC120\uD0DD":"\uB0A0\uC9DC \uC120\uD0DD",Z=s!=null?s:C?"\uAE30\uAC04 \uC120\uD0DD":"\uB0A0\uC9DC \uC120\uD0DD";function J(ue){ue||U(null),n===void 0&&A(ue),r==null||r(ue)}let ge={min:d,max:c,isDateDisabled:g,weekStartsOn:L,locale:x,"aria-label":Z};return(0,It.jsxs)(pR,{open:H,onOpenChange:J,children:[(0,It.jsx)(mR,{asChild:!0,children:(0,It.jsxs)("button",b(y({},h),{ref:p,type:"button",disabled:S,className:Be("ui-date-picker-trigger",e),"data-placeholder":j?void 0:"","data-mode":C?"range":void 0,children:[(0,It.jsx)("span",{children:j!=null?j:P}),(0,It.jsx)(Cr,{size:16,"aria-hidden":"true"})]}))}),(0,It.jsx)(_m,{className:"ui-date-picker-content","aria-label":Z,onOpenAutoFocus:ue=>{var ct;ue.preventDefault(),(ct=ue.currentTarget.querySelector('.ui-calendar-day[tabindex="0"]'))==null||ct.focus()},children:C?(0,It.jsx)(Nm,b(y({},ge),{mode:"range",value:M!=null?M:k,onValueChange:ue=>{if(!ue.from||!ue.to){U(ue);return}U(null),a===void 0&&R(ue),O==null||O({from:ue.from,to:ue.to}),J(!1)}})):(0,It.jsx)(Nm,b(y({},ge),{value:G,onValueChange:ue=>{a===void 0&&R(ue),O==null||O(ue),J(!1)}}))}),f&&(C?(0,It.jsxs)(It.Fragment,{children:[(0,It.jsx)("input",{type:"hidden",name:"".concat(f,"-start"),value:k!=null&&k.from?ka(k.from):"",disabled:S}),(0,It.jsx)("input",{type:"hidden",name:"".concat(f,"-end"),value:k!=null&&k.to?ka(k.to):"",disabled:S})]}):(0,It.jsx)("input",{type:"hidden",name:f,value:G?ka(G):"",disabled:S}))]})});vR.displayName="DatePicker";var SR=vR;var Af=T(B());var _=T(B(),1),uh=T(Va(),1);function km(e,[t,a]){return Math.min(a,Math.max(t,e))}var dB=!!(typeof window!="undefined"&&window.document&&window.document.createElement);function _e(e,t,{checkForDefaultPrevented:a=!0}={}){return function(l){if(e==null||e(l),a===!1||!l.defaultPrevented)return t==null?void 0:t(l)}}var ml=T(B(),1);var Mo=T(B(),1),xR=T(X(),1);function fu(e,t=[]){let a=[];function o(n,u){let r=Mo.createContext(u),i=a.length;a=[...a,u];let s=d=>{var p;let h=d,{scope:c,children:g}=h,L=D(h,["scope","children"]),x=((p=c==null?void 0:c[e])==null?void 0:p[i])||r,S=Mo.useMemo(()=>L,Object.values(L));return(0,xR.jsx)(x.Provider,{value:S,children:g})};s.displayName=n+"Provider";function f(d,c){var x;let g=((x=c==null?void 0:c[e])==null?void 0:x[i])||r,L=Mo.useContext(g);if(L)return L;if(u!==void 0)return u;throw new Error("\x60".concat(d,"\x60 must be used within \x60").concat(n,"\x60"))}return[s,f]}let l=()=>{let n=a.map(u=>Mo.createContext(u));return function(r){let i=(r==null?void 0:r[e])||n;return Mo.useMemo(()=>({["__scope".concat(e)]:b(y({},r),{[e]:i})}),[r,i])}};return l.scopeName=e,[o,$E(l,...t)]}function $E(...e){let t=e[0];if(e.length===1)return t;let a=()=>{let o=e.map(l=>({useScope:l(),scopeName:l.scopeName}));return function(n){let u=o.reduce((r,{useScope:i,scopeName:s})=>{let d=i(n)["__scope".concat(s)];return y(y({},r),d)},{});return Mo.useMemo(()=>({["__scope".concat(t.scopeName)]:u}),[u])}};return a.scopeName=t.scopeName,a}var yR=T(B(),1);function LR(e,t){if(typeof e=="function")return e(t);e!=null&&(e.current=t)}function Hm(...e){return t=>{let a=!1,o=e.map(l=>{let n=LR(l,t);return!a&&typeof n=="function"&&(a=!0),n});if(a)return()=>{for(let l=0;l<o.length;l++){let n=o[l];typeof n=="function"?n():LR(e[l],null)}}}}function Ue(...e){return yR.useCallback(Hm(...e),e)}var $e=T(B(),1);var hf=T(X(),1);function $l(e){let t=eD(e),a=$e.forwardRef((o,l)=>{let s=o,{children:n}=s,u=D(s,["children"]),r=$e.Children.toArray(n),i=r.find(aD);if(i){let f=i.props.children,d=r.map(c=>c===i?$e.Children.count(f)>1?$e.Children.only(null):$e.isValidElement(f)?f.props.children:null:c);return(0,hf.jsx)(t,b(y({},u),{ref:l,children:$e.isValidElement(f)?$e.cloneElement(f,void 0,d):null}))}return(0,hf.jsx)(t,b(y({},u),{ref:l,children:n}))});return a.displayName="".concat(e,".Slot"),a}function eD(e){let t=$e.forwardRef((a,o)=>{let u=a,{children:l}=u,n=D(u,["children"]);if($e.isValidElement(l)){let r=lD(l),i=oD(n,l.props);return l.type!==$e.Fragment&&(i.ref=o?Hm(o,r):r),$e.cloneElement(l,i)}return $e.Children.count(l)>1?$e.Children.only(null):null});return t.displayName="".concat(e,".SlotClone"),t}var tD=Symbol("radix.slottable");function aD(e){return $e.isValidElement(e)&&typeof e.type=="function"&&"__radixId"in e.type&&e.type.__radixId===tD}function oD(e,t){let a=y({},t);for(let o in t){let l=e[o],n=t[o];/^on[A-Z]/.test(o)?l&&n?a[o]=(...r)=>{let i=n(...r);return l(...r),i}:l&&(a[o]=l):o==="style"?a[o]=y(y({},l),n):o==="className"&&(a[o]=[l,n].filter(Boolean).join(" "))}return y(y({},e),a)}function lD(e){var o,l;let t=(o=Object.getOwnPropertyDescriptor(e.props,"ref"))==null?void 0:o.get,a=t&&"isReactWarning"in t&&t.isReactWarning;return a?e.ref:(t=(l=Object.getOwnPropertyDescriptor(e,"ref"))==null?void 0:l.get,a=t&&"isReactWarning"in t&&t.isReactWarning,a?e.props.ref:e.props.ref||e.ref)}var gf=T(X(),1),nD=T(B(),1);var uD=T(X(),1);function CR(e){let t=e+"CollectionProvider",[a,o]=fu(t),[l,n]=a(t,{collectionRef:{current:null},itemMap:new Map}),u=x=>{let{scope:S,children:h}=x,p=ml.default.useRef(null),m=ml.default.useRef(new Map).current;return(0,gf.jsx)(l,{scope:S,itemMap:m,collectionRef:p,children:h})};u.displayName=t;let r=e+"CollectionSlot",i=$l(r),s=ml.default.forwardRef((x,S)=>{let{scope:h,children:p}=x,m=n(r,h),v=Ue(S,m.collectionRef);return(0,gf.jsx)(i,{ref:v,children:p})});s.displayName=r;let f=e+"CollectionItemSlot",d="data-radix-collection-item",c=$l(f),g=ml.default.forwardRef((x,S)=>{let R=x,{scope:h,children:p}=R,m=D(R,["scope","children"]),v=ml.default.useRef(null),C=Ue(S,v),I=n(f,h);return ml.default.useEffect(()=>(I.itemMap.set(v,y({ref:v},m)),()=>{I.itemMap.delete(v)})),(0,gf.jsx)(c,{[d]:"",ref:C,children:p})});g.displayName=f;function L(x){let S=n(e+"CollectionConsumer",x);return ml.default.useCallback(()=>{let p=S.collectionRef.current;if(!p)return[];let m=Array.from(p.querySelectorAll("[".concat(d,"]")));return Array.from(S.itemMap.values()).sort((I,R)=>m.indexOf(I.ref.current)-m.indexOf(R.ref.current))},[S.collectionRef,S.itemMap])}return[{Provider:u,Slot:s,ItemSlot:g},L,o]}var vf=T(B(),1),rD=T(X(),1),iD=vf.createContext(void 0);function RR(e){let t=vf.useContext(iD);return e||t||"ltr"}var Me=T(B(),1);var IR=T(B(),1),bR=T(Va(),1);var wR=T(X(),1);var sD=["a","button","div","form","h2","h3","img","input","label","li","nav","ol","p","select","span","svg","ul"],Le=sD.reduce((e,t)=>{let a=$l("Primitive.".concat(t)),o=IR.forwardRef((l,n)=>{let s=l,{asChild:u}=s,r=D(s,["asChild"]),i=u?a:t;return typeof window!="undefined"&&(window[Symbol.for("radix-ui")]=!0),(0,wR.jsx)(i,b(y({},r),{ref:n}))});return o.displayName="Primitive.".concat(t),b(y({},e),{[t]:o})},{});function AR(e,t){e&&bR.flushSync(()=>e.dispatchEvent(t))}var cu=T(B(),1);function ao(e){let t=cu.useRef(e);return cu.useEffect(()=>{t.current=e}),cu.useMemo(()=>(...a)=>{var o;return(o=t.current)==null?void 0:o.call(t,...a)},[])}var ER=T(B(),1);var du=T(B(),1);function TR(e){let t=du.useRef(e);return du.useEffect(()=>{t.current=e}),du.useMemo(()=>(...a)=>{var o;return(o=t.current)==null?void 0:o.call(t,...a)},[])}function DR(e,t=globalThis==null?void 0:globalThis.document){let a=TR(e);ER.useEffect(()=>{let o=l=>{l.key==="Escape"&&a(l)};return t.addEventListener("keydown",o,{capture:!0}),()=>t.removeEventListener("keydown",o,{capture:!0})},[a,t])}var Fm=T(X(),1);var fD="DismissableLayer",zm="dismissableLayer.update",cD="dismissableLayer.pointerDownOutside",dD="dismissableLayer.focusOutside",MR,PR=Me.createContext({layers:new Set,layersWithOutsidePointerEventsDisabled:new Set,branches:new Set}),qm=Me.forwardRef((e,t)=>{var w;let R=e,{disableOutsidePointerEvents:a=!1,onEscapeKeyDown:o,onPointerDownOutside:l,onFocusOutside:n,onInteractOutside:u,onDismiss:r}=R,i=D(R,["disableOutsidePointerEvents","onEscapeKeyDown","onPointerDownOutside","onFocusOutside","onInteractOutside","onDismiss"]),s=Me.useContext(PR),[f,d]=Me.useState(null),c=(w=f==null?void 0:f.ownerDocument)!=null?w:globalThis==null?void 0:globalThis.document,[,g]=Me.useState({}),L=Ue(t,A=>d(A)),x=Array.from(s.layers),[S]=[...s.layersWithOutsidePointerEventsDisabled].slice(-1),h=x.indexOf(S),p=f?x.indexOf(f):-1,m=s.layersWithOutsidePointerEventsDisabled.size>0,v=p>=h,C=hD(A=>{let M=A.target,U=[...s.branches].some(N=>N.contains(M));!v||U||(l==null||l(A),u==null||u(A),A.defaultPrevented||r==null||r())},c),I=gD(A=>{let M=A.target;[...s.branches].some(N=>N.contains(M))||(n==null||n(A),u==null||u(A),A.defaultPrevented||r==null||r())},c);return DR(A=>{p===s.layers.size-1&&(o==null||o(A),!A.defaultPrevented&&r&&(A.preventDefault(),r()))},c),Me.useEffect(()=>{if(f)return a&&(s.layersWithOutsidePointerEventsDisabled.size===0&&(MR=c.body.style.pointerEvents,c.body.style.pointerEvents="none"),s.layersWithOutsidePointerEventsDisabled.add(f)),s.layers.add(f),OR(),()=>{a&&s.layersWithOutsidePointerEventsDisabled.size===1&&(c.body.style.pointerEvents=MR)}},[f,c,a,s]),Me.useEffect(()=>()=>{f&&(s.layers.delete(f),s.layersWithOutsidePointerEventsDisabled.delete(f),OR())},[f,s]),Me.useEffect(()=>{let A=()=>g({});return document.addEventListener(zm,A),()=>document.removeEventListener(zm,A)},[]),(0,Fm.jsx)(Le.div,b(y({},i),{ref:L,style:y({pointerEvents:m?v?"auto":"none":void 0},e.style),onFocusCapture:_e(e.onFocusCapture,I.onFocusCapture),onBlurCapture:_e(e.onBlurCapture,I.onBlurCapture),onPointerDownCapture:_e(e.onPointerDownCapture,C.onPointerDownCapture)}))});qm.displayName=fD;var pD="DismissableLayerBranch",mD=Me.forwardRef((e,t)=>{let a=Me.useContext(PR),o=Me.useRef(null),l=Ue(t,o);return Me.useEffect(()=>{let n=o.current;if(n)return a.branches.add(n),()=>{a.branches.delete(n)}},[a.branches]),(0,Fm.jsx)(Le.div,b(y({},e),{ref:l}))});mD.displayName=pD;function hD(e,t=globalThis==null?void 0:globalThis.document){let a=ao(e),o=Me.useRef(!1),l=Me.useRef(()=>{});return Me.useEffect(()=>{let n=r=>{if(r.target&&!o.current){let s=function(){BR(cD,a,f,{discrete:!0})};var i=s;let f={originalEvent:r};r.pointerType==="touch"?(t.removeEventListener("click",l.current),l.current=s,t.addEventListener("click",l.current,{once:!0})):s()}else t.removeEventListener("click",l.current);o.current=!1},u=window.setTimeout(()=>{t.addEventListener("pointerdown",n)},0);return()=>{window.clearTimeout(u),t.removeEventListener("pointerdown",n),t.removeEventListener("click",l.current)}},[t,a]),{onPointerDownCapture:()=>o.current=!0}}function gD(e,t=globalThis==null?void 0:globalThis.document){let a=ao(e),o=Me.useRef(!1);return Me.useEffect(()=>{let l=n=>{n.target&&!o.current&&BR(dD,a,{originalEvent:n},{discrete:!1})};return t.addEventListener("focusin",l),()=>t.removeEventListener("focusin",l)},[t,a]),{onFocusCapture:()=>o.current=!0,onBlurCapture:()=>o.current=!1}}function OR(){let e=new CustomEvent(zm);document.dispatchEvent(e)}function BR(e,t,a,{discrete:o}){let l=a.originalEvent.target,n=new CustomEvent(e,{bubbles:!1,cancelable:!0,detail:a});t&&l.addEventListener(e,t,{once:!0}),o?AR(l,n):l.dispatchEvent(n)}var UR=T(B(),1),Gm=0;function NR(){UR.useEffect(()=>{var t,a;let e=document.querySelectorAll("[data-radix-focus-guard]");return document.body.insertAdjacentElement("afterbegin",(t=e[0])!=null?t:_R()),document.body.insertAdjacentElement("beforeend",(a=e[1])!=null?a:_R()),Gm++,()=>{Gm===1&&document.querySelectorAll("[data-radix-focus-guard]").forEach(o=>o.remove()),Gm--}},[])}function _R(){let e=document.createElement("span");return e.setAttribute("data-radix-focus-guard",""),e.tabIndex=0,e.style.outline="none",e.style.opacity="0",e.style.position="fixed",e.style.pointerEvents="none",e}var Ia=T(B(),1);var qR=T(X(),1);var Vm="focusScope.autoFocusOnMount",Xm="focusScope.autoFocusOnUnmount",kR={bubbles:!1,cancelable:!0},vD="FocusScope",jm=Ia.forwardRef((e,t)=>{let x=e,{loop:a=!1,trapped:o=!1,onMountAutoFocus:l,onUnmountAutoFocus:n}=x,u=D(x,["loop","trapped","onMountAutoFocus","onUnmountAutoFocus"]),[r,i]=Ia.useState(null),s=ao(l),f=ao(n),d=Ia.useRef(null),c=Ue(t,S=>i(S)),g=Ia.useRef({paused:!1,pause(){this.paused=!0},resume(){this.paused=!1}}).current;Ia.useEffect(()=>{if(o){let m=function(R){if(g.paused||!r)return;let w=R.target;r.contains(w)?d.current=w:hl(d.current,{select:!0})},v=function(R){if(g.paused||!r)return;let w=R.relatedTarget;w!==null&&(r.contains(w)||hl(d.current,{select:!0}))},C=function(R){if(document.activeElement===document.body)for(let A of R)A.removedNodes.length>0&&hl(r)};var S=m,h=v,p=C;document.addEventListener("focusin",m),document.addEventListener("focusout",v);let I=new MutationObserver(C);return r&&I.observe(r,{childList:!0,subtree:!0}),()=>{document.removeEventListener("focusin",m),document.removeEventListener("focusout",v),I.disconnect()}}},[o,r,g.paused]),Ia.useEffect(()=>{if(r){zR.add(g);let S=document.activeElement;if(!r.contains(S)){let p=new CustomEvent(Vm,kR);r.addEventListener(Vm,s),r.dispatchEvent(p),p.defaultPrevented||(SD(RD(GR(r)),{select:!0}),document.activeElement===S&&hl(r))}return()=>{r.removeEventListener(Vm,s),setTimeout(()=>{let p=new CustomEvent(Xm,kR);r.addEventListener(Xm,f),r.dispatchEvent(p),p.defaultPrevented||hl(S!=null?S:document.body,{select:!0}),r.removeEventListener(Xm,f),zR.remove(g)},0)}}},[r,s,f,g]);let L=Ia.useCallback(S=>{if(!a&&!o||g.paused)return;let h=S.key==="Tab"&&!S.altKey&&!S.ctrlKey&&!S.metaKey,p=document.activeElement;if(h&&p){let m=S.currentTarget,[v,C]=xD(m);v&&C?!S.shiftKey&&p===C?(S.preventDefault(),a&&hl(v,{select:!0})):S.shiftKey&&p===v&&(S.preventDefault(),a&&hl(C,{select:!0})):p===m&&S.preventDefault()}},[a,o,g.paused]);return(0,qR.jsx)(Le.div,b(y({tabIndex:-1},u),{ref:c,onKeyDown:L}))});jm.displayName=vD;function SD(e,{select:t=!1}={}){let a=document.activeElement;for(let o of e)if(hl(o,{select:t}),document.activeElement!==a)return}function xD(e){let t=GR(e),a=HR(t,e),o=HR(t.reverse(),e);return[a,o]}function GR(e){let t=[],a=document.createTreeWalker(e,NodeFilter.SHOW_ELEMENT,{acceptNode:o=>{let l=o.tagName==="INPUT"&&o.type==="hidden";return o.disabled||o.hidden||l?NodeFilter.FILTER_SKIP:o.tabIndex>=0?NodeFilter.FILTER_ACCEPT:NodeFilter.FILTER_SKIP}});for(;a.nextNode();)t.push(a.currentNode);return t}function HR(e,t){for(let a of e)if(!LD(a,{upTo:t}))return a}function LD(e,{upTo:t}){if(getComputedStyle(e).visibility==="hidden")return!0;for(;e;){if(t!==void 0&&e===t)return!1;if(getComputedStyle(e).display==="none")return!0;e=e.parentElement}return!1}function yD(e){return e instanceof HTMLInputElement&&"select"in e}function hl(e,{select:t=!1}={}){if(e&&e.focus){let a=document.activeElement;e.focus({preventScroll:!0}),e!==a&&yD(e)&&t&&e.select()}}var zR=CD();function CD(){let e=[];return{add(t){let a=e[0];t!==a&&(a==null||a.pause()),e=FR(e,t),e.unshift(t)},remove(t){var a;e=FR(e,t),(a=e[0])==null||a.resume()}}}function FR(e,t){let a=[...e],o=a.indexOf(t);return o!==-1&&a.splice(o,1),a}function RD(e){return e.filter(t=>t.tagName!=="A")}var Ym=T(B(),1);var VR=T(B(),1),je=globalThis!=null&&globalThis.document?VR.useLayoutEffect:()=>{};var ID=Ym[" useId ".trim().toString()]||(()=>{}),bD=0;function Sf(e){let[t,a]=Ym.useState(ID());return je(()=>{e||a(o=>o!=null?o:String(bD++))},[e]),e||(t?"radix-".concat(t):"")}var Xt=T(B(),1);var XR=T(B(),1);var Km=T(X(),1);var wD="Arrow",jR=XR.forwardRef((e,t)=>{let u=e,{children:a,width:o=10,height:l=5}=u,n=D(u,["children","width","height"]);return(0,Km.jsx)(Le.svg,b(y({},n),{ref:t,width:o,height:l,viewBox:"0 0 30 10",preserveAspectRatio:"none",children:e.asChild?a:(0,Km.jsx)("polygon",{points:"0,0 30,0 15,10"})}))});jR.displayName=wD;var YR=jR;var KR=T(B(),1);function ZR(e){let[t,a]=KR.useState(void 0);return je(()=>{if(e){a({width:e.offsetWidth,height:e.offsetHeight});let o=new ResizeObserver(l=>{if(!Array.isArray(l)||!l.length)return;let n=l[0],u,r;if("borderBoxSize"in n){let i=n.borderBoxSize,s=Array.isArray(i)?i[0]:i;u=s.inlineSize,r=s.blockSize}else u=e.offsetWidth,r=e.offsetHeight;a({width:u,height:r})});return o.observe(e,{box:"border-box"}),()=>o.unobserve(e)}else a(void 0)},[e]),t}var gl=T(X(),1);var Zm="Popper",[QR,Qm]=fu(Zm),[TD,WR]=QR(Zm),JR=e=>{let{__scopePopper:t,children:a}=e,[o,l]=Xt.useState(null);return(0,gl.jsx)(TD,{scope:t,anchor:o,onAnchorChange:l,children:a})};JR.displayName=Zm;var $R="PopperAnchor",eI=Xt.forwardRef((e,t)=>{let s=e,{__scopePopper:a,virtualRef:o}=s,l=D(s,["__scopePopper","virtualRef"]),n=WR($R,a),u=Xt.useRef(null),r=Ue(t,u),i=Xt.useRef(null);return Xt.useEffect(()=>{let f=i.current;i.current=(o==null?void 0:o.current)||u.current,f!==i.current&&n.onAnchorChange(i.current)}),o?null:(0,gl.jsx)(Le.div,b(y({},l),{ref:r}))});eI.displayName=$R;var Wm="PopperContent",[ED,DD]=QR(Wm),tI=Xt.forwardRef((e,t)=>{var F,le,ve,$,ne,W,at,Ye;let tt=e,{__scopePopper:a,side:o="bottom",sideOffset:l=0,align:n="center",alignOffset:u=0,arrowPadding:r=0,avoidCollisions:i=!0,collisionBoundary:s=[],collisionPadding:f=0,sticky:d="partial",hideWhenDetached:c=!1,updatePositionStrategy:g="optimized",onPlaced:L}=tt,x=D(tt,["__scopePopper","side","sideOffset","align","alignOffset","arrowPadding","avoidCollisions","collisionBoundary","collisionPadding","sticky","hideWhenDetached","updatePositionStrategy","onPlaced"]),S=WR(Wm,a),[h,p]=Xt.useState(null),m=Ue(t,Pt=>p(Pt)),[v,C]=Xt.useState(null),I=ZR(v),R=(F=I==null?void 0:I.width)!=null?F:0,w=(le=I==null?void 0:I.height)!=null?le:0,A=o+(n!=="center"?"-"+n:""),M=typeof f=="number"?f:y({top:0,right:0,bottom:0,left:0},f),U=Array.isArray(s)?s:[s],N=U.length>0,H={padding:M,boundary:U.filter(OD),altBoundary:N},{refs:O,floatingStyles:G,placement:k,isPositioned:j,middlewareData:P}=Ys({strategy:"fixed",placement:A,whileElementsMounted:(...Pt)=>Pr(...Pt,{animationFrame:g==="always"}),elements:{reference:S.anchor},middleware:[Ks({mainAxis:l+w,alignmentAxis:u}),i&&Zs(y({mainAxis:!0,crossAxis:!1,limiter:d==="partial"?Qs():void 0},H)),i&&Ws(y({},H)),Js(b(y({},H),{apply:({elements:Pt,rects:wa,availableWidth:za,availableHeight:xl})=>{let{width:oo,height:hu}=wa.reference,Bo=Pt.floating.style;Bo.setProperty("--radix-popper-available-width","".concat(za,"px")),Bo.setProperty("--radix-popper-available-height","".concat(xl,"px")),Bo.setProperty("--radix-popper-anchor-width","".concat(oo,"px")),Bo.setProperty("--radix-popper-anchor-height","".concat(hu,"px"))}})),v&&ef({element:v,padding:r}),PD({arrowWidth:R,arrowHeight:w}),c&&$s(y({strategy:"referenceHidden"},H))]}),[Z,J]=lI(k),ge=ao(L);je(()=>{j&&(ge==null||ge())},[j,ge]);let ue=(ve=P.arrow)==null?void 0:ve.x,ct=($=P.arrow)==null?void 0:$.y,la=((ne=P.arrow)==null?void 0:ne.centerOffset)!==0,[jt,Ot]=Xt.useState();return je(()=>{h&&Ot(window.getComputedStyle(h).zIndex)},[h]),(0,gl.jsx)("div",{ref:O.setFloating,"data-radix-popper-content-wrapper":"",style:y(b(y({},G),{transform:j?G.transform:"translate(0, -200%)",minWidth:"max-content",zIndex:jt,"--radix-popper-transform-origin":[(W=P.transformOrigin)==null?void 0:W.x,(at=P.transformOrigin)==null?void 0:at.y].join(" ")}),((Ye=P.hide)==null?void 0:Ye.referenceHidden)&&{visibility:"hidden",pointerEvents:"none"}),dir:e.dir,children:(0,gl.jsx)(ED,{scope:a,placedSide:Z,onArrowChange:C,arrowX:ue,arrowY:ct,shouldHideArrow:la,children:(0,gl.jsx)(Le.div,b(y({"data-side":Z,"data-align":J},x),{ref:m,style:b(y({},x.style),{animation:j?void 0:"none"})}))})})});tI.displayName=Wm;var aI="PopperArrow",MD={top:"bottom",right:"left",bottom:"top",left:"right"},oI=Xt.forwardRef(function(t,a){let r=t,{__scopePopper:o}=r,l=D(r,["__scopePopper"]),n=DD(aI,o),u=MD[n.placedSide];return(0,gl.jsx)("span",{ref:n.onArrowChange,style:{position:"absolute",left:n.arrowX,top:n.arrowY,[u]:0,transformOrigin:{top:"",right:"0 0",bottom:"center 0",left:"100% 0"}[n.placedSide],transform:{top:"translateY(100%)",right:"translateY(50%) rotate(90deg) translateX(-50%)",bottom:"rotate(180deg)",left:"translateY(50%) rotate(-90deg) translateX(50%)"}[n.placedSide],visibility:n.shouldHideArrow?"hidden":void 0},children:(0,gl.jsx)(YR,b(y({},l),{ref:a,style:b(y({},l.style),{display:"block"})}))})});oI.displayName=aI;function OD(e){return e!==null}var PD=e=>({name:"transformOrigin",options:e,fn(t){var S,h,p,m,v;let{placement:a,rects:o,middlewareData:l}=t,u=((S=l.arrow)==null?void 0:S.centerOffset)!==0,r=u?0:e.arrowWidth,i=u?0:e.arrowHeight,[s,f]=lI(a),d={start:"0%",center:"50%",end:"100%"}[f],c=((p=(h=l.arrow)==null?void 0:h.x)!=null?p:0)+r/2,g=((v=(m=l.arrow)==null?void 0:m.y)!=null?v:0)+i/2,L="",x="";return s==="bottom"?(L=u?d:"".concat(c,"px"),x="".concat(-i,"px")):s==="top"?(L=u?d:"".concat(c,"px"),x="".concat(o.floating.height+i,"px")):s==="right"?(L="".concat(-i,"px"),x=u?d:"".concat(g,"px")):s==="left"&&(L="".concat(o.floating.width+i,"px"),x=u?d:"".concat(g,"px")),{data:{x:L,y:x}}}});function lI(e){let[t,a="center"]=e.split("-");return[t,a]}var nI=JR,uI=eI,rI=tI,iI=oI;var xf=T(B(),1),sI=T(Va(),1);var fI=T(X(),1);var _D="Portal",Jm=xf.forwardRef((e,t)=>{var i;let r=e,{container:a}=r,o=D(r,["container"]),[l,n]=xf.useState(!1);je(()=>n(!0),[]);let u=a||l&&((i=globalThis==null?void 0:globalThis.document)==null?void 0:i.body);return u?sI.default.createPortal((0,fI.jsx)(Le.div,b(y({},o),{ref:t})),u):null});Jm.displayName=_D;var ba=T(B(),1);var Lf=T(B(),1);var UD=ba[" useInsertionEffect ".trim().toString()]||je;function $m({prop:e,defaultProp:t,onChange:a=()=>{},caller:o}){let[l,n,u]=ND({defaultProp:t,onChange:a}),r=e!==void 0,i=r?e:l;{let f=ba.useRef(e!==void 0);ba.useEffect(()=>{let d=f.current;if(d!==r){let c=d?"controlled":"uncontrolled",g=r?"controlled":"uncontrolled";console.warn("".concat(o," is changing from ").concat(c," to ").concat(g,". Components should not switch from controlled to uncontrolled (or vice versa). Decide between using a controlled or uncontrolled value for the lifetime of the component."))}f.current=r},[r,o])}let s=ba.useCallback(f=>{var d;if(r){let c=kD(f)?f(e):f;c!==e&&((d=u.current)==null||d.call(u,c))}else n(f)},[r,e,n,u]);return[i,s]}function ND({defaultProp:e,onChange:t}){let[a,o]=ba.useState(e),l=ba.useRef(a),n=ba.useRef(t);return UD(()=>{n.current=t},[t]),ba.useEffect(()=>{var u;l.current!==a&&((u=n.current)==null||u.call(n,a),l.current=a)},[a,l]),[a,o,n]}function kD(e){return typeof e=="function"}var yf=T(B(),1);function cI(e){let t=yf.useRef({value:e,previous:e});return yf.useMemo(()=>(t.current.value!==e&&(t.current.previous=t.current.value,t.current.value=e),t.current.previous),[e])}var SI=T(B(),1);var hI=T(B(),1),XD=T(Va(),1);var et=T(B(),1);var HD=T(B(),1);function dI(e,t){if(typeof e=="function")return e(t);e!=null&&(e.current=t)}function pI(...e){return t=>{let a=!1,o=e.map(l=>{let n=dI(l,t);return!a&&typeof n=="function"&&(a=!0),n});if(a)return()=>{for(let l=0;l<o.length;l++){let n=o[l];typeof n=="function"?n():dI(e[l],null)}}}}var Cf=T(X(),1);function mI(e){let t=zD(e),a=et.forwardRef((o,l)=>{let s=o,{children:n}=s,u=D(s,["children"]),r=et.Children.toArray(n),i=r.find(qD);if(i){let f=i.props.children,d=r.map(c=>c===i?et.Children.count(f)>1?et.Children.only(null):et.isValidElement(f)?f.props.children:null:c);return(0,Cf.jsx)(t,b(y({},u),{ref:l,children:et.isValidElement(f)?et.cloneElement(f,void 0,d):null}))}return(0,Cf.jsx)(t,b(y({},u),{ref:l,children:n}))});return a.displayName="".concat(e,".Slot"),a}function zD(e){let t=et.forwardRef((a,o)=>{let u=a,{children:l}=u,n=D(u,["children"]);if(et.isValidElement(l)){let r=VD(l),i=GD(n,l.props);return l.type!==et.Fragment&&(i.ref=o?pI(o,r):r),et.cloneElement(l,i)}return et.Children.count(l)>1?et.Children.only(null):null});return t.displayName="".concat(e,".SlotClone"),t}var FD=Symbol("radix.slottable");function qD(e){return et.isValidElement(e)&&typeof e.type=="function"&&"__radixId"in e.type&&e.type.__radixId===FD}function GD(e,t){let a=y({},t);for(let o in t){let l=e[o],n=t[o];/^on[A-Z]/.test(o)?l&&n?a[o]=(...r)=>{let i=n(...r);return l(...r),i}:l&&(a[o]=l):o==="style"?a[o]=y(y({},l),n):o==="className"&&(a[o]=[l,n].filter(Boolean).join(" "))}return y(y({},e),a)}function VD(e){var o,l;let t=(o=Object.getOwnPropertyDescriptor(e.props,"ref"))==null?void 0:o.get,a=t&&"isReactWarning"in t&&t.isReactWarning;return a?e.ref:(t=(l=Object.getOwnPropertyDescriptor(e,"ref"))==null?void 0:l.get,a=t&&"isReactWarning"in t&&t.isReactWarning,a?e.props.ref:e.props.ref||e.ref)}var gI=T(X(),1);var jD=["a","button","div","form","h2","h3","img","input","label","li","nav","ol","p","select","span","svg","ul"],vI=jD.reduce((e,t)=>{let a=mI("Primitive.".concat(t)),o=hI.forwardRef((l,n)=>{let s=l,{asChild:u}=s,r=D(s,["asChild"]),i=u?a:t;return typeof window!="undefined"&&(window[Symbol.for("radix-ui")]=!0),(0,gI.jsx)(i,b(y({},r),{ref:n}))});return o.displayName="Primitive.".concat(t),b(y({},e),{[t]:o})},{});var xI=T(X(),1);var eh=Object.freeze({position:"absolute",border:0,width:1,height:1,padding:0,margin:-1,overflow:"hidden",clip:"rect(0, 0, 0, 0)",whiteSpace:"nowrap",wordWrap:"normal"}),YD="VisuallyHidden",KD=SI.forwardRef((e,t)=>(0,xI.jsx)(vI.span,b(y({},e),{ref:t,style:y(y({},eh),e.style)})));KD.displayName=YD;var q=T(X(),1);var ZD=[" ","Enter","ArrowUp","ArrowDown"],QD=[" ","Enter"],en="Select",[If,bf,WD]=CR(en),[pu,j_]=fu(en,[WD,Qm]),wf=Qm(),[JD,vl]=pu(en),[$D,e1]=pu(en),LI=e=>{let{__scopeSelect:t,children:a,open:o,defaultOpen:l,onOpenChange:n,value:u,defaultValue:r,onValueChange:i,dir:s,name:f,autoComplete:d,disabled:c,required:g,form:L}=e,x=wf(t),[S,h]=_.useState(null),[p,m]=_.useState(null),[v,C]=_.useState(!1),I=RR(s),[R,w]=$m({prop:o,defaultProp:l!=null?l:!1,onChange:n,caller:en}),[A,M]=$m({prop:u,defaultProp:r,onChange:i,caller:en}),U=_.useRef(null),N=S?L||!!S.closest("form"):!0,[H,O]=_.useState(new Set),G=Array.from(H).map(k=>k.props.value).join(";");return(0,q.jsx)(nI,b(y({},x),{children:(0,q.jsxs)(JD,{required:g,scope:t,trigger:S,onTriggerChange:h,valueNode:p,onValueNodeChange:m,valueNodeHasChildren:v,onValueNodeHasChildrenChange:C,contentId:Sf(),value:A,onValueChange:M,open:R,onOpenChange:w,dir:I,triggerPointerDownPosRef:U,disabled:c,children:[(0,q.jsx)(If.Provider,{scope:t,children:(0,q.jsx)($D,{scope:e.__scopeSelect,onNativeOptionAdd:_.useCallback(k=>{O(j=>new Set(j).add(k))},[]),onNativeOptionRemove:_.useCallback(k=>{O(j=>{let P=new Set(j);return P.delete(k),P})},[]),children:a})}),N?(0,q.jsxs)(qI,{"aria-hidden":!0,required:g,tabIndex:-1,name:f,autoComplete:d,value:A,onChange:k=>M(k.target.value),disabled:c,form:L,children:[A===void 0?(0,q.jsx)("option",{value:""}):null,Array.from(H)]},G):null]})}))};LI.displayName=en;var yI="SelectTrigger",CI=_.forwardRef((e,t)=>{let x=e,{__scopeSelect:a,disabled:o=!1}=x,l=D(x,["__scopeSelect","disabled"]),n=wf(a),u=vl(yI,a),r=u.disabled||o,i=Ue(t,u.onTriggerChange),s=bf(a),f=_.useRef("touch"),[d,c,g]=VI(S=>{let h=s().filter(v=>!v.disabled),p=h.find(v=>v.value===u.value),m=XI(h,S,p);m!==void 0&&u.onValueChange(m.value)}),L=S=>{r||(u.onOpenChange(!0),g()),S&&(u.triggerPointerDownPosRef.current={x:Math.round(S.pageX),y:Math.round(S.pageY)})};return(0,q.jsx)(uI,b(y({asChild:!0},n),{children:(0,q.jsx)(Le.button,b(y({type:"button",role:"combobox","aria-controls":u.contentId,"aria-expanded":u.open,"aria-required":u.required,"aria-autocomplete":"none",dir:u.dir,"data-state":u.open?"open":"closed",disabled:r,"data-disabled":r?"":void 0,"data-placeholder":GI(u.value)?"":void 0},l),{ref:i,onClick:_e(l.onClick,S=>{S.currentTarget.focus(),f.current!=="mouse"&&L(S)}),onPointerDown:_e(l.onPointerDown,S=>{f.current=S.pointerType;let h=S.target;h.hasPointerCapture(S.pointerId)&&h.releasePointerCapture(S.pointerId),S.button===0&&S.ctrlKey===!1&&S.pointerType==="mouse"&&(L(S),S.preventDefault())}),onKeyDown:_e(l.onKeyDown,S=>{let h=d.current!=="";!(S.ctrlKey||S.altKey||S.metaKey)&&S.key.length===1&&c(S.key),!(h&&S.key===" ")&&ZD.includes(S.key)&&(L(),S.preventDefault())})}))}))});CI.displayName=yI;var RI="SelectValue",II=_.forwardRef((e,t)=>{let c=e,{__scopeSelect:a,className:o,style:l,children:n,placeholder:u=""}=c,r=D(c,["__scopeSelect","className","style","children","placeholder"]),i=vl(RI,a),{onValueNodeHasChildrenChange:s}=i,f=n!==void 0,d=Ue(t,i.onValueNodeChange);return je(()=>{s(f)},[s,f]),(0,q.jsx)(Le.span,b(y({},r),{ref:d,style:{pointerEvents:"none"},children:GI(i.value)?(0,q.jsx)(q.Fragment,{children:u}):n}))});II.displayName=RI;var t1="SelectIcon",bI=_.forwardRef((e,t)=>{let n=e,{__scopeSelect:a,children:o}=n,l=D(n,["__scopeSelect","children"]);return(0,q.jsx)(Le.span,b(y({"aria-hidden":!0},l),{ref:t,children:o||"\u25BC"}))});bI.displayName=t1;var a1="SelectPortal",wI=e=>(0,q.jsx)(Jm,y({asChild:!0},e));wI.displayName=a1;var tn="SelectContent",AI=_.forwardRef((e,t)=>{let a=vl(tn,e.__scopeSelect),[o,l]=_.useState();if(je(()=>{l(new DocumentFragment)},[]),!a.open){let n=o;return n?uh.createPortal((0,q.jsx)(TI,{scope:e.__scopeSelect,children:(0,q.jsx)(If.Slot,{scope:e.__scopeSelect,children:(0,q.jsx)("div",{children:e.children})})}),n):null}return(0,q.jsx)(EI,b(y({},e),{ref:t}))});AI.displayName=tn;var Ha=10,[TI,Sl]=pu(tn),o1="SelectContentImpl",l1=$l("SelectContent.RemoveScroll"),EI=_.forwardRef((e,t)=>{let tt=e,{__scopeSelect:a,position:o="item-aligned",onCloseAutoFocus:l,onEscapeKeyDown:n,onPointerDownOutside:u,side:r,sideOffset:i,align:s,alignOffset:f,arrowPadding:d,collisionBoundary:c,collisionPadding:g,sticky:L,hideWhenDetached:x,avoidCollisions:S}=tt,h=D(tt,["__scopeSelect","position","onCloseAutoFocus","onEscapeKeyDown","onPointerDownOutside","side","sideOffset","align","alignOffset","arrowPadding","collisionBoundary","collisionPadding","sticky","hideWhenDetached","avoidCollisions"]),p=vl(tn,a),[m,v]=_.useState(null),[C,I]=_.useState(null),R=Ue(t,F=>v(F)),[w,A]=_.useState(null),[M,U]=_.useState(null),N=bf(a),[H,O]=_.useState(!1),G=_.useRef(!1);_.useEffect(()=>{if(m)return nf(m)},[m]),NR();let k=_.useCallback(F=>{let[le,...ve]=N().map(W=>W.ref.current),[$]=ve.slice(-1),ne=document.activeElement;for(let W of F)if(W===ne||(W==null||W.scrollIntoView({block:"nearest"}),W===le&&C&&(C.scrollTop=0),W===$&&C&&(C.scrollTop=C.scrollHeight),W==null||W.focus(),document.activeElement!==ne))return},[N,C]),j=_.useCallback(()=>k([w,m]),[k,w,m]);_.useEffect(()=>{H&&j()},[H,j]);let{onOpenChange:P,triggerPointerDownPosRef:Z}=p;_.useEffect(()=>{if(m){let F={x:0,y:0},le=$=>{var ne,W,at,Ye;F={x:Math.abs(Math.round($.pageX)-((W=(ne=Z.current)==null?void 0:ne.x)!=null?W:0)),y:Math.abs(Math.round($.pageY)-((Ye=(at=Z.current)==null?void 0:at.y)!=null?Ye:0))}},ve=$=>{F.x<=10&&F.y<=10?$.preventDefault():m.contains($.target)||P(!1),document.removeEventListener("pointermove",le),Z.current=null};return Z.current!==null&&(document.addEventListener("pointermove",le),document.addEventListener("pointerup",ve,{capture:!0,once:!0})),()=>{document.removeEventListener("pointermove",le),document.removeEventListener("pointerup",ve,{capture:!0})}}},[m,P,Z]),_.useEffect(()=>{let F=()=>P(!1);return window.addEventListener("blur",F),window.addEventListener("resize",F),()=>{window.removeEventListener("blur",F),window.removeEventListener("resize",F)}},[P]);let[J,ge]=VI(F=>{let le=N().filter(ne=>!ne.disabled),ve=le.find(ne=>ne.ref.current===document.activeElement),$=XI(le,F,ve);$&&setTimeout(()=>$.ref.current.focus())}),ue=_.useCallback((F,le,ve)=>{let $=!G.current&&!ve;(p.value!==void 0&&p.value===le||$)&&(A(F),$&&(G.current=!0))},[p.value]),ct=_.useCallback(()=>m==null?void 0:m.focus(),[m]),la=_.useCallback((F,le,ve)=>{let $=!G.current&&!ve;(p.value!==void 0&&p.value===le||$)&&U(F)},[p.value]),jt=o==="popper"?th:DI,Ot=jt===th?{side:r,sideOffset:i,align:s,alignOffset:f,arrowPadding:d,collisionBoundary:c,collisionPadding:g,sticky:L,hideWhenDetached:x,avoidCollisions:S}:{};return(0,q.jsx)(TI,{scope:a,content:m,viewport:C,onViewportChange:I,itemRefCallback:ue,selectedItem:w,onItemLeave:ct,itemTextRefCallback:la,focusSelectedItem:j,selectedItemText:M,position:o,isPositioned:H,searchRef:J,children:(0,q.jsx)(Nr,{as:l1,allowPinchZoom:!0,children:(0,q.jsx)(jm,{asChild:!0,trapped:p.open,onMountAutoFocus:F=>{F.preventDefault()},onUnmountAutoFocus:_e(l,F=>{var le;(le=p.trigger)==null||le.focus({preventScroll:!0}),F.preventDefault()}),children:(0,q.jsx)(qm,{asChild:!0,disableOutsidePointerEvents:!0,onEscapeKeyDown:n,onPointerDownOutside:u,onFocusOutside:F=>F.preventDefault(),onDismiss:()=>p.onOpenChange(!1),children:(0,q.jsx)(jt,b(y(y({role:"listbox",id:p.contentId,"data-state":p.open?"open":"closed",dir:p.dir,onContextMenu:F=>F.preventDefault()},h),Ot),{onPlaced:()=>O(!0),ref:R,style:y({display:"flex",flexDirection:"column",outline:"none"},h.style),onKeyDown:_e(h.onKeyDown,F=>{let le=F.ctrlKey||F.altKey||F.metaKey;if(F.key==="Tab"&&F.preventDefault(),!le&&F.key.length===1&&ge(F.key),["ArrowUp","ArrowDown","Home","End"].includes(F.key)){let $=N().filter(ne=>!ne.disabled).map(ne=>ne.ref.current);if(["ArrowUp","End"].includes(F.key)&&($=$.slice().reverse()),["ArrowUp","ArrowDown"].includes(F.key)){let ne=F.target,W=$.indexOf(ne);$=$.slice(W+1)}setTimeout(()=>k($)),F.preventDefault()}})}))})})})})});EI.displayName=o1;var n1="SelectItemAlignedPosition",DI=_.forwardRef((e,t)=>{let R=e,{__scopeSelect:a,onPlaced:o}=R,l=D(R,["__scopeSelect","onPlaced"]),n=vl(tn,a),u=Sl(tn,a),[r,i]=_.useState(null),[s,f]=_.useState(null),d=Ue(t,w=>f(w)),c=bf(a),g=_.useRef(!1),L=_.useRef(!0),{viewport:x,selectedItem:S,selectedItemText:h,focusSelectedItem:p}=u,m=_.useCallback(()=>{if(n.trigger&&n.valueNode&&r&&s&&x&&S&&h){let w=n.trigger.getBoundingClientRect(),A=s.getBoundingClientRect(),M=n.valueNode.getBoundingClientRect(),U=h.getBoundingClientRect();if(n.dir!=="rtl"){let ne=U.left-A.left,W=M.left-ne,at=w.left-W,Ye=w.width+at,Pt=Math.max(Ye,A.width),wa=window.innerWidth-Ha,za=km(W,[Ha,Math.max(Ha,wa-Pt)]);r.style.minWidth=Ye+"px",r.style.left=za+"px"}else{let ne=A.right-U.right,W=window.innerWidth-M.right-ne,at=window.innerWidth-w.right-W,Ye=w.width+at,Pt=Math.max(Ye,A.width),wa=window.innerWidth-Ha,za=km(W,[Ha,Math.max(Ha,wa-Pt)]);r.style.minWidth=Ye+"px",r.style.right=za+"px"}let N=c(),H=window.innerHeight-Ha*2,O=x.scrollHeight,G=window.getComputedStyle(s),k=parseInt(G.borderTopWidth,10),j=parseInt(G.paddingTop,10),P=parseInt(G.borderBottomWidth,10),Z=parseInt(G.paddingBottom,10),J=k+j+O+Z+P,ge=Math.min(S.offsetHeight*5,J),ue=window.getComputedStyle(x),ct=parseInt(ue.paddingTop,10),la=parseInt(ue.paddingBottom,10),jt=w.top+w.height/2-Ha,Ot=H-jt,tt=S.offsetHeight/2,F=S.offsetTop+tt,le=k+j+F,ve=J-le;if(le<=jt){let ne=N.length>0&&S===N[N.length-1].ref.current;r.style.bottom="0px";let W=s.clientHeight-x.offsetTop-x.offsetHeight,at=Math.max(Ot,tt+(ne?la:0)+W+P),Ye=le+at;r.style.height=Ye+"px"}else{let ne=N.length>0&&S===N[0].ref.current;r.style.top="0px";let at=Math.max(jt,k+x.offsetTop+(ne?ct:0)+tt)+ve;r.style.height=at+"px",x.scrollTop=le-jt+x.offsetTop}r.style.margin="".concat(Ha,"px 0"),r.style.minHeight=ge+"px",r.style.maxHeight=H+"px",o==null||o(),requestAnimationFrame(()=>g.current=!0)}},[c,n.trigger,n.valueNode,r,s,x,S,h,n.dir,o]);je(()=>m(),[m]);let[v,C]=_.useState();je(()=>{s&&C(window.getComputedStyle(s).zIndex)},[s]);let I=_.useCallback(w=>{w&&L.current===!0&&(m(),p==null||p(),L.current=!1)},[m,p]);return(0,q.jsx)(r1,{scope:a,contentWrapper:r,shouldExpandOnScrollRef:g,onScrollButtonChange:I,children:(0,q.jsx)("div",{ref:i,style:{display:"flex",flexDirection:"column",position:"fixed",zIndex:v},children:(0,q.jsx)(Le.div,b(y({},l),{ref:d,style:y({boxSizing:"border-box",maxHeight:"100%"},l.style)}))})})});DI.displayName=n1;var u1="SelectPopperPosition",th=_.forwardRef((e,t)=>{let r=e,{__scopeSelect:a,align:o="start",collisionPadding:l=Ha}=r,n=D(r,["__scopeSelect","align","collisionPadding"]),u=wf(a);return(0,q.jsx)(rI,b(y(y({},u),n),{ref:t,align:o,collisionPadding:l,style:b(y({boxSizing:"border-box"},n.style),{"--radix-select-content-transform-origin":"var(--radix-popper-transform-origin)","--radix-select-content-available-width":"var(--radix-popper-available-width)","--radix-select-content-available-height":"var(--radix-popper-available-height)","--radix-select-trigger-width":"var(--radix-popper-anchor-width)","--radix-select-trigger-height":"var(--radix-popper-anchor-height)"})}))});th.displayName=u1;var[r1,rh]=pu(tn,{}),ah="SelectViewport",MI=_.forwardRef((e,t)=>{let s=e,{__scopeSelect:a,nonce:o}=s,l=D(s,["__scopeSelect","nonce"]),n=Sl(ah,a),u=rh(ah,a),r=Ue(t,n.onViewportChange),i=_.useRef(0);return(0,q.jsxs)(q.Fragment,{children:[(0,q.jsx)("style",{dangerouslySetInnerHTML:{__html:"[data-radix-select-viewport]{scrollbar-width:none;-ms-overflow-style:none;-webkit-overflow-scrolling:touch;}[data-radix-select-viewport]::-webkit-scrollbar{display:none}"},nonce:o}),(0,q.jsx)(If.Slot,{scope:a,children:(0,q.jsx)(Le.div,b(y({"data-radix-select-viewport":"",role:"presentation"},l),{ref:r,style:y({position:"relative",flex:1,overflow:"hidden auto"},l.style),onScroll:_e(l.onScroll,f=>{let d=f.currentTarget,{contentWrapper:c,shouldExpandOnScrollRef:g}=u;if(g!=null&&g.current&&c){let L=Math.abs(i.current-d.scrollTop);if(L>0){let x=window.innerHeight-Ha*2,S=parseFloat(c.style.minHeight),h=parseFloat(c.style.height),p=Math.max(S,h);if(p<x){let m=p+L,v=Math.min(x,m),C=m-v;c.style.height=v+"px",c.style.bottom==="0px"&&(d.scrollTop=C>0?C:0,c.style.justifyContent="flex-end")}}}i.current=d.scrollTop})}))})]})});MI.displayName=ah;var OI="SelectGroup",[i1,s1]=pu(OI),f1=_.forwardRef((e,t)=>{let n=e,{__scopeSelect:a}=n,o=D(n,["__scopeSelect"]),l=Sf();return(0,q.jsx)(i1,{scope:a,id:l,children:(0,q.jsx)(Le.div,b(y({role:"group","aria-labelledby":l},o),{ref:t}))})});f1.displayName=OI;var PI="SelectLabel",c1=_.forwardRef((e,t)=>{let n=e,{__scopeSelect:a}=n,o=D(n,["__scopeSelect"]),l=s1(PI,a);return(0,q.jsx)(Le.div,b(y({id:l.id},o),{ref:t}))});c1.displayName=PI;var Rf="SelectItem",[d1,BI]=pu(Rf),_I=_.forwardRef((e,t)=>{let p=e,{__scopeSelect:a,value:o,disabled:l=!1,textValue:n}=p,u=D(p,["__scopeSelect","value","disabled","textValue"]),r=vl(Rf,a),i=Sl(Rf,a),s=r.value===o,[f,d]=_.useState(n!=null?n:""),[c,g]=_.useState(!1),L=Ue(t,m=>{var v;return(v=i.itemRefCallback)==null?void 0:v.call(i,m,o,l)}),x=Sf(),S=_.useRef("touch"),h=()=>{l||(r.onValueChange(o),r.onOpenChange(!1))};if(o==="")throw new Error("A <Select.Item /> must have a value prop that is not an empty string. This is because the Select value can be set to an empty string to clear the selection and show the placeholder.");return(0,q.jsx)(d1,{scope:a,value:o,disabled:l,textId:x,isSelected:s,onItemTextChange:_.useCallback(m=>{d(v=>{var C;return v||((C=m==null?void 0:m.textContent)!=null?C:"").trim()})},[]),children:(0,q.jsx)(If.ItemSlot,{scope:a,value:o,disabled:l,textValue:f,children:(0,q.jsx)(Le.div,b(y({role:"option","aria-labelledby":x,"data-highlighted":c?"":void 0,"aria-selected":s&&c,"data-state":s?"checked":"unchecked","aria-disabled":l||void 0,"data-disabled":l?"":void 0,tabIndex:l?void 0:-1},u),{ref:L,onFocus:_e(u.onFocus,()=>g(!0)),onBlur:_e(u.onBlur,()=>g(!1)),onClick:_e(u.onClick,()=>{S.current!=="mouse"&&h()}),onPointerUp:_e(u.onPointerUp,()=>{S.current==="mouse"&&h()}),onPointerDown:_e(u.onPointerDown,m=>{S.current=m.pointerType}),onPointerMove:_e(u.onPointerMove,m=>{var v;S.current=m.pointerType,l?(v=i.onItemLeave)==null||v.call(i):S.current==="mouse"&&m.currentTarget.focus({preventScroll:!0})}),onPointerLeave:_e(u.onPointerLeave,m=>{var v;m.currentTarget===document.activeElement&&((v=i.onItemLeave)==null||v.call(i))}),onKeyDown:_e(u.onKeyDown,m=>{var C;((C=i.searchRef)==null?void 0:C.current)!==""&&m.key===" "||(QD.includes(m.key)&&h(),m.key===" "&&m.preventDefault())})}))})})});_I.displayName=Rf;var Fr="SelectItemText",UI=_.forwardRef((e,t)=>{let h=e,{__scopeSelect:a,className:o,style:l}=h,n=D(h,["__scopeSelect","className","style"]),u=vl(Fr,a),r=Sl(Fr,a),i=BI(Fr,a),s=e1(Fr,a),[f,d]=_.useState(null),c=Ue(t,p=>d(p),i.onItemTextChange,p=>{var m;return(m=r.itemTextRefCallback)==null?void 0:m.call(r,p,i.value,i.disabled)}),g=f==null?void 0:f.textContent,L=_.useMemo(()=>(0,q.jsx)("option",{value:i.value,disabled:i.disabled,children:g},i.value),[i.disabled,i.value,g]),{onNativeOptionAdd:x,onNativeOptionRemove:S}=s;return je(()=>(x(L),()=>S(L)),[x,S,L]),(0,q.jsxs)(q.Fragment,{children:[(0,q.jsx)(Le.span,b(y({id:i.textId},n),{ref:c})),i.isSelected&&u.valueNode&&!u.valueNodeHasChildren?uh.createPortal(n.children,u.valueNode):null]})});UI.displayName=Fr;var NI="SelectItemIndicator",kI=_.forwardRef((e,t)=>{let n=e,{__scopeSelect:a}=n,o=D(n,["__scopeSelect"]);return BI(NI,a).isSelected?(0,q.jsx)(Le.span,b(y({"aria-hidden":!0},o),{ref:t})):null});kI.displayName=NI;var oh="SelectScrollUpButton",HI=_.forwardRef((e,t)=>{let a=Sl(oh,e.__scopeSelect),o=rh(oh,e.__scopeSelect),[l,n]=_.useState(!1),u=Ue(t,o.onScrollButtonChange);return je(()=>{if(a.viewport&&a.isPositioned){let i=function(){let f=s.scrollTop>0;n(f)};var r=i;let s=a.viewport;return i(),s.addEventListener("scroll",i),()=>s.removeEventListener("scroll",i)}},[a.viewport,a.isPositioned]),l?(0,q.jsx)(FI,b(y({},e),{ref:u,onAutoScroll:()=>{let{viewport:r,selectedItem:i}=a;r&&i&&(r.scrollTop=r.scrollTop-i.offsetHeight)}})):null});HI.displayName=oh;var lh="SelectScrollDownButton",zI=_.forwardRef((e,t)=>{let a=Sl(lh,e.__scopeSelect),o=rh(lh,e.__scopeSelect),[l,n]=_.useState(!1),u=Ue(t,o.onScrollButtonChange);return je(()=>{if(a.viewport&&a.isPositioned){let i=function(){let f=s.scrollHeight-s.clientHeight,d=Math.ceil(s.scrollTop)<f;n(d)};var r=i;let s=a.viewport;return i(),s.addEventListener("scroll",i),()=>s.removeEventListener("scroll",i)}},[a.viewport,a.isPositioned]),l?(0,q.jsx)(FI,b(y({},e),{ref:u,onAutoScroll:()=>{let{viewport:r,selectedItem:i}=a;r&&i&&(r.scrollTop=r.scrollTop+i.offsetHeight)}})):null});zI.displayName=lh;var FI=_.forwardRef((e,t)=>{let s=e,{__scopeSelect:a,onAutoScroll:o}=s,l=D(s,["__scopeSelect","onAutoScroll"]),n=Sl("SelectScrollButton",a),u=_.useRef(null),r=bf(a),i=_.useCallback(()=>{u.current!==null&&(window.clearInterval(u.current),u.current=null)},[]);return _.useEffect(()=>()=>i(),[i]),je(()=>{var d;let f=r().find(c=>c.ref.current===document.activeElement);(d=f==null?void 0:f.ref.current)==null||d.scrollIntoView({block:"nearest"})},[r]),(0,q.jsx)(Le.div,b(y({"aria-hidden":!0},l),{ref:t,style:y({flexShrink:0},l.style),onPointerDown:_e(l.onPointerDown,()=>{u.current===null&&(u.current=window.setInterval(o,50))}),onPointerMove:_e(l.onPointerMove,()=>{var f;(f=n.onItemLeave)==null||f.call(n),u.current===null&&(u.current=window.setInterval(o,50))}),onPointerLeave:_e(l.onPointerLeave,()=>{i()})}))}),p1="SelectSeparator",m1=_.forwardRef((e,t)=>{let l=e,{__scopeSelect:a}=l,o=D(l,["__scopeSelect"]);return(0,q.jsx)(Le.div,b(y({"aria-hidden":!0},o),{ref:t}))});m1.displayName=p1;var nh="SelectArrow",h1=_.forwardRef((e,t)=>{let r=e,{__scopeSelect:a}=r,o=D(r,["__scopeSelect"]),l=wf(a),n=vl(nh,a),u=Sl(nh,a);return n.open&&u.position==="popper"?(0,q.jsx)(iI,b(y(y({},l),o),{ref:t})):null});h1.displayName=nh;var g1="SelectBubbleInput",qI=_.forwardRef((l,o)=>{var n=l,{__scopeSelect:e,value:t}=n,a=D(n,["__scopeSelect","value"]);let u=_.useRef(null),r=Ue(o,u),i=cI(t);return _.useEffect(()=>{let s=u.current;if(!s)return;let f=window.HTMLSelectElement.prototype,c=Object.getOwnPropertyDescriptor(f,"value").set;if(i!==t&&c){let g=new Event("change",{bubbles:!0});c.call(s,t),s.dispatchEvent(g)}},[i,t]),(0,q.jsx)(Le.select,b(y({},a),{style:y(y({},eh),a.style),ref:r,defaultValue:t}))});qI.displayName=g1;function GI(e){return e===""||e===void 0}function VI(e){let t=ao(e),a=_.useRef(""),o=_.useRef(0),l=_.useCallback(u=>{let r=a.current+u;t(r),(function i(s){a.current=s,window.clearTimeout(o.current),s!==""&&(o.current=window.setTimeout(()=>i(""),1e3))})(r)},[t]),n=_.useCallback(()=>{a.current="",window.clearTimeout(o.current)},[]);return _.useEffect(()=>()=>window.clearTimeout(o.current),[]),[a,l,n]}function XI(e,t,a){let l=t.length>1&&Array.from(t).every(s=>s===t[0])?t[0]:t,n=a?e.indexOf(a):-1,u=v1(e,Math.max(n,0));l.length===1&&(u=u.filter(s=>s!==a));let i=u.find(s=>s.textValue.toLowerCase().startsWith(l.toLowerCase()));return i!==a?i:void 0}function v1(e,t){return e.map((a,o)=>e[(t+o)%e.length])}var jI=LI,YI=CI,KI=II,ZI=bI,QI=wI,WI=AI,JI=MI;var $I=_I,eb=UI,tb=kI,ab=HI,ob=zI;var bt=T(X()),lb=jI,nb=KI,ih=Af.forwardRef((l,o)=>{var n=l,{className:e,children:t}=n,a=D(n,["className","children"]);return(0,bt.jsxs)(YI,b(y({ref:o,className:Be("ui-select-trigger",e)},a),{children:[t,(0,bt.jsx)(ZI,{asChild:!0,children:(0,bt.jsx)(eu,{size:16,"aria-hidden":"true"})})]}))}),sh=Af.forwardRef((i,r)=>{var s=i,{className:e,children:t,position:a="popper",sideOffset:o=4,collisionPadding:l=8,style:n}=s,u=D(s,["className","children","position","sideOffset","collisionPadding","style"]);let f=mf();return(0,bt.jsx)(QI,{children:(0,bt.jsxs)(WI,b(y(b(y({},f),{ref:r,className:Be("ui-select-content",e),position:a,sideOffset:o,collisionPadding:l}),u),{style:y(y({},f==null?void 0:f.style),n),children:[(0,bt.jsx)(ab,{className:"ui-select-scroll",children:(0,bt.jsx)(wr,{size:16,"aria-hidden":"true"})}),(0,bt.jsx)(JI,{className:"ui-select-viewport",children:t}),(0,bt.jsx)(ob,{className:"ui-select-scroll",children:(0,bt.jsx)(eu,{size:16,"aria-hidden":"true"})})]}))})}),fh=Af.forwardRef((l,o)=>{var n=l,{className:e,children:t}=n,a=D(n,["className","children"]);return(0,bt.jsxs)($I,b(y({ref:o,className:Be("ui-select-item",e)},a),{children:[(0,bt.jsx)(eb,{children:t}),(0,bt.jsx)(tb,{className:"ui-select-item-indicator",children:(0,bt.jsx)(Rr,{size:14,"aria-hidden":"true"})})]}))});ih.displayName="SelectTrigger";sh.displayName="SelectContent";fh.displayName="SelectItem";var Oo=T(B());var Po=T(X()),ch=Oo.forwardRef((n,l)=>{var u=n,{className:e,tabIndex:t=0,role:a="region"}=u,o=D(u,["className","tabIndex","role"]);return(0,Po.jsx)("div",y({ref:l,tabIndex:t,role:a,className:Be("ui-table-container",e)},o))});ch.displayName="TableContainer";var dh=Oo.forwardRef((o,a)=>{var l=o,{className:e}=l,t=D(l,["className"]);return(0,Po.jsx)("table",y({ref:a,className:Be("ui-table",e)},t))});dh.displayName="Table";var ph=Oo.forwardRef((o,a)=>{var l=o,{className:e}=l,t=D(l,["className"]);return(0,Po.jsx)("thead",y({ref:a,className:Be("ui-table-header",e)},t))});ph.displayName="TableHeader";var mh=Oo.forwardRef((o,a)=>{var l=o,{className:e}=l,t=D(l,["className"]);return(0,Po.jsx)("tbody",y({ref:a,className:Be("ui-table-body",e)},t))});mh.displayName="TableBody";var qr=Oo.forwardRef((o,a)=>{var l=o,{className:e}=l,t=D(l,["className"]);return(0,Po.jsx)("tr",y({ref:a,className:Be("ui-table-row",e)},t))});qr.displayName="TableRow";var hh=Oo.forwardRef((o,a)=>{var l=o,{className:e}=l,t=D(l,["className"]);return(0,Po.jsx)("th",y({ref:a,className:Be("ui-table-head",e)},t))});hh.displayName="TableHead";var Tf=Oo.forwardRef((o,a)=>{var l=o,{className:e}=l,t=D(l,["className"]);return(0,Po.jsx)("td",y({ref:a,className:Be("ui-table-cell",e)},t))});Tf.displayName="TableCell";var x1=Oo.forwardRef((o,a)=>{var l=o,{className:e}=l,t=D(l,["className"]);return(0,Po.jsx)("caption",y({ref:a,className:Be("ui-table-caption",e)},t))});x1.displayName="TableCaption";var he=T(X()),Ef="__dashboard_all__",L1=[["today","\uC624\uB298"],["yesterday","\uC5B4\uC81C"],["7d","7\uC77C"],["30d","30\uC77C"],["month","\uC774\uBC88 \uB2EC"],["custom","\uC9C1\uC811"]],ub=e=>e?new Date(...e.split("-").map((t,a)=>Number(t)-(a===1?1:0))):null;function vh(e,t){let a=Object.getOwnPropertyDescriptor(Object.getPrototypeOf(e),"value");Object.defineProperty(e,"value",{configurable:!0,get(){return a.get.call(this)},set(o){a.set.call(this,o),t()}}),e.addEventListener("change",t),new MutationObserver(t).observe(e,{attributes:!0,childList:!0,subtree:!0})}var y1=document.getElementById("presetGroup"),C1=(0,Gr.createRoot)(y1);function ib(){C1.render((0,he.jsx)(jp,{id:"presetGroupControl",size:"sm",fullWidth:!0,value:state.currentPreset,"aria-label":"\uAE30\uAC04 \uC120\uD0DD",orientation:"horizontal",onValueChange:setPreset,children:L1.map(([e,t])=>(0,he.jsx)(Yp,{value:e,"data-preset":e,className:"preset-btn","aria-controls":e==="custom"?"customDateWrap":void 0,children:t},e))}))}window.renderDashboardPresets=ib;(0,mu.flushSync)(ib);var Df=document.getElementById("filterFromDate"),Mf=document.getElementById("filterToDate"),R1=(0,Gr.createRoot)(document.getElementById("dateRangeMount"));function Sh(){R1.render((0,he.jsx)(SR,{id:"filterDateRangeTrigger",mode:"range",value:{from:ub(Df.value),to:ub(Mf.value)},disabled:Df.disabled||Mf.disabled,"aria-label":"\uC870\uD68C \uAE30\uAC04",calendarLabel:"\uC870\uD68C \uAE30\uAC04 \uC120\uD0DD",locale:"ko-KR",onValueChange:e=>{Df.value=ka(e.from),Mf.value=ka(e.to),applyFilters()}}))}vh(Df,Sh);vh(Mf,Sh);(0,mu.flushSync)(Sh);var sb=new Map;function I1(e){var t,a;return e.time?(0,he.jsxs)(he.Fragment,{children:[(0,he.jsx)("span",{className:"record-time-full",children:e.time.full}),(0,he.jsx)("span",{className:"record-time-mobile",children:e.time.mobile})]}):e.badge?(0,he.jsx)("span",{className:e.badge.className,children:e.badge.text}):(a=(t=e.content)!=null?t:e.text)!=null?a:""}function fb({tableId:e,bodyId:t,label:a,wrapperClass:o,headers:l,rows:n,emptyText:u,tableClass:r}){return(0,he.jsx)(ch,{role:"region","aria-label":a,className:o,children:(0,he.jsxs)(dh,{id:e,className:r,children:[(0,he.jsx)(ph,{children:(0,he.jsx)(qr,{children:l.map((i,s)=>(0,he.jsx)(hh,{className:i.className,children:i.text},s))})}),(0,he.jsx)(mh,{id:t,children:n.length?n.map(i=>(0,he.jsx)(qr,{tabIndex:i.tabIndex,"data-index":i.dataIndex,"data-row-key":i.key,onClick:i.onClick,onKeyDown:i.onKeyDown,children:i.cells.map((s,f)=>(0,he.jsx)(Tf,{colSpan:s.colSpan,className:s.className,title:s.title,"data-change-key":s.changeKey,ref:d=>{d&&s.tone&&typeof window.flashValue=="function"&&window.flashValue(d,s.tone,s.label)},children:I1(s)},f))},i.key)):(0,he.jsx)(qr,{children:(0,he.jsx)(Tf,{colSpan:l.length,className:"text-muted table-empty",children:u||"\uB370\uC774\uD130 \uC5C6\uC74C"})})})]})})}var Of=[{text:"Provider"},{text:"\uC694\uCCAD",className:"text-right"},{text:"\uC2E4\uD328",className:"text-right"},{text:"\uD1A0\uD070",className:"text-right"}],cb=[{text:"Endpoint/Source"},...Of.slice(1)],db=[{text:"Model"},...Of.slice(1)],pb=[{text:"\uC2DC\uAC04"},{text:"Provider"},{text:"Endpoint"},{text:"Model"},{text:"\uC2E0\uADDC \uC785\uB825",className:"text-right"},{text:"\uCE90\uC2DC",className:"text-right"},{text:"\uCD9C\uB825",className:"text-right"},{text:"\uD569\uACC4",className:"text-right"},{text:"\uC9C0\uC5F0",className:"text-right"},{text:"\uC0C1\uD0DC"}],gh=[{key:"loading",cells:[{text:"\uBD88\uB7EC\uC624\uB294 \uC911...",className:"text-muted table-empty"},{text:""},{text:""},{text:""}]}],b1=[{key:"loading",cells:[{text:"\uB370\uC774\uD130\uB97C \uBD88\uB7EC\uC624\uB294 \uC911\uC785\uB2C8\uB2E4...",className:"text-muted table-empty"},...Array.from({length:9},()=>({text:""}))]}],mb=[{mountId:"providerAggTableMount",tableId:"providerAggTable",bodyId:"providerAggBody",label:"Provider \uC0C1\uC138 \uC9D1\uACC4",wrapperClass:"usage-table-scroll breakdown-table-wrap",headers:Of,rows:gh},{mountId:"endpointAggTableMount",tableId:"endpointAggTable",bodyId:"endpointAggBody",label:"Endpoint \uC0C1\uC138 \uC9D1\uACC4",wrapperClass:"usage-table-scroll breakdown-table-wrap",headers:cb,rows:gh},{mountId:"modelAggTableMount",tableId:"modelAggTable",bodyId:"modelAggBody",label:"Model \uC0C1\uC138 \uC9D1\uACC4",wrapperClass:"usage-table-scroll breakdown-table-wrap",headers:db,rows:gh},{mountId:"recordsTableMount",tableId:"recordsTable",bodyId:"recordsBody",label:"\uCD5C\uADFC \uC694\uCCAD \uAE30\uB85D \uD14C\uC774\uBE14 \uC2A4\uD06C\uB864 \uC601\uC5ED",wrapperClass:"usage-table-scroll table-responsive",headers:pb,rows:b1,tableClass:"records-table"}];for(let e of mb){let t=document.getElementById(e.mountId);if(!t)continue;let a=(0,Gr.createRoot)(t);sb.set(e.mountId,a),(0,mu.flushSync)(()=>a.render((0,he.jsx)(fb,y({},e))))}window.renderDohwiTable=(e,t,a)=>{let o=sb.get(e),l=mb.find(n=>n.mountId===e);o&&l&&(0,mu.flushSync)(()=>o.render((0,he.jsx)(fb,b(y({},l),{rows:t,emptyText:a}))))};window.dohwiTableHeaders={AGG_HEADERS:Of,ENDPOINT_HEADERS:cb,MODEL_HEADERS:db,RECORD_HEADERS:pb};for(let e of window.pendingDohwiTableUpdates||[])window.renderDohwiTable(e[0],e[1],e[2]);window.pendingDohwiTableUpdates=[];var rb;for(let e of document.querySelectorAll(".compact-select-wrap > select")){let o=function(){a.render((0,he.jsxs)(lb,{value:e.value||Ef,disabled:e.disabled,onValueChange:l=>{e.value=l===Ef?"":l,e.dispatchEvent(new Event("change",{bubbles:!0}))},children:[(0,he.jsx)(ih,{id:e.id+"Trigger",className:"compact-select","aria-label":e.getAttribute("aria-label"),title:e.title,children:(0,he.jsx)(nb,{})}),(0,he.jsx)(sh,{collisionPadding:16,children:Array.from(e.options).map(l=>(0,he.jsx)(fh,{value:l.value||Ef,disabled:l.disabled,children:l.textContent},l.value||Ef))})]}))};e.hidden=!0,e.tabIndex=-1;let t=document.createElement("div");t.className="dashboard-select-mount",e.parentElement.appendChild(t);let a=(0,Gr.createRoot)(t);vh(e,o),(rb=document.querySelector('label[for="'.concat(e.id,'"]')))==null||rb.setAttribute("for",e.id+"Trigger"),(0,mu.flushSync)(o)}})();
</script>
<script data-dohwi-motion="material">
(function() {
  var root = document.documentElement;
  var reduced = window.matchMedia('(prefers-reduced-motion: reduce)');
  function motionOn() { return !reduced.matches && !document.hidden; }

  /* ---- Sliding indicator (dohwi-ui SegmentedControl / Tabs, vanilla edition) ----
     Same contract as use-sliding-indicator.ts: the first position is placed instantly, and only later
     moves are animated (data-indicator-animate is enabled two frames after the first placement).
     offset* is used instead of getBoundingClientRect so a pressed (scaled) item never skews the box. */
  function slidingIndicator(group, itemSelector) {
    if (!group || group.getAttribute('data-indicator-bound')) return;
    var indicator = group.querySelector(':scope > .ui-segmented-indicator, :scope > .ui-tabs-indicator');
    if (!indicator) return;
    group.setAttribute('data-indicator-bound', 'true');
    var first = true;
    function items() { return Array.prototype.slice.call(group.querySelectorAll(itemSelector)); }
    function place() {
      var active = items().filter(function(el) { return el.classList.contains('active'); })[0];
      if (!active || !active.offsetWidth) { group.removeAttribute('data-indicator-ready'); group.removeAttribute('data-indicator-animate'); first = true; return; }
      var animate = !first && motionOn();
      if (!animate) group.removeAttribute('data-indicator-animate');
      indicator.style.width = active.offsetWidth + 'px';
      indicator.style.height = active.offsetHeight + 'px';
      indicator.style.transform = 'translate(' + active.offsetLeft + 'px, ' + active.offsetTop + 'px)';
      group.setAttribute('data-indicator-ready', '');
      if (animate) group.setAttribute('data-indicator-animate', '');
      else requestAnimationFrame(function() { requestAnimationFrame(function() { group.setAttribute('data-indicator-animate', ''); }); });
      first = false;
    }
    new MutationObserver(function() { place(); }).observe(group, { attributes: true, attributeFilter: ['class'], subtree: true });
    if (window.ResizeObserver) new ResizeObserver(function() { var keep = first; first = true; place(); first = keep; }).observe(group);
    if (document.fonts && document.fonts.ready) document.fonts.ready.then(function() { first = true; place(); });
    place();
  }
  // SegmentedControl owns its indicator via useSlidingIndicator.
  slidingIndicator(document.querySelector('.tab-nav'), '.tab-btn');

  /* ---- Material ripple (dohwi-ui vanilla edition) ---- */
  var live = new Map();
  function clearRipple(ripple) {
    clearTimeout(live.get(ripple)); live.delete(ripple);
    var layer = ripple.parentElement; ripple.remove();
    if (layer && !layer.childElementCount) layer.remove();
  }
  function addRipple(button, x, y) {
    if (!motionOn() || button.disabled || button.getAttribute('aria-disabled') === 'true') return;
    var rect = button.getBoundingClientRect();
    if (!rect.width || !rect.height) return;
    var px = x === null ? button.clientWidth / 2 : (x - rect.left) * button.offsetWidth / rect.width;
    var py = y === null ? button.clientHeight / 2 : (y - rect.top) * button.offsetHeight / rect.height;
    var layer = button.querySelector(':scope > .ui-button-ripple-layer');
    if (!layer) { layer = document.createElement('span'); layer.className = 'ui-button-ripple-layer'; layer.setAttribute('aria-hidden', 'true'); button.appendChild(layer); }
    while (layer.childElementCount >= 4) clearRipple(layer.firstElementChild);
    var ripple = document.createElement('span');
    ripple.className = 'ui-button-ripple';
    var size = Math.hypot(Math.max(px, button.offsetWidth - px), Math.max(py, button.offsetHeight - py)) * 2;
    ripple.style.left = px + 'px'; ripple.style.top = py + 'px'; ripple.style.width = size + 'px'; ripple.style.height = size + 'px';
    layer.appendChild(ripple);
    live.set(ripple, setTimeout(function() { clearRipple(ripple); }, 900));
  }
  var RIPPLE = '.btn';
  function bindRipples(scope) {
    Array.prototype.forEach.call((scope || document).querySelectorAll(RIPPLE), function(el) { el.classList.add('ui-ripple-host'); });
  }
  bindRipples(document);
  new MutationObserver(function() { bindRipples(document); }).observe(document.body, { childList: true, subtree: true });
  document.addEventListener('pointerdown', function(event) {
    if (event.button !== 0 || !event.target.closest) return;
    var button = event.target.closest('.ui-ripple-host');
    if (button) addRipple(button, event.clientX, event.clientY);
  }, true);
  document.addEventListener('click', function(event) {
    if (event.detail !== 0 || !event.target.closest) return;
    var button = event.target.closest('.ui-ripple-host');
    if (button) addRipple(button, null, null);
  }, true);
  document.addEventListener('animationend', function(event) { if (live.has(event.target)) clearRipple(event.target); }, true);

  /* ---- Staggered entrance for freshly rendered panels ---- */
  window.motionEnter = function(nodes) {
    if (!motionOn()) return;
    Array.prototype.forEach.call(nodes, function(node, index) {
      if (!node || !node.classList) return;
      node.style.setProperty('--enter-delay', Math.min(index, 8) * 40 + 'ms');
      node.classList.remove('is-entering');
      void node.offsetWidth;
      node.classList.add('is-entering');
      var done = function() { node.classList.remove('is-entering'); node.style.removeProperty('--enter-delay'); node.removeEventListener('animationend', done); };
      node.addEventListener('animationend', done);
    });
  };
  window.addEventListener('load', function() {
    window.motionEnter(document.querySelectorAll('#usageDataContent .kpi-group, #chartCard, #breakdownSection, #recordsPanel'));
  });
})();
</script>
</body>
</html>`
