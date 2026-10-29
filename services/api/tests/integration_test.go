package tests

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"monitopro/services/api/internal/ai"
	"monitopro/services/api/internal/auth"
	"monitopro/services/api/internal/domain"
	"monitopro/services/api/internal/events"
	"monitopro/services/api/internal/handlers"
	"monitopro/services/api/internal/middleware"
	"monitopro/services/api/internal/remediation"
	"monitopro/services/api/internal/storage"
)

type TestEnv struct {
	Router       http.Handler
	Store        storage.Storage
	TokenManager *auth.TokenManager
	Org          *domain.Organization
	AdminUser    *domain.User
	AdminToken   string
	OperatorUser *domain.User
	OperatorTok  string
	ViewerUser   *domain.User
	ViewerTok    string
}

func setupTestEnv(t *testing.T) *TestEnv {
	store := storage.NewMemoryStorage()
	secret := "dev_jwt_secret_key_minimum_32_characters_long"
	tm := auth.NewTokenManager(secret, 24*time.Hour)
	pub := events.NewNoopPublisher()
	cat := remediation.NewCatalogue()
	aiProv := ai.NewMockAdvisoryProvider()
	healthH := handlers.NewHealthHandler("api-test-replica", nil)

	router := handlers.NewRouter(&handlers.RouterConfig{
		Store:         store,
		TokenManager:  tm,
		Publisher:     pub,
		Catalogue:     cat,
		AIProvider:    aiProv,
		ReplicaID:     "api-test-replica",
		HealthHandler: healthH,
	})

	// Seed Organization
	orgID := uuid.New()
	org := &domain.Organization{
		ID:        orgID,
		Name:      "Demo Enterprise",
		Subdomain: "demo",
		AIEnabled: true,
	}
	if err := store.Tenants().Create(context.Background(), org); err != nil {
		t.Fatalf("Failed to create tenant: %v", err)
	}

	// Seed Admin User
	hash, _ := bcrypt.GenerateFromPassword([]byte("Password123!"), bcrypt.MinCost)
	adminUser := &domain.User{
		ID:           uuid.New(),
		Email:        "admin@demo.monitopro.com",
		PasswordHash: string(hash),
		FullName:     "Alice Admin",
		MFAEnabled:   false,
	}
	_ = store.Users().Create(context.Background(), adminUser)
	_ = store.Users().CreateMembership(context.Background(), &domain.Membership{
		OrganizationID: org.ID,
		UserID:         adminUser.ID,
		Role:           domain.RoleOwnerAdmin,
	})
	adminTok, _, _ := tm.GenerateToken(adminUser.ID, org.ID, domain.RoleOwnerAdmin, adminUser.Email)

	// Seed Operator User
	opUser := &domain.User{
		ID:           uuid.New(),
		Email:        "operator@demo.monitopro.com",
		PasswordHash: string(hash),
		FullName:     "Bob Operator",
	}
	_ = store.Users().Create(context.Background(), opUser)
	_ = store.Users().CreateMembership(context.Background(), &domain.Membership{
		OrganizationID: org.ID,
		UserID:         opUser.ID,
		Role:           domain.RoleITOperator,
	})
	opTok, _, _ := tm.GenerateToken(opUser.ID, org.ID, domain.RoleITOperator, opUser.Email)

	// Seed Viewer User
	viewerUser := &domain.User{
		ID:           uuid.New(),
		Email:        "viewer@demo.monitopro.com",
		PasswordHash: string(hash),
		FullName:     "Charlie Viewer",
	}
	_ = store.Users().Create(context.Background(), viewerUser)
	_ = store.Users().CreateMembership(context.Background(), &domain.Membership{
		OrganizationID: org.ID,
		UserID:         viewerUser.ID,
		Role:           domain.RoleReadonlyViewer,
	})
	viewerTok, _, _ := tm.GenerateToken(viewerUser.ID, org.ID, domain.RoleReadonlyViewer, viewerUser.Email)

	return &TestEnv{
		Router:       router,
		Store:        store,
		TokenManager: tm,
		Org:          org,
		AdminUser:    adminUser,
		AdminToken:   adminTok,
		OperatorUser: opUser,
		OperatorTok:  opTok,
		ViewerUser:   viewerUser,
		ViewerTok:    viewerTok,
	}
}

func TestHealthEndpoint(t *testing.T) {
	env := setupTestEnv(t)

	req := httptest.NewRequest("GET", "/api/health", nil)
	rec := httptest.NewRecorder()

	env.Router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK, got %d", rec.Code)
	}

	var resp map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if resp["status"] != "healthy" {
		t.Errorf("Expected status healthy, got %v", resp["status"])
	}
	if resp["replica_id"] != "api-test-replica" {
		t.Errorf("Expected replica_id api-test-replica, got %v", resp["replica_id"])
	}
}

