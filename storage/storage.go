// Package storage defines the storage port to persist the data.
package storage

import "context"

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
	Ring       int     `json:"ring"`
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
	// FillerOrigin the filler countries of origin.
	FillerOrigin []string `json:"fillerOrigin,omitempty"`
	// BinderOrigin the binder countries of origin.
	BinderOrigin []string `json:"binderOrigin,omitempty"`
	// WrapperType the wrapper leaf type, e.g., Sun Grown.
	WrapperType *string `json:"wrapperType,omitempty"`
	// OuterLeafTobaccoVariety the wrapper outer leaf tobacco's variety.
	OuterLeafTobaccoVariety *string `json:"outerLeafTobaccoVariety,omitempty"`
	// IsFlavoured indicates if the cigar is flavoured.
	IsFlavoured *bool `json:"isFlavoured,omitempty"`
	// AromaProfileManufacturer Array of aroma flavours according to the data source / website.
	AromaProfileManufacturer []string `json:"aromaProfileManufacturer,omitempty"`
	// AromaProfileCommunity aroma flavours with their weights from 0 to 1 according to the community.
	AromaProfileCommunity map[string]float64 `json:"aromaProfileCommunity,omitempty"`
	// Experience
	Strength        *string `json:"strength,omitempty"`
	FlavourStrength *string `json:"flavourStrength,omitempty"`
	// SmokingDuration reference smoking duration in minutes.
	SmokingDuration *string `json:"smokingDuration,omitempty"`
	// Purchase
	Price float64 `json:"price"`

	// AdditionalNotes additional info, e.g., barrel-aged
	AdditionalNotes *string `json:"additionalNotes,omitempty"`
}

type Writer interface {
	Write(ctx context.Context, r Record) (id string, err error)
	WriteBulk(ctx context.Context, r []Record) (ids []string, err error)
}

type Reader interface {
	Read(ctx context.Context, id string) (r Record, err error)
	ReadBulk(ctx context.Context, limit, page uint) (r []Record, nextPage uint, err error)
}

type Seeker interface {
	Seek(ctx context.Context, name string) (r Record, err error)
}

// ReadWriter defines the interface to write and read data in sync.
type ReadWriter interface {
	Writer
	Reader
}
