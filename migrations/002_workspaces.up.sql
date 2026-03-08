-- Phase 2: Workspaces, Leases, and Locks

-- Workspaces (isolated branches for agents/humans)
CREATE TABLE workspaces (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    repo_id UUID NOT NULL REFERENCES repositories(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    owner_id VARCHAR(255) NOT NULL,
    owner_type VARCHAR(50) NOT NULL, -- 'human' or 'agent'
    base_commit_id UUID REFERENCES commits(id),
    status VARCHAR(50) NOT NULL DEFAULT 'active', -- 'active', 'merged', 'abandoned'
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(repo_id, name)
);

CREATE INDEX idx_workspaces_repo ON workspaces(repo_id);
CREATE INDEX idx_workspaces_owner ON workspaces(owner_id);
CREATE INDEX idx_workspaces_status ON workspaces(status);

-- Workspace files (workspace-local file state)
CREATE TABLE workspace_files (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    file_id UUID NOT NULL REFERENCES files(id) ON DELETE CASCADE,
    content BYTEA NOT NULL,
    hash VARCHAR(64) NOT NULL,
    is_deleted BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(workspace_id, file_id)
);

CREATE INDEX idx_workspace_files_workspace ON workspace_files(workspace_id);
CREATE INDEX idx_workspace_files_file ON workspace_files(file_id);

-- Leases (intent declarations with TTL)
CREATE TABLE leases (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    file_id UUID REFERENCES files(id) ON DELETE CASCADE,
    path_pattern VARCHAR(4096), -- glob pattern for directory-level leases
    owner_id VARCHAR(255) NOT NULL,
    intent TEXT, -- description of planned changes
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT lease_target CHECK (file_id IS NOT NULL OR path_pattern IS NOT NULL)
);

CREATE INDEX idx_leases_workspace ON leases(workspace_id);
CREATE INDEX idx_leases_file ON leases(file_id);
CREATE INDEX idx_leases_expires ON leases(expires_at);
CREATE INDEX idx_leases_pattern ON leases(path_pattern);

-- Locks (explicit file/directory locks)
CREATE TABLE locks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    repo_id UUID NOT NULL REFERENCES repositories(id) ON DELETE CASCADE,
    file_id UUID REFERENCES files(id) ON DELETE CASCADE,
    path_pattern VARCHAR(4096), -- glob pattern for directory-level locks
    owner_id VARCHAR(255) NOT NULL,
    owner_type VARCHAR(50) NOT NULL,
    lock_type VARCHAR(50) NOT NULL, -- 'exclusive', 'shared'
    reason TEXT,
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT lock_target CHECK (file_id IS NOT NULL OR path_pattern IS NOT NULL)
);

CREATE INDEX idx_locks_repo ON locks(repo_id);
CREATE INDEX idx_locks_file ON locks(file_id);
CREATE INDEX idx_locks_owner ON locks(owner_id);
CREATE INDEX idx_locks_expires ON locks(expires_at);

-- Merge history
CREATE TABLE merge_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    workspace_id UUID NOT NULL REFERENCES workspaces(id) ON DELETE CASCADE,
    merge_commit_id UUID NOT NULL REFERENCES commits(id),
    merged_by VARCHAR(255) NOT NULL,
    merge_strategy VARCHAR(50) NOT NULL, -- 'fast_forward', 'three_way', 'force'
    conflicts_resolved JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_merge_history_workspace ON merge_history(workspace_id);

-- Workspace triggers
CREATE TRIGGER update_workspaces_updated_at
    BEFORE UPDATE ON workspaces
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER update_workspace_files_updated_at
    BEFORE UPDATE ON workspace_files
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();
