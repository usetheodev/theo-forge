package forge

import (
	"strings"
	"testing"

	"github.com/usetheodev/theo-forge/expr"
	"github.com/usetheodev/theo-forge/model"
)

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
	goldenTest(t, "hello-world.golden", yaml)
}

func TestBuildSteps(t *testing.T) {
	w := buildSteps()
	yaml, err := w.ToYAML()
	if err != nil {
		t.Fatal(err)
	}
	goldenTest(t, "steps.golden", yaml)
}

func TestBuildDagDiamond(t *testing.T) {
	w := buildDagDiamond()
	yaml, err := w.ToYAML()
	if err != nil {
		t.Fatal(err)
	}
	goldenTest(t, "dag-diamond.golden", yaml)
}

func TestBuildArgumentsParameters(t *testing.T) {
	w := buildArgumentsParameters()
	yaml, err := w.ToYAML()
	if err != nil {
		t.Fatal(err)
	}
	goldenTest(t, "arguments-parameters.golden", yaml)
}

func TestBuildCoinflip(t *testing.T) {
	w := buildCoinflip()
	yaml, err := w.ToYAML()
	if err != nil {
		t.Fatal(err)
	}
	goldenTest(t, "coinflip.golden", yaml)
}

func TestBuildConditionals(t *testing.T) {
	w := buildConditionals()
	yaml, err := w.ToYAML()
	if err != nil {
		t.Fatal(err)
	}
	goldenTest(t, "conditionals.golden", yaml)
}

func TestBuildScriptsPython(t *testing.T) {
	w := buildScriptsPython()
	yaml, err := w.ToYAML()
	if err != nil {
		t.Fatal(err)
	}
	goldenTest(t, "scripts-python.golden", yaml)
}

func TestBuildLoops(t *testing.T) {
	w := buildLoops()
	yaml, err := w.ToYAML()
	if err != nil {
		t.Fatal(err)
	}
	goldenTest(t, "loops.golden", yaml)
}

func TestBuildLoopsMaps(t *testing.T) {
	w := buildLoopsMaps()
	yaml, err := w.ToYAML()
	if err != nil {
		t.Fatal(err)
	}
	goldenTest(t, "loops-maps.golden", yaml)
}

func TestBuildOutputParameter(t *testing.T) {
	w := buildOutputParameter()
	yaml, err := w.ToYAML()
	if err != nil {
		t.Fatal(err)
	}
	goldenTest(t, "output-parameter.golden", yaml)
}

func TestBuildGlobalParameters(t *testing.T) {
	w := buildGlobalParameters()
	yaml, err := w.ToYAML()
	if err != nil {
		t.Fatal(err)
	}
	goldenTest(t, "global-parameters.golden", yaml)
}

func TestBuildExitHandlers(t *testing.T) {
	w := buildExitHandlers()
	yaml, err := w.ToYAML()
	if err != nil {
		t.Fatal(err)
	}
	goldenTest(t, "exit-handlers.golden", yaml)
}

func TestBuildForever(t *testing.T) {
	w := buildForever()
	yaml, err := w.ToYAML()
	if err != nil {
		t.Fatal(err)
	}
	goldenTest(t, "forever.golden", yaml)
}

func TestBuildArtifactPassing(t *testing.T) {
	w := buildArtifactPassing()
	yaml, err := w.ToYAML()
	if err != nil {
		t.Fatal(err)
	}
	goldenTest(t, "artifact-passing.golden", yaml)
}

func TestBuildSuspendTemplate(t *testing.T) {
	w := buildSuspendTemplate()
	yaml, err := w.ToYAML()
	if err != nil {
		t.Fatal(err)
	}
	goldenTest(t, "suspend-template.golden", yaml)
}

func TestBuildCronWorkflowExample(t *testing.T) {
	cw := buildCronWorkflow()
	yaml, err := cw.ToYAML()
	if err != nil {
		t.Fatal(err)
	}
	goldenTest(t, "cron-workflow.golden", yaml)
}

