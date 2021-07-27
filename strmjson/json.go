package strmjson

import (
	"encoding/json"
	"io"
	"reflect"

	"github.com/stdiopt/stream"
	"github.com/stdiopt/stream/strmio"
)

// JSONDecode parses the []byte input as json and send the object
// interface can the type of object that will parse and send
//		type MyType struct {
//			Name string
//		}
//		JSONDecode(&MyType{})
// if type is nil it will decode the input into an &interface{} which might
// produce different types map[string]interface{}, []interface{}, string, float64
// as the regular native json.Unmarshal
// if the input is not bytes it will error and cancel the pipeline
func Decode(v interface{}) stream.Processor {
	if v == nil {
		var l interface{}
		v = &l
	}
	typ := reflect.Indirect(reflect.ValueOf(v)).Type()

	return stream.Func(func(p stream.Proc) error {
		rd := strmio.AsReader(p)
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
		/*pr, pw := io.Pipe()
		dec := json.NewDecoder(pr)
		go func() {
			pw.CloseWithError(p.Consume(func(buf []byte) error {
				_, err := pw.Write(buf)
				if err != nil {
					pw.CloseWithError(err)
					return err
				}
				return err
			}))
		}()

		for {
			val := reflect.New(typ)

			err := dec.Decode(val.Interface())
			if err == io.EOF {
				break
			}
			if err != nil {
				pr.CloseWithError(err)
				return err
			}

			if vv, ok := val.Interface().(*interface{}); ok {
				v = *vv
			} else {
				v = val.Elem().Interface()
			}

			if err := p.Send(v); err != nil {
				pr.CloseWithError(err)
				return err
			}
		}
		return nil*/
	})
}

func Encode() stream.Processor {
	return stream.Func(func(p stream.Proc) error {
		wr := strmio.AsWriter(p)
		defer wr.Close()
		enc := json.NewEncoder(wr)

		return p.Consume(func(v interface{}) error {
			return enc.Encode(v)
		})
	})
}

/*func Encode() stream.Processor {
	return stream.Func(func(p stream.Proc) error {
		pr, pw := io.Pipe()
		go func() {
			enc := json.NewEncoder(pw)
			pw.CloseWithError(p.Consume(func(v interface{}) error {
				return enc.Encode(v)
			}))
		}()

		buf := make([]byte, 4096)
		for {
			select {
			case <-p.Context().Done():
				return p.Context().Err()
			default:
			}
			n, err := pr.Read(buf)
			if err == io.EOF {
				continue
			}
			if err != nil {
				return pr.CloseWithError(err)
			}
			sbuf := append([]byte{}, buf[:n]...)

			if err := p.Send(sbuf); err != nil {
				pw.CloseWithError(err)
				return err
			}
		}
	})
}*/

func Unmarshal(v interface{}) stream.Processor {
	if v == nil {
		var l interface{}
		v = &l
	}
	typ := reflect.Indirect(reflect.ValueOf(v)).Type()
	return stream.Func(func(p stream.Proc) error {
		return p.Consume(func(buf []byte) error {
			v := reflect.New(typ).Interface()

			if err := json.Unmarshal(buf, v); err != nil {
				return err
			}

			if vv, ok := v.(*interface{}); ok {
				v = *vv
			}

			return p.Send(v)
		})
	})
}

func Marshal() stream.Processor {
	return stream.Func(func(p stream.Proc) error {
		return p.Consume(func(v interface{}) error {
			buf, err := json.Marshal(v)
			if err != nil {
				return err
			}

			return p.Send(buf)
		})
	})
}

// Dump encodes the input as json into the writer
// TODO: {lpf} rename, this was meant for debug but might be good for general use
func Dump(w io.Writer) stream.Processor {
	return stream.Func(func(p stream.Proc) error {
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