func TestAuthFlow(t *testing.T) {
	env := setupTestEnv(t)

	// 1. Failed login with wrong password
	{
		body, _ := json.Marshal(map[string]string{
			"email":    env.AdminUser.Email,
			"password": "WrongPassword!",
		})
		req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(body))
		req.Header.Set("X-Tenant-ID", env.Org.ID.String())
		rec := httptest.NewRecorder()

		env.Router.ServeHTTP(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("Expected 401 Unauthorized for bad password, got %d", rec.Code)
		}
	}

	// 2. Successful login
	{
		body, _ := json.Marshal(map[string]string{
			"email":    env.AdminUser.Email,
			"password": "Password123!",
		})
		req := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(body))
		req.Header.Set("X-Tenant-ID", env.Org.ID.String())
		rec := httptest.NewRecorder()

		env.Router.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("Expected 200 OK for valid login, got %d", rec.Code)
		}

		var res handlers.LoginResponse
		_ = json.NewDecoder(rec.Body).Decode(&res)
		if res.Token == "" {
			t.Fatal("Expected non-empty token")
		}
		if res.User.Email != env.AdminUser.Email {
			t.Errorf("Expected email %s, got %s", env.AdminUser.Email, res.User.Email)
		}
	}

	// 3. MFA Setup and Activation
	var mfaSecret string
	{
		req := httptest.NewRequest("POST", "/api/v1/auth/mfa/setup", nil)
		req.Header.Set("Authorization", "Bearer "+env.AdminToken)
		rec := httptest.NewRecorder()

		env.Router.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("Expected 200 OK on MFA setup, got %d", rec.Code)
		}

		var mfaResp map[string]string
		_ = json.NewDecoder(rec.Body).Decode(&mfaResp)
		mfaSecret = mfaResp["secret"]
		if mfaSecret == "" || mfaResp["qr_code_url"] == "" {
			t.Fatal("Expected secret and qr_code_url in MFA setup response")
		}
	}

	// 4. MFA Verify
	{
		code, _ := auth.GenerateCode(mfaSecret, time.Now().UTC())
		body, _ := json.Marshal(map[string]string{"code": code})
		req := httptest.NewRequest("POST", "/api/v1/auth/mfa/verify", bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+env.AdminToken)
		rec := httptest.NewRecorder()

		env.Router.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("Expected 200 OK on MFA verify, got %d", rec.Code)
		}
	}

	// 5. Subsequent Login with MFA code required
	{
		// Attempt without MFA code -> 401
		body1, _ := json.Marshal(map[string]string{
			"email":    env.AdminUser.Email,
			"password": "Password123!",
		})
		req1 := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(body1))
		req1.Header.Set("X-Tenant-ID", env.Org.ID.String())
		rec1 := httptest.NewRecorder()
		env.Router.ServeHTTP(rec1, req1)
		if rec1.Code != http.StatusUnauthorized {
			t.Fatalf("Expected 401 when MFA code omitted, got %d", rec1.Code)
		}

		// Attempt with valid MFA code -> 200 OK
		code, _ := auth.GenerateCode(mfaSecret, time.Now().UTC())
		body2, _ := json.Marshal(map[string]string{
			"email":    env.AdminUser.Email,
			"password": "Password123!",
			"mfa_code": code,
		})
		req2 := httptest.NewRequest("POST", "/api/v1/auth/login", bytes.NewReader(body2))
		req2.Header.Set("X-Tenant-ID", env.Org.ID.String())
		rec2 := httptest.NewRecorder()
		env.Router.ServeHTTP(rec2, req2)
		if rec2.Code != http.StatusOK {
			t.Fatalf("Expected 200 OK with valid MFA code, got %d", rec2.Code)
		}
	}
}

