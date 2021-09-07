package strmsql

import (
	"bytes"
	"database/sql"
	"fmt"

	strm "github.com/stdiopt/stream"
	"github.com/stdiopt/stream/strmrefl"
)

type Dialect int

func (d Dialect) String() string {
	switch d {
	case MySQL:
		return "MySQL"
	case PSQL:
		return "PSQL"
	default:
		return "<undefined>"
	}
}

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

// Change this to receive params directly as []interface{}
// create a Tag thing to receive all "sql" params in order?!
func (d Dialect) InsertBatch(db *sql.DB, batchSize int, qry string, params ...interface{}) strm.Pipe {
	return strm.Func(func(p strm.Proc) error {
		batchParams := []interface{}{}
		err := p.Consume(func(v interface{}) error {
			pparams, err := solveParams(v, params...)
			if err != nil {
				return err
			}
			batchParams = append(batchParams, pparams...)

			if len(batchParams)/len(params) >= batchSize {
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

func Field(f ...interface{}) argFunc {
	return func(v interface{}) (interface{}, error) {
		f := strmrefl.FieldOf(v, f...)
		if f == nil {
			return nil, fmt.Errorf("invalid field: %v", f)
		}
		return f, nil
	}
}

type argFunc = func(interface{}) (interface{}, error)

func solveParams(v interface{}, ps ...interface{}) ([]interface{}, error) {
	var res []interface{}
	for _, p := range ps {
		switch p := p.(type) {
		case argFunc:
			v, err := p(v)
			if err != nil {
				return nil, err
			}
			res = append(res, v)
		default:
			res = append(res, p)
		}
	}
	return res, nil
}
