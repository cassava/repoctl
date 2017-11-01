package graph

import (
	"github.com/gonum/graph"
	"github.com/gonum/graph/traverse"
)

// Dependencies returns a list of all dependencies in the graph.
func Dependencies(g *Graph) []string {
	nodes := AllNodesBottomUp(g)
	list := make([]string, len(nodes))
	for i, n := range nodes {
		list[i] = n.(*Node).PkgName()
	}
	return list
}

// AURDependencies returns a list of dependencies that need to be fetched from AUR.
func AURDependencies(g *Graph) []string {
	nodes := AllNodesBottomUp(g)
	list := make([]string, 0, len(nodes))
	for _, n := range nodes {
		if n.(*Node).IsFromAUR() {
			list = append(list, n.(*Node).PkgName())
		}
	}
	return list
}

// RepoDependencies returns a list of dependencies that can be installed with pacman.
func RepoDependencies(g *Graph) []string {
	nodes := AllNodesBottomUp(g)
	list := make([]string, 0, len(nodes))
	for _, n := range nodes {
		if n.(*Node).IsFromAUR() {
			list = append(list, n.(*Node).PkgName())
		}
	}
	return list
}

// Roots returns all the root nodes for a directed graph.
func Roots(g graph.Directed) []graph.Node {
	roots := make([]graph.Node, 0)
	for _, n := range g.Nodes() {
		if g.To(n) == nil || len(g.To(n)) == 0 {
			roots = append(roots, n)
		}
	}
	return roots
}

// NodesBottomUp returns the subtree from the bottom levels upwards to the root.
// The nodes may appear multiple times however.
func NodesBottomUp(g graph.Directed, root graph.Node) []graph.Node {
	nodes := make([]graph.Node, 0)
	nodes = append(nodes, root)
	bfs := traverse.BreadthFirst{}
	bfs.Walk(g, root, func(v graph.Node, _ int) bool {
		nodes = append(nodes, v)
		return true
	})

	// Reverse the list
	sz := len(nodes)
	last := sz - 1
	for i := 0; i < sz/2; i++ {
		tmp := nodes[i]
		nodes[i] = nodes[last-i]
		nodes[last-i] = tmp
	}
	return uniqueNodes(nodes)
}

// AllNodesBottomUp returns for all roots the nodes bottom-up.
func AllNodesBottomUp(g graph.Directed) []graph.Node {
	nodes := make([]graph.Node, 0)
	for _, root := range Roots(g) {
		nodes = append(nodes, NodesBottomUp(g, root)...)
	}
	return uniqueNodes(nodes)
}

func uniqueStrings(list []string) []string {
	xs := make(map[string]bool)
	lst := make([]string, 0, len(list))
	for _, x := range list {
		if !xs[x] {
			xs[x] = true
			lst = append(lst, x)
		}
	}
	return lst
}

func uniqueNodes(list []graph.Node) []graph.Node {
	xs := make(map[int]bool)
	lst := make([]graph.Node, 0, len(list))
	for _, x := range list {
		if !xs[x.ID()] {
			xs[x.ID()] = true
			lst = append(lst, x)
		}
	}
	return lst
}
