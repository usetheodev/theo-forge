package forge

import (
	"fmt"

	"github.com/usetheo/theo/forge/model"
	"github.com/usetheo/theo/forge/serialize"
)

// WorkflowTemplate represents a namespace-scoped reusable workflow template.
type WorkflowTemplate struct {
	// Name is the template name.
	Name string
	// Namespace is the K8s namespace.
	Namespace string
	// APIVersion is the API version (default: argoproj.io/v1alpha1).
	APIVersion string
	// Labels for the template.
	Labels map[string]string
	// Annotations for the template.
	Annotations map[string]string
	// Entrypoint is the default starting template.
	Entrypoint string
	// Templates are the workflow templates.
	Templates []Templatable
	// Arguments are the template-level arguments.
	Arguments []Parameter
	// Volumes are the template-level volumes.
	Volumes []VolumeBuilder
	// ServiceAccountName for the template.
	ServiceAccountName string
}

func (wt *WorkflowTemplate) validate() error {
	if wt.Name == "" {
		return fmt.Errorf("workflow template name cannot be empty")
	}
	if len(wt.Name) > NameLimit {
		return fmt.Errorf("name must be no more than %d characters", NameLimit)
	}
	return nil
}

// Build converts the WorkflowTemplate to its serializable model.
func (wt *WorkflowTemplate) Build() (model.WorkflowTemplateModel, error) {
	if err := wt.validate(); err != nil {
		return model.WorkflowTemplateModel{}, err
	}

	apiVersion := wt.APIVersion
	if apiVersion == "" {
		apiVersion = DefaultAPIVersion
	}

	templates, err := buildTemplateModels(wt.Templates)
	if err != nil {
		return model.WorkflowTemplateModel{}, err
	}

	var args *model.ArgumentsModel
	if len(wt.Arguments) > 0 {
		args = &model.ArgumentsModel{}
		for _, p := range wt.Arguments {
			m, err := p.AsArgument()
			if err != nil {
				return model.WorkflowTemplateModel{}, fmt.Errorf("argument %q: %w", p.Name, err)
			}
			args.Parameters = append(args.Parameters, m)
		}
	}

	var vols []model.VolumeModel
	for _, v := range wt.Volumes {
		m, err := v.BuildVolume()
		if err != nil {
			return model.WorkflowTemplateModel{}, fmt.Errorf("volume: %w", err)
		}
		vols = append(vols, m)
	}

	return model.WorkflowTemplateModel{
		APIVersion: apiVersion,
		Kind:       "WorkflowTemplate",
		Metadata: model.WorkflowMetadata{
			Name:        wt.Name,
			Namespace:   wt.Namespace,
			Labels:      wt.Labels,
			Annotations: wt.Annotations,
		},
		Spec: model.WorkflowSpec{
			Entrypoint:         wt.Entrypoint,
			Templates:          templates,
			Arguments:          args,
			Volumes:            vols,
			ServiceAccountName: wt.ServiceAccountName,
		},
	}, nil
}

// ToYAML converts the WorkflowTemplate to YAML.
func (wt *WorkflowTemplate) ToYAML() (string, error) {
	m, err := wt.Build()
	if err != nil {
		return "", err
	}
	return serialize.WorkflowTemplateToYAML(m)
}

// ClusterWorkflowTemplate represents a cluster-scoped reusable workflow template.
type ClusterWorkflowTemplate struct {
	// Name is the template name.
	Name string
	// APIVersion is the API version.
	APIVersion string
	// Labels for the template.
	Labels map[string]string
	// Annotations for the template.
	Annotations map[string]string
	// Entrypoint is the default starting template.
	Entrypoint string
	// Templates are the workflow templates.
	Templates []Templatable
	// Arguments are the template-level arguments.
	Arguments []Parameter
	// ServiceAccountName for the template.
	ServiceAccountName string
}

func (cwt *ClusterWorkflowTemplate) validate() error {
	if cwt.Name == "" {
		return fmt.Errorf("cluster workflow template name cannot be empty")
	}
	return nil
}

// Build converts the ClusterWorkflowTemplate to its serializable model.
func (cwt *ClusterWorkflowTemplate) Build() (model.WorkflowTemplateModel, error) {
	if err := cwt.validate(); err != nil {
		return model.WorkflowTemplateModel{}, err
	}

	apiVersion := cwt.APIVersion
	if apiVersion == "" {
		apiVersion = DefaultAPIVersion
	}

	templates, err := buildTemplateModels(cwt.Templates)
	if err != nil {
		return model.WorkflowTemplateModel{}, err
	}

	var args *model.ArgumentsModel
	if len(cwt.Arguments) > 0 {
		args = &model.ArgumentsModel{}
		for _, p := range cwt.Arguments {
			m, err := p.AsArgument()
			if err != nil {
				return model.WorkflowTemplateModel{}, fmt.Errorf("argument %q: %w", p.Name, err)
			}
			args.Parameters = append(args.Parameters, m)
		}
	}

	return model.WorkflowTemplateModel{
		APIVersion: apiVersion,
		Kind:       "ClusterWorkflowTemplate",
		Metadata: model.WorkflowMetadata{
			Name:        cwt.Name,
			Labels:      cwt.Labels,
			Annotations: cwt.Annotations,
		},
		Spec: model.WorkflowSpec{
			Entrypoint:         cwt.Entrypoint,
			Templates:          templates,
			Arguments:          args,
			ServiceAccountName: cwt.ServiceAccountName,
		},
	}, nil
}

// ToYAML converts the ClusterWorkflowTemplate to YAML.
func (cwt *ClusterWorkflowTemplate) ToYAML() (string, error) {
	m, err := cwt.Build()
	if err != nil {
		return "", err
	}
	return serialize.WorkflowTemplateToYAML(m)
}

