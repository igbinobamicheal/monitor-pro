package handlers

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"monitopro/services/api/internal/domain"
	"monitopro/services/api/internal/events"
	"monitopro/services/api/internal/middleware"
	"monitopro/services/api/internal/storage"
)

type DeviceHandler struct {
	store     storage.Storage
	publisher events.EventPublisher
	jwtSecret []byte
}

func NewDeviceHandler(store storage.Storage, publisher events.EventPublisher, jwtSecret string) *DeviceHandler {
	return &DeviceHandler{
		store:     store,
		publisher: publisher,
		jwtSecret: []byte(jwtSecret),
	}
}

func (h *DeviceHandler) List(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.TenantIDFromContext(r.Context())
	if tenantID == uuid.Nil {
		writeError(w, http.StatusBadRequest, "Organization context required", "TENANT_REQUIRED")
		return
	}

	status := r.URL.Query().Get("status")
	search := r.URL.Query().Get("search")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 50
	}

	devices, total, err := h.store.Devices().List(r.Context(), tenantID, status, search, page, limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to query devices", "INTERNAL_ERROR")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"items": devices,
		"total": total,
		"page":  page,
	})
}

func (h *DeviceHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.TenantIDFromContext(r.Context())
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid device ID", "BAD_REQUEST")
		return
	}

	device, err := h.store.Devices().GetByID(r.Context(), tenantID, id)
	if err != nil {
		writeError(w, http.StatusNotFound, "Device not found", "NOT_FOUND")
		return
	}

	latestMetric, _ := h.store.Metrics().GetLatestByDevice(r.Context(), tenantID, id)
	details := domain.DeviceDetails{
		Device: *device,
	}

	if latestMetric != nil {
		details.UptimeSeconds = latestMetric.Uptime
		details.LatestMetrics = &domain.TelemetryMetrics{
			CPUUsage:     latestMetric.CPUUsage,
			RAMUsage:     latestMetric.RAMUsage,
			DiskUsage:    latestMetric.DiskUsage,
			NetworkInKB:  latestMetric.NetworkIn,
			NetworkOutKB: latestMetric.NetworkOut,
		}
	}

	writeJSON(w, http.StatusOK, details)
}

func (h *DeviceHandler) GetDiagnostics(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.TenantIDFromContext(r.Context())
	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid device ID", "BAD_REQUEST")
		return
	}

	diag, err := h.store.Devices().GetDiagnostics(r.Context(), tenantID, id)
	if err != nil {
		// Return empty structured object if not yet populated
		writeJSON(w, http.StatusOK, domain.DeviceDiagnostics{
			DeviceID:       id,
			OrganizationID: tenantID,
		})
		return
	}

	writeJSON(w, http.StatusOK, diag)
}

type CreateTokenRequest struct {
	ExpiresInHours int `json:"expires_in_hours"`
	MaxUses        int `json:"max_uses"`
}

func (h *DeviceHandler) CreateEnrollmentToken(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.TenantIDFromContext(r.Context())
	user := middleware.UserFromContext(r.Context())

	var req CreateTokenRequest
	_ = json.NewDecoder(r.Body).Decode(&req)

	if req.ExpiresInHours <= 0 {
		req.ExpiresInHours = 24
	}
	if req.MaxUses <= 0 {
		req.MaxUses = 1
	}

	tokenBytes := make([]byte, 24)
	_, _ = rand.Read(tokenBytes)
	tokenStr := hex.EncodeToString(tokenBytes)

	expiresAt := time.Now().UTC().Add(time.Duration(req.ExpiresInHours) * time.Hour)
	token := &domain.EnrollmentToken{
		ID:             uuid.New(),
		OrganizationID: tenantID,
		Token:          tokenStr,
		ExpiresAt:      expiresAt,
		MaxUses:        req.MaxUses,
		UsesCount:      0,
	}
	if user != nil {
		token.CreatedBy = &user.UserID
	}

	if err := h.store.Devices().CreateEnrollmentToken(r.Context(), token); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to create enrollment token", "INTERNAL_ERROR")
		return
	}

	tenant := middleware.TenantFromContext(r.Context())
	subdomain := "app"
	if tenant != nil {
		subdomain = tenant.Subdomain
	}
	cmd := fmt.Sprintf("proactive-agent.exe --enroll --token %s --server https://%s.monitopro.com", tokenStr, subdomain)

	writeJSON(w, http.StatusCreated, map[string]any{
		"token":             tokenStr,
		"expires_at":        expiresAt.Format(time.RFC3339),
		"installer_command": cmd,
	})
}

