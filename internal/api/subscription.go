package api

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/gorilla/websocket"
	"github.com/rokkovach/codedb/internal/db"
)

type SubscriptionService struct {
	subQueries   *db.SubscriptionQueries
	eventQueries *db.EventLogQueries
	mu           sync.RWMutex
	clients      map[string][]*websocket.Conn
}

func NewSubscriptionService(database *db.DB) *SubscriptionService {
	return &SubscriptionService{
		subQueries:   db.NewSubscriptionQueries(database),
		eventQueries: db.NewEventLogQueries(database),
		clients:      make(map[string][]*websocket.Conn),
	}
}

func (s *SubscriptionService) CreateSubscription(ctx context.Context, subscriberID string, repoID, workspaceID *string, eventTypes []db.EventType, pathPatterns []string) (*db.Subscription, error) {
	return s.subQueries.Create(ctx, subscriberID, repoID, workspaceID, eventTypes, pathPatterns)
}

func (s *SubscriptionService) ListSubscriptions(ctx context.Context, subscriberID string) ([]db.Subscription, error) {
	return s.subQueries.ListBySubscriber(ctx, subscriberID)
}

func (s *SubscriptionService) DeleteSubscription(ctx context.Context, id string) error {
	return s.subQueries.Deactivate(ctx, id)
}

func (s *SubscriptionService) LogEvent(ctx context.Context, eventType db.EventType, repoID, workspaceID *string, entityType, entityID *string, payload map[string]interface{}, actorID, actorType string) (*db.EventLog, error) {
	return s.eventQueries.Create(ctx, eventType, repoID, workspaceID, entityType, entityID, payload, actorID, actorType)
}

func (s *SubscriptionService) Broadcast(ctx context.Context, notification db.Notification) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for subscriberID, conns := range s.clients {
		for _, conn := range conns {
			if err := conn.WriteJSON(notification); err != nil {
				conn.Close()
				s.removeClient(subscriberID, conn)
			}
		}
	}
}

func (s *SubscriptionService) addClient(subscriberID string, conn *websocket.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.clients[subscriberID] = append(s.clients[subscriberID], conn)
}

func (s *SubscriptionService) removeClient(subscriberID string, conn *websocket.Conn) {
	s.mu.Lock()
	defer s.mu.Unlock()

	conns := s.clients[subscriberID]
	for i, c := range conns {
		if c == conn {
			s.clients[subscriberID] = append(conns[:i], conns[i+1:]...)
			break
		}
	}

	if len(s.clients[subscriberID]) == 0 {
		delete(s.clients, subscriberID)
	}
}

func (s *SubscriptionService) StartListener(ctx context.Context) error {
	ch, err := s.eventQueries.Listen(ctx)
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case notification, ok := <-ch:
				if !ok {
					return
				}
				s.Broadcast(ctx, notification)
			}
		}
	}()

	return nil
}

func (a *API) listSubscriptions(w http.ResponseWriter, r *http.Request) {
	repoID := chi.URLParam(r, "repoID")

	subs, err := a.subscriptionSvc.ListSubscriptions(r.Context(), repoID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to list subscriptions", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, subs)
}

type createSubscriptionRequest struct {
	SubscriberID string         `json:"subscriber_id"`
	WorkspaceID  *string        `json:"workspace_id"`
	EventTypes   []db.EventType `json:"event_types"`
	PathPatterns []string       `json:"path_patterns"`
}

func (a *API) createSubscription(w http.ResponseWriter, r *http.Request) {
	repoID := chi.URLParam(r, "repoID")

	var req createSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body", err.Error())
		return
	}

	if req.SubscriberID == "" {
		writeError(w, http.StatusBadRequest, "subscriber_id is required", "")
		return
	}

	if len(req.EventTypes) == 0 {
		writeError(w, http.StatusBadRequest, "event_types is required", "")
		return
	}

	sub, err := a.subscriptionSvc.CreateSubscription(r.Context(), req.SubscriberID, &repoID, req.WorkspaceID, req.EventTypes, req.PathPatterns)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "failed to create subscription", err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, sub)
}

func (a *API) deleteSubscription(w http.ResponseWriter, r *http.Request) {
	subID := chi.URLParam(r, "subID")

	if err := a.subscriptionSvc.DeleteSubscription(r.Context(), subID); err != nil {
		writeError(w, http.StatusInternalServerError, "failed to delete subscription", err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type wsMessage struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

func (a *API) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	subscriberID := r.URL.Query().Get("subscriber_id")
	if subscriberID == "" {
		writeError(w, http.StatusBadRequest, "subscriber_id is required", "")
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	a.subscriptionSvc.addClient(subscriberID, conn)

	defer func() {
		conn.Close()
		a.subscriptionSvc.removeClient(subscriberID, conn)
	}()

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}
