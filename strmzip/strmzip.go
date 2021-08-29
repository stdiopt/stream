// Package strmzip allows to stream zip contents.
package strmzip

import (
	"errors"
	"io"
	"path/filepath"

	"github.com/krolaw/zipstream"
	strm "github.com/stdiopt/stream"
	"github.com/stdiopt/stream/strmio"
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
