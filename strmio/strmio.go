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

func AsReader(p strm.Proc) ProcReader {
	done := make(chan struct{})
	pr, pw := io.Pipe()
	go func() {
		defer close(done)
		// nolint: errcheck
		pw.CloseWithError(p.Consume(func(b []byte) error {
			_, err := pw.Write(b)
			return err
		}))
	}()
	return ProcReader{pr, done}
}

func AsWriter(s strm.Sender) ProcWriter {
	return ProcWriter{s}
}

type ProcReader struct {
	*io.PipeReader
	done chan struct{}
}

func (r ProcReader) CloseWithError(err error) error {
	defer func() { <-r.done }()
	return r.PipeReader.CloseWithError(err)
}

func (r ProcReader) Close() error {
	defer func() { <-r.done }()
	return r.PipeReader.CloseWithError(nil)
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
