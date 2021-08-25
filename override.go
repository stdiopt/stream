package stream

import "context"

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

func (m Override) Consume(ifn interface{}) error {
	if m.ConsumeFunc != nil {
		fn := MakeConsumerFunc(ifn)
		return m.ConsumeFunc(fn)
	}
	if m.Proc != nil {
		return m.Proc.Consume(ifn)
	}
	return nil
}

func (m Override) cancel() {
	if m.Proc != nil {
		m.Proc.cancel()
	}
}

func (m Override) close() {
	if m.Proc != nil {
		m.close()
	}
}

func (m Override) Context() context.Context {
	if m.CTX != nil {
		return m.CTX
	}
	if m.Proc != nil {
		m.Proc.Context()
	}
	return context.TODO()
}
