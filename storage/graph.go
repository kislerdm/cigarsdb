package storage

type Node struct {
	Type       string
	Properties map[string]any
}

type Graph struct {
	Nodes map[string]Node
}
