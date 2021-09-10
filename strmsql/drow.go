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

type drowInsert struct {
	table     string
	batchSize int
	autoDDL   bool

	rowLen     int
	ddlChecked bool
}

//func NewDrowInsertOpts() *drowInsert {
//	return &drowInsert{}
//}

func (o *drowInsert) checkDDL(db *sql.DB, row strmdrow.Row) error {
	if o.rowLen == 0 {
		o.rowLen = len(row.Values)
	}
	if !o.autoDDL {
		return nil
	}
	if !o.ddlChecked {
		ddlQry, err := psqlDDLQry(o.table, row)
		if err != nil {
			return err
		}
		if _, err := db.Exec(ddlQry); err != nil {
			return err
		}
		o.ddlChecked = true
	}
	return nil
}

/*func (o *drowInsert) WithAutoDDL(b bool) *drowInsert {
	o.autoDDL = true
	return o
}

func (o *drowInsert) WithBatchSize(n int) *drowInsert {
	o.batchSize = n
	return o
}*/

type drowInsertOpts func(*drowInsert)

func DrowInsertOpts() drowInsertOpts {
	return func(o *drowInsert) {}
}

func (fn drowInsertOpts) WithAutoDDL(b bool) drowInsertOpts {
	return func(o *drowInsert) {
		fn(o)
		o.autoDDL = b
	}
}

func (fn drowInsertOpts) WithBatchSize(n int) drowInsertOpts {
	return func(o *drowInsert) {
		fn(o)
		o.batchSize = n
	}
}

// With options?!
// WithAutoDDL("table")

// InsertBatchStruct
func (d Dialect) InsertBatchDROW(db *sql.DB, tname string, opts ...drowInsertOpts) strm.Pipe {
	return strm.Func(func(p strm.Proc) error {
		o := drowInsert{table: tname, batchSize: 100}
		for _, fn := range opts {
			fn(&o)
		}
		batchParams := []interface{}{}
		qry := fmt.Sprintf(`insert into "%s" values`, tname)
		err := p.Consume(func(row strmdrow.Row) error {
			o.checkDDL(db, row)

			batchParams = append(batchParams, row.Values...)

			if len(batchParams)/o.rowLen >= o.batchSize {
				if _, err := d.ExecInsertQry(db, qry, o.rowLen, batchParams...); err != nil {
					return err
				}
				batchParams = batchParams[:0]
			}
			return nil
			// return p.Send(row)
		})
		if err != nil {
			return err
		}
		if len(batchParams) > 0 {
			_, err = d.ExecInsertQry(db, qry, o.rowLen, batchParams...)
			return err
		}
		return nil
	})
}

func psqlDDLQry(name string, row strmdrow.Row) (string, error) {
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
