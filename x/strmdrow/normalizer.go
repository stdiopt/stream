package strmdrow

import (
	"strings"
	"unicode"

	strm "github.com/stdiopt/stream"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

type normalizer struct {
	maxChars int
}

type NormalizeOpt func(*normalizer)

// Normalize to underscore
func (n normalizer) normalizeName(istr string) (string, error) {
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

func NormalizeColumns(opts ...NormalizeOpt) strm.Pipe {
	n := normalizer{}
	for _, fn := range opts {
		fn(&n)
	}
	var hdr Header
	return strm.S(func(s strm.Sender, v interface{}) error {
		d, ok := v.(Row)
		if !ok {
			return strm.NewTypeMismatchError(Row{}, v)
		}

		if hdr.Len() == 0 {
			for i := range d.Values {
				h := d.Header(i)
				colName, err := n.normalizeName(h.Name)
				if err != nil {
					return err
				}
				hdr.Add(Field{
					Name: colName,
					Type: h.Type,
				})
			}
		}
		d = d.WithHeader(&hdr)

		return s.Send(d)
	})
}

func WithMaxSize(m int) NormalizeOpt {
	return func(n *normalizer) {
		n.maxChars = m
	}
}