func TestBuildParallelismLimit(t *testing.T) {
	w := buildParallelismLimit()
	yaml, err := w.ToYAML()
	if err != nil {
		t.Fatal(err)
	}
	goldenTest(t, "parallelism-limit.golden", yaml)
}

func TestBuildNodeSelector(t *testing.T) {
	w := buildNodeSelector()
	yaml, err := w.ToYAML()
	if err != nil {
		t.Fatal(err)
	}
	goldenTest(t, "node-selector.golden", yaml)
}

func TestBuildRetryBackoff(t *testing.T) {
	w := buildRetryBackoff()
	yaml, err := w.ToYAML()
	if err != nil {
		t.Fatal(err)
	}
	goldenTest(t, "retry-backoff.golden", yaml)
}

func TestBuildDagEnhancedDepends(t *testing.T) {
	w := buildDagEnhancedDepends()
	yaml, err := w.ToYAML()
	if err != nil {
		t.Fatal(err)
	}
	goldenTest(t, "dag-enhanced-depends.golden", yaml)
}

func TestBuildDagMultiroot(t *testing.T) {
	w := buildDagMultiroot()
	yaml, err := w.ToYAML()
	if err != nil {
		t.Fatal(err)
	}
	goldenTest(t, "dag-multiroot.golden", yaml)
}

func TestBuildDagTargets(t *testing.T) {
	w := buildDagTargets()
	yaml, err := w.ToYAML()
	if err != nil {
		t.Fatal(err)
	}
	goldenTest(t, "dag-targets.golden", yaml)
}

func TestBuildHttpHelloWorld(t *testing.T) {
	w := buildHttpHelloWorld()
	yaml, err := w.ToYAML()
	if err != nil {
		t.Fatal(err)
	}
	goldenTest(t, "http-hello-world.golden", yaml)
}

func TestBuildVolumesEmptyDir(t *testing.T) {
	w := buildVolumesEmptyDir()
	yaml, err := w.ToYAML()
	if err != nil {
		t.Fatal(err)
	}
	goldenTest(t, "volumes-emptydir.golden", yaml)
}

func TestBuildSecrets(t *testing.T) {
	w := buildSecretsExample()
	yaml, err := w.ToYAML()
	if err != nil {
		t.Fatal(err)
	}
	goldenTest(t, "secrets.golden", yaml)
}

func TestBuildContinueOnFail(t *testing.T) {
	w := buildContinueOnFail()
	yaml, err := w.ToYAML()
	if err != nil {
		t.Fatal(err)
	}
	goldenTest(t, "continue-on-fail.golden", yaml)
}

func TestBuildGcTtl(t *testing.T) {
	w := buildGcTtl()
	yaml, err := w.ToYAML()
	if err != nil {
		t.Fatal(err)
	}
	goldenTest(t, "gc-ttl.golden", yaml)
}

func TestBuildPodGcStrategy(t *testing.T) {
	w := buildPodGcStrategy()
	yaml, err := w.ToYAML()
	if err != nil {
		t.Fatal(err)
	}
	goldenTest(t, "pod-gc-strategy.golden", yaml)
}

// --- Example tests (consolidated from example_test.go) ---

