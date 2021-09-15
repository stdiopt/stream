package strmcsv

import (
	"errors"
	"testing"

	"github.com/stdiopt/stream/utils/strmtest"
)

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
				[]byte("hdr1,hdr2\nfield1,field2\n"),
			},
			want: []interface{}{
				[]string{"hdr1", "hdr2"},
				[]string{"field1", "field2"},
			},
		},
		{
			name: "decodes multiple bytes csv into []string",
			args: args{','},
			send: []interface{}{
				[]byte("hdr1,hdr2\nfield1.1,field1.2\n"),
				[]byte("field2.1,field2.2\n"),
			},
			want: []interface{}{
				[]string{"hdr1", "hdr2"},
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
				[]byte("hdr1,hdr2\n\"field1,field2\n"),
			},
			want: []interface{}{
				[]string{"hdr1", "hdr2"},
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
