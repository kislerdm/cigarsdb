// Package storage defines the storage port to persist the data.
package storage

import "context"

type SpecializedRating struct {
	Who            string  `json:"who"`
	Year           string  `json:"year"`
	RatingOutOf100 float64 `json:"ratingOutOf100"`
}

// Record defines a fully denormalized data projection with a 360 overview of a cigar.
type Record struct {
	// Identification
	// Name cigar name.
	Name string `json:"name"`
	// URL link to the data source.
	URL string `json:"url"`
	// Brand manufacturer brand.
	Brand string `json:"brand"`
	// Series cigar's series.
	Series string `json:"series"`
	// VideoURLs URL to the videos with the maker's interview, or other sort of description.
	VideoURLs []string `json:"videoURLs,omitempty"`
	// Details freetext details, e.g., summary of the taste, or description of the cigar.
	Details map[string]string `json:"details,omitempty"`

	// Shape
	Diameter   float64 `json:"diameter_mm"`
	Ring       float64 `json:"ring"`
	Length     float64 `json:"length_mm"`
	LengthInch float64 `json:"length_inch"`
	// Format the cigar's format, e.g., robusto.
	Format string `json:"format"`

	// Manufacturing
	// Maker cigar maker, or blender who created the cigar, e.g., AJ Fernandez.
	Maker *string `json:"maker,omitempty"`
	// ManufactureOrigin the manufacturing country of origin.
	ManufactureOrigin string `json:"manufactureOrigin"`
	// TypeOfManufacturing how the cigar was manufactured.
	// Find details here: https://www.cigarworld.de/en/zigarrenlexikon/totalmente-a-mano.
	TypeOfManufacturing *string `json:"typeOfManufacturing,omitempty"`
	// Construction cigar's construction type, e.g., longfiller.
	Construction *string `json:"construction,omitempty"`
	// IsBoxpressed indicates if cigar is manufactured using the box-press technology.
	IsBoxpressed *bool `json:"isBoxpressed,omitempty"`
	// IsDiscontinued indicates is the cigar is no longer in making.
	IsDiscontinued *bool `json:"isDiscontinued,omitempty"`

	// Blend

	// WrapperOrigin the wrapper countries of origin.
	WrapperOrigin []string `json:"wrapperOrigin,omitempty"`
	// WrapperProperty the wrapper leaf type, e.g., Shade, Sun Grown etc.
	WrapperProperty []string `json:"wrapperProperty,omitempty"`
	// WrapperTobaccoVariety the wrapper outer leaf tobacco's variety.
	WrapperTobaccoVariety []string `json:"wrapperTobaccoVariety,omitempty"`

	// FillerOrigin the filler countries of origin.
	FillerOrigin []string `json:"fillerOrigin,omitempty"`
	// FillerProperty the filler property, e.g., Jalapa.
	FillerProperty []string `json:"fillerProperty,omitempty"`
	// FillerTobaccoVariety the filler property, e.g., Jalapa.
	FillerTobaccoVariety []string `json:"fillerTobaccoVariety,omitempty"`

	// BinderOrigin the binder countries of origin.
	BinderOrigin []string `json:"binderOrigin,omitempty"`
	// BinderProperty the binder leaf type.
	BinderProperty []string `json:"binderProperty,omitempty"`
	// BinderTobaccoVariety the binder property.
	BinderTobaccoVariety []string `json:"binderTobaccoVariety,omitempty"`

	// Color defines the wrapper's color.
	Color *string `json:"color,omitempty"`

	// IsFlavoured indicates if the cigar is flavoured.
	IsFlavoured *bool `json:"isFlavoured,omitempty"`
	// AromaProfileManufacturer Array of aroma flavours according to the data source / website.
	AromaProfileManufacturer []string `json:"aromaProfileManufacturer,omitempty"`
	// AromaProfileCommunity aroma flavours with their weights from 0 to 1 according to the community.
	AromaProfileCommunity *AromaProfileCommunity `json:"aromaProfileCommunity,omitempty"`

	// Experience
	Strength        *string `json:"strength,omitempty"`
	FlavourStrength *string `json:"flavourStrength,omitempty"`
	// SmokingDuration reference smoking duration in minutes.
	SmokingDuration *string `json:"smokingDuration,omitempty"`

	// Purchase
	Price float64 `json:"price"`

	// AdditionalNotes additional info, e.g., barrel-aged
	AdditionalNotes    *string             `json:"additionalNotes,omitempty"`
	SpecializedRatings []SpecializedRating `json:"specializedRatings,omitempty"`
}

func (r Record) IsEmpty() bool {
	return r.Name == ""
}

type AromaProfileCommunity struct {
	Weights       map[string]float64 `json:"weights"`
	NumberOfVotes int                `json:"numberOfVotes"`
}

type Writer interface {
	Write(ctx context.Context, r []Record) (ids []string, err error)
}

type Reader interface {
	Read(ctx context.Context, id string) (r Record, err error)
	ReadBulk(ctx context.Context, limit, page uint) (r []Record, nextPage uint, err error)
}

// ReadWriter defines the interface to write and read data in sync.
type ReadWriter interface {
	Writer
	Reader
}
