package main

import (
	_ "embed"
	"testing"

	"github.com/stretchr/testify/assert"
)

// the html contain the list with three items, two of which correspond to a single unit
//
//go:embed list.html
var list []byte

func Test_readList(t *testing.T) {
	want := []Record{
		{
			Name:           "Reposado Estate Blend Colorado Robusto",
			URL:            "https://www.noblego.de/reposado-estate-blend-colorado-robusto-zigarren/",
			OriginCountry:  "Nicaragua",
			Format:         "Robusto",
			Form:           "Rund",
			WrapperCountry: []string{"Ecuador"},
			FillerCountry:  []string{"Nicaragua"},
			Strength:       "Medium",
			Price:          2.1,
			Diameter:       19.8,
			Length:         127,
			Aromas:         []string{"Süß", "Würzig", "Zedernholz"},
		},
		{
			Name:           "Reposado Robusto",
			URL:            "https://www.noblego.de/reposado-robusto/",
			OriginCountry:  "Nicaragua",
			Form:           "Rund",
			WrapperCountry: []string{"Honduras"},
			FillerCountry:  []string{"Honduras", "Nicaragua"},
			Strength:       "Medium",
			Price:          2.1,
			Diameter:       19.8,
			Length:         127,
			Aromas:         []string{"Cremig", "Gras", "Holz", "Leder", "Nuss", "Süß"},
		},
	}
	got := readList(list)
	assert.Len(t, got, 2)
	assert.Equal(t, want, got)
}