// TestExampleDiamondDAG builds a complete diamond DAG workflow and validates YAML.
func TestExampleDiamondDAG(t *testing.T) {
	echoTpl := &Container{
		Name:    "echo",
		Image:   "alpine:3.18",
		Command: []string{"echo"},
		Args:    []string{expr.InputParam("msg")},
		Inputs:  []Parameter{{Name: "msg"}},
	}

	dag := &DAG{Name: "diamond"}
	A := &Task{Name: "A", Template: "echo", Arguments: []Parameter{{Name: "msg", Value: ptrStr("Task A")}}}
	B := &Task{Name: "B", Template: "echo", Arguments: []Parameter{{Name: "msg", Value: ptrStr("Task B")}}}
	C := &Task{Name: "C", Template: "echo", Arguments: []Parameter{{Name: "msg", Value: ptrStr("Task C")}}}
	D := &Task{Name: "D", Template: "echo", Arguments: []Parameter{{Name: "msg", Value: ptrStr("Task D")}}}

	A.Then(B)
	A.Then(C)
	B.Then(D)
	C.Then(D)
	dag.AddTasks(A, B, C, D)

	w := &Workflow{
		GenerateName: "diamond-",
		Namespace:    "argo",
		Entrypoint:   "diamond",
		Templates:    []Templatable{echoTpl, dag},
	}

	y, err := w.ToYAML()
	if err != nil {
		t.Fatal(err)
	}

	expected := []string{
		"apiVersion: argoproj.io/v1alpha1",
		"kind: Workflow",
		"generateName: diamond-",
		"namespace: argo",
		"entrypoint: diamond",
		"name: echo",
		"image: alpine:3.18",
		"name: diamond",
		"name: A",
		"name: B",
		"name: C",
		"name: D",
		"depends: A",
	}
	for _, s := range expected {
		if !strings.Contains(y, s) {
			t.Errorf("YAML missing: %q", s)
		}
	}
}

// TestExampleCoinflip builds a coinflip workflow with conditionals.
func TestExampleCoinflip(t *testing.T) {
	flip := &Script{
		Name:    "flip-coin",
		Image:   "python:3.11-alpine",
		Command: []string{"python"},
		Source: `import random
result = "heads" if random.randint(0, 1) == 0 else "tails"
print(result)`,
	}

	heads := &Container{
		Name:    "heads",
		Image:   "alpine:3.18",
		Command: []string{"echo"},
		Args:    []string{"it was heads"},
	}

	tails := &Container{
		Name:    "tails",
		Image:   "alpine:3.18",
		Command: []string{"echo"},
		Args:    []string{"it was tails"},
	}

	steps := &Steps{Name: "coinflip"}
	steps.AddSequentialStep(&Step{Name: "flip", Template: "flip-coin"})
	steps.AddParallelGroup(
		&Step{Name: "heads", Template: "heads", When: "{{steps.flip.outputs.result}} == heads"},
		&Step{Name: "tails", Template: "tails", When: "{{steps.flip.outputs.result}} == tails"},
	)

	w := &Workflow{
		GenerateName: "coinflip-",
		Entrypoint:   "coinflip",
		Templates:    []Templatable{flip, heads, tails, steps},
	}

	y, err := w.ToYAML()
	if err != nil {
		t.Fatal(err)
	}

	for _, s := range []string{"flip-coin", "heads", "tails", "coinflip", "when:"} {
		if !strings.Contains(y, s) {
			t.Errorf("YAML missing: %q", s)
		}
	}
}

// TestExampleParameterPassing builds a workflow that passes outputs between steps.
func TestExampleParameterPassing(t *testing.T) {
	generate := &Script{
		Name:    "generate",
		Image:   "alpine:3.18",
		Command: []string{"sh", "-c"},
		Source:  `echo "42" > /tmp/result`,
		Outputs: []Parameter{{Name: "result", ValueFrom: &ValueFrom{Path: "/tmp/result"}}},
	}

	consume := &Container{
		Name:    "consume",
		Image:   "alpine:3.18",
		Command: []string{"echo"},
		Args:    []string{expr.InputParam("msg")},
		Inputs:  []Parameter{{Name: "msg"}},
	}

	dag := &DAG{Name: "main"}
	genTask := &Task{Name: "generate", Template: "generate"}
	consumeTask := &Task{
		Name:     "consume",
		Template: "consume",
		Arguments: []Parameter{
			{Name: "msg", Value: ptrStr(expr.TaskOutputParam("generate", "result"))},
		},
	}
	genTask.Then(consumeTask)
	dag.AddTasks(genTask, consumeTask)

	w := &Workflow{
		GenerateName: "param-passing-",
		Entrypoint:   "main",
		Templates:    []Templatable{generate, consume, dag},
	}

	y, err := w.ToYAML()
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(y, "{{tasks.generate.outputs.parameters.result}}") {
		t.Error("YAML missing task output reference")
	}
}

