package neo4j

import (
	"cigarsdb/storage"
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

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
	var records = make([]map[string]any, len(r))
	for i, rec := range r {
		records[i], err = fromRecord(rec)
		if err != nil {
			err = fmt.Errorf("could not convert record %d to neo4j record: %w", i, err)
			break
		}
	}

	if err == nil {
		_, err = c.dbSession.ExecuteWrite(ctx, func(tx neo4j.ManagedTransaction) (any, error) {

			return tx.Run(ctx, `WITH $records AS records, apoc.date.currentTimestamp() AS now
UNWIND records AS rec
MERGE (n:CigarRaw{identifier:apoc.map.get(rec, "url", "", false)}) SET n += rec
, n.createdAt = apoc.map.get(n, "createdAt", now, false)
, n.updatedAt = now
`, map[string]interface{}{
				"records": records,
			})
		})
	}

	if err == nil {
		ids = make([]string, len(r))
		for i, rec := range r {
			ids[i] = rec.URL
		}
	}

	return ids, err
}

func fromRecord(rec any) (map[string]any, error) {
	var err error
	var o = make(map[string]any)

	t := reflect.TypeOf(rec)
	if t.Kind() != reflect.Struct {
		err = fmt.Errorf("imput must be struct")
	}

	if err == nil {
		v := reflect.ValueOf(rec)
		for i := 0; i < t.NumField(); i++ {
			fieldType := t.Field(i)
			key := strings.Split(fieldType.Tag.Get("json"), ",")[0]
			fieldVal := v.Field(i)
			switch fieldVal.Kind() {
			case reflect.Struct:
				if o[key], err = fromRecord(fieldVal.Interface()); err != nil {
					break
				}

			case reflect.Pointer:
				if !fieldVal.IsNil() && !fieldVal.IsZero() {
					elVal := fieldVal.Elem().Interface()
					if fieldVal.Elem().Kind() == reflect.Struct {
						if o[key], err = fromRecord(elVal); err != nil {
							break
						}
					} else {
						o[key] = elVal
					}
				}

			case reflect.Slice, reflect.Array:
				if fieldVal.Len() > 0 {
					switch fieldVal.Index(0).Kind() {
					case reflect.Struct:
						var vv = make([]map[string]any, fieldVal.Len())
						for j := 0; j < fieldVal.Len(); j++ {
							el := fieldVal.Index(j)
							vv[j], err = fromRecord(el.Interface())
							if err != nil {
								break
							}
						}
						o[key] = vv

					default:
						o[key] = fieldVal.Interface()
					}
				}

			case reflect.Map:
				if len(fieldVal.MapKeys()) > 0 {
					var vv []byte
					vv, err = json.Marshal(fieldVal.Interface())
					o[key] = string(vv)
				}

			default:
				o[key] = fieldVal.Interface()
			}
		}
	}

	return o, err
}
