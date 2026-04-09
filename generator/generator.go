package generator

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"strings"

	service "github.com/bashnko/drb99/services"
)

type Generator struct{}

func New() *Generator {
	return &Generator{}
}

func (g *Generator) Generate(cfg service.WrapperConfig) (map[string]string, error) {
	files := map[string]string{}

	if cfg.Features.NPMWrapper {
		packageJSON, err := g.renderPackageJSON(cfg)
		if err != nil {
			return nil, err
		}

		installJS, err := renderTemplate(installTemplate, cfg)
		if err != nil {
			return nil, err
		}

		indexJS, err := renderTemplate(indexTemplate, cfg)
		if err != nil {
			return nil, err
		}
		readme, err := renderTemplate(readmeTemplate, cfg)
		if err != nil {
			return nil, err
		}
		files["package.json"] = packageJSON
		files["install.js"] = installJS
		files["index.js"] = indexJS
		files["README.md"] = readme
	}
	if cfg.Features.GoRealeser {
		goreleaserYAML, err := renderTemplate(goreleaserTemplate, cfg)
		if err != nil {
			return nil, err
		}
		files[".gorealeaser.yml"] = goreleaserYAML
	}
	if cfg.Features.GithubActions {
		files[".github/workflows/release.yml"] = strings.ReplaceAll(githubActionsTemplate, "__GITHUB_TOKEN__", "${{ secrets.GITHUB_TOKEN }}")
	}
	return files, nil

}

func (g *Generator) renderPackageJSON(cfg service.WrapperConfig) (string, error) {
	pkg := map[string]any{
		"name":        cfg.PackageName,
		"version":     cfg.NPMVersion,
		"description": fmt.Sprintf("npm wrapper for %s", cfg.BinaryName),
		"license":     "MIT",
		"bin": map[string]string{
			cfg.BinaryName: "index.js",
		},
		"scripts": map[string]string{
			"postinstall": "node install.js",
		},
		"files": []string{"index.js", "install.js", "bin/", "README.md"},
	}
	buf, err := json.MarshalIndent(pkg, "", "  ")
	if err != nil {
		return "", err
	}
	return string(buf) + "\n", nil
}

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

const installTemplate = `#!/usr/bin/env node
'use strict';

const fs = require('fs');
const path = require('path');
const https = require('https');
const zlib = require('zlib');

const binaryName = {{ printf "%q" .BinaryName }};
const targetDir = path.join(__dirname, 'bin');
const platformKey = process.platform + '-' + process.arch;

const assets = {
{{- range .Platforms }}
  {{ printf "%q" .NodeKey }}: {
    url: {{ printf "%q" .URL }},
    fileName: {{ printf "%q" .BinaryFile }},
    archive: {{ printf "%q" .Archive }}
  },
{{- end }}
};

function fail(message, details) {
  const extra = details ? '\n' + details : '';
  console.error('[esdrb] ' + message + extra);
  process.exit(1);
}

function ensureDir(dir) {
  fs.mkdirSync(dir, { recursive: true });
}

function extractZipEntry(zipPath, outputPath) {
  const data = fs.readFileSync(zipPath);
  const eocdSignature = 0x06054b50;
  const centralSignature = 0x02014b50;
  const localSignature = 0x04034b50;

  if (data.length < 22) {
    fail('Downloaded archive is too small to be a valid zip file.');
  }

  let eocdOffset = -1;
  for (let i = data.length - 22; i >= Math.max(0, data.length - 65557); i -= 1) {
    if (data.readUInt32LE(i) === eocdSignature) {
      eocdOffset = i;
      break;
    }
  }

  if (eocdOffset === -1) {
    fail('Downloaded archive is not a valid zip file.');
  }

  const centralDirectoryOffset = data.readUInt32LE(eocdOffset + 16);
  const totalEntries = data.readUInt16LE(eocdOffset + 10);
  let cursor = centralDirectoryOffset;
  let selected = null;

  for (let entry = 0; entry < totalEntries; entry += 1) {
    if (data.readUInt32LE(cursor) !== centralSignature) {
      fail('Invalid zip central directory entry.');
    }

    const compressionMethod = data.readUInt16LE(cursor + 10);
    const compressedSize = data.readUInt32LE(cursor + 20);
    const fileNameLength = data.readUInt16LE(cursor + 28);
    const extraLength = data.readUInt16LE(cursor + 30);
    const commentLength = data.readUInt16LE(cursor + 32);
    const localHeaderOffset = data.readUInt32LE(cursor + 42);
    const fileName = data.slice(cursor + 46, cursor + 46 + fileNameLength).toString('utf8');

    if (fileName && !fileName.endsWith('/')) {
      selected = {
        compressionMethod: compressionMethod,
        compressedSize: compressedSize,
        localHeaderOffset: localHeaderOffset,
        fileName: fileName,
      };
      break;
    }

    cursor += 46 + fileNameLength + extraLength + commentLength;
  }

  if (!selected) {
    fail('Zip archive does not contain a usable binary.');
  }

  if (data.readUInt32LE(selected.localHeaderOffset) !== localSignature) {
    fail('Invalid zip local header.');
  }

  const localNameLength = data.readUInt16LE(selected.localHeaderOffset + 26);
  const localExtraLength = data.readUInt16LE(selected.localHeaderOffset + 28);
  const dataStart = selected.localHeaderOffset + 30 + localNameLength + localExtraLength;
  const payload = data.slice(dataStart, dataStart + selected.compressedSize);
  let extracted;

  if (selected.compressionMethod === 0) {
    extracted = payload;
  } else if (selected.compressionMethod === 8) {
    extracted = zlib.inflateRawSync(payload);
  } else {
    fail('Unsupported zip compression method: ' + selected.compressionMethod);
  }

  fs.writeFileSync(outputPath, extracted);
}

function download(url, destination, redirects = 0) {
  if (redirects > 5) {
    fail('Too many redirects while downloading binary.', url);
  }

  https.get(url, (res) => {
    if (res.statusCode >= 300 && res.statusCode < 400 && res.headers.location) {
      return download(res.headers.location, destination, redirects + 1);
    }

    if (res.statusCode !== 200) {
      return fail('Failed to download release asset.', 'HTTP ' + res.statusCode + ' from ' + url);
    }

    const tmpFile = destination + '.tmp';
    const file = fs.createWriteStream(tmpFile);

    res.pipe(file);

    file.on('finish', () => {
      file.close(() => {
        try {
          if (assets[platformKey].archive === 'zip') {
            extractZipEntry(tmpFile, destination);
            fs.unlinkSync(tmpFile);
          } else {
            fs.renameSync(tmpFile, destination);
          }

          if (process.platform !== 'win32') {
            fs.chmodSync(destination, 0o755);
          }
          console.log('[esdrb] Installed ' + binaryName + ' for ' + platformKey);
        } catch (err) {
          try {
            fs.unlinkSync(tmpFile);
          } catch (_) {
          }
          fail('Unable to install downloaded binary.', err.message);
        }
      });
    });

    file.on('error', (err) => {
      try {
        fs.unlinkSync(tmpFile);
      } catch (_) {
      }
      fail('Unable to write downloaded binary.', err.message);
    });
  }).on('error', (err) => {
    fail('Network failure while downloading binary.', err.message);
  });
}

function main() {
  const target = assets[platformKey];
  if (!target) {
    const supported = Object.keys(assets).join(', ');
    fail('Unsupported platform/architecture.', 'Detected ' + platformKey + '. Supported: ' + supported);
  }

  ensureDir(targetDir);
  const outputName = process.platform === 'win32' ? binaryName + '.exe' : binaryName;
  const outputPath = path.join(targetDir, outputName);
  download(target.url, outputPath);
}

main();
`

