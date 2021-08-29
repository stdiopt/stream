// Package strmzip allows to stream zip contents.
package strmzip

import (
	"archive/zip"
	"errors"
	"io"
	"os"
	"path/filepath"

	"github.com/krolaw/zipstream"
	strm "github.com/stdiopt/stream"
	"github.com/stdiopt/stream/strmutil"
	"github.com/stdiopt/stream/x/strmio"
	"golang.org/x/sync/errgroup"
)

// Stream files matched by pattern in the zip file.
func Stream(pattern string) strm.Pipe {
	return strm.Func(func(p strm.Proc) error {
		errDone := errors.New("done")
		r := strmio.AsReader(p)
		defer r.Close()
		w := strmio.AsWriter(p)

		pr, pw := io.Pipe()

		eg, _ := errgroup.WithContext(p.Context())
		eg.Go(func() error {
			zs := zipstream.NewReader(pr)
			for {
				hdr, err := zs.Next()
				if err == io.EOF {
					continue
				}
				if err != nil {
					r.CloseWithError(err)
					return err
				}
				if ok, err := filepath.Match(pattern, filepath.Base(hdr.Name)); !ok || err != nil {
					continue
				}
				_, err = io.Copy(w, zs)
				if err != nil {
					r.CloseWithError(err)
					return err
				}
			}
		})
		_, err := io.Copy(pw, r)
		if err != nil {
			return err
		}
		pw.CloseWithError(errDone)

		err = eg.Wait()
		if err == errDone {
			return nil
		}
		return err
	})
}

// EachArchive receives []byte and stores into a tempfile, it will send filename
// after EOF.
func EachArchive(pattern string, pps ...strm.Pipe) strm.Pipe {
	return strm.Func(func(p strm.Proc) error {
		r := strmio.AsReader(p)

		tmp, err := os.CreateTemp(os.TempDir(), "strmzip-*")
		if err != nil {
			return err
		}
		defer os.Remove(tmp.Name())

		// Copy to temporary file
		_, err = func() (int64, error) {
			defer tmp.Close()
			return io.Copy(tmp, r)
		}()
		if err != nil {
			return err
		}

		f, err := zip.OpenReader(tmp.Name())
		if err != nil {
			return err
		}
		defer f.Close()

		for _, e := range f.File {
			if ok, err := filepath.Match(pattern, e.Name); !ok || err != nil {
				continue
			}
			rd, err := e.Open()
			if err != nil {
				return err
			}
			err = func() error {
				defer rd.Close()
				return strm.RunWithContext(p.Context(),
					strmio.Reader(rd),
					strm.Line(pps...),
					strmutil.Pass(p),
				)
			}()
			if err != nil {
				return err
			}
		}
		return nil
	})
}