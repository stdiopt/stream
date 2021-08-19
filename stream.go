// Package stream provides stuff
package stream

import (
	"context"
)

type Chan interface {
	Consumer
	Sender
	close()
}

type PipeFunc = func(Proc) error

// Processor to build a processor use stream.Func
// type Processor interface {
//	run(*proc) error
// }

// Line will consume and pass a message sequentually on all ProcFuncs.
func Line(pfns ...PipeFunc) PipeFunc {
	if len(pfns) == 0 {
		return func(p Proc) error {
			return p.Consume(p.Send)
		}
	}
	if len(pfns) == 1 {
		return pfns[0]
	}
	return func(p Proc) error {
		eg, ctx := pGroupWithContext(p.Context())
		last := Consumer(p) // consumer should be nil
		for _, fn := range pfns[:len(pfns)-1] {
			l, fn := last, fn
			ch := newProcChan(ctx, 0)

			// Consuming from last and sending to channel
			eg.Go(func() error {
				defer l.Cancel()
				defer ch.close()
				return fn(newProc(ctx, l, ch))
			})
			last = ch
		}
		fn := pfns[len(pfns)-1]
		eg.Go(func() error {
			defer last.Cancel()
			return fn(newProc(ctx, last, p))
		})
		return eg.Wait()
	}
}

// Tee consumes and passes the consumed message to all pfs ProcFuncs.
func Tee(pfns ...PipeFunc) PipeFunc {
	return func(p Proc) error {
		eg, ctx := pGroupWithContext(p.Context())
		chs := make([]Chan, len(pfns))
		for i, fn := range pfns {
			ch := newProcChan(ctx, 0)
			fn := fn
			eg.Go(func() error {
				return fn(newProc(ctx, ch, p))
			})
			chs[i] = ch
		}
		eg.Go(func() error {
			defer func() {
				for _, ch := range chs {
					ch.close()
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
func Workers(n int, pfns ...PipeFunc) PipeFunc {
	pfn := Line(pfns...)
	if n <= 0 {
		n = 1
	}
	return func(pp Proc) error {
		p := pp.(*proc)
		ctx := p.Context()
		eg, ctx := pGroupWithContext(ctx)
		for i := 0; i < n; i++ {
			eg.Go(func() error {
				return pfn(newProc(ctx, p, p))
			})
		}
		return eg.Wait()
	}
}

// Buffer will create an extra buffered channel.
func Buffer(n int, pfns ...PipeFunc) PipeFunc {
	pfn := Line(pfns...)
	return func(p Proc) error {
		eg, ctx := pGroupWithContext(p.Context())
		ch := newProcChan(ctx, n)
		eg.Go(func() error {
			defer ch.close()
			np := newProc(ctx, p, ch)
			return np.Consume(np.Send)
		})
		eg.Go(func() error {
			defer ch.Cancel()
			return pfn(newProc(ctx, ch, p))
		})
		return eg.Wait()
	}
}

// Run will run the stream.
func Run(pfns ...PipeFunc) error {
	return RunWithContext(context.Background(), pfns...)
}

// RunWithContext runs the stream with a context.
func RunWithContext(ctx context.Context, pfns ...PipeFunc) error {
	pfn := Line(pfns...)
	return pfn(newProc(ctx, nil, nil))
}
