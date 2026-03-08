DROP TRIGGER IF EXISTS update_validators_updated_at ON validators;
DROP TRIGGER IF EXISTS update_validation_summaries_updated_at ON validation_summaries;
DROP TRIGGER IF EXISTS update_summary_on_validation_change ON validation_runs;
DROP FUNCTION IF EXISTS update_validation_summary();

DROP TABLE IF EXISTS validation_file_results;
DROP TABLE IF EXISTS validation_summaries;
DROP TABLE IF EXISTS validation_runs;
DROP TABLE IF EXISTS validators;
