package forge

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/usetheo/theo/forge/model"
	"github.com/usetheo/theo/forge/serialize"
	yamlconv "sigs.k8s.io/yaml"
)

// TestRoundTripAllExamples verifies that every Hera-generated YAML example can be
// parsed into a model and re-serialized without data loss.
// This proves the forge model types can represent ALL Argo Workflow examples programmatically.
func TestRoundTripAllExamples(t *testing.T) {
	heraDir := "hera/examples/workflows/upstream"

	entries, err := os.ReadDir(heraDir)
	if err != nil {
		t.Fatalf("read hera examples dir: %v", err)
	}

	var tested int
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
			continue
		}
		// Skip upstream YAML files (they're the original Argo examples, not Hera-generated)
		if strings.HasSuffix(entry.Name(), ".upstream.yaml") {
			continue
		}

		name := strings.TrimSuffix(entry.Name(), ".yaml")
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(heraDir, entry.Name())
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read %s: %v", path, err)
			}
			yamlStr := string(data)

			// Determine kind
			kind := detectKind(yamlStr)

			switch kind {
			case "Workflow":
				roundTripWorkflow(t, name, yamlStr)
			case "WorkflowTemplate", "ClusterWorkflowTemplate":
				roundTripWorkflowTemplate(t, name, yamlStr)
			case "CronWorkflow":
				roundTripCronWorkflow(t, name, yamlStr)
			default:
				t.Skipf("unknown kind %q in %s", kind, name)
			}
		})
		tested++
	}

	if tested == 0 {
		t.Fatal("no examples found")
	}
	t.Logf("Round-trip tested %d examples", tested)
}

func detectKind(yamlStr string) string {
	for _, line := range strings.Split(yamlStr, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "kind:") {
			kind := strings.TrimSpace(strings.TrimPrefix(line, "kind:"))
			// Strip inline comments (e.g., "Workflow  # comment")
			if idx := strings.Index(kind, "#"); idx > 0 {
				kind = strings.TrimSpace(kind[:idx])
			}
			return kind
		}
	}
	return ""
}

func roundTripWorkflow(t *testing.T, name, yamlStr string) {
	t.Helper()

	// Parse original YAML to model
	m, err := serialize.WorkflowFromYAML(yamlStr)
	if err != nil {
		t.Fatalf("parse %s: %v", name, err)
	}

	// Re-serialize to YAML
	gotYAML, err := serialize.WorkflowToYAML(m)
	if err != nil {
		t.Fatalf("serialize %s: %v", name, err)
	}

	// Compare semantically
	assertSemantic(t, name, gotYAML, yamlStr)
}

func roundTripWorkflowTemplate(t *testing.T, name, yamlStr string) {
	t.Helper()

	m, err := serialize.WorkflowTemplateFromYAML(yamlStr)
	if err != nil {
		t.Fatalf("parse %s: %v", name, err)
	}

	gotYAML, err := serialize.WorkflowTemplateToYAML(m)
	if err != nil {
		t.Fatalf("serialize %s: %v", name, err)
	}

	assertSemantic(t, name, gotYAML, yamlStr)
}

func roundTripCronWorkflow(t *testing.T, name, yamlStr string) {
	t.Helper()

	m, err := serialize.CronWorkflowFromYAML(yamlStr)
	if err != nil {
		t.Fatalf("parse %s: %v", name, err)
	}

	gotYAML, err := serialize.CronWorkflowToYAML(m)
	if err != nil {
		t.Fatalf("serialize %s: %v", name, err)
	}

	assertSemantic(t, name, gotYAML, yamlStr)
}

