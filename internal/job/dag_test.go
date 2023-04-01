package job_test

import (
	"errors"
	"testing"

	"github.com/Fzgt/cronflux/internal/job"
)

func indexOf(order []string, id string) int {
	for i, v := range order {
		if v == id {
			return i
		}
	}
	return -1
}

func TestBuildGraphTopoOrder(t *testing.T) {
	// Diamond: a -> b, a -> c, b -> d, c -> d.
	jobs := []job.Job{
		{ID: "a"},
		{ID: "b", DependsOn: []string{"a"}},
		{ID: "c", DependsOn: []string{"a"}},
		{ID: "d", DependsOn: []string{"b", "c"}},
	}
	g, err := job.BuildGraph(jobs)
	if err != nil {
		t.Fatalf("BuildGraph: %v", err)
	}
	order, err := g.TopoSort()
	if err != nil {
		t.Fatalf("TopoSort: %v", err)
	}
	if len(order) != 4 {
		t.Fatalf("order length = %d, want 4", len(order))
	}
	for _, pair := range [][2]string{{"a", "b"}, {"a", "c"}, {"b", "d"}, {"c", "d"}} {
		if indexOf(order, pair[0]) >= indexOf(order, pair[1]) {
			t.Errorf("%s should come before %s in %v", pair[0], pair[1], order)
		}
	}
}

func TestBuildGraphDetectsCycle(t *testing.T) {
	jobs := []job.Job{
		{ID: "a", DependsOn: []string{"b"}},
		{ID: "b", DependsOn: []string{"a"}},
	}
	if _, err := job.BuildGraph(jobs); !errors.Is(err, job.ErrCycle) {
		t.Fatalf("expected ErrCycle, got %v", err)
	}
}

func TestBuildGraphUnknownDependency(t *testing.T) {
	jobs := []job.Job{{ID: "a", DependsOn: []string{"ghost"}}}
	if _, err := job.BuildGraph(jobs); err == nil {
		t.Fatal("expected error for unknown dependency")
	}
}

func TestDependenciesAndDependents(t *testing.T) {
	g := job.NewGraph()
	g.AddEdge("a", "b")
	if deps := g.Dependencies("b"); len(deps) != 1 || deps[0] != "a" {
		t.Errorf("Dependencies(b) = %v, want [a]", deps)
	}
	if kids := g.Dependents("a"); len(kids) != 1 || kids[0] != "b" {
		t.Errorf("Dependents(a) = %v, want [b]", kids)
	}
}
