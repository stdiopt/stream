package strmparquet

import (
	"errors"
	"os"
	"reflect"

	goparquet "github.com/fraugster/parquet-go"
	"github.com/fraugster/parquet-go/floor"
	"github.com/fraugster/parquet-go/parquet"
	strm "github.com/stdiopt/stream"
	"github.com/stdiopt/stream/strmio"
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

// Encode receives a T and outputs encoded parquet in []byte
func Encode() strm.Pipe {
	return strm.Func(func(p strm.Proc) error {
		var fw *floor.Writer
		defer func() {
			if fw != nil {
				fw.Close()
			}
		}()

		return p.Consume(func(v interface{}) error {
			if v == nil {
				return errors.New("cannot encode nil value")
			}
			if fw == nil {
				w := strmio.AsWriter(p)
				// Setup parquet Writer
				pw := goparquet.NewFileWriter(w,
					goparquet.WithMaxRowGroupSize(1e6),
					goparquet.WithSchemaDefinition(schemaFrom(v)),
					goparquet.WithCompressionCodec(parquet.CompressionCodec_SNAPPY),
				)
				fw = floor.NewWriter(pw)
			}
			return fw.Write(v)
		})
	})
}