type AgentEnrollRequest struct {
	EnrollmentToken string `json:"enrollment_token"`
	MachineGUID     string `json:"machine_guid"`
	Hostname        string `json:"hostname"`
	OSVersion       string `json:"os_version"`
	PublicKey       string `json:"public_key"`
}

func (h *DeviceHandler) AgentEnroll(w http.ResponseWriter, r *http.Request) {
	var req AgentEnrollRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request payload", "BAD_REQUEST")
		return
	}

	if req.EnrollmentToken == "" || req.MachineGUID == "" || req.Hostname == "" || req.PublicKey == "" {
		writeError(w, http.StatusBadRequest, "Missing required enrollment fields", "MISSING_FIELDS")
		return
	}

	token, err := h.store.Devices().GetEnrollmentToken(r.Context(), req.EnrollmentToken)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Invalid enrollment token", "INVALID_TOKEN")
		return
	}

	now := time.Now().UTC()
	if now.After(token.ExpiresAt) {
		writeError(w, http.StatusUnauthorized, "Enrollment token has expired", "TOKEN_EXPIRED")
		return
	}

	if token.UsesCount >= token.MaxUses {
		writeError(w, http.StatusUnauthorized, "Enrollment token usage limit reached", "TOKEN_EXHAUSTED")
		return
	}

	_ = h.store.Devices().UseEnrollmentToken(r.Context(), token.Token)

	// Check if device already exists with this MachineGUID
	var device *domain.Device
	existing, err := h.store.Devices().GetByMachineGUID(r.Context(), req.MachineGUID)
	if err == nil && existing != nil {
		device = existing
		device.Hostname = req.Hostname
		device.OSVersion = req.OSVersion
		device.PublicKey = req.PublicKey
		device.Status = domain.DeviceStatusOnline
		device.LastSeenAt = &now
		_ = h.store.Devices().Update(r.Context(), device)
	} else {
		device = &domain.Device{
			ID:             uuid.New(),
			OrganizationID: token.OrganizationID,
			Hostname:       req.Hostname,
			IPAddress:      r.RemoteAddr,
			Status:         domain.DeviceStatusOnline,
			OSVersion:      req.OSVersion,
			MachineGUID:    req.MachineGUID,
			PublicKey:      req.PublicKey,
			LastSeenAt:     &now,
		}
		if err := h.store.Devices().Create(r.Context(), device); err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to register endpoint device", "INTERNAL_ERROR")
			return
		}
	}

	// Generate certificate/signature
	certData := fmt.Sprintf("%s:%s:%d", device.ID.String(), token.OrganizationID.String(), now.Unix())
	signedCert := base64.StdEncoding.EncodeToString([]byte(certData))

	writeJSON(w, http.StatusCreated, map[string]any{
		"device_id":              device.ID,
		"organization_id":        token.OrganizationID,
		"signed_certificate":     signedCert,
		"telemetry_interval_sec": 30,
	})
}

