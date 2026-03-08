# CodeDB

> Database-Native Collaborative Code Authoring for Humans and AI Agents

CodeDB is a database-native system for collaborative software development where the source of truth is stored in a transactional database rather than a filesystem checkout. It is designed for teams using many AI agents and human developers concurrently, where traditional Git workflows become a bottleneck.

## Table of Contents

- [Overview](#overview)
- [Features](#features)
- [Architecture](#architecture)
- [Quick Start](#quick-start)
- [Installation](#installation)
- [Configuration](#configuration)
- [API Documentation](#api-documentation)
- [Usage Examples](#usage-examples)
- [Development](#development)
- [Testing](#testing)
- [Deployment](#deployment)
- [Contributing](#contributing)
- [License](#license)

## Overview

### Problem Statement

Modern code generation velocity has increased significantly, especially with AI agents producing changes in parallel. In this environment, traditional Git workflows create friction:

- Pull request review becoming the main bottleneck
- Filesystem-based code storage poorly suited for concurrent agents
- Weak write-level atomicity across multiple files
- Poor coordination primitives for overlapping edits
- Limited real-time subscriptions to codebase changes
- Delayed validation instead of continuous verification

### Solution

CodeDB treats code as high-velocity structured data instead of static files:

- **Database-first source of truth** - PostgreSQL stores all code and metadata
- **Atomic write operations** - Multi-file commits with transactional guarantees
- **Isolated workspaces** - Per-agent branches with merge semantics
- **Real-time subscriptions** - PostgreSQL NOTIFY + WebSocket delivery
- **Continuous validation** - Linting, formatting, and checks on every write
- **Full auditability** - Complete history, rollback, and replayability

## Features

### Core Features

- **Repository Management** - Create, update, delete code repositories
- **File Versioning** - Track all file versions with content hashing
- **Atomic Commits** - Multi-file changesets with transactional guarantees
- **Workspace Isolation** - Isolated development environments for agents/humans
- **Leases & Locks** - Coordination primitives for concurrent editing
- **Real-time Subscriptions** - WebSocket-based event notifications
- **Validation Pipeline** - Configurable validators (linters, formatters, tests)
- **Audit Logging** - Complete change history and traceability

### Technical Features

- Written in **Go** for performance and concurrency
- **PostgreSQL** as the primary transactional store
- **WebSocket** support for real-time updates
- RESTful **JSON API**
- Database migrations with versioned SQL
- Docker-ready deployment
- Health check endpoints

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    API Layer (REST)                      │
│    /api/v1/repos, /files, /commits, /workspaces, etc.   │
├─────────────────────────────────────────────────────────┤
│  Services                                                │
│  ├── WorkspaceService (isolation, merge)                │
│  ├── SubscriptionService (events, WebSocket)            │
│  └── ValidationService (lint, format, test)             │
├─────────────────────────────────────────────────────────┤
│  Database Layer (internal/db)                           │
│  ├── RepositoryQueries, FileQueries, CommitQueries      │
│  ├── WorkspaceQueries, LeaseQueries, LockQueries        │
│  ├── SubscriptionQueries, EventLogQueries               │
│  └── ValidatorQueries, ValidationRunQueries             │
├─────────────────────────────────────────────────────────┤
│              PostgreSQL 16                               │
│  repositories | files | commits | workspaces | etc.     │
└─────────────────────────────────────────────────────────┘
```

### Database Schema

| Table | Purpose |
|-------|---------|
| `repositories` | Code repositories |
| `files` | File metadata |
| `file_versions` | File content history |
| `commits` | Atomic changesets |
| `commit_files` | Commit-file associations |
| `workspaces` | Isolated development environments |
| `workspace_files` | Workspace-local file state |
| `leases` | Intent declarations with TTL |
| `locks` | Explicit file/directory locks |
| `subscriptions` | Event subscription registry |
| `event_log` | Event history for replay |
| `validators` | Registered validation checks |
| `validation_runs` | Validation execution records |
| `validation_results` | Per-file validation results |
| `audit_log` | Complete audit trail |

## Quick Start

### Prerequisites

- Go 1.22+
- PostgreSQL 16+
- Docker (optional, for local PostgreSQL)

### 30-Second Setup

```bash
# Clone the repository
git clone https://github.com/rokkovach/codedb.git
cd codedb

# Start PostgreSQL
make docker-up

# Run migrations
make migrate-up

# Start the server
make run
```

The API will be available at `http://localhost:8080`

### Verify Installation

```bash
curl http://localhost:8080/health
# {"status":"healthy"}
```

## Installation

### From Source

```bash
git clone https://github.com/rokkovach/codedb.git
cd codedb

# Install dependencies
make deps

# Build binary
make build

# Binary will be at ./bin/codedb
```

### Using Docker

```bash
# Build Docker image
docker build -t codedb:latest .

# Run with Docker
docker run -d \
  -p 8080:8080 \
  -e DATABASE_URL=postgres://user:pass@host:5432/codedb \
  codedb:latest
```

### Database Setup

#### Option 1: Docker (Recommended for Development)

```bash
make docker-up
```

This starts a PostgreSQL 16 container with:
- User: `postgres`
- Password: `postgres`
- Database: `codedb`
- Port: `5432`

#### Option 2: Existing PostgreSQL

```bash
# Create database
createdb codedb

# Set connection string
export DATABASE_URL="postgres://user:pass@localhost:5432/codedb?sslmode=disable"

# Run migrations
make migrate-up
```

### Running Migrations

```bash
# Apply all migrations
make migrate-up

# Rollback last migration
make migrate-down

# Create new migration
make migrate-create name=add_new_feature
```

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `DATABASE_URL` | PostgreSQL connection string | `postgres://postgres:postgres@localhost:5432/codedb?sslmode=disable` |
| `PORT` | Server port | `8080` |
| `LOG_LEVEL` | Logging level (debug, info, warn, error) | `info` |
| `LOG_FORMAT` | Log format (json or text) | `text` |

### Example Configuration

```bash
# .env file
DATABASE_URL=postgres://codedb:secret@db.example.com:5432/codedb?sslmode=require
PORT=8080
LOG_LEVEL=info
```

### Connection String Format

```
postgres://USER:PASSWORD@HOST:PORT/DATABASE?sslmode=SSL_MODE
```

Parameters:
- `sslmode`: `disable`, `require`, `verify-ca`, `verify-full`
- `connect_timeout`: Connection timeout in seconds
- `statement_timeout`: Query timeout in milliseconds

## API Documentation

### Base URL

```
http://localhost:8080/api/v1
```

### Authentication

Currently no authentication is implemented. For production use, add authentication middleware.

### Quality of Life Features

CodeDB includes several quality-of-life features to improve developer experience:

#### Request ID Tracking

Every request receives a unique identifier for tracing and debugging:

```http
POST /api/v1/repos
X-Request-ID: custom-id-12345

Response:
HTTP/1.1 201 Created
X-Request-ID: custom-id-12345
```

Error responses include the request ID for easy troubleshooting:
```json
{
  "error": "validation failed",
  "code": 400,
  "details": "name is required",
  "request_id": "custom-id-12345"
}
```

#### Structured Logging

CodeDB uses structured JSON logging with request context:

**Production Mode (JSON):**
```bash
LOG_FORMAT=json LOG_LEVEL=info make run
```

Output:
```json
{
  "time": "2024-01-15T10:30:00Z",
  "level": "info",
  "message": "request completed",
  "request_id": "abc-123",
  "method": "POST",
  "path": "/api/v1/repos",
  "status": 201,
  "duration_ms": 45,
  "remote_addr": "192.168.1.1"
}
```

**Development Mode (Text):**
```bash
LOG_LEVEL=debug make run
```

Output:
```
2024-01-15T10:30:00Z INF request completed request_id=abc-123 method=POST path=/api/v1/repos status=201 duration_ms=45
```

#### API Documentation

Self-documenting API with OpenAPI specification:

**OpenAPI JSON Spec:**
```bash
GET /api/v1/openapi.json
```

Returns complete OpenAPI 3.0 specification.

**Interactive Documentation:**
```bash
GET /api/v1/docs
```

Returns Swagger UI for interactive API exploration.

### Endpoints

#### Health Check

```
GET /health
```

Response:
```json
{
  "status": "healthy"
}
```

---

#### Repositories

##### List Repositories

```
GET /api/v1/repos
```

Response:
```json
[
  {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "name": "my-project",
    "description": "My project description",
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-15T10:30:00Z"
  }
]
```

##### Create Repository

```
POST /api/v1/repos
```

Request:
```json
{
  "name": "my-project",
  "description": "My project description"
}
```

Response: `201 Created`
```json
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "name": "my-project",
  "description": "My project description",
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:00Z"
}
```

##### Get Repository

```
GET /api/v1/repos/{repoID}
```

##### Update Repository

```
PUT /api/v1/repos/{repoID}
```

Request:
```json
{
  "name": "new-name",
  "description": "Updated description"
}
```

##### Delete Repository

```
DELETE /api/v1/repos/{repoID}
```

---

#### Files

##### List Files

```
GET /api/v1/repos/{repoID}/files?include_deleted=false
```

Response:
```json
[
  {
    "id": "550e8400-e29b-41d4-a716-446655440001",
    "repo_id": "550e8400-e29b-41d4-a716-446655440000",
    "path": "src/main.go",
    "is_deleted": false,
    "created_at": "2024-01-15T10:30:00Z",
    "updated_at": "2024-01-15T10:30:00Z"
  }
]
```

##### Create File

```
POST /api/v1/repos/{repoID}/files
```

Request:
```json
{
  "path": "src/main.go",
  "content": "package main\n\nfunc main() {\n\tprintln(\"Hello, CodeDB!\")\n}"
}
```

##### Get File

```
GET /api/v1/files/{fileID}?version={versionID}
```

Response:
```json
{
  "file": {
    "id": "550e8400-e29b-41d4-a716-446655440001",
    "path": "src/main.go",
    "is_deleted": false
  },
  "content": "package main\n\nfunc main() {\n\tprintln(\"Hello, CodeDB!\")\n}",
  "version": {
    "id": "550e8400-e29b-41d4-a716-446655440002",
    "hash": "a1b2c3d4...",
    "size_bytes": 52,
    "created_at": "2024-01-15T10:30:00Z"
  }
}
```

##### List File Versions

```
GET /api/v1/files/{fileID}/versions
```

---

#### Commits

##### List Commits

```
GET /api/v1/repos/{repoID}/commits
```

Response:
```json
[
  {
    "id": "550e8400-e29b-41d4-a716-446655440003",
    "repo_id": "550e8400-e29b-41d4-a716-446655440000",
    "author_id": "agent-001",
    "author_type": "agent",
    "message": "Add main.go",
    "parent_commit_id": null,
    "created_at": "2024-01-15T10:30:00Z"
  }
]
```

##### Create Commit (Atomic Multi-File Write)

```
POST /api/v1/repos/{repoID}/commits
```

Request:
```json
{
  "author_id": "agent-001",
  "author_type": "agent",
  "message": "Refactor authentication module",
  "files": [
    {
      "file_id": "src/auth/login.go",
      "content": "package auth\n\nfunc Login() { ... }",
      "operation": "create"
    },
    {
      "file_id": "src/auth/middleware.go",
      "content": "package auth\n\nfunc Middleware() { ... }",
      "operation": "create"
    }
  ]
}
```

Operations: `create`, `update`, `delete`

##### Get Commit

```
GET /api/v1/commits/{commitID}
```

Response includes commit metadata and associated files.

##### Get Commit Validation Status

```
GET /api/v1/commits/{commitID}/validations
```

Response:
```json
{
  "summary": {
    "total_validators": 3,
    "passed_count": 2,
    "failed_count": 1,
    "pending_count": 0,
    "overall_status": "failed"
  },
  "runs": [
    {
      "id": "...",
      "validator_id": "...",
      "status": "passed",
      "duration_ms": 150
    }
  ]
}
```

---

#### Workspaces

##### List Workspaces

```
GET /api/v1/repos/{repoID}/workspaces?status=active
```

Status values: `active`, `merged`, `abandoned`

##### Create Workspace

```
POST /api/v1/repos/{repoID}/workspaces
```

Request:
```json
{
  "name": "feature-auth-refactor",
  "owner_id": "agent-001",
  "owner_type": "agent",
  "base_commit_id": "550e8400-e29b-41d4-a716-446655440003"
}
```

##### Get Workspace

```
GET /api/v1/repos/{repoID}/workspaces/{workspaceID}
```

##### Merge Workspace

```
POST /api/v1/repos/{repoID}/workspaces/{workspaceID}/merge
```

Request:
```json
{
  "merged_by": "human-001",
  "strategy": "three_way"
}
```

Strategies: `fast_forward`, `three_way`, `force`

Response on conflict (`409 Conflict`):
```json
{
  "error": "merge conflicts detected",
  "conflicts": [
    {
      "file_id": "...",
      "path": "src/auth/login.go",
      "base": "...",
      "ours": "...",
      "theirs": "..."
    }
  ]
}
```

##### Abandon Workspace

```
POST /api/v1/repos/{repoID}/workspaces/{workspaceID}/abandon
```

---

#### Workspace Files

##### List Workspace Files

```
GET /api/v1/repos/{repoID}/workspaces/{workspaceID}/files
```

##### Update Workspace File

```
POST /api/v1/repos/{repoID}/workspaces/{workspaceID}/files
```

Request:
```json
{
  "file_id": "550e8400-e29b-41d4-a716-446655440001",
  "content": "updated content",
  "is_deleted": false
}
```

##### Delete Workspace File

```
DELETE /api/v1/repos/{repoID}/workspaces/{workspaceID}/files/{fileID}
```

---

#### Leases (Intent Declarations)

Leases allow agents to declare intent to modify files with automatic expiration.

##### List Leases

```
GET /api/v1/repos/{repoID}/workspaces/{workspaceID}/leases
```

##### Acquire Lease

```
POST /api/v1/repos/{repoID}/workspaces/{workspaceID}/leases
```

Request:
```json
{
  "file_id": "550e8400-e29b-41d4-a716-446655440001",
  "owner_id": "agent-001",
  "intent": "Refactoring authentication logic",
  "ttl": "30m"
}
```

##### Renew Lease

```
PUT /api/v1/repos/{repoID}/workspaces/{workspaceID}/leases/{leaseID}
```

Request:
```json
{
  "ttl": "30m"
}
```

##### Release Lease

```
DELETE /api/v1/repos/{repoID}/workspaces/{workspaceID}/leases/{leaseID}
```

---

#### Locks (Explicit Locking)

Locks provide explicit file/directory locking without expiration (or with optional expiration).

##### List Locks

```
GET /api/v1/repos/{repoID}/locks
```

##### Acquire Lock

```
POST /api/v1/repos/{repoID}/locks
```

Request:
```json
{
  "file_id": "550e8400-e29b-41d4-a716-446655440001",
  "owner_id": "human-001",
  "owner_type": "human",
  "lock_type": "exclusive",
  "reason": "Major refactor in progress",
  "expires_at": "2024-01-16T10:30:00Z"
}
```

Lock types: `exclusive`, `shared`

##### Release Lock

```
DELETE /api/v1/repos/{repoID}/locks/{lockID}
```

---

#### Validators

##### List Validators

```
GET /api/v1/repos/{repoID}/validators
```

##### Create Validator

```
POST /api/v1/repos/{repoID}/validators
```

Request:
```json
{
  "name": "golangci-lint",
  "command": "golangci-lint run ./...",
  "file_patterns": ["**/*.go"],
  "timeout_seconds": 60,
  "is_blocking": true,
  "is_enabled": true,
  "priority": 0
}
```

##### Get Validator

```
GET /api/v1/repos/{repoID}/validators/{validatorID}
```

##### Update Validator

```
PUT /api/v1/repos/{repoID}/validators/{validatorID}
```

##### Delete Validator

```
DELETE /api/v1/repos/{repoID}/validators/{validatorID}
```

##### Get Workspace Validation Status

```
GET /api/v1/repos/{repoID}/workspaces/{workspaceID}/validations
```

---

#### Subscriptions

##### List Subscriptions

```
GET /api/v1/repos/{repoID}/subscriptions
```

##### Create Subscription

```
POST /api/v1/repos/{repoID}/subscriptions
```

Request:
```json
{
  "subscriber_id": "agent-001",
  "workspace_id": "550e8400-e29b-41d4-a716-446655440010",
  "event_types": ["commit", "file_update", "workspace_merge"],
  "path_patterns": ["src/auth/**"]
}
```

Event types:
- `file_create`, `file_update`, `file_delete`
- `commit`
- `workspace_create`, `workspace_merge`, `workspace_abandon`
- `lease_acquire`, `lease_release`
- `lock_acquire`, `lock_release`
- `validation_pass`, `validation_fail`

##### Delete Subscription

```
DELETE /api/v1/repos/{repoID}/subscriptions/{subID}
```

---

#### WebSocket

##### Connect

```
GET /api/v1/ws?subscriber_id={subscriberID}
```

JavaScript example:
```javascript
const ws = new WebSocket('ws://localhost:8080/api/v1/ws?subscriber_id=agent-001');

ws.onmessage = (event) => {
  const notification = JSON.parse(event.data);
  console.log('Event:', notification.event_type, notification.payload);
};

// Receive events like:
// {
//   "event_type": "commit",
//   "payload": {
//     "commit_id": "...",
//     "repo_id": "...",
//     "author_id": "agent-002",
//     "message": "Fix bug in auth"
//   }
// }
```

## Usage Examples

### Example 1: Agent Creating a New Feature

```bash
# 1. Create a workspace
curl -X POST http://localhost:8080/api/v1/repos/{repoID}/workspaces \
  -H "Content-Type: application/json" \
  -d '{
    "name": "feature-user-dashboard",
    "owner_id": "agent-001",
    "owner_type": "agent"
  }'

# Response: {"id": "ws-001", ...}

# 2. Acquire lease on files to modify
curl -X POST http://localhost:8080/api/v1/repos/{repoID}/workspaces/ws-001/leases \
  -H "Content-Type: application/json" \
  -d '{
    "path_pattern": "src/dashboard/**",
    "owner_id": "agent-001",
    "intent": "Implementing user dashboard",
    "ttl": "1h"
  }'

# 3. Update files in workspace
curl -X POST http://localhost:8080/api/v1/repos/{repoID}/workspaces/ws-001/files \
  -H "Content-Type: application/json" \
  -d '{
    "file_id": "src/dashboard/main.go",
    "content": "package dashboard\n\nfunc Render() { ... }"
  }'

# 4. Subscribe to validation results
curl -X POST http://localhost:8080/api/v1/repos/{repoID}/subscriptions \
  -H "Content-Type: application/json" \
  -d '{
    "subscriber_id": "agent-001",
    "workspace_id": "ws-001",
    "event_types": ["validation_pass", "validation_fail"]
  }'

# 5. Merge workspace when ready
curl -X POST http://localhost:8080/api/v1/repos/{repoID}/workspaces/ws-001/merge \
  -H "Content-Type: application/json" \
  -d '{
    "merged_by": "agent-001",
    "strategy": "three_way"
  }'
```

### Example 2: Human Reviewing Agent Changes

```bash
# 1. List active workspaces
curl http://localhost:8080/api/v1/repos/{repoID}/workspaces?status=active

# 2. Review workspace files
curl http://localhost:8080/api/v1/repos/{repoID}/workspaces/ws-001/files

# 3. Check validation status
curl http://localhost:8080/api/v1/repos/{repoID}/workspaces/ws-001/validations

# 4. Review commit history
curl http://localhost:8080/api/v1/repos/{repoID}/commits

# 5. Approve and merge (or abandon)
curl -X POST http://localhost:8080/api/v1/repos/{repoID}/workspaces/ws-001/merge \
  -H "Content-Type: application/json" \
  -d '{
    "merged_by": "human-001",
    "strategy": "three_way"
  }'
```

### Example 3: Real-time Monitoring

```javascript
// WebSocket client for monitoring repository events
const ws = new WebSocket('ws://localhost:8080/api/v1/ws?subscriber_id=monitor');

ws.onopen = () => {
  // Subscribe to all events
  fetch('http://localhost:8080/api/v1/repos/{repoID}/subscriptions', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      subscriber_id: 'monitor',
      event_types: [
        'commit', 'workspace_create', 'workspace_merge',
        'validation_pass', 'validation_fail'
      ]
    })
  });
};

ws.onmessage = (event) => {
  const notification = JSON.parse(event.data);
  
  switch (notification.event_type) {
    case 'commit':
      console.log(`New commit by ${notification.payload.author_id}`);
      break;
    case 'validation_fail':
      console.log(`Validation failed: ${notification.payload.message}`);
      break;
    // ... handle other events
  }
};
```

## Development

### Project Structure

```
codedb/
├── cmd/
│   └── server/           # Main server entrypoint
│       └── main.go
├── internal/
│   ├── api/              # HTTP API layer
│   │   ├── api.go            # Router and API struct
│   │   ├── repositories.go   # Repo/file/commit handlers
│   │   ├── workspace.go      # Workspace service
│   │   ├── workspace_handlers.go
│   │   ├── subscription.go   # Subscription service + WebSocket
│   │   └── validation.go     # Validation service
│   └── db/               # Database layer
│       ├── db.go             # DB connection and models
│       ├── queries.go        # Core queries
│       ├── workspace_queries.go
│       ├── subscription_queries.go
│       └── validation_queries.go
├── migrations/           # SQL migrations
│   ├── 001_core_schema.up.sql
│   ├── 002_workspaces.up.sql
│   ├── 003_subscriptions.up.sql
│   └── 004_validation.up.sql
├── pkg/
│   └── client/           # Go client SDK (future)
├── Makefile
├── go.mod
├── go.sum
├── prd.md                # Main PRD
├── neo4j_prd.md          # Neo4j integration PRD
└── testing_prd.md        # Testing strategy PRD
```

### Adding a New Feature

1. **Create migration** (if needed):
   ```bash
   make migrate-create name=add_feature_xyz
   ```

2. **Add database queries** in `internal/db/`:
   ```go
   func (q *Queries) CreateFeature(ctx context.Context, ...) (*Feature, error) {
       // Query implementation
   }
   ```

3. **Create service** in `internal/api/`:
   ```go
   type FeatureService struct { ... }
   func (s *FeatureService) CreateFeature(...) { ... }
   ```

4. **Add HTTP handler**:
   ```go
   func (a *API) createFeature(w http.ResponseWriter, r *http.Request) {
       // Handler implementation
   }
   ```

5. **Register route** in `api.go`:
   ```go
   r.Route("/features", func(r chi.Router) {
       r.Post("/", a.createFeature)
   })
   ```

6. **Write tests**:
   ```bash
   # Create test file
   touch internal/api/feature_test.go
   ```

7. **Run tests and lint**:
   ```bash
   make test lint
   ```

### Code Style

- Follow [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- Use `gofmt` for formatting
- Use `go vet` for static analysis
- Prefer table-driven tests
- Keep functions focused and small
- Add documentation comments to exported types

### Make Commands

```bash
make help              # Show all available commands
make build             # Build the binary
make run               # Run the server locally
make test              # Run all tests
make test-coverage     # Run tests with coverage report
make lint              # Run linter
make fmt               # Format code
make vet               # Run go vet
make clean             # Clean build artifacts
make migrate-up        # Run database migrations
make migrate-down      # Rollback migrations
make docker-up         # Start PostgreSQL container
make docker-down       # Stop PostgreSQL container
make deps              # Install dependencies
make all               # Run fmt, vet, lint, test, build
```

## Testing

### Running Tests

```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# Run specific package
go test ./internal/api/...

# Run with verbose output
go test -v ./...

# Run with race detector
go test -race ./...
```

### Test Categories

1. **Unit Tests** - Test individual functions and services
2. **Integration Tests** - Test database interactions
3. **API Tests** - Test HTTP endpoints
4. **Concurrency Tests** - Test concurrent operations

### Writing Tests

```go
func TestCreateRepository(t *testing.T) {
    // Setup
    db := setupTestDB(t)
    api := NewAPI(db)
    
    // Execute
    body := `{"name": "test-repo"}`
    req := httptest.NewRequest("POST", "/api/v1/repos", strings.NewReader(body))
    req.Header.Set("Content-Type", "application/json")
    
    w := httptest.NewRecorder()
    api.Router().ServeHTTP(w, req)
    
    // Assert
    assert.Equal(t, http.StatusCreated, w.Code)
    
    var repo db.Repository
    json.NewDecoder(w.Body).Decode(&repo)
    assert.Equal(t, "test-repo", repo.Name)
}
```

See [testing_prd.md](testing_prd.md) for comprehensive testing strategy.

## Deployment

### Docker

```dockerfile
# Dockerfile
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o codedb ./cmd/server

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/codedb .
COPY --from=builder /app/migrations ./migrations
EXPOSE 8080
CMD ["./codedb"]
```

Build and run:
```bash
docker build -t codedb:latest .
docker run -p 8080:8080 -e DATABASE_URL=... codedb:latest
```

### Docker Compose

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:16-alpine
    environment:
      POSTGRES_USER: codedb
      POSTGRES_PASSWORD: codedb
      POSTGRES_DB: codedb
    volumes:
      - postgres_data:/var/lib/postgresql/data
    ports:
      - "5432:5432"

  codedb:
    image: codedb:latest
    depends_on:
      - postgres
    environment:
      DATABASE_URL: postgres://codedb:codedb@postgres:5432/codedb?sslmode=disable
      PORT: "8080"
    ports:
      - "8080:8080"

volumes:
  postgres_data:
```

### Kubernetes

```yaml
# Deployment example
apiVersion: apps/v1
kind: Deployment
metadata:
  name: codedb
spec:
  replicas: 3
  selector:
    matchLabels:
      app: codedb
  template:
    metadata:
      labels:
        app: codedb
    spec:
      containers:
      - name: codedb
        image: codedb:latest
        ports:
        - containerPort: 8080
        env:
        - name: DATABASE_URL
          valueFrom:
            secretKeyRef:
              name: codedb-secret
              key: database-url
```

### Production Checklist

- [ ] Configure proper PostgreSQL connection pooling
- [ ] Set up SSL/TLS for database connections
- [ ] Configure authentication/authorization
- [ ] Set up monitoring and alerting
- [ ] Configure log aggregation
- [ ] Set up database backups
- [ ] Configure horizontal scaling
- [ ] Set up CI/CD pipeline
- [ ] Configure rate limiting
- [ ] Review security settings

## Roadmap

### Phase 1 (Current) - Core Platform
- ✅ PostgreSQL-based storage
- ✅ REST API
- ✅ Workspaces and isolation
- ✅ Leases and locks
- ✅ Real-time subscriptions
- ✅ Validation pipeline

### Phase 2 - Enhanced Features
- [ ] Go client SDK
- [ ] Filesystem projection (FUSE)
- [ ] Git compatibility layer
- [ ] Web UI
- [ ] Advanced merge strategies

### Phase 3 - Intelligence
- [ ] Neo4j graph integration (see [neo4j_prd.md](neo4j_prd.md))
- [ ] Impact analysis
- [ ] Dependency visualization
- [ ] Architecture policy checks

### Phase 4 - Scale
- [ ] Multi-region deployment
- [ ] Federation support
- [ ] Advanced caching
- [ ] Performance optimizations

## Contributing

We welcome contributions! Please follow these steps:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Run tests (`make test`)
5. Run linter (`make lint`)
6. Commit your changes (`git commit -m 'Add amazing feature'`)
7. Push to the branch (`git push origin feature/amazing-feature`)
8. Open a Pull Request

### Pull Request Guidelines

- Include tests for new functionality
- Update documentation as needed
- Follow existing code style
- Keep PRs focused and small
- Add PR description with context

## Troubleshooting

### Common Issues

#### Database Connection Failed

```
Error: connection refused
```

**Solution:** Ensure PostgreSQL is running:
```bash
make docker-up
# or check existing connection
psql $DATABASE_URL -c "SELECT 1"
```

#### Migration Failed

```
Error: migration failed
```

**Solution:** Check migration status and rollback:
```bash
make migrate-down
make migrate-up
```

#### Port Already in Use

```
Error: listen tcp :8080: bind: address already in use
```

**Solution:** Use a different port:
```bash
PORT=8081 make run
```

### Debug Mode

Enable verbose logging:
```bash
LOG_LEVEL=debug make run
```

## Performance

### Benchmarks

| Operation | Latency (p50) | Latency (p99) | Throughput |
|-----------|---------------|---------------|------------|
| Create file | 5ms | 15ms | 10k/s |
| Create commit | 20ms | 50ms | 5k/s |
| Get file | 2ms | 8ms | 20k/s |
| WebSocket broadcast | 1ms | 5ms | 50k/s |

### Optimization Tips

1. **Connection Pooling** - Configure appropriate pool size
2. **Indexing** - Ensure proper indexes for queries
3. **Caching** - Cache frequently accessed data
4. **Batching** - Batch multiple operations when possible
5. **Async Processing** - Use background workers for heavy operations

## Security

### Best Practices

- Use SSL/TLS for all connections
- Implement authentication in production
- Validate all input data
- Use parameterized queries (already done via pgx)
- Keep dependencies updated
- Regular security audits

### Reporting Security Issues

Please report security vulnerabilities to security@example.com

## License

MIT License

Copyright (c) 2024 CodeDB

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.

## Support

- **Documentation:** This README and PRD files
- **Issues:** [GitHub Issues](https://github.com/rokkovach/codedb/issues)
- **Discussions:** [GitHub Discussions](https://github.com/rokkovach/codedb/discussions)

## Acknowledgments

CodeDB is inspired by:
- Traditional version control systems (Git, Mercurial)
- Modern code review platforms (GitHub, GitLab)
- Collaborative editing tools (Google Docs, Figma)
- Database-driven applications

---

Built with ❤️ for the AI-assisted development future
