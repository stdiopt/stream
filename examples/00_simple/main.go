package main

import (
	"fmt"

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
		func(p stream.Proc) error {
			return p.Consume(func(v interface{}) error {
				fmt.Println("Consuming:", v)
				return nil
			})
		},
	)

	if err := stream.Run(l); err != nil {
		fmt.Println("err:", err)
	}
}