func (h *DeviceHandler) AgentTelemetry(w http.ResponseWriter, r *http.Request) {
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Failed to read request body", "BAD_REQUEST")
		return
	}

	var report domain.TelemetryReport
	if err := json.Unmarshal(bodyBytes, &report); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid telemetry JSON", "BAD_REQUEST")
		return
	}

	if report.DeviceID == uuid.Nil {
		writeError(w, http.StatusBadRequest, "device_id is required", "BAD_REQUEST")
		return
	}

	// Find device
	tenantID := middleware.TenantIDFromContext(r.Context())
	device, err := h.store.Devices().GetByID(r.Context(), tenantID, report.DeviceID)
	if err != nil {
		// Try lookup by bypass if tenant not in context
		device, err = h.store.Devices().GetByID(r.Context(), uuid.Nil, report.DeviceID)
		if err != nil {
			writeError(w, http.StatusNotFound, "Device not found", "NOT_FOUND")
			return
		}
	}

	// Verify cryptographic signature if public key is present and signature header is passed
	sigHeader := r.Header.Get(middleware.AgentSignatureHdr)
	if sigHeader != "" && device.PublicKey != "" {
		pubKeyBytes, err := base64.StdEncoding.DecodeString(device.PublicKey)
		if err == nil && len(pubKeyBytes) == ed25519.PublicKeySize {
			sigBytes, err := base64.StdEncoding.DecodeString(sigHeader)
			if err == nil {
				if !ed25519.Verify(pubKeyBytes, bodyBytes, sigBytes) {
					writeError(w, http.StatusUnauthorized, "Invalid agent cryptographic signature", "INVALID_SIGNATURE")
					return
				}
			}
		}
	}

	now := time.Now().UTC()
	if report.Timestamp.IsZero() {
		report.Timestamp = now
	}

	metric := &domain.Metric{
		OrganizationID: device.OrganizationID,
		DeviceID:       device.ID,
		CPUUsage:       report.Metrics.CPUUsage,
		RAMUsage:       report.Metrics.RAMUsage,
		DiskUsage:      report.Metrics.DiskUsage,
		NetworkIn:      report.Metrics.NetworkInKB,
		NetworkOut:     report.Metrics.NetworkOutKB,
		Timestamp:      report.Timestamp,
	}

	_ = h.store.Metrics().Insert(r.Context(), metric)
	_ = h.store.Devices().UpdateStatusAndLastSeen(r.Context(), device.ID, domain.DeviceStatusOnline, now)

	// If diagnostics delta is present, store it
	if report.DiagnosticsDelta != nil && len(*report.DiagnosticsDelta) > 0 {
		var raw map[string]json.RawMessage
		if err := json.Unmarshal(*report.DiagnosticsDelta, &raw); err == nil {
			diag := &domain.DeviceDiagnostics{
				DeviceID:          device.ID,
				OrganizationID:    device.OrganizationID,
				SystemInfo:        raw["system"],
				CPUInfo:           raw["cpu_cores"],
				MemoryInfo:        raw["memory"],
				StorageInfo:       raw["storage"],
				NetworkInfo:       raw["network"],
				SecurityInfo:      raw["security"],
				ProcessesInfo:     raw["processes"],
				ServicesInfo:      raw["services"],
				EventLogs:         raw["event_logs"],
				InstalledSoftware: raw["software"],
			}
			_ = h.store.Devices().UpsertDiagnostics(r.Context(), diag)
		}
	}

	// Publish to NATS
	_ = h.publisher.PublishTelemetry(r.Context(), device.OrganizationID, device.ID, map[string]any{
		"organization_id": device.OrganizationID,
		"device_id":       device.ID,
		"timestamp":       report.Timestamp.Format(time.RFC3339),
		"cpu_usage":       report.Metrics.CPUUsage,
		"ram_usage":       report.Metrics.RAMUsage,
		"disk_usage":      report.Metrics.DiskUsage,
		"network_in":      report.Metrics.NetworkInKB,
		"network_out":     report.Metrics.NetworkOutKB,
	})

	// Check thresholds for auto-alerting
	if report.Metrics.CPUUsage >= 90.0 {
		alert := &domain.Alert{
			ID:             uuid.New(),
			OrganizationID: device.OrganizationID,
			DeviceID:       device.ID,
			RuleName:       "High CPU Utilization",
			Message:        fmt.Sprintf("CPU utilization exceeded 90%% (Current: %.1f%%)", report.Metrics.CPUUsage),
			Severity:       "critical",
			Resolved:       false,
			CreatedAt:      now,
		}
		_ = h.store.Alerts().Create(r.Context(), alert)
		_ = h.publisher.PublishAlertTriggered(r.Context(), device.OrganizationID, alert)
	}

	if report.Metrics.RAMUsage >= 90.0 {
		alert := &domain.Alert{
			ID:             uuid.New(),
			OrganizationID: device.OrganizationID,
			DeviceID:       device.ID,
			RuleName:       "High Memory Usage",
			Message:        fmt.Sprintf("Memory usage exceeded 90%% (Current: %.1f%%)", report.Metrics.RAMUsage),
			Severity:       "warning",
			Resolved:       false,
			CreatedAt:      now,
		}
		_ = h.store.Alerts().Create(r.Context(), alert)
		_ = h.publisher.PublishAlertTriggered(r.Context(), device.OrganizationID, alert)
	}

	if report.Metrics.DiskUsage >= 90.0 {
		alert := &domain.Alert{
			ID:             uuid.New(),
			OrganizationID: device.OrganizationID,
			DeviceID:       device.ID,
			RuleName:       "High Disk Utilization",
			Message:        fmt.Sprintf("Disk usage exceeded 90%% (Current: %.1f%%)", report.Metrics.DiskUsage),
			Severity:       "warning",
			Resolved:       false,
			CreatedAt:      now,
		}
		_ = h.store.Alerts().Create(r.Context(), alert)
		_ = h.publisher.PublishAlertTriggered(r.Context(), device.OrganizationID, alert)
	}

	w.WriteHeader(http.StatusAccepted)
}
