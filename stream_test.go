package stream

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"sync"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestLine(t *testing.T) {
	tests := []struct {
		name string
		Pipe Pipe
		Send interface{}

		Want        []interface{}
		WantErrorRE string
	}{
		{
			name: "pass through when no pipes",
			Pipe: Line(),
			Send: 1,
			Want: []interface{}{1},
		},
		{
			name: "run single pipe ",
			Pipe: Line(
				Func(func(p Proc) error { return p.Consume(p.Send) }),
			),
			Send: 1,
			Want: []interface{}{1},
		},
		{
			name: "multiple pipe",
			Pipe: Line(
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
			),
			Send: 2,
			Want: []interface{}{"4"},
		},
		{
			name: "multiple output",
			Pipe: Line(
				Func(func(p Proc) error {
					return p.Consume(func(n int) error {
						for i := 0; i < n; i++ {
							if err := p.Send(i); err != nil {
								return err
							}
						}
						return nil
					})
				}),
			),
			Send: 5,
			Want: []interface{}{0, 1, 2, 3, 4},
		},
		{
			name: "returns error on run",
			Pipe: Line(
				Func(func(p Proc) error { return p.Consume(p.Send) }),
				Func(func(p Proc) error {
					return p.Consume(func(v interface{}) error {
						return errors.New("consumer error")
					})
				}),
			),
			Send:        1,
			WantErrorRE: "consumer error$",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got []interface{}
			ctx := context.Background()

			capture := Override{
				CTX: ctx,
				ConsumeFunc: func(fn ConsumerFunc) error {
					return fn(tt.Send)
				},
				SendFunc: func(v interface{}) error {
					got = append(got, v)
					return nil
				},
			}

			err := tt.Pipe.run(ctx, capture, capture)
			if !matchError(tt.WantErrorRE, err) {
				t.Errorf("wrong error\nwant: %v\n got: %v\n", tt.WantErrorRE, err)
			}
			if diff := cmp.Diff(tt.Want, got); diff != "" {
				t.Error("wrong output\n- want + got\n", diff)
			}
		})
	}
}

func TestWorkers(t *testing.T) {
	type Send struct {
		Value       interface{}
		WantErrorRE string
	}

	tests := []struct {
		name         string
		Pipe         Pipe
		Sends        []Send
		SenderError  error
		Want         map[interface{}]int
		WantErrorRE  string
		WantDuration time.Duration
	}{
		{
			name: "receives value",
			Pipe: Workers(1),
			Sends: []Send{
				{Value: 1},
			},
			Want: map[interface{}]int{1: 1},
		},
		{
			name: "process in parallel",
			Pipe: Workers(8, T(func(v interface{}) (interface{}, error) {
				time.Sleep(time.Second)
				return v, nil
			})),
			Sends: []Send{
				{Value: 1},
				{Value: 2},
				{Value: 3},
				{Value: 4},
			},
			WantDuration: 2 * time.Second,
			Want:         map[interface{}]int{1: 1, 2: 1, 3: 1, 4: 1},
		},
		{
			name:        "returns error when sender errors",
			Pipe:        Workers(1),
			SenderError: errors.New("sender error"),
			Sends: []Send{
				{Value: 1},
			},
			WantErrorRE: "sender error$",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			// consumer := newProcChan(ctx, 0)

			ch := make(chan interface{})
			go func() {
				defer close(ch)
				for _, m := range tt.Sends {
					ch <- m.Value
				}
			}()

			var got map[interface{}]int
			mu := sync.Mutex{}
			sender := Override{
				ConsumeFunc: func(fn ConsumerFunc) error {
					for v := range ch {
						if err := fn(v); err != nil {
							return err
						}
					}
					return nil
				},
				SendFunc: func(v interface{}) error {
					mu.Lock()
					defer mu.Unlock()
					if tt.SenderError != nil {
						return tt.SenderError
					}
					if got == nil {
						got = map[interface{}]int{}
					}
					got[v]++
					return nil
				},
			}
			mark := time.Now()
			err := tt.Pipe.run(ctx, sender, sender)
			if !matchError(tt.WantErrorRE, err) {
				t.Errorf("wrong error\nwant: %v\n got: %v\n", tt.WantErrorRE, err)
			}
			dur := time.Since(mark)
			if tt.WantDuration != 0 && dur > tt.WantDuration {
				t.Errorf("exceed time limit\nwant: <%v\n got: >%v\n", tt.WantDuration, dur)
			}
			if diff := cmp.Diff(tt.Want, got); diff != "" {
				t.Error("wrong output\n- want + got\n", diff)
			}
		})
	}
}

func matchError(match string, err error) bool {
	if match == "" {
		return err == nil
	}

	errStr := ""
	if err != nil {
		errStr = err.Error()
	}

	re := regexp.MustCompile(match)
	return re.MatchString(errStr)
}
