package storage

import (
	"context"
	"time"

	"github.com/google/uuid"
	"monitopro/services/api/internal/domain"
)

type TenantRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Organization, error)
	GetBySubdomain(ctx context.Context, subdomain string) (*domain.Organization, error)
	Create(ctx context.Context, org *domain.Organization) error
	UpdateAISettings(ctx context.Context, id uuid.UUID, aiEnabled bool) error
}

type UserRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	Create(ctx context.Context, user *domain.User) error
	UpdateMFA(ctx context.Context, userID uuid.UUID, secret *string, enabled bool) error
	GetMembership(ctx context.Context, orgID, userID uuid.UUID) (*domain.Membership, error)
	CreateMembership(ctx context.Context, membership *domain.Membership) error
}

type DeviceRepository interface {
	List(ctx context.Context, orgID uuid.UUID, status string, search string, page, limit int) ([]*domain.Device, int, error)
	GetByID(ctx context.Context, orgID, id uuid.UUID) (*domain.Device, error)
	GetByMachineGUID(ctx context.Context, machineGUID string) (*domain.Device, error)
	Create(ctx context.Context, device *domain.Device) error
	Update(ctx context.Context, device *domain.Device) error
	UpdateStatusAndLastSeen(ctx context.Context, id uuid.UUID, status domain.DeviceStatus, lastSeen time.Time) error
	GetDiagnostics(ctx context.Context, orgID, deviceID uuid.UUID) (*domain.DeviceDiagnostics, error)
	UpsertDiagnostics(ctx context.Context, diag *domain.DeviceDiagnostics) error
	CreateEnrollmentToken(ctx context.Context, token *domain.EnrollmentToken) error
	GetEnrollmentToken(ctx context.Context, tokenStr string) (*domain.EnrollmentToken, error)
	UseEnrollmentToken(ctx context.Context, tokenStr string) error
}

type MetricRepository interface {
	Insert(ctx context.Context, metric *domain.Metric) error
	GetLatestByDevice(ctx context.Context, orgID, deviceID uuid.UUID) (*domain.Metric, error)
	ListByDevice(ctx context.Context, orgID, deviceID uuid.UUID, limit int) ([]*domain.Metric, error)
}

type AlertRepository interface {
	List(ctx context.Context, orgID uuid.UUID, activeOnly bool) ([]*domain.Alert, error)
	GetByID(ctx context.Context, orgID, id uuid.UUID) (*domain.Alert, error)
	Create(ctx context.Context, alert *domain.Alert) error
	Resolve(ctx context.Context, orgID, id uuid.UUID) error
	ListActiveByDevice(ctx context.Context, orgID, deviceID uuid.UUID) ([]*domain.Alert, error)
}

type IncidentRepository interface {
	List(ctx context.Context, orgID uuid.UUID, status string) ([]*domain.Incident, error)
	GetByID(ctx context.Context, orgID, id uuid.UUID) (*domain.Incident, error)
	Create(ctx context.Context, incident *domain.Incident) error
	UpdateStatus(ctx context.Context, orgID, id uuid.UUID, status domain.IncidentStatus) error
	AddTimeline(ctx context.Context, item *domain.IncidentTimeline) error
	GetTimeline(ctx context.Context, incidentID uuid.UUID) ([]*domain.IncidentTimeline, error)
	AssociateAlert(ctx context.Context, incidentID, alertID uuid.UUID) error
	GetAlerts(ctx context.Context, incidentID uuid.UUID) ([]*domain.Alert, error)
	SaveAIReport(ctx context.Context, report *domain.AIAdvisoryReport) error
}

type RemediationRepository interface {
	CreateJob(ctx context.Context, job *domain.RemediationJob) error
	GetJobByID(ctx context.Context, orgID, id uuid.UUID) (*domain.RemediationJob, error)
	ListJobs(ctx context.Context, orgID uuid.UUID, limit int) ([]*domain.RemediationJob, error)
	UpdateJobStatus(ctx context.Context, orgID, id uuid.UUID, status domain.RemediationStatus, output string, passed *bool) error
	ApproveJob(ctx context.Context, orgID, id, approverID uuid.UUID) error
}

type RemoteSessionRepository interface {
	Create(ctx context.Context, session *domain.RemoteSession) error
	GetByID(ctx context.Context, orgID, id uuid.UUID) (*domain.RemoteSession, error)
	UpdateStatus(ctx context.Context, orgID, id uuid.UUID, status domain.RemoteSessionStatus, reason string) error
	Terminate(ctx context.Context, orgID, id uuid.UUID, reason string) error
}

type AuditRepository interface {
	Insert(ctx context.Context, log *domain.AuditLog) error
	List(ctx context.Context, orgID uuid.UUID, limit int) ([]*domain.AuditLog, error)
}

// Storage encapsulates all repositories
type Storage interface {
	Tenants() TenantRepository
	Users() UserRepository
	Devices() DeviceRepository
	Metrics() MetricRepository
	Alerts() AlertRepository
	Incidents() IncidentRepository
	Remediations() RemediationRepository
	RemoteSessions() RemoteSessionRepository
	Audit() AuditRepository
	Close()
}
