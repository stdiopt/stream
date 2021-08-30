// Package stream provides stuff
package stream

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestLine(t *testing.T) {
	type args struct {
		pps []Pipe
	}
	tests := []struct {
		name string
		ctx  context.Context
		args args

		send    interface{}
		want    []interface{}
		wantErr string
	}{
		{
			name: "pass through when no pipes",
			args: args{},
			send: 1,
			want: []interface{}{1},
		},
		{
			name: "run single pipe",
			args: args{[]Pipe{
				Func(func(p Proc) error {
					return p.Consume(p.Send)
				}),
			}},
			send: 1,
			want: []interface{}{1},
		},
		{
			name: "run multiple pipe",
			args: args{[]Pipe{
				Func(func(p Proc) error {
					return p.Consume(func(n int) error {
						return p.Send(n * n)
					})
				}),
				Func(func(p Proc) error {
					return p.Consume(func(n int) error {
						return p.Send(fmt.Sprint(n))
					})
				}),
			}},
			send: 2,
			want: []interface{}{"4"},
		},
		{
			name: "run multiple output",
			args: args{[]Pipe{
				S(func(s Sender, n int) error {
					for i := 0; i < n; i++ {
						if err := s.Send(i); err != nil {
							return err
						}
					}
					return nil
				}),
			}},
			send: 5,
			want: []interface{}{0, 1, 2, 3, 4},
		},
		{
			name: "returns error on run",
			args: args{[]Pipe{
				Func(func(p Proc) error { return p.Consume(p.Send) }),
				Func(func(p Proc) error {
					return p.Consume(func(v interface{}) error {
						return errors.New("consumer error")
					})
				}),
			}},
			send:    1,
			wantErr: "consumer error$",
		},
		{
			name: "break line",
			args: args{[]Pipe{
				S(func(s Sender, _ interface{}) error {
					for i := 0; i < 10; i++ {
						if err := s.Send(i); err != nil {
							return err
						}
					}
					return nil
				}),
				Func(func(p Proc) error {
					count := 0
					return p.Consume(func(v interface{}) error {
						count++
						if count > 2 {
							return ErrBreak
						}
						return p.Send(v)
					})
				}),
			}},
			send: 1,
			want: []interface{}{0, 1},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got []interface{}
			ctx := tt.ctx
			if ctx == nil {
				ctx = context.TODO()
			}

			sender := newPipeChan(ctx, 0)
			consumer := newPipeChan(ctx, 0)
			wg := sync.WaitGroup{}
			wg.Add(2)
			go func() {
				defer wg.Done()
				defer consumer.close()
				consumer.Send(tt.send) // nolint: errcheck
			}()
			go func() {
				defer wg.Done()
				// nolint: errcheck
				sender.Consume(func(v interface{}) error {
					got = append(got, v)
					return nil
				})
			}()
			pp := Line(tt.args.pps...)
			if pp == nil {
				t.Errorf("Line() is nil = %v, want %v", pp == nil, false)
			}
			if err := pp.Run(context.TODO(), consumer, sender); !matchError(tt.wantErr, err) {
				t.Errorf("Line() pp error = %v, wantErr %v", err, tt.wantErr)
			}
			sender.close()
			wg.Wait()

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Error("Line() wrong output\n- want + got\n", diff)
			}
		})
	}
}

