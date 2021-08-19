package main

import (
	"fmt"

	"github.com/stdiopt/stream"
)

func main() {
	l := stream.Line(
		producer(10),
		stream.Func(consumer),
	)

	if err := stream.Run(l); err != nil {
		fmt.Println("err:", err)
	}
}

func producer(n int) stream.PipeFunc {
	return stream.Func(func(p stream.Proc) error {
		for i := 0; i < 10; i++ {
			if err := p.Send(i); err != nil {
				return err
			}
		}
		return nil
	})
}

func consumer(p stream.Proc) error {
	return p.Consume(func(v interface{}) error {
		fmt.Println("Consuming:", v)
		return nil
	})
}
