#!/usr/bin/env bash
# Non-destructive inspection and sanitization runner for usage-audit JSONL logs.
# Enforces safe dry-run by default; requires --rewrite AND --backup-dir to modify.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

BIN="${REPO_ROOT}/scripts/.bin/usage-audit-sanitize"
MAIN_GO="${REPO_ROOT}/cmd/usage-audit-sanitize/main.go"

if [ ! -f "${BIN}" ] || [ "${BIN}" -ot "${MAIN_GO}" ]; then
  mkdir -p "$(dirname "${BIN}")"
  go -C "${REPO_ROOT}" build -trimpath -ldflags='-s -w' -o "${BIN}" ./cmd/usage-audit-sanitize
fi

exec "${BIN}" "$@"
