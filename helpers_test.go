package forge

import (
	"context"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/usetheo/theo/forge/client"
	"github.com/usetheo/theo/forge/expr"
)

// --- File I/O tests (consolidated from file_io_test.go) ---

func TestWorkflowToFile(t *testing.T) {
	tmpDir := t.TempDir()

	w := &Workflow{
		Name:       "file-test",
		Entrypoint: "main",
		Templates:  []Templatable{&Container{Name: "main", Image: "alpine"}},
	}

	path, err := w.ToFile(tmpDir, "")
	if err != nil {
		t.Fatal(err)
	}

	// Check file was created
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("file not created at %s", path)
	}

	// Check filename
	expectedName := "file-test.yaml"
	if filepath.Base(path) != expectedName {
		t.Errorf("filename = %q, want %q", filepath.Base(path), expectedName)
	}
}

func TestWorkflowToFileCustomName(t *testing.T) {
	tmpDir := t.TempDir()

	w := &Workflow{
		Name:       "test",
		Entrypoint: "main",
		Templates:  []Templatable{&Container{Name: "main", Image: "alpine"}},
	}

	path, err := w.ToFile(tmpDir, "custom-name.yaml")
	if err != nil {
		t.Fatal(err)
	}

	if filepath.Base(path) != "custom-name.yaml" {
		t.Errorf("filename = %q", filepath.Base(path))
	}
}

func TestWorkflowToFileGenerateName(t *testing.T) {
	tmpDir := t.TempDir()

	w := &Workflow{
		GenerateName: "my-wf-",
		Entrypoint:   "main",
		Templates:    []Templatable{&Container{Name: "main", Image: "alpine"}},
	}

	path, err := w.ToFile(tmpDir, "")
	if err != nil {
		t.Fatal(err)
	}

	if filepath.Base(path) != "my-wf.yaml" {
		t.Errorf("filename = %q, want 'my-wf.yaml'", filepath.Base(path))
	}
}

func TestWorkflowToFileAddsYAMLExtension(t *testing.T) {
	tmpDir := t.TempDir()

	w := &Workflow{
		Name:       "test",
		Entrypoint: "main",
		Templates:  []Templatable{&Container{Name: "main", Image: "alpine"}},
	}

	path, err := w.ToFile(tmpDir, "no-extension")
	if err != nil {
		t.Fatal(err)
	}

	if filepath.Base(path) != "no-extension.yaml" {
		t.Errorf("filename = %q", filepath.Base(path))
	}
}

func TestFromFile(t *testing.T) {
	tmpDir := t.TempDir()

	w := &Workflow{
		Name:       "round-trip-file",
		Namespace:  "argo",
		Entrypoint: "main",
		Templates:  []Templatable{&Container{Name: "main", Image: "alpine", Command: []string{"echo"}, Args: []string{"hello"}}},
	}

	path, err := w.ToFile(tmpDir, "")
	if err != nil {
		t.Fatal(err)
	}

	model, err := FromFile(path)
	if err != nil {
		t.Fatal(err)
	}

	if model.Metadata.Name != "round-trip-file" {
		t.Errorf("name = %q", model.Metadata.Name)
	}
	if model.Metadata.Namespace != "argo" {
		t.Errorf("namespace = %q", model.Metadata.Namespace)
	}
	if model.Spec.Entrypoint != "main" {
		t.Errorf("entrypoint = %q", model.Spec.Entrypoint)
	}
	if len(model.Spec.Templates) != 1 {
		t.Fatalf("templates = %d", len(model.Spec.Templates))
	}
	if model.Spec.Templates[0].Container == nil {
		t.Fatal("expected container template")
	}
	if model.Spec.Templates[0].Container.Image != "alpine" {
		t.Errorf("image = %q", model.Spec.Templates[0].Container.Image)
	}
}

