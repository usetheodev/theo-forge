package expr

import (
	"strings"
	"testing"
)

// TestExpr_Smoke confirms the package compiles and basic Expr construction works.
// Real coverage lives in dedicated test functions added by Phase 7 (T7.1).
func TestExpr_Smoke(t *testing.T) {
	e := C("x")
	if e.repr == "" {
		t.Fatal("C(\"x\") produced an Expr with empty repr")
	}
	if !strings.Contains(e.repr, "x") {
		t.Fatalf("C(\"x\") repr = %q, expected to contain \"x\"", e.repr)
	}
}
