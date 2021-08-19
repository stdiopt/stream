package stream

type override struct {
	Proc
	sendFunc    func(v interface{}) error
	consumeFunc func(v interface{}) error
}

func (p override) Send(v interface{}) error {
	return p.Proc.Send(v)
}

func (p override) Consume(fn interface{}) error {
	f := MakeConsumerFunc(fn)
	return p.Proc.Consume(func(v interface{}) error {
		return f(v)
	})
}
