package forge

import "testing"

func TestExprConstants(t *testing.T) {
	tests := []struct {
		name string
		expr Expr
		want string
	}{
		{"integer", C(1), "1"},
		{"nil", C(nil), "nil"},
		{"true", C(true), "true"},
		{"false", C(false), "false"},
		{"float", C(3.14), "3.14"},
		{"string", C("hello"), "'hello'"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expr.String() != tt.want {
				t.Errorf("got %q, want %q", tt.expr.String(), tt.want)
			}
		})
	}
}

func TestExprTemplateFormat(t *testing.T) {
	e := E("inputs.parameters.msg")
	if e.Tmpl() != "{{inputs.parameters.msg}}" {
		t.Errorf("Tmpl = %q", e.Tmpl())
	}
	if e.Eq() != "{{=inputs.parameters.msg}}" {
		t.Errorf("Eq = %q", e.Eq())
	}
}

func TestExprAttrChaining(t *testing.T) {
	e := E("tasks").Attr("task-a").Attr("outputs").Attr("result")
	want := "tasks.task-a.outputs.result"
	if e.String() != want {
		t.Errorf("got %q, want %q", e.String(), want)
	}
}

func TestExprIndex(t *testing.T) {
	tests := []struct {
		name string
		expr Expr
		want string
	}{
		{"index", E("test").Index(2), "test[2]"},
		{"key", E("test").Key("as"), `test["as"]`},
		{"slice", E("test").Slice(1, 9), "test[1:9]"},
		{"slice-from", E("test").SliceFrom(1), "test[1:]"},
		{"slice-to", E("test").SliceTo(9), "test[:9]"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expr.String() != tt.want {
				t.Errorf("got %q, want %q", tt.expr.String(), tt.want)
			}
		})
	}
}

func TestExprComparisons(t *testing.T) {
	x := E("x")
	y := E("y")
	tests := []struct {
		name string
		expr Expr
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
			if tt.expr.String() != tt.want {
				t.Errorf("got %q, want %q", tt.expr.String(), tt.want)
			}
		})
	}
}

func TestExprArithmetic(t *testing.T) {
	x := E("x")
	y := E("y")
	tests := []struct {
		name string
		expr Expr
		want string
	}{
		{"add", x.Add(y), "x + y"},
		{"sub", x.Sub(y), "x - y"},
		{"mul", x.Mul(y), "x * y"},
		{"div", x.Div(y), "x / y"},
		{"mod", x.Mod(y), "x % y"},
		{"pow", x.Pow(C(2)), "x ** 2"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expr.String() != tt.want {
				t.Errorf("got %q, want %q", tt.expr.String(), tt.want)
			}
		})
	}
}

func TestExprComplex(t *testing.T) {
	// x**2 + y
	expr := E("x").Pow(C(2)).Add(E("y"))
	if expr.String() != "x ** 2 + y" {
		t.Errorf("got %q", expr.String())
	}
}

func TestExprUnary(t *testing.T) {
	tests := []struct {
		name string
		expr Expr
		want string
	}{
		{"neg", E("y").Neg(), "-y"},
		{"not", E("y").Not(), "!y"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expr.String() != tt.want {
				t.Errorf("got %q, want %q", tt.expr.String(), tt.want)
			}
		})
	}
}

func TestExprLogical(t *testing.T) {
	a := E("a")
	b := E("b")
	tests := []struct {
		name string
		expr Expr
		want string
	}{
		{"and", a.And(b), "a && b"},
		{"or", a.OrExpr(b), "a || b"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expr.String() != tt.want {
				t.Errorf("got %q, want %q", tt.expr.String(), tt.want)
			}
		})
	}
}

func TestExprStringMethods(t *testing.T) {
	e := E("test")
	tests := []struct {
		name string
		expr Expr
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
			if tt.expr.String() != tt.want {
				t.Errorf("got %q, want %q", tt.expr.String(), tt.want)
			}
		})
	}
}

func TestExprConversions(t *testing.T) {
	e := E("value")
	tests := []struct {
		name string
		expr Expr
		want string
	}{
		{"toJson", e.ToJSON(), "value.toJson()"},
		{"asFloat", e.AsFloat(), "value.asFloat()"},
		{"asInt", e.AsInt(), "value.asInt()"},
		{"string", e.AsStr(), "value.string()"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expr.String() != tt.want {
				t.Errorf("got %q, want %q", tt.expr.String(), tt.want)
			}
		})
	}
}

func TestExprTernary(t *testing.T) {
	e := E("test").Check(E("test1"), E("test2"))
	want := "test ? test1 : test2"
	if e.String() != want {
		t.Errorf("got %q, want %q", e.String(), want)
	}
}

func TestExprCollections(t *testing.T) {
	tests := []struct {
		name string
		expr Expr
		want string
	}{
		{"map", E("list").Map(E("x, x * 2")), "list.map(x, x * 2)"},
		{"filter", E("list").Filter(E("x, x > 0")), "list.filter(x, x > 0)"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expr.String() != tt.want {
				t.Errorf("got %q, want %q", tt.expr.String(), tt.want)
			}
		})
	}
}

func TestSprigFunctions(t *testing.T) {
	tests := []struct {
		name string
		expr Expr
		want string
	}{
		{"trim", Sprig.Trim("c"), "sprig.trim('c')"},
		{"upper", Sprig.Upper("hello"), "sprig.upper('hello')"},
		{"lower", Sprig.Lower("HELLO"), "sprig.lower('HELLO')"},
		{"replace", Sprig.Replace("old", "new", "text"), "sprig.replace('old', 'new', 'text')"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expr.String() != tt.want {
				t.Errorf("got %q, want %q", tt.expr.String(), tt.want)
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
		{"InputParam", InputParam("msg"), "{{inputs.parameters.msg}}"},
		{"TaskOutputParam", TaskOutputParam("task-a", "result"), "{{tasks.task-a.outputs.parameters.result}}"},
		{"StepOutputParam", StepOutputParam("step-1", "output"), "{{steps.step-1.outputs.parameters.output}}"},
		{"TaskOutputResult", TaskOutputResult("task-a"), "{{tasks.task-a.outputs.result}}"},
		{"StepOutputResult", StepOutputResult("step-1"), "{{steps.step-1.outputs.result}}"},
		{"WorkflowParam", WorkflowParam("env"), "{{workflow.parameters.env}}"},
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
		expr Expr
		want string
	}{
		{"tasks", Tasks("my-task").Attr("outputs").Attr("result"), "tasks.my-task.outputs.result"},
		{"steps", StepsExpr("my-step").Attr("outputs").Attr("result"), "steps.my-step.outputs.result"},
		{"inputs", Inputs().Attr("parameters").Attr("msg"), "inputs.parameters.msg"},
		{"outputs", OutputsExpr().Attr("parameters").Attr("result"), "outputs.parameters.result"},
		{"item", Item(), "item"},
		{"item-attr", Item().Attr("name"), "item.name"},
		{"workflow", WorkflowExpr().Attr("name"), "workflow.name"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expr.String() != tt.want {
				t.Errorf("got %q, want %q", tt.expr.String(), tt.want)
			}
		})
	}
}

func TestExprConcat(t *testing.T) {
	result := Concat(" + ", E("a"), E("b"), E("c"))
	if result.String() != "a + b + c" {
		t.Errorf("got %q", result.String())
	}
}
