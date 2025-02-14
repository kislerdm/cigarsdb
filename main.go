package main

import (
	"cigarsdb/extract/cigarworld"
	"cigarsdb/extract/noblego"
	"cigarsdb/storage"
	"cigarsdb/storage/fs"
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"sync"
	"time"
)

var version = "dev"

func showVersion() bool {
	var ok bool
	for _, arg := range os.Args[1:] {
		switch arg {
		case "version", "V", "-v", "-version", "--version":
			_, _ = fmt.Fprintf(os.Stdout, "version: %s\n", version)
			ok = true
		}
	}
	return ok
}

func main() {
	if showVersion() {
		return
	}

	var (
		dumpDir        string
		limit, pageMin uint
		pageMax        uint
		s              string
	)
	flag.StringVar(&s, "i", "", "source")
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

	source, err := newSource(s, logs, destination)
	if err != nil {
		logs.Error("could not initialise the source fetching client", slog.Any("error", err))
		return
	}

	page := pageMin
	ctx := context.Background()
	for page > 0 {
		logs.Info("start fetching", slog.Uint64("page", uint64(page)))

		rec, nextPage, err := source.ReadBulk(ctx, limit, page)
		if err != nil {
			logs.Error("error fetching data", slog.Any("error", err), slog.Uint64("page", uint64(page)))
			return
		}
		_, err = destination.WriteBulk(ctx, rec)
		if err != nil {
			logs.Error("error persisting the data", slog.Any("error", err),
				slog.Uint64("page", uint64(page)))
		}

		logs.Info("end fetching", slog.Uint64("page", uint64(page)))

		if pageMax > 0 && page >= pageMax {
			break
		}

		page = nextPage
	}
}

func newSource(s string, logs *slog.Logger, writer storage.Writer) (source storage.Reader, err error) {
	c := newHTTPClient(5*time.Second, 5)

	switch s {
	case "noblego":
		source = noblego.Client{HTTPClient: c}
	case "cigarworld":
		source = cigarworld.Client{HTTPClient: c, Dumper: writer, Logs: logs}
	default:
		err = fmt.Errorf("data source is unknown")
	}
	return source, err
}

type httpClient struct {
	InitialDelay time.Duration
	MaxRetries   uint8
	attempt      uint8
	mu           *sync.Mutex
	c            *http.Client
}

func (h httpClient) Get(url string) (resp *http.Response, err error) {
	for h.attempt < h.MaxRetries {
		if resp, err = h.c.Get(url); err == nil {
			if resp.StatusCode == http.StatusTooManyRequests {
				h.mu.Lock()
				h.attempt++
				h.mu.Unlock()
				time.Sleep(h.InitialDelay)

			} else {
				h.mu.Lock()
				h.attempt = 0
				h.mu.Unlock()
				break
			}
		}
	}
	return resp, err
}

func newHTTPClient(initialDelay time.Duration, maxRetries uint8) httpClient {
	return httpClient{
		InitialDelay: initialDelay,
		MaxRetries:   maxRetries,
		mu:           new(sync.Mutex),
		c:            http.DefaultClient,
	}
}
