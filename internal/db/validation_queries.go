package db

import (
	"context"
	"database/sql"
	"time"
)

type Validator struct {
	ID           string    `json:"id"`
	RepoID       *string   `json:"repo_id"`
	Name         string    `json:"name"`
	Command      string    `json:"command"`
	FilePatterns []string  `json:"file_patterns"`
	TimeoutSecs  int       `json:"timeout_seconds"`
	IsBlocking   bool      `json:"is_blocking"`
	IsEnabled    bool      `json:"is_enabled"`
	Priority     int       `json:"priority"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type ValidationRun struct {
	ID           string     `json:"id"`
	CommitID     *string    `json:"commit_id"`
	WorkspaceID  *string    `json:"workspace_id"`
	ValidatorID  string     `json:"validator_id"`
	Status       string     `json:"status"`
	StartedAt    *time.Time `json:"started_at"`
	CompletedAt  *time.Time `json:"completed_at"`
	DurationMs   *int       `json:"duration_ms"`
	Output       *string    `json:"output"`
	ErrorMessage *string    `json:"error_message"`
	CreatedAt    time.Time  `json:"created_at"`
}

type ValidationFileResult struct {
	ID          string    `json:"id"`
	RunID       string    `json:"run_id"`
	FileID      string    `json:"file_id"`
	Status      string    `json:"status"`
	LineStart   *int      `json:"line_start"`
	LineEnd     *int      `json:"line_end"`
	ColumnStart *int      `json:"column_start"`
	ColumnEnd   *int      `json:"column_end"`
	Message     *string   `json:"message"`
	Severity    *string   `json:"severity"`
	RuleID      *string   `json:"rule_id"`
	CreatedAt   time.Time `json:"created_at"`
}

type ValidationSummary struct {
	ID              string    `json:"id"`
	CommitID        *string   `json:"commit_id"`
	WorkspaceID     *string   `json:"workspace_id"`
	TotalValidators int       `json:"total_validators"`
	PassedCount     int       `json:"passed_count"`
	FailedCount     int       `json:"failed_count"`
	PendingCount    int       `json:"pending_count"`
	SkippedCount    int       `json:"skipped_count"`
	IsComplete      bool      `json:"is_complete"`
	OverallStatus   string    `json:"overall_status"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type ValidatorQueries struct {
	db *DB
}

func NewValidatorQueries(db *DB) *ValidatorQueries {
	return &ValidatorQueries{db: db}
}

func (q *ValidatorQueries) Create(ctx context.Context, repoID *string, name, command string, filePatterns []string, timeoutSecs int, isBlocking, isEnabled bool, priority int) (*Validator, error) {
	var v Validator
	var repoIDVal interface{}
	if repoID != nil {
		repoIDVal = *repoID
	}

	err := q.db.Pool().QueryRow(ctx, `
		INSERT INTO validators (repo_id, name, command, file_patterns, timeout_seconds, is_blocking, is_enabled, priority)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, repo_id, name, command, file_patterns, timeout_seconds, is_blocking, is_enabled, priority, created_at, updated_at
	`, repoIDVal, name, command, filePatterns, timeoutSecs, isBlocking, isEnabled, priority).Scan(
		&v.ID, &v.RepoID, &v.Name, &v.Command, &v.FilePatterns, &v.TimeoutSecs, &v.IsBlocking,
		&v.IsEnabled, &v.Priority, &v.CreatedAt, &v.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func (q *ValidatorQueries) Get(ctx context.Context, id string) (*Validator, error) {
	var v Validator
	var repoID sql.NullString
	err := q.db.Pool().QueryRow(ctx, `
		SELECT id, repo_id, name, command, file_patterns, timeout_seconds, is_blocking, is_enabled, priority, created_at, updated_at
		FROM validators WHERE id = $1
	`, id).Scan(
		&v.ID, &repoID, &v.Name, &v.Command, &v.FilePatterns, &v.TimeoutSecs, &v.IsBlocking,
		&v.IsEnabled, &v.Priority, &v.CreatedAt, &v.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if repoID.Valid {
		v.RepoID = &repoID.String
	}
	return &v, nil
}

func (q *ValidatorQueries) ListByRepo(ctx context.Context, repoID string) ([]Validator, error) {
	rows, err := q.db.Pool().Query(ctx, `
		SELECT id, repo_id, name, command, file_patterns, timeout_seconds, is_blocking, is_enabled, priority, created_at, updated_at
		FROM validators WHERE repo_id = $1 OR repo_id IS NULL
		ORDER BY priority, name
	`, repoID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var validators []Validator
	for rows.Next() {
		var v Validator
		var repoID sql.NullString
		if err := rows.Scan(&v.ID, &repoID, &v.Name, &v.Command, &v.FilePatterns, &v.TimeoutSecs,
			&v.IsBlocking, &v.IsEnabled, &v.Priority, &v.CreatedAt, &v.UpdatedAt); err != nil {
			return nil, err
		}
		if repoID.Valid {
			v.RepoID = &repoID.String
		}
		validators = append(validators, v)
	}
	return validators, rows.Err()
}

func (q *ValidatorQueries) ListEnabled(ctx context.Context, repoID string) ([]Validator, error) {
	rows, err := q.db.Pool().Query(ctx, `
		SELECT id, repo_id, name, command, file_patterns, timeout_seconds, is_blocking, is_enabled, priority, created_at, updated_at
		FROM validators WHERE is_enabled = true AND (repo_id = $1 OR repo_id IS NULL)
		ORDER BY priority, name
	`, repoID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var validators []Validator
	for rows.Next() {
		var v Validator
		var repoIDNull sql.NullString
		if err := rows.Scan(&v.ID, &repoIDNull, &v.Name, &v.Command, &v.FilePatterns, &v.TimeoutSecs,
			&v.IsBlocking, &v.IsEnabled, &v.Priority, &v.CreatedAt, &v.UpdatedAt); err != nil {
			return nil, err
		}
		if repoIDNull.Valid {
			v.RepoID = &repoIDNull.String
		}
		validators = append(validators, v)
	}
	return validators, rows.Err()
}

func (q *ValidatorQueries) Update(ctx context.Context, id string, name *string, command *string, filePatterns []string, timeoutSecs *int, isBlocking *bool, isEnabled *bool, priority *int) (*Validator, error) {
	var v Validator
	var repoID sql.NullString
	err := q.db.Pool().QueryRow(ctx, `
		UPDATE validators SET
			name = COALESCE($2, name),
			command = COALESCE($3, command),
			file_patterns = CASE WHEN $4::varchar[] IS NOT NULL THEN $4 ELSE file_patterns END,
			timeout_seconds = COALESCE($5, timeout_seconds),
			is_blocking = COALESCE($6, is_blocking),
			is_enabled = COALESCE($7, is_enabled),
			priority = COALESCE($8, priority)
		WHERE id = $1
		RETURNING id, repo_id, name, command, file_patterns, timeout_seconds, is_blocking, is_enabled, priority, created_at, updated_at
	`, id, name, command, filePatterns, timeoutSecs, isBlocking, isEnabled, priority).Scan(
		&v.ID, &repoID, &v.Name, &v.Command, &v.FilePatterns, &v.TimeoutSecs, &v.IsBlocking,
		&v.IsEnabled, &v.Priority, &v.CreatedAt, &v.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	if repoID.Valid {
		v.RepoID = &repoID.String
	}
	return &v, nil
}

func (q *ValidatorQueries) Delete(ctx context.Context, id string) error {
	_, err := q.db.Pool().Exec(ctx, `DELETE FROM validators WHERE id = $1`, id)
	return err
}

type ValidationRunQueries struct {
	db *DB
}

func NewValidationRunQueries(db *DB) *ValidationRunQueries {
	return &ValidationRunQueries{db: db}
}

func (q *ValidationRunQueries) Create(ctx context.Context, commitID, workspaceID *string, validatorID string) (*ValidationRun, error) {
	var run ValidationRun
	var commitIDVal, workspaceIDVal interface{}
	if commitID != nil {
		commitIDVal = *commitID
	}
	if workspaceID != nil {
		workspaceIDVal = *workspaceID
	}

	err := q.db.Pool().QueryRow(ctx, `
		INSERT INTO validation_runs (commit_id, workspace_id, validator_id, status)
		VALUES ($1, $2, $3, 'pending')
		RETURNING id, commit_id, workspace_id, validator_id, status, started_at, completed_at, duration_ms, output, error_message, created_at
	`, commitIDVal, workspaceIDVal, validatorID).Scan(
		&run.ID, &run.CommitID, &run.WorkspaceID, &run.ValidatorID, &run.Status,
		&run.StartedAt, &run.CompletedAt, &run.DurationMs, &run.Output, &run.ErrorMessage, &run.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &run, nil
}

func (q *ValidationRunQueries) Get(ctx context.Context, id string) (*ValidationRun, error) {
	var run ValidationRun
	err := q.db.Pool().QueryRow(ctx, `
		SELECT id, commit_id, workspace_id, validator_id, status, started_at, completed_at, duration_ms, output, error_message, created_at
		FROM validation_runs WHERE id = $1
	`, id).Scan(
		&run.ID, &run.CommitID, &run.WorkspaceID, &run.ValidatorID, &run.Status,
		&run.StartedAt, &run.CompletedAt, &run.DurationMs, &run.Output, &run.ErrorMessage, &run.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &run, nil
}

func (q *ValidationRunQueries) Start(ctx context.Context, id string) error {
	now := time.Now()
	_, err := q.db.Pool().Exec(ctx, `
		UPDATE validation_runs SET status = 'running', started_at = $2 WHERE id = $1
	`, id, now)
	return err
}

func (q *ValidationRunQueries) Complete(ctx context.Context, id, status string, output, errorMessage *string) error {
	now := time.Now()
	_, err := q.db.Pool().Exec(ctx, `
		UPDATE validation_runs SET
			status = $2,
			completed_at = $3,
			output = $4,
			error_message = $5,
			duration_ms = EXTRACT(MILLISECONDS FROM $3 - started_at)
		WHERE id = $1
	`, id, status, now, output, errorMessage)
	return err
}

func (q *ValidationRunQueries) ListByCommit(ctx context.Context, commitID string) ([]ValidationRun, error) {
	rows, err := q.db.Pool().Query(ctx, `
		SELECT id, commit_id, workspace_id, validator_id, status, started_at, completed_at, duration_ms, output, error_message, created_at
		FROM validation_runs WHERE commit_id = $1
		ORDER BY created_at DESC
	`, commitID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var runs []ValidationRun
	for rows.Next() {
		var r ValidationRun
		if err := rows.Scan(&r.ID, &r.CommitID, &r.WorkspaceID, &r.ValidatorID, &r.Status,
			&r.StartedAt, &r.CompletedAt, &r.DurationMs, &r.Output, &r.ErrorMessage, &r.CreatedAt); err != nil {
			return nil, err
		}
		runs = append(runs, r)
	}
	return runs, rows.Err()
}

func (q *ValidationRunQueries) ListByWorkspace(ctx context.Context, workspaceID string) ([]ValidationRun, error) {
	rows, err := q.db.Pool().Query(ctx, `
		SELECT id, commit_id, workspace_id, validator_id, status, started_at, completed_at, duration_ms, output, error_message, created_at
		FROM validation_runs WHERE workspace_id = $1
		ORDER BY created_at DESC
	`, workspaceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var runs []ValidationRun
	for rows.Next() {
		var r ValidationRun
		if err := rows.Scan(&r.ID, &r.CommitID, &r.WorkspaceID, &r.ValidatorID, &r.Status,
			&r.StartedAt, &r.CompletedAt, &r.DurationMs, &r.Output, &r.ErrorMessage, &r.CreatedAt); err != nil {
			return nil, err
		}
		runs = append(runs, r)
	}
	return runs, rows.Err()
}

type ValidationFileResultQueries struct {
	db *DB
}

func NewValidationFileResultQueries(db *DB) *ValidationFileResultQueries {
	return &ValidationFileResultQueries{db: db}
}

func (q *ValidationFileResultQueries) Create(ctx context.Context, runID, fileID, status string, lineStart, lineEnd, columnStart, columnEnd *int, message, severity, ruleID *string) (*ValidationFileResult, error) {
	var result ValidationFileResult
	err := q.db.Pool().QueryRow(ctx, `
		INSERT INTO validation_file_results (run_id, file_id, status, line_start, line_end, column_start, column_end, message, severity, rule_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, run_id, file_id, status, line_start, line_end, column_start, column_end, message, severity, rule_id, created_at
	`, runID, fileID, status, lineStart, lineEnd, columnStart, columnEnd, message, severity, ruleID).Scan(
		&result.ID, &result.RunID, &result.FileID, &result.Status, &result.LineStart, &result.LineEnd,
		&result.ColumnStart, &result.ColumnEnd, &result.Message, &result.Severity, &result.RuleID, &result.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (q *ValidationFileResultQueries) ListByRun(ctx context.Context, runID string) ([]ValidationFileResult, error) {
	rows, err := q.db.Pool().Query(ctx, `
		SELECT id, run_id, file_id, status, line_start, line_end, column_start, column_end, message, severity, rule_id, created_at
		FROM validation_file_results WHERE run_id = $1
		ORDER BY file_id, line_start, column_start
	`, runID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []ValidationFileResult
	for rows.Next() {
		var r ValidationFileResult
		if err := rows.Scan(&r.ID, &r.RunID, &r.FileID, &r.Status, &r.LineStart, &r.LineEnd,
			&r.ColumnStart, &r.ColumnEnd, &r.Message, &r.Severity, &r.RuleID, &r.CreatedAt); err != nil {
			return nil, err
		}
		results = append(results, r)
	}
	return results, rows.Err()
}

type ValidationSummaryQueries struct {
	db *DB
}

func NewValidationSummaryQueries(db *DB) *ValidationSummaryQueries {
	return &ValidationSummaryQueries{db: db}
}

func (q *ValidationSummaryQueries) GetByCommit(ctx context.Context, commitID string) (*ValidationSummary, error) {
	var s ValidationSummary
	err := q.db.Pool().QueryRow(ctx, `
		SELECT id, commit_id, workspace_id, total_validators, passed_count, failed_count, pending_count, skipped_count, is_complete, overall_status, created_at, updated_at
		FROM validation_summaries WHERE commit_id = $1
	`, commitID).Scan(
		&s.ID, &s.CommitID, &s.WorkspaceID, &s.TotalValidators, &s.PassedCount, &s.FailedCount,
		&s.PendingCount, &s.SkippedCount, &s.IsComplete, &s.OverallStatus, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (q *ValidationSummaryQueries) GetByWorkspace(ctx context.Context, workspaceID string) (*ValidationSummary, error) {
	var s ValidationSummary
	err := q.db.Pool().QueryRow(ctx, `
		SELECT id, commit_id, workspace_id, total_validators, passed_count, failed_count, pending_count, skipped_count, is_complete, overall_status, created_at, updated_at
		FROM validation_summaries WHERE workspace_id = $1
	`, workspaceID).Scan(
		&s.ID, &s.CommitID, &s.WorkspaceID, &s.TotalValidators, &s.PassedCount, &s.FailedCount,
		&s.PendingCount, &s.SkippedCount, &s.IsComplete, &s.OverallStatus, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &s, nil
}
