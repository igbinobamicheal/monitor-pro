"use client";

import { useState, useEffect } from "react";
import { useParams } from "next/navigation";
import Link from "next/link";
import { DashboardSidebar } from "@/components/Sidebar";
import { AmbientBackground } from "@/components/AmbientBackground";
import {
  Sliders,
  Plus,
  Trash2,
  CheckCircle2,
  Search,
  Cpu,
  HardDrive,
  Database,
  Activity,
  X,
  SlidersHorizontal,
  Bell,
  Sparkles,
} from "lucide-react";

export interface ThresholdRule {
  id: string;
  name: string;
  description: string;
  metricType: "CPU" | "RAM" | "Disk" | "Network" | "Heartbeat";
  operator: ">" | ">=" | "<" | "<=";
  thresholdValue: number;
  unit: "%" | "MB/s" | "sec";
  durationMinutes: number;
  severity: "critical" | "warning" | "info";
  enabled: boolean;
  channels: string[];
}

export default function RulesEnginePage() {
  const params = useParams();
  const subdomain = (params?.subdomain as string) || "demo";

  const defaultRules: ThresholdRule[] = [
    {
      id: "rule-01",
      name: "Sustained High CPU Saturation",
      description: "Detects prolonged high computational load that indicates thread exhaustion or runaway worker processes.",
      metricType: "CPU",
      operator: ">",
      thresholdValue: 90,
      unit: "%",
      durationMinutes: 5,
      severity: "critical",
      enabled: true,
      channels: ["In-App", "PagerDuty", "Slack"],
    },
    {
      id: "rule-02",
      name: "Critical Disk Space Exhaustion",
      description: "Alerts when primary OS or database system volumes drop below critical headroom threshold.",
      metricType: "Disk",
      operator: ">=",
      thresholdValue: 92,
      unit: "%",
      durationMinutes: 3,
      severity: "critical",
      enabled: true,
      channels: ["In-App", "Email"],
    },
    {
      id: "rule-03",
      name: "Elevated Memory Pressure Warning",
      description: "Triggers early warning when system RAM consumption threatens memory paging stalls.",
      metricType: "RAM",
      operator: ">",
      thresholdValue: 85,
      unit: "%",
      durationMinutes: 10,
      severity: "warning",
      enabled: true,
      channels: ["In-App"],
    },
    {
      id: "rule-04",
      name: "Agent Heartbeat Missed",
      description: "Fires when endpoint background daemon fails to broadcast telemetry ping within SLA window.",
      metricType: "Heartbeat",
      operator: ">",
      thresholdValue: 60,
      unit: "sec",
      durationMinutes: 2,
      severity: "warning",
      enabled: true,
      channels: ["In-App", "Slack"],
    },
  ];

  const [rules, setRules] = useState<ThresholdRule[]>([]);

  useEffect(() => {
    if (typeof window !== "undefined") {
      const stored = localStorage.getItem(`monitopro_rules_${subdomain}`);
      if (stored) {
        try {
          setRules(JSON.parse(stored));
          return;
        } catch {}
      }
      setRules(defaultRules);
      localStorage.setItem(`monitopro_rules_${subdomain}`, JSON.stringify(defaultRules));
    }
  }, [subdomain]);

  const saveRules = (updated: ThresholdRule[]) => {
    setRules(updated);
    if (typeof window !== "undefined") {
      localStorage.setItem(`monitopro_rules_${subdomain}`, JSON.stringify(updated));
    }
  };

  const [searchTerm, setSearchTerm] = useState("");
  const [filterSeverity, setFilterSeverity] = useState<"all" | "critical" | "warning" | "info">("all");
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [toastMessage, setToastMessage] = useState<string | null>(null);

  // New Rule Form State
  const [formName, setFormName] = useState("");
  const [formDesc, setFormDesc] = useState("");
  const [formMetric, setFormMetric] = useState<"CPU" | "RAM" | "Disk" | "Network" | "Heartbeat">("CPU");
  const [formOperator, setFormOperator] = useState<">" | ">=" | "<" | "<=">(">");
  const [formThreshold, setFormThreshold] = useState<number>(85);
  const [formDuration, setFormDuration] = useState<number>(5);
  const [formSeverity, setFormSeverity] = useState<"critical" | "warning" | "info">("critical");
  const [formInApp, setFormInApp] = useState(true);
  const [formSlack, setFormSlack] = useState(true);
  const [formEmail, setFormEmail] = useState(false);

  const showToast = (msg: string) => {
    setToastMessage(msg);
    setTimeout(() => setToastMessage(null), 3000);
  };

  const handleToggle = (id: string) => {
    const updated = rules.map((r) => {
      if (r.id === id) {
        const next = !r.enabled;
        showToast(`Rule "${r.name}" ${next ? "enabled" : "disabled"}`);
        return { ...r, enabled: next };
      }
      return r;
    });
    saveRules(updated);
  };

  const handleDelete = (id: string) => {
    const target = rules.find((r) => r.id === id);
    const updated = rules.filter((r) => r.id !== id);
    saveRules(updated);
    showToast(`Rule "${target?.name || id}" deleted`);
  };

  const handleCreateRule = (e: React.FormEvent) => {
    e.preventDefault();
    if (!formName.trim()) return;

    const unit = formMetric === "Heartbeat" ? "sec" : formMetric === "Network" ? "MB/s" : "%";
    const channels: string[] = [];
    if (formInApp) channels.push("In-App");
    if (formSlack) channels.push("Slack");
    if (formEmail) channels.push("Email");

    const newRule: ThresholdRule = {
      id: `rule-${Date.now().toString().slice(-4)}`,
      name: formName.trim(),
      description: formDesc.trim() || `Threshold rule for ${formMetric} ${formOperator} ${formThreshold}${unit}`,
      metricType: formMetric,
      operator: formOperator,
      thresholdValue: formThreshold,
      unit,
      durationMinutes: formDuration,
      severity: formSeverity,
      enabled: true,
      channels: channels.length > 0 ? channels : ["In-App"],
    };

    saveRules([newRule, ...rules]);
    setIsModalOpen(false);
    showToast(`Successfully created rule "${newRule.name}"`);

    // Reset form
    setFormName("");
    setFormDesc("");
    setFormMetric("CPU");
    setFormOperator(">");
    setFormThreshold(85);
    setFormDuration(5);
    setFormSeverity("critical");
  };

  const filteredRules = rules.filter((r) => {
    const matchesSearch =
      r.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
      r.description.toLowerCase().includes(searchTerm.toLowerCase()) ||
      r.metricType.toLowerCase().includes(searchTerm.toLowerCase());
    if (!matchesSearch) return false;
    if (filterSeverity === "all") return true;
    return r.severity === filterSeverity;
  });

  const getMetricIcon = (metric: ThresholdRule["metricType"]) => {
    switch (metric) {
      case "CPU":
        return <Cpu className="w-4 h-4 text-primary" />;
      case "RAM":
        return <Database className="w-4 h-4 text-amber-700" />;
      case "Disk":
        return <HardDrive className="w-4 h-4 text-emerald-700" />;
      case "Network":
        return <Activity className="w-4 h-4 text-blue-600" />;
      case "Heartbeat":
        return <Activity className="w-4 h-4 text-purple-600" />;
    }
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
            <span className="text-textMain font-semibold">Monitoring Rules Engine</span>
          </div>

          <div className="flex items-center gap-2">
            <button
              onClick={() => setIsModalOpen(true)}
              className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-primary hover:bg-primaryHover text-white text-xs font-semibold shadow-sm transition-colors"
            >
              <Plus className="w-3.5 h-3.5" />
              Create Rule
            </button>
          </div>
        </header>

        {/* Toast */}
        {toastMessage && (
          <div className="fixed bottom-6 right-6 z-50 bg-[#151310] text-white px-4 py-3 rounded-lg shadow-2xl border border-white/10 flex items-center gap-3 animate-fade-in text-xs">
            <CheckCircle2 className="w-4 h-4 text-emerald-400 shrink-0" />
            <span>{toastMessage}</span>
          </div>
        )}

        <main className="p-8 max-w-7xl w-full mx-auto space-y-6">
          {/* Headline & Overview */}
          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
            <div>
              <h1 className="text-xl font-bold tracking-tight text-textMain flex items-center gap-2">
                <Sliders className="w-5 h-5 text-primary" />
                Monitoring Threshold Rules Engine
              </h1>
              <p className="text-xs text-textMuted mt-1">
                Establish automated threshold criteria to detect node degradation before it impacts production
              </p>
            </div>
            <div className="flex items-center gap-2">
              <span className="text-xs font-mono text-textSubtle">
                Active Engine: <strong>Deterministic Event Bus</strong>
              </span>
            </div>
          </div>

          {/* Search & Filter */}
          <div className="border border-white/50 bg-surface backdrop-blur-md rounded-xl p-4 shadow-sm glass-panel flex flex-col md:flex-row items-center justify-between gap-3">
            <div className="relative w-full md:w-80">
              <Search className="w-4 h-4 text-textSubtle absolute left-3 top-2.5" />
              <input
                type="text"
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                placeholder="Search rule name or metric..."
                className="w-full bg-white border border-borderSubtle rounded-lg pl-9 pr-3 py-1.5 text-xs text-textMain focus:outline-none focus:border-primary transition-colors shadow-inner"
              />
            </div>

            <div className="flex items-center gap-1.5">
              {(
                [
                  { key: "all", label: "All Rules" },
                  { key: "critical", label: "Critical" },
                  { key: "warning", label: "Warning" },
                  { key: "info", label: "Info" },
                ] as const
              ).map((tab) => (
                <button
                  key={tab.key}
                  onClick={() => setFilterSeverity(tab.key)}
                  className={`px-3 py-1 rounded-md text-xs font-medium transition-colors ${
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

          {/* Rules List */}
          <div className="space-y-3">
            {filteredRules.map((rule) => {
              const isCrit = rule.severity === "critical";
              const isWarn = rule.severity === "warning";

              return (
                <div
                  key={rule.id}
                  className={`border rounded-xl p-5 shadow-sm transition-all ${
                    rule.enabled
                      ? "border-white/60 bg-surface backdrop-blur-md glass-panel"
                      : "border-black/[0.06] bg-stone-100/60 opacity-60"
                  }`}
                >
                  <div className="flex flex-col md:flex-row md:items-center justify-between gap-4">
                    <div className="flex items-start gap-3.5">
                      <div className="p-2.5 rounded-lg bg-white border border-borderSubtle shadow-sm shrink-0">
                        {getMetricIcon(rule.metricType)}
                      </div>

                      <div className="space-y-1">
                        <div className="flex items-center gap-2">
                          <h3 className="text-sm font-bold text-textMain">{rule.name}</h3>
                          <span
                            className={`px-2 py-0.5 rounded text-[10px] font-mono font-semibold uppercase border ${
                              isCrit
                                ? "bg-rose-50 border-rose-200 text-rose-700"
                                : isWarn
                                ? "bg-amber-50 border-amber-200 text-amber-800"
                                : "bg-slate-100 border-slate-200 text-slate-700"
                            }`}
                          >
                            {rule.severity}
                          </span>
                          {!rule.enabled && (
                            <span className="text-[10px] text-textSubtle uppercase tracking-wider font-mono">
                              (Disabled)
                            </span>
                          )}
                        </div>
                        <p className="text-xs text-textMuted leading-relaxed max-w-2xl">{rule.description}</p>
                      </div>
                    </div>

                    <div className="flex items-center gap-6 self-end md:self-center shrink-0">
                      {/* Metric Expression Box */}
                      <div className="px-3.5 py-1.5 rounded-lg bg-white/70 border border-borderSubtle font-mono text-xs">
                        <span className="text-textSubtle uppercase text-[10px] block">Threshold Logic</span>
                        <span className="font-semibold text-textMain">
                          {rule.metricType} {rule.operator} {rule.thresholdValue}
                          {rule.unit}
                        </span>
                        <span className="text-textSubtle text-[11px] ml-1.5 font-sans">
                          for {rule.durationMinutes}m
                        </span>
                      </div>

                      {/* Notification Channels */}
                      <div className="hidden lg:flex items-center gap-1">
                        {rule.channels.map((ch) => (
                          <span
                            key={ch}
                            className="px-2 py-0.5 rounded bg-black/[0.03] border border-borderSubtle text-[10px] font-medium text-textMuted"
                          >
                            {ch}
                          </span>
                        ))}
                      </div>

                      {/* Toggle & Actions */}
                      <div className="flex items-center gap-3">
                        {/* Custom switch */}
                        <button
                          type="button"
                          onClick={() => handleToggle(rule.id)}
                          className={`w-11 h-6 flex items-center rounded-full p-1 cursor-pointer transition-colors ${
                            rule.enabled ? "bg-primary" : "bg-black/20"
                          }`}
                          title={rule.enabled ? "Disable rule" : "Enable rule"}
                        >
                          <div
                            className={`bg-white w-4 h-4 rounded-full shadow-md transform transition-transform ${
                              rule.enabled ? "translate-x-5" : "translate-x-0"
                            }`}
                          />
                        </button>

                        <button
                          onClick={() => handleDelete(rule.id)}
                          className="p-1.5 rounded-lg text-textSubtle hover:text-rose-600 hover:bg-rose-50 transition-colors"
                          title="Delete rule"
                        >
                          <Trash2 className="w-4 h-4" />
                        </button>
                      </div>
                    </div>
                  </div>
                </div>
              );
            })}
          </div>
        </main>
      </div>

      {/* Create Rule Modal */}
      {isModalOpen && (
        <div className="fixed inset-0 z-50 bg-black/40 backdrop-blur-sm flex items-center justify-center p-4 overflow-y-auto">
          <div className="bg-white rounded-xl shadow-2xl border border-borderSubtle max-w-xl w-full p-6 space-y-5 my-8">
            <div className="flex items-center justify-between border-b border-borderSubtle pb-3">
              <div className="flex items-center gap-2 text-textMain font-bold text-base">
                <SlidersHorizontal className="w-5 h-5 text-primary" />
                Configure New Threshold Rule
              </div>
              <button
                onClick={() => setIsModalOpen(false)}
                className="text-textSubtle hover:text-textMain p-1 rounded-md"
              >
                <X className="w-4 h-4" />
              </button>
            </div>

            <form onSubmit={handleCreateRule} className="space-y-4">
              <div className="space-y-1.5">
                <label className="text-xs font-semibold text-textMain">Rule Title</label>
                <input
                  type="text"
                  required
                  value={formName}
                  onChange={(e) => setFormName(e.target.value)}
                  placeholder="e.g. Host Memory Pressure Warning"
                  className="w-full bg-white border border-borderSubtle rounded-lg px-3 py-2 text-xs text-textMain focus:outline-none focus:border-primary shadow-sm"
                />
              </div>

              <div className="space-y-1.5">
                <label className="text-xs font-semibold text-textMain">Description</label>
                <input
                  type="text"
                  value={formDesc}
                  onChange={(e) => setFormDesc(e.target.value)}
                  placeholder="Explain why this threshold is significant..."
                  className="w-full bg-white border border-borderSubtle rounded-lg px-3 py-2 text-xs text-textMain focus:outline-none focus:border-primary shadow-sm"
                />
              </div>

              {/* Metric & Operator Grid */}
              <div className="grid grid-cols-2 gap-3">
                <div className="space-y-1.5">
                  <label className="text-xs font-semibold text-textMain">Target Metric</label>
                  <select
                    value={formMetric}
                    onChange={(e) => setFormMetric(e.target.value as any)}
                    className="w-full bg-white border border-borderSubtle rounded-lg px-3 py-2 text-xs text-textMain focus:outline-none focus:border-primary shadow-sm"
                  >
                    <option value="CPU">CPU Usage (%)</option>
                    <option value="RAM">RAM Allocation (%)</option>
                    <option value="Disk">Disk Capacity (%)</option>
                    <option value="Network">Network Bandwidth (MB/s)</option>
                    <option value="Heartbeat">Heartbeat Delay (sec)</option>
                  </select>
                </div>

                <div className="space-y-1.5">
                  <label className="text-xs font-semibold text-textMain">Evaluation Operator</label>
                  <select
                    value={formOperator}
                    onChange={(e) => setFormOperator(e.target.value as any)}
                    className="w-full bg-white border border-borderSubtle rounded-lg px-3 py-2 text-xs text-textMain focus:outline-none focus:border-primary shadow-sm"
                  >
                    <option value=">">Greater than (&gt;)</option>
                    <option value=">=">Greater than or equal (&gt;=)</option>
                    <option value="<">Less than (&lt;)</option>
                    <option value="<=">Less than or equal (&lt;=)</option>
                  </select>
                </div>
              </div>

              {/* Threshold Slider */}
              <div className="space-y-2 p-3 bg-stone-50 rounded-lg border border-borderSubtle">
                <div className="flex justify-between items-center text-xs">
                  <span className="font-semibold text-textMain">Threshold Level</span>
                  <span className="font-mono text-primary font-bold text-sm">
                    {formThreshold} {formMetric === "Heartbeat" ? "sec" : formMetric === "Network" ? "MB/s" : "%"}
                  </span>
                </div>
                <input
                  type="range"
                  min={formMetric === "Heartbeat" ? 10 : 5}
                  max={formMetric === "Heartbeat" ? 300 : 100}
                  step={formMetric === "Heartbeat" ? 5 : 1}
                  value={formThreshold}
                  onChange={(e) => setFormThreshold(Number(e.target.value))}
                  className="w-full accent-[#B87C3B] cursor-pointer"
                />
                <div className="flex justify-between text-[10px] text-textSubtle font-mono">
                  <span>{formMetric === "Heartbeat" ? "10s" : "5%"}</span>
                  <span>{formMetric === "Heartbeat" ? "300s" : "100%"}</span>
                </div>
              </div>

              {/* Duration and Severity */}
              <div className="grid grid-cols-2 gap-3">
                <div className="space-y-1.5">
                  <label className="text-xs font-semibold text-textMain">Sustained Window</label>
                  <select
                    value={formDuration}
                    onChange={(e) => setFormDuration(Number(e.target.value))}
                    className="w-full bg-white border border-borderSubtle rounded-lg px-3 py-2 text-xs text-textMain focus:outline-none focus:border-primary shadow-sm"
                  >
                    <option value={1}>1 Minute</option>
                    <option value={3}>3 Minutes</option>
                    <option value={5}>5 Minutes</option>
                    <option value={10}>10 Minutes</option>
                    <option value={15}>15 Minutes</option>
                  </select>
                </div>

                <div className="space-y-1.5">
                  <label className="text-xs font-semibold text-textMain">Severity Badge</label>
                  <select
                    value={formSeverity}
                    onChange={(e) => setFormSeverity(e.target.value as any)}
                    className="w-full bg-white border border-borderSubtle rounded-lg px-3 py-2 text-xs text-textMain focus:outline-none focus:border-primary shadow-sm"
                  >
                    <option value="critical">Critical (Red)</option>
                    <option value="warning">Warning (Amber)</option>
                    <option value="info">Info (Slate)</option>
                  </select>
                </div>
              </div>

              {/* Notification Channels */}
              <div className="space-y-1.5">
                <label className="text-xs font-semibold text-textMain">Notification Channels</label>
                <div className="flex items-center gap-4 text-xs text-textMuted">
                  <label className="flex items-center gap-1.5 cursor-pointer">
                    <input
                      type="checkbox"
                      checked={formInApp}
                      onChange={(e) => setFormInApp(e.target.checked)}
                      className="accent-[#B87C3B]"
                    />
                    <span>In-App Dashboard</span>
                  </label>
                  <label className="flex items-center gap-1.5 cursor-pointer">
                    <input
                      type="checkbox"
                      checked={formSlack}
                      onChange={(e) => setFormSlack(e.target.checked)}
                      className="accent-[#B87C3B]"
                    />
                    <span>Slack Webhook</span>
                  </label>
                  <label className="flex items-center gap-1.5 cursor-pointer">
                    <input
                      type="checkbox"
                      checked={formEmail}
                      onChange={(e) => setFormEmail(e.target.checked)}
                      className="accent-[#B87C3B]"
                    />
                    <span>Email Broadcast</span>
                  </label>
                </div>
              </div>

              {/* Live Preview Box */}
              <div className="p-3.5 bg-amber-50/60 border border-amber-200/60 rounded-lg flex items-start gap-2.5">
                <Sparkles className="w-4 h-4 text-primary shrink-0 mt-0.5" />
                <div className="text-xs space-y-0.5">
                  <div className="font-semibold text-textMain">Evaluation Preview:</div>
                  <div className="font-mono text-textMuted">
                    Trigger <strong className="text-primary uppercase">{formSeverity}</strong> alert if{" "}
                    <strong>{formMetric}</strong> {formOperator}{" "}
                    <strong>
                      {formThreshold}
                      {formMetric === "Heartbeat" ? "s" : "%"}
                    </strong>{" "}
                    for <strong>{formDuration} minutes</strong>.
                  </div>
                </div>
              </div>

              {/* Buttons */}
              <div className="flex justify-end gap-2 pt-2 border-t border-borderSubtle">
                <button
                  type="button"
                  onClick={() => setIsModalOpen(false)}
                  className="px-4 py-2 rounded-lg border border-borderSubtle text-xs font-medium text-textMuted hover:bg-stone-50 transition-colors"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="px-4 py-2 rounded-lg bg-primary hover:bg-primaryHover text-white text-xs font-semibold shadow-sm transition-colors"
                >
                  Save & Activate Rule
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
}
