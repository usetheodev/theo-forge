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
	Dependencies []string        `json:"dependencies,omitempty" yaml:"dependencies,omitempty"`
	Depends      string          `json:"depends,omitempty" yaml:"depends,omitempty"`
	Arguments    *ArgumentsModel `json:"arguments,omitempty" yaml:"arguments,omitempty"`
	When         string          `json:"when,omitempty" yaml:"when,omitempty"`
	ContinueOn   *ContinueOn    `json:"continueOn,omitempty" yaml:"continueOn,omitempty"`
	WithItems    []interface{}   `json:"withItems,omitempty" yaml:"withItems,omitempty"`
	WithParam    string          `json:"withParam,omitempty" yaml:"withParam,omitempty"`
}

// TemplateRef references a template in a WorkflowTemplate.
type TemplateRef struct {
	Name     string `json:"name" yaml:"name"`
	Template string `json:"template" yaml:"template"`
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
