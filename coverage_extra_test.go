package forge

import (
	"testing"
)

// Cover Workflow.buildVolumeClaimTemplates
func TestWorkflowWithVolumeClaimTemplates(t *testing.T) {
	w := &Workflow{
		Name:       "with-pvcs",
		Entrypoint: "main",
		VolumeClaimTemplates: []PVCVolume{
			{
				BaseVolume:       BaseVolume{Name: "work", MountPath: "/work"},
				Size:             "5Gi",
				StorageClassName: "fast",
				AccessModes:      []AccessMode{ReadWriteOnce},
			},
			{
				BaseVolume:  BaseVolume{Name: "cache", MountPath: "/cache"},
				Size:        "10Gi",
				AccessModes: []AccessMode{ReadWriteMany},
			},
		},
		Templates: []Templatable{
			&Container{
				Name:  "main",
				Image: "alpine",
				VolumeMounts: []VolumeBuilder{
					&PVCVolume{BaseVolume: BaseVolume{Name: "work", MountPath: "/work"}},
				},
			},
		},
	}
	model, err := w.Build()
	if err != nil {
		t.Fatal(err)
	}
	if len(model.Spec.VolumeClaimTemplates) != 2 {
		t.Fatalf("pvcs = %d, want 2", len(model.Spec.VolumeClaimTemplates))
	}
	if model.Spec.VolumeClaimTemplates[0].Metadata.Name != "work" {
		t.Errorf("pvc[0].name = %q", model.Spec.VolumeClaimTemplates[0].Metadata.Name)
	}
	if model.Spec.VolumeClaimTemplates[0].Spec.StorageClassName != "fast" {
		t.Errorf("storageClass = %q", model.Spec.VolumeClaimTemplates[0].Spec.StorageClassName)
	}
	if model.Spec.VolumeClaimTemplates[1].Spec.Resources.Requests.Storage != "10Gi" {
		t.Errorf("storage = %q", model.Spec.VolumeClaimTemplates[1].Spec.Resources.Requests.Storage)
	}
}

// Cover Workflow.buildArguments with artifacts
func TestWorkflowWithArgumentArtifacts(t *testing.T) {
	w := &Workflow{
		Name:       "with-art-args",
		Entrypoint: "main",
		ArgumentArtifacts: []ArtifactBuilder{
			&S3Artifact{
				Artifact: Artifact{Name: "input-data", Path: "/data"},
				Bucket:   "my-bucket",
				Key:      "input.csv",
			},
		},
		Templates: []Templatable{&Container{Name: "main", Image: "alpine"}},
	}
	model, err := w.Build()
	if err != nil {
		t.Fatal(err)
	}
	if model.Spec.Arguments == nil {
		t.Fatal("expected arguments")
	}
	if len(model.Spec.Arguments.Artifacts) != 1 {
		t.Fatalf("artifact args = %d", len(model.Spec.Arguments.Artifacts))
	}
	if model.Spec.Arguments.Artifacts[0].S3 == nil {
		t.Error("expected S3 artifact arg")
	}
}

// Cover Workflow.buildMetrics
func TestWorkflowWithMetrics(t *testing.T) {
	w := &Workflow{
		Name:       "with-metrics",
		Entrypoint: "main",
		Metrics: []Metric{
			{
				Name: "workflow_duration",
				Help: "Duration of workflow",
				Gauge: &Gauge{
					Value:    "{{workflow.duration}}",
					Realtime: ptrBool(true),
				},
			},
		},
		Templates: []Templatable{&Container{Name: "main", Image: "alpine"}},
	}
	model, err := w.Build()
	if err != nil {
		t.Fatal(err)
	}
	if model.Spec.Metrics == nil {
		t.Fatal("expected metrics")
	}
	if len(model.Spec.Metrics.Prometheus) != 1 {
		t.Fatalf("metrics = %d", len(model.Spec.Metrics.Prometheus))
	}
}

// Cover Container.buildMetrics
func TestContainerWithMetrics(t *testing.T) {
	c := &Container{
		Name:  "with-metrics",
		Image: "alpine",
		Metrics: []Metric{
			{Name: "step_duration", Help: "Step duration", Counter: &Counter{Value: "1"}},
		},
	}
	tpl, err := c.BuildTemplate()
	if err != nil {
		t.Fatal(err)
	}
	if tpl.Metrics == nil || len(tpl.Metrics.Prometheus) != 1 {
		t.Fatal("expected 1 metric")
	}
}

// Cover Script.buildMetrics and buildMetadata
func TestScriptWithMetricsAndMetadata(t *testing.T) {
	s := &Script{
		Name:        "with-meta",
		Image:       "python:3.11",
		Command:     []string{"python"},
		Source:      "print('hello')",
		Labels:      map[string]string{"team": "backend"},
		Annotations: map[string]string{"note": "test"},
		Metrics: []Metric{
			{Name: "script_runs", Help: "count", Counter: &Counter{Value: "1"}},
		},
	}
	tpl, err := s.BuildTemplate()
	if err != nil {
		t.Fatal(err)
	}
	if tpl.Metadata == nil {
		t.Fatal("expected metadata")
	}
	if tpl.Metadata.Labels["team"] != "backend" {
		t.Errorf("label = %q", tpl.Metadata.Labels["team"])
	}
	if tpl.Metrics == nil {
		t.Fatal("expected metrics")
	}
}

