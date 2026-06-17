package client

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/usetheodev/theo-forge/model"
)

// T2.3 regression tests (SEC-003). Covers validateK8sName, validateK8sNamespace,
// and observable URL shape via httptest.

func TestValidateK8sName(t *testing.T) {
	good := []string{"a", "valid", "valid-name", "wf.with.dots", "abc123"}
	for _, n := range good {
		if err := validateK8sName(n); err != nil {
			t.Errorf("validateK8sName(%q) = %v, want nil", n, err)
		}
	}
	bad := []string{"", "-leading-dash", "trailing-dash-", "UPPER", "with spaces", "default/../admin", "../escape", strings.Repeat("a", 254)}
	for _, n := range bad {
		if err := validateK8sName(n); !errors.Is(err, model.ErrInvalidName) {
			t.Errorf("validateK8sName(%q) = %v, want ErrInvalidName", n, err)
		}
	}
}

func TestValidateK8sNamespace(t *testing.T) {
	if err := validateK8sNamespace("default"); err != nil {
		t.Errorf("got %v, want nil", err)
	}
	if err := validateK8sNamespace("default/../admin"); !errors.Is(err, model.ErrInvalidNamespace) {
		t.Errorf("traversal namespace accepted: err=%v", err)
	}
	if err := validateK8sNamespace(""); !errors.Is(err, model.ErrInvalidNamespace) {
		t.Errorf("empty namespace accepted: err=%v", err)
	}
}

func TestClient_GetWorkflow_RejectsInvalidNamespace(t *testing.T) {
	svc := NewWorkflowsService("http://example.test", "tok", "default")
	_, err := svc.GetWorkflow(context.Background(), "wf", "default/../admin")
	if !errors.Is(err, model.ErrInvalidNamespace) {
		t.Fatalf("got %v, want ErrInvalidNamespace", err)
	}
}

func TestClient_GetWorkflow_RejectsInvalidName(t *testing.T) {
	svc := NewWorkflowsService("http://example.test", "tok", "default")
	_, err := svc.GetWorkflow(context.Background(), "../escape", "default")
	if !errors.Is(err, model.ErrInvalidName) {
		t.Fatalf("got %v, want ErrInvalidName", err)
	}
}

func TestClient_LifecycleOps_RejectInvalidName(t *testing.T) {
	svc := NewWorkflowsService("http://example.test", "tok", "default")
	ops := []struct {
		name string
		fn   func() error
	}{
		{"stop", func() error { return svc.StopWorkflow(context.Background(), "../bad", "default") }},
		{"terminate", func() error { return svc.TerminateWorkflow(context.Background(), "../bad", "default") }},
		{"suspend", func() error { return svc.SuspendWorkflow(context.Background(), "../bad", "default") }},
		{"resume", func() error { return svc.ResumeWorkflow(context.Background(), "../bad", "default") }},
	}
	for _, op := range ops {
		t.Run(op.name, func(t *testing.T) {
			if err := op.fn(); !errors.Is(err, model.ErrInvalidName) {
				t.Errorf("%s: got %v, want ErrInvalidName", op.name, err)
			}
		})
	}
}

// Observed-path test: confirm URL escaping is applied. Valid name "abc-1.2"
// (contains a dot which is valid in RFC1123 DNS subdomain) goes through
// url.PathEscape and matches what we expect.
func TestClient_GetWorkflow_AppliesPathEscape(t *testing.T) {
	var observedPath string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		observedPath = r.URL.EscapedPath()
		_, _ = w.Write([]byte(`{"metadata":{}}`))
	}))
	t.Cleanup(server.Close)

	svc := NewWorkflowsService(server.URL, "tok", "ns")
	_, err := svc.GetWorkflow(context.Background(), "abc-1.2", "ns")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	wantPath := "/api/v1/workflows/" + url.PathEscape("ns") + "/" + url.PathEscape("abc-1.2")
	if observedPath != wantPath {
		t.Fatalf("observedPath = %q, want %q", observedPath, wantPath)
	}
}
