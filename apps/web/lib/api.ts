export interface User {
  id: string;
  email: string;
  full_name: string;
  role: 'owner_admin' | 'it_admin' | 'it_operator' | 'readonly_viewer';
}

export interface Device {
  id: string;
  hostname: string;
  ip_address: string;
  status: 'online' | 'degraded' | 'offline';
  os_version: string;
  assigned_user_email?: string;
  last_seen_at?: string;
}

export interface MetricSnapshot {
  cpu_usage: number;
  ram_usage: number;
  disk_usage: number;
  network_in_kb: number;
  network_out_kb: number;
  uptime_seconds: number;
}

export interface Incident {
  id: string;
  title: string;
  severity: 'low' | 'medium' | 'high' | 'critical';
  status: 'open' | 'investigating' | 'resolved' | 'closed';
  created_at: string;
  alerts_count: number;
}

export interface Alert {
  id: string;
  device_id: string;
  device_hostname?: string;
  rule_id: string;
  rule_name?: string;
  severity: 'critical' | 'warning' | 'info';
  metric_name: string;
  current_value: number;
  threshold_value: number;
  triggered_at: string;
  acknowledged: boolean;
  acknowledged_by?: string;
  resolved: boolean;
  resolved_at?: string;
}

export interface AIAdvisory {
  incident_id: string;
  summary: string;
  hypotheses: Array<{
    rank: number;
    cause: string;
    confidence: number;
    supporting_evidence: string[];
    contradicting_evidence: string[];
  }>;
  recommended_verification_steps: string[];
  suggested_remediations: Array<{
    action_id: string;
    risk_level: string;
    expected_outcome: string;
  }>;
}

export interface RemediationAction {
  id: string;
  title: string;
  risk_level: 'read_only' | 'low_risk' | 'destructive';
  requires_approval: boolean;
  timeout_seconds: number;
}

export interface RemediationJob {
  id: string;
  action_id: string;
  device_id: string;
  status: 'proposed' | 'approved' | 'dispatched' | 'completed' | 'failed' | 'rejected';
  proposed_by: string;
  approved_by?: string;
  created_at: string;
  completed_at?: string;
  output?: string;
}

export interface RemoteSession {
  session_id: string;
  device_id: string;
  operator_id: string;
  status: 'awaiting_consent' | 'active' | 'rejected' | 'terminated' | 'expired';
  permissions: string[];
  reason: string;
}

