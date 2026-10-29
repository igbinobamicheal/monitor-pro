# Monitor Pro

> **Zero-Trust Windows Endpoint Observability, Deterministic AI Diagnostics & Policy-Governed Remediation**

Monitor Pro is an enterprise-grade, multi-tenant endpoint management platform engineered for proactive IT operations. It unifies low-overhead native Windows telemetry, deterministic AI root-cause hypothesis ranking with PII redaction, allowlisted one-click remediation, and unattended WebRTC remote support.

---

## Architecture Overview

```
                                      ┌─────────────────────────────────────────┐
                                      │             Edge Gateway                │
                                      │         Caddy Reverse Proxy             │
                                      │         (:80 / :443 TLS Term)           │
                                      └────────────────────┬────────────────────┘
                                                           │
                                   ┌───────────────────────┴───────────────────────┐
                                   ▼                                               ▼
                     ┌───────────────────────────┐                   ┌───────────────────────────┐
                     │     Next.js 15+ Web       │                   │    Go API Control Plane   │
                     │  Warm Canvas / Obsidian   │◄─────────────────►│   Multi-Replica (Chi)     │
                     │    Port :3000 (React)     │                   │   Ed25519 & JWT Auth      │
                     └───────────────────────────┘                   └─────────────┬─────────────┘
                                                                                   │
                                           ┌───────────────────────────────────────┴──────────────────┐
                                           ▼                                                          ▼
                             ┌───────────────────────────┐                              ┌───────────────────────────┐
                             │    PostgreSQL 15+ RLS     │                              │       NATS JetStream      │
                             │  Authoritative Storage    │                              │     Durable Telemetry     │
                             │  Row-Level Tenant Guard   │                              │     Port :4222 / :8222    │
                             └───────────────────────────┘                              └───────────────────────────┘
                                           ▲
                                           │ Signed Hardware Telemetry & Heartbeat
                                           │ (30-second cycles, Ed25519 Mutual Auth)
                                           │
                             ┌───────────────────────────┐
                             │  Native Rust Agent Daemon │
                             │  proactive-it-agent.exe   │
                             │  Win32 Hardware Sensor    │
                             └───────────────────────────┘
```

---

## Key Features

1. **Native Rust Endpoint Agent (`services/agent`)**
   - Standalone Windows binary (under 4 MB footprint, zero runtime dependencies).
   - High-performance Win32 hardware sensor (CPU, memory paging, drive volumes, network interfaces).
   - Ed25519 cryptographic identity provisioning and telemetry packet signing.
   - Built-in CLI flags: `--register <URL> <TOKEN>`, `--install`, `--uninstall`, `--run`.

2. **Go Control Plane API (`services/api`)**
   - Lightweight, concurrent Chi HTTP router with sub-millisecond route dispatch.
   - Multi-tenant architecture with mandatory Row-Level Security (RLS) enforcement.
   - RFC 6238 TOTP Multi-Factor Authentication (MFA) and tenant subdomain resolver.
   - Automated PII scrubbing regex pipeline safeguarding telemetry before AI inference.
   - Allowlisted remediation catalogue preventing arbitrary shell code execution.

3. **Next.js 15+ Operations Dashboard (`apps/web`)**
   - Tailored aesthetic with Warm Canvas (`#ECEAE5`), Obsidian Sidebar (`#151310`), and Rich Amber accents (`#B87C3B`).
   - 17 fully integrated routes:
     - Minimalist Marketing Landing Page with ambient lighting canvas.
     - Multi-step Organization Onboarding & TOTP MFA Setup.
     - Fleet Infrastructure Management & Deep Diagnostics Inspection.
     - 3-Step Guided Windows Agent Deployment Wizard.
     - Active Alerts Center & Dynamic Rule Policy Manager.
     - AI-Assisted Root Cause Incident Investigation.
     - Low-Latency Remote Support with User Consent Enforcement.
     - Policy-Governed Remediation Dispatch & Execution History.
     - Team RBAC & Role Management.
     - Workspace Settings & Zstandard Data Retention Controls.

4. **Inno Setup Windows Packaging (`installer/windows`)**
   - Single-file installer (`MonitorPro-Setup.exe`) bundling the native Rust agent.
   - Headless deployment support via `/VERYSILENT /SERVERURL=... /REGTOKEN=...` for Active Directory GPO, Intune, and Datto RMM mass rollouts.

---

## Directory Structure

```text
├── apps/
│   └── web/                   # Next.js 15+ App Router Frontend
│       ├── app/               # 17 Pages & Route Groups (auth, dashboard, marketing)
│       ├── components/        # Obsidian Sidebar, Ambient Background, Navbar
│       ├── lib/               # Typed API Client & Contract Interfaces
│       └── tailwind.config.ts # Precision V1 Color Palette & Tokens
├── services/
│   ├── agent/                 # Native Rust Windows Endpoint Agent
│   │   └── src/               # Telemetry Collector, Ed25519 Enrollment, Remote Protocol
│   ├── api/                   # Go 1.24+ Control Plane API
│   │   ├── cmd/server/        # Entrypoint & Signal Handlers
│   │   └── internal/          # Chi Handlers, Storage Adapters, AI Redactor, TOTP
│   └── workers/               # Go Background Event Processing & Archival
├── shared/
│   ├── contracts/             # OpenAPI 3.1, NATS Events, Remediation Catalogue
│   └── migrations/            # PostgreSQL Schema & Row-Level Security (RLS)
├── installer/
│   └── windows/               # Inno Setup Script (GUI & Silent CLI Deployment)
└── deploy/
    ├── docker-compose.yml     # Complete Multi-Container Stack (Caddy, 2x API, Postgres, NATS)
    └── caddy/Caddyfile        # Edge Reverse Proxy Configuration
```

---

## Getting Started

### Prerequisites

- **Go:** 1.22+
- **Node.js:** 18+ (Node 20+ recommended)
- **Rust:** 1.80+ (MSVC toolchain for Windows agent)
- **PostgreSQL:** 15+
- **Inno Setup:** 6.x (for building `.exe` installer)

---

### Running Locally

#### 1. Go Control Plane API

```bash
cd services/api
cp .env.example .env
go run ./cmd/server
# Listening on http://localhost:8080
```

#### 2. Next.js Web Dashboard

```bash
cd apps/web
npm install
cp .env.example .env.local
npm run dev
# Live at http://localhost:3000
```

#### 3. Native Rust Agent

```bash
cd services/agent
cargo build --release

# Register against local or remote control plane
.\target\release\proactive-it-agent.exe --register "http://localhost:8080" "<ENROLLMENT_TOKEN>"
```

---

### Running via Docker Compose

```bash
cd deploy
docker compose up -d --build
```

Services started:
- **Caddy Edge Proxy:** `http://localhost:80`
- **Go API Replicas:** `api-1` and `api-2`
- **PostgreSQL 15:** Port 5432 with auto-applied migrations
- **NATS JetStream:** Port 4222 (Client) & Port 8222 (HTTP Dashboard)

---

## Security & Zero-Trust Architecture

- **No Plaintext Agent Secrets:** Nodes authenticate exclusively via Ed25519 asymmetric signatures.
- **Strict Remediation Allowlist:** Only pre-audited diagnostic actions defined in `shared/contracts/remediation_catalogue.json` can be executed. Arbitrary shell execution is rejected by the agent parser.
- **Auditable Air-Gapped Consent:** Remote screen sessions cannot start without explicit user acceptance on the physical Windows desktop.

---

## License

Proprietary — All rights reserved.
