package forge

import (
	"encoding/json"
	"testing"
)

func TestArtifactNoNameCanBeCreated(t *testing.T) {
	a := Artifact{Path: "/tmp/path"}
	if a.Path != "/tmp/path" {
		t.Fatal("expected path to be set")
	}
}

func TestArtifactNoNameFailsBuildArtifact(t *testing.T) {
	a := Artifact{Path: "/tmp/path"}
	_, err := a.Build()
	if err == nil {
		t.Fatal("expected error when name is empty")
	}
	if err.Error() != "name cannot be empty when used" {
		t.Fatalf("unexpected error: %s", err.Error())
	}
}

func TestArtifactNoNamePassesWithName(t *testing.T) {
	a := Artifact{Path: "/tmp/path"}
	a2 := a.WithName("new")
	if a2.Name != "new" {
		t.Fatalf("expected name 'new', got '%s'", a2.Name)
	}
	if a2.Path != "/tmp/path" {
		t.Fatalf("expected path preserved, got '%s'", a2.Path)
	}
	// Original unchanged
	if a.Name != "" {
		t.Fatal("original artifact name should remain empty")
	}
}

func TestArtifactBuildWithName(t *testing.T) {
	a := Artifact{Name: "my-artifact", Path: "/tmp/output"}
	model, err := a.Build()
	if err != nil {
		t.Fatalf("Build() error: %v", err)
	}
	if model.Name != "my-artifact" {
		t.Errorf("name = %q, want 'my-artifact'", model.Name)
	}
	if model.Path != "/tmp/output" {
		t.Errorf("path = %q, want '/tmp/output'", model.Path)
	}
}

func TestArtifactOptionalField(t *testing.T) {
	a := Artifact{Name: "opt", Path: "/tmp/opt", Optional: ptrBool(true)}
	model, err := a.Build()
	if err != nil {
		t.Fatal(err)
	}
	if model.Optional == nil || !*model.Optional {
		t.Error("expected optional to be true")
	}
}

func TestS3ArtifactBuild(t *testing.T) {
	a := S3Artifact{
		Artifact: Artifact{Name: "s3-art", Path: "/tmp/data"},
		Bucket:   "my-bucket",
		Key:      "path/to/object",
		Endpoint: "s3.amazonaws.com",
	}
	model, err := a.Build()
	if err != nil {
		t.Fatal(err)
	}
	if model.Name != "s3-art" {
		t.Errorf("name = %q, want 's3-art'", model.Name)
	}
	if model.S3 == nil {
		t.Fatal("expected S3 field to be set")
	}
	if model.S3.Bucket != "my-bucket" {
		t.Errorf("bucket = %q, want 'my-bucket'", model.S3.Bucket)
	}
	if model.S3.Key != "path/to/object" {
		t.Errorf("key = %q, want 'path/to/object'", model.S3.Key)
	}
}

func TestGCSArtifactBuild(t *testing.T) {
	a := GCSArtifact{
		Artifact: Artifact{Name: "gcs-art", Path: "/tmp/data"},
		Bucket:   "my-gcs-bucket",
		Key:      "data/file.csv",
	}
	model, err := a.Build()
	if err != nil {
		t.Fatal(err)
	}
	if model.GCS == nil {
		t.Fatal("expected GCS field to be set")
	}
	if model.GCS.Bucket != "my-gcs-bucket" {
		t.Errorf("bucket = %q, want 'my-gcs-bucket'", model.GCS.Bucket)
	}
}

func TestHTTPArtifactBuild(t *testing.T) {
	a := HTTPArtifact{
		Artifact: Artifact{Name: "http-art", Path: "/tmp/data"},
		URL:      "https://example.com/data.tar.gz",
	}
	model, err := a.Build()
	if err != nil {
		t.Fatal(err)
	}
	if model.HTTP == nil {
		t.Fatal("expected HTTP field to be set")
	}
	if model.HTTP.URL != "https://example.com/data.tar.gz" {
		t.Errorf("url = %q, want 'https://example.com/data.tar.gz'", model.HTTP.URL)
	}
}

func TestGitArtifactBuild(t *testing.T) {
	a := GitArtifact{
		Artifact: Artifact{Name: "git-art", Path: "/tmp/repo"},
		Repo:     "https://github.com/example/repo.git",
		Revision: "main",
	}
	model, err := a.Build()
	if err != nil {
		t.Fatal(err)
	}
	if model.Git == nil {
		t.Fatal("expected Git field to be set")
	}
	if model.Git.Repo != "https://github.com/example/repo.git" {
		t.Errorf("repo = %q", model.Git.Repo)
	}
}

