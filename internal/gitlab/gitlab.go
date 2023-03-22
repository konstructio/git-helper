package gitlabcloud

import (
	"errors"
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/xanzy/go-gitlab"
)

func NewGitLabClient(token string) *gitlab.Client {
	if token == "" {
		log.Fatal("you must provide a token when using gitlab as a provider")
	}

	git, err := gitlab.NewClient(token)
	if err != nil {
		log.Fatal(err)
	}

	return git
}

// CheckProjectExists
func (gl *GitLabWrapper) CheckProjectExists(projectName string) (bool, error) {
	allprojects, err := gl.GetProjects()
	if err != nil {
		return false, err
	}

	var exists bool = false
	for _, project := range allprojects {
		if project.Name == projectName {
			exists = true
		}
	}

	return exists, nil
}

// GetProjectID
func (gl *GitLabWrapper) GetProjectID(projectName string) (int, error) {
	owned := true

	container := make([]gitlab.Project, 0)
	for nextPage := 1; nextPage > 0; {
		projects, resp, err := gl.Client.Projects.ListProjects(&gitlab.ListProjectsOptions{
			ListOptions: gitlab.ListOptions{
				Page:    nextPage,
				PerPage: 10,
			},
			Owned: &owned,
		})
		if err != nil {
			return 0, err
		}
		for _, project := range projects {
			container = append(container, *project)
		}
		nextPage = resp.NextPage
	}

	for _, project := range container {
		if project.Name == projectName {
			return project.ID, nil
		}
	}

	return 0, errors.New(fmt.Sprintf("could not get project ID for project %s", projectName))
}

// GetProjects
func (gl *GitLabWrapper) GetProjects() ([]gitlab.Project, error) {
	owned := true

	container := make([]gitlab.Project, 0)
	for nextPage := 1; nextPage > 0; {
		projects, resp, err := gl.Client.Projects.ListProjects(&gitlab.ListProjectsOptions{
			ListOptions: gitlab.ListOptions{
				Page:    nextPage,
				PerPage: 10,
			},
			Owned: &owned,
		})
		if err != nil {
			return []gitlab.Project{}, err
		}
		for _, project := range projects {
			container = append(container, *project)
		}
		nextPage = resp.NextPage
	}

	return container, nil
}

// Token & Key Management

// CreateProjectDeployToken
func (gl *GitLabWrapper) CreateProjectDeployToken(projectName string, p *DeployTokenCreateParameters) (string, error) {
	projectID, err := gl.GetProjectID(projectName)
	if err != nil {
		return "", err
	}

	// Check to see if the token already exists
	allTokens, err := gl.ListProjectDeployTokens(projectName)
	if err != nil {
		return "", err
	}

	var exists bool = false
	for _, token := range allTokens {
		if token.Name == p.Name {
			exists = true
		}
	}

	if !exists {
		token, _, err := gl.Client.DeployTokens.CreateProjectDeployToken(projectID, &gitlab.CreateProjectDeployTokenOptions{
			Name:     &p.Name,
			Username: &p.Username,
			Scopes:   &p.Scopes,
		})
		if err != nil {
			return "", err
		}
		log.Infof("created deploy token %s", token.Name)

		return token.Token, nil
	} else {
		log.Infof("deploy token %s already exists - skipping", p.Name)
		return "", nil
	}
}

// DeleteProjectDeployToken
func (gl *GitLabWrapper) DeleteProjectDeployToken(projectName string, tokenName string) error {
	projectID, err := gl.GetProjectID(projectName)
	if err != nil {
		return err
	}

	allTokens, err := gl.ListProjectDeployTokens(projectName)
	if err != nil {
		return err
	}

	var exists bool = false
	var tokenID int
	for _, token := range allTokens {
		if token.Name == tokenName {
			exists = true
			tokenID = token.ID
		}
	}

	if exists {
		_, err = gl.Client.DeployTokens.DeleteProjectDeployToken(projectID, tokenID)
		if err != nil {
			return err
		}
		log.Infof("deleted deploy token %s", tokenName)
	}

	return nil

}

