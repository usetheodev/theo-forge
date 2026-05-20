package serialize

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/usetheodev/theo-forge/model"
)

// Phase 7 (T7.3) coverage backfill for serialize package.

func TestWorkflowToYAML_RoundTrip(t *testing.T) {
	wf := model.WorkflowModel{Metadata: model.WorkflowMetadata{Name: "x"}}
	y, err := WorkflowToYAML(wf)
	if err != nil {
		t.Fatal(err)
	}
	got, err := WorkflowFromYAML(y)
	if err != nil {
		t.Fatal(err)
	}
	if got.Metadata.Name != "x" {
		t.Errorf("round-trip lost Name: %q", got.Metadata.Name)
	}
}

func TestWorkflowToJSON_RoundTrip(t *testing.T) {
	wf := model.WorkflowModel{Metadata: model.WorkflowMetadata{Name: "x"}}
	j, err := WorkflowToJSON(wf)
	if err != nil {
		t.Fatal(err)
	}
	got, err := WorkflowFromJSON(j)
	if err != nil {
		t.Fatal(err)
	}
	if got.Metadata.Name != "x" {
		t.Errorf("round-trip lost Name: %q", got.Metadata.Name)
	}
}

func TestWorkflowTemplateToYAML_RoundTrip(t *testing.T) {
	wt := model.WorkflowTemplateModel{Metadata: model.WorkflowMetadata{Name: "wt"}}
	y, err := WorkflowTemplateToYAML(wt)
	if err != nil {
		t.Fatal(err)
	}
	got, err := WorkflowTemplateFromYAML(y)
	if err != nil {
		t.Fatal(err)
	}
	if got.Metadata.Name != "wt" {
		t.Errorf("round-trip lost Name: %q", got.Metadata.Name)
	}
}

func TestCronWorkflowToYAML_RoundTrip(t *testing.T) {
	cw := model.CronWorkflowModel{Metadata: model.WorkflowMetadata{Name: "cw"}}
	y, err := CronWorkflowToYAML(cw)
	if err != nil {
		t.Fatal(err)
	}
	got, err := CronWorkflowFromYAML(y)
	if err != nil {
		t.Fatal(err)
	}
	if got.Metadata.Name != "cw" {
		t.Errorf("round-trip lost Name: %q", got.Metadata.Name)
	}
}

func TestWorkflowToFile_WritesValidYAML(t *testing.T) {
	dir := t.TempDir()
	path, err := WorkflowToFile("apiVersion: v1\n", dir, "x.yaml", "", "")
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path) //nolint:gosec // path is from t.TempDir
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "apiVersion") {
		t.Errorf("file content unexpected: %q", string(data))
	}
}

func TestWorkflowToFile_DerivesNameFromWfName(t *testing.T) {
	dir := t.TempDir()
	path, err := WorkflowToFile("apiVersion: v1\n", dir, "", "my-wf", "")
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(path) != "my-wf.yaml" {
		t.Errorf("derived name = %q, want \"my-wf.yaml\"", filepath.Base(path))
	}
}

func TestWorkflowToFile_DerivesNameFromGenerateName(t *testing.T) {
	dir := t.TempDir()
	path, err := WorkflowToFile("apiVersion: v1\n", dir, "", "", "auto-")
	if err != nil {
		t.Fatal(err)
	}
	if filepath.Base(path) != "auto.yaml" {
		t.Errorf("derived name = %q, want \"auto.yaml\" (trailing dash stripped)", filepath.Base(path))
	}
}

func TestWorkflowFromFile_ReadsBackWhatWeWrote(t *testing.T) {
	dir := t.TempDir()
	path, err := WorkflowToFile("apiVersion: argoproj.io/v1alpha1\nkind: Workflow\nmetadata:\n  name: roundtrip\n", dir, "x.yaml", "", "")
	if err != nil {
		t.Fatal(err)
	}
	wf, err := WorkflowFromFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if wf.Metadata.Name != "roundtrip" {
		t.Errorf("metadata.name = %q, want roundtrip", wf.Metadata.Name)
	}
}
