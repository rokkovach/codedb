DROP TRIGGER IF EXISTS notify_on_commit ON commits;
DROP TRIGGER IF EXISTS notify_on_workspace_change ON workspaces;
DROP FUNCTION IF EXISTS notify_event();

DROP TABLE IF EXISTS event_log;
DROP TABLE IF EXISTS subscriptions;
DROP TYPE IF EXISTS event_type;
