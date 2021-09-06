// Package strmagg performs aggregation on streams
package strmagg

import (
	"reflect"

	strm "github.com/stdiopt/stream"
	"github.com/stdiopt/stream/strmrefl"
)

type FieldFunc = func(interface{}) (interface{}, error)

func Field(f ...interface{}) FieldFunc {
	return func(v interface{}) (interface{}, error) {
		return strmrefl.FieldOf(v, f...)
	}
}

type AggEl struct {
	Field string
	Value interface{}
}

type Group struct {
	Field string      `json:"field"`
	Value interface{} `json:"value"`
	Count int         `json:"count"`
	Aggs  []*AggEl    `json:"aggs,omitempty"`
}

type aggOptField struct {
	Name       string
	FieldFunc  FieldFunc
	ReduceFunc func(a, v interface{}) interface{}
}

type aggOptions struct {
	name    string
	groupBy FieldFunc
	aggs    []aggOptField
}

type AggOptFunc func(a *aggOptions)

func Aggregate(opt ...AggOptFunc) strm.Pipe {
	o := aggOptions{}
	for _, fn := range opt {
		fn(&o)
	}

	return strm.Func(func(p strm.Proc) error {
		groupRef := map[interface{}]*Group{}
		group := []*Group{}

		err := p.Consume(func(v interface{}) error {
			key, err := o.groupBy(v)
			if err != nil {
				return err
			}

			g, ok := groupRef[key]
			if !ok {
				g = &Group{
					Field: o.name,
					Value: key,
					Aggs:  make([]*AggEl, len(o.aggs)),
				}
				groupRef[key] = g
				group = append(group, g)
			}
			g.Count++
			for i, a := range o.aggs {
				ar := g.Aggs[i]
				if ar == nil {
					ar = &AggEl{
						Field: a.Name,
						Value: nil,
					}
					g.Aggs[i] = ar
				}
				fi, err := a.FieldFunc(v)
				if err != nil {
					continue
				}
				ar.Value = a.ReduceFunc(ar.Value, fi)
			}

			return nil
		})
		if err != nil {
			return err
		}
		// use drow here
		return p.Send(group)
	})
}

func GroupBy(name string, fn FieldFunc) AggOptFunc {
	return func(a *aggOptions) {
		a.name = name
		a.groupBy = fn
	}
}

func Reduce(name string, field FieldFunc, fni interface{}) AggOptFunc {
	fn := makeReduceFunc(fni)
	return func(a *aggOptions) {
		a.aggs = append(a.aggs, aggOptField{name, field, fn})
	}
}

// func makeReduceFunc[Acc,Val any](fn func(a Acc, v Val) Acc)
func makeReduceFunc(fn interface{}) func(a, v interface{}) interface{} {
	fnVal := reflect.ValueOf(fn)
	typ := fnVal.Type()
	if typ.NumIn() != 2 {
		panic("requires 2 arguments")
	}
	if typ.NumOut() != 1 {
		panic("requires 1 return")
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
