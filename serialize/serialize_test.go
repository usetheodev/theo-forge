package serialize

import (
	"strings"
	"testing"

	"github.com/usetheodev/theo-forge/model"
)

// TestSerialize_Smoke confirms WorkflowToYAML round-trips a zero-value WorkflowModel
// to non-empty YAML. Real coverage (round-trip, path traversal, file I/O) lands in
// Phase 7 (T7.3) and Phase 2 (T2.1).
func TestSerialize_Smoke(t *testing.T) {
	yaml, err := WorkflowToYAML(model.WorkflowModel{})
	if err != nil {
		t.Fatalf("WorkflowToYAML returned err: %v", err)
	}
	if yaml == "" {
		t.Fatal("WorkflowToYAML returned empty string for zero-value WorkflowModel")
	}
	if !strings.Contains(yaml, "metadata") && !strings.Contains(yaml, "spec") {
		t.Logf("yaml output: %q", yaml)
	}
}
