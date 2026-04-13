package client

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

// mockHTTPClient records the request and returns a configurable response.
type mockHTTPClient struct {
	lastMethod string
	lastPath   string
	statusCode int
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	m.lastMethod = req.Method
	m.lastPath = req.URL.Path
	return &http.Response{
		StatusCode: m.statusCode,
		Body:       io.NopCloser(strings.NewReader(`{}`)),
	}, nil
}

func newTestService(mock *mockHTTPClient) *WorkflowsService {
	return &WorkflowsService{
		Host:       "http://localhost:2746",
		Token:      "test-token",
		Namespace:  "default",
		HTTPClient: mock,
	}
}

func TestStopWorkflow(t *testing.T) {
	mock := &mockHTTPClient{statusCode: 200}
	svc := newTestService(mock)

	err := svc.StopWorkflow(context.Background(), "my-wf", "build-ns")
	if err != nil {
		t.Fatalf("StopWorkflow: %v", err)
	}
	if mock.lastMethod != http.MethodPut {
		t.Errorf("method = %q, want PUT", mock.lastMethod)
	}
	if mock.lastPath != "/api/v1/workflows/build-ns/my-wf/stop" {
		t.Errorf("path = %q", mock.lastPath)
	}
}

func TestTerminateWorkflow(t *testing.T) {
	mock := &mockHTTPClient{statusCode: 200}
	svc := newTestService(mock)

	err := svc.TerminateWorkflow(context.Background(), "my-wf", "build-ns")
	if err != nil {
		t.Fatalf("TerminateWorkflow: %v", err)
	}
	if mock.lastMethod != http.MethodPut {
		t.Errorf("method = %q, want PUT", mock.lastMethod)
	}
	if mock.lastPath != "/api/v1/workflows/build-ns/my-wf/terminate" {
		t.Errorf("path = %q", mock.lastPath)
	}
}

func TestSuspendWorkflow(t *testing.T) {
	mock := &mockHTTPClient{statusCode: 200}
	svc := newTestService(mock)

	err := svc.SuspendWorkflow(context.Background(), "my-wf", "")
	if err != nil {
		t.Fatalf("SuspendWorkflow: %v", err)
	}
	if mock.lastPath != "/api/v1/workflows/default/my-wf/suspend" {
		t.Errorf("path = %q (namespace fallback failed)", mock.lastPath)
	}
}

func TestResumeWorkflow(t *testing.T) {
	mock := &mockHTTPClient{statusCode: 200}
	svc := newTestService(mock)

	err := svc.ResumeWorkflow(context.Background(), "my-wf", "")
	if err != nil {
		t.Fatalf("ResumeWorkflow: %v", err)
	}
	if mock.lastPath != "/api/v1/workflows/default/my-wf/resume" {
		t.Errorf("path = %q (namespace fallback failed)", mock.lastPath)
	}
}

func TestLifecycleOps_NamespaceFallback(t *testing.T) {
	tests := []struct {
		name     string
		fn       func(*WorkflowsService, context.Context) error
		wantPath string
	}{
		{
			name: "stop",
			fn: func(s *WorkflowsService, ctx context.Context) error {
				return s.StopWorkflow(ctx, "wf1", "")
			},
			wantPath: "/api/v1/workflows/default/wf1/stop",
		},
		{
			name: "terminate",
			fn: func(s *WorkflowsService, ctx context.Context) error {
				return s.TerminateWorkflow(ctx, "wf1", "")
			},
			wantPath: "/api/v1/workflows/default/wf1/terminate",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockHTTPClient{statusCode: 200}
			svc := newTestService(mock)
			err := tt.fn(svc, context.Background())
			if err != nil {
				t.Fatalf("error: %v", err)
			}
			if mock.lastPath != tt.wantPath {
				t.Errorf("path = %q, want %q", mock.lastPath, tt.wantPath)
			}
		})
	}
}
