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
		encodeOpts []encodeOpt
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
				encodeOpts: []encodeOpt{
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
			name: "returns error on invalid field",
			args: args{
				comma: ',',
				encodeOpts: []encodeOpt{
					Field("header 1", "Invalid"),
				},
			},
			send: []interface{}{
				sample{"name 1", 7},
			},
			want: []interface{}{
				[]byte(`header 1` + "\n"),
			},
			wantErr: "strmcsv.Encode.* struct: field invalid",
		},
		{
			name: "errors when sending a struct with no col mapping",
			args: args{
				comma: ',',
			},
			send: []interface{}{
				sample{"name 1", 7},
			},
			wantErr: "strmcsv.Encode.* invalid input type",
		},
		{
			name: "errors when sender errors",
			args: args{
				comma: ',',
				encodeOpts: []encodeOpt{
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

func TestDecode(t *testing.T) {
	type args struct {
		comma rune
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
			name: "ends gracefully on first EOF",
			args: args{','},
		},
		{
			name: "decodes csv into []string",
			args: args{','},
			send: []interface{}{
				[]byte("hdr1, hdr2\nfield1,field2\n"),
			},
			want: []interface{}{
				[]string{"field1", "field2"},
			},
		},
		{
			name: "decodes csv into []string",
			args: args{','},
			send: []interface{}{
				[]byte("hdr1, hdr2\nfield1.1,field1.2\n"),
				[]byte("field2.1,field2.2\n"),
			},
			want: []interface{}{
				[]string{"field1.1", "field1.2"},
				[]string{"field2.1", "field2.2"},
			},
		},
		{
			name: "returns error on header error",
			args: args{','},
			send: []interface{}{
				[]byte("\"hdr1, hdr2\nfield1,field2\n"),
			},
			wantErr: "strmcsv.Decode.* record on line 1",
		},
		{
			name: "returns error on row error",
			args: args{','},
			send: []interface{}{
				[]byte("hdr1, hdr2\n\"field1,field2\n"),
			},
			wantErr: "strmcsv.Decode.* parse error on line 2",
		},
		{
			name: "returns error on sender error",
			args: args{','},
			send: []interface{}{
				[]byte("hdr1, hdr2\nfield1,field2\n"),
			},
			senderError: errors.New("sender error"),
			wantErr:     "strmcsv.Decode.* sender error",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pp := Decode(tt.args.comma)
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

func TestDecodeAsJSON(t *testing.T) {
	type args struct {
		comma rune
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
			name: "eof on header read",
			args: args{','},
		},
		{
			name: "decodes as json",
			args: args{','},
			send: []interface{}{
				[]byte("hdr1,hdr2\nfield1.1,field1.2\n"),
				[]byte("field2.1,field2.2\n"),
			},
			want: []interface{}{
				[]byte(`{"hdr1":"field1.1","hdr2":"field1.2"}` + "\n"),
				[]byte(`{"hdr1":"field2.1","hdr2":"field2.2"}` + "\n"),
			},
		},
		{
			name: "decodes as json",
			args: args{','},
			send: []interface{}{
				[]byte("hdr1,hdr2\nfield1.1,field1.2\n"),
				[]byte("field2.1,field2.2\n"),
			},
			want: []interface{}{
				[]byte(`{"hdr1":"field1.1","hdr2":"field1.2"}` + "\n"),
				[]byte(`{"hdr1":"field2.1","hdr2":"field2.2"}` + "\n"),
			},
		},
		{
			name: "returns error on header error",
			args: args{','},
			send: []interface{}{
				[]byte("\"hdr1, hdr2\nfield1,field2\n"),
			},
			wantErr: "strmcsv.Decode.* record on line 1",
		},
		{
			name: "returns error on row error",
			args: args{','},
			send: []interface{}{
				[]byte("hdr1, hdr2\n\"field1,field2\n"),
			},
			wantErr: "strmcsv.Decode.* parse error on line 2",
		},
		{
			name: "returns error on sender error",
			args: args{','},
			send: []interface{}{
				[]byte("hdr1, hdr2\nfield1,field2\n"),
			},
			senderError: errors.New("sender error"),
			wantErr:     "strmcsv.Decode.* sender error",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pp := DecodeAsJSON(tt.args.comma)
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
