package strmfs

import (
	"io/fs"
	"path/filepath"

	strm "github.com/stdiopt/stream"
)

// ListFiles will list and send file names on a given a path, path will be joined using
// filepath.Join.
func ListFiles(pattern string) strm.Pipe {
	return strm.S(func(s strm.Sender, path string) error {
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
