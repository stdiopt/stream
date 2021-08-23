package strmrefl

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/stdiopt/stream"
)

// Field extracts A field from a struct and sends it forward
// on a map it will walk through map
// on a slice it's possible to have Field1.0.Field2
func Field(f string) stream.Pipe {
	return stream.Func(func(p stream.Proc) error {
		return p.Consume(func(v interface{}) error {
			val, err := FieldOf(v, f)
			if err != nil {
				return err
			}
			return p.Send(val)
		})
	})
}

type (
	FMap map[string]string
)

func FieldMap(target interface{}, fm FMap) stream.Pipe {
	typ := reflect.Indirect(reflect.ValueOf(target)).Type()
	return stream.Func(func(p stream.Proc) error {
		return p.Consume(func(v interface{}) error {
			sv := reflect.New(typ)
			vv := sv.Elem()
			for k, f := range fm {

				field := vv.FieldByName(k)
				if !field.IsValid() {
					return fmt.Errorf("field not found %q in %T", f, target)
				}

				val, err := FieldOf(v, f)
				if err != nil {
					return err
				}
				field.Set(reflect.ValueOf(val))
			}
			return p.Send(vv.Interface())
		})
	})
}

// Field returns a field of the value v by walking through the separators
// - on a struct it will walk through the struct Fields
// - on a map[string]interface{} it will walk through map
// - on a slice it's possible to have Field1.0.Field2
func FieldOf(v interface{}, s string) (interface{}, error) {
	if s == "" || s == "." {
		return v, nil
	}
	pp := strings.Split(s, ".")

	cur := reflect.Indirect(reflect.ValueOf(v))
	for _, k := range pp {
		switch cur.Kind() {
		case reflect.Struct:
			cur = cur.FieldByName(k)
			if !cur.IsValid() {
				return reflect.Value{}, fmt.Errorf("struct: field invalid: %q of %T", k, v)
			}
		case reflect.Slice:
			i, err := strconv.ParseUint(k, 10, 64)
			if err != nil {
				return reflect.Value{}, fmt.Errorf("slice: field invalid: %q of %T", k, v)
			}
			cur = cur.Index(int(i))
		case reflect.Map:
			if cur.Type().Key().Kind() != reflect.String {
				return reflect.Value{}, fmt.Errorf("map: key invalid: %q of %T", k, v)
			}
			cur = cur.MapIndex(reflect.ValueOf(k))
			if !cur.IsValid() {
				return nil, nil
			}
		// case reflect.String:
		case reflect.Invalid:
			return nil, fmt.Errorf("invalid type: %T", v)
			// default:
			//	return nil, fmt.Errorf("invalid type: %T", v)
		}
		cur = reflect.Indirect(cur)
		// This will solve stuff with underlying interface types
		// (e.g: interface{} in map[string]interface{})
		cur = reflect.ValueOf(cur.Interface())
	}
	return cur.Interface(), nil
}
