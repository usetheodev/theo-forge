package forge

import (
	"github.com/usetheo/theo/forge/model"
	"github.com/usetheo/theo/forge/serialize"
)

// ToFile writes the workflow YAML to a file.
// If name is empty, the workflow name is used as the filename.
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
