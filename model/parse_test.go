package model

import (
	"errors"
	"strings"
	"testing"
)

// Coverage backfill for model (T7.5 round 2).
// Targets ParseWorkflowStatus and *InvalidType.Error.

func TestParseWorkflowStatus_AllCases(t *testing.T) {
	tests := []struct {
		in   string
		want WorkflowStatus
		err  bool
	}{
		{"Pending", WorkflowPending, false},
		{"Running", WorkflowRunning, false},
		{"Succeeded", WorkflowSucceeded, false},
		{"Failed", WorkflowFailed, false},
		{"Error", WorkflowError, false},
		{"Terminated", WorkflowTerminated, false},
		{"garbage", "", true},
		{"", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got, err := ParseWorkflowStatus(tt.in)
			if (err != nil) != tt.err {
				t.Fatalf("err = %v, wantErr %v", err, tt.err)
			}
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func TestInvalidType_Error(t *testing.T) {
	e := &InvalidType{Expected: "int", Got: "string"}
	msg := e.Error()
	if !strings.Contains(msg, "int") || !strings.Contains(msg, "string") {
		t.Errorf("Error() = %q, want both type names", msg)
	}
	// Test that ParseImagePullPolicy returns *InvalidType.
	_, err := ParseImagePullPolicy("garbage")
	if err == nil {
		t.Fatal("expected error")
	}
	var iv *InvalidType
	if !errors.As(err, &iv) {
		t.Errorf("expected *InvalidType, got %T", err)
	}
}
