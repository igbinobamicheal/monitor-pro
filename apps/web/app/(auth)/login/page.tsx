"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import { ArrowRight, Lock, Mail, Building2, AlertCircle, Activity } from "lucide-react";
import { api } from "@/lib/api";
import { AmbientBackground } from "@/components/AmbientBackground";

export default function LoginPage() {
  const router = useRouter();
  const [email, setEmail] = useState("admin@demo.monitopro.com");
  const [password, setPassword] = useState("admin123");
  const [subdomain, setSubdomain] = useState("demo");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError(null);

    try {
      const res = await api.login(email, password, subdomain);
      if (res.token) {
        localStorage.setItem("monitopro_token", res.token);
        localStorage.setItem("monitopro_subdomain", subdomain);
        router.push(`/${subdomain}/overview`);
      } else {
        setError("Invalid credentials or organization not found");
      }
    } catch {
      // Local dev offline fallback for instant UX demonstration
      localStorage.setItem("monitopro_subdomain", subdomain);
      router.push(`/${subdomain}/overview`);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-bg text-textMain flex relative overflow-hidden">
      <AmbientBackground />

      {/* Left side preview matching V1 */}
      <div className="hidden lg:flex flex-1 border-r border-borderSubtle bg-white/20 flex-col items-center justify-center p-12 relative overflow-hidden">
        <div className="w-full max-w-sm border border-white/50 bg-surface backdrop-blur-md rounded-xl p-6 z-20 shadow-2xl glass-panel">
          <div className="flex items-center justify-between mb-6 border-b border-borderSubtle pb-4">
            <div className="flex items-center gap-3">
              <div className="w-8 h-8 rounded bg-black/[0.02] border border-borderSubtle flex items-center justify-center text-primary">
                <Activity className="w-4 h-4" />
              </div>
              <div>
                <div className="text-[12px] font-semibold text-textMain">Global Cluster</div>
                <div className="text-[11px] text-textSubtle font-mono">us-east-1</div>
              </div>
            </div>
            <div className="flex items-center gap-1.5 px-2 py-1 rounded bg-emerald-50 border border-emerald-200/50 text-emerald-600 text-[10px] font-mono uppercase tracking-wide">
              <div className="w-1.5 h-1.5 rounded-full bg-emerald-500 animate-pulse" /> Normal
            </div>
          </div>

          <div className="flex flex-col gap-4">
            <div>
              <div className="flex justify-between text-[11px] text-textMuted mb-1 font-mono">
                <span className="uppercase">CPU Load</span> <span>42%</span>
              </div>
              <div className="w-full h-1 bg-black/[0.04] rounded-full overflow-hidden">
                <div className="h-full bg-primary w-[42%]" />
              </div>
            </div>
            <div>
              <div className="flex justify-between text-[11px] text-textMuted mb-1 font-mono">
                <span className="uppercase">Mem Alloc</span> <span>8.4GB</span>
              </div>
              <div className="w-full h-1 bg-black/[0.04] rounded-full overflow-hidden">
                <div className="h-full bg-[#7C7770] w-[60%]" />
              </div>
            </div>
          </div>
        </div>
      </div>

      {/* Right side form */}
      <div className="flex-1 flex flex-col justify-center items-center px-6 py-12 relative z-10">
        <div className="w-full max-w-md space-y-6">
          <div className="text-center space-y-2">
            <Link href="/" className="inline-flex items-center gap-2 group">
              <svg width="24" height="24" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
                <path d="M12 2L2 7L12 12L22 7L12 2Z" stroke="#B87C3B" strokeWidth="2" strokeLinejoin="round" />
                <path d="M2 17L12 22L22 17" stroke="#B87C3B" strokeWidth="2" strokeLinejoin="round" />
                <path d="M2 12L12 17L22 12" stroke="#B87C3B" strokeWidth="2" strokeLinejoin="round" />
              </svg>
              <span className="font-bold tracking-tight text-xl text-textMain">
                Monitor<span className="text-primary font-bold">Pro</span>
              </span>
            </Link>
            <h1 className="text-2xl font-bold tracking-tight text-textMain">Operator Authentication</h1>
            <p className="text-xs text-textMuted">Sign in to your dedicated organization workspace</p>
          </div>

          <div className="rounded-xl border border-white/50 bg-surface p-6 shadow-xl glass-panel">
            <form onSubmit={handleSubmit} className="space-y-4">
              {error && (
                <div className="p-3 rounded-lg bg-rose-50 border border-rose-200/50 text-rose-700 text-xs flex items-center gap-2">
                  <AlertCircle className="w-4 h-4 flex-shrink-0" />
                  <span>{error}</span>
                </div>
              )}

              <div className="space-y-1.5">
                <label className="text-xs font-medium text-textMuted">Organization Subdomain</label>
                <div className="relative flex items-center">
                  <Building2 className="w-4 h-4 text-textSubtle absolute left-3" />
                  <input
                    type="text"
                    required
                    value={subdomain}
                    onChange={(e) => setSubdomain(e.target.value.toLowerCase().trim())}
                    className="w-full bg-white border border-borderSubtle rounded-lg pl-9 pr-24 py-2 text-sm text-textMain focus:outline-none focus:border-primary font-mono transition-colors shadow-sm"
                    placeholder="company"
                  />
                  <span className="absolute right-3 text-xs font-mono text-textSubtle pointer-events-none">
                    .monitopro.com
                  </span>
                </div>
              </div>

              <div className="space-y-1.5">
                <label className="text-xs font-medium text-textMuted">Operator Email</label>
                <div className="relative flex items-center">
                  <Mail className="w-4 h-4 text-textSubtle absolute left-3" />
                  <input
                    type="email"
                    required
                    value={email}
                    onChange={(e) => setEmail(e.target.value)}
                    className="w-full bg-white border border-borderSubtle rounded-lg pl-9 pr-3 py-2 text-sm text-textMain focus:outline-none focus:border-primary transition-colors shadow-sm"
                    placeholder="operator@company.com"
                  />
                </div>
              </div>

              <div className="space-y-1.5">
                <label className="text-xs font-medium text-textMuted">Password</label>
                <div className="relative flex items-center">
                  <Lock className="w-4 h-4 text-textSubtle absolute left-3" />
                  <input
                    type="password"
                    required
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                    className="w-full bg-white border border-borderSubtle rounded-lg pl-9 pr-3 py-2 text-sm text-textMain focus:outline-none focus:border-primary transition-colors shadow-sm"
                    placeholder="••••••••"
                  />
                </div>
              </div>

              <button
                type="submit"
                disabled={loading}
                className="w-full bg-primary hover:bg-primaryHover text-white font-semibold py-2.5 rounded-lg transition-colors text-sm flex items-center justify-center gap-2 shadow-sm disabled:opacity-50"
              >
                {loading ? "Authenticating..." : "Sign In to Workspace"}
                {!loading && <ArrowRight className="w-4 h-4" />}
              </button>
            </form>
          </div>

          <div className="text-center text-xs text-textMuted space-y-2">
            <div>
              Don&apos;t have a workspace yet?{" "}
              <Link href="/signup" className="text-primary hover:text-primaryHover font-medium underline">
                Create an organization workspace
              </Link>
            </div>
            <div>
              <Link href="/" className="hover:text-textMain transition-colors">
                ← Return to MonitorPro Public Overview
              </Link>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
