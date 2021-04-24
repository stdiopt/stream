package strmutil

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// Field extracts A field from a struct and sends it forward
// on a map it will walk through map
// on a slice it's possible to have Field1.0.Field2
func Field(f string) ProcFunc {
	return func(p Proc) error {
		return p.Consume(func(v interface{}) error {
			val, err := FieldOf(f, v)
			if err != nil {
				return err
			}
			return p.Send(val)
		})
	}
}

type (
	FMap map[string]string
)

func FieldMap(target interface{}, fm FMap) ProcFunc {
	typ := reflect.Indirect(reflect.ValueOf(target)).Type()
	return func(p Proc) error {
		return p.Consume(func(v interface{}) error {
			sv := reflect.New(typ)
			vv := sv.Elem()
			for k, f := range fm {

				field := vv.FieldByName(k)
				if !field.IsValid() {
					return fmt.Errorf("field not found %q in %T", f, target)
				}

				// Maybe use template per field

				/*tmpl := template.New("/")
				tmpl, err := tmpl.Parse(string(f))
				if err != nil {
					return err
				}
				buf := &bytes.Buffer{}
				tmpl.Execute(buf, v)
				field.Set(reflect.ValueOf(buf.String()))
				*/

				val, err := FieldOf(f, v)
				if err != nil {
					return err
				}
				field.Set(reflect.ValueOf(fmt.Sprint(val)))
			}
			return p.Send(vv.Interface())
		})
	}
}

func FieldOf(p string, v interface{}) (interface{}, error) {
	if p == "." {
		return v, nil
	}
	pp := strings.Split(p, ".")
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
		default:
			return nil, fmt.Errorf("invalid type: %T", v)
		}
		cur = reflect.Indirect(cur)
		// This will solve stuff with underlying interface types
		// (e.g: interface{} in map[string]interface{})
		cur = reflect.ValueOf(cur.Interface())
	}
	return cur.Interface(), nil
}
