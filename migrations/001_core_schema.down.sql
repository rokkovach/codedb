DROP TRIGGER IF EXISTS update_repositories_updated_at ON repositories;
DROP TRIGGER IF EXISTS update_files_updated_at ON files;
DROP TRIGGER IF EXISTS update_symbols_updated_at ON symbols;
DROP FUNCTION IF EXISTS update_updated_at();

DROP TABLE IF EXISTS symbols;
DROP TABLE IF EXISTS audit_log;
DROP TABLE IF EXISTS commit_files;
DROP TABLE IF EXISTS commits;
DROP TABLE IF EXISTS file_versions;
DROP TABLE IF EXISTS files;
DROP TABLE IF EXISTS repositories;
