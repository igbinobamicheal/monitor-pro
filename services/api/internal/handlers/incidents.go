package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"monitopro/services/api/internal/domain"
	"monitopro/services/api/internal/middleware"
	"monitopro/services/api/internal/storage"
)

type IncidentHandler struct {
	store storage.Storage
}

func NewIncidentHandler(store storage.Storage) *IncidentHandler {
	return &IncidentHandler{store: store}
}

func (h *IncidentHandler) ListAlerts(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.TenantIDFromContext(r.Context())
	if tenantID == uuid.Nil {
		writeError(w, http.StatusBadRequest, "Organization context required", "TENANT_REQUIRED")
		return
	}

	activeOnly := false
	if activeStr := r.URL.Query().Get("active"); activeStr != "" {
		activeOnly, _ = strconv.ParseBool(activeStr)
	}

	alerts, err := h.store.Alerts().List(r.Context(), tenantID, activeOnly)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to query alerts", "INTERNAL_ERROR")
		return
	}
	if alerts == nil {
		alerts = []*domain.Alert{}
	}

	writeJSON(w, http.StatusOK, alerts)
}

func (h *IncidentHandler) ListIncidents(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.TenantIDFromContext(r.Context())
	if tenantID == uuid.Nil {
		writeError(w, http.StatusBadRequest, "Organization context required", "TENANT_REQUIRED")
		return
	}

	status := r.URL.Query().Get("status")
	incidents, err := h.store.Incidents().List(r.Context(), tenantID, status)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to query incidents", "INTERNAL_ERROR")
		return
	}
	if incidents == nil {
		incidents = []*domain.Incident{}
	}

	writeJSON(w, http.StatusOK, incidents)
}

func (h *IncidentHandler) GetIncident(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.TenantIDFromContext(r.Context())
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid incident ID", "BAD_REQUEST")
		return
	}

	incident, err := h.store.Incidents().GetByID(r.Context(), tenantID, id)
	if err != nil {
		writeError(w, http.StatusNotFound, "Incident not found", "NOT_FOUND")
		return
	}

	// Fetch linked alerts and timeline
	alerts, _ := h.store.Incidents().GetAlerts(r.Context(), id)
	incident.Alerts = alerts
	incident.AlertsCount = len(alerts)

	writeJSON(w, http.StatusOK, incident)
}

func (h *IncidentHandler) GetTimeline(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid incident ID", "BAD_REQUEST")
		return
	}

	timeline, err := h.store.Incidents().GetTimeline(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to fetch timeline", "INTERNAL_ERROR")
		return
	}
	if timeline == nil {
		timeline = []*domain.IncidentTimeline{}
	}

	writeJSON(w, http.StatusOK, timeline)
}

type UpdateStatusRequest struct {
	Status domain.IncidentStatus `json:"status"`
}

func (h *IncidentHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.TenantIDFromContext(r.Context())
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid incident ID", "BAD_REQUEST")
		return
	}

	var req UpdateStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body", "BAD_REQUEST")
		return
	}

	switch req.Status {
	case domain.IncidentStatusOpen, domain.IncidentStatusInvestigating, domain.IncidentStatusResolved, domain.IncidentStatusClosed:
	default:
		writeError(w, http.StatusBadRequest, "Invalid incident status value", "BAD_REQUEST")
		return
	}

	if err := h.store.Incidents().UpdateStatus(r.Context(), tenantID, id, req.Status); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to update incident status", "INTERNAL_ERROR")
		return
	}

	user := middleware.UserFromContext(r.Context())
	var authorID *uuid.UUID
	authorName := "System"
	if user != nil {
		authorID = &user.UserID
		authorName = user.Email
	}

	_ = h.store.Incidents().AddTimeline(r.Context(), &domain.IncidentTimeline{
		IncidentID:   id,
		AuthorUserID: authorID,
		AuthorName:   authorName,
		EventType:    "status_change",
		Content:      "Status updated to " + string(req.Status),
		CreatedAt:    time.Now().UTC(),
	})

	incident, _ := h.store.Incidents().GetByID(r.Context(), tenantID, id)
	writeJSON(w, http.StatusOK, incident)
}
