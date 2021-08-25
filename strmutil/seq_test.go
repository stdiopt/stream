package strmutil

import (
	"errors"
	"testing"

	strm "github.com/stdiopt/stream"
	"github.com/stdiopt/stream/strmtest"
)

func TestSeq(t *testing.T) {
	type send struct {
		value       interface{}
		senderError error
		want        []interface{}
		wantErrorRE string
	}
	tests := []struct {
		name        string
		pipe        strm.Pipe
		sends       []send
		wantErrorRE string
	}{
		{
			name: "produces sequence",
			pipe: Seq(0, 5, 1),
			sends: []send{
				{value: 1, want: []interface{}{0, 1, 2, 3, 4, 5}},
			},
		},
		{
			name: "produces sequence when start > end and step < 0",
			pipe: Seq(5, 0, -1),
			sends: []send{
				{value: 1, want: []interface{}{5, 4, 3, 2, 1, 0}},
			},
		},
		{
			name: "produces sequence with larget steps",
			pipe: Seq(0, 5, 2),
			sends: []send{
				{value: 1, want: []interface{}{0, 2, 4}},
			},
		},
		{
			name: "produces sequence per value received",
			pipe: Seq(0, 5, 1),
			sends: []send{
				{value: 1, want: []interface{}{0, 1, 2, 3, 4, 5}},
				{value: 1, want: []interface{}{0, 1, 2, 3, 4, 5}},
			},
		},
		{
			name: "returns error if sequence start is bigger than end with a positive step",
			pipe: Seq(5, 1, 1),
			sends: []send{
				{value: 1, want: []interface{}{0, 1, 2, 3, 4}},
			},
			wantErrorRE: "strmutil.Seq.* invalid range: 5->1, step: 1 causes infinite loop$",
		},
		{
			name: "returns error if sequence step is 0",
			pipe: Seq(0, 5, 0),
			sends: []send{
				{value: 1, want: []interface{}{0, 1, 2, 3, 4}},
			},
			wantErrorRE: "strmutil.Seq.* invalid range: 0->5, step: 0 causes infinite loop$",
		},
		{
			name: "returns error if sender errors",
			pipe: Seq(0, 5, 1),
			sends: []send{
				{value: 1, senderError: errors.New("sender error")},
			},
			wantErrorRE: "strmutil.Seq.* sender error$",
		},
		{
			name: "returns error if sender errors reverse",
			pipe: Seq(5, 0, -2),
			sends: []send{
				{value: 1, senderError: errors.New("sender error")},
			},
			wantErrorRE: "strmutil.Seq.* sender error$",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := strmtest.New(t, tt.pipe)
			for _, s := range tt.sends {
				st.Send(s.value).
					WithSenderError(s.senderError).
					Expect(s.want...)
			}
			st.ExpectError(tt.wantErrorRE).Run()
		})
	}
}
