package strmtmpl

import (
	"bytes"
	"fmt"
	"text/template"

	strm "github.com/stdiopt/stream"
)

type OptionFunc func(t *template.Template) *template.Template

func WithFuncs(fm template.FuncMap) OptionFunc {
	return func(t *template.Template) *template.Template {
		return t.Funcs(fm)
	}
}

// Template receives any type and processes it through the template described
// on s and produces []byte
func Template(s string, opts ...OptionFunc) strm.Pipe {
	return strm.Func(func(p strm.Proc) error {
		tmpl := template.New("/")
		tmpl = tmpl.Option("missingkey=error")
		tmpl = tmpl.Funcs(template.FuncMap{
			"esc": func(s string) string {
				return fmt.Sprintf("%q", s)
			},
		})
		for _, fn := range opts {
			tmpl = fn(tmpl)
		}

		tmpl, err := tmpl.Parse(s)
		if err != nil {
			return err
		}
		return p.Consume(func(v interface{}) error {
			buf := &bytes.Buffer{}
			if err := tmpl.Execute(buf, v); err != nil {
				return err
			}
			return p.Send(buf.Bytes())
		})
	})
}
