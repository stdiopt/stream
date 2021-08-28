//go:generate mockgen -source=./stream.go -destination=proc_mock_test.go -package=stream Proc
package stream

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

	gomock "github.com/golang/mock/gomock"
)

func TestOverride_Send(t *testing.T) {
	type fields struct {
		Procfn      func(ctrl *gomock.Controller) Proc
		CTX         context.Context
		SendFunc    func(got *interface{}) func(v interface{}) error
		ConsumeFunc func(ConsumerFunc) error
	}
	type args struct {
		v interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		send    interface{}
		want    interface{}
		wantErr string
	}{
		{
			name:   "returns nil with no proc or sendfunc set",
			fields: fields{},
			args:   args{1},
		},
		{
			name: "calls SendFunc on send",
			fields: fields{
				Procfn: func(ctrl *gomock.Controller) Proc { return NewMockProc(ctrl) },
				SendFunc: func(got *interface{}) func(v interface{}) error {
					return func(v interface{}) error {
						*got = v
						return nil
					}
				},
			},
			args: args{1},
			want: 1,
		},
		{
			name: "calls Proc.Send on send",
			fields: fields{
				Procfn: func(ctrl *gomock.Controller) Proc {
					p := NewMockProc(ctrl)
					p.EXPECT().Send(1).Return(nil)
					return p
				},
			},
			args: args{1},
		},
		{
			name: "returns errors on send",
			fields: fields{
				Procfn: func(ctrl *gomock.Controller) Proc { return NewMockProc(ctrl) },
				SendFunc: func(got *interface{}) func(v interface{}) error {
					return func(v interface{}) error {
						return errors.New("send error")
					}
				},
			},
			args:    args{1},
			wantErr: "send error",
		},
		{
			name: "returns errors on proc send",
			fields: fields{
				Procfn: func(ctrl *gomock.Controller) Proc {
					p := NewMockProc(ctrl)
					p.EXPECT().
						Send(1).
						Return(errors.New("proc send error"))

					return p
				},
			},
			args:    args{1},
			wantErr: "proc send error",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got interface{}

			var sendFunc func(interface{}) error
			if tt.fields.SendFunc != nil {
				sendFunc = tt.fields.SendFunc(&got)
			}
			var p Proc
			if tt.fields.Procfn != nil {
				ctrl := gomock.NewController(t)
				defer ctrl.Finish()
				p = tt.fields.Procfn(ctrl)

			}
			o := Override{
				Proc:        p,
				CTX:         tt.fields.CTX,
				SendFunc:    sendFunc,
				ConsumeFunc: tt.fields.ConsumeFunc,
			}
			if err := o.Send(tt.args.v); !matchError(tt.wantErr, err) {
				t.Errorf("Override.Send() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Override.Send() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOverride_Consume(t *testing.T) {
	type fields struct {
		Procfn      func(ctrl *gomock.Controller) Proc
		CTX         context.Context
		SendFunc    func(v interface{}) error
		ConsumeFunc func(ConsumerFunc) error
	}
	type args struct {
		ifn interface{}
	}
	tests := []struct {
		name      string
		fields    fields
		argsfn    func(*interface{}) args
		want      interface{}
		wantErr   string
		wantPanic string
	}{
		{
			name: "calls consume func",
			fields: fields{
				ConsumeFunc: func(fn ConsumerFunc) error {
					return fn(1)
				},
			},
			argsfn: func(got *interface{}) args {
				fn := func(v interface{}) error {
					*got = v
					return nil
				}
				return args{fn}
			},
			want: 1,
		},
		{
			name: "calls consume func when a proc is set",
			fields: fields{
				Procfn: func(ctrl *gomock.Controller) Proc {
					return NewMockProc(ctrl)
				},
				ConsumeFunc: func(fn ConsumerFunc) error {
					return fn("test")
				},
			},
			argsfn: func(got *interface{}) args {
				fn := func(v interface{}) error {
					*got = v
					return nil
				}
				return args{fn}
			},
			want: "test",
		},
		{
			name: "use proc consume method",
			fields: fields{
				Procfn: func(ctrl *gomock.Controller) Proc {
					p := NewMockProc(ctrl)
					p.EXPECT().Consume(gomock.AssignableToTypeOf(
						func(interface{}) error { return nil },
					)).Return(nil)
					return p
				},
			},
			argsfn: func(got *interface{}) args {
				fn := func(v interface{}) error {
					*got = v
					return nil
				}
				return args{fn}
			},
		},
		{
			name:   "returns nil on no consumer overrides",
			fields: fields{},
			argsfn: func(got *interface{}) args {
				fn := func(v interface{}) error {
					*got = v
					return nil
				}
				return args{fn}
			},
		},
		{
			name: "returns error on consumerFunc error",
			fields: fields{
				Procfn: func(ctrl *gomock.Controller) Proc {
					p := NewMockProc(ctrl)
					p.EXPECT().
						Consume(gomock.Any()).
						Return(errors.New("proc consume error"))
					return p
				},
			},

			argsfn: func(got *interface{}) args {
				fn := func(v interface{}) error {
					*got = v
					return errors.New("func error")
				}
				return args{fn}
			},
			wantErr: "proc consume error",
		},
		{
			name: "returns error on consumerFunc error",
			fields: fields{
				ConsumeFunc: func(v ConsumerFunc) error {
					return errors.New("consumerFunc error")
				},
			},
			argsfn: func(got *interface{}) args {
				fn := func(v interface{}) error {
					*got = v
					return nil
				}
				return args{fn}
			},
			wantErr: "consumerFunc error",
		},
		{
			name: "returns error on invalid consumer type",
			fields: fields{
				ConsumeFunc: func(v ConsumerFunc) error { return nil },
			},
			argsfn: func(*interface{}) args {
				fn := func(a, b interface{}) error {
					return nil
				}
				return args{fn}
			},
			wantPanic: "func should have 1 params",
		},
		{
			name: "returns error on invalid consumer return type",
			fields: fields{
				ConsumeFunc: func(v ConsumerFunc) error { return nil },
			},
			argsfn: func(*interface{}) args {
				fn := func(a interface{}) int {
					return 0
				}
				return args{fn}
			},
			wantPanic: "func should have an error return only",
		},
		{
			name: "returns error on invalid consumer type",
			fields: fields{
				ConsumeFunc: func(fn ConsumerFunc) error {
					return fn("string")
				},
			},
			argsfn: func(*interface{}) args {
				fn := func(a int) error {
					return nil
				}
				return args{fn}
			},
			wantErr: "invalid type, want 'int' but got 'string'",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var p Proc
			if tt.fields.Procfn != nil {
				ctrl := gomock.NewController(t)
				defer ctrl.Finish()
				p = tt.fields.Procfn(ctrl)

			}
			o := Override{
				Proc:        p,
				CTX:         tt.fields.CTX,
				SendFunc:    tt.fields.SendFunc,
				ConsumeFunc: tt.fields.ConsumeFunc,
			}

			defer func() {
				var err error
				if p := recover(); p != nil {
					err = fmt.Errorf("%v", p)
				}
				if !matchError(tt.wantPanic, err) {
					t.Errorf("Override.Consume() panic = %v, wantPanic %v", err, tt.wantPanic)
				}
			}()
			var got interface{}
			args := tt.argsfn(&got)
			if err := o.Consume(args.ifn); !matchError(tt.wantErr, err) {
				t.Errorf("Override.Consume() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Override.Consume() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOverride_cancel(t *testing.T) {
	type fields struct {
		Procfn      func(ctrl *gomock.Controller) Proc
		CTX         context.Context
		SendFunc    func(v interface{}) error
		ConsumeFunc func(ConsumerFunc) error
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "call cancel on underlying proc",
			fields: fields{
				Procfn: func(ctrl *gomock.Controller) Proc {
					p := NewMockProc(ctrl)
					p.EXPECT().cancel()
					return p
				},
			},
		},
		{
			name:   "does nothing if proc not assigned",
			fields: fields{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var p Proc
			if tt.fields.Procfn != nil {
				ctrl := gomock.NewController(t)
				defer ctrl.Finish()
				p = tt.fields.Procfn(ctrl)

			}
			m := Override{
				Proc:        p,
				CTX:         tt.fields.CTX,
				SendFunc:    tt.fields.SendFunc,
				ConsumeFunc: tt.fields.ConsumeFunc,
			}
			m.cancel()
		})
	}
}

func TestOverride_close(t *testing.T) {
	type fields struct {
		Procfn      func(ctrl *gomock.Controller) Proc
		CTX         context.Context
		SendFunc    func(v interface{}) error
		ConsumeFunc func(ConsumerFunc) error
	}
	tests := []struct {
		name   string
		fields fields
	}{
		{
			name: "calls underlying proc close",
			fields: fields{
				Procfn: func(ctrl *gomock.Controller) Proc {
					p := NewMockProc(ctrl)
					p.EXPECT().close()
					return p
				},
			},
		},
		{
			name:   "does nothing if proc is not set",
			fields: fields{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var p Proc
			if tt.fields.Procfn != nil {
				ctrl := gomock.NewController(t)
				defer ctrl.Finish()
				p = tt.fields.Procfn(ctrl)

			}
			m := Override{
				Proc:        p,
				CTX:         tt.fields.CTX,
				SendFunc:    tt.fields.SendFunc,
				ConsumeFunc: tt.fields.ConsumeFunc,
			}
			m.close()
		})
	}
}

func TestOverride_Context(t *testing.T) {
	type fields struct {
		Procfn      func(ctrl *gomock.Controller) Proc
		CTX         context.Context
		SendFunc    func(v interface{}) error
		ConsumeFunc func(ConsumerFunc) error
	}
	tests := []struct {
		name   string
		fields fields
		want   context.Context
	}{
		{
			name:   "returns todo context if nothing set",
			fields: fields{},
			want:   context.TODO(),
		},
		{
			name: "returns context from Field",
			fields: fields{
				CTX: fakeContext{"a", nil},
			},
			want: fakeContext{"a", nil},
		},
		{
			name: "returns context from proc",
			fields: fields{
				Procfn: func(ctrl *gomock.Controller) Proc {
					p := NewMockProc(ctrl)
					p.EXPECT().
						Context().
						Return(fakeContext{"a", nil})
					return p
				},
			},
			want: fakeContext{"a", nil},
		},
		{
			name: "returns context from field if both set",
			fields: fields{
				CTX: fakeContext{"field", nil},
				Procfn: func(ctrl *gomock.Controller) Proc {
					p := NewMockProc(ctrl)
					p.EXPECT().
						Context().
						Return(fakeContext{"proc", nil}).AnyTimes()
					return p
				},
			},
			want: fakeContext{"field", nil},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var p Proc
			if tt.fields.Procfn != nil {
				ctrl := gomock.NewController(t)
				defer ctrl.Finish()
				p = tt.fields.Procfn(ctrl)
			}
			m := Override{
				Proc:        p,
				CTX:         tt.fields.CTX,
				SendFunc:    tt.fields.SendFunc,
				ConsumeFunc: tt.fields.ConsumeFunc,
			}
			if got := m.Context(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Override.Context() = %v, want %v", got, tt.want)
			}
		})
	}
}
