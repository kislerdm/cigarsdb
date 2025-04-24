package fs

import (
	"cigarsdb/storage"
	"cmp"
	"context"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test(t *testing.T) {
	dir := t.TempDir()
	c, err := NewClient(dir)
	assert.NoError(t, err)
	ctx := context.TODO()

	t.Run("read empty storage", func(t *testing.T) {
		got, err := c.Read(ctx, "foo")
		assert.Error(t, err)
		assert.Empty(t, got)

		gotBulk, nextPage, err := c.ReadBulk(ctx, 100, 0)
		assert.NoError(t, err)
		assert.Equal(t, uint(0), nextPage)
		assert.Nil(t, gotBulk)
	})

	wantBulk := []storage.Record{{Name: "0"}, {Name: "1"}}
	var wantIDs = make([]string, 0, len(wantBulk))
	t.Run("store two records as singles", func(t *testing.T) {
		for _, want := range wantBulk {
			id, err := c.Write(ctx, []storage.Record{want})
			assert.NoError(t, err)
			assert.Len(t, id, 1)
			wantIDs = append(wantIDs, id[0])
		}
	})

	t.Run("read written records one by one", func(t *testing.T) {
		for i, id := range wantIDs {
			got, err := c.Read(ctx, id)
			assert.NoError(t, err)
			assert.Equal(t, wantBulk[i], got)
		}
	})

	t.Run("store two records in bulk", func(t *testing.T) {
		gotIDs, err := c.Write(ctx, wantBulk)
		assert.NoError(t, err)
		assert.Equal(t, wantIDs, gotIDs)
	})

	t.Run("read written records in bulk in one go", func(t *testing.T) {
		gotBulk, nextPage, err := c.ReadBulk(ctx, uint(len(wantBulk)), 0)
		assert.NoError(t, err)
		assert.Equal(t, uint(0), nextPage)
		assert.Len(t, gotBulk, len(wantBulk))

		tmpWant := wantBulk
		slices.SortStableFunc(tmpWant, func(a, b storage.Record) int {
			return cmp.Compare(a.Name, b.Name)
		})

		tmpGot := gotBulk
		slices.SortStableFunc(tmpGot, func(a, b storage.Record) int {
			return cmp.Compare(a.Name, b.Name)
		})

		assert.Equal(t, tmpWant, tmpGot)
	})

	t.Run("read written records in pages record", func(t *testing.T) {
		var cnt int
		var page uint
		for {
			got, nextPage, err := c.ReadBulk(ctx, 1, page)
			assert.NoError(t, err)
			cnt++

			if nextPage > page {
				assert.Len(t, got, 1)
				assert.Contains(t, wantBulk, got[0])
				page = nextPage
			} else {
				break
			}
		}
		assert.Equal(t, len(wantBulk), cnt)
	})
}
