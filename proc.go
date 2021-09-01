package stream

import (
	"context"
	"fmt"
	"log"
	"reflect"
)

var errTyp = reflect.TypeOf((*error)(nil)).Elem()

// proc implements the Proc interface
type proc struct {
	caller callerInfo
	ctx    context.Context
	log    *log.Logger
	Consumer
	Sender
}

// newProc makes a Proc based on a Consumer and Sender.
func newProc(ctx context.Context, c Consumer, s Sender) *proc {
	return &proc{
		ctx:      ctx,
		Consumer: c,
		Sender:   s,
	}
}

// this can be overriden we should put here?
func (p proc) Consume(fn interface{}) error {
	if p.Consumer == nil {
		return MakeConsumerFunc(fn)(nil)
	}
	return p.Consumer.Consume(fn)
}

func (p proc) Send(v interface{}) error {
	if p.Sender == nil {
		return nil
	}

	return p.Sender.Send(v)
}

func (p proc) Context() context.Context {
	return p.ctx
}

func (p proc) cancel() {
	if p.Consumer == nil {
		return
	}
	p.Consumer.cancel()
}

func (p proc) Println(args ...interface{}) {
	if p.log == nil {
		return
	}
	p.log.Println(args...)
}

func (p proc) Printf(format string, args ...interface{}) {
	if p.log == nil {
		return
	}
	p.log.Printf(format, args...)
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
					want: fmt.Sprintf("%T", []byte(nil)),
					got:  fmt.Sprintf("%T", v),
				}
			}
			return fn(b)
		}
	case func(interface{}) error:
		return fn
	}

	// Any other param types
	fnVal := reflect.ValueOf(fn)
	fnTyp := fnVal.Type()
	if fnTyp.NumIn() != 1 {
		panic("func should have 1 params")
	}

	if fnTyp.NumOut() != 1 || fnTyp.Out(0) != errTyp {
		panic("func should have an error return only")
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
