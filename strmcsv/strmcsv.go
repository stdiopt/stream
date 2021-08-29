package strmcsv

// TODO: {lpf} Add read header option on decoder

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"

	strm "github.com/stdiopt/stream"
	"github.com/stdiopt/stream/strmrefl"
	"github.com/stdiopt/stream/x/strmio"
)

type column struct {
	hdr string
	fn  func(interface{}) (interface{}, error)
}

type encode struct {
	columns []column
}

type encodeOpt = func(*encode)

func Field(hdr string, f ...interface{}) encodeOpt {
	return func(e *encode) {
		e.columns = append(e.columns, column{
			hdr: hdr,
			fn: func(v interface{}) (interface{}, error) {
				return strmrefl.FieldOf(v, f...)
			},
		})
	}
}

// Encode receives a []string and encodes into []bytes writing the hdr first if any.
func Encode(comma rune, encodeOpts ...encodeOpt) strm.Pipe {
	e := encode{}
	for _, fn := range encodeOpts {
		fn(&e)
	}
	return strm.Func(func(p strm.Proc) (err error) {
		w := strmio.AsWriter(p)

		cw := csv.NewWriter(w)
		defer func() {
			cw.Flush()
			cwErr := cw.Error()
			if err == nil {
				err = cwErr
			}
		}()
		cw.Comma = comma

		headerSent := false
		return p.Consume(func(v interface{}) error {
			if !headerSent {
				hdr := []string{}
				for _, c := range e.columns {
					hdr = append(hdr, c.hdr)
				}
				if len(hdr) > 0 {
					cw.Write(hdr)
				}
				headerSent = true
			}
			if len(e.columns) == 0 {
				row, ok := v.([]string)
				if !ok {
					return fmt.Errorf(
						"invalid input type, requires []string when no columns defined",
					)
				}
				return cw.Write(row)
			}
			row := make([]string, len(e.columns))
			for i, c := range e.columns {
				raw, err := c.fn(v)
				if err != nil {
					return err
				}
				row[i] = fmt.Sprint(raw)
			}
			cw.Write(row)

			return cw.Error()
		})
	})
}

// Decode receives bytes and produces [][]string fields?
func Decode(comma rune) strm.Pipe {
	return strm.Func(func(p strm.Proc) error {
		rd := strmio.AsReader(p)
		defer rd.Close()

		csvReader := csv.NewReader(rd)
		csvReader.Comma = comma

		_, err := csvReader.Read()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		for {
			row, err := csvReader.Read()
			if err == io.EOF {
				return nil
			}
			if err != nil {
				rd.CloseWithError(err)
				return err
			}
			if err := p.Send(row); err != nil {
				return err
			}
		}
	})
}

func DecodeAsJSON(comma rune) strm.Pipe {
	return strm.Func(func(p strm.Proc) error {
		rd := strmio.AsReader(p)
		defer rd.Close()

		csvReader := csv.NewReader(rd)
		csvReader.Comma = comma

		hdr, err := csvReader.Read()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		w := strmio.AsWriter(p)
		enc := json.NewEncoder(w)
		for {
			row, err := csvReader.Read()
			if err == io.EOF {
				return nil
			}
			if err != nil {
				rd.CloseWithError(err)
				return err
			}

			m := map[string]interface{}{}
			for i, h := range hdr {
				m[h] = row[i]
			}
			if err := enc.Encode(m); err != nil {
				rd.CloseWithError(err)
				return err
			}
		}
	})
}
