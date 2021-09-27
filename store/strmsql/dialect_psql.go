package strmsql

import (
	"bytes"
	"database/sql"
	"fmt"
	"reflect"
	"time"

	"github.com/stdiopt/stream/drow"
)

var PSQL = pSQLDialect{}

type pSQLDialect struct{}

func (pSQLDialect) QryDDL(name string, row drow.Row) string {
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

func (pSQLDialect) QryBatch(qry string, nparams int, nrows int) string {
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

func (pSQLDialect) ColumnType(ct *sql.ColumnType) reflect.Type {
	t := ct.ScanType()
	switch t {
	case sqlNullBool:
		return reflect.PtrTo(boolTyp)
	case sqlNullString:
		return reflect.PtrTo(stringTyp)
	case sqlNullInt64:
		return reflect.PtrTo(int64Typ)
	case sqlNullTime:
		return reflect.PtrTo(timeTyp)
	case sqlRawBytes:
		return reflect.PtrTo(stringTyp)
	}
	return t
}
