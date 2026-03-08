# CodeDB Testing PRD

## Title
CodeDB Testing Strategy: Unit and End-to-End Coverage for a Go-Based Transactional Code Platform

## Overview
This PRD defines the testing strategy for CodeDB, a Go-based platform that stores and coordinates code changes through PostgreSQL, with optional Neo4j graph capabilities.

The goal is to establish strong confidence in correctness, reliability, and release safety through a layered testing strategy centered on:
- high-value unit test coverage
- critical-path end-to-end coverage
- deterministic integration testing for PostgreSQL and Neo4j components
- fast feedback in CI
- clear quality gates for merges and releases

## Problem
CodeDB is a transactional, event-driven system with multiple high-risk surfaces:
- multi-file atomic writes
- concurrency and conflict handling
- event propagation and subscriptions
- validation pipelines
- review enrichment
- graph projection and sync
- filesystem projection and compatibility behavior

Without a deliberate testing strategy, the platform risks:
- regressions in transactional behavior
- race conditions in concurrent agent workflows
- drift between PostgreSQL and Neo4j projections
- flaky change propagation and subscription failures
- broken review and promotion logic
- low confidence in releases

## Vision
Create a testing system that gives the team rapid feedback during development and strong confidence in production behavior, while remaining fast enough to support frequent iteration.

## Objective
Build a practical, enforceable testing program for CodeDB that delivers broad unit coverage, robust end-to-end validation of core workflows, and reliable CI/CD quality gates.

## Users
- Backend engineers building CodeDB services in Go
- Platform engineers maintaining CI/CD pipelines
- QA or developer productivity engineers
- Tech leads responsible for release quality
- Contributors building PostgreSQL, event, and Neo4j features

## Goals
1. Achieve strong unit coverage across core business logic
2. Achieve stable end-to-end coverage for critical product workflows
3. Validate PostgreSQL transactional correctness under realistic scenarios
4. Validate Neo4j graph sync and query correctness where enabled
5. Reduce regressions in concurrency, validation, review, and sync logic
6. Keep test suites reliable enough for branch protection and release gating
7. Make failures easy to diagnose and fix

## Non-Goals
- Chasing 100 percent coverage at the expense of usefulness
- Replacing production observability with tests
- Building a massive manual QA process
- Testing every edge case only through E2E flows
- Making local development depend on slow full-stack test runs for every change

## Testing Principles
- Prefer many small unit tests for logic-heavy code
- Reserve E2E tests for critical business flows
- Keep tests deterministic and isolated
- Avoid flaky network- and timing-dependent assertions where possible
- Test behavior and invariants, not implementation details
- Run the fastest tests most often
- Treat concurrency, transactions, and sync behavior as first-class test targets

## Scope
This testing PRD applies to:
- Go API services
- PostgreSQL transactional code paths
- event and subscription delivery
- validation and promotion pipelines
- review enrichment
- Neo4j graph projection and graph query behavior
- filesystem projection layer
- CI/CD test enforcement

## Test Pyramid

### 1. Unit Tests
Unit tests validate isolated functions, domain logic, validation rules, diff logic, conflict detection rules, policy evaluation, and transformation code.

These should be the largest test layer and provide the fastest feedback.

### 2. Integration Tests
Integration tests validate behavior at service and storage boundaries, including:
- PostgreSQL repositories
- transaction handling
- migrations
- event publication and consumption
- Neo4j graph writes and reads
- filesystem projection adapters

### 3. End-to-End Tests
E2E tests validate complete workflows across service boundaries, storage, eventing, validation, review, and graph sync where relevant.

These should cover only the highest-value user and system journeys.

## Product Requirements

## 1. Unit Testing Program

### Requirements
- All core domain packages must have unit tests
- Business logic should be structured to minimize dependency on I/O
- Table-driven tests should be used where input-output cases repeat
- Error paths and edge cases must be explicitly tested
- Critical invariants must be tested, including atomicity assumptions, conflict rules, and promotion rules

