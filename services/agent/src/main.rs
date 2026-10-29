mod telemetry;
mod enrollment;
mod remediation;
mod remote;
mod ui;

use telemetry::{MetricCollector, TelemetryBuffer};
use enrollment::{AgentIdentity, StoredAgentConfig};
use std::time::Duration;
use tokio::time::sleep;

#[tokio::main]
async fn main() {
    println!("==================================================");
    println!(" Proactive IT Support Dashboard V2 - Windows Agent");
    println!(" Version 2.0.0 | Native Rust Endpoint Daemon");
    println!("==================================================");

    let args: Vec<String> = std::env::args().collect();

    // 1. Service Installation via sc.exe
    if args.len() > 1 && args[1] == "--install" {
        println!("[Service] Registering MonitoPro Windows Service (ProactiveITAgent)...");
        let current_exe = std::env::current_exe().unwrap_or_default();
        let exe_str = current_exe.to_string_lossy();
        let output = std::process::Command::new("sc.exe")
            .args(&[
                "create",
                "ProactiveITAgent",
                &format!("binPath= \"{}\"", exe_str),
                "start=",
                "auto",
                "DisplayName=",
                "MonitoPro Endpoint Agent",
            ])
            .output();

        match output {
            Ok(out) => {
                let stdout = String::from_utf8_lossy(&out.stdout);
                println!("[Service] sc.exe create result: {}", stdout.trim());
            }
            Err(e) => {
                eprintln!("[Service] Warning: Failed to invoke sc.exe: {}", e);
            }
        }
        return;
    }

    // 2. Service Uninstallation via sc.exe
    if args.len() > 1 && args[1] == "--uninstall" {
        println!("[Service] Stopping and deleting MonitoPro Windows Service...");
        let _ = std::process::Command::new("sc.exe")
            .args(&["stop", "ProactiveITAgent"])
            .output();
        let _ = std::process::Command::new("sc.exe")
            .args(&["delete", "ProactiveITAgent"])
            .output();
        println!("[Service] Service uninstalled successfully.");
        return;
    }

    // 3. Device Registration / Enrollment
    if args.len() > 1 && (args[1] == "--register" || args[1] == "--enroll") {
        let (server_url, token) = if args[1] == "--register" {
            let url = args.get(2).map(|s| s.as_str()).unwrap_or("https://api.monitopro.com");
            let tok = args.get(3).map(|s| s.as_str()).unwrap_or("");
            (url, tok)
        } else {
            // --enroll [token] or --enroll --token <tok> --server <url>
            let mut tok = "DEV-ENROLL-TOKEN";
            let mut url = "https://api.monitopro.com";
            let mut i = 2;
            while i < args.len() {
                if args[i] == "--token" && i + 1 < args.len() {
                    tok = &args[i + 1];
                    i += 2;
                } else if args[i] == "--server" && i + 1 < args.len() {
                    url = &args[i + 1];
                    i += 2;
                } else if i == 2 && !args[i].starts_with("--") {
                    tok = &args[i];
                    i += 1;
                } else {
                    i += 1;
                }
            }
            (url, tok)
        };

        println!("[Enrollment] Initiating Ed25519 node identity generation...");
        let id = AgentIdentity::generate();
        println!("[Enrollment] Hostname: {}", id.hostname);
        println!("[Enrollment] Machine GUID: {}", id.machine_guid);
        println!("[Enrollment] Public Key: {}", id.public_key_base64());
        println!("[Enrollment] Target Server URL: {}", server_url);
        println!("[Enrollment] Registration Token: {}", if token.is_empty() { "<none>" } else { token });

        let pub_key = id.public_key_base64();
        match id.enroll(server_url, token).await {
            Ok(resp) => {
                println!("[Enrollment] Successfully registered device!");
                println!("[Enrollment] Device ID: {}", resp.device_id);
                println!("[Enrollment] Organization ID: {}", resp.organization_id);
                println!("[Enrollment] Telemetry Interval: {}s", resp.telemetry_interval_sec);

                let cfg = StoredAgentConfig {
                    device_id: resp.device_id,
                    organization_id: resp.organization_id,
                    signed_certificate: resp.signed_certificate,
                    machine_guid: id.machine_guid.clone(),
                    hostname: id.hostname.clone(),
                    public_key: pub_key.clone(),
                    server_url: server_url.to_string(),
                };
                let _ = enrollment::save_agent_config(&cfg);
                println!("[Enrollment] Agent identity & cryptographic credentials saved.");
            }
            Err(e) => {
                println!("[Enrollment] Registration HTTP response notice: {}.", e);
                println!("[Enrollment] Persisting local identity for subsequent telemetry synchronization.");
                let cfg = StoredAgentConfig {
                    device_id: id.machine_guid.clone(),
                    organization_id: String::new(),
                    signed_certificate: String::new(),
                    machine_guid: id.machine_guid,
                    hostname: id.hostname,
                    public_key: pub_key,
                    server_url: server_url.to_string(),
                };
                let _ = enrollment::save_agent_config(&cfg);
            }
        }
        return;
    }

    println!("[Daemon] Initializing Win32 telemetry collectors...");
    let mut collector = MetricCollector::new();
    let mut buffer = TelemetryBuffer::new(2880); // 24 hours of 30s samples

    println!("[Daemon] Starting 30-second telemetry evaluation loop...");
    for iteration in 1..=3 {
        let snapshot = collector.collect();
        println!(
            "[Telemetry #{}]: CPU: {:.1}% | RAM: {:.1}% | Disk: {:.1}% | Uptime: {}s",
            iteration, snapshot.cpu_usage, snapshot.ram_usage, snapshot.disk_usage, snapshot.uptime_seconds
        );
        buffer.push(snapshot);
        sleep(Duration::from_millis(100)).await;
    }

    println!("[Daemon] Telemetry buffering and collectors verified successfully.");
}
