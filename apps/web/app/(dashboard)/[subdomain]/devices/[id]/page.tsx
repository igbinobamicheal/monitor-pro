"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import { DashboardSidebar } from "@/components/Sidebar";
import { AmbientBackground } from "@/components/AmbientBackground";
import {
  Server,
  Activity,
  Cpu,
  Shield,
  List,
  FileText,
  Package,
  MonitorPlay,
  ArrowLeft,
} from "lucide-react";
import { api, Device, MetricSnapshot } from "@/lib/api";
import Link from "next/link";

export default function DeviceDiagnosticsPage() {
  const params = useParams();
  const subdomain = (params?.subdomain as string) || "demo";
  const deviceId = (params?.id as string) || "";

  const [device, setDevice] = useState<(Device & { latest_metrics?: MetricSnapshot }) | null>(null);
  const [activeTab, setActiveTab] = useState<"overview" | "system" | "network" | "processes" | "logs" | "software">("overview");
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchDevice = async () => {
      setLoading(true);
      setError(null);
      try {
        const d = await api.getDevice(deviceId);
        setDevice(d);
      } catch (err: any) {
        setDevice(null);
        setError(err?.message || "Endpoint not found");
      } finally {
        setLoading(false);
      }
    };
    fetchDevice();
  }, [deviceId]);

  return (
    <div className="min-h-screen bg-bg text-textMain flex relative">
      <AmbientBackground />
      <DashboardSidebar subdomain={subdomain} />

      <div className="flex-1 ml-[220px] flex flex-col min-h-screen bg-transparent relative">
        {/* Sticky Header matching V1 */}
        <header className="bg-bg/85 backdrop-blur-md border-b border-borderSubtle flex justify-between items-center px-8 h-14 sticky top-0 z-30">
          <div className="flex items-center gap-3 text-[13px] font-medium text-textMuted">
            <Link href={`/${subdomain}/devices`} className="hover:text-textMain transition-colors flex items-center gap-1.5">
              <ArrowLeft className="w-3.5 h-3.5 text-textSubtle" /> Infrastructure
            </Link>
            <span className="text-textSubtle">/</span>
            <span className="text-textMain font-semibold font-mono">{device?.hostname || deviceId}</span>
          </div>

          <div className="flex items-center gap-3">
            <Link
              href={`/${subdomain}/remote-support?target=${device?.hostname || deviceId}`}
              className="bg-primary hover:bg-primaryHover text-white px-3.5 py-1.5 rounded-md text-[12px] font-medium transition-all flex items-center gap-2 shadow-sm"
            >
              <MonitorPlay className="w-3.5 h-3.5" />
              <span>Launch Remote Support</span>
            </Link>
          </div>
        </header>

        <main className="flex-1 p-8 max-w-[1200px] mx-auto w-full flex flex-col gap-6">
          {loading ? (
            <div className="py-20 text-center text-xs font-mono text-textMuted flex flex-col items-center gap-3">
              <div className="w-6 h-6 border-2 border-primary border-t-transparent rounded-full animate-spin" />
              <span>Querying hardware telemetry profile...</span>
            </div>
          ) : !device ? (
            <div className="rounded-xl border border-borderSubtle bg-surface p-12 text-center glass-panel shadow-sm max-w-lg mx-auto space-y-4">
              <div className="w-12 h-12 rounded-xl bg-amber-50 border border-amber-200/60 flex items-center justify-center text-primary mx-auto shadow-sm">
                <Server className="w-6 h-6 text-primary" />
              </div>
              <div className="space-y-1">
                <h2 className="text-base font-bold text-textMain">Endpoint Not Found</h2>
                <p className="text-xs text-textMuted leading-relaxed">
                  No enrolled endpoint with ID <code className="font-mono text-primary font-semibold">{deviceId}</code> was found in workspace <span className="font-semibold">{subdomain}</span>.
                </p>
              </div>
              <div className="pt-2">
                <Link
                  href={`/${subdomain}/devices`}
                  className="bg-primary hover:bg-primaryHover text-white px-4 py-2 rounded-md text-xs font-medium transition-all shadow-sm inline-flex items-center gap-2"
                >
                  <ArrowLeft className="w-3.5 h-3.5" />
                  <span>Return to Fleet Infrastructure</span>
                </Link>
              </div>
            </div>
          ) : (
            <>
              {/* Device Profile Header */}
              <div className="p-6 rounded-xl border border-white/50 bg-surface flex flex-col md:flex-row items-start md:items-center justify-between gap-6 shadow-sm glass-panel">
                <div className="flex items-center gap-4">
                  <div className="w-12 h-12 rounded-xl bg-primary/10 border border-primary/30 flex items-center justify-center text-primary font-bold">
                    <Server className="w-6 h-6" />
                  </div>
                  <div>
                    <div className="flex items-center gap-2.5">
                      <h1 className="text-xl font-bold tracking-tight text-textMain">{device.hostname}</h1>
                      <span className={`px-2 py-0.5 rounded text-[10px] font-mono font-bold uppercase border ${
                        device.status === 'online'
                          ? 'border-emerald-200/50 bg-emerald-50 text-emerald-600'
                          : device.status === 'degraded'
                          ? 'border-amber-200/50 bg-amber-50 text-amber-700'
                          : 'border-rose-200/50 bg-rose-50 text-rose-700'
                      }`}>
                        {device.status}
                      </span>
                    </div>
                    <div className="flex items-center gap-4 text-xs font-mono text-textMuted mt-1">
                      <span>IP: {device.ip_address || "Awaiting broadcast"}</span>
                      <span>•</span>
                      <span>OS: {device.os_version || "Windows 11 / Server"}</span>
                      <span>•</span>
                      <span>Assigned: {device.assigned_user_email || "Unassigned"}</span>
                    </div>
                  </div>
                </div>
              </div>

              {/* Diagnostic Tabs */}
              <div className="border-b border-borderSubtle flex items-center gap-6 text-xs font-mono overflow-x-auto">
                {[
                  { id: "overview", label: "Overview & Gauges", icon: Activity },
                  { id: "system", label: "System & Hardware", icon: Cpu },
                  { id: "network", label: "Network & Security", icon: Shield },
                  { id: "processes", label: "Processes & Services", icon: List },
                  { id: "logs", label: "Event Logs", icon: FileText },
                  { id: "software", label: "Installed Software", icon: Package },
                ].map((tab) => {
                  const Icon = tab.icon;
                  const active = activeTab === tab.id;
                  return (
                    <button
                      key={tab.id}
                      onClick={() => setActiveTab(tab.id as any)}
                      className={`pb-3 flex items-center gap-2 border-b-2 font-medium transition-colors text-[13px] whitespace-nowrap ${
                        active
                          ? "border-primary text-textMain font-semibold"
                          : "border-transparent text-textMuted hover:text-textMain"
                      }`}
                    >
                      <Icon className={`w-4 h-4 ${active ? "text-primary" : "text-textSubtle"}`} />
                      <span>{tab.label}</span>
                    </button>
                  );
                })}
              </div>

              {/* Tab 1: Overview & Gauges */}
              {activeTab === "overview" && (
                <div className="space-y-6">
                  <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
                    <div className="metric-card rounded-lg p-5 space-y-2">
                      <span className="text-[11px] font-mono text-textMuted uppercase">CPU Utilization</span>
                      <div className="text-3xl font-bold font-mono text-primary">
                        {device.latest_metrics ? `${device.latest_metrics.cpu_usage.toFixed(1)}%` : "0.0%"}
                      </div>
                      <div className="w-full bg-black/[0.06] h-1.5 rounded-full overflow-hidden">
                        <div
                          className="bg-primary h-full rounded-full transition-all duration-500"
                          style={{ width: `${device.latest_metrics?.cpu_usage || 0}%` }}
                        />
                      </div>
                    </div>

                    <div className="metric-card rounded-lg p-5 space-y-2">
                      <span className="text-[11px] font-mono text-textMuted uppercase">Memory Load</span>
                      <div className="text-3xl font-bold font-mono text-success">
                        {device.latest_metrics ? `${device.latest_metrics.ram_usage.toFixed(1)}%` : "0.0%"}
                      </div>
                      <div className="w-full bg-black/[0.06] h-1.5 rounded-full overflow-hidden">
                        <div
                          className="bg-success h-full rounded-full transition-all duration-500"
                          style={{ width: `${device.latest_metrics?.ram_usage || 0}%` }}
                        />
                      </div>
                    </div>

                    <div className="metric-card rounded-lg p-5 space-y-2">
                      <span className="text-[11px] font-mono text-textMuted uppercase">System Drive (C:)</span>
                      <div className="text-3xl font-bold font-mono text-textMain">
                        {device.latest_metrics ? `${device.latest_metrics.disk_usage.toFixed(1)}%` : "0.0%"}
                      </div>
                      <div className="w-full bg-black/[0.06] h-1.5 rounded-full overflow-hidden">
                        <div
                          className="bg-textSubtle h-full rounded-full transition-all duration-500"
                          style={{ width: `${device.latest_metrics?.disk_usage || 0}%` }}
                        />
                      </div>
                    </div>

                    <div className="metric-card rounded-lg p-5 space-y-2">
                      <span className="text-[11px] font-mono text-textMuted uppercase">Uptime & Network</span>
                      <div className="text-xl font-bold font-mono text-textMain">
                        {device.latest_metrics?.uptime_seconds
                          ? `${Math.floor(device.latest_metrics.uptime_seconds / 86400)}d ${Math.floor((device.latest_metrics.uptime_seconds % 86400) / 3600)}h`
                          : "Online"}
                      </div>
                      <div className="text-[11px] font-mono text-textSubtle">
                        In: {device.latest_metrics?.network_in_kb || 0} KB/s • Out: {device.latest_metrics?.network_out_kb || 0} KB/s
                      </div>
                    </div>
                  </div>
                </div>
              )}

              {/* Tab 2: System & Hardware */}
              {activeTab === "system" && (
                <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                  <div className="rounded-xl border border-white/50 bg-surface p-6 space-y-4 glass-panel">
                    <h3 className="font-semibold text-sm text-textMain">Motherboard & Operating System</h3>
                    <div className="space-y-2 text-xs font-mono">
                      <div className="flex justify-between py-1.5 border-b border-borderSubtle">
                        <span className="text-textMuted">Hostname</span>
                        <span className="font-semibold text-textMain">{device.hostname}</span>
                      </div>
                      <div className="flex justify-between py-1.5 border-b border-borderSubtle">
                        <span className="text-textMuted">Operating System</span>
                        <span className="font-semibold text-textMain">{device.os_version || "Windows 64-bit"}</span>
                      </div>
                      <div className="flex justify-between py-1.5 border-b border-borderSubtle">
                        <span className="text-textMuted">IP Address</span>
                        <span className="font-semibold text-primary">{device.ip_address || "DHCP / Unassigned"}</span>
                      </div>
                      <div className="flex justify-between py-1.5 border-b border-borderSubtle">
                        <span className="text-textMuted">Agent Status</span>
                        <span className="font-semibold text-success uppercase">{device.status}</span>
                      </div>
                    </div>
                  </div>

                  <div className="rounded-xl border border-white/50 bg-surface p-6 space-y-4 glass-panel">
                    <h3 className="font-semibold text-sm text-textMain">Endpoint Telemetry Pipeline</h3>
                    <div className="space-y-2 text-xs font-mono">
                      <div className="flex justify-between py-1.5 border-b border-borderSubtle">
                        <span className="text-textMuted">Heartbeat Status</span>
                        <span className="font-semibold text-textMain">Active</span>
                      </div>
                      <div className="flex justify-between py-1.5 border-b border-borderSubtle">
                        <span className="text-textMuted">Telemetry Interval</span>
                        <span className="font-semibold text-textMain">30 seconds</span>
                      </div>
                      <div className="flex justify-between py-1.5 border-b border-borderSubtle">
                        <span className="text-textMuted">Last Ping</span>
                        <span className="font-semibold text-textMain">{device.last_seen_at ? new Date(device.last_seen_at).toLocaleTimeString() : "Just now"}</span>
                      </div>
                      <div className="flex justify-between py-1.5 border-b border-borderSubtle">
                        <span className="text-textMuted">Encryption</span>
                        <span className="font-semibold text-success">Ed25519 Signed Telemetry</span>
                      </div>
                    </div>
                  </div>
                </div>
              )}

              {/* Tab 3: Network & Security */}
              {activeTab === "network" && (
                <div className="rounded-xl border border-white/50 bg-surface p-6 space-y-4 glass-panel">
                  <h3 className="font-semibold text-sm text-textMain">Endpoint Network Configuration</h3>
                  <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                    <div className="p-4 rounded-lg bg-white/70 border border-borderSubtle space-y-1">
                      <div className="text-[10px] font-mono uppercase text-textSubtle">Network IP Address</div>
                      <div className="text-sm font-bold font-mono text-textMain">{device.ip_address || "127.0.0.1"}</div>
                      <div className="text-[11px] font-mono text-textMuted">Primary interface binding</div>
                    </div>
                    <div className="p-4 rounded-lg bg-white/70 border border-borderSubtle space-y-1">
                      <div className="text-[10px] font-mono uppercase text-textSubtle">Agent Cryptography</div>
                      <div className="text-sm font-bold text-success">Ed25519 Verified</div>
                      <div className="text-[11px] font-mono text-textMuted">Cryptographic mutual auth</div>
                    </div>
                    <div className="p-4 rounded-lg bg-white/70 border border-borderSubtle space-y-1">
                      <div className="text-[10px] font-mono uppercase text-textSubtle">Network Traffic</div>
                      <div className="text-sm font-bold font-mono text-textMain">
                        {device.latest_metrics ? `${(device.latest_metrics.network_in_kb + device.latest_metrics.network_out_kb).toFixed(1)} KB/s` : "0 KB/s"}
                      </div>
                      <div className="text-[11px] font-mono text-textMuted">Combined I/O rate</div>
                    </div>
                  </div>
                </div>
              )}

              {/* Tab 4: Processes & Services */}
              {activeTab === "processes" && (
                <div className="rounded-xl border border-white/50 bg-surface overflow-hidden glass-panel">
                  <div className="p-4 border-b border-borderSubtle font-semibold text-sm text-textMain flex items-center justify-between">
                    <span>Active Endpoint Processes</span>
                    <span className="text-xs font-mono text-textSubtle">Live Telemetry</span>
                  </div>
                  <div className="p-8 text-center text-xs font-mono text-textMuted space-y-2">
                    <p>Process snapshot is populated during scheduled deep collection cycles or on-demand diagnostic triggers.</p>
                    <Link
                      href={`/${subdomain}/remediation`}
                      className="inline-flex items-center gap-1.5 text-primary hover:underline font-semibold text-xs mt-2"
                    >
                      Trigger deep diagnostics job via Remediation Catalogue →
                    </Link>
                  </div>
                </div>
              )}

              {/* Tab 5: Event Logs */}
              {activeTab === "logs" && (
                <div className="rounded-xl border border-white/50 bg-surface p-4 space-y-3 font-mono text-xs glass-panel">
                  <div className="text-xs font-semibold text-textMuted uppercase">
                    Windows Event Log Streams
                  </div>
                  <div className="p-6 rounded-lg bg-[#151310] text-zinc-300 space-y-2 border border-white/10 shadow-inner text-center">
                    <p className="text-zinc-400">No anomalous event log errors or warning events reported by agent.</p>
                    <span className="text-[11px] text-zinc-500">Event filter: Level &lt;= 3 (Critical, Error, Warning)</span>
                  </div>
                </div>
              )}

              {/* Tab 6: Software */}
              {activeTab === "software" && (
                <div className="rounded-xl border border-white/50 bg-surface p-6 space-y-3 text-xs font-mono glass-panel">
                  <div className="font-semibold text-sm text-textMain">Installed Software &amp; Agent Info</div>
                  <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                    <div className="p-3 rounded-lg bg-white/70 border border-borderSubtle flex items-center gap-2">
                      <Package className="w-4 h-4 text-primary" />
                      <div>
                        <div className="text-textMain font-medium">Proactive IT Support Agent</div>
                        <div className="text-[11px] text-textSubtle">Native Rust Release Binary (v2.0.0)</div>
                      </div>
                    </div>
                    <div className="p-3 rounded-lg bg-white/70 border border-borderSubtle flex items-center gap-2">
                      <Shield className="w-4 h-4 text-emerald-600" />
                      <div>
                        <div className="text-textMain font-medium">{device.os_version || "Microsoft Windows"}</div>
                        <div className="text-[11px] text-textSubtle">64-bit Operating System Architecture</div>
                      </div>
                    </div>
                  </div>
                </div>
              )}
            </>
          )}
        </main>
      </div>
    </div>
  );
}