// Cover Script.buildVolumeMounts
func TestScriptWithVolumeMounts(t *testing.T) {
	s := &Script{
		Name:    "with-mounts",
		Image:   "python:3.11",
		Command: []string{"python"},
		Source:  "print('hello')",
		VolumeMounts: []VolumeBuilder{
			&EmptyDirVolume{BaseVolume: BaseVolume{Name: "tmp", MountPath: "/tmp/work"}},
		},
	}
	tpl, err := s.BuildTemplate()
	if err != nil {
		t.Fatal(err)
	}
	if len(tpl.Script.VolumeMounts) != 1 {
		t.Fatalf("mounts = %d", len(tpl.Script.VolumeMounts))
	}
}

// Cover DAG.buildOutputs with artifacts
func TestDAGWithInputAndOutputArtifacts(t *testing.T) {
	dag := &DAG{
		Name: "with-art-io",
		InputArtifacts: []ArtifactBuilder{
			&Artifact{Name: "input", Path: "/tmp/in"},
		},
		OutputArtifacts: []ArtifactBuilder{
			&Artifact{Name: "output", Path: "/tmp/out"},
		},
	}
	tpl, err := dag.BuildTemplate()
	if err != nil {
		t.Fatal(err)
	}
	if tpl.Inputs == nil || len(tpl.Inputs.Artifacts) != 1 {
		t.Fatal("expected 1 input artifact")
	}
	if tpl.Outputs == nil || len(tpl.Outputs.Artifacts) != 1 {
		t.Fatal("expected 1 output artifact")
	}
}

// Cover Steps.buildOutputs with artifacts
func TestStepsWithOutputArtifacts(t *testing.T) {
	steps := &Steps{
		Name: "with-art-out",
		OutputArtifacts: []ArtifactBuilder{
			&Artifact{Name: "result", Path: "/tmp/result"},
		},
	}
	tpl, err := steps.BuildTemplate()
	if err != nil {
		t.Fatal(err)
	}
	if tpl.Outputs == nil || len(tpl.Outputs.Artifacts) != 1 {
		t.Fatal("expected 1 output artifact")
	}
}

// Cover Step with artifact arguments
func TestStepWithArtifactArguments(t *testing.T) {
	s := &Step{
		Name:     "with-art-args",
		Template: "process",
		ArgumentArtifacts: []ArtifactBuilder{
			&Artifact{Name: "data", From: "{{steps.gen.outputs.artifacts.output}}"},
		},
	}
	model, err := s.BuildStep()
	if err != nil {
		t.Fatal(err)
	}
	if model.Arguments == nil || len(model.Arguments.Artifacts) != 1 {
		t.Fatal("expected 1 artifact argument")
	}
}

// Cover Task with artifact arguments
func TestTaskWithArtifactArguments(t *testing.T) {
	task := &Task{
		Name:     "with-art-args",
		Template: "process",
		ArgumentArtifacts: []ArtifactBuilder{
			&Artifact{Name: "data", From: "{{tasks.gen.outputs.artifacts.output}}"},
		},
	}
	model, err := task.BuildDAGTask()
	if err != nil {
		t.Fatal(err)
	}
	if model.Arguments == nil || len(model.Arguments.Artifacts) != 1 {
		t.Fatal("expected 1 artifact argument")
	}
}

// Cover PVCVolume.BuildPVC default access modes
func TestPVCVolumeDefaultAccessModes(t *testing.T) {
	v := PVCVolume{
		BaseVolume: BaseVolume{Name: "default-mode", MountPath: "/data"},
		Size:       "1Gi",
	}
	pvc, err := v.BuildPVC()
	if err != nil {
		t.Fatal(err)
	}
	if len(pvc.Spec.AccessModes) != 1 || pvc.Spec.AccessModes[0] != "ReadWriteOnce" {
		t.Errorf("default access modes = %v", pvc.Spec.AccessModes)
	}
}

// Cover Container with output parameters and artifacts
func TestContainerWithOutputs(t *testing.T) {
	c := &Container{
		Name:  "with-outputs",
		Image: "alpine",
		Outputs: []Parameter{
			{Name: "result", ValueFrom: &ValueFrom{Path: "/tmp/result"}},
		},
		OutputArtifacts: []ArtifactBuilder{
			&Artifact{Name: "logs", Path: "/tmp/logs"},
		},
	}
	tpl, err := c.BuildTemplate()
	if err != nil {
		t.Fatal(err)
	}
	if tpl.Outputs == nil {
		t.Fatal("expected outputs")
	}
	if len(tpl.Outputs.Parameters) != 1 {
		t.Errorf("output params = %d", len(tpl.Outputs.Parameters))
	}
	if len(tpl.Outputs.Artifacts) != 1 {
		t.Errorf("output artifacts = %d", len(tpl.Outputs.Artifacts))
	}
}

// Cover Workflow with retry strategy
func TestWorkflowWithRetryStrategy(t *testing.T) {
	limit := 5
	w := &Workflow{
		Name:       "with-retry",
		Entrypoint: "main",
		RetryStrategy: &RetryStrategy{
			Limit:       &limit,
			RetryPolicy: RetryAlways,
		},
		Templates: []Templatable{&Container{Name: "main", Image: "alpine"}},
	}
	model, err := w.Build()
	if err != nil {
		t.Fatal(err)
	}
	if model.Spec.RetryStrategy == nil {
		t.Fatal("expected retry strategy")
	}
	if model.Spec.RetryStrategy.Limit != "5" {
		t.Errorf("limit = %v, want \"5\"", model.Spec.RetryStrategy.Limit)
	}
}
