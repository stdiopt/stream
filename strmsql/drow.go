package strmsql

import (
	"database/sql"
	"reflect"

	"github.com/stdiopt/stream/drow"
)

func drowHeader(row *sql.Rows) (*drow.Header, error) {
	cols, err := row.Columns()
	if err != nil {
		return nil, err
	}
	typs, err := row.ColumnTypes()
	if err != nil {
		return nil, err
	}

	fields := []drow.Field{}
	for i, c := range cols {
		fields = append(fields, drow.Field{
			Name: c,
			Type: typs[i].ScanType(),
		})
	}
	return drow.NewHeader(fields...), nil
}

func drowScan(hdr *drow.Header, rows *sql.Rows) (drow.Row, error) {
	args := make([]interface{}, hdr.Len())
	vals := make([]reflect.Value, hdr.Len())
	row := drow.NewWithHeader(hdr)
	for i, f := range hdr.Fields() {
		val := reflect.New(f.Type)
		vals[i] = val
		args[i] = val.Interface()
	}
	if err := rows.Scan(args...); err != nil {
		return drow.Row{}, err
	}
	for i, v := range vals {
		row.Values[i] = v.Elem().Interface()
	}
	return row, nil
}
