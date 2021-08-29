package strmutil

import (
	"errors"
	"testing"
	"time"

	"github.com/stdiopt/stream/strmtest"
)

func TestWait(t *testing.T) {
	type args struct {
		d time.Duration
	}
	tests := []struct {
		name         string
		args         args
		send         []interface{}
		senderError  error
		want         []interface{}
		wantDuration time.Duration
		wantErr      string
	}{
		{
			name:         "waits the duration",
			args:         args{500 * time.Millisecond},
			send:         []interface{}{'x'},
			want:         []interface{}{'x'},
			wantDuration: 499 * time.Millisecond,
		},
		{
			name:         "waits the duration for each value",
			args:         args{200 * time.Millisecond},
			send:         []interface{}{1, 2},
			want:         []interface{}{1, 2},
			wantDuration: 399 * time.Millisecond,
		},
		{
			name:        "returns error when sender errors",
			args:        args{200 * time.Millisecond},
			send:        []interface{}{1},
			senderError: errors.New("sender error"),
			wantErr:     "strmutil.Wait.* sender error$",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pp := Wait(tt.args.d)
			if pp == nil {
				t.Errorf("Wait() is nil = %v, want %v", pp == nil, false)
			}

			st := strmtest.New(t, pp)
			for _, s := range tt.send {
				st.Send(s).WithSenderError(tt.senderError)
			}
			st.ExpectFull(tt.want...).
				ExpectError(tt.wantErr).
				ExpectMinDuration(tt.wantDuration).
				Run()
		})
	}
}
