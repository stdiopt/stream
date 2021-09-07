package strmfs

import (
	"testing"

	"github.com/stdiopt/stream/strmtest"
)

func TestListFiles(t *testing.T) {
	type args struct {
		pattern string
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
			name: "list files",
			args: args{"*"},
			send: []interface{}{"./testdata"},
			want: []interface{}{
				"testdata/01/test.txt",
				"testdata/01/test2.txt",
				"testdata/02/readonly.txt",
				"testdata/02/subdir/test.csv",
				"testdata/02/test.json",
			},
		},
		{
			name: "list files from multiple sends",
			args: args{"*"},
			send: []interface{}{
				"./testdata/02",
				"./testdata/01",
			},
			want: []interface{}{
				"testdata/02/readonly.txt",
				"testdata/02/subdir/test.csv",
				"testdata/02/test.json",
				"testdata/01/test.txt",
				"testdata/01/test2.txt",
			},
		},
		{
			name: "list patterned files",
			args: args{"*.txt"},
			send: []interface{}{"./testdata"},
			want: []interface{}{
				"testdata/01/test.txt",
				"testdata/01/test2.txt",
				"testdata/02/readonly.txt",
			},
		},
		{
			name:    "returns error on invalid pattern",
			args:    args{"["},
			send:    []interface{}{"./testdata"},
			wantErr: "strmfs.ListFiles.* syntax error in pattern",
		},
		{
			name:    "returns error on dir error",
			args:    args{"*"},
			send:    []interface{}{"nofile"},
			wantErr: "strmfs.ListFiles.* no such file or directory",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pp := FindFromInput(tt.args.pattern)
			if pp == nil {
				t.Errorf("ListFiles() is nil = %v, want %v", pp == nil, false)
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
