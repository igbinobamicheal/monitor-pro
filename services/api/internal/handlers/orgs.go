package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"monitopro/services/api/internal/domain"
	"monitopro/services/api/internal/middleware"
	"monitopro/services/api/internal/storage"
)

type OrgHandler struct {
	store storage.Storage
}

func NewOrgHandler(store storage.Storage) *OrgHandler {
	return &OrgHandler{store: store}
}

func (h *OrgHandler) GetCurrent(w http.ResponseWriter, r *http.Request) {
	tenant := middleware.TenantFromContext(r.Context())
	if tenant == nil {
		writeError(w, http.StatusNotFound, "Tenant not found", "NOT_FOUND")
		return
	}
	writeJSON(w, http.StatusOK, tenant)
}

func (h *OrgHandler) GetAISettings(w http.ResponseWriter, r *http.Request) {
	orgIDStr := chi.URLParam(r, "org_id")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid organization ID", "BAD_REQUEST")
		return
	}

	org, err := h.store.Tenants().GetByID(r.Context(), orgID)
	if err != nil {
		writeError(w, http.StatusNotFound, "Organization not found", "NOT_FOUND")
		return
	}

	settings := domain.AISettings{
		Enabled:   org.AIEnabled,
		RedactPII: true,
		Provider:  "mock_advisory",
	}

	writeJSON(w, http.StatusOK, settings)
}

func (h *OrgHandler) UpdateAISettings(w http.ResponseWriter, r *http.Request) {
	orgIDStr := chi.URLParam(r, "org_id")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid organization ID", "BAD_REQUEST")
		return
	}

	var req domain.AISettings
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body", "BAD_REQUEST")
		return
	}

	if err := h.store.Tenants().UpdateAISettings(r.Context(), orgID, req.Enabled); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to update AI settings", "INTERNAL_ERROR")
		return
	}

	req.RedactPII = true
	writeJSON(w, http.StatusOK, req)
}
