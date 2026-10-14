#!/usr/bin/env bash
# service-stop-sanitize-boundary.sh: Safe service-stop boundary runner for operational usage logs.
#
# Operational Safety Invariants:
#   1. Automatic cron deployment / operational execution is STRICTLY FORBIDDEN.
#      Unattended execution can disrupt service availability or cause unmonitored failure.
#   2. Modifying active logs while cliproxyapi writer is running is STRICTLY PROHIBITED.
#      Concurrent writes cause Linux inode desynchronization (writes to unlinked file)
#      and partial line write race conditions.
#   3. Rewrites MUST take place strictly within a verified service-stop boundary.
#   4. Restrictive verified backups (mode 0700) and SHA-256 pre-verification are mandatory.
#   5. On any failure during maintenance, immediate rollback and service restart are guaranteed.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

# No built-in default: the data directory depends on the deployment (USAGE_AUDIT_DATA_DIR, or the
# per-user default <user-config-dir>/cliproxyapi/usage-audit). Pass --file or set USAGE_LOG_PATH.
TARGET_FILE="${USAGE_LOG_PATH:-}"
BACKUP_ROOT="${BACKUP_ROOT:-/var/backups/cliproxyapi}"
CONTAINER_NAME="${CONTAINER_NAME:-cliproxyapi}"
HEALTH_URL="${HEALTH_URL:-http://127.0.0.1:8317/v0/management/plugins/usage-audit/dashboard}"

DRY_RUN=0
OPERATOR_CONFIRM=0
SKIP_CONTAINER_CHECK=0

usage() {
  cat <<EOF
Usage: $0 [options]

Service-Stop Boundary Operational Log Sanitizer Template

Options:
  --file <path>              Target JSONL log file (required unless USAGE_LOG_PATH is set)
  --backup-root <dir>        Backup directory base (default: ${BACKUP_ROOT})
  --container <name>         Container name to stop/start (default: ${CONTAINER_NAME})
  --dry-run, -n              Simulate the entire boundary sequence without stopping services or rewriting
  --operator-confirm         Mandatory confirmation flag for real execution
  --skip-container-check     Skip docker container stop/start (for local testing/simulation only)
  --health-url <url>         Health check URL after restart (default: ${HEALTH_URL})
  --help, -h                 Show this help message

CRITICAL SAFETY POLICY:
  - AUTOMATED CRON DEPLOYMENT / EXECUTION IS STRICTLY PROHIBITED.
  - Rewrites require stopping the cliproxyapi writer to prevent inode unlinking and data loss.
  - Requires manual operator execution during an approved maintenance window.
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --file)
      TARGET_FILE="$2"
      shift 2
      ;;
    --backup-root)
      BACKUP_ROOT="$2"
      shift 2
      ;;
    --container)
      CONTAINER_NAME="$2"
      shift 2
      ;;
    --dry-run|-n)
      DRY_RUN=1
      shift
      ;;
    --operator-confirm)
      OPERATOR_CONFIRM=1
      shift
      ;;
    --skip-container-check)
      SKIP_CONTAINER_CHECK=1
      shift
      ;;
    --health-url)
      HEALTH_URL="$2"
      shift 2
      ;;
    --help|-h)
      usage
      exit 0
      ;;
    *)
      echo "ERROR: Unknown option: $1" >&2
      usage >&2
      exit 2
      ;;
  esac
done

if [ -z "${TARGET_FILE}" ]; then
  echo "ERROR: no log file given. Pass --file <path> or set USAGE_LOG_PATH" >&2
  echo "       (the plugin writes <data dir>/usage.jsonl; see README, USAGE_AUDIT_DATA_DIR)." >&2
  exit 2
fi

# -----------------------------------------------------------------------------
# Anti-Cron & Operator Guard Enforcement
# -----------------------------------------------------------------------------
if [ "$DRY_RUN" -ne 1 ] && [ "$OPERATOR_CONFIRM" -ne 1 ]; then
  cat <<'ERR' >&2
================================================================================
CRITICAL SAFETY GUARD: OPERATOR CONFIRMATION REQUIRED
================================================================================
Automatic cron execution or unattended execution is STRICTLY PROHIBITED.
Rewriting operational logs while the writer daemon runs causes catastrophic
data loss (writes redirected to unlinked inode) or corrupted JSON lines.

