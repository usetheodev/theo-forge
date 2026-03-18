package forge

import "fmt"

// Operator defines how task dependencies are combined.
type Operator string

const (
	OperatorAnd Operator = "&&"
	OperatorOr  Operator = "||"
)

// TaskResult represents the result of a task for conditional dependencies.
type TaskResult string

const (
	TaskFailed       TaskResult = "Failed"
	TaskSucceeded    TaskResult = "Succeeded"
	TaskErrored      TaskResult = "Errored"
	TaskSkipped      TaskResult = "Skipped"
	TaskOmitted      TaskResult = "Omitted"
	TaskDaemoned     TaskResult = "Daemoned"
	TaskAnySucceeded TaskResult = "AnySucceeded"
	TaskAllFailed    TaskResult = "AllFailed"
)

// DAGModel is the serializable Argo DAG template.
type DAGModel struct {
	Tasks    []DAGTaskModel `json:"tasks" yaml:"tasks"`
	FailFast *bool          `json:"failFast,omitempty" yaml:"failFast,omitempty"`
	Target   string         `json:"target,omitempty" yaml:"target,omitempty"`
}

// DAGTaskModel is the serializable Argo DAG task.
type DAGTaskModel struct {
	Name         string           `json:"name" yaml:"name"`
	Template     string           `json:"template,omitempty" yaml:"template,omitempty"`
	TemplateRef  *TemplateRef     `json:"templateRef,omitempty" yaml:"templateRef,omitempty"`
	Dependencies []string         `json:"dependencies,omitempty" yaml:"dependencies,omitempty"`
	Depends      string           `json:"depends,omitempty" yaml:"depends,omitempty"`
	Arguments    *ArgumentsModel  `json:"arguments,omitempty" yaml:"arguments,omitempty"`
	When         string           `json:"when,omitempty" yaml:"when,omitempty"`
	ContinueOn   *ContinueOn     `json:"continueOn,omitempty" yaml:"continueOn,omitempty"`
	WithItems    []interface{}    `json:"withItems,omitempty" yaml:"withItems,omitempty"`
	WithParam    string           `json:"withParam,omitempty" yaml:"withParam,omitempty"`
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

// Task represents a node in a DAG.
type Task struct {
	// Name is the task name (must be unique within the DAG).
	Name string
	// Template is the template to invoke.
	Template string
	// TemplateRef references a template in a WorkflowTemplate.
	TemplateRef *TemplateRef
	// Dependencies are task names that must complete first.
	Dependencies []string
	// Depends is a complex dependency expression.
	Depends string
	// Arguments are the template arguments.
	Arguments []Parameter
	// ArgumentArtifacts are artifact arguments.
	ArgumentArtifacts []ArtifactBuilder
	// When is a conditional expression.
	When string
	// ContinueOn defines when to continue after failure.
	ContinueOn *ContinueOn
	// WithItems enables fan-out over a list.
	WithItems []interface{}
	// WithParam enables fan-out from a parameter.
	WithParam string
}

// GetOutputParameter returns a parameter reference for this task's output.
// Returns an Argo expression like "{{tasks.task-name.outputs.parameters.param-name}}".
func (t *Task) GetOutputParameter(paramName string) string {
	return fmt.Sprintf("{{tasks.%s.outputs.parameters.%s}}", t.Name, paramName)
}

// GetOutputResult returns a result reference for this task's output.
// Returns "{{tasks.task-name.outputs.result}}".
func (t *Task) GetOutputResult() string {
	return fmt.Sprintf("{{tasks.%s.outputs.result}}", t.Name)
}

// GetOutputArtifact returns an artifact reference for this task's output.
// Returns "{{tasks.task-name.outputs.artifacts.artifact-name}}".
func (t *Task) GetOutputArtifact(artifactName string) string {
	return fmt.Sprintf("{{tasks.%s.outputs.artifacts.%s}}", t.Name, artifactName)
}

// Then sets this task as a dependency of the other task.
// Returns the other task for chaining.
func (t *Task) Then(other *Task) *Task {
	if other.Depends == "" {
		other.Depends = t.Name
	} else {
		other.Depends = other.Depends + " " + string(OperatorAnd) + " " + t.Name
	}
	return other
}

// Or creates an OR dependency expression between tasks.
func (t *Task) Or(other *Task) string {
	return fmt.Sprintf("(%s %s %s)", t.Name, string(OperatorOr), other.Name)
}

// OnSuccess makes this task run when the other task succeeds.
func (t *Task) OnSuccess(other *Task) *Task {
	t.Depends = fmt.Sprintf("%s.%s", other.Name, TaskSucceeded)
	return t
}

// OnFailure makes this task run when the other task fails.
func (t *Task) OnFailure(other *Task) *Task {
	t.Depends = fmt.Sprintf("%s.%s", other.Name, TaskFailed)
	return t
}

// OnError makes this task run when the other task errors.
func (t *Task) OnError(other *Task) *Task {
	t.Depends = fmt.Sprintf("%s.%s", other.Name, TaskErrored)
	return t
}

// BuildDAGTask builds the serializable DAG task model.
func (t *Task) BuildDAGTask() (DAGTaskModel, error) {
	if t.Name == "" {
		return DAGTaskModel{}, fmt.Errorf("task name cannot be empty")
	}

	var args *ArgumentsModel
	if len(t.Arguments) > 0 || len(t.ArgumentArtifacts) > 0 {
		args = &ArgumentsModel{}
		for _, p := range t.Arguments {
			m, err := p.AsArgument()
			if err != nil {
				return DAGTaskModel{}, fmt.Errorf("task %q argument: %w", t.Name, err)
			}
			args.Parameters = append(args.Parameters, m)
		}
		for _, a := range t.ArgumentArtifacts {
			m, err := a.Build()
			if err != nil {
				return DAGTaskModel{}, fmt.Errorf("task %q artifact: %w", t.Name, err)
			}
			args.Artifacts = append(args.Artifacts, m)
		}
	}

	return DAGTaskModel{
		Name:         t.Name,
		Template:     t.Template,
		TemplateRef:  t.TemplateRef,
		Dependencies: t.Dependencies,
		Depends:      t.Depends,
		Arguments:    args,
		When:         t.When,
		ContinueOn:   t.ContinueOn,
		WithItems:    t.WithItems,
		WithParam:    t.WithParam,
	}, nil
}

// DAG represents an Argo Workflows DAG template.
type DAG struct {
	// Name is the template name.
	Name string
	// Tasks are the tasks in the DAG.
	Tasks []*Task
	// FailFast stops the DAG on the first task failure.
	FailFast *bool
	// Target is the target task to run.
	Target string
	// Inputs are the template inputs.
	Inputs []Parameter
	// Outputs are the template outputs.
	Outputs []Parameter
	// InputArtifacts are the input artifacts.
	InputArtifacts []ArtifactBuilder
	// OutputArtifacts are the output artifacts.
	OutputArtifacts []ArtifactBuilder
	// nodeNames tracks task names for conflict detection.
	nodeNames map[string]bool
}

// AddTask adds a task to the DAG. Returns error on name conflict.
func (d *DAG) AddTask(task *Task) error {
	if d.nodeNames == nil {
		d.nodeNames = make(map[string]bool)
	}
	if d.nodeNames[task.Name] {
		return &NodeNameConflict{Name: task.Name}
	}
	d.nodeNames[task.Name] = true
	d.Tasks = append(d.Tasks, task)
	return nil
}

// AddTasks adds multiple tasks. Stops on first error.
func (d *DAG) AddTasks(tasks ...*Task) error {
	for _, t := range tasks {
		if err := d.AddTask(t); err != nil {
			return err
		}
	}
	return nil
}

func (d *DAG) GetName() string {
	return d.Name
}

func (d *DAG) buildInputs() *InputsModel {
	var params []ParameterModel
	for _, p := range d.Inputs {
		m, err := p.AsInput()
		if err != nil {
			continue
		}
		params = append(params, m)
	}
	var arts []ArtifactModel
	for _, a := range d.InputArtifacts {
		m, err := a.Build()
		if err != nil {
			continue
		}
		arts = append(arts, m)
	}
	if len(params) == 0 && len(arts) == 0 {
		return nil
	}
	return &InputsModel{Parameters: params, Artifacts: arts}
}

func (d *DAG) buildOutputs() *OutputsModel {
	var params []ParameterModel
	for _, p := range d.Outputs {
		m, err := p.AsOutput()
		if err != nil {
			continue
		}
		params = append(params, m)
	}
	var arts []ArtifactModel
	for _, a := range d.OutputArtifacts {
		m, err := a.Build()
		if err != nil {
			continue
		}
		arts = append(arts, m)
	}
	if len(params) == 0 && len(arts) == 0 {
		return nil
	}
	return &OutputsModel{Parameters: params, Artifacts: arts}
}

// BuildTemplate builds the Argo Template for this DAG.
func (d *DAG) BuildTemplate() (TemplateModel, error) {
	if d.Name == "" {
		return TemplateModel{}, fmt.Errorf("DAG template name cannot be empty")
	}

	tasks := make([]DAGTaskModel, 0, len(d.Tasks))
	for _, t := range d.Tasks {
		m, err := t.BuildDAGTask()
		if err != nil {
			return TemplateModel{}, err
		}
		tasks = append(tasks, m)
	}

	return TemplateModel{
		Name: d.Name,
		DAG: &DAGModel{
			Tasks:    tasks,
			FailFast: d.FailFast,
			Target:   d.Target,
		},
		Inputs:  d.buildInputs(),
		Outputs: d.buildOutputs(),
	}, nil
}
