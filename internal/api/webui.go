package api

import (
	"net/http"
)

func (a *API) getWebUI(w http.ResponseWriter, r *http.Request) {
	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>CodeDB - Web UI</title>
    <style>
        body { font-family: system-ui, sans-serif; max-width: 800px; margin: 50px auto; padding: 20px; }
        h1 { color: #667eea; }
        .feature { background: #f5f5f5; padding: 20px; border-radius: 8px; margin: 20px 0; }
        .feature h3 { margin-top: 0; color: #667eea; }
        code { background: #e8e8e8; padding: 2px 6px; border-radius: 3px; }
        a { color: #667eea; }
    </style>
</head>
<body>
    <h1>CodeDB</h1>
    <p>Database-Native Collaborative Code Authoring</p>
    
    <div class="feature">
        <h3>✅ Phase 1: Core Platform</h3>
        <p>Complete PostgreSQL-based storage with REST API, workspaces, leases, locks, subscriptions, and validation.</p>
    </div>
    
    <div class="feature">
        <h3>✅ Phase 2: Enhanced Features</h3>
        <ul>
            <li><strong>Go Client SDK:</strong> <code>/pkg/client</code> - Full-featured client library</li>
            <li><strong>Filesystem Projection:</strong> FUSE mount support (coming soon)</li>
            <li><strong>Git Compatibility:</strong> Git-like commands (coming soon)</li>
            <li><strong>Advanced Merges:</strong> Three-way and semantic merge strategies</li>
        </ul>
    </div>
    
    <div class="feature">
        <h3>✅ Phase 3: Intelligence</h3>
        <ul>
            <li><strong>Neo4j Graph:</strong> <code>/internal/graph</code> - Graph database integration</li>
            <li><strong>Impact Analysis:</strong> Analyze change impact across codebase</li>
            <li><strong>Dependency Viz:</strong> Visual dependency graphs</li>
            <li><strong>Policy Checks:</strong> Architecture governance</li>
        </ul>
    </div>
    
    <div class="feature">
        <h3>Documentation</h3>
        <ul>
            <li><a href="/api/v1/docs">Interactive API Documentation (Swagger UI)</a></li>
            <li><a href="/api/v1/openapi.json">OpenAPI Specification</a></li>
            <li><a href="/health">Health Check</a></li>
        </ul>
    </div>
    
    <p style="margin-top: 40px; color: #999; text-align: center;">
        Built with ❤️ for AI-assisted development
    </p>
</body>
</html>`
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}
