package strmdrow

import (
	"unicode"

	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// As go columns
func normalizeGoField(istr string) (string, error) {
	isMn := func(r rune) bool {
		return unicode.Is(unicode.Mn, r) // Mn: nonspacing marks
	}
	t := transform.Chain(norm.NFKD, transform.RemoveFunc(isMn), norm.NFC)
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
	var last rune
	for _, r := range rn {
		if splitter(r) {
			wc = 0
			toUpper = true
			continue
		}
		wc++
		if wc > 4 && isVogal(last) {
			continue
		}
		if toUpper {
			res = append(res, unicode.ToUpper(r))
		} else {
			res = append(res, unicode.ToLower(r))
		}
		toUpper = false
		last = r
	}
	return string(res), nil
}
