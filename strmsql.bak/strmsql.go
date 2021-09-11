package strmsql

import (
	"bytes"
	"database/sql"
	"fmt"

	strm "github.com/stdiopt/stream"
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
func (d Dialect) InsertBatch(db *sql.DB, batchSize int, qry string) strm.Pipe {
	return strm.Func(func(p strm.Proc) error {
		batchParams := []interface{}{}
		paramLen := 0
		err := p.Consume(func(params []interface{}) error {
			if paramLen == 0 {
				// Use the first message to indicate values len
				paramLen = len(params)
			}
			batchParams = append(batchParams, params...)

			if len(batchParams)/len(params) >= batchSize {
				if _, err := d.ExecInsertQry(db, qry, paramLen, batchParams...); err != nil {
					return err
				}
				batchParams = batchParams[:0]
			}
			return nil
		})
		if err != nil {
			return err
		}
		if len(batchParams) > 0 {
			_, err = d.ExecInsertQry(db, qry, paramLen, batchParams...)
			return err
		}
		return nil
	})
}

func (d Dialect) Exec(db *sql.DB, qry string) strm.Pipe {
	return strm.S(func(_ strm.Sender, params []interface{}) error {
		_, err := db.Exec(qry, params...)
		return err
	})
}
