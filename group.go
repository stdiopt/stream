package stream

import (
	"context"
	"errors"
	"sync"
)

// pGroup based on golang.org/x/sync/errgroup but catches panics
type pGroup struct {
	cancel func()

	wg sync.WaitGroup

	errOnce sync.Once
	err     error
}

func pGroupWithContext(ctx context.Context) (*pGroup, context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	return &pGroup{cancel: cancel}, ctx
}

func (g *pGroup) Wait() error {
	g.wg.Wait()
	if g.cancel != nil {
		g.cancel()
	}
	return g.err
}

func (g *pGroup) Go(f func() error) {
	g.wg.Add(1)

	go func() {
		defer g.wg.Done()

		err := func() (err error) {
			/*defer func() {
				if p := recover(); p != nil {
					err = fmt.Errorf("%v", p)
				}
			}()*/
			err = f()
			// Not sure if this is the ideal place to check for Break
			// the function just returns and the stream compositors would
			// aggregate
			if errors.Is(err, ErrBreak) {
				return nil
			}
			return err
		}()
		if err != nil {
			g.errOnce.Do(func() {
				g.err = err
				if g.cancel != nil {
					g.cancel()
				}
			})
		}
	}()
}
