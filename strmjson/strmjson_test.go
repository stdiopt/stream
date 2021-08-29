package strmjson

import (
	"bytes"
	"reflect"
	"testing"

	strm "github.com/stdiopt/stream"
	"github.com/stdiopt/stream/strmtest"
)

func TestDecode(t *testing.T) {
	type args struct {
		v interface{}
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
			name: "Decodes json []byte",
			args: args{},
			send: []interface{}{
				[]byte(`{"name:"test"}`),
			},
			want: []interface{}{
				map[string]interface{}{
					"name": "test",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pp := Decode(tt.args.v)
			if pp == nil {
				t.Errorf("Decode() is nil = %v, want %v", pp == nil, false)
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

func TestEncode(t *testing.T) {
	tests := []struct {
		name string
		want strm.Pipe
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Encode(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Encode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDump(t *testing.T) {
	tests := []struct {
		name  string
		want  strm.Pipe
		wantW string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &bytes.Buffer{}
			if got := Dump(w); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Dump() = %v, want %v", got, tt.want)
			}
			if gotW := w.String(); gotW != tt.wantW {
				t.Errorf("Dump() = %v, want %v", gotW, tt.wantW)
			}
		})
	}
}
