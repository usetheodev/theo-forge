package expr

import (
	"strings"
	"testing"
)

// Phase 7 (T7.1) coverage backfill for expr package.

func TestC_AllTypes(t *testing.T) {
	tests := []struct {
		name string
		in   interface{}
		want string
	}{
		{"nil", nil, "nil"},
		{"true", true, "true"},
		{"false", false, "false"},
		{"int", 42, "42"},
		{"int64", int64(1234567890), "1234567890"},
		{"float64", 3.14, "3.14"},
		{"string", "x", "'x'"},
		{"default", struct{ A int }{1}, "{1}"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := C(tt.in).String()
			if got != tt.want {
				t.Errorf("C(%v) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestE_AndAccessors(t *testing.T) {
	e := E("foo").Attr("bar").Index(2).Key("k").Slice(0, 3).SliceFrom(1).SliceTo(2)
	if e.String() == "" {
		t.Fatal("chained accessors returned empty")
	}
}

func TestTmpl_AndRender(t *testing.T) {
	e := E("workflow.name")
	if e.Tmpl() != "{{workflow.name}}" {
		t.Errorf("Tmpl = %q", e.Tmpl())
	}
	if e.Render() != "{{=workflow.name}}" {
		t.Errorf("Render = %q", e.Render())
	}
	if e.Eq() != e.Render() {
		t.Errorf("Eq legacy alias broken")
	}
}

func TestComparisonOperators(t *testing.T) {
	a, b := E("a"), E("b")
	cases := map[string]string{
		a.Equals(b).String():    "a == b",
		a.NotEquals(b).String(): "a != b",
		a.GT(b).String():        "a > b",
		a.GTE(b).String():       "a >= b",
		a.LT(b).String():        "a < b",
		a.LTE(b).String():       "a <= b",
	}
	for got, want := range cases {
		if got != want {
			t.Errorf("got %q want %q", got, want)
		}
	}
}

func TestArithmeticOperators(t *testing.T) {
	a, b := E("a"), E("b")
	for got, want := range map[string]string{
		a.Add(b).String(): "a + b",
		a.Sub(b).String(): "a - b",
		a.Mul(b).String(): "a * b",
		a.Div(b).String(): "a / b",
		a.Mod(b).String(): "a % b",
		a.Pow(b).String(): "a ** b",
	} {
		if got != want {
			t.Errorf("got %q want %q", got, want)
		}
	}
}

func TestUnaryAndLogical(t *testing.T) {
	a, b := E("a"), E("b")
	if a.Neg().String() != "-a" {
		t.Errorf("Neg: %q", a.Neg().String())
	}
	if a.Not().String() != "!a" {
		t.Errorf("Not: %q", a.Not().String())
	}
	if a.And(b).String() != "a && b" {
		t.Errorf("And: %q", a.And(b).String())
	}
	if a.OrExpr(b).String() != "a || b" {
		t.Errorf("OrExpr: %q", a.OrExpr(b).String())
	}
}

func TestStringMethods(t *testing.T) {
	a := E("name")
	if !strings.Contains(a.Length().String(), ".length()") {
		t.Errorf("Length: %q", a.Length().String())
	}
	if !strings.Contains(a.ToJSON().String(), ".toJson()") {
		t.Errorf("ToJSON: %q", a.ToJSON().String())
	}
	if !strings.Contains(a.AsFloat().String(), ".asFloat()") {
		t.Errorf("AsFloat: %q", a.AsFloat().String())
	}
	if !strings.Contains(a.AsInt().String(), ".asInt()") {
		t.Errorf("AsInt: %q", a.AsInt().String())
	}
	if !strings.Contains(a.AsStr().String(), ".string()") {
		t.Errorf("AsStr: %q", a.AsStr().String())
	}
}

func TestCheck_Ternary(t *testing.T) {
	got := E("a").Check(C("yes"), C("no")).String()
	if got != "a ? 'yes' : 'no'" {
		t.Errorf("Check: %q", got)
	}
}

func TestMapFilter(t *testing.T) {
	got := E("xs").Map(E("# + 1")).String()
	if !strings.Contains(got, ".map(") {
		t.Errorf("Map: %q", got)
	}
	got = E("xs").Filter(E("# > 1")).String()
	if !strings.Contains(got, ".filter(") {
		t.Errorf("Filter: %q", got)
	}
}

func TestPackageRoots(t *testing.T) {
	if Tasks("x").String() != "tasks.x" {
		t.Errorf("Tasks: %q", Tasks("x").String())
	}
	if Steps("x").String() != "steps.x" {
		t.Errorf("Steps: %q", Steps("x").String())
	}
	if Inputs().String() != "inputs" {
		t.Errorf("Inputs: %q", Inputs().String())
	}
	if Outputs().String() != "outputs" {
		t.Errorf("Outputs: %q", Outputs().String())
	}
	if Item().String() != "item" {
		t.Errorf("Item: %q", Item().String())
	}
	if Workflow().String() != "workflow" {
		t.Errorf("Workflow: %q", Workflow().String())
	}
}

func TestRefBuilders(t *testing.T) {
	cases := []struct {
		got, want string
	}{
		{ParamRef("x.y"), "{{x.y}}"},
		{InputParam("p"), "{{inputs.parameters.p}}"},
		{TaskOutputParam("t", "p"), "{{tasks.t.outputs.parameters.p}}"},
		{StepOutputParam("s", "p"), "{{steps.s.outputs.parameters.p}}"},
		{TaskOutputResult("t"), "{{tasks.t.outputs.result}}"},
		{StepOutputResult("s"), "{{steps.s.outputs.result}}"},
		{WorkflowParam("p"), "{{workflow.parameters.p}}"},
	}
	for _, c := range cases {
		if c.got != c.want {
			t.Errorf("got %q want %q", c.got, c.want)
		}
	}
}

func TestConcat(t *testing.T) {
	got := Concat(", ", E("a"), E("b"), E("c")).String()
	if got != "a, b, c" {
		t.Errorf("Concat: %q", got)
	}
}
