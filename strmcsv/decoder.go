package strmcsv

import (
	"encoding/csv"
	"io"

	strm "github.com/stdiopt/stream"
	"github.com/stdiopt/stream/strmio"
)

// Decode receives bytes and produces []string fields
func Decode(comma rune) strm.Pipe {
	return strm.Func(func(p strm.Proc) error {
		rd := strmio.AsReader(p)
		defer rd.Close()

		csvReader := csv.NewReader(rd)
		csvReader.Comma = comma
		for {
			row, err := csvReader.Read()
			if err == io.EOF {
				return nil
			}
			if err != nil {
				rd.CloseWithError(err) // nolint: errcheck
				return err
			}
			if err := p.Send(row); err != nil {
				return err
			}
		}
	})
}

/*func DecodeAsStruct(comma rune, sample interface{}) strm.Pipe {
	return strm.Line(
		Decode(comma),
		AsStruct(sample),
	)
}*/

func DecodeAsDrow(comma rune) strm.Pipe {
	return strm.Line(
		Decode(comma),
		AsDrow(),
	)
}
