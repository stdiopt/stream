package strmsql

import (
	"bytes"
	"database/sql"
	"fmt"

	"github.com/lib/pq"
	"github.com/stdiopt/stream"
	"github.com/stdiopt/stream/strmrefl"
)

type (
	Field string
	Meta  string
)

type Dialect int

const (
	MySQL = Dialect(iota + 1)
	PSQL
)

func (d Dialect) ExecInsertQry(db *sql.DB, qry string, nparams int, batchParams ...interface{}) (sql.Result, error) {
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
		switch d {
		case MySQL:
			fmt.Fprintf(qryBuf, "?") // mysql
		case PSQL:
			fmt.Fprintf(qryBuf, "$%d", i+1) // postgres
		}
	}
	qryBuf.WriteString(")")

	return db.Exec(qryBuf.String(), batchParams...)
}

func (d Dialect) InsertBatch(db *sql.DB, batchSize int, qry string, params ...interface{}) stream.PipeFunc {
	return stream.Func(func(p stream.Proc) error {
		batchParams := []interface{}{}
		err := p.Consume(func(v interface{}) error {
			pparams := make([]interface{}, 0, len(params))
			for _, pp := range params {
				t := interface{}(pp)
				switch pp := pp.(type) {
				case func(interface{}) (interface{}, error):
					f, err := pp(v)
					if err != nil {
						return err
					}
					t = f
				case Field:
					f, err := strmrefl.FieldOf(v, string(pp))
					if err != nil {
						return err
					}
					t = f
				}
				pparams = append(pparams, t)
			}
			batchParams = append(batchParams, pparams...)

			if len(batchParams)/len(params) > batchSize {
				if _, err := d.ExecInsertQry(db, qry, len(params), batchParams...); err != nil {
					return err
				}
				batchParams = batchParams[:0]
			}
			return p.Send(v)
		})
		if err != nil {
			return err
		}
		if len(batchParams) > 0 {
			_, err = d.ExecInsertQry(db, qry, len(params), batchParams...)
			return err
		}
		return nil
	})
}

func (d Dialect) BulkInsert(db *sql.DB, table string, fields []string, params ...interface{}) stream.PipeFunc {
	// PSQL only for now?!
	return stream.Func(func(p stream.Proc) error {
		tx, err := db.Begin()
		if err != nil {
			return err
		}
		stmt, err := tx.Prepare(pq.CopyIn(table, fields...))
		if err != nil {
			return err
		}
		defer tx.Commit()
		err = p.Consume(func(v interface{}) error {
			pparams := make([]interface{}, 0, len(params))
			for _, pp := range params {
				t := interface{}(pp)
				switch pp := pp.(type) {
				case Field:
					f, err := strmrefl.FieldOf(v, string(pp))
					if err != nil {
						return err
					}
					t = f
				}
				pparams = append(pparams, t)
			}
			if _, err := stmt.Exec(pparams...); err != nil {
				return err
			}
			return p.Send(v)
		})
		if err != nil {
			tx.Rollback()
			return err
		}
		_, err = stmt.Exec()
		return err
	})
}

func (d Dialect) Exec(db *sql.DB, qry string, params ...interface{}) stream.PipeFunc {
	return stream.F(func(p stream.P, v interface{}) error {
		pparams := make([]interface{}, 0, len(params))
		for _, pp := range params {
			var t interface{} = p
			switch pp := pp.(type) {
			case Field:
				f, err := strmrefl.FieldOf(v, string(pp))
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
}
