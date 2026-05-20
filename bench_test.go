package forge

import (
	"testing"

	"github.com/usetheodev/theo-forge/serialize"
)

// Extra-2 — performance baselines for the hot paths.
//
// These benchmarks exist so reviewers can spot perf regressions in PRs
// (e.g., an inadvertent O(n^2) loop, a new allocation in a tight path).
// They do NOT define a SLO — Argo workflows are batch jobs and the SDK
// runs once per submission. The numbers are reference-only.
//
// Run all: make bench
// Run one: go test -bench=BenchmarkWorkflow_Build -benchmem -run='^$' ./...
//
// Benchmarks here mirror the canonical README quickstart shape (1
// container template + 1 entrypoint) plus a moderate DAG (4 tasks) to
// cover both the trivial and the realistic paths.

// helloWorldWorkflow is the fixture used by every Build/ToYAML/RoundTrip
// benchmark. Keeping the construction OUT of the benchmark loop avoids
// measuring allocator + struct-literal cost we don't care about.
func helloWorldWorkflow() *Workflow {
	return &Workflow{
		Name:       "hello-bench",
		Namespace:  "default",
		Entrypoint: "main",
		Templates: []Templatable{
			&Container{
				Name:    "main",
				Image:   "alpine:3.18",
				Command: []string{"echo"},
				Args:    []string{"hello"},
			},
		},
	}
}

// diamondDAGWorkflow exercises the more realistic translation path:
// 4 tasks + dependency edges + template lookup.
func diamondDAGWorkflow() *Workflow {
	echo := &Container{
		Name:    "echo",
		Image:   "alpine:3.18",
		Command: []string{"echo"},
	}
	mk := func(name string) *Task {
		return &Task{Name: name, Template: "echo"}
	}
	a, b, c, d := mk("a"), mk("b"), mk("c"), mk("d")
	a.Then(b)
	a.Then(c)
	b.Then(d)
	c.Then(d)
	dag := &DAG{Name: "diamond"}
	dag.AddTasks(a, b, c, d)
	return &Workflow{
		Name:       "diamond-bench",
		Entrypoint: "diamond",
		Templates:  []Templatable{echo, dag},
	}
}

func BenchmarkWorkflow_Build_HelloWorld(b *testing.B) {
	w := helloWorldWorkflow()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := w.Build(); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkWorkflow_Build_DiamondDAG(b *testing.B) {
	w := diamondDAGWorkflow()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := w.Build(); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkWorkflow_ToYAML_HelloWorld(b *testing.B) {
	w := helloWorldWorkflow()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := w.ToYAML(); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkWorkflow_ToYAML_DiamondDAG(b *testing.B) {
	w := diamondDAGWorkflow()
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := w.ToYAML(); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkWorkflow_RoundTrip_HelloWorld(b *testing.B) {
	w := helloWorldWorkflow()
	yamlStr, err := w.ToYAML()
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m, err := serialize.WorkflowFromYAML(yamlStr)
		if err != nil {
			b.Fatal(err)
		}
		if _, err := serialize.WorkflowToYAML(m); err != nil {
			b.Fatal(err)
		}
	}
}