// TestExampleArtifactPassing builds a workflow with artifact passing.
func TestExampleArtifactPassing(t *testing.T) {
	generate := &Script{
		Name:    "generate",
		Image:   "alpine:3.18",
		Command: []string{"sh", "-c"},
		Source:  `echo "hello world" > /tmp/output.txt`,
		OutputArtifacts: []ArtifactBuilder{
			&Artifact{Name: "output-file", Path: "/tmp/output.txt"},
		},
	}

	consume := &Container{
		Name:    "consume",
		Image:   "alpine:3.18",
		Command: []string{"cat"},
		Args:    []string{"/tmp/input.txt"},
	}

	w := &Workflow{
		GenerateName: "artifacts-",
		Entrypoint:   "main",
		Templates: []Templatable{
			generate,
			consume,
			&DAG{
				Name: "main",
				Tasks: []*Task{
					{Name: "gen", Template: "generate"},
					{Name: "use", Template: "consume", Depends: "gen"},
				},
			},
		},
	}

	y, err := w.ToYAML()
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(y, "output-file") {
		t.Error("YAML missing artifact name")
	}
}

// TestExampleCronWorkflow builds a scheduled workflow.
func TestExampleCronWorkflow(t *testing.T) {
	cw := &CronWorkflow{
		Name:              "hourly-cleanup",
		Namespace:         "ops",
		Schedule:          "0 * * * *",
		Timezone:          "UTC",
		ConcurrencyPolicy: "Forbid",
		Entrypoint:        "cleanup",
		Templates: []Templatable{
			&Container{
				Name:    "cleanup",
				Image:   "alpine:3.18",
				Command: []string{"sh", "-c"},
				Args:    []string{"echo 'cleaning up...'"},
			},
		},
	}

	y, err := cw.ToYAML()
	if err != nil {
		t.Fatal(err)
	}

	for _, s := range []string{
		"kind: CronWorkflow",
		"schedule: 0 * * * *",
		"timezone: UTC",
		"concurrencyPolicy: Forbid",
		"name: hourly-cleanup",
		"namespace: ops",
	} {
		if !strings.Contains(y, s) {
			t.Errorf("YAML missing: %q", s)
		}
	}
}

// TestExampleWorkflowTemplateRef builds a workflow that references a WorkflowTemplate.
func TestExampleWorkflowTemplateRef(t *testing.T) {
	// First, define the reusable template
	wt := &WorkflowTemplate{
		Name:       "echo-template",
		Namespace:  "default",
		Entrypoint: "echo",
		Templates: []Templatable{
			&Container{
				Name:    "echo",
				Image:   "alpine:3.18",
				Command: []string{"echo"},
				Args:    []string{expr.InputParam("msg")},
				Inputs:  []Parameter{{Name: "msg"}},
			},
		},
	}

	wtYAML, err := wt.ToYAML()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(wtYAML, "kind: WorkflowTemplate") {
		t.Error("missing kind in WorkflowTemplate YAML")
	}

	// Then, use it in a workflow via templateRef
	dag := &DAG{Name: "main"}
	dag.AddTask(&Task{
		Name: "call-echo",
		TemplateRef: &TemplateRef{
			Name:     "echo-template",
			Template: "echo",
		},
		Arguments: []Parameter{{Name: "msg", Value: ptrStr("Hello from ref!")}},
	})

	w := &Workflow{
		Name:       "use-template-ref",
		Entrypoint: "main",
		Templates:  []Templatable{dag},
	}

	y, err := w.ToYAML()
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(y, "templateRef:") {
		t.Error("YAML missing templateRef")
	}
	if !strings.Contains(y, "name: echo-template") {
		t.Error("YAML missing template ref name")
	}
}

