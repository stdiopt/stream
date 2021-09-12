package strmcsv

import (
	"fmt"
	"reflect"
	"strings"

	strm "github.com/stdiopt/stream"
	"github.com/stdiopt/stream/drow"
)

/*func AsStruct(sample interface{}) strm.Pipe {
	if sample == nil {
		panic("param sample is nil")
	}
	typ := reflect.Indirect(reflect.ValueOf(sample)).Type()
	return strm.Func(func(p strm.Proc) error {
		var indexMap map[int]int
		return p.Consume(func(row []string) error {
			if indexMap == nil {
				hdrMap := map[string]int{}
				for i, h := range row {
					hdrMap[h] = i
				}
				indexMap = map[int]int{}
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
					indexMap[index] = i
				}
				return nil
			}

			val := reflect.New(typ).Elem()
			for csvi, si := range indexMap {
				if csvi > len(row) {
					continue
				}
				val.Field(si).Set(reflect.ValueOf(row[csvi]))

			}
			return p.Send(val.Interface())
		})
	})
}*/

var stringTyp = reflect.TypeOf(string(""))

func AsDrow() strm.Pipe {
	return strm.Func(func(p strm.Proc) error {
		var header *drow.Header
		return p.Consume(func(row []string) error {
			if header == nil {
				fields := []drow.Field{}
				for _, h := range row {
					h := strings.TrimSpace(h)
					fields = append(fields, drow.Field{
						Name: h,
						Type: stringTyp,
						Tag:  reflect.StructTag(fmt.Sprintf(`csv:"%s"`, h)),
					})
				}
				header = drow.NewHeader(fields...)
				return nil
			}

			rec := drow.NewWithHeader(header)
			rec.Values = make([]interface{}, len(row))
			for i := range row {
				rec.Values[i] = row[i]
			}
			return p.Send(rec)
		})
	})
}
