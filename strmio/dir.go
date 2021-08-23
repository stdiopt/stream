package strmio

import (
	"io/fs"
	"path/filepath"

	"github.com/stdiopt/stream"
)

// ListFiles recursively and each send filename
func ListFiles(pattern string) stream.Pipe {
	return stream.S(func(s stream.Sender, path string) error {
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
				return s.Send(path)
			}
			return nil
		})
	})
}
