package forge

import "github.com/usetheodev/theo-forge/model"

// workflowLabelKey is the label Argo Controller injects on every pod
// of a workflow. Matching this label co-locates every step on the
// same node.
const workflowLabelKey = "workflows.argoproj.io/workflow"

// hostnameTopologyKey is the standard node-identity topology key.
const hostnameTopologyKey = "kubernetes.io/hostname"

// ColocateByLabel creates a PodAffinity that co-locates all pods
// with the same label on the same topology (typically hostname).
// Common pattern for workflows sharing a ReadWriteOnce PVC.
func ColocateByLabel(labelKey, labelValue, topologyKey string) *model.Affinity {
	return &model.Affinity{
		PodAffinity: &model.PodAffinity{
			RequiredDuringSchedulingIgnoredDuringExecution: []model.PodAffinityTerm{{
				LabelSelector: &model.LabelSelector{
					MatchLabels: map[string]string{labelKey: labelValue},
				},
				TopologyKey: topologyKey,
			}},
		},
	}
}

// DefaultPodAffinityFor builds the default podAffinity term used by
// Workflow.Build() when the user did not supply Affinity AND the workflow
// declares VolumeClaimTemplates (RWO PVCs shared across DAG steps).
//
// The term matches on Argo's workflow-identity label
// (workflows.argoproj.io/workflow) and topology key
// kubernetes.io/hostname, so every pod of the same workflow lands on the
// same node. This eliminates the PVC RWO Multi-Attach race that surfaces
// when DAG steps are scheduled across nodes.
//
// Behavior:
//   - Returns nil if w is nil.
//   - Returns nil if w.Name and w.GenerateName are both empty.
//   - When w.Name is set, the label value is the literal Name.
//   - When only w.GenerateName is set, the label value is the Argo
//     template variable "{{workflow.name}}", which Argo Controller
//     materializes at runtime after generating the final workflow name.
//
// Consumers normally don't call this directly — Workflow.Build() injects
// it automatically when the gate conditions hold. It's exported so
// consumers can opt in explicitly (e.g., wrap an external workflow) or
// inspect the canonical term in tests.
func DefaultPodAffinityFor(w *Workflow) *model.Affinity {
	if w == nil {
		return nil
	}
	labelValue := w.Name
	if labelValue == "" {
		if w.GenerateName == "" {
			return nil
		}
		labelValue = "{{workflow.name}}"
	}
	return ColocateByLabel(workflowLabelKey, labelValue, hostnameTopologyKey)
}
