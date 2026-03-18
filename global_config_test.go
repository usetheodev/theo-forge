package forge

import "testing"

func TestGlobalConfigDefaults(t *testing.T) {
	cfg := GetGlobalConfig()
	defer cfg.Reset()

	if cfg.GetImage() != "python:3.11" {
		t.Errorf("default image = %q, want 'python:3.11'", cfg.GetImage())
	}
	if !cfg.VerifySSL {
		t.Error("default VerifySSL should be true")
	}
}

func TestGlobalConfigSetImage(t *testing.T) {
	cfg := GetGlobalConfig()
	defer cfg.Reset()

	cfg.SetImage("alpine:3.18")
	if cfg.GetImage() != "alpine:3.18" {
		t.Errorf("image = %q, want 'alpine:3.18'", cfg.GetImage())
	}
}

func TestGlobalConfigSetNamespace(t *testing.T) {
	cfg := GetGlobalConfig()
	defer cfg.Reset()

	cfg.SetNamespace("workflows")
	if cfg.GetNamespace() != "workflows" {
		t.Errorf("namespace = %q", cfg.GetNamespace())
	}
}

func TestGlobalConfigTemplateHook(t *testing.T) {
	cfg := GetGlobalConfig()
	defer cfg.Reset()

	// Register hook that sets a default image
	cfg.RegisterTemplateHook(func(tpl *TemplateModel) {
		if tpl.Container != nil && tpl.Container.Image == "" {
			tpl.Container.Image = "default-image:latest"
		}
	})

	// Build a container with no image
	tpl := &TemplateModel{
		Name:      "test",
		Container: &ContainerModel{Image: ""},
	}
	cfg.DispatchTemplateHooks(tpl)

	if tpl.Container.Image != "default-image:latest" {
		t.Errorf("image after hook = %q", tpl.Container.Image)
	}
}

func TestGlobalConfigWorkflowHook(t *testing.T) {
	cfg := GetGlobalConfig()
	defer cfg.Reset()

	// Register hook that adds a label
	cfg.RegisterWorkflowHook(func(wf *WorkflowModel) {
		if wf.Metadata.Labels == nil {
			wf.Metadata.Labels = make(map[string]string)
		}
		wf.Metadata.Labels["managed-by"] = "forge"
	})

	wf := &WorkflowModel{
		Metadata: WorkflowMetadata{Name: "test"},
	}
	cfg.DispatchWorkflowHooks(wf)

	if wf.Metadata.Labels["managed-by"] != "forge" {
		t.Errorf("label = %q", wf.Metadata.Labels["managed-by"])
	}
}

func TestGlobalConfigMultipleHooks(t *testing.T) {
	cfg := GetGlobalConfig()
	defer cfg.Reset()

	callOrder := []string{}
	cfg.RegisterTemplateHook(func(tpl *TemplateModel) {
		callOrder = append(callOrder, "first")
	})
	cfg.RegisterTemplateHook(func(tpl *TemplateModel) {
		callOrder = append(callOrder, "second")
	})

	tpl := &TemplateModel{Name: "test"}
	cfg.DispatchTemplateHooks(tpl)

	if len(callOrder) != 2 {
		t.Fatalf("hooks called = %d, want 2", len(callOrder))
	}
	if callOrder[0] != "first" || callOrder[1] != "second" {
		t.Errorf("order = %v, want [first, second]", callOrder)
	}
}

func TestGlobalConfigClearHooks(t *testing.T) {
	cfg := GetGlobalConfig()
	defer cfg.Reset()

	called := false
	cfg.RegisterTemplateHook(func(tpl *TemplateModel) {
		called = true
	})
	cfg.ClearHooks()

	tpl := &TemplateModel{Name: "test"}
	cfg.DispatchTemplateHooks(tpl)

	if called {
		t.Error("hook should not be called after ClearHooks")
	}
}

func TestGlobalConfigReset(t *testing.T) {
	cfg := GetGlobalConfig()

	cfg.SetImage("custom:v1")
	cfg.SetNamespace("custom-ns")
	cfg.SetHost("https://custom.host")
	cfg.SetToken("secret")

	cfg.Reset()

	if cfg.GetImage() != "python:3.11" {
		t.Errorf("image after reset = %q", cfg.GetImage())
	}
	if cfg.GetNamespace() != "" {
		t.Errorf("namespace after reset = %q", cfg.GetNamespace())
	}
	if cfg.Host != "" {
		t.Errorf("host after reset = %q", cfg.Host)
	}
	if cfg.Token != "" {
		t.Errorf("token after reset = %q", cfg.Token)
	}
}
