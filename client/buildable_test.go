package client

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/usetheodev/theo-forge/model"
)

// Coverage backfill for client (T7.6 round 2):
// - CreateWorkflow / LintWorkflow (Buildable convenience methods)
// - NoopLogger / SlogLogger explicit calls
// - template_ops (CreateWorkflowTemplate / CreateCronWorkflow)
// - Error paths in CreateWorkflowFromModel / DeleteWorkflow / ListWorkflows / etc.

type fakeBuildable struct {
	wf model.WorkflowModel
	ns string
}

func (b fakeBuildable) Build() (model.WorkflowModel, error) { return b.wf, nil }
func (b fakeBuildable) GetNamespace() string                { return b.ns }

type errBuildable struct{}

func (errBuildable) Build() (model.WorkflowModel, error) {
	return model.WorkflowModel{}, errors.New("build failed")
}
func (errBuildable) GetNamespace() string { return "ns" }

func TestClient_CreateWorkflow_DelegatesToFromModel(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"metadata":{"name":"created"}}`))
	}))
	t.Cleanup(server.Close)
	svc := NewWorkflowsService(server.URL, "t", "default")
	got, err := svc.CreateWorkflow(context.Background(), fakeBuildable{wf: model.WorkflowModel{}, ns: "default"})
	if err != nil {
		t.Fatal(err)
	}
	if got.Metadata.Name != "created" {
		t.Errorf("Metadata.Name = %q", got.Metadata.Name)
	}
}

func TestClient_CreateWorkflow_PropagatesBuildError(t *testing.T) {
	svc := NewWorkflowsService("http://x", "t", "ns")
	if _, err := svc.CreateWorkflow(context.Background(), errBuildable{}); err == nil {
		t.Fatal("expected build error to propagate")
	}
}

func TestClient_LintWorkflow_DelegatesToFromModel(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"metadata":{"name":"linted"}}`))
	}))
	t.Cleanup(server.Close)
	svc := NewWorkflowsService(server.URL, "t", "default")
	got, err := svc.LintWorkflow(context.Background(), fakeBuildable{ns: "default"})
	if err != nil {
		t.Fatal(err)
	}
	if got.Metadata.Name != "linted" {
		t.Errorf("Metadata.Name = %q", got.Metadata.Name)
	}
}

func TestClient_LintWorkflow_PropagatesBuildError(t *testing.T) {
	svc := NewWorkflowsService("http://x", "t", "ns")
	if _, err := svc.LintWorkflow(context.Background(), errBuildable{}); err == nil {
		t.Fatal("expected build error to propagate")
	}
}

// Template ops

type fakeWfTemplate struct {
	m model.WorkflowTemplateModel
}

func (f fakeWfTemplate) Build() (model.WorkflowTemplateModel, error) { return f.m, nil }

func TestClient_CreateWorkflowTemplate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"metadata":{"name":"wft-created"}}`))
	}))
	t.Cleanup(server.Close)
	svc := NewWorkflowsService(server.URL, "t", "default")
	got, err := svc.CreateWorkflowTemplate(context.Background(), fakeWfTemplate{
		m: model.WorkflowTemplateModel{Metadata: model.WorkflowMetadata{Name: "wft"}},
	}, "default")
	if err != nil {
		t.Fatal(err)
	}
	if got.Metadata.Name != "wft-created" {
		t.Errorf("Metadata.Name = %q", got.Metadata.Name)
	}
}

func TestClient_CreateWorkflowTemplate_RejectsBadNamespace(t *testing.T) {
	svc := NewWorkflowsService("http://x", "t", "default/../admin")
	_, err := svc.CreateWorkflowTemplate(context.Background(), fakeWfTemplate{}, "")
	if !errors.Is(err, model.ErrInvalidNamespace) {
		t.Errorf("got %v, want ErrInvalidNamespace", err)
	}
}

type fakeCronWf struct {
	m model.CronWorkflowModel
}

func (f fakeCronWf) Build() (model.CronWorkflowModel, error) { return f.m, nil }

func TestClient_CreateCronWorkflow(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"metadata":{"name":"cw-created"}}`))
	}))
	t.Cleanup(server.Close)
	svc := NewWorkflowsService(server.URL, "t", "default")
	got, err := svc.CreateCronWorkflow(context.Background(), fakeCronWf{
		m: model.CronWorkflowModel{Metadata: model.WorkflowMetadata{Name: "cw"}},
	}, "default")
	if err != nil {
		t.Fatal(err)
	}
	if got.Metadata.Name != "cw-created" {
		t.Errorf("Metadata.Name = %q", got.Metadata.Name)
	}
}

func TestClient_CreateCronWorkflow_RejectsBadNamespace(t *testing.T) {
	svc := NewWorkflowsService("http://x", "t", "")
	_, err := svc.CreateCronWorkflow(context.Background(), fakeCronWf{}, "")
	if !errors.Is(err, model.ErrInvalidNamespace) {
		t.Errorf("got %v, want ErrInvalidNamespace", err)
	}
}

// Logger adapters

func TestNoopLogger_DoesNothing(t *testing.T) {
	var l Logger = NoopLogger{}
	l.Debug("msg", "k", "v")
	l.Error("err", "k", "v")
}

func TestSlogLogger_WithNilLoggerIsSafe(t *testing.T) {
	var l Logger = SlogLogger{} // L is nil
	l.Debug("msg", "k", "v")
	l.Error("err", "k", "v")
}
