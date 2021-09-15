package strmjson

import (
	"bytes"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stdiopt/stream/utils/strmtest"
)

func TestDecode(t *testing.T) {
	type sample struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}
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
			name: "decodes json []byte",
			args: args{},
			send: []interface{}{
				[]byte(`{"name":"test"}`),
			},
			want: []interface{}{
				map[string]interface{}{
					"name": "test",
				},
			},
		},
		{
			name: "decodes multiple incoming []byte's",
			args: args{},
			send: []interface{}{
				[]byte(`{"name"`),
				[]byte(`:"test"}`),
			},
			want: []interface{}{
				map[string]interface{}{
					"name": "test",
				},
			},
		},
		{
			name: "decodes simple value",
			args: args{},
			send: []interface{}{
				[]byte(`7`),
				[]byte(`7`),
			},
			want: []interface{}{
				float64(77),
			},
		},
		{
			name: "decodes into struct",
			args: args{sample{}},
			send: []interface{}{
				[]byte(`{"name":"test", "value":7}`),
			},
			want: []interface{}{
				sample{Name: "test", Value: 7},
			},
		},
		{
			name: "returns error when invalid json",
			args: args{sample{}},
			send: []interface{}{
				[]byte(`invalid json`),
			},
			wantErr: "strmjson.Decode.* invalid character",
		},
		{
			name: "returns error when sender errors",
			args: args{sample{}},
			send: []interface{}{
				[]byte(`{"name":"test", "value":7}`),
			},
			senderError: errors.New("sender error"),
			wantErr:     "strmjson.Decode.* sender error",
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
	type sample struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	tests := []struct {
		name        string
		send        []interface{}
		senderError error

		want    []interface{}
		wantErr string
	}{
		{
			name: "encodes json",
			send: []interface{}{"123"},
			want: []interface{}{[]byte("\"123\"\n")},
		},
		{
			name: "encodes struct into json",
			send: []interface{}{
				sample{Name: "test 1", Value: 7},
				sample{Name: "test 2", Value: 77},
			},
			want: []interface{}{
				[]byte(`{"name":"test 1","value":7}` + "\n"),
				[]byte(`{"name":"test 2","value":77}` + "\n"),
			},
		},
		{
			name: "returns error when sender errors",
			send: []interface{}{
				sample{Name: "test 1", Value: 7},
			},
			senderError: errors.New("sender error"),
			wantErr:     "sender error",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pp := Encode()
			if pp == nil {
				t.Errorf("Encode() is nil = %v, want %v", pp == nil, false)
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

func TestDump(t *testing.T) {
	type sample struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	tests := []struct {
		name        string
		send        []interface{}
		senderError error

		want       []interface{}
		wantErr    string
		wantOutput string
	}{
		{
			name: "dumps json into writer",
			send: []interface{}{
				sample{Name: "test", Value: 7},
			},
			want: []interface{}{
				sample{Name: "test", Value: 7},
			},
			wantOutput: "{\n  \"name\": \"test\",\n  \"value\": 7\n}\n",
		},
		{
			name: "returns error when encoder errors",
			send: []interface{}{
				func() {},
			},
			wantErr: "unsupported type",
		},
		{
			name: "returns error when sender errors",
			send: []interface{}{
				sample{Name: "test", Value: 7},
			},
			senderError: errors.New("sender error"),
			wantErr:     "sender error",
			wantOutput:  "{\n  \"name\": \"test\",\n  \"value\": 7\n}\n",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &bytes.Buffer{}
			pp := Dump(w)
			if pp == nil {
				t.Errorf("Dump() is nil = %v, want %v", pp == nil, false)
			}

			st := strmtest.New(t, pp)
			for _, s := range tt.send {
				st.Send(s).WithSenderError(tt.senderError)
			}
			st.ExpectFull(tt.want...).
				ExpectError(tt.wantErr).
				Run()

			if diff := cmp.Diff(w.String(), tt.wantOutput); diff != "" {
				t.Error(diff)
			}
		})
	}
}
