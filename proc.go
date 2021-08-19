package stream

import (
	"context"
	"fmt"
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

func Func(fn func(Proc) error) PipeFunc {
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

func F(fn interface{}) PipeFunc {
	pname := procName()
	return func(p Proc) error {
		if p, ok := p.(*proc); ok {
			p.name = pname
		}
		switch fn := fn.(type) {
		case func(P, []byte) error:
			err := p.Consume(func(b []byte) error {
				return fn(p, b)
			})
			return wrapStrmError(pname, err)
		case func(P, interface{}) error:
			err := p.Consume(func(v interface{}) error {
				return fn(p, v)
			})
			return wrapStrmError(pname, err)
		}

		fnVal := reflect.ValueOf(fn)
		fnTyp := fnVal.Type()
		args := make([]reflect.Value, 2)
		args[0] = reflect.ValueOf(P(p))
		err := p.Consume(func(v interface{}) error {
			if v == nil {
				args[1] = reflect.New(fnTyp.In(1)).Elem()
			} else {
				args[1] = reflect.ValueOf(v)
			}
			if args[1].Type() != fnTyp.In(1) {
				return TypeMismatchError{fnTyp.In(1).String(), args[1].Type().String()}
			}

			ret := fnVal.Call(args)
			if err, ok := ret[0].Interface().(error); ok && err != nil {
				return err
			}
			return nil
		})
		return wrapStrmError(pname, err)
	}
}

type Consumer interface {
	Consume(interface{}) error
	Cancel() // do we need cancel
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
	return fmt.Sprintf("%v", p.name)
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
		return MakeConsumerFunc(fn)(nil)
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

// MakeConsumerFunc returns a consumerFunc
// Is a bit slower but some what wrappers typed messages.
func MakeConsumerFunc(fn interface{}) ConsumerFunc {
	switch fn := fn.(type) {
	// optimize for []byte
	case func([]byte) error:
		return func(v interface{}) error {
			b, ok := v.([]byte)
			if !ok {
				return TypeMismatchError{
					want: fmt.Sprint("%T", []byte(nil)),
					got:  fmt.Sprint("%T", v),
				}
			}
			return fn(b)
		}
	case ConsumerFunc:
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
		if v == nil {
			args[0] = reflect.New(fnTyp.In(0)).Elem()
		} else {
			args[0] = reflect.ValueOf(v)
		}
		if args[0].Type() != fnTyp.In(0) {
			return TypeMismatchError{
				want: fnTyp.In(0).String(),
				got:  args[0].Type().String(),
			}
		}
		ret := fnVal.Call(args)
		if err, ok := ret[0].Interface().(error); ok && err != nil {
			return err
		}
		return nil
	}
}
