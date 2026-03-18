package forge

import "fmt"

// InvalidType is returned when a wrong type is submitted to a context.
type InvalidType struct {
	Expected string
	Got      string
}

func (e *InvalidType) Error() string {
	return fmt.Sprintf("invalid type: expected %s, got %s", e.Expected, e.Got)
}

// NodeNameConflict is returned when duplicate step/task names are detected.
type NodeNameConflict struct {
	Name string
}

func (e *NodeNameConflict) Error() string {
	return fmt.Sprintf("node name conflict: %q already exists in this context", e.Name)
}

// InvalidTemplateCall is returned when a template is called in an invalid context.
type InvalidTemplateCall struct {
	Name    string
	Context string
}

func (e *InvalidTemplateCall) Error() string {
	return fmt.Sprintf("template %q is not callable under a %s context", e.Name, e.Context)
}
