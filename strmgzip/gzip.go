package strmgzip

import (
	"compress/gzip"
	"io"

	"github.com/stdiopt/stream"
)

// Full zip the thing
func ZipX(lvl int) stream.Processor {
	return stream.Func(func(p stream.Proc) error {
		return p.Consume(func(buf []byte) error {
			pr, pw := io.Pipe()

			go func() {
				gw, err := gzip.NewWriterLevel(pw, lvl)
				if err != nil {
					pw.CloseWithError(err)
				}
				defer func() {
					gw.Close()
					pw.CloseWithError(err)
				}()

				_, err = gw.Write(buf)
			}()

			rbuf, err := io.ReadAll(pr)
			if err != nil {
				return err
			}

			return p.Send(rbuf)
		})
	})
}

// UnzipX full unzip thing
func UnzipX() stream.Processor {
	return stream.Func(func(p stream.Proc) error {
		return p.Consume(func(buf []byte) error {
			pr, pw := io.Pipe()

			go func() {
				_, err := pw.Write(buf)
				pw.CloseWithError(err)
			}()

			gr, err := gzip.NewReader(pr)
			if err != nil {
				return err
			}
			defer gr.Close()

			rbuf, err := io.ReadAll(gr)
			if err != nil {
				return err
			}
			return p.Send(rbuf)
		})
	})
}
