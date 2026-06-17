package forge

import "github.com/usetheodev/theo-forge/config"

// T4.2 — config wiring split from helpers.go. ADR-001 lives here.

// PreBuildHook is a function that transforms a TemplateModel before submission.
type PreBuildHook = config.PreBuildHook

// WorkflowPreBuildHook is a function that transforms a WorkflowModel before submission.
type WorkflowPreBuildHook = config.WorkflowPreBuildHook

// GlobalConfig holds default values applied to all workflows and templates.
type GlobalConfig = config.GlobalConfig

// (T3.1 / ADR-001) The previous `var globalConfig = config.GetGlobal()`
// captured the singleton pointer at package init, making config injection
// impossible. Use resolveConfig(explicit) for lazy resolution.

// NewConfig creates an independent GlobalConfig instance for dependency injection.
// Use this instead of GetGlobalConfig when you need isolated configuration
// (e.g., in tests or when building workflows with different settings concurrently).
func NewConfig() *config.GlobalConfig {
	return config.New()
}

// GetGlobalConfig returns the global configuration singleton.
// For isolated configuration (tests, concurrent builds), use NewConfig() instead.
func GetGlobalConfig() *config.GlobalConfig {
	return config.GetGlobal()
}

// resolveConfig returns the effective config for a Build() call.
// Resolution order (T3.1 / ADR-001):
//  1. explicit (e.g., Workflow.Config) when non-nil
//  2. otherwise, the package singleton config.GetGlobal()
//
// This replaces the package-level `var globalConfig = config.GetGlobal()`
// frozen at init time, which made NewConfig() hook isolation structurally
// impossible. Hooks registered on a NewConfig() instance now fire when
// that instance is supplied via WithConfig.
func resolveConfig(explicit *config.GlobalConfig) *config.GlobalConfig {
	if explicit != nil {
		return explicit
	}
	return config.GetGlobal()
}
