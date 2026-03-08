package client

import (
	"context"
	"time"
)

type Repository struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateRepositoryRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

type UpdateRepositoryRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

func (c *Client) ListRepositories(ctx context.Context) ([]Repository, error) {
	var repos []Repository
	err := c.get(ctx, "/api/v1/repos", &repos)
	return repos, err
}

func (c *Client) CreateRepository(ctx context.Context, req CreateRepositoryRequest) (*Repository, error) {
	var repo Repository
	err := c.post(ctx, "/api/v1/repos", req, &repo)
	return &repo, err
}

func (c *Client) GetRepository(ctx context.Context, repoID string) (*Repository, error) {
	var repo Repository
	err := c.get(ctx, "/api/v1/repos/"+repoID, &repo)
	return &repo, err
}

func (c *Client) UpdateRepository(ctx context.Context, repoID string, req UpdateRepositoryRequest) (*Repository, error) {
	var repo Repository
	err := c.put(ctx, "/api/v1/repos/"+repoID, req, &repo)
	return &repo, err
}

func (c *Client) DeleteRepository(ctx context.Context, repoID string) error {
	return c.delete(ctx, "/api/v1/repos/"+repoID)
}

type File struct {
	ID        string    `json:"id"`
	RepoID    string    `json:"repo_id"`
	Path      string    `json:"path"`
	IsDeleted bool      `json:"is_deleted"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type FileVersion struct {
	ID        string    `json:"id"`
	FileID    string    `json:"file_id"`
	Content   []byte    `json:"content"`
	Hash      string    `json:"hash"`
	SizeBytes int64     `json:"size_bytes"`
	CreatedAt time.Time `json:"created_at"`
}

type FileWithContent struct {
	File    File        `json:"file"`
	Content []byte      `json:"content"`
	Version FileVersion `json:"version"`
}

type CreateFileRequest struct {
	Path    string `json:"path"`
	Content []byte `json:"content"`
}

func (c *Client) ListFiles(ctx context.Context, repoID string, includeDeleted bool) ([]File, error) {
	var files []File
	path := "/api/v1/repos/" + repoID + "/files"
	if includeDeleted {
		path += "?include_deleted=true"
	}
	err := c.get(ctx, path, &files)
	return files, err
}

func (c *Client) CreateFile(ctx context.Context, repoID string, req CreateFileRequest) (*File, error) {
	var file File
	err := c.post(ctx, "/api/v1/repos/"+repoID+"/files", req, &file)
	return &file, err
}

func (c *Client) GetFile(ctx context.Context, fileID string, versionID *string) (*FileWithContent, error) {
	path := "/api/v1/files/" + fileID
	if versionID != nil {
		path += "?version=" + *versionID
	}
	var result FileWithContent
	err := c.get(ctx, path, &result)
	return &result, err
}

func (c *Client) ListFileVersions(ctx context.Context, fileID string) ([]FileVersion, error) {
	var versions []FileVersion
	err := c.get(ctx, "/api/v1/files/"+fileID+"/versions", &versions)
	return versions, err
}

type Commit struct {
	ID             string    `json:"id"`
	RepoID         string    `json:"repo_id"`
	AuthorID       string    `json:"author_id"`
	AuthorType     string    `json:"author_type"`
	Message        *string   `json:"message"`
	ParentCommitID *string   `json:"parent_commit_id"`
	CreatedAt      time.Time `json:"created_at"`
}

type CommitFileInput struct {
	FileID    string `json:"file_id"`
	Content   []byte `json:"content"`
	Operation string `json:"operation"`
}

type CreateCommitRequest struct {
	AuthorID   string            `json:"author_id"`
	AuthorType string            `json:"author_type"`
	Message    *string           `json:"message"`
	Files      []CommitFileInput `json:"files"`
}

type CommitWithFiles struct {
	Commit Commit       `json:"commit"`
	Files  []CommitFile `json:"files"`
}

type CommitFile struct {
	CommitID      string `json:"commit_id"`
	FileVersionID string `json:"file_version_id"`
	FileID        string `json:"file_id"`
	Operation     string `json:"operation"`
}

func (c *Client) ListCommits(ctx context.Context, repoID string) ([]Commit, error) {
	var commits []Commit
	err := c.get(ctx, "/api/v1/repos/"+repoID+"/commits", &commits)
	return commits, err
}

func (c *Client) CreateCommit(ctx context.Context, repoID string, req CreateCommitRequest) (*Commit, error) {
	var commit Commit
	err := c.post(ctx, "/api/v1/repos/"+repoID+"/commits", req, &commit)
	return &commit, err
}

func (c *Client) GetCommit(ctx context.Context, commitID string) (*CommitWithFiles, error) {
	var result CommitWithFiles
	err := c.get(ctx, "/api/v1/commits/"+commitID, &result)
	return &result, err
}

type Workspace struct {
	ID           string    `json:"id"`
	RepoID       string    `json:"repo_id"`
	Name         string    `json:"name"`
	OwnerID      string    `json:"owner_id"`
	OwnerType    string    `json:"owner_type"`
	BaseCommitID *string   `json:"base_commit_id"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type CreateWorkspaceRequest struct {
	Name         string  `json:"name"`
	OwnerID      string  `json:"owner_id"`
	OwnerType    string  `json:"owner_type"`
	BaseCommitID *string `json:"base_commit_id"`
}

type MergeWorkspaceRequest struct {
	MergedBy string `json:"merged_by"`
	Strategy string `json:"strategy"`
}

func (c *Client) ListWorkspaces(ctx context.Context, repoID string, status *string) ([]Workspace, error) {
	path := "/api/v1/repos/" + repoID + "/workspaces"
	if status != nil {
		path += "?status=" + *status
	}
	var workspaces []Workspace
	err := c.get(ctx, path, &workspaces)
	return workspaces, err
}

func (c *Client) CreateWorkspace(ctx context.Context, repoID string, req CreateWorkspaceRequest) (*Workspace, error) {
	var ws Workspace
	err := c.post(ctx, "/api/v1/repos/"+repoID+"/workspaces", req, &ws)
	return &ws, err
}

func (c *Client) GetWorkspace(ctx context.Context, repoID, workspaceID string) (*Workspace, error) {
	var ws Workspace
	err := c.get(ctx, "/api/v1/repos/"+repoID+"/workspaces/"+workspaceID, &ws)
	return &ws, err
}

func (c *Client) MergeWorkspace(ctx context.Context, repoID, workspaceID string, req MergeWorkspaceRequest) (*Commit, error) {
	var commit Commit
	err := c.post(ctx, "/api/v1/repos/"+repoID+"/workspaces/"+workspaceID+"/merge", req, &commit)
	return &commit, err
}

func (c *Client) AbandonWorkspace(ctx context.Context, repoID, workspaceID string) (*Workspace, error) {
	var ws Workspace
	err := c.post(ctx, "/api/v1/repos/"+repoID+"/workspaces/"+workspaceID+"/abandon", nil, &ws)
	return &ws, err
}

type WorkspaceFile struct {
	ID          string    `json:"id"`
	WorkspaceID string    `json:"workspace_id"`
	FileID      string    `json:"file_id"`
	Content     []byte    `json:"content"`
	Hash        string    `json:"hash"`
	IsDeleted   bool      `json:"is_deleted"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type UpdateWorkspaceFileRequest struct {
	FileID    string `json:"file_id"`
	Content   []byte `json:"content"`
	IsDeleted bool   `json:"is_deleted"`
}

func (c *Client) ListWorkspaceFiles(ctx context.Context, repoID, workspaceID string) ([]WorkspaceFile, error) {
	var files []WorkspaceFile
	err := c.get(ctx, "/api/v1/repos/"+repoID+"/workspaces/"+workspaceID+"/files", &files)
	return files, err
}

func (c *Client) UpdateWorkspaceFile(ctx context.Context, repoID, workspaceID string, req UpdateWorkspaceFileRequest) (*WorkspaceFile, error) {
	var file WorkspaceFile
	err := c.post(ctx, "/api/v1/repos/"+repoID+"/workspaces/"+workspaceID+"/files", req, &file)
	return &file, err
}

func (c *Client) DeleteWorkspaceFile(ctx context.Context, repoID, workspaceID, fileID string) error {
	return c.delete(ctx, "/api/v1/repos/"+repoID+"/workspaces/"+workspaceID+"/files/"+fileID)
}
