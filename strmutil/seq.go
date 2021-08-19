package strmutil

import "github.com/stdiopt/stream"

func Infinite(start, step int) stream.PipeFunc {
	return stream.Func(func(p stream.Proc) error {
		return p.Consume(func(interface{}) error {
			for i := start; ; i += step {
				if err := p.Send(i); err != nil {
					return err
				}
			}
		})
	})
}

func Seq(start, end, step int) stream.PipeFunc {
	return stream.Func(func(p stream.Proc) error {
		return p.Consume(func(interface{}) error {
			if start > end {
				for i := start; i >= end; i += step {
					if err := p.Send(i); err != nil {
						return err
					}
				}
				return nil
			}
			for i := start; i < end; i += step {
				if err := p.Send(i); err != nil {
					return err
				}
			}
			return nil
		})
	})
}
