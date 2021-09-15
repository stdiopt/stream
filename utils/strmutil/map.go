package strmutil

import (
	"reflect"

	strm "github.com/stdiopt/stream"
)

func Map(fn interface{}) strm.Pipe {
	fnVal := reflect.ValueOf(fn)
	fnTyp := fnVal.Type()
	if fnTyp.NumIn() != 1 {
		panic("map func should have 1 input")
	}
	if fnTyp.NumOut() != 1 {
		panic("map func should have 1 output")
	}

	return strm.S(func(s strm.Sender, v interface{}) error {
		var val reflect.Value
		if v == nil {
			val = reflect.New(fnTyp.In(0).Elem())
		} else {
			val = reflect.ValueOf(v)
			if val.Type().ConvertibleTo(fnTyp.In(0)) {
				val = val.Convert(fnTyp.In(0))
			} else {
				return strm.NewTypeMismatchError(
					fnTyp.In(0),
					v,
				)
			}
		}
		ret := fnVal.Call([]reflect.Value{val})
		return s.Send(ret[0].Interface())
	})
}

func FlatMap(fn interface{}) strm.Pipe {
	fnVal := reflect.ValueOf(fn)
	fnTyp := fnVal.Type()
	if fnTyp.NumIn() != 1 {
		panic("map func should have 1 input")
	}
	if fnTyp.NumOut() != 1 && fnTyp.Out(0).Kind() == reflect.Slice {
		panic("map func should have 1 output")
	}

	return strm.S(func(s strm.Sender, v interface{}) error {
		val := reflect.ValueOf(v)
		ret := fnVal.Call([]reflect.Value{val})

		for i := 0; i < ret[0].Len(); i++ {
			if err := s.Send(ret[0].Index(i).Interface()); err != nil {
				return err
			}
		}
		return nil
	})
}
