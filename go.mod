module github.com/usetheodev/theo-forge

// govulncheck (QG-6, 2026-05-20) detected 18 stdlib CVEs in Go 1.25.0;
// the highest fixed-in version observed across the report is 1.25.10
// (net/http GO-2026-4918, net GO-2026-4971). Require 1.25.10 to close all
// CVEs reachable through theo-forge's call graph.
go 1.25.10

require sigs.k8s.io/yaml v1.6.0

require (
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/kr/pretty v0.3.1 // indirect
	github.com/rogpeppe/go-internal v1.14.1 // indirect
	go.yaml.in/yaml/v2 v2.4.3 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
)
