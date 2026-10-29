package storage

import (
	"context"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"monitopro/services/api/internal/domain"
)

// MemoryStorage implements in-memory Storage for fast unit/integration testing
type MemoryStorage struct {
	mu          sync.RWMutex
	orgs        map[uuid.UUID]*domain.Organization
	orgsBySub   map[string]*domain.Organization
	users       map[uuid.UUID]*domain.User
	usersByMail map[string]*domain.User
	memberships map[string]*domain.Membership // "orgID:userID" -> Membership
	devices     map[uuid.UUID]*domain.Device
	diagnostics map[uuid.UUID]*domain.DeviceDiagnostics
	tokens      map[string]*domain.EnrollmentToken
	metrics     []*domain.Metric
	alerts      map[uuid.UUID]*domain.Alert
	incidents   map[uuid.UUID]*domain.Incident
	incAlerts   map[uuid.UUID][]uuid.UUID // incidentID -> []alertID
	timelines   map[uuid.UUID][]*domain.IncidentTimeline
	aiReports   map[uuid.UUID]*domain.AIAdvisoryReport
	jobs        map[uuid.UUID]*domain.RemediationJob
	sessions    map[uuid.UUID]*domain.RemoteSession
	auditLogs   []*domain.AuditLog
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		orgs:        make(map[uuid.UUID]*domain.Organization),
		orgsBySub:   make(map[string]*domain.Organization),
		users:       make(map[uuid.UUID]*domain.User),
		usersByMail: make(map[string]*domain.User),
		memberships: make(map[string]*domain.Membership),
		devices:     make(map[uuid.UUID]*domain.Device),
		diagnostics: make(map[uuid.UUID]*domain.DeviceDiagnostics),
		tokens:      make(map[string]*domain.EnrollmentToken),
		metrics:     make([]*domain.Metric, 0),
		alerts:      make(map[uuid.UUID]*domain.Alert),
		incidents:   make(map[uuid.UUID]*domain.Incident),
		incAlerts:   make(map[uuid.UUID][]uuid.UUID),
		timelines:   make(map[uuid.UUID][]*domain.IncidentTimeline),
		aiReports:   make(map[uuid.UUID]*domain.AIAdvisoryReport),
		jobs:        make(map[uuid.UUID]*domain.RemediationJob),
		sessions:    make(map[uuid.UUID]*domain.RemoteSession),
		auditLogs:   make([]*domain.AuditLog, 0),
	}
}

func (m *MemoryStorage) Close() {}
func (m *MemoryStorage) Tenants() TenantRepository         { return &memTenantRepo{m} }
func (m *MemoryStorage) Users() UserRepository             { return &memUserRepo{m} }
func (m *MemoryStorage) Devices() DeviceRepository         { return &memDeviceRepo{m} }
func (m *MemoryStorage) Metrics() MetricRepository         { return &memMetricRepo{m} }
func (m *MemoryStorage) Alerts() AlertRepository           { return &memAlertRepo{m} }
func (m *MemoryStorage) Incidents() IncidentRepository     { return &memIncidentRepo{m} }
func (m *MemoryStorage) Remediations() RemediationRepository { return &memRemediationRepo{m} }
func (m *MemoryStorage) RemoteSessions() RemoteSessionRepository {
	return &memRemoteSessionRepo{m}
}
func (m *MemoryStorage) Audit() AuditRepository { return &memAuditRepo{m} }

// memTenantRepo
type memTenantRepo struct{ m *MemoryStorage }

func (r *memTenantRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Organization, error) {
	r.m.mu.RLock()
	defer r.m.mu.RUnlock()
	org, ok := r.m.orgs[id]
	if !ok {
		return nil, errors.New("organization not found")
	}
	cp := *org
	return &cp, nil
}

func (r *memTenantRepo) GetBySubdomain(ctx context.Context, subdomain string) (*domain.Organization, error) {
	r.m.mu.RLock()
	defer r.m.mu.RUnlock()
	org, ok := r.m.orgsBySub[strings.ToLower(subdomain)]
	if !ok {
		return nil, errors.New("organization not found")
	}
	cp := *org
	return &cp, nil
}

