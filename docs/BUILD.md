# Build & Vendoring Notes

## Build prerequisites

- Go 1.25.0 or later (matches `go.mod`).
- `make` (optional; commands below use `go` directly).

## Standard build

```bash
go build ./...
go test -race -count=1 ./...
```

## Linting

```bash
gofmt -l .                        # must be empty
go vet ./...                      # must be clean
golangci-lint run ./...           # if installed
govulncheck ./...                 # if installed
```

CI enforces all four. See `.github/workflows/ci.yml`.

## Vendoring

theo-forge does NOT ship a `vendor/` directory. The default build path
uses `proxy.golang.org` (Go module proxy). This keeps the repository
small and the dependency tree auditable via `go.sum`.

### Air-gapped / no-proxy builds

If your environment cannot reach `proxy.golang.org`:

1. Use `GOFLAGS="-mod=mod"` to bypass the proxy.
2. Configure `GOPROXY` to point to your internal proxy:
   ```bash
   export GOPROXY=https://internal-proxy.corp.example/golang
   ```
3. Or generate a `vendor/` locally with `go mod vendor` — but DO NOT
   commit it; the repository policy is to remain vendorless.

(T5.7 / DEP-001 — closes the "no vendor + no explicit GONOSUMCHECK" gap by
documenting the intended posture.)

## Module verification

Before release, CI runs:

```bash
go mod verify    # checksums match downloaded modules
go mod tidy      # no unnecessary or missing entries
git diff --exit-code go.mod go.sum   # tidy was a no-op
```

The release workflow (`.github/workflows/release.yml`) also requires
`test`, `lint`, and `security` checks to pass before publishing
(T5.4 / INFRA-004).
