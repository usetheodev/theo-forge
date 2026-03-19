package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/usetheo/theo/forge/model"
)

// HTTPClient is an interface for HTTP requests (allows mocking).
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// Buildable is any type that can produce a WorkflowModel.
type Buildable interface {
	Build() (model.WorkflowModel, error)
	GetNamespace() string
}

// WorkflowsService is the REST client for the Argo Workflows API.
type WorkflowsService struct {
	// Host is the Argo server URL.
	Host string
	// Token is the Bearer token for authentication.
	Token string
	// Namespace is the default namespace.
	Namespace string
	// VerifySSL controls TLS verification.
	VerifySSL bool
	// HTTPClient is the underlying HTTP client (injectable for testing).
	HTTPClient HTTPClient
}

// NewWorkflowsService creates a new WorkflowsService.
func NewWorkflowsService(host, token, namespace string) *WorkflowsService {
	return &WorkflowsService{
		Host:       strings.TrimRight(host, "/"),
		Token:      token,
		Namespace:  namespace,
		VerifySSL:  true,
		HTTPClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// FormatToken formats the token for the Authorization header.
func (s *WorkflowsService) FormatToken() string {
	if s.Token == "" {
		return ""
	}
	if strings.HasPrefix(s.Token, "Bearer ") {
		return s.Token
	}
	return "Bearer " + s.Token
}

func (s *WorkflowsService) doRequest(ctx context.Context, method, path string, body interface{}) ([]byte, int, error) {
	url := s.Host + path

	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, 0, fmt.Errorf("marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, 0, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if token := s.FormatToken(); token != "" {
		req.Header.Set("Authorization", token)
	}

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return respBody, resp.StatusCode, &APIError{
			StatusCode: resp.StatusCode,
			Message:    string(respBody),
		}
	}

	return respBody, resp.StatusCode, nil
}

// APIError represents an error from the Argo API.
type APIError struct {
	StatusCode int
	Message    string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("argo API error (status %d): %s", e.StatusCode, e.Message)
}

// --- Workflow Operations ---

// WorkflowCreateRequest is the request body for creating a workflow.
type WorkflowCreateRequest struct {
	Workflow  model.WorkflowModel `json:"workflow"`
	Namespace string              `json:"namespace,omitempty"`
}

// CreateWorkflowFromModel submits a pre-built workflow model to the Argo server.
func (s *WorkflowsService) CreateWorkflowFromModel(ctx context.Context, wfModel model.WorkflowModel, namespace string) (model.WorkflowModel, error) {
	ns := namespace
	if ns == "" {
		ns = s.Namespace
	}

	body := WorkflowCreateRequest{Workflow: wfModel}
	respBody, _, err := s.doRequest(ctx, http.MethodPost, "/api/v1/workflows/"+ns, body)
	if err != nil {
		return model.WorkflowModel{}, err
	}

	var result model.WorkflowModel
	if err := json.Unmarshal(respBody, &result); err != nil {
		return model.WorkflowModel{}, fmt.Errorf("unmarshal response: %w", err)
	}
	return result, nil
}

// GetWorkflow retrieves a workflow by name.
func (s *WorkflowsService) GetWorkflow(ctx context.Context, name, namespace string) (model.WorkflowModel, error) {
	if namespace == "" {
		namespace = s.Namespace
	}
	respBody, _, err := s.doRequest(ctx, http.MethodGet, "/api/v1/workflows/"+namespace+"/"+name, nil)
	if err != nil {
		return model.WorkflowModel{}, err
	}

	var result model.WorkflowModel
	if err := json.Unmarshal(respBody, &result); err != nil {
		return model.WorkflowModel{}, fmt.Errorf("unmarshal response: %w", err)
	}
	return result, nil
}

// DeleteWorkflow deletes a workflow by name.
func (s *WorkflowsService) DeleteWorkflow(ctx context.Context, name, namespace string) error {
	if namespace == "" {
		namespace = s.Namespace
	}
	_, _, err := s.doRequest(ctx, http.MethodDelete, "/api/v1/workflows/"+namespace+"/"+name, nil)
	return err
}

// ListWorkflowsResponse is the response for listing workflows.
type ListWorkflowsResponse struct {
	Items []model.WorkflowModel `json:"items"`
}

// ListWorkflows lists workflows in a namespace.
func (s *WorkflowsService) ListWorkflows(ctx context.Context, namespace string) ([]model.WorkflowModel, error) {
	if namespace == "" {
		namespace = s.Namespace
	}
	respBody, _, err := s.doRequest(ctx, http.MethodGet, "/api/v1/workflows/"+namespace, nil)
	if err != nil {
		return nil, err
	}

	var result ListWorkflowsResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}
	return result.Items, nil
}

// LintWorkflowFromModel validates a pre-built workflow model with the Argo server.
func (s *WorkflowsService) LintWorkflowFromModel(ctx context.Context, wfModel model.WorkflowModel, namespace string) (model.WorkflowModel, error) {
	ns := namespace
	if ns == "" {
		ns = s.Namespace
	}

	body := WorkflowCreateRequest{Workflow: wfModel}
	respBody, _, err := s.doRequest(ctx, http.MethodPost, "/api/v1/workflows/"+ns+"/lint", body)
	if err != nil {
		return model.WorkflowModel{}, err
	}

	var result model.WorkflowModel
	if err := json.Unmarshal(respBody, &result); err != nil {
		return model.WorkflowModel{}, fmt.Errorf("unmarshal response: %w", err)
	}
	return result, nil
}

// --- High-Level Operations (accept Buildable) ---

// CreateWorkflow builds a Buildable and submits it to the Argo server.
func (s *WorkflowsService) CreateWorkflow(ctx context.Context, b Buildable) (model.WorkflowModel, error) {
	wfModel, err := b.Build()
	if err != nil {
		return model.WorkflowModel{}, fmt.Errorf("build workflow: %w", err)
	}
	return s.CreateWorkflowFromModel(ctx, wfModel, b.GetNamespace())
}

// LintWorkflow builds a Buildable and validates it with the Argo server.
func (s *WorkflowsService) LintWorkflow(ctx context.Context, b Buildable) (model.WorkflowModel, error) {
	wfModel, err := b.Build()
	if err != nil {
		return model.WorkflowModel{}, fmt.Errorf("build workflow: %w", err)
	}
	return s.LintWorkflowFromModel(ctx, wfModel, b.GetNamespace())
}

// --- Info Operations ---

// GetInfo returns server info.
func (s *WorkflowsService) GetInfo(ctx context.Context) (map[string]interface{}, error) {
	respBody, _, err := s.doRequest(ctx, http.MethodGet, "/api/v1/info", nil)
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// GetVersion returns server version.
func (s *WorkflowsService) GetVersion(ctx context.Context) (map[string]interface{}, error) {
	respBody, _, err := s.doRequest(ctx, http.MethodGet, "/api/v1/version", nil)
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}
	return result, nil
}
