package main

import (
	"bytes"
	_ "embed"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/net/html"
)

// the html contain the list with three items, two of which correspond to a single unit
//
//go:embed list.html
var list []byte

func Test_readList(t *testing.T) {
	want := []Record{
		{
			Name:           "Diesel Crucible Limited Edition 2021 Toro",
			URL:            "https://www.noblego.de/diesel-crucible-toro-zigarren/",
			OriginCountry:  "Nicaragua",
			Format:         "Toro",
			Form:           "Boxpressed",
			WrapperCountry: []string{"Ecuador"},
			FillerCountry:  []string{"Nicaragua"},
			Strength:       "Medium",
			Price:          11.5,
			Diameter:       19.8,
			Length:         152,
			Aromas:         []string{"Espresso", "Nougat", "Nuss", "Röstaromen", "Schokolade", "Schwarzer Pfeffer"},
			include:        true,
		},
	}
	got, err, warn := readList(bytes.NewReader(list))
	assert.NoError(t, err)
	assert.ErrorContains(t, warn, "no required data found")
	assert.Len(t, got, 1)
	assert.Equal(t, want, got)
}

func Test_readSpanChildrenValue(t *testing.T) {
	doc := strings.NewReader(`<span class="product-attribute-cig_diameter" title="Durchmesser in mm">
	<span class="label">Ø</span>
	<span class="value">19.8</span>
</span>`)
	node, err := html.Parse(doc)
	assert.NoError(t, err)
	want := "19.8"
	assert.Equal(t, want, readSpanChildrenValue(node))
}

func Test_readUrlAndName(t *testing.T) {
	doc := strings.NewReader(`<h2 class="product-name">
	<a href="https://www.noblego.de/reposado-estate-blend-colorado-robusto-zigarren/" title="Reposado Estate Blend Colorado Robusto">
	Reposado Estate Blend Colorado Robusto</a>
</h2>`)
	node, err := html.Parse(doc)
	assert.NoError(t, err)
	want := Record{
		URL:  "https://www.noblego.de/reposado-estate-blend-colorado-robusto-zigarren/",
		Name: "Reposado Estate Blend Colorado Robusto",
	}
	got := Record{}
	readUrlAndName(node, &got)
	assert.Equal(t, want, got)
}

func Test_readListProductDetails(t *testing.T) {
	doc := strings.NewReader(`<div class="product-attributes">
	<ul>
		<li class="product-attribute-herkunft">
			<span class="label">Herkunft</span>
			<span class="data">                            <a href="https://www.noblego.de/zigarren-nicaragua/">Nicaragua</a>
		</span>
		</li>
		<li class="product-attribute-cig_duration">
			<span class="label">Rauchdauer</span>
			<span class="data">45 bis 60 Min</span>
		</li>
		<li class="product-attribute-cig_size">
			<span class="label">Format</span>
			<span class="data">                            <a href="https://www.noblego.de/robusto-zigarren/">Robusto</a>
		</span>
		</li>
		<li class="product-attribute-cig_wrapper_origin">
			<span class="label">Deckblattherkunft</span>
			<span class="data">Ecuador</span>
		</li>
		<li class="product-attribute-cig_filler">
			<span class="label">Einlage</span>
			<span class="data">Nicaragua</span>
		</li>
		<li class="product-attribute-cig_aroma">
			<span class="label">Aroma</span>
			<span class="data">Süß, Würzig, Zedernholz</span>
		</li>
		<li class="product-attribute-cig_form">
			<span class="label">Form</span>
			<span class="data">Rund</span>
		</li>
</ul>
</div>`)
	node, err := html.Parse(doc)
	assert.NoError(t, err)
	want := Record{
		OriginCountry:  "Nicaragua",
		Format:         "Robusto",
		Form:           "Rund",
		WrapperCountry: []string{"Ecuador"},
		FillerCountry:  []string{"Nicaragua"},
		Aromas:         []string{"Süß", "Würzig", "Zedernholz"},
	}
	got := Record{}
	readListProductDetails(node, &got)
	assert.Equal(t, want, got)
}

