package strmcsv

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"

	strm "github.com/stdiopt/stream"
	"github.com/stdiopt/stream/drow"
	"github.com/stdiopt/stream/strmio"
)

// Encode receives a []string and encodes into []bytes writing the hdr first if any.
// Accepts input
// - []string
// - []interface{}
// - drow.Drow
// - struct?
func Encode(comma rune, opts ...encoderOpt) strm.Pipe {
	return strm.Func(func(p strm.Proc) (err error) {
		enc := newEncoder(strmio.AsWriter(p), comma, opts...)
		defer func() {
			closeErr := enc.close()
			if err == nil {
				err = closeErr
			}
		}()
		return p.Consume(enc.write)
	})
}

type encoder struct {
	cw    *csv.Writer
	comma rune

	typeChecked bool
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

func (e *encoder) close() error {
	e.cw.Flush()
	return e.cw.Error()
}

func (e *encoder) initWriter(v interface{}) error {
	if v == nil {
		return errors.New("cannot encode nil value")
	}
	switch v := v.(type) {
	case []string:
		return nil
	case drow.Row:
		hdr := []string{}
		for i := 0; i < v.NumField(); i++ {
			h := v.Header(i)
			hdr = append(hdr, h.Name)
		}
		return e.cw.Write(hdr)
	default:
		return fmt.Errorf("type %T is not supported", v)
	}
}

func (e *encoder) write(v interface{}) error {
	if !e.typeChecked {
		if err := e.initWriter(v); err != nil {
			return err
		}
		e.typeChecked = true
	}
	switch v := v.(type) {
	case []string:
		e.cw.Write(v) // nolint: errcheck
	case drow.Row:
		row := make([]string, v.NumField())
		for i := range row {
			row[i] = fmt.Sprint(v.Value(i))
		}
		e.cw.Write(row) // nolint: errcheck
	}

	return e.cw.Error()
}
