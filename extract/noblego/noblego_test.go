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
var wantDieselCrucibleToro = storage.Record{
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
	Details: map[string]string{
		"Genussverlauf": "Die im Boxpressed Stil gehaltene Diesel Crucible Limited Edition 2021 Toro " +
			"macht äußerlich einen sehr geschmeidigen Eindruck. " +
			"Sehr einladend wirkt dabei ihr leicht ölig schimmerndes Deckblatt von dunkelbrauner Farbe, " +
			"das genau wie ihr Umblatt aus Ecuador bezogen wurde. Die beiden Bauchbinden in silbrigem Grau " +
			"passen sehr gut zum Farbton des Deckblatts und runden das sehr gute Erscheinungsbild der " +
			"Crucible Toro optisch hervorragend ab. In der Einlage fanden ausschließlich nicaraguanische " +
			"Tabake Platz, die viel Würze und feine pfeffrige Akzente versprechen. Der komplex geartete " +
			"Aromaverlauf bringt wunderbar abwechslungsfreudige Espresso-, Zartbitterschokolade- und Nougatnoten hervor, " +
			"die mal von delikaten Tönen gerösteter Nüsse oder pfeffrigen Nuancen begleitet werden. " +
			"Ein intensiv vollmundiger Zigarrengenuss für gut 90 Minuten.",
		"Nice to know": "Importiert werden die Zigarren der Marke Diesel von der AKRA Kotschenreuther GmbH mit Sitz in Langenzenn/Bayern. Zurzeit sind mit den <a href=\"/diesel-cask-aged-zigarren/\" title=\"Die Diesel Cask Aged Zigarren kennenlernen! \" target=\"_blank\">Cask Aged</a> und <a href=\"/diesel-barrel-aged-zigarren/\" title=\"Die Diesel Barrel Aged Zigarren kennenlernen! \" target=\"_blank\">Barrel Aged</a> Zigarren zwei regulär erscheinende Serien der Marke Diesel in unserem Online-Shop erhältlich.",
		"Resümee": "Eine gelungene Zigarre mit viel Tiefgang! Erfahrenen Gaumen bereitet die Diesel " +
			"Crucible Limited Edition 2021 ein köstliches Genusserlebnis bis zum letzten Aschefall. " +
			"Jetzt bestellen, solange der Vorrat reicht!",
	},
}

//go:embed testdata/details-diesel-cask-aged-robusto.html
var detailsDieselCaskAgedRobusto []byte
var wantDieselCaskAgedRobusto = storage.Record{
	Name:                  "Diesel Cask Aged Robusto",
	URL:                   "https://www.noblego.de/diesel-cask-aged-robusto-zigarren/",
	Brand:                 "Diesel",
	Series:                "Cask Aged",
	Diameter:              20.6,
	Ring:                  52,
	Length:                127,
	Format:                "Robusto",
	Maker:                 pointer("AJ Fernandez"),
	ManufactureOrigin:     "Nicaragua",
	Construction:          pointer("Longfiller"),
	IsBoxpressed:          pointer(false),
	WrapperOrigin:         []string{"USA"},
	FillerOrigin:          []string{"Nicaragua"},
	BinderOrigin:          []string{"Brasilien"},
	WrapperTobaccoVariety: pointer("Broadleaf"),
	AromaProfileManufacturer: []string{
		"Cremig", "Erdig", "Fruchtig", "Pfeffer", "Schokolade", "Trockenes Holz", "Zedernholz",
	},
	AdditionalNotes: pointer("fassgereifter Tabak"),
	Strength:        pointer("Medium"),
	FlavourStrength: pointer("Medium-aromatisch"),
	SmokingDuration: pointer("45 bis 60 Min"),
	Price:           8.9,
	Details: map[string]string{
		"Genussverlauf": "Das Connecticut Broadleaf Deckblatt aus US-amerikanischem Anbau ist von " +
			"sattbrauner Farbe. Das brasilianische Arapiraca Umblatt der Cask Aged Robusto wurde für " +
			"circa ein Jahr in ausgedienten Sherry-Fässern gelagert. Unter dem vorbehandelten Umblatt liegt " +
			"eine handverlesene Auswahl nicaraguanischer Tabake. Die mittelkräftige Robusto startet mit " +
			"holzigen Aromen, welche von einer feinen Rosinensüße flankiert werden. Hin und wieder " +
			"aufkeimende Pfeffernoten sind Indiz für das Wechselspiel der Aromen. Im weiteren Verlauf " +
			"bilden sich schokoladige Noten und cremige Erd- sowie Trockenholzaromen. Final erscheinen " +
			"alle im Vorfeld geschilderten Aromen in unterschiedlicher Intensität. Nach gut 60 Minuten ist " +
			"dieser vollmundige Genuss vorbei und hinterlässt bei seinem Konsumenten einen sehr gut " +
			"unterhaltenen Gaumen.",
		"Nice to know": "Hergestellt werden die Diesel Cask Aged Zigarren bei A.J. Fernandez in Nicaragua. " +
			"Ihr aufregender Blend wurde von Abdel höchstpersönlich entwickelt.",
		"Resümee": "Eine sehr gut gemachte Zigarre mit komplexen Anleihen, die definitiv über das gewisse " +
			"Etwas verfügt. Probieren Sie die Diesel Cask Aged Robusto einfach. " +
			"Wir sind gespannt darauf, wie sie Ihnen zusagen wird.",
	},
}

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
			wantR:      wantDieselCaskAgedRobusto,
			wantErr:    assert.NoError,
		},
		"https://www.noblego.de/diesel-crucible-toro-zigarren/": {
			httpClient: mockHttp{Body: io.NopCloser(bytes.NewReader(detailsDieselCrucibleToro))},
			id:         "https://www.noblego.de/diesel-crucible-toro-zigarren/",
			wantR:      wantDieselCrucibleToro,
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
	c := Client{HTTPClient: mockHttp{
		Body: io.NopCloser(bytes.NewReader(listData)),
		BodyRoute: map[string]io.ReadCloser{
			"https://www.noblego.de/diesel-cask-aged-robusto-zigarren/": io.NopCloser(bytes.NewReader(
				detailsDieselCaskAgedRobusto)),
			"https://www.noblego.de/diesel-crucible-toro-zigarren/": io.NopCloser(bytes.NewReader(
				detailsDieselCrucibleToro)),
		},
	}}
	want := []storage.Record{
		wantDieselCaskAgedRobusto,
		wantDieselCrucibleToro,
	}
	got, nextPage, err := c.ReadBulk(context.TODO(), 0, 1)
	assert.NoError(t, err)
	assert.Zero(t, nextPage)
	assert.Len(t, got, len(want))
	for i, el := range want {
		recordsEqual(t, el, got[i])
	}
}

