package config

import (
	"sync/atomic"
	"testing"

	"github.com/usetheodev/theo-forge/model"
)

// Phase 7 (T7.2) coverage backfill for config package.

func TestGlobalConfig_SetGetImage(t *testing.T) {
	ResetGlobalConfigForTest(t)
	GetGlobal().SetImage("a:1")
	if GetGlobal().GetImage() != "a:1" {
		t.Fatal("SetImage/GetImage round-trip failed")
	}
}

func TestGlobalConfig_SetGetNamespace(t *testing.T) {
	ResetGlobalConfigForTest(t)
	GetGlobal().SetNamespace("argo")
	if GetGlobal().GetNamespace() != "argo" {
		t.Fatal("namespace round-trip failed")
	}
}

func TestGlobalConfig_SetHostAndToken(t *testing.T) {
	ResetGlobalConfigForTest(t)
	cfg := GetGlobal()
	cfg.SetHost("https://argo.example")
	cfg.SetToken("secret")
	if cfg.GetHost() != "https://argo.example" || cfg.GetToken() != "secret" {
		t.Fatal("host/token round-trip failed")
	}
}

func TestGlobalConfig_RegisterAndDispatchTemplateHook(t *testing.T) {
	ResetGlobalConfigForTest(t)
	var count int64
	cfg := GetGlobal()
	cfg.RegisterTemplateHook(func(*model.TemplateModel) { atomic.AddInt64(&count, 1) })
	cfg.DispatchTemplateHooks(&model.TemplateModel{})
	cfg.DispatchTemplateHooks(&model.TemplateModel{})
	if got := atomic.LoadInt64(&count); got != 2 {
		t.Fatalf("hook count = %d, want 2", got)
	}
}

func TestGlobalConfig_RegisterAndDispatchWorkflowHook(t *testing.T) {
	ResetGlobalConfigForTest(t)
	var count int64
	cfg := GetGlobal()
	cfg.RegisterWorkflowHook(func(*model.WorkflowModel) { atomic.AddInt64(&count, 1) })
	cfg.DispatchWorkflowHooks(&model.WorkflowModel{})
	if got := atomic.LoadInt64(&count); got != 1 {
		t.Fatalf("hook count = %d, want 1", got)
	}
}

func TestGlobalConfig_RegisterAndDispatchWorkflowTemplateHook(t *testing.T) {
	ResetGlobalConfigForTest(t)
	var count int64
	cfg := GetGlobal()
	cfg.RegisterWorkflowTemplateHook(func(*model.WorkflowTemplateModel) { atomic.AddInt64(&count, 1) })
	cfg.DispatchWorkflowTemplateHooks(&model.WorkflowTemplateModel{})
	if got := atomic.LoadInt64(&count); got != 1 {
		t.Fatalf("hook count = %d, want 1", got)
	}
}

func TestGlobalConfig_RegisterAndDispatchCronWorkflowHook(t *testing.T) {
	ResetGlobalConfigForTest(t)
	var count int64
	cfg := GetGlobal()
	cfg.RegisterCronWorkflowHook(func(*model.CronWorkflowModel) { atomic.AddInt64(&count, 1) })
	cfg.DispatchCronWorkflowHooks(&model.CronWorkflowModel{})
	if got := atomic.LoadInt64(&count); got != 1 {
		t.Fatalf("hook count = %d, want 1", got)
	}
}

func TestGlobalConfig_ClearHooks(t *testing.T) {
	ResetGlobalConfigForTest(t)
	cfg := GetGlobal()
	cfg.RegisterTemplateHook(func(*model.TemplateModel) {})
	cfg.RegisterWorkflowHook(func(*model.WorkflowModel) {})
	cfg.RegisterWorkflowTemplateHook(func(*model.WorkflowTemplateModel) {})
	cfg.RegisterCronWorkflowHook(func(*model.CronWorkflowModel) {})
	cfg.ClearHooks()

	// After ClearHooks, dispatching should be a no-op (no panic, no count).
	cfg.DispatchTemplateHooks(&model.TemplateModel{})
	cfg.DispatchWorkflowHooks(&model.WorkflowModel{})
	cfg.DispatchWorkflowTemplateHooks(&model.WorkflowTemplateModel{})
	cfg.DispatchCronWorkflowHooks(&model.CronWorkflowModel{})
}

func TestGlobalConfig_ApplyTemplateDefaults_ContainerImage(t *testing.T) {
	ResetGlobalConfigForTest(t)
	cfg := New()
	cfg.SetImage("base:1")
	tpl := &model.TemplateModel{Container: &model.ContainerModel{}}
	cfg.ApplyTemplateDefaults(tpl)
	if tpl.Container.Image != "base:1" {
		t.Fatalf("image not applied: %q", tpl.Container.Image)
	}
}

func TestGlobalConfig_ApplyTemplateDefaults_DoesNotOverwriteExplicit(t *testing.T) {
	ResetGlobalConfigForTest(t)
	cfg := New()
	cfg.SetImage("base:1")
	tpl := &model.TemplateModel{Container: &model.ContainerModel{Image: "user:9"}}
	cfg.ApplyTemplateDefaults(tpl)
	if tpl.Container.Image != "user:9" {
		t.Fatalf("explicit image overwritten: %q", tpl.Container.Image)
	}
}

func TestGlobalConfig_ApplyTemplateDefaults_NilSafe(t *testing.T) {
	var nilCfg *GlobalConfig
	nilCfg.ApplyTemplateDefaults(nil)                    // should not panic
	nilCfg.ApplyTemplateDefaults(&model.TemplateModel{}) // should not panic
}