func TestTenantIsolation(t *testing.T) {
	env := setupTestEnv(t)

	// Create second tenant
	org2 := &domain.Organization{
		ID:        uuid.New(),
		Name:      "Second Org",
		Subdomain: "second",
	}
	_ = env.Store.Tenants().Create(context.Background(), org2)

	// Device belonging to second tenant
	device2 := &domain.Device{
		ID:             uuid.New(),
		OrganizationID: org2.ID,
		Hostname:       "isolated-workstation",
		IPAddress:      "192.168.10.50",
		Status:         domain.DeviceStatusOnline,
	}
	_ = env.Store.Devices().Create(context.Background(), device2)

	// 1. Admin of Org 1 trying to access Org 2 device -> 404 Not Found (RLS isolation)
	{
		req := httptest.NewRequest("GET", "/api/v1/devices/"+device2.ID.String(), nil)
		req.Header.Set("Authorization", "Bearer "+env.AdminToken)
		req.Header.Set("X-Tenant-ID", env.Org.ID.String())
		rec := httptest.NewRecorder()

		env.Router.ServeHTTP(rec, req)
		if rec.Code != http.StatusNotFound {
			t.Fatalf("Expected 404 Device Not Found across tenant boundary, got %d", rec.Code)
		}
	}

	// 2. Org 1 user token presenting Org 2 as tenant header -> 403 Forbidden (TENANT_MISMATCH)
	{
		req := httptest.NewRequest("GET", "/api/v1/devices", nil)
		req.Header.Set("Authorization", "Bearer "+env.AdminToken)
		req.Header.Set("X-Tenant-ID", org2.ID.String())
		rec := httptest.NewRecorder()

		env.Router.ServeHTTP(rec, req)
		if rec.Code != http.StatusForbidden {
			t.Fatalf("Expected 403 Forbidden on token tenant mismatch, got %d", rec.Code)
		}
	}
}

