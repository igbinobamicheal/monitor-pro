package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"monitopro/services/api/internal/domain"
)

// PostgresStorage implements Storage using PostgreSQL pgxpool
type PostgresStorage struct {
	pool *pgxpool.Pool
}

func NewPostgresStorage(pool *pgxpool.Pool) *PostgresStorage {
	return &PostgresStorage{pool: pool}
}

// NewPostgresStorageFromConnString parses connection string, creates connection pool and verifies connection
func NewPostgresStorageFromConnString(ctx context.Context, connStr string) (*PostgresStorage, error) {
	poolConfig, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database configuration: %w", err)
	}
	poolConfig.MaxConns = 50
	poolConfig.MinConns = 5
	poolConfig.MaxConnLifetime = 1 * time.Hour
	poolConfig.MaxConnIdleTime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create postgres connection pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("postgres ping failed: %w", err)
	}
	return NewPostgresStorage(pool), nil
}

func (s *PostgresStorage) Pool() *pgxpool.Pool {
	return s.pool
}

func (s *PostgresStorage) Close() {
	if s.pool != nil {
		s.pool.Close()
	}
}

// withTenantTx executes a callback inside a transaction with app.current_tenant_id set for RLS
func (s *PostgresStorage) withTenantTx(ctx context.Context, orgID uuid.UUID, fn func(tx pgx.Tx) error) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if orgID != uuid.Nil {
		_, err = tx.Exec(ctx, "SELECT set_config('app.current_tenant_id', $1, true)", orgID.String())
	} else {
		_, err = tx.Exec(ctx, "SELECT set_config('app.current_tenant_id', 'bypass', true)")
	}
	if err != nil {
		return fmt.Errorf("failed to set rls tenant: %w", err)
	}

	if err := fn(tx); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (s *PostgresStorage) Tenants() TenantRepository         { return &pgTenantRepo{s} }
func (s *PostgresStorage) Users() UserRepository             { return &pgUserRepo{s} }
func (s *PostgresStorage) Devices() DeviceRepository         { return &pgDeviceRepo{s} }
func (s *PostgresStorage) Metrics() MetricRepository         { return &pgMetricRepo{s} }
func (s *PostgresStorage) Alerts() AlertRepository           { return &pgAlertRepo{s} }
func (s *PostgresStorage) Incidents() IncidentRepository     { return &pgIncidentRepo{s} }
func (s *PostgresStorage) Remediations() RemediationRepository { return &pgRemediationRepo{s} }
func (s *PostgresStorage) RemoteSessions() RemoteSessionRepository {
	return &pgRemoteSessionRepo{s}
}
func (s *PostgresStorage) Audit() AuditRepository { return &pgAuditRepo{s} }

// pgTenantRepo
type pgTenantRepo struct{ s *PostgresStorage }

func (r *pgTenantRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Organization, error) {
	query := `SELECT id, name, subdomain, ai_enabled, created_at, updated_at FROM organizations WHERE id = $1`
	var org domain.Organization
	err := r.s.pool.QueryRow(ctx, query, id).Scan(&org.ID, &org.Name, &org.Subdomain, &org.AIEnabled, &org.CreatedAt, &org.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("organization not found")
		}
		return nil, err
	}
	return &org, nil
}

func (r *pgTenantRepo) GetBySubdomain(ctx context.Context, subdomain string) (*domain.Organization, error) {
	query := `SELECT id, name, subdomain, ai_enabled, created_at, updated_at FROM organizations WHERE subdomain = $1`
	var org domain.Organization
	err := r.s.pool.QueryRow(ctx, query, strings.ToLower(subdomain)).Scan(&org.ID, &org.Name, &org.Subdomain, &org.AIEnabled, &org.CreatedAt, &org.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("organization not found")
		}
		return nil, err
	}
	return &org, nil
}

func (r *pgTenantRepo) Create(ctx context.Context, org *domain.Organization) error {
	query := `
		INSERT INTO organizations (id, name, subdomain, ai_enabled, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	if org.ID == uuid.Nil {
		org.ID = uuid.New()
	}
	now := time.Now().UTC()
	org.CreatedAt = now
	org.UpdatedAt = now
	_, err := r.s.pool.Exec(ctx, query, org.ID, org.Name, strings.ToLower(org.Subdomain), org.AIEnabled, org.CreatedAt, org.UpdatedAt)
	return err
}

func (r *pgTenantRepo) UpdateAISettings(ctx context.Context, id uuid.UUID, aiEnabled bool) error {
	query := `UPDATE organizations SET ai_enabled = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`
	_, err := r.s.pool.Exec(ctx, query, aiEnabled, id)
	return err
}

// pgUserRepo
type pgUserRepo struct{ s *PostgresStorage }

func (r *pgUserRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	query := `SELECT id, email, password_hash, full_name, mfa_secret, mfa_enabled, is_platform_admin, created_at, updated_at FROM users WHERE id = $1`
	var u domain.User
	err := r.s.pool.QueryRow(ctx, query, id).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.FullName, &u.MFASecret, &u.MFAEnabled, &u.IsPlatformAdmin, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &u, nil
}

func (r *pgUserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `SELECT id, email, password_hash, full_name, mfa_secret, mfa_enabled, is_platform_admin, created_at, updated_at FROM users WHERE LOWER(email) = LOWER($1)`
	var u domain.User
	err := r.s.pool.QueryRow(ctx, query, email).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.FullName, &u.MFASecret, &u.MFAEnabled, &u.IsPlatformAdmin, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &u, nil
}

func (r *pgUserRepo) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (id, email, password_hash, full_name, mfa_secret, mfa_enabled, is_platform_admin, created_at, updated_at)
		VALUES ($1, LOWER($2), $3, $4, $5, $6, $7, $8, $9)
	`
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}
	now := time.Now().UTC()
	user.CreatedAt = now
	user.UpdatedAt = now
	_, err := r.s.pool.Exec(ctx, query, user.ID, user.Email, user.PasswordHash, user.FullName, user.MFASecret, user.MFAEnabled, user.IsPlatformAdmin, user.CreatedAt, user.UpdatedAt)
	return err
}

