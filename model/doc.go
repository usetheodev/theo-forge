// Package model contains serializable types matching the Argo Workflows API schema.
//
// These types are the wire format — they serialize to JSON/YAML that the Argo
// Workflows controller understands. Builder types in the parent [forge] package
// construct these models via their Build() methods.
//
// Most users should not need to import this package directly. The [forge] package
// re-exports the most commonly used types as aliases.
package model
