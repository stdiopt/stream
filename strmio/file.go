package strmio

import (
	"io"
	"os"

	"github.com/stdiopt/stream"
)

/*func FileWrite(path string) stream.ProcFunc {
	return stream.Func(func(p stream.Proc) error {
		tmpl, err := template.New("path").Parse(path)
		if err != nil {
			return err
		}

		var lastFile string
		var f *os.File
		defer func() {
			if f != nil {
				f.Close()
			}
		}()
		return p.Consume(func(buf []byte) error {
			pathBuf := &bytes.Buffer{}
			if err := tmpl.Execute(pathBuf, p.Meta()); err != nil {
				return err
			}
			fpath := pathBuf.String()
			// Create another file
			if lastFile != fpath {
				lastFile = fpath
				if f != nil {
					f.Close()
				}
				dir := filepath.Dir(fpath)
				if err := os.MkdirAll(dir, os.FileMode(0755)); err != nil {
					return err
				}
				nf, err := os.OpenFile(fpath, os.O_RDWR|os.O_CREATE|os.O_APPEND, os.FileMode(0644))
				if err != nil {
					nf, err = os.Create(fpath)
					log.Println("PAth error:", err)
				}
				if err != nil {
					return err
				}
				f = nf
			}
			_, err := f.Write(buf)
			return err
		})
	})
}*/

// FileRead consume a filepath as a string and produces byte chunks from the file
/*func FileRead() stream.ProcFunc {
	return stream.Func(func(p stream.Proc) error {
		w := AsWriter(p)
		err := p.Consume(func(path string) error {
			f, err := os.Open(path)
			if err != nil {
				return err
			}
			_, err = io.Copy(w, f)
			return err
		})
		w.CloseWithError(err)
		return err
	})
}*/

func FileRead(path string) stream.ProcFunc {
	return stream.Func(func(p stream.Proc) error {
		w := AsWriter(p)
		err := p.Consume(func(interface{}) error {
			f, err := os.Open(path)
			if err != nil {
				return err
			}
			_, err = io.Copy(w, f)
			return err
		})
		w.CloseWithError(err)
		return err
	})
}
