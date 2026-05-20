package forge

import (
	"fmt"

	"github.com/usetheodev/theo-forge/model"
)

// Operator defines how task dependencies are combined.
type Operator string

// Operator literals.
const (
	OperatorAnd Operator = "&&"
	OperatorOr  Operator = "||"
)

// TaskResult represents the result of a task for conditional dependencies.
type TaskResult string

// TaskResult literals (Argo task lifecycle states).
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

// Task represents a node in a DAG.
type Task struct {
	// Name is the task name (must be unique within the DAG).
	Name string
	// Template is the template to invoke.
	Template string
	// TemplateRef references a template in a WorkflowTemplate.
	TemplateRef *model.TemplateRef
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
	ContinueOn *model.ContinueOn
	// WithItems enables fan-out over a list.
	WithItems []interface{}
	// WithParam enables fan-out from a parameter.
	WithParam string
	// WithSequence generates a list of numbers for fan-out.
	WithSequence *model.Sequence
	// Inline is an inline template definition.
	Inline Templatable
	// OnExit is the exit handler template name for this task.
	OnExit string
	// Hooks are lifecycle hooks for this task.
	Hooks map[string]model.LifecycleHook
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
//
// (T3.9 / completeness-h4-then-precedence) When other.Depends already contains
// an OR operator (||), the existing expression is wrapped in parentheses
// before the AND is appended. This preserves operator precedence so that
//
//	A.Then(B); X.Or(Y) into B.Depends           // existing: "X || Y"
//	A.Then(B)                                    // append A
//
// produces "(X || Y) && A" — NOT the ambiguous "X || Y && A" which would
// be parsed as "X || (Y && A)" by Argo's expression engine.
func (t *Task) Then(other *Task) *Task {
	if other.Depends == "" {
		other.Depends = t.Name
		return other
	}
	prev := other.Depends
	if needsParens(prev) {
		prev = "(" + prev + ")"
	}
	other.Depends = prev + " " + string(OperatorAnd) + " " + t.Name
	return other
}

// needsParens reports whether a Depends expression contains an OR operator
// at its top level and therefore must be parenthesized before further AND
// composition.
func needsParens(expr string) bool {
	return containsTopLevelOr(expr)
}

// containsTopLevelOr returns true if expr contains "||" outside any parens.
func containsTopLevelOr(expr string) bool {
	depth := 0
	for i := 0; i < len(expr); i++ {
		switch expr[i] {
		case '(':
			depth++
		case ')':
			if depth > 0 {
				depth--
			}
		case '|':
			if depth == 0 && i+1 < len(expr) && expr[i+1] == '|' {
				return true
			}
		}
	}
	return false
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
// Returns model.ErrTemplateAmbiguous when more than one of Template/TemplateRef/Inline
// is set; model.ErrTemplateMissing when none is set. (T3.3 / code-p4-template-ref-mutual-exclusion).
func (t *Task) BuildDAGTask() (model.DAGTaskModel, error) {
	if t.Name == "" {
		return model.DAGTaskModel{}, fmt.Errorf("task name cannot be empty")
	}
	if err := validateTemplateReference(t.Template, t.TemplateRef, t.Inline, "task "+t.Name); err != nil {
		return model.DAGTaskModel{}, err
	}

	var args *model.ArgumentsModel
	if len(t.Arguments) > 0 || len(t.ArgumentArtifacts) > 0 {
		args = &model.ArgumentsModel{}
		for _, p := range t.Arguments {
			m, err := p.AsArgument()
			if err != nil {
				return model.DAGTaskModel{}, fmt.Errorf("task %q argument: %w", t.Name, err)
			}
			args.Parameters = append(args.Parameters, m)
		}
		for _, a := range t.ArgumentArtifacts {
			m, err := a.Build()
			if err != nil {
				return model.DAGTaskModel{}, fmt.Errorf("task %q artifact: %w", t.Name, err)
			}
			args.Artifacts = append(args.Artifacts, m)
		}
	}

	var inline *model.TemplateModel
	if t.Inline != nil {
		m, err := t.Inline.BuildTemplate()
		if err != nil {
			return model.DAGTaskModel{}, fmt.Errorf("task %q inline: %w", t.Name, err)
		}
		inline = &m
	}

	return model.DAGTaskModel{
		Name:         t.Name,
		Template:     t.Template,
		TemplateRef:  t.TemplateRef,
		Inline:       inline,
		Dependencies: t.Dependencies,
		Depends:      t.Depends,
		Arguments:    args,
		When:         t.When,
		ContinueOn:   t.ContinueOn,
		WithItems:    t.WithItems,
		WithParam:    t.WithParam,
		WithSequence: t.WithSequence,
		OnExit:       t.OnExit,
		Hooks:        t.Hooks,
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

// GetName is the method.
func (d *DAG) GetName() string {
	return d.Name
}

// BuildTemplate builds the Argo Template for this DAG.
func (d *DAG) BuildTemplate() (model.TemplateModel, error) {
	if d.Name == "" {
		return model.TemplateModel{}, fmt.Errorf("DAG template name cannot be empty")
	}

	tasks := make([]model.DAGTaskModel, 0, len(d.Tasks))
	for _, t := range d.Tasks {
		m, err := t.BuildDAGTask()
		if err != nil {
			return model.TemplateModel{}, err
		}
		tasks = append(tasks, m)
	}

	inputs, err := buildInputsFromParams(d.Inputs, d.InputArtifacts)
	if err != nil {
		return model.TemplateModel{}, fmt.Errorf("DAG %q: %w", d.Name, err)
	}

	outputs, err := buildOutputsFromParams(d.Outputs, d.OutputArtifacts)
	if err != nil {
		return model.TemplateModel{}, fmt.Errorf("DAG %q: %w", d.Name, err)
	}

	return model.TemplateModel{
		Name: d.Name,
		DAG: &model.DAGModel{
			Tasks:    tasks,
			FailFast: d.FailFast,
			Target:   d.Target,
		},
		Inputs:  inputs,
		Outputs: outputs,
	}, nil
}