const indexTemplate = `#!/usr/bin/env node
'use strict';

const path = require('path');
const fs = require('fs');
const { spawn } = require('child_process');

const binaryName = {{ printf "%q" .BinaryName }};
const executable = process.platform === 'win32' ? binaryName + '.exe' : binaryName;
const binaryPath = path.join(__dirname, 'bin', executable);

if (!fs.existsSync(binaryPath)) {
  console.error('[esdrb] Binary is missing. Reinstall the package to trigger postinstall.');
  process.exit(1);
}

const child = spawn(binaryPath, process.argv.slice(2), { stdio: 'inherit' });

child.on('error', (err) => {
  console.error('[esdrb] Failed to start binary:', err.message);
  process.exit(1);
});

child.on('exit', (code, signal) => {
  if (signal) {
    process.kill(process.pid, signal);
    return;
  }
  process.exit(code === null ? 1 : code);
});
`

const readmeTemplate = `# {{ .PackageName }}

npm wrapper for **{{ .BinaryName }}** from [{{ .RepoURL }}]({{ .RepoURL }}).

## Install

    npm install -g {{ .PackageName }}

## Usage

    {{ .BinaryName }} --help

## How it works

- During postinstall, the package downloads the matching release asset for your platform.
- The binary is stored in ./bin and exposed through npm bin.
- No Go toolchain is required on end-user machines.

## Included platform mappings

{{- range .Platforms }}
- {{ .NodeKey }} -> {{ .BinaryFile }}
{{- end }}

## Release source

- Version: {{ .Version }}
- Repository: {{ .RepoURL }}
`

const goreleaserTemplate = `version: 2

project_name: {{ .BinaryName }}

before:
  hooks:
    - go mod tidy

builds:
  - id: default
    binary: {{ .BinaryName }}
    env:
      - CGO_ENABLED=0
    ldflags:
      - -s -w
      - -X main.version={{ "{{" }} .Version {{ "}}" }}
      - -X main.commit={{ "{{" }} .Commit {{ "}}" }}
      - -X main.date={{ "{{" }} .Date {{ "}}" }}
    goos:
      - linux
      - darwin
      - windows
    goarch:
      - amd64
      - arm64

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
      - '^ci:'
      - '\\bdocs?\\b'
`

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
