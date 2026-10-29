"use client";

import Link from "next/link";
import { ArrowRight } from "lucide-react";

export function PublicNavbar() {
  return (
    <header className="fixed top-0 left-0 right-0 z-50 bg-[#ECEAE5]/75 backdrop-blur-xl border-b border-white/60 shadow-[0_4px_24px_rgba(30,26,23,0.02)] transition-all">
      <div className="max-w-7xl mx-auto px-6 h-16 flex items-center justify-between">
        {/* Brand Logo */}
        <Link href="/" className="flex items-center gap-2.5 group">
          <div className="w-8 h-8 rounded-lg bg-gradient-to-br from-primary/15 to-primary/5 border border-primary/25 flex items-center justify-center transition-transform group-hover:scale-105">
            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
              <path d="M12 2L2 7L12 12L22 7L12 2Z" stroke="#B87C3B" strokeWidth="2.2" strokeLinejoin="round" />
              <path d="M2 17L12 22L22 17" stroke="#B87C3B" strokeWidth="2.2" strokeLinejoin="round" />
              <path d="M2 12L12 17L22 12" stroke="#B87C3B" strokeWidth="2.2" strokeLinejoin="round" />
            </svg>
          </div>
          <span className="font-semibold tracking-tight text-[16px] text-textMain">
            Monitor<span className="text-primary font-bold">Pro</span>
          </span>
          <span className="text-[10px] font-mono font-medium uppercase px-2 py-0.5 rounded-full bg-primary/10 border border-primary/20 text-primary ml-1">
            V2
          </span>
        </Link>

        {/* Minimalist Navigation */}
        <nav className="hidden md:flex items-center gap-8 text-[13px] font-medium text-textMuted">
          <Link href="#capabilities" className="hover:text-textMain transition-colors">
            Capabilities
          </Link>
          <Link href="#architecture" className="hover:text-textMain transition-colors">
            Architecture
          </Link>
          <Link href="#enrollment" className="hover:text-textMain transition-colors">
            Quick Start
          </Link>
        </nav>

        {/* Action Controls */}
        <div className="flex items-center gap-3">
          <Link
            href="/login"
            className="text-[13px] font-medium text-textMuted hover:text-textMain px-3 py-1.5 transition-colors"
          >
            Operator Sign In
          </Link>
          <Link
            href="/signup"
            className="flex items-center gap-1.5 text-[13px] font-medium bg-primary hover:bg-primaryHover text-white px-3.5 py-1.5 rounded-lg transition-all shadow-[0_2px_12px_rgba(184,124,59,0.25)] hover:shadow-[0_4px_16px_rgba(184,124,59,0.35)]"
          >
            Launch Free Workspace <ArrowRight className="w-3.5 h-3.5" />
          </Link>
        </div>
      </div>
    </header>
  );
}
