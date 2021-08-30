package strmjson

import (
	"encoding/json"
	"io"
	"reflect"

	strm "github.com/stdiopt/stream"
	"github.com/stdiopt/stream/strmio"
)

// Decode parses the []byte input as json and send the object
// interface can the type of object that will parse and send
//		type MyType struct {
//			Name string
//		}
//		JSONDecode(&MyType{})
// if type is nil it will decode the input into an &interface{} which might
// produce different types map[string]interface{}, []interface{}, string, float64
// as the regular native json.Unmarshal
// if the input is not bytes it will error and cancel the pipeline
func Decode(v interface{}) strm.Pipe {
	if v == nil {
		var l interface{}
		v = &l
	}
	typ := reflect.Indirect(reflect.ValueOf(v)).Type()

	return strm.Func(func(p strm.Proc) error {
		rd := strmio.AsReader(p)
		defer rd.Close()
		dec := json.NewDecoder(rd)
		for {
			val := reflect.New(typ)
			err := dec.Decode(val.Interface())
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				break
			}
			if err != nil {
				rd.CloseWithError(err)
				return err
			}

			if vv, ok := val.Interface().(*interface{}); ok {
				v = *vv
			} else {
				v = val.Elem().Interface()
			}

			if err := p.Send(v); err != nil {
				rd.CloseWithError(err)
				return err
			}
		}
		return nil
	})
}

func Encode() strm.Pipe {
	return strm.Func(func(p strm.Proc) error {
		wr := strmio.AsWriter(p)
		enc := json.NewEncoder(wr)

		return p.Consume(enc.Encode)
	})
}

// Dump encodes the input as json into the writer
// TODO: {lpf} rename, this was meant for debug but might be good for general use
func Dump(w io.Writer) strm.Pipe {
	return strm.Func(func(p strm.Proc) error {
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return p.Consume(func(v interface{}) error {
			if err := enc.Encode(v); err != nil {
				return err
			}
			return p.Send(v)
		})
	})
}