func TestAgentEnrollmentAndTelemetryIngestion(t *testing.T) {
	env := setupTestEnv(t)

	// 1. Admin creates an enrollment token
	var tokenStr string
	{
		body, _ := json.Marshal(map[string]int{"expires_in_hours": 24, "max_uses": 2})
		req := httptest.NewRequest("POST", "/api/v1/devices/enrollment-tokens", bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+env.AdminToken)
		req.Header.Set("X-Tenant-ID", env.Org.ID.String())
		rec := httptest.NewRecorder()

		env.Router.ServeHTTP(rec, req)
		if rec.Code != http.StatusCreated {
			t.Fatalf("Expected 201 Created on enrollment token generation, got %d", rec.Code)
		}

		var res map[string]any
		_ = json.NewDecoder(rec.Body).Decode(&res)
		tokenStr = res["token"].(string)
	}

	// 2. Endpoint Agent generates Ed25519 keypair and enrolls
	pubKey, privKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate Ed25519 key: %v", err)
	}
	pubKeyBase64 := base64.StdEncoding.EncodeToString(pubKey)

	var deviceID uuid.UUID
	{
		enrollBody, _ := json.Marshal(map[string]string{
			"enrollment_token": tokenStr,
			"machine_guid":     "GUID-WIN11-TEST-001",
			"hostname":         "DESKTOP-PRO-01",
			"os_version":       "Windows 11 Enterprise",
			"public_key":       pubKeyBase64,
		})
		req := httptest.NewRequest("POST", "/api/v1/agent/enroll", bytes.NewReader(enrollBody))
		rec := httptest.NewRecorder()

		env.Router.ServeHTTP(rec, req)
		if rec.Code != http.StatusCreated {
			t.Fatalf("Expected 201 Created on agent enrollment, got %d", rec.Code)
		}

		var res map[string]any
		_ = json.NewDecoder(rec.Body).Decode(&res)
		deviceID = uuid.MustParse(res["device_id"].(string))
		if res["signed_certificate"] == "" {
			t.Fatal("Expected signed certificate")
		}
	}

	// 3. Agent signs and submits Telemetry report with High CPU (to trigger auto-alert)
	{
		telemetry := domain.TelemetryReport{
			DeviceID:  deviceID,
			Timestamp: time.Now().UTC(),
			Metrics: domain.TelemetryMetrics{
				CPUUsage:     95.5, // > 90% threshold!
				RAMUsage:     72.0,
				DiskUsage:    45.0,
				NetworkInKB:  1024.5,
				NetworkOutKB: 512.2,
			},
		}
		rawBytes, _ := json.Marshal(telemetry)

		// Sign payload with private key
		sig := ed25519.Sign(privKey, rawBytes)
		sigBase64 := base64.StdEncoding.EncodeToString(sig)

		req := httptest.NewRequest("POST", "/api/v1/agent/telemetry", bytes.NewReader(rawBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set(middleware.AgentSignatureHdr, sigBase64)
		req.Header.Set("X-Tenant-ID", env.Org.ID.String())
		rec := httptest.NewRecorder()

		env.Router.ServeHTTP(rec, req)
		if rec.Code != http.StatusAccepted {
			t.Fatalf("Expected 202 Accepted on telemetry ingestion, got %d", rec.Code)
		}

		// Verify metric was stored
		metric, err := env.Store.Metrics().GetLatestByDevice(context.Background(), env.Org.ID, deviceID)
		if err != nil {
			t.Fatalf("Failed to retrieve stored metric: %v", err)
		}
		if metric.CPUUsage != 95.5 {
			t.Errorf("Expected CPU 95.5, got %f", metric.CPUUsage)
		}

		// Verify auto-alert was generated for high CPU
		alerts, err := env.Store.Alerts().ListActiveByDevice(context.Background(), env.Org.ID, deviceID)
		if err != nil || len(alerts) == 0 {
			t.Fatalf("Expected threshold alert to be generated for high CPU, got: %v", alerts)
		}
		if alerts[0].Severity != "critical" {
			t.Errorf("Expected severity critical, got %s", alerts[0].Severity)
		}
	}

	// 4. Verify invalid agent signature is rejected
	{
		telemetry := domain.TelemetryReport{
			DeviceID:  deviceID,
			Timestamp: time.Now().UTC(),
			Metrics:   domain.TelemetryMetrics{CPUUsage: 10.0},
		}
		rawBytes, _ := json.Marshal(telemetry)

		req := httptest.NewRequest("POST", "/api/v1/agent/telemetry", bytes.NewReader(rawBytes))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set(middleware.AgentSignatureHdr, base64.StdEncoding.EncodeToString([]byte("invalid_sig_bytes_padding_to_64_bytes_length_required_for_ed25519!!")))
		req.Header.Set("X-Tenant-ID", env.Org.ID.String())
		rec := httptest.NewRecorder()

		env.Router.ServeHTTP(rec, req)
		if rec.Code != http.StatusUnauthorized {
			t.Fatalf("Expected 401 Unauthorized for corrupted agent signature, got %d", rec.Code)
		}
	}
}

func TestRemediationLifecycle(t *testing.T) {
	env := setupTestEnv(t)

	// Create test device
	device := &domain.Device{
		ID:             uuid.New(),
		OrganizationID: env.Org.ID,
		Hostname:       "srv-app-01",
		IPAddress:      "10.0.1.20",
		Status:         domain.DeviceStatusOnline,
	}
	_ = env.Store.Devices().Create(context.Background(), device)

	// 1. Propose low-risk remediation job that requires approval
	var jobID uuid.UUID
	{
		body, _ := json.Marshal(map[string]any{
			"device_id":      device.ID,
			"action_id":      "restart_windows_service",
			"parameters":     map[string]string{"service_name": "Spooler"},
			"justification": "Spooler service unresponsive causing printing failures",
		})
		req := httptest.NewRequest("POST", "/api/v1/remediation/jobs", bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+env.OperatorTok)
		req.Header.Set("X-Tenant-ID", env.Org.ID.String())
		rec := httptest.NewRecorder()

		env.Router.ServeHTTP(rec, req)
		if rec.Code != http.StatusCreated {
			t.Fatalf("Expected 201 Created on job proposal, got %d (body: %s)", rec.Code, rec.Body.String())
		}

		var job domain.RemediationJob
		_ = json.NewDecoder(rec.Body).Decode(&job)
		jobID = job.ID
		if job.Status != domain.RemediationPendingApproval {
			t.Fatalf("Expected status pending_approval, got %s", job.Status)
		}
	}

	// 2. Readonly Viewer attempting to approve should be 403 Forbidden
	{
		req := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/remediation/jobs/%s/approve", jobID), nil)
		req.Header.Set("Authorization", "Bearer "+env.ViewerTok)
		req.Header.Set("X-Tenant-ID", env.Org.ID.String())
		rec := httptest.NewRecorder()

		env.Router.ServeHTTP(rec, req)
		if rec.Code != http.StatusForbidden {
			t.Fatalf("Expected 403 Forbidden for Viewer approval attempt, got %d", rec.Code)
		}
	}

	// 3. Admin approving the job -> 200 OK and status moves to queued
	{
		req := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/remediation/jobs/%s/approve", jobID), nil)
		req.Header.Set("Authorization", "Bearer "+env.AdminToken)
		req.Header.Set("X-Tenant-ID", env.Org.ID.String())
		rec := httptest.NewRecorder()

		env.Router.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("Expected 200 OK on job approval, got %d", rec.Code)
		}

		var job domain.RemediationJob
		_ = json.NewDecoder(rec.Body).Decode(&job)
		if job.Status != domain.RemediationQueued {
			t.Errorf("Expected job status queued after approval, got %s", job.Status)
		}
	}
}

