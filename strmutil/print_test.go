package strmutil

import (
	"bytes"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	strm "github.com/stdiopt/stream"
	"github.com/stdiopt/stream/strmtest"
)

func TestPrint(t *testing.T) {
	tests := []struct {
		name        string
		pipe        strm.Pipe
		send        []interface{}
		want        []interface{}
		senderError error

		wantErrorRE string
		matchOutput string
	}{
		{
			name:        "prints",
			pipe:        Print(""),
			send:        []interface{}{1},
			want:        []interface{}{1},
			matchOutput: "1\n",
		},
		{
			name:        "prints with a prefix",
			pipe:        Print("prefix"),
			send:        []interface{}{1},
			want:        []interface{}{1},
			matchOutput: "[prefix] 1\n",
		},
		{
			name:        "prints []byte as string",
			pipe:        Print("prefix"),
			send:        []interface{}{[]byte{'a', 'b', 'c'}},
			want:        []interface{}{[]byte{'a', 'b', 'c'}},
			matchOutput: "[prefix] abc\n",
		},
		{
			name:        "multiple types",
			pipe:        Print(""),
			send:        []interface{}{1, "two", 2.1},
			want:        []interface{}{1, "two", 2.1},
			matchOutput: "1\ntwo\n2.1\n",
		},
		{
			name:        "returns error when sender errors",
			pipe:        Print(""),
			send:        []interface{}{1, "two", 2.1},
			senderError: errors.New("sender error"),
			wantErrorRE: "strmutil.Print.* sender error$",
			matchOutput: "1\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			SetOutput(buf)

			st := strmtest.New(t, tt.pipe)
			for _, s := range tt.send {
				st.Send(s).WithSenderError(tt.senderError)
			}
			st.ExpectFull(tt.want...).
				ExpectError(tt.wantErrorRE).
				Run()

			if diff := cmp.Diff(buf.String(), tt.matchOutput); diff != "" {
				t.Error(diff)
			}
		})
	}
}
