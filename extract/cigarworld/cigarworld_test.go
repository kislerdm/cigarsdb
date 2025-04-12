package cigarworld

import (
	"bytes"
	"cigarsdb/extract"
	"cigarsdb/htmlfilter"
	"cigarsdb/storage"
	"context"
	_ "embed"
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/net/html"
)

//go:embed testdata/details-diesel-cask-aged-robusto-90016191_48509.html
var detailsDieselCaskAgedRobusto []byte
var wantDieselCaskAgedRobusto = storage.Record{
	Name:                 "Diesel Cask Aged Robusto",
	URL:                  "https://www.cigarworld.de/en/zigarren/nicaragua/diesel-cask-aged-robusto-90016191_48509",
	Brand:                "Diesel",
	Series:               "Robusto",
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
	WrapperProperty:      []string{"Broadleaf"},
	BinderTobaccoVariety: []string{"San Andrés"},
	TypeOfManufacturing:  pointer("TAM"),
	Price:                8.9,
	AromaProfileCommunity: &storage.AromaProfileCommunity{
		Weights: map[string]float64{
			//642111211257 -> sum = 33
			"Holz": 6. / 33, "Pfeffer": 4. / 33, "Gras": 2. / 33, "Frucht": 1. / 33, "Creme": 1. / 33, "Süß": 1. / 33,
			"Nuss": 2. / 33, "Schokolade": 1. / 33, "Kaffee": 1. / 33, "Toast": 2. / 33,
			"Leder": 5. / 33, "Erde": 7. / 33,
		},
		NumberOfVotes: 1,
	},
}

