package noblego

import (
	"bytes"
	"cigarsdb/storage"
	"context"
	_ "embed"
	"io"
	"net/http"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

//go:embed testdata/details-diesel-crucible-toro.html
var detailsDieselCrucibleToro []byte

//go:embed testdata/details-diesel-cask-aged-robusto.html
var detailsDieselCaskAgedRobusto []byte

func TestClient_Read(t *testing.T) {
	tests := map[string]struct {
		httpClient HTTPClient
		id         string
		wantR      storage.Record
		wantErr    assert.ErrorAssertionFunc
	}{
		"https://www.noblego.de/diesel-cask-aged-robusto-zigarren/": {
			httpClient: mockHttp{Body: io.NopCloser(bytes.NewReader(detailsDieselCaskAgedRobusto))},
			id:         "https://www.noblego.de/diesel-cask-aged-robusto-zigarren/",
			wantR: storage.Record{
				Name:              "Diesel Cask Aged Robusto",
				URL:               "https://www.noblego.de/diesel-cask-aged-robusto-zigarren/",
				Brand:             "Diesel",
				Series:            "Cask Aged",
				Diameter:          20.6,
				Ring:              52,
				Length:            127,
				Format:            "Robusto",
				Maker:             pointer("AJ Fernandez"),
				ManufactureOrigin: "Nicaragua",
				Construction:      pointer("Longfiller"),
				IsBoxpressed:      pointer(false),
				WrapperOrigin:     []string{"USA"},
				FillerOrigin:      []string{"Nicaragua"},
				BinderOrigin:      []string{"Brasilien"},
				WrapperType:       pointer("Broadleaf"),
				AromaProfileManufacturer: []string{
					"Cremig", "Erdig", "Fruchtig", "Pfeffer", "Schokolade", "Trockenes Holz", "Zedernholz",
				},
				AdditionalNotes: pointer("fassgereifter Tabak"),
				Strength:        pointer("Medium"),
				FlavourStrength: pointer("Medium-aromatisch"),
				SmokingDuration: pointer("45 bis 60 Min"),
				Price:           8.9,
			},
			wantErr: assert.NoError,
		},
		"https://www.noblego.de/diesel-crucible-toro-zigarren/": {
			httpClient: mockHttp{Body: io.NopCloser(bytes.NewReader(detailsDieselCrucibleToro))},
			id:         "https://www.noblego.de/diesel-crucible-toro-zigarren/",
			wantR: storage.Record{
				Name:              "Diesel Crucible Limited Edition 2021 Toro",
				URL:               "https://www.noblego.de/diesel-crucible-toro-zigarren/",
				Brand:             "Diesel",
				Series:            "Crucible",
				Diameter:          19.8,
				Ring:              50,
				Length:            152,
				Format:            "Toro",
				Maker:             pointer("AJ Fernandez"),
				ManufactureOrigin: "Nicaragua",
				Construction:      pointer("Longfiller"),
				IsBoxpressed:      pointer(true),
				WrapperOrigin:     []string{"Ecuador"},
				FillerOrigin:      []string{"Nicaragua"},
				BinderOrigin:      []string{"Ecuador"},
				AromaProfileManufacturer: []string{
					"Espresso", "Nougat", "Nuss", "Röstaromen", "Schokolade", "Schwarzer Pfeffer",
				},
				AdditionalNotes: pointer("Limited"),
				Strength:        pointer("Medium"),
				FlavourStrength: pointer("Medium-aromatisch"),
				SmokingDuration: pointer("60 bis 90 Min"),
				Price:           11.5,
			},
			wantErr: assert.NoError,
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

func Test_dataInCurrentPage(t *testing.T) {
	type args struct {
		page  uint
		limit uint
		total uint
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "1st page, 2 items per page, 3 items total",
			args: args{
				page:  1,
				limit: 2,
				total: 3,
			},
			want: true,
		},
		{
			name: "2st page, 2 items per page, 3 items total",
			args: args{
				page:  2,
				limit: 2,
				total: 3,
			},
			want: true,
		},
		{
			name: "3rd page, 2 items per page, 3 items total",
			args: args{
				page:  3,
				limit: 2,
				total: 3,
			},
			want: false,
		},
	}
	t.Parallel()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, dataInCurrentPage(tt.args.page, tt.args.limit, tt.args.total))
		})
	}
}

