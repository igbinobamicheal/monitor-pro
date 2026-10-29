package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"monitopro/services/api/internal/ai"
	"monitopro/services/api/internal/auth"
	"monitopro/services/api/internal/domain"
	"monitopro/services/api/internal/events"
	"monitopro/services/api/internal/middleware"
	"monitopro/services/api/internal/remediation"
	"monitopro/services/api/internal/storage"
)

// RouterConfig contains dependencies for setting up Chi routes
type RouterConfig struct {
	Store         storage.Storage
	TokenManager  *auth.TokenManager
	Publisher     events.EventPublisher
	Catalogue     *remediation.Catalogue
	AIProvider    ai.AIProvider
	ReplicaID     string
	HealthHandler *HealthHandler
}

// NewRouter constructs the complete Chi HTTP router with all middleware and routes
func NewRouter(cfg *RouterConfig) http.Handler {
	r := chi.NewRouter()

	// Global Middleware
	r.Use(middleware.RequestIDMiddleware)
	r.Use(middleware.CORSMiddleware)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Recoverer)
	r.Use(middleware.StructuredLogger)

	// Tenant resolution middleware (inspects Host and X-Tenant-ID)
	tenantResolver := middleware.NewTenantResolver(cfg.Store.Tenants())
	r.Use(tenantResolver.ResolveMiddleware)

	authMiddleware := middleware.NewAuthenticator(cfg.TokenManager)

	// Initialize handlers
	healthH := cfg.HealthHandler
	if healthH == nil {
		healthH = NewHealthHandler(cfg.ReplicaID, nil)
	}
	authH := NewAuthHandler(cfg.Store, cfg.TokenManager)
	orgH := NewOrgHandler(cfg.Store)
	deviceH := NewDeviceHandler(cfg.Store, cfg.Publisher, "secret")
	incidentH := NewIncidentHandler(cfg.Store)
	aiH := NewAIHandler(cfg.Store, cfg.AIProvider)
	remediationH := NewRemediationHandler(cfg.Store, cfg.Catalogue, cfg.Publisher)
	remoteH := NewRemoteSupportHandler(cfg.Store, cfg.Publisher)

	// Health endpoint (public)
	r.Get("/api/health", healthH.Check)
	r.Get("/health", healthH.Check)

	// WebSockets signaling for native remote support
	r.Get("/ws/remote-support/{id}", remoteH.HandleWebSocket)

	// API v1 routes
	r.Route("/api/v1", func(v1 chi.Router) {
		// Public Auth
		v1.Post("/auth/login", authH.Login)

		// Public Agent endpoints (Agent cryptographic enrollment & signed telemetry)
		v1.Post("/agent/enroll", deviceH.AgentEnroll)
		v1.Post("/agent/telemetry", deviceH.AgentTelemetry)

		// Authenticated Routes
		v1.Group(func(authGroup chi.Router) {
			authGroup.Use(authMiddleware.Authenticate)

			authGroup.Post("/auth/logout", authH.Logout)
			authGroup.Post("/auth/mfa/setup", authH.MFASetup)
			authGroup.Post("/auth/mfa/verify", authH.MFAVerify)

			// Tenant-scoped routes
			authGroup.Group(func(tenantGroup chi.Router) {
				tenantGroup.Use(tenantResolver.RequireTenant)

				// Organizations
				tenantGroup.Get("/orgs/current", orgH.GetCurrent)
				tenantGroup.Get("/orgs/{org_id}/settings/ai", orgH.GetAISettings)
				tenantGroup.Put("/orgs/{org_id}/settings/ai", orgH.UpdateAISettings)

				// Devices
				tenantGroup.Get("/devices", deviceH.List)
				tenantGroup.Get("/devices/{id}", deviceH.GetByID)
				tenantGroup.Get("/devices/{id}/diagnostics", deviceH.GetDiagnostics)
				tenantGroup.With(middleware.RequireRoles(domain.RoleOwnerAdmin, domain.RoleITAdmin)).
					Post("/devices/enrollment-tokens", deviceH.CreateEnrollmentToken)

				// Alerts
				tenantGroup.Get("/alerts", incidentH.ListAlerts)

				// Incidents
				tenantGroup.Get("/incidents", incidentH.ListIncidents)
				tenantGroup.Get("/incidents/{id}", incidentH.GetIncident)
				tenantGroup.Get("/incidents/{id}/timeline", incidentH.GetTimeline)
				tenantGroup.Put("/incidents/{id}/status", incidentH.UpdateStatus)

				// AI Advisory
				tenantGroup.Post("/incidents/{id}/ai-analysis", aiH.AnalyzeIncident)

				// Remediation
				tenantGroup.Get("/remediation/actions", remediationH.ListActions)
				tenantGroup.Get("/remediation/jobs", remediationH.ListJobs)
				tenantGroup.Get("/remediation/jobs/{id}", remediationH.GetJob)
				tenantGroup.With(middleware.RequireRoles(domain.RoleOwnerAdmin, domain.RoleITAdmin, domain.RoleITOperator)).
					Post("/remediation/jobs", remediationH.ProposeJob)
				tenantGroup.With(middleware.RequireRoles(domain.RoleOwnerAdmin, domain.RoleITAdmin)).
					Post("/remediation/jobs/{id}/approve", remediationH.ApproveJob)

				// Remote Support
				tenantGroup.With(middleware.RequireRoles(domain.RoleOwnerAdmin, domain.RoleITAdmin, domain.RoleITOperator)).
					Post("/remote-support/sessions", remoteH.RequestSession)
				tenantGroup.Get("/remote-support/sessions/{id}", remoteH.GetSession)
				tenantGroup.Post("/remote-support/sessions/{id}/signal", remoteH.PostSignal)
				tenantGroup.Post("/remote-support/sessions/{id}/terminate", remoteH.TerminateSession)
			})
		})
	})

	return r
}
