package strmutil

import (
	"errors"
	"testing"

	"github.com/stdiopt/stream/strmtest"
)

func TestSeq(t *testing.T) {
	type send struct {
		value       interface{}
		senderError error
		want        []interface{}
		wantErr     string
	}
	type args struct {
		start int
		end   int
		step  int
	}
	tests := []struct {
		name      string
		args      args
		sends     []send
		wantErr   string
		wantPanic string
	}{
		{
			name: "produces sequence",
			args: args{0, 5, 1},
			sends: []send{
				{value: 1, want: []interface{}{0, 1, 2, 3, 4, 5}},
			},
		},
		{
			name: "produces sequence when start > end and step < 0",
			args: args{5, 0, -1},
			sends: []send{
				{value: 1, want: []interface{}{5, 4, 3, 2, 1, 0}},
			},
		},
		{
			name: "produces sequence with larget steps",
			args: args{0, 5, 2},
			sends: []send{
				{value: 1, want: []interface{}{0, 2, 4}},
			},
		},
		{
			name: "produces sequence per value received",
			args: args{0, 5, 1},
			sends: []send{
				{value: 1, want: []interface{}{0, 1, 2, 3, 4, 5}},
				{value: 1, want: []interface{}{0, 1, 2, 3, 4, 5}},
			},
		},
		{
			name: "panics if sequence start is bigger than end with a positive step",
			args: args{5, 1, 1},
			sends: []send{
				{value: 1, want: []interface{}{0, 1, 2, 3, 4}},
			},
			wantPanic: "invalid range: 5->1, step: 1 causes infinite loop$",
		},
		{
			name: "panics error if sequence step is 0",
			args: args{0, 5, 0},
			sends: []send{
				{value: 1, want: []interface{}{0, 1, 2, 3, 4}},
			},
			wantPanic: "invalid range: 0->5, step: 0 causes infinite loop$",
		},
		{
			name: "returns error if sender errors",
			args: args{0, 5, 1},
			sends: []send{
				{value: 1, senderError: errors.New("sender error")},
			},
			wantErr: "strmutil.Seq.* sender error$",
		},
		{
			name: "returns error if sender errors reverse",
			args: args{5, 0, -2},
			sends: []send{
				{value: 1, senderError: errors.New("sender error")},
			},
			wantErr: "strmutil.Seq.* sender error$",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				p := recover()
				if !strmtest.MatchPanic(tt.wantPanic, p) {
					t.Errorf("Seq() panic = %v, wantPanic %v", p, tt.wantPanic)
				}
			}()
			pp := Seq(tt.args.start, tt.args.end, tt.args.step)
			if pp == nil {
				t.Errorf("Seq() is nil = %v, want %v", pp == nil, false)
			}

			st := strmtest.New(t, pp)
			for _, s := range tt.sends {
				st.Send(s.value).
					WithSenderError(s.senderError).
					Expect(s.want...)
			}
			st.ExpectError(tt.wantErr).Run()
		})
	}
}
