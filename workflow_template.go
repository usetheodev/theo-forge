package forge

import (
	"fmt"

	"github.com/usetheodev/theo-forge/model"
	"github.com/usetheodev/theo-forge/serialize"
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

// --- CronWorkflow ---

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
	// Schedules supports multiple cron schedules (alternative to Schedule).
	Schedules []string
	// When is a condition expression for running the workflow.
	When string
	// WorkflowMetadata sets metadata on the generated workflow.
	WorkflowMetadata *model.WorkflowMetadata
	// StopStrategy defines when to stop scheduling.
	StopStrategy *model.StopStrategy
	// WorkflowSpec is the workflow to run on schedule.
	Entrypoint         string
	Templates          []Templatable
	Arguments          []Parameter
	Volumes            []VolumeBuilder
	ServiceAccountName string
}

func (cw *CronWorkflow) validate() error {
	if cw.Name == "" {
		return fmt.Errorf("cron workflow name cannot be empty")
	}
	if cw.Schedule == "" && len(cw.Schedules) == 0 {
		return fmt.Errorf("cron workflow schedule or schedules must be set")
	}
	return nil
}

// Build converts the CronWorkflow to its serializable model.
func (cw *CronWorkflow) Build() (model.CronWorkflowModel, error) {
	if err := cw.validate(); err != nil {
		return model.CronWorkflowModel{}, err
	}

	apiVersion := cw.APIVersion
	if apiVersion == "" {
		apiVersion = DefaultAPIVersion
	}

	templates, err := buildTemplateModels(cw.Templates)
	if err != nil {
		return model.CronWorkflowModel{}, err
	}

	var args *model.ArgumentsModel
	if len(cw.Arguments) > 0 {
		args = &model.ArgumentsModel{}
		for _, p := range cw.Arguments {
			m, err := p.AsArgument()
			if err != nil {
				return model.CronWorkflowModel{}, fmt.Errorf("argument %q: %w", p.Name, err)
			}
			args.Parameters = append(args.Parameters, m)
		}
	}

	var vols []model.VolumeModel
	for _, v := range cw.Volumes {
		m, err := v.BuildVolume()
		if err != nil {
			return model.CronWorkflowModel{}, fmt.Errorf("volume: %w", err)
		}
		vols = append(vols, m)
	}

	return model.CronWorkflowModel{
		APIVersion: apiVersion,
		Kind:       "CronWorkflow",
		Metadata: model.WorkflowMetadata{
			Name:        cw.Name,
			Namespace:   cw.Namespace,
			Labels:      cw.Labels,
			Annotations: cw.Annotations,
		},
		Spec: model.CronWorkflowSpec{
			Schedule:                   cw.Schedule,
			Schedules:                  cw.Schedules,
			Timezone:                   cw.Timezone,
			When:                       cw.When,
			Suspend:                    cw.Suspend,
			ConcurrencyPolicy:          cw.ConcurrencyPolicy,
			StartingDeadlineSeconds:    cw.StartingDeadlineSeconds,
			SuccessfulJobsHistoryLimit: cw.SuccessfulJobsHistoryLimit,
			FailedJobsHistoryLimit:     cw.FailedJobsHistoryLimit,
			WorkflowMetadata:           cw.WorkflowMetadata,
			StopStrategy:               cw.StopStrategy,
			WorkflowSpec: model.WorkflowSpec{
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
	m, err := cw.Build()
	if err != nil {
		return "", err
	}
	return serialize.CronWorkflowToYAML(m)
}

// ToJSON converts the CronWorkflow to JSON.
func (cw *CronWorkflow) ToJSON() (string, error) {
	m, err := cw.Build()
	if err != nil {
		return "", err
	}
	return serialize.CronWorkflowToJSON(m)
}

