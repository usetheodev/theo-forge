// Package expr provides a fluent DSL for building Argo Workflows expressions.
//
// Expressions are used in conditional logic, parameter references, and
// template expressions within Argo Workflows. This package can be used
// independently of the [forge] builder package.
//
// # Example
//
//	e := expr.Tasks("task-a").Attr("outputs").Attr("result")
//	fmt.Println(e.Tmpl()) // "{{tasks.task-a.outputs.result}}"
//
//	cond := expr.E("inputs.parameters.env").Equals(expr.C("prod"))
//	fmt.Println(cond.Eq()) // "{{=inputs.parameters.env == 'prod'}}"
package expr
