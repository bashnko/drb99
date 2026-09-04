package service

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/h3yng/drb99/utils"
)

var (
	ErrNoFeaturesEnabled     = errors.New("at least one feature must be enabled")
	ErrBinaryNameRequired    = errors.New("binary_name is required")
	ErrPackageNameRequired   = errors.New("package name is required when npm wrapper is enabled")
	ErrVersionRequired       = errors.New("version is required")
	ErrAURVersionRequired    = errors.New("version is required when aur is enabled")
	ErrCurlVersionRequired   = errors.New("version is required when curl is enabled")
	ErrAssetURLsRequired     = errors.New("asset_urls is required in manual mode")
	ErrInvalidMode           = errors.New("mode must be either auto or manual")
	ErrInvalidAssetURL       = errors.New("invalid URL in asset_urls")
	ErrMissingAssetURL       = errors.New("missing manual asset URL for platform")
	ErrLatestReleaseEmptyTag = errors.New("latest release has empty tag name")
	ErrResolveReleaseVersion = errors.New("resolve latest release version")
	ErrResolveLatestVersion  = errors.New("resolve latest version")
	ErrValidateReleaseAsset  = errors.New("validate release asset")
	ErrFetchAssetDigest      = errors.New("fetch asset digest")
	ErrDuplicatePlatform     = errors.New("duplicate platform mapping for node target")
)

func (s *Service) prepareConfig(ctx context.Context, req GenerateRequest) (WrapperConfig, error) {
	features := normalizedFeatures(req.Features)
	if features.isEmpty() {
		return WrapperConfig{}, ErrNoFeaturesEnabled
	}

	owner, repo, err := utils.ParseGithubRepo(req.RepoURL)
	if err != nil {
		return WrapperConfig{}, err
	}

	binaryName := strings.TrimSpace(req.BinaryName)
	if binaryName == "" {
		return WrapperConfig{}, ErrBinaryNameRequired
	}

	packageName := strings.TrimSpace(req.PackageName)
	license := strings.TrimSpace(req.License)
	description := strings.TrimSpace(req.Description)
	runtimeImage := strings.TrimSpace(req.RuntimeImage)

	if features.NPMWrapper {
		if packageName == "" {
			return WrapperConfig{}, ErrPackageNameRequired
		}
		if err := validateNPMPackageName(packageName); err != nil {
			return WrapperConfig{}, err
		}
		if license == "" {
			license = "MIT"
		}
		if description == "" {
			description = fmt.Sprintf("npm wrapper for %s", binaryName)
		}
	}

	if features.DockerContainer {
		if runtimeImage == "" {
			runtimeImage = "golang:latest"
		}
	}

	if features.AUR {
		if license == "" {
			license = "MIT"
		}
		if description == "" {
			description = fmt.Sprintf("%s CLI", binaryName)
		}
	}

	mode := strings.ToLower(strings.TrimSpace(req.Mode))
	if features.NPMWrapper || features.Curl {
		if mode != "auto" && mode != "manual" {
			return WrapperConfig{}, ErrInvalidMode
		}
	} else if mode != "" && mode != "auto" && mode != "manual" {
		return WrapperConfig{}, ErrInvalidMode
	}

	selectedPlatforms := req.Platforms
	if len(selectedPlatforms) == 0 {
		if features.NPMWrapper || features.GoReleaser || features.Curl {
			selectedPlatforms = defaultPlatforms()
		}
	}

	version := utils.EnsureVersionPrefix(strings.TrimSpace(req.Version))
	if (features.NPMWrapper || features.Curl) && mode == "auto" && version == "" {
		release, err := s.gh.LatestRelease(ctx, owner, repo)
		if err != nil {
			return WrapperConfig{}, fmt.Errorf("%w: %v", ErrResolveReleaseVersion, err)
		}

		version = release.TagName
		if strings.TrimSpace(version) == "" {
			return WrapperConfig{}, ErrLatestReleaseEmptyTag
		}
	}

	if features.AUR && version == "" {
		release, err := s.gh.LatestRelease(ctx, owner, repo)
		if err != nil {
			return WrapperConfig{}, fmt.Errorf("%w: %v", ErrResolveLatestVersion, err)
		}

		version = strings.TrimSpace(release.TagName)
		if version == "" {
			return WrapperConfig{}, ErrLatestReleaseEmptyTag
		}
	}

	if features.NPMWrapper && version == "" {
		return WrapperConfig{}, ErrVersionRequired
	}
	if features.AUR && version == "" {
		return WrapperConfig{}, ErrAURVersionRequired
	}
	if features.Curl && version == "" {
		return WrapperConfig{}, ErrCurlVersionRequired
	}

	assets, err := buildPlatformAssets(binaryName, version, selectedPlatforms, features)
	if err != nil {
		return WrapperConfig{}, err
	}

	if features.NPMWrapper || features.Curl {
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
		BinaryName:        binaryName,
		Version:           version,
		NPMVersion:        utils.NPMVersion(version),
		AURVersion:        utils.NPMVersion(version),
		PackageName:       packageName,
		License:           license,
		Description:       description,
		Author:            owner,
		Features:          features,
		Platforms:         assets,
		GoReleaserTargets: goReleaserTargets,
		MainPath:          "./cmd/" + binaryName + "/main.go",
		RuntimeImage:      runtimeImage,
	}, nil
}
