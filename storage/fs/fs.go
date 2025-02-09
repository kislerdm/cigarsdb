// Package fs defines the FileSystem client which implements the storage.ReadWriter interface
package fs

import (
	"cigarsdb/storage"
	"context"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"strings"
)

func NewClient(dir string) (c *Client, err error) {
	if err = os.MkdirAll(dir, 0750); err == nil {
		c = &Client{Path: dir}
	}
	return c, err
}

type Client struct {
	Path string
}

func (c Client) Write(_ context.Context, r storage.Record) (string, error) {
	var err error
	id := c.newID(r)
	var f io.WriteCloser
	if f, err = os.OpenFile(c.filePath(id), os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0660); err == nil {
		defer func() { _ = f.Close() }()
		if err = json.NewEncoder(f).Encode(r); err != nil {
			id = ""
		}
	}
	return id, err
}

func (c Client) WriteBulk(_ context.Context, r []storage.Record) ([]string, error) {
	var (
		ids = make([]string, len(r))
		id  string
		err error
	)
	for i, el := range r {
		id = c.newID(el)
		var f io.WriteCloser
		if f, err = os.OpenFile(c.filePath(id), os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0660); err == nil {
			if err = json.NewEncoder(f).Encode(el); err == nil {
				ids[i] = id
			} else {
				break
			}
			_ = f.Close()
		}
	}
	return ids, err
}

func (c Client) Read(_ context.Context, id string) (storage.Record, error) {
	var (
		r   storage.Record
		err error
		f   io.ReadCloser
	)
	if f, err = os.Open(c.filePath(id)); err == nil {
		err = json.NewDecoder(f).Decode(&r)
		_ = f.Close()
	}
	return r, err
}

func (c Client) ReadBulk(_ context.Context, limit, page uint) ([]storage.Record, uint, error) {
	const defaultLimit = 100

	var (
		rs       []storage.Record
		err      error
		nextPage uint
	)

	if limit == 0 {
		limit = defaultLimit
	}

	skipStart := int(limit * page)
	skipEnd := skipStart + int(limit)
	var (
		skipCnt    int
		writtenCnt int
	)
	err = fs.WalkDir(os.DirFS(c.Path), ".", func(p string, d fs.DirEntry, err error) error {
		var found bool
		if !d.Type().IsDir() && strings.HasSuffix(p, ".json") {
			found = true
			skipCnt++
		}

		if found {
			var r storage.Record
			switch skipCnt > skipStart && writtenCnt < skipEnd {
			case true:
				if r, err = c.Read(nil, strings.TrimSuffix(p, ".json")); err == nil {
					rs = append(rs, r)
					writtenCnt++
				}

			case false:
				nextPage++
				err = fs.SkipAll
			}
		}
		return err
	})
	if err != nil {
		rs = nil
		nextPage = 0
	}
	return rs, nextPage, err
}

func (c Client) filePath(id string) string {
	return path.Join(c.Path, id) + ".json"
}

func (c Client) newID(r storage.Record) string {
	h := sha1.New()
	_, _ = io.WriteString(h, r.Name)
	return strings.TrimSpace(fmt.Sprintf("%x", h.Sum(nil)))
}
