package strmrefl

import (
	"reflect"

	strm "github.com/stdiopt/stream"
)

// Unslice consumes slices and sends each slice element.
func Unslice() strm.Pipe {
	return strm.S(func(p strm.Sender, v interface{}) error {
		val := reflect.Indirect(reflect.ValueOf(v))
		if val.Type().Kind() != reflect.Slice {
			return p.Send(v)
		}

		for i := 0; i < val.Len(); i++ {
			if err := p.Send(val.Index(i).Interface()); err != nil {
				return err
			}
		}
		return nil
	})
}

// Slice consumes elements and creates a slice if either downstream is done or
// it reaches 'max' elements
func Slice(max int) strm.Pipe {
	return strm.Func(func(p strm.Proc) error {
		slices := map[reflect.Type]reflect.Value{}
		err := p.Consume(func(v interface{}) error {
			typ := reflect.TypeOf(v)
			sl, ok := slices[typ]
			if !ok {
				sl = reflect.New(reflect.SliceOf(typ)).Elem()
			}
			sl = reflect.Append(sl, reflect.ValueOf(v))
			slices[typ] = sl
			if max > 0 && sl.Len() >= max {
				if err := p.Send(sl.Interface()); err != nil {
					return err
				}
				slices[typ] = reflect.New(reflect.SliceOf(typ)).Elem()
			}
			return nil
		})
		if err != nil {
			return err
		}
		for _, v := range slices {
			if v.IsNil() {
				continue
			}
			if err := p.Send(v.Interface()); err != nil {
				return err
			}
		}
		return nil
	})
}
