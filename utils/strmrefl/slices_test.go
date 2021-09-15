package strmrefl

import (
	"errors"
	"testing"

	"github.com/stdiopt/stream/utils/strmtest"
)

func TestUnslice(t *testing.T) {
	tests := []struct {
		name        string
		send        []interface{}
		senderError error
		want        []interface{}
		wantErr     string
		wantPanic   string
	}{
		{
			name: "unslice input",
			send: []interface{}{
				[]int{1, 2, 3},
				[]int{4, 5, 6},
			},
			want: []interface{}{
				1, 2, 3, 4, 5, 6,
			},
		},
		{
			name: "unslice different type input",
			send: []interface{}{
				[]int{1, 2, 3},
				[]string{"one", "two", "three"},
			},
			want: []interface{}{
				1, 2, 3, "one", "two", "three",
			},
		},
		{
			name: "sends value if not slice",
			send: []interface{}{1},
			want: []interface{}{1},
		},
		{
			name:        "return error when sender errors",
			send:        []interface{}{[]int{1, 2, 3}},
			senderError: errors.New("sender error"),
			wantErr:     "sender error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				p := recover()
				if !strmtest.MatchPanic(tt.wantPanic, p) {
					t.Errorf("Unslice() panic = %v, wantPanic %v", p, tt.wantPanic)
				}
			}()

			st := strmtest.New(t, Unslice())
			for _, s := range tt.send {
				st.Send(s).WithSenderError(tt.senderError)
			}
			st.ExpectFull(tt.want...).
				ExpectError(tt.wantErr).
				Run()
		})
	}
}

func TestSlice(t *testing.T) {
	type args struct {
		max int
	}
	tests := []struct {
		name        string
		args        args
		send        []interface{}
		senderError error
		want        []interface{}
		wantErr     string
		wantPanic   string
	}{
		{
			name: "slice input",
			args: args{0},
			send: []interface{}{1, 2, 3},
			want: []interface{}{[]int{1, 2, 3}},
		},
		{
			name: "group typed inputs",
			args: args{2},
			send: []interface{}{1, "1", 2, "2", "3", 3, 4},
			want: []interface{}{
				[]int{1, 2},
				[]string{"1", "2"},
				[]int{3, 4},
				[]string{"3"},
			},
		},
		{
			name:        "return error on sender error",
			args:        args{2},
			send:        []interface{}{1, 2, 3},
			senderError: errors.New("sender error"),
			wantErr:     "sender error",
		},
		{
			name:        "return error on final sender error",
			args:        args{4},
			send:        []interface{}{1, 2, 3},
			senderError: errors.New("sender error"),
			wantErr:     "sender error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				p := recover()
				if !strmtest.MatchPanic(tt.wantPanic, p) {
					t.Errorf("Slice() panic = %v, wantPanic %v", p, tt.wantPanic)
				}
			}()

			st := strmtest.New(t, Slice(tt.args.max))
			for _, s := range tt.send {
				st.Send(s).WithSenderError(tt.senderError)
			}
			st.ExpectFull(tt.want...).
				ExpectError(tt.wantErr).
				Run()
		})
	}
}
