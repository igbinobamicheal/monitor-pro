"use client";

import { useState } from "react";
import Link from "next/link";
import { PublicNavbar } from "@/components/Navbar";
import { AmbientBackground } from "@/components/AmbientBackground";
import {
  ArrowRight,
  Lock,
  Activity,
  ShieldCheck,
  MonitorPlay,
  Wrench,
  Check,
  Copy,
  Terminal,
  Cpu,
  Layers,
  Radio,
  Database,
  Sparkles,
} from "lucide-react";

export default function MarketingPage() {
  const [copied, setCopied] = useState(false);
  const [liveStreamActive, setLiveStreamActive] = useState(true);

  const installCommand =
    'powershell -ExecutionPolicy Bypass -Command "irm https://api.monitorpro.io/install.ps1 | iex"';

  const handleCopy = () => {
    navigator.clipboard.writeText(installCommand);
    setCopied(true);
    setTimeout(() => setCopied(false), 2400);
  };

  return (
    <div className="min-h-screen bg-bg text-textMain flex flex-col relative selection:bg-primary/20 selection:text-primary">
      <AmbientBackground />
      <PublicNavbar />

      {/* Hero Section */}
      <section className="pt-36 pb-24 lg:pt-44 lg:pb-32 px-6 max-w-5xl mx-auto w-full text-center flex flex-col items-center">
        {/* Subtle pill badge */}
        <div className="inline-flex items-center gap-2.5 px-3.5 py-1.5 rounded-full border border-primary/25 bg-white/40 backdrop-blur-md text-primary text-xs font-mono font-medium tracking-wide shadow-sm mb-8 animate-fade-in">
          <span className="w-2 h-2 rounded-full bg-primary animate-pulse" />
          <span>MONITORPRO V2 • ZERO-TRUST ENDPOINT OBSERVABILITY</span>
        </div>

        {/* Confident, clean typography */}
        <h1 className="text-4xl sm:text-6xl lg:text-7xl font-bold tracking-tight text-textMain leading-[1.08] max-w-4xl">
          Proactive IT Operations,{" "}
          <span className="text-primary italic font-serif font-normal">
            Simplified.
          </span>
        </h1>

        {/* Subtext: Clear, concise 2-line explanation */}
        <p className="mt-6 text-textMuted text-base sm:text-lg max-w-2xl mx-auto leading-relaxed font-normal">
          Sub-second Windows telemetry, AI-assisted root-cause diagnosis, and policy-governed remediation—packaged in a single, zero-footprint architecture.
        </p>

        {/* Focused Call-to-Actions */}
        <div className="flex flex-wrap items-center justify-center gap-4 mt-9">
          <Link
            href="/signup"
            className="flex items-center gap-2 bg-primary hover:bg-primaryHover text-white font-medium px-7 py-3.5 rounded-xl transition-all shadow-[0_4px_20px_rgba(184,124,59,0.32)] hover:shadow-[0_6px_26px_rgba(184,124,59,0.45)] hover:-translate-y-0.5 text-sm tracking-wide"
          >
            Launch Free Workspace <ArrowRight className="w-4 h-4" />
          </Link>

          <Link
            href="/login"
            className="flex items-center gap-2 border border-white/70 bg-white/50 hover:bg-white/80 backdrop-blur-xl px-7 py-3.5 rounded-xl transition-all text-sm font-medium text-textMain shadow-[0_4px_16px_rgba(30,26,23,0.03)] hover:-translate-y-0.5"
          >
            <Lock className="w-3.5 h-3.5 text-textSubtle" /> Operator Sign In
          </Link>
        </div>

        {/* Single Sleek Floating Glass Showcase */}
        <div className="w-full max-w-2xl mt-16 sm:mt-20">
          <div className="rounded-2xl bg-white/65 backdrop-blur-2xl border border-white/60 p-6 sm:p-7 shadow-[0_20px_50px_rgba(184,124,59,0.08)] transition-all hover:border-white/80">
            {/* Header: Live pulse beacon & Telemetry trigger */}
            <div className="flex items-center justify-between pb-5 border-b border-borderSubtle">
              <div className="flex items-center gap-3">
                <span className="relative flex h-2.5 w-2.5">
                  {liveStreamActive && (
                    <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-status-online opacity-75" />
                  )}
                  <span className="relative inline-flex rounded-full h-2.5 w-2.5 bg-status-online" />
                </span>
                <span className="font-mono text-xs font-semibold text-textMain tracking-tight">
                  srv-enterprise-01
                </span>
                <span className="text-[11px] font-mono text-textSubtle">• Online</span>
              </div>

              <button
                type="button"
                onClick={() => setLiveStreamActive(!liveStreamActive)}
                className="flex items-center gap-1.5 px-3 py-1 rounded-full text-[11px] font-mono font-medium transition-all bg-primary/10 text-primary border border-primary/20 hover:bg-primary/15"
              >
                <Activity className="w-3 h-3" />
                <span>{liveStreamActive ? "Live Telemetry Active" : "Paused"}</span>
              </button>
            </div>

            {/* 3 Clean Uncluttered Metric Readouts */}
            <div className="grid grid-cols-3 gap-4 pt-5">
              <div className="p-4 rounded-xl bg-white/50 border border-white/60 text-left space-y-1">
                <div className="text-[11px] font-mono uppercase tracking-wider text-textSubtle">
                  CPU
                </div>
                <div className="text-2xl font-mono font-bold text-primary">24%</div>
                <div className="text-[10px] text-textSubtle font-mono">16 Cores • Normal</div>
              </div>

              <div className="p-4 rounded-xl bg-white/50 border border-white/60 text-left space-y-1">
                <div className="text-[11px] font-mono uppercase tracking-wider text-textSubtle">
                  Memory
                </div>
                <div className="text-2xl font-mono font-bold text-textMain">6.2 GB</div>
                <div className="text-[10px] text-textSubtle font-mono">of 32 GB (19%)</div>
              </div>

              <div className="p-4 rounded-xl bg-white/50 border border-white/60 text-left space-y-1">
                <div className="text-[11px] font-mono uppercase tracking-wider text-textSubtle">
                  NVMe
                </div>
                <div className="text-2xl font-mono font-bold text-textMain">42%</div>
                <div className="text-[10px] text-textSubtle font-mono">Primary Volume C:</div>
              </div>
            </div>
          </div>
        </div>
      </section>

      {/* Feature Grid: 3 Spacious Frosted Glass Capability Cards */}
      <section id="capabilities" className="py-24 px-6 max-w-6xl mx-auto w-full">
        <div className="text-center max-w-2xl mx-auto mb-16 space-y-3">
          <span className="text-xs font-mono uppercase tracking-widest text-primary font-semibold">
            Core Capabilities
          </span>
          <h2 className="text-3xl sm:text-4xl font-bold tracking-tight text-textMain">
            Observability Without Cognitive Noise
          </h2>
          <p className="text-textMuted text-sm sm:text-base leading-relaxed">
            Eliminate alert fatigue. Every telemetry feed, support session, and remediation action is verified and accountable.
          </p>
        </div>

        <div className="grid grid-cols-1 md:grid-cols-3 gap-8">
          {/* Card 1: Continuous Telemetry */}
          <div className="p-8 sm:p-9 rounded-2xl bg-white/60 backdrop-blur-xl border border-white/60 shadow-[0_8px_30px_rgba(30,26,23,0.03)] hover:shadow-[0_12px_40px_rgba(184,124,59,0.08)] transition-all duration-300 flex flex-col justify-between group">
            <div className="space-y-5">
              <div className="w-12 h-12 rounded-xl bg-primary/10 border border-primary/20 flex items-center justify-center text-primary group-hover:scale-105 transition-transform">
                <Activity className="w-6 h-6" />
              </div>
              <h3 className="text-lg font-bold text-textMain tracking-tight">
                Continuous Telemetry
              </h3>
              <p className="text-textMuted text-sm leading-relaxed">
                Lightweight 30-second heartbeats streamed via native Rust agents with bounded disk buffering during network interruptions.
              </p>
            </div>
            <div className="pt-6 mt-6 border-t border-borderSubtle/60 flex items-center gap-2 text-xs font-mono text-textSubtle">
              <span className="w-1.5 h-1.5 rounded-full bg-status-online" />
              Win32 Perf Counters &bull; Zero Overhead
            </div>
          </div>

          {/* Card 2: Native Remote Support */}
          <div className="p-8 sm:p-9 rounded-2xl bg-white/60 backdrop-blur-xl border border-white/60 shadow-[0_8px_30px_rgba(30,26,23,0.03)] hover:shadow-[0_12px_40px_rgba(184,124,59,0.08)] transition-all duration-300 flex flex-col justify-between group">
            <div className="space-y-5">
              <div className="w-12 h-12 rounded-xl bg-primary/10 border border-primary/20 flex items-center justify-center text-primary group-hover:scale-105 transition-transform">
                <MonitorPlay className="w-6 h-6" />
              </div>
              <h3 className="text-lg font-bold text-textMain tracking-tight">
                Native Remote Support
              </h3>
              <p className="text-textMuted text-sm leading-relaxed">
                Direct DirectX Desktop Duplication with sub-100ms latency. Governed by unbypassable local user consent dialogs.
              </p>
            </div>
            <div className="pt-6 mt-6 border-t border-borderSubtle/60 flex items-center gap-2 text-xs font-mono text-textSubtle">
              <ShieldCheck className="w-3.5 h-3.5 text-primary" />
              Explicit Consent Gated
            </div>
          </div>

          {/* Card 3: Governed Remediation */}
          <div className="p-8 sm:p-9 rounded-2xl bg-white/60 backdrop-blur-xl border border-white/60 shadow-[0_8px_30px_rgba(30,26,23,0.03)] hover:shadow-[0_12px_40px_rgba(184,124,59,0.08)] transition-all duration-300 flex flex-col justify-between group">
            <div className="space-y-5">
              <div className="w-12 h-12 rounded-xl bg-primary/10 border border-primary/20 flex items-center justify-center text-primary group-hover:scale-105 transition-transform">
                <Wrench className="w-6 h-6" />
              </div>
              <h3 className="text-lg font-bold text-textMain tracking-tight">
                Governed Remediation
              </h3>
              <p className="text-textMuted text-sm leading-relaxed">
                Evidence-based AI diagnostics rank hypotheses with verified logs. Execution is restricted to allowlisted operator actions.
              </p>
            </div>
            <div className="pt-6 mt-6 border-t border-borderSubtle/60 flex items-center gap-2 text-xs font-mono text-textSubtle">
              <Sparkles className="w-3.5 h-3.5 text-primary" />
              Advisory AI &bull; Human-in-the-Loop
            </div>
          </div>
        </div>
      </section>

      {/* Live Deployment Snippet: Minimalist PowerShell Enrollment */}
      <section id="enrollment" className="py-20 px-6 max-w-4xl mx-auto w-full">
        <div className="rounded-2xl bg-white/70 backdrop-blur-2xl border border-white/70 p-8 sm:p-10 shadow-[0_16px_40px_rgba(30,26,23,0.04)] space-y-6">
          <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-4">
            <div className="space-y-1">
              <div className="flex items-center gap-2 text-xs font-mono uppercase text-primary font-semibold">
                <Terminal className="w-3.5 h-3.5" />
                <span>Zero-Touch Enrollment</span>
              </div>
              <h3 className="text-xl font-bold text-textMain tracking-tight">
                Deploy Endpoint in Seconds
              </h3>
            </div>
            <span className="text-xs font-mono text-textSubtle bg-black/[0.04] px-3 py-1 rounded-full w-fit">
              Windows Server & 10/11 &bull; 64-bit
            </span>
          </div>

          <p className="text-textMuted text-sm">
            Execute in an elevated PowerShell terminal. The service registers with cryptographic tenant isolation and initiates immediate telemetry.
          </p>

          {/* Minimalist Command Box */}
          <div className="relative group">
            <div className="rounded-xl bg-obsidian text-zinc-200 font-mono text-xs sm:text-sm p-4 sm:p-5 flex items-center justify-between gap-4 shadow-inner border border-white/10 overflow-x-auto">
              <div className="flex items-center gap-3 min-w-0">
                <span className="text-primary font-semibold select-none">&gt;</span>
                <span className="truncate select-all text-zinc-300 font-mono">
                  {installCommand}
                </span>
              </div>

              <button
                type="button"
                onClick={handleCopy}
                className="flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-mono font-medium transition-all bg-white/10 hover:bg-white/20 text-zinc-100 flex-shrink-0"
                aria-label="Copy enrollment command"
              >
                {copied ? (
                  <>
                    <Check className="w-3.5 h-3.5 text-status-online" />
                    <span className="text-status-online">Copied</span>
                  </>
                ) : (
                  <>
                    <Copy className="w-3.5 h-3.5 text-zinc-400" />
                    <span>Copy</span>
                  </>
                )}
              </button>
            </div>
          </div>
        </div>
      </section>

      {/* Architecture Highlights: Clean & Decoupled */}
      <section id="architecture" className="py-20 px-6 max-w-6xl mx-auto w-full">
        <div className="border-t border-borderSubtle/80 pt-16 space-y-12">
          <div className="text-center max-w-2xl mx-auto space-y-2">
            <span className="text-xs font-mono uppercase tracking-widest text-primary font-semibold">
              Distributed Foundations
            </span>
            <h2 className="text-2xl sm:text-3xl font-bold tracking-tight text-textMain">
              Engineering Foundations
            </h2>
          </div>

          <div className="grid grid-cols-2 md:grid-cols-4 gap-6">
            <div className="p-5 rounded-xl bg-white/40 backdrop-blur-md border border-white/50 space-y-2">
              <div className="flex items-center gap-2 text-primary">
                <Layers className="w-4 h-4" />
                <span className="font-semibold text-xs font-mono uppercase">Go 1.27 API</span>
              </div>
              <p className="text-xs text-textMuted leading-relaxed">
                Stateless load-balanced microservices with sub-millisecond route dispatch.
              </p>
            </div>

            <div className="p-5 rounded-xl bg-white/40 backdrop-blur-md border border-white/50 space-y-2">
              <div className="flex items-center gap-2 text-primary">
                <Cpu className="w-4 h-4" />
                <span className="font-semibold text-xs font-mono uppercase">Rust Agent</span>
              </div>
              <p className="text-xs text-textMuted leading-relaxed">
                Zero GC pauses, Ed25519 enrollment, and 24h offline disk caching.
              </p>
            </div>

            <div className="p-5 rounded-xl bg-white/40 backdrop-blur-md border border-white/50 space-y-2">
              <div className="flex items-center gap-2 text-primary">
                <Radio className="w-4 h-4" />
                <span className="font-semibold text-xs font-mono uppercase">NATS Stream</span>
              </div>
              <p className="text-xs text-textMuted leading-relaxed">
                Durable event streaming with strict subject partitioning and idempotency.
              </p>
            </div>

            <div className="p-5 rounded-xl bg-white/40 backdrop-blur-md border border-white/50 space-y-2">
              <div className="flex items-center gap-2 text-primary">
                <Database className="w-4 h-4" />
                <span className="font-semibold text-xs font-mono uppercase">Postgres RLS</span>
              </div>
              <p className="text-xs text-textMuted leading-relaxed">
                Row-Level Security cryptographic multi-tenant data isolation.
              </p>
            </div>
          </div>
        </div>
      </section>

      {/* Refined Minimalist Footer */}
      <footer className="border-t border-borderSubtle/60 py-12 px-6 mt-auto">
        <div className="max-w-6xl mx-auto flex flex-col sm:flex-row items-center justify-between gap-6 text-xs text-textSubtle">
          <div className="flex items-center gap-2.5">
            <div className="w-2 h-2 rounded-full bg-status-online" />
            <span className="font-medium text-textMain">MonitorPro V2</span>
            <span>&bull; Zero-Trust Endpoint Observability</span>
          </div>

          <div className="flex items-center gap-6 font-mono text-[11px]">
            <Link href="#capabilities" className="hover:text-textMain transition-colors">
              Capabilities
            </Link>
            <Link href="#architecture" className="hover:text-textMain transition-colors">
              Architecture
            </Link>
            <Link href="/login" className="hover:text-textMain transition-colors">
              Operator Sign In
            </Link>
            <Link href="/signup" className="hover:text-textMain transition-colors">
              Register
            </Link>
          </div>

          <div className="text-[11px] font-mono">
            &copy; {new Date().getFullYear()} MonitorPro Systems.
          </div>
        </div>
      </footer>
    </div>
  );
}
