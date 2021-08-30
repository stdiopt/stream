package stream

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/bwmarrin/snowflake"
	gomock "github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
)

func Test_newProc(t *testing.T) {
	type args struct {
		ctx context.Context
		c   Consumer
		s   Sender
	}
	tests := []struct {
		name string
		args args
		want *proc
	}{
		{
			name: "creates proc",
			args: args{
				ctx: context.TODO(),
				c:   fakeConsumer{"ca", nil},
				s:   fakeSender{"sa", nil},
			},
			want: &proc{
				ctx:      context.TODO(),
				Consumer: fakeConsumer{"ca", nil},
				Sender:   fakeSender{"sa", nil},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := newProc(tt.args.ctx, tt.args.c, tt.args.s); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newProc() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_proc_Consume(t *testing.T) {
	type fields struct {
		id         snowflake.ID
		ctx        context.Context
		Consumerfn func(ctrl *gomock.Controller) Consumer
		Sender     Sender
	}
	type args struct {
		fn interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		argsfn  func(*interface{}) args
		want    interface{}
		wantErr string
	}{
		{
			name: "calls underlying consumer",
			fields: fields{
				Consumerfn: func(ctrl *gomock.Controller) Consumer {
					p := NewMockConsumer(ctrl)
					p.EXPECT().
						Consume(
							gomock.AssignableToTypeOf((func(interface{}) error)(nil)),
						).Return(nil)
					return p
				},
			},
			argsfn: func(got *interface{}) args {
				return args{func(v interface{}) error {
					*got = v
					return nil
				}}
			},
		},
		{
			name: "returns underlying consumer error",
			fields: fields{
				Consumerfn: func(ctrl *gomock.Controller) Consumer {
					p := NewMockConsumer(ctrl)
					p.EXPECT().
						Consume(
							gomock.AssignableToTypeOf((func(interface{}) error)(nil)),
						).Return(errors.New("consumer error"))
					return p
				},
			},
			argsfn: func(got *interface{}) args {
				return args{func(v interface{}) error {
					*got = v
					return nil
				}}
			},
			wantErr: "consumer error",
		},
		{
			name:   "calls consumer func if no Consumer defined",
			fields: fields{},
			argsfn: func(got *interface{}) args {
				return args{func(v string) error {
					*got = v
					return nil
				}}
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			var consumer Consumer
			if tt.fields.Consumerfn != nil {
				consumer = tt.fields.Consumerfn(ctrl)
			}
			p := proc{
				id:       tt.fields.id,
				ctx:      tt.fields.ctx,
				Consumer: consumer,
				Sender:   tt.fields.Sender,
			}

			var got interface{}

			var args args
			if tt.argsfn != nil {
				args = tt.argsfn(&got)
			}

			if err := p.Consume(args.fn); !matchError(tt.wantErr, err) {
				t.Errorf("proc.Consume() error = %v, wantErr %v", err, tt.wantErr)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Error("wrong output\n- want + got\n", diff)
			}
		})
	}
}

func Test_proc_Send(t *testing.T) {
	type fields struct {
		id       snowflake.ID
		ctx      context.Context
		Consumer Consumer
		Senderfn func(ctrl *gomock.Controller) Sender
	}
	type args struct {
		v interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr string
	}{
		{
			name: "calls underlying sender",
			fields: fields{
				Senderfn: func(ctrl *gomock.Controller) Sender {
					p := NewMockSender(ctrl)
					p.EXPECT().Send(1)
					return p
				},
			},
			args: args{1},
		},
		{
			name: "returns error when underlying sender errors",
			fields: fields{
				Senderfn: func(ctrl *gomock.Controller) Sender {
					p := NewMockSender(ctrl)
					p.EXPECT().Send(1).Return(errors.New("sender error"))
					return p
				},
			},
			args:    args{1},
			wantErr: "sender error",
		},
		{
			name:   "returns nil when no underlying sender",
			fields: fields{},
			args:   args{1},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			var sender Sender
			if tt.fields.Senderfn != nil {
				sender = tt.fields.Senderfn(ctrl)
			}
			p := proc{
				id:       tt.fields.id,
				ctx:      tt.fields.ctx,
				Consumer: tt.fields.Consumer,
				Sender:   sender,
			}
			if err := p.Send(tt.args.v); !matchError(tt.wantErr, err) {
				t.Errorf("proc.Send() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_proc_Context(t *testing.T) {
	type fields struct {
		id       snowflake.ID
		ctx      context.Context
		Consumer Consumer
		Sender   Sender
	}
	tests := []struct {
		name   string
		fields fields
		want   context.Context
	}{
		{
			name: "returns context",
			fields: fields{
				ctx: context.TODO(),
			},
			want: context.TODO(),
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := proc{
				id:       tt.fields.id,
				ctx:      tt.fields.ctx,
				Consumer: tt.fields.Consumer,
				Sender:   tt.fields.Sender,
			}
			if got := p.Context(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("proc.Context() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_proc_cancel(t *testing.T) {
	type fields struct {
		id         snowflake.ID
		Consumerfn func(ctrl *gomock.Controller) Consumer
		Senderfn   Sender
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "cancel inner consumer",
			fields: fields{
				Consumerfn: func(ctrl *gomock.Controller) Consumer {
					p := NewMockConsumer(ctrl)
					p.EXPECT().cancel()
					return p
				},
			},
		},
		{
			name:   "call cancel with a nil consumer",
			fields: fields{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			var consumer Consumer

			if tt.fields.Consumerfn != nil {
				consumer = tt.fields.Consumerfn(ctrl)
			}
			p := proc{
				id:       tt.fields.id,
				Consumer: consumer,
			}
			p.cancel()
		})
	}
}

func TestMakeConsumerFunc(t *testing.T) {
	type args struct {
		fn interface{}
	}
	tests := []struct {
		name      string
		argsfn    func(*interface{}) args
		fnarg     interface{}
		want      interface{}
		wantErr   string
		wantPanic string
	}{
		{
			name: "calls built consumer func",
			argsfn: func(got *interface{}) args {
				return args{func(v interface{}) error {
					*got = v
					return nil
				}}
			},
			fnarg: 1,
			want:  1,
		},
		{
			name: "calls shortcut built consumer func",
			argsfn: func(got *interface{}) args {
				return args{func(v []byte) error {
					*got = v
					return nil
				}}
			},
			fnarg: []byte{1, 2},
			want:  []byte{1, 2},
		},
		{
			name: "returns error on func error",
			argsfn: func(got *interface{}) args {
				return args{func(v int) error {
					return errors.New("func error")
				}}
			},
			wantErr: "func error",
		},
		{
			name: "returns calls []byte shortcut built consumer func",
			argsfn: func(got *interface{}) args {
				return args{func(v []byte) error {
					*got = v
					return nil
				}}
			},
			fnarg:   1,
			wantErr: `invalid type, want '\[\]uint8' but got 'int'`,
		},
		{
			name: "panic when wrong params",
			argsfn: func(got *interface{}) args {
				return args{func(a, b int) error {
					return nil
				}}
			},
			wantPanic: `should have 1 param`,
		},
		{
			name: "panic when wrong return",
			argsfn: func(got *interface{}) args {
				return args{func(b int) int {
					return 0
				}}
			},
			wantPanic: `func should have an error return only`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				p := recover()
				if !matchPanic(tt.wantPanic, p) {
					t.Errorf("MakeConsumerFunc() panic = %v, wantPanic %v", p, tt.wantPanic)
				}
			}()

			var got interface{}
			args := tt.argsfn(&got)

			gotfn := MakeConsumerFunc(args.fn)
			if gotfn == nil {
				t.Errorf("MakeConsumerFunc() got is nil = %v, want %v", gotfn == nil, false)
			}

			if err := gotfn(tt.fnarg); !matchError(tt.wantErr, err) {
				t.Errorf("S().Run() error = %v, wantErr %v", err, tt.wantErr)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Error("wrong output\n- want + got\n", diff)
			}
		})
	}
}
