package forge

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
