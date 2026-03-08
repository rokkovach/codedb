package db

import (
	"context"
	"database/sql"
	"time"
)

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

type Lease struct {
	ID          string    `json:"id"`
	WorkspaceID string    `json:"workspace_id"`
	FileID      *string   `json:"file_id"`
	PathPattern *string   `json:"path_pattern"`
	OwnerID     string    `json:"owner_id"`
	Intent      *string   `json:"intent"`
	ExpiresAt   time.Time `json:"expires_at"`
	CreatedAt   time.Time `json:"created_at"`
}

type Lock struct {
	ID          string     `json:"id"`
	RepoID      string     `json:"repo_id"`
	FileID      *string    `json:"file_id"`
	PathPattern *string    `json:"path_pattern"`
	OwnerID     string     `json:"owner_id"`
	OwnerType   string     `json:"owner_type"`
	LockType    string     `json:"lock_type"`
	Reason      *string    `json:"reason"`
	ExpiresAt   *time.Time `json:"expires_at"`
	CreatedAt   time.Time  `json:"created_at"`
}

type MergeHistory struct {
	ID                string                 `json:"id"`
	WorkspaceID       string                 `json:"workspace_id"`
	MergeCommitID     string                 `json:"merge_commit_id"`
	MergedBy          string                 `json:"merged_by"`
	MergeStrategy     string                 `json:"merge_strategy"`
	ConflictsResolved map[string]interface{} `json:"conflicts_resolved"`
	CreatedAt         time.Time              `json:"created_at"`
}

type WorkspaceQueries struct {
	db *DB
}

func NewWorkspaceQueries(db *DB) *WorkspaceQueries {
	return &WorkspaceQueries{db: db}
}

