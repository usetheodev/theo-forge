package forge

import (
	"encoding/json"
	"testing"

	"github.com/usetheodev/theo-forge/model"
)

func TestParseWorkflowStatusFromUnstructured_Succeeded(t *testing.T) {
	obj := map[string]interface{}{
		"apiVersion": "argoproj.io/v1alpha1",
		"kind":       "Workflow",
		"status": map[string]interface{}{
			"phase":      "Succeeded",
			"startedAt":  "2026-01-01T00:00:00Z",
			"finishedAt": "2026-01-01T00:05:00Z",
			"nodes": map[string]interface{}{
				"node-1": map[string]interface{}{
					"id":          "node-1",
					"name":        "build",
					"type":        "Pod",
					"phase":       "Succeeded",
					"displayName": "build",
					"outputs": map[string]interface{}{
						"exitCode": "0",
					},
				},
				"node-2": map[string]interface{}{
					"id":    "node-2",
					"name":  "main",
					"type":  "DAG",
					"phase": "Succeeded",
				},
			},
		},
	}

	status, err := ParseWorkflowStatusFromUnstructured(obj)
	if err != nil {
		t.Fatalf("ParseWorkflowStatusFromUnstructured: %v", err)
	}
	if status == nil {
		t.Fatal("status is nil")
	}
	if status.Phase != model.WorkflowSucceeded {
		t.Errorf("Phase = %q, want Succeeded", status.Phase)
	}
	if len(status.Nodes) != 2 {
		t.Errorf("len(Nodes) = %d, want 2", len(status.Nodes))
	}
	node1 := status.Nodes["node-1"]
	if node1.Type != "Pod" {
		t.Errorf("node-1.Type = %q, want Pod", node1.Type)
	}
	if node1.Outputs == nil || node1.Outputs.ExitCode != "0" {
		t.Errorf("node-1.Outputs.ExitCode = %q, want 0", node1.Outputs.ExitCode)
	}
}

func TestParseWorkflowStatusFromUnstructured_Failed_MixedExitCodes(t *testing.T) {
	obj := map[string]interface{}{
		"status": map[string]interface{}{
			"phase": "Failed",
			"nodes": map[string]interface{}{
				"pod-ok": map[string]interface{}{
					"type":  "Pod",
					"phase": "Succeeded",
					"outputs": map[string]interface{}{
						"exitCode": "0",
					},
				},
				"pod-fail": map[string]interface{}{
					"type":  "Pod",
					"phase": "Failed",
					"outputs": map[string]interface{}{
						"exitCode": "1",
					},
				},
			},
		},
	}

	status, err := ParseWorkflowStatusFromUnstructured(obj)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if status.Phase != model.WorkflowFailed {
		t.Errorf("Phase = %q, want Failed", status.Phase)
	}
}

func TestParseWorkflowStatusFromUnstructured_NoStatus(t *testing.T) {
	obj := map[string]interface{}{
		"apiVersion": "argoproj.io/v1alpha1",
		"kind":       "Workflow",
	}

	status, err := ParseWorkflowStatusFromUnstructured(obj)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if status != nil {
		t.Errorf("expected nil status, got %+v", status)
	}
}

func TestAllPodNodesExitedZero_AllZero(t *testing.T) {
	status := &model.WorkflowStatusDetail{
		Nodes: map[string]model.NodeStatus{
			"p1": {Type: "Pod", Outputs: &model.OutputsModel{ExitCode: "0"}},
			"p2": {Type: "Pod", Outputs: &model.OutputsModel{ExitCode: "0"}},
			"d1": {Type: "DAG"},
		},
	}
	if !AllPodNodesExitedZero(status) {
		t.Error("expected true for all pods exit 0")
	}
}

func TestAllPodNodesExitedZero_MixedCodes(t *testing.T) {
	status := &model.WorkflowStatusDetail{
		Nodes: map[string]model.NodeStatus{
			"p1": {Type: "Pod", Outputs: &model.OutputsModel{ExitCode: "0"}},
			"p2": {Type: "Pod", Outputs: &model.OutputsModel{ExitCode: "1"}},
		},
	}
	if AllPodNodesExitedZero(status) {
		t.Error("expected false for mixed exit codes")
	}
}

func TestAllPodNodesExitedZero_NoPods(t *testing.T) {
	status := &model.WorkflowStatusDetail{
		Nodes: map[string]model.NodeStatus{
			"d1": {Type: "DAG"},
		},
	}
	if AllPodNodesExitedZero(status) {
		t.Error("expected false for zero Pod nodes")
	}
}

func TestAllPodNodesExitedZero_NilStatus(t *testing.T) {
	if AllPodNodesExitedZero(nil) {
		t.Error("expected false for nil status")
	}
}

func TestWorkflowStatusDetail_Deserialize_WithExitCode(t *testing.T) {
	// Simulate the JSON that comes from the Argo API / K8s dynamic client
	jsonData := `{
		"phase": "Succeeded",
		"nodes": {
			"build-pod": {
				"id": "build-pod",
				"type": "Pod",
				"phase": "Succeeded",
				"outputs": {
					"exitCode": "0",
					"parameters": [{"name": "digest", "value": "sha256:abc"}]
				}
			}
		}
	}`

	var status model.WorkflowStatusDetail
	if err := json.Unmarshal([]byte(jsonData), &status); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	node := status.Nodes["build-pod"]
	if node.Outputs == nil {
		t.Fatal("node outputs nil")
	}
	if node.Outputs.ExitCode != "0" {
		t.Errorf("ExitCode = %q, want 0", node.Outputs.ExitCode)
	}
	if len(node.Outputs.Parameters) != 1 || node.Outputs.Parameters[0].Name != "digest" {
		t.Errorf("Parameters unexpected: %+v", node.Outputs.Parameters)
	}
}

func TestWorkflowSpec_Shutdown_Serializes(t *testing.T) {
	w := &Workflow{
		Name:       "shutdown-test",
		Entrypoint: "main",
		Templates: []Templatable{
			&Container{Name: "main", Image: "alpine:3.18"},
		},
	}

	wfModel, err := w.Build()
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	// Set shutdown on the model directly (as the cancel operation does)
	wfModel.Spec.Shutdown = "Stop"

	data, err := json.Marshal(wfModel)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var roundtrip model.WorkflowModel
	if err := json.Unmarshal(data, &roundtrip); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if roundtrip.Spec.Shutdown != "Stop" {
		t.Errorf("Shutdown = %q, want Stop", roundtrip.Spec.Shutdown)
	}
}
