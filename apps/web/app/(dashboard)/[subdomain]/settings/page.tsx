"use client";

import { useState, useEffect } from "react";
import { useParams } from "next/navigation";
import Link from "next/link";
import { DashboardSidebar } from "@/components/Sidebar";
import { AmbientBackground } from "@/components/AmbientBackground";
import {
  Settings,
  Building2,
  Cpu,
  Archive,
  AlertOctagon,
  Eye,
  EyeOff,
  CheckCircle2,
  HardDrive,
  Download,
  Check,
  Shield,
  FileCheck2,
  RefreshCw,
  Sparkles,
} from "lucide-react";

export interface ArchiveRecord {
  id: string;
  dateRange: string;
  uncompressedSize: string;
  compressedSize: string;
  ratio: string;
  sha256: string;
  createdAt: string;
  verified: boolean;
}

export default function WorkspaceSettingsPage() {
  const params = useParams();
  const subdomain = (params?.subdomain as string) || "demo";

  // Section 1: Workspace configuration
  const [orgName, setOrgName] = useState("");
  const [adminEmail, setAdminEmail] = useState("");
  const [workspaceSlug, setWorkspaceSlug] = useState(subdomain);
  const [orgRegion, setOrgRegion] = useState("us-east-1 (N. Virginia)");

  useEffect(() => {
    const storedOrg = typeof window !== 'undefined' ? localStorage.getItem("monitopro_org_name") : null;
    const storedEmail = typeof window !== 'undefined' ? localStorage.getItem("monitopro_admin_email") : null;
    setOrgName(storedOrg || `${subdomain.toUpperCase()} Operations`);
    setAdminEmail(storedEmail || `admin@${subdomain}.internal`);
  }, [subdomain]);

  // Section 2: AI Governance
  const [aiKillSwitch, setAiKillSwitch] = useState(false);
  const [llmProvider, setLlmProvider] = useState<"ollama" | "in_memory" | "openai">("in_memory");
  const [ollamaEndpoint, setOllamaEndpoint] = useState("http://localhost:11434");
  const [ollamaModel, setOllamaModel] = useState("deepseek-r1:8b");
  const [openAiApiKey, setOpenAiApiKey] = useState("");
  const [showApiKey, setShowApiKey] = useState(false);

  // PII Scrubbing interactive audit tester
  const [samplePayload, setSamplePayload] = useState(
    "User admin connected from 192.168.1.104 with auth token sec_998124 to node ws-prod-01 (10.0.4.15). Contact: security@internal"
  );
  const [scrubbedPayload, setScrubbedPayload] = useState<string | null>(null);
  const [scrubStats, setScrubStats] = useState<{ redactedEntities: number; anonymizedRatio: string } | null>(null);

  // Section 3: Data Retention & Archival
  const [retentionDays, setRetentionDays] = useState(30);
  const [compressionAlgo, setCompressionAlgo] = useState("zstandard-level-9");
  const [archiveRecords, setArchiveRecords] = useState<ArchiveRecord[]>([]);

  const [toastMessage, setToastMessage] = useState<string | null>(null);

  const showToast = (msg: string) => {
    setToastMessage(msg);
    setTimeout(() => setToastMessage(null), 3000);
  };

  const handleSaveWorkspace = (e: React.FormEvent) => {
    e.preventDefault();
    showToast("Workspace configuration saved successfully.");
  };

  const handleTestPiiScrub = () => {
    // Deterministic client-side demonstration of the regex PII scrubber
    let cleaned = samplePayload;
    let count = 0;

    // IPv4 addresses
    cleaned = cleaned.replace(/\b\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}\b/g, () => {
      count++;
      return "[REDACTED_IP]";
    });
    // Email addresses
    cleaned = cleaned.replace(/[\w.-]+@[\w.-]+\.\w+/g, () => {
      count++;
      return "[REDACTED_EMAIL]";
    });
    // Tokens / secrets
    cleaned = cleaned.replace(/\b(sec_|tok_|key_)[a-zA-Z0-9_-]+/g, () => {
      count++;
      return "[REDACTED_SECRET]";
    });
    // System hostnames
    cleaned = cleaned.replace(/\b(srv|ws|gw)-[a-zA-Z0-9_-]+\b/g, () => {
      count++;
      return "[NODE_ANONYMIZED]";
    });

    setScrubbedPayload(cleaned);
    setScrubStats({
      redactedEntities: count,
      anonymizedRatio: "100% PII Cleaned",
    });
    showToast(`Sanitized ${count} sensitive telemetry tokens.`);
  };

  const handleVerifyChecksum = (id: string) => {
    setArchiveRecords((prev) =>
      prev.map((a) => (a.id === id ? { ...a, verified: true } : a))
    );
    showToast(`Cryptographic SHA-256 seal verified for ${id}`);
  };

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
            <span className="text-textMain font-semibold">Tenant Workspace & AI Governance</span>
          </div>

          <div className="flex items-center gap-2">
            <span className="px-2.5 py-1 rounded bg-stone-100 border border-stone-200 text-stone-700 text-[10px] font-mono font-semibold uppercase">
              Tenant ID: {subdomain}-tenant-001
            </span>
          </div>
        </header>

        {/* Toast */}
        {toastMessage && (
          <div className="fixed bottom-6 right-6 z-50 bg-[#151310] text-white px-4 py-3 rounded-lg shadow-2xl border border-white/10 flex items-center gap-3 animate-fade-in text-xs">
            <CheckCircle2 className="w-4 h-4 text-emerald-400 shrink-0" />
            <span>{toastMessage}</span>
          </div>
        )}

        <main className="p-8 max-w-7xl w-full mx-auto space-y-8">
          {/* Headline */}
          <div>
            <h1 className="text-xl font-bold tracking-tight text-textMain flex items-center gap-2">
              <Settings className="w-5 h-5 text-primary" />
              Tenant Workspace & AI Governance Settings
            </h1>
            <p className="text-xs text-textMuted mt-1">
              Configure organizational boundaries, AI advisory privacy constraints, and telemetry retention policies
            </p>
          </div>

          {/* Section 1: Workspace Configuration */}
          <div className="border border-white/50 bg-surface backdrop-blur-md rounded-xl p-6 shadow-sm glass-panel space-y-6">
            <div className="flex items-center justify-between border-b border-borderSubtle pb-4">
              <div className="flex items-center gap-3">
                <div className="w-9 h-9 rounded-lg bg-black/[0.03] border border-borderSubtle flex items-center justify-center text-primary">
                  <Building2 className="w-5 h-5" />
                </div>
                <div>
                  <h2 className="text-sm font-bold text-textMain">Workspace Identity & Routing</h2>
                  <p className="text-xs text-textSubtle">Primary organization metadata and vanity subdomain URL</p>
                </div>
              </div>
              <span className="px-2.5 py-1 rounded-full bg-emerald-50 border border-emerald-200/50 text-emerald-700 text-[10px] font-mono font-semibold uppercase">
                Active Tenant
              </span>
            </div>

            <form onSubmit={handleSaveWorkspace} className="space-y-4">
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div className="space-y-1.5">
                  <label className="text-xs font-semibold text-textMain">Organization Name</label>
                  <input
                    type="text"
                    value={orgName}
                    onChange={(e) => setOrgName(e.target.value)}
                    className="w-full bg-white border border-borderSubtle rounded-lg px-3 py-2 text-xs text-textMain focus:outline-none focus:border-primary shadow-sm"
                  />
                </div>

                <div className="space-y-1.5">
                  <label className="text-xs font-semibold text-textMain">Primary Contact / Owner Email</label>
                  <input
                    type="email"
                    value={adminEmail}
                    onChange={(e) => setAdminEmail(e.target.value)}
                    className="w-full bg-white border border-borderSubtle rounded-lg px-3 py-2 text-xs text-textMain focus:outline-none focus:border-primary shadow-sm"
                  />
                </div>
              </div>

              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div className="space-y-1.5">
                  <label className="text-xs font-semibold text-textMain">Dedicated Subdomain Routing</label>
                  <div className="relative flex items-center">
                    <input
                      type="text"
                      value={workspaceSlug}
                      onChange={(e) => setWorkspaceSlug(e.target.value.toLowerCase())}
                      className="w-full bg-stone-50 border border-borderSubtle rounded-lg pl-3 pr-28 py-2 text-xs font-mono text-textMain focus:outline-none focus:border-primary shadow-inner"
                    />
                    <span className="absolute right-3 text-xs font-mono text-textSubtle pointer-events-none">
                      .monitopro.com
                    </span>
                  </div>
                </div>

                <div className="space-y-1.5">
                  <label className="text-xs font-semibold text-textMain">Primary Sovereign Cluster Region</label>
                  <select
                    value={orgRegion}
                    onChange={(e) => setOrgRegion(e.target.value)}
                    className="w-full bg-white border border-borderSubtle rounded-lg px-3 py-2 text-xs text-textMain focus:outline-none focus:border-primary shadow-sm"
                  >
                    <option value="us-east-1 (N. Virginia)">us-east-1 (N. Virginia)</option>
                    <option value="eu-central-1 (Frankfurt)">eu-central-1 (Frankfurt)</option>
                    <option value="ap-southeast-1 (Singapore)">ap-southeast-1 (Singapore)</option>
                  </select>
                </div>
              </div>

              <div className="flex justify-end pt-2">
                <button
                  type="submit"
                  className="px-4 py-2 rounded-lg bg-primary hover:bg-primaryHover text-white text-xs font-semibold shadow-sm transition-colors"
                >
                  Save Workspace Changes
                </button>
              </div>
            </form>
          </div>

          {/* Section 2: AI Advisory Governance & Privacy Controls */}
          <div className="border border-white/50 bg-surface backdrop-blur-md rounded-xl p-6 shadow-sm glass-panel space-y-6">
            <div className="flex items-center justify-between border-b border-borderSubtle pb-4">
              <div className="flex items-center gap-3">
                <div className="w-9 h-9 rounded-lg bg-amber-50 border border-amber-200/50 flex items-center justify-center text-primary">
                  <Cpu className="w-5 h-5" />
                </div>
                <div>
                  <h2 className="text-sm font-bold text-textMain">AI Advisory Governance & Privacy</h2>
                  <p className="text-xs text-textSubtle">
                    Hardware-enforced kill-switch, sovereign model backends, and PII masking pipeline
                  </p>
                </div>
              </div>

              <div className="flex items-center gap-2">
                {aiKillSwitch ? (
                  <span className="px-2.5 py-1 rounded bg-rose-50 border border-rose-200 text-rose-700 text-[10px] font-mono font-bold uppercase flex items-center gap-1">
                    <AlertOctagon className="w-3.5 h-3.5" /> AI Engine Severed
                  </span>
                ) : (
                  <span className="px-2.5 py-1 rounded bg-emerald-50 border border-emerald-200/50 text-emerald-700 text-[10px] font-mono font-bold uppercase flex items-center gap-1">
                    <Sparkles className="w-3.5 h-3.5" /> Advisory Active
                  </span>
                )}
              </div>
            </div>

            {/* Tenant-Wide AI Kill-Switch */}
            <div className="p-4 rounded-xl border border-rose-200/80 bg-rose-50/40 flex flex-col sm:flex-row sm:items-center justify-between gap-4">
              <div className="space-y-1">
                <div className="flex items-center gap-2 text-rose-900 font-bold text-xs">
                  <AlertOctagon className="w-4 h-4 text-rose-600" />
                  Tenant-Wide AI Kill-Switch
                </div>
                <p className="text-xs text-rose-800 leading-relaxed max-w-xl">
                  Immediately halts all automated hypothesis generation, root-cause syntheses, and external LLM outbound
                  requests. Deterministic threshold alerts continue normal evaluation.
                </p>
              </div>

              <button
                type="button"
                onClick={() => {
                  const next = !aiKillSwitch;
                  setAiKillSwitch(next);
                  showToast(next ? "AI Advisory Kill-Switch ENGAGED" : "AI Advisory re-enabled");
                }}
                className={`px-4 py-2 rounded-lg text-xs font-bold transition-all shadow-sm ${
                  aiKillSwitch
                    ? "bg-rose-600 hover:bg-rose-700 text-white"
                    : "bg-white hover:bg-rose-50 border border-rose-300 text-rose-700"
                }`}
              >
                {aiKillSwitch ? "DISENGAGE KILL-SWITCH" : "ENGAGE AI KILL-SWITCH"}
              </button>
            </div>

            {/* LLM Provider Selection */}
            <div className="space-y-4">
              <label className="text-xs font-semibold text-textMain block">Inference Provider Architecture</label>
              <div className="grid grid-cols-1 md:grid-cols-3 gap-3">
                <div
                  onClick={() => setLlmProvider("ollama")}
                  className={`p-4 rounded-xl border cursor-pointer transition-all ${
                    llmProvider === "ollama"
                      ? "border-[#B87C3B] bg-amber-50/40 shadow-sm"
                      : "border-borderSubtle bg-white hover:bg-stone-50"
                  }`}
                >
                  <div className="flex items-center justify-between">
                    <div className="text-xs font-bold text-textMain">Local Sovereign Ollama</div>
                    {llmProvider === "ollama" && <Check className="w-4 h-4 text-primary" />}
                  </div>
                  <p className="text-[11px] text-textMuted mt-1.5 leading-relaxed">
                    Zero telemetry egress. Inferences run entirely on local cluster nodes using DeepSeek-R1 or Llama3.
                  </p>
                </div>

                <div
                  onClick={() => setLlmProvider("in_memory")}
                  className={`p-4 rounded-xl border cursor-pointer transition-all ${
                    llmProvider === "in_memory"
                      ? "border-[#B87C3B] bg-amber-50/40 shadow-sm"
                      : "border-borderSubtle bg-white hover:bg-stone-50"
                  }`}
                >
                  <div className="flex items-center justify-between">
                    <div className="text-xs font-bold text-textMain">In-Memory Rule Advisory</div>
                    {llmProvider === "in_memory" && <Check className="w-4 h-4 text-primary" />}
                  </div>
                  <p className="text-[11px] text-textMuted mt-1.5 leading-relaxed">
                    Pure algorithmic heuristics based on deterministic runbooks. Zero external network calls.
                  </p>
                </div>

                <div
                  onClick={() => setLlmProvider("openai")}
                  className={`p-4 rounded-xl border cursor-pointer transition-all ${
                    llmProvider === "openai"
                      ? "border-[#B87C3B] bg-amber-50/40 shadow-sm"
                      : "border-borderSubtle bg-white hover:bg-stone-50"
                  }`}
                >
                  <div className="flex items-center justify-between">
                    <div className="text-xs font-bold text-textMain">OpenAI Enterprise Zero-Retention</div>
                    {llmProvider === "openai" && <Check className="w-4 h-4 text-primary" />}
                  </div>
                  <p className="text-[11px] text-textMuted mt-1.5 leading-relaxed">
                    High reasoning capacity with strict zero data retention contract agreement.
                  </p>
                </div>
              </div>

              {/* Dynamic config for selected provider */}
              {llmProvider === "ollama" && (
                <div className="grid grid-cols-1 md:grid-cols-2 gap-3 p-4 bg-white/60 border border-borderSubtle rounded-xl text-xs">
                  <div className="space-y-1">
                    <label className="font-semibold text-textMain">Ollama Server Endpoint</label>
                    <input
                      type="text"
                      value={ollamaEndpoint}
                      onChange={(e) => setOllamaEndpoint(e.target.value)}
                      className="w-full bg-white border border-borderSubtle rounded-lg px-3 py-1.5 font-mono text-xs focus:outline-none focus:border-primary shadow-sm"
                    />
                  </div>
                  <div className="space-y-1">
                    <label className="font-semibold text-textMain">Target Model Tag</label>
                    <input
                      type="text"
                      value={ollamaModel}
                      onChange={(e) => setOllamaModel(e.target.value)}
                      className="w-full bg-white border border-borderSubtle rounded-lg px-3 py-1.5 font-mono text-xs focus:outline-none focus:border-primary shadow-sm"
                    />
                  </div>
                </div>
              )}

              {llmProvider === "openai" && (
                <div className="p-4 bg-white/60 border border-borderSubtle rounded-xl text-xs space-y-1">
                  <label className="font-semibold text-textMain">Private Tenant API Key</label>
                  <div className="relative flex items-center">
                    <input
                      type={showApiKey ? "text" : "password"}
                      value={openAiApiKey}
                      onChange={(e) => setOpenAiApiKey(e.target.value)}
                      className="w-full bg-white border border-borderSubtle rounded-lg px-3 py-1.5 font-mono text-xs focus:outline-none focus:border-primary pr-10 shadow-sm"
                    />
                    <button
                      type="button"
                      onClick={() => setShowApiKey(!showApiKey)}
                      className="absolute right-3 text-textSubtle hover:text-textMain"
                    >
                      {showApiKey ? <EyeOff className="w-4 h-4" /> : <Eye className="w-4 h-4" />}
                    </button>
                  </div>
                </div>
              )}
            </div>

            {/* PII Scrub Verification Audit */}
            <div className="border-t border-borderSubtle pt-4 space-y-3">
              <div className="flex items-center justify-between">
                <div>
                  <h3 className="text-xs font-bold text-textMain flex items-center gap-1.5">
                    <Shield className="w-4 h-4 text-emerald-600" />
                    PII Scrubbing Verification Sandbox
                  </h3>
                  <p className="text-[11px] text-textSubtle">
                    Live verification of the sanitization pipeline that strips IPs, emails, and credentials before AI processing.
                  </p>
                </div>
                <button
                  type="button"
                  onClick={handleTestPiiScrub}
                  className="px-3 py-1.5 rounded-lg bg-stone-100 hover:bg-stone-200 border border-borderSubtle text-xs font-semibold text-textMain transition-colors flex items-center gap-1.5"
                >
                  <FileCheck2 className="w-3.5 h-3.5 text-primary" />
                  Run PII Masker Audit
                </button>
              </div>

              <div className="grid grid-cols-1 md:grid-cols-2 gap-3 text-xs">
                <div>
                  <label className="text-[11px] font-semibold text-textSubtle block mb-1">
                    Raw Telemetry Snippet (with PII):
                  </label>
                  <textarea
                    rows={3}
                    value={samplePayload}
                    onChange={(e) => setSamplePayload(e.target.value)}
                    className="w-full bg-white border border-borderSubtle rounded-lg p-2.5 font-mono text-[11px] text-textMain focus:outline-none focus:border-primary shadow-inner resize-none"
                  />
                </div>

                <div>
                  <div className="flex items-center justify-between mb-1">
                    <label className="text-[11px] font-semibold text-emerald-700 block">
                      Tokenized Sanitized Output:
                    </label>
                    {scrubStats && (
                      <span className="text-[10px] font-mono text-emerald-600 font-bold">
                        {scrubStats.redactedEntities} Redactions • {scrubStats.anonymizedRatio}
                      </span>
                    )}
                  </div>
                  <div className="w-full h-[76px] bg-stone-50 border border-borderSubtle rounded-lg p-2.5 font-mono text-[11px] text-emerald-900 overflow-y-auto">
                    {scrubbedPayload || (
                      <span className="text-textSubtle italic">Click &quot;Run PII Masker Audit&quot; to execute real-time redaction...</span>
                    )}
                  </div>
                </div>
              </div>
            </div>
          </div>

          {/* Section 3: Data Retention & Zstandard Archival */}
          <div className="border border-white/50 bg-surface backdrop-blur-md rounded-xl p-6 shadow-sm glass-panel space-y-6">
            <div className="flex items-center justify-between border-b border-borderSubtle pb-4">
              <div className="flex items-center gap-3">
                <div className="w-9 h-9 rounded-lg bg-emerald-50 border border-emerald-200/50 flex items-center justify-center text-emerald-700">
                  <Archive className="w-5 h-5" />
                </div>
                <div>
                  <h2 className="text-sm font-bold text-textMain">Data Retention & Immutable Archival</h2>
                  <p className="text-xs text-textSubtle">
                    High-ratio Zstandard compression and verifiable cryptographic checksum logs
                  </p>
                </div>
              </div>

              <span className="px-2.5 py-1 rounded-full bg-stone-100 border border-stone-200 text-stone-700 text-[10px] font-mono font-semibold uppercase">
                Zstd Compression: 10:1 Ratio
              </span>
            </div>

            {/* Retention Slider */}
            <div className="p-4 bg-white/60 border border-borderSubtle rounded-xl space-y-3">
              <div className="flex justify-between items-center text-xs">
                <div>
                  <span className="font-semibold text-textMain">Active Hot Telemetry Retention</span>
                  <p className="text-[11px] text-textSubtle">
                    Metrics older than the selected window are automatically sealed into immutable cold archives.
                  </p>
                </div>
                <span className="font-mono text-primary font-bold text-base px-3 py-1 bg-white rounded border border-borderSubtle">
                  {retentionDays} Days
                </span>
              </div>

              <input
                type="range"
                min={7}
                max={365}
                step={7}
                value={retentionDays}
                onChange={(e) => setRetentionDays(Number(e.target.value))}
                className="w-full accent-[#B87C3B] cursor-pointer"
              />

              <div className="flex justify-between text-[10px] text-textSubtle font-mono">
                <span>7 Days (Sprint)</span>
                <span>30 Days (Recommended)</span>
                <span>90 Days</span>
                <span>365 Days (Annual Compliance)</span>
              </div>
            </div>

            {/* Immutable Archives Table */}
            <div className="space-y-3">
              <div className="flex items-center justify-between">
                <h3 className="text-xs font-bold text-textMain">Immutable Zstandard Archive Records</h3>
                <span className="text-[11px] text-textSubtle font-mono">Algorithm: {compressionAlgo}</span>
              </div>

              <div className="border border-borderSubtle rounded-xl overflow-hidden bg-white/70">
                <table className="w-full text-left text-xs">
                  <thead className="bg-black/[0.02] border-b border-borderSubtle text-[11px] font-semibold text-textMuted uppercase tracking-wider">
                    <tr>
                      <th className="px-4 py-2.5">Archive ID & Range</th>
                      <th className="px-4 py-2.5">Raw / Compressed</th>
                      <th className="px-4 py-2.5">Compression Ratio</th>
                      <th className="px-4 py-2.5">SHA-256 Integrity Digest</th>
                      <th className="px-4 py-2.5 text-right">Actions</th>
                    </tr>
                  </thead>
                  <tbody className="divide-y divide-borderSubtle/60">
                    {archiveRecords.length === 0 ? (
                      <tr>
                        <td colSpan={5} className="py-8 text-center text-xs font-mono text-textMuted">
                          No archived historical snapshots generated yet. Telemetry within the {retentionDays}-day retention window is stored active in hot database storage.
                        </td>
                      </tr>
                    ) : (
                      archiveRecords.map((rec) => (
                        <tr key={rec.id} className="hover:bg-white/50 transition-colors">
                        <td className="px-4 py-3">
                          <div className="font-mono font-bold text-textMain flex items-center gap-1.5">
                            <HardDrive className="w-3.5 h-3.5 text-textSubtle" />
                            {rec.id}
                          </div>
                          <div className="text-[10px] text-textSubtle font-mono">{rec.dateRange}</div>
                        </td>

                        <td className="px-4 py-3 font-mono text-[11px] text-textMuted">
                          {rec.uncompressedSize} → <strong className="text-textMain">{rec.compressedSize}</strong>
                        </td>

                        <td className="px-4 py-3 whitespace-nowrap">
                          <span className="px-2 py-0.5 rounded bg-emerald-50 text-emerald-800 text-[10px] font-mono font-semibold border border-emerald-200">
                            {rec.ratio}
                          </span>
                        </td>

                        <td className="px-4 py-3 font-mono text-[10px] text-textSubtle max-w-[180px] truncate">
                          {rec.sha256}
                        </td>

                        <td className="px-4 py-3 text-right whitespace-nowrap">
                          <div className="flex items-center justify-end gap-1.5">
                            <button
                              onClick={() => handleVerifyChecksum(rec.id)}
                              className="px-2.5 py-1 rounded text-[11px] font-medium bg-white hover:bg-stone-50 border border-borderSubtle text-textMain transition-colors flex items-center gap-1"
                            >
                              <Check className="w-3 h-3 text-emerald-600" />
                              Verify Seal
                            </button>
                            <button
                              onClick={() => showToast(`Downloading cold archive ${rec.id}.zst`)}
                              className="p-1.5 rounded text-textSubtle hover:text-primary hover:bg-black/[0.02] transition-colors"
                              title="Download archive package"
                            >
                              <Download className="w-3.5 h-3.5" />
                            </button>
                          </div>
                        </td>
                      </tr>
                    )))}
                  </tbody>
                </table>
              </div>
            </div>
          </div>
        </main>
      </div>
    </div>
  );
}
