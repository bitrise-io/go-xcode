package ruby

import (
	"bytes"
	"path/filepath"
	"text/template"

	"github.com/bitrise-io/go-utils/fileutil"
)

// gemfileTemplate is a template to generate Gemfile.
const gemfileTemplate = `source 'https://rubygems.org'
{{range $gem, $version := .}}
gem '{{$gem}}'{{if ne $version ""}}, '~> {{$version}}'{{end}}
{{end}}
`

// createGemfileContent generates a Gemfile with the given gems.
// gems can be provided as a gem_name: gem_version,
// define empty string for version, to use the latest version of the gem.
func createGemfileContent(gems map[string]string) (string, error) {
	t := template.New("Gemfile template")
	t, err := t.Parse(gemfileTemplate)
	if err != nil {
		return "", err
	}

	var gemfileBuffer bytes.Buffer
	if err := t.Execute(&gemfileBuffer, gems); err != nil {
		return "", err
	}

	return gemfileBuffer.String(), nil
}

// createGemfile generates a Gemfile in the given directory including the given gems and returns its path.
// gems can be provided as a gem_name: gem_version,
// define empty string for version, to use the latest version of the gem.
func createGemfile(dir string, gems map[string]string) (string, error) {
	gemfileContent, err := createGemfileContent(gems)
	if err != nil {
		return "", err
	}

	gemfilePth := filepath.Join(dir, "Gemfile")
	if err := fileutil.WriteStringToFile(gemfilePth, gemfileContent); err != nil {
		return "", err
	}

	return gemfilePth, nil
}
