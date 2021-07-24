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

type procFunc func(*proc) error

func (fn procFunc) run(p *proc) error { return fn(p) }
func (fn procFunc) internal()         {}

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

type consumerFunc = func(message) error

type consumer interface {
	consume(consumerFunc) error
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
	// State to be sent to next Processor
	meta       map[string]interface{}
	sendCalled bool
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

func (p *proc) Set(k string, v interface{}) {
	if p.meta == nil {
		p.meta = map[string]interface{}{}
	}
	p.meta[k] = v
}

func (p *proc) Value(k string) interface{} {
	if p.meta == nil {
		return nil
	}
	return p.meta[k]
}

func (p *proc) Meta() map[string]interface{} {
	if p.meta == nil {
		return nil
	}
	m := map[string]interface{}{}
	for k, v := range p.meta {
		m[k] = v
	}
	return m
}

func (p *proc) Consume(fn interface{}) error {
	cfn := makeConsumerFunc(fn)
	return p.consume(func(m message) error {
		// Lock
		if p.meta == nil {
			p.meta = map[string]interface{}{}
		}
		for k, v := range m.meta {
			p.meta[k] = v
		}
		// enter consume
		// defer exit consume
		p.sendCalled = false
		if err := cfn(m); err != nil {
			return err
		}
		if p.sendCalled {
			p.meta = map[string]interface{}{}
		}
		return nil
	})
}

func (p *proc) Send(v interface{}) error {
	p.sendCalled = true

	var meta map[string]interface{}
	if p.meta != nil {
		meta = map[string]interface{}{}
		for k, v := range p.meta {
			meta[k] = v
		}
	}

	if w, ok := meta["_debug"].(io.Writer); ok {
		fmt.Fprint(w, DebugProc(p, v))
	}
	return p.send(message{
		value: v,
		meta:  meta,
	})
}

func (p proc) Context() context.Context {
	return p.ctx
}

// the only two important funcs that could be discarded?
func (p proc) consume(fn func(message) error) error {
	if p.consumer == nil {
		return fn(message{})
	}
	return p.consumer.consume(fn)
}

func (p proc) send(m message) error {
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
