package strmutil

import "github.com/stdiopt/stream"

func Seq(start, end, step int) stream.Processor {
	return stream.Func(func(p stream.Proc) error {
		ctx := p.Context()
		if start > end {
			for i := start; i >= end; i += step {
				if err := p.Send(ctx, i); err != nil {
					return err
				}
			}
			return nil
		}
		for i := start; i < end; i += step {
			if err := p.Send(ctx, i); err != nil {
				return err
			}
		}
		return nil
	})
}
