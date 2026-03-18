package forge

import (
	"strings"
	"testing"
)

// TestExampleDefaultParameterOverwrite replicates Hera's default-parameters.yaml
func TestExampleDefaultParameterOverwrite(t *testing.T) {
	generator := &Script{
		Name:    "generator",
		Image:   "python:3.10",
		Command: []string{"python"},
		Source:  "print('Another message for the world!')",
	}

	consumer := &Script{
		Name:    "consumer",
		Image:   "python:3.10",
		Command: []string{"python"},
		Source:  "print('{{inputs.parameters.message}}')",
		Inputs: []Parameter{
			{Name: "message", Default: ptrStr("Hello, world!")},
			{Name: "foo", Default: ptrStr("42")},
		},
	}

	dag := &DAG{Name: "d"}
	genTask := &Task{Name: "generator", Template: "generator"}
	consumeDefault := &Task{Name: "consume-default", Template: "consumer"}
	consumeArg := &Task{
		Name:     "consume-argument",
		Template: "consumer",
		Arguments: []Parameter{
			{Name: "message", Value: ptrStr(genTask.GetOutputResult())},
		},
	}
	genTask.Then(consumeDefault)
	genTask.Then(consumeArg)
	dag.AddTasks(genTask, consumeDefault, consumeArg)

	w := &Workflow{
		GenerateName: "default-param-overwrite-",
		Entrypoint:   "d",
		Templates:    []Templatable{dag, generator, consumer},
	}

	y, err := w.ToYAML()
	if err != nil {
		t.Fatal(err)
	}

	// Verify key structural elements
	for _, s := range []string{
		"generateName: default-param-overwrite-",
		"entrypoint: d",
		"name: generator",
		"name: consumer",
		"name: consume-default",
		"name: consume-argument",
		"default: Hello, world!",
		"default: \"42\"",
	} {
		if !strings.Contains(y, s) {
			t.Errorf("YAML missing: %q\n\nFull YAML:\n%s", s, y)
		}
	}
}

// TestExampleOutputParameterPassing replicates Hera's output-parameters.yaml
func TestExampleOutputParameterPassing(t *testing.T) {
	outScript := &Script{
		Name:    "out",
		Image:   "python:3.10",
		Command: []string{"python"},
		Source:  "with open('/test', 'w') as f:\n    f.write('test')",
		Outputs: []Parameter{
			{Name: "a", ValueFrom: &ValueFrom{Path: "/test"}},
		},
	}

	inScript := &Script{
		Name:    "in-",
		Image:   "python:3.10",
		Command: []string{"python"},
		Source:  "print('{{inputs.parameters.a}}')",
		Inputs:  []Parameter{{Name: "a"}},
	}

	dag := &DAG{Name: "d"}
	outTask := &Task{Name: "out", Template: "out"}
	inTask := &Task{
		Name:     "in-",
		Template: "in-",
		Arguments: []Parameter{
			{Name: "a", Value: ptrStr(outTask.GetOutputParameter("a"))},
		},
	}
	outTask.Then(inTask)
	dag.AddTasks(outTask, inTask)

	w := &Workflow{
		GenerateName: "script-output-param-passing-",
		Entrypoint:   "d",
		Templates:    []Templatable{dag, outScript, inScript},
	}

	y, err := w.ToYAML()
	if err != nil {
		t.Fatal(err)
	}

	// Verify the output parameter reference is correct
	if !strings.Contains(y, "{{tasks.out.outputs.parameters.a}}") {
		t.Errorf("YAML missing output parameter reference\n\n%s", y)
	}
	// Verify outputs section
	if !strings.Contains(y, "valueFrom:") {
		t.Error("YAML missing valueFrom")
	}
	if !strings.Contains(y, "path: /test") {
		t.Error("YAML missing path: /test")
	}
}

// TestExampleWithItemsLoop replicates Hera's loop patterns
func TestExampleWithItemsLoop(t *testing.T) {
	echo := &Container{
		Name:    "echo",
		Image:   "alpine:3.18",
		Command: []string{"echo"},
		Args:    []string{"{{inputs.parameters.message}}"},
		Inputs:  []Parameter{{Name: "message"}},
	}

	dag := &DAG{Name: "main"}
	loopTask := &Task{
		Name:     "echo-loop",
		Template: "echo",
		Arguments: []Parameter{
			{Name: "message", Value: ptrStr("{{item}}")},
		},
		WithItems: []interface{}{"hello", "world", "foo"},
	}
	dag.AddTask(loopTask)

	w := &Workflow{
		GenerateName: "loops-",
		Entrypoint:   "main",
		Templates:    []Templatable{echo, dag},
	}

	y, err := w.ToYAML()
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(y, "withItems:") {
		t.Error("YAML missing withItems")
	}
	if !strings.Contains(y, "hello") || !strings.Contains(y, "world") {
		t.Error("YAML missing items")
	}
}

