package forge

import (
	"testing"
)

func TestScriptBuildTemplate(t *testing.T) {
	s := &Script{
		Name:    "hello-script",
		Image:   "python:3.11",
		Command: []string{"python"},
		Source:  `print("hello world")`,
	}
	tpl, err := s.BuildTemplate()
	if err != nil {
		t.Fatal(err)
	}
	if tpl.Name != "hello-script" {
		t.Errorf("name = %q", tpl.Name)
	}
	if tpl.Script == nil {
		t.Fatal("expected script to be set")
	}
	if tpl.Script.Image != "python:3.11" {
		t.Errorf("image = %q", tpl.Script.Image)
	}
	if tpl.Script.Source != `print("hello world")` {
		t.Errorf("source = %q", tpl.Script.Source)
	}
}

func TestScriptNoNameFails(t *testing.T) {
	s := &Script{Image: "python:3.11", Source: "print()"}
	_, err := s.BuildTemplate()
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestScriptNoSourceFails(t *testing.T) {
	s := &Script{Name: "test", Image: "python:3.11"}
	_, err := s.BuildTemplate()
	if err == nil {
		t.Fatal("expected error for empty source")
	}
}

func TestScriptWithBashCommand(t *testing.T) {
	s := &Script{
		Name:    "bash-script",
		Image:   "alpine",
		Command: []string{"sh", "-c"},
		Source:  `echo "hello"`,
	}
	tpl, err := s.BuildTemplate()
	if err != nil {
		t.Fatal(err)
	}
	if len(tpl.Script.Command) != 2 || tpl.Script.Command[0] != "sh" {
		t.Errorf("command = %v", tpl.Script.Command)
	}
}

func TestScriptWithEnv(t *testing.T) {
	s := &Script{
		Name:    "with-env",
		Image:   "python:3.11",
		Command: []string{"python"},
		Source:  "import os; print(os.environ['FOO'])",
		Env: []EnvBuilder{
			Env{Name: "FOO", Value: "bar"},
		},
	}
	tpl, err := s.BuildTemplate()
	if err != nil {
		t.Fatal(err)
	}
	if len(tpl.Script.Env) != 1 {
		t.Fatalf("env count = %d", len(tpl.Script.Env))
	}
	if tpl.Script.Env[0].Name != "FOO" {
		t.Errorf("env name = %q", tpl.Script.Env[0].Name)
	}
}

func TestScriptWithInputs(t *testing.T) {
	s := &Script{
		Name:    "with-inputs",
		Image:   "python:3.11",
		Command: []string{"python"},
		Source:  "print('{{inputs.parameters.msg}}')",
		Inputs: []Parameter{
			{Name: "msg", Default: ptrStr("hello")},
		},
	}
	tpl, err := s.BuildTemplate()
	if err != nil {
		t.Fatal(err)
	}
	if tpl.Inputs == nil || len(tpl.Inputs.Parameters) != 1 {
		t.Fatal("expected 1 input parameter")
	}
}

func TestScriptWithOutputs(t *testing.T) {
	s := &Script{
		Name:    "with-outputs",
		Image:   "python:3.11",
		Command: []string{"python"},
		Source:  "open('/tmp/result', 'w').write('42')",
		Outputs: []Parameter{
			{Name: "result", ValueFrom: &ValueFrom{Path: "/tmp/result"}},
		},
	}
	tpl, err := s.BuildTemplate()
	if err != nil {
		t.Fatal(err)
	}
	if tpl.Outputs == nil || len(tpl.Outputs.Parameters) != 1 {
		t.Fatal("expected 1 output parameter")
	}
}

func TestScriptWithResources(t *testing.T) {
	s := &Script{
		Name:    "with-resources",
		Image:   "python:3.11",
		Command: []string{"python"},
		Source:  "print('hello')",
		Resources: &ResourceRequirements{
			Requests: ResourceList{CPU: "100m", Memory: "128Mi"},
			Limits:   ResourceList{CPU: "500m", Memory: "512Mi"},
		},
	}
	tpl, err := s.BuildTemplate()
	if err != nil {
		t.Fatal(err)
	}
	if tpl.Script.Resources == nil {
		t.Fatal("expected resources")
	}
	if tpl.Script.Resources.Requests.CPU != "100m" {
		t.Errorf("cpu = %q", tpl.Script.Resources.Requests.CPU)
	}
}

func TestScriptWithRetry(t *testing.T) {
	limit := 5
	s := &Script{
		Name:    "with-retry",
		Image:   "python:3.11",
		Command: []string{"python"},
		Source:  "raise Exception('fail')",
		RetryStrategy: &RetryStrategy{
			Limit:       &limit,
			RetryPolicy: RetryAlways,
		},
	}
	tpl, err := s.BuildTemplate()
	if err != nil {
		t.Fatal(err)
	}
	if tpl.RetryStrategy == nil {
		t.Fatal("expected retry strategy")
	}
	if *tpl.RetryStrategy.Limit != 5 {
		t.Errorf("limit = %d", *tpl.RetryStrategy.Limit)
	}
}

func TestScriptMultilineSource(t *testing.T) {
	source := `import json
import sys

data = json.loads(sys.argv[1])
print(f"Name: {data['name']}")
print(f"Value: {data['value']}")
`
	s := &Script{
		Name:    "multiline",
		Image:   "python:3.11",
		Command: []string{"python"},
		Source:  source,
	}
	tpl, err := s.BuildTemplate()
	if err != nil {
		t.Fatal(err)
	}
	if tpl.Script.Source != source {
		t.Errorf("source mismatch")
	}
}
