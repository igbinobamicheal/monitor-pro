package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// Role defines role-based access control levels matching schema
type Role string

const (
	RoleOwnerAdmin     Role = "owner_admin"
	RoleITAdmin        Role = "it_admin"
	RoleITOperator     Role = "it_operator"
	RoleReadonlyViewer Role = "readonly_viewer"
)

func (r Role) IsValid() bool {
	switch r {
	case RoleOwnerAdmin, RoleITAdmin, RoleITOperator, RoleReadonlyViewer:
		return true
	default:
		return false
	}
}

// CanApproveRemediation returns true if role is authorized to approve remediation jobs
func (r Role) CanApproveRemediation() bool {
	return r == RoleOwnerAdmin || r == RoleITAdmin
}

// CanExecuteActions returns true if role can propose remediation or initiate remote sessions
func (r Role) CanExecuteActions() bool {
	return r == RoleOwnerAdmin || r == RoleITAdmin || r == RoleITOperator
}

// DeviceStatus defines connectivity states
type DeviceStatus string

const (
	DeviceStatusOnline   DeviceStatus = "online"
	DeviceStatusDegraded DeviceStatus = "degraded"
	DeviceStatusOffline  DeviceStatus = "offline"
)

// IncidentSeverity defines severity categories
type IncidentSeverity string

const (
	SeverityLow      IncidentSeverity = "low"
	SeverityMedium   IncidentSeverity = "medium"
	SeverityHigh     IncidentSeverity = "high"
	SeverityCritical IncidentSeverity = "critical"
)

// IncidentStatus defines resolution lifecycle states
type IncidentStatus string

const (
	IncidentStatusOpen          IncidentStatus = "open"
	IncidentStatusInvestigating IncidentStatus = "investigating"
	IncidentStatusResolved      IncidentStatus = "resolved"
	IncidentStatusClosed        IncidentStatus = "closed"
)

// RemediationStatus defines remediation lifecycle states
type RemediationStatus string

const (
	RemediationPendingApproval RemediationStatus = "pending_approval"
	RemediationApproved        RemediationStatus = "approved"
	RemediationQueued          RemediationStatus = "queued"
	RemediationExecuting       RemediationStatus = "executing"
	RemediationSucceeded       RemediationStatus = "succeeded"
	RemediationFailed          RemediationStatus = "failed"
	RemediationRejected        RemediationStatus = "rejected"
)

// RemoteSessionStatus defines screen sharing consent states
type RemoteSessionStatus string

const (
	RemoteSessionAwaitingConsent RemoteSessionStatus = "awaiting_consent"
	RemoteSessionActive          RemoteSessionStatus = "active"
	RemoteSessionRejected        RemoteSessionStatus = "rejected"
	RemoteSessionTerminated      RemoteSessionStatus = "terminated"
	RemoteSessionExpired         RemoteSessionStatus = "expired"
)

// Organization represents a multi-tenant workspace
type Organization struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Subdomain string    `json:"subdomain"`
	AIEnabled bool      `json:"ai_enabled"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// User represents an authenticated account
type User struct {
	ID              uuid.UUID `json:"id"`
	Email           string    `json:"email"`
	PasswordHash    string    `json:"-"`
	FullName        string    `json:"full_name"`
	MFASecret       *string   `json:"-"`
	MFAEnabled      bool      `json:"mfa_enabled"`
	IsPlatformAdmin bool      `json:"is_platform_admin"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// UserProfile is the public projection returned to clients
type UserProfile struct {
	ID         uuid.UUID `json:"id"`
	Email      string    `json:"email"`
	FullName   string    `json:"full_name"`
	Role       Role      `json:"role"`
	MFAEnabled bool      `json:"mfa_enabled"`
}

// Membership defines user role within a specific tenant
type Membership struct {
	ID             uuid.UUID `json:"id"`
	OrganizationID uuid.UUID `json:"organization_id"`
	UserID         uuid.UUID `json:"user_id"`
	Role           Role      `json:"role"`
	CreatedAt      time.Time `json:"created_at"`
}