func TestTee(t *testing.T) {
	type args struct {
		pps []Pipe
	}
	tests := []struct {
		name    string
		ctx     context.Context
		args    args
		sends   []interface{}
		want    map[interface{}]int
		wantErr string
	}{
		{
			name:  "send if tee is empty",
			args:  args{},
			sends: []interface{}{1},
			want:  map[interface{}]int{1: 1},
		},
		{
			name: "send on a single Tee",
			args: args{[]Pipe{
				Func(func(p Proc) error { return p.Consume(p.Send) }),
			}},
			sends: []interface{}{1},
			want:  map[interface{}]int{1: 1},
		},
		{
			name: "send to multiple pipes",
			args: args{[]Pipe{
				S(func(s Sender, v interface{}) error {
					return s.Send(fmt.Sprintf("Pipe 1: %v", v))
				}),
				S(func(s Sender, v interface{}) error {
					return s.Send(fmt.Sprintf("Pipe 2: %v", v))
				}),
			}},
			sends: []interface{}{1},
			want: map[interface{}]int{
				"Pipe 1: 1": 1,
				"Pipe 2: 1": 1,
			},
		},
		{
			name: "returns error when a pipe errors",
			args: args{[]Pipe{
				S(func(s Sender, v interface{}) error {
					return s.Send(fmt.Sprintf("Pipe 1: %v", v))
				}),
				S(func(s Sender, _ interface{}) error {
					return errors.New("pipe error")
				}),
			}},
			sends: []interface{}{1},
			want: map[interface{}]int{
				"Pipe 1: 1": 1,
			},
			wantErr: "pipe error",
		},
		{
			name: "returns error when a ch errors",
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(context.TODO())
				cancel()
				return ctx
			}(),
			args: args{[]Pipe{
				S(func(s Sender, v interface{}) error {
					time.Sleep(time.Millisecond)
					return s.Send(fmt.Sprintf("Pipe 1: %v", v))
				}),
				S(func(s Sender, v interface{}) error {
					return s.Send(fmt.Sprintf("Pipe 2: %v", v))
				}),
			}},
			sends:   []interface{}{1},
			wantErr: "context canceled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := tt.ctx
			if ctx == nil {
				ctx = context.TODO()
			}

			var got map[interface{}]int
			capture := Override{
				CTX: ctx,
				ConsumeFunc: func(fn ConsumerFunc) error {
					for _, v := range tt.sends {
						if err := fn(v); err != nil {
							return err
						}
					}
					return nil
				},
				SendFunc: func(v interface{}) error {
					if got == nil {
						got = map[interface{}]int{}
					}
					got[v]++
					return nil
				},
			}
			pp := Tee(tt.args.pps...)
			if pp == nil {
				t.Errorf("Tee() is nil = %v, want %v", pp == nil, false)
			}
			if err := pp.Run(ctx, capture, capture); !matchError(tt.wantErr, err) {
				t.Errorf("Tee().Run() error = %v, wantErr %v", err, tt.wantErr)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Error("Tee() wrong output\n- want + got\n", diff)
			}
		})
	}
}

func TestWorkers(t *testing.T) {
	type args struct {
		n   int
		pps []Pipe
	}
	tests := []struct {
		name  string
		args  args
		sends []interface{}

		want    map[interface{}]int
		wantDur time.Duration
		wantErr string
	}{
		{
			name:  "receives value",
			args:  args{n: 1},
			sends: []interface{}{1},
			want:  map[interface{}]int{1: 1},
		},

		{
			name:  "runs with one if n <= 0",
			args:  args{n: -1},
			sends: []interface{}{1},
			want:  map[interface{}]int{1: 1},
		},
		{
			name: "runs concurrently",
			args: args{
				n: 8,
				pps: []Pipe{
					S(func(s Sender, v interface{}) error {
						time.Sleep(time.Second)
						return s.Send(v)
					}),
				},
			},
			sends:   []interface{}{1, 2, 3, 4},
			wantDur: 2 * time.Second,
			want:    map[interface{}]int{1: 1, 2: 1, 3: 1, 4: 1},
		},
		{
			name: "returns error when a sender errors",
			args: args{
				n: 1,
				pps: []Pipe{
					Func(func(p Proc) error {
						return errors.New("proc error")
					}),
				},
			},
			wantErr: "proc error",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.TODO()

			var got map[interface{}]int

			sender := newPipeChan(ctx, 0)
			consumer := newPipeChan(ctx, 0)
			wg := sync.WaitGroup{}
			wg.Add(2)
			go func() {
				defer wg.Done()
				defer consumer.close()
				for _, v := range tt.sends {
					consumer.Send(v) // nolint: errcheck
				}
			}()
			go func() {
				defer wg.Done()
				// nolint: errcheck
				sender.Consume(func(v interface{}) error {
					if got == nil {
						got = map[interface{}]int{}
					}
					got[v]++
					return nil
				})
			}()

			pp := Workers(tt.args.n, tt.args.pps...)
			if pp == nil {
				t.Errorf("Workers() is nil = %v, want %v", pp == nil, false)
			}
			mark := time.Now()
			if err := pp.Run(context.TODO(), consumer, sender); !matchError(tt.wantErr, err) {
				t.Errorf("Workers().Run() error = %v, wantErr %v", err, tt.wantErr)
			}
			sender.close()
			wg.Wait()
			dur := time.Since(mark)
			if tt.wantDur != 0 && dur > tt.wantDur {
				t.Errorf("Workers() duration = %v, wantErr %v", dur, tt.wantDur)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Error("Workers() wrong output\n- want + got\n", diff)
			}
		})
	}
}