To execute this service-stop boundary procedure:
  1. Schedule an approved maintenance window.
  2. Execute manually from an interactive operator terminal.
  3. Pass the explicit flag: --operator-confirm

For read-only inspection without service interruption, use:
  usage-audit/scripts/audit-usage-logs.sh --file /path/to/usage.jsonl
================================================================================
ERR
  exit 1
fi

# Real rewrites require an attached operator terminal. Supplying the confirmation
# flag from cron/systemd is not sufficient authorization.
if [ "$DRY_RUN" -ne 1 ] && { [ ! -t 0 ] || [ ! -t 1 ]; }; then
  echo "ERROR: Live sanitization requires an interactive operator terminal." >&2
  echo "Use --dry-run for unattended/read-only verification." >&2
  exit 1
fi

# -----------------------------------------------------------------------------
# Single-Instance Concurrency Lock
# -----------------------------------------------------------------------------
LOCK_FILE="/tmp/cliproxyapi-sanitize-boundary.lock"
exec 200>"${LOCK_FILE}"
if ! flock -n 200; then
  echo "ERROR: Another sanitization or maintenance process holds ${LOCK_FILE}. Exiting." >&2
  exit 1
fi

echo "================================================================="
echo " CLIProxyAPI Service-Stop Boundary Sanitizer"
echo "================================================================="
echo "Mode:              $([ "$DRY_RUN" -eq 1 ] && echo "DRY-RUN (simulation only)" || echo "LIVE BOUNDARY EXECUTION")"
echo "Target Log:        ${TARGET_FILE}"
echo "Backup Base:       ${BACKUP_ROOT}"
echo "Container:         ${CONTAINER_NAME}"
echo "Container Action:  $([ "$SKIP_CONTAINER_CHECK" -eq 1 ] && echo "SKIPPED (testing flag)" || echo "STOP -> SANITIZE -> START")"
echo "WARP Sidecars:     UNTOUCHED (isolated network sidecars unaffected)"
echo "Cron Automation:   STRICTLY FORBIDDEN"
echo "================================================================="

# Ensure sanitizer binary is ready
SANITIZER_BIN="${REPO_ROOT}/scripts/.bin/usage-audit-sanitize"
MAIN_GO="${REPO_ROOT}/cmd/usage-audit-sanitize/main.go"
if [ ! -f "${SANITIZER_BIN}" ] || [ "${SANITIZER_BIN}" -ot "${MAIN_GO}" ]; then
  echo "[Prep] Compiling usage-audit-sanitize tool..."
  mkdir -p "$(dirname "${SANITIZER_BIN}")"
  go -C "${REPO_ROOT}" build -trimpath -ldflags='-s -w' -o "${SANITIZER_BIN}" ./cmd/usage-audit-sanitize
fi

# -----------------------------------------------------------------------------
# Phase 1: Preflight & Read-Only Baseline Inspection
# -----------------------------------------------------------------------------
echo "[1/6] Preflight inspection..."
if [ ! -f "${TARGET_FILE}" ]; then
  echo "ERROR: Target file does not exist: ${TARGET_FILE}" >&2
  exit 1
fi

TARGET_SIZE=$(stat -c %s "${TARGET_FILE}")
echo "  Target log size: ${TARGET_SIZE} bytes"

echo "[2/6] Running baseline dry-run inspection (read-only)..."
BASELINE_JSON=$("${SANITIZER_BIN}" --file "${TARGET_FILE}" --json)

# Parse summary fields using standard python3
BASELINE_EXPOSURES=$(python3 -c "import json, sys; d=json.loads(sys.argv[1]); print(d.get('exposed_records', 0))" "${BASELINE_JSON}")
BASELINE_TOTAL_LINES=$(python3 -c "import json, sys; d=json.loads(sys.argv[1]); print(d.get('total_lines', 0))" "${BASELINE_JSON}")
BASELINE_CORRUPTED=$(python3 -c "import json, sys; d=json.loads(sys.argv[1]); print(d.get('corrupted_lines', 0))" "${BASELINE_JSON}")
BASELINE_VALID=$(python3 -c "import json, sys; d=json.loads(sys.argv[1]); print(d.get('valid_records', 0))" "${BASELINE_JSON}")

echo "  Baseline: total_lines=${BASELINE_TOTAL_LINES}, valid=${BASELINE_VALID}, exposed=${BASELINE_EXPOSURES}, corrupted=${BASELINE_CORRUPTED}"

