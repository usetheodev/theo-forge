package model

import "fmt"

// WorkflowStatus represents the status of a workflow.
type WorkflowStatus string

const (
	WorkflowPending    WorkflowStatus = "Pending"
	WorkflowRunning    WorkflowStatus = "Running"
	WorkflowSucceeded  WorkflowStatus = "Succeeded"
	WorkflowFailed     WorkflowStatus = "Failed"
	WorkflowError      WorkflowStatus = "Error"
	WorkflowTerminated WorkflowStatus = "Terminated"
)

// ParseWorkflowStatus converts a string to a WorkflowStatus.
func ParseWorkflowStatus(s string) (WorkflowStatus, error) {
	switch s {
	case "Pending":
		return WorkflowPending, nil
	case "Running":
		return WorkflowRunning, nil
	case "Succeeded":
		return WorkflowSucceeded, nil
	case "Failed":
		return WorkflowFailed, nil
	case "Error":
		return WorkflowError, nil
	case "Terminated":
		return WorkflowTerminated, nil
	default:
		return "", fmt.Errorf("unknown workflow status %q, valid options: Pending, Running, Succeeded, Failed, Error, Terminated", s)
	}
}

// WorkflowStatusDetail is the detailed status of a workflow, including node-level information.
// Used for parsing the full status section from Argo Workflow CRs.
type WorkflowStatusDetail struct {
	Phase      WorkflowStatus        `json:"phase,omitempty" yaml:"phase,omitempty"`
	StartedAt  string                `json:"startedAt,omitempty" yaml:"startedAt,omitempty"`
	FinishedAt string                `json:"finishedAt,omitempty" yaml:"finishedAt,omitempty"`
	Message    string                `json:"message,omitempty" yaml:"message,omitempty"`
	Nodes      map[string]NodeStatus `json:"nodes,omitempty" yaml:"nodes,omitempty"`
	Outputs    *OutputsModel         `json:"outputs,omitempty" yaml:"outputs,omitempty"`
}

// NodeStatus is the status of a single node in a workflow.
type NodeStatus struct {
	ID           string        `json:"id,omitempty" yaml:"id,omitempty"`
	Name         string        `json:"name,omitempty" yaml:"name,omitempty"`
	DisplayName  string        `json:"displayName,omitempty" yaml:"displayName,omitempty"`
	Type         string        `json:"type,omitempty" yaml:"type,omitempty"`
	Phase        string        `json:"phase,omitempty" yaml:"phase,omitempty"`
	StartedAt    string        `json:"startedAt,omitempty" yaml:"startedAt,omitempty"`
	FinishedAt   string        `json:"finishedAt,omitempty" yaml:"finishedAt,omitempty"`
	Message      string        `json:"message,omitempty" yaml:"message,omitempty"`
	Outputs      *OutputsModel `json:"outputs,omitempty" yaml:"outputs,omitempty"`
	Children     []string      `json:"children,omitempty" yaml:"children,omitempty"`
	TemplateName string        `json:"templateName,omitempty" yaml:"templateName,omitempty"`
}
