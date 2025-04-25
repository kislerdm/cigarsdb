package cigargeeks

import (
	"cigarsdb/storage"
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClient_Read(t *testing.T) {
	t.Skip("manual integration test")
	c := Client{HTTPClient: http.DefaultClient}
	wantURL := "https://www.cigargeeks.com/index.php?action=cigars;area=showcig;cigar_id=193039;t=5_Vegas_Nicaragua_Churchill_LE"
	got, err := c.Read(context.TODO(), wantURL)
	assert.NoError(t, err)
	wantColor := "Colorado"
	wantStrength := "Medium"
	wantNodes := "https://www.cigarsinternational.com/p/5-vegas-nicaragua/2036078/"
	want := storage.Record{
		Name:              "Nicaragua Churchill LE",
		URL:               wantURL,
		Brand:             "5 Vegas",
		Ring:              49,
		LengthInch:        7,
		Format:            "Churchill",
		Maker:             nil,
		ManufactureOrigin: "Nicaragua",
		WrapperOrigin:     []string{"Ecuador", "Habano"},
		FillerOrigin:      []string{"Nicaragua"},
		BinderOrigin:      []string{"Nicaragua"},
		Color:             &wantColor,
		Strength:          &wantStrength,
		AdditionalNotes:   &wantNodes,
	}
	assert.Equal(t, want, got)
}

func TestClient_ReadBulk(t *testing.T) {
	t.Skip("manual integration test")
	c := Client{HTTPClient: http.DefaultClient}
	r, _, err := c.ReadBulk(context.TODO(), 0, 1)
	assert.NoError(t, err)
	assert.NotEmpty(t, r)
}
