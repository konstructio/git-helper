package gitlabcloud

import "github.com/xanzy/go-gitlab"

// GitLabWrapper holds gitlab cloud client info and provides and interface
// to its functions
type GitLabWrapper struct {
	Client          *gitlab.Client
	ParentGroupID   int
	ParentGroupPath string
}

// DeployTokenCreateParameters holds values to be passed to a function to create
// deploy tokens
type DeployTokenCreateParameters struct {
	Name     string
	Username string
	Scopes   []string
}

// ProjectWebhookRequest holds values to be passed to a function to create or manage
// project hooks
type ProjectHookRequest struct {
	ProjectName string

	CreateOpts *gitlab.AddProjectHookOptions
	PatchOpts  *gitlab.EditProjectHookOptions
}
