package service

import "context"

type Generator interface {
	Generate(wrapperConfig) (map[string]string, error)
}

type GithubClient interface {
	LatestRelease(ctx context.Context)
	ReleaseByTag(ctx context.Context)
	AssetExistByUrl(ctx context.Context)
}

type Service struct {
	gh  GithubClient
	gen Generator
}
