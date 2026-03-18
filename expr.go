package forge

import (
	"fmt"
	"strings"
)

// Expr represents an Argo expression node.
type Expr struct {
	repr string
}

// E creates an expression from a raw string.
func E(s string) Expr {
	return Expr{repr: s}
}

// C creates a constant expression.
func C(v interface{}) Expr {
	switch val := v.(type) {
	case nil:
		return Expr{repr: "nil"}
	case bool:
		if val {
			return Expr{repr: "true"}
		}
		return Expr{repr: "false"}
	case int:
		return Expr{repr: fmt.Sprintf("%d", val)}
	case int64:
		return Expr{repr: fmt.Sprintf("%d", val)}
	case float64:
		return Expr{repr: fmt.Sprintf("%g", val)}
	case string:
		return Expr{repr: fmt.Sprintf("'%s'", val)}
	default:
		return Expr{repr: fmt.Sprintf("%v", val)}
	}
}

// String returns the expression string.
func (e Expr) String() string {
	return e.repr
}

// Tmpl returns the expression wrapped in Argo template syntax {{...}}.
func (e Expr) Tmpl() string {
	return "{{" + e.repr + "}}"
}

// Eq returns the expression wrapped in Argo expression syntax {{=...}}.
func (e Expr) Eq() string {
	return "{{=" + e.repr + "}}"
}

// Attr accesses a field on the expression.
func (e Expr) Attr(name string) Expr {
	return Expr{repr: e.repr + "." + name}
}

// Index accesses an index on the expression.
func (e Expr) Index(i int) Expr {
	return Expr{repr: fmt.Sprintf("%s[%d]", e.repr, i)}
}

// Key accesses a key on the expression.
func (e Expr) Key(k string) Expr {
	return Expr{repr: fmt.Sprintf(`%s["%s"]`, e.repr, k)}
}

// Slice returns a slice expression.
func (e Expr) Slice(start, end int) Expr {
	return Expr{repr: fmt.Sprintf("%s[%d:%d]", e.repr, start, end)}
}

// SliceFrom returns a slice from start to end.
func (e Expr) SliceFrom(start int) Expr {
	return Expr{repr: fmt.Sprintf("%s[%d:]", e.repr, start)}
}

// SliceTo returns a slice from beginning to end.
func (e Expr) SliceTo(end int) Expr {
	return Expr{repr: fmt.Sprintf("%s[:%d]", e.repr, end)}
}

// --- Comparison operators ---

func (e Expr) Equals(other Expr) Expr {
	return Expr{repr: fmt.Sprintf("%s == %s", e.repr, other.repr)}
}

func (e Expr) NotEquals(other Expr) Expr {
	return Expr{repr: fmt.Sprintf("%s != %s", e.repr, other.repr)}
}

func (e Expr) GT(other Expr) Expr {
	return Expr{repr: fmt.Sprintf("%s > %s", e.repr, other.repr)}
}

func (e Expr) GTE(other Expr) Expr {
	return Expr{repr: fmt.Sprintf("%s >= %s", e.repr, other.repr)}
}

func (e Expr) LT(other Expr) Expr {
	return Expr{repr: fmt.Sprintf("%s < %s", e.repr, other.repr)}
}

func (e Expr) LTE(other Expr) Expr {
	return Expr{repr: fmt.Sprintf("%s <= %s", e.repr, other.repr)}
}

// --- Arithmetic operators ---

func (e Expr) Add(other Expr) Expr {
	return Expr{repr: fmt.Sprintf("%s + %s", e.repr, other.repr)}
}

func (e Expr) Sub(other Expr) Expr {
	return Expr{repr: fmt.Sprintf("%s - %s", e.repr, other.repr)}
}

func (e Expr) Mul(other Expr) Expr {
	return Expr{repr: fmt.Sprintf("%s * %s", e.repr, other.repr)}
}

func (e Expr) Div(other Expr) Expr {
	return Expr{repr: fmt.Sprintf("%s / %s", e.repr, other.repr)}
}

func (e Expr) Mod(other Expr) Expr {
	return Expr{repr: fmt.Sprintf("%s %% %s", e.repr, other.repr)}
}

func (e Expr) Pow(other Expr) Expr {
	return Expr{repr: fmt.Sprintf("%s ** %s", e.repr, other.repr)}
}

// --- Unary operators ---

func (e Expr) Neg() Expr {
	return Expr{repr: "-" + e.repr}
}

func (e Expr) Not() Expr {
	return Expr{repr: "!" + e.repr}
}

// --- Logical operators ---

func (e Expr) And(other Expr) Expr {
	return Expr{repr: fmt.Sprintf("%s && %s", e.repr, other.repr)}
}

func (e Expr) OrExpr(other Expr) Expr {
	return Expr{repr: fmt.Sprintf("%s || %s", e.repr, other.repr)}
}

// --- String methods ---

