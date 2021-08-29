package stream

import (
	"context"
	"reflect"
)

// Pipe interface for internal use along the pipe line constructor.
// Functions like Func, T, S, will return a Pipe.
type Pipe interface {
	Run(context.Context, Consumer, Sender) error
	newChan(context.Context, int) Proc
}

// pipe constructor returned by Func.
type pipe struct {
	fn func(Proc) error
}

// newChan creates a new channel.
func (p pipe) newChan(ctx context.Context, buffer int) Proc {
	return newPipeChan(ctx, buffer)
}

// run makes a proc and calls the pipe proc func
func (p pipe) Run(ctx context.Context, in Consumer, out Sender) error {
	return p.fn(proc{ctx: ctx, Consumer: in, Sender: out})
}

// Func calls accepts a func with a Proc interface which provides Consume and
// Send methods.
func Func(fn func(Proc) error) Pipe {
	pname := procName()
	return pipe{wrapProcFunc(pname, fn)}
}

// T accepts functions with signature func(T1)(T2, error) where it's called when
// it consumes value and sends the result if no error
func T(fn interface{}) Pipe {
	return pipe{wrapProcFunc(procName(), makeTProcFunc(fn))}
}

// S function that accepts a signature in the form of
// func(Sender, T) error and it's called when a value is received
// Sender can be called multiple times
func S(fn interface{}) Pipe {
	return pipe{wrapProcFunc(procName(), makeSProcFunc(fn))}
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
	default:
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
}

func makeTProcFunc(fn interface{}) procFunc {
	switch fn := fn.(type) {
	case func(interface{}) (interface{}, error):
		return func(p Proc) error {
			return p.Consume(func(v interface{}) error {
				r, err := fn(v)
				if err != nil {
					return err
				}
				return p.Send(r)
			})
		}
	default:
		return func(p Proc) error {
			fnVal := reflect.ValueOf(fn)
			fnTyp := fnVal.Type()
			if fnTyp.NumIn() != 1 {
				panic("func should have 1 param")
			}
			if fnTyp.NumOut() != 2 || fnTyp.Out(1) != errTyp {
				panic("func should have 2 outputs and last must be 'error'")
			}

			args := make([]reflect.Value, 1)
			err := p.Consume(func(v interface{}) error {
				if v == nil {
					args[0] = reflect.New(fnTyp.In(0)).Elem()
				} else {
					args[0] = reflect.ValueOf(v)
				}
				if fnTyp.In(0).Kind() != reflect.Interface && args[0].Type() != fnTyp.In(0) {
					return TypeMismatchError{fnTyp.In(0).String(), args[0].Type().String()}
				}

				ret := fnVal.Call(args)
				if err, ok := ret[1].Interface().(error); ok && err != nil {
					return err
				}
				return p.Send(ret[0].Interface())
			})
			return err
		}
	}
}
