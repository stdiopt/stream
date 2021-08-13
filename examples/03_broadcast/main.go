package main

import (
	"fmt"

	"github.com/stdiopt/stream"
)

func main() {
	l := stream.Line(
		generate(0, 2, 1),
		// This will send same message to any broadcast procfunc param
		// so it should print something like:
		//  one 0
		//  two 0
		//  one 1
		//  two 1
		// order is not guaranteed
		stream.Tee(
			printer("one "),
			printer("two "),
		),
	)
	if err := stream.Run(l); err != nil {
		fmt.Println("err:", err)
		return
	}

	l2 := stream.Line(
		generate(0, 10, 1),
		stream.Tee(
			stream.Line(
				// Only sends if number is even
				stream.Func(func(p stream.Proc) error {
					return p.Consume(func(v interface{}) error {
						if n, ok := v.(int); ok && n&1 == 0 {
							return p.Send(v)
						}
						return nil
					})
				}),
				termColor("\033[01;32m"),
			),
			// Only sends if number is odd
			stream.Line(
				stream.Func(func(p stream.Proc) error {
					return p.Consume(func(v interface{}) error {
						if n, ok := v.(int); ok && n&1 == 1 {
							return p.Send(v)
						}
						return nil
					})
				}),
				termColor("\033[01;31m"),
			),
		),
		printer(""),
	)
	if err := stream.Run(l2); err != nil {
		fmt.Println("err:", err)
	}
}

func termColor(c string) stream.ProcFunc {
	return stream.Func(func(p stream.Proc) error {
		return p.Consume(func(v interface{}) error {
			return p.Send(fmt.Sprintf("%s%v\033[0m", c, v))
		})
	})
}

func printer(prefix string) stream.ProcFunc {
	return stream.Func(func(p stream.Proc) error {
		return p.Consume(func(v interface{}) error {
			fmt.Printf("%s%v\n", prefix, v)
			return nil
		})
	})
}

// generate numbers
func generate(s, e, n int) stream.ProcFunc {
	return stream.Func(func(p stream.Proc) error {
		for i := s; i < e; i += n {
			if err := p.Send(i); err != nil {
				return err
			}
		}
		return nil
	})
}
