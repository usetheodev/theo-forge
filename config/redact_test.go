package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/usetheodev/theo-forge/model"
)

// T2.6 regression tests for GlobalConfig (SEC-007).

func TestGlobalConfig_StringRedactsToken(t *testing.T) {
	cfg := New()
	cfg.SetToken("super-secret-bearer-12345")
	got := fmt.Sprintf("%v", cfg)
	if strings.Contains(got, "super-secret") {
		t.Fatalf("String() leaked token: %q", got)
	}
	if !strings.Contains(got, "Token:***") {
		t.Fatalf("String() did not redact: %q", got)
	}
}

func TestGlobalConfig_MarshalJSONRedactsToken(t *testing.T) {
	cfg := New()
	cfg.SetToken("super-secret-bearer-12345")
	b, err := json.Marshal(cfg)
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

func TestGlobalConfig_UnmarshalRejectsRedactedToken(t *testing.T) {
	data := []byte(`{"host":"http://x","token":"***","namespace":"default","verifySSL":true}`)
	cfg := New()
	err := json.Unmarshal(data, cfg)
	if !errors.Is(err, model.ErrRedactedTokenLoaded) {
		t.Fatalf("got %v, want ErrRedactedTokenLoaded", err)
	}
}
