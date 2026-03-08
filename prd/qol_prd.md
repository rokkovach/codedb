# CodeDB Quality of Life PRD

## Title
CodeDB Quality of Life Improvements: Enhanced Developer Experience

## Overview
This PRD defines three quality-of-life improvements for CodeDB that significantly enhance the developer and operator experience: Request ID tracking, structured logging, and self-documenting API endpoints.

## Problem
CodeDB v1 provides core functionality but lacks several features that would improve daily usage:
- Difficult to trace requests through the system for debugging
- Logs lack structure and context for operational visibility
- No self-documenting API, requiring external documentation lookups

## Goals
1. Implement request ID tracking for all API calls
2. Add structured JSON logging with request context
3. Provide OpenAPI documentation via API endpoint
4. Improve error messages with traceable identifiers
5. Enable better observability for production operations

## Non-Goals
- Full observability platform (metrics, tracing)
- Authentication/authorization
- Rate limiting
- Advanced caching

## Features

### Feature 1: Request ID Tracking

#### Description
Every HTTP request receives a unique identifier that is:
- Generated if not provided
- Returned in response headers (`X-Request-ID`)
- Included in all log entries for that request
- Included in error responses

#### Requirements
- Generate UUID v4 for each request
- Accept `X-Request-ID` header to use caller-provided ID
- Add request ID to response headers
- Include request ID in error response bodies
- Propagate request ID through context

#### API Changes

Request:
```http
POST /api/v1/repos
X-Request-ID: custom-id-12345
```

Response:
```http
HTTP/1.1 201 Created
X-Request-ID: custom-id-12345
```

Error Response:
```json
{
  "error": "validation failed",
  "code": 400,
  "details": "name is required",
  "request_id": "custom-id-12345"
}
```

#### Implementation
- Middleware that wraps each request
- Request ID stored in context
- Response header injection
- Error response enrichment

---

### Feature 2: Structured Logging

#### Description
Implement structured JSON logging with:
- Request context (request ID, method, path)
- Response information (status, duration)
- Log levels (debug, info, warn, error)
- Configurable output format (JSON for production, text for dev)

#### Requirements
- JSON log format for production
- Human-readable format for development
- Include request context in all logs
- Log request start and end
- Log response duration
- Configurable log level

#### Log Format

Production (JSON):
```json
{
  "time": "2024-01-15T10:30:00Z",
  "level": "info",
  "msg": "request completed",
  "request_id": "abc-123",
  "method": "POST",
  "path": "/api/v1/repos",
  "status": 201,
  "duration_ms": 45,
  "remote_addr": "192.168.1.1"
}
```

Development (Text):
```
2024-01-15T10:30:00Z INFO request completed request_id=abc-123 method=POST path=/api/v1/repos status=201 duration_ms=45
```

#### Implementation
- Use `zerolog` for structured logging
- Middleware for request logging
- Log level configuration via environment
- Service-level logging helpers

---

### Feature 3: API Documentation Endpoint

#### Description
Provide OpenAPI 3.0 specification via API endpoint for self-documenting API.

#### Requirements
- Generate OpenAPI 3.0 spec
- Serve at `/api/v1/openapi.json` or `/api/v1/docs`
- Include all endpoints
- Include request/response schemas
- Include error responses

#### API Changes

```http
GET /api/v1/openapi.json
```

Response:
```json
{
  "openapi": "3.0.0",
  "info": {
    "title": "CodeDB API",
    "version": "1.0.0"
  },
  "paths": {
    "/api/v1/repos": {
      "get": {
        "summary": "List repositories",
        "responses": {
          "200": {
            "description": "List of repositories"
          }
        }
      }
    }
  }
}
```

#### Implementation
- Define OpenAPI spec as Go data structures
- Generate spec programmatically
- Serve as JSON endpoint
- Include in health check

---

## Implementation Plan

### Phase 1: Request ID Tracking
1. Create request ID middleware
2. Add request ID to context
3. Update error responses
4. Update response headers

### Phase 2: Structured Logging
1. Add zerolog dependency
2. Create logging middleware
3. Configure log levels
4. Add request/response logging

### Phase 3: API Documentation
1. Define OpenAPI structures
2. Generate spec from routes
3. Add documentation endpoint
4. Update README

## Testing Strategy

### Request ID Tests
- Verify ID generation for new requests
- Verify ID preservation for provided IDs
- Verify ID in response headers
- Verify ID in error responses

### Logging Tests
- Verify JSON format in production mode
- Verify text format in development mode
- Verify request context in logs
- Verify log level filtering

### Documentation Tests
- Verify OpenAPI endpoint returns valid spec
- Verify spec includes all endpoints
- Verify spec validates against OpenAPI 3.0 schema

## Success Metrics
- 100% of requests have traceable IDs
- All logs include request context
- API is self-documenting
- Reduced debugging time for issues