func TestRawArtifactBuild(t *testing.T) {
	a := RawArtifact{
		Artifact: Artifact{Name: "raw-art", Path: "/tmp/data"},
		Data:     "hello world",
	}
	model, err := a.Build()
	if err != nil {
		t.Fatal(err)
	}
	if model.Raw == nil {
		t.Fatal("expected Raw field to be set")
	}
	if model.Raw.Data != "hello world" {
		t.Errorf("data = %q, want 'hello world'", model.Raw.Data)
	}
}

func TestArtifactModelJSON(t *testing.T) {
	a := S3Artifact{
		Artifact: Artifact{Name: "s3-test", Path: "/data"},
		Bucket:   "bucket",
		Key:      "key",
	}
	model, err := a.Build()
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(model)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatal(err)
	}
	if m["name"] != "s3-test" {
		t.Errorf("json name = %v, want 's3-test'", m["name"])
	}
	s3, ok := m["s3"].(map[string]interface{})
	if !ok {
		t.Fatal("expected s3 to be a map")
	}
	if s3["bucket"] != "bucket" {
		t.Errorf("json s3.bucket = %v", s3["bucket"])
	}
}

// --- Additional artifact tests (consolidated from artifact_extra_test.go) ---

func TestAzureArtifactBuild(t *testing.T) {
	a := AzureArtifact{
		Artifact:  Artifact{Name: "azure-art", Path: "/tmp/data"},
		Endpoint:  "https://account.blob.core.windows.net",
		Container: "my-container",
		Blob:      "path/to/blob",
	}
	model, err := a.Build()
	if err != nil {
		t.Fatal(err)
	}
	if model.Azure == nil {
		t.Fatal("expected Azure field")
	}
	if model.Azure.Endpoint != "https://account.blob.core.windows.net" {
		t.Errorf("endpoint = %q", model.Azure.Endpoint)
	}
	if model.Azure.Container != "my-container" {
		t.Errorf("container = %q", model.Azure.Container)
	}
	if model.Azure.Blob != "path/to/blob" {
		t.Errorf("blob = %q", model.Azure.Blob)
	}
}

func TestOSSArtifactBuild(t *testing.T) {
	a := OSSArtifact{
		Artifact: Artifact{Name: "oss-art", Path: "/tmp/data"},
		Bucket:   "my-bucket",
		Key:      "path/to/object",
		Endpoint: "oss-cn-hangzhou.aliyuncs.com",
	}
	model, err := a.Build()
	if err != nil {
		t.Fatal(err)
	}
	if model.OSS == nil {
		t.Fatal("expected OSS field")
	}
	if model.OSS.Bucket != "my-bucket" {
		t.Errorf("bucket = %q", model.OSS.Bucket)
	}
	if model.OSS.Key != "path/to/object" {
		t.Errorf("key = %q", model.OSS.Key)
	}
}

func TestHDFSArtifactBuild(t *testing.T) {
	a := HDFSArtifact{
		Artifact:  Artifact{Name: "hdfs-art", Path: "/tmp/data"},
		HDFSPath:  "/data/output",
		Addresses: []string{"namenode:8020"},
		HDFSUser:  "hadoop",
	}
	model, err := a.Build()
	if err != nil {
		t.Fatal(err)
	}
	if model.HDFS == nil {
		t.Fatal("expected HDFS field")
	}
	if model.HDFS.Path != "/data/output" {
		t.Errorf("path = %q", model.HDFS.Path)
	}
	if len(model.HDFS.Addresses) != 1 || model.HDFS.Addresses[0] != "namenode:8020" {
		t.Errorf("addresses = %v", model.HDFS.Addresses)
	}
	if model.HDFS.HDFSUser != "hadoop" {
		t.Errorf("hdfsUser = %q", model.HDFS.HDFSUser)
	}
}

