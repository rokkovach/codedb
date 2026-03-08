package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"
)

type EventType string

const (
	EventTypeFileCreate       EventType = "file_create"
	EventTypeFileUpdate       EventType = "file_update"
	EventTypeFileDelete       EventType = "file_delete"
	EventTypeCommit           EventType = "commit"
	EventTypeWorkspaceCreate  EventType = "workspace_create"
	EventTypeWorkspaceMerge   EventType = "workspace_merge"
	EventTypeWorkspaceAbandon EventType = "workspace_abandon"
	EventTypeLeaseAcquire     EventType = "lease_acquire"
	EventTypeLeaseRelease     EventType = "lease_release"
	EventTypeLockAcquire      EventType = "lock_acquire"
	EventTypeLockRelease      EventType = "lock_release"
	EventTypeValidationPass   EventType = "validation_pass"
	EventTypeValidationFail   EventType = "validation_fail"
)

type Subscription struct {
	ID           string      `json:"id"`
	SubscriberID string      `json:"subscriber_id"`
	RepoID       *string     `json:"repo_id"`
	WorkspaceID  *string     `json:"workspace_id"`
	EventTypes   []EventType `json:"event_types"`
	PathPatterns []string    `json:"path_patterns"`
	IsActive     bool        `json:"is_active"`
	CreatedAt    time.Time   `json:"created_at"`
}

type EventLog struct {
	ID          string                 `json:"id"`
	EventType   EventType              `json:"event_type"`
	RepoID      *string                `json:"repo_id"`
	WorkspaceID *string                `json:"workspace_id"`
	EntityType  *string                `json:"entity_type"`
	EntityID    *string                `json:"entity_id"`
	Payload     map[string]interface{} `json:"payload"`
	ActorID     string                 `json:"actor_id"`
	ActorType   string                 `json:"actor_type"`
	CreatedAt   time.Time              `json:"created_at"`
}

type SubscriptionQueries struct {
	db *DB
}

func NewSubscriptionQueries(db *DB) *SubscriptionQueries {
	return &SubscriptionQueries{db: db}
}

