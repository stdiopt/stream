package strmutil

import (
	"errors"
	"testing"

	strm "github.com/stdiopt/stream"
	"github.com/stdiopt/stream/strmtest"
)

func TestValue(t *testing.T) {
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
			name: "sends value",
			pipe: Value(1),
			sends: []send{
				{value: 'x', want: []interface{}{1}},
			},
		},
		{
			name: "sends multiple values",
			pipe: Value(1, 2, "test"),
			sends: []send{
				{value: "test1", want: []interface{}{1, 2, "test"}},
				{value: "test2", want: []interface{}{1, 2, "test"}},
			},
		},
		{
			name: "does not send values if no params",
			pipe: Value(),
			sends: []send{
				{value: "test1"},
				{value: "test2"},
			},
		},
		{
			name: "returns error if sender errors",
			pipe: Value(1),
			sends: []send{
				{value: 1, senderError: errors.New("sender error")},
			},
			wantErrorRE: "strmutil.Value.* sender error$",
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
