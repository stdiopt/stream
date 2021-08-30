package strmcsv

import (
	"encoding/csv"
	"io"
	"reflect"

	strm "github.com/stdiopt/stream"
	"github.com/stdiopt/stream/strmio"
)

// Decode receives bytes and produces [][]string fields?
func DecodeAs(comma rune, v interface{}) strm.Pipe {
	if v == nil {
		panic("param v is nil")
	}
	typ := reflect.Indirect(reflect.ValueOf(v)).Type()
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
		hdrMap := map[string]int{}
		for i, h := range hdr {
			hdrMap[h] = i
		}

		for {
			row, err := csvReader.Read()
			if err == io.EOF {
				return nil
			}
			if err != nil {
				rd.CloseWithError(err) // nolint: errcheck
				return err
			}
			val := reflect.New(typ)
			for i := 0; i < typ.NumField(); i++ {
				ftyp := typ.Field(i)
				tag, ok := ftyp.Tag.Lookup("csv")
				if !ok {
					continue
				}
				index, ok := hdrMap[tag]
				if !ok {
					continue
				}
				val.Elem().Field(i).Set(reflect.ValueOf(row[index]))
			}

			if err := p.Send(val.Elem().Interface()); err != nil {
				return err
			}
		}
	})
}
