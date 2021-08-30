package strmutil

import (
	"errors"
	"testing"

	"github.com/stdiopt/stream/strmtest"
)

func TestRepeat(t *testing.T) {
	type send struct {
		value       interface{}
		senderError error
		want        []interface{}
	}
	type args struct {
		n int
	}
	tests := []struct {
		name      string
		args      args
		sends     []send
		wantErr   string
		wantPanic string
	}{

		{
			name: "repeats input",
			args: args{2},
			sends: []send{
				{value: 'x', want: []interface{}{'x', 'x'}},
			},
		},
		{
			name: "repeats multiple inputs",
			args: args{2},
			sends: []send{
				{value: "test1", want: []interface{}{"test1", "test1"}},
				{value: "test2", want: []interface{}{"test2", "test2"}},
			},
		},
		{
			name: "panic on invalid repeat param",
			args: args{0},
			sends: []send{
				{value: "test1", want: []interface{}{"test1", "test1"}},
				{value: "test2", want: []interface{}{"test2", "test2"}},
			},
			wantPanic: "invalid repeat param '0', should be > 0$",
		},
		{
			name: "returns error if sender errors",
			args: args{3},
			sends: []send{
				{value: 1, senderError: errors.New("sender error")},
			},
			wantErr: "strmutil.Repeat.* sender error$",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				p := recover()
				if !strmtest.MatchPanic(tt.wantPanic, p) {
					t.Errorf("Repeat() panic = %v, wantPanic %v", p, tt.wantPanic)
				}
			}()
			pp := Repeat(tt.args.n)
			if pp == nil {
				t.Errorf("Repeat() is nil = %v, want %v", pp == nil, false)
			}

			st := strmtest.New(t, pp)
			for _, s := range tt.sends {
				st.Send(s.value).
					WithSenderError(s.senderError).
					Expect(s.want...)
			}
			st.ExpectError(tt.wantErr).Run()
		})
	}
}
