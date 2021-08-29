package strmcsv

import (
	"encoding/csv"
	"encoding/json"
	"io"
	"regexp"

	strm "github.com/stdiopt/stream"
	"github.com/stdiopt/stream/x/strmio"
)

// Encode receives a []string and encodes into []bytes writing the hdr first if any.
func Encode(comma rune, hdr ...string) strm.Pipe {
	return strm.Func(func(p strm.Proc) error {
		w := strmio.AsWriter(p)

		cw := csv.NewWriter(w)
		defer cw.Flush()
		cw.Comma = comma
		if len(hdr) > 0 {
			if err := cw.Write(hdr); err != nil {
				return err
			}
		}
		return p.Consume(func(row []string) error {
			return cw.Write(row)
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

func DecodeMatch(comma rune, fields ...string) strm.Pipe {
	return strm.Func(func(p strm.Proc) error {
		rd := strmio.AsReader(p)
		defer rd.Close()
		csvReader := csv.NewReader(rd)
		csvReader.Comma = comma

		hdr, err := csvReader.Read()
		if err != nil {
			return err
		}
		indexes := make([]int, len(fields))
		for i, f := range fields {
			re, err := regexp.Compile(f)
			if err != nil {
				return err
			}
			indexes[i] = -1
			for hi, h := range hdr {
				if re.MatchString(h) {
					indexes[i] = hi
					break
				}
			}
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

			res := make([]string, len(fields))
			for i := range res {
				if indexes[i] == -1 {
					continue
				}
				res[i] = row[indexes[i]]
			}

			if err := p.Send(res); err != nil {
				rd.CloseWithError(err)
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
