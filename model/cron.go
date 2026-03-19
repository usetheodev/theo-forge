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
	Schedule                   string       `json:"schedule" yaml:"schedule"`
	Timezone                   string       `json:"timezone,omitempty" yaml:"timezone,omitempty"`
	Suspend                    *bool        `json:"suspend,omitempty" yaml:"suspend,omitempty"`
	ConcurrencyPolicy          string       `json:"concurrencyPolicy,omitempty" yaml:"concurrencyPolicy,omitempty"`
	StartingDeadlineSeconds    *int         `json:"startingDeadlineSeconds,omitempty" yaml:"startingDeadlineSeconds,omitempty"`
	SuccessfulJobsHistoryLimit *int         `json:"successfulJobsHistoryLimit,omitempty" yaml:"successfulJobsHistoryLimit,omitempty"`
	FailedJobsHistoryLimit     *int         `json:"failedJobsHistoryLimit,omitempty" yaml:"failedJobsHistoryLimit,omitempty"`
	WorkflowSpec               WorkflowSpec `json:"workflowSpec" yaml:"workflowSpec"`
}
