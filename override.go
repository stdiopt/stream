package stream

/*type override struct {
	Proc
}

func (p override) Send(v interface{}) error {
	log.Println("Sending:", v)
	return p.Proc.Send(v)
}

func (p override) Consume(fn interface{}) error {
	f := makeConsumerFunc(fn)
	return p.Proc.Consume(func(v interface{}) error {
		log.Println("Received:", v)
		return f(v)
	})
}*/