class ApiClient {
  private getHeaders(): HeadersInit {
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
    };
    if (typeof window !== 'undefined') {
      const token = localStorage.getItem('monitopro_token');
      if (token) {
        headers['Authorization'] = `Bearer ${token}`;
      }
      const tenant = localStorage.getItem('monitopro_active_tenant');
      if (tenant) {
        headers['X-Tenant-Subdomain'] = tenant;
      }
    }
    return headers;
  }

  async getHealth() {
    const res = await fetch('/api/health');
    if (!res.ok) throw new Error('Health check failed');
    return res.json();
  }

  async login(email: string, password: string, subdomain?: string) {
    const headers: Record<string, string> = { 'Content-Type': 'application/json' };
    if (subdomain) {
      headers['X-Tenant-Subdomain'] = subdomain;
    }
    const res = await fetch('/api/v1/auth/login', {
      method: 'POST',
      headers,
      body: JSON.stringify({ email, password }),
    });
    if (!res.ok) {
      const err = await res.json().catch(() => ({}));
      throw new Error(err.message || 'Authentication failed');
    }
    return res.json();
  }

  async listDevices(): Promise<{ items: Device[]; total: number }> {
    const res = await fetch('/api/v1/devices', { headers: this.getHeaders() });
    if (!res.ok) throw new Error('Failed to fetch devices');
    return res.json();
  }

  async getDevice(id: string): Promise<Device & { latest_metrics?: MetricSnapshot }> {
    const res = await fetch(`/api/v1/devices/${id}`, { headers: this.getHeaders() });
    if (!res.ok) throw new Error('Failed to fetch device details');
    return res.json();
  }

  async getDiagnostics(id: string): Promise<any> {
    const res = await fetch(`/api/v1/devices/${id}/diagnostics`, { headers: this.getHeaders() });
    if (!res.ok) return null;
    return res.json();
  }

  async listAlerts(activeOnly = false): Promise<Alert[]> {
    const url = `/api/v1/alerts${activeOnly ? '?active=true' : ''}`;
    const res = await fetch(url, { headers: this.getHeaders() });
    if (!res.ok) throw new Error('Failed to fetch alerts');
    const data = await res.json();
    return Array.isArray(data) ? data : [];
  }

  async listIncidents(): Promise<Incident[]> {
    const res = await fetch('/api/v1/incidents', { headers: this.getHeaders() });
    if (!res.ok) throw new Error('Failed to fetch incidents');
    const data = await res.json();
    return Array.isArray(data) ? data : [];
  }

  async getAIAdvisory(incidentId: string): Promise<AIAdvisory> {
    const res = await fetch(`/api/v1/incidents/${incidentId}/ai-analysis`, {
      method: 'POST',
      headers: this.getHeaders(),
    });
    if (!res.ok) throw new Error('Failed to generate AI advisory');
    return res.json();
  }

  async listRemediationActions(): Promise<RemediationAction[]> {
    const res = await fetch('/api/v1/remediation/actions', { headers: this.getHeaders() });
    if (!res.ok) throw new Error('Failed to fetch remediation catalogue');
    return res.json();
  }

  async listRemediationJobs(): Promise<RemediationJob[]> {
    const res = await fetch('/api/v1/remediation/jobs', { headers: this.getHeaders() });
    if (!res.ok) throw new Error('Failed to fetch remediation jobs');
    const data = await res.json();
    return Array.isArray(data) ? data : [];
  }

  async proposeRemediationJob(actionId: string, deviceId: string, parameters?: Record<string, any>): Promise<RemediationJob> {
    const res = await fetch('/api/v1/remediation/jobs', {
      method: 'POST',
      headers: this.getHeaders(),
      body: JSON.stringify({
        action_id: actionId,
        device_id: deviceId,
        parameters: parameters || {},
      }),
    });
    if (!res.ok) throw new Error('Failed to propose remediation job');
    return res.json();
  }

  async approveRemediationJob(jobId: string): Promise<RemediationJob> {
    const res = await fetch(`/api/v1/remediation/jobs/${jobId}/approve`, {
      method: 'POST',
      headers: this.getHeaders(),
    });
    if (!res.ok) throw new Error('Failed to approve remediation job');
    return res.json();
  }

  async requestRemoteSession(deviceId: string, permissions: string[], reason: string): Promise<RemoteSession> {
    const res = await fetch('/api/v1/remote-support/sessions', {
      method: 'POST',
      headers: this.getHeaders(),
      body: JSON.stringify({
        device_id: deviceId,
        permissions_requested: permissions,
        reason,
      }),
    });
    if (!res.ok) throw new Error('Failed to request remote session');
    return res.json();
  }

  async terminateRemoteSession(sessionId: string): Promise<void> {
    await fetch(`/api/v1/remote-support/sessions/${sessionId}/terminate`, {
      method: 'POST',
      headers: this.getHeaders(),
    });
  }

  async createEnrollmentToken(assignedUserId?: number | string | null): Promise<{ token: string; expires_at: string; installer_command: string }> {
    const res = await fetch('/api/v1/devices/enrollment-tokens', {
      method: 'POST',
      headers: this.getHeaders(),
      body: JSON.stringify({
        expires_in_hours: 24,
        max_uses: 1,
        assigned_user_id: assignedUserId || null,
      }),
    });
    if (!res.ok) {
      const err = await res.json().catch(() => ({}));
      throw new Error(err.message || 'Failed to generate enrollment token from backend');
    }
    return res.json();
  }

  async getDeviceUsers(): Promise<Array<{ id: number | string; full_name: string; email: string; department?: string }>> {
    const res = await fetch('/api/v1/users', { headers: this.getHeaders() });
    if (!res.ok) return [];
    const data = await res.json();
    return Array.isArray(data) ? data : (data.items || []);
  }

  async createDeviceUser(fullName: string, email: string, department?: string): Promise<{ id: number | string; full_name: string; email: string }> {
    const res = await fetch('/api/v1/users', {
      method: 'POST',
      headers: this.getHeaders(),
      body: JSON.stringify({ full_name: fullName, email, department }),
    });
    if (!res.ok) throw new Error('Failed to create device user');
    return res.json();
  }
}

export const api = new ApiClient();
