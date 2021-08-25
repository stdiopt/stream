package strmutil

import (
	"errors"
	"testing"

	strm "github.com/stdiopt/stream"
	"github.com/stdiopt/stream/strmtest"
)

func TestRepeat(t *testing.T) {
	type send struct {
		value       interface{}
		senderError error
		want        []interface{}
		wantErrorRE string
	}
	tests := []struct {
		name        string
		pipe        strm.Pipe
		sends       []send
		wantErrorRE string
	}{
		{
			name: "repeats input",
			pipe: Repeat(2),
			sends: []send{
				{value: 'x', want: []interface{}{'x', 'x'}},
			},
		},
		{
			name: "repeats multiple inputs",
			pipe: Repeat(2),
			sends: []send{
				{value: "test1", want: []interface{}{"test1", "test1"}},
				{value: "test2", want: []interface{}{"test2", "test2"}},
			},
		},
		{
			name: "errors on invalid repeat param",
			pipe: Repeat(0),
			sends: []send{
				{value: "test1", want: []interface{}{"test1", "test1"}},
				{value: "test2", want: []interface{}{"test2", "test2"}},
			},
			wantErrorRE: "strmutil.Repeat.* invalid repeat param '0', should be > 0$",
		},
		{
			name: "returns error if sender errors",
			pipe: Repeat(3),
			sends: []send{
				{value: 1, senderError: errors.New("sender error")},
			},
			wantErrorRE: "strmutil.Repeat.* sender error$",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := strmtest.New(t, tt.pipe)
			for _, s := range tt.sends {
				st.Send(s.value).
					WithSenderError(s.senderError).
					Expect(s.want...)
			}
			st.ExpectError(tt.wantErrorRE).Run()
		})
	}
}
