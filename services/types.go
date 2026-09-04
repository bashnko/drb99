package service

type GenerateRequest struct {
	RepoURL      string              `json:"repo_url" validate:"required,url"`
	BinaryName   string              `json:"binary_name" validate:"required,min=1,max=100"`
	PackageName  string              `json:"package_name,omitempty" validate:"omitempty,min=1,max=214"`
	License      string              `json:"license,omitempty" validate:"omitempty,spdx"`
	Description  string              `json:"description,omitempty" validate:"omitempty,max=500"`
	Version      string              `json:"version,omitempty" validate:"omitempty,semver"`
	Platforms    []string            `json:"platforms" validate:"omitempty,dive,platform"`
	Mode         string              `json:"mode" validate:"omitempty,oneof=auto manual"`
	Features     *Features           `json:"features,omitempty"`
	AssetURLs    map[string][]string `json:"asset_urls,omitempty"`
	RuntimeImage string              `json:"runtime_image" validate:"omitempty,min=1"`
}

type Features struct {
	NPMWrapper      bool `json:"npm_wrapper"`
	GoReleaser      bool `json:"goreleaser"`
	GithubActions   bool `json:"github_actions"`
	AUR             bool `json:"aur"`
	NixFlake        bool `json:"nix_flake"`
	DockerContainer bool `json:"docker_container"`
	Curl            bool `json:"curl"`
}
type WrapperConfig struct {
	RepoURL           string
	Owner             string
	Repo              string
	BinaryName        string
	Version           string
	NPMVersion        string
	AURVersion        string
	PackageName       string
	License           string
	Description       string
	Author            string
	Features          Features
	Platforms         []PlatformAsset
	GoReleaserTargets []string
	MainPath          string
	RuntimeImage      string
}

// type DockerRuntime struct {
// 	Golang     string
// 	JavaScript string
// 	TypeScript string
// }

type PlatformAsset struct {
	NodeKey    string
	InputKey   string
	GoOS       string
	GoArch     string
	GoSuffix   string
	BinaryFile string
	URLs       []string
	Archive    string
	SHA256     string
}

type GenerateResponse struct {
	Files map[string]string `json:"files"`
}

type PrefillRequest struct {
	RepoURL string `json:"repo_url" validate:"required,url"`
}

type PrefillResponse struct {
	RepoURL     string              `json:"repo_url"`
	Owner       string              `json:"owner"`
	Repo        string              `json:"repo"`
	Name        string              `json:"name"`
	Version     string              `json:"version,omitempty"`
	Author      string              `json:"author"`
	Description string              `json:"description,omitempty"`
	License     string              `json:"license,omitempty"`
	Assets      []ReleaseAsset      `json:"assets,omitempty"`
	AssetURLs   map[string][]string `json:"asset_urls,omitempty"`
}

type ReleaseAsset struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}
