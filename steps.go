package forge

import "fmt"

// StepModel is the serializable Argo Workflow step.
type StepModel struct {
	Name        string          `json:"name" yaml:"name"`
	Template    string          `json:"template,omitempty" yaml:"template,omitempty"`
	TemplateRef *TemplateRef    `json:"templateRef,omitempty" yaml:"templateRef,omitempty"`
	Arguments   *ArgumentsModel `json:"arguments,omitempty" yaml:"arguments,omitempty"`
	When        string          `json:"when,omitempty" yaml:"when,omitempty"`
	ContinueOn  *ContinueOn    `json:"continueOn,omitempty" yaml:"continueOn,omitempty"`
	WithItems   []interface{}   `json:"withItems,omitempty" yaml:"withItems,omitempty"`
	WithParam   string          `json:"withParam,omitempty" yaml:"withParam,omitempty"`
}

// Step represents a single step in a Steps template.
type Step struct {
	// Name is the step name.
	Name string
	// Template is the template to invoke.
	Template string
	// TemplateRef references a template in a WorkflowTemplate.
	TemplateRef *TemplateRef
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
func (s *Step) BuildStep() (StepModel, error) {
	if s.Name == "" {
		return StepModel{}, fmt.Errorf("step name cannot be empty")
	}

	var args *ArgumentsModel
	if len(s.Arguments) > 0 || len(s.ArgumentArtifacts) > 0 {
		args = &ArgumentsModel{}
		for _, p := range s.Arguments {
			m, err := p.AsArgument()
			if err != nil {
				return StepModel{}, fmt.Errorf("step %q argument: %w", s.Name, err)
			}
			args.Parameters = append(args.Parameters, m)
		}
		for _, a := range s.ArgumentArtifacts {
			m, err := a.Build()
			if err != nil {
				return StepModel{}, fmt.Errorf("step %q artifact: %w", s.Name, err)
			}
			args.Artifacts = append(args.Artifacts, m)
		}
	}

	return StepModel{
		Name:        s.Name,
		Template:    s.Template,
		TemplateRef: s.TemplateRef,
		Arguments:   args,
		When:        s.When,
		ContinueOn:  s.ContinueOn,
		WithItems:   s.WithItems,
		WithParam:   s.WithParam,
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

func (p *Parallel) buildSteps() ([]StepModel, error) {
	models := make([]StepModel, 0, len(p.Steps))
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

func (s *Steps) buildInputs() *InputsModel {
	var params []ParameterModel
	for _, p := range s.Inputs {
		m, err := p.AsInput()
		if err != nil {
			continue
		}
		params = append(params, m)
	}
	var arts []ArtifactModel
	for _, a := range s.InputArtifacts {
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

func (s *Steps) buildOutputs() *OutputsModel {
	var params []ParameterModel
	for _, p := range s.Outputs {
		m, err := p.AsOutput()
		if err != nil {
			continue
		}
		params = append(params, m)
	}
	var arts []ArtifactModel
	for _, a := range s.OutputArtifacts {
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

// BuildTemplate builds the Argo Template for this Steps template.
func (s *Steps) BuildTemplate() (TemplateModel, error) {
	if s.Name == "" {
		return TemplateModel{}, fmt.Errorf("steps template name cannot be empty")
	}

	steps := make([][]StepModel, 0, len(s.StepGroups))
	for _, group := range s.StepGroups {
		models, err := group.buildSteps()
		if err != nil {
			return TemplateModel{}, err
		}
		steps = append(steps, models)
	}

	return TemplateModel{
		Name:    s.Name,
		Steps:   steps,
		Inputs:  s.buildInputs(),
		Outputs: s.buildOutputs(),
	}, nil
}
