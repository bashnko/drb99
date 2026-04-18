package service

import (
	"context"
	"fmt"
	"regexp"
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
	Repository(ctx context.Context, owner, repo string) (gh.Repository, error)
	AssetExistByUrl(ctx context.Context, assetURL string) (bool, error)
}

type Service struct {
	gh  GithubClient
	gen Generator
}

var npmPackageNamePattern = regexp.MustCompile(`^(@[a-z0-9~][a-z0-9._~-]*/)?[a-z0-9~][a-z0-9._~-]*$`)

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

func (s *Service) Prefill(ctx context.Context, req PrefillRequest) (PrefillResponse, error) {
	owner, repo, err := utils.ParseGithubRepo(req.RepoURL)
	if err != nil {
		return PrefillResponse{}, err
	}

	repository, err := s.gh.Repository(ctx, owner, repo)
	if err != nil {
		return PrefillResponse{}, fmt.Errorf("resolve repository metadata: %w", err)
	}

	resp := PrefillResponse{
		RepoURL:     strings.TrimSpace(req.RepoURL),
		Owner:       owner,
		Repo:        repo,
		Name:        strings.TrimSpace(repository.Name),
		Author:      owner,
		Description: strings.TrimSpace(repository.Description),
		License:     preferredLicenseName(repository.License),
		AssetURLs:   map[string]string{},
	}

	if strings.TrimSpace(repository.Owner.Login) != "" {
		resp.Author = strings.TrimSpace(repository.Owner.Login)
	}

	release, err := s.gh.LatestRelease(ctx, owner, repo)
	if err != nil {
		if !gh.IsNotFound(err) {
			return PrefillResponse{}, fmt.Errorf("resolve latest release version: %w", err)
		}
		return resp, nil
	}

	resp.Version = strings.TrimSpace(release.TagName)
	for _, asset := range release.Assets {
		assetName := strings.TrimSpace(asset.Name)
		assetURL := strings.TrimSpace(asset.URL)
		if assetName == "" || assetURL == "" {
			continue
		}
		resp.Assets = append(resp.Assets, ReleaseAsset{Name: assetName, URL: assetURL})
		resp.AssetURLs[assetName] = assetURL
	}

	if len(resp.AssetURLs) == 0 {
		resp.AssetURLs = nil
	}

	return resp, nil
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

	packageName := strings.TrimSpace(req.PackageName)
	license := strings.TrimSpace(req.License)
	description := strings.TrimSpace(req.Description)
	if features.NPMWrapper {
		if packageName == "" {
			return WrapperConfig{}, fmt.Errorf("package name is required when npm wrapper is enabled")
		}
		if err := validateNPMPackageName(packageName); err != nil {
			return WrapperConfig{}, err
		}
		if license == "" {
			license = "MIT"
		}
		if description == "" {
			description = fmt.Sprintf("npm wrapper for %s", strings.TrimSpace(req.BinaryName))
		}
	}

	mode := strings.ToLower(strings.TrimSpace(req.Mode))
	if features.NPMWrapper {
		if mode != "auto" && mode != "manual" {
			return WrapperConfig{}, fmt.Errorf("mode must be either auto or manual when npm_wrapper is enabled")
		}
	} else if mode != "" && mode != "auto" && mode != "manual" {
		return WrapperConfig{}, fmt.Errorf("mode must be either auto or manual")
	}

	selectedPlatforms := req.Platforms
	if len(selectedPlatforms) == 0 {
		if features.NPMWrapper || features.GoReleaser {
			selectedPlatforms = defaultPlatforms()
		}
	}

	version := utils.EnsureVersionPrefix(strings.TrimSpace(req.Version))
	if features.NPMWrapper && mode == "auto" && version == "" {
		release, err := s.gh.LatestRelease(ctx, owner, repo)
		if err != nil {
			return WrapperConfig{}, fmt.Errorf("resolve latest release version: %w", err)
		}

		version = release.TagName
		if strings.TrimSpace(version) == "" {
			return WrapperConfig{}, fmt.Errorf("latest release has empty tag name")
		}
	}

	if features.NPMWrapper && version == "" {
		return WrapperConfig{}, fmt.Errorf("version is required")
	}

	assets, err := buildPlatformAssets(strings.TrimSpace(req.BinaryName), version, selectedPlatforms, features)
	if err != nil {
		return WrapperConfig{}, err
	}

	if features.NPMWrapper {
		assets, err = s.resolveAssets(ctx, mode, owner, repo, version, assets, req.AssetURLs)
		if err != nil {
			return WrapperConfig{}, err
		}
	}

	sort.Slice(assets, func(i, j int) bool {
		return assets[i].NodeKey < assets[j].NodeKey
	})

	goReleaserTargets := collectGoReleaserTargets(assets)

	return WrapperConfig{
		RepoURL:           req.RepoURL,
		Owner:             owner,
		Repo:              repo,
		BinaryName:        strings.TrimSpace(req.BinaryName),
		Version:           version,
		NPMVersion:        utils.NPMVersion(version),
		PackageName:       packageName,
		License:           license,
		Description:       description,
		Author:            owner,
		Features:          features,
		Platforms:         assets,
		GoReleaserTargets: goReleaserTargets,
	}, nil
}

