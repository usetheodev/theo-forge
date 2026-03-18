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

func ptrBool(b bool) *bool { return &b }
