# Plugin dashboard design

Usage-audit embeds the actual exported dohwi-ui **SegmentedControl**, **DatePicker** (`mode="range"`), **Calendar**, **Popover**, and **Select** React components with their Radix dependencies. Quota-monitor has no period/date controls and is unchanged by this integration.

## Components and form state

- `usage-audit/ui/vendor/` contains the dohwi-ui source components. `entry.tsx` connects them to the existing query state. `npm ci && npm run build` in that directory updates the marked CSS/JS regions in `usage-audit/internal/plugin/dashboard.html`.
- The dashboard is self-contained: React/Radix are bundled, with no CDN or external JavaScript request. Normal Go builds use the checked-in bundle and need no Node tooling.
- SegmentedControl owns its radio keyboard model and measured sliding indicator. The adapter only calls `setPreset` and supplies controlled values.
- One range DatePicker replaces the two single-date popups. First selection is an uncommitted draft; hovering/focusing previews the range. Second selection orders both endpoints and commits them atomically, issuing one query. Escape/outside dismissal discards the draft. A same-day range is valid.
- Hidden native dates and selects remain backing fields for the existing query/validation code. Native programmatic writes, option refreshes and resets update the React controls.
- Radix owns Select/Popover focus, pointer selection, dismissals, scrolling and collision handling. A real pointer press/release regression covers the old blur-before-click bug, which affected both standalone and iframe dashboards.

## Layout

- Existing dashboard colors, 38px desktop controls and responsive layout remain host tokens. Mobile controls/calendar days are at least 44px.
- Usage aggregation and recent-request ledgers are rendered with the actual exported dohwi-ui `TableContainer`, `Table`, `TableHeader`, `TableBody`, `TableRow`, `TableHead`, and `TableCell` React primitives. The source is vendored at `scripts/dashboard-ui/vendor/table.tsx`; the adapter only adds the existing dense widths, sticky header, right-aligned numeric cells, scroll height, and record click behavior. dohwi-ui owns the rounded outer border and row rules—there are no per-column vertical borders or handwritten replacement table markup. Empty/loading/data states all use the same Table component tree.
- Usage trends are always visible in a labelled section; no details/summary disclosure or collapse listener remains.
- Calendar week rows do not inherit the data-table Hover/cursor rules; only individual days/range bands are highlighted.
- Both page headings contain only a title and description. Header last-updated timestamps and relative-age timers are removed; account-level cache freshness and error notices remain.
- Initial summary/records requests start before React initialization, and each first-load panel renders as soon as its response arrives. Subsequent refreshes commit atomically so failures preserve the previous data, cursor and diff baseline.
- Recent-record paging scans the immutable snapshot frontier with a bounded `limit+1` index heap instead of copying/sorting every matching record. Timestamp/sequence ordering and cursor anchors remain unchanged; out-of-order/backdated records are not skipped. `BenchmarkRecordsPageTopK` and `BenchmarkRecordsPageChronological` compare the old reference path with the bounded implementation (in-memory synthetic workloads, not end-to-end latency).
- Reduced-motion preferences disable dohwi-ui animations. Popup styles and indicator motion come from exported dohwi-ui CSS.

## Validation

```sh
make verify-quota verify-usage
PLAYWRIGHT_MODULE=/path/to/playwright node usage-audit/scripts/test-usage-dashboard-loading.cjs
PLAYWRIGHT_MODULE=/path/to/playwright node usage-audit/scripts/test-usage-dashboard-layout.cjs
PLAYWRIGHT_MODULE=/path/to/playwright node scripts/test-dashboard-headers.cjs
bash scripts/preview-server/preview.sh start
PLAYWRIGHT_MODULE=/path/to/playwright node scripts/preview-server/ui-interaction-verify.cjs
PLAYWRIGHT_MODULE=/path/to/playwright node scripts/preview-server/segdate-verify.cjs
PLAYWRIGHT_MODULE=/path/to/playwright node scripts/preview-server/select-verify.cjs
PLAYWRIGHT_MODULE=/path/to/playwright node scripts/preview-server/motion-verify.cjs
```
