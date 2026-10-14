# Operational Runbook: Usage Log Sanitization & Service-Stop Boundary Management

**Document Version:** 2.0.0  
**Target Systems:** CLIProxyAPI `usage-audit` storage environments  
**Scope:** Remediation of `api_key_id` entries, concurrency protection, service-stop boundary enforcement, and rollback handling  
**Paths:** examples use `$USAGE_AUDIT_DATA_DIR`, the directory holding `usage.jsonl`. Set it to the directory the plugin actually uses: the `USAGE_AUDIT_DATA_DIR` value the host runs with, or the per-user default (`<user-config-dir>/cliproxyapi/usage-audit`, e.g. `~/.config/cliproxyapi/usage-audit` on Linux).  

---

## 1. Executive Summary & Operational Context

In earlier builds of `usage-audit`, incoming API key identifiers (`api_key_id`) were serialized to disk within JSONL records. Modern plugin releases have eliminated ingestion and retention of this field. However, historical active, rotated (`.1`, `.2`, ...), and archived log files may still contain raw keys or tokens.

### Current Operational Context (Post-v8.1 Rollback)
Following the rollback of `dashboard-hardening-v8.1` due to the empty quota accounts failure (`quota_accounts_empty`), production is currently running a legacy `usage-audit` binary (`ad076d7175c3452f376f118d966a56c23d170145bbb5dba60016e3b28646999a`).

**Operational Implication:**
Because the legacy binary is active, **incoming API requests may continue recording `api_key_id` into the active log (`usage.jsonl`)** until the hardened `v8.2` release is deployed.

This runbook defines the authoritative, non-destructive procedures for managing this operational state safely prior to `v8.2` redeployment without risking data loss, log truncation, or service outages.

---

## 2. Core Safety Invariants & Concurrency Protection

### A. Writer Concurrency Hazards (Why Rewriting Live Logs Is Forbidden)
The `cliproxyapi` daemon opens `usage.jsonl` with `O_APPEND|O_CREATE|O_WRONLY` and buffers writes via an internal 64 KiB `bufio.Writer`.

Rewriting or modifying an active log while `cliproxyapi` is running introduces severe hazards:

1. **Linux Inode Desynchronization (Silent Data Loss)**:
   The sanitizer uses atomic temp-file creation and `os.Rename`. When `os.Rename` replaces `usage.jsonl`, the directory entry points to a new inode. However, the running `cliproxyapi` process retains its open file descriptor pointing to the **original, now-unlinked inode**. All subsequent request logs written by `cliproxyapi` go into this unlinked inode and are permanently lost upon service termination. The new file receives zero subsequent logs.
2. **Buffer Flush Race & False Corruption**:
   If the sanitizer or reader reads `usage.jsonl` while the daemon is flushing a buffered line, it can encounter a partial JSON record at EOF. In rewrite mode, this could cause the partial record to be saved prematurely or corrupted.
3. **Dual-Writer Lock Collisions**:
   Concurrent modification violates single-writer expectations and can result in interleaved bytes or broken JSON objects.

**Mandatory Rule:** Active log files (`usage.jsonl`) **MUST NEVER be rewritten while `cliproxyapi` is running**. Rewriting is strictly restricted to a **service-stop boundary**.

---

### B. Strict Prohibition of Automated Cron Deployment / Execution
**Automatic cron deployment or scheduled operational execution is STRICTLY FORBIDDEN** (`자동 cron 배포/운영 실행 금지`).

Reasons:
- Automated scripts stopping the container can cause uncoordinated service outages.
- If a network, disk, or syntax error occurs during unattended execution, the container could remain stopped, causing prolonged downtime.
- Automated jobs cannot evaluate complex recovery situations (such as mid-log corrupted lines or filesystem full conditions).

**Policy:**
1. All sanitization rewrites must be executed manually by a human operator during an approved maintenance window.
2. Scripts must require explicit manual confirmation (`--operator-confirm`) and refuse unmonitored cron execution.
3. Periodic monitoring before `v8.2` redeployment must be **read-only inspection (dry-run)**.

---

### C. Non-Destructive Sanitizer Invariants
The underlying `usage-audit-sanitize` utility enforces:
1. **Dry-Run by Default**: Scans files and prints statistics; never alters or deletes files.
2. **Two-Key Modification Gate**: Modifying files requires `--rewrite` AND `--backup-dir <path>` simultaneously.
3. **Zero Secret Leakage**: Secret values and tokens are **never** output to stdout, stderr, logs, or reports.
4. **Byte-for-Byte Preservation of Corrupted Lines**:
   Non-JSON lines, truncated records, and malformed entries (such as historical lines 9451 and 10611) are preserved verbatim, preventing silent data truncation.
5. **Mode & Permission Preservation**: File permissions (`0640`) and ownership are preserved across rewrites.

---

## 3. Pre-Redeployment Management Procedures

### Procedure 1: Safe Non-Destructive Periodic Inspection (Read-Only)

Operators can inspect exposure levels periodically without service downtime.

#### A. Auditing Inactive/Rotated Logs (Zero Risk)
Rotated files (`usage.jsonl.1`, `usage.jsonl.2`, etc.) are closed and static. They can be safely audited or sanitized without stopping `cliproxyapi`:

