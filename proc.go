package stream

import (
	"context"
	"fmt"
	"io"
	"reflect"

	"github.com/bwmarrin/snowflake"
)

type P interface {
	Name() string
	Send(interface{}) error

	Context() context.Context

	MetaSet(string, interface{})
	MetaValue(string) interface{}
	Meta() map[string]interface{}
}

// Proc is the interface used by ProcFuncs to Consume and send data to the next
// func.
type Proc interface {
	P
	Consume(interface{}) error
}

// Message is an internal Message type that also passes ctx
type Message struct {
	Value interface{}
	Meta  map[string]interface{}
}

type procFunc func(*proc) error

func (fn procFunc) run(p *proc) (err error) { return fn(p) }

func Func(fn func(Proc) error) Processor {
	pname := procName()
	return procFunc(func(p *proc) error {
		p.name = pname
		if err := fn(p); err != nil {
			return strmError{
				pname: pname,
				err:   err,
			}
		}
		return nil
	})
}

func F(fn interface{}) Processor {
	pname := procName()
	fnVal := reflect.ValueOf(fn)
	return procFunc(func(p *proc) error {
		p.name = pname
		procVal := reflect.ValueOf(P(p))
		err := p.Consume(func(v interface{}) error {
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
		if err != nil {
			return strmError{pname: pname, err: err}
		}
		return nil
	})
}

type consumerFunc = func(Message) error

type consumer interface {
	consume(consumerFunc) error
	cancel()
}

type sender interface {
	send(Message) error
}

// proc implements the Proc interface
type proc struct {
	id  snowflake.ID
	ctx context.Context
	consumer
	sender

	name string
	// State to send to next Processor on message
	meta meta
}

// newProc makes a Proc based on a Consumer and Sender.
func newProc(ctx context.Context, c consumer, s sender) *proc {
	return &proc{
		ctx:      ctx,
		consumer: c,
		sender:   s,
	}
}

func (p *proc) Name() string {
	return p.name
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

// this can be overriden we should put here?
func (p *proc) Consume(fn interface{}) error {
	cfn := makeConsumerFunc(fn)
	err := p.consume(func(m Message) error {
		p.meta.SetAll(m.Meta)
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

	err := p.send(Message{
		Value: v,
		Meta:  meta,
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
func (p *proc) consume(fn func(Message) error) (err error) {
	defer func() {
		if p := recover(); p != nil {
			err = fmt.Errorf("consume panic: %v", p)
		}
	}()
	if p.consumer == nil {
		return fn(Message{})
	}

	return p.consumer.consume(fn)
}

func (p *proc) send(m Message) (err error) {
	defer func() {
		if p := recover(); p != nil {
			err = fmt.Errorf("send panic: %v", p)
		}
	}()
	if p.sender == nil {
		return nil
	}
	return p.sender.send(m)
}

// makeConsumerFunc returns a consumerFunc
// Is a bit slower but some what wrappers typed messages.
func makeConsumerFunc(fn interface{}) consumerFunc {
	switch fn := fn.(type) {
	case consumerFunc:
		return fn
	case func(interface{}) error:
		return func(m Message) error { return fn(m.Value) }
	}

	// Any other param types
	fnVal := reflect.ValueOf(fn)
	fnTyp := fnVal.Type()
	if fnTyp.NumIn() != 1 {
		panic("should have 1 params")
	}
	args := make([]reflect.Value, 1)
	return func(v Message) error {
		args[0] = reflect.ValueOf(v.Value) // will panic if wrong type passed
		ret := fnVal.Call(args)
		if err, ok := ret[0].Interface().(error); ok && err != nil {
			return err
		}
		return nil
	}
}
