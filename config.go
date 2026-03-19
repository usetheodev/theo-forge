package forge

import "github.com/usetheo/theo/forge/config"

// PreBuildHook is a function that transforms a TemplateModel before submission.
type PreBuildHook = config.PreBuildHook

// WorkflowPreBuildHook is a function that transforms a WorkflowModel before submission.
type WorkflowPreBuildHook = config.WorkflowPreBuildHook

// GlobalConfig holds default values applied to all workflows and templates.
type GlobalConfig = config.GlobalConfig

// globalConfig is the package-level reference to the global singleton.
var globalConfig = config.GetGlobal()

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

