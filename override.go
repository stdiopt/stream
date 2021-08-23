package stream

type Override struct {
	Proc
	SendFunc    func(v interface{}) error
	ConsumeFunc func(func(v interface{}) error) error
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

func (m Override) Consume(fn interface{}) error {
	if m.ConsumeFunc != nil {
		fn := MakeConsumerFunc(fn)
		return m.ConsumeFunc(fn)
	}
	if m.Proc != nil {
		return m.Proc.Consume(fn)
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
