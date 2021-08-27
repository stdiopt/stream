package stream

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"
)

func Test_newProcChan(t *testing.T) {
	type args struct {
		ctx    context.Context
		buffer int
	}
	tests := []struct {
		name string
		args args
		want *procChan
	}{
		{
			name: "returns a new procChan",
			args: args{context.TODO(), 0},
			want: &procChan{
				ctx:  context.TODO(),
				ch:   make(chan interface{}),
				done: make(chan struct{}),
			},
		},
		{
			name: "returns a new buffered chan",
			args: args{context.TODO(), 2},
			want: &procChan{
				ctx:  context.TODO(),
				ch:   make(chan interface{}, 2),
				done: make(chan struct{}),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := newProcChan(tt.args.ctx, tt.args.buffer)

			if !reflect.DeepEqual(got.ctx, tt.want.ctx) {
				t.Errorf("newProcChan().ctx = %v, want %v", got.ctx, tt.want.ctx)
			}

			if (got.done == nil) != (tt.want.done == nil) {
				t.Errorf("(newProcChan().done == nil) = %v, want %v", got.done == nil, tt.want.done == nil)
			}
			if cap(got.done) != cap(tt.want.done) {
				t.Errorf("cap(newProcChan().done) = %v, want %v", cap(got.done), cap(tt.want.done))
			}

			if (got.ch == nil) != (tt.want.ch == nil) {
				t.Errorf("(newProcChan().ch == nil) = %v, want %v", got.ch == nil, tt.want.ch == nil)
			}

			if cap(got.ch) != cap(tt.want.ch) {
				t.Errorf("cap(newProcChan().ch) = %v, want %v", cap(got.ch), cap(tt.want.ch))
			}
		})
	}
}

