package stream

import (
	"context"
	"fmt"
	"log"
	"reflect"
)

// Pipe interface for internal use along the pipe line constructor.
// Functions like Func, T, S, will return a Pipe.
type Pipe interface {
	Run(context.Context, consumer, sender) error
	newChan(context.Context, int) Chan
}

// pipe constructor returned by Func.
type pipe struct {
	caller callerInfo
	fn     func(Proc) error
}

// newChan creates a new channel.
func (p pipe) newChan(ctx context.Context, buffer int) Chan {
	return newPipeChan(ctx, buffer)
}

// run makes a proc and calls the pipe proc func
func (p pipe) Run(ctx context.Context, in consumer, out sender) error {
	prc := proc{
		caller:   p.caller,
		ctx:      ctx,
		consumer: in,
		sender:   out, // this could be created here and a sender returned?!
		log:      log.New(log.Writer(), fmt.Sprintf("[%s] ", p.caller.name), log.Flags()),
	}
	return p.fn(prc)
}

// Func calls accepts a func with a Proc interface which provides Consume and
// Send methods.
func Func(fn func(Proc) error) Pipe {
	// Build proc here?
	caller := procName()
	return pipe{
		caller,
		wrapProcFunc(caller.String(), fn),
	}
}

// S function that accepts a signature in the form of
// func(Sender, T) error and it's called when a value is received
// Sender can be called multiple times
func S(fn interface{}) Pipe {
	caller := procName()
	return pipe{
		caller,
		wrapProcFunc(caller.String(), makeSProcFunc(fn)),
	}
}

var senderTyp = reflect.TypeOf((*Sender)(nil)).Elem()

func makeSProcFunc(fn interface{}) procFunc {
	switch fn := fn.(type) {
	case func(Sender, []byte) error:
		return func(p Proc) error {
			return p.Consume(func(b []byte) error {
				return fn(p, b)
			})
		}
	case func(Sender, interface{}) error:
		return func(p Proc) error {
			return p.Consume(func(v interface{}) error {
				return fn(p, v)
			})
		}
	}
	return func(p Proc) error {
		fnVal := reflect.ValueOf(fn)
		fnTyp := fnVal.Type()

		if fnTyp.NumIn() != 2 {
			panic("func should have 2 params")
		}
		if fnTyp.In(0) != senderTyp {
			panic("first param should be 'strm.Sender'")
		}
		if fnTyp.NumOut() != 1 || fnTyp.Out(0) != errTyp {
			panic("func should have 1 output and must be 'error'")
		}
		args := make([]reflect.Value, 2)
		args[0] = reflect.ValueOf(Sender(p))
		return p.Consume(func(v interface{}) error {
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
	}
}
