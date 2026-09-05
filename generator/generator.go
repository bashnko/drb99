package generator

import (
	"maps"

	service "github.com/h3yng/drb99/services"
)

type Generator struct{}

func New() *Generator {
	return &Generator{}
}

func (g *Generator) Generate(cfg service.WrapperConfig) (map[string]string, error) {
	files := map[string]string{}

	if cfg.Features.NPMWrapper {
		npmFiles, err := g.generateNPMWrapper(cfg)
		if err != nil {
			return nil, err
		}
		maps.Copy(files, npmFiles)
	}

	if cfg.Features.GoReleaser {
		goreleaserFiles, err := g.generateGoReleaser(cfg)
		if err != nil {
			return nil, err
		}
		maps.Copy(files, goreleaserFiles)
	}

	if cfg.Features.GithubActions {
		githubActionsFiles := g.generateGithubActions(cfg)
		maps.Copy(files, githubActionsFiles)
	}

	if cfg.Features.AUR {
		aurFiles, err := g.generateAUR(cfg)
		if err != nil {
			return nil, err
		}
		maps.Copy(files, aurFiles)
	}

	if cfg.Features.NixFlake {
		nixFlakeFiles, err := g.generateNixFlake(cfg)
		if err != nil {
			return nil, err
		}
		maps.Copy(files, nixFlakeFiles)
	}

	if cfg.Features.DockerContainer {
		dockerFiles, err := g.generateDockerContainer(cfg)
		if err != nil {
			return nil, err
		}

		maps.Copy(files, dockerFiles)
	}

	if cfg.Features.Curl {
		curlFiles, err := g.generateCurl(cfg)
		if err != nil {
			return nil, err
		}

		maps.Copy(files, curlFiles)
	}

	if cfg.Features.Iex {
		iexFiles, err := g.generateIex(cfg)
		if err != nil {
			return nil, err
		}

		maps.Copy(files, iexFiles)
	}

	return files, nil
}
