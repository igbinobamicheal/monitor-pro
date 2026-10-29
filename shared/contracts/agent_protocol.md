# Endpoint Agent Communication Protocol Specification

**Protocol Version:** 2.0.0  
**Target:** Windows Rust Service (Linux in Phase 2+)  
**Security:** Asymmetric Cryptography (Ed25519) + TLS 1.3  

---

## 1. Enrollment & Cryptographic Identity

1. **Initial Provisioning:**
   - Operator creates an enrollment token bound to an organization.
   - Endpoint runs: `proactive-agent.exe --enroll --token <TOKEN> --server https://companya.monitopro.com`
2. **Keypair Generation:**
   - Agent generates an Ed25519 keypair locally via the OS crypto provider (`ring` or `ed25519-dalek`).
   - Private key is stored securely in Windows DPAPI / Credential Guard. It is never transmitted.
3. **Registration Request:**
   - Agent submits `machine_guid`, `hostname`, `os_version`, and the `public_key` to `/api/v1/agent/enroll`.
4. **Certificate / Token Issuance:**
   - Backend registers the node in PostgreSQL under the tenant with state `online`.
   - Backend signs an enrollment certificate containing `device_id` and `organization_id`.

---

## 2. Telemetry Reporting (30s Interval)

1. **Collection:**
   - CPU utilization (kernel + user times via Win32 `GetSystemTimes`).
   - RAM usage (via `GlobalMemoryStatusEx`).
   - Disk space and active drive status (via `GetDiskFreeSpaceExW`).
   - Network in/out throughput (via `GetIfTable2`).
   - System uptime (via `GetTickCount64`).
2. **Signing & Transmission:**
   - Payload is serialized to JSON.
   - Agent signs `hash(payload + timestamp + sequence_number)` with its private key.
   - Header `X-Agent-Signature: <BASE64_SIG>` and `X-Device-ID: <UUID>` are sent with HTTP `POST /api/v1/agent/telemetry`.
3. **Buffering & Backoff:**
   - If network is unreachable, metrics are stored in a local SQLite/sled ring buffer (maximum 24 hours of samples).
   - On reconnect, queued batches are flushed with bounded rate and exponential backoff + jitter.
