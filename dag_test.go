package forge

import (
	"errors"
	"testing"

	"github.com/usetheo/theo/forge/model"
)

func TestTaskThen(t *testing.T) {
	a := &Task{Name: "task-a", Template: "echo"}
	b := &Task{Name: "task-b", Template: "echo"}
	result := a.Then(b)
	if result != b {
		t.Fatal("Then should return the other task")
	}
	if b.Depends != "task-a" {
		t.Errorf("depends = %q, want 'task-a'", b.Depends)
	}
}

func TestTaskThenChaining(t *testing.T) {
	a := &Task{Name: "task-a", Template: "echo"}
	b := &Task{Name: "task-b", Template: "echo"}
	c := &Task{Name: "task-c", Template: "echo"}
	a.Then(b).Then(c)
	if b.Depends != "task-a" {
		t.Errorf("b.depends = %q", b.Depends)
	}
	if c.Depends != "task-b" {
		t.Errorf("c.depends = %q", c.Depends)
	}
}

func TestTaskOr(t *testing.T) {
	a := &Task{Name: "task-a", Template: "echo"}
	b := &Task{Name: "task-b", Template: "echo"}
	expr := a.Or(b)
	if expr != "(task-a || task-b)" {
		t.Errorf("or = %q, want '(task-a || task-b)'", expr)
	}
}

func TestTaskOnSuccess(t *testing.T) {
	a := &Task{Name: "task-a", Template: "echo"}
	b := &Task{Name: "task-b", Template: "echo"}
	b.OnSuccess(a)
	if b.Depends != "task-a.Succeeded" {
		t.Errorf("depends = %q, want 'task-a.Succeeded'", b.Depends)
	}
}

func TestTaskOnFailure(t *testing.T) {
	a := &Task{Name: "task-a", Template: "echo"}
	b := &Task{Name: "task-b", Template: "echo"}
	b.OnFailure(a)
	if b.Depends != "task-a.Failed" {
		t.Errorf("depends = %q, want 'task-a.Failed'", b.Depends)
	}
}

func TestTaskOnError(t *testing.T) {
	a := &Task{Name: "task-a", Template: "echo"}
	b := &Task{Name: "task-b", Template: "echo"}
	b.OnError(a)
	if b.Depends != "task-a.Errored" {
		t.Errorf("depends = %q, want 'task-a.Errored'", b.Depends)
	}
}

func TestTaskBuildDAGTask(t *testing.T) {
	task := &Task{
		Name:     "my-task",
		Template: "echo",
		Arguments: []Parameter{
			{Name: "msg", Value: ptrStr("hello")},
		},
	}
	model, err := task.BuildDAGTask()
	if err != nil {
		t.Fatal(err)
	}
	if model.Name != "my-task" {
		t.Errorf("name = %q", model.Name)
	}
	if model.Template != "echo" {
		t.Errorf("template = %q", model.Template)
	}
	if model.Arguments == nil || len(model.Arguments.Parameters) != 1 {
		t.Fatal("expected 1 argument")
	}
}

