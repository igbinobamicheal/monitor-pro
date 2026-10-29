use ed25519_dalek::SigningKey;
use rand::rngs::OsRng;
use serde::{Deserialize, Serialize};

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct EnrollmentRequest {
    pub enrollment_token: String,
    pub machine_guid: String,
    pub hostname: String,
    pub os_version: String,
    pub public_key: String,
}

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct EnrollmentResponse {
    pub device_id: String,
    pub organization_id: String,
    pub signed_certificate: String,
    pub telemetry_interval_sec: u64,
}

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct StoredAgentConfig {
    pub device_id: String,
    pub organization_id: String,
    pub signed_certificate: String,
    pub machine_guid: String,
    pub hostname: String,
    pub public_key: String,
    pub server_url: String,
}

pub struct AgentIdentity {
    pub signing_key: SigningKey,
    pub machine_guid: String,
    pub hostname: String,
}

impl AgentIdentity {
    pub fn generate() -> Self {
        let mut csprng = OsRng;
        let signing_key = SigningKey::generate(&mut csprng);
        
        let hostname = std::env::var("COMPUTERNAME")
            .or_else(|_| std::env::var("HOSTNAME"))
            .unwrap_or_else(|_| "Unknown-Node".to_string());

        let machine_guid = uuid::Uuid::new_v4().to_string();

        Self {
            signing_key,
            machine_guid,
            hostname,
        }
    }

    pub fn public_key_base64(&self) -> String {
        let verifying_key = self.signing_key.verifying_key();
        let bytes = verifying_key.to_bytes();
        bytes.iter().map(|b| format!("{:02x}", b)).collect()
    }

    pub fn build_enrollment_request(&self, token: &str) -> EnrollmentRequest {
        EnrollmentRequest {
            enrollment_token: token.to_string(),
            machine_guid: self.machine_guid.clone(),
            hostname: self.hostname.clone(),
            os_version: "Windows 11 Pro 64-bit".to_string(),
            public_key: self.public_key_base64(),
        }
    }

    pub async fn enroll(&self, server_url: &str, token: &str) -> Result<EnrollmentResponse, Box<dyn std::error::Error + Send + Sync>> {
        let req_body = self.build_enrollment_request(token);
        let trimmed_url = server_url.trim_end_matches('/');
        let endpoint = format!("{}/api/v1/agent/enroll", trimmed_url);

        let client = reqwest::Client::builder()
            .timeout(std::time::Duration::from_secs(10))
            .build()?;

        let resp = client.post(&endpoint)
            .json(&req_body)
            .send()
            .await?;

        if !resp.status().is_success() {
            let status = resp.status();
            let err_text = resp.text().await.unwrap_or_default();
            return Err(format!("Enrollment failed HTTP {}: {}", status, err_text).into());
        }

        let enrollment_resp: EnrollmentResponse = resp.json().await?;
        Ok(enrollment_resp)
    }
}

pub fn save_agent_config(cfg: &StoredAgentConfig) -> std::io::Result<()> {
    let json = serde_json::to_string_pretty(cfg).map_err(|e| std::io::Error::new(std::io::ErrorKind::Other, e))?;
    if let Ok(program_data) = std::env::var("ProgramData") {
        let dir = std::path::Path::new(&program_data).join("MonitoPro");
        let _ = std::fs::create_dir_all(&dir);
        let p = dir.join("agent-config.json");
        if std::fs::write(&p, &json).is_ok() {
            return Ok(());
        }
    }
    std::fs::write("agent-config.json", json)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_identity_generation() {
        let id = AgentIdentity::generate();
        assert!(!id.hostname.is_empty());
        assert!(!id.public_key_base64().is_empty());
        let req = id.build_enrollment_request("token-123");
        assert_eq!(req.enrollment_token, "token-123");
    }

    #[test]
    fn test_agent_config_serialization() {
        let cfg = StoredAgentConfig {
            device_id: "test-dev-id".to_string(),
            organization_id: "test-org-id".to_string(),
            signed_certificate: "cert-abc".to_string(),
            machine_guid: "guid-123".to_string(),
            hostname: "test-host".to_string(),
            public_key: "pubkey-hex".to_string(),
            server_url: "https://example.monitopro.com".to_string(),
        };
        let serialized = serde_json::to_string(&cfg).expect("serialize");
        let deserialized: StoredAgentConfig = serde_json::from_str(&serialized).expect("deserialize");
        assert_eq!(deserialized.device_id, "test-dev-id");
        assert_eq!(deserialized.server_url, "https://example.monitopro.com");
    }
}
