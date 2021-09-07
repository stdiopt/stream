package strmsql

import (
	"bytes"
	"database/sql"
	"fmt"
	"reflect"
	"time"

	strm "github.com/stdiopt/stream"
	"github.com/stdiopt/stream/x/strmdrow"
)

func psqlDDLDrow(name string, row strmdrow.Row) (string, error) {
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
	return qry.String(), nil
}

// InsertBatchStruct
func (d Dialect) InsertBatchDROW(db *sql.DB, batchSize int, tname string) strm.Pipe {
	return strm.Func(func(p strm.Proc) error {
		ddlCreated := false
		rowLen := 0
		batchParams := []interface{}{}
		qry := fmt.Sprintf(`insert into "%s" values`, tname)
		err := p.Consume(func(row strmdrow.Row) error {
			if !ddlCreated {
				ddlQry, err := psqlDDLDrow(tname, row)
				if err != nil {
					return err
				}
				p.Printf("DDL:\n%v\n", ddlQry)

				if _, err := db.Exec(ddlQry); err != nil {
					return err
				}

				rowLen = len(row.Values)
				ddlCreated = true
			}

			batchParams = append(batchParams, row.Values...)

			if len(batchParams)/rowLen >= batchSize {
				if _, err := d.ExecInsertQry(db, qry, rowLen, batchParams...); err != nil {
					return err
				}
				batchParams = batchParams[:0]
			}
			return p.Send(row)
		})
		if err != nil {
			return err
		}
		if len(batchParams) > 0 {
			_, err = d.ExecInsertQry(db, qry, rowLen, batchParams...)
			return err
		}
		return nil
	})
}
