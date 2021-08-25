package strmutil

import (
	"errors"
	"testing"

	strm "github.com/stdiopt/stream"
	"github.com/stdiopt/stream/strmtest"
)

func TestLimit(t *testing.T) {
	tests := []struct {
		name        string
		pipe        strm.Pipe
		send        []interface{}
		want        []interface{}
		senderError error

		wantErrorRE string
	}{
		{
			name: "sends all",
			pipe: Limit(3),
			send: []interface{}{1, 2},
			want: []interface{}{1, 2},
		},
		{
			name:        "limits consumption",
			pipe:        Limit(2),
			send:        []interface{}{1, 2, 3, 4},
			want:        []interface{}{1, 2},
			wantErrorRE: "strmutil.Limit.* break$",
		},
		{
			name:        "limits consumption to 0",
			pipe:        Limit(0),
			send:        []interface{}{1, 2, 3, 4},
			wantErrorRE: "strmutil.Limit.* break$",
		},
		{
			name:        "returns error when sender errors",
			pipe:        Limit(1),
			send:        []interface{}{1, 2, 3},
			senderError: errors.New("sender error"),
			wantErrorRE: "strmutil.Limit.* sender error$",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := strmtest.New(t, tt.pipe)
			for _, s := range tt.send {
				st.Send(s).WithSenderError(tt.senderError)
			}
			st.ExpectFull(tt.want...).
				ExpectError(tt.wantErrorRE).
				Run()
		})
	}
}
