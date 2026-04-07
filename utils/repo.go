package utils

import (
	"fmt"
	"net/url"
	"strings"
)

func ParseGithubRepo(repoURL string) (owner string, repo string, err error) {
	u, err := url.Parse(strings.TrimSpace(repoURL))
	if err != nil {
		return "", "", fmt.Errorf("invalid repo url: %w", err)
	}

	if u.Scheme != "https" || !strings.EqualFold(u.Host, "github.com") {
		return "", "", fmt.Errorf("invalid repo url: must be an https GitHub URL")
	}

	parts := strings.Split(strings.Trim(strings.TrimSuffix(u.Path, ".git"), "/"), "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("invalid repo url: expected format https://github.com/{owner}/{repo}")
	}
	return parts[0], parts[1], nil

}

func EnsureVersionPrefix(version string) string {
	if version == "" {
		return version
	}
	if strings.HasPrefix(version, "v") {
		return version
	}
	return "v" + version
}

func NPMVersion(version string) string {
	v := strings.TrimSpace(version)
	if strings.HasPrefix(v, "v") && len(v) > 1 {
		return v[1:]
	}
	return v
}
