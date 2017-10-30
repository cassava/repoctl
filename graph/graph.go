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
	"os"

	"github.com/gonum/graph"
	"github.com/goulash/errs"
	"github.com/goulash/pacman"
	"github.com/goulash/pacman/aur"
)

// Node implements graph.Node.
type Node struct {
	id int

	*pacman.Package
	fromAUR bool

	// TODO: Refactor this out.
	deps map[string]bool
}

// ID returns the unique (within the graph) ID of the node.
func (n *Node) ID() int { return n.id }

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

// IsInAUR returns true if the dependency needs to be fetched from AUR.
func (e *Edge) IsInAUR() bool { return e.to.fromAUR }

// Graph implements graph.Graph.
type Graph struct {
	// These are used for resolving dependencies:
	local map[string]*pacman.Package
	sync  map[string]*pacman.Package

	// For the interface:
	names     map[string]*Node
	nodes     []graph.Node
	nodeIDs   map[int]graph.Node
	edgesFrom map[int][]graph.Node
	edgesTo   map[int][]graph.Node
	edges     map[int]map[int]graph.Edge
	nextID    int

	// Statistics
	aurCalls int
}

// NewGraph returns a new dependency graph, ignoring repositories given in `ignore`.
//
// Note that it makes a difference between dependencies that are in AUR, and those
// availabe in the repositories. Ignoring repositories effectively "demotes" any
// packages available in them back to AUR.
//
// Any package that is in the dependency graph that is not from AUR is treated
// as a leaf in the graph, since we assume that pacman can resolve those
// dependencies.
func NewGraph(ignore ...string) (*Graph, error) {
	g := Graph{
		// local will be init in this function
		// sync will be init in this function

		names:     make(map[string]*Node),
		nodes:     make([]graph.Node, 0),
		nodeIDs:   make(map[int]graph.Node),
		edgesFrom: make(map[int][]graph.Node),
		edgesTo:   make(map[int][]graph.Node),
		edges:     make(map[int]map[int]graph.Edge),
		nextID:    0,

		aurCalls: 0,
	}

	// Read local database
	lpkgs, err := pacman.ReadLocalDatabase(errs.Print(os.Stderr))
	if err != nil {
		return nil, err
	}
	g.local = lpkgs.ToMap()

	// Read available packages
	var pkgs pacman.Packages
	if len(ignore) == 0 {
		pkgs, err = pacman.ReadAllSyncDatabases()
		if err != nil {
			return nil, err
		}
	} else {
		enabled, err := pacman.EnabledRepositories()
		if err != nil {
			return nil, err
		}
	nextRepo:
		for _, repo := range enabled {
			for _, ig := range ignore {
				if repo == ig {
					continue nextRepo
				}
			}

			rpkgs, err := pacman.ReadSyncDatabase(repo)
			if err != nil {
				return nil, err
			}
			pkgs = append(pkgs, rpkgs...)
		}
	}
	g.sync = pkgs.ToMap()

	return &g, nil
}

// NewNodeFromAUR returns a new node but does not otherwise modify the graph.
func (g *Graph) NewNodeFromAUR(pkg *aur.Package) *Node {
	return g.NewNode(pkg.Pkg(), true)
}

// NewNode returns a new node but does not otherwise modify the graph.
func (g *Graph) NewNode(pkg *pacman.Package, fromAUR bool) *Node {
	n := &Node{
		id:      g.NewNodeID(),
		Package: pkg,
		fromAUR: fromAUR,
		deps:    make(map[string]bool),
	}

	if !fromAUR {
		// If this node is not from AUR, then it's from a repository,
		// which means we don't have to build it and pacman will resolve
		// the dependencies.
		return n
	}

	// addDep adds the given dependency into deps if it is not
	// already installed.
	addDeps := func(deps []string) {
		for _, dep := range deps {
			if _, ok := g.local[dep]; ok {
				return
			}
			n.deps[dep] = true
		}
	}

	addDeps(pkg.MakeDepends)
	addDeps(pkg.Depends)
	return n
}

// NewNodeID returns a unique ID for a new node.
func (g *Graph) NewNodeID() int {
	g.nextID++
	return g.nextID
}

