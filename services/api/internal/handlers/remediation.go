package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"monitopro/services/api/internal/domain"
	"monitopro/services/api/internal/events"
	"monitopro/services/api/internal/middleware"
	"monitopro/services/api/internal/remediation"
	"monitopro/services/api/internal/storage"
)

type RemediationHandler struct {
	store     storage.Storage
	catalogue *remediation.Catalogue
	publisher events.EventPublisher
}

func NewRemediationHandler(store storage.Storage, cat *remediation.Catalogue, pub events.EventPublisher) *RemediationHandler {
	return &RemediationHandler{
		store:     store,
		catalogue: cat,
		publisher: pub,
	}
}

func (h *RemediationHandler) ListActions(w http.ResponseWriter, r *http.Request) {
	actions := h.catalogue.ListActions()
	writeJSON(w, http.StatusOK, actions)
}

type ProposeJobRequest struct {
	DeviceID      uuid.UUID       `json:"device_id"`
	ActionID      string          `json:"action_id"`
	Parameters    json.RawMessage `json:"parameters"`
	Justification string          `json:"justification"`
}

func (h *RemediationHandler) ProposeJob(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.TenantIDFromContext(r.Context())
	user := middleware.UserFromContext(r.Context())

	var req ProposeJobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body", "BAD_REQUEST")
		return
	}

	if req.DeviceID == uuid.Nil || req.ActionID == "" {
		writeError(w, http.StatusBadRequest, "device_id and action_id are required", "MISSING_FIELDS")
		return
	}

	// Validate action exists and check parameters against schema
	validatedParams, err := h.catalogue.ValidateParameters(req.ActionID, req.Parameters)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Parameter validation error: "+err.Error(), "INVALID_PARAMETERS")
		return
	}

	// Verify device exists
	device, err := h.store.Devices().GetByID(r.Context(), tenantID, req.DeviceID)
	if err != nil {
		writeError(w, http.StatusNotFound, "Device not found", "NOT_FOUND")
		return
	}

	action, _ := h.catalogue.GetAction(req.ActionID)

	status := domain.RemediationPendingApproval
	var approvedBy *uuid.UUID
	now := time.Now().UTC()

	// If action does not require approval, approve immediately
	if !action.RequiresApproval {
		status = domain.RemediationApproved
		if user != nil {
			approvedBy = &user.UserID
		}
	}

	paramBytes, _ := json.Marshal(validatedParams)
	job := &domain.RemediationJob{
		ID:             uuid.New(),
		OrganizationID: tenantID,
		DeviceID:       device.ID,
		ActionID:       req.ActionID,
		Parameters:     paramBytes,
		Status:         status,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if user != nil {
		job.RequestedBy = &user.UserID
		job.Requester = user.Email
	}
	if approvedBy != nil {
		job.ApprovedBy = approvedBy
		if user != nil {
			job.Approver = user.Email
		}
	}

	if err := h.store.Remediations().CreateJob(r.Context(), job); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to create remediation job", "INTERNAL_ERROR")
		return
	}

	// If approved, dispatch event
	if status == domain.RemediationApproved {
		job.Status = domain.RemediationQueued
		_ = h.store.Remediations().UpdateJobStatus(r.Context(), tenantID, job.ID, domain.RemediationQueued, "", nil)
		_ = h.publisher.PublishRemediationDispatched(r.Context(), tenantID, device.ID, map[string]any{
			"job_id":          job.ID,
			"organization_id": tenantID,
			"device_id":       device.ID,
			"action_id":       job.ActionID,
			"parameters":      validatedParams,
			"expires_at":      now.Add(time.Duration(action.TimeoutSeconds) * time.Second).Format(time.RFC3339),
		})
	}

	// Record audit log
	_ = h.store.Audit().Insert(r.Context(), &domain.AuditLog{
		OrganizationID: tenantID,
		ActorType:      "user",
		ActorID:        job.RequestedBy,
		Action:         "remediation_job_proposed",
		ResourceType:   "remediation_job",
		ResourceID:     &job.ID,
		IPAddress:      r.RemoteAddr,
	})

	writeJSON(w, http.StatusCreated, job)
}

func (h *RemediationHandler) ApproveJob(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.TenantIDFromContext(r.Context())
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		writeError(w, http.StatusUnauthorized, "Authentication required", "UNAUTHORIZED")
		return
	}

	idStr := chi.URLParam(r, "id")
	jobID, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid job ID", "BAD_REQUEST")
		return
	}

	job, err := h.store.Remediations().GetJobByID(r.Context(), tenantID, jobID)
	if err != nil {
		writeError(w, http.StatusNotFound, "Remediation job not found", "NOT_FOUND")
		return
	}

	if job.Status != domain.RemediationPendingApproval {
		writeError(w, http.StatusBadRequest, "Job is not in pending_approval status", "INVALID_STATUS")
		return
	}

	if err := h.store.Remediations().ApproveJob(r.Context(), tenantID, jobID, user.UserID); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to approve job: "+err.Error(), "INTERNAL_ERROR")
		return
	}

	// Transition to queued and dispatch to NATS
	_ = h.store.Remediations().UpdateJobStatus(r.Context(), tenantID, jobID, domain.RemediationQueued, "", nil)

	var params map[string]any
	_ = json.Unmarshal(job.Parameters, &params)
	action, _ := h.catalogue.GetAction(job.ActionID)
	timeout := 60
	if action.TimeoutSeconds > 0 {
		timeout = action.TimeoutSeconds
	}

	_ = h.publisher.PublishRemediationDispatched(r.Context(), tenantID, job.DeviceID, map[string]any{
		"job_id":          job.ID,
		"organization_id": tenantID,
		"device_id":       job.DeviceID,
		"action_id":       job.ActionID,
		"parameters":      params,
		"expires_at":      time.Now().UTC().Add(time.Duration(timeout) * time.Second).Format(time.RFC3339),
	})

	// Record audit
	_ = h.store.Audit().Insert(r.Context(), &domain.AuditLog{
		OrganizationID: tenantID,
		ActorType:      "user",
		ActorID:        &user.UserID,
		Action:         "remediation_job_approved",
		ResourceType:   "remediation_job",
		ResourceID:     &job.ID,
		IPAddress:      r.RemoteAddr,
	})

	updatedJob, _ := h.store.Remediations().GetJobByID(r.Context(), tenantID, jobID)
	writeJSON(w, http.StatusOK, updatedJob)
}

func (h *RemediationHandler) GetJob(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.TenantIDFromContext(r.Context())
	idStr := chi.URLParam(r, "id")
	jobID, err := uuid.Parse(idStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid job ID", "BAD_REQUEST")
		return
	}

	job, err := h.store.Remediations().GetJobByID(r.Context(), tenantID, jobID)
	if err != nil {
		writeError(w, http.StatusNotFound, "Remediation job not found", "NOT_FOUND")
		return
	}

	writeJSON(w, http.StatusOK, job)
}

func (h *RemediationHandler) ListJobs(w http.ResponseWriter, r *http.Request) {
	tenantID := middleware.TenantIDFromContext(r.Context())
	jobs, err := h.store.Remediations().ListJobs(r.Context(), tenantID, 50)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to list remediation jobs", "INTERNAL_ERROR")
		return
	}
	if jobs == nil {
		jobs = []*domain.RemediationJob{}
	}
	writeJSON(w, http.StatusOK, jobs)
}
