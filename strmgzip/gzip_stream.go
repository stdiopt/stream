package strmgzip

import (
	"compress/gzip"
	"io"

	"github.com/stdiopt/stream"
	"github.com/stdiopt/stream/strmio"
)

// Provide a way to send a Writer and still send Meta
func Zip(lvl int) stream.Processor {
	return stream.Func(func(p stream.Proc) error {
		wr := strmio.AsWriter(p)

		w, err := gzip.NewWriterLevel(wr, lvl)
		if err != nil {
			wr.CloseWithError(err)
			return err
		}
		defer func() {
			w.Flush()
			w.Close()
			wr.Close()
		}()
		return p.Consume(func(buf []byte) error {
			if _, err := w.Write(buf); err != nil {
				return err
			}
			return nil
		})
	})
}

func Unzip() stream.Processor {
	return stream.Func(func(p stream.Proc) error {
		rd := strmio.AsReader(p)
		gr, err := gzip.NewReader(rd)
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		wr := strmio.AsWriter(p)
		defer wr.Close()
		for {
			_, err := io.Copy(wr, gr)
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				return nil
			}
			if err != nil {
				return err
			}
		}
	})
}