func validateNPMPackageName(name string) error {
	if len(name) > 214 {
		return fmt.Errorf("package name must be 214 characters or less")
	}
	if strings.HasPrefix(name, ".") || strings.HasPrefix(name, "_") {
		return fmt.Errorf("package name cannot start with . or _")
	}
	if strings.Contains(name, " ") {
		return fmt.Errorf("package name cannot contain spaces")
	}

	if !npmPackageNamePattern.MatchString(name) {
		return fmt.Errorf("package name is invalid for npm")
	}
	return nil
}

func (s *Service) resolveAssets(ctx context.Context, mode, owner, repo, version string, assets []PlatformAsset, manualURL map[string]string) ([]PlatformAsset, error) {
	resolved := make([]PlatformAsset, len(assets))
	copy(resolved, assets)

	for i := range resolved {
		platform := resolved[i].InputKey
		binaryFile := resolved[i].BinaryFile

		switch mode {
		case "manual":
			if len(manualURL) == 0 {
				return nil, fmt.Errorf("asset urls is required in manual mode")
			}
			url := strings.TrimSpace(manualURL[platform])
			if url == "" {
				url = strings.TrimSpace(manualURL[resolved[i].NodeKey])
			}
			if url == "" {
				return nil, fmt.Errorf("missing manual asset URL for platform %s", platform)
			}
			resolved[i].URL = url
		case "auto":
			url := gh.BuildReleaseAssetURL(owner, repo, version, binaryFile)
			exists, err := s.gh.AssetExistByUrl(ctx, url)
			if err != nil {
				return nil, fmt.Errorf("validate release asset for %s: %w", platform, err)
			}
			if !exists {
				return nil, fmt.Errorf("missing release asset for %s (%s)", platform, binaryFile)
			}
			resolved[i].URL = url
		default:
			return nil, fmt.Errorf("unsupported mode: %s", mode)
		}
	}

	return resolved, nil
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
			return nil, fmt.Errorf("duplicate platform mapping for node target: %s", nodeKey)
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

func defaultPlatforms() []string {
	platformSpecs := utils.SupportedPlatformSpecs()
	platforms := make([]string, 0, len(platformSpecs))
	for platform := range platformSpecs {
		platforms = append(platforms, platform)
	}
	sort.Strings(platforms)
	return platforms
}

func collectGoReleaserTargets(assets []PlatformAsset) []string {
	targetSet := make(map[string]struct{}, len(assets))
	for _, asset := range assets {
		if asset.GoOS == "" || asset.GoArch == "" {
			continue
		}
		targetSet[asset.GoOS+"_"+asset.GoArch] = struct{}{}
	}

	targets := make([]string, 0, len(targetSet))
	for target := range targetSet {
		targets = append(targets, target)
	}
	sort.Strings(targets)
	return targets
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

func archiveTypeForPlatform(_ Features, platform string) string {
	if platform == "windows-amd64" {
		return "zip"
	}
	return "tar.gz"
}

func preferredLicenseName(license gh.License) string {
	if strings.TrimSpace(license.SPDXID) != "" && strings.TrimSpace(license.SPDXID) != "NOASSERTION" {
		return strings.TrimSpace(license.SPDXID)
	}
	return strings.TrimSpace(license.Name)
}
