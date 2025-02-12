package cigarworld

import (
	"bytes"
	"cigarsdb/extract"
	"cigarsdb/storage"
	"context"
	_ "embed"
	"io"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

//go:embed testdata/details-diesel-cask-aged-robusto-90016191_48509.html
var detailsDieselCaskAgedRobusto []byte
var wantDieselCaskAgedRobusto = storage.Record{
	Name:                    "Diesel Cask Aged Robusto",
	URL:                     "https://www.cigarworld.de/en/zigarren/nicaragua/diesel-cask-aged-robusto-90016191_48509",
	Brand:                   "Diesel",
	Diameter:                20.6,
	Ring:                    52,
	LengthInch:              5,
	Length:                  127,
	Format:                  "Robusto",
	Maker:                   pointer("A.J. Fernandez"),
	IsBoxpressed:            pointer(false),
	IsFlavoured:             pointer(false),
	WrapperOrigin:           []string{"USA"},
	FillerOrigin:            []string{"Nicaragua"},
	BinderOrigin:            []string{"Mexiko"},
	WrapperType:             pointer("Broadleaf"),
	OuterLeafTobaccoVariety: pointer("San Andrés"),
	TypeOfManufacturing:     pointer("TAM"),
	//Price:           8.9,
}

func TestClient_Read(t *testing.T) {
	tests := map[string]struct {
		httpClient extract.HTTPClient
		id         string
		wantR      storage.Record
		wantErr    assert.ErrorAssertionFunc
	}{
		"https://www.cigarworld.de/en/zigarren/nicaragua/diesel-cask-aged-robusto-90016191_48509": {
			httpClient: extract.MockHTTP{Body: io.NopCloser(bytes.NewReader(detailsDieselCaskAgedRobusto))},
			id:         "https://www.cigarworld.de/en/zigarren/nicaragua/diesel-cask-aged-robusto-90016191_48509",
			wantR:      wantDieselCaskAgedRobusto,
			wantErr:    assert.NoError,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			c := Client{
				HTTPClient: tt.httpClient,
			}
			gotR, err := c.Read(context.TODO(), tt.id)
			tt.wantErr(t, err)
			recordsEqual(t, tt.wantR, gotR)
		})
	}
}

func recordsEqual(t *testing.T, want storage.Record, got storage.Record) {
	wantT := reflect.TypeOf(want)
	wantV := reflect.ValueOf(want)
	gotV := reflect.ValueOf(got)
	for i := 0; i < wantT.NumField(); i++ {
		fieldWantT := wantT.Field(i)
		fieldWantV := wantV.Field(i)
		fieldGotV := gotV.Field(i)

		if fieldWantT.Type.Kind() == reflect.Pointer && fieldWantV.IsNil() {
			assert.Truef(t, fieldGotV.IsNil(), "field %s", fieldWantT.Name)
		} else {
			assert.Equalf(t, fieldWantV.Interface(), fieldGotV.Interface(),
				"field %s", fieldWantT.Name)
		}
	}
}
