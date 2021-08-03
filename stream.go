// Package stream provides stuff
package stream

import (
	"context"
)

type Chan interface {
	consumer
	sender
	close()
	cancel()
}

func createChan(ctx context.Context, n int) Chan {
	return newProcChan(ctx, n)
}

// Processor to build a processor use stream.Func
type Processor interface {
	run(*proc) error
}

// Line will consume and pass a message sequentually on all ProcFuncs.
func Line(pfns ...Processor) Processor {
	if len(pfns) == 0 {
		panic("no funcs")
	}
	if len(pfns) == 1 {
		return pfns[0]
	}
	return procFunc(func(p *proc) error {
		ctx := p.Context()
		eg, ctx := pGroupWithContext(ctx)
		last := consumer(p) // consumer should be nil
		for _, fn := range pfns[:len(pfns)-1] {
			l, fn := last, fn
			/*if i == len(pfns)-1 {
				// Last one will consume last to P
				eg.Go(func() error {
					// defer l.cancel()
					return fn.run(newProc(ctx, l, p))
				})
				break
			}*/
			ch := createChan(ctx, 0)
			// Consuming from last and sending to channel
			eg.Go(func() error {
				// defer l.cancel() // TODO: Causes issues with workers
				defer ch.close()
				return fn.run(newProc(ctx, l, ch))
			})
			last = ch
		}
		fn := pfns[len(pfns)-1]
		eg.Go(func() error {
			return fn.run(newProc(ctx, last, p))
		})
		return eg.Wait()
	})
}

// Broadcast consumes and passes the consumed message to all pfs ProcFuncs.
func Broadcast(pfns ...Processor) Processor {
	return procFunc(func(p *proc) error {
		eg, ctx := pGroupWithContext(p.Context())
		chs := make([]Chan, len(pfns))
		for i, fn := range pfns {
			ch := createChan(ctx, 0)
			fn := fn
			eg.Go(func() error {
				return fn.run(newProc(ctx, ch, p))
			})
			chs[i] = ch
		}
		eg.Go(func() error {
			defer func() {
				for _, ch := range chs {
					ch.close()
				}
			}()
			return p.consume(func(m Message) error {
				for _, ch := range chs {
					if err := ch.send(m); err != nil {
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
	return procFunc(func(p *proc) error {
		ctx := p.Context()
		eg, ctx := pGroupWithContext(ctx)
		for i := 0; i < n; i++ {
			eg.Go(func() error {
				return pfn.run(newProc(ctx, p, p))
			})
		}
		return eg.Wait()
	})
}

// Buffer will create an extra buffered channel.
func Buffer(n int, pfns ...Processor) Processor {
	pfn := Line(pfns...)
	return procFunc(func(p *proc) error {
		eg, ctx := pGroupWithContext(p.Context())
		ch := createChan(ctx, n)
		eg.Go(func() error {
			defer ch.close()
			return p.consume(ch.send)
		})
		eg.Go(func() error {
			return pfn.run(newProc(ctx, ch, p))
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
	return pfn.run(newProc(ctx, nil, nil))
}
