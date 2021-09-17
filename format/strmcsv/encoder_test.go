package strmcsv

import (
	"testing"

	"github.com/stdiopt/stream/utils/strmtest"
)

func TestEncode(t *testing.T) {
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
			name: "returns error when no columns are defined and it's not struct",
			args: args{
				comma: ',',
			},
			send:    []interface{}{1},
			wantErr: "strmcsv.Encode.* type int is not supported",
		},
		{
			name: "returns error when no columns are defined, and there is no exported fields",
			args: args{
				comma: ',',
			},
			send: []interface{}{
				struct{ s string }{},
			},
			wantErr: "strmcsv.Encode.* type struct \\{ s string \\} is not supported",
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
