package strmcsv

import (
	"reflect"

	strm "github.com/stdiopt/stream"
)

func init() {
	type consumerType = func([]string) error
	type senderType = func(strm.Sender, []string) error
	strm.ConsumerRegistry[reflect.TypeOf(consumerType(nil))] = func(f interface{}) strm.ConsumerFunc {
		fn := f.(consumerType)
		return func(v interface{}) error {
			s, ok := v.([]string)
			if !ok {
				return strm.NewTypeMismatchError([]string(nil), v)
			}
			return fn(s)
		}
	}
	strm.SRegistry[reflect.TypeOf(senderType(nil))] = func(f interface{}) strm.ProcFunc {
		fn := f.(senderType)
		return func(p strm.Proc) error {
			return p.Consume(func(v []string) error {
				return fn(p, v)
			})
		}
	}
}
