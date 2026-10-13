SHELL := /bin/bash

.PHONY: verify verify-sanitize build native-smoke audit-dry-run sanitize-dry-run boundary-sanitize-dry-run verify-layout clean

verify:
	go test -race -count=1 -timeout=120s ./...
	go vet ./...

verify-sanitize:
	go test -race -count=1 -timeout=120s ./cmd/usage-audit-sanitize/...
	go vet ./cmd/usage-audit-sanitize/...

build:
	mkdir -p artifacts
	go build -trimpath -ldflags='-s -w' -buildmode=c-shared -o artifacts/usage-audit.so .
	sha256sum artifacts/usage-audit.so

native-smoke: build
	gcc -O2 -Wall -Wextra -Werror -o /tmp/usage-audit-native-smoke verification/native_smoke.c -ldl
	data_dir="$$(mktemp -d)"; trap 'rm -rf "$$data_dir"' EXIT; USAGE_AUDIT_DATA_DIR="$$data_dir" /tmp/usage-audit-native-smoke artifacts/usage-audit.so
	rm -f /tmp/usage-audit-native-smoke

audit-dry-run sanitize-dry-run:
	@if [ -n "$$USAGE_LOG_PATH" ]; then \
		./scripts/audit-usage-logs.sh --file "$$USAGE_LOG_PATH"; \
	elif [ -n "$$USAGE_LOG_DIR" ]; then \
		./scripts/audit-usage-logs.sh --dir "$$USAGE_LOG_DIR"; \
	elif [ -f "testdata/usage-sample.jsonl" ]; then \
		echo "Notice: No USAGE_LOG_PATH specified; running demonstration dry-run against testdata/usage-sample.jsonl"; \
		echo "To audit an operational log, run: make audit-dry-run USAGE_LOG_PATH=/path/to/usage.jsonl"; \
		echo ""; \
		./scripts/audit-usage-logs.sh --file testdata/usage-sample.jsonl; \
	else \
		./scripts/audit-usage-logs.sh --help; \
	fi

boundary-sanitize-dry-run:
	@if [ -n "$$USAGE_LOG_PATH" ]; then \
		./scripts/service-stop-sanitize-boundary.sh --dry-run --file "$$USAGE_LOG_PATH"; \
	elif [ -f "testdata/usage-sample.jsonl" ]; then \
		echo "Notice: Running service-stop boundary dry-run against testdata/usage-sample.jsonl"; \
		./scripts/service-stop-sanitize-boundary.sh --dry-run --file testdata/usage-sample.jsonl; \
	else \
		./scripts/service-stop-sanitize-boundary.sh --help; \
	fi

verify-layout:
	node scripts/test-usage-dashboard-layout.cjs

clean:
	rm -rf artifacts
	rm -rf scripts/.bin
	rm -f /tmp/usage-audit-native-smoke
