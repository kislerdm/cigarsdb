package neo4j

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type inL1 struct {
	Foo string `json:"foo"`
	Bar *int64 `json:"bar,omitempty"`
}

type inL0 struct {
	Foo string    `json:"foo"`
	Bar *string   `json:"bar,omitempty"`
	Baz []inL1    `json:"baz,omitempty"`
	Qux []float64 `json:"qux,omitempty"`
}

func Test_fromRecord(t *testing.T) {
	var in = inL0{
		Foo: "foo",
		Bar: pointer("bar"),
		Baz: []inL1{
			{
				Foo: "fooL01",
				Bar: pointer(int64(1)),
			},
			{
				Foo: "fooL11",
			},
		},
		Qux: []float64{1.0, 2.0, 3.0},
	}

	got, err := fromRecord(in)
	assert.NoError(t, err)

	want := map[string]any{
		"foo": "foo",
		"bar": "bar",
		"baz": []map[string]any{
			{
				"foo": "fooL01",
				"bar": int64(1),
			},
			{"foo": "fooL11"},
		},
		"qux": []float64{1.0, 2.0, 3.0},
	}
	assert.Equal(t, want, got)
}

func pointer[V int64 | int | string](v V) *V {
	return &v
}
