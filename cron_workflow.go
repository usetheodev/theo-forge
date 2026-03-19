package forge

import (
	"fmt"

	"github.com/usetheo/theo/forge/model"
	"github.com/usetheo/theo/forge/serialize"
)

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