### Priority Areas
- changeset validation
- path and file operation rules
- diff and patch application
- conflict detection
- lease and intent handling
- review risk scoring
- policy evaluation
- graph relationship mapping
- event payload construction

### Quality Target
- Unit coverage target for core domain packages: 85 percent or higher
- Lower-level utility packages may target higher
- Thin transport or wiring layers do not need artificial coverage inflation

## 2. Integration Testing Program

### PostgreSQL Integration Tests
The system must validate:
- migration correctness
- repository CRUD behavior
- transaction rollback and commit semantics
- optimistic concurrency behavior
- event outbox or commit event behavior
- snapshot and workspace isolation
- lock or lease persistence semantics

### Neo4j Integration Tests
Where Neo4j is enabled, the system must validate:
- graph projection writes
- node and edge upserts
- stale relationship cleanup
- graph query correctness
- sync idempotency
- drift recovery behavior

### External Dependency Strategy
- PostgreSQL integration tests must run against a real PostgreSQL instance
- Neo4j integration tests must run against a real Neo4j instance when graph features are enabled
- Mocking should not replace real database validation for storage adapters
- Docker-based ephemeral test environments are acceptable for CI

## 3. End-to-End Testing Program

### Critical E2E Flows
The system must have E2E tests for at least the following:

#### Repository Write Flow
1. create repository
2. create workspace
3. stage multi-file change
4. commit changeset
5. verify persisted state
6. verify events emitted
7. verify validation triggered

#### Conflict Flow
1. create two workspaces from same base
2. apply overlapping edits
3. commit first changeset
4. verify second commit fails or resolves according to policy
5. verify conflict metadata is returned correctly

#### Review Flow
1. commit changeset
2. trigger validation pipeline
3. enrich review payload
4. verify risk metadata and validation status
5. approve or reject according to policy

#### Promotion Flow
1. commit valid changeset
2. pass required checks
3. promote to stable snapshot
4. verify lineage and history state

#### Event Subscription Flow
1. connect subscriber
2. commit relevant change
3. verify correct event delivery
4. verify payload content and ordering expectations

#### Graph Sync Flow
1. commit source code change
2. trigger graph projection update
3. verify graph nodes and edges updated
4. verify impact query returns expected results

#### Rollback / Recovery Flow
1. simulate failed validation or failed promotion
2. verify state consistency
3. verify partial writes are not visible
4. verify audit and event records remain correct

### E2E Quality Target
- 100 percent coverage of critical business flows
- E2E suite should optimize for confidence, not raw count of scenarios
- Flaky tests must be treated as production-blocking quality issues

## 4. Concurrency and Reliability Testing

### Requirements
The testing program must include scenarios for:
- simultaneous changeset creation
- overlapping file edits
- lease contention
- duplicate event handling
- retry and idempotency behavior
- graph sync race conditions
- subscription reconnect behavior

### Approach
- use deterministic concurrency tests where possible
- add targeted stress tests for hot paths
- run race detection in CI for Go packages where appropriate
- validate idempotency for consumers and sync workers

## 5. CI/CD Quality Gates

### Pull Request Gates
Every pull request must run:
- unit tests
- static analysis
- linting
- selected integration tests for impacted packages

### Main Branch Gates
Main branch validation must run:
- full unit suite
- full integration suite
- core E2E suite
- migration checks
- race detection for relevant packages

### Release Gates
Release validation must run:
- all branch and main checks
- full E2E regression suite
- optional performance smoke tests
- optional upgrade and rollback tests

## 6. Coverage Policy

### Coverage Expectations
- Core domain logic: 85 percent minimum
- Storage adapters: strong integration coverage required even if unit coverage is lower
- Critical workflows: must be represented in E2E
- New features: no merge without tests at the appropriate layer

### Enforcement
- Coverage thresholds should be measured per package category, not only globally
- A high global percentage must not hide untested critical modules
- PRs that reduce coverage in critical packages require explicit approval

