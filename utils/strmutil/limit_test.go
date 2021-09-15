package strmutil

import (
	"errors"
	"testing"

	"github.com/stdiopt/stream/utils/strmtest"
)

func TestLimit(t *testing.T) {
	type args struct {
		n int
	}
	tests := []struct {
		name        string
		args        args
		send        []interface{}
		senderError error

		want    []interface{}
		wantErr string
	}{

		{
			name: "sends all",
			args: args{3},
			send: []interface{}{1, 2},
			want: []interface{}{1, 2},
		},
		{
			name:    "limits consumption",
			args:    args{2},
			send:    []interface{}{1, 2, 3, 4},
			want:    []interface{}{1, 2},
			wantErr: "strmutil.Limit.* break$",
		},
		{
			name:    "limits consumption to 0",
			args:    args{0},
			send:    []interface{}{1, 2, 3, 4},
			wantErr: "strmutil.Limit.* break$",
		},
		{
			name:        "returns error when sender errors",
			args:        args{1},
			send:        []interface{}{1, 2, 3},
			senderError: errors.New("sender error"),
			wantErr:     "strmutil.Limit.* sender error$",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pp := Limit(tt.args.n)
			if pp == nil {
				t.Errorf("Limit() is nil = %v, want %v", pp == nil, false)
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