```bash
# Read-only inspection of rotated logs
./usage-audit/scripts/audit-usage-logs.sh --file "$USAGE_AUDIT_DATA_DIR/usage.jsonl.1"
```

#### B. Auditing Active Log (Read-Only Dry-Run)
The active log can be audited in read-only mode while the service is running:

```bash
# Read-only dry-run audit of active log
./usage-audit/scripts/audit-usage-logs.sh --file "$USAGE_AUDIT_DATA_DIR/usage.jsonl"
```

*Note on Reading Active Logs:* Because the daemon writer is active, a transient "corrupted line" might rarely be reported if the audit reads exactly while a line is being flushed at EOF. If an EOF corrupted line is reported, rerun the dry-run command after a moment to confirm whether it was a transient buffer flush or an actual corrupt record.

---

### Procedure 2: Service-Stop Boundary Sanitization (Operator Executed)

When sanitizing active logs (or active + rotated logs together), the operator must use the service-stop boundary script template: `usage-audit/scripts/service-stop-sanitize-boundary.sh`.

#### Boundary Execution Sequence

```
[1. Baseline Dry-Run] --> [2. Stop cliproxyapi] --> [3. Verify Stopped]
                                                            |
[6. Restart cliproxyapi] <-- [5. Verify Post-Sanitize] <-- [4. Backup (0700) & Rewrite]
         |
[7. Health Smoke 200]
```

#### Step-by-Step Operator Instructions

1. **Obtain Maintenance Approval**:
   Coordinate with team for a brief maintenance window (typical downtime: 2–5 seconds).

2. **Run Dry-Run Simulation First**:
   Simulate the boundary process to verify file paths and baseline exposure counts without stopping services:
   ```bash
   ./usage-audit/scripts/service-stop-sanitize-boundary.sh \
     --file "$USAGE_AUDIT_DATA_DIR/usage.jsonl" \
     --dry-run
   ```
   If `exposed_records == 0`, no rewrite is necessary!

3. **Execute Live Service-Stop Boundary Sanitization**:
   Execute manually with explicit confirmation:
   ```bash
   ./usage-audit/scripts/service-stop-sanitize-boundary.sh \
     --file "$USAGE_AUDIT_DATA_DIR/usage.jsonl" \
     --backup-root /var/backups/cliproxyapi \
     --container cliproxyapi \
     --operator-confirm
   ```

   **What the script executes under the hood:**
   - Acquires flock (`/tmp/cliproxyapi-sanitize-boundary.lock`) to prevent concurrent executions.
   - Performs pre-stop baseline audit.
   - Stops `cliproxyapi` container (`docker stop -t 15 cliproxyapi`). Note: WARP sidecar containers remain **untouched**.
   - Confirms container is fully stopped (writer process terminated).
   - Creates restricted backup directory with permissions `0700` (`chmod 700`).
   - Copies operational log and verifies SHA-256 integrity before modifying anything.
   - Invokes sanitizer to rewrite `usage.jsonl`: removes `api_key_id`, preserves corrupted lines verbatim, fsyncs, and preserves `0640` file mode.
   - Runs post-sanitization verification: confirms `exposed_records == 0`, `total_lines` unchanged, `corrupted_lines` unchanged.
   - Restarts `cliproxyapi` container (`docker start cliproxyapi`).
   - Polls `/v0/management/plugins/usage-audit/dashboard` for HTTP 200 health check.
   - Traps any failure: if any step fails while the container is stopped or during rewrite, the script automatically restores the original file from the verified backup, restarts the container, and reports the error.

---

## 4. Rollback & Disaster Recovery Procedure

If an unexpected error occurs or the service fails health verification:

1. **Automatic Script Rollback**:
   The `service-stop-sanitize-boundary.sh` script installs an emergency trap that automatically restores original files from the verified backup directory if an unhandled error occurs before completion.

2. **Manual Emergency Rollback**:
   If manual recovery is ever needed:
   ```bash
   # Identify latest backup directory
   ls -ld /var/backups/cliproxyapi/sanitizer-*

   # Stop container to prevent write collision
   docker stop -t 15 cliproxyapi

   # Restore verified backup
   cp -pf /var/backups/cliproxyapi/sanitizer-<TIMESTAMP>/usage.jsonl \
          "$USAGE_AUDIT_DATA_DIR/usage.jsonl"

   # Verify restored permissions
   chmod 640 "$USAGE_AUDIT_DATA_DIR/usage.jsonl"

   # Restart container
   docker start cliproxyapi

   # Verify health
   curl -s -f http://127.0.0.1:8317/v0/management/plugins/usage-audit/dashboard && echo "Restored OK"
   ```

---

## 5. Pre-Redeployment Checklist for v8.2

Before promoting and deploying `v8.2`:

- [ ] **Operational Log Sanitization Completed**: Active log verified with 0 `api_key_id` exposures via service-stop boundary.
- [ ] **Corrupted Line Verification**: Historical corrupted lines (e.g. lines 9451, 10611) identified and verified to prevent store decode errors.
- [ ] **Quota Initial Collection Gate**: Verified that deployment verification enforces non-empty accounts and completed collection before live promotion.
- [ ] **Anti-Cron Confirmation**: Confirmed no automated cron jobs are installed in operational environments.
- [ ] **Staged Release Verified**: Full `make verify`, sanitizer verification, native smokes, and `make package-dry-run` passed bit-for-bit reproducibility.