func TestFromFileNotFound(t *testing.T) {
	_, err := FromFile("/nonexistent/path/file.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestFromFileInvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "invalid.yaml")
	os.WriteFile(path, []byte("{{{{invalid yaml"), 0o644)

	_, err := FromFile(path)
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

// --- Units tests (consolidated from units_test.go) ---

func TestValidateBinaryUnit(t *testing.T) {
	valid := []string{"500Ki", "1Mi", "2Gi", "1Ti", "1.5Pi", "1.5Ei", "42", "0.5"}
	for _, v := range valid {
		if err := ValidateBinaryUnit(v); err != nil {
			t.Errorf("expected valid: %q, got error: %v", v, err)
		}
	}

	invalid := []string{"Mi", "5K", "Ti", "abc", "1.5Z", "500m", "2k"}
	for _, v := range invalid {
		if err := ValidateBinaryUnit(v); err == nil {
			t.Errorf("expected invalid: %q", v)
		}
	}
}

func TestValidateDecimalUnit(t *testing.T) {
	valid := []string{"0.5", "1", "500m", "2k", "1.5M", "42"}
	for _, v := range valid {
		if err := ValidateDecimalUnit(v); err != nil {
			t.Errorf("expected valid: %q, got error: %v", v, err)
		}
	}

	invalid := []string{"abc", "K", "2e", "1.5Z", "1.5Ki", "1.5Mi"}
	for _, v := range invalid {
		if err := ValidateDecimalUnit(v); err == nil {
			t.Errorf("expected invalid: %q", v)
		}
	}
}

func TestConvertDecimalUnit(t *testing.T) {
	tests := []struct {
		input string
		want  float64
	}{
		{"500m", 0.5},
		{"2k", 2000.0},
		{"1.5M", 1500000.0},
		{"42", 42.0},
		{"1", 1.0},
		{"0.5", 0.5},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ConvertDecimalUnit(tt.input)
			if err != nil {
				t.Fatal(err)
			}
			if math.Abs(got-tt.want) > 0.001 {
				t.Errorf("got %f, want %f", got, tt.want)
			}
		})
	}
}

func TestConvertDecimalUnitInvalid(t *testing.T) {
	invalid := []string{"1.5Z", "abc", "1.5Ki", "1.5Mi"}
	for _, v := range invalid {
		_, err := ConvertDecimalUnit(v)
		if err == nil {
			t.Errorf("expected error for %q", v)
		}
	}
}

func TestConvertBinaryUnit(t *testing.T) {
	tests := []struct {
		input string
		want  float64
	}{
		{"500Ki", 512000.0},
		{"1Mi", 1048576.0},
		{"2Gi", 2147483648.0},
		{"42", 42.0},
		{"0.5", 0.5},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ConvertBinaryUnit(tt.input)
			if err != nil {
				t.Fatal(err)
			}
			if math.Abs(got-tt.want) > 0.001 {
				t.Errorf("got %f, want %f", got, tt.want)
			}
		})
	}
}

func TestConvertBinaryUnitInvalid(t *testing.T) {
	invalid := []string{"1.5Z", "abc", "500m", "2k"}
	for _, v := range invalid {
		_, err := ConvertBinaryUnit(v)
		if err == nil {
			t.Errorf("expected error for %q", v)
		}
	}
}

func TestValidateResourceRequirementsValid(t *testing.T) {
	tests := []struct {
		name string
		res  ResourceRequirements
	}{
		{"cpu only", ResourceRequirements{
			Requests: ResourceList{CPU: "500m"},
			Limits:   ResourceList{CPU: "1"},
		}},
		{"memory only", ResourceRequirements{
			Requests: ResourceList{Memory: "256Mi"},
			Limits:   ResourceList{Memory: "1Gi"},
		}},
		{"cpu and memory", ResourceRequirements{
			Requests: ResourceList{CPU: "100m", Memory: "128Mi"},
			Limits:   ResourceList{CPU: "500m", Memory: "512Mi"},
		}},
		{"equal request and limit", ResourceRequirements{
			Requests: ResourceList{CPU: "1"},
			Limits:   ResourceList{CPU: "1"},
		}},
		{"request only", ResourceRequirements{
			Requests: ResourceList{CPU: "500m"},
		}},
		{"limit only", ResourceRequirements{
			Limits: ResourceList{CPU: "1"},
		}},
		{"ephemeral storage", ResourceRequirements{
			Requests: ResourceList{EphemeralStorage: "1Gi"},
			Limits:   ResourceList{EphemeralStorage: "50Gi"},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateResourceRequirements(tt.res); err != nil {
				t.Errorf("expected valid, got: %v", err)
			}
		})
	}
}

func TestValidateResourceRequirementsInvalid(t *testing.T) {
	tests := []struct {
		name string
		res  ResourceRequirements
		msg  string
	}{
		{"cpu request > limit", ResourceRequirements{
			Requests: ResourceList{CPU: "1"},
			Limits:   ResourceList{CPU: "500m"},
		}, "request must be smaller or equal to limit"},
		{"cpu millicores request > limit", ResourceRequirements{
			Requests: ResourceList{CPU: "1000m"},
			Limits:   ResourceList{CPU: "800m"},
		}, "request must be smaller or equal to limit"},
		{"memory request > limit", ResourceRequirements{
			Requests: ResourceList{Memory: "1Gi"},
			Limits:   ResourceList{Memory: "512Mi"},
		}, "request must be smaller or equal to limit"},
		{"ephemeral request > limit", ResourceRequirements{
			Requests: ResourceList{EphemeralStorage: "100Gi"},
			Limits:   ResourceList{EphemeralStorage: "50Gi"},
		}, "request must be smaller or equal to limit"},
		{"invalid cpu format", ResourceRequirements{
			Requests: ResourceList{CPU: "500a"},
		}, "invalid"},
		{"invalid memory format", ResourceRequirements{
			Requests: ResourceList{Memory: "500m"},
		}, "invalid"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateResourceRequirements(tt.res)
			if err == nil {
				t.Fatal("expected error")
			}
			if !contains(err.Error(), tt.msg) {
				t.Errorf("error = %q, want to contain %q", err.Error(), tt.msg)
			}
		})
	}
}

