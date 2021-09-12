package strmsql

import (
	"database/sql"

	strm "github.com/stdiopt/stream"
	"github.com/stdiopt/stream/drow"
)

type Dialecter interface {
	QryDDL(name string, row drow.Row) string
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
	return strm.Func(func(p strm.Proc) error {
		ctx := p.Context()
		b := newBatchInsert(ctx, d, qry, opts...)
		err := p.Consume(func(v interface{}) error {
			if err := b.Add(v); err != nil {
				return err
			}
			return p.Send(v)
		})
		if err != nil {
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
		var hdr *drow.Header
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
