package strmjson

import (
	"context"
	"encoding/json"
	"io"
	"reflect"

	"github.com/stdiopt/stream"
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
		ctx := p.Context()

		meta := stream.NewMeta()
		pr, pw := io.Pipe()
		dec := json.NewDecoder(pr)

		go func() {
			pw.CloseWithError(p.Consume(func(ctx context.Context, buf []byte) error {
				if m, ok := stream.MetaFromContext(ctx); ok {
					meta = meta.Merge(m)
				}

				_, err := pw.Write(buf)
				if err != nil {
					return pw.CloseWithError(err)
				}
				return err
			}))
		}()

		for {
			v := reflect.New(typ).Interface()

			err := dec.Decode(v)
			if err == io.EOF {
				break
			}
			if err != nil {
				pr.CloseWithError(err)
				return err
			}

			if vv, ok := v.(*interface{}); ok {
				v = *vv
			}

			ctx = stream.ContextWithMeta(ctx, meta)
			meta = stream.NewMeta()

			if err := p.Send(ctx, v); err != nil {
				pr.CloseWithError(err)
				return err
			}
		}
		return nil
	})
}

func Encode() stream.Processor {
	return stream.Func(func(p stream.Proc) error {
		ctx := p.Context()
		meta := stream.NewMeta()
		pr, pw := io.Pipe()

		go func() {
			enc := json.NewEncoder(pw)
			pw.CloseWithError(p.Consume(func(ctx context.Context, v interface{}) error {
				if m, ok := stream.MetaFromContext(ctx); ok {
					meta = meta.Merge(m)
				}
				return enc.Encode(v)
			}))
		}()

		buf := make([]byte, 4096)
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}
			n, err := pr.Read(buf)
			if err == io.EOF {
				break
			}
			if err != nil {
				return pr.CloseWithError(err)
			}
			sbuf := append([]byte{}, buf[:n]...)

			ctx = stream.ContextWithMeta(ctx, meta)
			meta = stream.NewMeta()

			if err := p.Send(ctx, sbuf); err != nil {
				pw.CloseWithError(err)
				return err
			}
		}
		return nil
	})
}

func Unmarshal(v interface{}) stream.Processor {
	if v == nil {
		var l interface{}
		v = &l
	}
	typ := reflect.Indirect(reflect.ValueOf(v)).Type()
	return stream.Func(func(p stream.Proc) error {
		return p.Consume(func(ctx context.Context, buf []byte) error {
			v := reflect.New(typ).Interface()

			if err := json.Unmarshal(buf, v); err != nil {
				return err
			}

			if vv, ok := v.(*interface{}); ok {
				v = *vv
			}

			return p.Send(ctx, v)
		})
	})
}

func Marshal() stream.Processor {
	return stream.Func(func(p stream.Proc) error {
		return p.Consume(func(ctx context.Context, v interface{}) error {
			buf, err := json.Marshal(v)
			if err != nil {
				return err
			}

			return p.Send(ctx, buf)
		})
	})
}

// Dump encodes the input as json into the writer
// TODO: {lpf} rename, this was meant for debug but might be good for general use
func Dump(w io.Writer) stream.Processor {
	return stream.Func(func(p stream.Proc) error {
		enc := json.NewEncoder(w)
		enc.SetIndent("", "  ")
		return p.Consume(func(ctx context.Context, v interface{}) error {
			if err := enc.Encode(v); err != nil {
				return err
			}
			return p.Send(ctx, v)
		})
	})
}
