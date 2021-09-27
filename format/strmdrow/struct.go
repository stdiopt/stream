package strmdrow

import (
	"fmt"
	"reflect"

	strm "github.com/stdiopt/stream"
	"github.com/stdiopt/stream/drow"
)

type fromStruct struct {
	tag string
}

type fromStructOpt func(*fromStruct)

func WithFromStructTag(t string) fromStructOpt {
	return func(o *fromStruct) {
		o.tag = t
	}
}

var FromStructOption = fromStructOpt(func(*fromStruct) {})

func (fn fromStructOpt) SelectTag(t string) fromStructOpt {
	return func(o *fromStruct) {
		fn(o)
		o.tag = t
	}
}

func FromStruct(opts ...fromStructOpt) strm.Pipe {
	return strm.Func(func(p strm.Proc) error {
		o := fromStruct{}
		for _, fn := range opts {
			fn(&o)
		}

		typeChecked := false
		var typ reflect.Type
		var hdr *drow.Header
		var fieldsI []int
		return p.Consume(func(v interface{}) error {
			val := reflect.Indirect(reflect.ValueOf(v))
			if !typeChecked {
				if val.Kind() != reflect.Struct {
					return fmt.Errorf("expected struct, unsupported %T type", v)
				}
				typ = val.Type()
				fields := []drow.Field{}
				for i := 0; i < typ.NumField(); i++ {
					ftyp := typ.Field(i)
					if !ftyp.IsExported() {
						continue
					}
					fieldsI = append(fieldsI, i)
					fields = append(fields, drow.Field{
						Name: ftyp.Name,
						Type: ftyp.Type,
						Tag:  ftyp.Tag,
					})
				}
				hdr = drow.NewHeader(fields...)
			}
			typeChecked = true
			row := drow.NewWithHeader(hdr)
			for i, fi := range fieldsI {
				row.Values[i] = val.Field(fi).Interface()
			}
			return p.Send(row)
		})
	})
}

// How slow is what?, slow, no point doing it honestly
func AsStruct() strm.Pipe {
	return strm.Func(func(p strm.Proc) error {
		var typ reflect.Type
		return p.Consume(func(row drow.Row) error {
			if typ == nil {
				ntyp, err := drow.StructType(row)
				if err != nil {
					return err
				}
				typ = ntyp
			}

			val := reflect.New(typ).Elem()
			for i, v := range row.Values {
				val.Field(i).Set(reflect.ValueOf(v))
			}
			return p.Send(val.Interface())
		})
	})
}
