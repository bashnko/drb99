package service

type GenerateRequest struct {
	RepoURL    string            `json:"repo_url"`
	BinaryName string            `json:"binary_name "`
	Version    string            `json:"version,omitempty"`
	Platform   []string          `json:"platform"`
	Mode       string            `json:"mode"`
	Features   *Features         `json:"features,omitempty"`
	AssetURLs  map[string]string `json:"asset_url,omitempty"`
}

type Features struct {
	NPMWrapper    bool `json:"npm_wrapper"`
	GoRealeser    bool `json:"go_releaser"`
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
}

type GenerateResponse struct {
	Files map[string]string `json:"files"`
}
