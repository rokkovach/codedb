package db

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	pool *pgxpool.Pool
}

func New(ctx context.Context, connString string) (*DB, error) {
	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, err
	}
	return &DB{pool: pool}, nil
}

func (db *DB) Close() {
	db.pool.Close()
}

func (db *DB) Pool() *pgxpool.Pool {
	return db.pool
}

type Repository struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
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

type Commit struct {
	ID             string    `json:"id"`
	RepoID         string    `json:"repo_id"`
	AuthorID       string    `json:"author_id"`
	AuthorType     string    `json:"author_type"`
	Message        *string   `json:"message"`
	ParentCommitID *string   `json:"parent_commit_id"`
	CreatedAt      time.Time `json:"created_at"`
}

type CommitFile struct {
	CommitID      string `json:"commit_id"`
	FileVersionID string `json:"file_version_id"`
	FileID        string `json:"file_id"`
	Operation     string `json:"operation"`
}

type AuditLog struct {
	ID         string                 `json:"id"`
	EntityType string                 `json:"entity_type"`
	EntityID   string                 `json:"entity_id"`
	Action     string                 `json:"action"`
	ActorID    string                 `json:"actor_id"`
	ActorType  string                 `json:"actor_type"`
	Metadata   map[string]interface{} `json:"metadata"`
	CreatedAt  time.Time              `json:"created_at"`
}

type Symbol struct {
	ID                 string                 `json:"id"`
	RepoID             string                 `json:"repo_id"`
	FileID             string                 `json:"file_id"`
	Name               string                 `json:"name"`
	Kind               string                 `json:"kind"`
	FullyQualifiedName string                 `json:"fully_qualified_name"`
	LineStart          int                    `json:"line_start"`
	LineEnd            *int                   `json:"line_end"`
	Signature          *string                `json:"signature"`
	Documentation      *string                `json:"documentation"`
	Metadata           map[string]interface{} `json:"metadata"`
	CreatedAt          time.Time              `json:"created_at"`
	UpdatedAt          time.Time              `json:"updated_at"`
}
