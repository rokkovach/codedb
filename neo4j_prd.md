# CodeDB Neo4j Add-on PRD

## Title
CodeDB Graph: Neo4j Add-on for Relationship Intelligence in a Go-Based CodeDB Platform

## Overview
CodeDB Graph is an optional Neo4j-powered add-on for CodeDB that models code relationships as a graph to support dependency analysis, impact tracing, semantic navigation, and agent reasoning.

The base CodeDB platform is implemented in Go and uses PostgreSQL as the primary transactional source of truth. The Neo4j add-on complements that core by maintaining a graph projection of repository structure, symbols, references, imports, ownership, and change relationships for graph-heavy queries and workflows.

## Problem
The core CodeDB platform can reliably store files, versions, changesets, validations, and event history in PostgreSQL. However, some developer and agent workflows are inherently graph-shaped and become awkward or inefficient in a purely relational model.

Examples include:
- Finding all transitive dependents of a symbol, module, or service
- Estimating blast radius from a proposed change
- Tracing call chains and reference chains across a codebase
- Detecting architectural boundary violations
- Identifying central or fragile nodes in the dependency graph
- Supporting agent planning over code relationships

These use cases benefit from a graph projection where relationships are first-class entities and traversal is a primary access pattern.

## Vision
Make the codebase navigable as a live graph so that humans and AI agents can reason over dependencies, ownership, structure, and impact in real time.

## Objective
Provide an optional graph intelligence layer that projects CodeDB repository and analysis data into Neo4j, enabling fast relationship traversal, review enrichment, architectural policy evaluation, and graph-aware agent workflows.

## Users
- AI agents that need dependency-aware planning
- Human reviewers assessing change impact
- Platform and architecture teams enforcing boundaries
- Developers exploring unfamiliar codebases
- Engineering managers investigating hotspots and coordination risk

## Goals
1. Build and maintain a graph projection of repository structure and code relationships
2. Support low-latency multi-hop dependency and impact queries
3. Improve review with graph-based blast-radius analysis
4. Enable agents to reason over symbols, modules, and service boundaries
5. Support architectural policy checks using graph traversals
6. Expose graph-backed APIs from Go services
7. Keep the graph synchronized with CodeDB snapshots and changesets

## Non-Goals
- Replacing PostgreSQL as the primary write path
- Making Neo4j mandatory for all deployments
- Providing perfect semantic understanding for every language in v1
- Replacing full static analysis or language servers
- Running all validation logic inside Neo4j

## Why Neo4j
Neo4j is a strong fit for relationship-heavy use cases involving:
- `CALLS`
- `IMPORTS`
- `DEFINES`
- `REFERENCES`
- `DEPENDS_ON`
- `OWNS`
- `MODIFIED`
- `AFFECTS`
- `VIOLATES`

The add-on should answer questions such as:
- What breaks if this symbol changes?
- Which services depend transitively on this package?
- What paths connect this API to this datastore?
- Which owners should review a changeset touching this path?
- Which modules violate the approved architecture graph?

## Platform Context

### Base Platform
- Language: Go
- Primary transactional database: PostgreSQL
- Primary DB driver: `pgx`
- Event and subscription model: PostgreSQL event stream plus application-level push channels
- Optional local or embedded components: SQLite where useful for isolated agent sandboxes
- Optional graph layer: Neo4j

### Integration Principle
PostgreSQL remains the source of truth for transactional code operations. Neo4j stores a derived graph projection optimized for traversal and relationship queries. Graph sync is event-driven from the Go application and CodeDB commit pipeline.

## Product Scope
The Neo4j add-on includes:
- Graph schema for code entities and relationships
- Go-based projection and sync services
- Sync pipeline from CodeDB snapshots and changesets
- Graph query APIs exposed by Go services
- Review-time impact analysis
- Architectural policy checks
- Optional graph analytics for hotspots and centrality

## User Problems to Solve

### For AI agents
- Determine relevant code before editing
- Identify transitive dependencies and risky paths
- Plan edits with architectural context
- Ask relationship-heavy questions over the codebase graph

