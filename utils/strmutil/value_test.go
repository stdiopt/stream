package strmutil

import (
	"errors"
	"testing"

	"github.com/stdiopt/stream/utils/strmtest"
)

func TestValue(t *testing.T) {
	type args struct {
		vs []interface{}
	}
	tests := []struct {
		name        string
		args        args
		send        []interface{}
		senderError error
		want        []interface{}
		wantErr     string
	}{

		{
			name: "sends value",
			args: args{[]interface{}{1}},
			send: []interface{}{'x'},
			want: []interface{}{1},
		},
		{
			name: "sends multiple values",
			args: args{[]interface{}{1, 2, "test"}},
			send: []interface{}{
				"test1", "test2",
			},
			want: []interface{}{
				1, 2, "test",
				1, 2, "test",
			},
		},
		{
			name: "does not send values if no params",
			args: args{},
			send: []interface{}{"test1", "test2"},
		},
		{
			name:        "returns error if sender errors",
			args:        args{[]interface{}{1}},
			send:        []interface{}{1},
			senderError: errors.New("sender error"),
			wantErr:     "strmutil.Value.* sender error$",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pp := Value(tt.args.vs...)
			if pp == nil {
				t.Errorf("Value() is nil = %v, want %v", pp == nil, false)
			}

			st := strmtest.New(t, pp)
			for _, s := range tt.send {
				st.Send(s).WithSenderError(tt.senderError)
			}
			st.ExpectFull(tt.want...).
				ExpectError(tt.wantErr).
				Run()
		})
	}
}
