package deps

// Definition represents a named proto construct with its dependencies.
type Definition struct {
	FQN        string   // Fully qualified name (e.g., "my.package.OrderService")
	Kind       string   // "service", "message", "enum"
	File       string   // Relative path of the containing file
	References []string // FQNs of types this definition depends on
}

// Graph tracks definitions and their dependency relationships.
type Graph struct {
	Nodes   map[string]*Definition // FQN → Definition
	Edges   map[string][]string    // FQN → list of FQNs it depends on
	FileMap map[string]string      // FQN → relative file path
}

// NewGraph creates an empty dependency graph.
func NewGraph() *Graph {
	return &Graph{
		Nodes:   make(map[string]*Definition),
		Edges:   make(map[string][]string),
		FileMap: make(map[string]string),
	}
}

// AddDefinition registers a definition in the graph.
func (g *Graph) AddDefinition(d *Definition) {
	g.Nodes[d.FQN] = d
	g.Edges[d.FQN] = d.References
	g.FileMap[d.FQN] = d.File
}

// TransitiveDeps returns all FQNs transitively required by the given
// set of FQNs, using BFS. The input FQNs are included in the result.
func (g *Graph) TransitiveDeps(fqns []string) []string {
	visited := make(map[string]bool)
	queue := make([]string, 0, len(fqns))

	for _, fqn := range fqns {
		if !visited[fqn] {
			visited[fqn] = true
			queue = append(queue, fqn)
		}
	}

	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		for _, dep := range g.Edges[current] {
			if !visited[dep] {
				visited[dep] = true
				queue = append(queue, dep)
			}
		}
	}

	result := make([]string, 0, len(visited))
	for fqn := range visited {
		result = append(result, fqn)
	}
	return result
}

// RequiredFiles returns all file paths that must appear in output
// to satisfy the given set of FQNs.
func (g *Graph) RequiredFiles(fqns []string) []string {
	fileSet := make(map[string]bool)
	for _, fqn := range fqns {
		if f, ok := g.FileMap[fqn]; ok {
			fileSet[f] = true
		}
	}
	files := make([]string, 0, len(fileSet))
	for f := range fileSet {
		files = append(files, f)
	}
	return files
}
