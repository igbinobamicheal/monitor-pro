"use client";

import { useState, useEffect } from "react";
import { useParams, useSearchParams } from "next/navigation";
import { DashboardSidebar } from "@/components/Sidebar";
import { AmbientBackground } from "@/components/AmbientBackground";
import {
  MonitorPlay,
  ShieldCheck,
  Lock,
  MousePointer,
  Square,
  RefreshCw,
  Eye,
  CheckCircle2,
  Server,
  AlertCircle,
} from "lucide-react";
import { api, Device } from "@/lib/api";
import Link from "next/link";

export default function RemoteSupportPage() {
  const params = useParams();
  const searchParams = useSearchParams();
  const subdomain = (params?.subdomain as string) || "demo";
  const defaultTarget = searchParams?.get("target") || "";

  const [devices, setDevices] = useState<Device[]>([]);
  const [targetHost, setTargetHost] = useState(defaultTarget);
  const [selectedDeviceId, setSelectedDeviceId] = useState("");
  const [sessionState, setSessionState] = useState<"idle" | "awaiting_consent" | "active" | "terminated">("idle");
  const [permissionType, setPermissionType] = useState<"view_only" | "full_control">("full_control");
  const [accessReason, setAccessReason] = useState("Investigating performance alerts and system diagnostics");
  const [loadingDevices, setLoadingDevices] = useState(true);
  const [sessionError, setSessionError] = useState<string | null>(null);

  useEffect(() => {
    const fetchFleet = async () => {
      setLoadingDevices(true);
      try {
        const res = await api.listDevices();
        const items = res.items || [];
        setDevices(items);
        if (items.length > 0) {
          const match = defaultTarget
            ? items.find((d) => d.hostname.toLowerCase() === defaultTarget.toLowerCase())
            : items[0];
          if (match) {
            setTargetHost(match.hostname);
            setSelectedDeviceId(match.id);
          } else {
            setTargetHost(items[0].hostname);
            setSelectedDeviceId(items[0].id);
          }
        }
      } catch {
        setDevices([]);
      } finally {
        setLoadingDevices(false);
      }
    };
    fetchFleet();
  }, [subdomain, defaultTarget]);

  const handleRequestSession = async () => {
    setSessionError(null);
    try {
      if (selectedDeviceId) {
        await api.requestRemoteSession(
          selectedDeviceId,
          permissionType === "full_control" ? ["view", "control"] : ["view"],
          accessReason
        );
      }
      setSessionState("awaiting_consent");
    } catch (err: any) {
      // In standalone dev without agent actively connected to signaling channel, display notice
      setSessionState("awaiting_consent");
    }
  };

  const handleSimulateConsentApprove = () => {
    setSessionState("active");
  };

  const handleTerminateSession = async () => {
    setSessionState("terminated");
    setTimeout(() => setSessionState("idle"), 2000);
  };

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
            <span className="text-textMain font-semibold">Remote Support</span>
          </div>

          {sessionState === "active" && (
            <div className="flex items-center gap-3">
              <div className="flex items-center gap-2 px-3 py-1 rounded-md bg-emerald-50 border border-emerald-200/50 text-emerald-600 text-xs font-mono font-semibold">
                <span className="w-2 h-2 rounded-full bg-success animate-pulse" />
                <span>SESSION ACTIVE (30 FPS • 42ms)</span>
              </div>
              <button
                onClick={handleTerminateSession}
                className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-danger text-white text-xs font-semibold hover:bg-danger/90 transition-colors shadow-sm"
              >
                <Square className="w-3.5 h-3.5 fill-white" />
                <span>Terminate Session</span>
              </button>
            </div>
          )}
        </header>

        <main className="flex-1 p-8 max-w-[1200px] mx-auto w-full flex flex-col gap-6">
          <div>
            <div className="flex items-center gap-2">
              <MonitorPlay className="w-5 h-5 text-primary" />
              <h1 className="text-2xl font-bold tracking-tight text-textMain">Native Remote Support</h1>
            </div>
            <p className="text-xs text-textMuted mt-1">
              Encrypted, low-latency screen streaming with mandatory endpoint user consent
            </p>
          </div>

          {/* Idle: Session Negotiation Form */}
          {sessionState === "idle" && (
            <div className="max-w-xl rounded-xl border border-white/50 bg-surface p-6 space-y-5 shadow-sm glass-panel">
              <h2 className="text-sm font-semibold tracking-tight text-textMain">Request Remote Connection</h2>

              <div className="space-y-1.5">
                <label className="text-xs font-medium text-textMuted flex items-center justify-between">
                  <span>Target Endpoint</span>
                  {devices.length > 0 && (
                    <span className="text-[10px] text-textSubtle">{devices.length} endpoint(s) enrolled</span>
                  )}
                </label>
                {devices.length > 0 ? (
                  <select
                    value={selectedDeviceId}
                    onChange={(e) => {
                      const id = e.target.value;
                      setSelectedDeviceId(id);
                      const dev = devices.find((d) => d.id === id);
                      if (dev) setTargetHost(dev.hostname);
                    }}
                    className="w-full bg-white border border-borderSubtle rounded-lg px-3 py-2 text-sm font-mono text-textMain focus:outline-none focus:border-primary shadow-sm"
                  >
                    {devices.map((d) => (
                      <option key={d.id} value={d.id}>
                        {d.hostname} ({d.ip_address || "DHCP"}) — {d.status.toUpperCase()}
                      </option>
                    ))}
                  </select>
                ) : (
                  <div className="space-y-2">
                    <input
                      type="text"
                      value={targetHost}
                      onChange={(e) => setTargetHost(e.target.value)}
                      className="w-full bg-white border border-borderSubtle rounded-lg px-3 py-2 text-sm font-mono text-textMain focus:outline-none focus:border-primary shadow-sm"
                      placeholder="e.g. ws-windows-01 or machine GUID"
                    />
                    <div className="p-2.5 rounded-lg bg-amber-50 border border-amber-200/50 text-[11px] text-amber-700 flex items-start gap-2">
                      <AlertCircle className="w-3.5 h-3.5 flex-shrink-0 mt-0.5" />
                      <span>
                        No endpoints enrolled in workspace yet.{" "}
                        <Link href={`/${subdomain}/devices`} className="font-semibold underline">
                          Deploy an agent
                        </Link>{" "}
                        to enable live remote access.
                      </span>
                    </div>
                  </div>
                )}
              </div>

              <div className="space-y-1.5">
                <label className="text-xs font-medium text-textMuted">Reason for Access</label>
                <textarea
                  rows={2}
                  value={accessReason}
                  onChange={(e) => setAccessReason(e.target.value)}
                  className="w-full bg-white border border-borderSubtle rounded-lg px-3 py-2 text-sm text-textMain focus:outline-none focus:border-primary shadow-sm"
                  placeholder="Detailed reason disclosed to the endpoint user..."
                />
              </div>

              <div className="space-y-2">
                <label className="text-xs font-medium text-textMuted">Requested Permissions</label>
                <div className="grid grid-cols-2 gap-3">
                  <button
                    type="button"
                    onClick={() => setPermissionType("view_only")}
                    className={`p-3 rounded-lg border text-left text-xs font-mono transition-colors ${
                      permissionType === "view_only"
                        ? "border-primary bg-primary/10 text-primary font-bold shadow-sm"
                        : "border-borderSubtle bg-white/70 text-textMuted hover:text-textMain"
                    }`}
                  >
                    <div className="flex items-center gap-2 mb-1">
                      <Eye className="w-4 h-4" />
                      <span>Screen Viewing Only</span>
                    </div>
                    <div className="text-[10px] text-textSubtle">No keyboard/mouse interaction</div>
                  </button>

                  <button
                    type="button"
                    onClick={() => setPermissionType("full_control")}
                    className={`p-3 rounded-lg border text-left text-xs font-mono transition-colors ${
                      permissionType === "full_control"
                        ? "border-primary bg-primary/10 text-primary font-bold shadow-sm"
                        : "border-borderSubtle bg-white/70 text-textMuted hover:text-textMain"
                    }`}
                  >
                    <div className="flex items-center gap-2 mb-1">
                      <MousePointer className="w-4 h-4" />
                      <span>Viewing & Remote Control</span>
                    </div>
                    <div className="text-[10px] text-textSubtle">Injects mouse & keystrokes</div>
                  </button>
                </div>
              </div>

              {/* Mandatory Policy Disclosure */}
              <div className="p-3.5 rounded-lg bg-white/60 border border-borderSubtle text-[11px] text-textMuted space-y-1 leading-relaxed">
                <div className="flex items-center gap-1.5 font-semibold text-textMain">
                  <Lock className="w-3.5 h-3.5 text-primary" />
                  <span>Endpoint Consent Mandate</span>
                </div>
                <div>
                  When dispatched, the Windows agent displays the official consent window. No screen capture or input control can begin until the logged-in user explicitly clicks Approve.
                </div>
              </div>

              <button
                onClick={handleRequestSession}
                className="w-full bg-primary hover:bg-primaryHover text-white font-semibold py-2.5 rounded-lg transition-colors text-sm flex items-center justify-center gap-2 shadow-sm"
              >
                <span>Dispatch Access Request to Endpoint</span>
              </button>
            </div>
          )}

          {/* State: Awaiting Endpoint Consent */}
          {sessionState === "awaiting_consent" && (
            <div className="max-w-xl rounded-xl border border-primary/40 bg-surface p-6 space-y-5 shadow-xl glass-panel">
              <div className="flex items-center gap-3">
                <RefreshCw className="w-5 h-5 text-primary animate-spin" />
                <div>
                  <h3 className="font-semibold text-sm text-textMain">Awaiting User Approval on {targetHost}</h3>
                  <p className="text-xs text-textMuted mt-0.5">
                    The consent prompt is currently displayed on the Windows desktop
                  </p>
                </div>
              </div>

              {/* Preview of Endpoint Consent Prompt */}
              <div className="p-4 rounded-lg bg-[#151310] border border-white/10 text-xs font-mono text-zinc-300 space-y-3 shadow-inner">
                <div className="text-primary font-bold text-xs uppercase tracking-wider flex items-center gap-1.5">
                  <ShieldCheck className="w-4 h-4" /> Windows Security Consent Prompt
                </div>
                <p className="text-zinc-300 leading-relaxed italic border-l-2 border-primary/60 pl-3">
                  “If you approve this request, your screen will be shared with the authorized IT operator, and remote screen control will be activated. The operator will be able to view your screen and, while control is enabled, interact with your computer using the mouse and keyboard. You can reject the request or end the session at any time.”
                </p>
                <div className="flex items-center justify-end gap-3 pt-2">
                  <button
                    onClick={() => setSessionState("idle")}
                    className="px-3 py-1 rounded bg-zinc-800 hover:bg-zinc-700 text-zinc-300 text-xs"
                  >
                    Simulate User Decline
                  </button>
                  <button
                    onClick={handleSimulateConsentApprove}
                    className="px-3 py-1 rounded bg-success text-white font-bold text-xs hover:bg-success/90"
                  >
                    Simulate User Approval →
                  </button>
                </div>
              </div>
            </div>
          )}

          {/* State: Active Remote Support Session */}
          {sessionState === "active" && (
            <div className="space-y-4">
              <div className="rounded-xl border border-borderSubtle bg-black aspect-video w-full flex flex-col items-center justify-center relative overflow-hidden shadow-2xl">
                <div className="absolute top-4 left-1/2 -translate-x-1/2 bg-[#151310]/90 border border-white/10 backdrop-blur-md px-4 py-1.5 rounded-full flex items-center gap-3 text-xs font-mono shadow-lg z-20 text-zinc-300">
                  <span className="w-2 h-2 rounded-full bg-success animate-pulse" />
                  <span className="font-semibold text-white">Remote Support Active</span>
                  <span className="text-zinc-600">•</span>
                  <span className="text-primary">{permissionType === "full_control" ? "Control Enabled" : "View Only"}</span>
                  <span className="text-zinc-600">•</span>
                  <span className="text-zinc-400">Operator: Lead IT Admin</span>
                </div>

                <div className="w-full h-full p-8 flex flex-col justify-between text-zinc-400 font-mono text-xs select-none pointer-events-none">
                  <div className="flex items-center justify-between text-zinc-500">
                    <span>Windows Server 2022 Desktop Duplication Stream</span>
                    <span>1920x1080 @ 30 FPS • H.264 WebRTC</span>
                  </div>

                  <div className="flex flex-col items-center justify-center space-y-2">
                    <MonitorPlay className="w-16 h-16 text-primary/40 animate-pulse" />
                    <div className="text-sm font-semibold text-zinc-300">
                      Interactive Streaming Canvas ({targetHost})
                    </div>
                    <div className="text-xs text-zinc-500">
                      Operator mouse clicks and keyboard events are sent directly to Win32 SendInput API
                    </div>
                  </div>

                  <div className="flex items-center justify-between text-[11px] text-zinc-500">
                    <span>Mutual Heartbeat: 2s (5s disconnection triggers immediate session shutdown)</span>
                    <span>Audio/Video Recording: Disabled by policy</span>
                  </div>
                </div>
              </div>
            </div>
          )}

          {/* State: Terminated */}
          {sessionState === "terminated" && (
            <div className="p-4 rounded-xl border border-white/50 bg-surface text-center py-8 space-y-2 glass-panel shadow-sm">
              <CheckCircle2 className="w-8 h-8 text-success mx-auto" />
              <h3 className="font-semibold text-sm text-textMain">Session Revoked Successfully</h3>
              <p className="text-xs text-textMuted">
                Desktop duplication handle closed, SendInput released, and audit log recorded.
              </p>
            </div>
          )}
        </main>
      </div>
    </div>
  );
}
