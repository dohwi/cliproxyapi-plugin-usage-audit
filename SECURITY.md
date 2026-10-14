# Security policy

## Deployment boundary

This is a native plugin loaded inside CLIProxyAPI, not an independently authenticated web service. The host process and its configuration are trusted.

**Route boundary.** CLIProxyAPI serves `/v0/resource/plugins/<id>/...` without authentication, so this plugin puts only static UI there: the dashboard page shell and its self-hosted fonts. All usage data is behind the host's management authentication under `/v0/management/plugins/usage-audit/...`:

| Route | Purpose |
|---|---|
| `GET /data/summary`, `GET /data/records` | Dashboard-shaped, bounded data (no `auth_index`) |
| `GET /summary`, `GET /records` | Management API (includes `auth_index`) |
| `DELETE /records` | Clear stored records |

The unauthenticated `/v0/resource/.../data/*` routes do not exist (HTTP 404). The static page therefore asks the operator for the management key; it is sent only as an `Authorization: Bearer` header (never in a URL), kept in the tab's `sessionStorage` after the host accepted it, and dropped on a 401/403. It is never written to `localStorage`, cookies or the server. Without a key the page sends no data request, because CLIProxyAPI counts every rejected management request toward a temporary IP ban. Usage data still exposes model/provider usage, timing, aliases and endpoint labels to anyone holding the key; aliases or source labels supplied by the host may themselves identify a deployment.

## Data handling

Raw API keys, auth IDs, prompts and failure response bodies are not persisted by the usage observer. Session/auth indexes and request metadata are operational data and must be protected. Store files are created with mode `0640` and directories with mode `0750`; use a trusted local data directory and restrict its ownership/group. Do not run sanitizing rewrites against an active writer: follow `docs/sanitization-runbook.md`.

Never commit operational logs, auth files, tokens or real account identifiers. `testdata/usage-sample.jsonl` is synthetic. The dashboard persists its language preference (`localStorage`) and, for the current tab only, the accepted management key (`sessionStorage`).

## Verification

Use the patched Go runtime pinned in `.mise.toml`. Run `make verify`, `make native-smoke`, `make verify-i18n`, `govulncheck ./...`, and `npm audit` in `ui/` before release; the Release workflow builds and smoke-tests every store platform natively. Check vendored component checksums and regenerate the checked-in dashboard after UI edits. These checks reduce risk but do not prove the absence of vulnerabilities.

## Reporting

Do not publish credentials or sensitive deployment details in issues. Use GitHub private vulnerability reporting when available, or contact the repository owner privately before sharing a security report.
