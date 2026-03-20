package forge

import (
	"encoding/json"
	"testing"
)

func TestEnvBuild(t *testing.T) {
	e := Env{Name: "MY_VAR", Value: "hello"}
	model := e.Build()
	if model.Name != "MY_VAR" {
		t.Errorf("name = %q, want 'MY_VAR'", model.Name)
	}
	if model.Value == nil || *model.Value != "hello" {
		t.Errorf("value = %v, want 'hello'", model.Value)
	}
}

func TestSecretEnvBuild(t *testing.T) {
	e := SecretEnv{
		Name:       "DB_PASS",
		SecretName: "db-credentials",
		SecretKey:  "password",
	}
	model := e.Build()
	if model.Name != "DB_PASS" {
		t.Errorf("name = %q", model.Name)
	}
	if model.ValueFrom == nil {
		t.Fatal("expected valueFrom to be set")
	}
	if model.ValueFrom.SecretKeyRef == nil {
		t.Fatal("expected secretKeyRef to be set")
	}
	if model.ValueFrom.SecretKeyRef.Name != "db-credentials" {
		t.Errorf("secretName = %q", model.ValueFrom.SecretKeyRef.Name)
	}
	if model.ValueFrom.SecretKeyRef.Key != "password" {
		t.Errorf("secretKey = %q", model.ValueFrom.SecretKeyRef.Key)
	}
}

func TestConfigMapEnvBuild(t *testing.T) {
	e := ConfigMapEnv{
		Name:           "APP_CONFIG",
		ConfigMapName:  "app-config",
		ConfigMapKey:   "setting",
	}
	model := e.Build()
	if model.Name != "APP_CONFIG" {
		t.Errorf("name = %q", model.Name)
	}
	if model.ValueFrom == nil {
		t.Fatal("expected valueFrom")
	}
	if model.ValueFrom.ConfigMapKeyRef == nil {
		t.Fatal("expected configMapKeyRef")
	}
	if model.ValueFrom.ConfigMapKeyRef.Name != "app-config" {
		t.Errorf("configMapName = %q", model.ValueFrom.ConfigMapKeyRef.Name)
	}
}

func TestFieldEnvBuild(t *testing.T) {
	e := FieldEnv{
		Name:      "NODE_NAME",
		FieldPath: "spec.nodeName",
	}
	model := e.Build()
	if model.ValueFrom == nil || model.ValueFrom.FieldRef == nil {
		t.Fatal("expected fieldRef")
	}
	if model.ValueFrom.FieldRef.FieldPath != "spec.nodeName" {
		t.Errorf("fieldPath = %q", model.ValueFrom.FieldRef.FieldPath)
	}
}

func TestResourceEnvBuild(t *testing.T) {
	e := ResourceEnv{
		Name:          "CPU_LIMIT",
		Resource:      "limits.cpu",
		ContainerName: "main",
	}
	model := e.Build()
	if model.ValueFrom == nil || model.ValueFrom.ResourceFieldRef == nil {
		t.Fatal("expected resourceFieldRef")
	}
	if model.ValueFrom.ResourceFieldRef.Resource != "limits.cpu" {
		t.Errorf("resource = %q", model.ValueFrom.ResourceFieldRef.Resource)
	}
	if model.ValueFrom.ResourceFieldRef.ContainerName != "main" {
		t.Errorf("containerName = %q", model.ValueFrom.ResourceFieldRef.ContainerName)
	}
}

func TestEnvModelJSON(t *testing.T) {
	e := SecretEnv{
		Name:       "TOKEN",
		SecretName: "auth",
		SecretKey:  "token",
	}
	model := e.Build()
	data, err := json.Marshal(model)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatal(err)
	}
	if m["name"] != "TOKEN" {
		t.Errorf("json name = %v", m["name"])
	}
}
