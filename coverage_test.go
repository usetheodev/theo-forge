package forge

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/usetheo/theo/forge/client"
	"github.com/usetheo/theo/forge/expr"
)

// Cover ParamRef
func TestParamRef(t *testing.T) {
	got := expr.ParamRef("inputs.parameters.msg")
	want := "{{inputs.parameters.msg}}"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// Cover APIError.Error
func TestAPIErrorString(t *testing.T) {
	e := &client.APIError{StatusCode: 404, Message: "not found"}
	if !strings.Contains(e.Error(), "404") {
		t.Errorf("error = %q", e.Error())
	}
	if !strings.Contains(e.Error(), "not found") {
		t.Errorf("error = %q", e.Error())
	}
}

// Cover GetVersion
func TestServiceGetVersion(t *testing.T) {
	svc := &client.WorkflowsService{
		Host: "https://argo.example.com",
		HTTPClient: &mockHTTPClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				if req.URL.Path != "/api/v1/version" {
					t.Errorf("path = %q", req.URL.Path)
				}
				return mockResponse(200, map[string]interface{}{"version": "v3.5.0"}), nil
			},
		},
	}
	v, err := svc.GetVersion(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if v["version"] != "v3.5.0" {
		t.Errorf("version = %v", v["version"])
	}
}

// Cover FromJSON
func TestFromJSON(t *testing.T) {
	w := &Workflow{
		Name:       "json-roundtrip",
		Entrypoint: "main",
		Templates:  []Templatable{&Container{Name: "main", Image: "alpine"}},
	}
	jsonStr, err := w.ToJSON()
	if err != nil {
		t.Fatal(err)
	}
	model, err := FromJSON(jsonStr)
	if err != nil {
		t.Fatal(err)
	}
	if model.Metadata.Name != "json-roundtrip" {
		t.Errorf("name = %q", model.Metadata.Name)
	}
}

func TestFromJSONInvalid(t *testing.T) {
	_, err := FromJSON("{invalid json")
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

// Cover PVCVolume.BuildVolume
func TestPVCVolumeBuildVolume(t *testing.T) {
	v := PVCVolume{
		BaseVolume:       BaseVolume{Name: "data", MountPath: "/data"},
		Size:             "10Gi",
		StorageClassName: "standard",
		AccessModes:      []AccessMode{ReadWriteOnce},
	}
	vol, err := v.BuildVolume()
	if err != nil {
		t.Fatal(err)
	}
	if vol.Name != "data" {
		t.Errorf("name = %q", vol.Name)
	}
	if vol.PersistentVolumeClaim == nil {
		t.Fatal("expected PVC ref")
	}
	if vol.PersistentVolumeClaim.ClaimName != "data" {
		t.Errorf("claimName = %q", vol.PersistentVolumeClaim.ClaimName)
	}
}

// Cover Expr.C with int64 and float64 branches
func TestExprConstantInt64(t *testing.T) {
	e := expr.C(int64(42))
	if e.String() != "42" {
		t.Errorf("got %q", e.String())
	}
}

func TestExprConstantFloat64(t *testing.T) {
	e := expr.C(float64(3.14))
	if e.String() != "3.14" {
		t.Errorf("got %q", e.String())
	}
}

// Cover WorkflowTemplate name-too-long validation
func TestWorkflowTemplateNameTooLong(t *testing.T) {
	wt := &WorkflowTemplate{
		Name:       strings.Repeat("a", NameLimit+1),
		Entrypoint: "main",
	}
	_, err := wt.Build()
	if err == nil {
		t.Fatal("expected error for name too long")
	}
}

// Cover HTTPTemplate with inputs and outputs
func TestHTTPTemplateWithInputsOutputs(t *testing.T) {
	h := &HTTPTemplate{
		Name:   "with-io",
		URL:    "https://example.com/api",
		Method: "POST",
		Inputs: []Parameter{{Name: "payload", Value: ptrStr("{}")}},
		Outputs: []Parameter{{
			Name:      "status",
			ValueFrom: &ValueFrom{Expression: "response.statusCode"},
		}},
	}
	tpl, err := h.BuildTemplate()
	if err != nil {
		t.Fatal(err)
	}
	if tpl.Inputs == nil || len(tpl.Inputs.Parameters) != 1 {
		t.Fatal("expected 1 input")
	}
	if tpl.Outputs == nil || len(tpl.Outputs.Parameters) != 1 {
		t.Fatal("expected 1 output")
	}
}

// Cover GlobalConfig.GetImage empty fallback
func TestGlobalConfigGetImageFallback(t *testing.T) {
	cfg := GetGlobalConfig()
	defer cfg.Reset()
	cfg.Image = ""
	if cfg.GetImage() != "python:3.11" {
		t.Errorf("fallback = %q", cfg.GetImage())
	}
}

// Cover ContainerSet BuildTemplate output path
func TestContainerSetWithOutputs(t *testing.T) {
	cs := &ContainerSet{
		Name: "with-out",
		Containers: []ContainerNode{
			{Name: "main", Image: "alpine"},
		},
		Outputs: []Parameter{{Name: "result", ValueFrom: &ValueFrom{Path: "/tmp/out"}}},
	}
	tpl, err := cs.BuildTemplate()
	if err != nil {
		t.Fatal(err)
	}
	if tpl.Outputs == nil || len(tpl.Outputs.Parameters) != 1 {
		t.Fatal("expected 1 output")
	}
}

// Cover ContainerSet with retry
func TestContainerSetWithRetry(t *testing.T) {
	limit := 2
	cs := &ContainerSet{
		Name: "with-retry",
		Containers: []ContainerNode{
			{Name: "main", Image: "alpine"},
		},
		RetryStrategy: &RetryStrategy{Limit: &limit},
	}
	tpl, err := cs.BuildTemplate()
	if err != nil {
		t.Fatal(err)
	}
	if tpl.RetryStrategy == nil {
		t.Fatal("expected retry strategy")
	}
}
