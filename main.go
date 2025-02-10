package main

import (
	"cigarsdb/extract/noblego"
	"cigarsdb/storage"
	"cigarsdb/storage/fs"
	"context"
	"flag"
	"log/slog"
	"net/http"
	"os"
)

func main() {
	var (
		dumpDir                 string
		limit, pageMin, pageMax uint
	)
	flag.StringVar(&dumpDir, "o", "/tmp", "output directory")
	flag.UintVar(&limit, "limit", 100, "fetch limit per page")
	flag.UintVar(&pageMin, "page-min", 1, "fetch starting from this page number")
	flag.UintVar(&pageMax, "page-max", 0, "fetch until this page number is reached")
	flag.Parse()

	var logs = slog.New(slog.NewJSONHandler(os.Stdin, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	destination, err := fs.NewClient(dumpDir)
	if err != nil {
		logs.Error("could not init the writer", slog.Any("error", err))
		return
	}

	source := noblego.Client{HTTPClient: http.DefaultClient}

	var rec []storage.Record
	var nextPage uint
	page := pageMin
	pMax := pageMax
	if pageMax == 0 {
		pMax = page + 1
	}
	ctx := context.Background()
	for page < pMax {
		rec, nextPage, err = source.ReadBulk(ctx, limit, page)
		if err != nil {
			logs.Error("error fetching data", slog.Any("error", err), slog.Uint64("page", uint64(page)))
			return
		}
		_, err = destination.WriteBulk(ctx, rec)
		if err != nil {
			logs.Error("error persisting the data", slog.Any("error", err),
				slog.Uint64("page", uint64(page)))
			return
		}
		if nextPage > page {
			page = nextPage
		}
		if pageMax == 0 {
			pMax = page + 1
		}
	}
}
