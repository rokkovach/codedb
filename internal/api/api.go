package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rokkovach/codedb/internal/db"
)

type API struct {
	db              *db.DB
	repoQueries     *db.RepositoryQueries
	fileQueries     *db.FileQueries
	commitQueries   *db.CommitQueries
	auditQueries    *db.AuditLogQueries
	workspaceSvc    *WorkspaceService
	subscriptionSvc *SubscriptionService
	validationSvc   *ValidationService
}

func NewAPI(database *db.DB) *API {
	return &API{
		db:              database,
		repoQueries:     db.NewRepositoryQueries(database),
		fileQueries:     db.NewFileQueries(database),
		commitQueries:   db.NewCommitQueries(database),
		auditQueries:    db.NewAuditLogQueries(database),
		workspaceSvc:    NewWorkspaceService(database),
		subscriptionSvc: NewSubscriptionService(database),
		validationSvc:   NewValidationService(database),
	}
}

func (a *API) Router() http.Handler {
	r := chi.NewRouter()
	r.Use(RequestIDMiddleware)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RealIP)

	r.Route("/api/v1", func(r chi.Router) {
		r.Get("/openapi.json", a.getOpenAPISpec)
		r.Get("/docs", a.getDocs)

		r.Route("/repos", func(r chi.Router) {
			r.Get("/", a.listRepos)
			r.Post("/", a.createRepo)
			r.Route("/{repoID}", func(r chi.Router) {
				r.Get("/", a.getRepo)
				r.Put("/", a.updateRepo)
				r.Delete("/", a.deleteRepo)
				r.Get("/files", a.listFiles)
				r.Post("/files", a.createFile)
				r.Get("/commits", a.listCommits)
				r.Post("/commits", a.createCommit)
				r.Get("/locks", a.listLocks)
				r.Post("/locks", a.acquireLock)
				r.Delete("/locks/{lockID}", a.releaseLock)
				r.Route("/validators", func(r chi.Router) {
					r.Get("/", a.listValidators)
					r.Post("/", a.createValidator)
					r.Route("/{validatorID}", func(r chi.Router) {
						r.Get("/", a.getValidator)
						r.Put("/", a.updateValidator)
						r.Delete("/", a.deleteValidator)
					})
				})
				r.Route("/workspaces", func(r chi.Router) {
					r.Get("/", a.listWorkspaces)
					r.Post("/", a.createWorkspace)
					r.Route("/{workspaceID}", func(r chi.Router) {
						r.Get("/", a.getWorkspace)
						r.Post("/merge", a.mergeWorkspace)
						r.Post("/abandon", a.abandonWorkspace)
						r.Get("/files", a.getWorkspaceFiles)
						r.Post("/files", a.updateWorkspaceFile)
						r.Delete("/files/{fileID}", a.deleteWorkspaceFile)
						r.Get("/leases", a.listLeases)
						r.Post("/leases", a.acquireLease)
						r.Put("/leases/{leaseID}", a.renewLease)
						r.Delete("/leases/{leaseID}", a.releaseLease)
						r.Get("/validations", a.getWorkspaceValidations)
					})
				})
				r.Get("/subscriptions", a.listSubscriptions)
				r.Post("/subscriptions", a.createSubscription)
				r.Delete("/subscriptions/{subID}", a.deleteSubscription)
			})
		})
		r.Get("/files/{fileID}", a.getFile)
		r.Get("/files/{fileID}/versions", a.listFileVersions)
		r.Get("/commits/{commitID}", a.getCommit)
		r.Get("/commits/{commitID}/validations", a.getCommitValidations)
		r.Get("/ws", a.handleWebSocket)
	})

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "healthy"})
	})

	return r
}

type errorResponse struct {
	Error     string `json:"error"`
	Code      int    `json:"code"`
	Details   string `json:"details,omitempty"`
	RequestID string `json:"request_id,omitempty"`
}

func writeError(w http.ResponseWriter, r *http.Request, code int, message string, details string) {
	requestID := GetRequestID(r.Context())
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(errorResponse{
		Error:     message,
		Code:      code,
		Details:   details,
		RequestID: requestID,
	})
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func (a *API) SubscriptionSvc() *SubscriptionService {
	return a.subscriptionSvc
}

type contextKey string

const (
	repoIDKey contextKey = "repoID"
)