func assertSemantic(t *testing.T, name, got, want string) {
	t.Helper()

	gotJSON, err := yamlconv.YAMLToJSON([]byte(got))
	if err != nil {
		t.Fatalf("convert got to JSON for %s: %v", name, err)
	}
	wantJSON, err := yamlconv.YAMLToJSON([]byte(want))
	if err != nil {
		t.Fatalf("convert want to JSON for %s: %v", name, err)
	}

	var gotMap, wantMap map[string]interface{}
	json.Unmarshal(gotJSON, &gotMap)
	json.Unmarshal(wantJSON, &wantMap)

	// Remove null/empty fields from both sides for comparison
	cleanMap(gotMap)
	cleanMap(wantMap)

	gotNorm := normalizeForComparison(gotMap)
	wantNorm := normalizeForComparison(wantMap)

	gotBytes, _ := json.MarshalIndent(gotNorm, "", "  ")
	wantBytes, _ := json.MarshalIndent(wantNorm, "", "  ")

	if string(gotBytes) != string(wantBytes) {
		// Find differences
		diffs := findDiffs("", gotNorm, wantNorm)
		if len(diffs) > 0 {
			t.Errorf("round-trip mismatch for %s:\n%s", name, strings.Join(diffs, "\n"))
		}
	}
}

// cleanMap removes null values and empty maps/slices recursively.
func cleanMap(m map[string]interface{}) {
	for k, v := range m {
		switch val := v.(type) {
		case nil:
			delete(m, k)
		case map[string]interface{}:
			cleanMap(val)
			if len(val) == 0 {
				delete(m, k)
			}
		case []interface{}:
			if len(val) == 0 {
				delete(m, k)
			} else {
				for _, item := range val {
					if sub, ok := item.(map[string]interface{}); ok {
						cleanMap(sub)
					}
				}
			}
		}
	}
}

// findDiffs finds differences between two normalized structures.
func findDiffs(path string, got, want interface{}) []string {
	var diffs []string

	switch g := got.(type) {
	case map[string]interface{}:
		w, ok := want.(map[string]interface{})
		if !ok {
			return []string{fmt.Sprintf("  %s: type mismatch (got map, want %T)", path, want)}
		}
		// Check missing and different keys
		for k, gv := range g {
			wv, exists := w[k]
			if !exists {
				diffs = append(diffs, fmt.Sprintf("  %s.%s: extra field in got", path, k))
				continue
			}
			diffs = append(diffs, findDiffs(path+"."+k, gv, wv)...)
		}
		for k := range w {
			if _, exists := g[k]; !exists {
				diffs = append(diffs, fmt.Sprintf("  %s.%s: missing field in got", path, k))
			}
		}
	case []interface{}:
		w, ok := want.([]interface{})
		if !ok {
			return []string{fmt.Sprintf("  %s: type mismatch (got array, want %T)", path, want)}
		}
		if len(g) != len(w) {
			return []string{fmt.Sprintf("  %s: array length mismatch (got %d, want %d)", path, len(g), len(w))}
		}
		for i := range g {
			diffs = append(diffs, findDiffs(fmt.Sprintf("%s[%d]", path, i), g[i], w[i])...)
		}
	default:
		if fmt.Sprint(got) != fmt.Sprint(want) {
			diffs = append(diffs, fmt.Sprintf("  %s: got %v, want %v", path, got, want))
		}
	}

	return diffs
}

// TestRoundTripTestdataExamples verifies that every YAML example in testdata/examples/
// (the upstream Argo Workflow examples) can be parsed into a model and re-serialized without data loss.
func TestRoundTripTestdataExamples(t *testing.T) {
	examplesDir := "testdata/examples"

	var tested int
	err := filepath.Walk(examplesDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(info.Name(), ".yaml") {
			return nil
		}

		// Get relative name for test identification
		relPath, _ := filepath.Rel(examplesDir, path)
		name := strings.TrimSuffix(relPath, ".yaml")

		t.Run(name, func(t *testing.T) {
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("read %s: %v", path, err)
			}
			yamlStr := string(data)

			kind := detectKind(yamlStr)

			switch kind {
			case "Workflow":
				roundTripWorkflow(t, name, yamlStr)
			case "WorkflowTemplate", "ClusterWorkflowTemplate":
				roundTripWorkflowTemplate(t, name, yamlStr)
			case "CronWorkflow":
				roundTripCronWorkflow(t, name, yamlStr)
			default:
				t.Skipf("unknown kind %q in %s", kind, name)
			}
		})
		tested++
		return nil
	})
	if err != nil {
		t.Fatalf("walk testdata/examples: %v", err)
	}

	if tested == 0 {
		t.Fatal("no examples found in testdata/examples/")
	}
	t.Logf("Round-trip tested %d testdata examples", tested)
}

