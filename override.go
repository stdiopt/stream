package stream

import (
	"context"
)

// Override a proc
type Override struct {
	Proc
	CTX         context.Context
	SendFunc    func(v interface{}) error
	ConsumeFunc func(ConsumerFunc) error
}

func (o Override) Send(v interface{}) error {
	if o.SendFunc != nil {
		return o.SendFunc(v)
	}
	if o.Proc != nil {
		return o.Proc.Send(v)
	}
	return nil
}

func (o Override) Consume(ifn interface{}) error {
	if o.ConsumeFunc != nil {
		fn := MakeConsumerFunc(ifn)
		return o.ConsumeFunc(fn)
	}
	if o.Proc != nil {
		return o.Proc.Consume(ifn)
	}
	return nil
}

func (o Override) cancel() {
	if o.Proc != nil {
		o.Proc.cancel()
	}
}

func (o Override) close() {
	if o.Proc != nil {
		o.Proc.close()
	}
}

func (o Override) Context() context.Context {
	if o.CTX != nil {
		return o.CTX
	}
	if o.Proc != nil {
		return o.Proc.Context()
	}
	return context.TODO()
}
