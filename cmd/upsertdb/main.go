package main

import (
	"cigarsdb/storage"
	fsClient "cigarsdb/storage/fs"
	"cigarsdb/storage/neo4j"
	"cmp"
	"context"
	"flag"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"strings"
)

func main() {
	var sourceDir string
	flag.StringVar(&sourceDir, "i", "/tmp", "directory to read the json files from")
	flag.Parse()

	var logs = slog.New(slog.NewJSONHandler(os.Stdin, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	from, err := fsClient.NewClient(sourceDir)
	if err != nil {
		logs.Error("could not initialise the FS reading client", slog.Any("error", err))
		return
	}

	c := neo4j.ConnectionConfig{
		DbURI:      os.Getenv("DB_URI"),
		DbPassword: os.Getenv("DB_PASSWORD"),
		DbName:     cmp.Or(os.Getenv("DB_NAME"), "neo4j"),
		DbUser:     cmp.Or(os.Getenv("DB_USER"), "neo4j"),
	}

	ctx := context.Background()

	to, err := neo4j.NewClient(ctx, c)
	if err != nil {
		logs.Error("could not initialise the neo4j writing client", slog.Any("error", err))
		return
	}

	err = filepath.WalkDir(sourceDir, func(p string, d fs.DirEntry, err error) error {
		if err == nil {
			if strings.HasSuffix(p, ".json") && !d.IsDir() {
				id := strings.TrimRight(path.Base(p), ".json")
				rec, er := from.Read(ctx, id)
				if er != nil {
					err = fmt.Errorf("could not read the file %s: %w", p, er)
				} else {
					if _, er = to.Write(ctx, []storage.Record{rec}); er != nil {
						err = fmt.Errorf("could not write the file %s: %w", p, er)
					}
				}
			}
		}
		return err
	})
	if err != nil {
		logs.Error("uploading error", slog.Any("error", err))
		return
	}
}
