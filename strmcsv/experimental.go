package strmcsv

import (
	"encoding/csv"
	"fmt"
	"reflect"

	strm "github.com/stdiopt/stream"
	"github.com/stdiopt/stream/strmio"
)

// experimental
func DecodeMap(comma rune, target interface{}, fields ...string) strm.Pipe {
	typ := reflect.Indirect(reflect.ValueOf(target)).Type()
	return strm.Func(func(p strm.Proc) error {
		rd := strmio.AsReader(p)
		defer rd.Close()

		csvReader := csv.NewReader(rd)
		csvReader.Comma = comma

		_, err := csvReader.Read()
		if err != nil {
			return err
		}
		for {
			row, err := csvReader.Read()
			if err != nil {
				rd.CloseWithError(err)
				return err
			}

			sv := reflect.New(typ)
			vv := sv.Elem()
			for i, f := range fields {
				field := vv.FieldByName(f)
				if !field.IsValid() {
					return fmt.Errorf("field not found %q in %T", f, target)
				}
				// Need to parse string into field
				field.Set(reflect.ValueOf(row[i]))
			}

			if err := p.Send(vv.Interface()); err != nil {
				return err
			}
		}
	})
}
