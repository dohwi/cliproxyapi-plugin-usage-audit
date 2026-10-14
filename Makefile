SHELL := /bin/bash

# Library extension and loader flags follow the build host (GOOS). Override with
# `make build GOOS=...` only for a matching native toolchain: the plugin is a CGO
# c-shared library and must be built natively per platform.
GOOS ?= $(shell go env GOOS)
ifeq ($(GOOS),windows)
LIB_EXT := dll
SMOKE_EXE := .exe
SMOKE_LIBS :=
else ifeq ($(GOOS),darwin)
LIB_EXT := dylib
SMOKE_EXE :=
SMOKE_LIBS :=
else
LIB_EXT := so
SMOKE_EXE :=
SMOKE_LIBS := -ldl
endif
LIB := artifacts/usage-audit.$(LIB_EXT)
SMOKE := artifacts/native-smoke$(SMOKE_EXE)
SHA256 := $(shell command -v sha256sum >/dev/null 2>&1 && echo sha256sum || echo "shasum -a 256")

.PHONY: verify-i18n release-check verify verify-sanitize build native-smoke audit-dry-run sanitize-dry-run boundary-sanitize-dry-run verify-layout clean

verify:
	go test -race -count=1 -timeout=120s ./...
	go vet ./...

verify-sanitize:
	go test -race -count=1 -timeout=120s ./cmd/usage-audit-sanitize/...
	go vet ./cmd/usage-audit-sanitize/...

build:
	mkdir -p artifacts
	go build -trimpath -ldflags='-s -w' -buildmode=c-shared -o $(LIB) .
	$(SHA256) $(LIB)

native-smoke:
	scripts/build-native.sh


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

verify-i18n:
	node scripts/test-usage-dashboard-language.cjs

# Usage: make release-check TAG=v0.1.0
release-check:
	./scripts/check-release-version.sh $(TAG)

clean:
	rm -rf artifacts dist
	rm -rf scripts/.bin
