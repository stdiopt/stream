// Package stream provides stuff
package stream

import (
	"context"

	"golang.org/x/sync/errgroup"
)

// public sender
type Sender interface {
	Logger
	Send(interface{}) error
	Context() context.Context
}

type sender interface {
	Send(interface{}) error
	close()
}

type consumer interface {
	Consume(interface{}) error
	cancel()
}

type Chan interface {
	consumer
	sender
}

type Logger interface {
	Println(...interface{})
	Printf(string, ...interface{})
}

// Proc is the interface used by ProcFuncs to Consume and send data to the next
// func.
type Proc interface {
	Context() context.Context
	Logger
	sender
	consumer
}

type ProcFunc = func(Proc) error

// Line will consume and pass a message sequentually on all ProcFuncs.
func Line(pps ...Pipe) Pipe {
	if len(pps) == 0 {
		return Func(func(p Proc) error {
			return p.Consume(p.Send)
		})
	}
	if len(pps) == 1 {
		return pps[0]
	}
	return pipe{fn: func(p Proc) error {
		eg, ctx := errgroup.WithContext(p.Context())
		last := consumer(p) // consumer should be nil
		for _, pp := range pps[:len(pps)-1] {
			ch := pp.newChan(ctx, 0)

			l := last
			pp := pp
			// Consuming from last and sending to channel
			eg.Go(func() error {
				defer l.cancel()
				defer ch.close()
				return pp.Run(ctx, l, ch)
			})
			last = ch
		}
		pp := pps[len(pps)-1]
		eg.Go(func() error {
			defer last.cancel()
			return pp.Run(ctx, last, p)
		})
		return eg.Wait()
	}}
}

// Tee consumes and passes the consumed message to all pfs ProcFuncs.
func Tee(pps ...Pipe) Pipe {
	if len(pps) == 0 {
		return pipe{fn: func(p Proc) error {
			return p.Consume(p.Send)
		}}
	}
	if len(pps) == 1 {
		return pps[0]
	}
	return pipe{fn: func(p Proc) error {
		eg, ctx := errgroup.WithContext(p.Context())
		// iproc
		chs := make([]Chan, len(pps))
		for i, pp := range pps {
			ch := pp.newChan(ctx, 0)
			pp := pp
			eg.Go(func() error {
				return pp.Run(ctx, ch, p)
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
	}}
}

// Workers will start N ProcFuncs consuming and sending on same channels.
func Workers(n int, pps ...Pipe) Pipe {
	pp := Line(pps...)
	if n <= 0 {
		n = 1
	}
	return pipe{fn: func(p Proc) error {
		// Wrap in a channel here? for output?
		// since sender might be something else
		eg, ctx := errgroup.WithContext(p.Context())
		for i := 0; i < n; i++ {
			eg.Go(func() error {
				return pp.Run(ctx, p, p)
			})
		}
		return eg.Wait()
	}}
}

// Buffer will create an extra buffered channel.
func Buffer(n int, pps ...Pipe) Pipe {
	pp := Line(pps...)
	return pipe{fn: func(p Proc) error {
		eg, ctx := errgroup.WithContext(p.Context())
		ch := pp.newChan(ctx, n)
		eg.Go(func() error {
			defer ch.close()
			return p.Consume(ch.Send)
		})
		eg.Go(func() error {
			defer ch.cancel()
			return pp.Run(ctx, ch, p)
		})
		return eg.Wait()
	}}
}

// Run will run the stream.
func Run(pps ...Pipe) error {
	return RunWithContext(context.Background(), pps...)
}

// RunWithContext runs the stream with a context.
func RunWithContext(ctx context.Context, pps ...Pipe) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	return Line(pps...).Run(ctx, nil, nil)
}

// RunFrom utility to run a pipeline from within a pipe func.
// it will consume from a Proc and send to Proc
func RunFrom(p Proc, pps ...Pipe) error {
	ctx, cancel := context.WithCancel(p.Context())
	defer cancel()
	return Line(pps...).Run(ctx, p, p)
}
