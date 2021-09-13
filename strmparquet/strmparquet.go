package strmparquet

import (
	"os"
	"reflect"

	goparquet "github.com/fraugster/parquet-go"
	"github.com/fraugster/parquet-go/floor"
	strm "github.com/stdiopt/stream"
)

// DecodeFile receives a string path and outputs T.
// Reason for lack of Decode is that we need a ReadSeeker
func DecodeFile(sample interface{}) strm.Pipe {
	typ := reflect.TypeOf(sample)
	return strm.S(func(p strm.Sender, path string) error {
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		pr, err := goparquet.NewFileReader(f)
		if err != nil {
			return err
		}
		fr := floor.NewReader(pr)
		for fr.Next() {
			v := reflect.New(typ)
			if err := fr.Scan(v.Interface()); err != nil {
				return err
			}
			if err := p.Send(v.Elem().Interface()); err != nil {
				return err
			}
		}
		return nil
	})
}
