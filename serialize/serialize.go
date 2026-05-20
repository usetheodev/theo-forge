// Package serialize provides serialization and deserialization functions
// for Argo Workflow model types (YAML, JSON, file I/O).
package serialize

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"sigs.k8s.io/yaml"

	"github.com/usetheodev/theo-forge/model"
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

// T8.3 — ClusterWorkflowTemplate serialization parity with WorkflowTemplate.
// (completeness-h6-clusterwftemplate-serialize)
//
// ClusterWorkflowTemplate shares the WorkflowTemplateModel shape (cluster-scoped
// variant identified by Kind). The functions below mirror WorkflowTemplate's
// surface so callers writing/reading cluster templates have first-class APIs.

// ClusterWorkflowTemplateToYAML converts a ClusterWorkflowTemplateModel to YAML.
func ClusterWorkflowTemplateToYAML(m model.WorkflowTemplateModel) (string, error) {
	if m.Kind == "" {
		m.Kind = "ClusterWorkflowTemplate"
	}
	return WorkflowTemplateToYAML(m)
}

// ClusterWorkflowTemplateFromYAML creates a WorkflowTemplateModel from a YAML string,
// asserting the Kind is ClusterWorkflowTemplate (or empty).
func ClusterWorkflowTemplateFromYAML(yamlStr string) (model.WorkflowTemplateModel, error) {
	m, err := WorkflowTemplateFromYAML(yamlStr)
	if err != nil {
		return model.WorkflowTemplateModel{}, err
	}
	if m.Kind != "" && m.Kind != "ClusterWorkflowTemplate" {
		return model.WorkflowTemplateModel{}, fmt.Errorf("serialize: expected Kind=ClusterWorkflowTemplate, got %q", m.Kind)
	}
	return m, nil
}

// ClusterWorkflowTemplateToFile writes a cluster template YAML to a file under
// outputDir. Same containment guarantees as WorkflowToFile (see containedJoin).
func ClusterWorkflowTemplateToFile(yamlStr, outputDir, fileName, name string) (string, error) {
	if fileName == "" {
		fileName = name + ".yaml"
	}
	if !strings.HasSuffix(fileName, ".yaml") && !strings.HasSuffix(fileName, ".yml") {
		fileName += ".yaml"
	}
	absDir, err := filepath.Abs(outputDir)
	if err != nil {
		return "", fmt.Errorf("resolve output directory: %w", err)
	}
	if err := os.MkdirAll(absDir, 0o750); err != nil {
		return "", fmt.Errorf("create output directory: %w", err)
	}
	path, err := containedJoin(absDir, fileName)
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(path, []byte(yamlStr), 0o600); err != nil {
		return "", fmt.Errorf("write file: %w", err)
	}
	return path, nil
}

// ClusterWorkflowTemplateFromFile reads a ClusterWorkflowTemplate from a YAML file.
func ClusterWorkflowTemplateFromFile(path string) (model.WorkflowTemplateModel, error) {
	cleanPath := filepath.Clean(path)
	data, err := os.ReadFile(cleanPath) // #nosec G304 — caller controls path
	if err != nil {
		return model.WorkflowTemplateModel{}, fmt.Errorf("read file: %w", err)
	}
	return ClusterWorkflowTemplateFromYAML(string(data))
}

// containedJoin returns filepath.Join(dir, name) only when the result stays
// strictly inside dir. Reject absolute names, ".", "..", and any name whose
// cleaned form escapes dir via traversal. Symlinks inside dir are resolved
// via EvalSymlinks (EC-3 from edge-case review) so a symlink whose target
// sits outside dir cannot bypass the Rel check.
//
// Callers must validate the application-level name shape (e.g., k8s name
// rules) upstream; this helper only enforces containment.
func containedJoin(dir, name string) (string, error) {
	if name == "" || name == "." || name == ".." {
		return "", fmt.Errorf("%w: name=%q", model.ErrPathTraversal, name)
	}
	if filepath.IsAbs(name) {
		return "", fmt.Errorf("%w: name is absolute (%q)", model.ErrPathTraversal, name)
	}
	absDir, err := filepath.Abs(filepath.Clean(dir))
	if err != nil {
		return "", fmt.Errorf("serialize: resolve dir: %w", err)
	}
	// EC-3: resolve symlinks if dir exists; ignore "not exist" so first-time
	// MkdirAll downstream still works.
	if resolved, errSym := filepath.EvalSymlinks(absDir); errSym == nil {
		absDir = resolved
	}
	joined := filepath.Clean(filepath.Join(absDir, name))
	rel, err := filepath.Rel(absDir, joined)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("%w: name=%q escapes %q", model.ErrPathTraversal, name, dir)
	}
	return joined, nil
}

// WorkflowToFile writes a workflow YAML to a file.
// If fileName is empty, a filename is derived from wfName or generateName.
//
// Returns model.ErrPathTraversal (wrapped) when fileName (or auto-derived
// name) would write outside outputDir. Forward subdirectories
// ("sub/file.yaml") are permitted; ".." traversal and absolute paths are not.
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

	if err := os.MkdirAll(absDir, 0o750); err != nil {
		return "", fmt.Errorf("create output directory: %w", err)
	}

	path, err := containedJoin(absDir, fileName)
	if err != nil {
		return "", err
	}
	if err := os.WriteFile(path, []byte(yamlStr), 0o600); err != nil {
		return "", fmt.Errorf("write file: %w", err)
	}

	return path, nil
}

// WorkflowFromFile reads a WorkflowModel from a YAML file.
func WorkflowFromFile(path string) (model.WorkflowModel, error) {
	cleanPath := filepath.Clean(path)
	data, err := os.ReadFile(cleanPath) // #nosec G304 — caller controls path
	if err != nil {
		return model.WorkflowModel{}, fmt.Errorf("read file: %w", err)
	}
	var m model.WorkflowModel
	if err := yaml.Unmarshal(data, &m); err != nil {
		return model.WorkflowModel{}, fmt.Errorf("unmarshal YAML: %w", err)
	}
	return m, nil
}