// --- Expr tests (consolidated from expr_test.go) ---

func TestExprConstants(t *testing.T) {
	tests := []struct {
		name string
		ex   expr.Expr
		want string
	}{
		{"integer", expr.C(1), "1"},
		{"nil", expr.C(nil), "nil"},
		{"true", expr.C(true), "true"},
		{"false", expr.C(false), "false"},
		{"float", expr.C(3.14), "3.14"},
		{"string", expr.C("hello"), "'hello'"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.ex.String() != tt.want {
				t.Errorf("got %q, want %q", tt.ex.String(), tt.want)
			}
		})
	}
}

func TestExprTemplateFormat(t *testing.T) {
	e := expr.E("inputs.parameters.msg")
	if e.Tmpl() != "{{inputs.parameters.msg}}" {
		t.Errorf("Tmpl = %q", e.Tmpl())
	}
	if e.Eq() != "{{=inputs.parameters.msg}}" {
		t.Errorf("Eq = %q", e.Eq())
	}
}

func TestExprAttrChaining(t *testing.T) {
	e := expr.E("tasks").Attr("task-a").Attr("outputs").Attr("result")
	want := "tasks.task-a.outputs.result"
	if e.String() != want {
		t.Errorf("got %q, want %q", e.String(), want)
	}
}

func TestExprIndex(t *testing.T) {
	tests := []struct {
		name string
		ex   expr.Expr
		want string
	}{
		{"index", expr.E("test").Index(2), "test[2]"},
		{"key", expr.E("test").Key("as"), `test["as"]`},
		{"slice", expr.E("test").Slice(1, 9), "test[1:9]"},
		{"slice-from", expr.E("test").SliceFrom(1), "test[1:]"},
		{"slice-to", expr.E("test").SliceTo(9), "test[:9]"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.ex.String() != tt.want {
				t.Errorf("got %q, want %q", tt.ex.String(), tt.want)
			}
		})
	}
}

func TestExprComparisons(t *testing.T) {
	x := expr.E("x")
	y := expr.E("y")
	tests := []struct {
		name string
		ex   expr.Expr
		want string
	}{
		{"equals", x.Equals(y), "x == y"},
		{"not-equals", x.NotEquals(y), "x != y"},
		{"gt", x.GT(y), "x > y"},
		{"gte", x.GTE(y), "x >= y"},
		{"lt", x.LT(y), "x < y"},
		{"lte", x.LTE(y), "x <= y"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.ex.String() != tt.want {
				t.Errorf("got %q, want %q", tt.ex.String(), tt.want)
			}
		})
	}
}

func TestExprArithmetic(t *testing.T) {
	x := expr.E("x")
	y := expr.E("y")
	tests := []struct {
		name string
		ex   expr.Expr
		want string
	}{
		{"add", x.Add(y), "x + y"},
		{"sub", x.Sub(y), "x - y"},
		{"mul", x.Mul(y), "x * y"},
		{"div", x.Div(y), "x / y"},
		{"mod", x.Mod(y), "x % y"},
		{"pow", x.Pow(expr.C(2)), "x ** 2"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.ex.String() != tt.want {
				t.Errorf("got %q, want %q", tt.ex.String(), tt.want)
			}
		})
	}
}

func TestExprComplex(t *testing.T) {
	// x**2 + y
	e := expr.E("x").Pow(expr.C(2)).Add(expr.E("y"))
	if e.String() != "x ** 2 + y" {
		t.Errorf("got %q", e.String())
	}
}

func TestExprUnary(t *testing.T) {
	tests := []struct {
		name string
		ex   expr.Expr
		want string
	}{
		{"neg", expr.E("y").Neg(), "-y"},
		{"not", expr.E("y").Not(), "!y"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.ex.String() != tt.want {
				t.Errorf("got %q, want %q", tt.ex.String(), tt.want)
			}
		})
	}
}

func TestExprLogical(t *testing.T) {
	a := expr.E("a")
	b := expr.E("b")
	tests := []struct {
		name string
		ex   expr.Expr
		want string
	}{
		{"and", a.And(b), "a && b"},
		{"or", a.OrExpr(b), "a || b"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.ex.String() != tt.want {
				t.Errorf("got %q, want %q", tt.ex.String(), tt.want)
			}
		})
	}
}

