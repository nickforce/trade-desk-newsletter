package render

import (
	"bytes"
	"os"
	"text/template"
)

func Markdown(tmplPath string, data any) (string, error) {
	b, err := os.ReadFile(tmplPath)
	if err != nil {
		return "", err
	}
	t, err := template.New("md").Parse(string(b))
	if err != nil {
		return "", err
	}
	var out bytes.Buffer
	if err := t.Execute(&out, data); err != nil {
		return "", err
	}
	return out.String(), nil
}
