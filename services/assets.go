package service

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	gh "github.com/h3yng/drb99/github"
	"github.com/h3yng/drb99/utils"
)

func (s *Service) resolveAssets(ctx context.Context, mode, owner, repo, version string, assets []PlatformAsset, manualURLs map[string][]string) ([]PlatformAsset, error) {
	resolved := make([]PlatformAsset, len(assets))
	copy(resolved, assets)

	for i := range resolved {
		platform := resolved[i].InputKey
		binaryFile := resolved[i].BinaryFile
		nodeKey := resolved[i].NodeKey

		switch mode {
		case "manual":
			if len(manualURLs) == 0 {
				return nil, ErrAssetURLsRequired
			}
			urls := findAssetURLs(manualURLs, platform, nodeKey, binaryFile)
			if len(urls) == 0 {
				return nil, fmt.Errorf("%w %s (tried keys: %s, %s, %s)", ErrMissingAssetURL, platform, platform, nodeKey, binaryFile)
			}
			trimmedURLs := make([]string, 0, len(urls))
			for _, u := range urls {
				if trimmed := strings.TrimSpace(u); trimmed != "" {
					if !isValidURL(trimmed) {
						return nil, fmt.Errorf("%w for platform %s: %s", ErrInvalidAssetURL, platform, trimmed)
					}
					trimmedURLs = append(trimmedURLs, trimmed)
				}
			}
			if len(trimmedURLs) == 0 {
				return nil, fmt.Errorf("%w %s", ErrMissingAssetURL, platform)
			}
			resolved[i].URLs = trimmedURLs
		case "auto":
			url := gh.BuildReleaseAssetURL(owner, repo, version, binaryFile)
			exists, err := s.gh.AssetExistByUrl(ctx, url)
			if err != nil {
				return nil, fmt.Errorf("%w for %s: %v", ErrValidateReleaseAsset, platform, err)
			}
			if !exists {
				return nil, fmt.Errorf("missing release asset for %s (%s)", platform, binaryFile)
			}
			resolved[i].URLs = []string{url}

			digest, err := s.gh.GetAssetDigest(ctx, owner, repo, version, binaryFile)
			if err != nil {
				return nil, fmt.Errorf("%w for %s: %v", ErrFetchAssetDigest, platform, err)
			}
			resolved[i].SHA256 = digest
		default:
			return nil, fmt.Errorf("unsupported mode: %s", mode)
		}
	}

	return resolved, nil
}

func findAssetURLs(manualURLs map[string][]string, platform, nodeKey, binaryFile string) []string {
	keys := []string{
		platform,
		nodeKey,
		binaryFile,
	}

	for _, key := range keys {
		if urls, ok := manualURLs[key]; ok && len(urls) > 0 {
			return urls
		}
	}

	lowerKeys := make(map[string][]string, len(manualURLs))
	for k, v := range manualURLs {
		lowerKeys[strings.ToLower(k)] = v
	}
	for _, key := range keys {
		if urls, ok := lowerKeys[strings.ToLower(key)]; ok && len(urls) > 0 {
			return urls
		}
	}

	return nil
}

func buildPlatformAssets(binaryName, version string, platforms []string, features Features) ([]PlatformAsset, error) {
	assets := make([]PlatformAsset, 0, len(platforms))
	usedNode := map[string]bool{}

	for _, inputPlatform := range platforms {
		spec, err := utils.ResolvePlatformSpec(inputPlatform)
		if err != nil {
			return nil, err
		}

		nodeKey := utils.NodeKey(spec)
		if usedNode[nodeKey] {
			return nil, fmt.Errorf("%w: %s", ErrDuplicatePlatform, nodeKey)
		}
		usedNode[nodeKey] = true

		archiveType := archiveTypeForPlatform(features, inputPlatform)
		asset := PlatformAsset{
			NodeKey:    nodeKey,
			InputKey:   inputPlatform,
			GoSuffix:   spec.GoSuffix,
			GoOS:       spec.GoOS,
			GoArch:     spec.GoArch,
			BinaryFile: utils.ReleaseAssetName(binaryName, version, spec, archiveType),
			Archive:    archiveType,
		}

		assets = append(assets, asset)
	}

	return assets, nil
}

func isValidURL(rawURL string) bool {
	parsed, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return false
	}
	return parsed.Scheme == "http" || parsed.Scheme == "https"
}
