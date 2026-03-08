package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

func (a *API) listWorkspaces(w http.ResponseWriter, r *http.Request) {
	repoID := chi.URLParam(r, "repoID")

	var status *string
	if s := r.URL.Query().Get("status"); s != "" {
		status = &s
	}

	workspaces, err := a.workspaceSvc.ListWorkspaces(r.Context(), repoID, status)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list workspaces", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, workspaces)
}

type createWorkspaceRequest struct {
	Name         string  `json:"name"`
	OwnerID      string  `json:"owner_id"`
	OwnerType    string  `json:"owner_type"`
	BaseCommitID *string `json:"base_commit_id"`
}

func (a *API) createWorkspace(w http.ResponseWriter, r *http.Request) {
	repoID := chi.URLParam(r, "repoID")

	var req createWorkspaceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body", err.Error())
		return
	}

	if req.Name == "" || req.OwnerID == "" {
		writeError(w, http.StatusBadRequest, "name and owner_id are required", "")
		return
	}

	if req.OwnerType == "" {
		req.OwnerType = "human"
	}

	ws, err := a.workspaceSvc.CreateWorkspace(r.Context(), repoID, req.Name, req.OwnerID, req.OwnerType, req.BaseCommitID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create workspace", err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, ws)
}

func (a *API) getWorkspace(w http.ResponseWriter, r *http.Request) {
	workspaceID := chi.URLParam(r, "workspaceID")

	ws, err := a.workspaceSvc.GetWorkspace(r.Context(), workspaceID)
	if err != nil {
		writeError(w, http.StatusNotFound, "workspace not found", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, ws)
}

type mergeWorkspaceRequest struct {
	MergedBy string `json:"merged_by"`
	Strategy string `json:"strategy"`
}

func (a *API) mergeWorkspace(w http.ResponseWriter, r *http.Request) {
	workspaceID := chi.URLParam(r, "workspaceID")

	var req mergeWorkspaceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body", err.Error())
		return
	}

	if req.Strategy == "" {
		req.Strategy = "three_way"
	}

	commit, conflicts, err := a.workspaceSvc.MergeWorkspace(r.Context(), workspaceID, req.MergedBy, req.Strategy)
	if err != nil {
		if len(conflicts) > 0 {
			writeJSON(w, http.StatusConflict, map[string]interface{}{
				"error":     err.Error(),
				"conflicts": conflicts,
			})
			return
		}
		writeError(w, http.StatusInternalServerError, "failed to merge workspace", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"commit":    commit,
		"conflicts": conflicts,
	})
}

