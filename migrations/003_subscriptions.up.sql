-- Phase 3: Subscriptions schema

-- Subscription channels (event types)
CREATE TYPE event_type AS ENUM (
    'file_create',
    'file_update',
    'file_delete',
    'commit',
    'workspace_create',
    'workspace_merge',
    'workspace_abandon',
    'lease_acquire',
    'lease_release',
    'lock_acquire',
    'lock_release',
    'validation_pass',
    'validation_fail'
);

-- Subscriptions
CREATE TABLE subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    subscriber_id VARCHAR(255) NOT NULL,
    repo_id UUID REFERENCES repositories(id) ON DELETE CASCADE,
    workspace_id UUID REFERENCES workspaces(id) ON DELETE CASCADE,
    event_types event_type[] NOT NULL,
    path_patterns VARCHAR(4096)[], -- only match certain paths
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_subscriptions_subscriber ON subscriptions(subscriber_id);
CREATE INDEX idx_subscriptions_repo ON subscriptions(repo_id);
CREATE INDEX idx_subscriptions_workspace ON subscriptions(workspace_id);

-- Event log (for replay/debugging)
CREATE TABLE event_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_type event_type NOT NULL,
    repo_id UUID REFERENCES repositories(id) ON DELETE CASCADE,
    workspace_id UUID REFERENCES workspaces(id) ON DELETE SET NULL,
    entity_type VARCHAR(50),
    entity_id UUID,
    payload JSONB NOT NULL,
    actor_id VARCHAR(255) NOT NULL,
    actor_type VARCHAR(50) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_event_log_type ON event_log(event_type);
CREATE INDEX idx_event_log_repo ON event_log(repo_id);
CREATE INDEX idx_event_log_created ON event_log(created_at);

-- Function to notify on events
CREATE OR REPLACE FUNCTION notify_event()
RETURNS TRIGGER AS $$
DECLARE
    event_type_str TEXT;
    payload JSONB;
BEGIN
    IF TG_TABLE_NAME = 'commits' AND TG_OP = 'INSERT' THEN
        event_type_str := 'commit';
        payload := jsonb_build_object(
            'commit_id', NEW.id,
            'repo_id', NEW.repo_id,
            'author_id', NEW.author_id,
            'message', NEW.message
        );
    ELSIF TG_TABLE_NAME = 'workspaces' AND TG_OP = 'INSERT' THEN
        event_type_str := 'workspace_create';
        payload := jsonb_build_object(
            'workspace_id', NEW.id,
            'repo_id', NEW.repo_id,
            'owner_id', NEW.owner_id,
            'name', NEW.name
        );
    ELSIF TG_TABLE_NAME = 'workspaces' AND TG_OP = 'UPDATE' AND NEW.status != OLD.status THEN
        IF NEW.status = 'merged' THEN
            event_type_str := 'workspace_merge';
        ELSIF NEW.status = 'abandoned' THEN
            event_type_str := 'workspace_abandon';
        END IF;
        payload := jsonb_build_object(
            'workspace_id', NEW.id,
            'repo_id', NEW.repo_id,
            'old_status', OLD.status,
            'new_status', NEW.status
        );
    END IF;

    IF event_type_str IS NOT NULL THEN
        PERFORM pg_notify('codedb_events', jsonb_build_object(
            'event_type', event_type_str,
            'payload', payload
        )::text);
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Event triggers
CREATE TRIGGER notify_on_commit
    AFTER INSERT ON commits
    FOR EACH ROW EXECUTE FUNCTION notify_event();

CREATE TRIGGER notify_on_workspace_change
    AFTER INSERT OR UPDATE ON workspaces
    FOR EACH ROW EXECUTE FUNCTION notify_event();
