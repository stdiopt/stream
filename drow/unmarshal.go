package drow

import (
	"fmt"
	"reflect"
	"strings"
)

func Unmarshal(r Row, v interface{}, opts ...UnmarshalOpt) error {
	u := unmarshal{}
	for _, fn := range opts {
		fn(&u)
	}
	return u.Unmarshal(r, v)
}

type Unmarshaler interface {
	UnmarshalDROW(Row) error
}

type unmarshal struct {
	tag string
}

func (u unmarshal) Unmarshal(row Row, v interface{}) error {
	if um, ok := v.(Unmarshaler); ok {
		return um.UnmarshalDROW(row)
	}

	val := reflect.ValueOf(v)
	if val.Type().Kind() != reflect.Ptr ||
		val.Type().Elem().Kind() != reflect.Struct {
		panic("param should be a pointer to struct")
	}
	val = val.Elem()
	typ := val.Type()

	for i := 0; i < typ.NumField(); i++ {
		ftyp := typ.Field(i)
		if !ftyp.IsExported() {
			continue
		}
		name := ftyp.Name
		if u.tag != "" {
			tag, ok := ftyp.Tag.Lookup(u.tag)
			if !ok {
				continue
			}
			tp := strings.Split(tag, ",")
			name = tp[0]
		}

		f := row.Meta(name)
		if f == nil {
			continue
		}
		if !ftyp.Type.ConvertibleTo(f.Type) {
			// continue? // panic?
			return fmt.Errorf("'%v' cannot be converted to '%v'", f.Type, ftyp.Type)
		}
		val.Field(i).Set(reflect.ValueOf(f.Value).Convert(ftyp.Type))

	}
	return nil
}

type UnmarshalOpt func(*unmarshal)

var UnmarshalOption = UnmarshalOpt(func(*unmarshal) {})

func (fn UnmarshalOpt) WithTag(t string) UnmarshalOpt {
	return func(u *unmarshal) {
		fn(u)
		u.tag = t
	}
}