func Test_readListProductPrice(t *testing.T) {
	doc := strings.NewReader(`<div class="add-to-cart">
												<form action="https://www.noblego.de/checkout/cart/add/" method="post">
													<ul class="product-prices">
														<li>
<span class="product-price-option-img">
	<img src="https://www.noblego.de/media/catalog/product/cache/2/small_image/89x/9df78eab33525d08d6e5fb8d27136e95/r/e/reposado_estate_blend_colorado_robusto_01_1.jpg" data-dest-src="https://www.noblego.de/media/catalog/product/cache/2/small_image/363x/9df78eab33525d08d6e5fb8d27136e95/r/e/reposado_estate_blend_colorado_robusto_01_1.jpg">
</span>
															<span class="product-price-option-and-availability">
					<span class="product-price-option-text product-price-option-type-cig_package_size">
		<span title="Verpackungseinheit">10er</span>
		<span class="availability-icon availability-suppliable-instock" title="Auf Lager und versandfertig "></span>
	</span>
	<span class="availability">
		<span class="availability-text">Auf Lager und versandfertig </span>
	</span>
</span>

															<span class="product-price">
	<span class="price">20,37&nbsp;€</span>                                                                                <span class="product-price-option-discount">inkl. 3% Rabatt</span>
													</span>

															<div class="qty-wrapper">
																<input type="text" class="input-text qty" name="qty[18555]">
																<div class="qty-buttons-wrapper">
																	<div class="qty-button increase"></div>
																	<div class="qty-button decrease"></div>
																</div>
															</div>
														</li>
														<li>
<span class="product-price-option-img">
	<img src="https://www.noblego.de/media/catalog/product/cache/2/small_image/89x/9df78eab33525d08d6e5fb8d27136e95/r/e/reposado_estate_blend_colorado_robusto_05_1.jpg" data-dest-src="https://www.noblego.de/media/catalog/product/cache/2/small_image/363x/9df78eab33525d08d6e5fb8d27136e95/r/e/reposado_estate_blend_colorado_robusto_05_1.jpg">
</span>
															<span class="product-price-option-and-availability">
					<span class="product-price-option-text product-price-option-type-cig_package_size">
		<span title="Verpackungseinheit">Einzeln</span>
		<span class="availability-icon availability-suppliable-instock" title="Auf Lager und versandfertig "></span>
	</span>
	<span class="availability">
		<span class="availability-text">Auf Lager und versandfertig </span>
	</span>
</span>

															<span class="product-price">
	<span class="price">2,10&nbsp;€</span>                                            </span>

															<div class="qty-wrapper">
																<input type="text" class="input-text qty" name="qty[18556]">
																<div class="qty-buttons-wrapper">
																	<div class="qty-button increase"></div>
																	<div class="qty-button decrease"></div>
																</div>
															</div>
														</li>
													</ul>
													<div class="add-to-box">
														<p class="validation-advice" style="display: none">Bitte geben Sie oben an, welche Menge Sie bestellen möchten!</p>
														<div class="add-to-box-text">Preise inkl. 19% MwSt zzgl. Versandkosten. <b>Versandkostenfrei mit DHL</b> ab 60€ Warenwert!</div>
														<button type="submit" title="In den Warenkorb" class="button btn-cart">
		<span>
			<span class="icon ic ic-cart"></span>
			<span>In den Warenkorb</span>
		</span>
	</button>
	</div></form></div>`)
	node, err := html.Parse(doc)
	assert.NoError(t, err)
	want := Record{Price: 2.1, include: true}
	got := Record{}
	readListProductPrice(node, &got)
	assert.Equal(t, want, got)
}

//go:embed details.html
var details []byte

func Test_readDetails(t *testing.T) {
	want := Record{
		Brand:           "Diesel",
		Collection:      "Crucible",
		Blender:         "AJ Fernandez",
		FillingType:     "Longfiller",
		WrapperType:     "Ecuador",
		BinderCountry:   []string{"Ecuador"},
		Gauge:           50,
		Aroma:           "Medium-aromatisch",
		LimitedEdition:  true,
		SmokingDuration: "60 bis 90 Min",
	}
	got := Record{}
	assert.NoError(t, readDetails(bytes.NewReader(details), &got))
	assert.Equal(t, want, got)
}
