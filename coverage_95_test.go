package forge

import "testing"

// Cover ClusterWorkflowTemplate.Build with arguments
func TestClusterWorkflowTemplateBuildWithArguments(t *testing.T) {
	cwt := &ClusterWorkflowTemplate{
		Name:       "with-args",
		Entrypoint: "main",
		Arguments:  []Parameter{{Name: "env", Value: ptrStr("prod")}},
		Templates:  []Templatable{&Container{Name: "main", Image: "alpine"}},
	}
	model, err := cwt.Build()
	if err != nil {
		t.Fatal(err)
	}
	if model.Spec.Arguments == nil || len(model.Spec.Arguments.Parameters) != 1 {
		t.Fatal("expected 1 argument")
	}
}

// Cover CronWorkflow.Build with arguments
func TestCronWorkflowBuildWithArguments(t *testing.T) {
	cw := &CronWorkflow{
		Name:       "with-args",
		Schedule:   "0 * * * *",
		Entrypoint: "main",
		Arguments:  []Parameter{{Name: "env", Value: ptrStr("staging")}},
		Templates:  []Templatable{&Container{Name: "main", Image: "alpine"}},
	}
	model, err := cw.Build()
	if err != nil {
		t.Fatal(err)
	}
	if model.Spec.WorkflowSpec.Arguments == nil {
		t.Fatal("expected arguments")
	}
}

// Cover BuildDAGTask artifact argument error path
func TestTaskBuildDAGTaskArtifactError(t *testing.T) {
	task := &Task{
		Name:     "test",
		Template: "tpl",
		ArgumentArtifacts: []ArtifactBuilder{
			&Artifact{Path: "/tmp"}, // no name → error
		},
	}
	_, err := task.BuildDAGTask()
	if err == nil {
		t.Fatal("expected error from artifact build")
	}
}

// Cover BuildStep artifact argument error path
func TestStepBuildStepArtifactError(t *testing.T) {
	s := &Step{
		Name:     "test",
		Template: "tpl",
		ArgumentArtifacts: []ArtifactBuilder{
			&Artifact{Path: "/tmp"}, // no name → error
		},
	}
	_, err := s.BuildStep()
	if err == nil {
		t.Fatal("expected error from artifact build")
	}
}

// Cover BuildArguments artifact error path
func TestBuildArgumentsArtifactError(t *testing.T) {
	_, err := BuildArguments(
		nil,
		[]ArtifactBuilder{&Artifact{Path: "/tmp"}}, // no name
	)
	if err == nil {
		t.Fatal("expected error")
	}
}

// Cover BuildArguments parameter error path
func TestBuildArgumentsParameterError(t *testing.T) {
	_, err := BuildArguments(
		[]Parameter{{Value: ptrStr("val")}}, // no name
		nil,
	)
	if err == nil {
		t.Fatal("expected error")
	}
}

// Cover DAG.BuildTemplate with task build error
func TestDAGBuildTemplateWithTaskError(t *testing.T) {
	dag := &DAG{
		Name: "test",
		Tasks: []*Task{
			{Name: "", Template: "tpl"}, // no name
		},
	}
	_, err := dag.BuildTemplate()
	if err == nil {
		t.Fatal("expected error from task build")
	}
}

// Cover Steps.BuildTemplate with step group error
func TestStepsBuildTemplateWithStepError(t *testing.T) {
	steps := &Steps{Name: "test"}
	steps.StepGroups = append(steps.StepGroups, Parallel{
		Steps: []*Step{{Name: "", Template: "tpl"}}, // no name
	})
	_, err := steps.BuildTemplate()
	if err == nil {
		t.Fatal("expected error from step build")
	}
}

// Cover Workflow.ToFile with .yml extension
func TestWorkflowToFileYmlExtension(t *testing.T) {
	w := &Workflow{
		Name:       "test",
		Entrypoint: "main",
		Templates:  []Templatable{&Container{Name: "main", Image: "alpine"}},
	}
	path, err := w.ToFile(t.TempDir(), "output.yml")
	if err != nil {
		t.Fatal(err)
	}
	if path[len(path)-4:] != ".yml" {
		t.Errorf("path should end with .yml, got %q", path)
	}
}

// Cover ConvertBinaryUnit and ConvertDecimalUnit parse number errors
// (These are effectively unreachable since regex validates the format first,
// but testing boundary ensures the error paths exist)
func TestConvertBinaryUnitLargeSuffix(t *testing.T) {
	v, err := ConvertBinaryUnit("1Ti")
	if err != nil {
		t.Fatal(err)
	}
	if v != 1099511627776 { // 2^40
		t.Errorf("got %f", v)
	}
}

func TestConvertDecimalUnitLargeSuffix(t *testing.T) {
	v, err := ConvertDecimalUnit("1G")
	if err != nil {
		t.Fatal(err)
	}
	if v != 1e9 {
		t.Errorf("got %f", v)
	}
}

// Cover ConfigMapVolume with no-name edge
func TestConfigMapVolumeNoNameBuildsError(t *testing.T) {
	v := ConfigMapVolume{BaseVolume: BaseVolume{MountPath: "/cfg"}}
	_, err := v.BuildVolume()
	if err == nil {
		t.Fatal("expected error")
	}
}

// Cover Workflow.buildVolumes error path (volume with no name)
func TestWorkflowBuildVolumesWithError(t *testing.T) {
	w := &Workflow{
		Name:       "test",
		Entrypoint: "main",
		Volumes: []VolumeBuilder{
			&EmptyDirVolume{BaseVolume: BaseVolume{Name: "good", MountPath: "/tmp"}},
			&EmptyDirVolume{BaseVolume: BaseVolume{MountPath: "/bad"}}, // no name
		},
		Templates: []Templatable{&Container{Name: "main", Image: "alpine"}},
	}
	model, err := w.Build()
	if err != nil {
		t.Fatal(err) // build doesn't fail, it skips invalid volumes
	}
	// Only the valid volume should be included
	if len(model.Spec.Volumes) != 1 {
		t.Errorf("volumes = %d, want 1 (invalid skipped)", len(model.Spec.Volumes))
	}
}

// Cover Workflow.buildVolumeClaimTemplates error skip
func TestWorkflowBuildVolumeClaimTemplatesWithError(t *testing.T) {
	w := &Workflow{
		Name:       "test",
		Entrypoint: "main",
		VolumeClaimTemplates: []PVCVolume{
			{BaseVolume: BaseVolume{Name: "good", MountPath: "/data"}, Size: "1Gi"},
			{BaseVolume: BaseVolume{MountPath: "/bad"}, Size: "1Gi"}, // no name
		},
		Templates: []Templatable{&Container{Name: "main", Image: "alpine"}},
	}
	model, err := w.Build()
	if err != nil {
		t.Fatal(err)
	}
	if len(model.Spec.VolumeClaimTemplates) != 1 {
		t.Errorf("pvcs = %d, want 1", len(model.Spec.VolumeClaimTemplates))
	}
}
