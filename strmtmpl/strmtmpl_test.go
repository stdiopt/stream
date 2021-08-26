package strmtmpl

import (
	"errors"
	"testing"
	"text/template"

	strm "github.com/stdiopt/stream"
	"github.com/stdiopt/stream/strmtest"
)

func TestTemplate(t *testing.T) {
	type sample struct {
		M map[string]interface{}
		S []interface{}
		I int
	}
	tests := []struct {
		name        string
		pipe        strm.Pipe
		send        []interface{}
		want        []interface{}
		senderError error

		wantErrorRE string
	}{
		{
			name: "executes template",
			pipe: Template("{{.}}"),
			send: []interface{}{1},
			want: []interface{}{[]byte("1")},
		},
		{
			name: "executes template on complex type",
			pipe: Template("Map {{.M.key}}, Slice: {{index .S 0}} {{.I}}"),
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
			pipe: Template("{{myfunc}}", WithFuncs(template.FuncMap{
				"myfunc": func() string { return "hello test" },
			})),
			send: []interface{}{1},
			want: []interface{}{[]byte("hello test")},
		},
		{
			name: "escape function",
			pipe: Template("{{esc .}}"),
			send: []interface{}{"test\nstring"},
			want: []interface{}{[]byte(`"test\nstring"`)},
		},
		{
			name:        "returns error on missing field",
			pipe:        Template("{{.Missing.key}}"),
			send:        []interface{}{sample{}},
			wantErrorRE: "strmtmpl.Template.* can't evaluate field Missing in type strmtmpl.sample",
		},
		{
			name:        "returns error on malformed template",
			pipe:        Template("{{"),
			send:        []interface{}{sample{}},
			wantErrorRE: "strmtmpl.Template.* unclosed action",
		},
		{
			name:        "returns error when sender errors",
			pipe:        Template(""),
			send:        []interface{}{1, 2, 3},
			senderError: errors.New("sender error"),
			wantErrorRE: "strmtmpl.Template.* sender error$",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st := strmtest.New(t, tt.pipe)
			for _, s := range tt.send {
				st.Send(s).WithSenderError(tt.senderError)
			}
			st.ExpectFull(tt.want...).
				ExpectError(tt.wantErrorRE).
				Run()
		})
	}
}
