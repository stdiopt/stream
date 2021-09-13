package strmparquet

import (
	"fmt"
	"reflect"
	"time"

	"github.com/fraugster/parquet-go/floor/interfaces"
	"github.com/fraugster/parquet-go/parquet"
	"github.com/fraugster/parquet-go/parquetschema"
	"github.com/stdiopt/stream/drow"
)

// Might need schema def
type drowMarshaller struct {
	row drow.Row
}

func (m drowMarshaller) MarshalParquet(record interfaces.MarshalObject) error {
	for i, v := range m.row.Values {
		f := m.row.Header(i)
		field := record.AddField(f.Name)
		val := reflect.ValueOf(v)

		if val.Kind() == reflect.Ptr {
			if val.IsNil() { // do not set
				continue
			}
			val = val.Elem()
		}

		if val.Type().ConvertibleTo(timeTyp) {
			ts := val.Interface().(time.Time).UnixNano()
			ts /= 1e6
			field.SetInt64(ts)
			/*if elem := schemaDef.SchemaElement(); elem.LogicalType != nil {
				switch {
				case elem.GetLogicalType().IsSetDATE():
					days := int32(value.Interface().(time.Time).Sub(time.Unix(0, 0).UTC()).Hours() / 24)
					field.SetInt32(days)
					return nil
				case elem.GetLogicalType().IsSetTIMESTAMP():
					return m.decodeTimestampValue(elem, field, value)
				}
			}*/
			continue
		}
		// only native types for now
		switch val.Kind() {
		case reflect.Bool:
			field.SetBool(val.Bool())
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32:
			field.SetInt32(int32(val.Int()))
		case reflect.Int64:
			field.SetInt64(val.Int())
		case reflect.Uint, reflect.Uint8, reflect.Uint16:
			field.SetInt32(int32(val.Uint()))
		case reflect.Uint32, reflect.Uint64:
			field.SetInt64(int64(val.Uint()))
		case reflect.Float32:
			field.SetFloat32(float32(val.Float()))
		case reflect.Float64:
			field.SetFloat64(val.Float())
		case reflect.String:
			field.SetByteArray([]byte(val.String()))
		default:
			return fmt.Errorf("unsupported type %s", val.Type())
		}
	}
	return nil
}

// Build schema definition from reflection
func schemaFromDROW(d drow.Row) *parquetschema.SchemaDefinition {
	root := &parquetschema.SchemaDefinition{
		RootColumn: &parquetschema.ColumnDefinition{
			SchemaElement: &parquet.SchemaElement{Name: "Thing"},
		},
	}
	for i := 0; i < d.NumField(); i++ {
		h := d.Header(i)
		col := columnDef(h.Name, h.Type)
		root.RootColumn.Children = append(root.RootColumn.Children, col)
	}
	return root
}
