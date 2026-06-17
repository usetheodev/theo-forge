// Package config — composable hook system (T8.2).
//
// The original hook API (RegisterTemplateHook + DispatchTemplateHooks) is
// fire-and-forget: hooks are anonymous, cannot be removed individually,
// and their errors are silently dropped. This file layers a NAMED hook
// API on top that addresses all three limitations without breaking
// existing callers.
//
// Old (still works):
//
//	cfg.RegisterTemplateHook(func(t *model.TemplateModel) { ... })  // returns nothing
//	cfg.DispatchTemplateHooks(t)                                     // ignores errors
//
// New (T8.2 composable):
//
//	id := cfg.RegisterNamedTemplateHook("inject-labels", func(t *model.TemplateModel) error { ... })
//	cfg.RemoveTemplateHook(id)
//	err := cfg.DispatchNamedTemplateHooks(t)   // first hook error short-circuits
package config

import (
	"fmt"

	"github.com/usetheodev/theo-forge/model"
)

// NamedHookID is an opaque handle returned by RegisterNamedXxxHook
// and consumed by RemoveXxxHook. Stable for the lifetime of the cfg.
type NamedHookID string

// NamedTemplateHook is a fallible template hook with an identity.
type NamedTemplateHook struct {
	ID NamedHookID
	Fn func(*model.TemplateModel) error
}

// NamedWorkflowHook is a fallible workflow hook with an identity.
type NamedWorkflowHook struct {
	ID NamedHookID
	Fn func(*model.WorkflowModel) error
}

// RegisterNamedTemplateHook registers a fallible template hook with a
// caller-supplied name. The name MUST be unique within this config;
// duplicate registration returns an error. (T8.2).
func (g *GlobalConfig) RegisterNamedTemplateHook(name string, fn func(*model.TemplateModel) error) (NamedHookID, error) {
	if name == "" {
		return "", fmt.Errorf("config: named hook requires a non-empty name")
	}
	if fn == nil {
		return "", fmt.Errorf("config: named hook %q requires a non-nil function", name)
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	for _, h := range g.namedTemplateHooks {
		if h.ID == NamedHookID(name) {
			return "", fmt.Errorf("config: template hook %q already registered", name)
		}
	}
	id := NamedHookID(name)
	g.namedTemplateHooks = append(g.namedTemplateHooks, NamedTemplateHook{ID: id, Fn: fn})
	return id, nil
}

// RemoveTemplateHook removes the named template hook registered under id.
// Returns true if a hook was found and removed; false otherwise. (T8.2).
func (g *GlobalConfig) RemoveTemplateHook(id NamedHookID) bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	for i, h := range g.namedTemplateHooks {
		if h.ID == id {
			g.namedTemplateHooks = append(g.namedTemplateHooks[:i], g.namedTemplateHooks[i+1:]...)
			return true
		}
	}
	return false
}

// DispatchNamedTemplateHooks invokes every named template hook in
// registration order. The FIRST hook that returns a non-nil error
// short-circuits; subsequent hooks are not invoked. The returned error
// is wrapped with the offending hook name for traceability. (T8.2).
func (g *GlobalConfig) DispatchNamedTemplateHooks(t *model.TemplateModel) error {
	g.mu.RLock()
	hooks := make([]NamedTemplateHook, len(g.namedTemplateHooks))
	copy(hooks, g.namedTemplateHooks)
	g.mu.RUnlock()
	for _, h := range hooks {
		if err := h.Fn(t); err != nil {
			return fmt.Errorf("template hook %q: %w", h.ID, err)
		}
	}
	return nil
}

// RegisterNamedWorkflowHook is the workflow-level counterpart of
// RegisterNamedTemplateHook. (T8.2).
func (g *GlobalConfig) RegisterNamedWorkflowHook(name string, fn func(*model.WorkflowModel) error) (NamedHookID, error) {
	if name == "" {
		return "", fmt.Errorf("config: named hook requires a non-empty name")
	}
	if fn == nil {
		return "", fmt.Errorf("config: named hook %q requires a non-nil function", name)
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	for _, h := range g.namedWorkflowHooks {
		if h.ID == NamedHookID(name) {
			return "", fmt.Errorf("config: workflow hook %q already registered", name)
		}
	}
	id := NamedHookID(name)
	g.namedWorkflowHooks = append(g.namedWorkflowHooks, NamedWorkflowHook{ID: id, Fn: fn})
	return id, nil
}

// RemoveWorkflowHook removes the named workflow hook registered under id. (T8.2).
func (g *GlobalConfig) RemoveWorkflowHook(id NamedHookID) bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	for i, h := range g.namedWorkflowHooks {
		if h.ID == id {
			g.namedWorkflowHooks = append(g.namedWorkflowHooks[:i], g.namedWorkflowHooks[i+1:]...)
			return true
		}
	}
	return false
}

// DispatchNamedWorkflowHooks invokes every named workflow hook;
// fail-fast on first error. (T8.2).
func (g *GlobalConfig) DispatchNamedWorkflowHooks(w *model.WorkflowModel) error {
	g.mu.RLock()
	hooks := make([]NamedWorkflowHook, len(g.namedWorkflowHooks))
	copy(hooks, g.namedWorkflowHooks)
	g.mu.RUnlock()
	for _, h := range hooks {
		if err := h.Fn(w); err != nil {
			return fmt.Errorf("workflow hook %q: %w", h.ID, err)
		}
	}
	return nil
}

// ListNamedTemplateHooks returns the IDs of every registered named template
// hook in registration order. Useful for diagnostics and tests. (T8.2).
func (g *GlobalConfig) ListNamedTemplateHooks() []NamedHookID {
	g.mu.RLock()
	defer g.mu.RUnlock()
	out := make([]NamedHookID, len(g.namedTemplateHooks))
	for i, h := range g.namedTemplateHooks {
		out[i] = h.ID
	}
	return out
}

// ListNamedWorkflowHooks returns the IDs of every registered named workflow
// hook in registration order. (T8.2).
func (g *GlobalConfig) ListNamedWorkflowHooks() []NamedHookID {
	g.mu.RLock()
	defer g.mu.RUnlock()
	out := make([]NamedHookID, len(g.namedWorkflowHooks))
	for i, h := range g.namedWorkflowHooks {
		out[i] = h.ID
	}
	return out
}
