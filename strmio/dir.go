package strmio

import (
	"io/fs"
	"path/filepath"

	"github.com/stdiopt/stream"
)

// ListFiles recursively and each send filename
func ListFiles() stream.PipeFunc {
	return stream.F(func(p stream.Proc, path string) error {
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

func Glob(pattern string) stream.PipeFunc {
	return stream.F(func(p stream.Proc, path string) error {
		return filepath.WalkDir(path, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			matched, err := filepath.Match(pattern, filepath.Base(path))
			if err != nil {
				return err
			}
			if matched {
				return p.Send(path)
			}
			return nil
		})
	})
}
