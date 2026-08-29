package service

import (
	"context"
	"fmt"
	"strings"

	gh "github.com/h3yng/drb99/github"
	"github.com/h3yng/drb99/utils"
)

type Generator interface {
	Generate(WrapperConfig) (map[string]string, error)
}

type GithubClient interface {
	LatestRelease(ctx context.Context, owner, repo string) (gh.Release, error)
	ReleaseByTag(ctx context.Context, owner, repo, tag string) (gh.Release, error)
	Repository(ctx context.Context, owner, repo string) (gh.Repository, error)
	AssetExistByUrl(ctx context.Context, assetURL string) (bool, error)
	GetAssetDigest(ctx context.Context, owner, repo, tag, assetName string) (string, error)
}

type Service struct {
	gh  GithubClient
	gen Generator
}

func New(ghClient GithubClient, gen Generator) *Service {
	return &Service{gh: ghClient, gen: gen}
}

func (s *Service) Generate(ctx context.Context, req GenerateRequest) (GenerateResponse, error) {
	if normalizedFeatures(req.Features).isEmpty() {
		return GenerateResponse{}, ErrNoFeaturesEnabled
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
		AssetURLs:   map[string][]string{},
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
		resp.AssetURLs[assetName] = append(resp.AssetURLs[assetName], assetURL)
	}

	if len(resp.AssetURLs) == 0 {
		resp.AssetURLs = nil
	}

	return resp, nil
}