func (r *memTenantRepo) Create(ctx context.Context, org *domain.Organization) error {
	r.m.mu.Lock()
	defer r.m.mu.Unlock()
	if org.ID == uuid.Nil {
		org.ID = uuid.New()
	}
	now := time.Now().UTC()
	org.CreatedAt = now
	org.UpdatedAt = now
	sub := strings.ToLower(org.Subdomain)
	r.m.orgs[org.ID] = org
	r.m.orgsBySub[sub] = org
	return nil
}

func (r *memTenantRepo) UpdateAISettings(ctx context.Context, id uuid.UUID, aiEnabled bool) error {
	r.m.mu.Lock()
	defer r.m.mu.Unlock()
	org, ok := r.m.orgs[id]
	if !ok {
		return errors.New("organization not found")
	}
	org.AIEnabled = aiEnabled
	org.UpdatedAt = time.Now().UTC()
	return nil
}

// memUserRepo
type memUserRepo struct{ m *MemoryStorage }

func (r *memUserRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	r.m.mu.RLock()
	defer r.m.mu.RUnlock()
	user, ok := r.m.users[id]
	if !ok {
		return nil, errors.New("user not found")
	}
	cp := *user
	return &cp, nil
}

func (r *memUserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	r.m.mu.RLock()
	defer r.m.mu.RUnlock()
	user, ok := r.m.usersByMail[strings.ToLower(email)]
	if !ok {
		return nil, errors.New("user not found")
	}
	cp := *user
	return &cp, nil
}

func (r *memUserRepo) Create(ctx context.Context, user *domain.User) error {
	r.m.mu.Lock()
	defer r.m.mu.Unlock()
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}
	now := time.Now().UTC()
	user.CreatedAt = now
	user.UpdatedAt = now
	r.m.users[user.ID] = user
	r.m.usersByMail[strings.ToLower(user.Email)] = user
	return nil
}

func (r *memUserRepo) UpdateMFA(ctx context.Context, userID uuid.UUID, secret *string, enabled bool) error {
	r.m.mu.Lock()
	defer r.m.mu.Unlock()
	u, ok := r.m.users[userID]
	if !ok {
		return errors.New("user not found")
	}
	u.MFASecret = secret
	u.MFAEnabled = enabled
	u.UpdatedAt = time.Now().UTC()
	return nil
}

func (r *memUserRepo) GetMembership(ctx context.Context, orgID, userID uuid.UUID) (*domain.Membership, error) {
	r.m.mu.RLock()
	defer r.m.mu.RUnlock()
	key := orgID.String() + ":" + userID.String()
	m, ok := r.m.memberships[key]
	if !ok {
		return nil, errors.New("membership not found")
	}
	cp := *m
	return &cp, nil
}

func (r *memUserRepo) CreateMembership(ctx context.Context, membership *domain.Membership) error {
	r.m.mu.Lock()
	defer r.m.mu.Unlock()
	if membership.ID == uuid.Nil {
		membership.ID = uuid.New()
	}
	membership.CreatedAt = time.Now().UTC()
	key := membership.OrganizationID.String() + ":" + membership.UserID.String()
	r.m.memberships[key] = membership
	return nil
}

// memDeviceRepo
type memDeviceRepo struct{ m *MemoryStorage }

func (r *memDeviceRepo) List(ctx context.Context, orgID uuid.UUID, status string, search string, page, limit int) ([]*domain.Device, int, error) {
	r.m.mu.RLock()
	defer r.m.mu.RUnlock()

	var matched []*domain.Device
	for _, d := range r.m.devices {
		if d.OrganizationID != orgID {
			continue
		}
		if status != "" && string(d.Status) != status {
			continue
		}
		if search != "" {
			if !strings.Contains(strings.ToLower(d.Hostname), strings.ToLower(search)) &&
				!strings.Contains(d.IPAddress, search) {
				continue
			}
		}
		cp := *d
		matched = append(matched, &cp)
	}

	total := len(matched)
	if limit <= 0 {
		limit = 50
	}
	if page <= 0 {
		page = 1
	}
	start := (page - 1) * limit
	if start >= total {
		return []*domain.Device{}, total, nil
	}
	end := start + limit
	if end > total {
		end = total
	}

	return matched[start:end], total, nil
}