### For reviewers
- Understand blast radius quickly
- See transitive dependencies and affected owners
- Evaluate architecture impact, not only file diff
- Prioritize risky changesets

### For platform teams
- Detect forbidden dependencies
- Enforce service and layer boundaries
- Track ownership and critical dependency chains
- Surface architectural drift over time

## Functional Requirements

### 1. Graph Projection
The system must project CodeDB entities into Neo4j, including:
- repositories
- directories
- files
- modules
- packages
- classes
- functions
- methods
- services
- APIs
- tests
- teams
- owners
- changesets
- validation runs

### 2. Relationship Modeling
The system must support relationship types such as:
- `CONTAINS`
- `DECLARES`
- `CALLS`
- `IMPORTS`
- `DEPENDS_ON`
- `IMPLEMENTS`
- `EXTENDS`
- `REFERENCES`
- `TESTS`
- `OWNS`
- `MODIFIED_IN`
- `AFFECTS`
- `VIOLATES`

Relationships should support metadata such as:
- confidence
- language
- origin type
- path depth
- criticality
- frequency
- revision
- freshness timestamp

### 3. Sync and Freshness
- Graph updates must be triggered from committed changesets or snapshot events
- Incremental sync is preferred over full rebuilds
- Full rebuild tooling must exist for drift recovery
- Sync status must be observable
- Graph reads should expose freshness metadata

### 4. Graph Query APIs
The add-on must expose Go service APIs for:
- transitive dependency lookup
- upstream/downstream traversal
- blast-radius analysis
- call chain exploration
- ownership tracing
- policy evaluation
- related-symbol discovery
- graph neighborhood queries

### 5. Review Enrichment
For each eligible changeset, the add-on should:
- estimate blast radius
- identify transitive dependents
- surface affected owners and teams
- attach graph-derived risk metadata
- identify architecture rule violations

### 6. Policy and Governance
The add-on should support graph-backed rules such as:
- forbidden service-to-service dependencies
- no direct API-to-database access
- restricted imports for sensitive packages
- owner approval required for critical paths
- warnings on edits to high-centrality nodes

### 7. Agent Interfaces
The system should support graph-aware agent queries such as:
- "find all symbols impacted by modifying X"
- "find shortest path from service A to datastore B"
- "rank candidate files related to this task"
- "identify safe edit region around this module"
- "show architecture constraints for this component"

## Example Queries
- Which services transitively depend on `billing/core/ledger`?
- Which files reference functions defined in this package?
- What is the shortest dependency path between this endpoint and persistence layer?
- Which owners should review a changeset touching this graph neighborhood?
- Which modules violate the approved architecture policy?

## Architecture

### High-Level Design
1. CodeDB stores transactional code and metadata in PostgreSQL
2. Go parser/indexer services extract symbols and relationships from changed files
3. A Go graph projection service writes graph updates into Neo4j
4. A Go query service exposes graph-backed APIs to UI, reviewers, and agents
5. A policy engine evaluates graph-based rules during review or promotion

### Core Components
- CodeDB API service in Go
- PostgreSQL transactional store
- Go event consumer for committed changesets
- Language parsers / AST extraction pipeline
- Go graph projection writer
- Neo4j database
- Graph query API service in Go
- Review enrichment service
- Policy evaluation service
- Observability and drift monitor

## Go Implementation Notes

### Recommended Libraries
- PostgreSQL access: `github.com/jackc/pgx`
- Neo4j access: official `github.com/neo4j/neo4j-go-driver/v5/neo4j`
- HTTP APIs: standard `net/http`, Chi, Gin, or Echo
- Background workers: native goroutines plus worker pool pattern
- Event distribution: PostgreSQL `LISTEN/NOTIFY` or application event bus
- Structured logging: Zap, Zerolog, or Logrus
- Metrics: Prometheus instrumentation

### Service Responsibilities
- API service handles repository and changeset requests
- Commit pipeline emits events after successful PostgreSQL transactions
- Graph sync worker consumes commit events and updates Neo4j
- Query service resolves graph traversal requests
- Review service enriches changesets with graph risk metadata

