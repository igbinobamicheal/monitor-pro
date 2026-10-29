package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"monitopro/services/api/internal/ai"
	"monitopro/services/api/internal/domain"
	"monitopro/services/api/internal/middleware"
	"monitopro/services/api/internal/storage"
)

type AIHandler struct {
	store    storage.Storage
	provider ai.AIProvider
}

func NewAIHandler(store storage.Storage, provider ai.AIProvider) *AIHandler {
	if provider == nil {
		provider = ai.NewMockAdvisoryProvider()
	}
	return &AIHandler{
		store:    store,
		provider: provider,
	}
}

func (h *AIHandler) AnalyzeIncident(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.TenantIDFromContext(r.Context())
	tenant := middleware.TenantFromContext(r.Context())
	if tenant != nil && !tenant.AIEnabled {
		writeError(w, http.StatusForbidden, "AI advisory is disabled for this organization", "AI_DISABLED")
		return
	}

	idStr := chi.URLParam(r, "id")
	incidentID, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid incident ID", "BAD_REQUEST")
		return
	}

	incident, err := h.store.Incidents().GetByID(r.Context(), tenantID, incidentID)
	if err != nil {
		writeError(w, http.StatusNotFound, "Incident not found", "NOT_FOUND")
		return
	}

	alerts, _ := h.store.Incidents().GetAlerts(r.Context(), incidentID)

	// Identify primary device from first alert or fallback default device
	var device *domain.Device
	var diags *domain.DeviceDiagnostics

	if len(alerts) > 0 {
		device, _ = h.store.Devices().GetByID(r.Context(), tenantID, alerts[0].DeviceID)
	}

	if device == nil {
		device = &domain.Device{
			ID:             uuid.New(),
			OrganizationID: tenantID,
			Hostname:       "endpoint-workstation-01",
			IPAddress:      "10.0.4.15",
			Status:         domain.DeviceStatusOnline,
			OSVersion:      "Windows 11 Pro 23H2",
		}
	} else {
		diags, _ = h.store.Devices().GetDiagnostics(r.Context(), tenantID, device.ID)
	}

	// Apply PII redaction
	sanitizedIncident := *incident
	sanitizedIncident.Title = ai.RedactPII(incident.Title)
	sanitizedDevice := *device
	sanitizedDevice.Hostname = ai.RedactPII(device.Hostname)
	sanitizedDevice.IPAddress = ai.RedactPII(device.IPAddress)
	sanitizedDevice.AssignedUserEmail = ai.RedactPII(device.AssignedUserEmail)

	sanitizedAlerts := make([]*domain.Alert, 0, len(alerts))
	for _, a := range alerts {
		cp := *a
		cp.Message = ai.RedactPII(a.Message)
		sanitizedAlerts = append(sanitizedAlerts, &cp)
	}

	resp, err := h.provider.GenerateAdvisory(r.Context(), &sanitizedIncident, sanitizedAlerts, &sanitizedDevice, diags)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to generate AI advisory: "+err.Error(), "ADVISORY_ERROR")
		return
	}

	// Persist advisory report in DB
	hypoBytes, _ := json.Marshal(resp.Hypotheses)
	recomBytes, _ := json.Marshal(resp.SuggestedRemediations)
	_ = h.store.Incidents().SaveAIReport(r.Context(), &domain.AIAdvisoryReport{
		ID:              uuid.New(),
		OrganizationID:  tenantID,
		IncidentID:      incidentID,
		Summary:         resp.Summary,
		Hypotheses:      hypoBytes,
		Recommendations: recomBytes,
		Provider:        h.provider.Name(),
	})

	// Add timeline event
	_ = h.store.Incidents().AddTimeline(r.Context(), &domain.IncidentTimeline{
		IncidentID: incidentID,
		EventType:  "ai_advisory",
		Content:    "AI Advisory generated root-cause hypotheses and recommended remediation actions",
		CreatedAt:  time.Now().UTC(),
	})

	writeJSON(w, http.StatusOK, resp)
}
