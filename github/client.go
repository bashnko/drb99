package github

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"strings"
	"time"
)

type APIError struct {
	StatusCode int
	Message    string
}

func (e APIError) Error() string {
	if strings.TrimSpace(e.Message) == "" {
		return fmt.Sprintf("github api returned status %d", e.StatusCode)
	}
	return fmt.Sprintf("github api returned status %d: %s", e.StatusCode, e.Message)
}

func IsNotFound(err error) bool {
	apiErr, ok := err.(APIError)
	return ok && apiErr.StatusCode == http.StatusNotFound
}

type Client struct {
	httpClient *http.Client
}

type Release struct {
	TagName string         `json:"tag_name"`
	Assets  []ReleasAssets `json:"assets"`
}

type Repository struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	License     License   `json:"license"`
	Owner       RepoOwner `json:"owner"`
}

type RepoOwner struct {
	Login string `json:"login"`
}

type License struct {
	SPDXID string `json:"spdx_id"`
	Name   string `json:"name"`
}

type ReleasAssets struct {
	Name string `json:"name"`
	URL  string `json:"browser_download_url"`
}

func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 20 * time.Second},
	}
}

func (c *Client) LatestRelease(ctx context.Context, owner, repo string) (Release, error) {
	endpoint := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo)
	return c.fetchRelease(ctx, endpoint)
}

func (c *Client) ReleaseByTag(ctx context.Context, owner, repo, tag string) (Release, error) {
	endpoint := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/tags/%s", owner, repo, tag)
	return c.fetchRelease(ctx, endpoint)
}

func (c *Client) Repository(ctx context.Context, owner, repo string) (Repository, error) {
	endpoint := fmt.Sprintf("https://api.github.com/repos/%s/%s", owner, repo)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return Repository{}, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "drb99/1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return Repository{}, err
	}

	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return Repository{}, APIError{StatusCode: resp.StatusCode}
	}
	var out Repository
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return Repository{}, fmt.Errorf("decode github response: %w", err)
	}
	return out, nil
}

func (c *Client) fetchRelease(ctx context.Context, endpoint string) (Release, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return Release{}, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "drb99/1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return Release{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return Release{}, APIError{StatusCode: resp.StatusCode}
	}
	var out Release
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return Release{}, fmt.Errorf("decode github response: %w", err)
	}
	return out, nil
}

func (c *Client) AssetExistByUrl(ctx context.Context, assetURL string) (bool, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, assetURL, nil)
	if err != nil {
		return false, err
	}
	req.Header.Set("User-Agent", "drb99/1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return false, nil
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		return true, nil
	}
	return false, fmt.Errorf("unexpected status while validating asset: %d", resp.StatusCode)

}

func BuildReleaseAssetURL(owner, repo, version, fileName string) string {
	clean := path.Clean(fileName)
	clean = strings.TrimPrefix(clean, "/")
	return fmt.Sprintf("https://github.com/%s/%s/releases/download/%s/%s", owner, repo, version, clean)
}
