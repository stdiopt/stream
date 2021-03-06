package main

import (
	"fmt"

	"github.com/stdiopt/stream"
)

func main() {
	stringAndReverse := stream.Line(
		stringify,
		reverse,
	)

	l := stream.Line(
		generate(1234, 1345, 1),
		stringAndReverse,
		print,
	)
	if err := stream.Run(l); err != nil {
		fmt.Println("err:", err)
	}
}

// Receive any value and produces strings
func stringify(p stream.Proc) error {
	return p.Consume(func(v interface{}) error {
		return p.Send(fmt.Sprint(v))
	})
}

// Receives strings and reverse
func reverse(p stream.Proc) error {
	return p.Consume(func(v interface{}) error {
		s, ok := v.(string)
		if !ok {
			return fmt.Errorf("wrong type: wants string got %T", v)
		}
		runes := []rune(s)
		for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
			runes[i], runes[j] = runes[j], runes[i]
		}
		return p.Send(string(runes))
	})
}

func print(p stream.Proc) error {
	return p.Consume(func(v interface{}) error {
		fmt.Println(v)
		return nil
	})
}

// generate numbers
func generate(s, e, n int) stream.ProcFunc {
	return func(p stream.Proc) error {
		for i := s; i < e; i++ {
			if err := p.Send(i); err != nil {
				return err
			}
		}
		return nil
	}
}
