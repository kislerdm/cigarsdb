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
	Name:                 "Diesel Cask Aged Robusto",
	URL:                  "https://www.cigarworld.de/en/zigarren/nicaragua/diesel-cask-aged-robusto-90016191_48509",
	Brand:                "Diesel",
	Diameter:             20.6,
	Ring:                 52,
	LengthInch:           5,
	Length:               127,
	Format:               "Robustos",
	Maker:                pointer("A.J. Fernandez"),
	IsBoxpressed:         pointer(false),
	IsFlavoured:          pointer(false),
	WrapperOrigin:        []string{"USA"},
	FillerOrigin:         []string{"Nicaragua"},
	BinderOrigin:         []string{"Mexiko"},
	WrapperProperty:      pointer("Broadleaf"),
	BinderTobaccoVariety: pointer("San Andrés"),
	TypeOfManufacturing:  pointer("TAM"),
	Price:                8.9,
}

//go:embed testdata/details-my-father-cigars-limited-edition-tatuaje-la-union-2023.html
var detailsMyFatherCigarsLimitedEditionTatuajeLaUnion []byte
var wantMyFatherCigarsLimitedEditionTatuajeLaUnion = storage.Record{
	Name:                  "My Father Cigars Limited Edition Tatuaje La Union 2023",
	URL:                   "https://www.cigarworld.de/zigarren/nicaragua/my-father-cigars-limited-edition-tatuaje-la-union-2023-90017088_56298",
	Brand:                 "My Father Cigars",
	Diameter:              19.8,
	Ring:                  50,
	LengthInch:            7.25,
	Length:                184.2,
	Maker:                 pointer("My Father Cigars S.A."),
	IsBoxpressed:          pointer(false),
	IsFlavoured:           pointer(false),
	WrapperOrigin:         []string{"Ecuador", "Nicaragua"},
	FillerOrigin:          []string{"Nicaragua"},
	BinderOrigin:          []string{"Nicaragua"},
	WrapperTobaccoVariety: pointer("Corojo, H-2000"),
	WrapperProperty:       pointer("Shade"),
	TypeOfManufacturing:   pointer("TAM"),
	Price:                 87.3,
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
		"https://www.cigarworld.de/zigarren/nicaragua/my-father-cigars-limited-edition-tatuaje-la-union-2023-90017088_56298": {
			httpClient: extract.MockHTTP{Body: io.NopCloser(bytes.NewReader(detailsMyFatherCigarsLimitedEditionTatuajeLaUnion))},
			id:         "https://www.cigarworld.de/zigarren/nicaragua/my-father-cigars-limited-edition-tatuaje-la-union-2023-90017088_56298",
			wantR:      wantMyFatherCigarsLimitedEditionTatuajeLaUnion,
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