func (r *memDeviceRepo) GetByID(ctx context.Context, orgID, id uuid.UUID) (*domain.Device, error) {
	r.m.mu.RLock()
	defer r.m.mu.RUnlock()
	d, ok := r.m.devices[id]
	if !ok || d.OrganizationID != orgID {
		return nil, errors.New("device not found")
	}
	cp := *d
	return &cp, nil
}

func (r *memDeviceRepo) GetByMachineGUID(ctx context.Context, machineGUID string) (*domain.Device, error) {
	r.m.mu.RLock()
	defer r.m.mu.RUnlock()
	for _, d := range r.m.devices {
		if d.MachineGUID == machineGUID {
			cp := *d
			return &cp, nil
		}
	}
	return nil, errors.New("device not found")
}

func (r *memDeviceRepo) Create(ctx context.Context, device *domain.Device) error {
	r.m.mu.Lock()
	defer r.m.mu.Unlock()
	if device.ID == uuid.Nil {
		device.ID = uuid.New()
	}
	now := time.Now().UTC()
	device.CreatedAt = now
	device.UpdatedAt = now
	r.m.devices[device.ID] = device
	return nil
}

func (r *memDeviceRepo) Update(ctx context.Context, device *domain.Device) error {
	r.m.mu.Lock()
	defer r.m.mu.Unlock()
	d, ok := r.m.devices[device.ID]
	if !ok || d.OrganizationID != device.OrganizationID {
		return errors.New("device not found")
	}
	device.UpdatedAt = time.Now().UTC()
	r.m.devices[device.ID] = device
	return nil
}

func (r *memDeviceRepo) UpdateStatusAndLastSeen(ctx context.Context, id uuid.UUID, status domain.DeviceStatus, lastSeen time.Time) error {
	r.m.mu.Lock()
	defer r.m.mu.Unlock()
	d, ok := r.m.devices[id]
	if !ok {
		return errors.New("device not found")
	}
	d.Status = status
	d.LastSeenAt = &lastSeen
	d.UpdatedAt = time.Now().UTC()
	return nil
}

func (r *memDeviceRepo) GetDiagnostics(ctx context.Context, orgID, deviceID uuid.UUID) (*domain.DeviceDiagnostics, error) {
	r.m.mu.RLock()
	defer r.m.mu.RUnlock()
	diag, ok := r.m.diagnostics[deviceID]
	if !ok || diag.OrganizationID != orgID {
		return nil, errors.New("diagnostics not found")
	}
	cp := *diag
	return &cp, nil
}

func (r *memDeviceRepo) UpsertDiagnostics(ctx context.Context, diag *domain.DeviceDiagnostics) error {
	r.m.mu.Lock()
	defer r.m.mu.Unlock()
	diag.UpdatedAt = time.Now().UTC()
	r.m.diagnostics[diag.DeviceID] = diag
	return nil
}

func (r *memDeviceRepo) CreateEnrollmentToken(ctx context.Context, token *domain.EnrollmentToken) error {
	r.m.mu.Lock()
	defer r.m.mu.Unlock()
	if token.ID == uuid.Nil {
		token.ID = uuid.New()
	}
	token.CreatedAt = time.Now().UTC()
	r.m.tokens[token.Token] = token
	return nil
}

func (r *memDeviceRepo) GetEnrollmentToken(ctx context.Context, tokenStr string) (*domain.EnrollmentToken, error) {
	r.m.mu.RLock()
	defer r.m.mu.RUnlock()
	t, ok := r.m.tokens[tokenStr]
	if !ok {
		return nil, errors.New("enrollment token not found")
	}
	cp := *t
	return &cp, nil
}

func (r *memDeviceRepo) UseEnrollmentToken(ctx context.Context, tokenStr string) error {
	r.m.mu.Lock()
	defer r.m.mu.Unlock()
	t, ok := r.m.tokens[tokenStr]
	if !ok {
		return errors.New("token not found")
	}
	t.UsesCount++
	return nil
}

// memMetricRepo
type memMetricRepo struct{ m *MemoryStorage }

func (r *memMetricRepo) Insert(ctx context.Context, metric *domain.Metric) error {
	r.m.mu.Lock()
	defer r.m.mu.Unlock()
	if metric.Timestamp.IsZero() {
		metric.Timestamp = time.Now().UTC()
	}
	metric.ID = int64(len(r.m.metrics) + 1)
	r.m.metrics = append(r.m.metrics, metric)
	return nil
}

