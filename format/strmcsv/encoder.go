package strmcsv

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"

	strm "github.com/stdiopt/stream"
	"github.com/stdiopt/stream/drow"
	"github.com/stdiopt/stream/utils/strmio"
)

// Encode receives a []string and encodes into []bytes writing the hdr first if any.
// Accepts input
// - []string
// - []interface{}
// - drow.Drow
// - struct?
func Encode(comma rune) strm.Pipe {
	return strm.Func(func(p strm.Proc) (err error) {
		enc := newEncoder(strmio.AsWriter(p), comma)
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

func newEncoder(w io.Writer, comma rune) *encoder {
	cw := csv.NewWriter(w)
	cw.Comma = comma
	return &encoder{
		cw:    cw,
		comma: comma,
	}
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
