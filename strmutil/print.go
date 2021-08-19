package strmutil

import (
	"fmt"

	"github.com/stdiopt/stream"
)

func Print(prefix string) stream.PipeFunc {
	return stream.Func(func(p stream.Proc) error {
		return p.Consume(func(v interface{}) error {
			switch v := v.(type) {
			case []byte:
				fmt.Printf("[%s] %v\n", prefix, string(v))
			default:
				fmt.Printf("[%s] %v\n", prefix, v)
			}
			return p.Send(v)
		})
	})
}
