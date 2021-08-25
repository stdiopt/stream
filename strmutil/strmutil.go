package strmutil

import (
	"fmt"
	"io"
	"os"
	"time"

	strm "github.com/stdiopt/stream"
)

var stdout = io.Writer(os.Stdout)

func SetOutput(w io.Writer) {
	stdout = w
}

// Limit passes N values and breaks the pipeline.
func Limit(n int) strm.Pipe {
	return strm.Func(func(p strm.Proc) error {
		count := 0
		return p.Consume(func(v interface{}) error {
			if n == 0 {
				return strm.ErrBreak
			}

			if err := p.Send(v); err != nil {
				return err
			}
			count++
			if count >= n {
				return strm.ErrBreak
			}
			return nil
		})
	})
}

// Wait consume, waits d Duration, and sends
func Wait(d time.Duration) strm.Pipe {
	return strm.S(func(s strm.Sender, v interface{}) error {
		time.Sleep(d)
		return s.Send(v)
	})
}

// Value returns a ProcFunc that sends a single value v.
func Value(vs ...interface{}) strm.Pipe {
	return strm.S(func(s strm.Sender, _ interface{}) error {
		for _, v := range vs {
			if err := s.Send(v); err != nil {
				return err
			}
		}
		return nil
	})
}

// Repeat repeats last consumed value n times.
func Repeat(n int) strm.Pipe {
	return strm.S(func(s strm.Sender, v interface{}) error {
		for i := 0; i < n; i++ {
			if err := s.Send(v); err != nil {
				return err
			}
		}
		return nil
	})
}

// Seq generates a sequence.
func Seq(start, end, step int) strm.Pipe {
	return strm.S(func(s strm.Sender, _ interface{}) error {
		if start > end {
			for i := start; i >= end; i += step {
				if err := s.Send(i); err != nil {
					return err
				}
			}
			return nil
		}
		for i := start; i < end; i += step {
			if err := s.Send(i); err != nil {
				return err
			}
		}
		return nil
	})
}

// Pass pass value to another sender, usefull for nested pipelines.
func Pass(pass strm.Sender) strm.Pipe {
	return strm.T(func(v interface{}) (interface{}, error) {
		if err := pass.Send(v); err != nil {
			return nil, err
		}
		return v, nil
	})
}

// Print value with prefix.
func Print(prefix string) strm.Pipe {
	if prefix != "" {
		prefix = fmt.Sprintf("[%s] ", prefix)
	}
	return strm.S(func(p strm.Sender, v interface{}) error {
		switch v := v.(type) {
		case []byte:
			fmt.Fprintf(stdout, "%s%v\n", prefix, string(v))
		default:
			fmt.Fprintf(stdout, "%s%v\n", prefix, v)
		}
		return p.Send(v)
	})
}
