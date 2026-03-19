package forge

import (
	"testing"

	"github.com/usetheo/theo/forge/model"
)

func TestParseWorkflowStatus(t *testing.T) {
	tests := []struct {
		input string
		want  WorkflowStatus
		err   bool
	}{
		{"Pending", WorkflowPending, false},
		{"Running", WorkflowRunning, false},
		{"Succeeded", WorkflowSucceeded, false},
		{"Failed", WorkflowFailed, false},
		{"Error", WorkflowError, false},
		{"Terminated", WorkflowTerminated, false},
		{"Unknown", "", true},
		{"", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := model.ParseWorkflowStatus(tt.input)
			if tt.err && err == nil {
				t.Fatal("expected error")
			}
			if !tt.err && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRetryStrategyBuild(t *testing.T) {
	limit := 3
	rs := RetryStrategy{
		Limit:       &limit,
		RetryPolicy: RetryOnFailure,
		Backoff: &Backoff{
			Duration:    "5s",
			Factor:      ptrInt(2),
			MaxDuration: "1m",
		},
	}
	model := rs.Build()
	if model.Limit != "3" {
		t.Errorf("limit = %v, want \"3\"", model.Limit)
	}
	if model.RetryPolicy != "OnFailure" {
		t.Errorf("policy = %q", model.RetryPolicy)
	}
	if model.Backoff == nil || model.Backoff.Duration != "5s" {
		t.Errorf("backoff duration = %v", model.Backoff)
	}
}

func TestMetricStructure(t *testing.T) {
	m := Metric{
		Name: "build_duration",
		Help: "Duration of build step",
		Labels: []Label{{Key: "step", Value: "build"}},
		Gauge:  &Gauge{Value: "{{duration}}", Realtime: ptrBool(true)},
	}
	if m.Name != "build_duration" {
		t.Errorf("name = %q", m.Name)
	}
	if m.Gauge == nil || m.Gauge.Realtime == nil || !*m.Gauge.Realtime {
		t.Error("expected realtime gauge")
	}
}

func TestErrorTypes(t *testing.T) {
	t.Run("InvalidType", func(t *testing.T) {
		err := &model.InvalidType{Expected: "Task", Got: "Step"}
		if err.Error() != "invalid type: expected Task, got Step" {
			t.Errorf("unexpected message: %s", err.Error())
		}
	})
	t.Run("NodeNameConflict", func(t *testing.T) {
		err := &NodeNameConflict{Name: "my-task"}
		if err.Error() != `node name conflict: "my-task" already exists in this context` {
			t.Errorf("unexpected message: %s", err.Error())
		}
	})
	t.Run("InvalidTemplateCall", func(t *testing.T) {
		err := &InvalidTemplateCall{Name: "echo", Context: "Workflow"}
		if err.Error() != `template "echo" is not callable under a Workflow context` {
			t.Errorf("unexpected message: %s", err.Error())
		}
	})
}

func ptrInt(i int) *int { return &i }
