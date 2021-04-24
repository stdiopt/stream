# stream

An experiment around building streamable pipelines in go

the idea it self is not about maximum performance but rather an simplier way to
abstract channels, workers, context cancellation and dynamic data mostly for
ETL jobs. It relies heavily on `interface{}` to pass data around and has some
heavy reflection usage on `github.com/stdiopt/stream/strmutil`.

it's be possible to build procs around serializing CSVs, selecting DB, producing
Parquet, crawling url's, consuming API's, etc...

## ProcFunc

a ProcFunc is the function signature to chain transforms in a stream
ProcFuncs should block until it doesn't have more messages to send, exiting
early closes the internal channel and will stop further processors

- Each procFunc runs in a go routine
- ProcFuncs should block until they don't have any more data to send
- Consume or Send will be cancelled if context is done

```go
type ProcFunc = func(p stream.Proc) error
```

the Proc interface:

```go
type Proc interface {
	Context() context.Context
	Consume(func(interface{}) error) error
	Send(interface{}) error
}
```

Consume is a blocking method that consumes messages from the previous processor
and calls the func passed as argument for each message

Send will send a value to the next processor

an usual ProcFunc looks like:

```go
func(p stream.Proc) error {
	// Initialize things, open files, db connections, whatever fits the
	// processor

	// Since consume blocks we can call it in the end to hold the function
	// until we don't have more to consume if the underlying context is
	// cancelled due to a previous error or timeout the Consume will cease and
	// return
	return p.Consume(func(v interface{}) error {
		// do something with consumed value
		return p.Send(transformed)
	})
}
```

## Usage

```go
package main

import (
	"fmt"

	"github.com/stdiopt/stream"
)

func main() {
	l := stream.Line(
		produce,
		consume,
	)
	if err := stream.Run(l); err != nil {
		fmt.Println("err:", err)
	}
}
func produce(p stream.Proc) error {
	for i := 0; i < 10; i++ {
		if err := p.Send(i); err != nil {
			return err
		}
	}
	return nil
}
func consume(p stream.Proc) error {
	return p.Consume(func(v interface{}) error {
		fmt.Println("Consuming:", v)
		return nil
	})
}
```

[examples](./examples)
