// Package stream provides stuff
package stream

import (
	"context"

	"golang.org/x/sync/errgroup"
)

// RenameToBuilder
type Processor interface {
	Run(Proc) error

	// Prevent implementations from outside
	internal()
}

// ConsumerFunc function type to receive messages.
type ConsumerFunc = func(context.Context, interface{}) error

type Consumer interface {
	Consume(interface{}) error
}

type Sender interface {
	Send(ctx context.Context, v interface{}) error
}

// Proc is the interface used by ProcFuncs to Consume and send data to the next
// func.
type Proc interface {
	Consumer
	Sender
	Context() context.Context
	WithContext(context.Context) Proc
}

// Line will consume and pass a message sequentually on all ProcFuncs.
func Line(pfns ...Processor) Processor {
	if len(pfns) == 0 {
		panic("no funcs")
	}
	if len(pfns) == 1 {
		return pfns[0]
	}
	return procFunc(func(p Proc) error {
		ctx := p.Context()
		eg, ctx := errgroup.WithContext(ctx)
		// eg := errgroup.Group{}
		last := p // consumer should be nil
		for i, fn := range pfns {
			l, fn := last, fn // shadow
			if i == len(pfns)-1 {
				// Last one will consume last to P
				np := makeProc(ctx, l, p)
				eg.Go(func() error {
					return fn.Run(np)
				})
				break
			}
			ch := NewChan(ctx, 0)
			// Consuming from last and sending to channel
			np := makeProc(ctx, l, ch)
			eg.Go(func() error {
				defer ch.Close()
				return fn.Run(np)
			})
			last = makeProc(ctx, ch, nil) // don't need a sender here
		}
		return eg.Wait()
	})
}

// Broadcast consumes and passes the consumed message to all pfs ProcFuncs.
func Broadcast(pfns ...Processor) Processor {
	return procFunc(func(p Proc) error {
		eg, ctx := errgroup.WithContext(p.Context())
		chs := make([]Chan, len(pfns))
		for i, fn := range pfns {
			ch := NewChan(ctx, 0)
			fn := fn
			eg.Go(func() error {
				return fn.Run(makeProc(ctx, ch, p))
			})
			chs[i] = ch
		}
		eg.Go(func() error {
			defer func() {
				for _, ch := range chs {
					ch.Close()
				}
			}()
			return p.Consume(func(ctx context.Context, v interface{}) error {
				for _, ch := range chs {
					if err := ch.Send(ctx, v); err != nil {
						return err
					}
				}
				return nil
			})
		})
		return eg.Wait()
	})
}

// Workers will start N ProcFuncs consuming and sending on same channels.
func Workers(n int, pfns ...Processor) Processor {
	pfn := Line(pfns...)
	if n <= 0 {
		n = 1
	}
	return procFunc(func(p Proc) error {
		ctx := p.Context()
		eg := errgroup.Group{}
		for i := 0; i < n; i++ {
			eg.Go(func() error {
				return pfn.Run(makeProc(ctx, p, p))
			})
		}
		return eg.Wait()
	})
}

// Buffer will create an extra buffered channel.
func Buffer(n int, pfns ...Processor) Processor {
	pfn := Line(pfns...)
	return procFunc(func(p Proc) error {
		eg, ctx := errgroup.WithContext(p.Context())
		ch := NewChan(ctx, n)
		eg.Go(func() error {
			defer ch.Close()
			return p.Consume(ch.Send)
		})
		eg.Go(func() error {
			return pfn.Run(makeProc(ctx, ch, p))
		})
		return eg.Wait()
	})
}

// Run will run the stream.
func Run(pfns ...Processor) error {
	return RunWithContext(context.Background(), pfns...)
}

// RunWithContext runs the stream with a context.
func RunWithContext(ctx context.Context, pfns ...Processor) error {
	pfn := Line(pfns...)
	return pfn.Run(makeProc(ctx, nil, nil))
}
