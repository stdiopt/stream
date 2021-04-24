package strmutil

import (
	"bytes"
	"fmt"
	"text/template"
)

func Template(s string) ProcFunc {
	return func(p Proc) error {
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
		return p.Consume(func(v interface{}) error {
			buf := &bytes.Buffer{}
			if err := tmpl.Execute(buf, v); err != nil {
				return err
			}
			return p.Send(buf.Bytes())
		})
	}
}
