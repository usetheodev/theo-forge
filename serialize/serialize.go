// Package serialize provides serialization and deserialization functions
// for Argo Workflow model types (YAML, JSON, file I/O).
package serialize

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/usetheodev/theo-forge/model"
	"sigs.k8s.io/yaml"
)

// WorkflowToYAML converts a WorkflowModel to a YAML string.
func WorkflowToYAML(m model.WorkflowModel) (string, error) {
	data, err := yaml.Marshal(m)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// WorkflowToJSON converts a WorkflowModel to an indented JSON string.
func WorkflowToJSON(m model.WorkflowModel) (string, error) {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// WorkflowToDict converts a WorkflowModel to a map (via JSON round-trip).
func WorkflowToDict(m model.WorkflowModel) (map[string]interface{}, error) {
	data, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// WorkflowFromYAML creates a WorkflowModel from a YAML string.
func WorkflowFromYAML(yamlStr string) (model.WorkflowModel, error) {
	var m model.WorkflowModel
	if err := yaml.Unmarshal([]byte(yamlStr), &m); err != nil {
		return model.WorkflowModel{}, err
	}
	return m, nil
}

// WorkflowFromJSON creates a WorkflowModel from a JSON string.
func WorkflowFromJSON(jsonStr string) (model.WorkflowModel, error) {
	var m model.WorkflowModel
	if err := json.Unmarshal([]byte(jsonStr), &m); err != nil {
		return model.WorkflowModel{}, err
	}
	return m, nil
}

// WorkflowTemplateToYAML converts a WorkflowTemplateModel to a YAML string.
func WorkflowTemplateToYAML(m model.WorkflowTemplateModel) (string, error) {
	data, err := yaml.Marshal(m)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// CronWorkflowToYAML converts a CronWorkflowModel to a YAML string.
func CronWorkflowToYAML(m model.CronWorkflowModel) (string, error) {
	data, err := yaml.Marshal(m)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// CronWorkflowToJSON converts a CronWorkflowModel to an indented JSON string.
func CronWorkflowToJSON(m model.CronWorkflowModel) (string, error) {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// WorkflowTemplateFromYAML creates a WorkflowTemplateModel from a YAML string.
func WorkflowTemplateFromYAML(yamlStr string) (model.WorkflowTemplateModel, error) {
	var m model.WorkflowTemplateModel
	if err := yaml.Unmarshal([]byte(yamlStr), &m); err != nil {
		return model.WorkflowTemplateModel{}, err
	}
	return m, nil
}

// CronWorkflowFromYAML creates a CronWorkflowModel from a YAML string.
func CronWorkflowFromYAML(yamlStr string) (model.CronWorkflowModel, error) {
	var m model.CronWorkflowModel
	if err := yaml.Unmarshal([]byte(yamlStr), &m); err != nil {
		return model.CronWorkflowModel{}, err
	}
	return m, nil
}

// WorkflowToFile writes a workflow YAML to a file.
// If fileName is empty, a filename is derived from wfName or generateName.
func WorkflowToFile(yamlStr, outputDir, fileName, wfName, generateName string) (string, error) {
	if fileName == "" {
		n := wfName
		if n == "" {
			n = strings.TrimSuffix(generateName, "-")
		}
		fileName = n + ".yaml"
	}
	if !strings.HasSuffix(fileName, ".yaml") && !strings.HasSuffix(fileName, ".yml") {
		fileName += ".yaml"
	}

	absDir, err := filepath.Abs(outputDir)
	if err != nil {
		return "", fmt.Errorf("resolve output directory: %w", err)
	}

	if err := os.MkdirAll(absDir, 0o755); err != nil {
		return "", fmt.Errorf("create output directory: %w", err)
	}

	path := filepath.Join(absDir, fileName)
	if err := os.WriteFile(path, []byte(yamlStr), 0o644); err != nil {
		return "", fmt.Errorf("write file: %w", err)
	}

	return path, nil
}

// WorkflowFromFile reads a WorkflowModel from a YAML file.
func WorkflowFromFile(path string) (model.WorkflowModel, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return model.WorkflowModel{}, fmt.Errorf("read file: %w", err)
	}
	var m model.WorkflowModel
	if err := yaml.Unmarshal(data, &m); err != nil {
		return model.WorkflowModel{}, fmt.Errorf("unmarshal YAML: %w", err)
	}
	return m, nil
}
