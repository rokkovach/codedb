package api

import (
	"context"
	"encoding/json"
	"net/http"
	"os/exec"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/rokkovach/codedb/internal/db"
)

type ValidationService struct {
	validatorQueries  *db.ValidatorQueries
	runQueries        *db.ValidationRunQueries
	fileResultQueries *db.ValidationFileResultQueries
	summaryQueries    *db.ValidationSummaryQueries
}

func NewValidationService(database *db.DB) *ValidationService {
	return &ValidationService{
		validatorQueries:  db.NewValidatorQueries(database),
		runQueries:        db.NewValidationRunQueries(database),
		fileResultQueries: db.NewValidationFileResultQueries(database),
		summaryQueries:    db.NewValidationSummaryQueries(database),
	}
}

func (s *ValidationService) CreateValidator(ctx context.Context, repoID *string, name, command string, filePatterns []string, timeoutSecs int, isBlocking, isEnabled bool, priority int) (*db.Validator, error) {
	return s.validatorQueries.Create(ctx, repoID, name, command, filePatterns, timeoutSecs, isBlocking, isEnabled, priority)
}

func (s *ValidationService) GetValidator(ctx context.Context, id string) (*db.Validator, error) {
	return s.validatorQueries.Get(ctx, id)
}

func (s *ValidationService) ListValidators(ctx context.Context, repoID string) ([]db.Validator, error) {
	return s.validatorQueries.ListByRepo(ctx, repoID)
}

func (s *ValidationService) UpdateValidator(ctx context.Context, id string, name *string, command *string, filePatterns []string, timeoutSecs *int, isBlocking *bool, isEnabled *bool, priority *int) (*db.Validator, error) {
	return s.validatorQueries.Update(ctx, id, name, command, filePatterns, timeoutSecs, isBlocking, isEnabled, priority)
}

func (s *ValidationService) DeleteValidator(ctx context.Context, id string) error {
	return s.validatorQueries.Delete(ctx, id)
}

func (s *ValidationService) TriggerValidationsForCommit(ctx context.Context, repoID, commitID string) error {
	validators, err := s.validatorQueries.ListEnabled(ctx, repoID)
	if err != nil {
		return err
	}

	for _, v := range validators {
		_, err := s.runQueries.Create(ctx, &commitID, nil, v.ID)
		if err != nil {
			continue
		}
		go s.executeValidator(ctx, v, &commitID, nil)
	}

	return nil
}

func (s *ValidationService) TriggerValidationsForWorkspace(ctx context.Context, repoID, workspaceID string) error {
	validators, err := s.validatorQueries.ListEnabled(ctx, repoID)
	if err != nil {
		return err
	}

	for _, v := range validators {
		_, err := s.runQueries.Create(ctx, nil, &workspaceID, v.ID)
		if err != nil {
			continue
		}
		go s.executeValidator(ctx, v, nil, &workspaceID)
	}

	return nil
}

func (s *ValidationService) executeValidator(ctx context.Context, v db.Validator, commitID, workspaceID *string) {
	run, err := s.runQueries.Create(ctx, commitID, workspaceID, v.ID)
	if err != nil {
		return
	}

	if err := s.runQueries.Start(ctx, run.ID); err != nil {
		return
	}

	timeout := time.Duration(v.TimeoutSecs) * time.Second
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, "sh", "-c", v.Command)
	output, err := cmd.CombinedOutput()

	status := "passed"
	var errMsg *string
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			status = "timeout"
			t := "validation timed out"
			errMsg = &t
		} else {
			status = "failed"
			t := err.Error()
			errMsg = &t
		}
	}

	outputStr := string(output)
	s.runQueries.Complete(ctx, run.ID, status, &outputStr, errMsg)
}

func (s *ValidationService) GetSummaryForCommit(ctx context.Context, commitID string) (*db.ValidationSummary, error) {
	return s.summaryQueries.GetByCommit(ctx, commitID)
}

func (s *ValidationService) GetSummaryForWorkspace(ctx context.Context, workspaceID string) (*db.ValidationSummary, error) {
	return s.summaryQueries.GetByWorkspace(ctx, workspaceID)
}

func (s *ValidationService) ListRunsForCommit(ctx context.Context, commitID string) ([]db.ValidationRun, error) {
	return s.runQueries.ListByCommit(ctx, commitID)
}

