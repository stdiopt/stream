package stream

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
)

func Test_pipe_newChan(t *testing.T) {
	type fields struct {
		fn func(Proc) error
	}
	type args struct {
		ctx    context.Context
		buffer int
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   *procChan
	}{
		{
			name:   "returns new proc channel",
			fields: fields{},
			args:   args{context.TODO(), 0},
			want: &procChan{
				ctx:  context.TODO(),
				ch:   make(chan interface{}),
				done: make(chan struct{}),
			},
		},
		{
			name:   "returns new buffered proc channel",
			fields: fields{},
			args:   args{context.TODO(), 5},
			want: &procChan{
				ctx:  context.TODO(),
				ch:   make(chan interface{}, 5),
				done: make(chan struct{}),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := pipe{
				fn: tt.fields.fn,
			}
			gotProc := p.newChan(tt.args.ctx, tt.args.buffer)

			got, ok := gotProc.(*procChan)
			if !ok {
				t.Errorf("newProcChan() should be *procChan = %v, want %v", ok, true)
			}

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

func Test_pipe_Run(t *testing.T) {
	type fields struct {
		fn func(Proc) error
	}
	type args struct {
		ctx   context.Context
		infn  func(ctrl *gomock.Controller) Consumer
		outfn func(ctrl *gomock.Controller) Sender
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr string
	}{
		{
			name: "calls proc func",
			fields: fields{
				fn: func(p Proc) error {
					return p.Consume(p.Send)
				},
			},
			args: args{
				ctx: context.TODO(),
				infn: func(ctrl *gomock.Controller) Consumer {
					p := NewMockConsumer(ctrl)
					p.EXPECT().Consume(
						gomock.AssignableToTypeOf((func(interface{}) error)(nil)),
					).Return(nil)

					return p
				},
				outfn: func(ctrl *gomock.Controller) Sender {
					p := NewMockSender(ctrl)
					return p
				},
			},
		},
		{
			name: "returns error from proc func",
			fields: fields{
				fn: func(p Proc) error {
					return errors.New("procfn error")
				},
			},
			args: args{
				ctx: context.TODO(),
				infn: func(ctrl *gomock.Controller) Consumer {
					p := NewMockConsumer(ctrl)
					return p
				},
				outfn: func(ctrl *gomock.Controller) Sender {
					p := NewMockSender(ctrl)
					return p
				},
			},
			wantErr: "procfn error",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			var c Consumer
			var s Sender
			if tt.args.infn != nil {
				c = tt.args.infn(ctrl)
			}
			if tt.args.outfn != nil {
				s = tt.args.outfn(ctrl)
			}
			p := pipe{
				fn: tt.fields.fn,
			}
			if err := p.Run(tt.args.ctx, c, s); !matchError(tt.wantErr, err) {
				t.Errorf("pipe.Run() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFunc(t *testing.T) {
	type args struct {
		fn func(Proc) error
	}
	tests := []struct {
		name      string
		args      args
		send      interface{}
		want      []interface{}
		wantErr   string
		wantPanic string
	}{
		{
			name: "returns pipe with proc func",
			args: args{func(p Proc) error {
				return p.Consume(p.Send)
			}},
			send: 1,
			want: []interface{}{1},
		},
		{
			name: "returns pipe with diferent type func",
			args: args{func(p Proc) error {
				return p.Consume(func(int) error {
					return nil
				})
			}},
			send: 1,
			want: []interface{}{1},
		},
		{
			name: "returns pipe with diferent type func",
			args: args{func(p Proc) error {
				return p.Consume(func(int) error {
					return nil
				})
			}},
			send:    "1",
			wantErr: "invalid type, want 'int' but got 'string'",
		},
		{
			name: "returns pipe with diferent type func",
			args: args{func(p Proc) error {
				return p.Consume(func(v interface{}) error {
					return errors.New("consumer error")
				})
			}},
			send:    1,
			wantErr: "consumer error",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				p := recover()
				if !matchPanic(tt.wantPanic, p) {
					t.Errorf("Func().Run() panic = %v, wantPanic %v", p, tt.wantPanic)
				}
			}()
			// here we test if the pipe returned behaves fine
			pp := Func(tt.args.fn)
			if pp == nil {
				t.Errorf("Func() is nil = %v, want %v", pp == nil, false)
			}

			var got []interface{}
			capture := Override{
				CTX: context.TODO(),
				ConsumeFunc: func(fn ConsumerFunc) error {
					return fn(tt.send)
				},
				SendFunc: func(v interface{}) error {
					got = append(got, v)
					return nil
				},
			}
			if err := pp.Run(context.TODO(), capture, capture); !matchError(tt.wantErr, errors.Unwrap(err)) {
				t.Errorf("Func().Run() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestT(t *testing.T) {
	type args struct {
		fn interface{}
	}
	tests := []struct {
		name      string
		args      args
		send      interface{}
		want      []interface{}
		wantErr   string
		wantPanic string
	}{
		{
			name: "builds proc func",
			args: args{func(n int) (string, error) {
				return fmt.Sprint(n), nil
			}},
			send: 1,
			want: []interface{}{"1"},
		},
		{
			name: "translates nil to zero",
			args: args{func(n int) (string, error) {
				return fmt.Sprint(n), nil
			}},
			send: nil,
			want: []interface{}{"0"},
		},
		{
			name: "transforms different type",
			args: args{func(v interface{}) (string, error) {
				return fmt.Sprintf("transform: %v", v), nil
			}},
			send: 1,
			want: []interface{}{"transform: 1"},
		},
		{
			name: "returns error when default consumer errors",
			args: args{func(interface{}) (interface{}, error) {
				return 1, errors.New("consumerFunc error")
			}},
			send:    1,
			wantErr: "consumerFunc error",
		},
		{
			name: "returns error on consumer error",
			args: args{func(int) (interface{}, error) {
				return 1, errors.New("consumerFunc error")
			}},
			send:    1,
			wantErr: "consumerFunc error",
		},
		{
			name: "returns error on different send type",
			args: args{func(int) (interface{}, error) {
				return "1", nil
			}},
			send:    "1",
			wantErr: "invalid type, want 'int' but got 'string'",
		},
		{
			name: "panics on wrong number of inputs",
			args: args{func(int, int) (string, error) {
				return "", nil
			}},
			wantPanic: "func should have 1 param",
		},
		{
			name: "panics on wrong number of output",
			args: args{func(int) error {
				return nil
			}},
			wantPanic: "func should have 2 outputs and last must be 'error'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				p := recover()
				if !matchPanic(tt.wantPanic, p) {
					t.Errorf("T().Run() panic = %v, wantPanic %v", p, tt.wantPanic)
				}
			}()
			// here we test if the pipe returned behaves fine
			pp := T(tt.args.fn)
			if pp == nil {
				t.Errorf("T().Run() is nil = %v, want %v", pp == nil, false)
			}

			var got []interface{}
			capture := Override{
				CTX: context.TODO(),
				ConsumeFunc: func(fn ConsumerFunc) error {
					return fn(tt.send)
				},
				SendFunc: func(v interface{}) error {
					got = append(got, v)
					return nil
				},
			}

			if err := pp.Run(context.TODO(), capture, capture); !matchError(tt.wantErr, err) {
				t.Errorf("T().Run() error = %v, wantErr %v", err, tt.wantErr)
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Error("wrong output\n- want + got\n", diff)
			}
		})
	}
}

func TestS(t *testing.T) {
	type args struct {
		fn interface{}
	}
	tests := []struct {
		name      string
		args      args
		send      interface{}
		want      []interface{}
		wantErr   string
		wantPanic string
	}{
		{
			name: "builds proc func",
			args: args{
				func(s Sender, v interface{}) error {
					return s.Send(v)
				},
			},
			send: 1,
			want: []interface{}{1},
		},
		{
			name: "builds default []byte proc func",
			args: args{
				func(s Sender, v []byte) error {
					return s.Send(v)
				},
			},
			send: []byte{1},
			want: []interface{}{[]byte{1}},
		},
		{
			name: "builds typed proc func",
			args: args{
				func(s Sender, v float64) error {
					return s.Send(v)
				},
			},
			send: float64(1),
			want: []interface{}{float64(1)},
		},
		{
			name: "calls with zero param if send is nil",
			args: args{
				func(s Sender, v string) error {
					return s.Send(v)
				},
			},
			send: nil,
			want: []interface{}{""},
		},
		{
			name: "sends multiple",
			args: args{
				func(s Sender, v int) error {
					s.Send(v)
					s.Send(v)
					return nil
				},
			},
			send: 1,
			want: []interface{}{1, 1},
		},
		{
			name: "returns error when default errors",
			args: args{
				func(s Sender, v interface{}) error {
					return errors.New("consumer error")
				},
			},
			send:    1,
			wantErr: "consumer error",
		},
		{
			name: "returns error when consumer errors",
			args: args{
				func(s Sender, v int) error {
					return errors.New("consumer error")
				},
			},
			send:    1,
			wantErr: "consumer error",
		},
		{
			name: "returns error on different send type",
			args: args{
				func(s Sender, v int) error {
					return nil
				},
			},
			send:    "1",
			wantErr: "invalid type, want 'int' but got 'string'",
		},
		{
			name: "panics on different number of inputs",
			args: args{
				func(s Sender, a, b int) error {
					return nil
				},
			},
			wantPanic: "func should have 2 params",
		},
		{
			name: "panics on wrong first param",
			args: args{
				func(s Proc, b int) error {
					return nil
				},
			},
			wantPanic: `first param should be '\w+\.Sender'`,
		},
		{
			name: "panics on wrong return ",
			args: args{
				func(s Sender, b int) (int, int) {
					return 0, 0
				},
			},
			wantPanic: `func should have 1 output and must be 'error'`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				p := recover()
				if !matchPanic(tt.wantPanic, p) {
					t.Errorf("S().Run() panic = %v, wantPanic %v", p, tt.wantPanic)
				}
			}()
			// here we test if the pipe returned behaves fine
			pp := S(tt.args.fn)
			if pp == nil {
				t.Errorf("S().Run() is nil = %v, want %v", pp == nil, false)
			}

			var got []interface{}
			capture := Override{
				CTX: context.TODO(),
				ConsumeFunc: func(fn ConsumerFunc) error {
					return fn(tt.send)
				},
				SendFunc: func(v interface{}) error {
					got = append(got, v)
					return nil
				},
			}
			if err := pp.Run(context.TODO(), capture, capture); !matchError(tt.wantErr, err) {
				t.Errorf("S().Run() error = %v, wantErr %v", err, tt.wantErr)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Error("wrong output\n- want + got\n", diff)
			}
		})
	}
}
