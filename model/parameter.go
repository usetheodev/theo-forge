package model

// ParameterModel is the serializable representation of a Parameter,
// matching the Argo Workflows API schema.
type ParameterModel struct {
	Name        string     `json:"name" yaml:"name"`
	Value       *string    `json:"value,omitempty" yaml:"value,omitempty"`
	Default     *string    `json:"default,omitempty" yaml:"default,omitempty"`
	Description string     `json:"description,omitempty" yaml:"description,omitempty"`
	Enum        []string   `json:"enum,omitempty" yaml:"enum,omitempty"`
	GlobalName  string     `json:"globalName,omitempty" yaml:"globalName,omitempty"`
	ValueFrom   *ValueFrom `json:"valueFrom,omitempty" yaml:"valueFrom,omitempty"`
}

// ValueFrom describes a location in which to obtain the value to a parameter.
type ValueFrom struct {
	// Path is a file path to read the value from.
	Path string `json:"path,omitempty" yaml:"path,omitempty"`
	// Expression is an expression to evaluate.
	Expression string `json:"expression,omitempty" yaml:"expression,omitempty"`
	// JSONPath is a JSONPath expression to evaluate against the resource.
	JSONPath string `json:"jsonPath,omitempty" yaml:"jsonPath,omitempty"`
	// JQFilter is a jq expression to evaluate against the resource.
	JQFilter string `json:"jqFilter,omitempty" yaml:"jqFilter,omitempty"`
	// Parameter is a reference to another parameter.
	Parameter string `json:"parameter,omitempty" yaml:"parameter,omitempty"`
	// ConfigMapKeyRef references a key in a ConfigMap.
	ConfigMapKeyRef *ConfigMapKeyRef `json:"configMapKeyRef,omitempty" yaml:"configMapKeyRef,omitempty"`
	// Default is the default value if the source cannot be resolved.
	Default *string `json:"default,omitempty" yaml:"default,omitempty"`
}
