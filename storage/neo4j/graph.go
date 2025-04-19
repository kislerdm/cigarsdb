package neo4j

import (
	"cigarsdb/storage"
	"maps"
)

type Node struct {
	Identifier string         `json:"identifier"`
	Type       string         `json:"type"`
	CreatedAt  int64          `json:"-"`
	UpdatedAt  int64          `json:"-"`
	Parameters map[string]any `json:"parameters"`
}

const (
	neo4jPropIdentifier = "identifier"
	neo4jPropType       = "type"
	neo4jPropCreatedAt  = "createdAt"
	neo4jPropUpdatedAt  = "updatedAt"
)

func (r Node) toWriteObject() map[string]any {
	// 4 = identifier+type+createdAt+updatedAt
	var o = make(map[string]any, 4+len(r.Parameters))
	o[neo4jPropIdentifier] = r.Identifier
	o[neo4jPropType] = r.Type
	o[neo4jPropCreatedAt] = r.CreatedAt
	o[neo4jPropUpdatedAt] = r.UpdatedAt
	maps.Copy(o, r.Parameters)
	return o
}

type Edge struct {
	Identifier string         `json:"identifier"`
	Type       string         `json:"type"`
	FromID     string         `json:"fromID"`
	FromType   string         `json:"-"`
	ToID       string         `json:"toID"`
	ToType     string         `json:"-"`
	CreatedAt  int64          `json:"-"`
	UpdatedAt  int64          `json:"-"`
	Parameters map[string]any `json:"parameters"`
}

const (
	neo4jEdgePropFromID   = "fromID"
	neo4jEdgePropFromType = "fromType"
	neo4jEdgePropToID     = "toID"
	neo4jEdgePropToType   = "toType"
)

func (r Edge) toWriteObject() map[string]any {
	// 8 = identifier+type+fromID+toID+fromType+toType+createdAt+updatedAt
	var o = make(map[string]any, 8+len(r.Parameters))
	o[neo4jPropIdentifier] = r.Identifier
	o[neo4jPropType] = r.Type
	o[neo4jPropCreatedAt] = r.CreatedAt
	o[neo4jPropUpdatedAt] = r.UpdatedAt
	o[neo4jEdgePropFromID] = r.FromID
	o[neo4jEdgePropFromType] = r.FromType
	o[neo4jEdgePropToID] = r.ToID
	o[neo4jEdgePropToType] = r.ToType
	maps.Copy(o, r.Parameters)
	return o
}

type Graph struct {
	Nodes map[string]Node
	Edges map[string]Edge
}

func newGraph(r storage.Record) Graph {
	panic("todo")
}