func (r *memMetricRepo) GetLatestByDevice(ctx context.Context, orgID, deviceID uuid.UUID) (*domain.Metric, error) {
	r.m.mu.RLock()
	defer r.m.mu.RUnlock()
	for i := len(r.m.metrics) - 1; i >= 0; i-- {
		m := r.m.metrics[i]
		if m.DeviceID == deviceID && m.OrganizationID == orgID {
			cp := *m
			return &cp, nil
		}
	}
	return nil, errors.New("no metrics found")
}

func (r *memMetricRepo) ListByDevice(ctx context.Context, orgID, deviceID uuid.UUID, limit int) ([]*domain.Metric, error) {
	r.m.mu.RLock()
	defer r.m.mu.RUnlock()
	var list []*domain.Metric
	for i := len(r.m.metrics) - 1; i >= 0; i-- {
		m := r.m.metrics[i]
		if m.DeviceID == deviceID && m.OrganizationID == orgID {
			cp := *m
			list = append(list, &cp)
			if len(list) >= limit {
				break
			}
		}
	}
	return list, nil
}

// memAlertRepo
type memAlertRepo struct{ m *MemoryStorage }

func (r *memAlertRepo) List(ctx context.Context, orgID uuid.UUID, activeOnly bool) ([]*domain.Alert, error) {
	r.m.mu.RLock()
	defer r.m.mu.RUnlock()
	var list []*domain.Alert
	for _, a := range r.m.alerts {
		if a.OrganizationID != orgID {
			continue
		}
		if activeOnly && a.Resolved {
			continue
		}
		cp := *a
		list = append(list, &cp)
	}
	return list, nil
}

func (r *memAlertRepo) GetByID(ctx context.Context, orgID, id uuid.UUID) (*domain.Alert, error) {
	r.m.mu.RLock()
	defer r.m.mu.RUnlock()
	a, ok := r.m.alerts[id]
	if !ok || a.OrganizationID != orgID {
		return nil, errors.New("alert not found")
	}
	cp := *a
	return &cp, nil
}

func (r *memAlertRepo) Create(ctx context.Context, alert *domain.Alert) error {
	r.m.mu.Lock()
	defer r.m.mu.Unlock()
	if alert.ID == uuid.Nil {
		alert.ID = uuid.New()
	}
	alert.CreatedAt = time.Now().UTC()
	r.m.alerts[alert.ID] = alert
	return nil
}

func (r *memAlertRepo) Resolve(ctx context.Context, orgID, id uuid.UUID) error {
	r.m.mu.Lock()
	defer r.m.mu.Unlock()
	a, ok := r.m.alerts[id]
	if !ok || a.OrganizationID != orgID {
		return errors.New("alert not found")
	}
	now := time.Now().UTC()
	a.Resolved = true
	a.ResolvedAt = &now
	return nil
}

func (r *memAlertRepo) ListActiveByDevice(ctx context.Context, orgID, deviceID uuid.UUID) ([]*domain.Alert, error) {
	r.m.mu.RLock()
	defer r.m.mu.RUnlock()
	var list []*domain.Alert
	for _, a := range r.m.alerts {
		if a.OrganizationID == orgID && a.DeviceID == deviceID && !a.Resolved {
			cp := *a
			list = append(list, &cp)
		}
	}
	return list, nil
}

// memIncidentRepo
type memIncidentRepo struct{ m *MemoryStorage }

func (r *memIncidentRepo) List(ctx context.Context, orgID uuid.UUID, status string) ([]*domain.Incident, error) {
	r.m.mu.RLock()
	defer r.m.mu.RUnlock()
	var list []*domain.Incident
	for _, inc := range r.m.incidents {
		if inc.OrganizationID != orgID {
			continue
		}
		if status != "" && string(inc.Status) != status {
			continue
		}
		cp := *inc
		cp.AlertsCount = len(r.m.incAlerts[inc.ID])
		list = append(list, &cp)
	}
	return list, nil
}

func (r *memIncidentRepo) GetByID(ctx context.Context, orgID, id uuid.UUID) (*domain.Incident, error) {
	r.m.mu.RLock()
	defer r.m.mu.RUnlock()
	inc, ok := r.m.incidents[id]
	if !ok || inc.OrganizationID != orgID {
		return nil, errors.New("incident not found")
	}
	cp := *inc
	cp.AlertsCount = len(r.m.incAlerts[inc.ID])
	return &cp, nil
}