func TestAzureArtifactNoNameFails(t *testing.T) {
	a := AzureArtifact{Artifact: Artifact{Path: "/tmp"}, Endpoint: "e", Container: "c", Blob: "b"}
	_, err := a.Build()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestOSSArtifactNoNameFails(t *testing.T) {
	a := OSSArtifact{Artifact: Artifact{Path: "/tmp"}, Key: "k"}
	_, err := a.Build()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestHDFSArtifactNoNameFails(t *testing.T) {
	a := HDFSArtifact{Artifact: Artifact{Path: "/tmp"}, HDFSPath: "/data"}
	_, err := a.Build()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestContainerWithOutputArtifacts(t *testing.T) {
	c := &Container{
		Name:  "with-artifacts",
		Image: "alpine",
		OutputArtifacts: []ArtifactBuilder{
			&S3Artifact{
				Artifact: Artifact{Name: "output", Path: "/tmp/output"},
				Bucket:   "results",
				Key:      "run/output.tar.gz",
			},
		},
	}
	tpl, err := c.BuildTemplate()
	if err != nil {
		t.Fatal(err)
	}
	if tpl.Outputs == nil || len(tpl.Outputs.Artifacts) != 1 {
		t.Fatal("expected 1 output artifact")
	}
	if tpl.Outputs.Artifacts[0].S3 == nil {
		t.Error("expected S3 artifact")
	}
}

func TestContainerWithInputArtifacts(t *testing.T) {
	c := &Container{
		Name:  "with-input-art",
		Image: "alpine",
		InputArtifacts: []ArtifactBuilder{
			&HTTPArtifact{
				Artifact: Artifact{Name: "data", Path: "/tmp/data.csv"},
				URL:      "https://example.com/data.csv",
			},
		},
	}
	tpl, err := c.BuildTemplate()
	if err != nil {
		t.Fatal(err)
	}
	if tpl.Inputs == nil || len(tpl.Inputs.Artifacts) != 1 {
		t.Fatal("expected 1 input artifact")
	}
}

func TestScriptWithOutputArtifacts(t *testing.T) {
	s := &Script{
		Name:    "with-artifacts",
		Image:   "python:3.11",
		Command: []string{"python"},
		Source:  "open('/tmp/out.txt','w').write('result')",
		OutputArtifacts: []ArtifactBuilder{
			&Artifact{Name: "result", Path: "/tmp/out.txt"},
		},
	}
	tpl, err := s.BuildTemplate()
	if err != nil {
		t.Fatal(err)
	}
	if tpl.Outputs == nil || len(tpl.Outputs.Artifacts) != 1 {
		t.Fatal("expected 1 output artifact")
	}
}

func TestDAGWithOutputArtifacts(t *testing.T) {
	dag := &DAG{
		Name: "with-outputs",
		OutputArtifacts: []ArtifactBuilder{
			&Artifact{Name: "final-output", Path: "/tmp/final"},
		},
	}
	tpl, err := dag.BuildTemplate()
	if err != nil {
		t.Fatal(err)
	}
	if tpl.Outputs == nil || len(tpl.Outputs.Artifacts) != 1 {
		t.Fatal("expected 1 output artifact")
	}
}

func TestStepsWithOutputs(t *testing.T) {
	steps := &Steps{
		Name:    "with-outputs",
		Outputs: []Parameter{{Name: "result", ValueFrom: &ValueFrom{Expression: "steps.last.outputs.result"}}},
	}
	tpl, err := steps.BuildTemplate()
	if err != nil {
		t.Fatal(err)
	}
	if tpl.Outputs == nil || len(tpl.Outputs.Parameters) != 1 {
		t.Fatal("expected 1 output parameter")
	}
}

func TestContainerGetName(t *testing.T) {
	c := &Container{Name: "test", Image: "alpine"}
	if c.GetName() != "test" {
		t.Errorf("GetName = %q", c.GetName())
	}
}

func TestScriptGetName(t *testing.T) {
	s := &Script{Name: "test", Image: "alpine", Command: []string{"sh"}, Source: "echo"}
	if s.GetName() != "test" {
		t.Errorf("GetName = %q", s.GetName())
	}
}

func TestDAGGetName(t *testing.T) {
	d := &DAG{Name: "test"}
	if d.GetName() != "test" {
		t.Errorf("GetName = %q", d.GetName())
	}
}

func TestStepsGetName(t *testing.T) {
	s := &Steps{Name: "test"}
	if s.GetName() != "test" {
		t.Errorf("GetName = %q", s.GetName())
	}
}

func TestContainerSetGetName(t *testing.T) {
	cs := &ContainerSet{Name: "test", Containers: []ContainerNode{{Name: "c", Image: "a"}}}
	if cs.GetName() != "test" {
		t.Errorf("GetName = %q", cs.GetName())
	}
}

func TestResourceTemplateGetName(t *testing.T) {
	r := &ResourceTemplate{Name: "test", Action: "create", Manifest: "a: b"}
	if r.GetName() != "test" {
		t.Errorf("GetName = %q", r.GetName())
	}
}

func TestSuspendGetName(t *testing.T) {
	s := &Suspend{Name: "test"}
	if s.GetName() != "test" {
		t.Errorf("GetName = %q", s.GetName())
	}
}

func TestHTTPTemplateGetName(t *testing.T) {
	h := &HTTPTemplate{Name: "test", URL: "https://example.com"}
	if h.GetName() != "test" {
		t.Errorf("GetName = %q", h.GetName())
	}
}