func Test_procChan_Context(t *testing.T) {
	type fields struct {
		ctx  context.Context
		ch   chan interface{}
		done chan struct{}
	}
	tests := []struct {
		name   string
		fields fields
		want   context.Context
	}{
		{
			name:   "return current context",
			fields: fields{ctx: context.TODO()},
			want:   context.TODO(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := procChan{
				ctx:  tt.fields.ctx,
				ch:   tt.fields.ch,
				done: tt.fields.done,
			}
			if got := c.Context(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("procChan.Context() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_procChan_Send(t *testing.T) {
	type fields struct {
		ctx  context.Context
		ch   chan interface{}
		done chan struct{}
	}
	type args struct {
		v interface{}
	}
	tests := []struct {
		name         string
		fields       fields
		args         args
		wantConsumed interface{}
		wantErr      string
	}{
		{
			name: "sends value through channel",
			fields: fields{
				ctx:  context.TODO(),
				ch:   make(chan interface{}),
				done: make(chan struct{}),
			},
			args:         args{1},
			wantConsumed: 1,
		},
		{
			name: "returns error on context canceled",
			fields: fields{
				ctx:  helperCanceledContext(),
				ch:   make(chan interface{}),
				done: make(chan struct{}),
			},
			args:    args{1},
			wantErr: "context canceled",
		},
		{
			name: "returns break error on context done close",
			fields: fields{
				ctx: context.TODO(),
				ch:  make(chan interface{}),
				done: func() chan struct{} {
					ch := make(chan struct{})
					close(ch)
					return ch
				}(),
			},
			args:    args{1},
			wantErr: "break",
		},
		{
			name: "returns error on context canceled while waiting for send",
			fields: fields{
				ctx:  helperTimeoutContext(time.Second),
				ch:   make(chan interface{}),
				done: make(chan struct{}),
			},
			args:    args{1},
			wantErr: "context.*",
		},
		{
			name: "returns break error on while waiting for send",
			fields: fields{
				ctx: context.TODO(),
				ch:  make(chan interface{}),
				done: func() chan struct{} {
					ch := make(chan struct{})
					go func() {
						time.Sleep(2 * time.Second)
						close(ch)
					}()
					return ch
				}(),
			},
			args:    args{1},
			wantErr: "break",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := procChan{
				ctx:  tt.fields.ctx,
				ch:   tt.fields.ch,
				done: tt.fields.done,
			}

			var consumed interface{}
			done := make(chan struct{})
			if tt.wantConsumed != nil {
				go func() {
					defer close(done)
					consumed = <-c.ch
				}()
			} else {
				close(done)
			}

			if err := c.Send(tt.args.v); !matchError(tt.wantErr, err) {
				t.Errorf("procChan.Send() error = %v, wantErr %v", err, tt.wantErr)
			}
			c.close()
			<-done
			if !reflect.DeepEqual(consumed, tt.wantConsumed) {
				t.Errorf("procChan.Send() consumed = %v, want %v", consumed, tt.wantConsumed)
			}
		})
	}
}

func Test_procChan_Consume(t *testing.T) {
	type fields struct {
		ctx  context.Context
		ch   chan interface{}
		done chan struct{}
	}
	type args struct {
		ifn interface{}
	}
	tests := []struct {
		name         string
		fields       fields
		argsfn       func(*interface{}) args
		closeSend    func(*procChan)
		send         interface{}
		wantConsumed interface{}
		wantErr      string
		wantPanic    string
	}{
		{
			name: "starts the consumer",
			fields: fields{
				ctx:  context.TODO(),
				ch:   make(chan interface{}),
				done: make(chan struct{}),
			},
			argsfn: func(out *interface{}) args {
				fn := func(v interface{}) error {
					*out = v
					return nil
				}
				return args{fn}
			},
			send:         1,
			wantConsumed: 1,
		},
		{
			name: "returns nil if returning break error",
			fields: fields{
				ctx:  context.TODO(),
				ch:   make(chan interface{}),
				done: make(chan struct{}),
			},
			argsfn: func(out *interface{}) args {
				fn := func(v interface{}) error {
					*out = v
					return ErrBreak
				}
				return args{fn}
			},
			send:         1,
			wantConsumed: 1,
		},
		{
			name: "returns error on consumer error",
			fields: fields{
				ctx:  context.TODO(),
				ch:   make(chan interface{}),
				done: make(chan struct{}),
			},
			argsfn: func(_ *interface{}) args {
				fn := func(v interface{}) error {
					return errors.New("consumer func error")
				}
				return args{fn}
			},
			send:    1,
			wantErr: "consumer func error",
		},
		{
			name: "panics on invalid function signature",
			fields: fields{
				ctx:  context.TODO(),
				ch:   make(chan interface{}),
				done: make(chan struct{}),
			},
			argsfn: func(*interface{}) args {
				fn := func(a, b, c interface{}) error {
					return nil
				}
				return args{fn}
			},
			send:      1,
			wantPanic: "should have 1 param",
		},
		{
			name: "errors on different consumption type",
			fields: fields{
				ctx:  context.TODO(),
				ch:   make(chan interface{}),
				done: make(chan struct{}),
			},
			argsfn: func(*interface{}) args {
				fn := func(a string) error {
					return nil
				}
				return args{fn}
			},
			send:    float64(1),
			wantErr: "^invalid type, want 'string' but got 'float64'$",
		},
		{
			name: "returns error on canceled context",
			fields: fields{
				ctx:  helperCanceledContext(),
				ch:   make(chan interface{}),
				done: make(chan struct{}),
			},
			argsfn: func(out *interface{}) args {
				fn := func(v interface{}) error {
					*out = v
					return nil
				}
				return args{fn}
			},
			send:    1,
			wantErr: "context canceled",
		},
		{
			name: "returns error on done closed",
			fields: fields{
				ctx: context.TODO(),
				ch:  make(chan interface{}),
				done: func() chan struct{} {
					ch := make(chan struct{})
					close(ch)
					return ch
				}(),
			},
			closeSend: func(c *procChan) {
				go func() {
					time.Sleep(time.Second)
					close(c.ch)
				}()
			},
			argsfn: func(out *interface{}) args {
				fn := func(v interface{}) error {
					*out = v
					return nil
				}
				return args{fn}
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := procChan{
				ctx:  tt.fields.ctx,
				ch:   tt.fields.ch,
				done: tt.fields.done,
			}
			closeSend := tt.closeSend
			if closeSend == nil {
				closeSend = func(*procChan) { c.close() }
			}
			if tt.send != nil {
				go func() {
					defer closeSend(&c)
					c.Send(tt.send)
				}()
			} else {
				closeSend(&c)
			}
			defer func() {
				var err error
				if p := recover(); p != nil {
					err = fmt.Errorf("%v", p)
				}
				if !matchError(tt.wantPanic, err) {
					t.Errorf("procChan.Send() panic = %v, wantPanic %v", err, tt.wantPanic)
				}
			}()
			var consumed interface{}
			a := tt.argsfn(&consumed)
			if err := c.Consume(a.ifn); !matchError(tt.wantErr, err) {
				t.Errorf("procChan.Consume() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(consumed, tt.wantConsumed) {
				t.Errorf("procChan.Consume() consumed = %v, want %v", consumed, tt.wantConsumed)
			}
		})
	}
}

func Test_procChan_cancel(t *testing.T) {
	type fields struct {
		ctx  context.Context
		ch   chan interface{}
		done chan struct{}
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "should close done channel",
			fields: fields{
				ctx:  context.TODO(),
				ch:   make(chan interface{}),
				done: make(chan struct{}),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := procChan{
				ctx:  tt.fields.ctx,
				ch:   tt.fields.ch,
				done: tt.fields.done,
			}
			c.cancel()
			var cancelled bool
			select {
			case <-c.done:
				cancelled = true
			case <-time.After(2 * time.Second):
			}
			if cancelled == false {
				t.Errorf("procChan.cancel() cancelled = %v, want %v", cancelled, true)
			}
		})
	}
}

func Test_procChan_close(t *testing.T) {
	type fields struct {
		ctx  context.Context
		ch   chan interface{}
		done chan struct{}
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "closes the message channel",
			fields: fields{
				ctx:  context.TODO(),
				ch:   make(chan interface{}),
				done: make(chan struct{}),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := procChan{
				ctx:  tt.fields.ctx,
				ch:   tt.fields.ch,
				done: tt.fields.done,
			}
			c.close()
			var closed bool
			select {
			case <-c.ch:
				closed = true
			case <-time.After(2 * time.Second):
			}
			if closed == false {
				t.Errorf("procChan.cancel() closed = %v, want %v", closed, true)
			}
		})
	}
}

type contextHelper struct {
	context.Context
	cancel func()
}

func helperContext() context.Context {
	ctx, cancel := context.WithCancel(context.TODO())
	return contextHelper{
		ctx,
		cancel,
	}
}

func helperCanceledContext() context.Context {
	ctx, cancel := context.WithCancel(context.TODO())
	cancel()
	return contextHelper{
		ctx,
		cancel,
	}
}

func helperTimeoutContext(d time.Duration) context.Context {
	ctx, cancel := context.WithTimeout(context.TODO(), d)
	return contextHelper{
		ctx,
		cancel,
	}
}
