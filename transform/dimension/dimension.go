// Package dimension defines the logic for attributes cardinality normalisation.
//
// For example, the values of the cigar's format "Pyramid", "Pirámide" and "Pirámide" will be converted to "Pirámide".
// Note that the Spanish and the singular nouns are used as opposed to English, or German and plural nouns.
package dimension

type Dimension interface {
	Convert() string
}
