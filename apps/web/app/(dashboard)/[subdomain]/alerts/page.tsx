"use client";

import { useState, useEffect } from "react";
import { useParams, useRouter } from "next/navigation";
import Link from "next/link";
import { api } from "@/lib/api";
import { DashboardSidebar } from "@/components/Sidebar";
import { AmbientBackground } from "@/components/AmbientBackground";
import {
  AlertTriangle,
  Flame,
  CheckCircle2,
  Clock,
  Server,
  Filter,
  Search,
  Check,
  ArrowUpRight,
  RefreshCw,
  Bell,
  AlertCircle,
  X,
  ShieldAlert,
} from "lucide-react";

export interface AlertItem {
  id: string;
  deviceId: string;
  hostname: string;
  metric: string;
  currentValue: string;
  threshold: string;
  severity: "critical" | "warning" | "info";
  triggeredAt: string;
  elapsed: string;
  acknowledged: boolean;
  acknowledgedBy?: string;
  resolved: boolean;
}

export default function AlertsCenterPage() {
  const params = useParams();
  const router = useRouter();
  const subdomain = (params?.subdomain as string) || "demo";

  // Alerts state from live telemetry
  const [alerts, setAlerts] = useState<AlertItem[]>([]);
  const [loading, setLoading] = useState(true);

  const loadAlerts = async () => {
    setLoading(true);
    try {
      const raw = await api.listAlerts();
      const mapped: AlertItem[] = (raw || []).map((a) => ({
        id: a.id,
        deviceId: a.device_id,
        hostname: a.device_hostname || a.device_id.substring(0, 8),
        metric: a.metric_name || "Telemetry Metric",
        currentValue: `${a.current_value !== undefined ? a.current_value.toFixed(1) : "0.0"}`,
        threshold: `Threshold: ${a.threshold_value}`,
        severity: a.severity || "warning",
        triggeredAt: a.triggered_at ? new Date(a.triggered_at).toLocaleTimeString() : "Just now",
        elapsed: "Active",
        acknowledged: a.acknowledged || false,
        acknowledgedBy: a.acknowledged_by,
        resolved: a.resolved || false,
      }));
      setAlerts(mapped);
    } catch {
      setAlerts([]);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadAlerts();
  }, [subdomain]);

  const [filterSeverity, setFilterSeverity] = useState<"all" | "critical" | "warning" | "info" | "ack">("all");
  const [searchTerm, setSearchTerm] = useState("");
  const [isPromoting, setIsPromoting] = useState<AlertItem | null>(null);
  const [incidentTitle, setIncidentTitle] = useState("");
  const [toastMessage, setToastMessage] = useState<string | null>(null);

  const showToast = (msg: string) => {
    setToastMessage(msg);
    setTimeout(() => setToastMessage(null), 3500);
  };

  const handleAcknowledge = (id: string) => {
    setAlerts((prev) =>
      prev.map((a) => {
        if (a.id === id) {
          const nextState = !a.acknowledged;
          showToast(nextState ? `Alert ${id} acknowledged by operator` : `Alert ${id} marked unacknowledged`);
          return {
            ...a,
            acknowledged: nextState,
            acknowledgedBy: nextState ? "Current Operator" : undefined,
          };
        }
        return a;
      })
    );
  };

  const handleResolve = (id: string) => {
    const target = alerts.find((a) => a.id === id);
    setAlerts((prev) => prev.filter((a) => a.id !== id));
    showToast(`Alert on ${target?.hostname || id} resolved`);
  };

  const handleStartPromote = (alert: AlertItem) => {
    setIsPromoting(alert);
    setIncidentTitle(`Investigate ${alert.metric} incident on ${alert.hostname}`);
  };

  const handleConfirmPromote = () => {
    if (!isPromoting) return;
    const promotedAlertId = isPromoting.id;
    // Remove from active alerts list or keep marked
    setAlerts((prev) => prev.filter((a) => a.id !== promotedAlertId));
    setIsPromoting(null);
    showToast(`Promoted alert to Incident: "${incidentTitle}". Redirecting to Incident Response...`);
    setTimeout(() => {
      router.push(`/${subdomain}/incidents`);
    }, 1200);
  };

  // Filtered alerts
  const filteredAlerts = alerts.filter((a) => {
    const matchesSearch =
      a.hostname.toLowerCase().includes(searchTerm.toLowerCase()) ||
      a.metric.toLowerCase().includes(searchTerm.toLowerCase()) ||
      a.id.toLowerCase().includes(searchTerm.toLowerCase());

    if (!matchesSearch) return false;

    if (filterSeverity === "all") return true;
    if (filterSeverity === "ack") return a.acknowledged;
    return a.severity === filterSeverity;
  });

  const criticalCount = alerts.filter((a) => a.severity === "critical").length;
  const warningCount = alerts.filter((a) => a.severity === "warning").length;
  const ackCount = alerts.filter((a) => a.acknowledged).length;

  return (
    <div className="min-h-screen bg-bg text-textMain flex relative">
      <AmbientBackground />
      <DashboardSidebar subdomain={subdomain} />

      <div className="flex-1 ml-[220px] flex flex-col min-h-screen bg-transparent relative">
        {/* Sticky Header */}
        <header className="bg-bg/85 backdrop-blur-md border-b border-borderSubtle flex justify-between items-center px-8 h-14 sticky top-0 z-30">
          <div className="flex items-center gap-3 text-[13px] font-medium text-textMuted">
            <Link href={`/${subdomain}/overview`} className="hover:text-textMain transition-colors">
              Overview
            </Link>
            <span className="text-textSubtle">/</span>
            <span className="text-textMain font-semibold">Active Alerts Center</span>
          </div>

          <div className="flex items-center gap-3">
            <div className="flex items-center gap-1.5 px-2.5 py-1 rounded bg-rose-50 border border-rose-200/50 text-rose-700 text-[10px] font-mono uppercase tracking-wide">
              <span className="w-1.5 h-1.5 rounded-full bg-rose-500 animate-pulse" />
              {criticalCount} Critical Active
            </div>
            <button
              onClick={() => showToast("Alerts refreshed from telemetry bus")}
              className="p-1.5 rounded-lg border border-borderSubtle bg-white hover:bg-black/[0.02] text-textMuted hover:text-textMain transition-colors"
              title="Refresh alerts"
            >
              <RefreshCw className="w-4 h-4" />
            </button>
          </div>
        </header>

        {/* Toast notification */}
        {toastMessage && (
          <div className="fixed bottom-6 right-6 z-50 bg-[#151310] text-white px-4 py-3 rounded-lg shadow-2xl border border-white/10 flex items-center gap-3 animate-fade-in text-xs">
            <CheckCircle2 className="w-4 h-4 text-emerald-400 shrink-0" />
            <span>{toastMessage}</span>
          </div>
        )}

        <main className="p-8 max-w-7xl w-full mx-auto space-y-6">
          {/* Top Headline */}
          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
            <div>
              <h1 className="text-xl font-bold tracking-tight text-textMain flex items-center gap-2">
                <AlertTriangle className="w-5 h-5 text-primary" />
                Active Alerts Center
              </h1>
              <p className="text-xs text-textMuted mt-1">
                Real-time threshold breaches evaluated against live node telemetry streams
              </p>
            </div>
            <div className="flex items-center gap-2">
              <Link
                href={`/${subdomain}/rules`}
                className="px-3.5 py-1.5 text-xs font-semibold rounded-lg border border-borderSubtle bg-white hover:bg-black/[0.02] text-textMain transition-colors shadow-sm"
              >
                Configure Threshold Rules →
              </Link>
            </div>
          </div>

          {/* Stat Cards */}
          <div className="grid grid-cols-1 sm:grid-cols-4 gap-4">
            <div className="border border-white/50 bg-surface backdrop-blur-md rounded-xl p-4 shadow-sm glass-panel flex items-center justify-between">
              <div>
                <div className="text-[11px] font-medium text-textSubtle uppercase tracking-wider">Total Active</div>
                <div className="text-2xl font-bold font-mono text-textMain mt-1">{alerts.length}</div>
              </div>
              <div className="w-10 h-10 rounded-lg bg-black/[0.02] border border-borderSubtle flex items-center justify-center text-textMuted">
                <Bell className="w-5 h-5" />
              </div>
            </div>

            <div className="border border-white/50 bg-surface backdrop-blur-md rounded-xl p-4 shadow-sm glass-panel flex items-center justify-between">
              <div>
                <div className="text-[11px] font-medium text-rose-600 uppercase tracking-wider">Critical Severity</div>
                <div className="text-2xl font-bold font-mono text-rose-600 mt-1">{criticalCount}</div>
              </div>
              <div className="w-10 h-10 rounded-lg bg-rose-50 border border-rose-200/50 flex items-center justify-center text-rose-600">
                <ShieldAlert className="w-5 h-5" />
              </div>
            </div>

            <div className="border border-white/50 bg-surface backdrop-blur-md rounded-xl p-4 shadow-sm glass-panel flex items-center justify-between">
              <div>
                <div className="text-[11px] font-medium text-amber-700 uppercase tracking-wider">Warning Level</div>
                <div className="text-2xl font-bold font-mono text-amber-700 mt-1">{warningCount}</div>
              </div>
              <div className="w-10 h-10 rounded-lg bg-amber-50 border border-amber-200/50 flex items-center justify-center text-primary">
                <AlertTriangle className="w-5 h-5" />
              </div>
            </div>

            <div className="border border-white/50 bg-surface backdrop-blur-md rounded-xl p-4 shadow-sm glass-panel flex items-center justify-between">
              <div>
                <div className="text-[11px] font-medium text-emerald-700 uppercase tracking-wider">Acknowledged</div>
                <div className="text-2xl font-bold font-mono text-emerald-700 mt-1">{ackCount}</div>
              </div>
              <div className="w-10 h-10 rounded-lg bg-emerald-50 border border-emerald-200/50 flex items-center justify-center text-emerald-600">
                <Check className="w-5 h-5" />
              </div>
            </div>
          </div>

          {/* Search & Filter Toolbar */}
          <div className="border border-white/50 bg-surface backdrop-blur-md rounded-xl p-4 shadow-sm glass-panel flex flex-col md:flex-row items-center justify-between gap-3">
            <div className="relative w-full md:w-80">
              <Search className="w-4 h-4 text-textSubtle absolute left-3 top-2.5" />
              <input
                type="text"
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                placeholder="Search hostname, metric, or alert ID..."
                className="w-full bg-white border border-borderSubtle rounded-lg pl-9 pr-3 py-1.5 text-xs text-textMain focus:outline-none focus:border-primary transition-colors shadow-inner"
              />
            </div>

            <div className="flex items-center gap-1.5 w-full md:w-auto overflow-x-auto pb-1 md:pb-0">
              <span className="text-[11px] text-textSubtle mr-1 flex items-center gap-1">
                <Filter className="w-3.5 h-3.5" /> Filter:
              </span>
              {(
                [
                  { key: "all", label: `All (${alerts.length})` },
                  { key: "critical", label: `Critical (${criticalCount})` },
                  { key: "warning", label: `Warning (${warningCount})` },
                  { key: "info", label: `Info (${alerts.filter((a) => a.severity === "info").length})` },
                  { key: "ack", label: `Acknowledged (${ackCount})` },
                ] as const
              ).map((tab) => (
                <button
                  key={tab.key}
                  onClick={() => setFilterSeverity(tab.key)}
                  className={`px-2.5 py-1 rounded-md text-xs font-medium transition-colors ${
                    filterSeverity === tab.key
                      ? "bg-primary text-white shadow-sm"
                      : "bg-white/60 hover:bg-white text-textMuted border border-borderSubtle"
                  }`}
                >
                  {tab.label}
                </button>
              ))}
            </div>
          </div>

          {/* Real-time Alerts Table */}
          <div className="border border-white/50 bg-surface backdrop-blur-md rounded-xl shadow-sm glass-panel overflow-hidden">
            <div className="px-6 py-4 border-b border-borderSubtle flex items-center justify-between">
              <h2 className="text-sm font-semibold text-textMain">Active Triggered Breaches</h2>
              <span className="text-xs text-textSubtle font-mono">
                Showing {filteredAlerts.length} of {alerts.length} alerts
              </span>
            </div>

            {loading ? (
              <div className="p-12 text-center text-xs font-mono text-textMuted flex flex-col items-center gap-3">
                <div className="w-5 h-5 border-2 border-primary border-t-transparent rounded-full animate-spin" />
                <span>Evaluating telemetry threshold rules...</span>
              </div>
            ) : filteredAlerts.length === 0 ? (
              <div className="p-12 text-center space-y-3">
                <CheckCircle2 className="w-10 h-10 text-emerald-500 mx-auto" />
                <div className="text-sm font-semibold text-textMain">
                  {alerts.length === 0 ? "No Active Alerts in Fleet" : "No Alerts Match Current Filter"}
                </div>
                <p className="text-xs text-textMuted max-w-sm mx-auto">
                  {alerts.length === 0
                    ? "All enrolled endpoints are transmitting healthy telemetry within normal operational thresholds."
                    : "No alerts match the active severity or search query criteria."}
                </p>
                {alerts.length === 0 && (
                  <div className="pt-2">
                    <Link
                      href={`/${subdomain}/rules`}
                      className="inline-flex items-center gap-1.5 px-3.5 py-1.5 rounded-md bg-primary hover:bg-primaryHover text-white text-xs font-medium transition-all shadow-sm"
                    >
                      <span>Configure Threshold Rules →</span>
                    </Link>
                  </div>
                )}
              </div>
            ) : (
              <div className="overflow-x-auto">
                <table className="w-full text-left text-xs">
                  <thead className="bg-black/[0.02] border-b border-borderSubtle text-[11px] font-semibold text-textMuted uppercase tracking-wider">
                    <tr>
                      <th className="px-6 py-3">Severity</th>
                      <th className="px-6 py-3">Target Node</th>
                      <th className="px-6 py-3">Breached Metric & Value</th>
                      <th className="px-6 py-3">Trigger Condition</th>
                      <th className="px-6 py-3">Elapsed</th>
                      <th className="px-6 py-3">Operator Status</th>
                      <th className="px-6 py-3 text-right">Actions</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-borderSubtle/60">
                    {filteredAlerts.map((alert) => {
                      const isCrit = alert.severity === "critical";
                      const isWarn = alert.severity === "warning";

                      return (
                        <tr key={alert.id} className="hover:bg-white/40 transition-colors">
                          <td className="px-6 py-4 whitespace-nowrap">
                            <span
                              className={`inline-flex items-center gap-1.5 px-2 py-0.5 rounded text-[10px] font-mono uppercase font-semibold border ${
                                isCrit
                                  ? "bg-rose-50 border-rose-200 text-rose-700"
                                  : isWarn
                                  ? "bg-amber-50 border-amber-200 text-amber-800"
                                  : "bg-slate-100 border-slate-200 text-slate-700"
                              }`}
                            >
                              <span
                                className={`w-1.5 h-1.5 rounded-full ${
                                  isCrit ? "bg-rose-500 animate-pulse" : isWarn ? "bg-amber-500" : "bg-slate-500"
                                }`}
                              />
                              {alert.severity}
                            </span>
                          </td>

                          <td className="px-6 py-4 whitespace-nowrap">
                            <Link
                              href={`/${subdomain}/devices`}
                              className="font-mono font-medium text-textMain hover:text-primary transition-colors flex items-center gap-1.5 group"
                            >
                              <Server className="w-3.5 h-3.5 text-textSubtle group-hover:text-primary transition-colors" />
                              {alert.hostname}
                              <ArrowUpRight className="w-3 h-3 opacity-0 group-hover:opacity-100 transition-opacity" />
                            </Link>
                            <span className="text-[10px] text-textSubtle font-mono block mt-0.5">{alert.id}</span>
                          </td>

                          <td className="px-6 py-4">
                            <div className="font-medium text-textMain">{alert.metric}</div>
                            <div className="font-mono text-[11px] text-rose-600 font-semibold mt-0.5">
                              {alert.currentValue}
                            </div>
                          </td>

                          <td className="px-6 py-4 font-mono text-[11px] text-textMuted whitespace-nowrap">
                            {alert.threshold}
                          </td>

                          <td className="px-6 py-4 whitespace-nowrap">
                            <span className="flex items-center gap-1 text-[11px] text-textSubtle font-mono">
                              <Clock className="w-3.5 h-3.5" />
                              {alert.elapsed}
                            </span>
                          </td>

                          <td className="px-6 py-4 whitespace-nowrap">
                            {alert.acknowledged ? (
                              <div className="text-[11px]">
                                <span className="inline-flex items-center gap-1 text-emerald-700 font-medium">
                                  <Check className="w-3 h-3" /> Acknowledged
                                </span>
                                {alert.acknowledgedBy && (
                                  <span className="block text-[10px] text-textSubtle truncate max-w-[140px]">
                                    {alert.acknowledgedBy}
                                  </span>
                                )}
                              </div>
                            ) : (
                              <span className="inline-flex items-center gap-1 text-amber-700 font-medium text-[11px]">
                                <AlertCircle className="w-3 h-3" /> Pending Review
                              </span>
                            )}
                          </td>

                          <td className="px-6 py-4 text-right whitespace-nowrap">
                            <div className="flex items-center justify-end gap-1.5">
                              <button
                                onClick={() => handleAcknowledge(alert.id)}
                                className={`px-2.5 py-1 rounded text-xs font-medium border transition-colors ${
                                  alert.acknowledged
                                    ? "bg-slate-100 hover:bg-slate-200 border-slate-300 text-slate-700"
                                    : "bg-white hover:bg-emerald-50 border-emerald-300 text-emerald-700"
                                }`}
                                title={alert.acknowledged ? "Unacknowledge" : "Acknowledge alert"}
                              >
                                {alert.acknowledged ? "Unack" : "Acknowledge"}
                              </button>

                              <button
                                onClick={() => handleResolve(alert.id)}
                                className="px-2.5 py-1 rounded text-xs font-medium bg-white hover:bg-emerald-50 border border-borderSubtle text-emerald-700 transition-colors"
                              >
                                Resolve
                              </button>

                              <button
                                onClick={() => handleStartPromote(alert)}
                                className="px-2.5 py-1 rounded text-xs font-semibold bg-rose-50 hover:bg-rose-100 border border-rose-300 text-rose-700 flex items-center gap-1 transition-colors"
                              >
                                <Flame className="w-3.5 h-3.5" />
                                Promote
                              </button>
                            </div>
                          </td>
                        </tr>
                      );
                    })}
                  </tbody>
                </table>
              </div>
            )}
          </div>
        </main>
      </div>

      {/* Promote to Incident Confirmation Modal */}
      {isPromoting && (
        <div className="fixed inset-0 z-50 bg-black/40 backdrop-blur-sm flex items-center justify-center p-4">
          <div className="bg-white rounded-xl shadow-2xl border border-borderSubtle max-w-lg w-full p-6 space-y-4">
            <div className="flex items-center justify-between border-b border-borderSubtle pb-3">
              <div className="flex items-center gap-2 text-rose-700 font-bold text-base">
                <Flame className="w-5 h-5 text-rose-600" />
                Promote Alert to Tracked Incident
              </div>
              <button
                onClick={() => setIsPromoting(null)}
                className="text-textSubtle hover:text-textMain p-1 rounded-md"
              >
                <X className="w-4 h-4" />
              </button>
            </div>

            <div className="space-y-3 text-xs text-textMuted">
              <p>
                Promoting this alert creates an active operational incident in the{" "}
                <strong className="text-textMain">Incident Response Command</strong>, initiates telemetry correlation,
                and prepares AI automated root-cause diagnostics.
              </p>

              <div className="p-3 rounded-lg bg-stone-50 border border-borderSubtle space-y-1 font-mono">
                <div><span className="text-textSubtle">Target Node:</span> <strong className="text-textMain">{isPromoting.hostname}</strong></div>
                <div><span className="text-textSubtle">Metric:</span> <span className="text-rose-700">{isPromoting.metric}</span></div>
                <div><span className="text-textSubtle">Current Reading:</span> {isPromoting.currentValue}</div>
              </div>

              <div className="space-y-1.5 pt-1">
                <label className="text-xs font-semibold text-textMain">Incident Title</label>
                <input
                  type="text"
                  value={incidentTitle}
                  onChange={(e) => setIncidentTitle(e.target.value)}
                  className="w-full bg-white border border-borderSubtle rounded-lg px-3 py-2 text-xs text-textMain focus:outline-none focus:border-primary shadow-sm"
                />
              </div>
            </div>

            <div className="flex justify-end gap-2 pt-2 border-t border-borderSubtle">
              <button
                type="button"
                onClick={() => setIsPromoting(null)}
                className="px-4 py-2 rounded-lg border border-borderSubtle text-xs font-medium text-textMuted hover:bg-stone-50 transition-colors"
              >
                Cancel
              </button>
              <button
                type="button"
                onClick={handleConfirmPromote}
                className="px-4 py-2 rounded-lg bg-rose-600 hover:bg-rose-700 text-white text-xs font-semibold flex items-center gap-1.5 shadow-sm transition-colors"
              >
                <Flame className="w-4 h-4" />
                Confirm & Open Incident
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