// Device represents an enrolled endpoint fleet item
type Device struct {
	ID                uuid.UUID    `json:"id"`
	OrganizationID    uuid.UUID    `json:"organization_id"`
	Hostname          string       `json:"hostname"`
	IPAddress         string       `json:"ip_address"`
	Status            DeviceStatus `json:"status"`
	OSVersion         string       `json:"os_version,omitempty"`
	MachineGUID       string       `json:"machine_guid,omitempty"`
	PublicKey         string       `json:"public_key,omitempty"`
	AssignedUserEmail string       `json:"assigned_user,omitempty"`
	LastSeenAt        *time.Time   `json:"last_seen,omitempty"`
	CreatedAt         time.Time    `json:"created_at"`
	UpdatedAt         time.Time    `json:"updated_at"`
}

// TelemetryMetrics represents resource utilization snapshot
type TelemetryMetrics struct {
	CPUUsage     float64 `json:"cpu_usage"`
	RAMUsage     float64 `json:"ram_usage"`
	DiskUsage    float64 `json:"disk_usage"`
	NetworkInKB  float64 `json:"network_in_kb"`
	NetworkOutKB float64 `json:"network_out_kb"`
}

// DeviceDetails embeds Device with latest telemetry snapshot
type DeviceDetails struct {
	Device
	UptimeSeconds uint64            `json:"uptime_seconds"`
	LatestMetrics *TelemetryMetrics `json:"latest_metrics,omitempty"`
}

// Metric represents a single time-series sample
type Metric struct {
	ID             int64     `json:"id"`
	OrganizationID uuid.UUID `json:"organization_id"`
	DeviceID       uuid.UUID `json:"device_id"`
	CPUUsage       float64   `json:"cpu_usage"`
	RAMUsage       float64   `json:"ram_usage"`
	DiskUsage      float64   `json:"disk_usage"`
	NetworkIn      float64   `json:"network_in"`
	NetworkOut     float64   `json:"network_out"`
	Uptime         uint64    `json:"uptime"`
	Timestamp      time.Time `json:"timestamp"`
}

// TelemetryReport is submitted by endpoint agents
type TelemetryReport struct {
	DeviceID         uuid.UUID        `json:"device_id"`
	Timestamp        time.Time        `json:"timestamp"`
	Metrics          TelemetryMetrics `json:"metrics"`
	DiagnosticsDelta *json.RawMessage `json:"diagnostics_delta,omitempty"`
}

// DeviceDiagnostics holds deep OS diagnostic data
type DeviceDiagnostics struct {
	DeviceID          uuid.UUID       `json:"device_id"`
	OrganizationID    uuid.UUID       `json:"organization_id"`
	SystemInfo        json.RawMessage `json:"system,omitempty"`
	CPUInfo           json.RawMessage `json:"cpu_cores,omitempty"`
	MemoryInfo        json.RawMessage `json:"memory,omitempty"`
	StorageInfo       json.RawMessage `json:"storage,omitempty"`
	NetworkInfo       json.RawMessage `json:"network,omitempty"`
	SecurityInfo      json.RawMessage `json:"security,omitempty"`
	ProcessesInfo     json.RawMessage `json:"processes,omitempty"`
	ServicesInfo      json.RawMessage `json:"services,omitempty"`
	EventLogs         json.RawMessage `json:"event_logs,omitempty"`
	InstalledSoftware json.RawMessage `json:"software,omitempty"`
	UpdatedAt         time.Time       `json:"updated_at"`
}

// Rule defines threshold alerting logic
type Rule struct {
	ID             uuid.UUID `json:"id"`
	OrganizationID uuid.UUID `json:"organization_id"`
	Name           string    `json:"name"`
	MetricType     string    `json:"metric_type"` // cpu, ram, disk
	Operator       string    `json:"operator"`    // >, >=, <, <=
	Threshold      float64   `json:"threshold"`
	Severity       string    `json:"severity"` // info, warning, critical
	Enabled        bool      `json:"enabled"`
	CreatedAt      time.Time `json:"created_at"`
}

