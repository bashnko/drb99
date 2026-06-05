package generator

import service "github.com/h3yng/drb99/services"

func (g *Generator) generateGoReleaser(cfg service.WrapperConfig) (map[string]string, error) {
	goreleaserYAML, err := renderTemplate(goreleaserTemplate, cfg)
	if err != nil {
		return nil, err
	}

	return map[string]string{
		".goreleaser.yml": goreleaserYAML,
	}, nil
}

const goreleaserTemplate = `version: 2

project_name: {{ .BinaryName }}

before:
  hooks:
    - go mod tidy

builds:
  - id: {{ .BinaryName }}
	main: ./cmd/{{ .BinaryName }}/main.go # replace this path with your main.go
    binary: {{ .BinaryName }}
    env:
      - CGO_ENABLED=0
    ldflags:
      - -s -w
      - -X main.version={{ "{{" }} .Version {{ "}}" }}
      - -X main.commit={{ "{{" }} .Commit {{ "}}" }}
      - -X main.date={{ "{{" }} .Date {{ "}}" }}
    targets:
{{- range .GoReleaserTargets }}
      - {{ . }}
{{- end }}

archives:
  - id: default
    formats:
      - binary
    format_overrides:
      - goos: windows
        formats:
          - zip
    name_template: >-
      {{ "{{" }} .Binary {{ "}}" }}_{{ "{{" }} .Version {{ "}}" }}_{{ "{{" }} if eq .Os "darwin" {{ "}}" }}macos{{ "{{" }} else {{ "}}" }}{{ "{{" }} .Os {{ "}}" }}{{ "{{" }} end {{ "}}" }}_{{ "{{" }} .Arch {{ "}}" }}
    wrap_in_directory: false

checksum:
  name_template: "checksums.txt"

changelog:
  sort: asc
  filters:
    exclude:
      - '^docs:'
      - '^test:'
      - '^tests:'
      - '^chore:'
      - '^ci:'
      - '\\bdocs?\\b'

release:
	github:
		owner: {{ .Owner }}
		name: {{ .BinaryName }}
	draft: false
	prerelease: auto
`
