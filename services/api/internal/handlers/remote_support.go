package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"monitopro/services/api/internal/domain"
	"monitopro/services/api/internal/events"
	"monitopro/services/api/internal/middleware"
	"monitopro/services/api/internal/storage"
)

type RemoteSupportHandler struct {
	store     storage.Storage
	publisher events.EventPublisher

	upgrader  websocket.Upgrader
	hubMu     sync.RWMutex
	hubRooms  map[uuid.UUID]map[*websocket.Conn]bool
}

func NewRemoteSupportHandler(store storage.Storage, publisher events.EventPublisher) *RemoteSupportHandler {
	return &RemoteSupportHandler{
		store:     store,
		publisher: publisher,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
		hubRooms: make(map[uuid.UUID]map[*websocket.Conn]bool),
	}
}

type RequestSessionBody struct {
	DeviceID             uuid.UUID `json:"device_id"`
	PermissionsRequested []string  `json:"permissions_requested"`
	Reason               string    `json:"reason"`
}

func (h *RemoteSupportHandler) RequestSession(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.TenantIDFromContext(r.Context())
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	var req RequestSessionBody
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body", "BAD_REQUEST")
		return
	}

	if req.DeviceID == uuid.Nil || len(req.PermissionsRequested) == 0 || req.Reason == "" {
		writeError(w, http.StatusBadRequest, "device_id, permissions_requested, and reason are required", "MISSING_FIELDS")
		return
	}

	// Validate permissions enum
	for _, p := range req.PermissionsRequested {
		if p != "screen_view" && p != "remote_control" {
			writeError(w, http.StatusBadRequest, "Invalid permission requested; must be screen_view or remote_control", "INVALID_PERMISSION")
			return
		}
	}

	// Check device exists
	device, err := h.store.Devices().GetByID(r.Context(), tenantID, req.DeviceID)
	if err != nil {
		writeError(w, http.StatusNotFound, "Device not found", "NOT_FOUND")
		return
	}

	sessionID := uuid.New()
	tokenBytes := make([]byte, 16)
	_, _ = rand.Read(tokenBytes)
	sessionToken := hex.EncodeToString(tokenBytes)

	now := time.Now().UTC()
	expiresAt := now.Add(1 * time.Hour) // Session max 60 minutes per protocol

	session := &domain.RemoteSession{
		SessionID:      sessionID,
		OrganizationID: tenantID,
		DeviceID:       device.ID,
		OperatorID:     user.UserID,
		OperatorName:   user.Email,
		Status:         domain.RemoteSessionAwaitingConsent,
		Permissions:    req.PermissionsRequested,
		Reason:         req.Reason,
		SignalingURL:   "/ws/remote-support/" + sessionID.String(),
		SessionToken:   sessionToken,
		ExpiresAt:      expiresAt,
		CreatedAt:      now,
	}

	if err := h.store.RemoteSessions().Create(r.Context(), session); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to create remote support session", "INTERNAL_ERROR")
		return
	}

	// Publish consent event
	_ = h.publisher.PublishRemoteConsent(r.Context(), tenantID, map[string]any{
		"session_id":          sessionID,
		"device_id":           device.ID,
		"operator_id":         user.UserID,
		"consent_decision":    "awaiting_consent",
		"granted_permissions": req.PermissionsRequested,
		"timestamp":           now.Format(time.RFC3339),
	})

	// Audit log
	_ = h.store.Audit().Insert(r.Context(), &domain.AuditLog{
		OrganizationID: tenantID,
		ActorType:      "user",
		ActorID:        &user.UserID,
		Action:         "remote_session_requested",
		ResourceType:   "remote_session",
		ResourceID:     &sessionID,
		IPAddress:      r.RemoteAddr,
	})

	writeJSON(w, http.StatusCreated, session)
}

