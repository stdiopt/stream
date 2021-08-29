package strmgzip

import (
	"compress/gzip"
	"io"

	strm "github.com/stdiopt/stream"
	"github.com/stdiopt/stream/x/strmio"
)

// Provide a way to send a Writer and still send Meta
func Zip(lvl int) strm.Pipe {
	return strm.Func(func(p strm.Proc) error {
		wr := strmio.AsWriter(p)

		w, err := gzip.NewWriterLevel(wr, lvl)
		if err != nil {
			return err
		}
		defer w.Close()
		return p.Consume(func(buf []byte) error {
			if _, err := w.Write(buf); err != nil {
				return err
			}
			return nil
		})
	})
}

func Unzip() strm.Pipe {
	return strm.Func(func(p strm.Proc) error {
		rd := strmio.AsReader(p)
		defer rd.Close()
		gr, err := gzip.NewReader(rd)
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		wr := strmio.AsWriter(p)
		// TODO: verify if we really need the loop, since we will receive until EOF regardless the underlying data?
		// for {
		_, err = io.Copy(wr, gr)
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			return nil
		}
		if err != nil {
			return err
		}
		//}
		return nil
	})
}
