package forge

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/usetheo/theo/forge/client"
	"github.com/usetheo/theo/forge/model"
)

// mockHTTPClient is a mock HTTP client for testing.
type mockHTTPClient struct {
	DoFunc func(req *http.Request) (*http.Response, error)
}

func (m *mockHTTPClient) Do(req *http.Request) (*http.Response, error) {
	return m.DoFunc(req)
}

func mockResponse(statusCode int, body interface{}) *http.Response {
	data, _ := json.Marshal(body)
	return &http.Response{
		StatusCode: statusCode,
		Body:       io.NopCloser(bytes.NewReader(data)),
		Header:     http.Header{},
	}
}

func TestNewWorkflowsService(t *testing.T) {
	svc := client.NewWorkflowsService("https://argo.example.com", "my-token", "default")
	if svc.Host != "https://argo.example.com" {
		t.Errorf("host = %q", svc.Host)
	}
	if svc.Token != "my-token" {
		t.Errorf("token = %q", svc.Token)
	}
	if svc.Namespace != "default" {
		t.Errorf("namespace = %q", svc.Namespace)
	}
}

func TestServiceTokenFormatting(t *testing.T) {
	tests := []struct {
		name  string
		token string
		want  string
	}{
		{"plain token", "my-token", "Bearer my-token"},
		{"bearer prefix", "Bearer my-token", "Bearer my-token"},
		{"empty", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := &client.WorkflowsService{Token: tt.token}
			got := svc.FormatToken()
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestServiceCreateWorkflow(t *testing.T) {
	expectedModel := model.WorkflowModel{
		APIVersion: DefaultAPIVersion,
		Kind:       DefaultKind,
		Metadata:   model.WorkflowMetadata{Name: "created-workflow"},
	}

	svc := &client.WorkflowsService{
		Host:      "https://argo.example.com",
		Token:     "test-token",
		Namespace: "default",
		HTTPClient: &mockHTTPClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				if req.Method != http.MethodPost {
					t.Errorf("method = %q, want POST", req.Method)
				}
				if req.URL.Path != "/api/v1/workflows/default" {
					t.Errorf("path = %q", req.URL.Path)
				}
				if req.Header.Get("Authorization") != "Bearer test-token" {
					t.Errorf("auth = %q", req.Header.Get("Authorization"))
				}
				if req.Header.Get("Content-Type") != "application/json" {
					t.Errorf("content-type = %q", req.Header.Get("Content-Type"))
				}
				return mockResponse(200, expectedModel), nil
			},
		},
	}

	w := &Workflow{
		Name:       "test-workflow",
		Entrypoint: "main",
		Templates:  []Templatable{&Container{Name: "main", Image: "alpine"}},
	}

	result, err := svc.CreateWorkflow(context.Background(), w)
	if err != nil {
		t.Fatal(err)
	}
	if result.Metadata.Name != "created-workflow" {
		t.Errorf("name = %q", result.Metadata.Name)
	}
}

func TestServiceCreateWorkflowUsesWorkflowNamespace(t *testing.T) {
	svc := &client.WorkflowsService{
		Host:      "https://argo.example.com",
		Namespace: "default",
		HTTPClient: &mockHTTPClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				if req.URL.Path != "/api/v1/workflows/custom-ns" {
					t.Errorf("path = %q, want /api/v1/workflows/custom-ns", req.URL.Path)
				}
				return mockResponse(200, model.WorkflowModel{}), nil
			},
		},
	}

	w := &Workflow{
		Name:       "test",
		Namespace:  "custom-ns",
		Entrypoint: "main",
		Templates:  []Templatable{&Container{Name: "main", Image: "alpine"}},
	}
	svc.CreateWorkflow(context.Background(), w)
}

func TestServiceGetWorkflow(t *testing.T) {
	expectedModel := model.WorkflowModel{
		APIVersion: DefaultAPIVersion,
		Kind:       DefaultKind,
		Metadata:   model.WorkflowMetadata{Name: "my-workflow"},
	}

	svc := &client.WorkflowsService{
		Host:      "https://argo.example.com",
		Namespace: "default",
		HTTPClient: &mockHTTPClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				if req.Method != http.MethodGet {
					t.Errorf("method = %q, want GET", req.Method)
				}
				if req.URL.Path != "/api/v1/workflows/default/my-workflow" {
					t.Errorf("path = %q", req.URL.Path)
				}
				return mockResponse(200, expectedModel), nil
			},
		},
	}

	result, err := svc.GetWorkflow(context.Background(), "my-workflow", "")
	if err != nil {
		t.Fatal(err)
	}
	if result.Metadata.Name != "my-workflow" {
		t.Errorf("name = %q", result.Metadata.Name)
	}
}