func TestRemoteSupportSessionFlow(t *testing.T) {
	env := setupTestEnv(t)

	device := &domain.Device{
		ID:             uuid.New(),
		OrganizationID: env.Org.ID,
		Hostname:       "finance-laptop-03",
		IPAddress:      "10.0.5.12",
		Status:         domain.DeviceStatusOnline,
	}
	_ = env.Store.Devices().Create(context.Background(), device)

	var sessionID uuid.UUID

	// 1. Operator requests remote session
	{
		body, _ := json.Marshal(map[string]any{
			"device_id":             device.ID,
			"permissions_requested": []string{"screen_view", "remote_control"},
			"reason":                "Assisting user with Excel lockup",
		})
		req := httptest.NewRequest("POST", "/api/v1/remote-support/sessions", bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+env.OperatorTok)
		req.Header.Set("X-Tenant-ID", env.Org.ID.String())
		rec := httptest.NewRecorder()

		env.Router.ServeHTTP(rec, req)
		if rec.Code != http.StatusCreated {
			t.Fatalf("Expected 201 Created for remote session request, got %d", rec.Code)
		}

		var session domain.RemoteSession
		_ = json.NewDecoder(rec.Body).Decode(&session)
		sessionID = session.SessionID
		if session.Status != domain.RemoteSessionAwaitingConsent {
			t.Errorf("Expected awaiting_consent status, got %s", session.Status)
		}
	}

	// 2. Endpoint user approves consent via signal
	{
		signalBody, _ := json.Marshal(map[string]any{
			"type": "consent_decision",
			"payload": map[string]any{
				"approved": true,
				"reason":   "User clicked approve",
			},
		})
		req := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/remote-support/sessions/%s/signal", sessionID), bytes.NewReader(signalBody))
		req.Header.Set("Authorization", "Bearer "+env.OperatorTok)
		req.Header.Set("X-Tenant-ID", env.Org.ID.String())
		rec := httptest.NewRecorder()

		env.Router.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("Expected 200 OK on signal post, got %d", rec.Code)
		}

		// Verify status changed to active
		sess, _ := env.Store.RemoteSessions().GetByID(context.Background(), env.Org.ID, sessionID)
		if sess.Status != domain.RemoteSessionActive {
			t.Errorf("Expected active session status after consent approval, got %s", sess.Status)
		}
	}

	// 3. Terminate session
	{
		req := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/remote-support/sessions/%s/terminate", sessionID), nil)
		req.Header.Set("Authorization", "Bearer "+env.OperatorTok)
		req.Header.Set("X-Tenant-ID", env.Org.ID.String())
		rec := httptest.NewRecorder()

		env.Router.ServeHTTP(rec, req)
		if rec.Code != http.StatusNoContent {
			t.Fatalf("Expected 204 No Content on terminate, got %d", rec.Code)
		}

		sess, _ := env.Store.RemoteSessions().GetByID(context.Background(), env.Org.ID, sessionID)
		if sess.Status != domain.RemoteSessionTerminated {
			t.Errorf("Expected terminated status, got %s", sess.Status)
		}
	}
}

func TestAIAnalysisEndpoint(t *testing.T) {
	env := setupTestEnv(t)

	// Seed incident
	incidentID := uuid.New()
	incident := &domain.Incident{
		ID:             incidentID,
		OrganizationID: env.Org.ID,
		Title:          "Memory Leak in Background Service on 10.0.1.20",
		Severity:       domain.SeverityHigh,
		Status:         domain.IncidentStatusInvestigating,
	}
	_ = env.Store.Incidents().Create(context.Background(), incident)

	req := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/incidents/%s/ai-analysis", incidentID), nil)
	req.Header.Set("Authorization", "Bearer "+env.AdminToken)
	req.Header.Set("X-Tenant-ID", env.Org.ID.String())
	rec := httptest.NewRecorder()

	env.Router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK on AI analysis, got %d (body: %s)", rec.Code, rec.Body.String())
	}

	var res domain.AIAdvisoryResponse
	if err := json.NewDecoder(rec.Body).Decode(&res); err != nil {
		t.Fatalf("Failed to decode AI advisory response: %v", err)
	}

	if res.IncidentID != incidentID {
		t.Errorf("Expected incident ID %s, got %s", incidentID, res.IncidentID)
	}
	if len(res.Hypotheses) == 0 {
		t.Fatal("Expected at least one AI hypothesis")
	}
	if len(res.RecommendedVerificationSteps) == 0 {
		t.Fatal("Expected recommended verification steps")
	}
}