//go:embed testdata/details-my-father-cigars-limited-edition-tatuaje-la-union-2023.html
var detailsMyFatherCigarsLimitedEditionTatuajeLaUnion []byte
var wantMyFatherCigarsLimitedEditionTatuajeLaUnion = storage.Record{
	Name:                  "My Father Cigars Limited Edition Tatuaje La Union 2023",
	URL:                   "https://www.cigarworld.de/zigarren/nicaragua/my-father-cigars-limited-edition-tatuaje-la-union-2023-90017088_56298",
	Brand:                 "My Father Cigars",
	Series:                "Tatuaje La Union 2023",
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
	WrapperTobaccoVariety: []string{"Corojo, H-2000"},
	WrapperProperty:       []string{"Shade"},
	TypeOfManufacturing:   pointer("TAM"),
	Price:                 87.3,
	Details: map[string]string{
		//nolint:misspell // Kollaboration is a correct German word
		"description": "<p>Die exklusive<strong> Kollaboration</strong> zwischen " +
			"<a href=\"/my-father-cigars\" target=\"_self\"><strong>My Father Cigars</strong>" +
			"</a> und <a href=\"/tatuaje\" target=\"_self\"><strong>Tatuaje</strong>" +
			"</a> hat den limitierten <strong>Sampler La Union</strong> für 2023 herausgebracht." +
			"</p><div class=\"blanktag\" style=\"background: url('https://www.cigarworld.de/binary/shop/blank');\">" +
			"</div>\n                                            " +
			"<p>Dieses einzigartige Set enthält je 20 Zigarren der Sorten <strong>Prominente Especial Tatuaje" +
			"</strong> mit einem Deckblatt aus nicaraguanischem Shade Grown Corojo 99 und <strong>" +
			"Prominente Especial My Father</strong> mit einem Deckblatt aus Ecuador H-2000.</p>" +
			"<div class=\"blanktag\" style=\"background: url('https://www.cigarworld.de/binary/shop/blank');\"></div>" +
			"\n                                            " +
			"<p><span>Wir freuen uns sehr, einige der <strong>weltweit nur 1.500" +
			"</strong> produzierten schwarz-glänzenden Kisten ergattert zu haben, von denen lediglich " +
			"30 ihren Weg nach Deutschland gefunden haben.</span></p><div class=\"blanktag\" style=\"background: " +
			"url('https://www.cigarworld.de/binary/shop/blank');\"></div>" +
			"\n                                            " +
			"<p>Die Warnhinweise befinden sich NUR auf dem <strong>äußeren Karton</strong>, nicht wie bei unseren " +
			"Bildern dargestellt, in oder auf der Kiste. D.h. die Kiste ist in ihrem <strong>schönen</strong>, " +
			"<strong>unbeklebten</strong> Zustand. Wir haben diese Warnhinweise digital hinzugefügt, um " +
			"gegebenenfalls Probleme mit Behörden zu vermeiden.</p><div class=\"blanktag\" style=\"background: " +
			"url('https://www.cigarworld.de/binary/shop/blank');\"></div>",
	},
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

func Test_readAromaCat(t *testing.T) {
	in := `
		    		var NameArrObj = {
		    	AromaNamenArr: ["Holz","Pfeffer","Gras","Frucht","Creme","Süß","Nuss","Schokolade","Kaffee","Toast","Leder","Erde"],
				AromaTabacNamenArr: ["Vanille","Süße","Erdig","Rauchig","Seifig","Fruchtig","Aromatisierung","Würze/Umami","Säure","Nussig","Grasig","Röstaromen"],
				PropsNameArr: ["Preis/Leistung","Verarbeitung","Stärke","Rauchvolumen","Zugwiderstand","Abbrandverhalten","Aromavielfalt","Aromaintensität"],
		    };

			$(document).ready(function(){
			    paintAroma(NameArrObj);
			    createSchieber();
			});
					$('img[data-src]:not(.swiper-lazy)').unveil(100, function() {
				$(this).on('load', function() {
					$(this).addClass('unveiled');
				});
			});
`
	wantCat := []string{
		"Holz", "Pfeffer", "Gras", "Frucht", "Creme", "Süß", "Nuss", "Schokolade", "Kaffee", "Toast", "Leder", "Erde",
	}
	wantCatAlt := []string{
		"Vanille", "Süße", "Erdig", "Rauchig", "Seifig", "Fruchtig", "Aromatisierung",
		"Würze/Umami", "Säure", "Nussig", "Grasig", "Röstaromen",
	}
	gotCat, gotCatAlt := readAromaCats(in)
	assert.Equal(t, wantCat, gotCat)
	assert.Equal(t, wantCatAlt, gotCatAlt)
}

//go:embed testdata/aggregate-page.html
var aggregatePage []byte

func Test_extractURLs(t *testing.T) {
	gotUrls, err := extractURLs(aggregatePage)
	assert.NoError(t, err)
	assert.Len(t, gotUrls, 27)
}

func Test_readPrice(t *testing.T) {
	t.Run("space and unbreakable space", func(t *testing.T) {
		in := `<div class="ws-g ws-c DetailOrderbox" data-addtocart="loader">
			<div class="ws-u-1">
                				<h4 class="h-alt h-alt-nm d-none d-lg-block nobr">Chicos</h4>

                
				
				
				<div class="ws-g DetailOrderbox-row DetailOrderbox--titlerow mt-2 mt-lg-3">
					<div class="ws-u-4-24 DetailOrderbox-col DetailOrderbox-image DetailOrderbox--title">&nbsp;</div>
					<div class="ws-u-7-24 DetailOrderbox-col DetailOrderbox-price DetailOrderbox--title">Preis</div>
					<div class="ws-u-5-24 DetailOrderbox-col DetailOrderbox-quantity DetailOrderbox--title">Menge</div>
					<div class="ws-u-8-24 DetailOrderbox-col DetailOrderbox-unit DetailOrderbox--title">Einheit</div>
				</div>
				

<form class="ws-form" method="post" action="/warenkorb/show" data-ajaction="/ajax/warenkorb/show" data-addtocart="form">
        <div class="ws-g DetailOrderbox-row">
        <div class="ws-u-4-24 DetailOrderbox-col DetailOrderbox-image" style="position: relative;">
        	              
        </div>
        <div class="ws-u-7-24 DetailOrderbox-col DetailOrderbox-price">
        	<span class="preis"><span data-eurval="1.60" data-curiso="EUR">€1.60</span></span>                        <nobr><span class="grundbetrag"></span></nobr>        </div>
        <div class="ws-u-5-24 DetailOrderbox-col DetailOrderbox-quantity">
                                </div>
        <div class="ws-u-8-24 DetailOrderbox-col DetailOrderbox-unit">
            <label for="wk_anzahl_auswahl_1" title="Ausverkauft">
                <input type="hidden" name="wk_einheit[1]" value="1"><span class="einheitlabel avail_4" title="Nicht Lieferbar">1&nbsp; Stk.</span>
            </label>
                        <div><span class="avail_4">Ausverkauft</span></div>        </div>
    </div>

        <div class="ws-g DetailOrderbox-row">
        <div class="ws-u-4-24 DetailOrderbox-col DetailOrderbox-image" style="position: relative;">
        	              
        </div>
        <div class="ws-u-7-24 DetailOrderbox-col DetailOrderbox-price">
        	<span class="preis"><span data-eurval="38.80" data-curiso="EUR">€38.80</span><del><span data-eurval="40.00" data-curiso="EUR">€40.00</span></del></span>                        <nobr><span class="grundbetrag"></span></nobr>        </div>
        <div class="ws-u-5-24 DetailOrderbox-col DetailOrderbox-quantity">
                                </div>
        <div class="ws-u-8-24 DetailOrderbox-col DetailOrderbox-unit">
            <label for="wk_anzahl_auswahl_2" title="Ausverkauft">
                <input type="hidden" name="wk_einheit[2]" value="2"><span class="einheitlabel avail_4" title="Nicht Lieferbar">25&nbsp; Stk.</span>
            </label>
            <small style="color: #666">inkl. 3% Rabatt</small>            <div><span class="avail_4">Ausverkauft</span></div>        </div>
    </div>

    
        
    
    <div class="ws-g mt-3">
        <div class="ws-u-1 text-right" style="font-size: 0.8em; opacity: 0.8;">Preise inkl. MwSt., ggf. zzgl. <a href="/service/versandkosten"><u>Versand</u></a></div>
    </div>
    
    <div class="ws-g DetailOrderbox-buttons mt-3">

        <div class="ws-u-1 ws-u-md-1-2 DetailOrderbox-addtocart">
                    </div>
        
                    <div class="ws-u-0 ws-u-md-1-2 mt-2 mt-lg-0 DetailOrderbox-addtowatchlist">&nbsp;</div>
                
    </div>

</form>




			</div>
		</div>`
		n, err := html.Parse(strings.NewReader(in))
		assert.NoError(t, err)

		var got storage.Record
		assert.NoError(t, readPrice(htmlfilter.Node{Node: n}, &got))
		wantPrice := 1.6
		assert.Equal(t, wantPrice, got.Price)
	})

	t.Run("unbreakable space", func(t *testing.T) {
		in := `<div class="ws-g ws-c DetailOrderbox" data-addtocart="loader">
			<div class="ws-u-1">
                				<h4 class="h-alt h-alt-nm d-none d-lg-block nobr">SMALL Panatela (SOFT TOUCH-BOX)</h4>

                
				
				
				<div class="ws-g DetailOrderbox-row DetailOrderbox--titlerow mt-2 mt-lg-3">
					<div class="ws-u-4-24 DetailOrderbox-col DetailOrderbox-image DetailOrderbox--title">&nbsp;</div>
					<div class="ws-u-7-24 DetailOrderbox-col DetailOrderbox-price DetailOrderbox--title">Preis</div>
					<div class="ws-u-5-24 DetailOrderbox-col DetailOrderbox-quantity DetailOrderbox--title">Menge</div>
					<div class="ws-u-8-24 DetailOrderbox-col DetailOrderbox-unit DetailOrderbox--title">Einheit</div>
				</div>
				

<form class="ws-form" method="post" action="/warenkorb/show" data-ajaction="/ajax/warenkorb/show" data-addtocart="form">
        <div class="ws-g DetailOrderbox-row">
        <div class="ws-u-4-24 DetailOrderbox-col DetailOrderbox-image" style="position: relative;">
        	            <img src="/bilder/detail/small/1735_28510_53229.jpg" class="produktbild" alt="Balmoral Dominican Selection SMALL Panatela (SOFT TOUCH-BOX)">  
        </div>
        <div class="ws-u-7-24 DetailOrderbox-col DetailOrderbox-price">
        	<span class="preis"><span data-eurval="6.31" data-curiso="EUR">€6.31</span><del><span data-eurval="6.50" data-curiso="EUR">€6.50</span></del></span>                        <nobr><span class="grundbetrag"></span></nobr>        </div>
        <div class="ws-u-5-24 DetailOrderbox-col DetailOrderbox-quantity">
                                </div>
        <div class="ws-u-8-24 DetailOrderbox-col DetailOrderbox-unit">
            <label for="wk_anzahl_auswahl_1" title="Ausverkauft">
                <input type="hidden" name="wk_einheit[1]" value="18364"><span class="einheitlabel avail_4" title="Nicht Lieferbar">5&nbsp;Dose</span>
            </label>
            <small style="color: #666">inkl. 3% Rabatt</small>            <div><span class="avail_4">Ausverkauft</span></div>        </div>
    </div>

    
        
    
    <div class="ws-g mt-3">
        <div class="ws-u-1 text-right" style="font-size: 0.8em; opacity: 0.8;">Preise inkl. MwSt., ggf. zzgl. <a href="/service/versandkosten"><u>Versand</u></a></div>
    </div>
    
    <div class="ws-g DetailOrderbox-buttons mt-3">

        <div class="ws-u-1 ws-u-md-1-2 DetailOrderbox-addtocart">
                    </div>
        
                    <div class="ws-u-0 ws-u-md-1-2 mt-2 mt-lg-0 DetailOrderbox-addtowatchlist">&nbsp;</div>
                
    </div>

</form>




			</div>
		</div>`
		n, err := html.Parse(strings.NewReader(in))
		assert.NoError(t, err)
		var got storage.Record
		assert.NoError(t, readPrice(htmlfilter.Node{Node: n}, &got))
		wantPrice := 1.26
		assert.Equal(t, wantPrice, got.Price)
	})
}
