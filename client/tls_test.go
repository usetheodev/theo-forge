package client

import (
	"context"
	"crypto/tls"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/usetheodev/theo-forge/model"
)

// T2.4 regression tests (SEC-004, code-p4-verifyssl-dead, completeness-h1-verifyssl)
// and EC-5 (honor externally-set *http.Client).

func TestWorkflowsService_DefaultsToVerifyTLS(t *testing.T) {
	svc := NewWorkflowsService("http://example.test", "tok", "default")
	if !svc.VerifySSL {
		t.Fatal("default VerifySSL = false, want true")
	}
}

func TestWorkflowsService_VerifySSL_True_RejectsSelfSigned(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"items":[]}`))
	}))
	t.Cleanup(server.Close)

	svc := NewWorkflowsService(server.URL, "tok", "default")
	svc.VerifySSL = true
	_, err := svc.ListWorkflows(context.Background(), "default")
	if err == nil {
		t.Fatal("expected TLS verification error, got nil")
	}
	if !strings.Contains(err.Error(), "x509") && !strings.Contains(err.Error(), "certificate") {
		t.Fatalf("expected TLS-related error, got %v", err)
	}
}

func TestWorkflowsService_VerifySSL_False_AcceptsSelfSigned(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"items":[]}`))
	}))
	t.Cleanup(server.Close)

	svc := NewWorkflowsService(server.URL, "tok", "default")
	svc.VerifySSL = false
	if _, err := svc.ListWorkflows(context.Background(), "default"); err != nil {
		t.Fatalf("VerifySSL=false should bypass TLS verification, got err: %v", err)
	}
}

// EC-5: pre-set HTTPClient MUST NOT be silently overwritten by httpClient().
func TestWorkflowsService_HonorsExternalHTTPClient(t *testing.T) {
	customTransport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true, MinVersion: tls.VersionTLS12}, //nolint:gosec
	}
	custom := &http.Client{Transport: customTransport}
	svc := NewWorkflowsService("http://example.test", "tok", "default")
	svc.HTTPClient = custom
	// First call: must return custom, must NOT replace.
	if got := svc.httpClient(); got != custom {
		t.Fatal("first httpClient() did not return injected *http.Client")
	}
	// Second call: still the same pointer.
	if got := svc.httpClient(); got != custom {
		t.Fatal("second httpClient() replaced the injected client")
	}
}

// T2.5 regression test (SEC-005). Server returns more than the limit;
// readBoundedBody (and downstream the public client methods) must
// surface model.ErrResponseTooLarge.
func TestClient_RejectsOversizedBody(t *testing.T) {
	bigChunk := strings.Repeat("a", 1024)
	totalKiB := 33 * 1024 // 33 MiB
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// Write 33 MiB so we exceed the 32 MiB limit.
		for i := 0; i < totalKiB; i++ {
			_, _ = w.Write([]byte(bigChunk))
		}
	}))
	t.Cleanup(server.Close)

	svc := NewWorkflowsService(server.URL, "tok", "default")
	_, err := svc.ListWorkflows(context.Background(), "default")
	if !errors.Is(err, model.ErrResponseTooLarge) {
		t.Fatalf("got err %v, want ErrResponseTooLarge", err)
	}
}

func TestReadBoundedBody_AtLimitAccepted(t *testing.T) {
	const limit int64 = 1024
	r := strings.NewReader(strings.Repeat("a", int(limit)))
	body, err := readBoundedBody(r, limit)
	if err != nil {
		t.Fatalf("err = %v, want nil at exactly the limit", err)
	}
	if int64(len(body)) != limit {
		t.Fatalf("len = %d, want %d", len(body), limit)
	}
}

func TestReadBoundedBody_OverLimitRejected(t *testing.T) {
	const limit int64 = 1024
	r := strings.NewReader(strings.Repeat("a", int(limit+1)))
	_, err := readBoundedBody(r, limit)
	if !errors.Is(err, model.ErrResponseTooLarge) {
		t.Fatalf("err = %v, want ErrResponseTooLarge", err)
	}
}
