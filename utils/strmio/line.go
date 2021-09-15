package strmio

import (
	"bufio"
	"io"

	strm "github.com/stdiopt/stream"
)

func ScanLine() strm.Pipe {
	return strm.Func(func(p strm.Proc) error {
		r := AsReader(p)
		defer r.Close()

		w := AsWriter(p)

		s := bufio.NewScanner(r)
		for s.Scan() {
			err := s.Err()
			if err == io.EOF {
				return nil
			}
			if err != nil {
				return err
			}
			if _, err := w.Write(s.Bytes()); err != nil {
				return err
			}
		}
		return nil
	})
}
