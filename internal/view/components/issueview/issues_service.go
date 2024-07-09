package issueview

import (
	"context"

	"github.com/google/go-github/v61/github"
)

//go:generate mockgen -source=issues_service.go -destination=mock/issues_service.go

type GithubIssueService interface {
	Get(ctx context.Context, owner string, repo string, number int) (*github.Issue, *github.Response, error)
}
