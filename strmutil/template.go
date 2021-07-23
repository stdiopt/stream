package strmutil

import (
	"bytes"
	"context"
	"fmt"
	"text/template"

	"github.com/stdiopt/stream"
)

// Template receives any type and processes it through the template described
// on s and produces []byte
func Template(s string) stream.Processor {
	return stream.Func(func(p stream.Proc) error {
		tmpl := template.New("/")
		tmpl = tmpl.Option("missingkey=error")
		tmpl = tmpl.Funcs(template.FuncMap{
			"esc": func(s string) string {
				return fmt.Sprintf("%q", s)
			},
		})
		tmpl, err := tmpl.Parse(s)
		if err != nil {
			return err
		}
		return p.Consume(func(ctx context.Context, v interface{}) error {
			buf := &bytes.Buffer{}
			if err := tmpl.Execute(buf, v); err != nil {
				return err
			}
			return p.Send(ctx, buf.Bytes())
		})
	})
}
