package stream

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"reflect"
	"runtime"

	"github.com/bwmarrin/snowflake"
)

// Proc is the interface used by ProcFuncs to Consume and send data to the next
// func.
type Proc interface {
	Consume(interface{}) error
	Send(interface{}) error

	Context() context.Context

	MetaSet(string, interface{})
	MetaValue(string) interface{}
	Meta() map[string]interface{}
}

type procFunc func(*proc) error

func (fn procFunc) run(p *proc) error { return fn(p) }

var dbgCtxKey = struct{ s string }{"dbg"}

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

	return procFunc(func(p *proc) error {
		p.dname = dname
		if err := fn(p); err != nil {
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

func F(fn interface{}) Processor {
	fnVal := reflect.ValueOf(fn)
	return Func(func(p Proc) error {
		procVal := reflect.ValueOf(p)
		return p.Consume(func(v interface{}) error {
			var arg reflect.Value
			if v == nil {
				arg = reflect.ValueOf(&v).Elem()
			} else {
				arg = reflect.ValueOf(v)
			}
			ret := fnVal.Call([]reflect.Value{procVal, arg})
			if err, ok := ret[0].Interface().(error); ok && err != nil {
				return err
			}
			return nil
		})
	})
}

type consumerFunc = func(message) error

type consumer interface {
	consume(consumerFunc) error
	cancel()
}

type sender interface {
	send(message) error
}

// proc implements the Proc interface
type proc struct {
	id  snowflake.ID
	ctx context.Context
	consumer
	sender

	dname string
	// State to send to next Processor on message
	meta meta
}

// newProc makes a Proc based on a Consumer and Sender.
func newProc(ctx context.Context, c consumer, s sender) *proc {
	return &proc{
		id:       flake.Generate(),
		ctx:      ctx,
		consumer: c,
		sender:   s,
	}
}

func (p *proc) MetaSet(k string, v interface{}) {
	p.meta.Set(k, v)
}

func (p *proc) MetaValue(k string) interface{} {
	return p.meta.Get(k)
}

func (p *proc) Meta() map[string]interface{} {
	return p.meta.Values()
}

func (p *proc) Consume(fn interface{}) error {
	cfn := makeConsumerFunc(fn)
	err := p.consume(func(m message) error {
		p.meta.SetAll(m.meta)
		// enter consume
		// defer exit consume
		if err := cfn(m); err != nil {
			return err
		}
		return nil
	})
	if err != nil && err != ErrBreak {
		return err
	}
	return nil
}

func (p *proc) Send(v interface{}) error {
	meta := p.meta.Values()

	if w, ok := meta["_debug"].(io.Writer); ok {
		DebugProc(w, p, v)
	}

	err := p.send(message{
		value: v,
		meta:  meta,
	})

	return err
}

func (p *proc) Context() context.Context {
	return p.ctx
}

func (p *proc) Cancel() {
	p.cancel()
}

func (p *proc) cancel() {
	if p.consumer == nil {
		return
	}
	p.consumer.cancel()
}

// the only two important funcs that could be discarded?
func (p *proc) consume(fn func(message) error) error {
	if p.consumer == nil {
		return fn(message{})
	}
	return p.consumer.consume(fn)
}

func (p *proc) send(m message) error {
	if p.sender == nil {
		return nil
	}
	return p.sender.send(m)
}

// Is a bit slower but some what wrappers typed messages.
func makeConsumerFunc(fn interface{}) consumerFunc {
	switch fn := fn.(type) {
	case consumerFunc:
		return fn
	case func(interface{}) error:
		return func(m message) error { return fn(m.value) }
	}

	// Any other param types
	fnVal := reflect.ValueOf(fn)
	fnTyp := fnVal.Type()
	if fnTyp.NumIn() != 1 {
		panic("should have 1 params")
	}
	args := make([]reflect.Value, 1)
	return func(v message) error {
		args[0] = reflect.ValueOf(v.value) // will panic if wrong type passed
		ret := fnVal.Call(args)
		if err, ok := ret[0].Interface().(error); ok && err != nil {
			return err
		}
		return nil
	}
}