func TestTaskBuildDAGTaskNoName(t *testing.T) {
	task := &Task{Template: "echo"}
	_, err := task.BuildDAGTask()
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestDAGAddTask(t *testing.T) {
	dag := &DAG{Name: "my-dag"}
	a := &Task{Name: "task-a", Template: "echo"}
	b := &Task{Name: "task-b", Template: "echo"}
	if err := dag.AddTasks(a, b); err != nil {
		t.Fatal(err)
	}
	if len(dag.Tasks) != 2 {
		t.Fatalf("tasks count = %d, want 2", len(dag.Tasks))
	}
}

func TestDAGAddTaskNameConflict(t *testing.T) {
	dag := &DAG{Name: "my-dag"}
	a := &Task{Name: "task-a", Template: "echo"}
	b := &Task{Name: "task-a", Template: "echo"} // same name
	err := dag.AddTasks(a, b)
	if err == nil {
		t.Fatal("expected name conflict error")
	}
	var conflict *NodeNameConflict
	if !errors.As(err, &conflict) {
		t.Fatalf("expected NodeNameConflict, got %T", err)
	}
	if conflict.Name != "task-a" {
		t.Errorf("conflict name = %q", conflict.Name)
	}
}

func TestDAGDifferentNamesForSameTemplate(t *testing.T) {
	dag := &DAG{Name: "my-dag"}
	a := &Task{Name: "call-1", Template: "echo"}
	b := &Task{Name: "call-2", Template: "echo"} // same template, different name
	if err := dag.AddTasks(a, b); err != nil {
		t.Fatalf("should allow different names for same template: %v", err)
	}
}

func TestDAGBuildTemplate(t *testing.T) {
	dag := &DAG{Name: "diamond"}
	a := &Task{Name: "A", Template: "echo"}
	b := &Task{Name: "B", Template: "echo"}
	c := &Task{Name: "C", Template: "echo"}
	d := &Task{Name: "D", Template: "echo"}

	a.Then(b)
	a.Then(c)
	b.Then(d)
	c.Then(d)

	dag.AddTasks(a, b, c, d)

	tpl, err := dag.BuildTemplate()
	if err != nil {
		t.Fatal(err)
	}
	if tpl.Name != "diamond" {
		t.Errorf("name = %q", tpl.Name)
	}
	if tpl.DAG == nil {
		t.Fatal("expected DAG to be set")
	}
	if len(tpl.DAG.Tasks) != 4 {
		t.Fatalf("dag tasks = %d, want 4", len(tpl.DAG.Tasks))
	}

	// Verify dependencies
	taskMap := make(map[string]model.DAGTaskModel)
	for _, task := range tpl.DAG.Tasks {
		taskMap[task.Name] = task
	}
	if taskMap["B"].Depends != "A" {
		t.Errorf("B depends = %q, want 'A'", taskMap["B"].Depends)
	}
	if taskMap["C"].Depends != "A" {
		t.Errorf("C depends = %q, want 'A'", taskMap["C"].Depends)
	}
	// D depends on both B and C
	if taskMap["D"].Depends != "B && C" {
		t.Errorf("D depends = %q, want 'B && C'", taskMap["D"].Depends)
	}
}

func TestDAGNoNameFails(t *testing.T) {
	dag := &DAG{}
	_, err := dag.BuildTemplate()
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestDAGWithInputs(t *testing.T) {
	dag := &DAG{
		Name:   "with-inputs",
		Inputs: []Parameter{{Name: "msg", Value: ptrStr("hello")}},
	}
	tpl, err := dag.BuildTemplate()
	if err != nil {
		t.Fatal(err)
	}
	if tpl.Inputs == nil || len(tpl.Inputs.Parameters) != 1 {
		t.Fatal("expected 1 input parameter")
	}
}

func TestDAGFailFast(t *testing.T) {
	ff := true
	dag := &DAG{
		Name:     "fail-fast",
		FailFast: &ff,
	}
	tpl, err := dag.BuildTemplate()
	if err != nil {
		t.Fatal(err)
	}
	if tpl.DAG.FailFast == nil || !*tpl.DAG.FailFast {
		t.Error("expected failFast to be true")
	}
}

func TestTaskWithTemplateRef(t *testing.T) {
	task := &Task{
		Name: "ref-task",
		TemplateRef: &TemplateRef{
			Name:     "my-workflow-template",
			Template: "echo",
		},
	}
	model, err := task.BuildDAGTask()
	if err != nil {
		t.Fatal(err)
	}
	if model.TemplateRef == nil {
		t.Fatal("expected templateRef")
	}
	if model.TemplateRef.Name != "my-workflow-template" {
		t.Errorf("templateRef.name = %q", model.TemplateRef.Name)
	}
}

func TestTaskWithContinueOn(t *testing.T) {
	task := &Task{
		Name:     "continue",
		Template: "may-fail",
		ContinueOn: &ContinueOn{
			Error:  true,
			Failed: true,
		},
	}
	model, err := task.BuildDAGTask()
	if err != nil {
		t.Fatal(err)
	}
	if model.ContinueOn == nil {
		t.Fatal("expected continueOn")
	}
	if !model.ContinueOn.Error || !model.ContinueOn.Failed {
		t.Error("expected both error and failed to be true")
	}
}

func TestTaskWithItems(t *testing.T) {
	task := &Task{
		Name:      "loop",
		Template:  "process",
		WithItems: []interface{}{"a", "b", "c"},
	}
	model, err := task.BuildDAGTask()
	if err != nil {
		t.Fatal(err)
	}
	if len(model.WithItems) != 3 {
		t.Errorf("withItems len = %d, want 3", len(model.WithItems))
	}
}
