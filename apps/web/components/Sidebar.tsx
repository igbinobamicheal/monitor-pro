"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import {
  LayoutDashboard,
  Server,
  DownloadCloud,
  AlertTriangle,
  Sliders,
  Flame,
  MonitorPlay,
  Wrench,
  Shield,
  Users,
  Settings,
  LogOut,
} from "lucide-react";

interface SidebarProps {
  subdomain: string;
}

export function DashboardSidebar({ subdomain }: SidebarProps) {
  const pathname = usePathname();

  const navSections = [
    {
      title: "Platform",
      items: [
        { label: "Overview", href: `/${subdomain}/overview`, icon: LayoutDashboard },
        { label: "Infrastructure", href: `/${subdomain}/devices`, icon: Server },
        { label: "Deploy Agent", href: `/${subdomain}/deploy`, icon: DownloadCloud },
        { label: "Alerts", href: `/${subdomain}/alerts`, icon: AlertTriangle },
        { label: "Alert Rules", href: `/${subdomain}/rules`, icon: Sliders },
      ],
    },
    {
      title: "Investigate & Action",
      items: [
        { label: "Incidents", href: `/${subdomain}/incidents`, icon: Flame },
        { label: "Remote Support", href: `/${subdomain}/remote-support`, icon: MonitorPlay },
        { label: "Remediation", href: `/${subdomain}/remediation`, icon: Wrench },
        { label: "Audit Logs", href: `/${subdomain}/audit-logs`, icon: Shield },
      ],
    },
    {
      title: "Administration",
      items: [
        { label: "Team", href: `/${subdomain}/team`, icon: Users },
        { label: "Settings", href: `/${subdomain}/settings`, icon: Settings },
      ],
    },
  ];

  return (
    <aside className="bg-[#151310] h-screen w-[220px] flex-shrink-0 flex flex-col border-r border-white/5 fixed left-0 top-0 z-40 select-none text-zinc-400">
      {/* Brand Header */}
      <div className="px-5 py-4 flex items-center justify-between border-b border-white/5">
        <Link href={`/${subdomain}/overview`} className="flex items-center gap-2.5">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
            <path d="M12 2L2 7L12 12L22 7L12 2Z" stroke="#B87C3B" strokeWidth="2" strokeLinejoin="round" />
            <path d="M2 17L12 22L22 17" stroke="#B87C3B" strokeWidth="2" strokeLinejoin="round" />
            <path d="M2 12L12 17L22 12" stroke="#B87C3B" strokeWidth="2" strokeLinejoin="round" />
          </svg>
          <span className="text-[14px] font-semibold tracking-tight text-white">MonitorPro</span>
        </Link>
      </div>

      {/* Navigation Sections */}
      <div className="flex-1 overflow-y-auto px-3 py-4 space-y-6">
        {navSections.map((section) => (
          <div key={section.title}>
            <div className="text-[10px] font-semibold text-zinc-500 px-3 mb-2 uppercase tracking-widest">
              {section.title}
            </div>
            <ul className="flex flex-col gap-0.5">
              {section.items.map((item) => {
                const Icon = item.icon;
                const active = pathname === item.href || (item.href.endsWith("/devices") && pathname?.startsWith(`/${subdomain}/devices/`));
                return (
                  <li key={item.label}>
                    <Link
                      href={item.href}
                      className={`flex items-center gap-3 pl-3 pr-4 py-2 border-l-2 text-[13px] transition-all rounded-r-md ${
                        active
                          ? "border-[#B87C3B] text-[#B87C3B] font-semibold bg-white/[0.02]"
                          : "border-transparent text-zinc-400 hover:text-white hover:bg-white/[0.01] font-medium"
                      }`}
                    >
                      <Icon className={`w-4 h-4 ${active ? "text-[#B87C3B]" : "text-zinc-500"}`} />
                      <span>{item.label}</span>
                    </Link>
                  </li>
                );
              })}
            </ul>
          </div>
        ))}
      </div>

      {/* User Footer matching V1 */}
      <div className="p-3 border-t border-white/5">
        <div className="mb-3 px-3 py-2 bg-white/[0.02] border border-white/5 rounded-md">
          <div className="text-[12px] font-medium text-white truncate capitalize">{subdomain} Workspace</div>
          <div className="text-[10px] text-zinc-500 font-mono uppercase mt-0.5">Administrator Console</div>
        </div>
        <Link
          href="/login"
          className="flex items-center gap-3 w-full p-2 rounded-md hover:bg-white/[0.02] text-zinc-400 hover:text-white text-[13px] font-medium transition-all text-left border border-transparent hover:border-white/5"
        >
          <LogOut className="w-4 h-4 text-zinc-500" />
          <span>Sign Out</span>
        </Link>
      </div>
    </aside>
  );
}
