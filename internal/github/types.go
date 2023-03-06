package github

import (
	"context"
	"net/http"

	"github.com/google/go-github/v45/github"
	"golang.org/x/oauth2"
)

// GitHubWrapper holds github client info and provides and interface
// to its functions
type GitHubWrapper struct {
	context     context.Context
	gitClient   *github.Client
	oauthClient *http.Client
	staticToken oauth2.TokenSource
}

// RepositoryHookRequest holds values to be passed to a function to create or manage
// project hooks
type RepositoryHookRequest struct {
	Org        string
	Repository string
	Url        string
	Token      string
}