### Data Flow
1. A changeset is committed transactionally in PostgreSQL
2. The Go service emits an event after commit
3. A graph sync worker receives the event
4. A parser extracts updated symbols and relationships
5. Neo4j is updated with changed nodes and edges
6. Review and agent services query Neo4j for impact and relationship analysis

## Graph Data Model

### Node Types
- Repository
- Directory
- File
- Module
- Package
- Class
- Function
- Method
- Interface
- Service
- API
- Test
- Team
- Owner
- Changeset
- ValidationRun

### Relationship Types
- `CONTAINS`
- `DECLARES`
- `CALLS`
- `IMPORTS`
- `DEPENDS_ON`
- `IMPLEMENTS`
- `EXTENDS`
- `REFERENCES`
- `TESTS`
- `OWNS`
- `MODIFIED_IN`
- `TRIGGERED`
- `AFFECTS`
- `VIOLATES`

## Sync Strategy

### Initial Load
- Build graph from a stable repository snapshot
- Parse supported languages
- Populate nodes and relationships
- Record projection version and snapshot lineage

### Incremental Sync
- Listen for committed changesets
- Re-parse only affected files or modules
- Upsert changed nodes and edges
- Remove stale relationships
- Recompute derived metadata selectively

### Recovery
- Support full rebuild when drift is detected
- Compare graph lineage against repository lineage
- Surface lag, failure, and drift metrics

## MVP Scope
The MVP should include:
- Go-based services for sync and query
- Python codebase support first if product scope still targets Python repos, otherwise first target language should match initial customer demand
- File, module, class, and function nodes
- `IMPORTS`, `CALLS`, `DECLARES`, and `DEPENDS_ON` edges
- Sync from committed changesets
- Basic blast-radius API
- Review enrichment with transitive dependents
- Simple architecture policy checks
- Ownership overlay support

## Out of Scope for MVP
- Full polyglot semantic support
- Runtime trace ingestion
- Cross-repo enterprise graph federation
- Automatic code rewriting from graph algorithms
- Advanced analytics dashboards
- Perfect dynamic call graph reconstruction

## Success Metrics
- Median latency from commit to graph freshness
- P95 latency for transitive dependency queries
- Percentage of reviews enriched with graph data
- Reduction in reviewer time to assess impact
- Number of architecture violations detected automatically
- Improvement in agent task success with graph context
- Graph drift rate relative to repository state

## Risks
- Parser quality may limit graph completeness
- Multi-language support may create inconsistent graph fidelity
- Sync drift can reduce trust in graph results
- Operating PostgreSQL and Neo4j together increases system complexity
- Teams may over-trust graph output as complete truth

## Open Questions
- Should graph sync happen synchronously on commit or asynchronously after commit?
- Which repository languages should be supported first?
- How much metadata should be stored in Neo4j vs derived at query time?
- Should ownership and architecture policy data live in PostgreSQL, Neo4j, or both?
- When should graph algorithms such as centrality and community detection be introduced?

## Rollout Plan

### Phase 1
- Define graph schema
- Build Go projection service
- Build parser pipeline for initial language support
- Project repository and symbol relationships into Neo4j
- Expose basic impact and dependency APIs

### Phase 2
- Add review enrichment and architecture policy checks
- Add ownership overlays
- Improve incremental sync and drift monitoring

### Phase 3
- Add hotspot and centrality analysis
- Expand language support
- Add richer agent graph planning features
- Evaluate cross-repository graph federation

## Technical Recommendation
Use Neo4j as an optional graph intelligence layer in a Go-based CodeDB architecture. Keep PostgreSQL as the canonical transactional store and use Go services to synchronize, query, and operationalize the graph.

## Summary
CodeDB Graph adds Neo4j-powered relationship intelligence to a Go-based CodeDB platform. It enables dependency-aware development, faster impact analysis, stronger architecture governance, and better agent planning without moving the source of truth away from PostgreSQL.
