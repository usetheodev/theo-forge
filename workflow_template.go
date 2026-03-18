package forge

import (
	"encoding/json"
	"fmt"

	"sigs.k8s.io/yaml"
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

// WorkflowTemplateModel is the serializable Argo WorkflowTemplate.
type WorkflowTemplateModel struct {
	APIVersion string            `json:"apiVersion" yaml:"apiVersion"`
	Kind       string            `json:"kind" yaml:"kind"`
	Metadata   WorkflowMetadata  `json:"metadata" yaml:"metadata"`
	Spec       WorkflowSpec      `json:"spec" yaml:"spec"`
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
func (wt *WorkflowTemplate) Build() (WorkflowTemplateModel, error) {
	if err := wt.validate(); err != nil {
		return WorkflowTemplateModel{}, err
	}

	apiVersion := wt.APIVersion
	if apiVersion == "" {
		apiVersion = DefaultAPIVersion
	}

	templates := make([]TemplateModel, 0, len(wt.Templates))
	for _, t := range wt.Templates {
		m, err := t.BuildTemplate()
		if err != nil {
			return WorkflowTemplateModel{}, fmt.Errorf("template %q: %w", t.GetName(), err)
		}
		templates = append(templates, m)
	}

	var args *ArgumentsModel
	if len(wt.Arguments) > 0 {
		args = &ArgumentsModel{}
		for _, p := range wt.Arguments {
			m, err := p.AsArgument()
			if err != nil {
				continue
			}
			args.Parameters = append(args.Parameters, m)
		}
	}

	var vols []VolumeModel
	for _, v := range wt.Volumes {
		m, err := v.BuildVolume()
		if err != nil {
			continue
		}
		vols = append(vols, m)
	}

	return WorkflowTemplateModel{
		APIVersion: apiVersion,
		Kind:       "WorkflowTemplate",
		Metadata: WorkflowMetadata{
			Name:        wt.Name,
			Namespace:   wt.Namespace,
			Labels:      wt.Labels,
			Annotations: wt.Annotations,
		},
		Spec: WorkflowSpec{
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
	model, err := wt.Build()
	if err != nil {
		return "", err
	}
	data, err := yaml.Marshal(model)
	if err != nil {
		return "", err
	}
	return string(data), nil
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
func (cwt *ClusterWorkflowTemplate) Build() (WorkflowTemplateModel, error) {
	if err := cwt.validate(); err != nil {
		return WorkflowTemplateModel{}, err
	}

	apiVersion := cwt.APIVersion
	if apiVersion == "" {
		apiVersion = DefaultAPIVersion
	}

	templates := make([]TemplateModel, 0, len(cwt.Templates))
	for _, t := range cwt.Templates {
		m, err := t.BuildTemplate()
		if err != nil {
			return WorkflowTemplateModel{}, fmt.Errorf("template %q: %w", t.GetName(), err)
		}
		templates = append(templates, m)
	}

	var args *ArgumentsModel
	if len(cwt.Arguments) > 0 {
		args = &ArgumentsModel{}
		for _, p := range cwt.Arguments {
			m, err := p.AsArgument()
			if err != nil {
				continue
			}
			args.Parameters = append(args.Parameters, m)
		}
	}

	return WorkflowTemplateModel{
		APIVersion: apiVersion,
		Kind:       "ClusterWorkflowTemplate",
		Metadata: WorkflowMetadata{
			Name:        cwt.Name,
			Labels:      cwt.Labels,
			Annotations: cwt.Annotations,
		},
		Spec: WorkflowSpec{
			Entrypoint:         cwt.Entrypoint,
			Templates:          templates,
			Arguments:          args,
			ServiceAccountName: cwt.ServiceAccountName,
		},
	}, nil
}

// ToYAML converts the ClusterWorkflowTemplate to YAML.
func (cwt *ClusterWorkflowTemplate) ToYAML() (string, error) {
	model, err := cwt.Build()
	if err != nil {
		return "", err
	}
	data, err := yaml.Marshal(model)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// CronWorkflow represents a scheduled workflow.
type CronWorkflow struct {
	// Name is the cron workflow name.
	Name string
	// Namespace is the K8s namespace.
	Namespace string
	// APIVersion is the API version.
	APIVersion string
	// Labels for the cron workflow.
	Labels map[string]string
	// Annotations for the cron workflow.
	Annotations map[string]string
	// Schedule is the cron expression (e.g., "0 * * * *").
	Schedule string
	// Timezone for the schedule.
	Timezone string
	// Suspend pauses the cron schedule.
	Suspend *bool
	// ConcurrencyPolicy defines how concurrent runs are handled (Allow, Replace, Forbid).
	ConcurrencyPolicy string
	// StartingDeadlineSeconds is the deadline for starting missed runs.
	StartingDeadlineSeconds *int
	// SuccessfulJobsHistoryLimit is the number of successful runs to keep.
	SuccessfulJobsHistoryLimit *int
	// FailedJobsHistoryLimit is the number of failed runs to keep.
	FailedJobsHistoryLimit *int
	// WorkflowSpec is the workflow to run on schedule.
	Entrypoint         string
	Templates          []Templatable
	Arguments          []Parameter
	Volumes            []VolumeBuilder
	ServiceAccountName string
}

// CronWorkflowModel is the serializable Argo CronWorkflow.
type CronWorkflowModel struct {
	APIVersion string             `json:"apiVersion" yaml:"apiVersion"`
	Kind       string             `json:"kind" yaml:"kind"`
	Metadata   WorkflowMetadata   `json:"metadata" yaml:"metadata"`
	Spec       CronWorkflowSpec   `json:"spec" yaml:"spec"`
}

// CronWorkflowSpec is the spec for a cron workflow.
type CronWorkflowSpec struct {
	Schedule                   string       `json:"schedule" yaml:"schedule"`
	Timezone                   string       `json:"timezone,omitempty" yaml:"timezone,omitempty"`
	Suspend                    *bool        `json:"suspend,omitempty" yaml:"suspend,omitempty"`
	ConcurrencyPolicy          string       `json:"concurrencyPolicy,omitempty" yaml:"concurrencyPolicy,omitempty"`
	StartingDeadlineSeconds    *int         `json:"startingDeadlineSeconds,omitempty" yaml:"startingDeadlineSeconds,omitempty"`
	SuccessfulJobsHistoryLimit *int         `json:"successfulJobsHistoryLimit,omitempty" yaml:"successfulJobsHistoryLimit,omitempty"`
	FailedJobsHistoryLimit     *int         `json:"failedJobsHistoryLimit,omitempty" yaml:"failedJobsHistoryLimit,omitempty"`
	WorkflowSpec               WorkflowSpec `json:"workflowSpec" yaml:"workflowSpec"`
}

func (cw *CronWorkflow) validate() error {
	if cw.Name == "" {
		return fmt.Errorf("cron workflow name cannot be empty")
	}
	if cw.Schedule == "" {
		return fmt.Errorf("cron workflow schedule cannot be empty")
	}
	return nil
}

// Build converts the CronWorkflow to its serializable model.
func (cw *CronWorkflow) Build() (CronWorkflowModel, error) {
	if err := cw.validate(); err != nil {
		return CronWorkflowModel{}, err
	}

	apiVersion := cw.APIVersion
	if apiVersion == "" {
		apiVersion = DefaultAPIVersion
	}

	templates := make([]TemplateModel, 0, len(cw.Templates))
	for _, t := range cw.Templates {
		m, err := t.BuildTemplate()
		if err != nil {
			return CronWorkflowModel{}, fmt.Errorf("template %q: %w", t.GetName(), err)
		}
		templates = append(templates, m)
	}

	var args *ArgumentsModel
	if len(cw.Arguments) > 0 {
		args = &ArgumentsModel{}
		for _, p := range cw.Arguments {
			m, err := p.AsArgument()
			if err != nil {
				continue
			}
			args.Parameters = append(args.Parameters, m)
		}
	}

	var vols []VolumeModel
	for _, v := range cw.Volumes {
		m, err := v.BuildVolume()
		if err != nil {
			continue
		}
		vols = append(vols, m)
	}

	return CronWorkflowModel{
		APIVersion: apiVersion,
		Kind:       "CronWorkflow",
		Metadata: WorkflowMetadata{
			Name:        cw.Name,
			Namespace:   cw.Namespace,
			Labels:      cw.Labels,
			Annotations: cw.Annotations,
		},
		Spec: CronWorkflowSpec{
			Schedule:                   cw.Schedule,
			Timezone:                   cw.Timezone,
			Suspend:                    cw.Suspend,
			ConcurrencyPolicy:          cw.ConcurrencyPolicy,
			StartingDeadlineSeconds:    cw.StartingDeadlineSeconds,
			SuccessfulJobsHistoryLimit: cw.SuccessfulJobsHistoryLimit,
			FailedJobsHistoryLimit:     cw.FailedJobsHistoryLimit,
			WorkflowSpec: WorkflowSpec{
				Entrypoint:         cw.Entrypoint,
				Templates:          templates,
				Arguments:          args,
				Volumes:            vols,
				ServiceAccountName: cw.ServiceAccountName,
			},
		},
	}, nil
}

// ToYAML converts the CronWorkflow to YAML.
func (cw *CronWorkflow) ToYAML() (string, error) {
	model, err := cw.Build()
	if err != nil {
		return "", err
	}
	data, err := yaml.Marshal(model)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ToJSON converts the CronWorkflow to JSON.
func (cw *CronWorkflow) ToJSON() (string, error) {
	model, err := cw.Build()
	if err != nil {
		return "", err
	}
	data, err := json.MarshalIndent(model, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}
