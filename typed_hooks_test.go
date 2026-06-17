package forge

import (
	"sync/atomic"
	"testing"

	"github.com/usetheodev/theo-forge/config"
	"github.com/usetheodev/theo-forge/model"
)

// T3.2 regression tests (completeness-workflow-hook-not-fired-for-wftemplate).

func TestWorkflowTemplate_PreBuildHookDispatched(t *testing.T) {
	config.ResetGlobalConfigForTest(t)
	cfg := NewConfig()
	var count int64
	cfg.RegisterWorkflowTemplateHook(func(_ *model.WorkflowTemplateModel) {
		atomic.AddInt64(&count, 1)
	})
	wt := (&WorkflowTemplate{
		Name:       "wft",
		Entrypoint: "main",
		Templates: []Templatable{
			&Container{Name: "main", Image: "alpine:3.18"},
		},
	}).WithConfig(cfg)
	if _, err := wt.Build(); err != nil {
		t.Fatalf("Build err: %v", err)
	}
	if got := atomic.LoadInt64(&count); got != 1 {
		t.Fatalf("WorkflowTemplate hook fired %d times, want 1", got)
	}
}

func TestClusterWorkflowTemplate_PreBuildHookDispatched(t *testing.T) {
	config.ResetGlobalConfigForTest(t)
	cfg := NewConfig()
	var count int64
	cfg.RegisterWorkflowTemplateHook(func(_ *model.WorkflowTemplateModel) {
		atomic.AddInt64(&count, 1)
	})
	cwt := (&ClusterWorkflowTemplate{
		Name:       "cwft",
		Entrypoint: "main",
		Templates: []Templatable{
			&Container{Name: "main", Image: "alpine:3.18"},
		},
	}).WithConfig(cfg)
	if _, err := cwt.Build(); err != nil {
		t.Fatalf("Build err: %v", err)
	}
	if got := atomic.LoadInt64(&count); got != 1 {
		t.Fatalf("ClusterWorkflowTemplate hook fired %d times, want 1", got)
	}
}

func TestCronWorkflow_PreBuildHookDispatched(t *testing.T) {
	config.ResetGlobalConfigForTest(t)
	cfg := NewConfig()
	var count int64
	cfg.RegisterCronWorkflowHook(func(_ *model.CronWorkflowModel) {
		atomic.AddInt64(&count, 1)
	})
	cw := (&CronWorkflow{
		Name:       "cw",
		Schedule:   "*/5 * * * *",
		Entrypoint: "main",
		Templates: []Templatable{
			&Container{Name: "main", Image: "alpine:3.18"},
		},
	}).WithConfig(cfg)
	if _, err := cw.Build(); err != nil {
		t.Fatalf("Build err: %v", err)
	}
	if got := atomic.LoadInt64(&count); got != 1 {
		t.Fatalf("CronWorkflow hook fired %d times, want 1", got)
	}
}