// TestExampleWithVolumesAndSecrets builds a complete workflow with volumes.
func TestExampleWithVolumesAndSecrets(t *testing.T) {
	w := &Workflow{
		Name:       "with-volumes",
		Entrypoint: "main",
		Volumes: []VolumeBuilder{
			&SecretVolume{BaseVolume: BaseVolume{Name: "creds", MountPath: "/etc/creds"}, SecretName: "app-creds"},
			&EmptyDirVolume{BaseVolume: BaseVolume{Name: "workspace", MountPath: "/workspace"}},
		},
		Templates: []Templatable{
			&Container{
				Name:    "main",
				Image:   "alpine:3.18",
				Command: []string{"sh", "-c"},
				Args:    []string{"cat /etc/creds/password && ls /workspace"},
				Env: []EnvBuilder{
					SecretEnv{Name: "DB_PASS", SecretName: "db-creds", SecretKey: "password"},
				},
				VolumeMounts: []VolumeBuilder{
					&SecretVolume{BaseVolume: BaseVolume{Name: "creds", MountPath: "/etc/creds"}},
					&EmptyDirVolume{BaseVolume: BaseVolume{Name: "workspace", MountPath: "/workspace"}},
				},
			},
		},
	}

	y, err := w.ToYAML()
	if err != nil {
		t.Fatal(err)
	}

	for _, s := range []string{"secretName: app-creds", "emptyDir:", "mountPath: /etc/creds", "mountPath: /workspace"} {
		if !strings.Contains(y, s) {
			t.Errorf("YAML missing: %q", s)
		}
	}
}

// --- Advanced example tests (consolidated from example_advanced_test.go) ---

// TestExampleDefaultParameterOverwrite replicates Hera's default-parameters.yaml
func TestExampleDefaultParameterOverwrite(t *testing.T) {
	generator := &Script{
		Name:    "generator",
		Image:   "python:3.10",
		Command: []string{"python"},
		Source:  "print('Another message for the world!')",
	}

	consumer := &Script{
		Name:    "consumer",
		Image:   "python:3.10",
		Command: []string{"python"},
		Source:  "print('{{inputs.parameters.message}}')",
		Inputs: []Parameter{
			{Name: "message", Default: ptrStr("Hello, world!")},
			{Name: "foo", Default: ptrStr("42")},
		},
	}

	dag := &DAG{Name: "d"}
	genTask := &Task{Name: "generator", Template: "generator"}
	consumeDefault := &Task{Name: "consume-default", Template: "consumer"}
	consumeArg := &Task{
		Name:     "consume-argument",
		Template: "consumer",
		Arguments: []Parameter{
			{Name: "message", Value: ptrStr(genTask.GetOutputResult())},
		},
	}
	genTask.Then(consumeDefault)
	genTask.Then(consumeArg)
	dag.AddTasks(genTask, consumeDefault, consumeArg)

	w := &Workflow{
		GenerateName: "default-param-overwrite-",
		Entrypoint:   "d",
		Templates:    []Templatable{dag, generator, consumer},
	}

	y, err := w.ToYAML()
	if err != nil {
		t.Fatal(err)
	}

	// Verify key structural elements
	for _, s := range []string{
		"generateName: default-param-overwrite-",
		"entrypoint: d",
		"name: generator",
		"name: consumer",
		"name: consume-default",
		"name: consume-argument",
		"default: Hello, world!",
		"default: \"42\"",
	} {
		if !strings.Contains(y, s) {
			t.Errorf("YAML missing: %q\n\nFull YAML:\n%s", s, y)
		}
	}
}

