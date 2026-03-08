package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rokkovach/codedb/internal/db"
)

func (a *API) listRepos(w http.ResponseWriter, r *http.Request) {
	limit := 50
	offset := 0

	repos, err := a.repoQueries.List(r.Context(), limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list repositories", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, repos)
}

type createRepoRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

func (a *API) createRepo(w http.ResponseWriter, r *http.Request) {
	var req createRepoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body", err.Error())
		return
	}

	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required", "")
		return
	}

	repo, err := a.repoQueries.Create(r.Context(), req.Name, req.Description)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create repository", err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, repo)
}

func (a *API) getRepo(w http.ResponseWriter, r *http.Request) {
	repoID := chi.URLParam(r, "repoID")

	repo, err := a.repoQueries.Get(r.Context(), repoID)
	if err != nil {
		writeError(w, http.StatusNotFound, "repository not found", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, repo)
}

type updateRepoRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

func (a *API) updateRepo(w http.ResponseWriter, r *http.Request) {
	repoID := chi.URLParam(r, "repoID")

	var req updateRepoRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body", err.Error())
		return
	}

	repo, err := a.repoQueries.Update(r.Context(), repoID, req.Name, req.Description)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to update repository", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, repo)
}

func (a *API) deleteRepo(w http.ResponseWriter, r *http.Request) {
	repoID := chi.URLParam(r, "repoID")

	if err := a.repoQueries.Delete(r.Context(), repoID); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete repository", err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (a *API) listFiles(w http.ResponseWriter, r *http.Request) {
	repoID := chi.URLParam(r, "repoID")
	includeDeleted := r.URL.Query().Get("include_deleted") == "true"

	files, err := a.fileQueries.ListByRepo(r.Context(), repoID, includeDeleted)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list files", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, files)
}

type createFileRequest struct {
	Path    string `json:"path"`
	Content []byte `json:"content"`
}

func (a *API) createFile(w http.ResponseWriter, r *http.Request) {
	repoID := chi.URLParam(r, "repoID")

	var req createFileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body", err.Error())
		return
	}

	if req.Path == "" {
		writeError(w, http.StatusBadRequest, "path is required", "")
		return
	}

	file, err := a.fileQueries.Create(r.Context(), repoID, req.Path)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create file", err.Error())
		return
	}

	if len(req.Content) > 0 {
		_, err = a.fileQueries.CreateVersion(r.Context(), file.ID, req.Content)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "failed to create file version", err.Error())
			return
		}
	}

	writeJSON(w, http.StatusCreated, file)
}

func (a *API) getFile(w http.ResponseWriter, r *http.Request) {
	fileID := chi.URLParam(r, "fileID")

	file, err := a.fileQueries.Get(r.Context(), fileID)
	if err != nil {
		writeError(w, http.StatusNotFound, "file not found", err.Error())
		return
	}

	version := r.URL.Query().Get("version")
	var fv *db.FileVersion
	if version != "" {
		fv, err = a.fileQueries.GetVersion(r.Context(), version)
	} else {
		fv, err = a.fileQueries.GetLatestVersion(r.Context(), fileID)
	}
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"file":    file,
			"content": nil,
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"file":    file,
		"content": fv.Content,
		"version": fv,
	})
}

func (a *API) listFileVersions(w http.ResponseWriter, r *http.Request) {
	fileID := chi.URLParam(r, "fileID")
	limit := 50

	versions, err := a.fileQueries.ListVersions(r.Context(), fileID, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list file versions", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, versions)
}

func (a *API) listCommits(w http.ResponseWriter, r *http.Request) {
	repoID := chi.URLParam(r, "repoID")
	limit := 50

	commits, err := a.commitQueries.ListByRepo(r.Context(), repoID, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list commits", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, commits)
}

type createCommitRequest struct {
	AuthorID   string               `json:"author_id"`
	AuthorType string               `json:"author_type"`
	Message    *string              `json:"message"`
	Files      []db.CommitFileInput `json:"files"`
}

func (a *API) createCommit(w http.ResponseWriter, r *http.Request) {
	repoID := chi.URLParam(r, "repoID")

	var req createCommitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body", err.Error())
		return
	}

	if req.AuthorID == "" {
		writeError(w, http.StatusBadRequest, "author_id is required", "")
		return
	}

	latest, _ := a.commitQueries.GetLatestCommit(r.Context(), repoID)
	var parentID *string
	if latest != nil {
		parentID = &latest.ID
	}

	commit, err := a.commitQueries.Create(r.Context(), repoID, req.AuthorID, req.AuthorType, req.Message, parentID, req.Files)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create commit", err.Error())
		return
	}

	go a.validationSvc.TriggerValidationsForCommit(r.Context(), repoID, commit.ID)

	writeJSON(w, http.StatusCreated, commit)
}

func (a *API) getCommit(w http.ResponseWriter, r *http.Request) {
	commitID := chi.URLParam(r, "commitID")

	commit, err := a.commitQueries.Get(r.Context(), commitID)
	if err != nil {
		writeError(w, http.StatusNotFound, "commit not found", err.Error())
		return
	}

	files, err := a.commitQueries.GetCommitFiles(r.Context(), commitID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get commit files", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"commit": commit,
		"files":  files,
	})
}

func (a *API) getCommitValidations(w http.ResponseWriter, r *http.Request) {
	commitID := chi.URLParam(r, "commitID")

	summary, err := a.validationSvc.GetSummaryForCommit(r.Context(), commitID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get validation summary", err.Error())
		return
	}

	runs, err := a.validationSvc.ListRunsForCommit(r.Context(), commitID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to get validation runs", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"summary": summary,
		"runs":    runs,
	})
}
