-- +goose Up
-- Enable Row Level Security (RLS) across all tenant-scoped tables

ALTER TABLE devices ENABLE ROW LEVEL SECURITY;
ALTER TABLE metrics ENABLE ROW LEVEL SECURITY;
ALTER TABLE alerts ENABLE ROW LEVEL SECURITY;
ALTER TABLE incidents ENABLE ROW LEVEL SECURITY;
ALTER TABLE remediation_jobs ENABLE ROW LEVEL SECURITY;
ALTER TABLE remote_sessions ENABLE ROW LEVEL SECURITY;
ALTER TABLE audit_logs ENABLE ROW LEVEL SECURITY;
ALTER TABLE device_diagnostics ENABLE ROW LEVEL SECURITY;
ALTER TABLE ai_advisory_reports ENABLE ROW LEVEL SECURITY;

-- Tenant isolation policies: check against the session variable 'app.current_tenant_id'
-- If app.current_tenant_id is set to 'bypass', allow platform admin queries

CREATE POLICY tenant_isolation_devices ON devices
    USING (organization_id = NULLIF(current_setting('app.current_tenant_id', true), '')::UUID OR current_setting('app.current_tenant_id', true) = 'bypass')
    WITH CHECK (organization_id = NULLIF(current_setting('app.current_tenant_id', true), '')::UUID OR current_setting('app.current_tenant_id', true) = 'bypass');

CREATE POLICY tenant_isolation_metrics ON metrics
    USING (organization_id = NULLIF(current_setting('app.current_tenant_id', true), '')::UUID OR current_setting('app.current_tenant_id', true) = 'bypass')
    WITH CHECK (organization_id = NULLIF(current_setting('app.current_tenant_id', true), '')::UUID OR current_setting('app.current_tenant_id', true) = 'bypass');

CREATE POLICY tenant_isolation_alerts ON alerts
    USING (organization_id = NULLIF(current_setting('app.current_tenant_id', true), '')::UUID OR current_setting('app.current_tenant_id', true) = 'bypass')
    WITH CHECK (organization_id = NULLIF(current_setting('app.current_tenant_id', true), '')::UUID OR current_setting('app.current_tenant_id', true) = 'bypass');

CREATE POLICY tenant_isolation_incidents ON incidents
    USING (organization_id = NULLIF(current_setting('app.current_tenant_id', true), '')::UUID OR current_setting('app.current_tenant_id', true) = 'bypass')
    WITH CHECK (organization_id = NULLIF(current_setting('app.current_tenant_id', true), '')::UUID OR current_setting('app.current_tenant_id', true) = 'bypass');

CREATE POLICY tenant_isolation_remediation_jobs ON remediation_jobs
    USING (organization_id = NULLIF(current_setting('app.current_tenant_id', true), '')::UUID OR current_setting('app.current_tenant_id', true) = 'bypass')
    WITH CHECK (organization_id = NULLIF(current_setting('app.current_tenant_id', true), '')::UUID OR current_setting('app.current_tenant_id', true) = 'bypass');

CREATE POLICY tenant_isolation_remote_sessions ON remote_sessions
    USING (organization_id = NULLIF(current_setting('app.current_tenant_id', true), '')::UUID OR current_setting('app.current_tenant_id', true) = 'bypass')
    WITH CHECK (organization_id = NULLIF(current_setting('app.current_tenant_id', true), '')::UUID OR current_setting('app.current_tenant_id', true) = 'bypass');

CREATE POLICY tenant_isolation_audit_logs ON audit_logs
    USING (organization_id = NULLIF(current_setting('app.current_tenant_id', true), '')::UUID OR current_setting('app.current_tenant_id', true) = 'bypass')
    WITH CHECK (organization_id = NULLIF(current_setting('app.current_tenant_id', true), '')::UUID OR current_setting('app.current_tenant_id', true) = 'bypass');

CREATE POLICY tenant_isolation_diagnostics ON device_diagnostics
    USING (organization_id = NULLIF(current_setting('app.current_tenant_id', true), '')::UUID OR current_setting('app.current_tenant_id', true) = 'bypass')
    WITH CHECK (organization_id = NULLIF(current_setting('app.current_tenant_id', true), '')::UUID OR current_setting('app.current_tenant_id', true) = 'bypass');

CREATE POLICY tenant_isolation_ai_reports ON ai_advisory_reports
    USING (organization_id = NULLIF(current_setting('app.current_tenant_id', true), '')::UUID OR current_setting('app.current_tenant_id', true) = 'bypass')
    WITH CHECK (organization_id = NULLIF(current_setting('app.current_tenant_id', true), '')::UUID OR current_setting('app.current_tenant_id', true) = 'bypass');

-- +goose Down
DROP POLICY IF EXISTS tenant_isolation_ai_reports ON ai_advisory_reports;
DROP POLICY IF EXISTS tenant_isolation_diagnostics ON device_diagnostics;
DROP POLICY IF EXISTS tenant_isolation_audit_logs ON audit_logs;
DROP POLICY IF EXISTS tenant_isolation_remote_sessions ON remote_sessions;
DROP POLICY IF EXISTS tenant_isolation_remediation_jobs ON remediation_jobs;
DROP POLICY IF EXISTS tenant_isolation_incidents ON incidents;
DROP POLICY IF EXISTS tenant_isolation_alerts ON alerts;
DROP POLICY IF EXISTS tenant_isolation_metrics ON metrics;
DROP POLICY IF EXISTS tenant_isolation_devices ON devices;
