# Security policy

## Deployment boundary

This is a native plugin loaded inside CLIProxyAPI, not an independently authenticated web service. The host process and its configuration are trusted. Keep plugin resource routes behind a trusted network or an authenticated reverse proxy; do not expose them directly to the public Internet.

The dashboard, resource summary and resource records do not require the management key. They intentionally expose model/provider usage, timing, aliases and endpoint labels. Management queries and DELETE depend on the host enforcing management authentication before dispatch. Sanitized resource responses are not anonymous; aliases or source labels supplied by the host may themselves identify a deployment.

## Data handling

Raw API keys, auth IDs, prompts and failure response bodies are not persisted by the usage observer. Session/auth indexes and request metadata are operational data and must be protected. Store files are created with mode `0640` and directories with mode `0750`; use a trusted local data directory and restrict its ownership/group. Do not run sanitizing rewrites against an active writer: follow `docs/sanitization-runbook.md`.

Never commit operational logs, auth files, tokens or real account identifiers. `testdata/usage-sample.jsonl` is synthetic. The dashboard persists only its language preference.

## Verification

Use the patched Go runtime pinned in `.mise.toml`. Run `make verify`, `make native-smoke`, `make verify-i18n`, `govulncheck ./...`, and `npm audit` in `ui/` before release. Check vendored component checksums and regenerate the checked-in dashboard after UI edits. These checks reduce risk but do not prove the absence of vulnerabilities.

## Reporting

Do not publish credentials or sensitive deployment details in issues. Use GitHub private vulnerability reporting when available, or contact the repository owner privately before sharing a security report.
