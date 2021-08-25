package strmutil

import (
	"errors"
	"testing"
	"time"

	strm "github.com/stdiopt/stream"
	"github.com/stdiopt/stream/strmtest"
)

func TestWait(t *testing.T) {
	tests := []struct {
		name         string
		pipe         strm.Pipe
		send         []interface{}
		senderError  error
		want         []interface{}
		wantDuration time.Duration
		wantErrorRE  string
	}{
		{
			name:         "waits the duration",
			pipe:         Wait(500 * time.Millisecond),
			send:         []interface{}{'x'},
			want:         []interface{}{'x'},
			wantDuration: 499 * time.Millisecond,
		},
		{
			name:         "waits the duration for each value",
			pipe:         Wait(200 * time.Millisecond),
			send:         []interface{}{1, 2},
			want:         []interface{}{1, 2},
			wantDuration: 399 * time.Millisecond,
		},
		{
			name:        "returns error when sender errors",
			pipe:        Wait(200 * time.Millisecond),
			send:        []interface{}{1},
			senderError: errors.New("sender error"),
			wantErrorRE: "strmutil.Wait.* sender error$",
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
				ExpectMinDuration(tt.wantDuration).
				Run()
		})
	}
}