func (q *WorkspaceQueries) Create(ctx context.Context, repoID, name, ownerID, ownerType string, baseCommitID *string) (*Workspace, error) {
	var ws Workspace
	err := q.db.Pool().QueryRow(ctx, `
		INSERT INTO workspaces (repo_id, name, owner_id, owner_type, base_commit_id)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, repo_id, name, owner_id, owner_type, base_commit_id, status, created_at, updated_at
	`, repoID, name, ownerID, ownerType, baseCommitID).Scan(
		&ws.ID, &ws.RepoID, &ws.Name, &ws.OwnerID, &ws.OwnerType, &ws.BaseCommitID,
		&ws.Status, &ws.CreatedAt, &ws.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &ws, nil
}

func (q *WorkspaceQueries) Get(ctx context.Context, id string) (*Workspace, error) {
	var ws Workspace
	err := q.db.Pool().QueryRow(ctx, `
		SELECT id, repo_id, name, owner_id, owner_type, base_commit_id, status, created_at, updated_at
		FROM workspaces WHERE id = $1
	`, id).Scan(
		&ws.ID, &ws.RepoID, &ws.Name, &ws.OwnerID, &ws.OwnerType, &ws.BaseCommitID,
		&ws.Status, &ws.CreatedAt, &ws.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &ws, nil
}

func (q *WorkspaceQueries) GetByName(ctx context.Context, repoID, name string) (*Workspace, error) {
	var ws Workspace
	err := q.db.Pool().QueryRow(ctx, `
		SELECT id, repo_id, name, owner_id, owner_type, base_commit_id, status, created_at, updated_at
		FROM workspaces WHERE repo_id = $1 AND name = $2
	`, repoID, name).Scan(
		&ws.ID, &ws.RepoID, &ws.Name, &ws.OwnerID, &ws.OwnerType, &ws.BaseCommitID,
		&ws.Status, &ws.CreatedAt, &ws.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &ws, nil
}

func (q *WorkspaceQueries) ListByRepo(ctx context.Context, repoID string, status *string) ([]Workspace, error) {
	query := `SELECT id, repo_id, name, owner_id, owner_type, base_commit_id, status, created_at, updated_at
		FROM workspaces WHERE repo_id = $1`
	args := []interface{}{repoID}
	if status != nil {
		query += " AND status = $2"
		args = append(args, *status)
	}
	query += " ORDER BY created_at DESC"

	rows, err := q.db.Pool().Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var workspaces []Workspace
	for rows.Next() {
		var ws Workspace
		if err := rows.Scan(&ws.ID, &ws.RepoID, &ws.Name, &ws.OwnerID, &ws.OwnerID, &ws.OwnerType,
			&ws.BaseCommitID, &ws.Status, &ws.CreatedAt, &ws.UpdatedAt); err != nil {
			return nil, err
		}
		workspaces = append(workspaces, ws)
	}
	return workspaces, rows.Err()
}

func (q *WorkspaceQueries) UpdateStatus(ctx context.Context, id, status string) (*Workspace, error) {
	var ws Workspace
	err := q.db.Pool().QueryRow(ctx, `
		UPDATE workspaces SET status = $2 WHERE id = $1
		RETURNING id, repo_id, name, owner_id, owner_type, base_commit_id, status, created_at, updated_at
	`, id, status).Scan(
		&ws.ID, &ws.RepoID, &ws.Name, &ws.OwnerID, &ws.OwnerType, &ws.BaseCommitID,
		&ws.Status, &ws.CreatedAt, &ws.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &ws, nil
}

func (q *WorkspaceQueries) Delete(ctx context.Context, id string) error {
	_, err := q.db.Pool().Exec(ctx, `DELETE FROM workspaces WHERE id = $1`, id)
	return err
}

type WorkspaceFileQueries struct {
	db *DB
}

func NewWorkspaceFileQueries(db *DB) *WorkspaceFileQueries {
	return &WorkspaceFileQueries{db: db}
}

func (q *WorkspaceFileQueries) Upsert(ctx context.Context, workspaceID, fileID string, content []byte, isDeleted bool) (*WorkspaceFile, error) {
	hash := hashContent(content)
	var wf WorkspaceFile
	err := q.db.Pool().QueryRow(ctx, `
		INSERT INTO workspace_files (workspace_id, file_id, content, hash, is_deleted)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (workspace_id, file_id) DO UPDATE SET
			content = EXCLUDED.content,
			hash = EXCLUDED.hash,
			is_deleted = EXCLUDED.is_deleted,
			updated_at = now()
		RETURNING id, workspace_id, file_id, content, hash, is_deleted, created_at, updated_at
	`, workspaceID, fileID, content, hash, isDeleted).Scan(
		&wf.ID, &wf.WorkspaceID, &wf.FileID, &wf.Content, &wf.Hash, &wf.IsDeleted,
		&wf.CreatedAt, &wf.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &wf, nil
}

func (q *WorkspaceFileQueries) Get(ctx context.Context, workspaceID, fileID string) (*WorkspaceFile, error) {
	var wf WorkspaceFile
	err := q.db.Pool().QueryRow(ctx, `
		SELECT id, workspace_id, file_id, content, hash, is_deleted, created_at, updated_at
		FROM workspace_files WHERE workspace_id = $1 AND file_id = $2
	`, workspaceID, fileID).Scan(
		&wf.ID, &wf.WorkspaceID, &wf.FileID, &wf.Content, &wf.Hash, &wf.IsDeleted,
		&wf.CreatedAt, &wf.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &wf, nil
}

func (q *WorkspaceFileQueries) ListByWorkspace(ctx context.Context, workspaceID string) ([]WorkspaceFile, error) {
	rows, err := q.db.Pool().Query(ctx, `
		SELECT id, workspace_id, file_id, content, hash, is_deleted, created_at, updated_at
		FROM workspace_files WHERE workspace_id = $1 ORDER BY file_id
	`, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []WorkspaceFile
	for rows.Next() {
		var f WorkspaceFile
		if err := rows.Scan(&f.ID, &f.WorkspaceID, &f.FileID, &f.Content, &f.Hash, &f.IsDeleted,
			&f.CreatedAt, &f.UpdatedAt); err != nil {
			return nil, err
		}
		files = append(files, f)
	}
	return files, rows.Err()
}

func (q *WorkspaceFileQueries) Delete(ctx context.Context, workspaceID, fileID string) error {
	_, err := q.db.Pool().Exec(ctx, `
		DELETE FROM workspace_files WHERE workspace_id = $1 AND file_id = $2
	`, workspaceID, fileID)
	return err
}

type LeaseQueries struct {
	db *DB
}

func NewLeaseQueries(db *DB) *LeaseQueries {
	return &LeaseQueries{db: db}
}

func (q *LeaseQueries) Create(ctx context.Context, workspaceID string, fileID *string, pathPattern *string, ownerID string, intent *string, ttl time.Duration) (*Lease, error) {
	expiresAt := time.Now().Add(ttl)
	var lease Lease
	err := q.db.Pool().QueryRow(ctx, `
		INSERT INTO leases (workspace_id, file_id, path_pattern, owner_id, intent, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, workspace_id, file_id, path_pattern, owner_id, intent, expires_at, created_at
	`, workspaceID, fileID, pathPattern, ownerID, intent, expiresAt).Scan(
		&lease.ID, &lease.WorkspaceID, &lease.FileID, &lease.PathPattern, &lease.OwnerID,
		&lease.Intent, &lease.ExpiresAt, &lease.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &lease, nil
}

func (q *LeaseQueries) Get(ctx context.Context, id string) (*Lease, error) {
	var lease Lease
	err := q.db.Pool().QueryRow(ctx, `
		SELECT id, workspace_id, file_id, path_pattern, owner_id, intent, expires_at, created_at
		FROM leases WHERE id = $1
	`, id).Scan(
		&lease.ID, &lease.WorkspaceID, &lease.FileID, &lease.PathPattern, &lease.OwnerID,
		&lease.Intent, &lease.ExpiresAt, &lease.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &lease, nil
}

func (q *LeaseQueries) ListByWorkspace(ctx context.Context, workspaceID string) ([]Lease, error) {
	rows, err := q.db.Pool().Query(ctx, `
		SELECT id, workspace_id, file_id, path_pattern, owner_id, intent, expires_at, created_at
		FROM leases WHERE workspace_id = $1 AND expires_at > now()
		ORDER BY created_at DESC
	`, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var leases []Lease
	for rows.Next() {
		var l Lease
		if err := rows.Scan(&l.ID, &l.WorkspaceID, &l.FileID, &l.PathPattern, &l.OwnerID,
			&l.Intent, &l.ExpiresAt, &l.CreatedAt); err != nil {
			return nil, err
		}
		leases = append(leases, l)
	}
	return leases, rows.Err()
}

func (q *LeaseQueries) Renew(ctx context.Context, id string, ttl time.Duration) (*Lease, error) {
	expiresAt := time.Now().Add(ttl)
	var lease Lease
	err := q.db.Pool().QueryRow(ctx, `
		UPDATE leases SET expires_at = $2 WHERE id = $1
		RETURNING id, workspace_id, file_id, path_pattern, owner_id, intent, expires_at, created_at
	`, id, expiresAt).Scan(
		&lease.ID, &lease.WorkspaceID, &lease.FileID, &lease.PathPattern, &lease.OwnerID,
		&lease.Intent, &lease.ExpiresAt, &lease.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &lease, nil
}

func (q *LeaseQueries) Release(ctx context.Context, id string) error {
	_, err := q.db.Pool().Exec(ctx, `DELETE FROM leases WHERE id = $1`, id)
	return err
}

func (q *LeaseQueries) CheckConflict(ctx context.Context, workspaceID string, fileID *string, pathPattern *string, ownerID string) ([]Lease, error) {
	query := `SELECT id, workspace_id, file_id, path_pattern, owner_id, intent, expires_at, created_at
		FROM leases WHERE workspace_id = $1 AND owner_id != $2 AND expires_at > now()`
	args := []interface{}{workspaceID, ownerID}

	if fileID != nil {
		query += " AND (file_id = $3 OR path_pattern IS NOT NULL)"
		args = append(args, *fileID)
	} else if pathPattern != nil {
		query += " AND (path_pattern IS NOT NULL OR file_id IS NOT NULL)"
	}

	rows, err := q.db.Pool().Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var leases []Lease
	for rows.Next() {
		var l Lease
		if err := rows.Scan(&l.ID, &l.WorkspaceID, &l.FileID, &l.PathPattern, &l.OwnerID,
			&l.Intent, &l.ExpiresAt, &l.CreatedAt); err != nil {
			return nil, err
		}
		leases = append(leases, l)
	}
	return leases, rows.Err()
}

type LockQueries struct {
	db *DB
}

func NewLockQueries(db *DB) *LockQueries {
	return &LockQueries{db: db}
}

func (q *LockQueries) Create(ctx context.Context, repoID string, fileID *string, pathPattern *string, ownerID, ownerType, lockType string, reason *string, expiresAt *time.Time) (*Lock, error) {
	var lock Lock
	err := q.db.Pool().QueryRow(ctx, `
		INSERT INTO locks (repo_id, file_id, path_pattern, owner_id, owner_type, lock_type, reason, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, repo_id, file_id, path_pattern, owner_id, owner_type, lock_type, reason, expires_at, created_at
	`, repoID, fileID, pathPattern, ownerID, ownerType, lockType, reason, expiresAt).Scan(
		&lock.ID, &lock.RepoID, &lock.FileID, &lock.PathPattern, &lock.OwnerID, &lock.OwnerType,
		&lock.LockType, &lock.Reason, &lock.ExpiresAt, &lock.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &lock, nil
}

func (q *LockQueries) Get(ctx context.Context, id string) (*Lock, error) {
	var lock Lock
	var expiresAt sql.NullTime
	err := q.db.Pool().QueryRow(ctx, `
		SELECT id, repo_id, file_id, path_pattern, owner_id, owner_type, lock_type, reason, expires_at, created_at
		FROM locks WHERE id = $1
	`, id).Scan(
		&lock.ID, &lock.RepoID, &lock.FileID, &lock.PathPattern, &lock.OwnerID, &lock.OwnerType,
		&lock.LockType, &lock.Reason, &expiresAt, &lock.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	if expiresAt.Valid {
		lock.ExpiresAt = &expiresAt.Time
	}
	return &lock, nil
}

func (q *LockQueries) ListByRepo(ctx context.Context, repoID string) ([]Lock, error) {
	rows, err := q.db.Pool().Query(ctx, `
		SELECT id, repo_id, file_id, path_pattern, owner_id, owner_type, lock_type, reason, expires_at, created_at
		FROM locks WHERE repo_id = $1 AND (expires_at IS NULL OR expires_at > now())
		ORDER BY created_at DESC
	`, repoID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var locks []Lock
	for rows.Next() {
		var l Lock
		var expiresAt sql.NullTime
		if err := rows.Scan(&l.ID, &l.RepoID, &l.FileID, &l.PathPattern, &l.OwnerID, &l.OwnerType,
			&l.LockType, &l.Reason, &expiresAt, &l.CreatedAt); err != nil {
			return nil, err
		}
		if expiresAt.Valid {
			l.ExpiresAt = &expiresAt.Time
		}
		locks = append(locks, l)
	}
	return locks, rows.Err()
}

func (q *LockQueries) Release(ctx context.Context, id string) error {
	_, err := q.db.Pool().Exec(ctx, `DELETE FROM locks WHERE id = $1`, id)
	return err
}

func (q *LockQueries) CheckConflict(ctx context.Context, repoID string, fileID *string, pathPattern *string, ownerID string) ([]Lock, error) {
	query := `SELECT id, repo_id, file_id, path_pattern, owner_id, owner_type, lock_type, reason, expires_at, created_at
		FROM locks WHERE repo_id = $1 AND owner_id != $2 AND (expires_at IS NULL OR expires_at > now())`
	args := []interface{}{repoID, ownerID}

	if fileID != nil {
		query += " AND (file_id = $3 OR path_pattern IS NOT NULL)"
		args = append(args, *fileID)
	} else if pathPattern != nil {
		query += " AND (path_pattern IS NOT NULL OR file_id IS NOT NULL)"
	}

	rows, err := q.db.Pool().Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var locks []Lock
	for rows.Next() {
		var l Lock
		var expiresAt sql.NullTime
		if err := rows.Scan(&l.ID, &l.RepoID, &l.FileID, &l.PathPattern, &l.OwnerID, &l.OwnerType,
			&l.LockType, &l.Reason, &expiresAt, &l.CreatedAt); err != nil {
			return nil, err
		}
		if expiresAt.Valid {
			l.ExpiresAt = &expiresAt.Time
		}
		locks = append(locks, l)
	}
	return locks, rows.Err()
}

type MergeHistoryQueries struct {
	db *DB
}

func NewMergeHistoryQueries(db *DB) *MergeHistoryQueries {
	return &MergeHistoryQueries{db: db}
}

func (q *MergeHistoryQueries) Create(ctx context.Context, workspaceID, mergeCommitID, mergedBy, mergeStrategy string, conflictsResolved map[string]interface{}) (*MergeHistory, error) {
	var mh MergeHistory
	err := q.db.Pool().QueryRow(ctx, `
		INSERT INTO merge_history (workspace_id, merge_commit_id, merged_by, merge_strategy, conflicts_resolved)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, workspace_id, merge_commit_id, merged_by, merge_strategy, conflicts_resolved, created_at
	`, workspaceID, mergeCommitID, mergedBy, mergeStrategy, conflictsResolved).Scan(
		&mh.ID, &mh.WorkspaceID, &mh.MergeCommitID, &mh.MergedBy, &mh.MergeStrategy,
		&mh.ConflictsResolved, &mh.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &mh, nil
}

func (q *MergeHistoryQueries) ListByWorkspace(ctx context.Context, workspaceID string) ([]MergeHistory, error) {
	rows, err := q.db.Pool().Query(ctx, `
		SELECT id, workspace_id, merge_commit_id, merged_by, merge_strategy, conflicts_resolved, created_at
		FROM merge_history WHERE workspace_id = $1 ORDER BY created_at DESC
	`, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []MergeHistory
	for rows.Next() {
		var h MergeHistory
		if err := rows.Scan(&h.ID, &h.WorkspaceID, &h.MergeCommitID, &h.MergedBy, &h.MergeStrategy,
			&h.ConflictsResolved, &h.CreatedAt); err != nil {
			return nil, err
		}
		history = append(history, h)
	}
	return history, rows.Err()
}
