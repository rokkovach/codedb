package api

import (
	"context"
	"errors"
	"time"

	"github.com/rokkovach/codedb/internal/db"
)

type WorkspaceService struct {
	workspaceQueries *db.WorkspaceQueries
	wsFileQueries    *db.WorkspaceFileQueries
	leaseQueries     *db.LeaseQueries
	lockQueries      *db.LockQueries
	mergeQueries     *db.MergeHistoryQueries
	commitQueries    *db.CommitQueries
	fileQueries      *db.FileQueries
	auditQueries     *db.AuditLogQueries
}

func NewWorkspaceService(database *db.DB) *WorkspaceService {
	return &WorkspaceService{
		workspaceQueries: db.NewWorkspaceQueries(database),
		wsFileQueries:    db.NewWorkspaceFileQueries(database),
		leaseQueries:     db.NewLeaseQueries(database),
		lockQueries:      db.NewLockQueries(database),
		mergeQueries:     db.NewMergeHistoryQueries(database),
		commitQueries:    db.NewCommitQueries(database),
		fileQueries:      db.NewFileQueries(database),
		auditQueries:     db.NewAuditLogQueries(database),
	}
}

func (s *WorkspaceService) CreateWorkspace(ctx context.Context, repoID, name, ownerID, ownerType string, baseCommitID *string) (*db.Workspace, error) {
	return s.workspaceQueries.Create(ctx, repoID, name, ownerID, ownerType, baseCommitID)
}

func (s *WorkspaceService) GetWorkspace(ctx context.Context, id string) (*db.Workspace, error) {
	return s.workspaceQueries.Get(ctx, id)
}

func (s *WorkspaceService) ListWorkspaces(ctx context.Context, repoID string, status *string) ([]db.Workspace, error) {
	return s.workspaceQueries.ListByRepo(ctx, repoID, status)
}

func (s *WorkspaceService) AbandonWorkspace(ctx context.Context, id string) (*db.Workspace, error) {
	return s.workspaceQueries.UpdateStatus(ctx, id, "abandoned")
}

type MergeConflict struct {
	FileID string `json:"file_id"`
	Path   string `json:"path"`
	Base   []byte `json:"base"`
	Ours   []byte `json:"ours"`
	Theirs []byte `json:"theirs"`
}

func (s *WorkspaceService) MergeWorkspace(ctx context.Context, workspaceID, mergedBy, strategy string) (*db.Commit, []MergeConflict, error) {
	ws, err := s.workspaceQueries.Get(ctx, workspaceID)
	if err != nil {
		return nil, nil, err
	}

	if ws.Status != "active" {
		return nil, nil, errors.New("workspace is not active")
	}

	conflicts, err := s.detectConflicts(ctx, ws)
	if err != nil {
		return nil, nil, err
	}

	if len(conflicts) > 0 && strategy != "force" {
		return nil, conflicts, errors.New("merge conflicts detected")
	}

	commit, err := s.createMergeCommit(ctx, ws, mergedBy)
	if err != nil {
		return nil, nil, err
	}

	_, err = s.mergeQueries.Create(ctx, workspaceID, commit.ID, mergedBy, strategy, nil)
	if err != nil {
		return nil, nil, err
	}

	_, err = s.workspaceQueries.UpdateStatus(ctx, workspaceID, "merged")
	if err != nil {
		return nil, nil, err
	}

	return commit, nil, nil
}

func (s *WorkspaceService) detectConflicts(ctx context.Context, ws *db.Workspace) ([]MergeConflict, error) {
	wsFiles, err := s.wsFileQueries.ListByWorkspace(ctx, ws.ID)
	if err != nil {
		return nil, err
	}

	var conflicts []MergeConflict

	for _, wsFile := range wsFiles {
		file, err := s.fileQueries.Get(ctx, wsFile.FileID)
		if err != nil {
			continue
		}

		latestVersion, err := s.fileQueries.GetLatestVersion(ctx, wsFile.FileID)
		if err != nil {
			continue
		}

		if string(latestVersion.Content) != string(wsFile.Content) {
			conflicts = append(conflicts, MergeConflict{
				FileID: wsFile.FileID,
				Path:   file.Path,
				Base:   latestVersion.Content,
				Ours:   wsFile.Content,
				Theirs: latestVersion.Content,
			})
		}
	}

	return conflicts, nil
}

