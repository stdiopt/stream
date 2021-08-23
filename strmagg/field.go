package strmagg

import "github.com/stdiopt/stream/strmrefl"

type FieldFunc = func(interface{}) (interface{}, error)

func Field(f ...interface{}) FieldFunc {
	return func(v interface{}) (interface{}, error) {
		return strmrefl.FieldOf(v, f...)
	}
}
