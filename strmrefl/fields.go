package strmrefl

import (
	"fmt"
	"reflect"

	strm "github.com/stdiopt/stream"
)

type Fields []interface{}

func F(f ...interface{}) Fields {
	return Fields(f)
}

type FMap map[string]Fields

// Extract extracts A field from a struct and sends it forward
// on a map it will walk through map
// on a slice it's possible to have Field1.0.Field2
func Extract(f ...interface{}) strm.Pipe {
	return strm.S(func(s strm.Sender, v interface{}) error {
		val, err := FieldOf(v, f...)
		if err != nil {
			return err
		}
		return s.Send(val)
	})
}

func StructMap(target interface{}, fm FMap) strm.Pipe {
	if target == nil {
		panic("target value is nil")
	}
	typ := reflect.Indirect(reflect.ValueOf(target)).Type()
	return strm.S(func(p strm.Sender, v interface{}) error {
		sv := reflect.New(typ)
		vv := sv.Elem()
		for k, f := range fm {

			field := vv.FieldByName(k)
			if !field.IsValid() {
				return fmt.Errorf("field not found %q in %T", k, target)
			}

			val, err := FieldOf(v, f...)
			if err != nil {
				return err
			}
			field.Set(reflect.ValueOf(val))
		}
		return p.Send(vv.Interface())
	})
}

// FieldOf returns a field of the value v by walking through the field params.
func FieldOf(v interface{}, ff ...interface{}) (interface{}, error) {
	cur := reflect.Indirect(reflect.ValueOf(v))
	for _, k := range ff {
		switch cur.Kind() {
		case reflect.Struct:
			cur = cur.FieldByName(k.(string))
			if !cur.IsValid() {
				return nil, fmt.Errorf("struct: field invalid: %q of %T", k, v)
			}
		case reflect.Slice:
			i, ok := k.(int)
			if !ok {
				return nil, fmt.Errorf("slice: field invalid: %q of %T", k, v)
			}
			cur = cur.Index(i)
		case reflect.Map:
			//if cur.Type().Key().Kind() != reflect.String {
			//	return reflect.Value{}, fmt.Errorf("map: key invalid: %q of %T", k, v)
			//}
			cur = cur.MapIndex(reflect.ValueOf(k))
			continue
		case reflect.Invalid:
			return nil, fmt.Errorf("invalid type: %v", v)
		default:
			return nil, fmt.Errorf("invalid type: '%T' for field: %q", v, k)
		}
		cur = reflect.Indirect(cur)
		// This will solve stuff with underlying interface types
		// (e.g: interface{} in map[string]interface{})
		cur = reflect.ValueOf(cur.Interface())
	}
	return cur.Interface(), nil
}
