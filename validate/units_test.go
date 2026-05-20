package validate

import "testing"

// TestValidate_BinaryUnit_Smoke confirms a known-good binary unit parses.
// Real coverage (table-driven, edge cases, resource requirements) lands in Phase 7 (T7.4).
func TestValidate_BinaryUnit_Smoke(t *testing.T) {
	if err := BinaryUnit("1Gi"); err != nil {
		t.Fatalf("BinaryUnit(\"1Gi\") = %v, want nil", err)
	}
}

// TestValidate_DecimalUnit_Smoke confirms a known-good decimal unit parses.
func TestValidate_DecimalUnit_Smoke(t *testing.T) {
	if err := DecimalUnit("100m"); err != nil {
		t.Fatalf("DecimalUnit(\"100m\") = %v, want nil", err)
	}
}