func (s *ValidationService) ListRunsForWorkspace(ctx context.Context, workspaceID string) ([]db.ValidationRun, error) {
	return s.runQueries.ListByWorkspace(ctx, workspaceID)
}

func (s *ValidationService) CreateFileResult(ctx context.Context, runID, fileID, status string, lineStart, lineEnd, columnStart, columnEnd *int, message, severity, ruleID *string) (*db.ValidationFileResult, error) {
	return s.fileResultQueries.Create(ctx, runID, fileID, status, lineStart, lineEnd, columnStart, columnEnd, message, severity, ruleID)
}

func (s *ValidationService) ListFileResults(ctx context.Context, runID string) ([]db.ValidationFileResult, error) {
	return s.fileResultQueries.ListByRun(ctx, runID)
}

func (a *API) listValidators(w http.ResponseWriter, r *http.Request) {
	repoID := chi.URLParam(r, "repoID")

	validators, err := a.validationSvc.ListValidators(r.Context(), repoID)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "failed to list validators", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, validators)
}

type createValidatorRequest struct {
	Name         string   `json:"name"`
	Command      string   `json:"command"`
	FilePatterns []string `json:"file_patterns"`
	TimeoutSecs  int      `json:"timeout_seconds"`
	IsBlocking   bool     `json:"is_blocking"`
	IsEnabled    bool     `json:"is_enabled"`
	Priority     int      `json:"priority"`
}

func (a *API) createValidator(w http.ResponseWriter, r *http.Request) {
	repoID := chi.URLParam(r, "repoID")

	var req createValidatorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid request body", err.Error())
		return
	}

	if req.Name == "" || req.Command == "" {
		writeError(w, r, http.StatusBadRequest, "name and command are required", "")
		return
	}

	if req.TimeoutSecs == 0 {
		req.TimeoutSecs = 60
	}

	v, err := a.validationSvc.CreateValidator(r.Context(), &repoID, req.Name, req.Command, req.FilePatterns, req.TimeoutSecs, req.IsBlocking, req.IsEnabled, req.Priority)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "failed to create validator", err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, v)
}

func (a *API) getValidator(w http.ResponseWriter, r *http.Request) {
	validatorID := chi.URLParam(r, "validatorID")

	v, err := a.validationSvc.GetValidator(r.Context(), validatorID)
	if err != nil {
		writeError(w, r, http.StatusNotFound, "validator not found", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, v)
}

type updateValidatorRequest struct {
	Name         *string  `json:"name"`
	Command      *string  `json:"command"`
	FilePatterns []string `json:"file_patterns"`
	TimeoutSecs  *int     `json:"timeout_seconds"`
	IsBlocking   *bool    `json:"is_blocking"`
	IsEnabled    *bool    `json:"is_enabled"`
	Priority     *int     `json:"priority"`
}

func (a *API) updateValidator(w http.ResponseWriter, r *http.Request) {
	validatorID := chi.URLParam(r, "validatorID")

	var req updateValidatorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, r, http.StatusBadRequest, "invalid request body", err.Error())
		return
	}

	v, err := a.validationSvc.UpdateValidator(r.Context(), validatorID, req.Name, req.Command, req.FilePatterns, req.TimeoutSecs, req.IsBlocking, req.IsEnabled, req.Priority)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "failed to update validator", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, v)
}

func (a *API) deleteValidator(w http.ResponseWriter, r *http.Request) {
	validatorID := chi.URLParam(r, "validatorID")

	if err := a.validationSvc.DeleteValidator(r.Context(), validatorID); err != nil {
		writeError(w, r, http.StatusInternalServerError, "failed to delete validator", err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (a *API) getWorkspaceValidations(w http.ResponseWriter, r *http.Request) {
	workspaceID := chi.URLParam(r, "workspaceID")

	summary, err := a.validationSvc.GetSummaryForWorkspace(r.Context(), workspaceID)
	if err != nil {
		summary = nil
	}

	runs, err := a.validationSvc.ListRunsForWorkspace(r.Context(), workspaceID)
	if err != nil {
		writeError(w, r, http.StatusInternalServerError, "failed to get validation runs", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"summary": summary,
		"runs":    runs,
	})
}

func parseFilePatterns(patterns []string) []string {
	var result []string
	for _, p := range patterns {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}