func TestExprStringMethods(t *testing.T) {
	e := expr.E("test")
	tests := []struct {
		name string
		ex   expr.Expr
		want string
	}{
		{"contains", e.Contains("hello"), "test.contains('hello')"},
		{"matches", e.Matches("^a.*"), "test.matches('^a.*')"},
		{"startsWith", e.StartsWith("pre"), "test.startsWith('pre')"},
		{"endsWith", e.EndsWith("suf"), "test.endsWith('suf')"},
		{"length", e.Length(), "test.length()"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.ex.String() != tt.want {
				t.Errorf("got %q, want %q", tt.ex.String(), tt.want)
			}
		})
	}
}

func TestExprConversions(t *testing.T) {
	e := expr.E("value")
	tests := []struct {
		name string
		ex   expr.Expr
		want string
	}{
		{"toJson", e.ToJSON(), "value.toJson()"},
		{"asFloat", e.AsFloat(), "value.asFloat()"},
		{"asInt", e.AsInt(), "value.asInt()"},
		{"string", e.AsStr(), "value.string()"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.ex.String() != tt.want {
				t.Errorf("got %q, want %q", tt.ex.String(), tt.want)
			}
		})
	}
}

func TestExprTernary(t *testing.T) {
	e := expr.E("test").Check(expr.E("test1"), expr.E("test2"))
	want := "test ? test1 : test2"
	if e.String() != want {
		t.Errorf("got %q, want %q", e.String(), want)
	}
}

func TestExprCollections(t *testing.T) {
	tests := []struct {
		name string
		ex   expr.Expr
		want string
	}{
		{"map", expr.E("list").Map(expr.E("x, x * 2")), "list.map(x, x * 2)"},
		{"filter", expr.E("list").Filter(expr.E("x, x > 0")), "list.filter(x, x > 0)"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.ex.String() != tt.want {
				t.Errorf("got %q, want %q", tt.ex.String(), tt.want)
			}
		})
	}
}

func TestSprigFunctions(t *testing.T) {
	tests := []struct {
		name string
		ex   expr.Expr
		want string
	}{
		{"trim", expr.Sprig.Trim("c"), "sprig.trim('c')"},
		{"upper", expr.Sprig.Upper("hello"), "sprig.upper('hello')"},
		{"lower", expr.Sprig.Lower("HELLO"), "sprig.lower('HELLO')"},
		{"replace", expr.Sprig.Replace("old", "new", "text"), "sprig.replace('old', 'new', 'text')"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.ex.String() != tt.want {
				t.Errorf("got %q, want %q", tt.ex.String(), tt.want)
			}
		})
	}
}

func TestParamRefHelpers(t *testing.T) {
	tests := []struct {
		name string
		got  string
		want string
	}{
		{"InputParam", expr.InputParam("msg"), "{{inputs.parameters.msg}}"},
		{"TaskOutputParam", expr.TaskOutputParam("task-a", "result"), "{{tasks.task-a.outputs.parameters.result}}"},
		{"StepOutputParam", expr.StepOutputParam("step-1", "output"), "{{steps.step-1.outputs.parameters.output}}"},
		{"TaskOutputResult", expr.TaskOutputResult("task-a"), "{{tasks.task-a.outputs.result}}"},
		{"StepOutputResult", expr.StepOutputResult("step-1"), "{{steps.step-1.outputs.result}}"},
		{"WorkflowParam", expr.WorkflowParam("env"), "{{workflow.parameters.env}}"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("got %q, want %q", tt.got, tt.want)
			}
		})
	}
}

func TestExprHelperFunctions(t *testing.T) {
	tests := []struct {
		name string
		ex   expr.Expr
		want string
	}{
		{"tasks", expr.Tasks("my-task").Attr("outputs").Attr("result"), "tasks.my-task.outputs.result"},
		{"steps", expr.Steps("my-step").Attr("outputs").Attr("result"), "steps.my-step.outputs.result"},
		{"inputs", expr.Inputs().Attr("parameters").Attr("msg"), "inputs.parameters.msg"},
		{"outputs", expr.Outputs().Attr("parameters").Attr("result"), "outputs.parameters.result"},
		{"item", expr.Item(), "item"},
		{"item-attr", expr.Item().Attr("name"), "item.name"},
		{"workflow", expr.Workflow().Attr("name"), "workflow.name"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.ex.String() != tt.want {
				t.Errorf("got %q, want %q", tt.ex.String(), tt.want)
			}
		})
	}
}

func TestExprConcat(t *testing.T) {
	result := expr.Concat(" + ", expr.E("a"), expr.E("b"), expr.E("c"))
	if result.String() != "a + b + c" {
		t.Errorf("got %q", result.String())
	}
}
