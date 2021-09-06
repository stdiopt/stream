package stream

import (
	"context"
	"log"
	"reflect"
)

var errTyp = reflect.TypeOf((*error)(nil)).Elem()

// proc implements the Proc interface
type proc struct {
	caller callerInfo
	ctx    context.Context
	log    *log.Logger
	consumer
	sender
}

// newProc makes a Proc based on a Consumer and Sender.
func newProc(ctx context.Context, c consumer, s sender) *proc {
	return &proc{
		ctx:      ctx,
		consumer: c,
		sender:   s,
	}
}

// this can be overriden we should put here?
func (p proc) Consume(fn interface{}) error {
	if p.consumer == nil {
		return MakeConsumerFunc(fn)(nil)
	}
	return p.consumer.Consume(fn)
}

func (p proc) Send(v interface{}) error {
	if p.sender == nil {
		return nil
	}

	return p.sender.Send(v)
}

func (p proc) Context() context.Context {
	return p.ctx
}

func (p proc) cancel() {
	if p.consumer == nil {
		return
	}
	p.consumer.cancel()
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

var ConsumerRegistry = map[reflect.Type]func(interface{}) ConsumerFunc{}

// MakeConsumerFunc returns a consumerFunc
// Is a bit slower but some what wrappers typed messages.
func MakeConsumerFunc(fn interface{}) ConsumerFunc {
	switch fn := fn.(type) {
	// optimize for []byte
	case func([]byte) error:
		return func(v interface{}) error {
			b, ok := v.([]byte)
			if !ok {
				return NewTypeMismatchError([]byte(nil), v)
			}
			return fn(b)
		}
	case func(interface{}) error:
		return fn
	}

	// Any other param types
	fnVal := reflect.ValueOf(fn)
	fnTyp := fnVal.Type()
	if cfn, ok := ConsumerRegistry[fnTyp]; ok {
		return cfn(fn)
	}

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
			return NewTypeMismatchError(fnTyp.In(0), args[0].Type())
		}
		ret := fnVal.Call(args)
		if err, ok := ret[0].Interface().(error); ok && err != nil {
			return err
		}
		return nil
	}
}