// AddNode adds the given node to the graph, creating new nodes and
// resolving their dependencies as required.
func (g *Graph) AddNode(node graph.Node) error {
	// Checking preconditions:
	n, ok := node.(*Node)
	if !ok {
		panic("only accept our own nodes")
	}
	if g.HasName(n.PkgName()) {
		panic("package name already in graph")
	}
	if g.Has(node) {
		panic("node id already here")
	}

	// To avoid running in infinite circles if there is a circular
	// dependency, we add the node now to the graph before resolving
	// dependencies.
	g.names[n.PkgName()] = n
	g.nodes = append(g.nodes, n)
	g.nodeIDs[n.ID()] = n

	// Resolving dependencies:
	unavailable := make([]string, 0)
	for d := range n.deps {
		if g.HasName(d) {
			// Dependency already in the graph.
			continue
		}

		if p, ok := g.sync[d]; ok {
			// Dependency is available in a repository, and is a leaf.
			err := g.AddNode(g.NewNode(p, false))
			if err != nil {
				panic("unexpected error adding synced dependency")
			}
		} else {
			// Dependency is not available and is hopefully in AUR.
			unavailable = append(unavailable, d)
		}
	}

	// Get unavailable dependencies and add them to the graph.
	g.aurCalls++
	pkgs, err := aur.ReadAll(unavailable)
	if err != nil {
		return err
	}
	if err := g.AddPackagesFromAUR(pkgs); err != nil {
		return err
	}

	// Add edges to the graph.
	g.setEdgesFrom(n)

	return nil
}

// AddPackagesFromAUR adds the packages to the graph as nodes.
func (g *Graph) AddPackagesFromAUR(pkgs aur.Packages) error {
	for _, p := range pkgs {
		err := g.AddNode(g.NewNodeFromAUR(p))
		if err != nil {
			return err
		}
	}
	g.fixDependenciesFromAUR()

	return nil
}

func (g *Graph) fixEdgeFromTo(from, to *Node) {
	from.deps[to.PkgName()] = true
	g.addEdgeFromTo(from, to)
}

func (g *Graph) addEdgeFromTo(from, to *Node) {
	fid := from.ID()
	g.edgesFrom[fid] = append(g.edgesFrom[fid], to)
	tid := to.ID()
	g.edgesTo[tid] = append(g.edgesTo[tid], from)
	// FIXME: Continue here by adding g.edges entry!
}

// setEdgesFrom should be called only once per node.
func (g *Graph) setEdgesFrom(v *Node) {
	if g.edges[v.ID()] != nil {
		panic("setEdgesFrom should only be called once per node")
	}

	g.edges[v.ID()] = make(map[int]graph.Edge)
	dnodes := make([]graph.Node, 0, len(v.deps))
	for d := range v.deps {
		dn := g.names[d]
		dnodes = append(dnodes, dn)
		g.edgesTo[dn.ID()] = append(g.edgesTo[dn.ID()], v)
	}
	g.edgesFrom[v.ID()] = dnodes
}

// fixDependenciesFromAUR adds depdency edges back in if they relate to AUR packages.
//
// Rationale: Dependencies are lost between AUR packages if they are already installed,
// but if they were explicitely added (i.e. `Node.fromAUR==true`), then we want these
// dependencies to remain, so that the graph is ordered correctly.
//
// This operation takes O(n) time.
func (g *Graph) fixDependenciesFromAUR() {
	fix := func(v *Node, deps []string) {
		for _, d := range deps {
			dn, ok := g.names[d]
			if !ok {
				panic("unexpected error")
			}

			// Ignore non-AUR dependencies
			if !dn.fromAUR {
				continue
			}

			// Fix if edge is missing
			if !g.HasEdgeFromTo(v, dn) {
				g.fixEdgeFromTo(v, dn)
			}

		}
	}
	for _, n := range g.names {
		if !n.fromAUR {
			continue
		}

		fix(n, n.Package.Depends)
		fix(n, n.Package.MakeDepends)
	}
}

// Has returns whether the node exists within the graph.
func (g *Graph) Has(n graph.Node) bool {
	_, ok := g.nodeIDs[n.ID()]
	return ok
}

// HasName returns whether a particular package name is in the graph.
func (g *Graph) HasName(key string) bool {
	_, ok := g.names[key]
	return ok
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

}

func (g *Graph) Dependencies() []string {
	nodes := AllNodesBottomUp(g)
	list := make([]string, len(nodes))
	for i, n := range nodes {
		list[i] = n.PkgName()
	}
	return list
}

func (g *Graph) DependenciesFromAUR() []string {
	nodes := AllNodesBottomUp(g)
	list := make([]string, 0, len(nodes))
	for i, n := range nodes {
		if n.fromAUR {
			list = append(list, n.PkgName())
		}
	}
	return list
}

func (g *Graph) DependenciesFromRepos() []string {
	nodes := AllNodesBottomUp(g)
	list := make([]string, 0, len(nodes))
	for i, n := range nodes {
		if n.fromAUR {
			list = append(list, n.PkgName())
		}
	}
	return list
}
