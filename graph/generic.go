package graph

import (
	"github.com/gonum/graph"
	"github.com/gonum/graph/traverse"
)

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
func NodesBottomUp(g graph.Directed, root graph.Node) []graph.Node {
	nodes := make([]graph.Node, 0)
	bfs := traverse.BreadthFirst{}
	bfs.Walk(g, root, func(v graph.Node, _ int) bool {
		nodes = append(nodes, v)
		return true
	})
	nodes = append(nodes, root)

	// Reverse the list
	sz := len(nodes)
	last := sz - 1
	for i := 0; i < sz/2; i++ {
		tmp := nodes[i]
		nodes[i] = nodes[last-i]
		nodes[last-i] = tmp
	}
	return nodes
}

// AllNodesBottomUp returns for all roots the nodes bottom-up.
func AllNodesBottomUp(g graph.Directed) []graph.Node {
	nodes := make([]graph.Node, 0)
	for _, root := range Roots(g) {
		nodes = append(nodes, NodesBottomUp(g, root)...)
	}
	return nodes
}
