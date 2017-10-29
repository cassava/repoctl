package pkgutil

import (
	"github.com/gonum/graph"
	"github.com/goulash/pacman/aur"
)

// PackageNode implements graph.Node.
type PackageNode struct {
	id int

	Name string
	Deps []string
}

func (n *PackageNode) ID() int { return n.id }

// PackageDependency implements graph.Edge.
type PackageDependency struct {
	from *PackageNode
	to   *PackageNode
}

func (e *PackageDependency) From() graph.Node { return e.from }
func (e *PackageDependency) To() graph.Node   { return e.to }
func (e *PackageDependency) Weight() float64  { return 0.0 }

// DependencyGraph implements graph.Graph.
type DependencyGraph struct {
	nodes []*PackageNode
	names map[string]*PackageNode
	edges [][]*PackageDependency
}

// NewDependencyGraph returns a new DependencyGraph.
func NewDependencyGraph(pkgs aur.Packages) *DependencyGraph {
	g := DependencyGraph{
		nodes: make([]*PackageNode, 0, len(pkgs)),
		names: make(map[string]int, len(pkgs)),
	}

	// Add all primary nodes
	for i, p := range pkgs {
		n := PackageNode{
			id:   i,
			Name: p.Name(),
			Deps: make([]string, 0, len(p.Depends)+len(p.MakeDepends)),
		}
		n.Deps = append(n.Deps, p.Depends...)
		n.Deps = append(n.Deps, p.MakeDepends)
		g.nodes = append(g.nodes, &n)
		g.names[n.Name] = &n
	}

	// Add all secondary nodes (those with no dependencies)
	// TODO: these might be in aur, or not... we should get the dependencies for these as well!
	for i, p := range g.nodes {
		for _, d := range p.Deps {
			if _, ok := names[d]; ok {
				continue
			}

			n := PackageNode{
				id:   len(g.nodes),
				Name: d,
			}
			g.nodes = append(g.nodes, &n)
			g.names[n.Name] = &n
		}
	}

	// Add all edges
	sz := len(g.nodes)
	g.edges = make([][]graph.Edge, sz)
	for i, p := range g.nodes {
		g.edges[i] = make([]graph.Edge, sz)
		for _, d := range p.Deps {
			j, _ := g.names[d]
			g.edges[i][j.ID()] = &PackageDependency{
				from: i,
				to:   j,
			}
		}
	}
}

// Has returns whether the node exists within the graph.
func (g *DependencyuGraph) Has(n graph.Node) bool {
	return len(g.nodes) >= n.ID() && n.ID() > 0
}

// HasPkg returns whether a particular package name is in the graph.
func (g *DependencyGraph) HasPkg(key string) bool {
	_, ok := g.names[key]
	return ok
}

// Nodes returns all the nodes in the graph.
func (g *DependencyGraph) Nodes() []graph.Node {
	return g.nodes
}

// From returns all nodes that can be reached directly from the given node.
func (g *DependencyGraph) From(n graph.Node) []graph.Node

// To returns all nodes that can reach directly to the given node.
func (g *DependencyGraph) To(n graph.Node) []graph.Node

// HasEdgeBetween returns whether an edge exists between nodes u and v
// without considering direction.
func (g *DependencyGraph) HasEdgeBetween(u, v graph.Node) bool

// HasEdgeFromTo returns whether an edge exists in the graph from u to v.
func (g *DependencyGraph) HasEdgeFromTo(u, v graph.Node) bool

// Edge returns the edge from u to v if such an edge exists and nil
// otherwise. The node v must be directly reachable from u as defined
// by the From method.
func (g *DependencyGraph) Edge(u, v graph.Node) graph.Edge {
	return g.edges[u][v]
}
