package model

// DAGModel is the serializable Argo DAG template.
type DAGModel struct {
	Tasks    []DAGTaskModel `json:"tasks" yaml:"tasks"`
	FailFast *bool          `json:"failFast,omitempty" yaml:"failFast,omitempty"`
	Target   string         `json:"target,omitempty" yaml:"target,omitempty"`
}

// DAGTaskModel is the serializable Argo DAG task.
type DAGTaskModel struct {
	Name         string          `json:"name" yaml:"name"`
	Template     string          `json:"template,omitempty" yaml:"template,omitempty"`
	TemplateRef  *TemplateRef    `json:"templateRef,omitempty" yaml:"templateRef,omitempty"`
	Inline       *TemplateModel  `json:"inline,omitempty" yaml:"inline,omitempty"`
	Dependencies []string        `json:"dependencies,omitempty" yaml:"dependencies,omitempty"`
	Depends      string          `json:"depends,omitempty" yaml:"depends,omitempty"`
	Arguments    *ArgumentsModel `json:"arguments,omitempty" yaml:"arguments,omitempty"`
	When         string          `json:"when,omitempty" yaml:"when,omitempty"`
	ContinueOn   *ContinueOn    `json:"continueOn,omitempty" yaml:"continueOn,omitempty"`
	WithItems    []interface{}   `json:"withItems,omitempty" yaml:"withItems,omitempty"`
	WithParam    string          `json:"withParam,omitempty" yaml:"withParam,omitempty"`
	WithSequence *Sequence       `json:"withSequence,omitempty" yaml:"withSequence,omitempty"`
	OnExit       string          `json:"onExit,omitempty" yaml:"onExit,omitempty"`
	Hooks        map[string]LifecycleHook `json:"hooks,omitempty" yaml:"hooks,omitempty"`
}

// Sequence generates a list of numbers for fan-out.
type Sequence struct {
	Count  string `json:"count,omitempty" yaml:"count,omitempty"`
	Start  string `json:"start,omitempty" yaml:"start,omitempty"`
	End    string `json:"end,omitempty" yaml:"end,omitempty"`
	Format string `json:"format,omitempty" yaml:"format,omitempty"`
}

// TemplateRef references a template in a WorkflowTemplate.
type TemplateRef struct {
	Name         string `json:"name" yaml:"name"`
	Template     string `json:"template" yaml:"template"`
	ClusterScope bool   `json:"clusterScope,omitempty" yaml:"clusterScope,omitempty"`
}

// ArgumentsModel is the serializable Argo Arguments.
type ArgumentsModel struct {
	Parameters []ParameterModel `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	Artifacts  []ArtifactModel  `json:"artifacts,omitempty" yaml:"artifacts,omitempty"`
}

// ContinueOn defines when to continue after a step/task fails.
type ContinueOn struct {
	Error  bool `json:"error,omitempty" yaml:"error,omitempty"`
	Failed bool `json:"failed,omitempty" yaml:"failed,omitempty"`
}
