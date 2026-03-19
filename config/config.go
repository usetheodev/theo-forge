// Package config provides global configuration and hook management
// for the forge workflow builder.
package config

import (
	"sync"

	"github.com/usetheo/theo/forge/model"
)

// PreBuildHook is a function that transforms a TemplateModel before submission.
type PreBuildHook func(*model.TemplateModel)

// WorkflowPreBuildHook is a function that transforms a WorkflowModel before submission.
type WorkflowPreBuildHook func(*model.WorkflowModel)

// GlobalConfig holds default values applied to all workflows and templates.
type GlobalConfig struct {
	mu sync.RWMutex

	// Host is the default Argo server URL.
	Host string
	// Token is the default Bearer token.
	Token string
	// Namespace is the default namespace.
	Namespace string
	// Image is the default container image.
	Image string
	// ServiceAccountName is the default service account.
	ServiceAccountName string
	// ImagePullPolicy is the default pull policy.
	ImagePullPolicy model.ImagePullPolicy
	// VerifySSL is the default SSL verification setting.
	VerifySSL bool

	// templateHooks are pre-build hooks for templates.
	templateHooks []PreBuildHook
	// workflowHooks are pre-build hooks for workflows.
	workflowHooks []WorkflowPreBuildHook
}

// globalConfig is the package-level singleton.
var globalConfig = &GlobalConfig{
	Image:     "python:3.11",
	VerifySSL: true,
}

// New creates an independent GlobalConfig instance for dependency injection.
// Use this instead of GetGlobal when you need isolated configuration
// (e.g., in tests or when building workflows with different settings concurrently).
func New() *GlobalConfig {
	return &GlobalConfig{
		Image:     "python:3.11",
		VerifySSL: true,
	}
}

// GetGlobal returns the global configuration singleton.
// For isolated configuration (tests, concurrent builds), use New() instead.
func GetGlobal() *GlobalConfig {
	return globalConfig
}

// SetImage sets the default container image.
func (g *GlobalConfig) SetImage(image string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.Image = image
}

// GetImage returns the default image, or fallback if not set.
func (g *GlobalConfig) GetImage() string {
	g.mu.RLock()
	defer g.mu.RUnlock()
	if g.Image == "" {
		return "python:3.11"
	}
	return g.Image
}

// SetNamespace sets the default namespace.
func (g *GlobalConfig) SetNamespace(namespace string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.Namespace = namespace
}

// GetNamespace returns the default namespace.
func (g *GlobalConfig) GetNamespace() string {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.Namespace
}

// SetHost sets the default Argo server host.
func (g *GlobalConfig) SetHost(host string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.Host = host
}

// SetToken sets the default Bearer token.
func (g *GlobalConfig) SetToken(token string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.Token = token
}

// RegisterTemplateHook registers a pre-build hook for templates.
func (g *GlobalConfig) RegisterTemplateHook(hook PreBuildHook) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.templateHooks = append(g.templateHooks, hook)
}

// RegisterWorkflowHook registers a pre-build hook for workflows.
func (g *GlobalConfig) RegisterWorkflowHook(hook WorkflowPreBuildHook) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.workflowHooks = append(g.workflowHooks, hook)
}

// DispatchTemplateHooks applies all registered template hooks.
func (g *GlobalConfig) DispatchTemplateHooks(tpl *model.TemplateModel) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	for _, hook := range g.templateHooks {
		hook(tpl)
	}
}

// DispatchWorkflowHooks applies all registered workflow hooks.
func (g *GlobalConfig) DispatchWorkflowHooks(wf *model.WorkflowModel) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	for _, hook := range g.workflowHooks {
		hook(wf)
	}
}

// ClearHooks removes all registered hooks.
func (g *GlobalConfig) ClearHooks() {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.templateHooks = nil
	g.workflowHooks = nil
}

// Reset restores the global config to defaults.
func (g *GlobalConfig) Reset() {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.Host = ""
	g.Token = ""
	g.Namespace = ""
	g.Image = "python:3.11"
	g.ServiceAccountName = ""
	g.ImagePullPolicy = ""
	g.VerifySSL = true
	g.templateHooks = nil
	g.workflowHooks = nil
}
