package forge

import (
	"context"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/usetheodev/theo-forge/client"
	"github.com/usetheodev/theo-forge/expr"
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
	_ = os.WriteFile(path, []byte("{{{{invalid yaml"), 0o644)

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

// --- Coverage tests (consolidated from coverage_test.go) ---

// Cover ParamRef
func TestParamRef(t *testing.T) {
	got := expr.ParamRef("inputs.parameters.msg")
	want := "{{inputs.parameters.msg}}"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// Cover APIError.Error
func TestAPIErrorString(t *testing.T) {
	e := &client.APIError{StatusCode: 404, Message: "not found"}
	if !strings.Contains(e.Error(), "404") {
		t.Errorf("error = %q", e.Error())
	}
	if !strings.Contains(e.Error(), "not found") {
		t.Errorf("error = %q", e.Error())
	}
}

// Cover GetVersion
func TestServiceGetVersion(t *testing.T) {
	svc := &client.WorkflowsService{
		Host: "https://argo.example.com",
		HTTPClient: &mockHTTPClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				if req.URL.Path != "/api/v1/version" {
					t.Errorf("path = %q", req.URL.Path)
				}
				return mockResponse(200, map[string]interface{}{"version": "v3.5.0"}), nil
			},
		},
	}
	v, err := svc.GetVersion(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if v["version"] != "v3.5.0" {
		t.Errorf("version = %v", v["version"])
	}
}

// Cover FromJSON
func TestFromJSONCoverage(t *testing.T) {
	w := &Workflow{
		Name:       "json-roundtrip",
		Entrypoint: "main",
		Templates:  []Templatable{&Container{Name: "main", Image: "alpine"}},
	}
	jsonStr, err := w.ToJSON()
	if err != nil {
		t.Fatal(err)
	}
	model, err := FromJSON(jsonStr)
	if err != nil {
		t.Fatal(err)
	}
	if model.Metadata.Name != "json-roundtrip" {
		t.Errorf("name = %q", model.Metadata.Name)
	}
}

func TestFromJSONInvalid(t *testing.T) {
	_, err := FromJSON("{invalid json")
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

// Cover PVCVolume.BuildVolume
func TestPVCVolumeBuildVolume(t *testing.T) {
	v := PVCVolume{
		BaseVolume:       BaseVolume{Name: "data", MountPath: "/data"},
		Size:             "10Gi",
		StorageClassName: "standard",
		AccessModes:      []AccessMode{ReadWriteOnce},
	}
	vol, err := v.BuildVolume()
	if err != nil {
		t.Fatal(err)
	}
	if vol.Name != "data" {
		t.Errorf("name = %q", vol.Name)
	}
	if vol.PersistentVolumeClaim == nil {
		t.Fatal("expected PVC ref")
	}
	if vol.PersistentVolumeClaim.ClaimName != "data" {
		t.Errorf("claimName = %q", vol.PersistentVolumeClaim.ClaimName)
	}
}

// Cover Expr.C with int64 and float64 branches
func TestExprConstantInt64(t *testing.T) {
	e := expr.C(int64(42))
	if e.String() != "42" {
		t.Errorf("got %q", e.String())
	}
}

func TestExprConstantFloat64(t *testing.T) {
	e := expr.C(float64(3.14))
	if e.String() != "3.14" {
		t.Errorf("got %q", e.String())
	}
}

// Cover WorkflowTemplate name-too-long validation
func TestWorkflowTemplateNameTooLong(t *testing.T) {
	wt := &WorkflowTemplate{
		Name:       strings.Repeat("a", NameLimit+1),
		Entrypoint: "main",
	}
	_, err := wt.Build()
	if err == nil {
		t.Fatal("expected error for name too long")
	}
}

// Cover HTTPTemplate with inputs and outputs
func TestHTTPTemplateWithInputsOutputs(t *testing.T) {
	h := &HTTPTemplate{
		Name:   "with-io",
		URL:    "https://example.com/api",
		Method: "POST",
		Inputs: []Parameter{{Name: "payload", Value: ptrStr("{}")}},
		Outputs: []Parameter{{
			Name:      "status",
			ValueFrom: &ValueFrom{Expression: "response.statusCode"},
		}},
	}
	tpl, err := h.BuildTemplate()
	if err != nil {
		t.Fatal(err)
	}
	if tpl.Inputs == nil || len(tpl.Inputs.Parameters) != 1 {
		t.Fatal("expected 1 input")
	}
	if tpl.Outputs == nil || len(tpl.Outputs.Parameters) != 1 {
		t.Fatal("expected 1 output")
	}
}

// Cover GlobalConfig.GetImage empty fallback
func TestGlobalConfigGetImageFallback(t *testing.T) {
	cfg := GetGlobalConfig()
	defer cfg.Reset()
	cfg.Image = ""
	if cfg.GetImage() != "python:3.11" {
		t.Errorf("fallback = %q", cfg.GetImage())
	}
}

// Cover ContainerSet BuildTemplate output path
func TestContainerSetWithOutputs(t *testing.T) {
	cs := &ContainerSet{
		Name: "with-out",
		Containers: []ContainerNode{
			{Name: "main", Image: "alpine"},
		},
		Outputs: []Parameter{{Name: "result", ValueFrom: &ValueFrom{Path: "/tmp/out"}}},
	}
	tpl, err := cs.BuildTemplate()
	if err != nil {
		t.Fatal(err)
	}
	if tpl.Outputs == nil || len(tpl.Outputs.Parameters) != 1 {
		t.Fatal("expected 1 output")
	}
}

// Cover ContainerSet with retry
func TestContainerSetWithRetry(t *testing.T) {
	limit := 2
	cs := &ContainerSet{
		Name: "with-retry",
		Containers: []ContainerNode{
			{Name: "main", Image: "alpine"},
		},
		RetryStrategy: &RetryStrategy{Limit: &limit},
	}
	tpl, err := cs.BuildTemplate()
	if err != nil {
		t.Fatal(err)
	}
	if tpl.RetryStrategy == nil {
		t.Fatal("expected retry strategy")
	}
}

// --- Coverage 90% tests (consolidated from coverage_90_test.go) ---

// Cover ToYAML/ToJSON/ToDict error paths (invalid workflow -> Build fails -> propagates)
func TestWorkflowToYAMLBuildError(t *testing.T) {
	w := &Workflow{Entrypoint: "main"} // no name
	_, err := w.ToYAML()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestWorkflowToJSONBuildError(t *testing.T) {
	w := &Workflow{Entrypoint: "main"}
	_, err := w.ToJSON()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestWorkflowToDictBuildError(t *testing.T) {
	w := &Workflow{Entrypoint: "main"}
	_, err := w.ToDict()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestFromYAMLInvalid(t *testing.T) {
	_, err := FromYAML("{{{{bad yaml")
	if err == nil {
		t.Fatal("expected error")
	}
}

// Cover WorkflowTemplate ToYAML error path
func TestWorkflowTemplateToYAMLBuildError(t *testing.T) {
	wt := &WorkflowTemplate{} // no name
	_, err := wt.ToYAML()
	if err == nil {
		t.Fatal("expected error")
	}
}

// Cover ClusterWorkflowTemplate ToYAML error path
func TestClusterWorkflowTemplateToYAMLBuildError(t *testing.T) {
	cwt := &ClusterWorkflowTemplate{} // no name
	_, err := cwt.ToYAML()
	if err == nil {
		t.Fatal("expected error")
	}
}

// Cover CronWorkflow ToYAML/ToJSON error paths
func TestCronWorkflowToYAMLBuildError(t *testing.T) {
	cw := &CronWorkflow{} // no name, no schedule
	_, err := cw.ToYAML()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCronWorkflowToJSONBuildError(t *testing.T) {
	cw := &CronWorkflow{}
	_, err := cw.ToJSON()
	if err == nil {
		t.Fatal("expected error")
	}
}

// Cover Workflow.Build with failing template
func TestWorkflowBuildWithFailingTemplate(t *testing.T) {
	w := &Workflow{
		Name:       "test",
		Entrypoint: "main",
		Templates: []Templatable{
			&Container{Name: "", Image: "alpine"}, // no name -> fails
		},
	}
	_, err := w.Build()
	if err == nil {
		t.Fatal("expected error from failing template")
	}
}

// Cover WorkflowTemplate.Build with failing template
func TestWorkflowTemplateBuildWithFailingTemplate(t *testing.T) {
	wt := &WorkflowTemplate{
		Name:       "test",
		Entrypoint: "main",
		Templates: []Templatable{
			&Script{Name: "", Image: "alpine", Command: []string{"sh"}, Source: "echo"}, // no name
		},
	}
	_, err := wt.Build()
	if err == nil {
		t.Fatal("expected error")
	}
}

// Cover ClusterWorkflowTemplate.Build with failing template
func TestClusterWorkflowTemplateBuildWithFailingTemplate(t *testing.T) {
	cwt := &ClusterWorkflowTemplate{
		Name:       "test",
		Entrypoint: "main",
		Templates: []Templatable{
			&Container{Name: "", Image: "alpine"},
		},
	}
	_, err := cwt.Build()
	if err == nil {
		t.Fatal("expected error")
	}
}

// Cover CronWorkflow.Build with failing template
func TestCronWorkflowBuildWithFailingTemplate(t *testing.T) {
	cw := &CronWorkflow{
		Name:       "test",
		Schedule:   "0 * * * *",
		Entrypoint: "main",
		Templates: []Templatable{
			&Container{Name: "", Image: "alpine"},
		},
	}
	_, err := cw.Build()
	if err == nil {
		t.Fatal("expected error")
	}
}

// Cover WorkflowTemplate.Build with volumes
func TestWorkflowTemplateBuildWithVolumes(t *testing.T) {
	wt := &WorkflowTemplate{
		Name:       "with-vols",
		Entrypoint: "main",
		Volumes: []VolumeBuilder{
			&EmptyDirVolume{BaseVolume: BaseVolume{Name: "tmp", MountPath: "/tmp"}},
		},
		Templates: []Templatable{&Container{Name: "main", Image: "alpine"}},
	}
	m, err := wt.Build()
	if err != nil {
		t.Fatal(err)
	}
	if len(m.Spec.Volumes) != 1 {
		t.Fatalf("volumes = %d", len(m.Spec.Volumes))
	}
}

// Cover CronWorkflow.Build with volumes and arguments
func TestCronWorkflowBuildWithVolsAndArgs(t *testing.T) {
	cw := &CronWorkflow{
		Name:       "full",
		Schedule:   "0 * * * *",
		Entrypoint: "main",
		Arguments:  []Parameter{{Name: "env", Value: ptrStr("prod")}},
		Volumes: []VolumeBuilder{
			&EmptyDirVolume{BaseVolume: BaseVolume{Name: "tmp", MountPath: "/tmp"}},
		},
		Templates: []Templatable{&Container{Name: "main", Image: "alpine"}},
	}
	m, err := cw.Build()
	if err != nil {
		t.Fatal(err)
	}
	if m.Spec.WorkflowSpec.Arguments == nil {
		t.Fatal("expected arguments")
	}
	if len(m.Spec.WorkflowSpec.Volumes) != 1 {
		t.Fatal("expected 1 volume")
	}
}

// Cover Workflow.ToFile error paths
func TestWorkflowToFileBuildError(t *testing.T) {
	w := &Workflow{Entrypoint: "main"} // no name
	_, err := w.ToFile(t.TempDir(), "")
	if err == nil {
		t.Fatal("expected error")
	}
}

// Cover DAG.buildOutputs with output parameters
func TestDAGWithOutputParameters(t *testing.T) {
	dag := &DAG{
		Name: "with-out-params",
		Outputs: []Parameter{
			{Name: "result", ValueFrom: &ValueFrom{Expression: "tasks.final.outputs.result"}},
		},
	}
	tpl, err := dag.BuildTemplate()
	if err != nil {
		t.Fatal(err)
	}
	if tpl.Outputs == nil || len(tpl.Outputs.Parameters) != 1 {
		t.Fatal("expected 1 output parameter")
	}
}

// Cover LintWorkflow build error path
func TestServiceLintWorkflowBuildError(t *testing.T) {
	svc := &client.WorkflowsService{Host: "https://argo.example.com", Namespace: "default"}
	w := &Workflow{Entrypoint: "main"} // no name
	_, err := svc.LintWorkflow(context.TODO(), w)
	if err == nil {
		t.Fatal("expected build error")
	}
}

// Cover CreateWorkflow build error path
func TestServiceCreateWorkflowBuildError(t *testing.T) {
	svc := &client.WorkflowsService{Host: "https://argo.example.com", Namespace: "default"}
	w := &Workflow{Entrypoint: "main"} // no name
	_, err := svc.CreateWorkflow(context.TODO(), w)
	if err == nil {
		t.Fatal("expected build error")
	}
}

// Cover AddParallelGroup with name conflict within group
func TestStepsAddParallelGroupInternalConflict(t *testing.T) {
	steps := &Steps{Name: "test"}
	err := steps.AddParallelGroup(
		&Step{Name: "dup", Template: "a"},
		&Step{Name: "dup", Template: "b"},
	)
	if err == nil {
		t.Fatal("expected name conflict within parallel group")
	}
}

// Cover S3/GCS/HTTP/Git/Raw artifact error paths
func TestS3ArtifactNoNameFails(t *testing.T) {
	a := S3Artifact{Artifact: Artifact{Path: "/tmp"}, Bucket: "b", Key: "k"}
	_, err := a.Build()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGCSArtifactNoNameFails(t *testing.T) {
	a := GCSArtifact{Artifact: Artifact{Path: "/tmp"}, Bucket: "b", Key: "k"}
	_, err := a.Build()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestGitArtifactNoNameFails(t *testing.T) {
	a := GitArtifact{Artifact: Artifact{Path: "/tmp"}, Repo: "r"}
	_, err := a.Build()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestRawArtifactNoNameFails(t *testing.T) {
	a := RawArtifact{Artifact: Artifact{Path: "/tmp"}, Data: "d"}
	_, err := a.Build()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestHTTPArtifactNoNameFails(t *testing.T) {
	a := HTTPArtifact{Artifact: Artifact{Path: "/tmp"}, URL: "u"}
	_, err := a.Build()
	if err == nil {
		t.Fatal("expected error")
	}
}

// Cover Expr.C default branch
func TestExprConstantDefault(t *testing.T) {
	type custom struct{ X int }
	e := expr.C(custom{X: 42})
	if !strings.Contains(e.String(), "42") {
		t.Errorf("got %q", e.String())
	}
}

// --- Coverage 95% tests (consolidated from coverage_95_test.go) ---

// Cover ClusterWorkflowTemplate.Build with arguments
func TestClusterWorkflowTemplateBuildWithArguments(t *testing.T) {
	cwt := &ClusterWorkflowTemplate{
		Name:       "with-args",
		Entrypoint: "main",
		Arguments:  []Parameter{{Name: "env", Value: ptrStr("prod")}},
		Templates:  []Templatable{&Container{Name: "main", Image: "alpine"}},
	}
	m, err := cwt.Build()
	if err != nil {
		t.Fatal(err)
	}
	if m.Spec.Arguments == nil || len(m.Spec.Arguments.Parameters) != 1 {
		t.Fatal("expected 1 argument")
	}
}

// Cover CronWorkflow.Build with arguments
func TestCronWorkflowBuildWithArguments(t *testing.T) {
	cw := &CronWorkflow{
		Name:       "with-args",
		Schedule:   "0 * * * *",
		Entrypoint: "main",
		Arguments:  []Parameter{{Name: "env", Value: ptrStr("staging")}},
		Templates:  []Templatable{&Container{Name: "main", Image: "alpine"}},
	}
	m, err := cw.Build()
	if err != nil {
		t.Fatal(err)
	}
	if m.Spec.WorkflowSpec.Arguments == nil {
		t.Fatal("expected arguments")
	}
}

// Cover BuildDAGTask artifact argument error path
func TestTaskBuildDAGTaskArtifactError(t *testing.T) {
	task := &Task{
		Name:     "test",
		Template: "tpl",
		ArgumentArtifacts: []ArtifactBuilder{
			&Artifact{Path: "/tmp"}, // no name -> error
		},
	}
	_, err := task.BuildDAGTask()
	if err == nil {
		t.Fatal("expected error from artifact build")
	}
}

// Cover BuildStep artifact argument error path
func TestStepBuildStepArtifactError(t *testing.T) {
	s := &Step{
		Name:     "test",
		Template: "tpl",
		ArgumentArtifacts: []ArtifactBuilder{
			&Artifact{Path: "/tmp"}, // no name -> error
		},
	}
	_, err := s.BuildStep()
	if err == nil {
		t.Fatal("expected error from artifact build")
	}
}

// Cover BuildArguments artifact error path
func TestBuildArgumentsArtifactError(t *testing.T) {
	_, err := BuildArguments(
		nil,
		[]ArtifactBuilder{&Artifact{Path: "/tmp"}}, // no name
	)
	if err == nil {
		t.Fatal("expected error")
	}
}

// Cover BuildArguments parameter error path
func TestBuildArgumentsParameterError(t *testing.T) {
	_, err := BuildArguments(
		[]Parameter{{Value: ptrStr("val")}}, // no name
		nil,
	)
	if err == nil {
		t.Fatal("expected error")
	}
}

// Cover DAG.BuildTemplate with task build error
func TestDAGBuildTemplateWithTaskError(t *testing.T) {
	dag := &DAG{
		Name: "test",
		Tasks: []*Task{
			{Name: "", Template: "tpl"}, // no name
		},
	}
	_, err := dag.BuildTemplate()
	if err == nil {
		t.Fatal("expected error from task build")
	}
}

// Cover Steps.BuildTemplate with step group error
func TestStepsBuildTemplateWithStepError(t *testing.T) {
	steps := &Steps{Name: "test"}
	steps.StepGroups = append(steps.StepGroups, Parallel{
		Steps: []*Step{{Name: "", Template: "tpl"}}, // no name
	})
	_, err := steps.BuildTemplate()
	if err == nil {
		t.Fatal("expected error from step build")
	}
}

// Cover Workflow.ToFile with .yml extension
func TestWorkflowToFileYmlExtension(t *testing.T) {
	w := &Workflow{
		Name:       "test",
		Entrypoint: "main",
		Templates:  []Templatable{&Container{Name: "main", Image: "alpine"}},
	}
	path, err := w.ToFile(t.TempDir(), "output.yml")
	if err != nil {
		t.Fatal(err)
	}
	if path[len(path)-4:] != ".yml" {
		t.Errorf("path should end with .yml, got %q", path)
	}
}

func TestConvertBinaryUnitLargeSuffix(t *testing.T) {
	v, err := ConvertBinaryUnit("1Ti")
	if err != nil {
		t.Fatal(err)
	}
	if v != 1099511627776 { // 2^40
		t.Errorf("got %f", v)
	}
}

func TestConvertDecimalUnitLargeSuffix(t *testing.T) {
	v, err := ConvertDecimalUnit("1G")
	if err != nil {
		t.Fatal(err)
	}
	if v != 1e9 {
		t.Errorf("got %f", v)
	}
}

// Cover ConfigMapVolume with no-name edge
func TestConfigMapVolumeNoNameBuildsError(t *testing.T) {
	v := ConfigMapVolume{BaseVolume: BaseVolume{MountPath: "/cfg"}}
	_, err := v.BuildVolume()
	if err == nil {
		t.Fatal("expected error")
	}
}

// Cover Workflow.buildVolumes error path (volume with no name)
func TestWorkflowBuildVolumesWithError(t *testing.T) {
	w := &Workflow{
		Name:       "test",
		Entrypoint: "main",
		Volumes: []VolumeBuilder{
			&EmptyDirVolume{BaseVolume: BaseVolume{Name: "good", MountPath: "/tmp"}},
			&EmptyDirVolume{BaseVolume: BaseVolume{MountPath: "/bad"}}, // no name
		},
		Templates: []Templatable{&Container{Name: "main", Image: "alpine"}},
	}
	_, err := w.Build()
	if err == nil {
		t.Fatal("expected error for volume with no name")
	}
}

// Cover Workflow.buildVolumeClaimTemplates error propagation
func TestWorkflowBuildVolumeClaimTemplatesWithError(t *testing.T) {
	w := &Workflow{
		Name:       "test",
		Entrypoint: "main",
		VolumeClaimTemplates: []PVCVolume{
			{BaseVolume: BaseVolume{Name: "good", MountPath: "/data"}, Size: "1Gi"},
			{BaseVolume: BaseVolume{MountPath: "/bad"}, Size: "1Gi"}, // no name
		},
		Templates: []Templatable{&Container{Name: "main", Image: "alpine"}},
	}
	_, err := w.Build()
	if err == nil {
		t.Fatal("expected error for PVC with no name")
	}
}

// --- Coverage extra tests (consolidated from coverage_extra_test.go) ---

// Cover Workflow.buildVolumeClaimTemplates
func TestWorkflowWithVolumeClaimTemplates(t *testing.T) {
	w := &Workflow{
		Name:       "with-pvcs",
		Entrypoint: "main",
		VolumeClaimTemplates: []PVCVolume{
			{
				BaseVolume:       BaseVolume{Name: "work", MountPath: "/work"},
				Size:             "5Gi",
				StorageClassName: "fast",
				AccessModes:      []AccessMode{ReadWriteOnce},
			},
			{
				BaseVolume:  BaseVolume{Name: "cache", MountPath: "/cache"},
				Size:        "10Gi",
				AccessModes: []AccessMode{ReadWriteMany},
			},
		},
		Templates: []Templatable{
			&Container{
				Name:  "main",
				Image: "alpine",
				VolumeMounts: []VolumeBuilder{
					&PVCVolume{BaseVolume: BaseVolume{Name: "work", MountPath: "/work"}},
				},
			},
		},
	}
	m, err := w.Build()
	if err != nil {
		t.Fatal(err)
	}
	if len(m.Spec.VolumeClaimTemplates) != 2 {
		t.Fatalf("pvcs = %d, want 2", len(m.Spec.VolumeClaimTemplates))
	}
	if m.Spec.VolumeClaimTemplates[0].Metadata.Name != "work" {
		t.Errorf("pvc[0].name = %q", m.Spec.VolumeClaimTemplates[0].Metadata.Name)
	}
	if m.Spec.VolumeClaimTemplates[0].Spec.StorageClassName != "fast" {
		t.Errorf("storageClass = %q", m.Spec.VolumeClaimTemplates[0].Spec.StorageClassName)
	}
	if m.Spec.VolumeClaimTemplates[1].Spec.Resources.Requests.Storage != "10Gi" {
		t.Errorf("storage = %q", m.Spec.VolumeClaimTemplates[1].Spec.Resources.Requests.Storage)
	}
}

// Cover Workflow.buildArguments with artifacts
func TestWorkflowWithArgumentArtifacts(t *testing.T) {
	w := &Workflow{
		Name:       "with-art-args",
		Entrypoint: "main",
		ArgumentArtifacts: []ArtifactBuilder{
			&S3Artifact{
				Artifact: Artifact{Name: "input-data", Path: "/data"},
				Bucket:   "my-bucket",
				Key:      "input.csv",
			},
		},
		Templates: []Templatable{&Container{Name: "main", Image: "alpine"}},
	}
	m, err := w.Build()
	if err != nil {
		t.Fatal(err)
	}
	if m.Spec.Arguments == nil {
		t.Fatal("expected arguments")
	}
	if len(m.Spec.Arguments.Artifacts) != 1 {
		t.Fatalf("artifact args = %d", len(m.Spec.Arguments.Artifacts))
	}
	if m.Spec.Arguments.Artifacts[0].S3 == nil {
		t.Error("expected S3 artifact arg")
	}
}

// Cover Workflow.buildMetrics
func TestWorkflowWithMetrics(t *testing.T) {
	w := &Workflow{
		Name:       "with-metrics",
		Entrypoint: "main",
		Metrics: []Metric{
			{
				Name: "workflow_duration",
				Help: "Duration of workflow",
				Gauge: &Gauge{
					Value:    "{{workflow.duration}}",
					Realtime: ptrBool(true),
				},
			},
		},
		Templates: []Templatable{&Container{Name: "main", Image: "alpine"}},
	}
	m, err := w.Build()
	if err != nil {
		t.Fatal(err)
	}
	if m.Spec.Metrics == nil {
		t.Fatal("expected metrics")
	}
	if len(m.Spec.Metrics.Prometheus) != 1 {
		t.Fatalf("metrics = %d", len(m.Spec.Metrics.Prometheus))
	}
}

// Cover Container.buildMetrics
func TestContainerWithMetrics(t *testing.T) {
	c := &Container{
		Name:  "with-metrics",
		Image: "alpine",
		Metrics: []Metric{
			{Name: "step_duration", Help: "Step duration", Counter: &Counter{Value: "1"}},
		},
	}
	tpl, err := c.BuildTemplate()
	if err != nil {
		t.Fatal(err)
	}
	if tpl.Metrics == nil || len(tpl.Metrics.Prometheus) != 1 {
		t.Fatal("expected 1 metric")
	}
}

// Cover Script.buildMetrics and buildMetadata
func TestScriptWithMetricsAndMetadata(t *testing.T) {
	s := &Script{
		Name:        "with-meta",
		Image:       "python:3.11",
		Command:     []string{"python"},
		Source:      "print('hello')",
		Labels:      map[string]string{"team": "backend"},
		Annotations: map[string]string{"note": "test"},
		Metrics: []Metric{
			{Name: "script_runs", Help: "count", Counter: &Counter{Value: "1"}},
		},
	}
	tpl, err := s.BuildTemplate()
	if err != nil {
		t.Fatal(err)
	}
	if tpl.Metadata == nil {
		t.Fatal("expected metadata")
	}
	if tpl.Metadata.Labels["team"] != "backend" {
		t.Errorf("label = %q", tpl.Metadata.Labels["team"])
	}
	if tpl.Metrics == nil {
		t.Fatal("expected metrics")
	}
}

// Cover Script.buildVolumeMounts
func TestScriptWithVolumeMounts(t *testing.T) {
	s := &Script{
		Name:    "with-mounts",
		Image:   "python:3.11",
		Command: []string{"python"},
		Source:  "print('hello')",
		VolumeMounts: []VolumeBuilder{
			&EmptyDirVolume{BaseVolume: BaseVolume{Name: "tmp", MountPath: "/tmp/work"}},
		},
	}
	tpl, err := s.BuildTemplate()
	if err != nil {
		t.Fatal(err)
	}
	if len(tpl.Script.VolumeMounts) != 1 {
		t.Fatalf("mounts = %d", len(tpl.Script.VolumeMounts))
	}
}

// Cover DAG.buildOutputs with artifacts
func TestDAGWithInputAndOutputArtifacts(t *testing.T) {
	dag := &DAG{
		Name: "with-art-io",
		InputArtifacts: []ArtifactBuilder{
			&Artifact{Name: "input", Path: "/tmp/in"},
		},
		OutputArtifacts: []ArtifactBuilder{
			&Artifact{Name: "output", Path: "/tmp/out"},
		},
	}
	tpl, err := dag.BuildTemplate()
	if err != nil {
		t.Fatal(err)
	}
	if tpl.Inputs == nil || len(tpl.Inputs.Artifacts) != 1 {
		t.Fatal("expected 1 input artifact")
	}
	if tpl.Outputs == nil || len(tpl.Outputs.Artifacts) != 1 {
		t.Fatal("expected 1 output artifact")
	}
}

// Cover Steps.buildOutputs with artifacts
func TestStepsWithOutputArtifacts(t *testing.T) {
	steps := &Steps{
		Name: "with-art-out",
		OutputArtifacts: []ArtifactBuilder{
			&Artifact{Name: "result", Path: "/tmp/result"},
		},
	}
	tpl, err := steps.BuildTemplate()
	if err != nil {
		t.Fatal(err)
	}
	if tpl.Outputs == nil || len(tpl.Outputs.Artifacts) != 1 {
		t.Fatal("expected 1 output artifact")
	}
}

// Cover Step with artifact arguments
func TestStepWithArtifactArguments(t *testing.T) {
	s := &Step{
		Name:     "with-art-args",
		Template: "process",
		ArgumentArtifacts: []ArtifactBuilder{
			&Artifact{Name: "data", From: "{{steps.gen.outputs.artifacts.output}}"},
		},
	}
	m, err := s.BuildStep()
	if err != nil {
		t.Fatal(err)
	}
	if m.Arguments == nil || len(m.Arguments.Artifacts) != 1 {
		t.Fatal("expected 1 artifact argument")
	}
}

// Cover Task with artifact arguments
func TestTaskWithArtifactArguments(t *testing.T) {
	task := &Task{
		Name:     "with-art-args",
		Template: "process",
		ArgumentArtifacts: []ArtifactBuilder{
			&Artifact{Name: "data", From: "{{tasks.gen.outputs.artifacts.output}}"},
		},
	}
	m, err := task.BuildDAGTask()
	if err != nil {
		t.Fatal(err)
	}
	if m.Arguments == nil || len(m.Arguments.Artifacts) != 1 {
		t.Fatal("expected 1 artifact argument")
	}
}

// Cover PVCVolume.BuildPVC default access modes
func TestPVCVolumeDefaultAccessModes(t *testing.T) {
	v := PVCVolume{
		BaseVolume: BaseVolume{Name: "default-mode", MountPath: "/data"},
		Size:       "1Gi",
	}
	pvc, err := v.BuildPVC()
	if err != nil {
		t.Fatal(err)
	}
	if len(pvc.Spec.AccessModes) != 1 || pvc.Spec.AccessModes[0] != "ReadWriteOnce" {
		t.Errorf("default access modes = %v", pvc.Spec.AccessModes)
	}
}

// Cover Container with output parameters and artifacts
func TestContainerWithOutputs(t *testing.T) {
	c := &Container{
		Name:  "with-outputs",
		Image: "alpine",
		Outputs: []Parameter{
			{Name: "result", ValueFrom: &ValueFrom{Path: "/tmp/result"}},
		},
		OutputArtifacts: []ArtifactBuilder{
			&Artifact{Name: "logs", Path: "/tmp/logs"},
		},
	}
	tpl, err := c.BuildTemplate()
	if err != nil {
		t.Fatal(err)
	}
	if tpl.Outputs == nil {
		t.Fatal("expected outputs")
	}
	if len(tpl.Outputs.Parameters) != 1 {
		t.Errorf("output params = %d", len(tpl.Outputs.Parameters))
	}
	if len(tpl.Outputs.Artifacts) != 1 {
		t.Errorf("output artifacts = %d", len(tpl.Outputs.Artifacts))
	}
}

// Cover Workflow with retry strategy
func TestWorkflowWithRetryStrategy(t *testing.T) {
	limit := 5
	w := &Workflow{
		Name:       "with-retry",
		Entrypoint: "main",
		RetryStrategy: &RetryStrategy{
			Limit:       &limit,
			RetryPolicy: RetryAlways,
		},
		Templates: []Templatable{&Container{Name: "main", Image: "alpine"}},
	}
	m, err := w.Build()
	if err != nil {
		t.Fatal(err)
	}
	if m.Spec.RetryStrategy == nil {
		t.Fatal("expected retry strategy")
	}
	if m.Spec.RetryStrategy.Limit != "5" {
		t.Errorf("limit = %v, want \"5\"", m.Spec.RetryStrategy.Limit)
	}
}

// --- Coverage final tests (consolidated from coverage_final_test.go) ---

// Cover volume BuildVolume error paths (no-name failures)
func TestHostPathVolumeNoNameFails(t *testing.T) {
	v := HostPathVolume{Path: "/data"}
	_, err := v.BuildVolume()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestSecretVolumeNoNameFails(t *testing.T) {
	v := SecretVolume{SecretName: "s"}
	_, err := v.BuildVolume()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestExistingVolumeNoNameFails(t *testing.T) {
	v := ExistingVolume{ClaimName: "pvc"}
	_, err := v.BuildVolume()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestPVCVolumeNoNameFails(t *testing.T) {
	v := PVCVolume{Size: "1Gi"}
	_, err := v.BuildVolume()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestPVCVolumeBuildPVCNoNameFails(t *testing.T) {
	v := PVCVolume{Size: "1Gi"}
	_, err := v.BuildPVC()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestNFSVolumeNoNameFails(t *testing.T) {
	v := NFSVolume{Server: "nfs", Path: "/data"}
	_, err := v.BuildVolume()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestConfigMapVolumeNoNameFails(t *testing.T) {
	v := ConfigMapVolume{}
	_, err := v.BuildVolume()
	if err == nil {
		t.Fatal("expected error")
	}
}

// Cover Parameter.String success path
func TestParameterStringSuccess(t *testing.T) {
	p := Parameter{Name: "test", Value: ptrStr("hello")}
	s, err := p.String()
	if err != nil {
		t.Fatal(err)
	}
	if s != "hello" {
		t.Errorf("got %q", s)
	}
}

// Cover service unmarshal error paths
func TestServiceCreateWorkflowBadResponse(t *testing.T) {
	svc := &client.WorkflowsService{
		Host:      "https://argo.example.com",
		Namespace: "default",
		HTTPClient: &mockHTTPClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				return mockResponse(200, "not-json-object"), nil
			},
		},
	}
	w := &Workflow{Name: "test", Entrypoint: "main", Templates: []Templatable{&Container{Name: "main", Image: "a"}}}
	_, err := svc.CreateWorkflow(context.Background(), w)
	if err == nil {
		t.Fatal("expected unmarshal error")
	}
}

func TestServiceListWorkflowsBadResponse(t *testing.T) {
	svc := &client.WorkflowsService{
		Host:      "https://argo.example.com",
		Namespace: "default",
		HTTPClient: &mockHTTPClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				return mockResponse(200, "not-json"), nil
			},
		},
	}
	_, err := svc.ListWorkflows(context.Background(), "")
	if err == nil {
		t.Fatal("expected unmarshal error")
	}
}

func TestServiceLintWorkflowBadResponse(t *testing.T) {
	svc := &client.WorkflowsService{
		Host:      "https://argo.example.com",
		Namespace: "default",
		HTTPClient: &mockHTTPClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				return mockResponse(200, "bad"), nil
			},
		},
	}
	w := &Workflow{Name: "test", Entrypoint: "main", Templates: []Templatable{&Container{Name: "main", Image: "a"}}}
	_, err := svc.LintWorkflow(context.Background(), w)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestServiceGetInfoBadResponse(t *testing.T) {
	svc := &client.WorkflowsService{
		Host: "https://argo.example.com",
		HTTPClient: &mockHTTPClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				return mockResponse(200, "bad"), nil
			},
		},
	}
	_, err := svc.GetInfo(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestServiceGetVersionBadResponse(t *testing.T) {
	svc := &client.WorkflowsService{
		Host: "https://argo.example.com",
		HTTPClient: &mockHTTPClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				return mockResponse(200, "bad"), nil
			},
		},
	}
	_, err := svc.GetVersion(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestServiceGetWorkflowBadResponse(t *testing.T) {
	svc := &client.WorkflowsService{
		Host:      "https://argo.example.com",
		Namespace: "default",
		HTTPClient: &mockHTTPClient{
			DoFunc: func(req *http.Request) (*http.Response, error) {
				return mockResponse(200, "bad"), nil
			},
		},
	}
	_, err := svc.GetWorkflow(context.Background(), "test", "")
	if err == nil {
		t.Fatal("expected error")
	}
}

// Cover Steps.buildInputs/buildOutputs artifact paths
func TestStepsWithInputArtifacts(t *testing.T) {
	steps := &Steps{
		Name: "with-art-in",
		InputArtifacts: []ArtifactBuilder{
			&Artifact{Name: "data", Path: "/tmp/data"},
		},
	}
	tpl, err := steps.BuildTemplate()
	if err != nil {
		t.Fatal(err)
	}
	if tpl.Inputs == nil || len(tpl.Inputs.Artifacts) != 1 {
		t.Fatal("expected 1 input artifact")
	}
}

// Cover DAG.buildInputs artifact path
func TestDAGWithInputArtifacts(t *testing.T) {
	dag := &DAG{
		Name: "with-art-in",
		InputArtifacts: []ArtifactBuilder{
			&Artifact{Name: "data", Path: "/tmp/data"},
		},
	}
	tpl, err := dag.BuildTemplate()
	if err != nil {
		t.Fatal(err)
	}
	if tpl.Inputs == nil || len(tpl.Inputs.Artifacts) != 1 {
		t.Fatal("expected 1 input artifact")
	}
}

// Cover Script.buildInputs artifact path
func TestScriptWithInputArtifacts(t *testing.T) {
	s := &Script{
		Name:    "with-art-in",
		Image:   "python:3.11",
		Command: []string{"python"},
		Source:  "print('hi')",
		InputArtifacts: []ArtifactBuilder{
			&Artifact{Name: "model", Path: "/tmp/model.pkl"},
		},
	}
	tpl, err := s.BuildTemplate()
	if err != nil {
		t.Fatal(err)
	}
	if tpl.Inputs == nil || len(tpl.Inputs.Artifacts) != 1 {
		t.Fatal("expected 1 input artifact")
	}
}

// Cover ValidateResourceRequirements - limit-only validations
func TestValidateResourceLimitOnlyInvalid(t *testing.T) {
	err := ValidateResourceRequirements(ResourceRequirements{
		Limits: ResourceList{CPU: "abc"},
	})
	if err == nil {
		t.Fatal("expected error for invalid CPU limit")
	}
}

func TestValidateResourceMemoryLimitOnlyInvalid(t *testing.T) {
	err := ValidateResourceRequirements(ResourceRequirements{
		Limits: ResourceList{Memory: "500m"},
	})
	if err == nil {
		t.Fatal("expected error for invalid memory limit (decimal unit)")
	}
}

func TestValidateResourceEphemeralLimitOnlyInvalid(t *testing.T) {
	err := ValidateResourceRequirements(ResourceRequirements{
		Limits: ResourceList{EphemeralStorage: "abc"},
	})
	if err == nil {
		t.Fatal("expected error for invalid ephemeral limit")
	}
}

func TestValidateResourceEphemeralRequestInvalid(t *testing.T) {
	err := ValidateResourceRequirements(ResourceRequirements{
		Requests: ResourceList{EphemeralStorage: "abc"},
	})
	if err == nil {
		t.Fatal("expected error for invalid ephemeral request")
	}
}

func TestValidateResourceEphemeralLimitInvalid(t *testing.T) {
	err := ValidateResourceRequirements(ResourceRequirements{
		Requests: ResourceList{EphemeralStorage: "1Gi"},
		Limits:   ResourceList{EphemeralStorage: "abc"},
	})
	if err == nil {
		t.Fatal("expected error for invalid ephemeral limit with valid request")
	}
}