//go:embed testdata/details-rocky-patel-vintage-connecticut-1999.html
var detailsRockyPatelVintageConnecticut1999 []byte

func Test_ReadVideo(t *testing.T) {
	c := Client{HTTPClient: mockHttp{Body: io.NopCloser(bytes.NewReader(detailsRockyPatelVintageConnecticut1999))}}
	want := storage.Record{
		Name:      "Rocky Patel Vintage Connecticut 1999 Robusto",
		Brand:     "Rocky Patel",
		Series:    "Vintage Connecticut 1999",
		VideoURLs: []string{"https://www.youtube.com/watch?v=jYW29PKpjyY"},
		Details: map[string]string{
			"Genuss": "Die Rocky Patel Vintage Connecticut 1999 Robusto ist mit ihren 12,7 Zentimetern Länge und dem " +
				"50er Ringmaß eine klassische Robusto. Und doch ist sie so vieles mehr als nur das. Sie ist sehr mild " +
				"und cremig im Rauch, bei den Aromen allerdings schöpft sie aus dem Vollen. Kaffee, fruchtige " +
				"Töne und eine subtile Süße sind nur die Spitze des Eisberges dieses fein komponierten Zigarre. " +
				"Das mindestens 7 Jahre alte Deckblatt aus Connecticut sorgt für einen unnachahmlichen Genuss. " +
				//nolint:misspell //Perfektion is a correct German word
				"Der Abbrand ist kreisrund, die Asche stabil und das Zugverhalten nahe der Perfektion. " +
				"Eine feine Zigarre ist diese Robusto von Superstar der internationalen Zigarrenproduktion: " +
				"Rocky Patel.",
			"Fazit": "Für fast jede Gelegenheit eignet sich diese Zigarre aber am angenehmsten ist doch der Sommerabend " +
				"im Freien mit einem Glas Riesling aus dem Rheingau.\n" +
				"Rocky Patel Vintage Connecticut 1999 Robusto - ein wunderbarer Smoke!",
		},
		Diameter:          20,
		Ring:              50,
		Length:            127,
		Format:            "Robusto",
		Maker:             pointer("Rocky Patel"),
		ManufactureOrigin: "Honduras",
		Construction:      pointer("Longfiller"),
		WrapperOrigin: []string{
			"Connecticut River Valley",
		},
		FillerOrigin: []string{
			"Dominikanische Republik", "Nicaragua",
		},
		BinderOrigin:             []string{"Nicaragua"},
		WrapperTobaccoVariety:    pointer("Connecticut Shade"),
		AromaProfileManufacturer: []string{"Cremig", "Fruchtig", "Kaffee", "Süß"},
		Strength:                 pointer("Mild"),
		FlavourStrength:          pointer("Medium-kräftig"),
		SmokingDuration:          pointer("45 bis 90 Min"),
		Price:                    8.6,
	}
	got, err := c.Read(context.TODO(), "")
	assert.NoError(t, err)
	recordsEqual(t, want, got)
}