## 7. Test Environment Strategy

### Local Development
Developers should be able to run:
- unit tests without external services
- targeted integration tests with local or containerized PostgreSQL
- optional Neo4j integration tests when working on graph features

### CI Environment
CI should provide:
- isolated PostgreSQL instance
- isolated Neo4j instance when needed
- reproducible migration setup
- deterministic test data seeding
- artifact collection for logs and failure diagnostics

## 8. Test Data Strategy
- Use small deterministic fixtures
- Prefer builder helpers over large static fixtures
- Seed only what a test needs
- Avoid hidden shared mutable test state
- Use unique repository and workspace identifiers per test
- Reset or recreate database state between integration and E2E runs

## 9. Observability for Tests
The platform should expose test-friendly diagnostics including:
- structured logs
- event traces
- sync lag metrics
- validation run status
- graph projection freshness
- changeset lineage and audit history

These diagnostics should be attachable to CI failures to reduce debugging time.

## 10. Tooling Recommendations

### Go Testing
- standard `go test`
- table-driven tests
- `t.Cleanup()` for cleanup behavior
- race detector for concurrency-sensitive packages

### PostgreSQL Testing
- run integration tests against real PostgreSQL
- use migration setup once, then isolated test databases or schemas
- use transaction rollback or ephemeral databases for test isolation

### Neo4j Testing
- run graph integration tests against a real Neo4j instance
- validate projection correctness through explicit graph assertions
- test sync idempotency and stale edge cleanup

### E2E Harness
- start services in a controlled environment
- drive API-level workflows
- assert final persisted state and emitted events
- avoid brittle UI-driven E2E unless UI is a product requirement

## 11. Acceptance Criteria

### Unit Coverage
- Core domain packages reach at least 85 percent coverage
- Every critical business rule has direct unit tests
- New domain logic cannot merge without tests

### Integration Coverage
- PostgreSQL repositories and transaction boundaries are tested against real PostgreSQL
- Neo4j projection and query behavior are tested against real Neo4j when enabled
- Migration tests run in CI

### End-to-End Coverage
- All critical workflows listed in this PRD have automated E2E tests
- E2E failure rate remains low and tracked over time
- Flaky E2E tests are triaged immediately

### CI/CD Enforcement
- Branch protection requires passing test gates
- Coverage regression rules are enforced on critical packages
- Release process includes full regression execution

## 12. Metrics
- unit test runtime
- integration test runtime
- E2E runtime
- package-level coverage
- flaky test rate
- mean time to diagnose failed CI runs
- escaped defect rate in production
- number of regressions caught pre-merge
- graph sync defect rate
- concurrency defect rate

## 13. Risks
- Over-reliance on E2E tests can slow delivery
- Excessive mocking can hide real storage and sync bugs
- Poor test isolation can create flaky failures
- Coverage targets can become vanity metrics if not tied to risk
- Slow CI can cause teams to bypass quality discipline

## 14. Open Questions
- What are the exact minimum merge gates by branch type?
- Should Neo4j tests run on every PR or only when graph-related code changes?
- Which packages are considered core for the 85 percent threshold?
- Should subscription delivery be validated through E2E only or also through dedicated integration harnesses?
- What release cadence should determine full regression frequency?

## 15. Rollout Plan

### Phase 1
- define package test categories
- establish coverage reporting
- add unit tests for core domain logic
- add PostgreSQL integration baseline
- add CI branch gates

### Phase 2
- add E2E coverage for repository, conflict, validation, and review flows
- add race detection in CI
- add Neo4j integration coverage for graph-enabled deployments

### Phase 3
- expand E2E to promotion, recovery, and subscription flows
- add stress and idempotency tests
- optimize CI runtime and failure diagnostics

## Summary
CodeDB requires a layered testing strategy centered on broad unit coverage and focused end-to-end validation of critical workflows. The testing program should prioritize correctness of transactions, concurrency, eventing, validation, review, and optional graph sync while remaining fast, deterministic, and enf
