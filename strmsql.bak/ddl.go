package strmsql

import (
	"bytes"
	"database/sql"
	"fmt"
	"reflect"
	"time"

	strm "github.com/stdiopt/stream"
)

func (d Dialect) StructDDL(db *sql.DB, name string) strm.Pipe {
	switch d {
	case PSQL:
		return psqlStructDDL(db, name)
	}
	panic(fmt.Sprintf("%v not supported yet", d))
}

func psqlStructDDL(db *sql.DB, name string) strm.Pipe {
	timeTyp := reflect.TypeOf(time.Time{})
	return strm.Func(func(p strm.Proc) error {
		checked := false
		return p.Consume(func(v interface{}) error {
			if checked {
				return p.Send(v)
			}
			typ := reflect.TypeOf(v)
			qry := &bytes.Buffer{}
			fmt.Fprintf(qry, "create table if not exists \"%s\" (\n", name)
			for i := 0; i < typ.NumField(); i++ {
				h := typ.Field(i)
				var sqlType string
				var sqlNull string = "not null"
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
				if i < typ.NumField()-1 {
					qry.WriteRune(',')
				}
				qry.WriteRune('\n')
			}
			qry.WriteString(")\n")

			p.Printf("Qry is:\n%v\n", qry)
			if _, err := db.Exec(qry.String()); err != nil {
				return err
			}
			checked = true
			return p.Send(v)
		})
	})
}
