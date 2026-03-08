# CodeDB PRD

## Title
CodeDB: Database-Native Collaborative Code Authoring for Humans and AI Agents

## Overview
CodeDB is a database-native system for collaborative software development where the source of truth is stored in a transactional database rather than a filesystem checkout. It is designed for teams using many AI agents and human developers concurrently, where traditional Git workflows become a bottleneck for coordination, review, and validation.

## Problem
Modern code generation velocity has increased significantly, especially with AI agents producing changes in parallel. In this environment, the traditional workflow of checkout, branch, push, pull request, review, and rebase creates friction.

Current issues include:
- Pull request review becoming the main bottleneck for high-output teams
- Filesystem-based code storage being poorly suited for 10 to 1000 concurrent agents
- Weak write-level atomicity across multiple files
- Poor coordination primitives across agents making overlapping edits hard to manage
- Limited real-time subscriptions to codebase state changes
- Delayed linting, formatting, and review instead of continuous verification

## Vision
Treat code as high-velocity structured data instead of static files. Build a system where code changes are committed transactionally, coordinated through database-native primitives, and continuously validated in real time.

## Objective
Enable teams with many concurrent AI agents to write, validate, coordinate, and review code through a transactional database-native workflow that reduces reliance on filesystem-first branching and pull-request-centric synchronization.

## Target Users
- AI-first software engineering teams
- Internal platform teams building coding agents
- Small teams experimenting with parallel agent-based development
- Research teams exploring high-concurrency code generation workflows

## Goals
1. Store code and metadata in a transactional source of truth
2. Support concurrent agent writes with atomic multi-file commits
3. Provide real-time subscriptions to codebase changes
4. Run linting, formatting, and validation continuously at write time
5. Improve coordination across agents through leases, locks, and intent declarations
6. Preserve complete audit history, rollback, and replayability
7. Support compatibility with existing tools through filesystem projections

## Non-Goals
- Replacing Git for all external or open-source collaboration in v1
- Building a full IDE in v1
- Supporting deep semantic merge for every programming language in v1
- Eliminating the filesystem entirely for all downstream tools in v1

## Core Principles
- Database-first source of truth
- Atomic write operations
- Isolated workspaces for humans and agents
- Event-driven coordination
- Continuous automated verification
- Full auditability and traceability
- Compatibility with existing developer tooling

## Proposed Solution
CodeDB stores repositories, files, changesets, symbols, validation state, and agent activity in a transactional database. Instead of treating the checked-out filesystem as the canonical source, the database becomes the primary representation of the codebase.

Agents and humans interact with the system through APIs and subscriptions. Filesystem views are generated as derived projections for compatibility with editors, compilers, and existing developer tools.

## User Problems to Solve
### For AI agents
- Safely editing multiple files at once
- Avoiding conflicts with other agents
- Receiving immediate feedback on invalid changes
- Subscribing to codebase changes that affect current tasks

### For human reviewers
- Reviewing semantic changes instead of raw line diffs
- Prioritizing risky or high-impact changes
- Reducing manual review load on low-risk edits
- Understanding change history and agent intent