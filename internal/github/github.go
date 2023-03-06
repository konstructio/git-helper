package github

import (
	"context"
	"errors"
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/google/go-github/v45/github"
	"golang.org/x/oauth2"
)

var (
	defaultEvents []string = []string{"pull_request_review", "push", "issue_comment", "pull_request"}
)

// NewGitHubClient instantiates a new GitHub client wrapper
func NewGitHubClient(token string) GitHubWrapper {
	if token == "" {
		log.Fatal("you must provide a token when using github as a provider")
	}

	var gSession GitHubWrapper
	gSession.context = context.Background()
	gSession.staticToken = oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	gSession.oauthClient = oauth2.NewClient(gSession.context, gSession.staticToken)
	gSession.gitClient = github.NewClient(gSession.oauthClient)

	return gSession

}

// ListRepoWebhooks returns all webhooks for a repository
func (gh *GitHubWrapper) ListRepoWebhooks(owner string, repo string) ([]*github.Hook, error) {
	container := make([]*github.Hook, 0)
	for nextPage := 1; nextPage > 0; {
		hooks, resp, err := gh.gitClient.Repositories.ListHooks(gh.context, owner, repo, &github.ListOptions{
			Page:    nextPage,
			PerPage: 10,
		})
		if err != nil {
			return []*github.Hook{}, err
		}
		for _, hook := range hooks {
			container = append(container, hook)
		}
		nextPage = resp.NextPage
	}
	return container, nil
}

// CreateRepositoryWebhook
func (gh *GitHubWrapper) CreateRepositoryWebhook(req RepositoryHookRequest) error {
	_, _, err := gh.gitClient.Repositories.CreateHook(gh.context, req.Org, req.Repository, &github.Hook{
		Events: defaultEvents,
		Config: map[string]interface{}{
			"content_type": "json",
			"insecure_ssl": 0,
			"url":          req.Url,
			"secret":       req.Token,
		},
	})
	if err != nil {
		return errors.New(fmt.Sprintf("error when creating a webhook: %v", err))
	}
	log.Infof("created hook %s/%s/%s", req.Org, req.Repository, req.Url)

	return nil
}

// DeleteRepositoryWebhook
func (gh *GitHubWrapper) DeleteRepositoryWebhook(req RepositoryHookRequest) error {
	webhooks, err := gh.ListRepoWebhooks(req.Org, req.Repository)
	if err != nil {
		return err
	}

	var hookID int64 = 0
	for _, hook := range webhooks {
		if req.Url == hook.Config["url"] {
			hookID = hook.GetID()
		}
	}
	if hookID != 0 {
		_, err := gh.gitClient.Repositories.DeleteHook(gh.context, req.Org, req.Repository, hookID)
		if err != nil {
			return err
		}
		log.Infof("deleted hook %s/%s/%s", req.Org, req.Repository, req.Url)
	} else {
		return errors.New(fmt.Sprintf("hook %s/%s/%s not found", req.Org, req.Repository, req.Url))
	}

	return nil
}

// UpdateRepositoryWebhook
func (gh *GitHubWrapper) UpdateRepositoryWebhook(req RepositoryHookRequest) error {
	webhooks, err := gh.ListRepoWebhooks(req.Org, req.Repository)
	if err != nil {
		return err
	}

	var hookID int64 = 0
	for _, hook := range webhooks {
		if req.Url == hook.Config["url"] {
			hookID = hook.GetID()
		}
	}
	if hookID != 0 {
		_, _, err := gh.gitClient.Repositories.EditHook(gh.context, req.Org, req.Repository, hookID, &github.Hook{
			Events: defaultEvents,
			Config: map[string]interface{}{
				"content_type": "json",
				"insecure_ssl": 0,
				"url":          req.Url,
				"secret":       req.Token,
			},
		})
		if err != nil {
			return errors.New(fmt.Sprintf("error when creating a webhook: %v", err))
		}
		log.Infof("updated hook %s/%s/%s", req.Org, req.Repository, req.Url)
	} else {
		return errors.New(fmt.Sprintf("hook %s/%s/%s not found", req.Org, req.Repository, req.Url))
	}

	return nil
}
