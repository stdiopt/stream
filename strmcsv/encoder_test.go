package strmcsv

import (
	"errors"
	"testing"

	"github.com/stdiopt/stream/strmtest"
)

func TestEncode(t *testing.T) {
	type sample struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	type args struct {
		comma      rune
		encodeOpts []encoderOpt
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
			name: "encodes csv []string to []byte",
			args: args{comma: ','},
			send: []interface{}{
				[]string{"one", "two"},
			},
			want: []interface{}{
				[]byte(`one,two` + "\n"),
			},
		},
		{
			name: "encodes csv []string to []byte",
			args: args{comma: ';'},
			send: []interface{}{
				[]string{"one", "two"},
				[]string{"three", "four"},
			},
			want: []interface{}{
				[]byte(
					`one;two` + "\n" +
						`three;four` + "\n",
				),
			},
		},
		{
			name: "encodes struct into csv []byte",
			args: args{
				comma: ',',
				encodeOpts: []encoderOpt{
					Field("header 1", "Name"),
					Field("header 2", "Value"),
				},
			},
			send: []interface{}{
				sample{"name 1", 7},
				sample{"name 2", 77},
			},
			want: []interface{}{
				[]byte(
					`header 1,header 2` + "\n" +
						`name 1,7` + "\n" +
						`name 2,77` + "\n",
				),
			},
		},
		{
			name: "encodes struct into csv []byte",
			args: args{
				comma: ',',
				encodeOpts: []encoderOpt{
					Field("Name"),
				},
			},
			send: []interface{}{
				sample{"name 1", 7},
				sample{"name 2", 77},
			},
			want: []interface{}{
				[]byte(
					`Name` + "\n" +
						`name 1` + "\n" +
						`name 2` + "\n",
				),
			},
		},
		{
			name: "encodes all struct fields when no col mapping",
			args: args{
				comma: ',',
			},
			send: []interface{}{
				sample{"name 1", 7},
			},
			want: []interface{}{
				[]byte(
					`Name,Value` + "\n" +
						`name 1,7` + "\n",
				),
			},
		},
		{
			name: "encodes all ptr struct fields when no col mapping",
			args: args{
				comma: ',',
			},
			send: []interface{}{
				&sample{"name 1", 7},
			},
			want: []interface{}{
				[]byte(
					`Name,Value` + "\n" +
						`name 1,7` + "\n",
				),
			},
		},
		{
			name: "returns error when no columns are defined and it's not struct",
			args: args{
				comma: ',',
			},
			send:    []interface{}{1},
			wantErr: "strmcsv.Encode.* invalid input type",
		},
		{
			name: "returns error when no columns are defined, and there is no exported fields",
			args: args{
				comma: ',',
			},
			send: []interface{}{
				struct{ s string }{},
			},
			wantErr: "strmcsv.Encode.* invalid input type",
		},
		{
			name: "returns error on invalid field",
			args: args{
				comma: ',',
				encodeOpts: []encoderOpt{
					Field("header 1", "Invalid"),
				},
			},
			send: []interface{}{
				sample{"name 1", 7},
			},
			want: []interface{}{
				[]byte(`header 1` + "\n"),
			},
			wantErr: "strmcsv.Encode.* field invalid",
		},
		{
			name: "errors when sender errors",
			args: args{
				comma: ',',
				encodeOpts: []encoderOpt{
					Field("header 1", "Name"),
					Field("header 2", "Value"),
				},
			},
			send: []interface{}{
				sample{"name 1", 7},
				sample{"name 2", 77},
			},
			senderError: errors.New("sender error"),
			wantErr:     "strmcsv.Encode.* sender error",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pp := Encode(tt.args.comma, tt.args.encodeOpts...)
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
