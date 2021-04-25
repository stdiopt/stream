// Package stream provides stuff
package stream

import (
	"context"

	"golang.org/x/sync/errgroup"
)

// ConsumerFunc function type to receive messages.
type ConsumerFunc = func(v interface{}) error

type Consumer interface {
	Consume(ConsumerFunc) error
}

type Sender interface {
	Send(v interface{}) error
}

// Proc is the interface used by ProcFuncs to Consume and send data to the next
// func.
type Proc interface {
	Consumer
	Sender
	Context() context.Context
}

// proc implements the Proc interface
type proc struct {
	ctx context.Context
	Consumer
	Sender
}

// MakeProc makes a Proc based on a Consumer and Sender.
func MakeProc(ctx context.Context, c Consumer, s Sender) Proc {
	return proc{ctx, c, s}
}

func (p proc) Consume(fn func(v interface{}) error) error {
	if p.Consumer == nil {
		return nil
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

type ProcFunc = func(Proc) error

// Line will consume and pass a message sequentually on all ProcFuncs.
func Line(pfns ...ProcFunc) ProcFunc {
	if len(pfns) == 0 {
		panic("no funcs")
	}
	if len(pfns) == 1 {
		return pfns[0]
	}
	return func(p Proc) error {
		ctx := p.Context()
		if ctx == nil {
			ctx = context.Background()
		}
		eg, ctx := errgroup.WithContext(ctx)
		last := p // consumer should be nil
		for i, fn := range pfns {
			l, fn := last, fn // shadow
			if i == len(pfns)-1 {
				// Last one will consume last to P
				eg.Go(func() error {
					return fn(proc{ctx, l, p})
				})
				break
			}
			ch := NewChan(ctx, 0)
			// Consuming from last and sending to channel
			np := proc{ctx, l, ch}
			eg.Go(func() error {
				defer ch.Close()
				return fn(np)
			})
			last = proc{ctx, ch, nil} // don't need a sender here
		}
		return eg.Wait()
	}
}

// Broadcast consumes and passes the consumed message to all pfs ProcFuncs.
func Broadcast(pfns ...ProcFunc) ProcFunc {
	return func(p Proc) error {
		eg, ctx := errgroup.WithContext(p.Context())
		chs := make([]Chan, len(pfns))
		for i, fn := range pfns {
			ch := NewChan(ctx, 0)
			fn := fn
			eg.Go(func() error {
				return fn(proc{ctx, ch, p})
			})
			chs[i] = ch
		}
		eg.Go(func() error {
			defer func() {
				for _, ch := range chs {
					ch.Close()
				}
			}()
			return p.Consume(func(v interface{}) error {
				for _, ch := range chs {
					if err := ch.Send(v); err != nil {
						return err
					}
				}
				return nil
			})
		})
		return eg.Wait()
	}
}

// Workers will start N ProcFuncs consuming and sending on same channels.
func Workers(n int, pfns ...ProcFunc) ProcFunc {
	pfn := Line(pfns...)
	if n <= 0 {
		n = 1
	}
	return func(p Proc) error {
		eg, ctx := errgroup.WithContext(p.Context())
		for i := 0; i < n; i++ {
			eg.Go(func() error {
				return pfn(proc{ctx, p, p})
			})
		}
		return eg.Wait()
	}
}

// Buffer will create an extra buffered channel.
func Buffer(n int, pfns ...ProcFunc) ProcFunc {
	pfn := Line(pfns...)
	return func(p Proc) error {
		eg, ctx := errgroup.WithContext(p.Context())

		ch := NewChan(ctx, n)
		eg.Go(func() error {
			defer ch.Close()
			return p.Consume(ch.Send)
		})
		eg.Go(func() error {
			return pfn(proc{ctx, ch, p})
		})
		return eg.Wait()
	}
}

// Run will run the stream.
func Run(pfns ...ProcFunc) error {
	return RunWithContext(context.Background(), pfns...)
}

// RunWithContext runs the stream with a context.
func RunWithContext(ctx context.Context, pfns ...ProcFunc) error {
	pfn := Line(pfns...)
	return pfn(proc{ctx, nil, nil})
}
