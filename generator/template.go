package generator

import (
	"bytes"
	"text/template"

	service "github.com/h3yng/drb99/services"
)

func renderTemplate(raw string, cfg service.WrapperConfig) (string, error) {
	tpl, err := template.New("file").Parse(raw)
	if err != nil {
		return "", err
	}
	var out bytes.Buffer
	if err := tpl.Execute(&out, cfg); err != nil {
		return "", err
	}
	return out.String(), nil
}
