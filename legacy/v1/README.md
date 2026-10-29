# Proactive IT Support Dashboard

GitHub Repository - https://github.com/igbinobamicheal/PROACTIVE-IT-SUPPORT-DASHBOARD

[Vercel Deployment] - https://proactive-it-support-dashboard.vercel.app

> **Repository URL:** [https://github.com/igbinobamicheal/PROACTIVE-IT-SUPPORT-DASHBOARD](https://github.com/igbinobamicheal/PROACTIVE-IT-SUPPORT-DASHBOARD)  
> **Live Web App (Vercel):** [https://proactive-it-support-dashboard.vercel.app](https://proactive-it-support-dashboard.vercel.app)

---

## 📌 Overview

**Proactive IT Support Dashboard** is an enterprise-grade system health and performance monitoring solution. Built for speed, reliability, and security, it combines a high-performance C++20 REST API backend, PostgreSQL relational storage (with full Supabase compatibility), a lightweight C++ monitoring agent, and a modern glassmorphic web UI.

The platform actively tracks machine metrics (CPU utilization, memory usage, disk storage, uptime, system load), triggers automated health alerts, and enables remote agent installation for both Windows and Linux hosts.

---

## ✨ Key Features

- **🚀 C++20 REST API Backend**: Built with Crow, featuring high-throughput async I/O, prepared PostgreSQL statements via `libpqxx`, JWT-based authentication, and BCrypt password hashing.
- **🖥️ C++ System Monitoring Agent**: Cross-platform background service for Windows and Linux that collects host telemetry (CPU, RAM, Disk, Uptime, process statistics) and securely transmits it to the API backend.
- **💎 Glassmorphic Web Interface**: Responsive, high-aesthetic web dashboard for real-time telemetry visualizers, alert management, host device telemetry, and automated deployment script generation.
- **🗄️ PostgreSQL / Supabase Integration**: Robust data retention schema with automated indexing, prepared queries, and connection pooler compatibility.
- **⚙️ Automated Agent Installers**: Dynamically generates single-command PowerShell and Shell scripts for instant host enrollment.
- **☁️ Cloud & Serverless Ready**: Automated Vercel deployments for the static frontend and Docker/Railway ready for the C++ backend.

---

## 📁 Repository Structure

```text
PROACTIVE-IT-SUPPORT-DASHBOARD/
├── backend/            # C++20 Crow REST Web Server source code & vcpkg.json
├── agent/              # C++ system metric collection agent source code
├── database/           # PostgreSQL schema (schema.sql) and seed data (seed.sql)
├── frontend/           # Modern glassmorphic HTML/CSS/JS frontend interface
├── installer/          # Deployment installer templates (Windows & Linux)
├── scripts/            # Local Windows build and launch batch scripts
├── docs/               # Architecture, API specifications, and deployment guides
├── generate-config.js  # Build script for injecting API configuration on Vercel
├── vercel.json         # Vercel deployment, rewrites, and security headers config
└── docker-compose.yml  # Multi-container local production deployment setup
```

---

## 🛠️ Prerequisites

Before building and running the project locally, ensure your system has:

1. **C++20 Compiler**: MSVC 2022+ (Windows), GCC 11+ (Linux), or Clang 13+ (macOS/Linux).
2. **CMake**: Version 3.15 or higher.
3. **vcpkg**: C++ package manager (integrated automatically via `vcpkg.json` manifest mode).
4. **PostgreSQL 15+**: Local PostgreSQL server or a hosted instance (e.g., [Supabase](https://supabase.com)).
5. **Node.js**: Version 14+ (Required for frontend config generation during Vercel builds).

---

## 🚀 Setup & Installation Guide

### 1. Database Setup

1. Connect to your PostgreSQL database instance.
2. Execute the schema script to create the necessary tables, indexes, and constraints:
   ```bash
   psql -U postgres -d postgres -f database/schema.sql
   ```
3. Seed the default database user (Administrator):
   ```bash
   psql -U postgres -d postgres -f database/seed.sql
   ```
   > 🔐 **Default Admin Credentials**:  
   > **Username:** `admin` | **Password:** `admin123` _(Please change after first login)_

---

### 2. Backend API Setup

#### Option A: Local Build (CMake + vcpkg)

1. Navigate to the backend directory:
   ```bash
   cd backend
   ```
2. Create your configuration file from the template:
   - **Windows (Command Prompt / PowerShell)**:
     ```cmd
     copy config\config.json.template config\config.json
     ```
   - **Linux / macOS**:
     ```bash
     cp config/config.json.template config/config.json
     ```
3. Update `backend/config/config.json` with your database credentials and secret keys.
4. Configure and build using CMake presets (dependencies will automatically download via vcpkg manifest mode):
   ```bash
   cmake --preset default
   cmake --build --preset default
   ```
5. Run the compiled backend executable:
   - **Windows**:
     ```cmd
     .\build\Release\proactive_it_dashboard.exe
     ```
   - **Linux**:
     ```bash
     ./build/Release/proactive_it_dashboard
     ```

#### Option B: Automated Windows Batch Scripts

If you are on Windows, you can use the pre-configured scripts in `/scripts`:

```cmd
scripts\build_backend.bat
scripts\start_backend.bat
```

---

### 3. Monitoring Agent Setup

1. Navigate to the agent directory:
   ```bash
   cd agent
   ```
2. Build the agent executable:
   ```bash
   cmake --preset default
   cmake --build --preset default
   ```
3. Launch the agent:
   - **Windows**: `.\build\Release\proactive_it_agent.exe`
   - **Linux**: `./build/Release/proactive_it_agent`
4. On first startup, authorize the agent using your Admin JWT Token or Registration Key to enroll the device into your dashboard.

> 💡 **Automated Deployment**: You can also generate PowerShell (`.ps1`) or Bash (`.sh`) installer scripts directly from the live web UI at `/deploy-agent` to deploy agents across remote machines.

---

### 4. Frontend Setup & Vercel Deployment

#### Local Development

The frontend in `/frontend` consists of modern vanilla HTML5, CSS3, and JavaScript.
To run locally:

1. Serve the `frontend` folder using any static server (e.g., VS Code Live Server, Python `http.server`, or `npx serve frontend`).
2. Open `http://localhost:3000` or `http://localhost:8080` in your browser.
3. By default, local frontend queries `http://localhost:8080/api`.

#### Vercel Deployment

This repository is pre-configured for automated Vercel deployment:

1. Import your GitHub repository ([`igbinobamicheal/PROACTIVE-IT-SUPPORT-DASHBOARD`](https://github.com/igbinobamicheal/PROACTIVE-IT-SUPPORT-DASHBOARD)) into **Vercel**.
2. Keep the **Root Directory** as `/` (workspace root).
3. Set the project settings:
   - **Framework Preset**: `Other`
   - **Build Command**: `node generate-config.js`
   - **Output Directory**: `frontend`
4. Configure the **Environment Variable**:
   | Key | Value | Description |
   |---|---|---|
   | `API_URL` | `https://your-backend-api-url.com/api` | Public endpoint of your live C++ backend |
5. Deploy! Vercel will run `generate-config.js`, inject the `API_URL` into `frontend/js/config.js`, and host the static app.
6. Clean URLs are configured via `vercel.json` (`/dashboard`, `/login`, `/devices`, `/alerts`, `/device-details`).

---

## 🔑 Environment Variables Reference

When deploying the backend (on Railway, Render, Docker, etc.), configure the following environment variables:

| Variable           | Description                            | Default / Example                   |
| ------------------ | -------------------------------------- | ----------------------------------- |
| `DB_HOST`          | PostgreSQL Hostname                    | `localhost` or `db.xxx.supabase.co` |
| `DB_PORT`          | PostgreSQL Port                        | `5432`                              |
| `DB_USER`          | PostgreSQL Username                    | `postgres`                          |
| `DB_PASSWORD`      | PostgreSQL Password                    | `your_db_password`                  |
| `DB_NAME`          | Database Name                          | `postgres`                          |
| `JWT_SECRET`       | Secret key for signing session tokens  | `random_secret_string_32_chars`     |
| `REGISTRATION_KEY` | Shared secret key for agent enrollment | `agent_enrollment_key`              |
| `PORT`             | API Server listening port              | `8080`                              |

---

## 🐳 Docker Deployment

Run the backend API and PostgreSQL database together using Docker Compose:

```bash
# Build and start services in detached mode
docker-compose up -d --build

# View container logs
docker-compose logs -f
```

---

## 📚 Documentation

For more detailed technical specifications and guides, consult the `/docs` directory:

- [📄 Architecture Overview](docs/architecture.md)
- [🔌 REST API Specifications](docs/api.md)
- [☁️ Production Deployment Guide](docs/deployment.md)

---

## 📄 License

Distributed under the MIT License. See `LICENSE` for more information.
