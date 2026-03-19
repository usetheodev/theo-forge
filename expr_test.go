package forge

import (
	"testing"

	"github.com/usetheo/theo/forge/expr"
)

func TestExprConstants(t *testing.T) {
	tests := []struct {
		name string
		ex expr.Expr
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
		ex expr.Expr
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
		ex expr.Expr
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
		ex expr.Expr
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
	expr := expr.E("x").Pow(expr.C(2)).Add(expr.E("y"))
	if expr.String() != "x ** 2 + y" {
		t.Errorf("got %q", expr.String())
	}
}

func TestExprUnary(t *testing.T) {
	tests := []struct {
		name string
		ex expr.Expr
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
		ex expr.Expr
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
		ex expr.Expr
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
		ex expr.Expr
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
		ex expr.Expr
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
		ex expr.Expr
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
		ex expr.Expr
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
