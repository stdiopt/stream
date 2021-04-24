package stream

import "context"

// proc implements the Proc interface
type proc struct {
	ctx context.Context
	Consumer
	Sender
}

// MakeProc makes a Proc based on a Consumer and Sender.
func MakeProc(ctx context.Context, c Consumer, s Sender) Proc {
	return proc{ctx, c, s}
}

func (p proc) Consume(fn ConsumerFunc) error {
	if p.Consumer == nil {
		return nil
	}
	return p.Consumer.Consume(fn)
}

func (p proc) Send(ctx context.Context, v interface{}) error {
	if p.Sender == nil {
		return nil
	}
	return p.Sender.Send(ctx, v)
}

func (p proc) Context() context.Context {
	return p.ctx
}
