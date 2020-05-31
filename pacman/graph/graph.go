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
// with the following function:
//
//   graph.Dependencies(g)
package graph

import (
	"github.com/cassava/repoctl/pacman"
	"github.com/cassava/repoctl/pacman/aur"

	"gonum.org/v1/gonum/graph"
	"gonum.org/v1/gonum/graph/simple"
)

// Node implements graph.Node.
type Node struct {
	simple.Node
	pacman.AnyPackage
}

// IsFromAUR returns whether the node comes from AUR.
func (n *Node) IsFromAUR() bool {
	_, ok := n.AnyPackage.(*aur.Package)
	return ok
}

// AllDepends returns a (newly created) string slice of the installation
// and make dependencies of this package.
func (n *Node) AllDepends() []string {
	deps := make([]string, 0, n.NumAllDepends())
	deps = append(deps, n.PkgDepends()...)
	deps = append(deps, n.PkgMakeDepends()...)
	return deps
}

// NumAllDepends returns the number of make and installation dependencies the
// package has.
func (n *Node) NumAllDepends() int {
	return len(n.PkgDepends()) + len(n.PkgMakeDepends())
}

func (n *Node) String() string {
	return n.PkgName()
}

// Graph implements graph.Graph.
type Graph struct {
	*simple.DirectedGraph

	names map[string]int64
}

// NewGraph returns a new graph.
func NewGraph() *Graph {
	return &Graph{
		DirectedGraph: simple.NewDirectedGraph(),
		names:         make(map[string]int64),
	}
}

// NewNode returns a new node.
func (g *Graph) NewNode(pkg pacman.AnyPackage) *Node {
	return &Node{
		Node:       g.DirectedGraph.NewNode().(simple.Node),
		AnyPackage: pkg,
	}
}

// Has returns whether the node exists within the graph.
func (g *Graph) Has(id int64) bool {
	return g.Node(id) != nil
}

// HasName returns whether the package with the given name exists within the
// graph.
func (g *Graph) HasName(name string) bool {
	_, ok := g.names[name]
	return ok
}

// NodeWithName returns the node with the given name, or nil.
func (g *Graph) NodeWithName(name string) *Node {
	id, ok := g.names[name]
	if !ok {
		return nil
	}
	return g.Node(id).(*Node)
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
	if g.Has(n.ID()) {
		panic("node id already here")
	}

	g.DirectedGraph.AddNode(v)
	g.names[n.PkgName()] = n.ID()
}

// AddEdgeFromTo adds a directed edge from u to v.
func (g *Graph) AddEdgeFromTo(u, v graph.Node) {
	g.SetEdge(g.NewEdge(u, v))
}