func (r *pgUserRepo) UpdateMFA(ctx context.Context, userID uuid.UUID, secret *string, enabled bool) error {
	query := `UPDATE users SET mfa_secret = $1, mfa_enabled = $2, updated_at = CURRENT_TIMESTAMP WHERE id = $3`
	_, err := r.s.pool.Exec(ctx, query, secret, enabled, userID)
	return err
}

func (r *pgUserRepo) GetMembership(ctx context.Context, orgID, userID uuid.UUID) (*domain.Membership, error) {
	query := `SELECT id, organization_id, user_id, role, created_at FROM memberships WHERE organization_id = $1 AND user_id = $2`
	var m domain.Membership
	err := r.s.pool.QueryRow(ctx, query, orgID, userID).Scan(&m.ID, &m.OrganizationID, &m.UserID, &m.Role, &m.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("membership not found")
		}
		return nil, err
	}
	return &m, nil
}

func (r *pgUserRepo) CreateMembership(ctx context.Context, membership *domain.Membership) error {
	query := `INSERT INTO memberships (id, organization_id, user_id, role, created_at) VALUES ($1, $2, $3, $4, $5)`
	if membership.ID == uuid.Nil {
		membership.ID = uuid.New()
	}
	membership.CreatedAt = time.Now().UTC()
	_, err := r.s.pool.Exec(ctx, query, membership.ID, membership.OrganizationID, membership.UserID, membership.Role, membership.CreatedAt)
	return err
}

// pgDeviceRepo
type pgDeviceRepo struct{ s *PostgresStorage }

func (r *pgDeviceRepo) List(ctx context.Context, orgID uuid.UUID, status string, search string, page, limit int) ([]*domain.Device, int, error) {
	var devices []*domain.Device
	var total int

	err := r.s.withTenantTx(ctx, orgID, func(tx pgx.Tx) error {
		var conditions []string
		var args []any
		idx := 1

		conditions = append(conditions, fmt.Sprintf("organization_id = $%d", idx))
		args = append(args, orgID)
		idx++

		if status != "" {
			conditions = append(conditions, fmt.Sprintf("status = $%d", idx))
			args = append(args, status)
			idx++
		}

		if search != "" {
			conditions = append(conditions, fmt.Sprintf("(hostname ILIKE $%d OR ip_address ILIKE $%d)", idx, idx))
			args = append(args, "%"+search+"%")
			idx++
		}

		whereClause := strings.Join(conditions, " AND ")

		countQuery := fmt.Sprintf("SELECT COUNT(*) FROM devices WHERE %s", whereClause)
		if err := tx.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
			return err
		}

		if limit <= 0 {
			limit = 50
		}
		if page <= 0 {
			page = 1
		}
		offset := (page - 1) * limit

		listQuery := fmt.Sprintf(`
			SELECT id, organization_id, hostname, ip_address, status, os_version, machine_guid, public_key, assigned_user_email, last_seen_at, created_at, updated_at
			FROM devices
			WHERE %s
			ORDER BY last_seen_at DESC NULLS LAST, created_at DESC
			LIMIT $%d OFFSET $%d
		`, whereClause, idx, idx+1)

		args = append(args, limit, offset)

		rows, err := tx.Query(ctx, listQuery, args...)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var d domain.Device
			var osVer, mGuid, pubKey, assigned *string
			if err := rows.Scan(&d.ID, &d.OrganizationID, &d.Hostname, &d.IPAddress, &d.Status, &osVer, &mGuid, &pubKey, &assigned, &d.LastSeenAt, &d.CreatedAt, &d.UpdatedAt); err != nil {
				return err
			}
			if osVer != nil {
				d.OSVersion = *osVer
			}
			if mGuid != nil {
				d.MachineGUID = *mGuid
			}
			if pubKey != nil {
				d.PublicKey = *pubKey
			}
			if assigned != nil {
				d.AssignedUserEmail = *assigned
			}
			devices = append(devices, &d)
		}

		return rows.Err()
	})

	return devices, total, err
}

