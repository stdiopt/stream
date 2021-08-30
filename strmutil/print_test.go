package strmutil

import (
	"bytes"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stdiopt/stream/strmtest"
)

func TestPrint(t *testing.T) {
	type args struct {
		prefix string
	}
	tests := []struct {
		name        string
		args        args
		send        []interface{}
		senderError error

		want       []interface{}
		wantErr    string
		wantOutput string
	}{
		{
			name:       "prints",
			args:       args{""},
			send:       []interface{}{1},
			want:       []interface{}{1},
			wantOutput: "1\n",
		},
		{
			name:       "prints with a prefix",
			args:       args{"prefix"},
			send:       []interface{}{1},
			want:       []interface{}{1},
			wantOutput: "[prefix] 1\n",
		},
		{
			name:       "prints []byte as string",
			args:       args{"prefix"},
			send:       []interface{}{[]byte{'a', 'b', 'c'}},
			want:       []interface{}{[]byte{'a', 'b', 'c'}},
			wantOutput: "[prefix] abc\n",
		},
		{
			name:       "multiple types",
			args:       args{""},
			send:       []interface{}{1, "two", 2.1},
			want:       []interface{}{1, "two", 2.1},
			wantOutput: "1\ntwo\n2.1\n",
		},
		{
			name:        "returns error when sender errors",
			args:        args{""},
			send:        []interface{}{1, "two", 2.1},
			senderError: errors.New("sender error"),
			wantErr:     "strmutil.Print.* sender error$",
			wantOutput:  "1\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			SetOutput(buf)

			pp := Print(tt.args.prefix)
			if pp == nil {
				t.Errorf("Print() is nil = %v, want %v", pp == nil, false)
			}

			st := strmtest.New(t, pp)
			for _, s := range tt.send {
				st.Send(s).WithSenderError(tt.senderError)
			}
			st.ExpectFull(tt.want...).
				ExpectError(tt.wantErr).
				Run()

			if diff := cmp.Diff(buf.String(), tt.wantOutput); diff != "" {
				t.Error(diff)
			}
		})
	}
}
