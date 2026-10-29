package plugin

import (
	"bytes"
	"testing"
)

// The cached dashboard envelope must equal a freshly built one byte for byte,
// and repeated calls must reuse the same buffer.
func TestDashboardResponseCachedMatchesFresh(t *testing.T) {
	headers := map[string][]string{"Content-Type": {"text/html; charset=utf-8"}, "Cache-Control": {"no-store"}}
	for k, v := range dashboardSecurityHeaders {
		headers[k] = v
	}
	fresh := OKEnvelope(managementResponse{StatusCode: 200, Headers: headers, Body: []byte(DashboardHTML)})
	got := handleGetDashboard()
	if !bytes.Equal(got, fresh) {
		t.Fatal("cached dashboard envelope differs from a freshly built one")
	}
	if &handleGetDashboard()[0] != &got[0] {
		t.Fatal("dashboard envelope not reused")
	}
}