func TestBuffer(t *testing.T) {
	type args struct {
		n   int
		pps []Pipe
	}
	tests := []struct {
		name    string
		ctx     context.Context
		args    args
		send    interface{}
		want    []interface{}
		wantErr string
	}{

		{
			name: "pass through when no pipes",
			args: args{},
			send: 1,
			want: []interface{}{1},
		},
		// TODO: Add test cases.
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got []interface{}
			ctx := tt.ctx
			if ctx == nil {
				ctx = context.TODO()
			}

			sender := newPipeChan(ctx, 0)
			consumer := newPipeChan(ctx, 0)
			wg := sync.WaitGroup{}
			wg.Add(2)
			go func() {
				defer wg.Done()
				defer consumer.close()
				consumer.Send(tt.send) // nolint: errcheck
			}()
			go func() {
				defer wg.Done()
				// nolint: errcheck
				sender.Consume(func(v interface{}) error {
					got = append(got, v)
					return nil
				})
			}()
			pp := Buffer(tt.args.n, tt.args.pps...)
			if pp == nil {
				t.Errorf("Buffer() is nil = %v, want %v", pp == nil, false)
			}
			if err := pp.Run(context.TODO(), consumer, sender); !matchError(tt.wantErr, err) {
				t.Errorf("Buffer() pp error = %v, wantErr %v", err, tt.wantErr)
			}
			sender.close()
			wg.Wait()

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Error("Buffer() wrong output\n- want + got\n", diff)
			}
		})
	}
	/*for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Buffer(tt.args.n, tt.args.pps...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Buffer() = %v, want %v", got, tt.want)
			}
		})
	}*/
}

func TestRun(t *testing.T) {
	type args struct {
		pps []Pipe
	}
	tests := []struct {
		name    string
		args    args
		wantErr string
	}{
		{
			name: "run line",
			args: args{[]Pipe{
				Func(func(p Proc) error { return p.Consume(p.Send) }),
				Func(func(p Proc) error { return p.Consume(p.Send) }),
			}},
		},
		{
			name: "returns error",
			args: args{[]Pipe{
				Func(func(p Proc) error { return p.Consume(p.Send) }),
				Func(func(p Proc) error {
					return errors.New("run error")
				}),
			}},
			wantErr: "run error",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Run(tt.args.pps...); !matchError(tt.wantErr, err) {
				t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRunWithContext(t *testing.T) {
	type args struct {
		ctx context.Context
		pps []Pipe
	}
	tests := []struct {
		name    string
		args    args
		wantErr string
	}{
		{
			name: "runs with context",
			args: args{
				ctx: context.TODO(),
				pps: []Pipe{
					Func(func(p Proc) error { return p.Consume(p.Send) }),
					Func(func(p Proc) error { return p.Consume(p.Send) }),
				},
			},
		},
		{
			name: "runs with context and cancel",
			args: args{
				ctx: func() context.Context {
					ctx, cancel := context.WithCancel(context.TODO())
					cancel()
					return ctx
				}(),
				pps: []Pipe{
					Func(func(p Proc) error { return p.Consume(p.Send) }),
					Func(func(p Proc) error { return p.Consume(p.Send) }),
				},
			},
			wantErr: "context canceled",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := RunWithContext(tt.args.ctx, tt.args.pps...); !matchError(tt.wantErr, err) {
				t.Errorf("RunWithContext() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