func (r *pgDeviceRepo) GetByID(ctx context.Context, orgID, id uuid.UUID) (*domain.Device, error) {
	var d domain.Device
	err := r.s.withTenantTx(ctx, orgID, func(tx pgx.Tx) error {
		query := `
			SELECT id, organization_id, hostname, ip_address, status, os_version, machine_guid, public_key, assigned_user_email, last_seen_at, created_at, updated_at
			FROM devices
			WHERE id = $1 AND organization_id = $2
		`
		var osVer, mGuid, pubKey, assigned *string
		err := tx.QueryRow(ctx, query, id, orgID).Scan(&d.ID, &d.OrganizationID, &d.Hostname, &d.IPAddress, &d.Status, &osVer, &mGuid, &pubKey, &assigned, &d.LastSeenAt, &d.CreatedAt, &d.UpdatedAt)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return errors.New("device not found")
			}
			return err
		}
		if osVer != nil {
			d.OSVersion = *osVer
		}
		if mGuid != nil {
			d.MachineGUID = *mGuid
		}
		if pubKey != nil {
			d.PublicKey = *pubKey
		}
		if assigned != nil {
			d.AssignedUserEmail = *assigned
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *pgDeviceRepo) GetByMachineGUID(ctx context.Context, machineGUID string) (*domain.Device, error) {
	var d domain.Device
	query := `
		SELECT id, organization_id, hostname, ip_address, status, os_version, machine_guid, public_key, assigned_user_email, last_seen_at, created_at, updated_at
		FROM devices
		WHERE machine_guid = $1
		LIMIT 1
	`
	var osVer, mGuid, pubKey, assigned *string
	err := r.s.pool.QueryRow(ctx, query, machineGUID).Scan(&d.ID, &d.OrganizationID, &d.Hostname, &d.IPAddress, &d.Status, &osVer, &mGuid, &pubKey, &assigned, &d.LastSeenAt, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("device not found")
		}
		return nil, err
	}
	if osVer != nil {
		d.OSVersion = *osVer
	}
	if mGuid != nil {
		d.MachineGUID = *mGuid
	}
	if pubKey != nil {
		d.PublicKey = *pubKey
	}
	if assigned != nil {
		d.AssignedUserEmail = *assigned
	}
	return &d, nil
}

func (r *pgDeviceRepo) Create(ctx context.Context, device *domain.Device) error {
	if device.ID == uuid.Nil {
		device.ID = uuid.New()
	}
	now := time.Now().UTC()
	device.CreatedAt = now
	device.UpdatedAt = now

	return r.s.withTenantTx(ctx, device.OrganizationID, func(tx pgx.Tx) error {
		query := `
			INSERT INTO devices (id, organization_id, hostname, ip_address, status, os_version, machine_guid, public_key, assigned_user_email, last_seen_at, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		`
		_, err := tx.Exec(ctx, query, device.ID, device.OrganizationID, device.Hostname, device.IPAddress, device.Status, device.OSVersion, device.MachineGUID, device.PublicKey, device.AssignedUserEmail, device.LastSeenAt, device.CreatedAt, device.UpdatedAt)
		return err
	})
}

func (r *pgDeviceRepo) Update(ctx context.Context, device *domain.Device) error {
	device.UpdatedAt = time.Now().UTC()
	return r.s.withTenantTx(ctx, device.OrganizationID, func(tx pgx.Tx) error {
		query := `
			UPDATE devices
			SET hostname = $1, ip_address = $2, status = $3, os_version = $4, public_key = $5, assigned_user_email = $6, updated_at = $7
			WHERE id = $8 AND organization_id = $9
		`
		_, err := tx.Exec(ctx, query, device.Hostname, device.IPAddress, device.Status, device.OSVersion, device.PublicKey, device.AssignedUserEmail, device.UpdatedAt, device.ID, device.OrganizationID)
		return err
	})
}

func (r *pgDeviceRepo) UpdateStatusAndLastSeen(ctx context.Context, id uuid.UUID, status domain.DeviceStatus, lastSeen time.Time) error {
	query := `UPDATE devices SET status = $1, last_seen_at = $2, updated_at = CURRENT_TIMESTAMP WHERE id = $3`
	_, err := r.s.pool.Exec(ctx, query, status, lastSeen, id)
	return err
}