// TestExampleWithParamLoop tests withParam-based fan-out
func TestExampleWithParamLoop(t *testing.T) {
	generate := &Script{
		Name:    "generate-list",
		Image:   "python:3.11",
		Command: []string{"python"},
		Source:  `import json; print(json.dumps(["a", "b", "c"]))`,
	}

	process := &Container{
		Name:    "process",
		Image:   "alpine:3.18",
		Command: []string{"echo"},
		Args:    []string{"{{inputs.parameters.item}}"},
		Inputs:  []Parameter{{Name: "item"}},
	}

	dag := &DAG{Name: "main"}
	genTask := &Task{Name: "gen", Template: "generate-list"}
	processTask := &Task{
		Name:     "process",
		Template: "process",
		Arguments: []Parameter{
			{Name: "item", Value: ptrStr("{{item}}")},
		},
		WithParam: genTask.GetOutputResult(),
	}
	genTask.Then(processTask)
	dag.AddTasks(genTask, processTask)

	w := &Workflow{
		GenerateName: "param-loop-",
		Entrypoint:   "main",
		Templates:    []Templatable{generate, process, dag},
	}

	y, err := w.ToYAML()
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(y, "withParam:") {
		t.Error("YAML missing withParam")
	}
	if !strings.Contains(y, "{{tasks.gen.outputs.result}}") {
		t.Error("YAML missing result reference in withParam")
	}
}

// TestExampleRetryWithBackoff tests retry configuration
func TestExampleRetryWithBackoff(t *testing.T) {
	limit := 3
	factor := 2
	w := &Workflow{
		GenerateName: "retry-",
		Entrypoint:   "main",
		Templates: []Templatable{
			&Script{
				Name:    "main",
				Image:   "python:3.11",
				Command: []string{"python"},
				Source:  "import random; assert random.random() > 0.5",
				RetryStrategy: &RetryStrategy{
					Limit:       &limit,
					RetryPolicy: RetryOnFailure,
					Backoff: &Backoff{
						Duration:    "5s",
						Factor:      &factor,
						MaxDuration: "1m",
					},
				},
			},
		},
	}

	y, err := w.ToYAML()
	if err != nil {
		t.Fatal(err)
	}

	for _, s := range []string{"retryStrategy:", "retryPolicy: OnFailure", "duration: 5s", "maxDuration: 1m"} {
		if !strings.Contains(y, s) {
			t.Errorf("YAML missing: %q", s)
		}
	}
}

// TestExampleSuspendApprovalGate tests manual approval pattern
func TestExampleSuspendApprovalGate(t *testing.T) {
	steps := &Steps{Name: "approval-flow"}
	steps.AddSequentialStep(&Step{Name: "deploy-staging", Template: "deploy"})
	steps.AddSequentialStep(&Step{Name: "wait-approval", Template: "approve"})
	steps.AddSequentialStep(&Step{Name: "deploy-prod", Template: "deploy"})

	w := &Workflow{
		Name:       "approval-gate",
		Entrypoint: "approval-flow",
		Templates: []Templatable{
			&Container{Name: "deploy", Image: "alpine", Command: []string{"echo"}, Args: []string{"deploying..."}},
			&Suspend{Name: "approve"},
			steps,
		},
	}

	y, err := w.ToYAML()
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(y, "suspend:") {
		t.Error("YAML missing suspend template")
	}
	if !strings.Contains(y, "deploy-staging") || !strings.Contains(y, "deploy-prod") {
		t.Error("YAML missing step names")
	}
}

// TestExampleMultiClusterTemplateRef tests referencing ClusterWorkflowTemplate
func TestExampleMultiClusterTemplateRef(t *testing.T) {
	// Define cluster-wide template
	cwt := &ClusterWorkflowTemplate{
		Name:       "shared-build",
		Entrypoint: "build",
		Templates: []Templatable{
			&Container{
				Name:    "build",
				Image:   "golang:1.22",
				Command: []string{"go"},
				Args:    []string{"build", "./..."},
				Inputs:  []Parameter{{Name: "repo"}},
			},
		},
	}
	cwtYAML, err := cwt.ToYAML()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(cwtYAML, "kind: ClusterWorkflowTemplate") {
		t.Error("CWT YAML missing kind")
	}

	// Use it via templateRef
	dag := &DAG{Name: "pipeline"}
	dag.AddTask(&Task{
		Name: "build",
		TemplateRef: &TemplateRef{
			Name:     "shared-build",
			Template: "build",
		},
		Arguments: []Parameter{
			{Name: "repo", Value: ptrStr("https://github.com/example/app.git")},
		},
	})

	w := &Workflow{
		GenerateName: "ci-",
		Entrypoint:   "pipeline",
		Templates:    []Templatable{dag},
	}

	y, err := w.ToYAML()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(y, "templateRef:") {
		t.Error("YAML missing templateRef")
	}
	if !strings.Contains(y, "name: shared-build") {
		t.Error("YAML missing CWT reference")
	}
}
