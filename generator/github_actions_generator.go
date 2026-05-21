package generator

import (
  "strings"

  service "github.com/h3yng/drb99/services"
)

func (g *Generator) generateGithubActions(_ service.WrapperConfig) map[string]string {
	return map[string]string{
		".github/workflows/release.yml": strings.ReplaceAll(githubActionsTemplate, "__GITHUB_TOKEN__", "${{ secrets.GITHUB_TOKEN }}"),
	}
}

const githubActionsTemplate = `name: release

on:
  push:
    tags:
      - "v*"

permissions:
  contents: write

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Setup Go
        uses: actions/setup-go@v5
        with:
          go-version: stable

      - name: Run GoReleaser
        uses: goreleaser/goreleaser-action@v6
        with:
          version: latest
          args: release --clean
        env:
          GITHUB_TOKEN: __GITHUB_TOKEN__
`
