"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import { DashboardSidebar } from "@/components/Sidebar";
import { AmbientBackground } from "@/components/AmbientBackground";
import {
  Server,
  Activity,
  AlertTriangle,
  Flame,
  CheckCircle2,
  RefreshCw,
  ArrowUpRight,
  Shield,
  Plus,
} from "lucide-react";
import { api, Device, Incident } from "@/lib/api";
import Link from "next/link";

export default function OverviewPage() {
  const params = useParams();
  const subdomain = (params?.subdomain as string) || "demo";

  const [devices, setDevices] = useState<Device[]>([]);
  const [incidents, setIncidents] = useState<Incident[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchData = async () => {
    setLoading(true);
    setError(null);
    try {
      const [devRes, incRes] = await Promise.allSettled([
        api.listDevices(),
        api.listIncidents(),
      ]);

      if (devRes.status === "fulfilled") {
        setDevices(devRes.value.items || []);
      }
      if (incRes.status === "fulfilled") {
        setIncidents(incRes.value || []);
      }
    } catch {
      setError("Unable to connect to Go control plane API. Check service status.");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchData();
  }, [subdomain]);

  const onlineCount = devices.filter((d) => d.status === "online").length;
  const degradedCount = devices.filter((d) => d.status === "degraded").length;
  const offlineCount = devices.filter((d) => d.status === "offline").length;

  return (
    <div className="min-h-screen bg-bg text-textMain flex relative">
      <AmbientBackground />
      <DashboardSidebar subdomain={subdomain} />

      <div className="flex-1 ml-[220px] flex flex-col min-h-screen bg-transparent relative">
        {/* Sticky Header matching V1 */}
        <header className="bg-bg/85 backdrop-blur-md border-b border-borderSubtle flex justify-between items-center px-8 h-14 sticky top-0 z-30">
          <div className="flex items-center gap-3 text-[13px] font-medium text-textMuted">
            <span className="text-textMain font-semibold">Overview</span>
          </div>

          <div className="flex items-center gap-3">
            <button
              onClick={fetchData}
              disabled={loading}
              className="border border-borderSubtle bg-surface text-textMain px-3 py-1.5 rounded-md text-[12px] font-medium hover:bg-black/[0.02] transition-colors flex items-center gap-2 glass-panel"
            >
              <RefreshCw className={`w-3.5 h-3.5 ${loading ? "animate-spin text-primary" : "text-textSubtle"}`} />
              <span>Refresh Fleet</span>
            </button>
            <Link
              href={`/${subdomain}/devices`}
              className="bg-primary hover:bg-primaryHover text-white px-3 py-1.5 rounded-md text-[12px] font-medium transition-all flex items-center gap-2 shadow-sm"
            >
              <Plus className="w-3.5 h-3.5" />
              <span>Add Device</span>
            </Link>
            <div className="w-px h-4 bg-borderSubtle" />
            <div className="flex items-center gap-1.5 px-2 py-1 rounded bg-emerald-50 border border-emerald-200/50 text-emerald-600 text-[10px] font-mono uppercase tracking-wide">
              <div className="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse" />
              Live Events Connected
            </div>
          </div>
        </header>

        <main className="flex-1 p-8 max-w-[1200px] mx-auto w-full flex flex-col gap-6">
          <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-2">
            <div>
              <h2 className="text-[20px] font-semibold text-textMain tracking-tight">Global Health</h2>
              <p className="text-[12px] text-textSubtle mt-0.5">
                Workspace <span className="font-mono text-primary font-medium">{subdomain}.monitopro.com</span> • 30-second telemetry interval
              </p>
            </div>
          </div>

          {/* Error Alert if backend unreachable */}
          {error && (
            <div className="p-4 rounded-lg bg-amber-500/10 border border-amber-500/30 text-amber-700 text-xs flex items-center justify-between">
              <div className="flex items-center gap-2.5">
                <AlertTriangle className="w-4 h-4 flex-shrink-0" />
                <span>{error}</span>
              </div>
              <button onClick={fetchData} className="underline font-semibold hover:text-textMain">
                Retry Connection
              </button>
            </div>
          )}

          {/* Stats Card Row matching V1 */}
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
            <div className="metric-card rounded-lg p-4 flex flex-col gap-3 relative overflow-hidden">
              <div className="flex justify-between items-center">
                <span className="text-[12px] font-medium text-textMuted">Total Nodes</span>
                <Server className="w-3.5 h-3.5 text-textSubtle" />
              </div>
              <div className="text-[24px] font-bold font-mono tracking-tight text-textMain">
                {loading ? "..." : devices.length}
              </div>
              <div className="text-[11px] font-mono text-textSubtle flex items-center gap-2">
                <span className="text-success font-semibold">{onlineCount} online</span>
                <span>•</span>
                <span className="text-danger font-semibold">{offlineCount} offline</span>
              </div>
            </div>

            <div className="metric-card rounded-lg p-4 flex flex-col gap-3 relative overflow-hidden">
              <div className="flex justify-between items-center">
                <span className="text-[12px] font-medium text-textMuted">Active (Online)</span>
                <CheckCircle2 className="w-3.5 h-3.5 text-success" />
              </div>
              <div className="text-[24px] font-bold font-mono tracking-tight text-success">
                {loading ? "..." : onlineCount}
              </div>
              <div className="text-[11px] font-mono text-textSubtle">
                {devices.length === 0 ? "100% optimal" : `${Math.round((onlineCount / (devices.length || 1)) * 100)}% fleet operational`}
              </div>
            </div>

            <div className="metric-card rounded-lg p-4 flex flex-col gap-3 relative overflow-hidden">
              <div className="flex justify-between items-center">
                <span className="text-[12px] font-medium text-textMuted">Active Alerts & Incidents</span>
                <Flame className="w-3.5 h-3.5 text-warning" />
              </div>
              <div className="text-[24px] font-bold font-mono tracking-tight text-warning">
                {loading ? "..." : incidents.filter((i) => i.status === "open").length}
              </div>
              <div className="text-[11px] font-mono text-textSubtle">
                {incidents.length} total correlated events
              </div>
            </div>

            <div className="metric-card rounded-lg p-4 flex flex-col gap-3 relative overflow-hidden">
              <div className="flex justify-between items-center">
                <span className="text-[12px] font-medium text-textMuted">Advisory AI Status</span>
                <Shield className="w-3.5 h-3.5 text-primary" />
              </div>
              <div className="text-[24px] font-bold font-mono tracking-tight text-primary">Active</div>
              <div className="text-[11px] font-mono text-textSubtle">Read-only diagnostic mode</div>
            </div>
          </div>

          {/* Tables Section matching V1 */}
          <div className="bg-surface border border-borderSubtle rounded-lg overflow-hidden flex flex-col glass-panel shadow-sm">
            <div className="p-3.5 border-b border-borderSubtle flex justify-between items-center bg-black/[0.01]">
              <span className="text-[13px] font-semibold text-textMain">Recent Devices</span>
              <Link
                href={`/${subdomain}/devices`}
                className="text-[11px] font-medium text-textSubtle hover:text-textMain transition-colors flex items-center gap-1"
              >
                View Full Fleet ({devices.length}) →
              </Link>
            </div>

            <div className="overflow-x-auto w-full">
              <table className="w-full text-left border-collapse text-[12px]">
                <thead>
                  <tr className="border-b border-borderSubtle bg-black/[0.005]">
                    <th className="py-2.5 px-4 font-medium text-textSubtle font-mono uppercase tracking-widest text-[10px]">Device Name</th>
                    <th className="py-2.5 px-4 font-medium text-textSubtle font-mono uppercase tracking-widest text-[10px]">Status</th>
                    <th className="py-2.5 px-4 font-medium text-textSubtle font-mono uppercase tracking-widest text-[10px]">IP Address</th>
                    <th className="py-2.5 px-4 font-medium text-textSubtle font-mono uppercase tracking-widest text-[10px]">Operating System</th>
                    <th className="py-2.5 px-4 font-medium text-textSubtle font-mono uppercase tracking-widest text-[10px]">Assigned User</th>
                    <th className="py-2.5 px-4 font-medium text-textSubtle font-mono uppercase tracking-widest text-[10px] text-right">Action</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-borderSubtle text-textMain">
                  {devices.length === 0 && !loading && (
                    <tr>
                      <td colSpan={6} className="text-center text-textMuted py-8">
                        No managed endpoints enrolled yet under <span className="font-semibold">{subdomain}</span>.
                        <div className="mt-2">
                          <Link
                            href={`/${subdomain}/devices`}
                            className="text-primary hover:underline font-semibold"
                          >
                            Generate Enrollment Token →
                          </Link>
                        </div>
                      </td>
                    </tr>
                  )}
                  {devices.slice(0, 5).map((device) => (
                    <tr key={device.id} className="hover:bg-black/[0.01] transition-colors">
                      <td className="py-3 px-4 font-semibold text-textMain flex items-center gap-2">
                        <div
                          className={`w-2 h-2 rounded-full ${
                            device.status === "online"
                              ? "bg-success"
                              : device.status === "degraded"
                              ? "bg-warning"
                              : "bg-danger"
                          }`}
                        />
                        <span>{device.hostname}</span>
                      </td>
                      <td className="py-3 px-4 uppercase text-[10px] font-mono">
                        <span
                          className={`px-2 py-0.5 rounded border ${
                            device.status === "online"
                              ? "border-emerald-200/50 bg-emerald-50 text-emerald-600"
                              : device.status === "degraded"
                              ? "border-amber-200/50 bg-amber-50 text-amber-700"
                              : "border-rose-200/50 bg-rose-50 text-rose-700"
                          }`}
                        >
                          {device.status}
                        </span>
                      </td>
                      <td className="py-3 px-4 text-textSubtle font-mono">{device.ip_address}</td>
                      <td className="py-3 px-4 text-textSubtle truncate max-w-[200px]">
                        {device.os_version || "Windows 11 Pro"}
                      </td>
                      <td className="py-3 px-4 text-textMuted">
                        {device.assigned_user_email || "Unassigned"}
                      </td>
                      <td className="py-3 px-4 text-right space-x-2">
                        <Link
                          href={`/${subdomain}/devices/${device.id}`}
                          className="text-primary hover:underline font-semibold text-[12px]"
                        >
                          Inspect
                        </Link>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        </main>
      </div>
    </div>
  );
}
