// Package streamtest performs tests on stream pipes.
package strmtest

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stdiopt/stream"
)

// Case to be used with Test func and test pipelines.
type Case struct {
	PipeFunc    stream.PipeFunc
	SenderError error
	CloserError error

	Sends            []Send
	WantInitErrorRE  string
	WantCloseErrorRE string

	WantInitPanicRE  string
	WantClosePanicRE string
}

// Send per message basis testing struct.
type Send struct {
	Value       interface{}
	Want        []interface{}
	WantErrorRE string
}

type CaseOptFunc = func(*Case)

func NewCase(opt ...CaseOptFunc) Case {
	c := Case{}
	for _, fn := range opt {
		fn(&c)
	}
	return c
}

func WithPipeFunc(p stream.PipeFunc) CaseOptFunc {
	return func(c *Case) { c.PipeFunc = p }
}

func WithSenderError(err error) CaseOptFunc {
	return func(c *Case) { c.SenderError = err }
}

func WithCloserError(err error) CaseOptFunc {
	return func(c *Case) { c.CloserError = err }
}

func WithSend(send ...Send) CaseOptFunc {
	return func(c *Case) {
		c.Sends = append(c.Sends, send...)
	}
}

func WithWantInitError(match string) CaseOptFunc {
	return func(c *Case) { c.WantInitErrorRE = match }
}

func WithWantCloseError(match string) CaseOptFunc {
	return func(c *Case) { c.WantCloseErrorRE = match }
}

func WithWantInitPanic(match string) CaseOptFunc {
	return func(c *Case) { c.WantInitPanicRE = match }
}

func WithWantClosePanic(match string) CaseOptFunc {
	return func(c *Case) { c.WantClosePanicRE = match }
}

func (tt Case) Test(t *testing.T) {
	t.Helper()

	type result struct {
		out []interface{}
	}
	cr := &result{}
	var closed int

	capture := func(p stream.Proc) error {
		err := p.Consume(func(v interface{}) error {
			if tt.SenderError != nil {
				return tt.SenderError
			}
			cr.out = append(cr.out, v)
			return nil
		})
		closed++
		return err
	}

	pfn := func() stream.PipeFunc {
		defer func() {
			p := recover()
			if !matchPanic(tt.WantInitPanicRE, p) {
				t.Fatalf("wrong init panic\nwant: %v\n got: %v\n", tt.WantInitPanicRE, p)
			}
		}()
		pfn := tt.PipeFunc(capture)
		if !matchError(tt.WantInitErrorRE, err) {
			t.Fatalf("wrong init error\nwant: %v\n got: %v\n", tt.WantInitErrorRE, err)
		}
		return s
	}()
	if pfn == nil {
		return
	}

	results := []*result{}
	for i, m := range tt.Sends {
		// Should we check send panic?
		cr = &result{}
		results = append(results, cr)

		err := s.Send(m.Value)
		if !matchError(m.WantErrorRE, err) {
			t.Errorf("wrong send#%d error\nwant: %v\n got: %v\n", i, m.WantErrorRE, err)
		}
	}

	func() {
		defer func() {
			p := recover()
			if !matchPanic(tt.WantClosePanicRE, p) {
				t.Fatalf("wrong close panic\nwant: %v\n got: %v\n", tt.WantClosePanicRE, p)
			}
		}()
		err := s.Close()
		if !matchError(tt.WantCloseErrorRE, err) {
			t.Errorf("wrong close error\nwant: %v\n got: %v\n", tt.WantCloseErrorRE, err)
		}
	}()

	for i, m := range tt.Sends {
		got := results[i].out
		if diff := cmp.Diff(m.Want, got); diff != "" {
			t.Errorf("wrong output#%d\n- want + got\n%v", i, diff)
		}

	}

	if closed != 1 {
		t.Fatalf("wrong close count\nwant: %v\n got: %v\n", 1, closed)
	}
}

// Test tests a case.
func Test(t *testing.T, tt Case) {
	tt.Test(t)
}
