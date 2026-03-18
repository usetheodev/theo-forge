package forge

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"sigs.k8s.io/yaml"
)

// ToFile writes the workflow YAML to a file.
// If name is empty, the workflow name is used as the filename.
func (w *Workflow) ToFile(outputDir string, name string) (string, error) {
	yamlStr, err := w.ToYAML()
	if err != nil {
		return "", err
	}

	if name == "" {
		n := w.Name
		if n == "" {
			n = strings.TrimSuffix(w.GenerateName, "-")
		}
		name = n + ".yaml"
	}
	if !strings.HasSuffix(name, ".yaml") && !strings.HasSuffix(name, ".yml") {
		name += ".yaml"
	}

	absDir, err := filepath.Abs(outputDir)
	if err != nil {
		return "", fmt.Errorf("resolve output directory: %w", err)
	}

	if err := os.MkdirAll(absDir, 0o755); err != nil {
		return "", fmt.Errorf("create output directory: %w", err)
	}

	path := filepath.Join(absDir, name)
	if err := os.WriteFile(path, []byte(yamlStr), 0o644); err != nil {
		return "", fmt.Errorf("write file: %w", err)
	}

	return path, nil
}

// FromFile reads a WorkflowModel from a YAML file.
func FromFile(path string) (WorkflowModel, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return WorkflowModel{}, fmt.Errorf("read file: %w", err)
	}
	var model WorkflowModel
	if err := yaml.Unmarshal(data, &model); err != nil {
		return WorkflowModel{}, fmt.Errorf("unmarshal YAML: %w", err)
	}
	return model, nil
}
