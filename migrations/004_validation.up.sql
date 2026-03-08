-- Phase 4: Validation Pipeline Schema

-- Validators (registered validation checks)
CREATE TABLE validators (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    repo_id UUID REFERENCES repositories(id) ON DELETE CASCADE, -- NULL = global validator
    name VARCHAR(255) NOT NULL,
    command TEXT NOT NULL, -- shell command to run
    file_patterns VARCHAR(4096)[], -- glob patterns for files to validate
    timeout_seconds INTEGER NOT NULL DEFAULT 60,
    is_blocking BOOLEAN NOT NULL DEFAULT TRUE, -- fail blocks merge
    is_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    priority INTEGER NOT NULL DEFAULT 0, -- lower runs first
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(repo_id, name)
);

CREATE INDEX idx_validators_repo ON validators(repo_id);
CREATE INDEX idx_validators_enabled ON validators(is_enabled);

-- Validation runs
CREATE TABLE validation_runs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    commit_id UUID REFERENCES commits(id) ON DELETE CASCADE,
    workspace_id UUID REFERENCES workspaces(id) ON DELETE CASCADE,
    validator_id UUID NOT NULL REFERENCES validators(id) ON DELETE CASCADE,
    status VARCHAR(50) NOT NULL, -- 'pending', 'running', 'passed', 'failed', 'timeout', 'skipped'
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    duration_ms INTEGER,
    output TEXT,
    error_message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_validation_runs_commit ON validation_runs(commit_id);
CREATE INDEX idx_validation_runs_workspace ON validation_runs(workspace_id);
CREATE INDEX idx_validation_runs_status ON validation_runs(status);

-- File-level validation results
CREATE TABLE validation_file_results (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    run_id UUID NOT NULL REFERENCES validation_runs(id) ON DELETE CASCADE,
    file_id UUID NOT NULL REFERENCES files(id) ON DELETE CASCADE,
    status VARCHAR(50) NOT NULL, -- 'passed', 'failed', 'skipped'
    line_start INTEGER,
    line_end INTEGER,
    column_start INTEGER,
    column_end INTEGER,
    message TEXT,
    severity VARCHAR(50), -- 'error', 'warning', 'info'
    rule_id VARCHAR(255),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_validation_file_results_run ON validation_file_results(run_id);
CREATE INDEX idx_validation_file_results_file ON validation_file_results(file_id);

-- Validation summaries (denormalized for quick lookup)
CREATE TABLE validation_summaries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    commit_id UUID REFERENCES commits(id) ON DELETE CASCADE,
    workspace_id UUID REFERENCES workspaces(id) ON DELETE CASCADE,
    total_validators INTEGER NOT NULL DEFAULT 0,
    passed_count INTEGER NOT NULL DEFAULT 0,
    failed_count INTEGER NOT NULL DEFAULT 0,
    pending_count INTEGER NOT NULL DEFAULT 0,
    skipped_count INTEGER NOT NULL DEFAULT 0,
    is_complete BOOLEAN NOT NULL DEFAULT FALSE,
    overall_status VARCHAR(50) NOT NULL DEFAULT 'pending', -- 'pending', 'passed', 'failed'
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT summary_target CHECK (commit_id IS NOT NULL OR workspace_id IS NOT NULL)
);

CREATE INDEX idx_validation_summaries_commit ON validation_summaries(commit_id);
CREATE INDEX idx_validation_summaries_workspace ON validation_summaries(workspace_id);

-- Validator trigger
CREATE TRIGGER update_validators_updated_at
    BEFORE UPDATE ON validators
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER update_validation_summaries_updated_at
    BEFORE UPDATE ON validation_summaries
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- Function to update validation summaries
CREATE OR REPLACE FUNCTION update_validation_summary()
RETURNS TRIGGER AS $$
DECLARE
    v_commit_id UUID;
    v_workspace_id UUID;
    v_total INTEGER;
    v_passed INTEGER;
    v_failed INTEGER;
    v_pending INTEGER;
    v_skipped INTEGER;
BEGIN
    v_commit_id := NEW.commit_id;
    v_workspace_id := NEW.workspace_id;

    SELECT COUNT(*), 
           COUNT(*) FILTER (WHERE status = 'passed'),
           COUNT(*) FILTER (WHERE status = 'failed'),
           COUNT(*) FILTER (WHERE status = 'pending'),
           COUNT(*) FILTER (WHERE status = 'skipped')
    INTO v_total, v_passed, v_failed, v_pending, v_skipped
    FROM validation_runs
    WHERE (commit_id = v_commit_id OR workspace_id = v_workspace_id);

    INSERT INTO validation_summaries (commit_id, workspace_id, total_validators, passed_count, failed_count, pending_count, skipped_count, is_complete, overall_status)
    VALUES (v_commit_id, v_workspace_id, v_total, v_passed, v_failed, v_pending, v_skipped, v_pending = 0, CASE WHEN v_failed > 0 THEN 'failed' WHEN v_pending = 0 THEN 'passed' ELSE 'pending' END)
    ON CONFLICT (commit_id, workspace_id) DO UPDATE SET
        total_validators = v_total,
        passed_count = v_passed,
        failed_count = v_failed,
        pending_count = v_pending,
        skipped_count = v_skipped,
        is_complete = (v_pending = 0),
        overall_status = CASE WHEN v_failed > 0 THEN 'failed' WHEN v_pending = 0 THEN 'passed' ELSE 'pending' END,
        updated_at = NOW();

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_summary_on_validation_change
    AFTER INSERT OR UPDATE ON validation_runs
    FOR EACH ROW EXECUTE FUNCTION update_validation_summary();
