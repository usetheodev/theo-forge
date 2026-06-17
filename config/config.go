// Package config provides global configuration and hook management
// for the forge workflow builder.
package config

import (
	"sync"

	"github.com/usetheodev/theo-forge/model"
)

// PreBuildHook is a function that transforms a TemplateModel before submission.
type PreBuildHook func(*model.TemplateModel)

// WorkflowPreBuildHook is a function that transforms a WorkflowModel before submission.
type WorkflowPreBuildHook func(*model.WorkflowModel)

// WorkflowTemplatePreBuildHook is a function that transforms a WorkflowTemplateModel
// before submission. Dispatched by both WorkflowTemplate.Build() and
// ClusterWorkflowTemplate.Build(). (T3.2 / completeness-workflow-hook-not-fired-for-wftemplate).
type WorkflowTemplatePreBuildHook func(*model.WorkflowTemplateModel)

// CronWorkflowPreBuildHook is a function that transforms a CronWorkflowModel before submission.
// (T3.2 / completeness-workflow-hook-not-fired-for-wftemplate).
type CronWorkflowPreBuildHook func(*model.CronWorkflowModel)

// GlobalConfig holds default values applied to all workflows and templates.
//
// T4.5 — data fields are unexported. Use the SetX/GetX methods to access
// them; this guarantees the mutex is always held and prevents data races
// from direct field access. Fields used in JSON marshaling are still
// emitted under their canonical lowercase JSON tag via the redact.go
// codec (the struct literally cannot leak race-prone state because no
// caller can read or write its data outside the locked accessors).
type GlobalConfig struct {
	mu sync.RWMutex

	host               string
	token              string
	namespace          string
	image              string
	serviceAccountName string
	imagePullPolicy    model.ImagePullPolicy
	verifySSL          bool

	templateHooks         []PreBuildHook
	workflowHooks         []WorkflowPreBuildHook
	workflowTemplateHooks []WorkflowTemplatePreBuildHook
	cronWorkflowHooks     []CronWorkflowPreBuildHook

	// namedTemplateHooks / namedWorkflowHooks are the composable hook
	// system introduced by T8.2 (named, removable, fallible). See
	// named_hooks.go for the API.
	namedTemplateHooks []NamedTemplateHook
	namedWorkflowHooks []NamedWorkflowHook
}

// globalConfig is the package-level singleton.
var globalConfig = &GlobalConfig{
	image:     "python:3.11",
	verifySSL: true,
}

// New creates an independent GlobalConfig instance for dependency injection.
// Use this instead of GetGlobal when you need isolated configuration
// (e.g., in tests or when building workflows with different settings concurrently).
func New() *GlobalConfig {
	return &GlobalConfig{
		image:     "python:3.11",
		verifySSL: true,
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
	g.image = image
}

// GetImage returns the default image, or fallback if not set.
func (g *GlobalConfig) GetImage() string {
	g.mu.RLock()
	defer g.mu.RUnlock()
	if g.image == "" {
		return "python:3.11"
	}
	return g.image
}

// SetNamespace sets the default namespace.
func (g *GlobalConfig) SetNamespace(namespace string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.namespace = namespace
}

// GetNamespace returns the default namespace.
func (g *GlobalConfig) GetNamespace() string {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.namespace
}

// SetHost sets the default Argo server host.
func (g *GlobalConfig) SetHost(host string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.host = host
}

// GetHost returns the default Argo server host.
func (g *GlobalConfig) GetHost() string {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.host
}

// SetToken sets the default Bearer token.
func (g *GlobalConfig) SetToken(token string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.token = token
}

// GetToken returns the default Bearer token.
func (g *GlobalConfig) GetToken() string {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.token
}

// SetServiceAccountName sets the default service account name applied during Build().
func (g *GlobalConfig) SetServiceAccountName(name string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.serviceAccountName = name
}

// GetServiceAccountName returns the default service account name.
func (g *GlobalConfig) GetServiceAccountName() string {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.serviceAccountName
}

// SetImagePullPolicy sets the default ImagePullPolicy applied during Build().
func (g *GlobalConfig) SetImagePullPolicy(p model.ImagePullPolicy) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.imagePullPolicy = p
}

