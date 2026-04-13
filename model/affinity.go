package model

// Affinity defines scheduling constraints for workflow pods.
// Mirrors k8s.io/api/core/v1.Affinity without importing the dependency.
type Affinity struct {
	NodeAffinity    *NodeAffinity    `json:"nodeAffinity,omitempty" yaml:"nodeAffinity,omitempty"`
	PodAffinity     *PodAffinity     `json:"podAffinity,omitempty" yaml:"podAffinity,omitempty"`
	PodAntiAffinity *PodAntiAffinity `json:"podAntiAffinity,omitempty" yaml:"podAntiAffinity,omitempty"`
}

// PodAffinity defines pod affinity scheduling rules.
type PodAffinity struct {
	RequiredDuringSchedulingIgnoredDuringExecution  []PodAffinityTerm         `json:"requiredDuringSchedulingIgnoredDuringExecution,omitempty" yaml:"requiredDuringSchedulingIgnoredDuringExecution,omitempty"`
	PreferredDuringSchedulingIgnoredDuringExecution []WeightedPodAffinityTerm `json:"preferredDuringSchedulingIgnoredDuringExecution,omitempty" yaml:"preferredDuringSchedulingIgnoredDuringExecution,omitempty"`
}

// PodAntiAffinity defines pod anti-affinity scheduling rules.
type PodAntiAffinity struct {
	RequiredDuringSchedulingIgnoredDuringExecution  []PodAffinityTerm         `json:"requiredDuringSchedulingIgnoredDuringExecution,omitempty" yaml:"requiredDuringSchedulingIgnoredDuringExecution,omitempty"`
	PreferredDuringSchedulingIgnoredDuringExecution []WeightedPodAffinityTerm `json:"preferredDuringSchedulingIgnoredDuringExecution,omitempty" yaml:"preferredDuringSchedulingIgnoredDuringExecution,omitempty"`
}

// PodAffinityTerm defines a set of pods for affinity scheduling.
type PodAffinityTerm struct {
	LabelSelector *LabelSelector `json:"labelSelector,omitempty" yaml:"labelSelector,omitempty"`
	TopologyKey   string         `json:"topologyKey" yaml:"topologyKey"`
	Namespaces    []string       `json:"namespaces,omitempty" yaml:"namespaces,omitempty"`
}

// WeightedPodAffinityTerm defines a weighted pod affinity term.
type WeightedPodAffinityTerm struct {
	Weight          int32           `json:"weight" yaml:"weight"`
	PodAffinityTerm PodAffinityTerm `json:"podAffinityTerm" yaml:"podAffinityTerm"`
}

// NodeAffinity defines node affinity scheduling rules.
type NodeAffinity struct {
	RequiredDuringSchedulingIgnoredDuringExecution  *NodeSelector             `json:"requiredDuringSchedulingIgnoredDuringExecution,omitempty" yaml:"requiredDuringSchedulingIgnoredDuringExecution,omitempty"`
	PreferredDuringSchedulingIgnoredDuringExecution []PreferredSchedulingTerm `json:"preferredDuringSchedulingIgnoredDuringExecution,omitempty" yaml:"preferredDuringSchedulingIgnoredDuringExecution,omitempty"`
}

// NodeSelector represents a node selector requirement.
type NodeSelector struct {
	NodeSelectorTerms []NodeSelectorTerm `json:"nodeSelectorTerms" yaml:"nodeSelectorTerms"`
}

// NodeSelectorTerm defines node selection criteria.
type NodeSelectorTerm struct {
	MatchExpressions []NodeSelectorRequirement `json:"matchExpressions,omitempty" yaml:"matchExpressions,omitempty"`
	MatchFields      []NodeSelectorRequirement `json:"matchFields,omitempty" yaml:"matchFields,omitempty"`
}

// NodeSelectorRequirement is a single node selector requirement.
type NodeSelectorRequirement struct {
	Key      string   `json:"key" yaml:"key"`
	Operator string   `json:"operator" yaml:"operator"`
	Values   []string `json:"values,omitempty" yaml:"values,omitempty"`
}

// PreferredSchedulingTerm defines a weighted preferred scheduling term.
type PreferredSchedulingTerm struct {
	Weight     int32            `json:"weight" yaml:"weight"`
	Preference NodeSelectorTerm `json:"preference" yaml:"preference"`
}