if [ "${BASELINE_EXPOSURES}" -eq 0 ]; then
  echo ""
  echo "✓ Notice: Target log file contains 0 api_key_id exposures."
  echo "  No sanitization rewrite needed. Service stop skipped. Exiting safely."
  exit 0
fi

TIMESTAMP=$(date -u +"%Y%m%dT%H%M%SZ")
BACKUP_DIR="${BACKUP_ROOT}/sanitizer-${TIMESTAMP}"

if [ "$DRY_RUN" -eq 1 ]; then
  echo ""
  echo "[DRY-RUN PLAN]"
  echo "  1. Would check free space in ${BACKUP_ROOT} (> ${TARGET_SIZE} bytes)"
  if [ "$SKIP_CONTAINER_CHECK" -ne 1 ]; then
    echo "  2. Would stop container '${CONTAINER_NAME}' (WARP sidecars untouched)"
    echo "  3. Would verify container '${CONTAINER_NAME}' has stopped"
  else
    echo "  2. Would skip container check (--skip-container-check active)"
  fi
  echo "  4. Would create restricted backup directory: ${BACKUP_DIR} (mode 0700)"
  echo "  5. Would create byte-verified backup of ${TARGET_FILE}"
  echo "  6. Would execute atomic rewrite removing api_key_id and preserving ${BASELINE_CORRUPTED} corrupted lines"
  echo "  7. Would verify post-sanitization: exposed_records=0, total_lines=${BASELINE_TOTAL_LINES}"
  if [ "$SKIP_CONTAINER_CHECK" -ne 1 ]; then
    echo "  8. Would restart container '${CONTAINER_NAME}'"
    echo "  9. Would verify health check: ${HEALTH_URL} HTTP 200"
  fi
  echo ""
  echo "✓ DRY-RUN SUCCESSFUL: Sequence simulated safely. No services stopped. No files modified."
  exit 0
fi

# -----------------------------------------------------------------------------
# Phase 2: Live Service-Stop Boundary Execution
# -----------------------------------------------------------------------------
SERVICE_WAS_STOPPED=0
ROLLBACK_TRIGGERED=0

cleanup_and_rollback() {
  local exit_code=$?
  if [ $exit_code -ne 0 ] && [ $ROLLBACK_TRIGGERED -eq 0 ]; then
    ROLLBACK_TRIGGERED=1
    echo ""
    echo "!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!"
    echo " EMERGENCY: An error occurred during service-stop sanitization."
    echo " Initiating immediate rollback..."
    echo "!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!"
    if [ -d "${BACKUP_DIR}" ]; then
      local base_name
      base_name="$(basename "${TARGET_FILE}")"
      if [ -f "${BACKUP_DIR}/${base_name}" ]; then
        echo " Restoring ${TARGET_FILE} from verified backup ${BACKUP_DIR}/${base_name}..."
        cp -pf "${BACKUP_DIR}/${base_name}" "${TARGET_FILE}"
      fi
    fi
    if [ "${SERVICE_WAS_STOPPED}" -eq 1 ] && [ "$SKIP_CONTAINER_CHECK" -ne 1 ]; then
      echo " Restarting container ${CONTAINER_NAME}..."
      docker start "${CONTAINER_NAME}" || true
    fi
    echo " Rollback completed. Check errors above."
  fi
}
trap cleanup_and_rollback EXIT

# Step A: Stop Container
if [ "$SKIP_CONTAINER_CHECK" -ne 1 ]; then
  echo "[3/6] Entering service-stop boundary: stopping container ${CONTAINER_NAME}..."
  if docker ps --format '{{.Names}}' | grep -wq "^${CONTAINER_NAME}$"; then
    docker stop -t 15 "${CONTAINER_NAME}"
    SERVICE_WAS_STOPPED=1
    echo "  Container ${CONTAINER_NAME} stopped successfully."
  else
    echo "  Container ${CONTAINER_NAME} is not currently running."
  fi

  # Verify stopped
  if docker ps --format '{{.Names}}' | grep -wq "^${CONTAINER_NAME}$"; then
    echo "ERROR: Failed to stop container ${CONTAINER_NAME}; writer may still be active!" >&2
    exit 1
  fi
else
  echo "[3/6] Container stop skipped (--skip-container-check active)..."
fi

# Step B: Create Restricted Verified Backup
echo "[4/6] Creating secure verified backup directory..."
mkdir -p "${BACKUP_DIR}"
chmod 700 "${BACKUP_DIR}"

