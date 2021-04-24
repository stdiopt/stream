// Package strmagg performs aggregation on streams
package strmagg

import (
	"context"
	"reflect"

	"github.com/stdiopt/stream"
	streamu "github.com/stdiopt/stream/strmutil"
)

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
	Name  string
	Field string
	Func  func(a, v interface{}) interface{}
}

type aggOptions struct {
	name    string
	groupBy func(interface{}) interface{}
	aggs    []aggOptField
}

type AggOptFunc func(a *aggOptions)

func Aggregate(opt ...AggOptFunc) stream.ProcFunc {
	o := aggOptions{}
	for _, fn := range opt {
		fn(&o)
	}

	return func(p stream.Proc) error {
		ctx := p.Context()
		groupRef := map[interface{}]*Group{}
		group := []*Group{}

		err := p.Consume(func(ctx context.Context, v interface{}) error {
			if v == streamu.End {
				return p.Send(ctx, group)
			}
			key := o.groupBy(v)

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
				fi, err := streamu.FieldOf(v, a.Field)
				if err != nil {
					continue
				}
				ar.Value = a.Func(ar.Value, fi)
			}

			return nil
		})
		if err != nil {
			return err
		}
		return p.Send(ctx, group)
	}
}

func GroupBy(name string, fn func(interface{}) interface{}) AggOptFunc {
	return func(a *aggOptions) {
		a.name = name
		a.groupBy = fn
	}
}

func Reduce(name, field string, fni interface{}) AggOptFunc {
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
