// Package strmagg performs aggregation on streams
package strmagg

import (
	"reflect"

	strm "github.com/stdiopt/stream"
	"github.com/stdiopt/stream/drow"
)

type aggEl struct {
	Field string
	Value interface{}
}

type Group struct {
	Field string      `json:"field"`
	Value interface{} `json:"value"`
	Count int         `json:"count"`
	Aggs  []*aggEl    `json:"aggs,omitempty"`
}

type aggOptField struct {
	Name       string
	ReduceFunc func(a, v interface{}) interface{}
}

type aggOptions struct {
	name    string
	groupBy func(interface{}) (interface{}, error)
	aggs    []aggOptField
}

type aggOptFunc func(a *aggOptions)

type aggAll struct{}

func Aggregate(opt ...aggOptFunc) strm.Pipe {
	o := aggOptions{}
	for _, fn := range opt {
		fn(&o)
	}

	return strm.Func(func(p strm.Proc) error {
		groupRef := map[interface{}]*Group{}
		group := []*Group{}

		err := p.Consume(func(v interface{}) error {
			key := interface{}(aggAll{}) // defaults to aggAll
			if o.groupBy != nil {
				gkey, err := o.groupBy(v)
				if err != nil {
					return err
				}
				key = gkey
			}

			g, ok := groupRef[key]
			if !ok {
				g = &Group{
					Field: o.name,
					Value: key,
					Aggs:  make([]*aggEl, len(o.aggs)),
				}
				groupRef[key] = g
				group = append(group, g)
			}
			g.Count++
			for i, a := range o.aggs {
				ar := g.Aggs[i]
				if ar == nil {
					ar = &aggEl{
						Field: a.Name,
						Value: nil,
					}
					g.Aggs[i] = ar
				}
				ar.Value = a.ReduceFunc(ar.Value, v)
			}

			return nil
		})
		if err != nil {
			return err
		}

		for _, g := range group {
			rec := drow.New()
			if o.groupBy != nil {
				rec.SetOrAdd(g.Field, g.Value)
			}
			for _, a := range g.Aggs {
				rec.SetOrAdd(a.Field, a.Value)
			}
			if err := p.Send(rec); err != nil {
				return err
			}
		}
		return nil
	})
}

// GroupBy groups the input in a map using the function
func GroupBy(name string, ifn interface{}) aggOptFunc {
	fn := makeGroupFunc(ifn)
	return func(a *aggOptions) {
		a.name = name
		a.groupBy = fn
	}
}

// Reduce performs a reduce function on the group
func Reduce(name string, ifn interface{}) aggOptFunc {
	fn := makeReduceFunc(ifn)
	return func(a *aggOptions) {
		a.aggs = append(a.aggs, aggOptField{name, fn})
	}
}

// func makeReduceFunc[Acc,Val any](fn func(a Acc, v Val) Acc)
func makeReduceFunc(fn interface{}) func(a, v interface{}) interface{} {
	fnVal := reflect.ValueOf(fn)
	typ := fnVal.Type()
	if typ.NumIn() != 2 {
		panic("reduce func requires 2 inputs")
	}
	if typ.NumOut() != 1 {
		panic("reduce func requires 1 output")
	}
	if typ.In(0) != typ.Out(0) {
		panic("return type should be equal to first argument")
	}
	return func(a, v interface{}) interface{} {
		argA := reflect.ValueOf(a)
		if a == nil {
			argA = reflect.New(fnVal.Type().In(0)).Elem()
		}
		argV := reflect.ValueOf(v)
		if v == nil {
			argV = reflect.New(fnVal.Type().In(1)).Elem()
		}
		args := []reflect.Value{argA, argV}

		ret := fnVal.Call(args)
		return ret[0].Interface()
	}
}

func makeGroupFunc(fn interface{}) func(interface{}) (interface{}, error) {
	fnVal := reflect.ValueOf(fn)
	typ := fnVal.Type()
	if typ.NumIn() != 1 {
		panic("group func requires 1 input")
	}
	if typ.NumOut() < 1 || typ.NumOut() > 2 {
		panic("group func requires at least 1 output")
	}
	return func(v interface{}) (interface{}, error) {
		arg := reflect.ValueOf(v)

		args := []reflect.Value{arg}
		ret := fnVal.Call(args)

		if len(ret) > 1 {
			if err, ok := ret[1].Interface().(error); ok && err != nil {
				return ret[1].Interface(), err
			}
		}
		return ret[0].Interface(), nil
	}
}
