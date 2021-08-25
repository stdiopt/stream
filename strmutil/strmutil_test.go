package strmutil

import (
	"bytes"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stdiopt/stream/strmtest"
)

func TestPrint(t *testing.T) {
	tests := []struct {
		name        string
		arg         string
		send        []interface{}
		senderError error

		want        []interface{}
		wantErrorRE string
		matchOutput string
	}{
		{
			name:        "prints",
			arg:         "",
			send:        []interface{}{1},
			want:        []interface{}{1},
			matchOutput: "1\n",
		},
		{
			name:        "prints with a prefix",
			arg:         "prefix",
			send:        []interface{}{1},
			want:        []interface{}{1},
			matchOutput: "[prefix] 1\n",
		},

		{
			name:        "prints []byte as string",
			arg:         "prefix",
			send:        []interface{}{[]byte{'a', 'b', 'c'}},
			want:        []interface{}{[]byte{'a', 'b', 'c'}},
			matchOutput: "[prefix] abc\n",
		},
		{
			name:        "multiple types",
			arg:         "",
			send:        []interface{}{1, "two", 2.1},
			want:        []interface{}{1, "two", 2.1},
			matchOutput: "1\ntwo\n2.1\n",
		},
		{
			name:        "returns error when sender errors",
			arg:         "",
			send:        []interface{}{1, "two", 2.1},
			senderError: errors.New("sender error"),
			wantErrorRE: "sender error$",
			matchOutput: "1\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			SetOutput(buf)

			strmtest.New(t, Print(tt.arg)).
				Send(tt.send...).
				SenderError(tt.senderError).
				Expect(tt.want...).
				ExpectMatchError(tt.wantErrorRE).
				Run()

			if diff := cmp.Diff(buf.String(), tt.matchOutput); diff != "" {
				t.Error(diff)
			}
		})
	}
}
