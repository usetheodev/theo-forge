package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

// T6.1 regression (val-006-no-logger-injection-client).

type spyLogger struct {
	mu     sync.Mutex
	events []map[string]any
}

func (s *spyLogger) Debug(msg string, kv ...any) { s.record("debug", msg, kv) }
func (s *spyLogger) Error(msg string, kv ...any) { s.record("error", msg, kv) }
func (s *spyLogger) record(level, msg string, kv []any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	entry := map[string]any{"level": level, "msg": msg}
	for i := 0; i+1 < len(kv); i += 2 {
		if k, ok := kv[i].(string); ok {
			entry[k] = kv[i+1]
		}
	}
	s.events = append(s.events, entry)
}

func TestWorkflowsService_LoggerCalledOnSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"items":[]}`))
	}))
	t.Cleanup(server.Close)

	spy := &spyLogger{}
	svc := NewWorkflowsService(server.URL, "secret-token", "default")
	svc.Logger = spy
	if _, err := svc.ListWorkflows(context.Background(), "default"); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if len(spy.events) == 0 {
		t.Fatal("logger never called")
	}
	// EC: assert no event carries the Authorization header or body.
	for _, ev := range spy.events {
		for k, v := range ev {
			lk := strings.ToLower(k)
			if lk == "authorization" || lk == "token" {
				t.Errorf("logger leaked %q: %v", k, v)
			}
			if vs, ok := v.(string); ok && strings.Contains(vs, "secret-token") {
				t.Errorf("logger leaked token value in %q=%q", k, vs)
			}
		}
	}
}

func TestWorkflowsService_LoggerCalledOnError(t *testing.T) {
	// Point at a closed server so Do() returns an error immediately.
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	server.Close()

	spy := &spyLogger{}
	svc := NewWorkflowsService(server.URL, "secret", "default")
	svc.Logger = spy
	_, _ = svc.ListWorkflows(context.Background(), "default")
	gotError := false
	for _, ev := range spy.events {
		if ev["level"] == "error" {
			gotError = true
		}
	}
	if !gotError {
		t.Fatal("expected at least one error-level event from logger")
	}
}

// EC: a panicking logger MUST NOT crash the HTTP request path.
type panicLogger struct{}

func (panicLogger) Debug(string, ...any) { panic("debug-panic") }
func (panicLogger) Error(string, ...any) { panic("error-panic") }

func TestWorkflowsService_PanickingLoggerDoesNotCrashRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"items":[]}`))
	}))
	t.Cleanup(server.Close)

	svc := NewWorkflowsService(server.URL, "tok", "default")
	svc.Logger = panicLogger{}
	if _, err := svc.ListWorkflows(context.Background(), "default"); err != nil {
		t.Fatalf("panicking logger crashed the request: %v", err)
	}
}
