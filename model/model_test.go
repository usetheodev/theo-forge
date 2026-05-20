package model

import (
	"encoding/json"
	"testing"
)

// TestModel_Smoke confirms a zero-value WorkflowModel marshals to non-empty JSON.
// Real coverage (tag fidelity, IntOrString round-trip) lands in Phase 7 (T7.5).
func TestModel_Smoke(t *testing.T) {
	b, err := json.Marshal(WorkflowModel{})
	if err != nil {
		t.Fatalf("json.Marshal(WorkflowModel{}) = %v", err)
	}
	if string(b) == "{}" {
		t.Fatalf("json.Marshal(WorkflowModel{}) = %q, want non-empty struct dump", string(b))
	}
}
