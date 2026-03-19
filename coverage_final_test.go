package forge

import (
	"context"
	"net/http"
	"testing"

	"github.com/usetheo/theo/forge/client"
)

// Cover volume BuildVolume error paths (no-name failures)
func TestHostPathVolumeNoNameFails(t *testing.T) {
	v := HostPathVolume{Path: "/data"}
	_, err := v.BuildVolume()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestSecretVolumeNoNameFails(t *testing.T) {
	v := SecretVolume{SecretName: "s"}
	_, err := v.BuildVolume()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestExistingVolumeNoNameFails(t *testing.T) {
	v := ExistingVolume{ClaimName: "pvc"}
	_, err := v.BuildVolume()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestPVCVolumeNoNameFails(t *testing.T) {
	v := PVCVolume{Size: "1Gi"}
	_, err := v.BuildVolume()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestPVCVolumeBuildPVCNoNameFails(t *testing.T) {
	v := PVCVolume{Size: "1Gi"}
	_, err := v.BuildPVC()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestNFSVolumeNoNameFails(t *testing.T) {
	v := NFSVolume{Server: "nfs", Path: "/data"}
	_, err := v.BuildVolume()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestConfigMapVolumeNoNameFails(t *testing.T) {
	v := ConfigMapVolume{}
	_, err := v.BuildVolume()
	if err == nil {
		t.Fatal("expected error")
	}
}

// Cover Parameter.String success path
func TestParameterStringSuccess(t *testing.T) {
	p := Parameter{Name: "test", Value: ptrStr("hello")}
	s, err := p.String()
	if err != nil {
		t.Fatal(err)
	}
	if s != "hello" {
		t.Errorf("got %q", s)
	}
}

// Cover service unmarshal error paths
func TestServiceCreateWorkflowBadResponse(t *testing.T) {
	svc := &client.WorkflowsService{
		Host:      "https://argo.example.com",
		Namespace: "default",
		HTTPClient: &mockHTTPClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				return mockResponse(200, "not-json-object"), nil
			},
		},
	}
	w := &Workflow{Name: "test", Entrypoint: "main", Templates: []Templatable{&Container{Name: "main", Image: "a"}}}
	_, err := svc.CreateWorkflow(context.Background(), w)
	if err == nil {
		t.Fatal("expected unmarshal error")
	}
}

func TestServiceListWorkflowsBadResponse(t *testing.T) {
	svc := &client.WorkflowsService{
		Host:      "https://argo.example.com",
		Namespace: "default",
		HTTPClient: &mockHTTPClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				return mockResponse(200, "not-json"), nil
			},
		},
	}
	_, err := svc.ListWorkflows(context.Background(), "")
	if err == nil {
		t.Fatal("expected unmarshal error")
	}
}

func TestServiceLintWorkflowBadResponse(t *testing.T) {
	svc := &client.WorkflowsService{
		Host:      "https://argo.example.com",
		Namespace: "default",
		HTTPClient: &mockHTTPClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				return mockResponse(200, "bad"), nil
			},
		},
	}
	w := &Workflow{Name: "test", Entrypoint: "main", Templates: []Templatable{&Container{Name: "main", Image: "a"}}}
	_, err := svc.LintWorkflow(context.Background(), w)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestServiceGetInfoBadResponse(t *testing.T) {
	svc := &client.WorkflowsService{
		Host: "https://argo.example.com",
		HTTPClient: &mockHTTPClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				return mockResponse(200, "bad"), nil
			},
		},
	}
	_, err := svc.GetInfo(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestServiceGetVersionBadResponse(t *testing.T) {
	svc := &client.WorkflowsService{
		Host: "https://argo.example.com",
		HTTPClient: &mockHTTPClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				return mockResponse(200, "bad"), nil
			},
		},
	}
	_, err := svc.GetVersion(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestServiceGetWorkflowBadResponse(t *testing.T) {
	svc := &client.WorkflowsService{
		Host:      "https://argo.example.com",
		Namespace: "default",
		HTTPClient: &mockHTTPClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				return mockResponse(200, "bad"), nil
			},
		},
	}
	_, err := svc.GetWorkflow(context.Background(), "test", "")
	if err == nil {
		t.Fatal("expected error")
	}
}

// Cover Steps.buildInputs/buildOutputs artifact paths
func TestStepsWithInputArtifacts(t *testing.T) {
	steps := &Steps{
		Name: "with-art-in",
		InputArtifacts: []ArtifactBuilder{
			&Artifact{Name: "data", Path: "/tmp/data"},
		},
	}
	tpl, err := steps.BuildTemplate()
	if err != nil {
		t.Fatal(err)
	}
	if tpl.Inputs == nil || len(tpl.Inputs.Artifacts) != 1 {
		t.Fatal("expected 1 input artifact")
	}
}

// Cover DAG.buildInputs artifact path
func TestDAGWithInputArtifacts(t *testing.T) {
	dag := &DAG{
		Name: "with-art-in",
		InputArtifacts: []ArtifactBuilder{
			&Artifact{Name: "data", Path: "/tmp/data"},
		},
	}
	tpl, err := dag.BuildTemplate()
	if err != nil {
		t.Fatal(err)
	}
	if tpl.Inputs == nil || len(tpl.Inputs.Artifacts) != 1 {
		t.Fatal("expected 1 input artifact")
	}
}

// Cover Script.buildInputs artifact path
func TestScriptWithInputArtifacts(t *testing.T) {
	s := &Script{
		Name:    "with-art-in",
		Image:   "python:3.11",
		Command: []string{"python"},
		Source:  "print('hi')",
		InputArtifacts: []ArtifactBuilder{
			&Artifact{Name: "model", Path: "/tmp/model.pkl"},
		},
	}
	tpl, err := s.BuildTemplate()
	if err != nil {
		t.Fatal(err)
	}
	if tpl.Inputs == nil || len(tpl.Inputs.Artifacts) != 1 {
		t.Fatal("expected 1 input artifact")
	}
}

// Cover ValidateResourceRequirements - limit-only validations
func TestValidateResourceLimitOnlyInvalid(t *testing.T) {
	err := ValidateResourceRequirements(ResourceRequirements{
		Limits: ResourceList{CPU: "abc"},
	})
	if err == nil {
		t.Fatal("expected error for invalid CPU limit")
	}
}

func TestValidateResourceMemoryLimitOnlyInvalid(t *testing.T) {
	err := ValidateResourceRequirements(ResourceRequirements{
		Limits: ResourceList{Memory: "500m"},
	})
	if err == nil {
		t.Fatal("expected error for invalid memory limit (decimal unit)")
	}
}

func TestValidateResourceEphemeralLimitOnlyInvalid(t *testing.T) {
	err := ValidateResourceRequirements(ResourceRequirements{
		Limits: ResourceList{EphemeralStorage: "abc"},
	})
	if err == nil {
		t.Fatal("expected error for invalid ephemeral limit")
	}
}

func TestValidateResourceEphemeralRequestInvalid(t *testing.T) {
	err := ValidateResourceRequirements(ResourceRequirements{
		Requests: ResourceList{EphemeralStorage: "abc"},
	})
	if err == nil {
		t.Fatal("expected error for invalid ephemeral request")
	}
}

func TestValidateResourceEphemeralLimitInvalid(t *testing.T) {
	err := ValidateResourceRequirements(ResourceRequirements{
		Requests: ResourceList{EphemeralStorage: "1Gi"},
		Limits:   ResourceList{EphemeralStorage: "abc"},
	})
	if err == nil {
		t.Fatal("expected error for invalid ephemeral limit with valid request")
	}
}
