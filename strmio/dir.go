package strmio

import (
	"io/fs"
	"path/filepath"

	"github.com/stdiopt/stream"
)

// Read dir recursively and each send filename
func DirPaths(path string) stream.ProcFunc {
	return stream.F(func(p stream.Proc, _ interface{}) error {
		return filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			return p.Send(path)
		})
	})
}
