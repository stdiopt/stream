package strmio

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	strm "github.com/stdiopt/stream"
	"github.com/stdiopt/stream/strmtest"
)

func TestReader(t *testing.T) {
	type args struct {
		r io.Reader
	}
	tests := []struct {
		name        string
		args        args
		send        []interface{}
		senderError error

		want      []interface{}
		wantErr   string
		wantPanic string
	}{
		{
			name: "reads from reader",
			args: args{strings.NewReader("test")},
			send: []interface{}{7},

			want: []interface{}{
				[]byte(`test`),
			},
		},
		{
			name: "reads only on first message",
			args: args{strings.NewReader("test 2")},
			send: []interface{}{7, 7},

			want: []interface{}{
				[]byte(`test 2`),
			},
		},
		{
			name:        "returns errors when sender errors",
			args:        args{strings.NewReader("test 3")},
			send:        []interface{}{7},
			senderError: errors.New("sender error"),

			wantErr: "strmio.Reader.* sender error",
		},
		{
			name: "panics when reader is nil",
			args: args{},
			send: []interface{}{7},

			wantPanic: "reader is nil",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				p := recover()
				if !strmtest.MatchPanic(tt.wantPanic, p) {
					t.Errorf("Reader() panic = %v, wantPanic %v", p, tt.wantPanic)
				}
			}()

			pp := Reader(tt.args.r)
			if pp == nil {
				t.Errorf("Reader() is nil = %v, want %v", pp == nil, false)
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

func TestWriter(t *testing.T) {
	type args struct {
		w io.Writer
	}
	tests := []struct {
		name        string
		args        *args
		send        []interface{}
		senderError error

		want       []interface{}
		wantErr    string
		wantPanic  string
		wantOutput string
	}{
		{
			name: "writes to writer",
			send: []interface{}{
				[]byte(`test`),
			},
			want: []interface{}{
				[]byte(`test`),
			},
			wantOutput: "test",
		},
		{
			name: "writes multiple to writer",
			send: []interface{}{
				[]byte(`test`),
			},
			want: []interface{}{
				[]byte(`test`),
			},
			wantOutput: "test",
		},
		{
			name: "panics if writer is nil",
			args: &args{},
			send: []interface{}{
				[]byte(`test`),
			},
			wantPanic: "writer is nil",
		},
		{
			name: "returns error when sender errors",
			send: []interface{}{
				[]byte(`test`),
			},
			senderError: errors.New("sender error"),
			wantOutput:  "test",
			wantErr:     "sender error",
		},
		{
			name: "returns error when writer errors",
			args: &args{testWriter{errors.New("writer error")}},
			send: []interface{}{
				[]byte(`test`),
			},
			wantOutput: "",
			wantErr:    "writer error",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				p := recover()
				if !strmtest.MatchPanic(tt.wantPanic, p) {
					t.Errorf("Writer() panic = %v, wantPanic %v", p, tt.wantPanic)
				}
			}()
			var w io.Writer
			var buf *bytes.Buffer
			if tt.args == nil {
				buf = &bytes.Buffer{}
				w = buf
			} else {
				w = tt.args.w
			}
			pp := Writer(w)
			if pp == nil {
				t.Errorf("Writer() is nil = %v, want %v", pp == nil, false)
			}

			st := strmtest.New(t, pp)
			for _, s := range tt.send {
				st.Send(s).WithSenderError(tt.senderError)
			}
			st.ExpectFull(tt.want...).
				ExpectError(tt.wantErr).
				Run()

			if buf != nil {
				if diff := cmp.Diff(buf.String(), tt.wantOutput); diff != "" {
					t.Error("Writer() wantOutput:\n", diff)
				}
			}
		})
	}
}

func TestAsReader(t *testing.T) {
	tests := []struct {
		name string

		send          []interface{}
		consumerError error

		wantErr  string
		wantRead string
	}{
		{
			name:     "reads a proc",
			send:     []interface{}{[]byte(`test 1`)},
			wantRead: "test 1",
		},
		{
			name:          "returns error when consumer errors",
			send:          []interface{}{[]byte(`test 1`)},
			consumerError: errors.New("consumer error"),
			wantErr:       "consumer error",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o := strm.Override{
				ConsumeFunc: func(fn strm.ConsumerFunc) error {
					if tt.consumerError != nil {
						return tt.consumerError
					}
					for _, s := range tt.send {
						fn(s)
					}
					return nil
				},
			}
			rd := AsReader(o)
			if rd == nil {
				t.Errorf("AsReader() is nil = %v, want %v", rd == nil, false)
			}
			buf := bytes.Buffer{}
			if _, err := io.Copy(&buf, rd); !strmtest.MatchError(tt.wantErr, err) {
				t.Errorf("AsReader() error = %v, wantErr %v", err, tt.wantErr)
			}
			if diff := cmp.Diff(buf.String(), tt.wantRead); diff != "" {
				t.Error("AsReader() wantOutput:\n", diff)
			}
		})
	}
}

func TestAsWriter(t *testing.T) {
	tests := []struct {
		name string

		write       []byte
		senderError error

		want     []interface{}
		wantErr  string
		wantRead string
	}{
		{
			name:  "writes into a sender",
			write: []byte(`test 1`),

			want: []interface{}{
				[]byte(`test 1`),
			},
		},
		{
			name:        "returns error when sender errors",
			write:       []byte(`test 2`),
			senderError: errors.New("sender error"),
			wantErr:     "sender error",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got []interface{}
			o := strm.Override{
				SendFunc: func(v interface{}) error {
					if tt.senderError != nil {
						return tt.senderError
					}
					got = append(got, v)
					return nil
				},
			}
			wr := AsWriter(o)
			if _, err := wr.Write(tt.write); !strmtest.MatchError(tt.wantErr, err) {
				t.Errorf("AsWriter() error = %v, wantErr %v", err, tt.wantErr)
			}

			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Error("AsWriter() wrong full any output\n- want + got\n", diff)
			}
		})
	}
}

type testWriter struct {
	err error
}

func (w testWriter) Write([]byte) (int, error) {
	return 0, w.err
}