func (h *RemoteSupportHandler) GetSession(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.TenantIDFromContext(r.Context())
	idStr := chi.URLParam(r, "id")
	sessionID, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid session ID", "BAD_REQUEST")
		return
	}

	session, err := h.store.RemoteSessions().GetByID(r.Context(), tenantID, sessionID)
	if err != nil {
		writeError(w, http.StatusNotFound, "Remote session not found", "NOT_FOUND")
		return
	}

	session.SignalingURL = "/ws/remote-support/" + session.SessionID.String()
	writeJSON(w, http.StatusOK, session)
}

type SignalMessage struct {
	Type    string          `json:"type"` // offer, answer, ice_candidate, consent_decision
	Payload json.RawMessage `json:"payload"`
}

func (h *RemoteSupportHandler) PostSignal(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.TenantIDFromContext(r.Context())
	idStr := chi.URLParam(r, "id")
	sessionID, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid session ID", "BAD_REQUEST")
		return
	}

	var sig SignalMessage
	if err := json.NewDecoder(r.Body).Decode(&sig); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid signal body", "BAD_REQUEST")
		return
	}

	if sig.Type == "consent_decision" {
		var decision struct {
			Approved bool   `json:"approved"`
			Reason   string `json:"reason"`
		}
		_ = json.Unmarshal(sig.Payload, &decision)
		newStatus := domain.RemoteSessionRejected
		if decision.Approved {
			newStatus = domain.RemoteSessionActive
		}
		_ = h.store.RemoteSessions().UpdateStatus(r.Context(), tenantID, sessionID, newStatus, decision.Reason)
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "signaled"})
}

func (h *RemoteSupportHandler) TerminateSession(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.TenantIDFromContext(r.Context())
	idStr := chi.URLParam(r, "id")
	sessionID, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid session ID", "BAD_REQUEST")
		return
	}

	_ = h.store.RemoteSessions().Terminate(r.Context(), tenantID, sessionID, "operator_manual_termination")

	// Notify room connections via websocket if open
	h.hubMu.Lock()
	if conns, ok := h.hubRooms[sessionID]; ok {
		for c := range conns {
			_ = c.WriteJSON(map[string]string{
				"type":   "session_terminated",
				"reason": "operator_manual_termination",
			})
			_ = c.Close()
		}
		delete(h.hubRooms, sessionID)
	}
	h.hubMu.Unlock()

	// Audit log
	user := middleware.UserFromContext(r.Context())
	var actorID *uuid.UUID
	if user != nil {
		actorID = &user.UserID
	}
	_ = h.store.Audit().Insert(r.Context(), &domain.AuditLog{
		OrganizationID: tenantID,
		ActorType:      "user",
		ActorID:        actorID,
		Action:         "remote_session_terminated",
		ResourceType:   "remote_session",
		ResourceID:     &sessionID,
		IPAddress:      r.RemoteAddr,
	})

	w.WriteHeader(http.StatusNoContent)
}

func (h *RemoteSupportHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	sessionID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Invalid session ID", http.StatusBadRequest)
		return
	}

	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	h.hubMu.Lock()
	if _, ok := h.hubRooms[sessionID]; !ok {
		h.hubRooms[sessionID] = make(map[*websocket.Conn]bool)
	}
	h.hubRooms[sessionID][conn] = true
	h.hubMu.Unlock()

	defer func() {
		h.hubMu.Lock()
		if conns, ok := h.hubRooms[sessionID]; ok {
			delete(conns, conn)
			if len(conns) == 0 {
				delete(h.hubRooms, sessionID)
			}
		}
		h.hubMu.Unlock()
	}()

	for {
		messageType, msg, err := conn.ReadMessage()
		if err != nil {
			break
		}

		// Broadcast message to other peers in session room
		h.hubMu.RLock()
		peers := make([]*websocket.Conn, 0)
		for peer := range h.hubRooms[sessionID] {
			if peer != conn {
				peers = append(peers, peer)
			}
		}
		h.hubMu.RUnlock()

		for _, peer := range peers {
			_ = peer.WriteMessage(messageType, msg)
		}
	}
}
