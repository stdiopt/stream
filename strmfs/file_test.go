package strmfs

import (
	"errors"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stdiopt/stream/strmtest"
)

func TestWriteFile(t *testing.T) {
	type args struct {
		path string
	}
	tests := []struct {
		name        string
		args        args
		send        []interface{}
		senderError error
		clear       func()

		want       []interface{}
		wantErr    string
		skipOutput bool
		wantOutput string
	}{
		{
			name: "writes a file",
			args: args{"./testdata/temp-test"},
			clear: func() {
				os.RemoveAll("./testdata/temp-test")
			},
			send: []interface{}{
				[]byte(`hello world`),
			},
			wantOutput: "hello world",
		},
		{
			name: "writes a file with multiple sends",
			args: args{"./testdata/temp-test"},
			clear: func() {
				os.RemoveAll("./testdata/temp-test")
			},
			send: []interface{}{
				[]byte(`hello world`),
				[]byte(` again`),
			},
			wantOutput: "hello world again",
		},
		{
			name: "create sub dir",
			args: args{"./testdata/test/temp-test"},
			clear: func() {
				os.RemoveAll("./testdata/test")
			},
			send: []interface{}{
				[]byte(`hello world`),
				[]byte(` again`),
			},
			wantOutput: "hello world again",
		},
		{
			name: "errors when Mkdir errors",
			args: args{"./testdata/01/test.txt/otherthing"},
			send: []interface{}{
				[]byte(`hello world`),
			},
			skipOutput: true,
			wantErr:    "strmfs.WriteFile.* mkdir testdata/01/test.txt: not a directory",
		},
		{
			name: "errors when Create errors",
			args: args{"./testdata/01"},
			send: []interface{}{
				[]byte(`hello world`),
			},
			skipOutput: true,
			wantErr:    "strmfs.WriteFile.* open ./testdata/01: is a directory",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.clear != nil {
				defer tt.clear()
			}

			pp := WriteFile(tt.args.path)
			if pp == nil {
				t.Errorf("WriteFile() is nil = %v, want %v", pp == nil, false)
			}

			st := strmtest.New(t, pp)
			for _, s := range tt.send {
				st.Send(s).WithSenderError(tt.senderError)
			}
			st.ExpectFull(tt.want...).
				ExpectError(tt.wantErr).
				Run()

			if tt.skipOutput {
				return
			}
			data, err := os.ReadFile(tt.args.path)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(string(data), tt.wantOutput); diff != "" {
				t.Error(diff)
			}
		})
	}
}

func TestReadFile(t *testing.T) {
	tests := []struct {
		name        string
		send        []interface{}
		senderError error

		want    []interface{}
		wantErr string
	}{
		{
			name: "reads a file",
			send: []interface{}{
				"./testdata/01/test.txt",
			},
			want: []interface{}{
				[]byte("1\n"),
			},
		},
		{
			name: "reads multiple files",
			send: []interface{}{
				"./testdata/01/test.txt",
				"./testdata/01/test2.txt",
			},
			want: []interface{}{
				[]byte("1\n"),
				[]byte("test2\n"),
			},
		},
		{
			name: "returns error when opening file errors",
			send: []interface{}{
				"./testdata/notfound",
			},
			wantErr: "strmfs.ReadFile.* open ./testdata/notfound: no such file or directory",
		},
		{
			name: "returns error when sender errors",
			send: []interface{}{
				"./testdata/01/test.txt",
			},
			senderError: errors.New("sender error"),
			wantErr:     "strmfs.ReadFile.* sender error",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pp := ReadFile()
			if pp == nil {
				t.Errorf("ReadFile() is nil = %v, want %v", pp == nil, false)
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
