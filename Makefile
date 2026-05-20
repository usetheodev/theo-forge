# theo-forge — Makefile
#
# Single source of truth for local + CI verification. CI workflows in
# .github/workflows/ delegate to `make verify` rather than re-defining
# commands so behavior is identical end-to-end.
#
# Authority:
#   - docs/QUALITY-GATES.md (gate definitions + rationale)
#   - .claude/rules/{testing,golang-conventions,error-handling}.md
#
# Tool versions (pin everything for reproducibility).
GOLANGCI_LINT_VERSION   ?= v1.61.0
GOVULNCHECK_VERSION     ?= latest
OSV_SCANNER_VERSION     ?= latest
NANCY_VERSION           ?= latest
GO_TEST_COVERAGE_VERSION ?= latest

# Default target prints help.
.DEFAULT_GOAL := help

# Phony targets (no actual files).
.PHONY: help install-tools verify verify-fmt verify-vet verify-lint \
        verify-test verify-race verify-coverage verify-sec \
        verify-vuln verify-osv verify-nancy verify-build \
        test test-race coverage coverage-html clean

# ---------------------------------------------------------------------------
# Help
# ---------------------------------------------------------------------------
help: ## Print this help.
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / { printf "\033[36m%-25s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

# ---------------------------------------------------------------------------
# Tooling
# ---------------------------------------------------------------------------
install-tools: ## Install all required dev tools into $GOPATH/bin.
	@echo "==> installing golangci-lint $(GOLANGCI_LINT_VERSION)"
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | \
		sh -s -- -b $$(go env GOPATH)/bin $(GOLANGCI_LINT_VERSION)
	@echo "==> installing govulncheck $(GOVULNCHECK_VERSION)"
	@go install golang.org/x/vuln/cmd/govulncheck@$(GOVULNCHECK_VERSION)
	@echo "==> installing osv-scanner $(OSV_SCANNER_VERSION)"
	@go install github.com/google/osv-scanner/cmd/osv-scanner@$(OSV_SCANNER_VERSION)
	@echo "==> installing nancy $(NANCY_VERSION)"
	@go install github.com/sonatype-nexus-community/nancy@$(NANCY_VERSION)
	@echo "==> installing go-test-coverage $(GO_TEST_COVERAGE_VERSION)"
	@go install github.com/vladopajic/go-test-coverage/v2@$(GO_TEST_COVERAGE_VERSION)
	@echo "==> all tools installed in $$(go env GOPATH)/bin"

# ---------------------------------------------------------------------------
# Aggregate verify gate (the rigorous quality gate).
# Order matters: cheap checks first, expensive scans last.
# Any failure stops the chain (-` after &&` semantics).
# ---------------------------------------------------------------------------
verify: verify-fmt verify-vet verify-lint verify-build verify-race verify-coverage verify-sec ## Run ALL quality gates.
	@echo "==> ✅ verify: ALL QUALITY GATES PASSED"

# ---------------------------------------------------------------------------
# Individual gates
# ---------------------------------------------------------------------------
verify-fmt: ## gofmt MUST produce no output.
	@echo "==> verify-fmt"
	@out=$$(gofmt -l .); \
	if [ -n "$$out" ]; then \
		echo "::error::gofmt found unformatted files:"; \
		echo "$$out"; \
		exit 1; \
	fi
	@echo "    OK (gofmt -l . empty)"

verify-vet: ## go vet MUST be clean.
	@echo "==> verify-vet"
	@go vet ./...
	@echo "    OK"

verify-lint: ## golangci-lint MUST be clean (zero warnings, strict config).
	@echo "==> verify-lint (golangci-lint $(GOLANGCI_LINT_VERSION))"
	@command -v golangci-lint >/dev/null 2>&1 || { echo "::error::golangci-lint not installed. Run: make install-tools"; exit 1; }
	@golangci-lint run ./...
	@echo "    OK"

verify-build: ## go build MUST succeed for every package.
	@echo "==> verify-build"
	@go build ./...
	@echo "    OK"

verify-race: ## go test -race MUST pass on every package (data-race detection gate).
	@echo "==> verify-race"
	@go test -race -count=1 ./...
	@echo "    OK"

verify-coverage: coverage.out ## Per-package coverage MUST satisfy .testcoverage.yml.
	@echo "==> verify-coverage (.testcoverage.yml)"
	@command -v go-test-coverage >/dev/null 2>&1 || { echo "::error::go-test-coverage not installed. Run: make install-tools"; exit 1; }
	@go-test-coverage --config=.testcoverage.yml

verify-sec: verify-vuln verify-osv verify-nancy ## All vulnerability scanners MUST report zero HIGH/CRITICAL.

verify-vuln: ## govulncheck (Go stdlib + module CVEs against go.sum).
	@echo "==> verify-vuln (govulncheck)"
	@command -v govulncheck >/dev/null 2>&1 || { echo "::error::govulncheck not installed. Run: make install-tools"; exit 1; }
	@govulncheck ./...
	@echo "    OK"

verify-osv: ## osv-scanner (Google OSV.dev DB; broader than govulncheck for non-Go transitive deps).
	@echo "==> verify-osv (osv-scanner)"
	@command -v osv-scanner >/dev/null 2>&1 || { echo "::error::osv-scanner not installed. Run: make install-tools"; exit 1; }
	@osv-scanner --lockfile=go.mod
	@echo "    OK"

verify-nancy: ## nancy (Sonatype OSS Index). Skipped if NANCY_USERNAME/NANCY_TOKEN not set.
	@echo "==> verify-nancy (nancy)"
	@command -v nancy >/dev/null 2>&1 || { echo "::error::nancy not installed. Run: make install-tools"; exit 1; }
	@bash -c 'if [ -z "$$NANCY_USERNAME" ] || [ -z "$$NANCY_TOKEN" ]; then \
		echo "    SKIP (nancy needs NANCY_USERNAME + NANCY_TOKEN; free signup: https://ossindex.sonatype.org/)"; \
		exit 0; \
	fi; \
	go list -json -deps ./... | nancy sleuth --quiet --username "$$NANCY_USERNAME" --token "$$NANCY_TOKEN" && echo "    OK"'

# ---------------------------------------------------------------------------
# Coverage artifact (consumed by verify-coverage)
# ---------------------------------------------------------------------------
coverage.out: ## Generate atomic-mode coverage profile.
	@echo "==> generating coverage profile"
	@go test -race -count=1 -coverprofile=coverage.out -covermode=atomic ./...

coverage: coverage.out ## Print per-package coverage summary.
	@go tool cover -func=coverage.out | tail -20

coverage-html: coverage.out ## Open per-line coverage in browser.
	@go tool cover -html=coverage.out

# ---------------------------------------------------------------------------
# Convenience aliases
# ---------------------------------------------------------------------------
test: ## Quick test run (no race detection, no coverage).
	@go test -count=1 ./...

test-race: verify-race ## Alias for verify-race.

clean: ## Remove generated artifacts.
	@rm -f coverage.out coverage.html sbom.cdx.json
	@echo "==> cleaned"
