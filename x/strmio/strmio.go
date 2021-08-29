package strmio

import (
	"io"

	strm "github.com/stdiopt/stream"
)

// Reader reads bytes from r and sends down the line.
func Reader(r io.Reader) strm.Pipe {
	return strm.Func(func(p strm.Proc) error {
		buf := make([]byte, 4096)
		isEOF := false
		for !isEOF {
			n, err := r.Read(buf)
			if err == io.EOF {
				isEOF = true
			} else if err != nil {
				return err
			}
			b := append([]byte{}, buf[:n]...)
			if err := p.Send(b); err != nil {
				return err
			}
		}
		return nil
	})
}

// WithReader expects a io.ReadCloser as input and sends the reader bytes.
func WithReader() strm.Pipe {
	return strm.S(func(s strm.Sender, r io.ReadCloser) error {
		defer r.Close()
		buf := make([]byte, 4096)
		for {
			n, err := r.Read(buf)
			if err == io.EOF {
				if n > 0 {
					return s.Send(buf[:n])
				}
				return nil
			}
			if err != nil {
				return err
			}
			b := append([]byte{}, buf[:n]...)
			if err := s.Send(b); err != nil {
				return err
			}
		}
	})
}

// Writer consumes []byte and writes to io.Writer
func Writer(w io.Writer) strm.Pipe {
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
		pw.CloseWithError(p.Consume(func(b []byte) error {
			_, err := pw.Write(b)
			return err
		}))
	}()
	return pr
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

func AsWriter(s strm.Sender) ProcWriter {
	return ProcWriter{s}
}
