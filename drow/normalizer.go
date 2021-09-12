package drow

import (
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

type Normalizer struct {
	maxChars int
}

func NewNormalizer(opts ...NormalizeOpt) Normalizer {
	n := Normalizer{}
	for _, fn := range opts {
		fn(&n)
	}
	return n
}

type NormalizeOpt func(*Normalizer)

var NormalizeOption = NormalizeOpt(func(*Normalizer) {})

func (fn NormalizeOpt) WithMaxSize(m int) NormalizeOpt {
	return func(n *Normalizer) {
		fn(n)
		n.maxChars = m
	}
}

// Normalize to underscore
func (n Normalizer) Name(istr string) (string, error) {
	// if transform to to regular
	t := transform.Chain(norm.NFKD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	str, _, err := transform.String(t, istr)
	if err != nil {
		return "", err
	}

	splitter := func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	}
	rn := []rune(str)

	// toUpper := true
	res := make([]rune, 0, len(rn))
	wc := 0
	var last rune
	for _, r := range rn {
		switch {
		case splitter(r):
			wc = 0
			// toUpper = true
			if last != '_' {
				res = append(res, '_')
			}
			last = '_'
			continue
		case n.maxChars > 0 && wc > n.maxChars:
			continue
		default:
			res = append(res, unicode.ToLower(r))
			wc++
			// toUpper = false
			last = r
		}
	}
	return strings.Trim(string(res), "_"), nil
}
