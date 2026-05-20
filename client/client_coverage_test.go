package client

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/usetheodev/theo-forge/model"
)

// Phase 7 (T7.6) coverage backfill for client package.

func TestNewWorkflowsService_TrimsHostTrailingSlash(t *testing.T) {
	svc := NewWorkflowsService("https://argo.example.com/", "tok", "ns")
	if svc.Host != "https://argo.example.com" {
		t.Errorf("Host = %q, want trimmed", svc.Host)
	}
}

func TestFormatToken_AddsBearerPrefix(t *testing.T) {
	svc := NewWorkflowsService("http://x", "raw-token", "ns")
	if svc.FormatToken() != "Bearer raw-token" {
		t.Errorf("FormatToken = %q", svc.FormatToken())
	}
}

func TestFormatToken_PreservesExistingBearer(t *testing.T) {
	svc := NewWorkflowsService("http://x", "Bearer alreadyset", "ns")
	if svc.FormatToken() != "Bearer alreadyset" {
		t.Errorf("FormatToken = %q", svc.FormatToken())
	}
}

func TestFormatToken_EmptyReturnsEmpty(t *testing.T) {
	svc := NewWorkflowsService("http://x", "", "ns")
	if svc.FormatToken() != "" {
		t.Errorf("expected empty, got %q", svc.FormatToken())
	}
}

func TestAPIError_Format(t *testing.T) {
	e := &APIError{StatusCode: 418, Message: "I'm a teapot"}
	if !strings.Contains(e.Error(), "418") || !strings.Contains(e.Error(), "teapot") {
		t.Errorf("APIError.Error() = %q", e.Error())
	}
}

func TestClient_DoesNotIncludeAuthHeaderWhenTokenEmpty(t *testing.T) {
	var observedAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		observedAuth = r.Header.Get("Authorization")
		_, _ = w.Write([]byte(`{"items":[]}`))
	}))
	t.Cleanup(server.Close)
	svc := NewWorkflowsService(server.URL, "", "ns")
	_, _ = svc.ListWorkflows(context.Background(), "ns")
	if observedAuth != "" {
		t.Errorf("expected no Authorization header, got %q", observedAuth)
	}
}

func TestClient_IncludesAuthHeaderWhenTokenSet(t *testing.T) {
	var observedAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		observedAuth = r.Header.Get("Authorization")
		_, _ = w.Write([]byte(`{"items":[]}`))
	}))
	t.Cleanup(server.Close)
	svc := NewWorkflowsService(server.URL, "abc", "ns")
	_, _ = svc.ListWorkflows(context.Background(), "ns")
	if observedAuth != "Bearer abc" {
		t.Errorf("Authorization = %q, want \"Bearer abc\"", observedAuth)
	}
}

func TestClient_APIErrorOnNon2xx(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(`{"error":"nope"}`))
	}))
	t.Cleanup(server.Close)
	svc := NewWorkflowsService(server.URL, "tok", "ns")
	_, err := svc.GetWorkflow(context.Background(), "wf", "ns")
	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected APIError, got %v", err)
	}
	if apiErr.StatusCode != http.StatusForbidden {
		t.Errorf("StatusCode = %d, want 403", apiErr.StatusCode)
	}
}

func TestClient_LifecycleOps_SendCorrectMethod(t *testing.T) {
	var observedMethod, observedPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		observedMethod = r.Method
		observedPath = r.URL.Path
		w.WriteHeader(200)
	}))
	t.Cleanup(server.Close)
	svc := NewWorkflowsService(server.URL, "t", "ns")
	if err := svc.StopWorkflow(context.Background(), "wf", "ns"); err != nil {
		t.Fatal(err)
	}
	if observedMethod != http.MethodPut {
		t.Errorf("method = %q, want PUT", observedMethod)
	}
	if !strings.HasSuffix(observedPath, "/stop") {
		t.Errorf("path = %q, want suffix /stop", observedPath)
	}
}

func TestClient_DeleteWorkflow(t *testing.T) {
	var observed string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		observed = r.Method
		w.WriteHeader(200)
	}))
	t.Cleanup(server.Close)
	svc := NewWorkflowsService(server.URL, "t", "ns")
	if err := svc.DeleteWorkflow(context.Background(), "wf", "ns"); err != nil {
		t.Fatal(err)
	}
	if observed != http.MethodDelete {
		t.Errorf("method = %q", observed)
	}
}

func TestClient_CreateWorkflowFromModel(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"metadata":{"name":"submitted"}}`))
	}))
	t.Cleanup(server.Close)
	svc := NewWorkflowsService(server.URL, "t", "ns")
	got, err := svc.CreateWorkflowFromModel(context.Background(), model.WorkflowModel{
		Metadata: model.WorkflowMetadata{Name: "x"},
	}, "ns")
	if err != nil {
		t.Fatal(err)
	}
	if got.Metadata.Name != "submitted" {
		t.Errorf("Metadata.Name = %q, want \"submitted\"", got.Metadata.Name)
	}
}

func TestClient_LintWorkflowFromModel(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"metadata":{"name":"linted"}}`))
	}))
	t.Cleanup(server.Close)
	svc := NewWorkflowsService(server.URL, "t", "ns")
	got, err := svc.LintWorkflowFromModel(context.Background(), model.WorkflowModel{}, "ns")
	if err != nil {
		t.Fatal(err)
	}
	if got.Metadata.Name != "linted" {
		t.Errorf("Metadata.Name = %q", got.Metadata.Name)
	}
}

func TestClient_GetVersionAndInfoTyped(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/version":
			_, _ = w.Write([]byte(`{"version":"v3.5.0","platform":"linux/amd64"}`))
		case "/api/v1/info":
			_, _ = w.Write([]byte(`{"managedNamespace":"argo"}`))
		}
	}))
	t.Cleanup(server.Close)
	svc := NewWorkflowsService(server.URL, "t", "ns")
	v, err := svc.GetVersion(context.Background())
	if err != nil || v.Version != "v3.5.0" || v.Platform != "linux/amd64" {
		t.Errorf("GetVersion: %+v, err %v", v, err)
	}
	i, err := svc.GetInfo(context.Background())
	if err != nil || i.ManagedNamespace != "argo" {
		t.Errorf("GetInfo: %+v, err %v", i, err)
	}
}

func TestNoopLoggerSilent(t *testing.T) {
	(NoopLogger{}).Debug("x", "k", "v")
	(NoopLogger{}).Error("y", "k", "v")
}

func TestSlogLogger_NilSafe(t *testing.T) {
	(SlogLogger{}).Debug("x")
	(SlogLogger{}).Error("y")
}

func TestKvToAttrs_OddLength(t *testing.T) {
	attrs := kvToAttrs([]any{"k1", "v1", "k2"})
	if len(attrs) != 2 {
		t.Errorf("expected 2 attrs, got %d", len(attrs))
	}
}

func TestKvToAttrs_NonStringKey(t *testing.T) {
	attrs := kvToAttrs([]any{123, "v"})
	if len(attrs) != 1 || attrs[0].Key != "INVALID_KEY" {
		t.Errorf("expected INVALID_KEY fallback, got %+v", attrs)
	}
}
