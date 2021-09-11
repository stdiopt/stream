package strmsql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	strm "github.com/stdiopt/stream"
	"github.com/stdiopt/stream/x/strmdrow"
)

func InsertBatch(db *sql.DB, qry string, opts ...batchInsertOpt) strm.Pipe {
	return strm.Func(func(p strm.Proc) (err error) {
		b := newBatchInsert(p.Context(), db, qry)
		for _, fn := range opts {
			fn(&b)
		}
		if err := p.Consume(b.Add); err != nil {
			return err
		}
		return b.flush()
	})
}

type batchInsert struct {
	ctx      context.Context
	dialect  Dialecter
	db       *sql.DB
	batch    []interface{}
	rowLen   int
	rowCount int

	typeChecked bool

	qry       string
	batchSize int
	autoDDL   string
}

func newBatchInsert(ctx context.Context, db *sql.DB, qry string) batchInsert {
	return batchInsert{
		ctx:       ctx,
		db:        db,
		qry:       qry,
		batchSize: 2,
	}
}

func (b *batchInsert) checkInitial(v interface{}) error {
	if b.typeChecked {
		return nil
	}
	b.typeChecked = true
	if b.autoDDL != "" {
		row, ok := v.(strmdrow.Row)
		if !ok {
			return errors.New("auto DDL only supported on drow")
		}
		// Grab dialect from somewhere
		qry := b.dialect.QryDDL(b.autoDDL, row)
		if _, err := b.db.Exec(qry); err != nil {
			return err
		}
	}

	switch v := v.(type) {
	case strmdrow.Row:
		b.rowLen = len(v.Values)
	case []interface{}:
		b.rowLen = len(v)
	default:
		return fmt.Errorf("invalid type %T, only strmdrow.Row or []interface{} supported", v)
	}
	return nil
}

// Maybe support drow only?!
func (b *batchInsert) Add(v interface{}) error {
	if err := b.checkInitial(v); err != nil {
		return err
	}
	switch v := v.(type) {
	case strmdrow.Row:
		b.batch = append(b.batch, v.Values...)
	case []interface{}:
		b.batch = append(b.batch, v...)
	default:
		b.batch = append(b.batch, v)
	}

	b.rowCount++
	if b.rowCount >= b.batchSize {
		return b.flush()
	}
	return nil
}

func (b *batchInsert) flush() error {
	qry := b.dialect.QryBatch(b.qry, b.rowLen, b.rowCount)
	if _, err := b.db.ExecContext(b.ctx, qry, b.batch...); err != nil {
		return err
	}
	b.batch = b.batch[:0]
	b.rowCount = 0
	return nil
}

type batchInsertOpt func(*batchInsert)

func BatchInsertOpt() batchInsertOpt {
	return func(*batchInsert) {}
}

func (fn batchInsertOpt) WithBatchSize(n int) batchInsertOpt {
	return func(o *batchInsert) { fn(o); o.batchSize = n }
}

func (fn batchInsertOpt) WithAutoDDL(t string) batchInsertOpt {
	return func(o *batchInsert) { fn(o); o.autoDDL = t }
}

func (fn batchInsertOpt) WithDialect(d Dialecter) batchInsertOpt {
	return func(o *batchInsert) { fn(o); o.dialect = d }
}
