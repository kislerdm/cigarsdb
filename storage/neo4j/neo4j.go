package neo4j

import (
	"cigarsdb/storage"
	"context"
	"fmt"
	"reflect"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

type ConnectionConfig struct {
	DbURI      string
	DbPassword string
	DbName     string
	DbUser     string
}

type Client struct {
	dbSession neo4j.SessionWithContext
}

func NewClient(ctx context.Context, cfg ConnectionConfig) (c Client, err error) {
	d, err := neo4j.NewDriverWithContext(cfg.DbURI, neo4j.BasicAuth(cfg.DbUser, cfg.DbPassword, ""))
	switch err != nil {
	case true:
		err = fmt.Errorf("could not init neo4j driver: %w", err)
	case false:

		if err = d.VerifyConnectivity(ctx); err != nil {
			err = fmt.Errorf("could not connect to neo4j: %w", err)
		} else {
			sess := d.NewSession(ctx, neo4j.SessionConfig{DatabaseName: cfg.DbName, ImpersonatedUser: cfg.DbUser})
			c = Client{dbSession: sess}
		}
	}

	return c, err
}

func (c Client) Write(ctx context.Context, r []storage.Record) (ids []string, err error) {
	_, err = c.dbSession.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {
		records := fromRecords(r)
		return tx.Run(ctx, `WITH $records AS records, apoc.date.currentTimestamp() AS now
UNWIND records AS rec
MERGE (n:CigarRaw{identifier:apoc.map.get(rec, "url", "", false)}) SET n += rec
, n.createdAt = apoc.map.get(n, "createdAt", now, false)
, n.updatedAt = now
`, map[string]interface{}{
			"records": records,
		})
	})

	if err == nil {
		ids = make([]string, len(r))
		for i, rec := range r {
			ids[i] = rec.URL
		}
	}
	return ids, err
}

func fromRecords(r []storage.Record) []map[string]any {
	var o = make([]map[string]any, 0, len(r))
	for _, rec := range r {
		t := reflect.TypeOf(rec)
		v := reflect.ValueOf(rec)
		for i := 0; i < t.NumField(); i++ {
			_ = v.Field(i)
			panic("todo")
		}

		el := map[string]any{"identifier": rec.URL}
		o = append(o, el)
	}
	return o
}
