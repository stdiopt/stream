package strmutil

import (
	"context"
	"fmt"
	"reflect"
)

// Unslice consumes slices and sends each slice element.
func Unslice() ProcFunc {
	return func(p Proc) error {
		return p.Consume(func(ctx context.Context, v interface{}) error {
			val := reflect.Indirect(reflect.ValueOf(v))
			if val.Type().Kind() != reflect.Slice {
				return fmt.Errorf("not a slice: %T", v)
			}

			for i := 0; i < val.Len(); i++ {
				if err := p.Send(ctx, val.Index(i).Interface()); err != nil {
					return err
				}
			}
			return nil
		})
	}
}

// Slice consumes elements and creates a slice if either downstream is done or
// it reaches 'max' elements
func Slice(max int) ProcFunc {
	return func(p Proc) error {
		slices := map[reflect.Type]reflect.Value{}
		err := p.Consume(func(ctx context.Context, v interface{}) error {
			typ := reflect.TypeOf(v)
			sl, ok := slices[typ]
			if !ok {
				sl = reflect.New(reflect.SliceOf(typ)).Elem()
			}
			sl = reflect.Append(sl, reflect.ValueOf(v))
			slices[typ] = sl
			if max > 0 && sl.Len() >= max {
				if err := p.Send(ctx, sl.Interface()); err != nil {
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
			if err := p.Send(p.Context(), v.Interface()); err != nil {
				return err
			}
		}
		return nil
	}
}
