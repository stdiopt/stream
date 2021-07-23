package strmgzip

import (
	"compress/gzip"
	"context"
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
		ctx := p.Context()

		meta := stream.NewMeta()

		pr, pw := io.Pipe()
		go func() {
			w, err := gzip.NewWriterLevel(pw, lvl)
			if err != nil {
				pw.CloseWithError(err)
				return
			}
			pw.CloseWithError(p.Consume(func(ctx context.Context, buf []byte) error {
				// might need lock
				if m, ok := stream.MetaFromContext(ctx); ok {
					meta = meta.Merge(m)
				}
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

			ctx = stream.ContextWithMeta(ctx, meta)
			meta = stream.Meta{}

			sbuf := append([]byte{}, buf[:n]...)
			if err := p.Send(ctx, sbuf); err != nil {
				pw.CloseWithError(err)
				return err
			}
		}
		return nil
	})
}

func Reader() stream.Processor {
	return stream.Func(func(p stream.Proc) error {
		ctx := p.Context()

		meta := stream.NewMeta()

		pr, pw := io.Pipe()
		go func() {
			pw.CloseWithError(p.Consume(func(ctx context.Context, buf []byte) error {
				if m, ok := stream.MetaFromContext(ctx); ok {
					meta = meta.Merge(m)
				}
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

			ctx = stream.ContextWithMeta(ctx, meta)
			meta = stream.NewMeta()

			sbuf := append([]byte{}, buf[:n]...)
			if err := p.Send(ctx, sbuf); err != nil {
				return pr.CloseWithError(err)
			}
		}
		return nil
	})
}
