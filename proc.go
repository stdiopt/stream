package stream

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"runtime"
)

type procFunc func(Proc) error

func (fn procFunc) Run(p Proc) error { return fn(p) }
func (fn procFunc) internal()        {}

var (
	fnCtxKey  = struct{ s string }{"fn"}
	dbgCtxKey = struct{ s string }{"dbg"}
)

func ContextWithDebug(ctx context.Context, w io.Writer) context.Context {
	return context.WithValue(ctx, dbgCtxKey, w)
}

func Func(fn func(Proc) error) Processor {
	// information to be attached in the dbg process
	// Caller is the function name that calls the NewProc
	var name string
	{
		pc, _, _, _ := runtime.Caller(1)
		fi := runtime.FuncForPC(pc)
		name = fi.Name()
		_, name = filepath.Split(name)
	}

	// The function that calls the builder func
	_, f, l, _ := runtime.Caller(2)
	_, file := filepath.Split(f)

	dname := fmt.Sprintf("%s %s:%d", name, file, l)

	return procFunc(func(p Proc) error {
		pp := override{
			Proc: p,
			SenderFunc: func(ctx context.Context, v interface{}) error {
				meta, ok := MetaFromContext(ctx)
				if !ok {
					meta = NewMeta()
					ctx = ContextWithMeta(ctx, meta)
				}
				meta = meta.WithCaller(name)
				ctx = ContextWithMeta(ctx, meta)

				ctx = context.WithValue(ctx, fnCtxKey, dname)

				if w, ok := ctx.Value(dbgCtxKey).(io.Writer); ok {
					fmt.Fprint(w, DebugCtx(ctx, dname, v))
				}
				return p.Send(ctx, v)
			},
		}

		if err := fn(pp); err != nil {
			return strmError{
				name: name,
				file: file,
				line: l,
				err:  err,
			}
		}
		return nil
	})
}

// proc implements the Proc interface
type proc struct {
	ctx context.Context
	Consumer
	Sender
}

// makeProc makes a Proc based on a Consumer and Sender.
func makeProc(ctx context.Context, c Consumer, s Sender) *proc {
	return &proc{ctx, c, s}
}

func (p proc) Consume(fn interface{}) error {
	if p.Consumer == nil {
		return nil
	}
	return p.Consumer.Consume(fn)
}

func (p proc) Send(ctx context.Context, v interface{}) error {
	if p.Sender == nil {
		return nil
	}
	return p.Sender.Send(ctx, v)
}

func (p proc) Context() context.Context {
	return p.ctx
}

func (p proc) WithContext(ctx context.Context) Proc {
	return proc{
		ctx:      ctx,
		Consumer: p.Consumer,
		Sender:   p.Sender,
	}
}

type override struct {
	Proc
	ConsumerFunc func(ConsumerFunc) error
	SenderFunc   func(context.Context, interface{}) error
}

func (w override) Consume(fn interface{}) error {
	if w.ConsumerFunc != nil {
		return w.ConsumerFunc(MakeConsumerFunc(fn))
	}
	return w.Proc.Consume(fn)
}

func (w override) Send(ctx context.Context, v interface{}) error {
	if w.SenderFunc != nil {
		return w.SenderFunc(ctx, v)
	}
	return w.Proc.Send(ctx, v)
}
