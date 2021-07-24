package strmgzip

import (
	"compress/gzip"
	"io"

	"github.com/stdiopt/stream"
)

// Create something that automate the byte proc and put it in io.Reader

/*func ByteOTransform(p stream.Proc, trans func(io.Writer) io.Writer) error {
	ctx := p.Context()
	meta := stream.Meta{}

	pr, pw := io.Pipe()
	go func() {
		w := trans(pw)
		pw.CloseWithError(p.Consume(func(ctx context.Context, buf []byte) error {
			meta.Merge(stream.MetaFromContext(ctx))
			if _, err := w.Write(buf); err != nil {
				return err
			}
			if f, ok := w.(interface{ Flush() error }); ok {
				return f.Flush()
			}
			return nil
		}))
	}()

	buf := make([]byte, 4096)
	for {
		n, err := pr.Read(buf)
		if err == io.EOF {
			break
		}

		if err != nil {
			pr.CloseWithError(err)
			return err
		}

		ctx = stream.ContextWithMeta(ctx, meta)
		meta = stream.Meta{}

		sbuf := append([]byte{}, buf[:n]...)
		if err := p.Send(ctx, sbuf); err != nil {
			pw.CloseWithError(err)
			return err
		}
	}
	return nil
}*/

// Provide a way to send a Writer and still send Meta
func Writer(lvl int) stream.Processor {
	return stream.Func(func(p stream.Proc) error {
		pr, pw := io.Pipe()
		go func() {
			w, err := gzip.NewWriterLevel(pw, lvl)
			if err != nil {
				pw.CloseWithError(err)
				return
			}
			pw.CloseWithError(p.Consume(func(buf []byte) error {
				if _, err := w.Write(buf); err != nil {
					return err
				}
				return w.Flush()
			}))
		}()

		buf := make([]byte, 4096)
		for {
			n, err := pr.Read(buf)
			if err == io.EOF {
				break
			}

			if err != nil {
				pr.CloseWithError(err)
				return err
			}

			sbuf := append([]byte{}, buf[:n]...)
			if err := p.Send(sbuf); err != nil {
				pw.CloseWithError(err)
				return err
			}
		}
		return nil
	})
}

func Reader() stream.Processor {
	return stream.Func(func(p stream.Proc) error {
		pr, pw := io.Pipe()
		go func() {
			pw.CloseWithError(p.Consume(func(buf []byte) error {
				_, err := pw.Write(buf)
				return err
			}))
		}()

		rd, err := gzip.NewReader(pr)
		if err != nil {
			return pr.CloseWithError(err)
		}
		defer rd.Close()
		buf := make([]byte, 4096)
		for {
			n, err := rd.Read(buf)
			if err == io.EOF {
				break
			}
			if err != nil {
				return pr.CloseWithError(err)
			}

			sbuf := append([]byte{}, buf[:n]...)
			if err := p.Send(sbuf); err != nil {
				return pr.CloseWithError(err)
			}
		}
		return nil
	})
}
