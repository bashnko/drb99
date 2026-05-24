package generator

import (
	"strings"

	service "github.com/h3yng/drb99/services"
)

func (g *Generator) generateNixFlake(cfg service.WrapperConfig) (map[string]string, error) {
	flakeNix, err := renderTemplate(nixFlakeTemplate, nixFlakeTemplateConfig(cfg))
	if err != nil {
		return nil, err
	}

	return map[string]string{
		"flake.nix": flakeNix,
	}, nil
}

type nixFlakeConfig struct {
	Description string
	BinaryName  string
	Version     string
}

func nixFlakeTemplateConfig(cfg service.WrapperConfig) nixFlakeConfig {
	description := strings.TrimSpace(cfg.Description)
	if description == "" {
		description = strings.TrimSpace(cfg.BinaryName + " CLI")
	}

	version := strings.TrimSpace(cfg.Version)
	if version == "" {
		version = "0.1.0"
	}
	version = strings.TrimPrefix(version, "v")

	return nixFlakeConfig{
		Description: description,
		BinaryName:  strings.TrimSpace(cfg.BinaryName),
		Version:     version,
	}
}

const nixFlakeTemplate = `{ 
  description = {{ printf "%q" .Description }};

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  };

  outputs = { self, nixpkgs }:
    let
      systems = [
        "x86_64-linux"
        "aarch64-linux"
        "x86_64-darwin"
        "aarch64-darwin"
      ];

      forAllSystems = f:
        nixpkgs.lib.genAttrs systems (
          system:
            f {
              pkgs = import nixpkgs { inherit system; };
            }
        );
    in {
      packages = forAllSystems ({ pkgs }: {
        default = pkgs.buildGoModule {
          pname = {{ printf "%q" .BinaryName }};
          version = {{ printf "%q" .Version }};
          src = ./.;
          vendorHash = null;

          subPackages = [ "cmd/{{ .BinaryName }}" ];

          ldflags = [
            "-s"
            "-w"
          ];

          meta = {
            mainProgram = {{ printf "%q" .BinaryName }};
          };
        };
      });

      apps = forAllSystems ({ pkgs }: {
        default = {
          type = "app";
          program = "${self.packages.${pkgs.system}.default}/bin/{{ .BinaryName }}";
        };
      });
    };
}
`
