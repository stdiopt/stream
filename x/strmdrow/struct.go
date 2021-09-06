package strmdrow

import (
	"reflect"
	"strings"

	strm "github.com/stdiopt/stream"
)

// How slow is what?, slow, no point doing it honestly
func AsStruct() strm.Pipe {
	return strm.Func(func(p strm.Proc) error {
		var typ reflect.Type
		return p.Consume(func(d Row) error {
			if typ == nil {
				typ = toStruct(d)
			}

			val := reflect.New(typ).Elem()
			for i, v := range d.Values {
				val.Field(i).Set(reflect.ValueOf(v))
			}
			return p.Send(val.Interface())
		})
	})
}

func toStruct(d Row) reflect.Type {
	fields := []reflect.StructField{}

	for i := range d.Values {
		h := d.Header(i)
		colName := strings.ToTitle(h.Name)
		fields = append(fields, reflect.StructField{
			Name: colName,
			Type: h.Type,
		})
	}

	return reflect.StructOf(fields)
}
