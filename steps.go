package forge

import (
	"fmt"

	"github.com/usetheodev/theo-forge/model"
)

// Step represents a single step in a Steps template.
type Step struct {
	// Name is the step name.
	Name string
	// Template is the template to invoke.
	Template string
	// TemplateRef references a template in a WorkflowTemplate.
	TemplateRef *model.TemplateRef
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
	// OnExit is the exit handler template name for this step.
	OnExit string
	// Hooks are lifecycle hooks for this step.
	Hooks map[string]model.LifecycleHook
}

// GetOutputParameter returns a parameter reference for this step's output.
func (s *Step) GetOutputParameter(paramName string) string {
	return fmt.Sprintf("{{steps.%s.outputs.parameters.%s}}", s.Name, paramName)
}

// GetOutputResult returns a result reference for this step's output.
func (s *Step) GetOutputResult() string {
	return fmt.Sprintf("{{steps.%s.outputs.result}}", s.Name)
}

// GetOutputArtifact returns an artifact reference for this step's output.
func (s *Step) GetOutputArtifact(artifactName string) string {
	return fmt.Sprintf("{{steps.%s.outputs.artifacts.%s}}", s.Name, artifactName)
}

// BuildStep builds the serializable step model.
func (s *Step) BuildStep() (model.StepModel, error) {
	if s.Name == "" {
		return model.StepModel{}, fmt.Errorf("step name cannot be empty")
	}

	var args *model.ArgumentsModel
	if len(s.Arguments) > 0 || len(s.ArgumentArtifacts) > 0 {
		args = &model.ArgumentsModel{}
		for _, p := range s.Arguments {
			m, err := p.AsArgument()
			if err != nil {
				return model.StepModel{}, fmt.Errorf("step %q argument: %w", s.Name, err)
			}
			args.Parameters = append(args.Parameters, m)
		}
		for _, a := range s.ArgumentArtifacts {
			m, err := a.Build()
			if err != nil {
				return model.StepModel{}, fmt.Errorf("step %q artifact: %w", s.Name, err)
			}
			args.Artifacts = append(args.Artifacts, m)
		}
	}

	var inline *model.TemplateModel
	if s.Inline != nil {
		m, err := s.Inline.BuildTemplate()
		if err != nil {
			return model.StepModel{}, fmt.Errorf("step %q inline: %w", s.Name, err)
		}
		inline = &m
	}

	return model.StepModel{
		Name:         s.Name,
		Template:     s.Template,
		TemplateRef:  s.TemplateRef,
		Inline:       inline,
		Arguments:    args,
		When:         s.When,
		ContinueOn:   s.ContinueOn,
		WithItems:    s.WithItems,
		WithParam:    s.WithParam,
		WithSequence: s.WithSequence,
		OnExit:       s.OnExit,
		Hooks:        s.Hooks,
	}, nil
}

// Parallel represents a group of steps that run in parallel.
type Parallel struct {
	Steps []*Step
	// nodeNames tracks step names for conflict detection.
	nodeNames map[string]bool
}

// AddStep adds a step to the parallel group. Returns error on name conflict.
func (p *Parallel) AddStep(step *Step) error {
	if p.nodeNames == nil {
		p.nodeNames = make(map[string]bool)
	}
	if p.nodeNames[step.Name] {
		return &NodeNameConflict{Name: step.Name}
	}
	p.nodeNames[step.Name] = true
	p.Steps = append(p.Steps, step)
	return nil
}

func (p *Parallel) buildSteps() ([]model.StepModel, error) {
	models := make([]model.StepModel, 0, len(p.Steps))
	for _, s := range p.Steps {
		m, err := s.BuildStep()
		if err != nil {
			return nil, err
		}
		models = append(models, m)
	}
	return models, nil
}

// Steps represents an Argo Workflows steps template.
// Each element in StepGroups runs sequentially; steps within a group run in parallel.
type Steps struct {
	// Name is the template name.
	Name string
	// StepGroups are groups of steps. Steps within a group run in parallel.
	StepGroups []Parallel
	// Inputs are the template inputs.
	Inputs []Parameter
	// Outputs are the template outputs.
	Outputs []Parameter
	// InputArtifacts are the input artifacts.
	InputArtifacts []ArtifactBuilder
	// OutputArtifacts are the output artifacts.
	OutputArtifacts []ArtifactBuilder
	// nodeNames tracks step names across all groups.
	nodeNames map[string]bool
}

// AddSequentialStep adds a single step as a new sequential group.
func (s *Steps) AddSequentialStep(step *Step) error {
	if s.nodeNames == nil {
		s.nodeNames = make(map[string]bool)
	}
	if s.nodeNames[step.Name] {
		return &NodeNameConflict{Name: step.Name}
	}
	s.nodeNames[step.Name] = true
	s.StepGroups = append(s.StepGroups, Parallel{
		Steps:     []*Step{step},
		nodeNames: map[string]bool{step.Name: true},
	})
	return nil
}

// AddParallelGroup adds a group of steps that run in parallel.
func (s *Steps) AddParallelGroup(steps ...*Step) error {
	if s.nodeNames == nil {
		s.nodeNames = make(map[string]bool)
	}
	p := Parallel{nodeNames: make(map[string]bool)}
	for _, step := range steps {
		if s.nodeNames[step.Name] {
			return &NodeNameConflict{Name: step.Name}
		}
		s.nodeNames[step.Name] = true
		if err := p.AddStep(step); err != nil {
			return err
		}
	}
	s.StepGroups = append(s.StepGroups, p)
	return nil
}

func (s *Steps) GetName() string {
	return s.Name
}

// BuildTemplate builds the Argo Template for this Steps template.
func (s *Steps) BuildTemplate() (model.TemplateModel, error) {
	if s.Name == "" {
		return model.TemplateModel{}, fmt.Errorf("steps template name cannot be empty")
	}

	steps := make([][]model.StepModel, 0, len(s.StepGroups))
	for _, group := range s.StepGroups {
		models, err := group.buildSteps()
		if err != nil {
			return model.TemplateModel{}, err
		}
		steps = append(steps, models)
	}

	inputs, err := buildInputsFromParams(s.Inputs, s.InputArtifacts)
	if err != nil {
		return model.TemplateModel{}, fmt.Errorf("steps %q: %w", s.Name, err)
	}

	outputs, err := buildOutputsFromParams(s.Outputs, s.OutputArtifacts)
	if err != nil {
		return model.TemplateModel{}, fmt.Errorf("steps %q: %w", s.Name, err)
	}

	return model.TemplateModel{
		Name:    s.Name,
		Steps:   steps,
		Inputs:  inputs,
		Outputs: outputs,
	}, nil
}