func (r *pgDeviceRepo) GetDiagnostics(ctx context.Context, orgID, deviceID uuid.UUID) (*domain.DeviceDiagnostics, error) {
	var diag domain.DeviceDiagnostics
	diag.DeviceID = deviceID
	diag.OrganizationID = orgID

	err := r.s.withTenantTx(ctx, orgID, func(tx pgx.Tx) error {
		query := `
			SELECT system_info, cpu_info, memory_info, storage_info, network_info, security_info, processes_info, services_info, event_logs, installed_software, updated_at
			FROM device_diagnostics
			WHERE device_id = $1 AND organization_id = $2
		`
		err := tx.QueryRow(ctx, query, deviceID, orgID).Scan(
			&diag.SystemInfo, &diag.CPUInfo, &diag.MemoryInfo, &diag.StorageInfo, &diag.NetworkInfo,
			&diag.SecurityInfo, &diag.ProcessesInfo, &diag.ServicesInfo, &diag.EventLogs, &diag.InstalledSoftware,
			&diag.UpdatedAt,
		)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return errors.New("diagnostics not found")
			}
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &diag, nil
}

func (r *pgDeviceRepo) UpsertDiagnostics(ctx context.Context, diag *domain.DeviceDiagnostics) error {
	return r.s.withTenantTx(ctx, diag.OrganizationID, func(tx pgx.Tx) error {
		query := `
			INSERT INTO device_diagnostics (device_id, organization_id, system_info, cpu_info, memory_info, storage_info, network_info, security_info, processes_info, services_info, event_logs, installed_software, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, CURRENT_TIMESTAMP)
			ON CONFLICT (device_id) DO UPDATE SET
				system_info = EXCLUDED.system_info,
				cpu_info = EXCLUDED.cpu_info,
				memory_info = EXCLUDED.memory_info,
				storage_info = EXCLUDED.storage_info,
				network_info = EXCLUDED.network_info,
				security_info = EXCLUDED.security_info,
				processes_info = EXCLUDED.processes_info,
				services_info = EXCLUDED.services_info,
				event_logs = EXCLUDED.event_logs,
				installed_software = EXCLUDED.installed_software,
				updated_at = CURRENT_TIMESTAMP
		`
		_, err := tx.Exec(ctx, query, diag.DeviceID, diag.OrganizationID, diag.SystemInfo, diag.CPUInfo, diag.MemoryInfo, diag.StorageInfo, diag.NetworkInfo, diag.SecurityInfo, diag.ProcessesInfo, diag.ServicesInfo, diag.EventLogs, diag.InstalledSoftware)
		return err
	})
}

func (r *pgDeviceRepo) CreateEnrollmentToken(ctx context.Context, token *domain.EnrollmentToken) error {
	if token.ID == uuid.Nil {
		token.ID = uuid.New()
	}
	token.CreatedAt = time.Now().UTC()
	query := `
		INSERT INTO enrollment_tokens (id, organization_id, token, expires_at, max_uses, uses_count, created_by, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.s.pool.Exec(ctx, query, token.ID, token.OrganizationID, token.Token, token.ExpiresAt, token.MaxUses, token.UsesCount, token.CreatedBy, token.CreatedAt)
	return err
}

func (r *pgDeviceRepo) GetEnrollmentToken(ctx context.Context, tokenStr string) (*domain.EnrollmentToken, error) {
	query := `SELECT id, organization_id, token, expires_at, max_uses, uses_count, created_by, created_at FROM enrollment_tokens WHERE token = $1`
	var t domain.EnrollmentToken
	err := r.s.pool.QueryRow(ctx, query, tokenStr).Scan(&t.ID, &t.OrganizationID, &t.Token, &t.ExpiresAt, &t.MaxUses, &t.UsesCount, &t.CreatedBy, &t.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("enrollment token not found")
		}
		return nil, err
	}
	return &t, nil
}

func (r *pgDeviceRepo) UseEnrollmentToken(ctx context.Context, tokenStr string) error {
	query := `UPDATE enrollment_tokens SET uses_count = uses_count + 1 WHERE token = $1`
	_, err := r.s.pool.Exec(ctx, query, tokenStr)
	return err
}

// pgMetricRepo
type pgMetricRepo struct{ s *PostgresStorage }

func (r *pgMetricRepo) Insert(ctx context.Context, m *domain.Metric) error {
	return r.s.withTenantTx(ctx, m.OrganizationID, func(tx pgx.Tx) error {
		query := `
			INSERT INTO metrics (organization_id, device_id, cpu_usage, ram_usage, disk_usage, network_in, network_out, uptime, timestamp)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		`
		if m.Timestamp.IsZero() {
			m.Timestamp = time.Now().UTC()
		}
		_, err := tx.Exec(ctx, query, m.OrganizationID, m.DeviceID, m.CPUUsage, m.RAMUsage, m.DiskUsage, m.NetworkIn, m.NetworkOut, m.Uptime, m.Timestamp)
		return err
	})
}

func (r *pgMetricRepo) GetLatestByDevice(ctx context.Context, orgID, deviceID uuid.UUID) (*domain.Metric, error) {
	var m domain.Metric
	err := r.s.withTenantTx(ctx, orgID, func(tx pgx.Tx) error {
		query := `
			SELECT id, organization_id, device_id, cpu_usage, ram_usage, disk_usage, network_in, network_out, uptime, timestamp
			FROM metrics
			WHERE device_id = $1 AND organization_id = $2
			ORDER BY timestamp DESC
			LIMIT 1
		`
		err := tx.QueryRow(ctx, query, deviceID, orgID).Scan(&m.ID, &m.OrganizationID, &m.DeviceID, &m.CPUUsage, &m.RAMUsage, &m.DiskUsage, &m.NetworkIn, &m.NetworkOut, &m.Uptime, &m.Timestamp)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return errors.New("no metrics found")
			}
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (r *pgMetricRepo) ListByDevice(ctx context.Context, orgID, deviceID uuid.UUID, limit int) ([]*domain.Metric, error) {
	var metrics []*domain.Metric
	if limit <= 0 {
		limit = 100
	}
	err := r.s.withTenantTx(ctx, orgID, func(tx pgx.Tx) error {
		query := `
			SELECT id, organization_id, device_id, cpu_usage, ram_usage, disk_usage, network_in, network_out, uptime, timestamp
			FROM metrics
			WHERE device_id = $1 AND organization_id = $2
			ORDER BY timestamp DESC
			LIMIT $3
		`
		rows, err := tx.Query(ctx, query, deviceID, orgID, limit)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var m domain.Metric
			if err := rows.Scan(&m.ID, &m.OrganizationID, &m.DeviceID, &m.CPUUsage, &m.RAMUsage, &m.DiskUsage, &m.NetworkIn, &m.NetworkOut, &m.Uptime, &m.Timestamp); err != nil {
				return err
			}
			metrics = append(metrics, &m)
		}
		return rows.Err()
	})
	return metrics, err
}

// pgAlertRepo
type pgAlertRepo struct{ s *PostgresStorage }

func (r *pgAlertRepo) List(ctx context.Context, orgID uuid.UUID, activeOnly bool) ([]*domain.Alert, error) {
	var alerts []*domain.Alert
	err := r.s.withTenantTx(ctx, orgID, func(tx pgx.Tx) error {
		var query string
		if activeOnly {
			query = `
				SELECT a.id, a.organization_id, a.device_id, a.rule_id, COALESCE(r.name, 'Threshold Alert'), a.message, a.severity, a.resolved, a.resolved_at, a.created_at
				FROM alerts a
				LEFT JOIN rules r ON a.rule_id = r.id
				WHERE a.organization_id = $1 AND a.resolved = FALSE
				ORDER BY a.created_at DESC
			`
		} else {
			query = `
				SELECT a.id, a.organization_id, a.device_id, a.rule_id, COALESCE(r.name, 'Threshold Alert'), a.message, a.severity, a.resolved, a.resolved_at, a.created_at
				FROM alerts a
				LEFT JOIN rules r ON a.rule_id = r.id
				WHERE a.organization_id = $1
				ORDER BY a.created_at DESC
			`
		}
		rows, err := tx.Query(ctx, query, orgID)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var a domain.Alert
			if err := rows.Scan(&a.ID, &a.OrganizationID, &a.DeviceID, &a.RuleID, &a.RuleName, &a.Message, &a.Severity, &a.Resolved, &a.ResolvedAt, &a.CreatedAt); err != nil {
				return err
			}
			alerts = append(alerts, &a)
		}
		return rows.Err()
	})
	return alerts, err
}

func (r *pgAlertRepo) GetByID(ctx context.Context, orgID, id uuid.UUID) (*domain.Alert, error) {
	var a domain.Alert
	err := r.s.withTenantTx(ctx, orgID, func(tx pgx.Tx) error {
		query := `
			SELECT a.id, a.organization_id, a.device_id, a.rule_id, COALESCE(r.name, 'Threshold Alert'), a.message, a.severity, a.resolved, a.resolved_at, a.created_at
			FROM alerts a
			LEFT JOIN rules r ON a.rule_id = r.id
			WHERE a.id = $1 AND a.organization_id = $2
		`
		return tx.QueryRow(ctx, query, id, orgID).Scan(&a.ID, &a.OrganizationID, &a.DeviceID, &a.RuleID, &a.RuleName, &a.Message, &a.Severity, &a.Resolved, &a.ResolvedAt, &a.CreatedAt)
	})
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *pgAlertRepo) Create(ctx context.Context, alert *domain.Alert) error {
	if alert.ID == uuid.Nil {
		alert.ID = uuid.New()
	}
	alert.CreatedAt = time.Now().UTC()
	return r.s.withTenantTx(ctx, alert.OrganizationID, func(tx pgx.Tx) error {
		query := `
			INSERT INTO alerts (id, organization_id, device_id, rule_id, message, severity, resolved, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`
		_, err := tx.Exec(ctx, query, alert.ID, alert.OrganizationID, alert.DeviceID, alert.RuleID, alert.Message, alert.Severity, alert.Resolved, alert.CreatedAt)
		return err
	})
}

func (r *pgAlertRepo) Resolve(ctx context.Context, orgID, id uuid.UUID) error {
	return r.s.withTenantTx(ctx, orgID, func(tx pgx.Tx) error {
		query := `UPDATE alerts SET resolved = TRUE, resolved_at = CURRENT_TIMESTAMP WHERE id = $1 AND organization_id = $2`
		_, err := tx.Exec(ctx, query, id, orgID)
		return err
	})
}

func (r *pgAlertRepo) ListActiveByDevice(ctx context.Context, orgID, deviceID uuid.UUID) ([]*domain.Alert, error) {
	var alerts []*domain.Alert
	err := r.s.withTenantTx(ctx, orgID, func(tx pgx.Tx) error {
		query := `
			SELECT a.id, a.organization_id, a.device_id, a.rule_id, COALESCE(r.name, 'Threshold Alert'), a.message, a.severity, a.resolved, a.resolved_at, a.created_at
			FROM alerts a
			LEFT JOIN rules r ON a.rule_id = r.id
			WHERE a.organization_id = $1 AND a.device_id = $2 AND a.resolved = FALSE
			ORDER BY a.created_at DESC
		`
		rows, err := tx.Query(ctx, query, orgID, deviceID)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var a domain.Alert
			if err := rows.Scan(&a.ID, &a.OrganizationID, &a.DeviceID, &a.RuleID, &a.RuleName, &a.Message, &a.Severity, &a.Resolved, &a.ResolvedAt, &a.CreatedAt); err != nil {
				return err
			}
			alerts = append(alerts, &a)
		}
		return rows.Err()
	})
	return alerts, err
}

// pgIncidentRepo
type pgIncidentRepo struct{ s *PostgresStorage }

func (r *pgIncidentRepo) List(ctx context.Context, orgID uuid.UUID, status string) ([]*domain.Incident, error) {
	var incidents []*domain.Incident
	err := r.s.withTenantTx(ctx, orgID, func(tx pgx.Tx) error {
		var query string
		var args []any
		if status != "" {
			query = `
				SELECT i.id, i.organization_id, i.title, i.severity, i.status, i.assigned_user_id, u.full_name, i.resolved_at, i.created_at, i.updated_at,
					(SELECT COUNT(*) FROM incident_alerts ia WHERE ia.incident_id = i.id) as alerts_count
				FROM incidents i
				LEFT JOIN users u ON i.assigned_user_id = u.id
				WHERE i.organization_id = $1 AND i.status = $2
				ORDER BY i.created_at DESC
			`
			args = []any{orgID, status}
		} else {
			query = `
				SELECT i.id, i.organization_id, i.title, i.severity, i.status, i.assigned_user_id, u.full_name, i.resolved_at, i.created_at, i.updated_at,
					(SELECT COUNT(*) FROM incident_alerts ia WHERE ia.incident_id = i.id) as alerts_count
				FROM incidents i
				LEFT JOIN users u ON i.assigned_user_id = u.id
				WHERE i.organization_id = $1
				ORDER BY i.created_at DESC
			`
			args = []any{orgID}
		}

		rows, err := tx.Query(ctx, query, args...)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var inc domain.Incident
			var assignedName *string
			if err := rows.Scan(&inc.ID, &inc.OrganizationID, &inc.Title, &inc.Severity, &inc.Status, &inc.AssignedUserID, &assignedName, &inc.ResolvedAt, &inc.CreatedAt, &inc.UpdatedAt, &inc.AlertsCount); err != nil {
				return err
			}
			inc.AssignedTo = assignedName
			incidents = append(incidents, &inc)
		}
		return rows.Err()
	})
	return incidents, err
}

func (r *pgIncidentRepo) GetByID(ctx context.Context, orgID, id uuid.UUID) (*domain.Incident, error) {
	var inc domain.Incident
	err := r.s.withTenantTx(ctx, orgID, func(tx pgx.Tx) error {
		query := `
			SELECT i.id, i.organization_id, i.title, i.severity, i.status, i.assigned_user_id, u.full_name, i.resolved_at, i.created_at, i.updated_at,
				(SELECT COUNT(*) FROM incident_alerts ia WHERE ia.incident_id = i.id) as alerts_count
			FROM incidents i
			LEFT JOIN users u ON i.assigned_user_id = u.id
			WHERE i.id = $1 AND i.organization_id = $2
		`
		var assignedName *string
		err := tx.QueryRow(ctx, query, id, orgID).Scan(&inc.ID, &inc.OrganizationID, &inc.Title, &inc.Severity, &inc.Status, &inc.AssignedUserID, &assignedName, &inc.ResolvedAt, &inc.CreatedAt, &inc.UpdatedAt, &inc.AlertsCount)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return errors.New("incident not found")
			}
			return err
		}
		inc.AssignedTo = assignedName
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &inc, nil
}

func (r *pgIncidentRepo) Create(ctx context.Context, incident *domain.Incident) error {
	if incident.ID == uuid.Nil {
		incident.ID = uuid.New()
	}
	now := time.Now().UTC()
	incident.CreatedAt = now
	incident.UpdatedAt = now

	return r.s.withTenantTx(ctx, incident.OrganizationID, func(tx pgx.Tx) error {
		query := `
			INSERT INTO incidents (id, organization_id, title, severity, status, assigned_user_id, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`
		_, err := tx.Exec(ctx, query, incident.ID, incident.OrganizationID, incident.Title, incident.Severity, incident.Status, incident.AssignedUserID, incident.CreatedAt, incident.UpdatedAt)
		return err
	})
}

func (r *pgIncidentRepo) UpdateStatus(ctx context.Context, orgID, id uuid.UUID, status domain.IncidentStatus) error {
	return r.s.withTenantTx(ctx, orgID, func(tx pgx.Tx) error {
		var resolvedAt *time.Time
		if status == domain.IncidentStatusResolved || status == domain.IncidentStatusClosed {
			now := time.Now().UTC()
			resolvedAt = &now
		}
		query := `UPDATE incidents SET status = $1, resolved_at = $2, updated_at = CURRENT_TIMESTAMP WHERE id = $3 AND organization_id = $4`
		_, err := tx.Exec(ctx, query, status, resolvedAt, id, orgID)
		return err
	})
}

func (r *pgIncidentRepo) AddTimeline(ctx context.Context, item *domain.IncidentTimeline) error {
	if item.ID == uuid.Nil {
		item.ID = uuid.New()
	}
	item.CreatedAt = time.Now().UTC()
	query := `INSERT INTO incident_timeline (id, incident_id, author_user_id, event_type, content, created_at) VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.s.pool.Exec(ctx, query, item.ID, item.IncidentID, item.AuthorUserID, item.EventType, item.Content, item.CreatedAt)
	return err
}

func (r *pgIncidentRepo) GetTimeline(ctx context.Context, incidentID uuid.UUID) ([]*domain.IncidentTimeline, error) {
	query := `
		SELECT t.id, t.incident_id, t.author_user_id, COALESCE(u.full_name, 'System'), t.event_type, t.content, t.created_at
		FROM incident_timeline t
		LEFT JOIN users u ON t.author_user_id = u.id
		WHERE t.incident_id = $1
		ORDER BY t.created_at ASC
	`
	rows, err := r.s.pool.Query(ctx, query, incidentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*domain.IncidentTimeline
	for rows.Next() {
		var t domain.IncidentTimeline
		if err := rows.Scan(&t.ID, &t.IncidentID, &t.AuthorUserID, &t.AuthorName, &t.EventType, &t.Content, &t.CreatedAt); err != nil {
			return nil, err
		}
		list = append(list, &t)
	}
	return list, rows.Err()
}

func (r *pgIncidentRepo) AssociateAlert(ctx context.Context, incidentID, alertID uuid.UUID) error {
	query := `INSERT INTO incident_alerts (incident_id, alert_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`
	_, err := r.s.pool.Exec(ctx, query, incidentID, alertID)
	return err
}

func (r *pgIncidentRepo) GetAlerts(ctx context.Context, incidentID uuid.UUID) ([]*domain.Alert, error) {
	query := `
		SELECT a.id, a.organization_id, a.device_id, a.rule_id, COALESCE(r.name, 'Threshold Alert'), a.message, a.severity, a.resolved, a.resolved_at, a.created_at
		FROM alerts a
		JOIN incident_alerts ia ON a.id = ia.alert_id
		LEFT JOIN rules r ON a.rule_id = r.id
		WHERE ia.incident_id = $1
		ORDER BY a.created_at ASC
	`
	rows, err := r.s.pool.Query(ctx, query, incidentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var alerts []*domain.Alert
	for rows.Next() {
		var a domain.Alert
		if err := rows.Scan(&a.ID, &a.OrganizationID, &a.DeviceID, &a.RuleID, &a.RuleName, &a.Message, &a.Severity, &a.Resolved, &a.ResolvedAt, &a.CreatedAt); err != nil {
			return nil, err
		}
		alerts = append(alerts, &a)
	}
	return alerts, rows.Err()
}

func (r *pgIncidentRepo) SaveAIReport(ctx context.Context, report *domain.AIAdvisoryReport) error {
	if report.ID == uuid.Nil {
		report.ID = uuid.New()
	}
	report.CreatedAt = time.Now().UTC()
	return r.s.withTenantTx(ctx, report.OrganizationID, func(tx pgx.Tx) error {
		query := `
			INSERT INTO ai_advisory_reports (id, organization_id, incident_id, summary, hypotheses, recommendations, provider, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`
		_, err := tx.Exec(ctx, query, report.ID, report.OrganizationID, report.IncidentID, report.Summary, report.Hypotheses, report.Recommendations, report.Provider, report.CreatedAt)
		return err
	})
}

// pgRemediationRepo
type pgRemediationRepo struct{ s *PostgresStorage }

func (r *pgRemediationRepo) CreateJob(ctx context.Context, job *domain.RemediationJob) error {
	if job.ID == uuid.Nil {
		job.ID = uuid.New()
	}
	now := time.Now().UTC()
	job.CreatedAt = now
	job.UpdatedAt = now

	return r.s.withTenantTx(ctx, job.OrganizationID, func(tx pgx.Tx) error {
		query := `
			INSERT INTO remediation_jobs (id, organization_id, device_id, action_id, parameters, status, requested_by, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		`
		_, err := tx.Exec(ctx, query, job.ID, job.OrganizationID, job.DeviceID, job.ActionID, job.Parameters, job.Status, job.RequestedBy, job.CreatedAt, job.UpdatedAt)
		return err
	})
}

func (r *pgRemediationRepo) GetJobByID(ctx context.Context, orgID, id uuid.UUID) (*domain.RemediationJob, error) {
	var job domain.RemediationJob
	err := r.s.withTenantTx(ctx, orgID, func(tx pgx.Tx) error {
		query := `
			SELECT j.id, j.organization_id, j.device_id, j.action_id, j.parameters, j.status, j.requested_by, COALESCE(u1.full_name, ''), j.approved_by, COALESCE(u2.full_name, ''), COALESCE(j.execution_output, ''), j.verification_passed, j.created_at, j.updated_at
			FROM remediation_jobs j
			LEFT JOIN users u1 ON j.requested_by = u1.id
			LEFT JOIN users u2 ON j.approved_by = u2.id
			WHERE j.id = $1 AND j.organization_id = $2
		`
		err := tx.QueryRow(ctx, query, id, orgID).Scan(
			&job.ID, &job.OrganizationID, &job.DeviceID, &job.ActionID, &job.Parameters,
			&job.Status, &job.RequestedBy, &job.Requester, &job.ApprovedBy, &job.Approver,
			&job.ExecutionOutput, &job.VerificationPassed, &job.CreatedAt, &job.UpdatedAt,
		)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return errors.New("remediation job not found")
			}
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &job, nil
}

func (r *pgRemediationRepo) ListJobs(ctx context.Context, orgID uuid.UUID, limit int) ([]*domain.RemediationJob, error) {
	var jobs []*domain.RemediationJob
	if limit <= 0 {
		limit = 50
	}
	err := r.s.withTenantTx(ctx, orgID, func(tx pgx.Tx) error {
		query := `
			SELECT j.id, j.organization_id, j.device_id, j.action_id, j.parameters, j.status, j.requested_by, COALESCE(u1.full_name, ''), j.approved_by, COALESCE(u2.full_name, ''), COALESCE(j.execution_output, ''), j.verification_passed, j.created_at, j.updated_at
			FROM remediation_jobs j
			LEFT JOIN users u1 ON j.requested_by = u1.id
			LEFT JOIN users u2 ON j.approved_by = u2.id
			WHERE j.organization_id = $1
			ORDER BY j.created_at DESC
			LIMIT $2
		`
		rows, err := tx.Query(ctx, query, orgID, limit)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var job domain.RemediationJob
			if err := rows.Scan(&job.ID, &job.OrganizationID, &job.DeviceID, &job.ActionID, &job.Parameters, &job.Status, &job.RequestedBy, &job.Requester, &job.ApprovedBy, &job.Approver, &job.ExecutionOutput, &job.VerificationPassed, &job.CreatedAt, &job.UpdatedAt); err != nil {
				return err
			}
			jobs = append(jobs, &job)
		}
		return rows.Err()
	})
	return jobs, err
}

func (r *pgRemediationRepo) UpdateJobStatus(ctx context.Context, orgID, id uuid.UUID, status domain.RemediationStatus, output string, passed *bool) error {
	return r.s.withTenantTx(ctx, orgID, func(tx pgx.Tx) error {
		query := `
			UPDATE remediation_jobs
			SET status = $1, execution_output = $2, verification_passed = $3, updated_at = CURRENT_TIMESTAMP
			WHERE id = $4 AND organization_id = $5
		`
		_, err := tx.Exec(ctx, query, status, output, passed, id, orgID)
		return err
	})
}

func (r *pgRemediationRepo) ApproveJob(ctx context.Context, orgID, id, approverID uuid.UUID) error {
	return r.s.withTenantTx(ctx, orgID, func(tx pgx.Tx) error {
		query := `
			UPDATE remediation_jobs
			SET status = 'approved', approved_by = $1, updated_at = CURRENT_TIMESTAMP
			WHERE id = $2 AND organization_id = $3 AND status = 'pending_approval'
		`
		res, err := tx.Exec(ctx, query, approverID, id, orgID)
		if err != nil {
			return err
		}
		if res.RowsAffected() == 0 {
			return errors.New("job not pending approval or not found")
		}
		return nil
	})
}

// pgRemoteSessionRepo
type pgRemoteSessionRepo struct{ s *PostgresStorage }

func (r *pgRemoteSessionRepo) Create(ctx context.Context, session *domain.RemoteSession) error {
	if session.SessionID == uuid.Nil {
		session.SessionID = uuid.New()
	}
	session.CreatedAt = time.Now().UTC()
	return r.s.withTenantTx(ctx, session.OrganizationID, func(tx pgx.Tx) error {
		query := `
			INSERT INTO remote_sessions (id, organization_id, device_id, operator_user_id, status, permissions, reason, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`
		_, err := tx.Exec(ctx, query, session.SessionID, session.OrganizationID, session.DeviceID, session.OperatorID, session.Status, session.Permissions, session.Reason, session.CreatedAt)
		return err
	})
}

func (r *pgRemoteSessionRepo) GetByID(ctx context.Context, orgID, id uuid.UUID) (*domain.RemoteSession, error) {
	var s domain.RemoteSession
	err := r.s.withTenantTx(ctx, orgID, func(tx pgx.Tx) error {
		query := `
			SELECT rs.id, rs.organization_id, rs.device_id, rs.operator_user_id, COALESCE(u.full_name, ''), rs.status, rs.permissions, rs.reason, rs.consent_granted_at, rs.started_at, rs.ended_at, rs.termination_reason, rs.created_at
			FROM remote_sessions rs
			LEFT JOIN users u ON rs.operator_user_id = u.id
			WHERE rs.id = $1 AND rs.organization_id = $2
		`
		var termReason *string
		err := tx.QueryRow(ctx, query, id, orgID).Scan(
			&s.SessionID, &s.OrganizationID, &s.DeviceID, &s.OperatorID, &s.OperatorName,
			&s.Status, &s.Permissions, &s.Reason, &s.ConsentGrantedAt, &s.StartedAt, &s.EndedAt,
			&termReason, &s.CreatedAt,
		)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return errors.New("remote session not found")
			}
			return err
		}
		if termReason != nil {
			s.TerminationReason = *termReason
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *pgRemoteSessionRepo) UpdateStatus(ctx context.Context, orgID, id uuid.UUID, status domain.RemoteSessionStatus, reason string) error {
	return r.s.withTenantTx(ctx, orgID, func(tx pgx.Tx) error {
		now := time.Now().UTC()
		var consentAt, startedAt, endedAt *time.Time
		if status == domain.RemoteSessionActive {
			consentAt = &now
			startedAt = &now
		} else if status == domain.RemoteSessionTerminated || status == domain.RemoteSessionRejected {
			endedAt = &now
		}

		query := `
			UPDATE remote_sessions
			SET status = $1, termination_reason = NULLIF($2, ''),
				consent_granted_at = COALESCE(consent_granted_at, $3),
				started_at = COALESCE(started_at, $4),
				ended_at = COALESCE(ended_at, $5)
			WHERE id = $6 AND organization_id = $7
		`
		_, err := tx.Exec(ctx, query, status, reason, consentAt, startedAt, endedAt, id, orgID)
		return err
	})
}

func (r *pgRemoteSessionRepo) Terminate(ctx context.Context, orgID, id uuid.UUID, reason string) error {
	return r.UpdateStatus(ctx, orgID, id, domain.RemoteSessionTerminated, reason)
}

// pgAuditRepo
type pgAuditRepo struct{ s *PostgresStorage }

func (r *pgAuditRepo) Insert(ctx context.Context, log *domain.AuditLog) error {
	if log.ID == uuid.Nil {
		log.ID = uuid.New()
	}
	log.CreatedAt = time.Now().UTC()
	return r.s.withTenantTx(ctx, log.OrganizationID, func(tx pgx.Tx) error {
		query := `
			INSERT INTO audit_logs (id, organization_id, actor_type, actor_id, action, resource_type, resource_id, metadata, ip_address, created_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		`
		var meta any
		if len(log.Metadata) > 0 {
			meta = log.Metadata
		}
		_, err := tx.Exec(ctx, query, log.ID, log.OrganizationID, log.ActorType, log.ActorID, log.Action, log.ResourceType, log.ResourceID, meta, log.IPAddress, log.CreatedAt)
		return err
	})
}

func (r *pgAuditRepo) List(ctx context.Context, orgID uuid.UUID, limit int) ([]*domain.AuditLog, error) {
	var logs []*domain.AuditLog
	if limit <= 0 {
		limit = 100
	}
	err := r.s.withTenantTx(ctx, orgID, func(tx pgx.Tx) error {
		query := `
			SELECT id, organization_id, actor_type, actor_id, action, resource_type, resource_id, metadata, ip_address, created_at
			FROM audit_logs
			WHERE organization_id = $1
			ORDER BY created_at DESC
			LIMIT $2
		`
		rows, err := tx.Query(ctx, query, orgID, limit)
		if err != nil {
			return err
		}
		defer rows.Close()

		for rows.Next() {
			var l domain.AuditLog
			var meta []byte
			var ip *string
			if err := rows.Scan(&l.ID, &l.OrganizationID, &l.ActorType, &l.ActorID, &l.Action, &l.ResourceType, &l.ResourceID, &meta, &ip, &l.CreatedAt); err != nil {
				return err
			}
			if len(meta) > 0 {
				l.Metadata = json.RawMessage(meta)
			}
			if ip != nil {
				l.IPAddress = *ip
			}
			logs = append(logs, &l)
		}
		return rows.Err()
	})
	return logs, err
}