// ListProjectDeployTokens
func (gl *GitLabWrapper) ListProjectDeployTokens(projectName string) ([]gitlab.DeployToken, error) {
	projectID, err := gl.GetProjectID(projectName)
	if err != nil {
		return []gitlab.DeployToken{}, err
	}

	container := make([]gitlab.DeployToken, 0)
	for nextPage := 1; nextPage > 0; {
		tokens, resp, err := gl.Client.DeployTokens.ListProjectDeployTokens(projectID, &gitlab.ListProjectDeployTokensOptions{
			Page:    nextPage,
			PerPage: 10,
		})
		if err != nil {
			return []gitlab.DeployToken{}, err
		}
		for _, token := range tokens {
			container = append(container, *token)
		}
		nextPage = resp.NextPage
	}

	return container, nil
}

// Webhooks

// ListProjectWebhooks returns all webhooks for a project
func (gl *GitLabWrapper) ListProjectWebhooks(projectID int) ([]gitlab.ProjectHook, error) {
	container := make([]gitlab.ProjectHook, 0)
	for nextPage := 1; nextPage > 0; {
		hooks, resp, err := gl.Client.Projects.ListProjectHooks(projectID, &gitlab.ListProjectHooksOptions{
			Page:    nextPage,
			PerPage: 10,
		})
		if err != nil {
			return []gitlab.ProjectHook{}, err
		}
		for _, hook := range hooks {
			container = append(container, *hook)
		}
		nextPage = resp.NextPage
	}
	return container, nil
}

// CreateProjectWebhook
func (gl *GitLabWrapper) CreateProjectWebhook(req *ProjectHookRequest) error {
	projectID, err := gl.GetProjectID(req.ProjectName)
	if err != nil {
		return err
	}

	_, _, err = gl.Client.Projects.AddProjectHook(projectID, req.CreateOpts)
	if err != nil {
		return err
	}
	log.Infof("created hook %s/%s", req.ProjectName, *req.CreateOpts.URL)

	return nil
}

// DeleteProjectWebhook
func (gl *GitLabWrapper) DeleteProjectWebhook(req *ProjectHookRequest) error {
	projectID, err := gl.GetProjectID(req.ProjectName)
	if err != nil {
		return err
	}

	webhooks, err := gl.ListProjectWebhooks(projectID)
	if err != nil {
		return err
	}

	var hookID int = 0
	for _, hook := range webhooks {
		if hook.ProjectID == projectID && hook.URL == *req.CreateOpts.URL {
			hookID = hook.ID
		}
	}
	if hookID == 0 {
		return errors.New(fmt.Sprintf("no webhooks were found for project %s given search parameters", req.ProjectName))
	}
	_, err = gl.Client.Projects.DeleteProjectHook(projectID, hookID)
	if err != nil {
		return err
	}
	log.Infof("deleted hook %s/%s", req.ProjectName, *req.CreateOpts.URL)

	return nil
}

// UpdateProjectWebhook
func (gl *GitLabWrapper) UpdateProjectWebhook(req *ProjectHookRequest) error {
	projectID, err := gl.GetProjectID(req.ProjectName)
	if err != nil {
		return err
	}

	webhooks, err := gl.ListProjectWebhooks(projectID)
	if err != nil {
		return err
	}

	var hookID int = 0
	for _, hook := range webhooks {
		if hook.ProjectID == projectID && hook.URL == *req.CreateOpts.URL {
			hookID = hook.ID
		}
	}
	if hookID == 0 {
		return errors.New(fmt.Sprintf("no webhooks were found for project %s given search parameters", req.ProjectName))
	}
	_, _, err = gl.Client.Projects.EditProjectHook(projectID, hookID, req.PatchOpts)
	if err != nil {
		return err
	}

	return nil
}
