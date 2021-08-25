package strmutil

import (
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	strm "github.com/stdiopt/stream"
	"github.com/stdiopt/stream/strmtest"
)

func TestPass(t *testing.T) {
	tests := []struct {
		name        string
		send        []interface{}
		want        []interface{}
		wantPass    []interface{}
		senderError error
		passerError error

		wantErrorRE string
	}{
		{
			name:     "pass value to other proc",
			send:     []interface{}{1},
			want:     []interface{}{1},
			wantPass: []interface{}{1},
		},
		{
			name:        "returns error when sender errors",
			send:        []interface{}{1, 2, 3},
			senderError: errors.New("sender error"),
			wantPass:    []interface{}{1},
			wantErrorRE: "strmutil.Pass.* sender error$",
		},
		{
			name:        "returns error when passer errors",
			send:        []interface{}{1, 2, 3},
			passerError: errors.New("passer error"),
			wantErrorRE: "strmutil.Pass.* passer error$",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got []interface{}

			p := strm.Override{SendFunc: func(v interface{}) error {
				if tt.passerError != nil {
					return tt.passerError
				}
				got = append(got, v)
				return nil
			}}
			st := strmtest.New(t, Pass(p))
			for _, s := range tt.send {
				st.Send(s).WithSenderError(tt.senderError)
			}
			st.ExpectFull(tt.want...).
				ExpectError(tt.wantErrorRE).
				Run()

			if diff := cmp.Diff(got, tt.wantPass); diff != "" {
				t.Error("wrong pass output\n- want + got\n", diff)
			}
		})
	}
}
