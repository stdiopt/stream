package main

import (
	"context"
	"fmt"

	"github.com/stdiopt/stream"
)

func main() {
	stringAndReverse := stream.Line(
		stream.Func(stringify),
		stream.Func(reverse),
	)

	l := stream.Line(
		generate(1234, 1345, 1),
		stringAndReverse,
		stream.Func(print),
	)
	if err := stream.Run(l); err != nil {
		fmt.Println("err:", err)
	}
}

// Receive any value and produces strings
func stringify(p stream.Proc) error {
	return p.Consume(func(ctx context.Context, v interface{}) error {
		return p.Send(ctx, fmt.Sprint(v))
	})
}

// Receives strings and reverse
func reverse(p stream.Proc) error {
	return p.Consume(func(ctx context.Context, v interface{}) error {
		s, ok := v.(string)
		if !ok {
			return fmt.Errorf("wrong type: wants string got %T", v)
		}
		runes := []rune(s)
		for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
			runes[i], runes[j] = runes[j], runes[i]
		}
		return p.Send(ctx, string(runes))
	})
}

func print(p stream.Proc) error {
	return p.Consume(func(_ context.Context, v interface{}) error {
		fmt.Println(v)
		return nil
	})
}

// generate numbers
func generate(s, e, n int) stream.Processor {
	return stream.Func(func(p stream.Proc) error {
		for i := s; i < e; i++ {
			if err := p.Send(p.Context(), i); err != nil {
				return err
			}
		}
		return nil
	})
}
