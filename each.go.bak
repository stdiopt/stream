package stream

// Runs this line for each received value.
func Each(pps ...Pipe) Pipe {
	return pipe{func(p Proc) error {
		return p.Consume(func(v interface{}) error {
			consumer := Override{
				ConsumeFunc: func(fn ConsumerFunc) error {
					return fn(v)
				},
			}
			return Line(pps...).Run(p.Context(), consumer, p)
		})
	}}
}
