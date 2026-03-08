package client

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_Health(t *testing.T) {
	client := NewClient(WithBaseURL("http://localhost:8080"))

	ctx := context.Background()
	health, err := client.Health(ctx)

	require.NoError(t, err)
	assert.Equal(t, "healthy", health.Status)
}

func TestClient_CreateRepository(t *testing.T) {
	client := NewClient(
		WithBaseURL("http://localhost:8080"),
		WithRequestID("test-request-123"),
	)

	ctx := context.Background()
	repo, err := client.CreateRepository(ctx, CreateRepositoryRequest{
		Name:        "test-repo",
		Description: strPtr("Test repository"),
	})

	require.NoError(t, err)
	assert.NotEmpty(t, repo.ID)
	assert.Equal(t, "test-repo", repo.Name)
}

func TestClient_ListRepositories(t *testing.T) {
	client := NewClient(WithBaseURL("http://localhost:8080"))

	ctx := context.Background()
	repos, err := client.ListRepositories(ctx)

	require.NoError(t, err)
	assert.NotNil(t, repos)
}

func TestClient_Workflow(t *testing.T) {
	client := NewClient(WithBaseURL("http://localhost:8080"))
	ctx := context.Background()

	repo, err := client.CreateRepository(ctx, CreateRepositoryRequest{
		Name: "workflow-test",
	})
	require.NoError(t, err)

	file, err := client.CreateFile(ctx, repo.ID, CreateFileRequest{
		Path:    "main.go",
		Content: []byte("package main\n\nfunc main() {}"),
	})
	require.NoError(t, err)

	commit, err := client.CreateCommit(ctx, repo.ID, CreateCommitRequest{
		AuthorID:   "test-user",
		AuthorType: "human",
		Message:    strPtr("Initial commit"),
		Files: []CommitFileInput{
			{
				FileID:    file.Path,
				Content:   []byte("package main\n\nfunc main() {}"),
				Operation: "create",
			},
		},
	})
	require.NoError(t, err)
	assert.NotEmpty(t, commit.ID)

	ws, err := client.CreateWorkspace(ctx, repo.ID, CreateWorkspaceRequest{
		Name:         "test-workspace",
		OwnerID:      "test-user",
		OwnerType:    "human",
		BaseCommitID: &commit.ID,
	})
	require.NoError(t, err)
	assert.NotEmpty(t, ws.ID)

	_, err = client.UpdateWorkspaceFile(ctx, repo.ID, ws.ID, UpdateWorkspaceFileRequest{
		FileID:    file.ID,
		Content:   []byte("package main\n\nfunc main() { println(\"updated\") }"),
		IsDeleted: false,
	})
	require.NoError(t, err)

	_, err = client.MergeWorkspace(ctx, repo.ID, ws.ID, MergeWorkspaceRequest{
		MergedBy: "test-user",
		Strategy: "three_way",
	})
	require.NoError(t, err)

	err = client.DeleteRepository(ctx, repo.ID)
	require.NoError(t, err)
}

func strPtr(s string) *string {
	return &s
}
