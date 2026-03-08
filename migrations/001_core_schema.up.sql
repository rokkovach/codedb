-- Phase 1: Core Schema
-- Repositories
CREATE TABLE repositories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Files with versioning
CREATE TABLE files (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    repo_id UUID NOT NULL REFERENCES repositories(id) ON DELETE CASCADE,
    path VARCHAR(4096) NOT NULL,
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(repo_id, path)
);

CREATE INDEX idx_files_repo_id ON files(repo_id);
CREATE INDEX idx_files_path ON files(path);

-- File content versions
CREATE TABLE file_versions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    file_id UUID NOT NULL REFERENCES files(id) ON DELETE CASCADE,
    content BYTEA NOT NULL,
    hash VARCHAR(64) NOT NULL,
    size_bytes BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_file_versions_file_id ON file_versions(file_id);

-- Commits (atomic multi-file changesets)
CREATE TABLE commits (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    repo_id UUID NOT NULL REFERENCES repositories(id) ON DELETE CASCADE,
    author_id VARCHAR(255) NOT NULL,
    author_type VARCHAR(50) NOT NULL, -- 'human' or 'agent'
    message TEXT,
    parent_commit_id UUID REFERENCES commits(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_commits_repo_id ON commits(repo_id);
CREATE INDEX idx_commits_parent ON commits(parent_commit_id);

-- Commit-file associations
CREATE TABLE commit_files (
    commit_id UUID NOT NULL REFERENCES commits(id) ON DELETE CASCADE,
    file_version_id UUID NOT NULL REFERENCES file_versions(id) ON DELETE CASCADE,
    file_id UUID NOT NULL REFERENCES files(id) ON DELETE CASCADE,
    operation VARCHAR(20) NOT NULL, -- 'create', 'update', 'delete'
    PRIMARY KEY(commit_id, file_id)
);

CREATE INDEX idx_commit_files_commit ON commit_files(commit_id);
CREATE INDEX idx_commit_files_file ON commit_files(file_id);

-- Audit log
CREATE TABLE audit_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    entity_type VARCHAR(50) NOT NULL,
    entity_id UUID NOT NULL,
    action VARCHAR(50) NOT NULL,
    actor_id VARCHAR(255) NOT NULL,
    actor_type VARCHAR(50) NOT NULL,
    metadata JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_log_entity ON audit_log(entity_type, entity_id);
CREATE INDEX idx_audit_log_created ON audit_log(created_at);

-- Symbols (for semantic code analysis)
CREATE TABLE symbols (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    repo_id UUID NOT NULL REFERENCES repositories(id) ON DELETE CASCADE,
    file_id UUID NOT NULL REFERENCES files(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    kind VARCHAR(50) NOT NULL, -- 'function', 'class', 'variable', etc.
    fully_qualified_name TEXT NOT NULL,
    line_start INTEGER NOT NULL,
    line_end INTEGER,
    signature TEXT,
    documentation TEXT,
    metadata JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_symbols_repo ON symbols(repo_id);
CREATE INDEX idx_symbols_file ON symbols(file_id);
CREATE INDEX idx_symbols_name ON symbols(name);
CREATE INDEX idx_symbols_fqn ON symbols(fully_qualified_name);

-- Update trigger function
CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_repositories_updated_at
    BEFORE UPDATE ON repositories
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER update_files_updated_at
    BEFORE UPDATE ON files
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER update_symbols_updated_at
    BEFORE UPDATE ON symbols
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();
