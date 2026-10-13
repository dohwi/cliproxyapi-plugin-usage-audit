# Storage and query structural hardening

## Contracts

- Records pagination is descending `(effective timestamp, append sequence)` keyset pagination. Sequence is a 1-based retained append ordinal scoped to a Store generation; existing JSONL does not require migration. The first page fixes the sequence frontier. Later arrivals, including backdated completions and timestamp ties, are excluded from that traversal.
- Cursor v1 includes normalized filter/scope fingerprint, generation, frontier, last timestamp and sequence. Canonical JSON and a process-local random HMAC-SHA256 key authenticate it. Length, encoding, version, fields, boundary and anchor are validated. Legacy base64 offsets and altered filters return **400**, not reinterpretation. Clear, rotation, detected external changes or re-registration invalidate the generation (**409**, restart without cursor); process restart invalidates the HMAC (**400**).
- Store keeps one decoded cache with an estimated payload cap of **32 MiB** (record structs plus strings). Caller-owned slice copies protect it from mutation. File identity, size and mtime detect external changes. Append invalidates the cache; flush alone does not invalidate a generation. Summary followed by records without intervening append decodes once. Cache counters are exposed for measurement.
- DELETE advances an epoch under the same mutex as consumer commits, then clears storage. Observers capture the epoch before decoding. Queued and already-dequeued old entries cannot commit after the clear, including observers still decoding or waiting to enqueue. Epoch advances even if filesystem deletion fails; the API still returns 500 on partial deletion. Post-boundary observers remain writable. Epoch-rejected entries are intentional deletions, not queue overflow drops.
- Rotation prepares an append-only replacement descriptor, stages existing regular files in the same directory, installs the new namespace and removes oldest history only after all renames succeed. Failures reverse all completed moves. Failed rollback closes the Store and leaves staged files for operator recovery. Tests inject every rename stage, replacement creation, retention removal and rollback failure.
- Summary and records expose `storage_health`: write/flush/rotation/decode/read failures, corrupt/truncated line counts, decode mode/degraded state, snapshot decode/cache-hit counts, last error category and timestamp. The operational default is resilient JSONL decoding: every malformed non-empty line is skipped and later valid records continue across active and rotated files. An unterminated malformed final line is skipped and counted in both `corrupt_lines` and `truncated_lines`; a valid final object without a newline remains valid. Administrative strict mode is opt-in with `USAGE_AUDIT_STRICT_DECODING=true` and fails on any malformed line. Error details, filesystem paths, line numbers and raw bytes are never placed in health, logs, or API responses. Legacy `api_key_id` is stripped from live snapshots; the existing append and resource field allowlists remain in force.
- Resource summary defaults to **30 days**, allows at most **366 days**, and returns **400** when its combined summary/daily/provider/endpoint/model bucket count exceeds **4096**. Explicit historical ranges remain supported. Resource records retain limits **1..100**; management records retain **1..500**. Inverted ranges return 400. Management summary retains all-history compatibility. Dashboard today/30-day requests remain supported; dashboard files were not modified.

## Deliberate limits / residual risks

- This is a bounded retained cache, not an index. Cache misses still decode complete retained JSONL and filters still sort copies. The 32 MiB cap does **not** cap transient decoding or concurrent response allocations. Histories above the cap are not cached. Continuous append invalidates caching and can still cause repeated scans. Incremental indexing / bounded streaming summaries / host-level rate limiting are follow-up work, not implied guarantees.
- File metadata detection assumes a single writer. It is not cross-process locking and cannot detect an adversarial edit that preserves identity, size and mtime, or prevent all races with an unsupported external writer.
- Rotation rollback protects ordinary filesystem-operation failures, not abrupt process/power loss. No fsync/journal or automatic recovery of `.usage-rotation-*` directories is implemented. Failed rollback requires operator inspection; staged data is deliberately not removed. A successful rotation expires existing cursors even if all records relevant to that cursor remain retained.
- Clear remains a potentially partial filesystem deletion on failure; failure is explicit. The epoch barrier prevents queued pre-clear records from reappearing, not records deliberately submitted again after the boundary.
- Existing filesystem calls have no cancellable OS-I/O timeout. Test/subprocess execution uses explicit outer timeouts. Shutdown retains the existing bounded wait behavior.

## Verification

Executed in this module:

```sh
gofmt -w <changed Go files>
go test -race ./...
go vet ./...
go build -buildmode=c-shared -o /tmp/usage-audit-structural.so .
cc -Wall -Wextra -O2 -o /tmp/usage-audit-native-smoke verification/native_smoke.c -ldl
USAGE_AUDIT_DATA_DIR=<temporary directory> timeout 20s /tmp/usage-audit-native-smoke /tmp/usage-audit-structural.so
```

Race tests and vet passed; native ABI register/usage/summary/records/shutdown smoke printed `NATIVE_SMOKE_PASS`. Smoke used a disposable directory, not production data.
