DROP TRIGGER IF EXISTS update_workspaces_updated_at ON workspaces;
DROP TRIGGER IF EXISTS update_workspace_files_updated_at ON workspace_files;

DROP TABLE IF EXISTS merge_history;
DROP TABLE IF EXISTS locks;
DROP TABLE IF EXISTS leases;
DROP TABLE IF EXISTS workspace_files;
DROP TABLE IF EXISTS workspaces;
