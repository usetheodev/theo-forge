package client

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/usetheodev/theo-forge/model"
)

// defaultClientTimeout applies to HTTP requests when callers do not
// supply their own *http.Client via the HTTPClient field.
const defaultClientTimeout = 30 * time.Second

// maxResponseBodyBytes is the maximum response body the client will read.
// Bounded read prevents memory exhaustion via malicious or buggy server
// responses (T2.5 / SEC-005). 32 MiB is comfortably above typical Argo
// responses and below realistic OOM thresholds.
const maxResponseBodyBytes = 32 << 20

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
	// Logger receives structured Debug/Error events. Nil falls back to
	// NoopLogger (silent). NEVER logs the Authorization header or request
	// body. (T6.1 / ADR-007 / val-006)
	Logger Logger
}

// logger returns Logger or a NoopLogger fallback.
func (s *WorkflowsService) logger() Logger {
	if s.Logger == nil {
		return NoopLogger{}
	}
	return s.Logger
}

// NewWorkflowsService creates a new WorkflowsService.
// VerifySSL defaults to true (TLS verification ON). The HTTPClient field
// is left nil and resolved lazily by httpClient(); see that method for the
// rules (T2.4 / SEC-004 + EC-5).
func NewWorkflowsService(host, token, namespace string) *WorkflowsService {
	return &WorkflowsService{
		Host:      strings.TrimRight(host, "/"),
		Token:     token,
		Namespace: namespace,
		VerifySSL: true,
	}
}

// httpClient returns the HTTP client used for requests.
//
// Resolution order (EC-5 from edge-case review):
//  1. If s.HTTPClient is non-nil, return it as-is (consumer-injected
//     clients are honored — e.g., for tracing, proxy, custom transport).
//  2. Otherwise, construct a default client wired to s.VerifySSL via
//     tls.Config (T2.4 / SEC-004). Memoize on s.HTTPClient.
//
// MinVersion is fixed to TLS 1.2 per
// .claude/rules/golang-conventions.md §HTTP client.
func (s *WorkflowsService) httpClient() HTTPClient {
	if s.HTTPClient != nil {
		return s.HTTPClient
	}
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: !s.VerifySSL, //nolint:gosec // intentionally configurable via VerifySSL
			MinVersion:         tls.VersionTLS12,
		},
	}
	s.HTTPClient = &http.Client{Timeout: defaultClientTimeout, Transport: tr}
	return s.HTTPClient
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

func (s *WorkflowsService) doRequest(ctx context.Context, method, path string, body interface{}) ([]byte, error) {
	url := s.Host + path

	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	if token := s.FormatToken(); token != "" {
		req.Header.Set("Authorization", token)
	}

	start := time.Now()
	resp, err := s.httpClient().Do(req)
	if err != nil {
		logSafe(s.logger().Error, "argo request failed",
			"method", method, "path", path, "err", err.Error())
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	logSafe(s.logger().Debug, "argo response",
		"method", method, "path", path, "status", resp.StatusCode, "latency_ms", time.Since(start).Milliseconds())

	respBody, err := readBoundedBody(resp.Body, maxResponseBodyBytes)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		return respBody, &APIError{
			StatusCode: resp.StatusCode,
			Message:    string(respBody),
		}
	}

	return respBody, nil
}

