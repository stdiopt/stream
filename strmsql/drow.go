package strmsql

import (
	"database/sql"
	"reflect"

	"github.com/stdiopt/stream/x/strmdrow"
)

func drowHeader(row *sql.Rows) (*strmdrow.Header, error) {
	cols, err := row.Columns()
	if err != nil {
		return nil, err
	}
	typs, err := row.ColumnTypes()
	if err != nil {
		return nil, err
	}

	fields := []strmdrow.Field{}
	for i, c := range cols {
		fields = append(fields, strmdrow.Field{
			Name: c,
			Type: typs[i].ScanType(),
		})
	}
	return strmdrow.NewHeader(fields...), nil
}

func drowScan(hdr *strmdrow.Header, rows *sql.Rows) (strmdrow.Row, error) {
	args := make([]interface{}, hdr.Len())
	vals := make([]reflect.Value, hdr.Len())
	row := strmdrow.NewWithHeader(hdr)
	for i, f := range hdr.Fields() {
		val := reflect.New(f.Type)
		vals[i] = val
		args[i] = val.Interface()
	}
	if err := rows.Scan(args...); err != nil {
		return strmdrow.Row{}, err
	}
	for i, v := range vals {
		row.Values[i] = v.Elem().Interface()
	}
	return row, nil
}
