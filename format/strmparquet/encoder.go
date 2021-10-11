package strmparquet

import (
	"errors"
	"io"

	goparquet "github.com/fraugster/parquet-go"
	"github.com/fraugster/parquet-go/floor"
	"github.com/fraugster/parquet-go/parquet"
	"github.com/fraugster/parquet-go/parquetschema"
	strm "github.com/stdiopt/stream"
	"github.com/stdiopt/stream/drow"
	"github.com/stdiopt/stream/utils/strmio"
)

// Encode receives a drow or a struct and encodes into []bytes,
func Encode(opts ...encoderOpt) strm.Pipe {
	return strm.Func(func(p strm.Proc) (err error) {
		enc := newEncoder(strmio.AsWriter(p), opts...)
		defer func() {
			closeErr := enc.close()
			if err == nil {
				err = closeErr
			}
		}()

		return p.Consume(enc.write)
	})
}

// encoder encodes parquet format into a writer
type encoder struct {
	writer io.Writer
	fw     *floor.Writer

	rowGroupSize int64
	compression  parquet.CompressionCodec
}

type encoderOpt func(*encoder)

func newEncoder(w io.Writer, opts ...encoderOpt) encoder {
	enc := encoder{
		writer:       w,
		rowGroupSize: 1000,
		compression:  parquet.CompressionCodec_SNAPPY,
	}
	for _, fn := range opts {
		fn(&enc)
	}
	return enc
}

func (e encoder) close() error {
	if e.fw == nil {
		return nil
	}
	return e.fw.Close()
}

func (e *encoder) initWriter(v interface{}) error {
	if v == nil {
		return errors.New("cannot encode nil value")
	}
	var schema *parquetschema.SchemaDefinition
	switch v := v.(type) {
	case drow.Row:
		schema = schemaFromDROW(v)
	default:
		schema = schemaFromStruct(v) // might fail if not struct
	}
	pw := goparquet.NewFileWriter(e.writer,
		goparquet.WithMaxRowGroupSize(e.rowGroupSize),
		goparquet.WithCompressionCodec(e.compression), // make this as an option
		goparquet.WithSchemaDefinition(schema),
	)
	e.fw = floor.NewWriter(pw)
	return nil
}

func (e *encoder) write(v interface{}) error {
	if e.fw == nil {
		if err := e.initWriter(v); err != nil {
			return err
		}
	}
	switch v := v.(type) {
	case drow.Row:
		return e.fw.Write(drowMarshaller{v})
	default:
		return e.fw.Write(v)
	}
}

func WithEncodeRowGroupSize(n int64) encoderOpt {
	return func(e *encoder) {
		e.rowGroupSize = n
	}
}
