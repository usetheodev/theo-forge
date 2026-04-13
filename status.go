package forge

import (
	"encoding/json"
	"fmt"

	"github.com/usetheodev/theo-forge/model"
)

// ParseWorkflowStatusFromUnstructured extracts WorkflowStatusDetail from a
// K8s unstructured object (map[string]interface{}).
// Returns nil if there is no status section.
func ParseWorkflowStatusFromUnstructured(obj map[string]interface{}) (*model.WorkflowStatusDetail, error) {
	statusRaw, ok := obj["status"]
	if !ok {
		return nil, nil
	}
	data, err := json.Marshal(statusRaw)
	if err != nil {
		return nil, fmt.Errorf("marshal status: %w", err)
	}
	var status model.WorkflowStatusDetail
	if err := json.Unmarshal(data, &status); err != nil {
		return nil, fmt.Errorf("unmarshal status: %w", err)
	}
	return &status, nil
}

// AllPodNodesExitedZero checks if all Pod-type nodes have exitCode "0".
// Detects false failures from daemon sidecar termination.
func AllPodNodesExitedZero(status *model.WorkflowStatusDetail) bool {
	if status == nil || len(status.Nodes) == 0 {
		return false
	}
	podCount := 0
	for _, node := range status.Nodes {
		if node.Type != "Pod" {
			continue
		}
		podCount++
		if node.Outputs == nil || node.Outputs.ExitCode != "0" {
			return false
		}
	}
	return podCount > 0
}
