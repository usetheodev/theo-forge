package client

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/usetheodev/theo-forge/model"
)

// T4.11 — WorkflowTemplate and CronWorkflow client operations
// (arch-client-buildable-interface-duplication).
//
// TemplateBuildable and CronBuildable are role interfaces defined locally
// to avoid importing the root forge package (preserves the inviolable
// dependency direction documented in .claude/rules/architecture.md).

// TemplateBuildable produces a serializable WorkflowTemplateModel.
type TemplateBuildable interface {
	Build() (model.WorkflowTemplateModel, error)
}

// CronBuildable produces a serializable CronWorkflowModel.
type CronBuildable interface {
	Build() (model.CronWorkflowModel, error)
}

// WorkflowTemplateCreateRequest is the request body for creating a WorkflowTemplate.
type WorkflowTemplateCreateRequest struct {
	Template  model.WorkflowTemplateModel `json:"template"`
	Namespace string                      `json:"namespace,omitempty"`
}

// CronWorkflowCreateRequest is the request body for creating a CronWorkflow.
type CronWorkflowCreateRequest struct {
	CronWorkflow model.CronWorkflowModel `json:"cronWorkflow"`
	Namespace    string                  `json:"namespace,omitempty"`
}

// CreateWorkflowTemplateFromModel submits a pre-built WorkflowTemplate model to the Argo server.
func (s *WorkflowsService) CreateWorkflowTemplateFromModel(ctx context.Context, m model.WorkflowTemplateModel, namespace string) (model.WorkflowTemplateModel, error) {
	ns := namespace
	if ns == "" {
		ns = s.Namespace
	}
	if err := validateK8sNamespace(ns); err != nil {
		return model.WorkflowTemplateModel{}, err
	}
	body := WorkflowTemplateCreateRequest{Template: m}
	respBody, err := s.doRequest(ctx, http.MethodPost, "/api/v1/workflow-templates/"+ns, body)
	if err != nil {
		return model.WorkflowTemplateModel{}, err
	}
	var result model.WorkflowTemplateModel
	if err := json.Unmarshal(respBody, &result); err != nil {
		return model.WorkflowTemplateModel{}, fmt.Errorf("unmarshal response: %w", err)
	}
	return result, nil
}

// CreateWorkflowTemplate builds a TemplateBuildable and submits it.
func (s *WorkflowsService) CreateWorkflowTemplate(ctx context.Context, b TemplateBuildable, namespace string) (model.WorkflowTemplateModel, error) {
	m, err := b.Build()
	if err != nil {
		return model.WorkflowTemplateModel{}, fmt.Errorf("build workflow template: %w", err)
	}
	return s.CreateWorkflowTemplateFromModel(ctx, m, namespace)
}

// CreateCronWorkflowFromModel submits a pre-built CronWorkflow model to the Argo server.
func (s *WorkflowsService) CreateCronWorkflowFromModel(ctx context.Context, m model.CronWorkflowModel, namespace string) (model.CronWorkflowModel, error) {
	ns := namespace
	if ns == "" {
		ns = s.Namespace
	}
	if err := validateK8sNamespace(ns); err != nil {
		return model.CronWorkflowModel{}, err
	}
	body := CronWorkflowCreateRequest{CronWorkflow: m}
	respBody, err := s.doRequest(ctx, http.MethodPost, "/api/v1/cron-workflows/"+ns, body)
	if err != nil {
		return model.CronWorkflowModel{}, err
	}
	var result model.CronWorkflowModel
	if err := json.Unmarshal(respBody, &result); err != nil {
		return model.CronWorkflowModel{}, fmt.Errorf("unmarshal response: %w", err)
	}
	return result, nil
}

// CreateCronWorkflow builds a CronBuildable and submits it.
func (s *WorkflowsService) CreateCronWorkflow(ctx context.Context, b CronBuildable, namespace string) (model.CronWorkflowModel, error) {
	m, err := b.Build()
	if err != nil {
		return model.CronWorkflowModel{}, fmt.Errorf("build cron workflow: %w", err)
	}
	return s.CreateCronWorkflowFromModel(ctx, m, namespace)
}
