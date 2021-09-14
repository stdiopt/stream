package strmdrow

import (
	"fmt"
	"reflect"
	"strings"

	strm "github.com/stdiopt/stream"
	"github.com/stdiopt/stream/drow"
)

type fromStruct struct {
	tag string
}

type fromStructOpt func(*fromStruct)

var FromStructOption = fromStructOpt(func(*fromStruct) {})

func (fn fromStructOpt) WithTag(t string) fromStructOpt {
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
		return p.Consume(func(v interface{}) error {
			val := reflect.ValueOf(v)
			if val.Kind() != reflect.Struct {
				return fmt.Errorf("expected struct, unsupported %T type", v)
			}
			typ := val.Type()
			row := drow.New()
			for i := 0; i < typ.NumField(); i++ {
				ftyp := typ.Field(i)
				name := ftyp.Name
				if o.tag != "" {
					tag, ok := ftyp.Tag.Lookup(o.tag)
					if !ok {
						continue
					}
					tp := strings.Split(tag, ",")
					name = tp[0]
				}
				row.SetOrAdd(name, val.Field(i).Interface())
			}
			return p.Send(row)
		})
	})
}
