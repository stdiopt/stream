package strmutil

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
)

func FileReader(path string) ProcFunc {
	return func(p Proc) error {
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		return IOReader(f)(p)
	}
}

func IOReader(r io.Reader) ProcFunc {
	return func(p Proc) error {
		ctx := p.Context()
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
			if err := p.Send(ctx, b); err != nil {
				return err
			}
		}
		return nil
	}
}

// IOWithReader expects a io.ReadCloser as input and sends the reader bytes.
func IOWithReader() ProcFunc {
	return func(p Proc) error {
		return p.Consume(func(ctx context.Context, v interface{}) error {
			r, ok := v.(io.ReadCloser)
			if !ok {
				return fmt.Errorf("wrong type: want io.ReadCloser, got: %T", v)
			}
			defer r.Close()
			buf := make([]byte, 4096)
			for {
				n, err := r.Read(buf)
				if err == io.EOF {
					if n > 0 {
						return p.Send(p.Context(), buf[:n])
					}
					return nil
				}
				if err != nil {
					return err
				}
				b := append([]byte{}, buf[:n]...)
				if err := p.Send(ctx, b); err != nil {
					return err
				}
			}
		})
	}
}

// IOWriter consumes []byte and writes to io.Writer w.
func IOWriter(w io.Writer) ProcFunc {
	return func(p Proc) error {
		return p.Consume(func(_ context.Context, v interface{}) error {
			b, ok := v.([]byte)
			if !ok {
				return fmt.Errorf("wrong type: want []byte, got: %T", v)
			}
			_, err := w.Write(b)
			return err
		})
	}
}

type ReadErrorCloser interface {
	Read([]byte) (int, error)
	CloseWithError(error) error
}

func AsReader(p Proc) ReadErrorCloser {
	pr, pw := io.Pipe()
	go func() {
		err := p.Consume(func(_ context.Context, v interface{}) error {
			b, ok := v.([]byte)
			if !ok {
				return errors.New("input must be []byte")
			}
			if _, err := pw.Write(b); err != nil {
				return pw.CloseWithError(err)
			}
			return nil
		})
		pw.CloseWithError(err) // nolint: errcheck
	}()
	return pr
}
