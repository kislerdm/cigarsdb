package dimension

import "strings"

type Country string

func (s Country) Convert() string {
	lookup := map[string]string{
		"ecuador":                 "Ecuador",
		"nicaragua":               "Nicaragua",
		"honduras":                "Honduras",
		"dominikanische republik": "Dominican Republic",
		"kuba":                    "Cuba",
		"brasilien":               "Brazil",
		"usa":                     "USA",
		"mexiko":                  "Mexico",
		"kamerun":                 "Cameroon",
		"sumatra":                 "Sumatra",
		"costa rica":              "Costa Rica",
		"indonesien":              "Indonesia",
		"panama":                  "Panama",
		"indonesia":               "Indonesia",
		"peru":                    "Peru",
		"java":                    "Java",
		"philippinen":             "Philippines",
		"san andres":              "San Andres",
		"italien":                 "Italy",
		"deutschland":             "Germany",
		"pennsylvania":            "Pennsylvania",
		"karibik":                 "Caribbean",
		"kanaren":                 "Canary Islands",
		"kanarische inseln":       "Canary Islands",
		"mosambik":                "Mozambique",
		"kolumbien":               "Colombia",
		"unbekannt / geheim":      "",
		"unbekannt":               "",
		"geheim":                  "",
		"ohne":                    "",
	}

	var o string
	tmp := strings.ToLower(string(s))
	tmp = strings.NewReplacer("_", " ", "-", " ").Replace(tmp)
	o, ok := lookup[tmp]
	if !ok {
		o = toCapFirstLetters(tmp)
	}
	return o
}
