package strmdrow

import (
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// nolint: unused, deadcode
func isVogal(r rune) bool {
	switch unicode.ToLower(r) {
	case 'a', 'e', 'i', 'o', 'u':
		return true
	}
	return false
}

// nolint: unused, deadcode
func normalizeGoField(istr string) (string, error) {
	t := transform.Chain(norm.NFKD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	str, _, err := transform.String(t, istr)
	if err != nil {
		return "", err
	}

	splitter := func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	}
	rn := []rune(str)

	toUpper := true
	res := make([]rune, 0, len(rn))
	wc := 0
	// var last rune
	for _, r := range rn {
		if splitter(r) {
			wc = 0
			toUpper = true
			continue
		}
		wc++
		if toUpper {
			res = append(res, unicode.ToUpper(r))
		} else {
			res = append(res, unicode.ToLower(r))
		}
		toUpper = false
	}
	return string(res), nil
}
