package api

import (
	"net/http"
)

type OpenAPIInfo struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Version     string `json:"version"`
}

type OpenAPIServer struct {
	URL string `json:"url"`
}

type OpenAPIParameter struct {
	Name     string `json:"name"`
	In       string `json:"in"`
	Required bool   `json:"required"`
	Schema   struct {
		Type string `json:"type"`
	} `json:"schema"`
}

type OpenAPIResponse struct {
	Description string                 `json:"description"`
	Content     map[string]interface{} `json:"content,omitempty"`
}

type OpenAPIPath struct {
	Summary     string                     `json:"summary,omitempty"`
	Description string                     `json:"description,omitempty"`
	Parameters  []OpenAPIParameter         `json:"parameters,omitempty"`
	Responses   map[string]OpenAPIResponse `json:"responses"`
}

type OpenAPIDocument struct {
	OpenAPI    string                 `json:"openapi"`
	Info       OpenAPIInfo            `json:"info"`
	Servers    []OpenAPIServer        `json:"servers"`
	Paths      map[string]OpenAPIPath `json:"paths"`
	Components map[string]interface{} `json:"components,omitempty"`
}

func GenerateOpenAPISpec() *OpenAPIDocument {
	return &OpenAPIDocument{
		OpenAPI: "3.0.0",
		Info: OpenAPIInfo{
			Title:       "CodeDB API",
			Description: "Database-Native Collaborative Code Authoring for Humans and AI Agents",
			Version:     "1.0.0",
		},
		Servers: []OpenAPIServer{
			{URL: "http://localhost:8080"},
		},
		Paths: map[string]OpenAPIPath{
			"/health": {
				Summary: "Health check endpoint",
				Responses: map[string]OpenAPIResponse{
					"200": {
						Description: "Service is healthy",
						Content: map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"type": "object",
									"properties": map[string]interface{}{
										"status": map[string]string{"type": "string"},
									},
								},
							},
						},
					},
				},
			},
			"/api/v1/repos": {
				Summary: "Repository operations",
				Responses: map[string]OpenAPIResponse{
					"200": {
						Description: "List of repositories",
						Content: map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"type": "array",
									"items": map[string]interface{}{
										"$ref": "#/components/schemas/Repository",
									},
								},
							},
						},
					},
				},
			},
			"/api/v1/repos/{repoID}": {
				Summary: "Single repository operations",
				Parameters: []OpenAPIParameter{
					{
						Name:     "repoID",
						In:       "path",
						Required: true,
						Schema: struct {
							Type string `json:"type"`
						}{Type: "string"},
					},
				},
				Responses: map[string]OpenAPIResponse{
					"200": {
						Description: "Repository details",
						Content: map[string]interface{}{
							"application/json": map[string]interface{}{
								"schema": map[string]interface{}{
									"$ref": "#/components/schemas/Repository",
								},
							},
						},
					},
					"404": {
						Description: "Repository not found",
					},
				},
			},
			"/api/v1/repos/{repoID}/files": {
				Summary: "File operations within a repository",
				Parameters: []OpenAPIParameter{
					{
						Name:     "repoID",
						In:       "path",
						Required: true,
						Schema: struct {
							Type string `json:"type"`
						}{Type: "string"},
					},
				},
				Responses: map[string]OpenAPIResponse{
					"200": {
						Description: "List of files",
					},
				},
			},
			"/api/v1/repos/{repoID}/commits": {
				Summary: "Commit operations within a repository",
				Parameters: []OpenAPIParameter{
					{
						Name:     "repoID",
						In:       "path",
						Required: true,
						Schema: struct {
							Type string `json:"type"`
						}{Type: "string"},
					},
				},
				Responses: map[string]OpenAPIResponse{
					"200": {
						Description: "List of commits",
					},
				},
			},
			"/api/v1/repos/{repoID}/workspaces": {
				Summary: "Workspace operations",
				Parameters: []OpenAPIParameter{
					{
						Name:     "repoID",
						In:       "path",
						Required: true,
						Schema: struct {
							Type string `json:"type"`
						}{Type: "string"},
					},
				},
				Responses: map[string]OpenAPIResponse{
					"200": {
						Description: "List of workspaces",
					},
				},
			},
			"/api/v1/ws": {
				Summary: "WebSocket endpoint for real-time subscriptions",
				Responses: map[string]OpenAPIResponse{
					"101": {
						Description: "Switching to WebSocket protocol",
					},
				},
			},
		},
		Components: map[string]interface{}{
			"schemas": map[string]interface{}{
				"Repository": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"id":          map[string]string{"type": "string"},
						"name":        map[string]string{"type": "string"},
						"description": map[string]string{"type": "string"},
						"created_at":  map[string]string{"type": "string", "format": "date-time"},
						"updated_at":  map[string]string{"type": "string", "format": "date-time"},
					},
				},
				"Error": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"error":      map[string]string{"type": "string"},
						"code":       map[string]interface{}{"type": "integer"},
						"details":    map[string]string{"type": "string"},
						"request_id": map[string]string{"type": "string"},
					},
				},
			},
		},
	}
}

func (a *API) getOpenAPISpec(w http.ResponseWriter, r *http.Request) {
	spec := GenerateOpenAPISpec()
	writeJSON(w, http.StatusOK, spec)
}

func (a *API) getDocs(w http.ResponseWriter, r *http.Request) {
	html := `<!DOCTYPE html>
<html>
<head>
    <title>CodeDB API Documentation</title>
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@3.52.5/swagger-ui.css">
    <style>
        body { margin: 0; padding: 0; }
    </style>
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@3.52.5/swagger-ui-bundle.js"></script>
    <script src="https://unpkg.com/swagger-ui-dist@3.52.5/swagger-ui-standalone-preset.js"></script>
    <script>
        window.onload = function() {
            const ui = SwaggerUIBundle({
                url: "/api/v1/openapi.json",
                dom_id: '#swagger-ui',
                presets: [
                    SwaggerUIBundle.presets.apis,
                    SwaggerUIStandalonePreset
                ],
                layout: "StandaloneLayout"
            })
        }
    </script>
</body>
</html>`
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(html))
}
