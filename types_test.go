package forge

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"sigs.k8s.io/yaml"
)

// Phase 1 / T1.2 — Resources presets.

// presetFactories holds the 3 presets so table-driven tests reuse the same set.
var presetFactories = []struct {
	name       string
	make       func() *ResourceRequirements
	goldenFile string
}{
	{"tiny", ResourcesTiny, "resources-tiny.golden.yaml"},
	{"small", ResourcesSmall, "resources-small.golden.yaml"},
	{"medium", ResourcesMedium, "resources-medium.golden.yaml"},
}

func TestResources_AllocatesFresh(t *testing.T) {
	for _, tc := range presetFactories {
		t.Run(tc.name, func(t *testing.T) {
			a := tc.make()
			b := tc.make()
			if a == b {
				t.Errorf("%s: two calls returned the SAME pointer — aliasing risk", tc.name)
			}
		})
	}
}

// EC-3 — mutation in one instance MUST NOT affect another.
func TestResourcesTiny_MutationIsolation(t *testing.T) {
	a := ResourcesTiny()
	b := ResourcesTiny()
	a.Requests.CPU = "999m"
	a.Limits.Memory = "999Gi"
	if b.Requests.CPU == "999m" {
		t.Errorf("mutation on a leaked into b.Requests.CPU = %q", b.Requests.CPU)
	}
	if b.Limits.Memory == "999Gi" {
		t.Errorf("mutation on a leaked into b.Limits.Memory = %q", b.Limits.Memory)
	}
}

func TestResources_RequestLessThanLimit(t *testing.T) {
	// Cheap sanity check on the preset numbers: requests <= limits.
	// We do byte compare on the suffix-stripped numerals; this is
	// adequate for the 3 hardcoded presets here.
	cases := []struct {
		name                   string
		r                      *ResourceRequirements
		wantReqCPU, wantLimCPU string
		wantReqMem, wantLimMem string
	}{
		{"tiny", ResourcesTiny(), "50m", "100m", "32Mi", "64Mi"},
		{"small", ResourcesSmall(), "100m", "500m", "128Mi", "256Mi"},
		{"medium", ResourcesMedium(), "500m", "2000m", "512Mi", "2Gi"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.r.Requests.CPU != tc.wantReqCPU {
				t.Errorf("Requests.CPU = %q, want %q", tc.r.Requests.CPU, tc.wantReqCPU)
			}
			if tc.r.Limits.CPU != tc.wantLimCPU {
				t.Errorf("Limits.CPU = %q, want %q", tc.r.Limits.CPU, tc.wantLimCPU)
			}
			if tc.r.Requests.Memory != tc.wantReqMem {
				t.Errorf("Requests.Memory = %q, want %q", tc.r.Requests.Memory, tc.wantReqMem)
			}
			if tc.r.Limits.Memory != tc.wantLimMem {
				t.Errorf("Limits.Memory = %q, want %q", tc.r.Limits.Memory, tc.wantLimMem)
			}
		})
	}
}

func TestResources_GoldenYAML(t *testing.T) {
	for _, tc := range presetFactories {
		t.Run(tc.name, func(t *testing.T) {
			got, err := yaml.Marshal(tc.make())
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			goldenPath := filepath.Join("testdata", tc.goldenFile)
			want, err := os.ReadFile(goldenPath) //nolint:gosec // test fixture path is constant
			if err != nil {
				t.Fatalf("read golden %s: %v (run with -update-golden to create)", goldenPath, err)
			}
			if strings.TrimRight(string(got), "\n") != strings.TrimRight(string(want), "\n") {
				t.Errorf("%s: golden mismatch\n=== got ===\n%s\n=== want ===\n%s",
					tc.name, got, want)
			}
		})
	}
}