// TestExampleOutputParameterPassing replicates Hera's output-parameters.yaml
func TestExampleOutputParameterPassing(t *testing.T) {
	outScript := &Script{
		Name:    "out",
		Image:   "python:3.10",
		Command: []string{"python"},
		Source:  "with open('/test', 'w') as f:\n    f.write('test')",
		Outputs: []Parameter{
			{Name: "a", ValueFrom: &ValueFrom{Path: "/test"}},
		},
	}

	inScript := &Script{
		Name:    "in-",
		Image:   "python:3.10",
		Command: []string{"python"},
		Source:  "print('{{inputs.parameters.a}}')",
		Inputs:  []Parameter{{Name: "a"}},
	}

	dag := &DAG{Name: "d"}
	outTask := &Task{Name: "out", Template: "out"}
	inTask := &Task{
		Name:     "in-",
		Template: "in-",
		Arguments: []Parameter{
			{Name: "a", Value: ptrStr(outTask.GetOutputParameter("a"))},
		},
	}
	outTask.Then(inTask)
	dag.AddTasks(outTask, inTask)

	w := &Workflow{
		GenerateName: "script-output-param-passing-",
		Entrypoint:   "d",
		Templates:    []Templatable{dag, outScript, inScript},
	}

	y, err := w.ToYAML()
	if err != nil {
		t.Fatal(err)
	}

	// Verify the output parameter reference is correct
	if !strings.Contains(y, "{{tasks.out.outputs.parameters.a}}") {
		t.Errorf("YAML missing output parameter reference\n\n%s", y)
	}
	// Verify outputs section
	if !strings.Contains(y, "valueFrom:") {
		t.Error("YAML missing valueFrom")
	}
	if !strings.Contains(y, "path: /test") {
		t.Error("YAML missing path: /test")
	}
}

// TestExampleWithItemsLoop replicates Hera's loop patterns
func TestExampleWithItemsLoop(t *testing.T) {
	echo := &Container{
		Name:    "echo",
		Image:   "alpine:3.18",
		Command: []string{"echo"},
		Args:    []string{"{{inputs.parameters.message}}"},
		Inputs:  []Parameter{{Name: "message"}},
	}

	dag := &DAG{Name: "main"}
	loopTask := &Task{
		Name:     "echo-loop",
		Template: "echo",
		Arguments: []Parameter{
			{Name: "message", Value: ptrStr("{{item}}")},
		},
		WithItems: []interface{}{"hello", "world", "foo"},
	}
	dag.AddTask(loopTask)

	w := &Workflow{
		GenerateName: "loops-",
		Entrypoint:   "main",
		Templates:    []Templatable{echo, dag},
	}

	y, err := w.ToYAML()
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(y, "withItems:") {
		t.Error("YAML missing withItems")
	}
	if !strings.Contains(y, "hello") || !strings.Contains(y, "world") {
		t.Error("YAML missing items")
	}
}

// TestExampleWithParamLoop tests withParam-based fan-out
func TestExampleWithParamLoop(t *testing.T) {
	generate := &Script{
		Name:    "generate-list",
		Image:   "python:3.11",
		Command: []string{"python"},
		Source:  `import json; print(json.dumps(["a", "b", "c"]))`,
	}

	process := &Container{
		Name:    "process",
		Image:   "alpine:3.18",
		Command: []string{"echo"},
		Args:    []string{"{{inputs.parameters.item}}"},
		Inputs:  []Parameter{{Name: "item"}},
	}

	dag := &DAG{Name: "main"}
	genTask := &Task{Name: "gen", Template: "generate-list"}
	processTask := &Task{
		Name:     "process",
		Template: "process",
		Arguments: []Parameter{
			{Name: "item", Value: ptrStr("{{item}}")},
		},
		WithParam: genTask.GetOutputResult(),
	}
	genTask.Then(processTask)
	dag.AddTasks(genTask, processTask)

	w := &Workflow{
		GenerateName: "param-loop-",
		Entrypoint:   "main",
		Templates:    []Templatable{generate, process, dag},
	}

	y, err := w.ToYAML()
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(y, "withParam:") {
		t.Error("YAML missing withParam")
	}
	if !strings.Contains(y, "{{tasks.gen.outputs.result}}") {
		t.Error("YAML missing result reference in withParam")
	}
}

