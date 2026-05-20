package serialize

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/usetheodev/theo-forge/model"
)

// Coverage backfill for serialize (T7.3 round 2).
// Targets the 0% / 75% functions: WorkflowToDict, CronWorkflowToJSON,
// ClusterWorkflowTemplate{ToYAML,FromYAML,ToFile,FromFile}, and error paths.

func TestWorkflowToDict_OK(t *testing.T) {
	wf := model.WorkflowModel{Metadata: model.WorkflowMetadata{Name: "x"}}
	d, err := WorkflowToDict(wf)
	if err != nil {
		t.Fatal(err)
	}
	md, ok := d["metadata"].(map[string]interface{})
	if !ok {
		t.Fatalf("metadata missing or wrong type: %v", d)
	}
	if md["name"] != "x" {
		t.Errorf("name = %v, want x", md["name"])
	}
}

func TestCronWorkflowToJSON_OK(t *testing.T) {
	cw := model.CronWorkflowModel{Metadata: model.WorkflowMetadata{Name: "cw"}}
	j, err := CronWorkflowToJSON(cw)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(j, `"name"`) {
		t.Errorf("unexpected json: %s", j)
	}
}

func TestClusterWorkflowTemplate_RoundTrip(t *testing.T) {
	cwt := model.WorkflowTemplateModel{
		Kind:     "ClusterWorkflowTemplate",
		Metadata: model.WorkflowMetadata{Name: "cwt"},
	}
	y, err := ClusterWorkflowTemplateToYAML(cwt)
	if err != nil {
		t.Fatal(err)
	}
	back, err := ClusterWorkflowTemplateFromYAML(y)
	if err != nil {
		t.Fatalf("FromYAML err: %v", err)
	}
	if back.Metadata.Name != "cwt" {
		t.Errorf("round-trip name lost: %q", back.Metadata.Name)
	}
}

func TestClusterWorkflowTemplate_DefaultKindInjected(t *testing.T) {
	cwt := model.WorkflowTemplateModel{Metadata: model.WorkflowMetadata{Name: "cwt2"}}
	y, err := ClusterWorkflowTemplateToYAML(cwt)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(y, "ClusterWorkflowTemplate") {
		t.Errorf("expected default Kind injection, got: %s", y)
	}
}

func TestClusterWorkflowTemplate_FromYAML_RejectsWrongKind(t *testing.T) {
	wrong := "apiVersion: argoproj.io/v1alpha1\nkind: Workflow\nmetadata:\n  name: x\n"
	_, err := ClusterWorkflowTemplateFromYAML(wrong)
	if err == nil {
		t.Fatal("expected error for wrong kind")
	}
	if !strings.Contains(err.Error(), "ClusterWorkflowTemplate") {
		t.Errorf("error should mention ClusterWorkflowTemplate: %v", err)
	}
}

func TestClusterWorkflowTemplate_FromYAML_AcceptsEmptyKind(t *testing.T) {
	y := "apiVersion: argoproj.io/v1alpha1\nmetadata:\n  name: x\n"
	if _, err := ClusterWorkflowTemplateFromYAML(y); err != nil {
		t.Fatalf("empty kind should be accepted: %v", err)
	}
}

func TestClusterWorkflowTemplateToFile_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path, err := ClusterWorkflowTemplateToFile("apiVersion: v1\nkind: ClusterWorkflowTemplate\nmetadata:\n  name: x\n", dir, "", "x")
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(path) != "x.yaml" {
		t.Errorf("default filename from name: %q", filepath.Base(path))
	}
	back, err := ClusterWorkflowTemplateFromFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if back.Metadata.Name != "x" {
		t.Errorf("round-trip name lost: %q", back.Metadata.Name)
	}
}

func TestClusterWorkflowTemplateToFile_ExplicitName(t *testing.T) {
	dir := t.TempDir()
	path, err := ClusterWorkflowTemplateToFile("kind: ClusterWorkflowTemplate\n", dir, "explicit.yml", "ignored")
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(path) != "explicit.yml" {
		t.Errorf("explicit filename preserved: got %q", filepath.Base(path))
	}
}

func TestClusterWorkflowTemplateToFile_AppendsExtension(t *testing.T) {
	dir := t.TempDir()
	path, err := ClusterWorkflowTemplateToFile("kind: ClusterWorkflowTemplate\n", dir, "noext", "")
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Ext(path) != ".yaml" {
		t.Errorf("expected .yaml extension, got %q", path)
	}
}

func TestWorkflowFromYAML_RejectsMalformed(t *testing.T) {
	_, err := WorkflowFromYAML("not: valid: yaml: ::")
	if err == nil {
		t.Fatal("expected parse error")
	}
}

func TestWorkflowFromJSON_RejectsMalformed(t *testing.T) {
	_, err := WorkflowFromJSON("{not json")
	if err == nil {
		t.Fatal("expected parse error")
	}
}

func TestCronWorkflowFromYAML_RejectsMalformed(t *testing.T) {
	_, err := CronWorkflowFromYAML("not: valid: yaml: ::")
	if err == nil {
		t.Fatal("expected parse error")
	}
}

func TestWorkflowToFile_RejectsBadOutputDir(t *testing.T) {
	// Use a path that contains a NUL byte — guaranteed to fail on POSIX.
	_, err := WorkflowToFile("", "/nope\x00bad", "x.yaml", "", "")
	if err == nil {
		t.Fatal("expected error on invalid output dir")
	}
}

func TestWorkflowFromFile_NotFound(t *testing.T) {
	_, err := WorkflowFromFile(filepath.Join(t.TempDir(), "missing.yaml"))
	if err == nil {
		t.Fatal("expected ENOENT")
	}
	if !os.IsNotExist(err) && !strings.Contains(err.Error(), "no such file") {
		t.Errorf("expected file-not-found, got: %v", err)
	}
}