func (q *SubscriptionQueries) Create(ctx context.Context, subscriberID string, repoID, workspaceID *string, eventTypes []EventType, pathPatterns []string) (*Subscription, error) {
	var sub Subscription
	var repoIDVal, workspaceIDVal interface{}
	if repoID != nil {
		repoIDVal = *repoID
	}
	if workspaceID != nil {
		workspaceIDVal = *workspaceID
	}

	err := q.db.Pool().QueryRow(ctx, `
		INSERT INTO subscriptions (subscriber_id, repo_id, workspace_id, event_types, path_patterns)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, subscriber_id, repo_id, workspace_id, event_types, path_patterns, is_active, created_at
	`, subscriberID, repoIDVal, workspaceIDVal, eventTypes, pathPatterns).Scan(
		&sub.ID, &sub.SubscriberID, &sub.RepoID, &sub.WorkspaceID, &sub.EventTypes, &sub.PathPatterns,
		&sub.IsActive, &sub.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &sub, nil
}

func (q *SubscriptionQueries) Get(ctx context.Context, id string) (*Subscription, error) {
	var sub Subscription
	var repoID, workspaceID sql.NullString
	err := q.db.Pool().QueryRow(ctx, `
		SELECT id, subscriber_id, repo_id, workspace_id, event_types, path_patterns, is_active, created_at
		FROM subscriptions WHERE id = $1
	`, id).Scan(
		&sub.ID, &sub.SubscriberID, &repoID, &workspaceID, &sub.EventTypes, &sub.PathPatterns,
		&sub.IsActive, &sub.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	if repoID.Valid {
		sub.RepoID = &repoID.String
	}
	if workspaceID.Valid {
		sub.WorkspaceID = &workspaceID.String
	}
	return &sub, nil
}

func (q *SubscriptionQueries) ListBySubscriber(ctx context.Context, subscriberID string) ([]Subscription, error) {
	rows, err := q.db.Pool().Query(ctx, `
		SELECT id, subscriber_id, repo_id, workspace_id, event_types, path_patterns, is_active, created_at
		FROM subscriptions WHERE subscriber_id = $1 AND is_active = true
		ORDER BY created_at DESC
	`, subscriberID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subs []Subscription
	for rows.Next() {
		var sub Subscription
		var repoID, workspaceID sql.NullString
		if err := rows.Scan(&sub.ID, &sub.SubscriberID, &repoID, &workspaceID, &sub.EventTypes,
			&sub.PathPatterns, &sub.IsActive, &sub.CreatedAt); err != nil {
			return nil, err
		}
		if repoID.Valid {
			sub.RepoID = &repoID.String
		}
		if workspaceID.Valid {
			sub.WorkspaceID = &workspaceID.String
		}
		subs = append(subs, sub)
	}
	return subs, rows.Err()
}

func (q *SubscriptionQueries) Deactivate(ctx context.Context, id string) error {
	_, err := q.db.Pool().Exec(ctx, `UPDATE subscriptions SET is_active = false WHERE id = $1`, id)
	return err
}

func (q *SubscriptionQueries) FindMatching(ctx context.Context, eventType EventType, repoID, workspaceID *string, path string) ([]Subscription, error) {
	query := `
		SELECT id, subscriber_id, repo_id, workspace_id, event_types, path_patterns, is_active, created_at
		FROM subscriptions
		WHERE is_active = true
			AND $1 = ANY(event_types)
			AND (repo_id IS NULL OR repo_id = $2)
			AND (workspace_id IS NULL OR workspace_id = $3)
	`
	var repoIDVal, workspaceIDVal interface{}
	if repoID != nil {
		repoIDVal = *repoID
	}
	if workspaceID != nil {
		workspaceIDVal = *workspaceID
	}

	rows, err := q.db.Pool().Query(ctx, query, eventType, repoIDVal, workspaceIDVal)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subs []Subscription
	for rows.Next() {
		var sub Subscription
		var repoIDNull, workspaceIDNull sql.NullString
		if err := rows.Scan(&sub.ID, &sub.SubscriberID, &repoIDNull, &workspaceIDNull, &sub.EventTypes,
			&sub.PathPatterns, &sub.IsActive, &sub.CreatedAt); err != nil {
			return nil, err
		}
		if repoIDNull.Valid {
			sub.RepoID = &repoIDNull.String
		}
		if workspaceIDNull.Valid {
			sub.WorkspaceID = &workspaceIDNull.String
		}
		subs = append(subs, sub)
	}
	return subs, rows.Err()
}

type EventLogQueries struct {
	db *DB
}

func NewEventLogQueries(db *DB) *EventLogQueries {
	return &EventLogQueries{db: db}
}

func (q *EventLogQueries) Create(ctx context.Context, eventType EventType, repoID, workspaceID *string, entityType, entityID *string, payload map[string]interface{}, actorID, actorType string) (*EventLog, error) {
	var event EventLog
	var repoIDVal, workspaceIDVal, entityTypeVal, entityIDVal interface{}
	if repoID != nil {
		repoIDVal = *repoID
	}
	if workspaceID != nil {
		workspaceIDVal = *workspaceID
	}
	if entityType != nil {
		entityTypeVal = *entityType
	}
	if entityID != nil {
		entityIDVal = *entityID
	}

	err := q.db.Pool().QueryRow(ctx, `
		INSERT INTO event_log (event_type, repo_id, workspace_id, entity_type, entity_id, payload, actor_id, actor_type)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, event_type, repo_id, workspace_id, entity_type, entity_id, payload, actor_id, actor_type, created_at
	`, eventType, repoIDVal, workspaceIDVal, entityTypeVal, entityIDVal, payload, actorID, actorType).Scan(
		&event.ID, &event.EventType, &event.RepoID, &event.WorkspaceID, &event.EntityType, &event.EntityID,
		&event.Payload, &event.ActorID, &event.ActorType, &event.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &event, nil
}

func (q *EventLogQueries) ListByRepo(ctx context.Context, repoID string, limit int) ([]EventLog, error) {
	rows, err := q.db.Pool().Query(ctx, `
		SELECT id, event_type, repo_id, workspace_id, entity_type, entity_id, payload, actor_id, actor_type, created_at
		FROM event_log WHERE repo_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`, repoID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []EventLog
	for rows.Next() {
		var e EventLog
		if err := rows.Scan(&e.ID, &e.EventType, &e.RepoID, &e.WorkspaceID, &e.EntityType, &e.EntityID,
			&e.Payload, &e.ActorID, &e.ActorType, &e.CreatedAt); err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, rows.Err()
}

func (q *EventLogQueries) ListByWorkspace(ctx context.Context, workspaceID string, limit int) ([]EventLog, error) {
	rows, err := q.db.Pool().Query(ctx, `
		SELECT id, event_type, repo_id, workspace_id, entity_type, entity_id, payload, actor_id, actor_type, created_at
		FROM event_log WHERE workspace_id = $1
		ORDER BY created_at DESC
		LIMIT $2
	`, workspaceID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []EventLog
	for rows.Next() {
		var e EventLog
		if err := rows.Scan(&e.ID, &e.EventType, &e.RepoID, &e.WorkspaceID, &e.EntityType, &e.EntityID,
			&e.Payload, &e.ActorID, &e.ActorType, &e.CreatedAt); err != nil {
			return nil, err
		}
		events = append(events, e)
	}
	return events, rows.Err()
}

type Notification struct {
	EventType EventType              `json:"event_type"`
	Payload   map[string]interface{} `json:"payload"`
}

func (q *EventLogQueries) Listen(ctx context.Context) (<-chan Notification, error) {
	conn, err := q.db.Pool().Acquire(ctx)
	if err != nil {
		return nil, err
	}

	_, err = conn.Exec(ctx, `LISTEN codedb_events`)
	if err != nil {
		conn.Release()
		return nil, err
	}

	ch := make(chan Notification, 100)

	go func() {
		defer close(ch)
		defer conn.Release()

		for {
			select {
			case <-ctx.Done():
				return
			default:
				notification, err := conn.Conn().WaitForNotification(ctx)
				if err != nil {
					if ctx.Err() != nil {
						return
					}
					continue
				}

				var n Notification
				if err := json.Unmarshal([]byte(notification.Payload), &n); err != nil {
					continue
				}

				select {
				case ch <- n:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return ch, nil
}
