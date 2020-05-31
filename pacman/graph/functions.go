package graph

import (
	"github.com/cassava/repoctl/pacman"
	"github.com/cassava/repoctl/pacman/aur"

	"gonum.org/v1/gonum/graph/topo"
)

// Dependencies returns a list of all dependencies in the graph,
// those in repositories, those from AUR, and those unknown.
func Dependencies(g *Graph) (pacman.Packages, aur.Packages, []string) {
	rps := make(pacman.Packages, 0)
	aps := make(aur.Packages, 0)
	ups := make([]string, 0)

	names := make(map[string]bool)
	nodes, _ := topo.Sort(g)
	for _, vn := range nodes {
		n := vn.(*Node)
		if names[n.PkgName()] {
			continue
		}

		names[n.PkgName()] = true
		switch p := n.AnyPackage.(type) {
		case *aur.Package:
			aps = append(aps, p)
		case *pacman.Package:
			if p.Origin == pacman.UnknownOrigin {
				ups = append(ups, p.Name)
			} else {
				rps = append(rps, p)
			}
		default:
			panic("unexpected type of package in graph")
		}
	}
	return rps, aps, ups
}
