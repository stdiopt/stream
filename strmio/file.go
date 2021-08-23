package strmio

import (
	"io"
	"os"
	"path/filepath"

	"github.com/stdiopt/stream"
)

// WriteFile create a file in path and writes to it.
func WriteFile(path string) stream.Pipe {
	return stream.Func(func(p stream.Proc) error {
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

func ReadFile() stream.Pipe {
	return stream.S(func(s stream.Sender, path string) error {
		w := AsWriter(s)
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		_, err = io.Copy(w, f)
		return err
	})
}
