package neo4j

import (
	"cigarsdb/storage"
	"context"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type Client struct {
	dbSession neo4j.SessionWithContext
}

func (c Client) Write(ctx context.Context, r storage.Record) (id string, err error) {
	//TODO implement me
	panic("implement me")
}

func (c Client) WriteBulk(ctx context.Context, r []storage.Record) (ids []string, err error) {
	//TODO implement me
	panic("implement me")
}
