package strmsql

import (
	"database/sql"
	"reflect"
	"time"

	"github.com/stdiopt/stream/drow"
)

func drowHeader(d Dialecter, row *sql.Rows) (*drow.Header, error) {
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
			Type: d.ColumnType(typs[i]),
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

var (
	sqlNullBool   = reflect.TypeOf(sql.NullBool{})
	sqlNullString = reflect.TypeOf(sql.NullString{})
	sqlNullInt64  = reflect.TypeOf(sql.NullInt64{})
	sqlNullTime   = reflect.TypeOf(sql.NullTime{})
	sqlRawBytes   = reflect.TypeOf(sql.RawBytes{})

	boolTyp   = reflect.TypeOf(bool(false))
	int64Typ  = reflect.TypeOf(int64(0))
	timeTyp   = reflect.TypeOf(time.Time{})
	stringTyp = reflect.TypeOf("")
)
