// Package streamtest performs tests on stream pipes.
package strmtest

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stdiopt/stream"
)

// Case to be used with Test func and test pipelines.
type Case struct {
	PipeFunc    stream.Pipe
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

func WithPipeFunc(p stream.Pipe) CaseOptFunc {
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

type procMock struct {
	ctx         context.Context
	err         chan error
	ConsumeFunc func(v interface{}) error
	SendFunc    func(v interface{}) error
}

func (p *procMock) Context() context.Context {
	return p.ctx
}

func (p *procMock) Cancel() {
}

func (p *procMock) Consume(fn interface{}) error {
	p.ConsumeFunc = stream.MakeConsumerFunc(fn)
	return <-p.err
}

func (p *procMock) Send(v interface{}) error {
	return p.SendFunc(v)
}

func (tt Case) Test(t *testing.T) {
	t.Helper()

	var got []interface{}

	mock := &procMock{
		ctx: context.Background(),
		err: make(chan error),
		SendFunc: func(v interface{}) error {
			if tt.SenderError != nil {
				return tt.SenderError
			}
			got = append(got, v)
			return nil
		},
	}

	go func() {
		defer func() {
			p := recover()
			if !matchPanic(tt.WantInitPanicRE, p) {
				t.Fatalf("wrong init panic\nwant: %v\n got: %v\n", tt.WantInitPanicRE, p)
			}
		}()
		err := tt.PipeFunc(mock)
		if !matchError(tt.WantInitErrorRE, err) {
			t.Fatalf("wrong init error\nwant: %v\n got: %v\n", tt.WantInitErrorRE, err)
		}
	}()

	for i, m := range tt.Sends {
		// Should we check send panic?
		cr = &result{}

		err := mock.ConsumeFunc(m.Value)
		if !matchError(m.WantErrorRE, err) {
			t.Errorf("wrong send#%d error\nwant: %v\n got: %v\n", i, m.WantErrorRE, err)
		}

		if diff := cmp.Diff(m.Want, cr.out); diff != "" {
			t.Errorf("wrong output#%d\n- want + got\n%v", i, diff)
		}
	}
	close(mock.err)

	if closed != 1 {
		t.Fatalf("wrong close count\nwant: %v\n got: %v\n", 1, closed)
	}
}

// Test tests a case.
func Test(t *testing.T, tt Case) {
	tt.Test(t)
}
