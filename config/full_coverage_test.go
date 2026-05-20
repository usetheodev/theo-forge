package config

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/usetheodev/theo-forge/model"
)

// Coverage backfill for config (T7.2 round 2).

func TestGetImage_FallbackWhenEmpty(t *testing.T) {
	ResetGlobalConfigForTest(t)
	GetGlobal().SetImage("")
	if got := GetGlobal().GetImage(); got != "python:3.11" {
		t.Errorf("fallback expected python:3.11, got %q", got)
	}
}

func TestApplyTemplateDefaults_AppliesServiceAccount(t *testing.T) {
	cfg := New()
	cfg.serviceAccountName = "sa-x"
	tpl := &model.TemplateModel{Container: &model.ContainerModel{Image: "alpine"}}
	cfg.ApplyTemplateDefaults(tpl)
	if tpl.ServiceAccountName != "sa-x" {
		t.Errorf("ServiceAccountName not applied: %q", tpl.ServiceAccountName)
	}
}

func TestApplyTemplateDefaults_PreservesExplicitServiceAccount(t *testing.T) {
	cfg := New()
	cfg.serviceAccountName = "default-sa"
	tpl := &model.TemplateModel{ServiceAccountName: "explicit-sa", Container: &model.ContainerModel{Image: "alpine"}}
	cfg.ApplyTemplateDefaults(tpl)
	if tpl.ServiceAccountName != "explicit-sa" {
		t.Errorf("explicit ServiceAccountName overwritten: %q", tpl.ServiceAccountName)
	}
}

func TestApplyTemplateDefaults_ScriptReceivesDefaults(t *testing.T) {
	cfg := New()
	cfg.image = "python:3.12"
	cfg.imagePullPolicy = model.ImagePullAlways
	tpl := &model.TemplateModel{Script: &model.ScriptModel{}}
	cfg.ApplyTemplateDefaults(tpl)
	if tpl.Script.Image != "python:3.12" {
		t.Errorf("Script.Image = %q", tpl.Script.Image)
	}
	if tpl.Script.ImagePullPolicy != string(model.ImagePullAlways) {
		t.Errorf("Script.ImagePullPolicy = %q", tpl.Script.ImagePullPolicy)
	}
}

func TestApplyTemplateDefaults_ContainerImagePullPolicyDefault(t *testing.T) {
	cfg := New()
	cfg.imagePullPolicy = model.ImagePullIfNotPresent
	tpl := &model.TemplateModel{Container: &model.ContainerModel{Image: "alpine"}}
	cfg.ApplyTemplateDefaults(tpl)
	if tpl.Container.ImagePullPolicy != string(model.ImagePullIfNotPresent) {
		t.Errorf("Container.ImagePullPolicy = %q", tpl.Container.ImagePullPolicy)
	}
}

func TestString_NilReceiver(t *testing.T) {
	var nilCfg *GlobalConfig
	if got := nilCfg.String(); got != "<nil>" {
		t.Errorf("expected <nil>, got %q", got)
	}
}

func TestMarshalJSON_NilReceiver(t *testing.T) {
	var nilCfg *GlobalConfig
	b, err := nilCfg.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != "null" {
		t.Errorf("expected null, got %q", string(b))
	}
}

func TestUnmarshalJSON_AcceptsRealToken(t *testing.T) {
	data := []byte(`{"host":"https://x","token":"real","namespace":"ns","image":"alpine:3.18","imagePullPolicy":"Always","verifySSL":false}`)
	cfg := New()
	if err := json.Unmarshal(data, cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.token != "real" || cfg.image != "alpine:3.18" || cfg.verifySSL != false {
		t.Errorf("unmarshal lost fields: %+v", cfg)
	}
	if cfg.imagePullPolicy != model.ImagePullAlways {
		t.Errorf("ImagePullPolicy = %q", cfg.imagePullPolicy)
	}
}

func TestUnmarshalJSON_PreservesServiceAccountAndNamespace(t *testing.T) {
	data := []byte(`{"namespace":"argo","serviceAccountName":"my-sa","verifySSL":true}`)
	cfg := New()
	if err := json.Unmarshal(data, cfg); err != nil {
		t.Fatal(err)
	}
	if cfg.namespace != "argo" || cfg.serviceAccountName != "my-sa" {
		t.Errorf("fields lost: %+v", cfg)
	}
}

func TestUnmarshalJSON_RejectsMalformed(t *testing.T) {
	cfg := New()
	if err := json.Unmarshal([]byte("{bad"), cfg); err == nil {
		t.Fatal("expected parse error")
	}
}

func TestString_EmptyTokenShowsEmpty(t *testing.T) {
	cfg := New()
	s := cfg.String()
	if !strings.Contains(s, "<empty>") {
		t.Errorf("expected <empty>, got: %s", s)
	}
}

func TestUnmarshalJSON_RejectsRedactedToken_Error(t *testing.T) {
	cfg := New()
	err := json.Unmarshal([]byte(`{"token":"***"}`), cfg)
	if !errors.Is(err, model.ErrRedactedTokenLoaded) {
		t.Errorf("got %v, want ErrRedactedTokenLoaded", err)
	}
}