func (s *WorkspaceService) createMergeCommit(ctx context.Context, ws *db.Workspace, mergedBy string) (*db.Commit, error) {
	wsFiles, err := s.wsFileQueries.ListByWorkspace(ctx, ws.ID)
	if err != nil {
		return nil, err
	}

	var commitFiles []db.CommitFileInput
	for _, f := range wsFiles {
		if f.IsDeleted {
			commitFiles = append(commitFiles, db.CommitFileInput{
				FileID:    f.FileID,
				Content:   nil,
				Operation: "delete",
			})
		} else {
			commitFiles = append(commitFiles, db.CommitFileInput{
				FileID:    f.FileID,
				Content:   f.Content,
				Operation: "update",
			})
		}
	}

	message := "Merge workspace: " + ws.Name
	return s.commitQueries.Create(ctx, ws.RepoID, mergedBy, "human", &message, ws.BaseCommitID, commitFiles)
}

func (s *WorkspaceService) GetFile(ctx context.Context, workspaceID, fileID string) (*db.WorkspaceFile, error) {
	return s.wsFileQueries.Get(ctx, workspaceID, fileID)
}

func (s *WorkspaceService) ListFiles(ctx context.Context, workspaceID string) ([]db.WorkspaceFile, error) {
	return s.wsFileQueries.ListByWorkspace(ctx, workspaceID)
}

func (s *WorkspaceService) UpsertFile(ctx context.Context, workspaceID, fileID string, content []byte, isDeleted bool) (*db.WorkspaceFile, error) {
	return s.wsFileQueries.Upsert(ctx, workspaceID, fileID, content, isDeleted)
}

func (s *WorkspaceService) DeleteFile(ctx context.Context, workspaceID, fileID string) error {
	return s.wsFileQueries.Delete(ctx, workspaceID, fileID)
}

func (s *WorkspaceService) AcquireLease(ctx context.Context, workspaceID string, fileID *string, pathPattern *string, ownerID string, intent *string, ttl time.Duration) (*db.Lease, error) {
	conflicts, err := s.leaseQueries.CheckConflict(ctx, workspaceID, fileID, pathPattern, ownerID)
	if err != nil {
		return nil, err
	}
	if len(conflicts) > 0 {
		return nil, errors.New("lease conflict detected")
	}

	return s.leaseQueries.Create(ctx, workspaceID, fileID, pathPattern, ownerID, intent, ttl)
}

func (s *WorkspaceService) RenewLease(ctx context.Context, leaseID string, ttl time.Duration) (*db.Lease, error) {
	return s.leaseQueries.Renew(ctx, leaseID, ttl)
}

func (s *WorkspaceService) ReleaseLease(ctx context.Context, leaseID string) error {
	return s.leaseQueries.Release(ctx, leaseID)
}

func (s *WorkspaceService) ListLeases(ctx context.Context, workspaceID string) ([]db.Lease, error) {
	return s.leaseQueries.ListByWorkspace(ctx, workspaceID)
}

func (s *WorkspaceService) AcquireLock(ctx context.Context, repoID string, fileID *string, pathPattern *string, ownerID, ownerType, lockType string, reason *string, expiresAt *time.Time) (*db.Lock, error) {
	conflicts, err := s.lockQueries.CheckConflict(ctx, repoID, fileID, pathPattern, ownerID)
	if err != nil {
		return nil, err
	}
	if len(conflicts) > 0 {
		return nil, errors.New("lock conflict detected")
	}

	return s.lockQueries.Create(ctx, repoID, fileID, pathPattern, ownerID, ownerType, lockType, reason, expiresAt)
}

func (s *WorkspaceService) ReleaseLock(ctx context.Context, lockID string) error {
	return s.lockQueries.Release(ctx, lockID)
}

func (s *WorkspaceService) ListLocks(ctx context.Context, repoID string) ([]db.Lock, error) {
	return s.lockQueries.ListByRepo(ctx, repoID)
}
