package strmio

import (
	"errors"
	"io"

	"github.com/stdiopt/stream"
)

// Reader reads bytes from r and sends down the line.
func Reader(r io.Reader) stream.Pipe {
	return stream.Func(func(p stream.Proc) error {
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
func WithReader() stream.Pipe {
	return stream.Func(func(p stream.Proc) error {
		return p.Consume(func(r io.ReadCloser) error {
			defer r.Close()
			buf := make([]byte, 4096)
			for {
				n, err := r.Read(buf)
				if err == io.EOF {
					if n > 0 {
						return p.Send(buf[:n])
					}
					return nil
				}
				if err != nil {
					return err
				}
				b := append([]byte{}, buf[:n]...)
				if err := p.Send(b); err != nil {
					return err
				}
			}
		})
	})
}

// Writer consumes []byte and writes to io.Writer
func Writer(w io.Writer) stream.Pipe {
	return stream.Func(func(p stream.Proc) error {
		return p.Consume(func(b []byte) error {
			_, err := w.Write(b)
			if err != nil {
				return err
			}
			return p.Send(b)
		})
	})
}

type ReadErrorCloser interface {
	Read([]byte) (int, error)
	Close() error
	CloseWithError(error) error
}

func AsReader(p stream.Proc) ReadErrorCloser {
	pr, pw := io.Pipe()
	go func() {
		pw.CloseWithError(p.Consume(func(b []byte) error {
			_, err := pw.Write(b)
			return err
		}))
	}()
	return pr
}

type (
	pipeWriter = io.PipeWriter
	pipeReader = io.PipeReader
)

var ErrClosed = errors.New("closed")

type ProcWriter struct {
	sender stream.Sender
}

func (w ProcWriter) Write(buf []byte) (int, error) {
	sbuf := append([]byte{}, buf...)
	if err := w.sender.Send(sbuf); err != nil {
		return 0, err
	}
	return len(buf), nil
}

func AsWriter(s stream.Sender) ProcWriter {
	return ProcWriter{s}
}

/*type ProcWriter struct {
	proc stream.Proc
	done chan struct{}
	*pipeWriter
}

func (p ProcWriter) Close() error {
	if err := p.pipeWriter.Close(); err != nil {
		return err
	}
	<-p.done
	return nil
}

func (p ProcWriter) CloseWithError(err error) error {
	if err := p.pipeWriter.CloseWithError(err); err != nil {
		return err
	}
	<-p.done
	return nil
}

func AsWriter(p stream.Proc) *ProcWriter {
	done := make(chan struct{})
	pr, pw := io.Pipe()
	go func() {
		defer close(done)
		buf := make([]byte, 4096*4)
		for {
			n, err := pr.Read(buf)
			if err != nil {
				pr.CloseWithError(err)
				return
			}
			sbuf := append([]byte{}, buf[:n]...)
			if err := p.Send(sbuf); err != nil {
				pr.CloseWithError(err)
				return
			}
		}
	}()
	return &ProcWriter{
		proc:       p,
		done:       done,
		pipeWriter: pw,
	}
}*/
