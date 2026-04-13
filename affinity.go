package forge

import "github.com/usetheodev/theo-forge/model"

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
