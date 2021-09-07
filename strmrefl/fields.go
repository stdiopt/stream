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
		val := FieldOf(v, f...)
		if val == nil {
			return fmt.Errorf("field invalid: %v", f)
		}
		return s.Send(val)
	})
}

// FieldOf returns a field of the value v by walking through the field params.
func FieldOf(v interface{}, ff ...interface{}) interface{} {
	return fieldOf(reflect.ValueOf(v), ff...)
}

// Same as field but returns multiple fields as a []interface{}
// Good for arguments
func RowOf(v interface{}, mf ...Fields) []interface{} {
	res := make([]interface{}, len(mf))
	r := reflect.ValueOf(v)
	for i, ff := range mf {
		res[i] = fieldOf(r, ff...)
	}
	return res
}

// FieldOf returns a field of the value v by walking through the field params.
func fieldOf(val reflect.Value, ff ...interface{}) interface{} {
	cur := reflect.Indirect(val)
	for _, k := range ff {
		switch cur.Kind() {
		case reflect.Struct:
			// Should be optimized by a cache map
			cur = cachedByName(cur, k.(string))
		case reflect.Slice:
			i, ok := k.(int)
			if !ok {
				return nil
			}
			cur = cur.Index(i)
		case reflect.Map:
			cur = cur.MapIndex(reflect.ValueOf(k))
		default:
			return nil
		}
		cur = reflect.Indirect(cur)
		if !cur.IsValid() {
			return nil
		}
		if cur.Kind() == reflect.Interface {
			cur = cur.Elem()
		}
		// This will solve stuff with underlying interface types
		// (e.g: interface{} in map[string]interface{}) !?
		// cur = reflect.ValueOf(cur.Interface())
	}
	if !cur.IsValid() {
		return nil
	}
	return cur.Interface()
}

// FieldOf returns a field of the value v by walking through the field params.
/*func FieldOf(v interface{}, ff ...interface{}) (interface{}, error) {
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
			cur = cur.MapIndex(reflect.ValueOf(k))
		case reflect.Invalid:
			return nil, fmt.Errorf("invalid type %v", v)
		default:
			return nil, fmt.Errorf("invalid type '%T' for field: %q", v, k)
		}
		cur = reflect.Indirect(cur)
		// This will solve stuff with underlying interface types
		// (e.g: interface{} in map[string]interface{})
		cur = reflect.ValueOf(cur.Interface())
	}
	return cur.Interface(), nil
}*/
