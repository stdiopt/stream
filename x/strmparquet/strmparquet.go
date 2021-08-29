package strmparquet

import (
	"errors"
	"log"
	"os"
	"reflect"
	"time"

	goparquet "github.com/fraugster/parquet-go"
	"github.com/fraugster/parquet-go/floor"
	"github.com/fraugster/parquet-go/parquet"
	"github.com/fraugster/parquet-go/parquetschema"
	strm "github.com/stdiopt/stream"
	"github.com/stdiopt/stream/x/strmio"
)

// DecodeFile receives a string path and outputs T.
func DecodeFile(sample interface{}) strm.Pipe {
	typ := reflect.TypeOf(sample)
	return strm.Func(func(p strm.Proc) error {
		return p.Consume(func(path string) error {
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
	})
}

// Encode receives a T and outputs encoded parquet in []byte
func Encode() strm.Pipe {
	return strm.Func(func(p strm.Proc) error {
		var fw *floor.Writer
		closefn := func() {}
		defer func() { closefn() }()

		return p.Consume(func(v interface{}) error {
			if v == nil {
				return errors.New("cannot encode nil value")
			}
			if fw == nil {
				log.Println("Create writer")
				w := strmio.AsWriter(p)
				// Setup parquet Writer
				pw := goparquet.NewFileWriter(w,
					goparquet.WithMaxRowGroupSize(100000),
					goparquet.WithSchemaDefinition(schemaFrom(v)),
					goparquet.WithCompressionCodec(parquet.CompressionCodec_SNAPPY),
				)
				fw = floor.NewWriter(pw)
				closefn = func() {
					fw.Close()
					pw.Close()
				}
			}
			return fw.Write(v)
		})
	})
}

// Build schema definition from reflection
func schemaFrom(v interface{}) *parquetschema.SchemaDefinition {
	val := reflect.Indirect(reflect.ValueOf(v))
	typ := val.Type()

	root := &parquetschema.SchemaDefinition{
		RootColumn: &parquetschema.ColumnDefinition{
			SchemaElement: &parquet.SchemaElement{Name: "Thing"},
		},
	}

	for i := 0; i < typ.NumField(); i++ {
		ftyp := typ.Field(i)
		t, ok := ftyp.Tag.Lookup("parquet")
		if !ok {
			continue
		}

		var ptyp parquet.Type
		var convTyp *parquet.ConvertedType
		var logTyp *parquet.LogicalType
		var rep parquet.FieldRepetitionType = parquet.FieldRepetitionType_REQUIRED

		ityp := val.Field(i).Type()
		// vi := val.Field(i).Interface()
		if val.Field(i).Kind() == reflect.Ptr {
			rep = parquet.FieldRepetitionType_OPTIONAL
			ityp = val.Field(i).Type().Elem()
		}
		switch ityp.Kind() {
		case reflect.Int, reflect.Int32:
			ptyp = parquet.Type_INT32
		case reflect.Int64:
			ptyp = parquet.Type_INT64
		case reflect.Float32:
			ptyp = parquet.Type_FLOAT
		case reflect.Float64:
			ptyp = parquet.Type_DOUBLE
		case reflect.String:
			ptyp = parquet.Type_BYTE_ARRAY
			convTyp = new(parquet.ConvertedType)
			*convTyp = parquet.ConvertedType_UTF8
			logTyp = &parquet.LogicalType{
				STRING: &parquet.StringType{},
			}
		case reflect.Slice:
			if ityp.Elem().Kind() != reflect.Uint8 {
				panic("wrong type")
			}

			ptyp = parquet.Type_BYTE_ARRAY
			convTyp = new(parquet.ConvertedType)
			*convTyp = parquet.ConvertedType_UTF8
			logTyp = &parquet.LogicalType{
				STRING: &parquet.StringType{},
			}
			// regular slice
		case reflect.Struct:
			if ityp != reflect.TypeOf(time.Time{}) {
				panic("only type supported for now")
			}
			ptyp = parquet.Type_INT64
			logTyp = &parquet.LogicalType{
				TIMESTAMP: &parquet.TimestampType{
					IsAdjustedToUTC: true,
					Unit: &parquet.TimeUnit{
						NANOS: &parquet.NanoSeconds{},
					},
				},
			}
		}
		col := &parquetschema.ColumnDefinition{
			SchemaElement: &parquet.SchemaElement{
				Name:           t,
				Type:           &ptyp,
				RepetitionType: &rep,
				ConvertedType:  convTyp,
				LogicalType:    logTyp,
			},
		}
		root.RootColumn.Children = append(root.RootColumn.Children, col)
	}
	return root
}
