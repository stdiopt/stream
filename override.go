package stream

import "log"

type override struct {
	Proc
	sendFunc    func(v interface{}) error
	consumeFunc func(v interface{}) error
}

func (p override) Send(v interface{}) error {
	log.Println("Sending:", v)
	return p.Proc.Send(v)
}

func (p override) Consume(fn interface{}) error {
	f := makeConsumerFunc(fn)
	return p.Proc.Consume(func(v Message) error {
		log.Println("Received:", v)
		return f(v)
	})
}
