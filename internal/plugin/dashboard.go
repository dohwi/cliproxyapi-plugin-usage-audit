package plugin

import (
	_ "embed"
	"strings"
)

// The dashboard page is split into plain files: dashboard.html (markup plus the
// generated dohwi-ui regions written by ui/build.mjs), style.css and app.js (the
// hand-written page styles and logic). They are assembled once at start-up into a
// single self-contained page (no CDN, no extra requests).
//
//go:embed dashboard.html
var dashboardShell string

//go:embed style.css
var dashboardStyles string

//go:embed app.js
var dashboardScript string

// DashboardHTML is the self-contained dashboard page for usage-audit.
var DashboardHTML = assembleDashboard(dashboardShell, dashboardStyles, dashboardScript)

func assembleDashboard(shell, styles, script string) string {
	page := strings.Replace(shell, "@@STYLE@@\n", styles, 1)
	return strings.Replace(page, "@@SCRIPT@@\n", script, 1)
}
