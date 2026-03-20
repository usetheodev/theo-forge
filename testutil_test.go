package forge

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/usetheodev/theo-forge/serialize"
	yamlconv "sigs.k8s.io/yaml"
)

// --- Pointer helpers ---

func ptrStr(s string) *string   { return &s }
func ptrInt(i int) *int         { return &i }
func ptrBool(b bool) *bool      { return &b }

// --- String helpers ---

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsStr(s, sub))
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}

// --- YAML comparison helpers ---

// normalizeForComparison recursively normalizes a value for comparison.
// Arrays of objects with a "name" field are sorted by name to make comparison order-independent.
func normalizeForComparison(v interface{}) interface{} {
	switch val := v.(type) {
	case map[string]interface{}:
		result := make(map[string]interface{}, len(val))
		for k, v2 := range val {
			result[k] = normalizeForComparison(v2)
		}
		return result
	case []interface{}:
		normalized := make([]interface{}, len(val))
		for i, item := range val {
			normalized[i] = normalizeForComparison(item)
		}
		// Sort arrays of objects with "name" key by the "name" value
		if len(normalized) > 0 {
			if _, ok := normalized[0].(map[string]interface{}); ok {
				allHaveName := true
				for _, item := range normalized {
					m, ok := item.(map[string]interface{})
					if !ok {
						allHaveName = false
						break
					}
					if _, has := m["name"]; !has {
						allHaveName = false
						break
					}
				}
				if allHaveName {
					sorted := make([]interface{}, len(normalized))
					copy(sorted, normalized)
					for i := 0; i < len(sorted); i++ {
						for j := i + 1; j < len(sorted); j++ {
							nameI := fmt.Sprint(sorted[i].(map[string]interface{})["name"])
							nameJ := fmt.Sprint(sorted[j].(map[string]interface{})["name"])
							if nameI > nameJ {
								sorted[i], sorted[j] = sorted[j], sorted[i]
							}
						}
					}
					return sorted
				}
			}
		}
		return normalized
	default:
		return v
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

// assertSemantic compares two YAML strings semantically.
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

	cleanMap(gotMap)
	cleanMap(wantMap)

	gotNorm := normalizeForComparison(gotMap)
	wantNorm := normalizeForComparison(wantMap)

	gotBytes, _ := json.MarshalIndent(gotNorm, "", "  ")
	wantBytes, _ := json.MarshalIndent(wantNorm, "", "  ")

	if string(gotBytes) != string(wantBytes) {
		diffs := findDiffs("", gotNorm, wantNorm)
		if len(diffs) > 0 {
			t.Errorf("round-trip mismatch for %s:\n%s", name, strings.Join(diffs, "\n"))
		}
	}
}

// --- Golden file helpers ---

var updateGolden = flag.Bool("update-golden", false, "update golden test files")

func goldenTest(t *testing.T, name string, got string) {
	t.Helper()

	goldenPath := filepath.Join("testdata", name+".yaml")

	if *updateGolden {
		if err := os.WriteFile(goldenPath, []byte(got), 0o644); err != nil {
			t.Fatalf("write golden file: %v", err)
		}
		return
	}

	expected, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("read golden file %s: %v (run with -update-golden to create)", goldenPath, err)
	}

	if got != string(expected) {
		t.Errorf("YAML output does not match golden file %s\n\nGot:\n%s\n\nExpected:\n%s", goldenPath, got, string(expected))
	}
}

// --- Round-trip helpers ---

func detectKind(yamlStr string) string {
	for _, line := range strings.Split(yamlStr, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "kind:") {
			kind := strings.TrimSpace(strings.TrimPrefix(line, "kind:"))
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

	m, err := serialize.WorkflowFromYAML(yamlStr)
	if err != nil {
		t.Fatalf("parse %s: %v", name, err)
	}

	gotYAML, err := serialize.WorkflowToYAML(m)
	if err != nil {
		t.Fatalf("serialize %s: %v", name, err)
	}

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
