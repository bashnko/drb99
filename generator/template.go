package generator

import (
	"bytes"
	"text/template"
)

func renderTemplate(raw string, data any) (string, error) {
	tpl, err := template.New("file").Parse(raw)
	if err != nil {
		return "", err
	}
	var out bytes.Buffer
	if err := tpl.Execute(&out, data); err != nil {
		return "", err
	}
	return out.String(), nil
}