BASE_NAME="$(basename "${TARGET_FILE}")"
BACKUP_FILE="${BACKUP_DIR}/${BASE_NAME}"
echo "  Copying ${TARGET_FILE} -> ${BACKUP_FILE}..."
cp -p "${TARGET_FILE}" "${BACKUP_FILE}"

SRC_HASH=$(sha256sum "${TARGET_FILE}" | awk '{print $1}')
DST_HASH=$(sha256sum "${BACKUP_FILE}" | awk '{print $1}')
if [ "${SRC_HASH}" != "${DST_HASH}" ]; then
  echo "ERROR: Backup SHA-256 hash mismatch! Aborting rewrite." >&2
  exit 1
fi
echo "  Backup verified: SHA-256 ${DST_HASH}"

# Step C: Execute Sanitizer Rewrite
echo "[5/6] Executing atomic sanitization rewrite..."
"${SANITIZER_BIN}" --file "${TARGET_FILE}" --rewrite --backup-dir "${BACKUP_DIR}"

# Step D: Post-Sanitization Verification
echo "[6/6] Verifying sanitization results..."
POST_JSON=$("${SANITIZER_BIN}" --file "${TARGET_FILE}" --json)
POST_EXPOSURES=$(python3 -c "import json, sys; d=json.loads(sys.argv[1]); print(d.get('exposed_records', 0))" "${POST_JSON}")
POST_TOTAL_LINES=$(python3 -c "import json, sys; d=json.loads(sys.argv[1]); print(d.get('total_lines', 0))" "${POST_JSON}")
POST_CORRUPTED=$(python3 -c "import json, sys; d=json.loads(sys.argv[1]); print(d.get('corrupted_lines', 0))" "${POST_JSON}")

echo "  Post-sanitize stats: total_lines=${POST_TOTAL_LINES}, exposed=${POST_EXPOSURES}, corrupted=${POST_CORRUPTED}"

if [ "${POST_EXPOSURES}" -ne 0 ]; then
  echo "ERROR: Verification failed: exposed records remaining: ${POST_EXPOSURES}" >&2
  exit 1
fi

if [ "${POST_TOTAL_LINES}" -ne "${BASELINE_TOTAL_LINES}" ]; then
  echo "ERROR: Verification failed: total lines changed from ${BASELINE_TOTAL_LINES} to ${POST_TOTAL_LINES}" >&2
  exit 1
fi

if [ "${POST_CORRUPTED}" -ne "${BASELINE_CORRUPTED}" ]; then
  echo "ERROR: Verification failed: corrupted lines count changed from ${BASELINE_CORRUPTED} to ${POST_CORRUPTED}" >&2
  exit 1
fi

# Step E: Restart Service
if [ "${SERVICE_WAS_STOPPED}" -eq 1 ] && [ "$SKIP_CONTAINER_CHECK" -ne 1 ]; then
  echo "  Restarting container ${CONTAINER_NAME}..."
  docker start "${CONTAINER_NAME}"
  SERVICE_WAS_STOPPED=0

  echo "  Checking service health..."
  HEALTH_OK=0
  for i in {1..15}; do
    if curl -s -f -m 3 "${HEALTH_URL}" >/dev/null 2>&1; then
      HEALTH_OK=1
      echo "  Health check PASSED (attempt $i)."
      break
    fi
    sleep 1
  done

  if [ "${HEALTH_OK}" -ne 1 ]; then
    echo "ERROR: Health check did not return HTTP 200 within 15 seconds." >&2
    echo "Inspect container logs: docker logs --tail 50 ${CONTAINER_NAME}" >&2
    SERVICE_WAS_STOPPED=1
    exit 1
  fi
fi

# Clear rollback trap
trap - EXIT

echo ""
echo "================================================================="
echo "✓ SERVICE-STOP SANITIZATION COMPLETED SUCCESSFULLY"
echo "================================================================="
echo "Target File:         ${TARGET_FILE}"
echo "Lines Processed:     ${POST_TOTAL_LINES}"
echo "Exposures Removed:   ${BASELINE_EXPOSURES}"
echo "Corrupted Preserved: ${POST_CORRUPTED}"
echo "Secure Backup Path:  ${BACKUP_FILE}"
echo "Backup Directory:    ${BACKUP_DIR} (mode 0700)"
echo "WARP Sidecars:       UNTOUCHED"
echo "================================================================="