func (e Expr) Contains(s string) Expr {
	return Expr{repr: fmt.Sprintf("%s.contains('%s')", e.repr, s)}
}

func (e Expr) Matches(pattern string) Expr {
	return Expr{repr: fmt.Sprintf("%s.matches('%s')", e.repr, pattern)}
}

func (e Expr) StartsWith(prefix string) Expr {
	return Expr{repr: fmt.Sprintf("%s.startsWith('%s')", e.repr, prefix)}
}

func (e Expr) EndsWith(suffix string) Expr {
	return Expr{repr: fmt.Sprintf("%s.endsWith('%s')", e.repr, suffix)}
}

func (e Expr) Length() Expr {
	return Expr{repr: e.repr + ".length()"}
}

// --- Conversion methods ---

func (e Expr) ToJSON() Expr {
	return Expr{repr: e.repr + ".toJson()"}
}

func (e Expr) AsFloat() Expr {
	return Expr{repr: e.repr + ".asFloat()"}
}

func (e Expr) AsInt() Expr {
	return Expr{repr: e.repr + ".asInt()"}
}

func (e Expr) AsStr() Expr {
	return Expr{repr: e.repr + ".string()"}
}

// --- Ternary ---

func (e Expr) Check(ifTrue, ifFalse Expr) Expr {
	return Expr{repr: fmt.Sprintf("%s ? %s : %s", e.repr, ifTrue.repr, ifFalse.repr)}
}

// --- Collection methods ---

func (e Expr) Map(lambda Expr) Expr {
	return Expr{repr: fmt.Sprintf("%s.map(%s)", e.repr, lambda.repr)}
}

func (e Expr) Filter(lambda Expr) Expr {
	return Expr{repr: fmt.Sprintf("%s.filter(%s)", e.repr, lambda.repr)}
}

// --- Sprig functions ---

// Sprig provides access to sprig template functions.
var Sprig = sprigNS{}

type sprigNS struct{}

func (sprigNS) Trim(s string) Expr {
	return Expr{repr: fmt.Sprintf("sprig.trim('%s')", s)}
}

func (sprigNS) Upper(s string) Expr {
	return Expr{repr: fmt.Sprintf("sprig.upper('%s')", s)}
}

func (sprigNS) Lower(s string) Expr {
	return Expr{repr: fmt.Sprintf("sprig.lower('%s')", s)}
}

func (sprigNS) Replace(old, new, s string) Expr {
	return Expr{repr: fmt.Sprintf("sprig.replace('%s', '%s', '%s')", old, new, s)}
}

// --- Global expression root ---

// G is the global expression root, similar to Hera's `g`.
// Usage: G.Attr("tasks").Attr("task-a").Attr("outputs").Attr("result")
var G = Expr{repr: ""}

// Tasks returns a tasks expression root.
func Tasks(name string) Expr {
	return Expr{repr: "tasks." + name}
}

// StepsExpr returns a steps expression root.
func StepsExpr(name string) Expr {
	return Expr{repr: "steps." + name}
}

// Inputs returns the inputs expression.
func Inputs() Expr {
	return Expr{repr: "inputs"}
}

// Outputs returns the outputs expression.
func OutputsExpr() Expr {
	return Expr{repr: "outputs"}
}

// Item returns the {{item}} expression for withItems loops.
func Item() Expr {
	return Expr{repr: "item"}
}

// Workflow returns a workflow-level expression.
func WorkflowExpr() Expr {
	return Expr{repr: "workflow"}
}

// --- Helper for building parameter references ---

// ParamRef creates a parameter reference expression like "{{inputs.parameters.name}}".
func ParamRef(path string) string {
	return "{{" + path + "}}"
}

// InputParam creates an input parameter reference.
func InputParam(name string) string {
	return "{{inputs.parameters." + name + "}}"
}

// OutputParam creates a task output parameter reference.
func TaskOutputParam(taskName, paramName string) string {
	return "{{tasks." + taskName + ".outputs.parameters." + paramName + "}}"
}

// StepOutputParam creates a step output parameter reference.
func StepOutputParam(stepName, paramName string) string {
	return "{{steps." + stepName + ".outputs.parameters." + paramName + "}}"
}

// TaskOutputResult creates a task output result reference.
func TaskOutputResult(taskName string) string {
	return "{{tasks." + taskName + ".outputs.result}}"
}

// StepOutputResult creates a step output result reference.
func StepOutputResult(stepName string) string {
	return "{{steps." + stepName + ".outputs.result}}"
}

// WorkflowParam creates a workflow parameter reference.
func WorkflowParam(name string) string {
	return "{{workflow.parameters." + name + "}}"
}

// Concat joins multiple expressions with a separator.
func Concat(sep string, exprs ...Expr) Expr {
	parts := make([]string, len(exprs))
	for i, e := range exprs {
		parts[i] = e.repr
	}
	return Expr{repr: strings.Join(parts, sep)}
}
