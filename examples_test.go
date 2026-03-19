package forge

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/usetheo/theo/forge/model"
	yamlconv "sigs.k8s.io/yaml"
)

// helpers are defined in other test files (ptrStr, ptrInt, ptrBool)

// yamlToMap parses YAML string into a generic map for comparison.
func yamlToMap(yamlStr string) (map[string]interface{}, error) {
	jsonBytes, err := yamlconv.YAMLToJSON([]byte(yamlStr))
	if err != nil {
		return nil, err
	}
	var m map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &m); err != nil {
		return nil, err
	}
	return m, nil
}

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
					// Sort by name
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

// semanticEqual compares two YAML strings semantically (ignoring field order and template ordering).
func semanticEqual(got, want string) (bool, error) {
	gotMap, err := yamlToMap(got)
	if err != nil {
		return false, fmt.Errorf("parsing got: %w", err)
	}
	wantMap, err := yamlToMap(want)
	if err != nil {
		return false, fmt.Errorf("parsing want: %w", err)
	}
	gotNorm := normalizeForComparison(gotMap)
	wantNorm := normalizeForComparison(wantMap)
	return reflect.DeepEqual(gotNorm, wantNorm), nil
}

// readExpectedYAML reads the Hera-generated YAML for comparison.
func readExpectedYAML(name string) (string, error) {
	path := fmt.Sprintf("hera/examples/workflows/upstream/%s.yaml", name)
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// assertYAMLEqual compares the built YAML with the expected Hera YAML.
func assertYAMLEqual(t *testing.T, name string, gotYAML string) {
	t.Helper()
	wantYAML, err := readExpectedYAML(name)
	if err != nil {
		t.Fatalf("read expected YAML for %s: %v", name, err)
	}
	equal, err := semanticEqual(gotYAML, wantYAML)
	if err != nil {
		t.Fatalf("compare YAML for %s: %v", name, err)
	}
	if !equal {
		t.Errorf("YAML mismatch for %s\n\nGot:\n%s\n\nWant:\n%s", name, gotYAML, wantYAML)
	}
}

// === BUILDERS ===

func buildHelloWorld() *Workflow {
	return &Workflow{
		GenerateName: "hello-world-",
		Labels:       map[string]string{"workflows.argoproj.io/archive-strategy": "false"},
		Annotations:  map[string]string{"workflows.argoproj.io/description": "This is a simple hello world example.\n"},
		Entrypoint:   "hello-world",
		Templates: []Templatable{
			&Container{
				Name:    "hello-world",
				Image:   "busybox",
				Command: []string{"echo"},
				Args:    []string{"hello world"},
			},
		},
	}
}

func buildSteps() *Workflow {
	return &Workflow{
		GenerateName: "steps-",
		Entrypoint:   "hello-hello-hello",
		Templates: []Templatable{
			func() Templatable {
				s := &Steps{Name: "hello-hello-hello"}
				s.AddSequentialStep(&Step{
					Name:     "hello1",
					Template: "print-message",
					Arguments: []Parameter{{Name: "message", Value: ptrStr("hello1")}},
				})
				s.AddParallelGroup(
					&Step{
						Name:     "hello2a",
						Template: "print-message",
						Arguments: []Parameter{{Name: "message", Value: ptrStr("hello2a")}},
					},
					&Step{
						Name:     "hello2b",
						Template: "print-message",
						Arguments: []Parameter{{Name: "message", Value: ptrStr("hello2b")}},
					},
				)
				return s
			}(),
			&Container{
				Name:    "print-message",
				Image:   "busybox",
				Command: []string{"echo"},
				Args:    []string{"{{inputs.parameters.message}}"},
				Inputs:  []Parameter{{Name: "message"}},
			},
		},
	}
}

func buildDagDiamond() *Workflow {
	dag := &DAG{Name: "diamond"}
	A := &Task{Name: "A", Template: "echo", Arguments: []Parameter{{Name: "message", Value: ptrStr("A")}}}
	B := &Task{Name: "B", Template: "echo", Depends: "A", Arguments: []Parameter{{Name: "message", Value: ptrStr("B")}}}
	C := &Task{Name: "C", Template: "echo", Depends: "A", Arguments: []Parameter{{Name: "message", Value: ptrStr("C")}}}
	D := &Task{Name: "D", Template: "echo", Depends: "B && C", Arguments: []Parameter{{Name: "message", Value: ptrStr("D")}}}
	dag.AddTasks(A, B, C, D)

	return &Workflow{
		GenerateName: "dag-diamond-",
		Entrypoint:   "diamond",
		Templates: []Templatable{
			dag,
			&Container{
				Name:    "echo",
				Image:   "alpine:3.7",
				Command: []string{"echo", "{{inputs.parameters.message}}"},
				Inputs:  []Parameter{{Name: "message"}},
			},
		},
	}
}

func buildArgumentsParameters() *Workflow {
	return &Workflow{
		GenerateName: "arguments-parameters-",
		Entrypoint:   "print-message",
		Arguments:    []Parameter{{Name: "message", Value: ptrStr("hello world")}},
		Templates: []Templatable{
			&Container{
				Name:    "print-message",
				Image:   "busybox",
				Command: []string{"echo"},
				Args:    []string{"{{inputs.parameters.message}}"},
				Inputs:  []Parameter{{Name: "message"}},
			},
		},
	}
}

func buildCoinflip() *Workflow {
	return &Workflow{
		GenerateName: "coinflip-",
		Annotations:  map[string]string{"workflows.argoproj.io/description": "This is an example of coin flip defined as a sequence of conditional steps."},
		Entrypoint:   "coinflip",
		Templates: []Templatable{
			&Script{
				Name:    "flip-coin",
				Image:   "python:alpine3.6",
				Command: []string{"python"},
				Source:  "import random\nresult = 'heads' if random.randint(0, 1) == 0 else 'tails'\nprint(result)",
			},
			&Container{
				Name:    "heads",
				Image:   "alpine:3.6",
				Command: []string{"sh", "-c"},
				Args:    []string{"echo \"it was heads\""},
			},
			&Container{
				Name:    "tails",
				Image:   "alpine:3.6",
				Command: []string{"sh", "-c"},
				Args:    []string{"echo \"it was tails\""},
			},
			func() Templatable {
				s := &Steps{Name: "coinflip"}
				s.AddSequentialStep(&Step{Name: "flip-coin", Template: "flip-coin"})
				s.AddParallelGroup(
					&Step{Name: "heads", Template: "heads", When: "{{steps.flip-coin.outputs.result}} == heads"},
					&Step{Name: "tails", Template: "tails", When: "{{steps.flip-coin.outputs.result}} == tails"},
				)
				return s
			}(),
		},
	}
}

func buildConditionals() *Workflow {
	return &Workflow{
		GenerateName: "conditional-",
		Entrypoint:   "conditional-example",
		Arguments:    []Parameter{{Name: "should-print", Value: ptrStr("true")}},
		Templates: []Templatable{
			func() Templatable {
				s := &Steps{
					Name:   "conditional-example",
					Inputs: []Parameter{{Name: "should-print"}},
				}
				s.AddParallelGroup(
					&Step{Name: "print-hello-govaluate", Template: "argosay", When: "{{inputs.parameters.should-print}} == true"},
					&Step{Name: "print-hello-expr", Template: "argosay", When: "{{= inputs.parameters[\"should-print\"] == 'true'}}"},
					&Step{Name: "print-hello-expr-json", Template: "argosay", When: "{{=jsonpath(workflow.parameters.json, '$[0].value') == 'true'}}"},
				)
				return s
			}(),
			&Container{
				Name:    "argosay",
				Image:   "argoproj/argosay:v1",
				Command: []string{"sh", "-c"},
				Args:    []string{"cowsay hello"},
			},
		},
	}
}

func buildScriptsPython() *Workflow {
	return &Workflow{
		GenerateName: "scripts-python-",
		Entrypoint:   "python-script-example",
		Templates: []Templatable{
			&Script{
				Name:    "gen-random-int",
				Image:   "python:alpine3.6",
				Command: []string{"python"},
				Source:  "\nimport random\ni = random.randint(1, 100)\nprint(i)",
			},
			&Container{
				Name:    "print-message",
				Image:   "alpine:latest",
				Command: []string{"sh", "-c"},
				Args:    []string{"echo result was: {{inputs.parameters.message}}"},
				Inputs:  []Parameter{{Name: "message"}},
			},
			func() Templatable {
				s := &Steps{Name: "python-script-example"}
				s.AddSequentialStep(&Step{Name: "generate", Template: "gen-random-int"})
				s.AddSequentialStep(&Step{
					Name:     "print",
					Template: "print-message",
					Arguments: []Parameter{{Name: "message", Value: ptrStr("{{steps.generate.outputs.result}}")}},
				})
				return s
			}(),
		},
	}
}

func buildColoredLogs() *Workflow {
	return &Workflow{
		GenerateName: "colored-logs-",
		Entrypoint:   "colored-logs-example",
		Annotations: map[string]string{
			"workflows.argoproj.io/description": "This example demonstrates colored logs of the logs viewer on Argo UI.\n",
		},
		Templates: []Templatable{
			&Container{
				Name:    "colored-logs-example",
				Image:   "python:3.7",
				Command: []string{"python", "-c"},
				Args:    []string{"from datetime import datetime; print(datetime.utcnow().isoformat())"},
			},
		},
	}
}

func buildLoops() *Workflow {
	return &Workflow{
		GenerateName: "loops-",
		Entrypoint:   "loop-example",
		Templates: []Templatable{
			func() Templatable {
				s := &Steps{Name: "loop-example"}
				s.AddSequentialStep(&Step{
					Name:     "print-message-loop",
					Template: "print-message",
					Arguments: []Parameter{{Name: "message", Value: ptrStr("{{item}}")}},
					WithItems: []interface{}{"hello world", "goodbye world"},
				})
				return s
			}(),
			&Container{
				Name:    "print-message",
				Image:   "busybox",
				Command: []string{"echo"},
				Args:    []string{"{{inputs.parameters.message}}"},
				Inputs:  []Parameter{{Name: "message"}},
			},
		},
	}
}

func buildLoopsMaps() *Workflow {
	return &Workflow{
		GenerateName: "loops-maps-",
		Entrypoint:   "loop-map-example",
		Templates: []Templatable{
			func() Templatable {
				s := &Steps{Name: "loop-map-example"}
				s.AddSequentialStep(&Step{
					Name:     "test-linux",
					Template: "cat-os-release",
					Arguments: []Parameter{
						{Name: "image", Value: ptrStr("{{item.image}}")},
						{Name: "tag", Value: ptrStr("{{item.tag}}")},
					},
					WithItems: []interface{}{
						map[string]interface{}{"image": "debian", "tag": "9.1"},
						map[string]interface{}{"image": "debian", "tag": "8.9"},
						map[string]interface{}{"image": "alpine", "tag": "3.6"},
						map[string]interface{}{"image": "ubuntu", "tag": "17.10"},
					},
				})
				return s
			}(),
			&Container{
				Name:    "cat-os-release",
				Image:   "{{inputs.parameters.image}}:{{inputs.parameters.tag}}",
				Command: []string{"cat"},
				Args:    []string{"/etc/os-release"},
				Inputs: []Parameter{
					{Name: "image"},
					{Name: "tag"},
				},
			},
		},
	}
}

func buildOutputParameter() *Workflow {
	return &Workflow{
		GenerateName: "output-parameter-",
		Entrypoint:   "output-parameter",
		Templates: []Templatable{
			&Container{
				Name:    "hello-world-to-file",
				Image:   "busybox",
				Command: []string{"sh", "-c"},
				Args:    []string{"sleep 1; echo -n hello world > /tmp/hello_world.txt"},
				Outputs: []Parameter{{Name: "hello-param", ValueFrom: &ValueFrom{Path: "/tmp/hello_world.txt", Default: ptrStr("Foobar")}}},
			},
			&Container{
				Name:    "print-message",
				Image:   "busybox",
				Command: []string{"echo"},
				Args:    []string{"{{inputs.parameters.message}}"},
				Inputs:  []Parameter{{Name: "message"}},
			},
			func() Templatable {
				s := &Steps{Name: "output-parameter"}
				s.AddSequentialStep(&Step{Name: "generate-parameter", Template: "hello-world-to-file"})
				s.AddSequentialStep(&Step{
					Name:     "consume-parameter",
					Template: "print-message",
					Arguments: []Parameter{{Name: "message", Value: ptrStr("{{steps.generate-parameter.outputs.parameters.hello-param}}")}},
				})
				return s
			}(),
		},
	}
}

func buildGlobalParameters() *Workflow {
	return &Workflow{
		GenerateName: "global-parameters-",
		Entrypoint:   "print-message",
		Arguments:    []Parameter{{Name: "message", Value: ptrStr("hello world")}},
		Templates: []Templatable{
			&Container{
				Name:    "print-message",
				Image:   "busybox",
				Command: []string{"echo"},
				Args:    []string{"{{workflow.parameters.message}}"},
			},
		},
	}
}

func buildExitHandlers() *Workflow {
	return &Workflow{
		GenerateName: "exit-handlers-",
		Entrypoint:   "intentional-fail",
		OnExit:       "exit-handler",
		Templates: []Templatable{
			&Container{
				Name:    "intentional-fail",
				Image:   "alpine:latest",
				Command: []string{"sh", "-c"},
				Args:    []string{"echo intentional failure; exit 1"},
			},
			func() Templatable {
				s := &Steps{Name: "exit-handler"}
				s.AddParallelGroup(
					&Step{Name: "notify", Template: "send-email"},
					&Step{Name: "celebrate", Template: "celebrate", When: "{{workflow.status}} == Succeeded"},
					&Step{Name: "cry", Template: "cry", When: "{{workflow.status}} != Succeeded"},
				)
				return s
			}(),
			&Container{
				Name:    "send-email",
				Image:   "alpine:latest",
				Command: []string{"sh", "-c"},
				Args:    []string{"echo send e-mail: {{workflow.name}} {{workflow.status}} {{workflow.duration}}. Failed steps {{workflow.failures}}"},
			},
			&Container{
				Name:    "celebrate",
				Image:   "alpine:latest",
				Command: []string{"sh", "-c"},
				Args:    []string{"echo hooray!"},
			},
			&Container{
				Name:    "cry",
				Image:   "alpine:latest",
				Command: []string{"sh", "-c"},
				Args:    []string{"echo boohoo!"},
			},
		},
	}
}

func buildForever() *Workflow {
	return &Workflow{
		Name:       "forever",
		Entrypoint: "main",
		Templates: []Templatable{
			&Container{
				Name:    "main",
				Image:   "busybox",
				Command: []string{"sh", "-c", "for I in $(seq 1 1000) ; do echo $I ; sleep 1s; done"},
			},
		},
	}
}

func buildArtifactPassing() *Workflow {
	return &Workflow{
		GenerateName: "artifact-passing-",
		Entrypoint:   "artifact-example",
		Templates: []Templatable{
			&Container{
				Name:    "hello-world-to-file",
				Image:   "busybox",
				Command: []string{"sh", "-c"},
				Args:    []string{"sleep 1; echo hello world | tee /tmp/hello_world.txt"},
				OutputArtifacts: []ArtifactBuilder{
					&Artifact{Name: "hello-art", Path: "/tmp/hello_world.txt"},
				},
			},
			&Container{
				Name:    "print-message-from-file",
				Image:   "alpine:latest",
				Command: []string{"sh", "-c"},
				Args:    []string{"cat /tmp/message"},
				InputArtifacts: []ArtifactBuilder{
					&Artifact{Name: "message", Path: "/tmp/message"},
				},
			},
			func() Templatable {
				s := &Steps{Name: "artifact-example"}
				s.AddSequentialStep(&Step{Name: "generate-artifact", Template: "hello-world-to-file"})
				s.AddSequentialStep(&Step{
					Name:     "consume-artifact",
					Template: "print-message-from-file",
					ArgumentArtifacts: []ArtifactBuilder{
						&Artifact{Name: "message", From: "{{steps.generate-artifact.outputs.artifacts.hello-art}}"},
					},
				})
				return s
			}(),
		},
	}
}

func buildSuspendTemplate() *Workflow {
	return &Workflow{
		GenerateName: "suspend-template-",
		Entrypoint:   "suspend",
		Templates: []Templatable{
			&Container{
				Name:    "hello-world",
				Image:   "busybox",
				Command: []string{"echo"},
				Args:    []string{"hello world"},
			},
			&Suspend{Name: "approve"},
			&Suspend{Name: "delay", Duration: "20"},
			func() Templatable {
				s := &Steps{Name: "suspend"}
				s.AddSequentialStep(&Step{Name: "build", Template: "hello-world"})
				s.AddSequentialStep(&Step{Name: "approve", Template: "approve"})
				s.AddSequentialStep(&Step{Name: "delay", Template: "delay"})
				s.AddSequentialStep(&Step{Name: "release", Template: "hello-world"})
				return s
			}(),
		},
	}
}

func buildCronWorkflow() *CronWorkflow {
	return &CronWorkflow{
		Name:                       "hello-world",
		Schedules:                  []string{"* * * * *"},
		Timezone:                   "America/Los_Angeles",
		StartingDeadlineSeconds:    ptrInt(0),
		ConcurrencyPolicy:         "Replace",
		SuccessfulJobsHistoryLimit: ptrInt(4),
		FailedJobsHistoryLimit:     ptrInt(4),
		Suspend:                    ptrBool(false),
		Entrypoint:                 "hello-world-with-time",
		Templates: []Templatable{
			&Container{
				Name:    "hello-world-with-time",
				Image:   "busybox",
				Command: []string{"echo"},
				Args:    []string{"\U0001F553 hello world. Scheduled on: {{workflow.scheduledTime}}"},
			},
		},
	}
}

func buildParallelismLimit() *Workflow {
	return &Workflow{
		GenerateName: "parallelism-limit-",
		Entrypoint:   "parallelism-limit",
		Parallelism:  ptrInt(2),
		Templates: []Templatable{
			&Container{
				Name:    "sleep",
				Image:   "alpine:latest",
				Command: []string{"sh", "-c", "sleep 10"},
			},
			func() Templatable {
				s := &Steps{Name: "parallelism-limit"}
				s.AddSequentialStep(&Step{
					Name:     "sleep",
					Template: "sleep",
					WithItems: []interface{}{"this", "workflow", "should", "take", "at", "least", 60, "seconds", "to", "complete"},
				})
				return s
			}(),
		},
	}
}

func buildNodeSelector() *Workflow {
	return &Workflow{
		GenerateName: "node-selector-",
		Entrypoint:   "print-arch",
		Arguments:    []Parameter{{Name: "arch", Value: ptrStr("amd64")}},
		Templates: []Templatable{
			&Container{
				Name:         "print-arch",
				Image:        "alpine:latest",
				Command:      []string{"sh", "-c"},
				Args:         []string{"uname -a"},
				Inputs:       []Parameter{{Name: "arch"}},
				NodeSelector: map[string]string{"beta.kubernetes.io/arch": "{{inputs.parameters.arch}}"},
			},
		},
	}
}

func buildRetryBackoff() *Workflow {
	factor := 2
	return &Workflow{
		GenerateName: "retry-backoff-",
		Entrypoint:   "retry-backoff",
		Templates: []Templatable{
			&Container{
				Name:    "retry-backoff",
				Image:   "python:alpine3.6",
				Command: []string{"python", "-c"},
				Args:    []string{"import random; import sys; exit_code = random.choice([0, 1, 1]); sys.exit(exit_code)"},
				RetryStrategy: &RetryStrategy{
					Limit: ptrInt(10),
					Backoff: &Backoff{
						Duration:    "1",
						Factor:      &factor,
						MaxDuration: "1m",
						Cap:         "5",
					},
				},
			},
		},
	}
}

func buildDagEnhancedDepends() *Workflow {
	dag := &DAG{Name: "diamond"}
	dag.AddTasks(
		&Task{Name: "A", Template: "pass"},
		&Task{Name: "B", Template: "pass", Depends: "A"},
		&Task{Name: "C", Template: "fail", Depends: "A"},
		&Task{Name: "should-execute-1", Template: "pass", Depends: "A && (C.Succeeded || C.Failed)"},
		&Task{Name: "should-execute-2", Template: "pass", Depends: "B || C"},
		&Task{Name: "should-not-execute", Template: "pass", Depends: "B && C"},
		&Task{Name: "should-execute-3", Template: "pass", Depends: "should-execute-2.Succeeded || should-not-execute"},
	)

	return &Workflow{
		GenerateName: "dag-diamond-",
		Entrypoint:   "diamond",
		Templates: []Templatable{
			&Container{Name: "pass", Image: "alpine:3.7", Command: []string{"sh", "-c", "exit 0"}},
			&Container{Name: "fail", Image: "alpine:3.7", Command: []string{"sh", "-c", "exit 1"}},
			dag,
		},
	}
}

func buildDagMultiroot() *Workflow {
	dag := &DAG{Name: "multiroot"}
	A := &Task{Name: "A", Template: "echo", Arguments: []Parameter{{Name: "message", Value: ptrStr("A")}}}
	B := &Task{Name: "B", Template: "echo", Arguments: []Parameter{{Name: "message", Value: ptrStr("B")}}}
	C := &Task{Name: "C", Template: "echo", Depends: "A", Arguments: []Parameter{{Name: "message", Value: ptrStr("C")}}}
	D := &Task{Name: "D", Template: "echo", Depends: "A && B", Arguments: []Parameter{{Name: "message", Value: ptrStr("D")}}}
	dag.AddTasks(A, B, C, D)

	return &Workflow{
		GenerateName: "dag-multiroot-",
		Entrypoint:   "multiroot",
		Templates: []Templatable{
			dag,
			&Container{
				Name:    "echo",
				Image:   "alpine:3.7",
				Command: []string{"echo", "{{inputs.parameters.message}}"},
				Inputs:  []Parameter{{Name: "message"}},
			},
		},
	}
}

func buildDagTargets() *Workflow {
	dag := &DAG{
		Name:   "dag-target",
		Target: "{{workflow.parameters.target}}",
	}
	dag.AddTasks(
		&Task{Name: "A", Template: "echo", Arguments: []Parameter{{Name: "message", Value: ptrStr("A")}}},
		&Task{Name: "B", Template: "echo", Depends: "A", Arguments: []Parameter{{Name: "message", Value: ptrStr("B")}}},
		&Task{Name: "C", Template: "echo", Depends: "A", Arguments: []Parameter{{Name: "message", Value: ptrStr("C")}}},
		&Task{Name: "D", Template: "echo", Depends: "B && C", Arguments: []Parameter{{Name: "message", Value: ptrStr("D")}}},
		&Task{Name: "E", Template: "echo", Depends: "C", Arguments: []Parameter{{Name: "message", Value: ptrStr("E")}}},
	)

	return &Workflow{
		GenerateName: "dag-target-",
		Entrypoint:   "dag-target",
		Arguments:    []Parameter{{Name: "target", Value: ptrStr("E")}},
		Templates: []Templatable{
			&Container{
				Name:    "echo",
				Image:   "alpine:3.7",
				Command: []string{"echo", "{{inputs.parameters.message}}"},
				Inputs:  []Parameter{{Name: "message"}},
			},
			dag,
		},
	}
}

func buildHttpHelloWorld() *Workflow {
	return &Workflow{
		GenerateName: "http-template-",
		Annotations: map[string]string{
			"workflows.argoproj.io/description": "Http template will demostrate http template functionality\n",
			"workflows.argoproj.io/version":     ">= 3.2.0",
		},
		Labels:     map[string]string{"workflows.argoproj.io/test": "true"},
		Entrypoint: "main",
		Templates: []Templatable{
			func() Templatable {
				s := &Steps{Name: "main"}
				s.AddParallelGroup(
					&Step{
						Name: "good", Template: "http",
						Arguments: []Parameter{{Name: "url", Value: ptrStr("https://raw.githubusercontent.com/argoproj/argo-workflows/4e450e250168e6b4d51a126b784e90b11a0162bc/pkg/apis/workflow/v1alpha1/generated.swagger.json")}},
					},
					&Step{
						Name: "bad", Template: "http",
						Arguments:  []Parameter{{Name: "url", Value: ptrStr("https://raw.githubusercontent.com/argoproj/argo-workflows/thisisnotahash/pkg/apis/workflow/v1alpha1/generated.swagger.json")}},
						ContinueOn: &model.ContinueOn{Failed: true},
					},
				)
				return s
			}(),
			&HTTPTemplate{
				Name:   "http",
				URL:    "{{inputs.parameters.url}}",
				Inputs: []Parameter{{Name: "url"}},
			},
		},
	}
}

func buildVolumesEmptyDir() *Workflow {
	return &Workflow{
		GenerateName: "volumes-emptydir-",
		Entrypoint:   "volumes-emptydir-example",
		Volumes: []VolumeBuilder{
			&EmptyDirVolume{BaseVolume: BaseVolume{Name: "workdir"}},
		},
		Templates: []Templatable{
			&Container{
				Name:    "volumes-emptydir-example",
				Image:   "debian:latest",
				Command: []string{"/bin/bash", "-c"},
				Args:    []string{" vol_found=`mount | grep /mnt/vol` && if [[ -n $vol_found ]]; then echo \"Volume mounted and found\"; else echo \"Not found\"; fi "},
				VolumeMounts: []VolumeBuilder{
					&EmptyDirVolume{BaseVolume: BaseVolume{Name: "workdir", MountPath: "/mnt/vol"}},
				},
			},
		},
	}
}

func buildSecretsExample() *Workflow {
	return &Workflow{
		GenerateName: "secrets-",
		Entrypoint:   "print-secret",
		Volumes: []VolumeBuilder{
			&SecretVolume{BaseVolume: BaseVolume{Name: "my-secret-vol"}, SecretName: "my-secret"},
		},
		Templates: []Templatable{
			&Container{
				Name:    "print-secret",
				Image:   "alpine:3.7",
				Command: []string{"sh", "-c"},
				Args:    []string{" echo \"secret from env: $MYSECRETPASSWORD\"; echo \"secret from file: `cat /secret/mountpath/mypassword`\" "},
				VolumeMounts: []VolumeBuilder{
					&SecretVolume{BaseVolume: BaseVolume{Name: "my-secret-vol", MountPath: "/secret/mountpath"}},
				},
				Env: []EnvBuilder{
					SecretEnv{Name: "MYSECRETPASSWORD", SecretName: "my-secret", SecretKey: "mypassword"},
				},
			},
		},
	}
}

func buildContinueOnFail() *Workflow {
	return &Workflow{
		GenerateName: "continue-on-fail-",
		Entrypoint:   "workflow-ignore",
		Parallelism:  ptrInt(1),
		Templates: []Templatable{
			&Container{
				Name:    "hello-world",
				Image:   "busybox",
				Command: []string{"echo"},
				Args:    []string{"hello world"},
			},
			&Container{
				Name:    "intentional-fail",
				Image:   "alpine:latest",
				Command: []string{"sh", "-c"},
				Args:    []string{"echo intentional failure; exit 1"},
			},
			func() Templatable {
				s := &Steps{Name: "workflow-ignore"}
				s.AddSequentialStep(&Step{Name: "A", Template: "hello-world"})
				s.AddParallelGroup(
					&Step{Name: "B", Template: "hello-world"},
					&Step{Name: "C", Template: "intentional-fail", ContinueOn: &model.ContinueOn{Failed: true}},
				)
				s.AddSequentialStep(&Step{Name: "D", Template: "hello-world"})
				return s
			}(),
		},
	}
}

func buildGcTtl() *Workflow {
	return &Workflow{
		GenerateName: "gc-ttl-",
		Entrypoint:   "hello-world",
		TTLStrategy: &model.TTLStrategy{
			SecondsAfterCompletion: ptrInt(10),
			SecondsAfterSuccess:    ptrInt(5),
			SecondsAfterFailure:    ptrInt(5),
		},
		Templates: []Templatable{
			&Container{
				Name:    "hello-world",
				Image:   "busybox",
				Command: []string{"echo"},
				Args:    []string{"hello world"},
			},
		},
	}
}

func buildPodGcStrategy() *Workflow {
	return &Workflow{
		GenerateName: "pod-gc-strategy-",
		Entrypoint:   "pod-gc-strategy",
		PodGC:        &model.PodGC{Strategy: "OnPodSuccess", DeleteDelayDuration: "30s"},
		Templates: []Templatable{
			func() Templatable {
				s := &Steps{Name: "pod-gc-strategy"}
				s.AddParallelGroup(
					&Step{Name: "fail", Template: "fail"},
					&Step{Name: "succeed", Template: "succeed"},
				)
				return s
			}(),
			&Container{
				Name:    "fail",
				Image:   "alpine:3.7",
				Command: []string{"sh", "-c"},
				Args:    []string{"exit 1"},
			},
			&Container{
				Name:    "succeed",
				Image:   "alpine:3.7",
				Command: []string{"sh", "-c"},
				Args:    []string{"exit 0"},
			},
		},
	}
}

// === TESTS ===

func TestBuildHelloWorld(t *testing.T) {
	w := buildHelloWorld()
	yaml, err := w.ToYAML()
	if err != nil {
		t.Fatal(err)
	}
	assertYAMLEqual(t, "hello-world", yaml)
}

func TestBuildSteps(t *testing.T) {
	w := buildSteps()
	yaml, err := w.ToYAML()
	if err != nil {
		t.Fatal(err)
	}
	assertYAMLEqual(t, "steps", yaml)
}

func TestBuildDagDiamond(t *testing.T) {
	w := buildDagDiamond()
	yaml, err := w.ToYAML()
	if err != nil {
		t.Fatal(err)
	}
	assertYAMLEqual(t, "dag-diamond", yaml)
}

func TestBuildArgumentsParameters(t *testing.T) {
	w := buildArgumentsParameters()
	yaml, err := w.ToYAML()
	if err != nil {
		t.Fatal(err)
	}
	assertYAMLEqual(t, "arguments-parameters", yaml)
}

func TestBuildCoinflip(t *testing.T) {
	w := buildCoinflip()
	yaml, err := w.ToYAML()
	if err != nil {
		t.Fatal(err)
	}
	assertYAMLEqual(t, "coinflip", yaml)
}

func TestBuildConditionals(t *testing.T) {
	w := buildConditionals()
	yaml, err := w.ToYAML()
	if err != nil {
		t.Fatal(err)
	}
	assertYAMLEqual(t, "conditionals", yaml)
}

func TestBuildScriptsPython(t *testing.T) {
	w := buildScriptsPython()
	yaml, err := w.ToYAML()
	if err != nil {
		t.Fatal(err)
	}
	assertYAMLEqual(t, "scripts-python", yaml)
}

func TestBuildLoops(t *testing.T) {
	w := buildLoops()
	yaml, err := w.ToYAML()
	if err != nil {
		t.Fatal(err)
	}
	assertYAMLEqual(t, "loops", yaml)
}

func TestBuildLoopsMaps(t *testing.T) {
	w := buildLoopsMaps()
	yaml, err := w.ToYAML()
	if err != nil {
		t.Fatal(err)
	}
	assertYAMLEqual(t, "loops-maps", yaml)
}

func TestBuildOutputParameter(t *testing.T) {
	w := buildOutputParameter()
	yaml, err := w.ToYAML()
	if err != nil {
		t.Fatal(err)
	}
	assertYAMLEqual(t, "output-parameter", yaml)
}

func TestBuildGlobalParameters(t *testing.T) {
	w := buildGlobalParameters()
	yaml, err := w.ToYAML()
	if err != nil {
		t.Fatal(err)
	}
	assertYAMLEqual(t, "global-parameters", yaml)
}

func TestBuildExitHandlers(t *testing.T) {
	w := buildExitHandlers()
	yaml, err := w.ToYAML()
	if err != nil {
		t.Fatal(err)
	}
	assertYAMLEqual(t, "exit-handlers", yaml)
}

func TestBuildForever(t *testing.T) {
	w := buildForever()
	yaml, err := w.ToYAML()
	if err != nil {
		t.Fatal(err)
	}
	assertYAMLEqual(t, "forever", yaml)
}

func TestBuildArtifactPassing(t *testing.T) {
	w := buildArtifactPassing()
	yaml, err := w.ToYAML()
	if err != nil {
		t.Fatal(err)
	}
	assertYAMLEqual(t, "artifact-passing", yaml)
}

func TestBuildSuspendTemplate(t *testing.T) {
	w := buildSuspendTemplate()
	yaml, err := w.ToYAML()
	if err != nil {
		t.Fatal(err)
	}
	assertYAMLEqual(t, "suspend-template", yaml)
}

func TestBuildCronWorkflowExample(t *testing.T) {
	cw := buildCronWorkflow()
	yaml, err := cw.ToYAML()
	if err != nil {
		t.Fatal(err)
	}
	assertYAMLEqual(t, "cron-workflow", yaml)
}

func TestBuildParallelismLimit(t *testing.T) {
	w := buildParallelismLimit()
	yaml, err := w.ToYAML()
	if err != nil {
		t.Fatal(err)
	}
	assertYAMLEqual(t, "parallelism-limit", yaml)
}

func TestBuildNodeSelector(t *testing.T) {
	w := buildNodeSelector()
	yaml, err := w.ToYAML()
	if err != nil {
		t.Fatal(err)
	}
	assertYAMLEqual(t, "node-selector", yaml)
}

func TestBuildRetryBackoff(t *testing.T) {
	w := buildRetryBackoff()
	yaml, err := w.ToYAML()
	if err != nil {
		t.Fatal(err)
	}
	assertYAMLEqual(t, "retry-backoff", yaml)
}

func TestBuildDagEnhancedDepends(t *testing.T) {
	w := buildDagEnhancedDepends()
	yaml, err := w.ToYAML()
	if err != nil {
		t.Fatal(err)
	}
	assertYAMLEqual(t, "dag-enhanced-depends", yaml)
}

func TestBuildDagMultiroot(t *testing.T) {
	w := buildDagMultiroot()
	yaml, err := w.ToYAML()
	if err != nil {
		t.Fatal(err)
	}
	assertYAMLEqual(t, "dag-multiroot", yaml)
}

func TestBuildDagTargets(t *testing.T) {
	w := buildDagTargets()
	yaml, err := w.ToYAML()
	if err != nil {
		t.Fatal(err)
	}
	assertYAMLEqual(t, "dag-targets", yaml)
}

func TestBuildHttpHelloWorld(t *testing.T) {
	w := buildHttpHelloWorld()
	yaml, err := w.ToYAML()
	if err != nil {
		t.Fatal(err)
	}
	assertYAMLEqual(t, "http-hello-world", yaml)
}

func TestBuildVolumesEmptyDir(t *testing.T) {
	w := buildVolumesEmptyDir()
	yaml, err := w.ToYAML()
	if err != nil {
		t.Fatal(err)
	}
	assertYAMLEqual(t, "volumes-emptydir", yaml)
}

func TestBuildSecrets(t *testing.T) {
	w := buildSecretsExample()
	yaml, err := w.ToYAML()
	if err != nil {
		t.Fatal(err)
	}
	assertYAMLEqual(t, "secrets", yaml)
}

func TestBuildContinueOnFail(t *testing.T) {
	w := buildContinueOnFail()
	yaml, err := w.ToYAML()
	if err != nil {
		t.Fatal(err)
	}
	assertYAMLEqual(t, "continue-on-fail", yaml)
}

func TestBuildGcTtl(t *testing.T) {
	w := buildGcTtl()
	yaml, err := w.ToYAML()
	if err != nil {
		t.Fatal(err)
	}
	assertYAMLEqual(t, "gc-ttl", yaml)
}

func TestBuildPodGcStrategy(t *testing.T) {
	w := buildPodGcStrategy()
	yaml, err := w.ToYAML()
	if err != nil {
		t.Fatal(err)
	}
	assertYAMLEqual(t, "pod-gc-strategy", yaml)
}
