"use client";

import { useEffect, useState } from "react";
import { useParams } from "next/navigation";
import { DashboardSidebar } from "@/components/Sidebar";
import { AmbientBackground } from "@/components/AmbientBackground";
import {
  Flame,
  Sparkles,
  CheckCircle2,
  Clock,
  ShieldCheck,
  ArrowRight,
} from "lucide-react";
import { api, Incident, AIAdvisory } from "@/lib/api";
import Link from "next/link";

export default function IncidentsPage() {
  const params = useParams();
  const subdomain = (params?.subdomain as string) || "demo";

  const [incidents, setIncidents] = useState<Incident[]>([]);
  const [selectedIncident, setSelectedIncident] = useState<Incident | null>(null);
  const [aiReport, setAiReport] = useState<AIAdvisory | null>(null);
  const [aiLoading, setAiLoading] = useState(false);
  const [aiError, setAiError] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchIncidents = async () => {
      setLoading(true);
      try {
        const data = await api.listIncidents();
        setIncidents(data || []);
        if (data && data.length > 0) {
          setSelectedIncident(data[0]);
        } else {
          setSelectedIncident(null);
        }
      } catch {
        setIncidents([]);
        setSelectedIncident(null);
      } finally {
        setLoading(false);
      }
    };
    fetchIncidents();
  }, [subdomain]);

  const requestAIAnalysis = async (incident: Incident) => {
    setAiLoading(true);
    setAiError(null);
    try {
      const report = await api.getAIAdvisory(incident.id);
      setAiReport(report);
    } catch (err: any) {
      setAiError(err?.message || "Failed to contact AI Advisory service. Verify backend AI provider.");
    } finally {
      setAiLoading(false);
    }
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
            <span className="text-textMain font-semibold">Incidents</span>
          </div>

          <div className="flex items-center gap-1.5 px-2 py-1 rounded bg-amber-50 border border-amber-200/50 text-amber-700 text-[10px] font-mono uppercase tracking-wide">
            Automated Triage Active
          </div>
        </header>

        <main className="flex-1 p-8 max-w-[1200px] mx-auto w-full flex flex-col gap-6">
          <div>
            <div className="flex items-center gap-2">
              <Flame className="w-5 h-5 text-warning" />
              <h1 className="text-2xl font-bold tracking-tight text-textMain">Incidents & Root Cause Triage</h1>
            </div>
            <p className="text-xs text-textMuted mt-1">
              Correlated fleet anomalies with evidence-based AI hypothesis ranking
            </p>
          </div>

          {loading ? (
            <div className="py-20 text-center text-xs font-mono text-textMuted flex flex-col items-center gap-3">
              <div className="w-6 h-6 border-2 border-primary border-t-transparent rounded-full animate-spin" />
              <span>Querying correlated incident graph...</span>
            </div>
          ) : incidents.length === 0 ? (
            <div className="rounded-xl border border-borderSubtle bg-surface p-12 text-center glass-panel shadow-sm max-w-2xl mx-auto space-y-4">
              <div className="w-14 h-14 rounded-2xl bg-emerald-50 border border-emerald-200/60 flex items-center justify-center text-emerald-600 mx-auto shadow-sm">
                <ShieldCheck className="w-7 h-7" />
              </div>
              <div className="space-y-1">
                <h2 className="text-base font-bold text-textMain">All Systems Healthy</h2>
                <p className="text-xs text-textMuted leading-relaxed max-w-md mx-auto">
                  Zero active correlated incidents or anomalous thresholds detected across your fleet. Telemetry is being actively monitored against your defined policy baselines.
                </p>
              </div>
              <div className="flex items-center justify-center gap-3 pt-2">
                <Link
                  href={`/${subdomain}/alerts`}
                  className="border border-borderSubtle bg-surface hover:bg-black/[0.02] text-textMain px-3.5 py-1.5 rounded-md text-xs font-medium transition-all"
                >
                  View Alert History
                </Link>
                <Link
                  href={`/${subdomain}/rules`}
                  className="bg-primary hover:bg-primaryHover text-white px-3.5 py-1.5 rounded-md text-xs font-medium transition-all shadow-sm flex items-center gap-1.5"
                >
                  <span>Configure Alert Rules</span>
                  <ArrowRight className="w-3.5 h-3.5" />
                </Link>
              </div>
            </div>
          ) : (
            <div className="grid grid-cols-1 lg:grid-cols-12 gap-6">
              {/* Incidents List Column */}
              <div className="lg:col-span-5 space-y-3">
                <h2 className="text-[12px] font-mono uppercase tracking-wider text-textMuted font-semibold">
                  Active Correlated Incidents ({incidents.length})
                </h2>

                <div className="space-y-2">
                  {incidents.map((inc) => {
                    const isSelected = selectedIncident?.id === inc.id;
                    return (
                      <div
                        key={inc.id}
                        onClick={() => {
                          setSelectedIncident(inc);
                          setAiReport(null);
                          setAiError(null);
                        }}
                        className={`p-4 rounded-xl border cursor-pointer transition-all ${
                          isSelected
                            ? "border-primary bg-primary/10 shadow-md"
                            : "border-borderSubtle bg-surface hover:border-primary/40 glass-panel"
                        }`}
                      >
                        <div className="flex items-center justify-between text-[10px] font-mono mb-1.5">
                          <span
                            className={`px-2 py-0.5 rounded uppercase font-bold border ${
                              inc.severity === "critical"
                                ? "border-rose-200/50 bg-rose-50 text-rose-700"
                                : "border-amber-200/50 bg-amber-50 text-amber-700"
                            }`}
                          >
                            {inc.severity}
                          </span>
                          <span className="text-textSubtle uppercase">{inc.status}</span>
                        </div>

                        <h3 className="font-semibold text-sm text-textMain">{inc.title}</h3>

                        <div className="flex items-center justify-between text-[11px] font-mono text-textSubtle mt-3 pt-2 border-t border-borderSubtle">
                          <span>{inc.alerts_count} aggregated alerts</span>
                          <span>Active</span>
                        </div>
                      </div>
                    );
                  })}
                </div>
              </div>

              {/* Incident Detail & AI Advisory Column */}
              <div className="lg:col-span-7 space-y-6">
                {selectedIncident ? (
                  <div className="rounded-xl border border-white/50 bg-surface p-6 space-y-6 glass-panel shadow-sm">
                    <div className="flex items-start justify-between pb-4 border-b border-borderSubtle">
                      <div className="space-y-1">
                        <span className="text-xs font-mono text-primary font-semibold">{selectedIncident.id}</span>
                        <h2 className="text-lg font-bold tracking-tight text-textMain">{selectedIncident.title}</h2>
                      </div>

                      <button
                        onClick={() => requestAIAnalysis(selectedIncident)}
                        disabled={aiLoading}
                        className="flex items-center gap-1.5 px-3.5 py-2 rounded-lg bg-primary hover:bg-primaryHover text-white font-semibold text-xs transition-colors shadow-sm disabled:opacity-50"
                      >
                        <Sparkles className="w-4 h-4 fill-white" />
                        <span>{aiLoading ? "Analyzing..." : "Generate AI Advisory"}</span>
                      </button>
                    </div>

                    {aiError && (
                      <div className="p-3.5 rounded-lg bg-rose-50 border border-rose-200/70 text-xs text-rose-700">
                        {aiError}
                      </div>
                    )}

                    {/* AI Advisory Panel */}
                    {aiReport && (
                      <div className="p-5 rounded-xl border border-primary/30 bg-primary/5 space-y-4">
                        <div className="flex items-center justify-between pb-2 border-b border-primary/20">
                          <div className="flex items-center gap-2 text-xs font-mono font-bold text-primary">
                            <Sparkles className="w-4 h-4" />
                            <span>AI Advisory Hypothesis (Read-Only)</span>
                          </div>
                          <span className="text-[10px] font-mono uppercase bg-primary/10 border border-primary/30 px-2 py-0.5 rounded text-primary font-bold">
                            Human Authorization Required
                          </span>
                        </div>

                        <p className="text-xs text-textMain font-mono leading-relaxed">
                          {aiReport.summary}
                        </p>

                        {aiReport.hypotheses.map((h) => (
                          <div key={h.rank} className="p-3.5 rounded-lg bg-white/70 border border-borderSubtle space-y-2 text-xs font-mono">
                            <div className="flex items-center justify-between">
                              <span className="font-bold text-textMain">Hypothesis #{h.rank}: {h.cause}</span>
                              <span className="text-primary font-bold">{Math.round(h.confidence * 100)}% Confidence</span>
                            </div>

                            <div className="space-y-1 pt-1">
                              <div className="text-[10px] uppercase font-bold text-success">Supporting Evidence:</div>
                              <div className="text-textMuted pl-2 border-l-2 border-success/50 space-y-0.5">
                                {h.supporting_evidence.map((ev, i) => (
                                  <div key={i}>• {ev}</div>
                                ))}
                              </div>
                            </div>

                            <div className="space-y-1 pt-1">
                              <div className="text-[10px] uppercase font-bold text-primary">Contradicting Evidence:</div>
                              <div className="text-textMuted pl-2 border-l-2 border-primary/50 space-y-0.5">
                                {h.contradicting_evidence.map((ev, i) => (
                                  <div key={i}>• {ev}</div>
                                ))}
                              </div>
                            </div>
                          </div>
                        ))}

                        <div className="space-y-2 pt-2">
                          <div className="text-xs font-semibold text-textMain">Recommended Verification Steps:</div>
                          <div className="space-y-1 text-xs font-mono text-textMuted">
                            {aiReport.recommended_verification_steps.map((step, i) => (
                              <div key={i} className="flex items-center gap-2">
                                <CheckCircle2 className="w-3.5 h-3.5 text-success" />
                                <span>{step}</span>
                              </div>
                            ))}
                          </div>
                        </div>
                      </div>
                    )}

                    {/* Timeline Notes */}
                    <div className="space-y-3 pt-2">
                      <h3 className="font-semibold text-xs uppercase font-mono text-textMuted">
                        Investigation Timeline
                      </h3>
                      <div className="space-y-2 font-mono text-xs">
                        <div className="p-3 rounded-lg bg-white/60 border border-borderSubtle flex items-start gap-3">
                          <Clock className="w-4 h-4 text-textSubtle flex-shrink-0 mt-0.5" />
                          <div>
                            <div className="text-textMain font-semibold">Incident Registered in Fleet Graph</div>
                            <div className="text-textSubtle text-[11px] mt-0.5">{new Date(selectedIncident.created_at).toLocaleString()}</div>
                          </div>
                        </div>
                      </div>
                    </div>
                  </div>
                ) : (
                  <div className="p-8 rounded-xl border border-white/50 bg-surface text-center text-textMuted text-xs font-mono glass-panel">
                    Select an incident to view details and advisory analysis.
                  </div>
                )}
              </div>
            </div>
          )}
        </main>
      </div>
    </div>
  );
}
