package strmrefl

import (
	"reflect"
)

var structCache = map[reflect.Type]*structType{}

type structType struct {
	typ    reflect.Type
	fields map[string]int
}

func cachedByName(v reflect.Value, name string) reflect.Value {
	typ := v.Type()
	st, ok := structCache[typ]
	if !ok {
		st = &structType{typ: typ, fields: map[string]int{}}
		for i := 0; i < typ.NumField(); i++ {
			st.fields[typ.Field(i).Name] = i
		}
		structCache[typ] = st
	}
	fi, ok := st.fields[name]
	if !ok {
		return reflect.Value{}
	}
	return v.Field(fi)
}
