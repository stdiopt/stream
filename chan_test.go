package stream

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

func TestConsumex(t *testing.T) {
	testError := errors.New("test")
	type testCase struct {
		ctx      context.Context
		fn       func(Chan, *[]interface{}) consumerFunc
		wantData []interface{}
		wantErr  error
	}

	test := func(tt testCase) func(t *testing.T) {
		return func(t *testing.T) {
			ch := newChan(tt.ctx, 0)
			go func() {
				defer ch.Close()
				for i := 0; i < 4; i++ {
					ch.send(message{value: i}) // nolint: errcheck
				}
			}()

			consumed := []interface{}{}
			fn := tt.fn(ch, &consumed)

			err := ch.consume(fn)
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
		}
	}

	t.Run("consume with no error", test(testCase{
		ctx: context.Background(),
		fn: func(c Chan, out *[]interface{}) consumerFunc {
			return func(m message) error {
				*out = append(*out, m.value)
				return nil
			}
		},
		wantData: []interface{}{0, 1, 2, 3},
		wantErr:  nil,
	}))
	t.Run("consume should return testError", test(testCase{
		ctx: context.Background(),
		fn: func(c Chan, out *[]interface{}) consumerFunc {
			return func(m message) error {
				*out = append(*out, m.value)
				return testError
			}
		},
		wantData: []interface{}{0},
		wantErr:  testError,
	}))
	t.Run("returns if context is cancelled", test(testCase{
		ctx: func() context.Context {
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			return ctx
		}(),
		fn: func(c Chan, out *[]interface{}) consumerFunc {
			return func(m message) error {
				*out = append(*out, m.value)
				return nil
			}
		},
		wantData: []interface{}{},
		wantErr:  context.Canceled,
	}))
}

func TestSend(t *testing.T) {
	type testCase struct {
		ctx         context.Context
		values      []interface{}
		consumerErr error
		wantErr     error
		wantData    []interface{}
	}
	test := func(tt testCase) func(t *testing.T) {
		return func(t *testing.T) {
			t.Helper()
			consumed := []interface{}{}
			wg := sync.WaitGroup{}

			ch := newChan(tt.ctx, 0)
			wg.Add(1)
			go func() {
				defer wg.Done()
				ch.consume(func(m message) error { // nolint: errcheck
					consumed = append(consumed, m.value)
					return tt.consumerErr
				})
			}()

			for _, s := range tt.values {
				err := ch.send(message{value: s})
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
		}
	}
	t.Run("send without error",
		test(testCase{
			ctx:      context.Background(),
			values:   []interface{}{1, 2, 3},
			wantErr:  (error)(nil),
			wantData: []interface{}{1, 2, 3},
		}),
	)
	t.Run("returns canceled err on a cancelled context",
		test(testCase{
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			}(),
			values:   []interface{}{1, 2, 3},
			wantErr:  context.Canceled,
			wantData: []interface{}{},
		}),
	)
	t.Run("returns deadline exceed err on a cancelled context",
		test(testCase{
			ctx: func() context.Context {
				ctx, cancel := context.WithDeadline(context.Background(), time.Now())
				cancel()
				return ctx
			}(),
			values:   []interface{}{1, 2, 3},
			wantErr:  context.DeadlineExceeded,
			wantData: []interface{}{},
		}),
	)
}
