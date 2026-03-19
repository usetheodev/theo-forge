package model

// StepModel is the serializable Argo Workflow step.
type StepModel struct {
	Name         string          `json:"name" yaml:"name"`
	Template     string          `json:"template,omitempty" yaml:"template,omitempty"`
	TemplateRef  *TemplateRef    `json:"templateRef,omitempty" yaml:"templateRef,omitempty"`
	Inline       *TemplateModel  `json:"inline,omitempty" yaml:"inline,omitempty"`
	Arguments    *ArgumentsModel `json:"arguments,omitempty" yaml:"arguments,omitempty"`
	When         string          `json:"when,omitempty" yaml:"when,omitempty"`
	ContinueOn   *ContinueOn    `json:"continueOn,omitempty" yaml:"continueOn,omitempty"`
	WithItems    []interface{}   `json:"withItems,omitempty" yaml:"withItems,omitempty"`
	WithParam    string          `json:"withParam,omitempty" yaml:"withParam,omitempty"`
	WithSequence *Sequence       `json:"withSequence,omitempty" yaml:"withSequence,omitempty"`
	OnExit       string          `json:"onExit,omitempty" yaml:"onExit,omitempty"`
	Hooks        map[string]LifecycleHook `json:"hooks,omitempty" yaml:"hooks,omitempty"`
}
