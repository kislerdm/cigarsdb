package cigarcentury

import (
	"bytes"
	"cigarsdb/extract"
	"context"
	_ "embed"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

//go:embed fixtures/arturo-fuente-casa-cuba-divine-inspiration.html
var ArturoFuenteCasaCubaDivineInspiration []byte

func TestClient_ReadArturoFuenteCasaCubaDivineInspiration(t *testing.T) {

	c := Client{
		HTTPClient: &extract.MockHTTP{
			Body: io.NopCloser(bytes.NewReader(ArturoFuenteCasaCubaDivineInspiration)),
		},
	}

	wantURL := "https://www.cigarcentury.com/en/cigars/arturo-fuente-casa-cuba-divine-inspiration"
	got, err := c.Read(context.TODO(), wantURL)
	assert.NoError(t, err)

	t.Parallel()

	t.Run("name", func(t *testing.T) {
		assert.Equal(t, "Arturo Fuente - Casa Cuba - Divine Inspiration", got.Name)
	})

	t.Run("origin", func(t *testing.T) {
		assert.Equal(t, []string{"Dominican Republic"}, got.FillerOrigin)
	})

	t.Run("manufacturer", func(t *testing.T) {
		assert.Equal(t, "Tabacalera Fuente", *got.Maker)
	})

	t.Run("brand", func(t *testing.T) {
		assert.Equal(t, "Arturo Fuente, Casa Cuba", got.Brand)
	})

	t.Run("wrapper", func(t *testing.T) {
		assert.Equal(t, []string{"Ecuadorian"}, got.WrapperOrigin)
	})

	t.Run("binder", func(t *testing.T) {
		assert.Equal(t, []string{"Dominican Republic"}, got.BinderOrigin)
	})

	t.Run("filler", func(t *testing.T) {
		assert.Equal(t, []string{"Dominican Republic"}, got.FillerOrigin)
	})

	t.Run("format", func(t *testing.T) {
		assert.Equal(t, "Corona Grande", got.Format)
	})

	t.Run("box-pressed", func(t *testing.T) {
		assert.False(t, *got.IsBoxpressed)
	})

	t.Run("strength", func(t *testing.T) {
		assert.Equal(t, "Medium", *got.Strength)
	})

	t.Run("flavours", func(t *testing.T) {
		flavours := *got.AromaProfileCommunity
		assert.Empty(t, flavours.NumberOfVotes)

		const wantWeight = float64(1. / 5)
		wantWeights := map[string]float64{
			"Oak":          wantWeight,
			"Raisin":       wantWeight,
			"Spices":       wantWeight,
			"Coffee Beans": wantWeight,
			"White Pepper": wantWeight,
		}
		assert.Equal(t, wantWeights, flavours.Weights)
	})

	t.Run("length", func(t *testing.T) {
		assert.Equal(t, 156., got.Length)
		assert.Empty(t, got.LengthInch)
	})

	t.Run("ring gauge", func(t *testing.T) {
		assert.Equal(t, float64(47), got.Ring)
		assert.Empty(t, got.Diameter)
	})

	t.Run("colour", func(t *testing.T) {
		assert.Equal(t, "Colorado Maduro", *got.Color)
	})

	t.Run("specialized rating", func(t *testing.T) {
		assert.Len(t, got.SpecializedRatings, 8)
		for i, v := range got.SpecializedRatings {
			t.Run(fmt.Sprintf("element %d 'who' is not empty", i), func(t *testing.T) {
				assert.NotEmpty(t, v.Who)
			})
			t.Run(fmt.Sprintf("element %d 'rating' is over zero", i), func(t *testing.T) {
				assert.Greater(t, v.RatingOutOf100, 0.)
			})
			t.Run(fmt.Sprintf("element %d 'year' is not empty", i), func(t *testing.T) {
				assert.NotEmpty(t, v.Year)
			})
		}
	})

	t.Run("url is set", func(t *testing.T) {
		assert.Equal(t, wantURL, got.URL)
	})
}

func Test_readVotes(t *testing.T) {
	tests := map[string]int{
		"(1)":     1,
		" (1) ":   1,
		"(1) ":    1,
		"(1)\n":   1,
		"(1)\n\t": 1,
		"(1)\t":   1,
	}
	for in, want := range tests {
		t.Run(in, func(t *testing.T) {
			assert.Equalf(t, want, readVotes(in), "readVotes(%v)", in)
		})
	}
}

//go:embed fixtures/cigars-p-473.html
var AllCigars []byte

func TestClient_ReadBulk(t *testing.T) {
	c := Client{
		HTTPClient: &extract.MockHTTP{
			Body: io.NopCloser(bytes.NewReader(AllCigars)),
		},
	}

	got, nextPage, err := c.ReadBulk(context.TODO(), 0, 0)
	assert.NoError(t, err)
	assert.Equal(t, nextPage, 0)
	assert.Len(t, got, 5675)
}
