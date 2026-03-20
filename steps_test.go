package forge

import (
	"errors"
	"testing"
)

func TestStepBuild(t *testing.T) {
	s := &Step{
		Name:     "echo-step",
		Template: "echo",
		Arguments: []Parameter{
			{Name: "msg", Value: ptrStr("hello")},
		},
	}
	model, err := s.BuildStep()
	if err != nil {
		t.Fatal(err)
	}
	if model.Name != "echo-step" {
		t.Errorf("name = %q", model.Name)
	}
	if model.Template != "echo" {
		t.Errorf("template = %q", model.Template)
	}
	if model.Arguments == nil || len(model.Arguments.Parameters) != 1 {
		t.Fatal("expected 1 argument")
	}
}

func TestStepNoNameFails(t *testing.T) {
	s := &Step{Template: "echo"}
	_, err := s.BuildStep()
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestStepWithCondition(t *testing.T) {
	s := &Step{
		Name:     "conditional",
		Template: "process",
		When:     "{{steps.check.outputs.result}} == pass",
	}
	model, err := s.BuildStep()
	if err != nil {
		t.Fatal(err)
	}
	if model.When != "{{steps.check.outputs.result}} == pass" {
		t.Errorf("when = %q", model.When)
	}
}

func TestParallelAddStep(t *testing.T) {
	p := &Parallel{}
	a := &Step{Name: "step-a", Template: "echo"}
	b := &Step{Name: "step-b", Template: "echo"}
	if err := p.AddStep(a); err != nil {
		t.Fatal(err)
	}
	if err := p.AddStep(b); err != nil {
		t.Fatal(err)
	}
	if len(p.Steps) != 2 {
		t.Fatalf("steps count = %d", len(p.Steps))
	}
}

func TestParallelNameConflict(t *testing.T) {
	p := &Parallel{}
	a := &Step{Name: "step-a", Template: "echo"}
	b := &Step{Name: "step-a", Template: "echo"} // same name
	_ = p.AddStep(a)
	err := p.AddStep(b)
	if err == nil {
		t.Fatal("expected name conflict")
	}
	var conflict *NodeNameConflict
	if !errors.As(err, &conflict) {
		t.Fatalf("expected NodeNameConflict, got %T", err)
	}
}

func TestStepsSequential(t *testing.T) {
	steps := &Steps{Name: "sequential"}
	a := &Step{Name: "first", Template: "echo"}
	b := &Step{Name: "second", Template: "echo"}
	c := &Step{Name: "third", Template: "echo"}

	if err := steps.AddSequentialStep(a); err != nil {
		t.Fatal(err)
	}
	if err := steps.AddSequentialStep(b); err != nil {
		t.Fatal(err)
	}
	if err := steps.AddSequentialStep(c); err != nil {
		t.Fatal(err)
	}

	tpl, err := steps.BuildTemplate()
	if err != nil {
		t.Fatal(err)
	}
	if tpl.Name != "sequential" {
		t.Errorf("name = %q", tpl.Name)
	}
	if len(tpl.Steps) != 3 {
		t.Fatalf("step groups = %d, want 3", len(tpl.Steps))
	}
	// Each group has 1 step
	for i, group := range tpl.Steps {
		if len(group) != 1 {
			t.Errorf("group[%d] steps = %d, want 1", i, len(group))
		}
	}
}

func TestStepsParallel(t *testing.T) {
	steps := &Steps{Name: "parallel"}
	a := &Step{Name: "step-a", Template: "echo"}
	b := &Step{Name: "step-b", Template: "echo"}

	if err := steps.AddParallelGroup(a, b); err != nil {
		t.Fatal(err)
	}

	tpl, err := steps.BuildTemplate()
	if err != nil {
		t.Fatal(err)
	}
	if len(tpl.Steps) != 1 {
		t.Fatalf("step groups = %d, want 1", len(tpl.Steps))
	}
	if len(tpl.Steps[0]) != 2 {
		t.Fatalf("parallel steps = %d, want 2", len(tpl.Steps[0]))
	}
}

func TestStepsMixed(t *testing.T) {
	steps := &Steps{Name: "mixed"}
	a := &Step{Name: "setup", Template: "init"}
	b := &Step{Name: "build", Template: "build"}
	c := &Step{Name: "test", Template: "test"}
	d := &Step{Name: "cleanup", Template: "cleanup"}

	// Sequential: setup
	_ = steps.AddSequentialStep(a)
	// Parallel: build + test
	_ = steps.AddParallelGroup(b, c)
	// Sequential: cleanup
	_ = steps.AddSequentialStep(d)

	tpl, err := steps.BuildTemplate()
	if err != nil {
		t.Fatal(err)
	}
	if len(tpl.Steps) != 3 {
		t.Fatalf("step groups = %d, want 3", len(tpl.Steps))
	}
	if len(tpl.Steps[0]) != 1 {
		t.Errorf("group[0] = %d, want 1", len(tpl.Steps[0]))
	}
	if len(tpl.Steps[1]) != 2 {
		t.Errorf("group[1] = %d, want 2", len(tpl.Steps[1]))
	}
	if len(tpl.Steps[2]) != 1 {
		t.Errorf("group[2] = %d, want 1", len(tpl.Steps[2]))
	}
}

func TestStepsNameConflictAcrossGroups(t *testing.T) {
	steps := &Steps{Name: "conflict"}
	a := &Step{Name: "step-a", Template: "echo"}
	b := &Step{Name: "step-a", Template: "echo"} // same name, different group

	_ = steps.AddSequentialStep(a)
	err := steps.AddSequentialStep(b)
	if err == nil {
		t.Fatal("expected name conflict across groups")
	}
}

func TestStepsNoNameFails(t *testing.T) {
	steps := &Steps{}
	_, err := steps.BuildTemplate()
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestStepsWithInputs(t *testing.T) {
	steps := &Steps{
		Name:   "with-inputs",
		Inputs: []Parameter{{Name: "msg", Value: ptrStr("hello")}},
	}
	tpl, err := steps.BuildTemplate()
	if err != nil {
		t.Fatal(err)
	}
	if tpl.Inputs == nil || len(tpl.Inputs.Parameters) != 1 {
		t.Fatal("expected 1 input parameter")
	}
}