// TestRoundTripWorkflowBuilder verifies that workflows built programmatically
// produce valid models that round-trip cleanly.
func TestRoundTripWorkflowBuilder(t *testing.T) {
	builders := map[string]func() (*Workflow, error){
		"hello-world": func() (*Workflow, error) { return buildHelloWorld(), nil },
		"steps":       func() (*Workflow, error) { return buildSteps(), nil },
		"dag-diamond": func() (*Workflow, error) { return buildDagDiamond(), nil },
	}

	for name, builder := range builders {
		t.Run(name, func(t *testing.T) {
			w, err := builder()
			if err != nil {
				t.Fatalf("build %s: %v", name, err)
			}

			// Build to model
			m, err := w.Build()
			if err != nil {
				t.Fatalf("build model %s: %v", name, err)
			}

			// Serialize
			yaml1, err := serialize.WorkflowToYAML(m)
			if err != nil {
				t.Fatalf("serialize %s: %v", name, err)
			}

			// Parse back
			m2, err := serialize.WorkflowFromYAML(yaml1)
			if err != nil {
				t.Fatalf("parse back %s: %v", name, err)
			}

			// Re-serialize
			yaml2, err := serialize.WorkflowToYAML(m2)
			if err != nil {
				t.Fatalf("re-serialize %s: %v", name, err)
			}

			// Should be identical
			if yaml1 != yaml2 {
				t.Errorf("round-trip not stable for %s", name)
			}
		})
	}
}

// Ensure CronWorkflowFromYAML exists
func init() {
	// Verify serialize package has the necessary functions
	_ = serialize.WorkflowFromYAML
	_ = serialize.WorkflowToYAML
	_ = serialize.WorkflowTemplateFromYAML
	_ = serialize.WorkflowTemplateToYAML
	_ = serialize.CronWorkflowFromYAML
	_ = serialize.CronWorkflowToYAML
}

// Additional test to verify we can programmatically create a workflow using model types
// directly for any feature.
func TestModelDirectConstruction(t *testing.T) {
	// Build a workflow with synchronization, hooks, memoize, etc.
	m := model.WorkflowModel{
		APIVersion: DefaultAPIVersion,
		Kind:       "Workflow",
		Metadata:   model.WorkflowMetadata{GenerateName: "feature-rich-"},
		Spec: model.WorkflowSpec{
			Entrypoint: "main",
			Synchronization: &model.SynchronizationModel{
				Mutex: &model.MutexModel{Name: "test-mutex"},
			},
			Hooks: map[string]model.LifecycleHook{
				"exit": {Template: "exit-handler"},
			},
			PodSpecPatch: `{"containers":[{"name":"main","resources":{"limits":{"cpu":"1"}}}]}`,
			Templates: []model.TemplateModel{
				{
					Name: "main",
					Container: &model.ContainerModel{
						Image:   "alpine:3.18",
						Command: []string{"echo", "hello"},
					},
					Memoize: &model.MemoizeModel{
						Key:    "{{inputs.parameters.msg}}",
						MaxAge: "1h",
						Cache: &model.CacheModel{
							ConfigMap: &model.ConfigMapKeyRef{
								Name: "my-cache",
								Key:  "data",
							},
						},
					},
					Daemon: ptrBool(true),
					Synchronization: &model.SynchronizationModel{
						Semaphore: &model.SemaphoreModel{
							ConfigMapKeyRef: &model.ConfigMapKeyRef{
								Name: "semaphore-config",
								Key:  "workflow",
							},
						},
					},
				},
				{
					Name: "exit-handler",
					Container: &model.ContainerModel{
						Image:   "alpine:3.18",
						Command: []string{"echo", "done"},
					},
				},
			},
		},
	}

	yamlStr, err := serialize.WorkflowToYAML(m)
	if err != nil {
		t.Fatal(err)
	}

	// Verify key features are in the YAML
	checks := []string{
		"synchronization:", "mutex:", "test-mutex",
		"hooks:", "exit:", "exit-handler",
		"podSpecPatch:", "memoize:", "daemon:",
		"semaphore:", "configMapKeyRef:",
	}
	for _, check := range checks {
		if !strings.Contains(yamlStr, check) {
			t.Errorf("YAML missing %q", check)
		}
	}
}
