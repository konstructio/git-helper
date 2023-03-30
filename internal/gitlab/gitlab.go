package gitlabcloud

import (
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/xanzy/go-gitlab"
)

// NewGitLabClient instantiates a wrapper to communicate with GitLab
// It sets the path and ID of the group under which resources will be managed
func NewGitLabClient(token string, parentGroupName string) (GitLabWrapper, error) {
	git, err := gitlab.NewClient(token)
	if err != nil {
		return GitLabWrapper{}, fmt.Errorf("error instantiating gitlab client: %s", err)
	}

	// Get parent group ID
	minAccessLevel := gitlab.AccessLevelValue(gitlab.DeveloperPermissions)
	container := make([]gitlab.Group, 0)
	for nextPage := 1; nextPage > 0; {
		groups, resp, err := git.Groups.ListGroups(&gitlab.ListGroupsOptions{
			ListOptions: gitlab.ListOptions{
				Page:    nextPage,
				PerPage: 10,
			},
			MinAccessLevel: &minAccessLevel,
		})
		if err != nil {
			return GitLabWrapper{}, fmt.Errorf("could not get gitlab groups: %s", err)
		}
		for _, group := range groups {
			container = append(container, *group)
		}
		nextPage = resp.NextPage
	}
	var gid int = 0
	for _, group := range container {
		if group.FullPath == parentGroupName {
			gid = group.ID
		} else {
			continue
		}
	}
	if gid == 0 {
		return GitLabWrapper{}, fmt.Errorf("could not find gitlab group %s", parentGroupName)
	}

	// Get parent group path
	group, _, err := git.Groups.GetGroup(gid, &gitlab.GetGroupOptions{})
	if err != nil {
		return GitLabWrapper{}, fmt.Errorf("could not get gitlab parent group path: %s", err)
	}

	return GitLabWrapper{
		Client:          git,
		ParentGroupID:   gid,
		ParentGroupPath: group.FullPath,
	}, nil
}

// CheckProjectExists within a parent group
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

// GetProjectID returns a project's ID scoped to the parent group
func (gl *GitLabWrapper) GetProjectID(projectName string) (int, error) {
	container := make([]gitlab.Project, 0)
	for nextPage := 1; nextPage > 0; {
		projects, resp, err := gl.Client.Groups.ListGroupProjects(gl.ParentGroupID, &gitlab.ListGroupProjectsOptions{
			ListOptions: gitlab.ListOptions{
				Page:    nextPage,
				PerPage: 10,
			},
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
		if !strings.Contains(project.Name, "deleted") &&
			strings.ToLower(project.Name) == projectName {
			return project.ID, nil
		}
	}

	return 0, fmt.Errorf("could not get project ID for project %s", projectName)
}

// GetProjects for a specific parent group by ID
func (gl *GitLabWrapper) GetProjects() ([]gitlab.Project, error) {
	container := make([]gitlab.Project, 0)
	for nextPage := 1; nextPage > 0; {
		projects, resp, err := gl.Client.Groups.ListGroupProjects(gl.ParentGroupID, &gitlab.ListGroupProjectsOptions{
			ListOptions: gitlab.ListOptions{
				Page:    nextPage,
				PerPage: 10,
			},
		})
		if err != nil {
			return []gitlab.Project{}, err
		}
		for _, project := range projects {
			// Skip deleted projects
			if !strings.Contains(project.Name, "deleted") {
				container = append(container, *project)
			}
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
	log.Infof("created hook %s / %s", req.ProjectName, *req.CreateOpts.URL)

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
		return fmt.Errorf("no webhooks were found for project %s given search parameters", req.ProjectName)
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
		return fmt.Errorf("no webhooks were found for project %s given search parameters", req.ProjectName)
	}
	_, _, err = gl.Client.Projects.EditProjectHook(projectID, hookID, req.PatchOpts)
	if err != nil {
		return err
	}

	return nil
}
