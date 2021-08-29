package strmfs

import (
	"io"
	"os"
	"path/filepath"

	strm "github.com/stdiopt/stream"
	"github.com/stdiopt/stream/x/strmio"
)

// WriteFile create a file in path and writes to it.
func WriteFile(path string) strm.Pipe {
	return strm.Func(func(p strm.Proc) error {
		closefn := func() {}
		defer closefn()

		var f *os.File
		return p.Consume(func(buf []byte) error {
			// Only create file if something received
			if f == nil {
				dir := filepath.Dir(path)
				if err := os.MkdirAll(dir, os.FileMode(0755)); err != nil {
					return err
				}

				nf, err := os.Create(path)
				if err != nil {
					return err
				}

				closefn = func() { f.Close() }
				f = nf
			}

			if _, err := f.Write(buf); err != nil {
				return err
			}
			return p.Send(buf)
		})
	})
}

func ReadFile() strm.Pipe {
	return strm.Func(func(p strm.Proc) error {
		return p.Consume(func(path string) error {
			w := strmio.AsWriter(p)
			f, err := os.Open(path)
			if err != nil {
				return err
			}
			_, err = io.Copy(w, f)
			return err
		})
	})
}
