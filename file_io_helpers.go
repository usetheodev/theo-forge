package forge

import (
	"github.com/usetheodev/theo-forge/model"
	"github.com/usetheodev/theo-forge/serialize"
)

// T4.2 — file I/O helpers split from helpers.go.

// ToFile writes the workflow YAML to a file.
// If name is empty, the workflow name is used as the filename.
// Returns model.ErrPathTraversal (wrapped) on path-traversal attempts.
func (w *Workflow) ToFile(outputDir string, name string) (string, error) {
	yamlStr, err := w.ToYAML()
	if err != nil {
		return "", err
	}
	return serialize.WorkflowToFile(yamlStr, outputDir, name, w.Name, w.GenerateName)
}

// FromFile reads a WorkflowModel from a YAML file.
func FromFile(path string) (model.WorkflowModel, error) {
	return serialize.WorkflowFromFile(path)
}