// GetImagePullPolicy returns the default ImagePullPolicy.
func (g *GlobalConfig) GetImagePullPolicy() model.ImagePullPolicy {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.imagePullPolicy
}

// SetVerifySSL sets the default TLS-verification setting.
func (g *GlobalConfig) SetVerifySSL(v bool) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.verifySSL = v
}

// GetVerifySSL returns the default TLS-verification setting.
func (g *GlobalConfig) GetVerifySSL() bool {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.verifySSL
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

// RegisterWorkflowTemplateHook registers a pre-build hook for WorkflowTemplate
// + ClusterWorkflowTemplate. (T3.2).
func (g *GlobalConfig) RegisterWorkflowTemplateHook(hook WorkflowTemplatePreBuildHook) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.workflowTemplateHooks = append(g.workflowTemplateHooks, hook)
}

// DispatchWorkflowTemplateHooks applies all registered WorkflowTemplate hooks. (T3.2).
func (g *GlobalConfig) DispatchWorkflowTemplateHooks(wt *model.WorkflowTemplateModel) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	for _, hook := range g.workflowTemplateHooks {
		hook(wt)
	}
}

// RegisterCronWorkflowHook registers a pre-build hook for CronWorkflow. (T3.2).
func (g *GlobalConfig) RegisterCronWorkflowHook(hook CronWorkflowPreBuildHook) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.cronWorkflowHooks = append(g.cronWorkflowHooks, hook)
}

// DispatchCronWorkflowHooks applies all registered CronWorkflow hooks. (T3.2).
func (g *GlobalConfig) DispatchCronWorkflowHooks(cw *model.CronWorkflowModel) {
	g.mu.RLock()
	defer g.mu.RUnlock()
	for _, hook := range g.cronWorkflowHooks {
		hook(cw)
	}
}

// ClearHooks removes all registered hooks (both anonymous and named).
func (g *GlobalConfig) ClearHooks() {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.templateHooks = nil
	g.workflowHooks = nil
	g.workflowTemplateHooks = nil
	g.cronWorkflowHooks = nil
	g.namedTemplateHooks = nil
	g.namedWorkflowHooks = nil
}

// ApplyTemplateDefaults overlays default scalar fields onto a TemplateModel
// when those fields are unset. Never overwrites a value the consumer set
// explicitly. Holds the read lock for the duration of the copy.
//
// T3.1 / ADR-001 — closes completeness-globalconfig-dead-scalars by making
// Image, ServiceAccountName, ImagePullPolicy actually take effect during Build().
//
// is high (15) because there are 3 separate fields × 2 receivers (Container,
// Script) + nil-safety. Splitting into per-field helpers would just rename
// the conditions. Tracked: revisit after T4.1 BaseTemplate extraction.
//
//nolint:gocyclo,cyclop // Each field overlay is a 2-line guarded assignment; total
func (g *GlobalConfig) ApplyTemplateDefaults(t *model.TemplateModel) {
	if g == nil || t == nil {
		return
	}
	g.mu.RLock()
	defer g.mu.RUnlock()
	if t.Container != nil {
		if t.Container.Image == "" && g.image != "" {
			t.Container.Image = g.image
		}
		if t.Container.ImagePullPolicy == "" && g.imagePullPolicy != "" {
			t.Container.ImagePullPolicy = string(g.imagePullPolicy)
		}
	}
	if t.Script != nil {
		if t.Script.Image == "" && g.image != "" {
			t.Script.Image = g.image
		}
		if t.Script.ImagePullPolicy == "" && g.imagePullPolicy != "" {
			t.Script.ImagePullPolicy = string(g.imagePullPolicy)
		}
	}
	if t.ServiceAccountName == "" && g.serviceAccountName != "" {
		t.ServiceAccountName = g.serviceAccountName
	}
}

// Reset restores the global config to defaults.
func (g *GlobalConfig) Reset() {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.host = ""
	g.token = ""
	g.namespace = ""
	g.image = "python:3.11"
	g.serviceAccountName = ""
	g.imagePullPolicy = ""
	g.verifySSL = true
	g.templateHooks = nil
	g.workflowHooks = nil
	g.workflowTemplateHooks = nil
	g.cronWorkflowHooks = nil
	g.namedTemplateHooks = nil
	g.namedWorkflowHooks = nil
}
