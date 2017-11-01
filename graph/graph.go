// Package graph provides dependency resolution for AUR packages.
// It is not very optimized, but it should work (hopefully).
//
// Usage
//
// First, you have to create a graph:
//
//   pkgs, err := aur.ReadAll(list)
//   if err != nil {
//      return err
//   }
//   g, err := graph.NewGraph()
//   if err != nil {
//      return err
//   }
//
// Once you have a graph, you can then get the ordered dependency list
// with one of the following functions:
//
//   g.Dependencies()
//   g.DependenciesFromAUR()
//   g.DependenciesFromRepos()
package graph

import (
	"github.com/gonum/graph"
	"github.com/goulash/pacman"
)

// Node implements graph.Node.
type Node struct {
	id int

	*pacman.Package
}

// ID returns the unique (within the graph) ID of the node.
func (n *Node) ID() int { return n.id }

// IsFromAUR returns whether the node comes from AUR.
func (n *Node) IsFromAUR() bool {
	return n.Package.Origin == pacman.AUROrigin
}

// Dependencies returns a (newly created) string slice of the installation
// and make dependencies of this package.
func (n *Node) Dependencies() []string {
	deps := make([]string, 0, len(n.Package.Depends)+len(n.Package.MakeDepends))
	deps = append(deps, n.Package.Depends...)
	deps = append(deps, n.Package.MakeDepends...)
	return deps
}

// NumDependencies returns the number of make and installation dependencies the package has.
func (n *Node) NumDependencies() int {
	return len(n.Package.Depends) + len(n.Package.MakeDepends)
}

// Edge implements the graph.Edge interface.
type Edge struct {
	from *Node
	to   *Node
}

// From returns the node that has the dependency.
func (e *Edge) From() graph.Node { return e.from }

// To returns the depdency that the from node has.
func (e *Edge) To() graph.Node { return e.to }

// Weight returns zero, because depdencies are not weighted.
func (e *Edge) Weight() float64 { return 0.0 }

// IsFromAUR returns true if the dependency needs to be fetched from AUR.
func (e *Edge) IsFromAUR() bool { return e.to.IsFromAUR() }

// Graph implements graph.Graph.
type Graph struct {
	nodes     []graph.Node
	nodeIDs   map[int]graph.Node
	edgesFrom map[int][]graph.Node
	edgesTo   map[int][]graph.Node
	edges     map[int]map[int]graph.Edge
	nextID    int
}

// NewGraph returns a new graph.
func NewGraph() *Graph {
	return &Graph{
		names:     make(map[string]*Node),
		nodes:     make([]graph.Node, 0),
		nodeIDs:   make(map[int]graph.Node),
		edgesFrom: make(map[int][]graph.Node),
		edgesTo:   make(map[int][]graph.Node),
		edges:     make(map[int]map[int]graph.Edge),
		nextID:    0,
	}
}

// Has returns whether the node exists within the graph.
func (g *Graph) Has(n graph.Node) bool {
	_, ok := g.nodeIDs[n.ID()]
	return ok
}

// HasPkgName returns whether the package with the given name exists within the
// graph.
func (g *Graph) HasPkgName(name string) bool {
	_, ok := g.names[name]
	return ok
}

// NodeWithName returns the node with the given name, or nil.
func (g *Graph) NodeWithName(name string) *Node {
	return g.names[name]
}

// Nodes returns all the nodes in the graph.
func (g *Graph) Nodes() []graph.Node {
	return g.nodes
}

// From returns all nodes that can be reached directly from the given node.
func (g *Graph) From(v graph.Node) []graph.Node {
	return g.edgesFrom[v.ID()]
}

// To returns all nodes that can reach directly to the given node.
func (g *Graph) To(v graph.Node) []graph.Node {
	return g.edgesTo[v.ID()]
}

// HasEdgeBetween returns whether an edge exists between nodes u and v
// without considering direction.
func (g *Graph) HasEdgeBetween(u, v graph.Node) bool {
	return g.HasEdgeFromTo(u, v) || g.HasEdgeFromTo(v, u)
}

// HasEdgeFromTo returns whether an edge exists in the graph from u to v.
func (g *Graph) HasEdgeFromTo(u, v graph.Node) bool {
	for _, n := range g.edgesFrom[u.ID()] {
		if n == v {
			return true
		}
	}
	return false
}

// Edge returns the edge from u to v if such an edge exists and nil
// otherwise. The node v must be directly reachable from u as defined
// by the From method.
func (g *Graph) Edge(u, v graph.Node) graph.Edge {
	return g.edges[u][v]
}

// NewNodeID returns a unique ID for a new node.
func (g *Graph) NewNodeID() int {
	g.nextID++
	return g.nextID
}

// NewNode returns a new node.
func (g *Graph) NewNode(pkg *pacman.Package) *Node {
	return &Node{
		id:      g.NewNodeID(),
		Package: pkg,
	}
}

// AddNode adds the node and initializes data structures but does nothing else.
func (g *Graph) AddNode(v graph.Node) {
	// Checking preconditions:
	n, ok := v.(*Node)
	if !ok {
		panic("only accept our own nodes")
	}
	if g.HasName(n.PkgName()) {
		panic("package name already in graph")
	}
	if g.Has(node) {
		panic("node id already here")
	}

	g.names[n.Pkgname()] = n
	g.nodes = append(g.nodes, n)
	id := n.ID()
	g.nodeIDs[id] = n
	g.edgesFrom[id] = make([]graph.Node, 0, n.NumDependencies())
	g.edgesTo[id] = make([]graph.Node, 0)
	g.edges[id] = make(map[int]graph.Edge)
}

// AddEdgeFromTo adds an edge betwewen the two nodes.
func (g *Graph) AddEdgeFromTo(u, v graph.Node) {
	g.edges[uid][vid] = &Edge{from: u, to: v}
	g.edgesFrom[uid] = append(g.edgesFrom[uid], u)
	g.edgesTo[vid] = append(g.edgesTo[vid], v)
}
