package job

import (
	"errors"
	"fmt"
	"sort"
)

// ErrCycle is returned when a dependency graph cannot be ordered because it
// contains a cycle.
var ErrCycle = errors.New("job: dependency cycle detected")

// Graph is a directed dependency graph over job IDs. An edge from A to B means
// "B depends on A": A must finish before B may run. It is the structure the
// scheduler walks to order a DAG of jobs.
type Graph struct {
	nodes    map[string]bool
	parents  map[string]map[string]bool
	children map[string]map[string]bool
}

// NewGraph returns an empty Graph.
func NewGraph() *Graph {
	return &Graph{
		nodes:    map[string]bool{},
		parents:  map[string]map[string]bool{},
		children: map[string]map[string]bool{},
	}
}

// AddNode records a job ID with no edges. Adding an existing node is a no-op.
func (g *Graph) AddNode(id string) {
	g.nodes[id] = true
	if g.parents[id] == nil {
		g.parents[id] = map[string]bool{}
	}
	if g.children[id] == nil {
		g.children[id] = map[string]bool{}
	}
}

// AddEdge adds a dependency edge from -> to, meaning "to depends on from". Both
// endpoints are created if they do not yet exist.
func (g *Graph) AddEdge(from, to string) {
	g.AddNode(from)
	g.AddNode(to)
	g.children[from][to] = true
	g.parents[to][from] = true
}

// Dependencies returns the IDs that id directly depends on (its parents).
func (g *Graph) Dependencies(id string) []string {
	return keys(g.parents[id])
}

// Dependents returns the IDs that directly depend on id (its children).
func (g *Graph) Dependents(id string) []string {
	return keys(g.children[id])
}

func keys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

// TopoSort returns the job IDs in an order where every job appears after all of
// its dependencies. Ties are broken by ID so the ordering is deterministic. It
// returns ErrCycle if the graph cannot be linearised.
func (g *Graph) TopoSort() ([]string, error) {
	indeg := make(map[string]int, len(g.nodes))
	for id := range g.nodes {
		indeg[id] = len(g.parents[id])
	}

	queue := make([]string, 0)
	for id, d := range indeg {
		if d == 0 {
			queue = append(queue, id)
		}
	}

	order := make([]string, 0, len(g.nodes))
	for len(queue) > 0 {
		sort.Strings(queue)
		n := queue[0]
		queue = queue[1:]
		order = append(order, n)
		for child := range g.children[n] {
			indeg[child]--
			if indeg[child] == 0 {
				queue = append(queue, child)
			}
		}
	}

	if len(order) != len(g.nodes) {
		return nil, ErrCycle
	}
	return order, nil
}

// BuildGraph constructs a Graph from a set of jobs, adding an edge from each
// dependency to the job that declares it. It errors if a job depends on an
// unknown ID or if the resulting graph contains a cycle.
func BuildGraph(jobs []Job) (*Graph, error) {
	g := NewGraph()
	known := make(map[string]bool, len(jobs))
	for _, j := range jobs {
		g.AddNode(j.ID)
		known[j.ID] = true
	}
	for _, j := range jobs {
		for _, dep := range j.DependsOn {
			if !known[dep] {
				return nil, fmt.Errorf("job %q depends on unknown job %q", j.ID, dep)
			}
			g.AddEdge(dep, j.ID)
		}
	}
	if _, err := g.TopoSort(); err != nil {
		return nil, err
	}
	return g, nil
}
