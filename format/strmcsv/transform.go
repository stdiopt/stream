package strmcsv

import (
	"fmt"
	"reflect"
	"strings"

	strm "github.com/stdiopt/stream"
	"github.com/stdiopt/stream/drow"
)

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
