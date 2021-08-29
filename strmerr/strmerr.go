// Package strmerr deal with errors
package strmerr

import (
	"io"
	"log"

	strm "github.com/stdiopt/stream"
)

var Logger = log.New(io.Discard, "", 0)

// Skip on error
func Skip(pps ...strm.Pipe) strm.Pipe {
	return strm.Func(func(p strm.Proc) error {
		o := strm.Override{
			ConsumeFunc: func(fn strm.ConsumerFunc) error {
				return p.Consume(func(v interface{}) error {
					select {
					case <-p.Context().Done():
						return p.Context().Err()
					default:
					}
					if err := fn(v); err != nil {
						Logger.Println("skip error:", err)
					}
					return nil
				})
			},
		}
		// Should we rebuild this on retry?
		return strm.Line(pps...).Run(p.Context(), o, p)
	})
}

func Retry(n int, pps ...strm.Pipe) strm.Pipe {
	return strm.Func(func(p strm.Proc) error {
		o := strm.Override{
			ConsumeFunc: func(fn strm.ConsumerFunc) error {
				return p.Consume(func(v interface{}) error {
					var err error
					for i := 0; i < n; i++ {
						select {
						case <-p.Context().Done():
							return p.Context().Err()
						default:
						}
						err = fn(v)
						if err == nil { // not error
							if i != 0 {
								Logger.Printf("recovered from attempt %d", i+1)
							}
							break
						}
						Logger.Printf("retry#%d error: %v", i+1, err)
					}
					return err
				})
			},
		}
		return strm.Line(pps...).Run(p.Context(), o, p)
	})
}