type mockHttp struct {
	Body      io.ReadCloser
	BodyRoute map[string]io.ReadCloser
	Err       error
}

func (m mockHttp) Get(url string) (*http.Response, error) {
	var r *http.Response
	if m.Err == nil {
		r = &http.Response{
			StatusCode: http.StatusOK,
			Body:       m.Body,
		}
		if v, ok := m.BodyRoute[url]; ok {
			r.Body = v
		}
	}
	return r, m.Err
}

//go:embed testdata/list.html
var listData []byte

func TestClient_ReadBulk(t *testing.T) {
	listUrl := "https://www.noblego.de/zigarren/?limit=96&p=1"
	c := Client{HTTPClient: mockHttp{
		BodyRoute: map[string]io.ReadCloser{
			listUrl: io.NopCloser(bytes.NewReader(listData)),
			"https://www.noblego.de/diesel-cask-aged-robusto-zigarren/": io.NopCloser(bytes.NewReader(
				detailsDieselCaskAgedRobusto)),
			"https://www.noblego.de/diesel-crucible-toro-zigarren/": io.NopCloser(bytes.NewReader(
				detailsDieselCrucibleToro)),
		},
	}}
	want := []storage.Record{
		{
			Name:              "Diesel Cask Aged Robusto",
			URL:               "https://www.noblego.de/diesel-cask-aged-robusto-zigarren/",
			Brand:             "Diesel",
			Series:            "Cask Aged",
			Diameter:          20.6,
			Ring:              52,
			Length:            127,
			Format:            "Robusto",
			Maker:             pointer("AJ Fernandez"),
			ManufactureOrigin: "Nicaragua",
			Construction:      pointer("Longfiller"),
			IsBoxpressed:      pointer(false),
			WrapperOrigin:     []string{"USA"},
			FillerOrigin:      []string{"Nicaragua"},
			BinderOrigin:      []string{"Brasilien"},
			WrapperType:       pointer("Broadleaf"),
			AromaProfileManufacturer: []string{
				"Cremig", "Erdig", "Fruchtig", "Pfeffer", "Schokolade", "Trockenes Holz", "Zedernholz",
			},
			AdditionalNotes: pointer("fassgereifter Tabak"),
			Strength:        pointer("Medium"),
			FlavourStrength: pointer("Medium-aromatisch"),
			SmokingDuration: pointer("45 bis 60 Min"),
			Price:           8.9,
		},
		{
			Name:              "Diesel Crucible Limited Edition 2021 Toro",
			URL:               "https://www.noblego.de/diesel-crucible-toro-zigarren/",
			Brand:             "Diesel",
			Series:            "Crucible",
			Diameter:          19.8,
			Ring:              50,
			Length:            152,
			Format:            "Toro",
			Maker:             pointer("AJ Fernandez"),
			ManufactureOrigin: "Nicaragua",
			Construction:      pointer("Longfiller"),
			IsBoxpressed:      pointer(true),
			WrapperOrigin:     []string{"Ecuador"},
			FillerOrigin:      []string{"Nicaragua"},
			BinderOrigin:      []string{"Ecuador"},
			AromaProfileManufacturer: []string{
				"Espresso", "Nougat", "Nuss", "Röstaromen", "Schokolade", "Schwarzer Pfeffer",
			},
			AdditionalNotes: pointer("Limited"),
			Strength:        pointer("Medium"),
			FlavourStrength: pointer("Medium-aromatisch"),
			SmokingDuration: pointer("60 bis 90 Min"),
			Price:           11.5,
		},
	}
	got, nextPage, err := c.ReadBulk(context.TODO(), 0, 1)
	assert.NoError(t, err)
	assert.Zero(t, nextPage)
	assert.Len(t, got, len(want))
	for i, el := range want {
		recordsEqual(t, el, got[i])
	}
}
