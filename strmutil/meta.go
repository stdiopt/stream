package strmutil

import (
	"bytes"
	"text/template"

	"github.com/stdiopt/stream"
)

func TemplateTransform(p stream.Proc, s string) (string, error) {
	tmpl, err := template.New("strmutil.Transform").Parse(s)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, p.Meta()); err != nil {
		return "", err
	}
	return buf.String(), nil
}
