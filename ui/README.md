# Embedded dashboard controls

The usage dashboard now consumes the actual exported dohwi-ui React components, not handwritten popup/calendar/radio approximations.

- `vendor/`: unchanged copies from `dohwi-ui/src/components/ui` (SegmentedControl, DatePicker, Calendar, Popover, Select and their local dependencies/styles).
- `entry.tsx`: dashboard adapter. Native dates/selects remain query-state backing fields. A complete date range is written atomically and applied once; unfinished drafts are owned by DatePicker.
- `build.mjs`: bundles React, Radix and the components into the marked CSS/JS regions in `usage-audit/internal/plugin/dashboard.html`. `internal/plugin/dashboard.go` embeds that file together with `style.css` and `app.js` and assembles the page at start-up. No CDN, external JS request or Node runtime is needed by the plugin.

Rebuild after adapter/vendor changes:

```sh
cd usage-audit/ui
npm ci
npm run build
```

The generated dashboard stays checked in so normal Go builds do not require frontend tooling. Dependencies are pinned in `package-lock.json`. Update vendor sources as a unit with dohwi-ui and rebuild; do not edit the generated bundle. The adapter contains layout/token adjustments only; keyboard, focus, popup collision, range drafts and selection behavior belong to dohwi-ui/Radix.
