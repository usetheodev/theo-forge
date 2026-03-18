package forge

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWorkflowToFile(t *testing.T) {
	tmpDir := t.TempDir()

	w := &Workflow{
		Name:       "file-test",
		Entrypoint: "main",
		Templates:  []Templatable{&Container{Name: "main", Image: "alpine"}},
	}

	path, err := w.ToFile(tmpDir, "")
	if err != nil {
		t.Fatal(err)
	}

	// Check file was created
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("file not created at %s", path)
	}

	// Check filename
	expectedName := "file-test.yaml"
	if filepath.Base(path) != expectedName {
		t.Errorf("filename = %q, want %q", filepath.Base(path), expectedName)
	}
}

func TestWorkflowToFileCustomName(t *testing.T) {
	tmpDir := t.TempDir()

	w := &Workflow{
		Name:       "test",
		Entrypoint: "main",
		Templates:  []Templatable{&Container{Name: "main", Image: "alpine"}},
	}

	path, err := w.ToFile(tmpDir, "custom-name.yaml")
	if err != nil {
		t.Fatal(err)
	}

	if filepath.Base(path) != "custom-name.yaml" {
		t.Errorf("filename = %q", filepath.Base(path))
	}
}

func TestWorkflowToFileGenerateName(t *testing.T) {
	tmpDir := t.TempDir()

	w := &Workflow{
		GenerateName: "my-wf-",
		Entrypoint:   "main",
		Templates:    []Templatable{&Container{Name: "main", Image: "alpine"}},
	}

	path, err := w.ToFile(tmpDir, "")
	if err != nil {
		t.Fatal(err)
	}

	if filepath.Base(path) != "my-wf.yaml" {
		t.Errorf("filename = %q, want 'my-wf.yaml'", filepath.Base(path))
	}
}

func TestWorkflowToFileAddsYAMLExtension(t *testing.T) {
	tmpDir := t.TempDir()

	w := &Workflow{
		Name:       "test",
		Entrypoint: "main",
		Templates:  []Templatable{&Container{Name: "main", Image: "alpine"}},
	}

	path, err := w.ToFile(tmpDir, "no-extension")
	if err != nil {
		t.Fatal(err)
	}

	if filepath.Base(path) != "no-extension.yaml" {
		t.Errorf("filename = %q", filepath.Base(path))
	}
}

func TestFromFile(t *testing.T) {
	tmpDir := t.TempDir()

	w := &Workflow{
		Name:       "round-trip-file",
		Namespace:  "argo",
		Entrypoint: "main",
		Templates:  []Templatable{&Container{Name: "main", Image: "alpine", Command: []string{"echo"}, Args: []string{"hello"}}},
	}

	path, err := w.ToFile(tmpDir, "")
	if err != nil {
		t.Fatal(err)
	}

	model, err := FromFile(path)
	if err != nil {
		t.Fatal(err)
	}

	if model.Metadata.Name != "round-trip-file" {
		t.Errorf("name = %q", model.Metadata.Name)
	}
	if model.Metadata.Namespace != "argo" {
		t.Errorf("namespace = %q", model.Metadata.Namespace)
	}
	if model.Spec.Entrypoint != "main" {
		t.Errorf("entrypoint = %q", model.Spec.Entrypoint)
	}
	if len(model.Spec.Templates) != 1 {
		t.Fatalf("templates = %d", len(model.Spec.Templates))
	}
	if model.Spec.Templates[0].Container == nil {
		t.Fatal("expected container template")
	}
	if model.Spec.Templates[0].Container.Image != "alpine" {
		t.Errorf("image = %q", model.Spec.Templates[0].Container.Image)
	}
}

func TestFromFileNotFound(t *testing.T) {
	_, err := FromFile("/nonexistent/path/file.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestFromFileInvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "invalid.yaml")
	os.WriteFile(path, []byte("{{{{invalid yaml"), 0o644)

	_, err := FromFile(path)
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}