func Test_readYoutubeVideoID(t *testing.T) {
	tests := map[string]string{
		"//www.youtube.com/v/jYW29PKpjyY?version=3&amp;hl=de_DE":       "https://www.youtube.com/watch?v=jYW29PKpjyY",
		"https://www.youtube.com/v/jYW29PKpjyY?version=3&amp;hl=de_DE": "https://www.youtube.com/watch?v=jYW29PKpjyY",
		"foo.bar": "foo.bar",
		"https://www.youtube.com/embed/VAiR_IlzA9c?rel=0": "https://www.youtube.com/watch?v=VAiR_IlzA9c",
		"https://www.youtube.com/v/watch?v=jYW29PKpjyY":   "https://www.youtube.com/watch?v=jYW29PKpjyY",
	}
	t.Parallel()
	for in, want := range tests {
		t.Run(in, func(t *testing.T) {
			assert.Equal(t, want, newVideoURL(in))
		})
	}
}

//go:embed testdata/details-la-aroma-del-caribe-no-2.html
var detailsLaAroma []byte

//go:embed testdata/details-carlos-torano-casa-torano-toro.html
var detailsCarlosToranoCasaTorano []byte

func TestClient_ReadExceptionalStructure(t *testing.T) {
	tests := map[string]struct {
		in   []byte
		want storage.Record
	}{
		"La Aroma del Caribe No. 2": {
			in: detailsLaAroma,
			want: storage.Record{
				Name:                     "La Aroma del Caribe No. 2",
				Brand:                    "La Aroma del Caribe",
				Series:                   "Edición Especial",
				VideoURLs:                []string{"https://www.youtube.com/watch?v=5ax4TwIiQCo"},
				Diameter:                 19.8,
				Ring:                     50,
				Length:                   127,
				Format:                   "Robusto",
				ManufactureOrigin:        "Nicaragua",
				WrapperOrigin:            []string{"Ecuador"},
				FillerOrigin:             []string{"Nicaragua"},
				BinderOrigin:             []string{"Nicaragua"},
				AromaProfileManufacturer: []string{"Erdig", "Fruchtig", "Leder", "Zedernholz", "Zimt"},
				Strength:                 pointer("Stark"),
				FlavourStrength:          pointer("Medium-aromatisch"),
				SmokingDuration:          pointer("45 bis 90 Min"),
				Price:                    8.5,
			},
		},
		"Carlos Toraño Casa Toraño Toro": {
			in: detailsCarlosToranoCasaTorano,
			want: storage.Record{
				Name:   "Carlos Toraño Casa Toraño Toro",
				Brand:  "Carlos Toraño",
				Series: "Casa Toraño",
				Details: map[string]string{
					"Genuss": "Die Serie Casa Toraño beinhaltet einen Blend mit Tabaken aus Honduras, Nicaragua und einer zusätzlichen Spezialmischung anderer zentralamerikanischer Tabake. Das Umblatt stammt aus Nicaragua, das Deckblatt ist ein sehr schönes Ecuador Connecticut. Im Geschmack liefert die Carlos Toraño Casa Toraño Toro ein wunderschönes Beispiel für das Zusammenspiel von Röstaromen, leichten Räucheraromen, nussigen Nuancen und auch Zitrusfrüchten. Der Duft der Casa Torano Toro überzeugt ebenfalls den verwöhnten Aficionado. Am Gaumen offenbart die Toro vielschichtige Noten von Holz, Nuss und Gewürzen. Diese Tabakware berührt somit alle Sinneseindrücke auf das Feinste.",
					"Fazit":  "Carlos Toraño Casa Toraño Toro: Ein milder und dabei mittelkräftiger Smoke zu einem sehr fairen Preis.",
				},
				Diameter:                 19,
				Ring:                     48,
				Length:                   158,
				Format:                   "Toro",
				ManufactureOrigin:        "Honduras",
				TypeOfManufacturing:      nil,
				Construction:             pointer("Longfiller"),
				WrapperOrigin:            []string{"Ecuador"},
				IsBoxpressed:             pointer(false),
				FillerOrigin:             []string{"Honduras", "Nicaragua", "Peru"},
				BinderOrigin:             []string{"Nicaragua"},
				WrapperTobaccoVariety:    pointer("Connecticut Shade"),
				AromaProfileManufacturer: []string{"Cremig", "Erdig", "Gewürze", "Nuss", "Pfeffer", "Röstaromen"},
				Strength:                 pointer("Medium"),
				FlavourStrength:          pointer("Mild-aromatisch"),
				SmokingDuration:          pointer("45 bis 90 Min"),
				Price:                    7.5,
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			c := Client{HTTPClient: mockHttp{Body: io.NopCloser(bytes.NewReader(tt.in))}}
			got, err := c.Read(context.TODO(), "")
			assert.NoError(t, err)
			recordsEqual(t, tt.want, got)
		})
	}
}
