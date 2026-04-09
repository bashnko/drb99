package service

type GenerateRequest struct {
	RepoURL    string            `json:"repo_url"`
	BinaryName string            `json:"binary_name"`
	Version    string            `json:"version,omitempty"`
	Platforms  []string          `json:"platforms"`
	Mode       string            `json:"mode"`
	Features   *Features         `json:"features,omitempty"`
	AssetURLs  map[string]string `json:"asset_urls,omitempty"`
}

type Features struct {
	NPMWrapper    bool `json:"npm_wrapper"`
	GoReleaser    bool `json:"goreleaser"`
	GithubActions bool `json:"github_actions"`
}
type WrapperConfig struct {
	RepoURL     string
	Owner       string
	Repo        string
	BinaryName  string
	Version     string
	NPMVersion  string
	PackageName string
	Features    Features
	Platforms   []PlatformAsset
}

type PlatformAsset struct {
	NodeKey    string
	InputKey   string
	GoSuffix   string
	BinaryFile string
	URL        string
	Archive    string
}

type GenerateResponse struct {
	Files map[string]string `json:"files"`
}
