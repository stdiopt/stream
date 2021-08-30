package strmio

import (
	"io"

	strm "github.com/stdiopt/stream"
)

// Reader reads bytes from r and sends down the line.
// consequent received values will do nothing, it will panic if reader is nil.
func Reader(r io.Reader) strm.Pipe {
	if r == nil {
		panic("reader is nil")
	}
	return strm.S(func(s strm.Sender, _ interface{}) error {
		w := AsWriter(s)
		_, err := io.Copy(w, r)
		return err
	})
}

// Writer consumes []byte and writes to io.Writer it will panic if writer is nil
func Writer(w io.Writer) strm.Pipe {
	if w == nil {
		panic("writer is nil")
	}
	return strm.S(func(s strm.Sender, b []byte) error {
		_, err := w.Write(b)
		if err != nil {
			return err
		}
		return s.Send(b)
	})
}

func AsReader(p strm.Proc) *io.PipeReader {
	pr, pw := io.Pipe()
	go func() {
		// nolint: errcheck
		pw.CloseWithError(p.Consume(func(b []byte) error {
			_, err := pw.Write(b)
			return err
		}))
	}()
	return pr
}

func AsWriter(s strm.Sender) ProcWriter {
	return ProcWriter{s}
}

type ProcWriter struct {
	sender strm.Sender
}

func (w ProcWriter) Write(buf []byte) (int, error) {
	sbuf := append([]byte{}, buf...)
	if err := w.sender.Send(sbuf); err != nil {
		return 0, err
	}
	return len(buf), nil
}