func (a *API) abandonWorkspace(w http.ResponseWriter, r *http.Request) {
	workspaceID := chi.URLParam(r, "workspaceID")

	ws, err := a.workspaceSvc.AbandonWorkspace(r.Context(), workspaceID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to abandon workspace", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, ws)
}

func (a *API) getWorkspaceFiles(w http.ResponseWriter, r *http.Request) {
	workspaceID := chi.URLParam(r, "workspaceID")

	files, err := a.workspaceSvc.ListFiles(r.Context(), workspaceID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list workspace files", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, files)
}

type updateWorkspaceFileRequest struct {
	FileID    string `json:"file_id"`
	Content   []byte `json:"content"`
	IsDeleted bool   `json:"is_deleted"`
}

func (a *API) updateWorkspaceFile(w http.ResponseWriter, r *http.Request) {
	workspaceID := chi.URLParam(r, "workspaceID")

	var req updateWorkspaceFileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body", err.Error())
		return
	}

	if req.FileID == "" {
		writeError(w, http.StatusBadRequest, "file_id is required", "")
		return
	}

	file, err := a.workspaceSvc.UpsertFile(r.Context(), workspaceID, req.FileID, req.Content, req.IsDeleted)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update workspace file", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, file)
}

func (a *API) deleteWorkspaceFile(w http.ResponseWriter, r *http.Request) {
	workspaceID := chi.URLParam(r, "workspaceID")
	fileID := chi.URLParam(r, "fileID")

	if err := a.workspaceSvc.DeleteFile(r.Context(), workspaceID, fileID); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete workspace file", err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (a *API) listLeases(w http.ResponseWriter, r *http.Request) {
	workspaceID := chi.URLParam(r, "workspaceID")

	leases, err := a.workspaceSvc.ListLeases(r.Context(), workspaceID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list leases", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, leases)
}

type acquireLeaseRequest struct {
	FileID      *string       `json:"file_id"`
	PathPattern *string       `json:"path_pattern"`
	OwnerID     string        `json:"owner_id"`
	Intent      *string       `json:"intent"`
	TTL         time.Duration `json:"ttl"`
}

func (a *API) acquireLease(w http.ResponseWriter, r *http.Request) {
	workspaceID := chi.URLParam(r, "workspaceID")

	var req acquireLeaseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body", err.Error())
		return
	}

	if req.OwnerID == "" {
		writeError(w, http.StatusBadRequest, "owner_id is required", "")
		return
	}

	if req.TTL == 0 {
		req.TTL = 30 * time.Minute
	}

	lease, err := a.workspaceSvc.AcquireLease(r.Context(), workspaceID, req.FileID, req.PathPattern, req.OwnerID, req.Intent, req.TTL)
	if err != nil {
		writeError(w, http.StatusConflict, err.Error(), "")
		return
	}

	writeJSON(w, http.StatusCreated, lease)
}

type renewLeaseRequest struct {
	TTL time.Duration `json:"ttl"`
}

func (a *API) renewLease(w http.ResponseWriter, r *http.Request) {
	leaseID := chi.URLParam(r, "leaseID")

	var req renewLeaseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body", err.Error())
		return
	}

	if req.TTL == 0 {
		req.TTL = 30 * time.Minute
	}

	lease, err := a.workspaceSvc.RenewLease(r.Context(), leaseID, req.TTL)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to renew lease", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, lease)
}

func (a *API) releaseLease(w http.ResponseWriter, r *http.Request) {
	leaseID := chi.URLParam(r, "leaseID")

	if err := a.workspaceSvc.ReleaseLease(r.Context(), leaseID); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to release lease", err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (a *API) listLocks(w http.ResponseWriter, r *http.Request) {
	repoID := chi.URLParam(r, "repoID")

	locks, err := a.workspaceSvc.ListLocks(r.Context(), repoID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list locks", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, locks)
}

type acquireLockRequest struct {
	FileID      *string    `json:"file_id"`
	PathPattern *string    `json:"path_pattern"`
	OwnerID     string     `json:"owner_id"`
	OwnerType   string     `json:"owner_type"`
	LockType    string     `json:"lock_type"`
	Reason      *string    `json:"reason"`
	ExpiresAt   *time.Time `json:"expires_at"`
}

func (a *API) acquireLock(w http.ResponseWriter, r *http.Request) {
	repoID := chi.URLParam(r, "repoID")

	var req acquireLockRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body", err.Error())
		return
	}

	if req.OwnerID == "" {
		writeError(w, http.StatusBadRequest, "owner_id is required", "")
		return
	}

	if req.OwnerType == "" {
		req.OwnerType = "human"
	}

	if req.LockType == "" {
		req.LockType = "exclusive"
	}

	lock, err := a.workspaceSvc.AcquireLock(r.Context(), repoID, req.FileID, req.PathPattern, req.OwnerID, req.OwnerType, req.LockType, req.Reason, req.ExpiresAt)
	if err != nil {
		writeError(w, http.StatusConflict, err.Error(), "")
		return
	}

	writeJSON(w, http.StatusCreated, lock)
}

func (a *API) releaseLock(w http.ResponseWriter, r *http.Request) {
	lockID := chi.URLParam(r, "lockID")

	if err := a.workspaceSvc.ReleaseLock(r.Context(), lockID); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to release lock", err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
