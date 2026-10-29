"use client";

import { useParams } from "next/navigation";
import { DashboardSidebar } from "@/components/Sidebar";
import { AmbientBackground } from "@/components/AmbientBackground";
import { Shield } from "lucide-react";
import Link from "next/link";

export default function AuditLogsPage() {
  const params = useParams();
  const subdomain = (params?.subdomain as string) || "demo";

  const auditEvents = [
    {
      id: "aud-001",
      timestamp: "2026-10-27 16:42:10 UTC",
      actor: "Lead IT Admin",
      action: "REMOTE_SESSION_APPROVED",
      resource: "srv-prod-sql-01",
      details: "User explicitly approved remote screen sharing and control after reading mandatory consent text",
    },
    {
      id: "aud-002",
      timestamp: "2026-10-27 16:38:05 UTC",
      actor: "Lead IT Admin",
      action: "REMEDIATION_APPROVED",
      resource: "clean_temp_files",
      details: "Dual-operator approval granted for low-risk temporary file reclamation",
    },
    {
      id: "aud-003",
      timestamp: "2026-10-27 16:30:00 UTC",
      actor: "System Sentinel",
      action: "ENDPOINT_ENROLLED",
      resource: "ws-eng-micheal",
      details: "Cryptographic Ed25519 identity verification and certificate issued",
    },
  ];

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
            <span className="text-textMain font-semibold">Audit Logs</span>
          </div>

          <div className="flex items-center gap-1.5 px-2 py-1 rounded bg-emerald-50 border border-emerald-200/50 text-emerald-600 text-[10px] font-mono uppercase tracking-wide">
            Immutable Audit Trail
          </div>
        </header>

        <main className="flex-1 p-8 max-w-[1200px] mx-auto w-full flex flex-col gap-6">
          <div>
            <div className="flex items-center gap-2">
              <Shield className="w-5 h-5 text-primary" />
              <h1 className="text-2xl font-bold tracking-tight text-textMain">Security Audit Logs</h1>
            </div>
            <p className="text-xs text-textMuted mt-1">
              Immutable, tenant-scoped audit trail of all security-critical operations
            </p>
          </div>

          <div className="bg-surface border border-borderSubtle rounded-lg overflow-hidden flex flex-col glass-panel shadow-sm">
            <div className="overflow-x-auto w-full">
              <table className="w-full text-left border-collapse text-[12px] font-mono">
                <thead>
                  <tr className="border-b border-borderSubtle bg-black/[0.005]">
                    <th className="py-2.5 px-4 font-medium text-textSubtle uppercase tracking-widest text-[10px]">Timestamp (UTC)</th>
                    <th className="py-2.5 px-4 font-medium text-textSubtle uppercase tracking-widest text-[10px]">Actor</th>
                    <th className="py-2.5 px-4 font-medium text-textSubtle uppercase tracking-widest text-[10px]">Action Type</th>
                    <th className="py-2.5 px-4 font-medium text-textSubtle uppercase tracking-widest text-[10px]">Target Resource</th>
                    <th className="py-2.5 px-4 font-medium text-textSubtle uppercase tracking-widest text-[10px]">Audit Details</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-borderSubtle text-textMain">
                  {auditEvents.map((ev) => (
                    <tr key={ev.id} className="hover:bg-black/[0.01] transition-colors">
                      <td className="py-3 px-4 text-textSubtle whitespace-nowrap">{ev.timestamp}</td>
                      <td className="py-3 px-4 font-semibold text-textMain">{ev.actor}</td>
                      <td className="py-3 px-4">
                        <span className="px-2 py-0.5 rounded border border-primary/30 bg-primary/10 text-primary font-bold text-[10px]">
                          {ev.action}
                        </span>
                      </td>
                      <td className="py-3 px-4 text-textMain font-semibold">{ev.resource}</td>
                      <td className="py-3 px-4 text-textMuted max-w-md truncate">{ev.details}</td>
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
