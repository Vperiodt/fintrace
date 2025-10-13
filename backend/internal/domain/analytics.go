package domain

// PathNode represents a node within a graph path.
type PathNode struct {
	ID     string
	Type   string
	Label  string
	Weight float64
}

// PathEdge represents an edge between two nodes in a path.
type PathEdge struct {
	Type   string
	Source string
	Target string
	Label  string
	Weight float64
}

// ShortestPath encapsulates nodes and edges connecting a source and target user.
type ShortestPath struct {
	SourceUserID string
	TargetUserID string
	Nodes        []PathNode
	Edges        []PathEdge
	Hops         int
}