func TestServiceDeleteWorkflow(t *testing.T) {
	svc := &client.WorkflowsService{
		Host:      "https://argo.example.com",
		Namespace: "default",
		HTTPClient: &mockHTTPClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				if req.Method != http.MethodDelete {
					t.Errorf("method = %q, want DELETE", req.Method)
				}
				if req.URL.Path != "/api/v1/workflows/default/my-workflow" {
					t.Errorf("path = %q", req.URL.Path)
				}
				return mockResponse(200, map[string]string{}), nil
			},
		},
	}

	err := svc.DeleteWorkflow(context.Background(), "my-workflow", "")
	if err != nil {
		t.Fatal(err)
	}
}

func TestServiceListWorkflows(t *testing.T) {
	listResp := client.ListWorkflowsResponse{
		Items: []model.WorkflowModel{
			{Metadata: model.WorkflowMetadata{Name: "wf-1"}},
			{Metadata: model.WorkflowMetadata{Name: "wf-2"}},
		},
	}

	svc := &client.WorkflowsService{
		Host:      "https://argo.example.com",
		Namespace: "default",
		HTTPClient: &mockHTTPClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				if req.Method != http.MethodGet {
					t.Errorf("method = %q, want GET", req.Method)
				}
				return mockResponse(200, listResp), nil
			},
		},
	}

	items, err := svc.ListWorkflows(context.Background(), "")
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 {
		t.Fatalf("items = %d, want 2", len(items))
	}
	if items[0].Metadata.Name != "wf-1" {
		t.Errorf("items[0].name = %q", items[0].Metadata.Name)
	}
}

func TestServiceLintWorkflow(t *testing.T) {
	svc := &client.WorkflowsService{
		Host:      "https://argo.example.com",
		Namespace: "default",
		HTTPClient: &mockHTTPClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				if req.URL.Path != "/api/v1/workflows/default/lint" {
					t.Errorf("path = %q", req.URL.Path)
				}
				return mockResponse(200, model.WorkflowModel{Metadata: model.WorkflowMetadata{Name: "linted"}}), nil
			},
		},
	}

	w := &Workflow{
		Name:       "test",
		Entrypoint: "main",
		Templates:  []Templatable{&Container{Name: "main", Image: "alpine"}},
	}

	result, err := svc.LintWorkflow(context.Background(), w)
	if err != nil {
		t.Fatal(err)
	}
	if result.Metadata.Name != "linted" {
		t.Errorf("name = %q", result.Metadata.Name)
	}
}

func TestServiceAPIError(t *testing.T) {
	svc := &client.WorkflowsService{
		Host:      "https://argo.example.com",
		Namespace: "default",
		HTTPClient: &mockHTTPClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				return mockResponse(404, map[string]string{"message": "not found"}), nil
			},
		},
	}

	_, err := svc.GetWorkflow(context.Background(), "nonexistent", "")
	if err == nil {
		t.Fatal("expected error")
	}
	var apiErr *client.APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected APIError, got %T: %v", err, err)
	}
	if apiErr.StatusCode != 404 {
		t.Errorf("status = %d, want 404", apiErr.StatusCode)
	}
}

func TestServiceNetworkError(t *testing.T) {
	svc := &client.WorkflowsService{
		Host:      "https://argo.example.com",
		Namespace: "default",
		HTTPClient: &mockHTTPClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				return nil, fmt.Errorf("connection refused")
			},
		},
	}

	_, err := svc.GetWorkflow(context.Background(), "test", "")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "connection refused") {
		t.Errorf("error = %q", err.Error())
	}
}

func TestServiceGetInfo(t *testing.T) {
	svc := &client.WorkflowsService{
		Host: "https://argo.example.com",
		HTTPClient: &mockHTTPClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				if req.URL.Path != "/api/v1/info" {
					t.Errorf("path = %q", req.URL.Path)
				}
				return mockResponse(200, map[string]interface{}{"managedNamespace": "argo"}), nil
			},
		},
	}

	info, err := svc.GetInfo(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if info["managedNamespace"] != "argo" {
		t.Errorf("managedNamespace = %v", info["managedNamespace"])
	}
}

func TestServiceHostTrailingSlash(t *testing.T) {
	svc := client.NewWorkflowsService("https://argo.example.com/", "token", "ns")
	if svc.Host != "https://argo.example.com" {
		t.Errorf("host = %q, trailing slash should be removed", svc.Host)
	}
}
