package db

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
)

type RepositoryQueries struct {
	db *DB
}

func NewRepositoryQueries(db *DB) *RepositoryQueries {
	return &RepositoryQueries{db: db}
}

func (q *RepositoryQueries) Create(ctx context.Context, name string, description *string) (*Repository, error) {
	var repo Repository
	err := q.db.Pool().QueryRow(ctx, `
		INSERT INTO repositories (name, description)
		VALUES ($1, $2)
		RETURNING id, name, description, created_at, updated_at
	`, name, description).Scan(&repo.ID, &repo.Name, &repo.Description, &repo.CreatedAt, &repo.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &repo, nil
}

func (q *RepositoryQueries) Get(ctx context.Context, id string) (*Repository, error) {
	var repo Repository
	err := q.db.Pool().QueryRow(ctx, `
		SELECT id, name, description, created_at, updated_at
		FROM repositories WHERE id = $1
	`, id).Scan(&repo.ID, &repo.Name, &repo.Description, &repo.CreatedAt, &repo.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &repo, nil
}

func (q *RepositoryQueries) GetByName(ctx context.Context, name string) (*Repository, error) {
	var repo Repository
	err := q.db.Pool().QueryRow(ctx, `
		SELECT id, name, description, created_at, updated_at
		FROM repositories WHERE name = $1
	`, name).Scan(&repo.ID, &repo.Name, &repo.Description, &repo.CreatedAt, &repo.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &repo, nil
}

func (q *RepositoryQueries) List(ctx context.Context, limit, offset int) ([]Repository, error) {
	rows, err := q.db.Pool().Query(ctx, `
		SELECT id, name, description, created_at, updated_at
		FROM repositories ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var repos []Repository
	for rows.Next() {
		var repo Repository
		if err := rows.Scan(&repo.ID, &repo.Name, &repo.Description, &repo.CreatedAt, &repo.UpdatedAt); err != nil {
			return nil, err
		}
		repos = append(repos, repo)
	}
	return repos, rows.Err()
}

func (q *RepositoryQueries) Update(ctx context.Context, id string, name *string, description *string) (*Repository, error) {
	var repo Repository
	err := q.db.Pool().QueryRow(ctx, `
		UPDATE repositories
		SET name = COALESCE($2, name), description = COALESCE($3, description)
		WHERE id = $1
		RETURNING id, name, description, created_at, updated_at
	`, id, name, description).Scan(&repo.ID, &repo.Name, &repo.Description, &repo.CreatedAt, &repo.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &repo, nil
}

func (q *RepositoryQueries) Delete(ctx context.Context, id string) error {
	_, err := q.db.Pool().Exec(ctx, `DELETE FROM repositories WHERE id = $1`, id)
	return err
}

type FileQueries struct {
	db *DB
}

func NewFileQueries(db *DB) *FileQueries {
	return &FileQueries{db: db}
}

func (q *FileQueries) Create(ctx context.Context, repoID, path string) (*File, error) {
	var file File
	err := q.db.Pool().QueryRow(ctx, `
		INSERT INTO files (repo_id, path)
		VALUES ($1, $2)
		RETURNING id, repo_id, path, is_deleted, created_at, updated_at
	`, repoID, path).Scan(&file.ID, &file.RepoID, &file.Path, &file.IsDeleted, &file.CreatedAt, &file.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &file, nil
}

func (q *FileQueries) Get(ctx context.Context, id string) (*File, error) {
	var file File
	err := q.db.Pool().QueryRow(ctx, `
		SELECT id, repo_id, path, is_deleted, created_at, updated_at
		FROM files WHERE id = $1
	`, id).Scan(&file.ID, &file.RepoID, &file.Path, &file.IsDeleted, &file.CreatedAt, &file.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &file, nil
}

func (q *FileQueries) GetByPath(ctx context.Context, repoID, path string) (*File, error) {
	var file File
	err := q.db.Pool().QueryRow(ctx, `
		SELECT id, repo_id, path, is_deleted, created_at, updated_at
		FROM files WHERE repo_id = $1 AND path = $2
	`, repoID, path).Scan(&file.ID, &file.RepoID, &file.Path, &file.IsDeleted, &file.CreatedAt, &file.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &file, nil
}

func (q *FileQueries) ListByRepo(ctx context.Context, repoID string, includeDeleted bool) ([]File, error) {
	query := `
		SELECT id, repo_id, path, is_deleted, created_at, updated_at
		FROM files WHERE repo_id = $1
	`
	if !includeDeleted {
		query += " AND is_deleted = false"
	}
	query += " ORDER BY path"

	rows, err := q.db.Pool().Query(ctx, query, repoID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []File
	for rows.Next() {
		var file File
		if err := rows.Scan(&file.ID, &file.RepoID, &file.Path, &file.IsDeleted, &file.CreatedAt, &file.UpdatedAt); err != nil {
			return nil, err
		}
		files = append(files, file)
	}
	return files, rows.Err()
}

func (q *FileQueries) MarkDeleted(ctx context.Context, id string) error {
	_, err := q.db.Pool().Exec(ctx, `
		UPDATE files SET is_deleted = true WHERE id = $1
	`, id)
	return err
}

func (q *FileQueries) CreateVersion(ctx context.Context, fileID string, content []byte) (*FileVersion, error) {
	hash := sha256.Sum256(content)
	hashStr := hex.EncodeToString(hash[:])

	var version FileVersion
	err := q.db.Pool().QueryRow(ctx, `
		INSERT INTO file_versions (file_id, content, hash, size_bytes)
		VALUES ($1, $2, $3, $4)
		RETURNING id, file_id, content, hash, size_bytes, created_at
	`, fileID, content, hashStr, len(content)).Scan(
		&version.ID, &version.FileID, &version.Content, &version.Hash, &version.SizeBytes, &version.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &version, nil
}

func (q *FileQueries) GetLatestVersion(ctx context.Context, fileID string) (*FileVersion, error) {
	var version FileVersion
	err := q.db.Pool().QueryRow(ctx, `
		SELECT id, file_id, content, hash, size_bytes, created_at
		FROM file_versions
		WHERE file_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`, fileID).Scan(
		&version.ID, &version.FileID, &version.Content, &version.Hash, &version.SizeBytes, &version.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &version, nil
}

func (q *FileQueries) GetVersion(ctx context.Context, versionID string) (*FileVersion, error) {
	var version FileVersion
	err := q.db.Pool().QueryRow(ctx, `
		SELECT id, file_id, content, hash, size_bytes, created_at
		FROM file_versions WHERE id = $1
	`, versionID).Scan(
		&version.ID, &version.FileID, &version.Content, &version.Hash, &version.SizeBytes, &version.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &version, nil
}

func (q *FileQueries) ListVersions(ctx context.Context, fileID string, limit int) ([]FileVersion, error) {
	rows, err := q.db.Pool().Query(ctx, `
		SELECT id, file_id, content, hash, size_bytes, created_at
		FROM file_versions
		WHERE file_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`, fileID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var versions []FileVersion
	for rows.Next() {
		var v FileVersion
		if err := rows.Scan(&v.ID, &v.FileID, &v.Content, &v.Hash, &v.SizeBytes, &v.CreatedAt); err != nil {
			return nil, err
		}
		versions = append(versions, v)
	}
	return versions, rows.Err()
}

type CommitQueries struct {
	db *DB
}

func NewCommitQueries(db *DB) *CommitQueries {
	return &CommitQueries{db: db}
}

type CommitFileInput struct {
	FileID    string
	Content   []byte
	Operation string
}

func (q *CommitQueries) Create(ctx context.Context, repoID, authorID, authorType string, message *string, parentCommitID *string, files []CommitFileInput) (*Commit, error) {
	tx, err := q.db.Pool().Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var commit Commit
	err = tx.QueryRow(ctx, `
		INSERT INTO commits (repo_id, author_id, author_type, message, parent_commit_id)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, repo_id, author_id, author_type, message, parent_commit_id, created_at
	`, repoID, authorID, authorType, message, parentCommitID).Scan(
		&commit.ID, &commit.RepoID, &commit.AuthorID, &commit.AuthorType, &commit.Message,
		&commit.ParentCommitID, &commit.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	for _, f := range files {
		var fileID string
		err := tx.QueryRow(ctx, `
			INSERT INTO files (repo_id, path)
			VALUES ($1, $2)
			ON CONFLICT (repo_id, path) DO UPDATE SET is_deleted = false, updated_at = now()
			RETURNING id
		`, repoID, f.FileID).Scan(&fileID)
		if err != nil {
			return nil, err
		}

		var versionID string
		err = tx.QueryRow(ctx, `
			INSERT INTO file_versions (file_id, content, hash, size_bytes)
			VALUES ($1, $2, $3, $4)
			RETURNING id
		`, fileID, f.Content, hashContent(f.Content), len(f.Content)).Scan(&versionID)
		if err != nil {
			return nil, err
		}

		_, err = tx.Exec(ctx, `
			INSERT INTO commit_files (commit_id, file_version_id, file_id, operation)
			VALUES ($1, $2, $3, $4)
		`, commit.ID, versionID, fileID, f.Operation)
		if err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &commit, nil
}

func (q *CommitQueries) Get(ctx context.Context, id string) (*Commit, error) {
	var commit Commit
	err := q.db.Pool().QueryRow(ctx, `
		SELECT id, repo_id, author_id, author_type, message, parent_commit_id, created_at
		FROM commits WHERE id = $1
	`, id).Scan(
		&commit.ID, &commit.RepoID, &commit.AuthorID, &commit.AuthorType, &commit.Message,
		&commit.ParentCommitID, &commit.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &commit, nil
}

func (q *CommitQueries) ListByRepo(ctx context.Context, repoID string, limit int) ([]Commit, error) {
	rows, err := q.db.Pool().Query(ctx, `
		SELECT id, repo_id, author_id, author_type, message, parent_commit_id, created_at
		FROM commits WHERE repo_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`, repoID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var commits []Commit
	for rows.Next() {
		var c Commit
		if err := rows.Scan(&c.ID, &c.RepoID, &c.AuthorID, &c.AuthorType, &c.Message, &c.ParentCommitID, &c.CreatedAt); err != nil {
			return nil, err
		}
		commits = append(commits, c)
	}
	return commits, rows.Err()
}

func (q *CommitQueries) GetCommitFiles(ctx context.Context, commitID string) ([]CommitFile, error) {
	rows, err := q.db.Pool().Query(ctx, `
		SELECT cf.commit_id, cf.file_version_id, cf.file_id, cf.operation
		FROM commit_files cf WHERE cf.commit_id = $1
	`, commitID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []CommitFile
	for rows.Next() {
		var f CommitFile
		if err := rows.Scan(&f.CommitID, &f.FileVersionID, &f.FileID, &f.Operation); err != nil {
			return nil, err
		}
		files = append(files, f)
	}
	return files, rows.Err()
}

func (q *CommitQueries) GetLatestCommit(ctx context.Context, repoID string) (*Commit, error) {
	var commit Commit
	err := q.db.Pool().QueryRow(ctx, `
		SELECT id, repo_id, author_id, author_type, message, parent_commit_id, created_at
		FROM commits WHERE repo_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`, repoID).Scan(
		&commit.ID, &commit.RepoID, &commit.AuthorID, &commit.AuthorType, &commit.Message,
		&commit.ParentCommitID, &commit.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &commit, nil
}

type AuditLogQueries struct {
	db *DB
}

func NewAuditLogQueries(db *DB) *AuditLogQueries {
	return &AuditLogQueries{db: db}
}

func (q *AuditLogQueries) Create(ctx context.Context, entityType, entityID, action, actorID, actorType string, metadata map[string]interface{}) (*AuditLog, error) {
	var log AuditLog
	err := q.db.Pool().QueryRow(ctx, `
		INSERT INTO audit_log (entity_type, entity_id, action, actor_id, actor_type, metadata)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, entity_type, entity_id, action, actor_id, actor_type, metadata, created_at
	`, entityType, entityID, action, actorID, actorType, metadata).Scan(
		&log.ID, &log.EntityType, &log.EntityID, &log.Action, &log.ActorID, &log.ActorType, &log.Metadata, &log.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &log, nil
}

func (q *AuditLogQueries) ListByEntity(ctx context.Context, entityType, entityID string, limit int) ([]AuditLog, error) {
	rows, err := q.db.Pool().Query(ctx, `
		SELECT id, entity_type, entity_id, action, actor_id, actor_type, metadata, created_at
		FROM audit_log
		WHERE entity_type = $1 AND entity_id = $2
		ORDER BY created_at DESC
		LIMIT $3
	`, entityType, entityID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []AuditLog
	for rows.Next() {
		var l AuditLog
		if err := rows.Scan(&l.ID, &l.EntityType, &l.EntityID, &l.Action, &l.ActorID, &l.ActorType, &l.Metadata, &l.CreatedAt); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	return logs, rows.Err()
}

func hashContent(content []byte) string {
	hash := sha256.Sum256(content)
	return hex.EncodeToString(hash[:])
}
