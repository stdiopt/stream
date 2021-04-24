package stream_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stdiopt/stream"
)

func TestConsume(t *testing.T) {
	testError := errors.New("test")
	tests := []struct {
		name     string
		ctx      context.Context
		fn       func(stream.Chan, *[]interface{}) stream.ConsumerFunc
		wantData []interface{}
		wantErr  error
	}{
		{
			name: "consume with no error",
			ctx:  context.Background(),
			fn: func(c stream.Chan, out *[]interface{}) stream.ConsumerFunc {
				return func(_ context.Context, v interface{}) error {
					*out = append(*out, v)
					return nil
				}
			},
			wantData: []interface{}{0, 1, 2, 3},
			wantErr:  nil,
		},
		{
			name: "consume should return testError",
			ctx:  context.Background(),
			fn: func(c stream.Chan, out *[]interface{}) stream.ConsumerFunc {
				return func(_ context.Context, v interface{}) error {
					*out = append(*out, v)
					return testError
				}
			},
			wantData: []interface{}{0},
			wantErr:  testError,
		},
		{
			name: "returns if context is cancelled",
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			}(),
			fn: func(c stream.Chan, out *[]interface{}) stream.ConsumerFunc {
				return func(_ context.Context, v interface{}) error {
					*out = append(*out, v)
					return nil
				}
			},
			wantData: []interface{}{},
			wantErr:  context.Canceled,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ch := stream.NewChan(tt.ctx, 0)
			go func() {
				defer ch.Close()
				for i := 0; i < 4; i++ {
					ch.Send(tt.ctx, i) // nolint: errcheck
				}
			}()

			consumed := []interface{}{}
			fn := tt.fn(ch, &consumed)

			err := ch.Consume(fn)
			if want := tt.wantErr; err != want {
				t.Errorf("\nwant: %v\n got: %v\n", want, err)
			}

			if want := len(tt.wantData); len(consumed) != want {
				t.Errorf("\nwant: %v\n got: %v\n", want, len(consumed))
			}
			for i, v := range consumed {
				if want := tt.wantData[i]; v != want {
					t.Errorf("\nwant: %v\n got: %v\n", want, v)
				}
			}
		})
	}
}

func TestSend(t *testing.T) {
	tests := []struct {
		name        string
		ctx         context.Context
		values      []interface{}
		consumerErr error
		wantErr     error
		wantData    []interface{}
	}{
		{
			name:     "send without error",
			ctx:      context.Background(),
			values:   []interface{}{1, 2, 3},
			wantErr:  (error)(nil),
			wantData: []interface{}{1, 2, 3},
		},
		{
			name: "returns canceled err on a cancelled context",
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			}(),
			values:   []interface{}{1, 2, 3},
			wantErr:  context.Canceled,
			wantData: []interface{}{},
		},
		{
			name: "returns deadline exceed err on a cancelled context",
			ctx: func() context.Context {
				ctx, cancel := context.WithDeadline(context.Background(), time.Now())
				cancel()
				return ctx
			}(),
			values:   []interface{}{1, 2, 3},
			wantErr:  context.DeadlineExceeded,
			wantData: []interface{}{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			consumed := []interface{}{}
			wg := sync.WaitGroup{}

			ch := stream.NewChan(tt.ctx, 0)
			wg.Add(1)
			go func() {
				defer wg.Done()
				ch.Consume(func(_ context.Context, v interface{}) error { // nolint: errcheck
					consumed = append(consumed, v)
					return tt.consumerErr
				})
			}()

			for _, s := range tt.values {
				err := ch.Send(tt.ctx, s)
				if want := tt.wantErr; err != want {
					t.Errorf("\nwant: %v\n got: %v\n", want, err)
					break
				}
			}
			ch.Close()
			wg.Wait()

			if want := len(tt.wantData); len(consumed) != want {
				t.Fatalf("\nwant: %v\n got: %v\n", want, len(consumed))
			}
			for i, c := range tt.wantData {
				if want := c; consumed[i] != want {
					t.Errorf("\nwant: %v\n got: %v\n", want, consumed[i])
				}
			}
		})
	}
}
