package strmsql

import (
	"context"
	"errors"
	"fmt"

	"github.com/stdiopt/stream/x/strmdrow"
)

type batchInsert struct {
	ctx    context.Context
	strmdb *DB
	qry    string

	batch    []interface{}
	rowLen   int
	rowCount int

	typeChecked bool

	batchSize int
	autoDDL   string
}

func newBatchInsert(ctx context.Context, db *DB, qry string, opts ...batchInsertOpt) batchInsert {
	b := batchInsert{
		ctx:       ctx,
		strmdb:    db,
		qry:       qry,
		batchSize: 2,
	}
	for _, fn := range opts {
		fn(&b)
	}
	return b
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
		qry := b.strmdb.dialect.QryDDL(b.autoDDL, row)
		if _, err := b.strmdb.db.Exec(qry); err != nil {
			return fmt.Errorf("query error: %w on %v", err, qry)
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
	if b.rowCount == 0 {
		return nil
	}
	qry := b.strmdb.dialect.QryBatch(b.qry, b.rowLen, b.rowCount)
	if _, err := b.strmdb.db.ExecContext(b.ctx, qry, b.batch...); err != nil {
		return fmt.Errorf("query error %w on query %s", err, qry)
	}
	b.batch = b.batch[:0]
	b.rowCount = 0
	return nil
}

type batchInsertOpt func(*batchInsert)

var BatchOptions = batchInsertOpt(func(*batchInsert) {})

func (fn batchInsertOpt) WithSize(n int) batchInsertOpt {
	return func(o *batchInsert) { fn(o); o.batchSize = n }
}

func (fn batchInsertOpt) WithAutoDDL(t string) batchInsertOpt {
	return func(o *batchInsert) { fn(o); o.autoDDL = t }
}
