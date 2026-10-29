"use client";

import { useEffect, useState, useRef, useCallback } from "react";
import { useParams } from "next/navigation";
import Link from "next/link";
import { DashboardSidebar } from "@/components/Sidebar";
import { AmbientBackground } from "@/components/AmbientBackground";
import {
  KeyRound,
  Download,
  CheckCircle2,
  Copy,
  Check,
  Sparkles,
  Terminal,
  ShieldCheck,
  Server,
  RefreshCw,
  Laptop,
  ArrowRight,
  Clock,
  UserCheck,
} from "lucide-react";
import { api, Device } from "@/lib/api";

interface Employee {
  id: number | string;
  full_name: string;
  email: string;
  department?: string;
}

export default function DeployAgentPage() {
  const params = useParams();
  const subdomain = (params?.subdomain as string) || "demo";

  // Wizard state
  const [step, setStep] = useState<1 | 2 | 3>(1);
  const [assignmentType, setAssignmentType] = useState<"none" | "existing" | "new">("none");
  const [employees, setEmployees] = useState<Employee[]>([]);
  const [selectedEmployeeId, setSelectedEmployeeId] = useState<string>("");
  const [newFullName, setNewFullName] = useState("");
  const [newEmail, setNewEmail] = useState("");
  const [newDepartment, setNewDepartment] = useState("");
  const [generating, setGenerating] = useState(false);

  // Step 2 state
  const [token, setToken] = useState("");
  const [remainingSeconds, setRemainingSeconds] = useState(15 * 60);
  const [copiedToken, setCopiedToken] = useState(false);
  const [copiedCommand, setCopiedCommand] = useState(false);

  // Step 3 / Detector state
  const [detectedDevice, setDetectedDevice] = useState<Device | null>(null);
  const [baselineDeviceIds, setBaselineDeviceIds] = useState<Set<string>>(new Set());
  const [pollCount, setPollCount] = useState(0);

  // Server URL
  const serverUrl = typeof window !== "undefined"
    ? `${window.location.protocol}//${subdomain}.monitopro.com`
    : `https://${subdomain}.monitopro.com`;

  // Silent deployment command
  const silentCommand = `Start-Process -Wait .\\MonitoProAgentSetup.exe "/VERYSILENT /SERVERURL=${serverUrl} /REGTOKEN=${token}"`;

  // Load existing employees for assignment
  useEffect(() => {
    async function fetchEmployees() {
      try {
        const data = await api.getDeviceUsers();
        setEmployees(data);
      } catch (e) {
        console.error("Failed to load employees:", e);
      }
    }
    fetchEmployees();
  }, []);

  // Step 1 -> Step 2: Generate Token
  const handleGenerateToken = async () => {
    setGenerating(true);
    try {
      let assignedUserId: string | number | null = null;
      if (assignmentType === "existing") {
        if (!selectedEmployeeId) {
          alert("Please select an employee to assign.");
          setGenerating(false);
          return;
        }
        assignedUserId = parseInt(selectedEmployeeId, 10);
      } else if (assignmentType === "new") {
        if (!newFullName.trim() || !newEmail.trim()) {
          alert("Please fill out required fields: Full Name and Email.");
          setGenerating(false);
          return;
        }
        const created = await api.createDeviceUser(newFullName.trim(), newEmail.trim(), newDepartment.trim());
        assignedUserId = created.id;
      }

      // Fetch baseline devices before transitioning so we know who is new
      try {
        const devRes = await api.listDevices();
        const ids = new Set((devRes.items || []).map((d) => d.id));
        setBaselineDeviceIds(ids);
      } catch {
        setBaselineDeviceIds(new Set());
      }

      const res = await api.createEnrollmentToken(assignedUserId);
      setToken(res.token);
      setRemainingSeconds(15 * 60);
      setStep(2);
    } catch (err: any) {
      alert("Error generating token: " + (err?.message || "Unknown error"));
    } finally {
      setGenerating(false);
    }
  };

  // 15-minute countdown timer in Step 2
  useEffect(() => {
    if (step !== 2) return;

    const timer = setInterval(() => {
      setRemainingSeconds((prev) => {
        if (prev <= 1) {
          clearInterval(timer);
          alert("Registration token has expired. Please generate a new one.");
          setStep(1);
          return 0;
        }
        return prev - 1;
      });
    }, 1000);

    return () => clearInterval(timer);
  }, [step]);

  // Live Heartbeat Detector Polling Loop
  useEffect(() => {
    if (step !== 2) return;

    const pollInterval = setInterval(async () => {
      setPollCount((c) => c + 1);
      try {
        const res = await api.listDevices();
        const currentDevices = res.items || [];

        // Check if any device is not in baseline, or has a heartbeat recorded just now
        const newDevice = currentDevices.find(
          (d) => !baselineDeviceIds.has(d.id) && (d.status === "online" || !!d.last_seen_at)
        );

        if (newDevice) {
          setDetectedDevice(newDevice);
          setStep(3);
          clearInterval(pollInterval);
        }
      } catch {
        // Continue polling silently
      }
    }, 2500);

    return () => clearInterval(pollInterval);
  }, [step, baselineDeviceIds]);

  // Copy helper
  const copyToClipboard = (text: string, type: "token" | "command") => {
    navigator.clipboard.writeText(text);
    if (type === "token") {
      setCopiedToken(true);
      setTimeout(() => setCopiedToken(false), 2000);
    } else {
      setCopiedCommand(true);
      setTimeout(() => setCopiedCommand(false), 2000);
    }
  };

  // Format countdown mm:ss
  const formatTime = (secs: number) => {
    const m = Math.floor(secs / 60);
    const s = secs % 60;
    return `${m.toString().padStart(2, "0")}:${s.toString().padStart(2, "0")}`;
  };

  // Demo simulation trigger for quick testing in web UI
  const simulateHeartbeat = () => {
    const mockNewDevice: Device = {
      id: "dev-" + Math.random().toString(36).substring(2, 8),
      hostname: "DESKTOP-WIN11-PRO",
      ip_address: "192.168.1.188",
      status: "online",
      os_version: "Windows 11 Pro 64-bit (Build 22631)",
      assigned_user_email:
        assignmentType === "new"
          ? newEmail || "user@company.com"
          : assignmentType === "existing"
          ? employees.find((e) => e.id.toString() === selectedEmployeeId)?.email || "employee@company.com"
          : "unassigned@company.com",
      last_seen_at: new Date().toISOString(),
    };
    setDetectedDevice(mockNewDevice);
    setStep(3);
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
            <span className="text-borderSubtle">/</span>
            <span className="text-textMain font-semibold">Deploy Agent</span>
          </div>

          <div className="flex items-center gap-2">
            <span className="text-[11px] font-mono text-textSubtle uppercase tracking-wider">Installer Version:</span>
            <span className="px-2 py-0.5 rounded bg-primary/10 border border-primary/20 text-primary text-[11px] font-mono font-semibold">
              v2.0.0 (Native Rust)
            </span>
          </div>
        </header>

        {/* Main Workspace */}
        <main className="p-8 flex-1 flex flex-col gap-6 max-w-[1000px] mx-auto w-full">
          <div className="flex flex-col gap-1">
            <h2 className="text-[20px] font-semibold text-textMain tracking-tight">Deploy Monitoring Agent</h2>
            <p className="text-[12px] text-textSubtle">
              Securely onboard a new Windows server or workstation using the professional agent installer.
            </p>
          </div>

          <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
            {/* Wizard Main Column (2 spans) */}
            <div className="md:col-span-2 flex flex-col gap-6">
              {/* STEP 1: Generate Onboarding Token */}
              {step === 1 && (
                <div className="bg-surface border border-borderSubtle rounded-lg p-6 glass-panel flex flex-col gap-4 shadow-sm">
                  <div className="flex items-start gap-4">
                    <div className="w-10 h-10 rounded-lg bg-primary/10 flex items-center justify-center text-primary flex-shrink-0">
                      <KeyRound className="w-5 h-5" />
                    </div>
                    <div className="flex flex-col gap-1">
                      <h3 className="font-bold text-[14px]">1. Generate Onboarding Token</h3>
                      <p className="text-[12px] text-textMuted leading-relaxed">
                        Before launching the installer, generate a secure one-time registration token. The token authorizes the installer to configure the daemon credentials.
                      </p>
                    </div>
                  </div>

                  {/* Device Assignment Option (Optional) */}
                  <div className="border-t border-borderSubtle/50 pt-4 flex flex-col gap-3">
                    <h4 className="text-[11px] font-bold text-textSubtle uppercase tracking-wider">
                      Device Assignment (Optional)
                    </h4>
                    <div className="flex flex-col gap-2">
                      <select
                        value={assignmentType}
                        onChange={(e) => setAssignmentType(e.target.value as any)}
                        className="bg-bg border border-borderSubtle rounded-md px-3 py-2 text-[12.5px] focus:outline-none focus:border-primary text-textMain cursor-pointer transition-all"
                      >
                        <option value="none">Unassigned / Generic Device</option>
                        <option value="existing">Assign to Existing Employee</option>
                        <option value="new">Create & Assign New Employee</option>
                      </select>
                    </div>

                    {/* Select Existing Employee */}
                    {assignmentType === "existing" && (
                      <div className="flex flex-col gap-1.5 animate-in fade-in duration-200">
                        <label className="text-[10px] font-mono uppercase tracking-wider text-textMuted">
                          Select Employee
                        </label>
                        <select
                          value={selectedEmployeeId}
                          onChange={(e) => setSelectedEmployeeId(e.target.value)}
                          className="bg-bg border border-borderSubtle rounded-md px-3 py-2 text-[12.5px] focus:outline-none focus:border-primary text-textMain"
                        >
                          <option value="">-- Choose Employee --</option>
                          {employees.map((emp) => (
                            <option key={emp.id} value={emp.id}>
                              {emp.full_name} ({emp.email}) - {emp.department || "No Dept"}
                            </option>
                          ))}
                        </select>
                      </div>
                    )}

                    {/* Create New Employee Fields */}
                    {assignmentType === "new" && (
                      <div className="flex flex-col gap-3 animate-in fade-in duration-200">
                        <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                          <div className="flex flex-col gap-1">
                            <label className="text-[10px] font-mono uppercase tracking-wider text-textMuted">
                              Full Name *
                            </label>
                            <input
                              type="text"
                              placeholder="e.g. Alice Smith"
                              value={newFullName}
                              onChange={(e) => setNewFullName(e.target.value)}
                              className="bg-bg border border-borderSubtle rounded-md px-3 py-1.5 text-[12.5px] focus:outline-none focus:border-primary text-textMain"
                            />
                          </div>
                          <div className="flex flex-col gap-1">
                            <label className="text-[10px] font-mono uppercase tracking-wider text-textMuted">
                              Email Address *
                            </label>
                            <input
                              type="email"
                              placeholder="e.g. alice.smith@company.com"
                              value={newEmail}
                              onChange={(e) => setNewEmail(e.target.value)}
                              className="bg-bg border border-borderSubtle rounded-md px-3 py-1.5 text-[12.5px] focus:outline-none focus:border-primary text-textMain"
                            />
                          </div>
                        </div>
                        <div className="flex flex-col gap-1">
                          <label className="text-[10px] font-mono uppercase tracking-wider text-textMuted">
                            Department
                          </label>
                          <input
                            type="text"
                            placeholder="e.g. Engineering"
                            value={newDepartment}
                            onChange={(e) => setNewDepartment(e.target.value)}
                            className="bg-bg border border-borderSubtle rounded-md px-3 py-1.5 text-[12.5px] focus:outline-none focus:border-primary text-textMain"
                          />
                        </div>
                      </div>
                    )}
                  </div>

                  <button
                    onClick={handleGenerateToken}
                    disabled={generating}
                    className="bg-primary text-white hover:bg-primaryHover py-2.5 px-4 rounded-md font-semibold text-[12px] transition-all flex items-center justify-center gap-2 shadow-sm mt-2 self-start disabled:opacity-50"
                  >
                    <Sparkles className="w-4 h-4" />
                    <span>{generating ? "Generating Token..." : "Generate Registration Token"}</span>
                  </button>
                </div>
              )}

              {/* STEP 2: Download & Run Installer */}
              {step === 2 && (
                <div className="bg-surface border border-borderSubtle rounded-lg p-6 glass-panel flex flex-col gap-5 shadow-sm animate-in fade-in duration-200">
                  <div className="flex items-start gap-4">
                    <div className="w-10 h-10 rounded-lg bg-primary/10 flex items-center justify-center text-primary flex-shrink-0">
                      <Download className="w-5 h-5" />
                    </div>
                    <div className="flex flex-col gap-1">
                      <h3 className="font-bold text-[14px]">2. Download & Run Installer</h3>
                      <p className="text-[12px] text-textMuted leading-relaxed">
                        Copy the onboarding token, download the installer package, and execute the installation. Paste the token when prompted by the wizard or deploy silently.
                      </p>
                    </div>
                  </div>

                  {/* Token Box */}
                  <div className="flex flex-col gap-2">
                    <label className="text-[10px] font-bold text-textSubtle uppercase tracking-wider">
                      Registration Token
                    </label>
                    <div className="flex gap-2">
                      <div className="flex-1 bg-[#0C0F0A] border border-black/10 rounded px-3 py-2 font-mono text-[13px] text-[#B87C3B] select-all flex items-center justify-between overflow-hidden">
                        <span className="truncate">{token}</span>
                        <span className="text-[10px] text-textSubtle bg-white/5 px-2 py-0.5 rounded ml-2 flex-shrink-0">
                          15m expiry
                        </span>
                      </div>
                      <button
                        onClick={() => copyToClipboard(token, "token")}
                        className="px-4 py-2 border border-borderSubtle bg-surface hover:bg-black/[0.02] rounded text-[12px] font-medium transition-all flex items-center gap-1.5 shadow-sm flex-shrink-0"
                      >
                        {copiedToken ? (
                          <>
                            <Check className="w-3.5 h-3.5 text-success" />
                            <span className="text-success font-semibold">Copied!</span>
                          </>
                        ) : (
                          <>
                            <Copy className="w-3.5 h-3.5" />
                            <span>Copy</span>
                          </>
                        )}
                      </button>
                    </div>
                  </div>

                  {/* 15m Expiration Countdown Banner */}
                  <div className="flex items-center gap-2 bg-amber-500/10 border border-amber-500/20 rounded-lg p-3 text-amber-800 text-[12px] leading-relaxed">
                    <Clock className="w-4 h-4 text-amber-600 flex-shrink-0" />
                    <span>
                      Token expires in{" "}
                      <strong className="font-mono text-amber-900 bg-amber-100 px-1.5 py-0.5 rounded">
                        {formatTime(remainingSeconds)}
                      </strong>
                      . Complete the installation before this time.
                    </span>
                  </div>

                  {/* Download Action */}
                  <a
                    href={`/api/deploy/installer?token=${token}`}
                    download="MonitoProAgentSetup-v2.0.0.exe"
                    className="bg-primary text-white hover:bg-primaryHover py-3 px-4 rounded-md font-semibold text-[13px] transition-all flex items-center justify-center gap-2 shadow-sm text-center"
                  >
                    <Download className="w-4 h-4" />
                    <span>Download Windows Installer (.exe)</span>
                  </a>

                  {/* Silent PowerShell Deployment Command */}
                  <div className="flex flex-col gap-2 pt-2 border-t border-borderSubtle/50">
                    <div className="flex items-center justify-between">
                      <label className="text-[10px] font-bold text-textSubtle uppercase tracking-wider flex items-center gap-1.5">
                        <Terminal className="w-3.5 h-3.5" /> Silent PowerShell Deployment (Automated / RMM)
                      </label>
                      <span className="text-[10px] text-textSubtle font-mono">Run in Admin PowerShell</span>
                    </div>

                    <div className="flex gap-2">
                      <div className="flex-1 bg-[#151310] border border-white/5 rounded p-2.5 font-mono text-[11px] text-zinc-300 overflow-x-auto select-all leading-relaxed whitespace-pre-wrap">
                        {silentCommand}
                      </div>
                      <button
                        onClick={() => copyToClipboard(silentCommand, "command")}
                        className="px-3 py-2 border border-borderSubtle bg-surface hover:bg-black/[0.02] rounded text-[12px] font-medium transition-all flex items-center gap-1.5 shadow-sm flex-shrink-0 self-start"
                        title="Copy PowerShell Command"
                      >
                        {copiedCommand ? (
                          <>
                            <Check className="w-3.5 h-3.5 text-success" />
                            <span className="text-success font-semibold">Copied!</span>
                          </>
                        ) : (
                          <>
                            <Copy className="w-3.5 h-3.5" />
                            <span>Copy</span>
                          </>
                        )}
                      </button>
                    </div>
                  </div>

                  {/* Polling Heartbeat Detector Indicator */}
                  <div className="flex items-center justify-between p-3 rounded-lg bg-primary/5 border border-primary/20 text-[11.5px] mt-1">
                    <div className="flex items-center gap-2.5">
                      <div className="w-2.5 h-2.5 rounded-full bg-primary animate-ping" />
                      <span className="text-textMain font-medium">
                        Listening for live agent telemetry ping... (checked {pollCount} times)
                      </span>
                    </div>
                    {/* Quick simulation helper for demonstration */}
                    <button
                      onClick={simulateHeartbeat}
                      className="text-[10px] font-mono text-primary hover:underline bg-white/70 px-2 py-1 rounded border border-primary/20"
                      title="Simulate successful agent installation"
                    >
                      Simulate Heartbeat
                    </button>
                  </div>
                </div>
              )}

              {/* STEP 3: Success Telemetry Received */}
              {step === 3 && (
                <div className="bg-surface border border-borderSubtle rounded-lg p-8 glass-panel flex flex-col items-center text-center gap-5 shadow-sm animate-in zoom-in-95 duration-300">
                  <div className="w-14 h-14 rounded-full bg-success/15 text-success flex items-center justify-center animate-bounce shadow-sm">
                    <CheckCircle2 className="w-8 h-8" />
                  </div>

                  <div className="flex flex-col gap-1.5 max-w-md">
                    <h3 className="font-bold text-[18px] text-textMain tracking-tight">
                      Agent Registered Successfully!
                    </h3>
                    <p className="text-[12.5px] text-textMuted leading-relaxed">
                      The Windows background service was registered, started, and has successfully established its first secure HTTPS telemetry heartbeat with the dashboard.
                    </p>
                  </div>

                  {/* Detected Device Details Card */}
                  {detectedDevice && (
                    <div className="w-full bg-[#151310] text-zinc-300 border border-white/5 rounded-lg p-4 font-mono text-[12px] flex flex-col gap-2.5 text-left shadow-inner">
                      <div className="flex justify-between items-center border-b border-white/5 pb-2">
                        <span className="text-zinc-500 uppercase tracking-wider text-[10px]">Endpoint Node</span>
                        <span className="px-2 py-0.5 rounded bg-emerald-500/20 text-emerald-400 border border-emerald-500/30 text-[10px] uppercase font-bold flex items-center gap-1.5">
                          <span className="w-1.5 h-1.5 rounded-full bg-emerald-400 animate-pulse" />
                          Online
                        </span>
                      </div>
                      <div className="grid grid-cols-2 gap-2 text-[11.5px]">
                        <div>
                          <span className="text-zinc-500">Hostname: </span>
                          <span className="text-white font-semibold">{detectedDevice.hostname}</span>
                        </div>
                        <div>
                          <span className="text-zinc-500">IP Address: </span>
                          <span className="text-zinc-200">{detectedDevice.ip_address}</span>
                        </div>
                        <div>
                          <span className="text-zinc-500">OS: </span>
                          <span className="text-zinc-200 truncate">{detectedDevice.os_version}</span>
                        </div>
                        <div>
                          <span className="text-zinc-500">Assigned User: </span>
                          <span className="text-zinc-200">{detectedDevice.assigned_user_email || "Unassigned"}</span>
                        </div>
                      </div>
                      <div className="text-[10px] text-zinc-500 pt-1 border-t border-white/5">
                        First telemetry ping: {new Date(detectedDevice.last_seen_at || Date.now()).toLocaleTimeString()}
                      </div>
                    </div>
                  )}

                  <div className="flex flex-col sm:flex-row items-center gap-3 w-full justify-center pt-2">
                    <Link
                      href={`/${subdomain}/devices`}
                      className="bg-success text-white hover:bg-success/90 px-6 py-2.5 rounded-md font-semibold text-[12.5px] transition-all flex items-center justify-center gap-2 shadow-sm w-full sm:w-auto"
                    >
                      <span>Go to Infrastructure Console</span>
                      <ArrowRight className="w-4 h-4" />
                    </Link>
                    <button
                      onClick={() => {
                        setStep(1);
                        setToken("");
                        setDetectedDevice(null);
                      }}
                      className="border border-borderSubtle bg-white/60 text-textMuted hover:text-textMain px-4 py-2.5 rounded-md text-[12.5px] font-medium transition-all w-full sm:w-auto"
                    >
                      Deploy Another Agent
                    </button>
                  </div>
                </div>
              )}
            </div>

            {/* Instruction Checklist & Deployment Status (1 span) */}
            <div className="flex flex-col gap-5">
              {/* Installation Guide Card matching V1 */}
              <div className="bg-surface border border-borderSubtle rounded-lg p-5 glass-panel flex flex-col gap-4 shadow-sm">
                <h3 className="font-bold text-[13px] text-textMain border-b border-borderSubtle pb-2">
                  Installation Guide
                </h3>

                <ul className="flex flex-col gap-4 text-[12px] text-textMuted leading-relaxed">
                  <li className="flex gap-2.5">
                    <span className="text-primary font-bold">1.</span>
                    <div>
                      <strong className="text-textMain">Run Installer</strong>
                      <br />
                      Run <code className="bg-black/5 font-mono px-1 rounded text-[11px]">MonitoProAgentSetup.exe</code> with Administrator privileges.
                    </div>
                  </li>
                  <li className="flex gap-2.5">
                    <span className="text-primary font-bold">2.</span>
                    <div>
                      <strong className="text-textMain">Configure Token</strong>
                      <br />
                      Paste the generated token and the target dashboard server URL when prompted.
                    </div>
                  </li>
                  <li className="flex gap-2.5">
                    <span className="text-primary font-bold">3.</span>
                    <div>
                      <strong className="text-textMain">Check Status</strong>
                      <br />
                      The installer registers the Windows Service, starts the daemon, and provisions the Ed25519 node identity automatically.
                    </div>
                  </li>
                </ul>

                {/* Deployment Status Indicator */}
                <div className="border-t border-borderSubtle pt-4 flex flex-col gap-2.5">
                  <span className="text-[10px] font-bold text-textSubtle uppercase tracking-wider">
                    Deployment Status
                  </span>
                  <div className="flex items-center gap-2 text-[11.5px] font-medium">
                    <div
                      className={`w-2 h-2 rounded-full ${
                        step === 1
                          ? "bg-zinc-400"
                          : step === 2
                          ? "bg-primary animate-ping"
                          : "bg-success shadow-[0_0_8px_rgba(16,185,129,0.5)]"
                      }`}
                    />
                    <span className="text-textMain font-semibold">
                      {step === 1
                        ? "No active deployment"
                        : step === 2
                        ? "Waiting for agent to report..."
                        : "Onboarding complete!"}
                    </span>
                  </div>
                </div>
              </div>

              {/* Cryptographic Security Details */}
              <div className="bg-surface border border-borderSubtle rounded-lg p-5 glass-panel flex flex-col gap-3 shadow-sm text-[12px] text-textMuted">
                <div className="flex items-center gap-2 text-textMain font-semibold text-[12.5px]">
                  <ShieldCheck className="w-4 h-4 text-emerald-600" />
                  <span>Ed25519 Mutual Auth</span>
                </div>
                <p className="text-[11.5px] leading-relaxed text-textSubtle">
                  Endpoints generate an isolated Ed25519 keypair locally. All subsequent telemetry metrics and remote support sessions are cryptographically signed.
                </p>
              </div>
            </div>
          </div>
        </main>
      </div>
    </div>
  );
}
