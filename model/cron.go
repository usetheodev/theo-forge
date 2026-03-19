package model

// CronWorkflowModel is the serializable Argo CronWorkflow.
type CronWorkflowModel struct {
	APIVersion string           `json:"apiVersion" yaml:"apiVersion"`
	Kind       string           `json:"kind" yaml:"kind"`
	Metadata   WorkflowMetadata `json:"metadata" yaml:"metadata"`
	Spec       CronWorkflowSpec `json:"spec" yaml:"spec"`
}

// CronWorkflowSpec is the spec for a cron workflow.
type CronWorkflowSpec struct {
	Schedule                   string       `json:"schedule,omitempty" yaml:"schedule,omitempty"`
	Schedules                  []string     `json:"schedules,omitempty" yaml:"schedules,omitempty"`
	Timezone                   string       `json:"timezone,omitempty" yaml:"timezone,omitempty"`
	When                       string       `json:"when,omitempty" yaml:"when,omitempty"`
	Suspend                    *bool        `json:"suspend,omitempty" yaml:"suspend,omitempty"`
	ConcurrencyPolicy          string       `json:"concurrencyPolicy,omitempty" yaml:"concurrencyPolicy,omitempty"`
	StartingDeadlineSeconds    *int         `json:"startingDeadlineSeconds,omitempty" yaml:"startingDeadlineSeconds,omitempty"`
	SuccessfulJobsHistoryLimit *int         `json:"successfulJobsHistoryLimit,omitempty" yaml:"successfulJobsHistoryLimit,omitempty"`
	FailedJobsHistoryLimit     *int         `json:"failedJobsHistoryLimit,omitempty" yaml:"failedJobsHistoryLimit,omitempty"`
	WorkflowSpec               WorkflowSpec `json:"workflowSpec" yaml:"workflowSpec"`
	WorkflowMetadata           *WorkflowMetadata `json:"workflowMetadata,omitempty" yaml:"workflowMetadata,omitempty"`
	StopStrategy               *StopStrategy `json:"stopStrategy,omitempty" yaml:"stopStrategy,omitempty"`
}

// StopStrategy defines when to stop scheduling.
type StopStrategy struct {
	Expression string `json:"expression,omitempty" yaml:"expression,omitempty"`
}
