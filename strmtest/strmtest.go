// Package streamtest performs tests on stream pipes.
package strmtest

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	strm "github.com/stdiopt/stream"
)

type wantMode int

const (
	wantModeOrdered = wantMode(iota + 1)
	wantModeAny
)

type Capturer struct {
	t     *testing.T
	pipe  strm.Pipe
	sends []*Send

	wantMode        wantMode
	want            []interface{}
	wantMaxDuration time.Duration
	wantMinDuration time.Duration
	wantErrorRE     string
}

func New(t *testing.T, pipe strm.Pipe) *Capturer {
	return &Capturer{
		t:    t,
		pipe: pipe,
	}
}

func (c *Capturer) Send(v interface{}) *Send {
	s := &Send{
		Capturer: c,
		value:    v,
	}
	c.sends = append(c.sends, s)

	return s
}

func (c *Capturer) ExpectMaxDuration(d time.Duration) *Capturer {
	c.wantMaxDuration = d
	return c
}

func (c *Capturer) ExpectMinDuration(d time.Duration) *Capturer {
	c.wantMinDuration = d
	return c
}

func (c *Capturer) ExpectError(m string) *Capturer {
	c.wantErrorRE = m
	return c
}

func (c *Capturer) ExpectFull(v ...interface{}) *Capturer {
	c.wantMode = wantModeOrdered
	c.want = v
	return c
}

func (c *Capturer) ExpectAnyFull(v ...interface{}) *Capturer {
	c.wantMode = wantModeAny
	c.want = v
	return c
}

func (c *Capturer) Run() {
	c.t.Helper()
	type result struct {
		send *Send
		got  []interface{}
	}

	var mu sync.Mutex
	var cur *result
	var results []*result
	var gotFull []interface{}
	capturer := strm.Override{
		ConsumeFunc: func(fn strm.ConsumerFunc) error {
			for _, v := range c.sends {
				mu.Lock()
				cur = &result{send: v}
				results = append(results, cur)
				mu.Unlock()
				if err := fn(v.value); err != nil {
					return err
				}
			}
			return nil
		},
		SendFunc: func(v interface{}) error {
			mu.Lock()
			defer mu.Unlock()
			if cur.send.senderError != nil {
				return cur.send.senderError
			}
			cur.got = append(cur.got, v)
			gotFull = append(gotFull, v)
			return nil
		},
	}

	mark := time.Now()
	err := c.pipe.Run(context.TODO(), capturer, capturer)
	if !MatchError(c.wantErrorRE, err) {
		c.t.Errorf("wrong error\nwant: %v\n got: %v\n", c.wantErrorRE, err)
	}

	dur := time.Since(mark)
	if c.wantMaxDuration != 0 && dur > c.wantMaxDuration {
		c.t.Errorf("exceed time limit\nwant: <%v\n got: %v\n", c.wantMaxDuration, dur)
	}

	if c.wantMinDuration != 0 && dur < c.wantMinDuration {
		c.t.Errorf("below duration\nwant: >%v\n got: %v\n", c.wantMinDuration, dur)
	}

	// Per send
	for _, r := range results {
		switch r.send.wantMode {
		case wantModeOrdered:
			if diff := cmp.Diff(r.send.want, r.got); diff != "" {
				c.t.Error("wrong output\n- want + got\n", diff)
			}
		case wantModeAny:
			got := map[interface{}]int{}
			for _, v := range r.got {
				got[v]++
			}
			want := map[interface{}]int{}
			for _, v := range r.send.want {
				want[v]++
			}
			if diff := cmp.Diff(want, got); diff != "" {
				c.t.Error("wrong any output\n- want + got\n", diff)
			}
		}
	}
	// Full pipe results
	switch c.wantMode {
	case wantModeOrdered:
		if diff := cmp.Diff(c.want, gotFull); diff != "" {
			c.t.Error("wrong full output\n- want + got\n", diff)
		}
	case wantModeAny:
		got := map[interface{}]int{}
		for _, v := range gotFull {
			got[v]++
		}
		want := map[interface{}]int{}
		for _, v := range c.want {
			want[v]++
		}
		if diff := cmp.Diff(want, got); diff != "" {
			c.t.Error("wrong full any output\n- want + got\n", diff)
		}
	}
}

type Send struct {
	*Capturer
	value       interface{}
	senderError error

	wantMode wantMode
	want     []interface{}
}

func (s *Send) Expect(v ...interface{}) *Capturer {
	if s.wantMode != 0 {
		s.t.Fatal("expect already used")
	}
	s.wantMode = wantModeOrdered
	s.want = v
	return s.Capturer
}

func (s *Send) ExpectAny(v ...interface{}) *Capturer {
	if s.wantMode != 0 {
		s.t.Fatal("expect already used")
	}
	s.wantMode = wantModeAny
	s.want = v
	return s.Capturer
}

func (s *Send) WithSenderError(err error) *Send {
	s.senderError = err
	return s
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
