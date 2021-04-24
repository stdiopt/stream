package strmutil

import (
	"errors"
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
	}
}

func IOWriter(w io.Writer) ProcFunc {
	return func(p Proc) error {
		return p.Consume(func(v interface{}) error {
			b, ok := v.([]byte)
			if !ok {
				return errors.New("wrong type")
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
		err := p.Consume(func(v interface{}) error {
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
