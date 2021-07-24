package strmio

import (
	"io"

	"github.com/stdiopt/stream"
)

// FileReader reads file and sends byte down the line.
/*func File(path string) stream.Processor {
	return stream.Func(func(p stream.Proc) error {
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		return Reader(f).Run(p)
	})
}*/

// Reader reads bytes from r and sends down the line.
func Reader(r io.Reader) stream.Processor {
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
func WithReader() stream.Processor {
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
func Writer(w io.Writer) stream.Processor {
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
	CloseWithError(error) error
}

func AsReader(p stream.Proc) ReadErrorCloser {
	pr, pw := io.Pipe()
	go func() {
		err := p.Consume(func(b []byte) error {
			if _, err := pw.Write(b); err != nil {
				pw.CloseWithError(err)
				return err
			}
			return nil
		})
		pw.CloseWithError(err) // nolint: errcheck
	}()
	return pr
}
