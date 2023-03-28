package job

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