func (r *memIncidentRepo) Create(ctx context.Context, incident *domain.Incident) error {
	r.m.mu.Lock()
	defer r.m.mu.Unlock()
	if incident.ID == uuid.Nil {
		incident.ID = uuid.New()
	}
	now := time.Now().UTC()
	incident.CreatedAt = now
	incident.UpdatedAt = now
	r.m.incidents[incident.ID] = incident
	return nil
}

func (r *memIncidentRepo) UpdateStatus(ctx context.Context, orgID, id uuid.UUID, status domain.IncidentStatus) error {
	r.m.mu.Lock()
	defer r.m.mu.Unlock()
	inc, ok := r.m.incidents[id]
	if !ok || inc.OrganizationID != orgID {
		return errors.New("incident not found")
	}
	inc.Status = status
	now := time.Now().UTC()
	inc.UpdatedAt = now
	if status == domain.IncidentStatusResolved || status == domain.IncidentStatusClosed {
		inc.ResolvedAt = &now
	}
	return nil
}

func (r *memIncidentRepo) AddTimeline(ctx context.Context, item *domain.IncidentTimeline) error {
	r.m.mu.Lock()
	defer r.m.mu.Unlock()
	if item.ID == uuid.Nil {
		item.ID = uuid.New()
	}
	item.CreatedAt = time.Now().UTC()
	r.m.timelines[item.IncidentID] = append(r.m.timelines[item.IncidentID], item)
	return nil
}

func (r *memIncidentRepo) GetTimeline(ctx context.Context, incidentID uuid.UUID) ([]*domain.IncidentTimeline, error) {
	r.m.mu.RLock()
	defer r.m.mu.RUnlock()
	list := r.m.timelines[incidentID]
	var res []*domain.IncidentTimeline
	for _, t := range list {
		cp := *t
		res = append(res, &cp)
	}
	return res, nil
}

func (r *memIncidentRepo) AssociateAlert(ctx context.Context, incidentID, alertID uuid.UUID) error {
	r.m.mu.Lock()
	defer r.m.mu.Unlock()
	r.m.incAlerts[incidentID] = append(r.m.incAlerts[incidentID], alertID)
	return nil
}

func (r *memIncidentRepo) GetAlerts(ctx context.Context, incidentID uuid.UUID) ([]*domain.Alert, error) {
	r.m.mu.RLock()
	defer r.m.mu.RUnlock()
	alertIDs := r.m.incAlerts[incidentID]
	var list []*domain.Alert
	for _, id := range alertIDs {
		if a, ok := r.m.alerts[id]; ok {
			cp := *a
			list = append(list, &cp)
		}
	}
	return list, nil
}

func (r *memIncidentRepo) SaveAIReport(ctx context.Context, report *domain.AIAdvisoryReport) error {
	r.m.mu.Lock()
	defer r.m.mu.Unlock()
	if report.ID == uuid.Nil {
		report.ID = uuid.New()
	}
	report.CreatedAt = time.Now().UTC()
	r.m.aiReports[report.ID] = report
	return nil
}

// memRemediationRepo
type memRemediationRepo struct{ m *MemoryStorage }

func (r *memRemediationRepo) CreateJob(ctx context.Context, job *domain.RemediationJob) error {
	r.m.mu.Lock()
	defer r.m.mu.Unlock()
	if job.ID == uuid.Nil {
		job.ID = uuid.New()
	}
	now := time.Now().UTC()
	job.CreatedAt = now
	job.UpdatedAt = now
	r.m.jobs[job.ID] = job
	return nil
}

func (r *memRemediationRepo) GetJobByID(ctx context.Context, orgID, id uuid.UUID) (*domain.RemediationJob, error) {
	r.m.mu.RLock()
	defer r.m.mu.RUnlock()
	j, ok := r.m.jobs[id]
	if !ok || j.OrganizationID != orgID {
		return nil, errors.New("remediation job not found")
	}
	cp := *j
	return &cp, nil
}

