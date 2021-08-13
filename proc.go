package stream

import (
	"context"
	"reflect"

	"github.com/bwmarrin/snowflake"
)

type P interface {
	Context() context.Context
	Name() string
	Sender
}

// Proc is the interface used by ProcFuncs to Consume and send data to the next
// func.
type Proc interface {
	P
	Consumer
}

type procFunc func(*proc) error

func (fn procFunc) run(p *proc) (err error) { return fn(p) }

func Func(fn func(Proc) error) ProcFunc {
	pname := procName()
	return func(p Proc) error {
		if p, ok := p.(*proc); ok {
			p.name = pname
		}
		if err := fn(p); err != nil {
			return strmError{
				pname: pname,
				err:   err,
			}
		}
		return nil
	}
}

func F(fn interface{}) ProcFunc {
	pname := procName()
	fnVal := reflect.ValueOf(fn)
	return func(p Proc) error {
		if p, ok := p.(*proc); ok {
			p.name = pname
		}
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
	}
}

type Consumer interface {
	Consume(interface{}) error
	Cancel()
}

type Sender interface {
	Send(interface{}) error
}

// proc implements the Proc interface
type proc struct {
	id  snowflake.ID
	ctx context.Context
	Consumer
	Sender

	name string
}

func (p *proc) String() string {
	return p.name
}

// newProc makes a Proc based on a Consumer and Sender.
func newProc(ctx context.Context, c Consumer, s Sender) *proc {
	return &proc{
		ctx:      ctx,
		Consumer: c,
		Sender:   s,
	}
}

func (p *proc) Name() string {
	return p.name
}

// this can be overriden we should put here?
func (p *proc) Consume(fn interface{}) error {
	if p.Consumer == nil {
		return makeConsumerFunc(fn)(nil)
	}
	return p.Consumer.Consume(fn)
}

func (p *proc) Send(v interface{}) error {
	if p.Sender == nil {
		return nil
	}

	return p.Sender.Send(v)
}

func (p *proc) Context() context.Context {
	return p.ctx
}

func (p *proc) Cancel() {
	if p.Consumer == nil {
		return
	}
	p.Consumer.Cancel()
}

/*
func (p *proc) close() {
	if p.sender == nil {
		return
	}
	p.sender.close()
}
*/
// makeConsumerFunc returns a consumerFunc
// Is a bit slower but some what wrappers typed messages.
func makeConsumerFunc(fn interface{}) consumerFunc {
	switch fn := fn.(type) {
	case consumerFunc:
		return fn
	}

	// Any other param types
	fnVal := reflect.ValueOf(fn)
	fnTyp := fnVal.Type()
	if fnTyp.NumIn() != 1 {
		panic("should have 1 params")
	}
	args := make([]reflect.Value, 1)
	return func(v interface{}) error {
		args[0] = reflect.ValueOf(v) // will panic if wrong type passed
		ret := fnVal.Call(args)
		if err, ok := ret[0].Interface().(error); ok && err != nil {
			return err
		}
		return nil
	}
}
