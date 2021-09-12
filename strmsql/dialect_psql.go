package strmsql

import (
	"bytes"
	"fmt"
	"reflect"
	"time"

	"github.com/stdiopt/stream/drow"
)

type PSQL struct{}

func (PSQL) QryDDL(name string, row drow.Row) string {
	timeTyp := reflect.TypeOf(time.Time{})

	qry := &bytes.Buffer{}
	fmt.Fprintf(qry, "create table if not exists \"%s\" (\n", name)
	for i := 0; i < row.NumField(); i++ {
		h := row.Header(i)
		var sqlType string
		sqlNull := "not null"
		ftyp := h.Type
		if ftyp.Kind() == reflect.Ptr {
			sqlNull = "null"
			ftyp = ftyp.Elem()
		}
		switch ftyp.Kind() {
		case reflect.Int, reflect.Int16, reflect.Int32, reflect.Int64:
			sqlType = "integer"
		case reflect.Uint, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			sqlType = "unsigned integer"
		case reflect.String: // or blob?
			sqlType = "varchar"
		case reflect.Struct:
			switch ftyp {
			case timeTyp:
				sqlType = "datetime"
			}
		}
		fmt.Fprintf(qry, "\t\"%s\" %s %s", h.Name, sqlType, sqlNull)
		if i < row.NumField()-1 {
			qry.WriteRune(',')
		}
		qry.WriteRune('\n')
	}
	qry.WriteString(")\n")
	return qry.String()
}

func (PSQL) QryBatch(qry string, nparams int, nrows int) string {
	qryBuf := &bytes.Buffer{}
	qryBuf.WriteString(qry)
	for i := 0; i < nparams*nrows; i++ {
		if i%nparams == 0 {
			if i == 0 {
				qryBuf.WriteString("(")
			} else {
				qryBuf.WriteString("),\n(")
			}
		} else {
			qryBuf.WriteString(",")
		}
		fmt.Fprintf(qryBuf, "$%d", i+1) // mysql
	}
	qryBuf.WriteString(")")
	return qryBuf.String()
}
