// Package streamtest performs tests on stream pipes.
package strmtest

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	strm "github.com/stdiopt/stream"
)

type Capturer struct {
	t    *testing.T
	pipe strm.Pipe
}

func New(t *testing.T, pipe strm.Pipe) *Capturer {
	return &Capturer{t, pipe}
}

func (c *Capturer) Send(v ...interface{}) *Run {
	// Schedule send
	return &Run{
		pipe:  c.pipe,
		t:     c.t,
		sends: v,
	}
}

type wantMode int

const (
	wantModeOrdered = wantMode(iota + 1)
	wantModeAny
)

type Run struct {
	pipe        strm.Pipe
	t           *testing.T
	sends       []interface{}
	senderError error

	wantDuration time.Duration
	wantMode     wantMode
	wantOrdered  []interface{}
	wantAny      map[interface{}]int
	wantErrorRE  string
}

func (r *Run) SenderError(err error) *Run {
	r.senderError = err
	return r
}

func (r *Run) Expect(v ...interface{}) *Run {
	if r.wantMode != 0 {
		r.t.Fatal("expect already used")
	}
	r.wantMode = wantModeOrdered
	r.wantOrdered = append(r.wantOrdered, v...)
	return r
}

func (r *Run) ExpectNonOrdered(v ...interface{}) *Run {
	if r.wantMode != 0 {
		r.t.Fatal("expect already used")
	}
	r.wantMode = wantModeAny
	r.wantAny = map[interface{}]int{}
	for _, vv := range v {
		r.wantAny[vv]++
	}
	return r
}

func (r *Run) ExpectDuration(d time.Duration) *Run {
	r.wantDuration = d
	return r
}

func (r *Run) ExpectMatchError(estr string) *Run {
	r.wantErrorRE = estr
	return r
}

func (r *Run) Run() {
	r.t.Helper()
	var got []interface{}
	var gotAny map[interface{}]int

	in := make(chan interface{})
	go func() {
		defer close(in)
		for _, m := range r.sends {
			in <- m
		}
	}()
	out := make(chan interface{})
	done := make(chan struct{})
	go func() {
		defer close(done)
		for v := range out {
			switch r.wantMode {
			case wantModeOrdered:
				got = append(got, v)
			case wantModeAny:
				if gotAny == nil {
					gotAny = map[interface{}]int{}
				}
				gotAny[v]++
			}
		}
	}()

	capturer := strm.Override{
		ConsumeFunc: func(fn strm.ConsumerFunc) error {
			for v := range in {
				if err := fn(v); err != nil {
					return err
				}
			}
			return nil
		},
		SendFunc: func(v interface{}) error {
			if r.senderError != nil {
				return r.senderError
			}
			out <- v
			return nil
		},
	}
	mark := time.Now()
	func() {
		defer close(out)
		err := r.pipe.Run(context.TODO(), capturer, capturer)
		if !matchError(r.wantErrorRE, err) {
			r.t.Errorf("wrong error\nwant: %v\n got: %v\n", r.wantErrorRE, err)
		}
	}()
	<-done

	dur := time.Since(mark)
	if r.wantDuration != 0 && dur > r.wantDuration {
		r.t.Errorf("exceed time limit\nwant: <%v\n got: >%v\n", r.wantDuration, dur)
	}

	switch r.wantMode {
	case wantModeOrdered:
		if diff := cmp.Diff(r.wantOrdered, got); diff != "" {
			r.t.Error("wrong output\n- want + got\n", diff)
		}
	case wantModeAny:
		if diff := cmp.Diff(r.wantAny, got); diff != "" {
			r.t.Error("wrong output\n- want + got\n", diff)
		}
	}
}

// Try to do in the Send Expect way
// tt := new strmtest.NewController(t)
// tt.With(Print("
// tt.Send(value).Expect(value)
// tt.Send(value).Expect(errors.New("stuff"))
// tt.Send(value,value,value)
// tt.Expect(1, 2, 3)
// tt.Send(1).Expect(1)
// tt.ExpectError("123"),
// tt.Test(Pipe)
// if err := tt.ExpectationsMet(); err != nil {
//
//