// Alert represents an anomaly event triggered on an endpoint
type Alert struct {
	ID             uuid.UUID  `json:"id"`
	OrganizationID uuid.UUID  `json:"organization_id"`
	DeviceID       uuid.UUID  `json:"device_id"`
	RuleID         *uuid.UUID `json:"rule_id,omitempty"`
	RuleName       string     `json:"rule_name"`
	Message        string     `json:"message"`
	Severity       string     `json:"severity"` // info, warning, critical
	Resolved       bool       `json:"resolved"`
	ResolvedAt     *time.Time `json:"resolved_at,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}

// Incident represents a correlated group of issues requiring resolution
type Incident struct {
	ID             uuid.UUID        `json:"id"`
	OrganizationID uuid.UUID        `json:"organization_id"`
	Title          string           `json:"title"`
	Severity       IncidentSeverity `json:"severity"`
	Status         IncidentStatus   `json:"status"`
	AssignedTo     *string          `json:"assigned_to,omitempty"`
	AssignedUserID *uuid.UUID       `json:"assigned_user_id,omitempty"`
	AlertsCount    int              `json:"alerts_count"`
	Alerts         []*Alert         `json:"alerts,omitempty"`
	ResolvedAt     *time.Time       `json:"resolved_at,omitempty"`
	CreatedAt      time.Time        `json:"created_at"`
	UpdatedAt      time.Time        `json:"updated_at"`
}

// IncidentTimeline represents an audit/progress entry on an incident
type IncidentTimeline struct {
	ID             uuid.UUID  `json:"id"`
	IncidentID     uuid.UUID  `json:"incident_id"`
	AuthorUserID   *uuid.UUID `json:"author_user_id,omitempty"`
	AuthorName     string     `json:"author_name,omitempty"`
	EventType      string     `json:"event_type"`
	Content        string     `json:"content"`
	CreatedAt      time.Time  `json:"created_at"`
}

// RemediationAction is a validated catalogue item
type RemediationAction struct {
	ID               string          `json:"id"`
	Title            string          `json:"title"`
	RiskLevel        string          `json:"risk_level"` // read_only, low_risk, destructive
	RequiresApproval bool            `json:"requires_approval"`
	TimeoutSeconds   int             `json:"timeout_seconds"`
	ParameterSchema  json.RawMessage `json:"parameter_schema"`
}

// RemediationJob represents a single requested execution
type RemediationJob struct {
	ID                 uuid.UUID         `json:"id"`
	OrganizationID     uuid.UUID         `json:"organization_id"`
	DeviceID           uuid.UUID         `json:"device_id"`
	ActionID           string            `json:"action_id"`
	Parameters         json.RawMessage   `json:"parameters"`
	Status             RemediationStatus `json:"status"`
	RequestedBy        *uuid.UUID        `json:"requested_by_id,omitempty"`
	Requester          string            `json:"requester,omitempty"`
	ApprovedBy         *uuid.UUID        `json:"approved_by_id,omitempty"`
	Approver           string            `json:"approver,omitempty"`
	ExecutionOutput    string            `json:"execution_output,omitempty"`
	VerificationPassed *bool             `json:"verification_passed,omitempty"`
	CreatedAt          time.Time         `json:"created_at"`
	UpdatedAt          time.Time         `json:"updated_at"`
}

// RemoteSession represents a consent-negotiated screen sharing session
type RemoteSession struct {
	SessionID         uuid.UUID           `json:"session_id"`
	OrganizationID    uuid.UUID           `json:"organization_id"`
	DeviceID          uuid.UUID           `json:"device_id"`
	OperatorID        uuid.UUID           `json:"operator_id"`
	OperatorName      string              `json:"operator_name,omitempty"`
	Status            RemoteSessionStatus `json:"status"`
	Permissions       []string            `json:"permissions"`
	Reason            string              `json:"reason"`
	SignalingURL      string              `json:"signaling_url,omitempty"`
	SessionToken      string              `json:"session_token,omitempty"`
	ConsentGrantedAt  *time.Time          `json:"consent_granted_at,omitempty"`
	StartedAt         *time.Time          `json:"started_at,omitempty"`
	EndedAt           *time.Time          `json:"ended_at,omitempty"`
	TerminationReason string              `json:"termination_reason,omitempty"`
	ExpiresAt         time.Time           `json:"expires_at"`
	CreatedAt         time.Time           `json:"created_at"`
}

// AuditLog records security-critical actions
type AuditLog struct {
	ID             uuid.UUID       `json:"id"`
	OrganizationID uuid.UUID       `json:"organization_id"`
	ActorType      string          `json:"actor_type"` // user, agent, system
	ActorID        *uuid.UUID      `json:"actor_id,omitempty"`
	Action         string          `json:"action"`
	ResourceType   string          `json:"resource_type"`
	ResourceID     *uuid.UUID      `json:"resource_id,omitempty"`
	Metadata       json.RawMessage `json:"metadata,omitempty"`
	IPAddress      string          `json:"ip_address,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
}

