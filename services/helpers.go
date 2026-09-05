package service

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	gh "github.com/h3yng/drb99/github"
	"github.com/h3yng/drb99/utils"
)

var npmPackageNamePattern = regexp.MustCompile(`^(@[a-z0-9~][a-z0-9._~-]*/)?[a-z0-9~][a-z0-9._~-]*$`)

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
	return !f.NPMWrapper && !f.GoReleaser && !f.GithubActions && !f.AUR && !f.NixFlake && !f.DockerContainer && !f.Curl && !f.Iex
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
