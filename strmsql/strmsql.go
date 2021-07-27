package strmsql

import (
	"bytes"
	"database/sql"
	"fmt"

	"github.com/stdiopt/stream"
	"github.com/stdiopt/stream/strmutil"
)

type (
	Field string
	Meta  string
)

func execInsertQry(db *sql.DB, qry string, nparams int, batchParams ...interface{}) (sql.Result, error) {
	qryBuf := &bytes.Buffer{}
	qryBuf.WriteString(qry)
	for i := range batchParams {
		if i%nparams == 0 {
			if i == 0 {
				qryBuf.WriteString("(")
			} else {
				qryBuf.WriteString("),\n(")
			}
		} else {
			qryBuf.WriteString(",")
		}
		fmt.Fprintf(qryBuf, "$%d", i+1)
	}
	qryBuf.WriteString(")")

	return db.Exec(qryBuf.String(), batchParams...)
}

func InsertBatch(db *sql.DB, batchSize int, qry string, params ...interface{}) stream.Processor {
	return stream.Func(func(p stream.Proc) error {
		batchParams := []interface{}{}
		err := p.Consume(func(v interface{}) error {
			pparams := make([]interface{}, 0, len(params))
			for _, pp := range params {
				var t interface{} = p
				switch pp := pp.(type) {
				case Meta:
					t = p.MetaValue(string(pp))
				case Field:
					f, err := strmutil.FieldOf(v, string(pp))
					if err != nil {
						return err
					}
					t = f
				}
				pparams = append(pparams, t)
			}
			batchParams = append(batchParams, pparams...)

			if len(batchParams)/len(params) > batchSize {
				if _, err := execInsertQry(db, qry, len(params), batchParams...); err != nil {
					return err
				}
				batchParams = batchParams[:0]
			}
			return p.Send(v)
		})
		if err != nil {
			return err
		}

		_, err = execInsertQry(db, qry, len(params), batchParams...)
		return err
	})
}

func Exec(db *sql.DB, qry string, params ...interface{}) stream.Processor {
	return stream.Func(func(p stream.Proc) error {
		return p.Consume(func(v interface{}) error {
			pparams := make([]interface{}, 0, len(params))
			for _, pp := range params {
				var t interface{} = p
				switch pp := pp.(type) {
				case Meta:
					t = p.MetaValue(string(pp))
				case Field:
					f, err := strmutil.FieldOf(v, string(pp))
					if err != nil {
						return err
					}
					t = f
				}
				pparams = append(pparams, t)
			}
			if _, err := db.Exec(qry, pparams...); err != nil {
				return err
			}
			return p.Send(v)
		})
	})
}