// TestExampleRetryWithBackoff tests retry configuration
func TestExampleRetryWithBackoff(t *testing.T) {
	limit := 3
	factor := 2
	w := &Workflow{
		GenerateName: "retry-",
		Entrypoint:   "main",
		Templates: []Templatable{
			&Script{
				Name:    "main",
				Image:   "python:3.11",
				Command: []string{"python"},
				Source:  "import random; assert random.random() > 0.5",
				RetryStrategy: &RetryStrategy{
					Limit:       &limit,
					RetryPolicy: RetryOnFailure,
					Backoff: &Backoff{
						Duration:    "5s",
						Factor:      &factor,
						MaxDuration: "1m",
					},
				},
			},
		},
	}

	y, err := w.ToYAML()
	if err != nil {
		t.Fatal(err)
	}

	for _, s := range []string{"retryStrategy:", "retryPolicy: OnFailure", "duration: 5s", "maxDuration: 1m"} {
		if !strings.Contains(y, s) {
			t.Errorf("YAML missing: %q", s)
		}
	}
}

// TestExampleSuspendApprovalGate tests manual approval pattern
func TestExampleSuspendApprovalGate(t *testing.T) {
	steps := &Steps{Name: "approval-flow"}
	steps.AddSequentialStep(&Step{Name: "deploy-staging", Template: "deploy"})
	steps.AddSequentialStep(&Step{Name: "wait-approval", Template: "approve"})
	steps.AddSequentialStep(&Step{Name: "deploy-prod", Template: "deploy"})

	w := &Workflow{
		Name:       "approval-gate",
		Entrypoint: "approval-flow",
		Templates: []Templatable{
			&Container{Name: "deploy", Image: "alpine", Command: []string{"echo"}, Args: []string{"deploying..."}},
			&Suspend{Name: "approve"},
			steps,
		},
	}

	y, err := w.ToYAML()
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(y, "suspend:") {
		t.Error("YAML missing suspend template")
	}
	if !strings.Contains(y, "deploy-staging") || !strings.Contains(y, "deploy-prod") {
		t.Error("YAML missing step names")
	}
}

// TestExampleMultiClusterTemplateRef tests referencing ClusterWorkflowTemplate
func TestExampleMultiClusterTemplateRef(t *testing.T) {
	// Define cluster-wide template
	cwt := &ClusterWorkflowTemplate{
		Name:       "shared-build",
		Entrypoint: "build",
		Templates: []Templatable{
			&Container{
				Name:    "build",
				Image:   "golang:1.22",
				Command: []string{"go"},
				Args:    []string{"build", "./..."},
				Inputs:  []Parameter{{Name: "repo"}},
			},
		},
	}
	cwtYAML, err := cwt.ToYAML()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(cwtYAML, "kind: ClusterWorkflowTemplate") {
		t.Error("CWT YAML missing kind")
	}

	// Use it via templateRef
	dag := &DAG{Name: "pipeline"}
	dag.AddTask(&Task{
		Name: "build",
		TemplateRef: &TemplateRef{
			Name:     "shared-build",
			Template: "build",
		},
		Arguments: []Parameter{
			{Name: "repo", Value: ptrStr("https://github.com/example/app.git")},
		},
	})

	w := &Workflow{
		GenerateName: "ci-",
		Entrypoint:   "pipeline",
		Templates:    []Templatable{dag},
	}

	y, err := w.ToYAML()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(y, "templateRef:") {
		t.Error("YAML missing templateRef")
	}
	if !strings.Contains(y, "name: shared-build") {
		t.Error("YAML missing CWT reference")
	}
}
