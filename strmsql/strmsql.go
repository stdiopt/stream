package strmsql

import (
	"database/sql"

	strm "github.com/stdiopt/stream"
	"github.com/stdiopt/stream/x/strmdrow"
)

type Dialecter interface {
	QryDDL(name string, row strmdrow.Row) string
	QryBatch(qry string, nparams, nrows int) string
}

type DB struct {
	db      *sql.DB
	dialect Dialecter
}

func New(db *sql.DB, dialect Dialecter) *DB {
	return &DB{
		db,
		dialect,
	}
}

func (d *DB) BatchInsert(qry string, opts ...batchInsertOpt) strm.Pipe {
	return strm.Func(func(p strm.Proc) (err error) {
		ctx := p.Context()
		b := newBatchInsert(ctx, d, qry, opts...)
		if err := p.Consume(b.Add); err != nil {
			return err
		}
		return b.flush()
	})
}

func (d *DB) Exec(qry string) strm.Pipe {
	return strm.S(func(_ strm.Sender, params []interface{}) error {
		_, err := d.db.Exec(qry, params...)
		return err
	})
}

func (d *DB) Query(qry string) strm.Pipe {
	return strm.Func(func(p strm.Proc) error {
		var hdr *strmdrow.Header
		return p.Consume(func(params []interface{}) error {
			rows, err := d.db.QueryContext(p.Context(), qry, params...)
			if err != nil {
				return err
			}

			if hdr == nil {
				nhdr, err := drowHeader(rows)
				if err != nil {
					return err
				}
				hdr = nhdr
			}
			for rows.Next() {
				row, err := drowScan(hdr, rows)
				if err != nil {
					return err
				}
				if err := p.Send(row); err != nil {
					return err
				}
			}
			return nil
		})
	})
}
