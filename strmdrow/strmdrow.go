package strmdrow

import (
	"reflect"

	strm "github.com/stdiopt/stream"
	"github.com/stdiopt/stream/drow"
)

func init() {
	type consumerType = func(drow.Row) error
	type senderType = func(strm.Sender, drow.Row) error
	// Registers this type on consumer registry for a fast path
	strm.ConsumerRegistry[reflect.TypeOf(consumerType(nil))] = func(f interface{}) strm.ConsumerFunc {
		fn := f.(consumerType)
		return func(v interface{}) error {
			r, ok := v.(drow.Row)
			if !ok {
				return strm.NewTypeMismatchError(drow.Row{}, v)
			}
			return fn(r)
		}
	}
	strm.SRegistry[reflect.TypeOf(senderType(nil))] = func(f interface{}) strm.ProcFunc {
		fn := f.(senderType)
		return func(p strm.Proc) error {
			return p.Consume(func(r drow.Row) error {
				return fn(p, r)
			})
		}
	}
}

func NormalizeColumns(opts ...drow.NormalizeOpt) strm.Pipe {
	return strm.Func(func(p strm.Proc) error {
		n := drow.NewNormalizer(opts...)
		var hdr drow.Header
		return p.Consume(func(row drow.Row) error {
			if hdr.Len() == 0 {
				for i := range row.Values {
					h := row.Header(i)
					colName, err := n.Name(h.Name)
					if err != nil {
						return err
					}
					hdr.Add(drow.Field{
						Name: colName,
						Type: h.Type,
					})
				}
			}
			row = row.WithHeader(&hdr)
			return p.Send(row)
		})
	})
}

func Unmarshal(sample interface{}, opts ...drow.UnmarshalOpt) strm.Pipe {
	return strm.Func(func(p strm.Proc) error {
		typ := reflect.TypeOf(sample)
		return p.Consume(func(row drow.Row) error {
			val := reflect.New(typ)
			if err := drow.Unmarshal(row, val.Interface(), opts...); err != nil {
				return err
			}
			return p.Send(val.Elem().Interface())
		})
	})
}

// How slow is what?, slow, no point doing it honestly
func AsStruct() strm.Pipe {
	return strm.Func(func(p strm.Proc) error {
		var typ reflect.Type
		return p.Consume(func(row drow.Row) error {
			if typ == nil {
				ntyp, err := drow.StructType(row)
				if err != nil {
					return err
				}
				typ = ntyp
			}

			val := reflect.New(typ).Elem()
			for i, v := range row.Values {
				val.Field(i).Set(reflect.ValueOf(v))
			}
			return p.Send(val.Interface())
		})
	})
}
