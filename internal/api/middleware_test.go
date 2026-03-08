package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rokkovach/codedb/internal/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestIDMiddleware(t *testing.T) {
	t.Run("generates request ID when not provided", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		handler := RequestIDMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := GetRequestID(r.Context())
			assert.NotEmpty(t, requestID)
			w.WriteHeader(http.StatusOK)
		}))

		handler.ServeHTTP(w, req)

		assert.NotEmpty(t, w.Header().Get("X-Request-ID"))
	})

	t.Run("preserves provided request ID", func(t *testing.T) {
		providedID := "custom-request-id-12345"
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Request-ID", providedID)
		w := httptest.NewRecorder()

		handler := RequestIDMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := GetRequestID(r.Context())
			assert.Equal(t, providedID, requestID)
			w.WriteHeader(http.StatusOK)
		}))

		handler.ServeHTTP(w, req)

		assert.Equal(t, providedID, w.Header().Get("X-Request-ID"))
	})
}

func TestErrorResponses(t *testing.T) {
	_ = &API{}

	t.Run("error response includes request ID", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Request-ID", "test-id-123")
		w := httptest.NewRecorder()

		ctx := context.WithValue(req.Context(), RequestIDKey, "test-id-123")
		req = req.WithContext(ctx)

		writeError(w, req, http.StatusBadRequest, "test error", "test details")

		var resp errorResponse
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)

		assert.Equal(t, "test error", resp.Error)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
		assert.Equal(t, "test details", resp.Details)
		assert.Equal(t, "test-id-123", resp.RequestID)
	})
}

func TestOpenAPISpec(t *testing.T) {
	t.Run("generates valid OpenAPI spec", func(t *testing.T) {
		spec := GenerateOpenAPISpec()

		assert.Equal(t, "3.0.0", spec.OpenAPI)
		assert.Equal(t, "CodeDB API", spec.Info.Title)
		assert.Equal(t, "1.0.0", spec.Info.Version)
		assert.NotEmpty(t, spec.Paths)
		assert.Contains(t, spec.Paths, "/health")
		assert.Contains(t, spec.Paths, "/api/v1/repos")
	})
}

func TestOpenAPIEndpoint(t *testing.T) {
	mockDB := &db.DB{}
	api := NewAPI(mockDB)

	t.Run("returns OpenAPI spec", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/openapi.json", nil)
		w := httptest.NewRecorder()

		api.Router().ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

		var spec map[string]interface{}
		err := json.NewDecoder(w.Body).Decode(&spec)
		require.NoError(t, err)

		assert.Equal(t, "3.0.0", spec["openapi"])
	})
}

func TestDocsEndpoint(t *testing.T) {
	mockDB := &db.DB{}
	api := NewAPI(mockDB)

	t.Run("returns HTML documentation", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/v1/docs", nil)
		w := httptest.NewRecorder()

		api.Router().ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "text/html", w.Header().Get("Content-Type"))
		assert.Contains(t, w.Body.String(), "CodeDB API Documentation")
	})
}

func TestHealthEndpoint(t *testing.T) {
	mockDB := &db.DB{}
	api := NewAPI(mockDB)

	t.Run("returns healthy status", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/health", nil)
		w := httptest.NewRecorder()

		api.Router().ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var resp map[string]string
		err := json.NewDecoder(w.Body).Decode(&resp)
		require.NoError(t, err)
		assert.Equal(t, "healthy", resp["status"])
	})
}
