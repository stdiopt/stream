package strmtext

import (
	"errors"
	"reflect"
	"testing"
	"text/template"

	"github.com/stdiopt/stream/utils/strmtest"
)

func TestWithFuncs(t *testing.T) {
	type args struct {
		fm template.FuncMap
	}
	tests := []struct {
		name string
		args args
		want OptionFunc
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := WithTemplateFuncs(tt.args.fm); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("WithFuncs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTemplate(t *testing.T) {
	type sample struct {
		M map[string]interface{}
		S []interface{}
		I int
	}

	type args struct {
		s    string
		opts []OptionFunc
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
			name: "executes template",
			args: args{s: "{{.}}"},
			send: []interface{}{1},
			want: []interface{}{[]byte("1")},
		},
		{
			name: "executes template on complex type",
			args: args{s: "Map {{.M.key}}, Slice: {{index .S 0}} {{.I}}"},
			send: []interface{}{
				sample{
					M: map[string]interface{}{
						"key": "test1",
					},
					S: []interface{}{"1", 2, 3},
					I: 10,
				},
				sample{
					M: map[string]interface{}{
						"key": "test2",
					},
					S: []interface{}{"2", 2, 3},
					I: 20,
				},
			},
			want: []interface{}{
				[]byte("Map test1, Slice: 1 10"),
				[]byte("Map test2, Slice: 2 20"),
			},
		},
		{
			name: "custom functions",
			args: args{
				s: "{{myfunc}}",
				opts: []OptionFunc{
					WithTemplateFuncs(template.FuncMap{
						"myfunc": func() string { return "hello test" },
					}),
				},
			},
			send: []interface{}{1},
			want: []interface{}{[]byte("hello test")},
		},
		{
			name: "escape function",
			args: args{s: "{{esc .}}"},
			send: []interface{}{"test\nstring"},
			want: []interface{}{[]byte(`"test\nstring"`)},
		},
		{
			name:    "returns error on missing field",
			args:    args{s: "{{.Missing.key}}"},
			send:    []interface{}{sample{}},
			wantErr: "strmtmpl.Template.* can't evaluate field Missing in type strmtmpl.sample",
		},
		{
			name:      "panics error on malformed template",
			args:      args{s: "{{"},
			send:      []interface{}{sample{}},
			wantPanic: "unclosed action",
		},
		{
			name:        "returns error when sender errors",
			args:        args{s: ""},
			send:        []interface{}{1, 2, 3},
			senderError: errors.New("sender error"),
			wantErr:     "strmtmpl.Template.* sender error$",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				p := recover()
				if !strmtest.MatchPanic(tt.wantPanic, p) {
					t.Errorf("Template() panic = %v, wantPanic %v", p, tt.wantPanic)
				}
			}()
			pp := Template(tt.args.s, tt.args.opts...)
			if pp == nil {
				t.Errorf("Template() is nil = %v, want %v", pp == nil, false)
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
