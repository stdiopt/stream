package strmparquet

import (
	"reflect"
	"strings"
	"time"

	"github.com/fraugster/parquet-go/parquet"
	"github.com/fraugster/parquet-go/parquetschema"
)

var timeTyp = reflect.TypeOf(time.Time{})

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

		var colName string
		t, ok := ftyp.Tag.Lookup("parquet")
		switch {
		case ok && t == "-":
			continue
		case !ok:
			colName = strings.ToLower(ftyp.Name)
		default:
			colName = t
		}
		col := columnDef(colName, ftyp.Type)
		root.RootColumn.Children = append(root.RootColumn.Children, col)
	}
	return root
}

func columnDef(name string, typ reflect.Type) *parquetschema.ColumnDefinition {
	var ptyp parquet.Type
	var convTyp *parquet.ConvertedType
	var logTyp *parquet.LogicalType
	rep := parquet.FieldRepetitionType_REQUIRED

	// vi := val.Field(i).Interface()
	if typ.Kind() == reflect.Ptr {
		rep = parquet.FieldRepetitionType_OPTIONAL
		typ = typ.Elem()
	}
	switch typ.Kind() {
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
		if typ.Elem().Kind() != reflect.Uint8 {
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
		if !typ.ConvertibleTo(timeTyp) {
			panic("only time supported for now")
		}
		ptyp = parquet.Type_INT64
		logTyp = &parquet.LogicalType{
			TIMESTAMP: &parquet.TimestampType{
				IsAdjustedToUTC: true,
				Unit: &parquet.TimeUnit{
					MILLIS: &parquet.MilliSeconds{},
				},
			},
		}
	}
	return &parquetschema.ColumnDefinition{
		SchemaElement: &parquet.SchemaElement{
			Name:           name,
			Type:           &ptyp,
			RepetitionType: &rep,
			ConvertedType:  convTyp,
			LogicalType:    logTyp,
		},
	}
}
