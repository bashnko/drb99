package generator

import (
	service "github.com/h3yng/drb99/services"
)

func (g *Generator) generateDockerContainer(cfg service.WrapperConfig) (map[string]string, error) {
	dockerfile, err := renderTemplate(dockerfileTemplate, cfg)
	if err != nil {
		return nil, err
	}

	gitignore, err := renderTemplate(gitignoreFileTemplate, cfg)
	if err != nil {
		return nil, err
	}

	return map[string]string{
		"Dockerfile": dockerfile,
		".gitignore": gitignore,
	}, nil
}

const dockerfileTemplate = `# This Dockerfile was generated using drb99
# https://github.com/h3yng/drb99.git
From  {{ .RuntimeImage}} 

WORKDIR /app

COPY . .

RUN go build -o {{ .BinaryName }}
CMD [ "./{{ .BinaryName }}" ]

`

const gitignoreFileTemplate = `ignore paths`
