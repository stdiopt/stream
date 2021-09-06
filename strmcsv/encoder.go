package strmcsv

import (
	"encoding/csv"
	"fmt"
	"io"
	"reflect"

	strm "github.com/stdiopt/stream"
	"github.com/stdiopt/stream/strmio"
	"github.com/stdiopt/stream/strmrefl"
)

// Encode receives a []string and encodes into []bytes writing the hdr first if any.
func Encode(comma rune, opts ...encoderOpt) strm.Pipe {
	return strm.Func(func(p strm.Proc) (err error) {
		enc := newEncoder(strmio.AsWriter(p), comma, opts...)
		defer func() {
			closeErr := enc.Close()
			if err == nil {
				err = closeErr
			}
		}()
		return p.Consume(enc.Write)
	})
}

type column struct {
	hdr string
	fn  func(interface{}) (interface{}, error)
}

type encoder struct {
	cw         *csv.Writer
	comma      rune
	columns    []column
	headerSent bool
}

type encoderOpt = func(*encoder)

func newEncoder(w io.Writer, comma rune, opts ...encoderOpt) *encoder {
	cw := csv.NewWriter(w)
	cw.Comma = comma
	enc := &encoder{
		cw:    cw,
		comma: comma,
	}
	for _, fn := range opts {
		fn(enc)
	}
	return enc
}

func (e *encoder) Close() error {
	e.cw.Flush()
	return e.cw.Error()
}

func (e *encoder) autoColumns(v interface{}) {
	if _, ok := v.([]string); ok {
		return
	}
	typ := reflect.TypeOf(v)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	if typ.Kind() != reflect.Struct {
		return
	}
	for i := 0; i < typ.NumField(); i++ {
		ftyp := typ.Field(i)
		if !ftyp.IsExported() {
			continue
		}
		name := ftyp.Name
		e.columns = append(e.columns, column{
			hdr: name,
			fn: func(v interface{}) (interface{}, error) {
				return strmrefl.FieldOf(v, name)
			},
		})
	}
}

func (e *encoder) sendHeader(v interface{}) {
	if len(e.columns) == 0 {
		e.autoColumns(v)
	}

	hdr := []string{}
	for _, c := range e.columns {
		hdr = append(hdr, c.hdr)
	}
	if len(hdr) > 0 {
		e.cw.Write(hdr) // nolint: errcheck
	}
	e.headerSent = true
}

func (e *encoder) Write(v interface{}) error {
	if !e.headerSent {
		e.sendHeader(v)
	}
	if len(e.columns) == 0 {
		row, ok := v.([]string)
		if !ok {
			return fmt.Errorf(
				"invalid input type, requires []string when no columns defined",
			)
		}
		return e.cw.Write(row)
	}
	row := make([]string, len(e.columns))
	for i, c := range e.columns {
		raw, err := c.fn(v)
		if err != nil {
			return err
		}
		row[i] = fmt.Sprint(raw)
	}
	e.cw.Write(row) // nolint: errcheck

	return e.cw.Error()
}

func Field(hdr string, f ...interface{}) encoderOpt {
	if len(f) == 0 {
		f = []interface{}{hdr}
	}
	return func(e *encoder) {
		e.columns = append(e.columns, column{
			hdr: hdr,
			fn: func(v interface{}) (interface{}, error) {
				return strmrefl.FieldOf(v, f...)
			},
		})
	}
}
