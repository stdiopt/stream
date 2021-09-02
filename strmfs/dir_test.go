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
				"testdata/02/subdir/test.csv",
				"testdata/02/test.json",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pp := ListFiles(tt.args.pattern)
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
