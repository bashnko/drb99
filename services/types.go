package service

type GenerateRequest struct {
	RepoURL     string            `json:"repo_url"`
	BinaryName  string            `json:"binary_name"`
	PackageName string            `json:"package_name,omitempty"`
	License     string            `json:"license,omitempty"`
	Description string            `json:"description,omitempty"`
	Version     string            `json:"version,omitempty"`
	Platforms   []string          `json:"platforms"`
	Mode        string            `json:"mode"`
	Features    *Features         `json:"features,omitempty"`
	AssetURLs   map[string]string `json:"asset_urls,omitempty"`
}

type Features struct {
	NPMWrapper    bool `json:"npm_wrapper"`
	GoReleaser    bool `json:"goreleaser"`
	GithubActions bool `json:"github_actions"`
}
type WrapperConfig struct {
	RepoURL           string
	Owner             string
	Repo              string
	BinaryName        string
	Version           string
	NPMVersion        string
	PackageName       string
	License           string
	Description       string
	Author            string
	Features          Features
	Platforms         []PlatformAsset
	GoReleaserTargets []string
}

type PlatformAsset struct {
	NodeKey    string
	InputKey   string
	GoOS       string
	GoArch     string
	GoSuffix   string
	BinaryFile string
	URL        string
	Archive    string
}

type GenerateResponse struct {
	Files map[string]string `json:"files"`
}

type PrefillRequest struct {
	RepoURL string `json:"repo_url"`
}

type PrefillResponse struct {
	RepoURL     string            `json:"repo_url"`
	Owner       string            `json:"owner"`
	Repo        string            `json:"repo"`
	Name        string            `json:"name"`
	Version     string            `json:"version,omitempty"`
	Author      string            `json:"author"`
	Description string            `json:"description,omitempty"`
	License     string            `json:"license,omitempty"`
	Assets      []ReleaseAsset    `json:"assets,omitempty"`
	AssetURLs   map[string]string `json:"asset_urls,omitempty"`
}

type ReleaseAsset struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}
