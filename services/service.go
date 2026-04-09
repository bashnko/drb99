package service

import (
	"context"
	"fmt"
	"sort"
	"strings"

	gh "github.com/bashnko/drb99/github"
	"github.com/bashnko/drb99/utils"
)

type Generator interface {
	Generate(WrapperConfig) (map[string]string, error)
}

type GithubClient interface {
	LatestRelease(ctx context.Context, owner, repo string) (gh.Release, error)
	ReleaseByTag(ctx context.Context, owner, repo, tag string) (gh.Release, error)
	AssetExistByUrl(ctx context.Context, assetURL string) (bool, error)
}

type Service struct {
	gh  GithubClient
	gen Generator
}

func New(ghClient GithubClient, gen Generator) *Service {
	return &Service{gh: ghClient, gen: gen}
}

func (s *Service) Generate(ctx context.Context, req GenerateRequest) (GenerateResponse, error) {
	if normalizedFeatures(req.Features).isEmpty() {
		return GenerateResponse{}, fmt.Errorf("at least one feature must be enabled")
	}

	cfg, err := s.prepareConfig(ctx, req)
	if err != nil {
		return GenerateResponse{}, err
	}

	files, err := s.gen.Generate(cfg)
	if err != nil {
		return GenerateResponse{}, err
	}

	return GenerateResponse{Files: files}, nil
}

func (s *Service) prepareConfig(ctx context.Context, req GenerateRequest) (WrapperConfig, error) {
	features := normalizedFeatures(req.Features)
	if features.isEmpty() {
		return WrapperConfig{}, fmt.Errorf("at least one feature must be enabled")
	}

	owner, repo, err := utils.ParseGithubRepo(req.RepoURL)
	if err != nil {
		return WrapperConfig{}, err
	}

	if strings.TrimSpace(req.BinaryName) == "" {
		return WrapperConfig{}, fmt.Errorf("binary_name is required")
	}

	if len(req.Platforms) == 0 {
		return WrapperConfig{}, fmt.Errorf("platforms must not be empty")
	}

	mode := strings.ToLower(strings.TrimSpace(req.Mode))
	if mode != "auto" && mode != "manual" {
		return WrapperConfig{}, fmt.Errorf("mode must be either auto or manual")
	}

	version := utils.EnsureVersionPrefix(strings.TrimSpace(req.Version))
	if mode == "auto" && version == "" {
		release, err := s.gh.LatestRelease(ctx, owner, repo)
		if err != nil {
			return WrapperConfig{}, fmt.Errorf("resolve latest release version: %w", err)
		}

		version = release.TagName
		if strings.TrimSpace(version) == "" {
			return WrapperConfig{}, fmt.Errorf("latest release has empty tag name")
		}
	}

	if version == "" {
		return WrapperConfig{}, fmt.Errorf("version is required")
	}

	assets, err := s.resolveAssets(ctx, mode, owner, repo, strings.TrimSpace(req.BinaryName), version, req.Platforms, req.AssetURLs, features)
	if err != nil {
		return WrapperConfig{}, err
	}

	sort.Slice(assets, func(i, j int) bool {
		return assets[i].NodeKey < assets[j].NodeKey
	})

	return WrapperConfig{
		RepoURL:     req.RepoURL,
		Owner:       owner,
		Repo:        repo,
		BinaryName:  strings.TrimSpace(req.BinaryName),
		Version:     version,
		NPMVersion:  utils.NPMVersion(version),
		PackageName: strings.ToLower(strings.TrimSpace(req.BinaryName)) + "-npm",
		Features:    features,
		Platforms:   assets,
	}, nil
}

func (s *Service) resolveAssets(ctx context.Context, mode, owner, repo, binaryName, version string, platform []string, manualURL map[string]string, features Features) ([]PlatformAsset, error) {
	assets := make([]PlatformAsset, 0, len(platform))
	usedNode := map[string]bool{}
	for _, platform := range platform {
		spec, err := utils.ResolvePlatformSpec(platform)
		if err != nil {
			return nil, err
		}
		NodeKey := utils.NodeKey(spec)
		if usedNode[NodeKey] {
			return nil, fmt.Errorf("duplicate platform mapping for node target: %s", NodeKey)
		}
		usedNode[NodeKey] = true

		archiveType := archiveTypeForPlatform(features, platform)
		binaryFile := utils.ReleaseAssetName(binaryName, version, spec, archiveType)
		asset := PlatformAsset{
			NodeKey:    NodeKey,
			InputKey:   platform,
			GoSuffix:   spec.GoSuffix,
			BinaryFile: binaryFile,
			Archive:    archiveType,
		}

		switch mode {
		case "manual":
			if len(manualURL) == 0 {
				return nil, fmt.Errorf("asset urls is required in manual mode")
			}
			url := strings.TrimSpace(manualURL[platform])
			if url == "" {
				url = strings.TrimSpace(manualURL[NodeKey])
			}
			if url == "" {
				return nil, fmt.Errorf("missing manual asset URL for platform %s", platform)
			}
			asset.URL = url
		case "auto":
			url := gh.BuildReleaseAssetURL(owner, repo, version, binaryFile)
			exists, err := s.gh.AssetExistByUrl(ctx, url)
			if err != nil {
				return nil, fmt.Errorf("validate release asset for %s: %w", platform, err)
			}
			if !exists {
				return nil, fmt.Errorf("missing release asset for %s (%s)", platform, binaryFile)
			}
			asset.URL = url
		default:
			return nil, fmt.Errorf("unsopported mode: %s", mode)
		}
		assets = append(assets, asset)
	}
	return assets, nil
}

func normalizedFeatures(features *Features) Features {
	if features == nil {
		return Features{NPMWrapper: true}
	}
	return *features
}

func (f Features) isEmpty() bool {
	return !f.NPMWrapper && !f.GoReleaser && !f.GithubActions
}

func archiveTypeForPlatform(features Features, platform string) string {
	if features.GoReleaser {
		if platform == "windows-amd64" {
			return "zip"
		}
		return "binary"
	}
	return "binary"
}
