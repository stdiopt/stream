package stream

import (
	"reflect"

	"golang.org/x/sync/errgroup"
)

// Experimental routing

// Route will route data through specific pipe, pipes will only be cleared if
// route pipeline goes away.
func Route(fn interface{}, sub func(k string) Pipe) Pipe {
	routefn := makeRouteFunc(fn)
	return pipe{fn: func(p Proc) (err error) {
		eg, ctx := errgroup.WithContext(p.Context())
		senders := map[string]*pipeChan{}

		defer func() {
			e := eg.Wait()
			if err == nil {
				err = e
			}
		}()

		defer func() {
			for _, v := range senders {
				v.close()
			}
		}()

		// When we stop consuming we close the channel
		return p.Consume(func(v interface{}) error {
			key := routefn(v)
			if key == "" {
				return nil
			}
			s, ok := senders[key]
			if !ok {
				ch := newPipeChan(ctx, 0)
				pp := sub(key)
				eg.Go(func() error {
					return pp.Run(ctx, ch, p)
				})
				s = ch
				senders[key] = s
			}
			return s.Send(v)
		})
	}}
}

func makeRouteFunc(fn interface{}) func(interface{}) string {
	switch fn := fn.(type) {
	case func(interface{}) string:
		return fn
	}

	fnVal := reflect.ValueOf(fn)
	fnTyp := fnVal.Type()
	if fnTyp.NumIn() != 1 {
		panic("func should have 1 param")
	}
	if fnTyp.NumOut() != 1 {
		panic("func should have 1 output")
	}
	args := make([]reflect.Value, 1)
	return func(v interface{}) string {
		if v == nil {
			args[0] = reflect.New(fnTyp.In(0)).Elem()
		} else {
			args[0] = reflect.ValueOf(v)
		}
		if fnTyp.In(0).Kind() != reflect.Interface && args[0].Type() != fnTyp.In(0) {
			return ""
		}
		ret := fnVal.Call(args)
		s, ok := ret[0].Interface().(string)
		if !ok {
			return ""
		}
		return s
	}
}