// EnrollmentToken allows an endpoint agent to register
type EnrollmentToken struct {
	ID             uuid.UUID  `json:"id"`
	OrganizationID uuid.UUID  `json:"organization_id"`
	Token          string     `json:"token"`
	ExpiresAt      time.Time  `json:"expires_at"`
	MaxUses        int        `json:"max_uses"`
	UsesCount      int        `json:"uses_count"`
	CreatedBy      *uuid.UUID `json:"created_by,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
}

// AISettings configures tenant-level advisory engine
type AISettings struct {
	Enabled   bool   `json:"enabled"`
	RedactPII bool   `json:"redact_pii"`
	Provider  string `json:"provider"` // mock_advisory, anthropic, openai, local_ollama
}

// AIAdvisoryReport represents stored diagnostic analysis
type AIAdvisoryReport struct {
	ID              uuid.UUID       `json:"id"`
	OrganizationID  uuid.UUID       `json:"organization_id"`
	IncidentID      uuid.UUID       `json:"incident_id"`
	Summary         string          `json:"summary"`
	Hypotheses      json.RawMessage `json:"hypotheses"`
	Recommendations json.RawMessage `json:"recommendations"`
	Provider        string          `json:"provider"`
	CreatedAt       time.Time       `json:"created_at"`
}

// AIHypothesis defines root-cause hypothesis from advisory AI
type AIHypothesis struct {
	Rank                  int      `json:"rank"`
	Cause                 string   `json:"cause"`
	Confidence            float64  `json:"confidence"`
	SupportingEvidence    []string `json:"supporting_evidence"`
	ContradictingEvidence []string `json:"contradicting_evidence"`
}

// AISuggestedRemediation links hypotheses to allowlisted remediation catalogue
type AISuggestedRemediation struct {
	ActionID        string `json:"action_id"`
	RiskLevel       string `json:"risk_level"`
	ExpectedOutcome string `json:"expected_outcome"`
}

// AIAdvisoryResponse is the structured diagnostic response
type AIAdvisoryResponse struct {
	IncidentID                   uuid.UUID                `json:"incident_id"`
	Summary                      string                   `json:"summary"`
	Hypotheses                   []AIHypothesis           `json:"hypotheses"`
	RecommendedVerificationSteps []string                 `json:"recommended_verification_steps"`
	SuggestedRemediations        []AISuggestedRemediation `json:"suggested_remediations"`
}

// ErrorResponse matches OpenAPI error format
type ErrorResponse struct {
	Error string `json:"error"`
	Code  string `json:"code"`
}
