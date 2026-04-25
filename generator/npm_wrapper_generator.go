package generator

import (
	"encoding/json"

	service "github.com/bashnko/drb99/services"
)

func (g *Generator) generateNPMWrapper(cfg service.WrapperConfig) (map[string]string, error) {
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

	return map[string]string{
		"package.json": packageJSON,
		"install.js":   installJS,
		"index.js":     indexJS,
		"README.md":    readme,
	}, nil
}

func (g *Generator) renderPackageJSON(cfg service.WrapperConfig) (string, error) {
	pkg := map[string]any{
		"name":        cfg.PackageName,
		"version":     cfg.NPMVersion,
		"description": cfg.Description,
		"license":     cfg.License,
		"author":      cfg.Author,
		"bin": map[string]string{
			cfg.BinaryName: "index.js",
		},
		"dependencies": map[string]string{
			"tar": "^7.5.2",
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

const installTemplate = `#!/usr/bin/env node
'use strict';

const fs = require('fs');
const path = require('path');
const https = require('https');
const tar = require('tar');
const zlib = require('zlib');

const binaryName = {{ printf "%q" .BinaryName }};
const targetDir = path.join(__dirname, 'bin');
const platformKey = process.platform + '-' + process.arch;

const assets = {
{{- range .Platforms }}
  {{ printf "%q" .NodeKey }}: {
    urls: [{{ range .URLs }}{{ printf "%q" . }},{{ end }}],
    fileName: {{ printf "%q" .BinaryFile }},
    archive: {{ printf "%q" .Archive }}
  },
{{- end }}
};

function fail(message, details) {
  const extra = details ? '\n' + details : '';
  console.error('[drb99] ' + message + extra);
  process.exit(1);
}

function ensureDir(dir) {
  fs.mkdirSync(dir, { recursive: true });
}

function findBinaryFile(dir) {
  const expectedNames = new Set([binaryName, binaryName + '.exe']);
  const entries = fs.readdirSync(dir, { withFileTypes: true }).sort((a, b) => a.name.localeCompare(b.name));

  for (const entry of entries) {
    const entryPath = path.join(dir, entry.name);
    if (entry.isFile()) {
      if (expectedNames.has(entry.name)) {
        return entryPath;
      }
      continue;
    }
    if (entry.isDirectory()) {
      const nested = findBinaryFile(entryPath);
      if (nested) {
        return nested;
      }
    }
  }

  return null;
}

async function extractTarGzEntry(archivePath, outputPath) {
  try {
    const extractDir = fs.mkdtempSync(path.join(targetDir, 'tar-'));
    try {
      await tar.x({
        file: archivePath,
        cwd: extractDir,
        gzip: true,
      });

      const extracted = findBinaryFile(extractDir);
      if (!extracted) {
        fail('Tar archive does not contain the expected binary: ' + binaryName);
      }

      try {
        fs.rmSync(outputPath, { force: true });
      } catch (_) {
      }

      fs.renameSync(extracted, outputPath);
    } finally {
      fs.rmSync(extractDir, { recursive: true, force: true });
    }
  } catch (err) {
    throw err;
  }
}

function extractZipEntry(zipPath, outputPath) {
  const data = fs.readFileSync(zipPath);
  const eocdSignature = 0x06054b50;
  const centralSignature = 0x02014b50;
  const localSignature = 0x04034b50;

  if (data.length < 22) {
    throw new Error('Downloaded archive is too small to be a valid zip file.');
  }

  let eocdOffset = -1;
  for (let i = data.length - 22; i >= Math.max(0, data.length - 65557); i -= 1) {
    if (data.readUInt32LE(i) === eocdSignature) {
      eocdOffset = i;
      break;
    }
  }

  if (eocdOffset === -1) {
    throw new Error('Downloaded archive is not a valid zip file.');
  }

  const centralDirectoryOffset = data.readUInt32LE(eocdOffset + 16);
  const totalEntries = data.readUInt16LE(eocdOffset + 10);
  let cursor = centralDirectoryOffset;
  let selected = null;

  for (let entry = 0; entry < totalEntries; entry += 1) {
    if (data.readUInt32LE(cursor) !== centralSignature) {
      throw new Error('Invalid zip central directory entry.');
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
    throw new Error('Zip archive does not contain a usable binary.');
  }

  if (data.readUInt32LE(selected.localHeaderOffset) !== localSignature) {
    throw new Error('Invalid zip local header.');
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
    throw new Error('Unsupported zip compression method: ' + selected.compressionMethod);
  }

  fs.writeFileSync(outputPath, extracted);
}

function download(url, destination, redirects = 0) {
  return new Promise((resolve, reject) => {
    if (redirects > 5) {
      reject(new Error('Too many redirects while downloading binary: ' + url));
      return;
    }

    https.get(url, (res) => {
      if (res.statusCode >= 300 && res.statusCode < 400 && res.headers.location) {
        return download(res.headers.location, destination, redirects + 1).then(resolve).catch(reject);
      }

      if (res.statusCode !== 200) {
        reject(new Error('HTTP ' + res.statusCode + ' from ' + url));
        return;
      }

      const tmpFile = destination + '.tmp';
      const file = fs.createWriteStream(tmpFile);

      res.pipe(file);

      file.on('finish', () => {
        file.close(async () => {
          try {
            if (assets[platformKey].archive === 'zip') {
              extractZipEntry(tmpFile, destination);
              fs.unlinkSync(tmpFile);
            } else if (assets[platformKey].archive === 'tar.gz') {
              await extractTarGzEntry(tmpFile, destination);
              fs.unlinkSync(tmpFile);
            } else {
              fs.renameSync(tmpFile, destination);
            }

            if (process.platform !== 'win32') {
              fs.chmodSync(destination, 0o755);
            }
            resolve();
          } catch (err) {
            try {
              fs.unlinkSync(tmpFile);
            } catch (_) {
            }
            reject(err);
          }
        });
      });

      file.on('error', (err) => {
        try {
          fs.unlinkSync(tmpFile);
        } catch (_) {
        }
        reject(err);
      });
    }).on('error', reject);
  });
}

async function tryDownloadUrls(urls, destination, urlIndex = 0) {
  if (urlIndex >= urls.length) {
    throw new Error('All asset URLs failed. Tried: ' + urls.join(', '));
  }

  const url = urls[urlIndex];
  try {
    console.log('[drb99] Attempting to download from ' + (urlIndex + 1) + '/' + urls.length + ': ' + url);
    await download(url, destination);
    console.log('[drb99] Installed ' + binaryName + ' for ' + platformKey);
  } catch (err) {
    if (urlIndex + 1 < urls.length) {
      console.warn('[drb99] Download failed: ' + err.message + '. Trying next URL...');
      return tryDownloadUrls(urls, destination, urlIndex + 1);
    } else {
      throw err;
    }
  }
}

async function main() {
  const target = assets[platformKey];
  if (!target) {
    const supported = Object.keys(assets).join(', ');
    fail('Unsupported platform/architecture.', 'Detected ' + platformKey + '. Supported: ' + supported);
  }

  if (!target.urls || target.urls.length === 0) {
    fail('No asset URLs configured for this platform.', platformKey);
  }

  ensureDir(targetDir);
  const outputName = process.platform === 'win32' ? binaryName + '.exe' : binaryName;
  const outputPath = path.join(targetDir, outputName);

  try {
    await tryDownloadUrls(target.urls, outputPath);
  } catch (err) {
    fail('Unable to install binary.', err.message);
  }
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
  console.error('[drb99] Binary is missing. Reinstall the package to trigger postinstall.');
  process.exit(1);
}

const child = 1pawn(binaryPath, process.argv.slice(2), { stdio: 'inherit' });

child.on('error', (err) => {
  console.error('[drb99] Failed to start binary:', err.message);
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
