package strmsql

import (
	"bytes"
	"database/sql"
	"fmt"

	"github.com/lib/pq"
	strm "github.com/stdiopt/stream"
	"github.com/stdiopt/stream/strmrefl"
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

func (d Dialect) InsertBatch(db *sql.DB, batchSize int, qry string, params ...interface{}) strm.Pipe {
	return strm.Func(func(p strm.Proc) error {
		batchParams := []interface{}{}
		err := p.Consume(func(v interface{}) error {
			pparams, err := solveParams(v, params...)
			if err != nil {
				return err
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

func (d Dialect) BulkInsert(db *sql.DB, table string, fields []string, params ...interface{}) strm.Pipe {
	// PSQL only for now?!
	return strm.Func(func(p strm.Proc) error {
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
			pparams, err := solveParams(v, params...)
			if err != nil {
				return err
			}
			if _, err := stmt.Exec(pparams...); err != nil {
				return fmt.Errorf("stmt.Exec: %w", err)
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

func (d Dialect) Exec(db *sql.DB, qry string, params ...interface{}) strm.Pipe {
	return strm.S(func(p strm.Sender, v interface{}) error {
		pparams, err := solveParams(v, params...)
		if err != nil {
			return err
		}
		if _, err := db.Exec(qry, pparams...); err != nil {
			return err
		}
		return p.Send(v)
	})
}

type FieldFunc = func(interface{}) (interface{}, error)

func Field(f ...interface{}) FieldFunc {
	return func(v interface{}) (interface{}, error) {
		return strmrefl.FieldOf(v, f...)
	}
}

func solveParams(v interface{}, ps ...interface{}) ([]interface{}, error) {
	var res []interface{}
	for _, p := range ps {
		switch p := p.(type) {
		case FieldFunc:
			v, err := p(v)
			if err != nil {
				return nil, err
			}
			if v, ok := v.([]interface{}); ok {
				res = append(res, v...)
			}
			res = append(res, v)
		case []interface{}:
			res = append(res, p...)
		default:
			res = append(res, p)
		}
	}
	return res, nil
}
