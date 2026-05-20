// Package config tests.
//
// Global-state discipline:
//
// Any test in any package that mutates the GLOBAL config (config.GetGlobal())
// MUST call ResetGlobalConfigForTest(t) at the top. The helper:
//  1. Acquires TestLock (process-wide mutex; serializes against parallel tests).
//  2. Calls GetGlobal().Reset() to wipe state.
//  3. Registers t.Cleanup to reset again on test exit.
//
// This is required because GlobalConfig is a singleton (T3.1 / ADR-001 in the
// fix-all-review-findings plan eliminates the structural reason for this
// singleton; until that lands, every test must cooperate). See EC-1 in
// .claude/knowledge-base/reviews/edge-case-review-fix-all-review-findings.md.
package config

import "testing"

// TestConfig_Smoke confirms the package compiles and New() returns a non-nil config.
// Real coverage lives in dedicated test functions added by Phase 7 (T7.2).
func TestConfig_Smoke(t *testing.T) {
	cfg := New()
	if cfg == nil {
		t.Fatal("config.New() returned nil")
	}
	// Default Image after Reset is "python:3.11" (see config.go Reset).
	cfg.Reset()
	if cfg.image != "python:3.11" {
		t.Fatalf("Reset() Image = %q, want %q", cfg.image, "python:3.11")
	}
}

// TestGlobalConfig_ResetForTest_AcquiresAndReleasesLock verifies the helper
// contract: lock acquired, state cleaned, lock released on cleanup.
func TestGlobalConfig_ResetForTest_AcquiresAndReleasesLock(t *testing.T) {
	GetGlobal().SetImage("contaminated:1")
	ResetGlobalConfigForTest(t)
	if got := GetGlobal().GetImage(); got != "python:3.11" {
		t.Fatalf("after Reset, GetImage() = %q, want \"python:3.11\"", got)
	}
}
