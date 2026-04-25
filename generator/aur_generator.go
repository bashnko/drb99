package generator

import (
	"strings"

	service "github.com/bashnko/drb99/services"
)

func (g *Generator) generateAUR(cfg service.WrapperConfig) (map[string]string, error) {
	pkgbuild, err := renderTemplate(aurPKGBUILDTemplate, cfg)
	if err != nil {
		return nil, err
	}

	return map[string]string{
		"PKGBUILD":                  pkgbuild,
		".github/workflows/aur.yml": strings.ReplaceAll(aurWorkflowTemplate, "__GITHUB_TOKEN__", "${{ secrets.GITHUB_TOKEN }}"),
	}, nil
}

const aurPKGBUILDTemplate = `pkgname={{ .BinaryName }}
pkgver={{ .AURVersion }}
pkgrel=1
pkgdesc="{{ .Description }}"
arch=('x86_64' 'aarch64')
url="{{ .RepoURL }}"
license=('{{ .License }}')
depends=()
makedepends=('go')
source=("${pkgname}-${pkgver}.tar.gz::https://github.com/{{ .Owner }}/{{ .Repo }}/archive/refs/tags/v${pkgver}.tar.gz")
sha256sums=('SKIP')

build() {
  cd "${srcdir}/{{ .Repo }}-v${pkgver}"
  go build -trimpath -ldflags "-s -w" -o "${pkgname}" "./cmd/{{ .BinaryName }}"
}

package() {
  install -Dm755 "${srcdir}/{{ .Repo }}-v${pkgver}/${pkgname}" "${pkgdir}/usr/bin/${pkgname}"
}
`

const aurWorkflowTemplate = `name: aur-pkgbuild-update

on:
  push:
    tags:
      - "v*"

permissions:
  contents: write

jobs:
  update-pkgbuild:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Configure Git
        run: |
          git config user.name "github-actions[bot]"
          git config user.email "41898282+github-actions[bot]@users.noreply.github.com"

      - name: Update PKGBUILD version and checksum
        env:
          TAG: ${{ github.ref_name }}
          OWNER: {{ .Owner }}
          REPO: {{ .Repo }}
        run: |
          version="${TAG#v}"
          source_url="https://github.com/${OWNER}/${REPO}/archive/refs/tags/${TAG}.tar.gz"
          sha256="$(curl -fsSL "${source_url}" | sha256sum | awk '{print $1}')"

          sed -i "s/^pkgver=.*/pkgver=${version}/" PKGBUILD
          sed -i "s|^sha256sums=.*|sha256sums=('${sha256}')|" PKGBUILD

      - name: Commit PKGBUILD update
        env:
          DEFAULT_BRANCH: ${{ github.event.repository.default_branch }}
        run: |
          if git diff --quiet PKGBUILD; then
            echo "No PKGBUILD changes detected"
            exit 0
          fi

          git add PKGBUILD
          git commit -m "chore(aur): bump PKGBUILD to ${GITHUB_REF_NAME}"
          git push origin "HEAD:${DEFAULT_BRANCH}"

      - name: Optional AUR publish hint
        run: |
          echo "PKGBUILD updated in repo. Add your own AUR push step if you mirror to aur.archlinux.org."
`
