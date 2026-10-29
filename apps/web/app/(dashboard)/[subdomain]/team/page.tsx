"use client";

import { useState, useEffect } from "react";
import { useParams } from "next/navigation";
import Link from "next/link";
import { DashboardSidebar } from "@/components/Sidebar";
import { AmbientBackground } from "@/components/AmbientBackground";
import {
  Users,
  UserPlus,
  ShieldCheck,
  ShieldAlert,
  Copy,
  Check,
  Search,
  CheckCircle2,
  X,
  Mail,
  UserCheck,
  Trash2,
  Shield,
  KeyRound,
  Sparkles,
} from "lucide-react";

export type RoleType = "owner_admin" | "it_admin" | "it_operator" | "readonly_viewer";

export interface TeamMember {
  id: string;
  fullName: string;
  email: string;
  role: RoleType;
  mfaEnrolled: boolean;
  joinedAt: string;
  lastActive: string;
  assignedEndpointsCount: number;
}

export default function TeamPage() {
  const params = useParams();
  const subdomain = (params?.subdomain as string) || "demo";

  const [members, setMembers] = useState<TeamMember[]>([]);

  useEffect(() => {
    const adminName = typeof window !== 'undefined' ? localStorage.getItem("monitopro_admin_name") || "Workspace Administrator" : "Workspace Administrator";
    const adminEmail = typeof window !== 'undefined' ? localStorage.getItem("monitopro_admin_email") || `admin@${subdomain}.internal` : `admin@${subdomain}.internal`;
    const stored = typeof window !== 'undefined' ? localStorage.getItem(`monitopro_team_${subdomain}`) : null;

    if (stored) {
      try {
        setMembers(JSON.parse(stored));
        return;
      } catch {}
    }

    const initial: TeamMember[] = [
      {
        id: "owner-01",
        fullName: adminName,
        email: adminEmail,
        role: "owner_admin",
        mfaEnrolled: true,
        joinedAt: "Initial Provisioning",
        lastActive: "Active Session",
        assignedEndpointsCount: 0,
      },
    ];
    setMembers(initial);
    if (typeof window !== 'undefined') {
      localStorage.setItem(`monitopro_team_${subdomain}`, JSON.stringify(initial));
    }
  }, [subdomain]);

  const saveMembers = (updated: TeamMember[]) => {
    setMembers(updated);
    if (typeof window !== 'undefined') {
      localStorage.setItem(`monitopro_team_${subdomain}`, JSON.stringify(updated));
    }
  };

  const [searchTerm, setSearchTerm] = useState("");
  const [filterRole, setFilterRole] = useState<"all" | RoleType>("all");
  const [isInviteModalOpen, setIsInviteModalOpen] = useState(false);
  const [toastMessage, setToastMessage] = useState<string | null>(null);

  // Invite Form
  const [inviteName, setInviteName] = useState("");
  const [inviteEmail, setInviteEmail] = useState("");
  const [inviteRole, setInviteRole] = useState<RoleType>("it_operator");
  const [generatedInviteLink, setGeneratedInviteLink] = useState<string | null>(null);
  const [copiedLink, setCopiedLink] = useState(false);

  const showToast = (msg: string) => {
    setToastMessage(msg);
    setTimeout(() => setToastMessage(null), 3000);
  };

  const handleRoleChange = (userId: string, newRole: RoleType) => {
    const updated = members.map((m) => (m.id === userId ? { ...m, role: newRole } : m));
    saveMembers(updated);
    showToast(`Updated role to ${newRole}`);
  };

  const handleRemoveMember = (userId: string) => {
    const target = members.find((m) => m.id === userId);
    const updated = members.filter((m) => m.id !== userId);
    saveMembers(updated);
    showToast(`Removed ${target?.fullName || "member"} from organization roster`);
  };

  const handleGenerateInvite = (e: React.FormEvent) => {
    e.preventDefault();
    if (!inviteEmail.trim() || !inviteName.trim()) return;

    const token = "inv_" + Math.random().toString(36).substring(2, 12);
    const link = `https://${subdomain}.monitopro.com/join?token=${token}&role=${inviteRole}`;
    setGeneratedInviteLink(link);

    const newMember: TeamMember = {
      id: `usr-${Date.now().toString().slice(-4)}`,
      fullName: inviteName.trim(),
      email: inviteEmail.trim(),
      role: inviteRole,
      mfaEnrolled: false,
      joinedAt: "Pending Invitation",
      lastActive: "Never",
      assignedEndpointsCount: 0,
    };
    saveMembers([newMember, ...members]);
    showToast(`Invitation created for ${inviteEmail}`);
  };

  const copyInviteLink = () => {
    if (!generatedInviteLink) return;
    navigator.clipboard.writeText(generatedInviteLink);
    setCopiedLink(true);
    setTimeout(() => setCopiedLink(false), 2000);
  };

  const filteredMembers = members.filter((m) => {
    const matchesSearch =
      m.fullName.toLowerCase().includes(searchTerm.toLowerCase()) ||
      m.email.toLowerCase().includes(searchTerm.toLowerCase());
    if (!matchesSearch) return false;
    if (filterRole === "all") return true;
    return m.role === filterRole;
  });

  const getRoleBadge = (role: RoleType) => {
    switch (role) {
      case "owner_admin":
        return (
          <span className="px-2.5 py-0.5 rounded-full text-[10px] font-mono font-bold uppercase tracking-wider bg-purple-50 text-purple-700 border border-purple-200">
            Owner Admin
          </span>
        );
      case "it_admin":
        return (
          <span className="px-2.5 py-0.5 rounded-full text-[10px] font-mono font-bold uppercase tracking-wider bg-amber-50 text-amber-800 border border-amber-200">
            IT Admin
          </span>
        );
      case "it_operator":
        return (
          <span className="px-2.5 py-0.5 rounded-full text-[10px] font-mono font-bold uppercase tracking-wider bg-blue-50 text-blue-700 border border-blue-200">
            IT Operator
          </span>
        );
      case "readonly_viewer":
        return (
          <span className="px-2.5 py-0.5 rounded-full text-[10px] font-mono font-bold uppercase tracking-wider bg-stone-100 text-stone-700 border border-stone-200">
            Read-Only Viewer
          </span>
        );
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
            <span className="text-textMain font-semibold">Team Members & RBAC</span>
          </div>

          <div className="flex items-center gap-2">
            <button
              onClick={() => {
                setGeneratedInviteLink(null);
                setInviteName("");
                setInviteEmail("");
                setIsInviteModalOpen(true);
              }}
              className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg bg-primary hover:bg-primaryHover text-white text-xs font-semibold shadow-sm transition-colors"
            >
              <UserPlus className="w-3.5 h-3.5" />
              Invite Member
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
          {/* Headline */}
          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
            <div>
              <h1 className="text-xl font-bold tracking-tight text-textMain flex items-center gap-2">
                <Users className="w-5 h-5 text-primary" />
                Team Members & RBAC Security
              </h1>
              <p className="text-xs text-textMuted mt-1">
                Manage organization operator accounts, granular permission boundaries, and MFA enforcement
              </p>
            </div>
            <div className="flex items-center gap-2">
              <span className="px-2.5 py-1 rounded bg-emerald-50 border border-emerald-200/50 text-emerald-700 text-[10px] font-mono font-semibold uppercase">
                MFA Enforced Workspace
              </span>
            </div>
          </div>

          {/* RBAC Info Card */}
          <div className="border border-white/50 bg-surface backdrop-blur-md rounded-xl p-5 shadow-sm glass-panel grid grid-cols-1 md:grid-cols-4 gap-4 text-xs">
            <div className="space-y-1">
              <div className="font-semibold text-textMain flex items-center gap-1.5">
                <Shield className="w-3.5 h-3.5 text-purple-600" /> Owner Admin
              </div>
              <p className="text-textSubtle leading-relaxed">
                Full authority: billing, tenant lifecycle, AI kill-switch override, user deletion.
              </p>
            </div>
            <div className="space-y-1">
              <div className="font-semibold text-textMain flex items-center gap-1.5">
                <ShieldCheck className="w-3.5 h-3.5 text-amber-700" /> IT Admin
              </div>
              <p className="text-textSubtle leading-relaxed">
                Can approve destructive remediations, configure alert rules, initiate remote sessions.
              </p>
            </div>
            <div className="space-y-1">
              <div className="font-semibold text-textMain flex items-center gap-1.5">
                <UserCheck className="w-3.5 h-3.5 text-blue-600" /> IT Operator
              </div>
              <p className="text-textSubtle leading-relaxed">
                Requests remote sessions, acknowledges alerts, executes non-destructive runbooks.
              </p>
            </div>
            <div className="space-y-1">
              <div className="font-semibold text-textMain flex items-center gap-1.5">
                <KeyRound className="w-3.5 h-3.5 text-stone-600" /> Read-Only Viewer
              </div>
              <p className="text-textSubtle leading-relaxed">
                Auditor view: inspects live graphs, audit trails, and incidents without execution access.
              </p>
            </div>
          </div>

          {/* Filter Bar */}
          <div className="border border-white/50 bg-surface backdrop-blur-md rounded-xl p-4 shadow-sm glass-panel flex flex-col md:flex-row items-center justify-between gap-3">
            <div className="relative w-full md:w-80">
              <Search className="w-4 h-4 text-textSubtle absolute left-3 top-2.5" />
              <input
                type="text"
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                placeholder="Search member name or email..."
                className="w-full bg-white border border-borderSubtle rounded-lg pl-9 pr-3 py-1.5 text-xs text-textMain focus:outline-none focus:border-primary transition-colors shadow-inner"
              />
            </div>

            <div className="flex items-center gap-1.5 overflow-x-auto pb-1 md:pb-0">
              {(
                [
                  { key: "all", label: "All Roles" },
                  { key: "owner_admin", label: "Owner" },
                  { key: "it_admin", label: "IT Admins" },
                  { key: "it_operator", label: "Operators" },
                  { key: "readonly_viewer", label: "Read-Only" },
                ] as const
              ).map((tab) => (
                <button
                  key={tab.key}
                  onClick={() => setFilterRole(tab.key)}
                  className={`px-3 py-1 rounded-md text-xs font-medium transition-colors ${
                    filterRole === tab.key
                      ? "bg-primary text-white shadow-sm"
                      : "bg-white/60 hover:bg-white text-textMuted border border-borderSubtle"
                  }`}
                >
                  {tab.label}
                </button>
              ))}
            </div>
          </div>

          {/* Members Table */}
          <div className="border border-white/50 bg-surface backdrop-blur-md rounded-xl shadow-sm glass-panel overflow-hidden">
            <div className="px-6 py-4 border-b border-borderSubtle flex items-center justify-between">
              <h2 className="text-sm font-semibold text-textMain">Active Organization Operators</h2>
              <span className="text-xs text-textSubtle font-mono">
                {members.length} Registered Accounts
              </span>
            </div>

            <div className="overflow-x-auto">
              <table className="w-full text-left text-xs">
                <thead className="bg-black/[0.02] border-b border-borderSubtle text-[11px] font-semibold text-textMuted uppercase tracking-wider">
                  <tr>
                    <th className="px-6 py-3">Operator Name</th>
                    <th className="px-6 py-3">Assigned Role</th>
                    <th className="px-6 py-3">MFA Security</th>
                    <th className="px-6 py-3">Endpoints</th>
                    <th className="px-6 py-3">Last Active</th>
                    <th className="px-6 py-3 text-right">Actions</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-borderSubtle/60">
                  {filteredMembers.map((member) => (
                    <tr key={member.id} className="hover:bg-white/40 transition-colors">
                      <td className="px-6 py-4">
                        <div className="font-semibold text-textMain">{member.fullName}</div>
                        <div className="font-mono text-[11px] text-textSubtle flex items-center gap-1 mt-0.5">
                          <Mail className="w-3 h-3 text-textSubtle" />
                          {member.email}
                        </div>
                      </td>

                      <td className="px-6 py-4 whitespace-nowrap">
                        {getRoleBadge(member.role)}
                      </td>

                      <td className="px-6 py-4 whitespace-nowrap">
                        {member.mfaEnrolled ? (
                          <span className="inline-flex items-center gap-1 text-emerald-700 font-medium text-[11px]">
                            <ShieldCheck className="w-3.5 h-3.5 text-emerald-600" />
                            TOTP Active
                          </span>
                        ) : (
                          <span className="inline-flex items-center gap-1 text-amber-700 font-medium text-[11px]">
                            <ShieldAlert className="w-3.5 h-3.5 text-amber-600" />
                            Pending Setup
                          </span>
                        )}
                      </td>

                      <td className="px-6 py-4 whitespace-nowrap font-mono text-[11px] text-textMuted">
                        {member.assignedEndpointsCount} nodes
                      </td>

                      <td className="px-6 py-4 whitespace-nowrap text-[11px] text-textSubtle font-mono">
                        {member.lastActive}
                      </td>

                      <td className="px-6 py-4 text-right whitespace-nowrap">
                        <div className="flex items-center justify-end gap-2">
                          {/* Role selector dropdown */}
                          <select
                            value={member.role}
                            disabled={member.role === "owner_admin"}
                            onChange={(e) => handleRoleChange(member.id, e.target.value as RoleType)}
                            className="bg-white border border-borderSubtle rounded px-2 py-1 text-[11px] text-textMain focus:outline-none focus:border-primary disabled:opacity-50"
                          >
                            <option value="owner_admin">Owner Admin</option>
                            <option value="it_admin">IT Admin</option>
                            <option value="it_operator">IT Operator</option>
                            <option value="readonly_viewer">Read-Only</option>
                          </select>

                          {member.role !== "owner_admin" && (
                            <button
                              onClick={() => handleRemoveMember(member.id)}
                              className="p-1 rounded text-textSubtle hover:text-rose-600 hover:bg-rose-50 transition-colors"
                              title="Revoke access"
                            >
                              <Trash2 className="w-3.5 h-3.5" />
                            </button>
                          )}
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        </main>
      </div>

      {/* Invite Member Modal */}
      {isInviteModalOpen && (
        <div className="fixed inset-0 z-50 bg-black/40 backdrop-blur-sm flex items-center justify-center p-4">
          <div className="bg-white rounded-xl shadow-2xl border border-borderSubtle max-w-lg w-full p-6 space-y-4">
            <div className="flex items-center justify-between border-b border-borderSubtle pb-3">
              <div className="flex items-center gap-2 text-textMain font-bold text-base">
                <UserPlus className="w-5 h-5 text-primary" />
                Invite Team Member
              </div>
              <button
                onClick={() => setIsInviteModalOpen(false)}
                className="text-textSubtle hover:text-textMain p-1 rounded-md"
              >
                <X className="w-4 h-4" />
              </button>
            </div>

            {!generatedInviteLink ? (
              <form onSubmit={handleGenerateInvite} className="space-y-4">
                <div className="space-y-1.5">
                  <label className="text-xs font-semibold text-textMain">Operator Full Name</label>
                  <input
                    type="text"
                    required
                    value={inviteName}
                    onChange={(e) => setInviteName(e.target.value)}
                    placeholder="Jordan Blake"
                    className="w-full bg-white border border-borderSubtle rounded-lg px-3 py-2 text-xs text-textMain focus:outline-none focus:border-primary shadow-sm"
                  />
                </div>

                <div className="space-y-1.5">
                  <label className="text-xs font-semibold text-textMain">Work Email Address</label>
                  <input
                    type="email"
                    required
                    value={inviteEmail}
                    onChange={(e) => setInviteEmail(e.target.value)}
                    placeholder="jordan.blake@company.com"
                    className="w-full bg-white border border-borderSubtle rounded-lg px-3 py-2 text-xs text-textMain focus:outline-none focus:border-primary shadow-sm"
                  />
                </div>

                <div className="space-y-1.5">
                  <label className="text-xs font-semibold text-textMain">Assigned RBAC Role</label>
                  <select
                    value={inviteRole}
                    onChange={(e) => setInviteRole(e.target.value as RoleType)}
                    className="w-full bg-white border border-borderSubtle rounded-lg px-3 py-2 text-xs text-textMain focus:outline-none focus:border-primary shadow-sm"
                  >
                    <option value="it_admin">IT Admin (Runbooks, remote sessions & approvals)</option>
                    <option value="it_operator">IT Operator (Standard incident investigator & alerts)</option>
                    <option value="readonly_viewer">Read-Only Viewer (Telemetry inspection only)</option>
                  </select>
                </div>

                <div className="p-3 bg-stone-50 rounded-lg border border-borderSubtle text-[11px] text-textMuted flex items-start gap-2">
                  <Sparkles className="w-4 h-4 text-primary shrink-0 mt-0.5" />
                  <span>
                    The invitee will receive an onboard invitation requiring mandatory RFC 6238 TOTP enrollment before accessing workspace devices.
                  </span>
                </div>

                <div className="flex justify-end gap-2 pt-2 border-t border-borderSubtle">
                  <button
                    type="button"
                    onClick={() => setIsInviteModalOpen(false)}
                    className="px-4 py-2 rounded-lg border border-borderSubtle text-xs font-medium text-textMuted hover:bg-stone-50 transition-colors"
                  >
                    Cancel
                  </button>
                  <button
                    type="submit"
                    className="px-4 py-2 rounded-lg bg-primary hover:bg-primaryHover text-white text-xs font-semibold shadow-sm transition-colors"
                  >
                    Generate Invitation Link
                  </button>
                </div>
              </form>
            ) : (
              <div className="space-y-4">
                <div className="p-4 bg-emerald-50 rounded-xl border border-emerald-200/50 space-y-2">
                  <div className="flex items-center gap-2 text-emerald-800 font-semibold text-xs">
                    <CheckCircle2 className="w-4 h-4 text-emerald-600" />
                    Invitation Link Ready
                  </div>
                  <p className="text-xs text-emerald-700">
                    Send this single-use secure link to <strong>{inviteEmail}</strong> to activate their account.
                  </p>
                </div>

                <div className="space-y-1.5">
                  <label className="text-xs font-semibold text-textMain">One-Click Join Link</label>
                  <div className="flex items-center gap-2">
                    <input
                      type="text"
                      readOnly
                      value={generatedInviteLink}
                      className="w-full bg-stone-50 border border-borderSubtle rounded-lg px-3 py-2 text-xs font-mono text-textMuted shadow-inner select-all"
                    />
                    <button
                      onClick={copyInviteLink}
                      className="px-3 py-2 rounded-lg bg-primary hover:bg-primaryHover text-white text-xs font-semibold flex items-center gap-1.5 shrink-0 shadow-sm transition-colors"
                    >
                      {copiedLink ? <Check className="w-4 h-4 text-emerald-200" /> : <Copy className="w-4 h-4" />}
                      {copiedLink ? "Copied" : "Copy"}
                    </button>
                  </div>
                </div>

                <div className="flex justify-end pt-2 border-t border-borderSubtle">
                  <button
                    type="button"
                    onClick={() => setIsInviteModalOpen(false)}
                    className="px-4 py-2 rounded-lg bg-white border border-borderSubtle text-xs font-semibold text-textMain hover:bg-stone-50 transition-colors"
                  >
                    Done
                  </button>
                </div>
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );
}
