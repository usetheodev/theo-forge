package forge

import (
	"sync/atomic"
	"testing"

	"github.com/usetheodev/theo-forge/config"
	"github.com/usetheodev/theo-forge/model"
)

// T3.1 regression tests (arch-globalconfig-frozen-singleton, val-004, completeness-globalconfig-dead-scalars)
// Covers ADR-001 (resolveConfig + WithConfig + ApplyTemplateDefaults).

func TestWorkflow_WithConfig_HookDispatched(t *testing.T) {
	config.ResetGlobalConfigForTest(t)
	cfg := NewConfig()
	var count int64
	cfg.RegisterTemplateHook(func(_ *model.TemplateModel) {
		atomic.AddInt64(&count, 1)
	})
	w := (&Workflow{
		Name:       "wf",
		Entrypoint: "main",
		Templates: []Templatable{
			&Container{Name: "main", Image: "alpine:3.18"},
		},
	}).WithConfig(cfg)

	if _, err := w.Build(); err != nil {
		t.Fatalf("Build err: %v", err)
	}
	if got := atomic.LoadInt64(&count); got != 1 {
		t.Fatalf("hook dispatched %d times via WithConfig, want 1", got)
	}
}

func TestWorkflow_WithConfig_HookDoesNotLeakToGlobal(t *testing.T) {
	config.ResetGlobalConfigForTest(t)
	cfg := NewConfig()
	var isolated int64
	cfg.RegisterTemplateHook(func(_ *model.TemplateModel) {
		atomic.AddInt64(&isolated, 1)
	})

	// Build WITHOUT WithConfig: must use the (empty) global singleton, not cfg.
	w := &Workflow{
		Name:       "wf",
		Entrypoint: "main",
		Templates: []Templatable{
			&Container{Name: "main", Image: "alpine:3.18"},
		},
	}
	if _, err := w.Build(); err != nil {
		t.Fatalf("Build err: %v", err)
	}
	if got := atomic.LoadInt64(&isolated); got != 0 {
		t.Fatalf("isolated hook leaked to global singleton: count=%d", got)
	}
}

func TestWorkflow_WithConfig_AppliesImageDefault(t *testing.T) {
	config.ResetGlobalConfigForTest(t)
	cfg := NewConfig()
	cfg.SetImage("custom:9.9")
	w := (&Workflow{
		Name:       "wf",
		Entrypoint: "main",
		Templates: []Templatable{
			&Container{Name: "main"}, // Image intentionally unset
		},
	}).WithConfig(cfg)

	wf, err := w.Build()
	if err != nil {
		t.Fatalf("Build err: %v", err)
	}
	if len(wf.Spec.Templates) == 0 || wf.Spec.Templates[0].Container == nil {
		t.Fatalf("Build produced no Container template")
	}
	if wf.Spec.Templates[0].Container.Image != "custom:9.9" {
		t.Fatalf("Image default not applied: got %q, want \"custom:9.9\"", wf.Spec.Templates[0].Container.Image)
	}
}

func TestWorkflow_WithConfig_ExplicitImageWins(t *testing.T) {
	config.ResetGlobalConfigForTest(t)
	cfg := NewConfig()
	cfg.SetImage("default:1")
	w := (&Workflow{
		Name:       "wf",
		Entrypoint: "main",
		Templates: []Templatable{
			&Container{Name: "main", Image: "explicit:2"},
		},
	}).WithConfig(cfg)
	wf, _ := w.Build()
	if wf.Spec.Templates[0].Container.Image != "explicit:2" {
		t.Fatalf("explicit Image overridden: got %q, want \"explicit:2\"", wf.Spec.Templates[0].Container.Image)
	}
}

func TestWorkflow_WithoutConfig_UsesGlobalSingleton(t *testing.T) {
	config.ResetGlobalConfigForTest(t)
	config.GetGlobal().SetImage("singleton:default")

	w := &Workflow{
		Name:       "wf",
		Entrypoint: "main",
		Templates: []Templatable{
			&Container{Name: "main"},
		},
	}
	wf, err := w.Build()
	if err != nil {
		t.Fatalf("Build err: %v", err)
	}
	if wf.Spec.Templates[0].Container.Image != "singleton:default" {
		t.Fatalf("singleton default not applied: got %q", wf.Spec.Templates[0].Container.Image)
	}
}

func TestWorkflow_WithConfig_ExplicitOverridesGlobal(t *testing.T) {
	config.ResetGlobalConfigForTest(t)
	config.GetGlobal().SetImage("singleton:wrong")
	cfg := NewConfig()
	cfg.SetImage("explicit:right")

	w := (&Workflow{
		Name:       "wf",
		Entrypoint: "main",
		Templates: []Templatable{
			&Container{Name: "main"},
		},
	}).WithConfig(cfg)
	wf, _ := w.Build()
	if wf.Spec.Templates[0].Container.Image != "explicit:right" {
		t.Fatalf("explicit Config did not win over singleton: got %q", wf.Spec.Templates[0].Container.Image)
	}
}
