"use client";

import { useState } from "react";
import { useRouter } from "next/navigation";
import Link from "next/link";
import {
  ArrowRight,
  Lock,
  Mail,
  Building2,
  User,
  ShieldCheck,
  CheckCircle2,
  Copy,
  Check,
  KeyRound,
  AlertCircle,
  QrCode,
  Sparkles,
} from "lucide-react";
import { AmbientBackground } from "@/components/AmbientBackground";

export default function SignupPage() {
  const router = useRouter();

  // Multi-step onboarding flow: 1 = Org & Admin, 2 = TOTP MFA Setup, 3 = Completed / Redirect
  const [step, setStep] = useState<1 | 2>(1);

  // Form states
  const [companyName, setCompanyName] = useState("");
  const [subdomain, setSubdomain] = useState("");
  const [fullName, setFullName] = useState("");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [mfaCode, setMfaCode] = useState("");

  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [copiedSecret, setCopiedSecret] = useState(false);

  // RFC 6238 TOTP Secret Mock
  const totpSecret = "JBSWY3DPEHPK3PXP";

  // Auto-slugify company name to subdomain
  const handleCompanyChange = (val: string) => {
    setCompanyName(val);
    const slug = val
      .toLowerCase()
      .replace(/[^a-z0-9]/g, "-")
      .replace(/-+/g, "-")
      .replace(/^-|-$/g, "");
    setSubdomain(slug);
  };

  const handleStep1Submit = (e: React.FormEvent) => {
    e.preventDefault();
    setError(null);

    if (!companyName.trim()) {
      setError("Please enter your organization or company name.");
      return;
    }
    if (!subdomain.trim() || subdomain.length < 3) {
      setError("Please choose a valid workspace subdomain (at least 3 characters).");
      return;
    }
    if (!fullName.trim()) {
      setError("Please provide your full administrator name.");
      return;
    }
    if (!email.trim() || !email.includes("@")) {
      setError("Please enter a valid work email address.");
      return;
    }
    if (password.length < 8) {
      setError("Password must be at least 8 characters long.");
      return;
    }

    // Proceed to Step 2 (TOTP MFA setup)
    setStep(2);
  };

  const handleMfaSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    setError(null);

    if (mfaCode.trim().length !== 6 || !/^\d{6}$/.test(mfaCode.trim())) {
      setError("Please enter a valid 6-digit numeric authenticator code.");
      setLoading(false);
      return;
    }

    try {
      // In production, this verifies with POST /api/v1/auth/register
      // Here we store tenant token & subdomain in localStorage and redirect
      const generatedToken = "mp-jwt-" + Math.random().toString(36).substring(2, 12);
      localStorage.setItem("monitopro_token", generatedToken);
      localStorage.setItem("monitopro_subdomain", subdomain);
      localStorage.setItem("monitopro_admin_name", fullName);
      localStorage.setItem("monitopro_admin_email", email);
      localStorage.setItem("monitopro_org_name", companyName);

      // Brief delay for UX animation
      await new Promise((resolve) => setTimeout(resolve, 600));
      router.push(`/${subdomain}/overview`);
    } catch {
      setError("Registration failed. Please try again.");
    } finally {
      setLoading(false);
    }
  };

  const copySecretToClipboard = () => {
    navigator.clipboard.writeText(totpSecret);
    setCopiedSecret(true);
    setTimeout(() => setCopiedSecret(false), 2000);
  };

  return (
    <div className="min-h-screen bg-bg text-textMain flex relative overflow-hidden">
      <AmbientBackground />

      {/* Left side preview matching V1 */}
      <div className="hidden lg:flex flex-1 border-r border-borderSubtle bg-white/20 flex-col items-center justify-center p-12 relative overflow-hidden">
        <div className="w-full max-w-md border border-white/50 bg-surface backdrop-blur-md rounded-xl p-8 z-20 shadow-2xl glass-panel space-y-6">
          <div className="flex items-center justify-between border-b border-borderSubtle pb-4">
            <div className="flex items-center gap-3">
              <div className="w-9 h-9 rounded-lg bg-black/[0.03] border border-borderSubtle flex items-center justify-center text-primary">
                <ShieldCheck className="w-5 h-5" />
              </div>
              <div>
                <div className="text-[13px] font-semibold text-textMain">Enterprise Multi-Tenant</div>
                <div className="text-[11px] text-textSubtle font-mono">Isolated Cryptographic Vault</div>
              </div>
            </div>
            <span className="px-2.5 py-1 rounded-full bg-emerald-50 border border-emerald-200/50 text-emerald-600 text-[10px] font-mono uppercase tracking-wider font-semibold">
              Zero-Trust
            </span>
          </div>

          <div className="space-y-4">
            <div className="text-xs text-textMuted leading-relaxed">
              Every MonitorPro tenant operates on an isolated security boundary with RFC 6238 Time-Based One-Time Passwords (TOTP), Ed25519 node mutual authentication, and immutable event journaling.
            </div>

            <div className="p-4 rounded-lg bg-white/40 border border-borderSubtle space-y-2.5 font-mono text-[11px]">
              <div className="flex items-center justify-between text-textMuted">
                <span>Subdomain Routing:</span>
                <span className="text-primary font-bold">
                  {subdomain ? `${subdomain}.monitopro.com` : "your-tenant.monitopro.com"}
                </span>
              </div>
              <div className="flex items-center justify-between text-textMuted">
                <span>Security Engine:</span>
                <span className="text-emerald-700 font-medium">Dual-Operator Enforced</span>
              </div>
              <div className="flex items-center justify-between text-textMuted">
                <span>AI Governance:</span>
                <span className="text-textMain font-medium">PII Scrubbed + Kill-Switch</span>
              </div>
            </div>

            <div className="flex items-center gap-2 pt-2 text-[11px] text-textSubtle">
              <Sparkles className="w-4 h-4 text-primary shrink-0" />
              <span>Includes 14-day full infrastructure telemetry trial with no limits.</span>
            </div>
          </div>
        </div>
      </div>

      {/* Right side form */}
      <div className="flex-1 flex flex-col justify-center items-center px-6 py-12 relative z-10 overflow-y-auto">
        <div className="w-full max-w-md space-y-6">
          {/* Header */}
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
            <h1 className="text-2xl font-bold tracking-tight text-textMain">Create Organization Workspace</h1>
            <p className="text-xs text-textMuted">
              {step === 1 ? "Configure your dedicated tenant domain and primary administrator" : "Enroll your authenticator app for RFC 6238 TOTP multi-factor security"}
            </p>
          </div>

          {/* Stepper indicator */}
          <div className="flex items-center justify-center gap-2 py-1">
            <div className={`flex items-center gap-1.5 text-xs font-medium ${step === 1 ? "text-primary font-semibold" : "text-emerald-600"}`}>
              <div className={`w-5 h-5 rounded-full flex items-center justify-center text-[10px] font-bold ${step === 1 ? "bg-primary text-white" : "bg-emerald-500 text-white"}`}>
                {step === 1 ? "1" : "✓"}
              </div>
              <span>Workspace Info</span>
            </div>
            <div className="w-8 h-[1px] bg-borderSubtle" />
            <div className={`flex items-center gap-1.5 text-xs font-medium ${step === 2 ? "text-primary font-semibold" : "text-textSubtle"}`}>
              <div className={`w-5 h-5 rounded-full flex items-center justify-center text-[10px] font-bold ${step === 2 ? "bg-primary text-white" : "bg-black/10 text-textSubtle"}`}>
                2
              </div>
              <span>MFA Setup</span>
            </div>
          </div>

          {/* Form Card */}
          <div className="rounded-xl border border-white/50 bg-surface p-6 shadow-xl glass-panel">
            {error && (
              <div className="mb-4 p-3 rounded-lg bg-rose-50 border border-rose-200/50 text-rose-700 text-xs flex items-center gap-2">
                <AlertCircle className="w-4 h-4 flex-shrink-0" />
                <span>{error}</span>
              </div>
            )}

            {step === 1 ? (
              <form onSubmit={handleStep1Submit} className="space-y-4">
                <div className="space-y-1.5">
                  <label className="text-xs font-medium text-textMuted">Organization / Company Name</label>
                  <div className="relative flex items-center">
                    <Building2 className="w-4 h-4 text-textSubtle absolute left-3" />
                    <input
                      type="text"
                      required
                      value={companyName}
                      onChange={(e) => handleCompanyChange(e.target.value)}
                      className="w-full bg-white border border-borderSubtle rounded-lg pl-9 pr-3 py-2 text-sm text-textMain focus:outline-none focus:border-primary transition-colors shadow-sm"
                      placeholder="Acme Global Inc"
                    />
                  </div>
                </div>

                <div className="space-y-1.5">
                  <label className="text-xs font-medium text-textMuted">Tenant Subdomain URL</label>
                  <div className="relative flex items-center">
                    <span className="text-xs text-textSubtle font-mono absolute left-3 select-none">https://</span>
                    <input
                      type="text"
                      required
                      value={subdomain}
                      onChange={(e) => setSubdomain(e.target.value.toLowerCase().replace(/[^a-z0-9-]/g, ""))}
                      className="w-full bg-white border border-borderSubtle rounded-lg pl-16 pr-24 py-2 text-sm text-textMain focus:outline-none focus:border-primary font-mono transition-colors shadow-sm"
                      placeholder="acme"
                    />
                    <span className="absolute right-3 text-xs font-mono text-textSubtle pointer-events-none">
                      .monitopro.com
                    </span>
                  </div>
                  <p className="text-[11px] text-textSubtle">This will be your team’s dedicated private login gateway.</p>
                </div>

                <div className="border-t border-borderSubtle pt-3 space-y-4">
                  <div className="space-y-1.5">
                    <label className="text-xs font-medium text-textMuted">Administrator Full Name</label>
                    <div className="relative flex items-center">
                      <User className="w-4 h-4 text-textSubtle absolute left-3" />
                      <input
                        type="text"
                        required
                        value={fullName}
                        onChange={(e) => setFullName(e.target.value)}
                        className="w-full bg-white border border-borderSubtle rounded-lg pl-9 pr-3 py-2 text-sm text-textMain focus:outline-none focus:border-primary transition-colors shadow-sm"
                        placeholder="Alex Morgan"
                      />
                    </div>
                  </div>

                  <div className="space-y-1.5">
                    <label className="text-xs font-medium text-textMuted">Work Email</label>
                    <div className="relative flex items-center">
                      <Mail className="w-4 h-4 text-textSubtle absolute left-3" />
                      <input
                        type="email"
                        required
                        value={email}
                        onChange={(e) => setEmail(e.target.value)}
                        className="w-full bg-white border border-borderSubtle rounded-lg pl-9 pr-3 py-2 text-sm text-textMain focus:outline-none focus:border-primary transition-colors shadow-sm"
                        placeholder="alex.morgan@company.com"
                      />
                    </div>
                  </div>

                  <div className="space-y-1.5">
                    <label className="text-xs font-medium text-textMuted">Master Password (min 8 characters)</label>
                    <div className="relative flex items-center">
                      <Lock className="w-4 h-4 text-textSubtle absolute left-3" />
                      <input
                        type="password"
                        required
                        minLength={8}
                        value={password}
                        onChange={(e) => setPassword(e.target.value)}
                        className="w-full bg-white border border-borderSubtle rounded-lg pl-9 pr-3 py-2 text-sm text-textMain focus:outline-none focus:border-primary transition-colors shadow-sm"
                        placeholder="••••••••••••"
                      />
                    </div>
                  </div>
                </div>

                <button
                  type="submit"
                  className="w-full mt-2 bg-primary hover:bg-primaryHover text-white font-semibold py-2.5 rounded-lg transition-colors text-sm flex items-center justify-center gap-2 shadow-sm"
                >
                  Continue to MFA Setup
                  <ArrowRight className="w-4 h-4" />
                </button>
              </form>
            ) : (
              <form onSubmit={handleMfaSubmit} className="space-y-5">
                <div className="text-center space-y-1">
                  <div className="text-sm font-semibold text-textMain">Scan Authenticator QR Code</div>
                  <p className="text-xs text-textMuted">
                    Use Google Authenticator, 1Password, or Microsoft Authenticator to scan.
                  </p>
                </div>

                {/* Stylized QR Code Graphic Container */}
                <div className="flex flex-col items-center justify-center p-5 bg-white rounded-xl border border-borderSubtle shadow-inner">
                  <div className="relative w-36 h-36 border-2 border-dashed border-[#B87C3B]/40 rounded-lg p-2 bg-stone-50 flex items-center justify-center">
                    {/* Simulated Clean QR Code Grid */}
                    <div className="grid grid-cols-6 gap-1 w-full h-full p-1">
                      {Array.from({ length: 36 }).map((_, i) => {
                        const isCorner =
                          (i < 3 && i % 6 < 3) ||
                          (i >= 18 && i < 21 && i % 6 < 3) ||
                          (i % 6 >= 3 && i < 12);
                        const isFilled = isCorner || (i * 7) % 3 === 0 || (i * 11) % 5 === 0;
                        return (
                          <div
                            key={i}
                            className={`rounded-sm ${isFilled ? "bg-[#1E1A17]" : "bg-transparent"}`}
                          />
                        );
                      })}
                    </div>
                    <div className="absolute inset-0 flex items-center justify-center pointer-events-none">
                      <div className="w-7 h-7 bg-white rounded shadow border border-borderSubtle flex items-center justify-center">
                        <KeyRound className="w-4 h-4 text-primary" />
                      </div>
                    </div>
                  </div>

                  <div className="mt-3 text-center">
                    <span className="text-[11px] text-textSubtle font-medium">Manual Secret Key:</span>
                    <div className="mt-1 flex items-center justify-center gap-1.5 font-mono text-xs font-bold text-textMain bg-black/[0.04] px-2.5 py-1 rounded border border-borderSubtle">
                      <span>{totpSecret}</span>
                      <button
                        type="button"
                        onClick={copySecretToClipboard}
                        className="text-textMuted hover:text-primary transition-colors ml-1 p-0.5"
                        title="Copy Secret"
                      >
                        {copiedSecret ? <Check className="w-3.5 h-3.5 text-emerald-600" /> : <Copy className="w-3.5 h-3.5" />}
                      </button>
                    </div>
                  </div>
                </div>

                <div className="space-y-1.5">
                  <label className="text-xs font-medium text-textMuted text-center block">
                    Enter 6-digit verification code from app
                  </label>
                  <input
                    type="text"
                    inputMode="numeric"
                    pattern="[0-9]*"
                    maxLength={6}
                    required
                    autoFocus
                    value={mfaCode}
                    onChange={(e) => setMfaCode(e.target.value.replace(/[^0-9]/g, ""))}
                    className="w-full text-center tracking-[0.5em] text-lg font-mono font-bold bg-white border border-borderSubtle rounded-lg py-2.5 text-textMain focus:outline-none focus:border-primary transition-colors shadow-sm"
                    placeholder="000000"
                  />
                </div>

                <div className="flex gap-2 pt-1">
                  <button
                    type="button"
                    onClick={() => setStep(1)}
                    className="flex-1 border border-borderSubtle hover:bg-white text-textMuted font-medium py-2 rounded-lg transition-colors text-xs"
                  >
                    Back
                  </button>
                  <button
                    type="submit"
                    disabled={loading || mfaCode.length !== 6}
                    className="flex-[2] bg-primary hover:bg-primaryHover text-white font-semibold py-2 rounded-lg transition-colors text-sm flex items-center justify-center gap-2 shadow-sm disabled:opacity-50"
                  >
                    {loading ? "Activating Workspace..." : "Complete Setup & Launch"}
                    {!loading && <CheckCircle2 className="w-4 h-4" />}
                  </button>
                </div>
              </form>
            )}
          </div>

          {/* Footer Navigation */}
          <div className="text-center text-xs text-textMuted space-y-2">
            <div>
              Already have an organization workspace?{" "}
              <Link href="/login" className="text-primary hover:text-primaryHover font-medium underline">
                Sign In
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
