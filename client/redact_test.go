package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/usetheodev/theo-forge/model"
)

// T2.6 regression tests (SEC-006, SEC-007) + EC-6 (UnmarshalJSON rejects "***").

func TestWorkflowsService_StringRedactsToken(t *testing.T) {
	svc := NewWorkflowsService("http://argo.example", "super-secret-bearer-12345", "default")
	got := fmt.Sprintf("%v", svc)
	if strings.Contains(got, "super-secret") {
		t.Fatalf("String() leaked token: %q", got)
	}
	if !strings.Contains(got, "Token:***") {
		t.Fatalf("String() did not redact: %q", got)
	}
}

func TestWorkflowsService_StringEmptyTokenShowsEmpty(t *testing.T) {
	svc := NewWorkflowsService("http://argo.example", "", "default")
	got := fmt.Sprintf("%v", svc)
	if !strings.Contains(got, "<empty>") {
		t.Fatalf("expected <empty> for unset token, got %q", got)
	}
}

func TestWorkflowsService_MarshalJSONRedactsToken(t *testing.T) {
	svc := NewWorkflowsService("http://argo.example", "super-secret-bearer-12345", "default")
	b, err := json.Marshal(svc)
	if err != nil {
		t.Fatalf("Marshal err: %v", err)
	}
	if strings.Contains(string(b), "super-secret") {
		t.Fatalf("MarshalJSON leaked token: %s", b)
	}
	if !strings.Contains(string(b), `"token":"***"`) {
		t.Fatalf("MarshalJSON did not redact: %s", b)
	}
}

// EC-6: round-tripping the marshaled output (where Token is "***") MUST fail
// loudly, not silently load a useless token.
func TestWorkflowsService_UnmarshalRejectsRedactedToken(t *testing.T) {
	data := []byte(`{"host":"http://x","token":"***","namespace":"default","verifySSL":true}`)
	var svc WorkflowsService
	err := json.Unmarshal(data, &svc)
	if !errors.Is(err, model.ErrRedactedTokenLoaded) {
		t.Fatalf("got %v, want ErrRedactedTokenLoaded", err)
	}
}

func TestWorkflowsService_UnmarshalAcceptsRealToken(t *testing.T) {
	data := []byte(`{"host":"http://x","token":"real-token","namespace":"default","verifySSL":true}`)
	var svc WorkflowsService
	if err := json.Unmarshal(data, &svc); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if svc.Token != "real-token" {
		t.Fatalf("Token = %q, want real-token", svc.Token)
	}
}
