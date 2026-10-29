"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import { DashboardSidebar } from "@/components/Sidebar";
import { AmbientBackground } from "@/components/AmbientBackground";
import {
  Server,
  Search,
  Plus,
  MonitorPlay,
  KeyRound,
  Copy,
  Check,
  RefreshCw,
  Download,
} from "lucide-react";
import { api, Device } from "@/lib/api";
import Link from "next/link";

export default function DevicesPage() {
  const params = useParams();
  const subdomain = (params?.subdomain as string) || "demo";

  const [devices, setDevices] = useState<Device[]>([]);
  const [search, setSearch] = useState("");
  const [statusFilter, setStatusFilter] = useState("all");
  const [loading, setLoading] = useState(true);
  const [showEnrollModal, setShowEnrollModal] = useState(false);
  const [enrollToken, setEnrollToken] = useState("");
  const [copied, setCopied] = useState(false);

  const [enrollCommand, setEnrollCommand] = useState("");
  const [tokenError, setTokenError] = useState<string | null>(null);

  const loadDevices = async () => {
    setLoading(true);
    try {
      const res = await api.listDevices();
      setDevices(res.items || []);
    } catch {
      setDevices([]);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadDevices();
  }, [subdomain]);

  const generateEnrollment = async () => {
    setTokenError(null);
    try {
      const res = await api.createEnrollmentToken();
      setEnrollToken(res.token);
      setEnrollCommand(res.installer_command);
      setShowEnrollModal(true);
    } catch (err: any) {
      setTokenError(err?.message || "Failed to generate enrollment token. Ensure Go API is running.");
      setShowEnrollModal(true);
    }
  };

  const copyToClipboard = (text: string) => {
    navigator.clipboard.writeText(text);
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };

  const filteredDevices = devices.filter((d) => {
    const matchesSearch =
      d.hostname.toLowerCase().includes(search.toLowerCase()) ||
      d.ip_address.includes(search);
    const matchesStatus = statusFilter === "all" || d.status === statusFilter;
    return matchesSearch && matchesStatus;
  });

  return (
    <div className="min-h-screen bg-bg text-textMain flex relative">
      <AmbientBackground />
      <DashboardSidebar subdomain={subdomain} />

      <div className="flex-1 ml-[220px] flex flex-col min-h-screen bg-transparent relative">
        {/* Sticky Header matching V1 */}
        <header className="bg-bg/85 backdrop-blur-md border-b border-borderSubtle flex justify-between items-center px-8 h-14 sticky top-0 z-30">
          <div className="flex items-center gap-3 text-[13px] font-medium text-textMuted">
            <Link href={`/${subdomain}/overview`} className="hover:text-textMain transition-colors">
              Overview
            </Link>
            <span className="text-textSubtle">/</span>
            <span className="text-textMain font-semibold">Infrastructure</span>
          </div>

          <div className="flex items-center gap-3">
            <button
              onClick={loadDevices}
              className="border border-borderSubtle bg-surface text-textMain p-1.5 rounded-md hover:bg-black/[0.02] transition-colors glass-panel"
              title="Refresh Devices"
            >
              <RefreshCw className={`w-3.5 h-3.5 ${loading ? "animate-spin text-primary" : "text-textSubtle"}`} />
            </button>
            <button
              onClick={generateEnrollment}
              className="bg-primary hover:bg-primaryHover text-white px-3 py-1.5 rounded-md text-[12px] font-medium transition-all flex items-center gap-2 shadow-sm"
            >
              <Plus className="w-3.5 h-3.5" />
              <span>Add Device</span>
            </button>
            <div className="w-px h-4 bg-borderSubtle" />
            <div className="flex items-center gap-1.5 px-2 py-1 rounded bg-emerald-50 border border-emerald-200/50 text-emerald-600 text-[10px] font-mono uppercase tracking-wide">
              <div className="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse" />
              Live Fleet
            </div>
          </div>
        </header>

        <main className="flex-1 p-8 max-w-[1200px] mx-auto w-full flex flex-col gap-6">
          <div className="flex flex-col gap-1">
            <h2 className="text-[20px] font-semibold text-textMain tracking-tight">Devices & Infrastructure</h2>
            <p className="text-[12px] text-textSubtle">Manage monitoring agents and dynamic registration configurations.</p>
          </div>

          {/* Filter & Search Bar */}
          <div className="flex flex-col sm:flex-row items-center justify-between gap-4 bg-surface border border-borderSubtle p-2.5 rounded-lg glass-panel shadow-sm">
            <div className="flex items-center gap-2 w-full sm:w-80">
              <Search className="w-4 h-4 text-textSubtle ml-1" />
              <input
                type="text"
                placeholder="Search hostname or IP..."
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                className="bg-transparent border-none text-[13px] text-textMain w-full focus:outline-none placeholder:text-textSubtle font-mono"
              />
            </div>

            <div className="flex items-center gap-2 text-xs font-mono w-full sm:w-auto overflow-x-auto">
              {["all", "online", "degraded", "offline"].map((status) => (
                <button
                  key={status}
                  onClick={() => setStatusFilter(status)}
                  className={`px-3 py-1 rounded-md capitalize transition-colors text-[11px] ${
                    statusFilter === status
                      ? "bg-primary text-white font-semibold shadow-sm"
                      : "border border-borderSubtle bg-white/40 text-textMuted hover:text-textMain"
                  }`}
                >
                  {status}
                </button>
              ))}
            </div>
          </div>

          {/* Fleet Table */}
          <div className="bg-surface border border-borderSubtle rounded-lg overflow-hidden flex flex-col glass-panel shadow-sm">
            <div className="overflow-x-auto w-full">
              <table className="w-full text-left border-collapse text-[12px]">
                <thead>
                  <tr className="border-b border-borderSubtle bg-black/[0.005]">
                    <th className="py-2.5 px-4 font-medium text-textSubtle font-mono uppercase tracking-widest text-[10px]">Hostname</th>
                    <th className="py-2.5 px-4 font-medium text-textSubtle font-mono uppercase tracking-widest text-[10px]">Status</th>
                    <th className="py-2.5 px-4 font-medium text-textSubtle font-mono uppercase tracking-widest text-[10px]">IP Address</th>
                    <th className="py-2.5 px-4 font-medium text-textSubtle font-mono uppercase tracking-widest text-[10px]">Operating System</th>
                    <th className="py-2.5 px-4 font-medium text-textSubtle font-mono uppercase tracking-widest text-[10px]">Assigned User</th>
                    <th className="py-2.5 px-4 font-medium text-textSubtle font-mono uppercase tracking-widest text-[10px] text-right">Actions</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-borderSubtle text-textMain">
                  {filteredDevices.length === 0 && (
                    <tr>
                      <td colSpan={6} className="py-12 px-6 text-center">
                        <div className="max-w-md mx-auto flex flex-col items-center justify-center space-y-3">
                          <div className="w-12 h-12 rounded-xl bg-primary/10 border border-primary/20 flex items-center justify-center text-primary shadow-sm">
                            <Server className="w-6 h-6 text-primary" />
                          </div>
                          {devices.length === 0 ? (
                            <>
                              <h3 className="font-semibold text-textMain text-sm">No endpoints enrolled yet</h3>
                              <p className="text-xs text-textMuted leading-relaxed">
                                Deploy the lightweight Monitor Pro Windows agent to initiate real-time hardware telemetry, automatic heartbeat detection, and secure remote support.
                              </p>
                              <div className="flex items-center gap-3 pt-2">
                                <button
                                  onClick={generateEnrollment}
                                  className="bg-primary hover:bg-primaryHover text-white px-3.5 py-1.5 rounded-md text-xs font-medium transition-all shadow-sm flex items-center gap-2"
                                >
                                  <KeyRound className="w-3.5 h-3.5" />
                                  <span>Generate Token</span>
                                </button>
                                <Link
                                  href={`/${subdomain}/deploy`}
                                  className="border border-borderSubtle bg-surface hover:bg-black/[0.02] text-textMain px-3.5 py-1.5 rounded-md text-xs font-medium transition-all flex items-center gap-2"
                                >
                                  <Download className="w-3.5 h-3.5 text-primary" />
                                  <span>Deployment Wizard →</span>
                                </Link>
                              </div>
                            </>
                          ) : (
                            <>
                              <h3 className="font-semibold text-textMain text-sm">No matching endpoints found</h3>
                              <p className="text-xs text-textMuted">
                                No endpoints in your fleet match the search query or status filter.
                              </p>
                              <button
                                onClick={() => {
                                  setSearch("");
                                  setStatusFilter("all");
                                }}
                                className="text-xs text-primary hover:underline font-medium pt-1"
                              >
                                Clear filters
                              </button>
                            </>
                          )}
                        </div>
                      </td>
                    </tr>
                  )}
                  {filteredDevices.map((device) => (
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
                      <td className="py-3 px-4 text-textSubtle truncate max-w-[200px]">{device.os_version}</td>
                      <td className="py-3 px-4 text-textMuted">{device.assigned_user_email || "Unassigned"}</td>
                      <td className="py-3 px-4 text-right space-x-3">
                        <Link
                          href={`/${subdomain}/remote-support?target=${device.hostname}`}
                          className="text-primary hover:underline font-semibold inline-flex items-center gap-1 text-[12px]"
                        >
                          <MonitorPlay className="w-3.5 h-3.5" /> Remote Support
                        </Link>
                        <Link
                          href={`/${subdomain}/devices/${device.id}`}
                          className="text-textMain hover:text-primary hover:underline font-semibold text-[12px]"
                        >
                          Diagnostics →
                        </Link>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>

          {/* Enrollment Modal matching V1 */}
          {showEnrollModal && (
            <div className="fixed inset-0 bg-black/40 backdrop-blur-sm z-50 flex items-center justify-center p-4">
              <div className="bg-surface border border-white/50 backdrop-blur-2xl rounded-xl max-w-lg w-full p-6 space-y-4 shadow-2xl glass-panel">
                <div className="flex items-center justify-between pb-3 border-b border-borderSubtle">
                  <div className="flex items-center gap-2">
                    <KeyRound className="w-4 h-4 text-primary" />
                    <h3 className="font-semibold text-sm text-textMain">Deploy Windows Agent</h3>
                  </div>
                  <button
                    onClick={() => setShowEnrollModal(false)}
                    className="text-xs text-textSubtle hover:text-textMain"
                  >
                    ✕ Close
                  </button>
                </div>

                {tokenError ? (
                  <div className="p-3 rounded-lg bg-rose-50 border border-rose-200/60 text-xs text-rose-700">
                    <p className="font-semibold">Unable to generate enrollment token</p>
                    <p className="mt-1">{tokenError}</p>
                  </div>
                ) : (
                  <>
                    <p className="text-xs text-textMuted leading-relaxed">
                      Run this command on the target Windows endpoint. It authenticates with the cryptographically verified enrollment token:
                    </p>

                    <div className="p-3 rounded-lg bg-[#151310] border border-white/10 font-mono text-xs text-primary flex items-center justify-between gap-3 shadow-inner">
                      <span className="truncate text-amber-300">
                        {enrollCommand || `proactive-agent.exe --enroll --token ${enrollToken} --server https://${subdomain}.monitopro.com`}
                      </span>
                      <button
                        onClick={() =>
                          copyToClipboard(
                            enrollCommand || `proactive-agent.exe --enroll --token ${enrollToken} --server https://${subdomain}.monitopro.com`
                          )
                        }
                        className="p-1.5 rounded hover:bg-white/10 text-white flex-shrink-0"
                        title="Copy to clipboard"
                      >
                        {copied ? <Check className="w-4 h-4 text-success" /> : <Copy className="w-4 h-4" />}
                      </button>
                    </div>

                    <div className="text-[11px] font-mono text-textSubtle">
                      Token expires in 24 hours. Single-use only.
                    </div>
                  </>
                )}
              </div>
            </div>
          )}
        </main>
      </div>
    </div>
  );
}
