package config

import (
	"errors"
	"strings"
	"testing"

	"github.com/usetheodev/theo-forge/model"
)

// T8.2 — Composable hook system tests.

func TestRegisterNamedTemplateHook_DispatchesInOrder(t *testing.T) {
	cfg := New()
	var calls []string
	_, err := cfg.RegisterNamedTemplateHook("first", func(*model.TemplateModel) error {
		calls = append(calls, "first")
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	_, err = cfg.RegisterNamedTemplateHook("second", func(*model.TemplateModel) error {
		calls = append(calls, "second")
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := cfg.DispatchNamedTemplateHooks(&model.TemplateModel{}); err != nil {
		t.Fatal(err)
	}
	if len(calls) != 2 || calls[0] != "first" || calls[1] != "second" {
		t.Errorf("registration order not preserved: %v", calls)
	}
}

func TestRegisterNamedTemplateHook_RejectsDuplicate(t *testing.T) {
	cfg := New()
	if _, err := cfg.RegisterNamedTemplateHook("dup", func(*model.TemplateModel) error { return nil }); err != nil {
		t.Fatal(err)
	}
	_, err := cfg.RegisterNamedTemplateHook("dup", func(*model.TemplateModel) error { return nil })
	if err == nil || !strings.Contains(err.Error(), "already registered") {
		t.Errorf("expected duplicate error, got %v", err)
	}
}

func TestRegisterNamedTemplateHook_RejectsEmptyName(t *testing.T) {
	cfg := New()
	if _, err := cfg.RegisterNamedTemplateHook("", func(*model.TemplateModel) error { return nil }); err == nil {
		t.Errorf("expected empty-name error")
	}
}

func TestRegisterNamedTemplateHook_RejectsNilFn(t *testing.T) {
	cfg := New()
	if _, err := cfg.RegisterNamedTemplateHook("x", nil); err == nil {
		t.Errorf("expected nil-fn error")
	}
}

func TestRemoveTemplateHook_RemovesByID(t *testing.T) {
	cfg := New()
	id, _ := cfg.RegisterNamedTemplateHook("rm", func(*model.TemplateModel) error { return nil })
	if !cfg.RemoveTemplateHook(id) {
		t.Errorf("RemoveTemplateHook returned false for existing hook")
	}
	if cfg.RemoveTemplateHook(id) {
		t.Errorf("RemoveTemplateHook returned true for already-removed hook")
	}
}

func TestDispatchNamedTemplateHooks_ShortCircuitsOnError(t *testing.T) {
	cfg := New()
	hit := 0
	_, _ = cfg.RegisterNamedTemplateHook("first", func(*model.TemplateModel) error {
		hit++
		return errors.New("boom")
	})
	_, _ = cfg.RegisterNamedTemplateHook("second", func(*model.TemplateModel) error {
		hit++
		return nil
	})
	err := cfg.DispatchNamedTemplateHooks(&model.TemplateModel{})
	if err == nil || !strings.Contains(err.Error(), "first") {
		t.Errorf("expected wrap with 'first', got %v", err)
	}
	if hit != 1 {
		t.Errorf("expected short-circuit after 1 hook, hit count = %d", hit)
	}
}

func TestListNamedTemplateHooks(t *testing.T) {
	cfg := New()
	_, _ = cfg.RegisterNamedTemplateHook("a", func(*model.TemplateModel) error { return nil })
	_, _ = cfg.RegisterNamedTemplateHook("b", func(*model.TemplateModel) error { return nil })
	ids := cfg.ListNamedTemplateHooks()
	if len(ids) != 2 || ids[0] != "a" || ids[1] != "b" {
		t.Errorf("ListNamedTemplateHooks = %v", ids)
	}
}

// Workflow-level symmetrical tests.

func TestRegisterNamedWorkflowHook_RoundTrip(t *testing.T) {
	cfg := New()
	id, err := cfg.RegisterNamedWorkflowHook("wf", func(*model.WorkflowModel) error { return nil })
	if err != nil {
		t.Fatal(err)
	}
	ids := cfg.ListNamedWorkflowHooks()
	if len(ids) != 1 || ids[0] != id {
		t.Errorf("ListNamedWorkflowHooks = %v", ids)
	}
	if !cfg.RemoveWorkflowHook(id) {
		t.Errorf("RemoveWorkflowHook returned false")
	}
}

func TestRegisterNamedWorkflowHook_RejectsDuplicate(t *testing.T) {
	cfg := New()
	if _, err := cfg.RegisterNamedWorkflowHook("dup", func(*model.WorkflowModel) error { return nil }); err != nil {
		t.Fatal(err)
	}
	_, err := cfg.RegisterNamedWorkflowHook("dup", func(*model.WorkflowModel) error { return nil })
	if err == nil {
		t.Errorf("expected duplicate error")
	}
}

func TestRegisterNamedWorkflowHook_RejectsEmptyName(t *testing.T) {
	cfg := New()
	if _, err := cfg.RegisterNamedWorkflowHook("", func(*model.WorkflowModel) error { return nil }); err == nil {
		t.Errorf("expected empty-name error")
	}
}

func TestRegisterNamedWorkflowHook_RejectsNilFn(t *testing.T) {
	cfg := New()
	if _, err := cfg.RegisterNamedWorkflowHook("x", nil); err == nil {
		t.Errorf("expected nil-fn error")
	}
}

func TestDispatchNamedWorkflowHooks_PropagatesError(t *testing.T) {
	cfg := New()
	_, _ = cfg.RegisterNamedWorkflowHook("wf", func(*model.WorkflowModel) error {
		return errors.New("nope")
	})
	err := cfg.DispatchNamedWorkflowHooks(&model.WorkflowModel{})
	if err == nil || !strings.Contains(err.Error(), "wf") {
		t.Errorf("expected wrap with 'wf', got %v", err)
	}
}

func TestRemoveWorkflowHook_NonExistent(t *testing.T) {
	cfg := New()
	if cfg.RemoveWorkflowHook("ghost") {
		t.Errorf("RemoveWorkflowHook returned true for non-existent hook")
	}
}

func TestClearHooks_AlsoClearsNamed(t *testing.T) {
	cfg := New()
	_, _ = cfg.RegisterNamedTemplateHook("a", func(*model.TemplateModel) error { return nil })
	_, _ = cfg.RegisterNamedWorkflowHook("b", func(*model.WorkflowModel) error { return nil })
	cfg.ClearHooks()
	if len(cfg.ListNamedTemplateHooks()) != 0 || len(cfg.ListNamedWorkflowHooks()) != 0 {
		t.Errorf("ClearHooks did not clear named hooks")
	}
}

func TestReset_AlsoClearsNamed(t *testing.T) {
	cfg := New()
	_, _ = cfg.RegisterNamedTemplateHook("a", func(*model.TemplateModel) error { return nil })
	cfg.Reset()
	if len(cfg.ListNamedTemplateHooks()) != 0 {
		t.Errorf("Reset did not clear named hooks")
	}
}
