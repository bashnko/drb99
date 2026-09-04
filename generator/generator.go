package generator

import (
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
		for name, content := range npmFiles {
			files[name] = content
		}
	}

	if cfg.Features.GoReleaser {
		goreleaserFiles, err := g.generateGoReleaser(cfg)
		if err != nil {
			return nil, err
		}
		for name, content := range goreleaserFiles {
			files[name] = content
		}
	}

	if cfg.Features.GithubActions {
		githubActionsFiles := g.generateGithubActions(cfg)
		for name, content := range githubActionsFiles {
			files[name] = content
		}
	}

	if cfg.Features.AUR {
		aurFiles, err := g.generateAUR(cfg)
		if err != nil {
			return nil, err
		}
		for name, content := range aurFiles {
			files[name] = content
		}
	}

	if cfg.Features.NixFlake {
		nixFlakeFiles, err := g.generateNixFlake(cfg)
		if err != nil {
			return nil, err
		}
		for name, content := range nixFlakeFiles {
			files[name] = content
		}
	}

	if cfg.Features.DockerContainer {
		dockerFiles, err := g.generateDockerContainer(cfg)
		if err != nil {
			return nil, err
		}

		for name, content := range dockerFiles {
			files[name] = content
		}
	}

	if cfg.Features.Curl {
		curlFiles, err := g.generateCurl(cfg)
		if err != nil {
			return nil, err
		}

		for name, content := range curlFiles {
			files[name] = content
		}
	}

	return files, nil
}