// readBoundedBody reads up to limit bytes from r. If r would yield more
// than limit bytes (i.e., one additional Read would succeed), returns
// model.ErrResponseTooLarge (wrapped). T2.5 / SEC-005.
func readBoundedBody(r io.Reader, limit int64) ([]byte, error) {
	limited := io.LimitReader(r, limit+1) // +1 to detect overflow
	body, err := io.ReadAll(limited)
	if err != nil {
		return nil, fmt.Errorf("client: read response: %w", err)
	}
	if int64(len(body)) > limit {
		return nil, fmt.Errorf("%w (limit %d bytes)", model.ErrResponseTooLarge, limit)
	}
	return body, nil
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
// Returns model.ErrInvalidNamespace (wrapped) if namespace fails RFC1123 validation. (T2.3 / SEC-003).
func (s *WorkflowsService) CreateWorkflowFromModel(ctx context.Context, wfModel model.WorkflowModel, namespace string) (model.WorkflowModel, error) {
	ns := namespace
	if ns == "" {
		ns = s.Namespace
	}

	path, err := workflowsPath(ns)
	if err != nil {
		return model.WorkflowModel{}, err
	}
	body := WorkflowCreateRequest{Workflow: wfModel}
	respBody, err := s.doRequest(ctx, http.MethodPost, path, body)
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
// Returns model.ErrInvalidName / ErrInvalidNamespace (wrapped) on validation failure.
func (s *WorkflowsService) GetWorkflow(ctx context.Context, name, namespace string) (model.WorkflowModel, error) {
	if namespace == "" {
		namespace = s.Namespace
	}
	path, err := workflowPath(namespace, name)
	if err != nil {
		return model.WorkflowModel{}, err
	}
	respBody, err := s.doRequest(ctx, http.MethodGet, path, nil)
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
	path, err := workflowPath(namespace, name)
	if err != nil {
		return err
	}
	_, err = s.doRequest(ctx, http.MethodDelete, path, nil)
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
	path, err := workflowsPath(namespace)
	if err != nil {
		return nil, err
	}
	respBody, err := s.doRequest(ctx, http.MethodGet, path, nil)
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

	base, err := workflowsPath(ns)
	if err != nil {
		return model.WorkflowModel{}, err
	}
	body := WorkflowCreateRequest{Workflow: wfModel}
	respBody, err := s.doRequest(ctx, http.MethodPost, base+"/lint", body)
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

// --- Workflow Lifecycle Operations ---

// StopWorkflow stops a running workflow (allows running steps to complete).
func (s *WorkflowsService) StopWorkflow(ctx context.Context, name, namespace string) error {
	return s.lifecycleOp(ctx, name, namespace, "stop")
}

// TerminateWorkflow terminates a running workflow immediately.
func (s *WorkflowsService) TerminateWorkflow(ctx context.Context, name, namespace string) error {
	return s.lifecycleOp(ctx, name, namespace, "terminate")
}

// SuspendWorkflow suspends a running workflow.
func (s *WorkflowsService) SuspendWorkflow(ctx context.Context, name, namespace string) error {
	return s.lifecycleOp(ctx, name, namespace, "suspend")
}

// ResumeWorkflow resumes a suspended workflow.
func (s *WorkflowsService) ResumeWorkflow(ctx context.Context, name, namespace string) error {
	return s.lifecycleOp(ctx, name, namespace, "resume")
}

// lifecycleOp factors stop/terminate/suspend/resume — all share the same
// PUT /api/v1/workflows/{ns}/{name}/{action} shape with validation + URL escape.
func (s *WorkflowsService) lifecycleOp(ctx context.Context, name, namespace, action string) error {
	if namespace == "" {
		namespace = s.Namespace
	}
	path, err := workflowActionPath(namespace, name, action)
	if err != nil {
		return err
	}
	_, err = s.doRequest(ctx, http.MethodPut, path, nil)
	return err
}

// --- Info Operations ---

// ArgoServerInfo is the known-fields subset of the /api/v1/info response.
// Additional fields returned by the server are preserved in Extra. (T4.9 /
// code-p4-getinfo-getversion-untyped).
type ArgoServerInfo struct {
	ManagedNamespace string                 `json:"managedNamespace,omitempty"`
	Links            []ArgoServerInfoLink   `json:"links,omitempty"`
	ModalsClosed     []string               `json:"modals,omitempty"`
	NavColor         string                 `json:"navColor,omitempty"`
	Extra            map[string]interface{} `json:"-"`
}

// ArgoServerInfoLink mirrors the items in /api/v1/info "links".
type ArgoServerInfoLink struct {
	Name  string `json:"name,omitempty"`
	Scope string `json:"scope,omitempty"`
	URL   string `json:"url,omitempty"`
}

// ArgoServerVersion is the known-fields subset of the /api/v1/version response.
type ArgoServerVersion struct {
	Version      string `json:"version,omitempty"`
	BuildDate    string `json:"buildDate,omitempty"`
	GitCommit    string `json:"gitCommit,omitempty"`
	GitTag       string `json:"gitTag,omitempty"`
	GitTreeState string `json:"gitTreeState,omitempty"`
	GoVersion    string `json:"goVersion,omitempty"`
	Compiler     string `json:"compiler,omitempty"`
	Platform     string `json:"platform,omitempty"`
}

// GetInfo returns typed server info. (T4.9).
func (s *WorkflowsService) GetInfo(ctx context.Context) (ArgoServerInfo, error) {
	respBody, err := s.doRequest(ctx, http.MethodGet, "/api/v1/info", nil)
	if err != nil {
		return ArgoServerInfo{}, err
	}
	var result ArgoServerInfo
	if err := json.Unmarshal(respBody, &result); err != nil {
		return ArgoServerInfo{}, fmt.Errorf("unmarshal info: %w", err)
	}
	return result, nil
}

// GetVersion returns typed server version. (T4.9).
func (s *WorkflowsService) GetVersion(ctx context.Context) (ArgoServerVersion, error) {
	respBody, err := s.doRequest(ctx, http.MethodGet, "/api/v1/version", nil)
	if err != nil {
		return ArgoServerVersion{}, err
	}
	var result ArgoServerVersion
	if err := json.Unmarshal(respBody, &result); err != nil {
		return ArgoServerVersion{}, fmt.Errorf("unmarshal version: %w", err)
	}
	return result, nil
}
