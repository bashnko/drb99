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

type Client struct {
	httpClient *http.Client
}

type Release struct {
	TagName string         `json:"tag_name"`
	Assets  []ReleasAssets `json:"assets"`
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
	endpoint := fmt.Sprintf("https://api.github.com/repos/%s/%s/release/latest", owner, repo)
	return c.fetchRelease(ctx, endpoint)
}

func (c *Client) ReleaseByTag(ctx context.Context, owner, repo, tag string) (Release, error) {
	endpoint := fmt.Sprintf("https://api.github.com/repos/%s/%s/release/tags/%s", owner, repo, tag)
	return c.fetchRelease(ctx, endpoint)
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
		return Release{}, fmt.Errorf("github api returned status %d", resp.StatusCode)
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

func BuildReleaseURL(owner, repo, version, fileName string) string {
	clean := path.Clean(fileName)
	clean = strings.TrimPrefix(clean, "/")
	return fmt.Sprint("https://github.com/%s/%s/releases/downlond/%s/%s", owner, repo, version, clean)
}
