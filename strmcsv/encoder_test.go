package strmcsv

import (
	"testing"

	"github.com/stdiopt/stream/strmtest"
)

func TestEncode(t *testing.T) {
	type sample struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pp := Encode(tt.args.comma)
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
