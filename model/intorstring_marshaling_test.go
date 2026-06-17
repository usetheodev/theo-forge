package model

import (
	"encoding/json"
	"strings"
	"testing"
)

// Phase 7 (T7.5) coverage backfill for model package — JSON tag fidelity.
// IntOrString is tracked for v0.6.0 (T4.3, ADR-002) and not yet implemented;
// this file covers what's currently shippable.

func TestWorkflowModel_RoundTripJSON(t *testing.T) {
	wf := WorkflowModel{
		APIVersion: "argoproj.io/v1alpha1",
		Kind:       "Workflow",
		Metadata: WorkflowMetadata{
			Name:      "my-wf",
			Namespace: "argo",
		},
	}
	b, err := json.Marshal(wf)
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	if !strings.Contains(s, `"apiVersion":"argoproj.io/v1alpha1"`) {
		t.Errorf("apiVersion tag missing: %s", s)
	}
	if !strings.Contains(s, `"name":"my-wf"`) {
		t.Errorf("metadata.name tag missing: %s", s)
	}
	var back WorkflowModel
	if err := json.Unmarshal(b, &back); err != nil {
		t.Fatal(err)
	}
	if back.Metadata.Name != wf.Metadata.Name {
		t.Errorf("round-trip name mismatch: got %q want %q", back.Metadata.Name, wf.Metadata.Name)
	}
}

func TestResourceList_OmitsEmptyScalars(t *testing.T) {
	rr := ResourceRequirements{
		Requests: ResourceList{CPU: "100m"},
	}
	b, _ := json.Marshal(rr)
	s := string(b)
	if !strings.Contains(s, `"cpu":"100m"`) {
		t.Errorf("cpu missing: %s", s)
	}
	if strings.Contains(s, `"memory"`) {
		t.Errorf("memory should be omitempty: %s", s)
	}
	// NOTE: Limits is a struct (not pointer), so Go renders "limits":{} even
	// when empty. omitempty on the wrapping ResourceRequirements field does
	// not affect struct fields. Tracked as documentation gap, not a bug.
	if !strings.Contains(s, `"limits"`) {
		t.Logf("limits absent (acceptable): %s", s)
	}
}

func TestParseImagePullPolicy_AllCases(t *testing.T) {
	tests := []struct {
		in   string
		want ImagePullPolicy
		err  bool
	}{
		{"Always", ImagePullAlways, false},
		{"always", ImagePullAlways, false},
		{"Never", ImagePullNever, false},
		{"never", ImagePullNever, false},
		{"IfNotPresent", ImagePullIfNotPresent, false},
		{"ifNotPresent", ImagePullIfNotPresent, false},
		{"if_not_present", ImagePullIfNotPresent, false},
		{"garbage", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			got, err := ParseImagePullPolicy(tt.in)
			if (err != nil) != tt.err {
				t.Fatalf("err = %v, wantErr %v", err, tt.err)
			}
			if got != tt.want {
				t.Errorf("got %q want %q", got, tt.want)
			}
		})
	}
}

func TestSentinelErrors_AreDistinct(t *testing.T) {
	// Sanity: every sentinel returns a non-empty message and is distinguishable.
	sentinels := []error{
		ErrPathTraversal, ErrEmptyImage, ErrEntrypointMissing,
		ErrTemplateAmbiguous, ErrTemplateMissing,
		ErrInvalidName, ErrInvalidNamespace,
		ErrResponseTooLarge, ErrRedactedTokenLoaded, ErrInvalidTimeout,
	}
	seen := map[string]bool{}
	for _, e := range sentinels {
		if e == nil || e.Error() == "" {
			t.Fatal("sentinel is nil or empty")
		}
		if seen[e.Error()] {
			t.Errorf("duplicate sentinel message: %q", e.Error())
		}
		seen[e.Error()] = true
	}
}
