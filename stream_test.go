package stream

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestLine(t *testing.T) {
	tests := []struct {
		name     string
		PipeFunc Pipe
		Send     interface{}

		Want        []interface{}
		WantErrorRE string
	}{

		{
			name:     "pass through when no pipes",
			PipeFunc: Line(),
			Send:     1,
			Want:     []interface{}{1},
		},
		{
			name: "run single pipe ",
			PipeFunc: Line(func(p Proc) error {
				return p.Consume(p.Send)
			}),
			Send: 1,
			Want: []interface{}{1},
		},
		{
			name: "multiple pipe",
			PipeFunc: Line(
				func(p Proc) error {
					return p.Consume(func(n int) error {
						return p.Send(n * n)
					})
				},
				func(p Proc) error {
					return p.Consume(func(n int) error {
						return p.Send(fmt.Sprint(n))
					})
				},
			),
			Send: 2,
			Want: []interface{}{"4"},
		},
		{
			name: "multiple output",
			PipeFunc: Line(func(p Proc) error {
				return p.Consume(func(n int) error {
					for i := 0; i < n; i++ {
						if err := p.Send(i); err != nil {
							return err
						}
					}
					return nil
				})
			}),
			Send: 5,
			Want: []interface{}{0, 1, 2, 3, 4},
		},
		{
			name: "line error",
			PipeFunc: Line(func(p Proc) error {
				return p.Consume(func(v interface{}) error {
					return errors.New("consumer error")
				})
			}),
			Send:        1,
			WantErrorRE: "^consumer error$",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got []interface{}
			ctx := context.Background()

			c := newProcChan(ctx, 0)
			s := newProcChan(ctx, 0)

			done := make(chan struct{})
			go func() {
				defer close(done)
				c.Consume(func(v interface{}) error {
					got = append(got, v)
					return nil
				})
			}()

			if tt.Send != nil {
				go func() {
					defer s.close()
					s.Send(tt.Send)
				}()
			} else {
				s.close()
			}
			// inverted on purpose
			var err error
			go func() {
				defer c.close()
				err = tt.PipeFunc(newProc(ctx, s, c))
			}()
			<-done

			if !matchError(tt.WantErrorRE, err) {
				t.Errorf("wrong error\nwant: %v\n got: %v\n", tt.WantErrorRE, err)
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
