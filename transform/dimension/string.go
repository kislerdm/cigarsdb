package dimension

import (
	"bytes"
	"unicode"
)

func toCapFirstLetters(s string) string {
	var o bytes.Buffer
	for i, el := range s {
		switch i == 0 {
		case true:
			el = unicode.ToUpper(el)
		default:
			if s[i-1] == ' ' {
				el = unicode.ToUpper(el)
			}
		}
		_, _ = o.WriteRune(el)
	}
	return o.String()
}
