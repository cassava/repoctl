package graph

func (g *Graph) Dependencies() []string {
	nodes := AllNodesBottomUp(g)
	list := make([]string, len(nodes))
	for i, n := range nodes {
		list[i] = n.(*Node).PkgName()
	}
	return list
}

func (g *Graph) DependenciesFromAUR() []string {
	nodes := AllNodesBottomUp(g)
	list := make([]string, 0, len(nodes))
	for _, n := range nodes {
		if n.(*Node).IsFromAUR() {
			list = append(list, n.(*Node).PkgName())
		}
	}
	return list
}

func (g *Graph) DependenciesFromRepos() []string {
	nodes := AllNodesBottomUp(g)
	list := make([]string, 0, len(nodes))
	for _, n := range nodes {
		if n.(*Node).IsFromAUR() {
			list = append(list, n.(*Node).PkgName())
		}
	}
	return list
}