func (r *memRemediationRepo) ListJobs(ctx context.Context, orgID uuid.UUID, limit int) ([]*domain.RemediationJob, error) {
	r.m.mu.RLock()
	defer r.m.mu.RUnlock()
	var list []*domain.RemediationJob
	for _, j := range r.m.jobs {
		if j.OrganizationID == orgID {
			cp := *j
			list = append(list, &cp)
			if limit > 0 && len(list) >= limit {
				break
			}
		}
	}
	return list, nil
}

func (r *memRemediationRepo) UpdateJobStatus(ctx context.Context, orgID, id uuid.UUID, status domain.RemediationStatus, output string, passed *bool) error {
	r.m.mu.Lock()
	defer r.m.mu.Unlock()
	j, ok := r.m.jobs[id]
	if !ok || j.OrganizationID != orgID {
		return errors.New("remediation job not found")
	}
	j.Status = status
	j.ExecutionOutput = output
	j.VerificationPassed = passed
	j.UpdatedAt = time.Now().UTC()
	return nil
}

func (r *memRemediationRepo) ApproveJob(ctx context.Context, orgID, id, approverID uuid.UUID) error {
	r.m.mu.Lock()
	defer r.m.mu.Unlock()
	j, ok := r.m.jobs[id]
	if !ok || j.OrganizationID != orgID {
		return errors.New("remediation job not found")
	}
	if j.Status != domain.RemediationPendingApproval {
		return errors.New("job not pending approval")
	}
	j.Status = domain.RemediationApproved
	j.ApprovedBy = &approverID
	j.UpdatedAt = time.Now().UTC()
	return nil
}

// memRemoteSessionRepo
type memRemoteSessionRepo struct{ m *MemoryStorage }

func (r *memRemoteSessionRepo) Create(ctx context.Context, session *domain.RemoteSession) error {
	r.m.mu.Lock()
	defer r.m.mu.Unlock()
	if session.SessionID == uuid.Nil {
		session.SessionID = uuid.New()
	}
	session.CreatedAt = time.Now().UTC()
	r.m.sessions[session.SessionID] = session
	return nil
}

func (r *memRemoteSessionRepo) GetByID(ctx context.Context, orgID, id uuid.UUID) (*domain.RemoteSession, error) {
	r.m.mu.RLock()
	defer r.m.mu.RUnlock()
	s, ok := r.m.sessions[id]
	if !ok || s.OrganizationID != orgID {
		return nil, errors.New("remote session not found")
	}
	cp := *s
	return &cp, nil
}

func (r *memRemoteSessionRepo) UpdateStatus(ctx context.Context, orgID, id uuid.UUID, status domain.RemoteSessionStatus, reason string) error {
	r.m.mu.Lock()
	defer r.m.mu.Unlock()
	s, ok := r.m.sessions[id]
	if !ok || s.OrganizationID != orgID {
		return errors.New("remote session not found")
	}
	s.Status = status
	s.TerminationReason = reason
	now := time.Now().UTC()
	if status == domain.RemoteSessionActive {
		s.ConsentGrantedAt = &now
		s.StartedAt = &now
	} else if status == domain.RemoteSessionTerminated || status == domain.RemoteSessionRejected {
		s.EndedAt = &now
	}
	return nil
}

func (r *memRemoteSessionRepo) Terminate(ctx context.Context, orgID, id uuid.UUID, reason string) error {
	return r.UpdateStatus(ctx, orgID, id, domain.RemoteSessionTerminated, reason)
}

// memAuditRepo
type memAuditRepo struct{ m *MemoryStorage }

func (r *memAuditRepo) Insert(ctx context.Context, log *domain.AuditLog) error {
	r.m.mu.Lock()
	defer r.m.mu.Unlock()
	if log.ID == uuid.Nil {
		log.ID = uuid.New()
	}
	log.CreatedAt = time.Now().UTC()
	r.m.auditLogs = append(r.m.auditLogs, log)
	return nil
}

func (r *memAuditRepo) List(ctx context.Context, orgID uuid.UUID, limit int) ([]*domain.AuditLog, error) {
	r.m.mu.RLock()
	defer r.m.mu.RUnlock()
	var list []*domain.AuditLog
	for i := len(r.m.auditLogs) - 1; i >= 0; i-- {
		l := r.m.auditLogs[i]
		if l.OrganizationID == orgID {
			cp := *l
			list = append(list, &cp)
			if limit > 0 && len(list) >= limit {
				break
			}
		}
	}
	return list, nil
}
