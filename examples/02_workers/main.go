package main

import (
	"fmt"
	"time"

	"github.com/stdiopt/stream"
)

func main() {
	l := stream.Line(
		func(p stream.Proc) error {
			for i := 0; i < 10; i++ {
				if err := p.Send(i); err != nil {
					return err
				}
			}
			return nil
		},
		// if ran without workers it would take at least 10 seconds for the input above
		stream.Workers(10, func(p stream.Proc) error {
			return p.Consume(func(v interface{}) error {
				n := v.(int)
				time.Sleep(time.Second) // Simulate work
				return p.Send(n * n)
			})
		}),
		// buffer creates an underlying channel with 100 capacity
		stream.Buffer(100, func(p stream.Proc) error {
			return p.Consume(func(v interface{}) error {
				fmt.Println(v)
				return nil
			})
		}),
	)

	if err := stream.Run(l); err != nil {
		fmt.Println("err:", err)
	}
}
