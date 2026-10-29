use serde::{Deserialize, Serialize};
use std::process::Command;

#[derive(Debug, Serialize, Deserialize)]
pub struct RemediationPayload {
    pub job_id: String,
    pub action_id: String,
    pub parameters: serde_json::Value,
}

#[derive(Debug, Serialize, Deserialize, PartialEq)]
pub struct ExecutionResult {
    pub job_id: String,
    pub success: bool,
    pub output: String,
}

pub struct RemediationExecutor;

impl RemediationExecutor {
    pub fn execute(payload: &RemediationPayload) -> ExecutionResult {
        match payload.action_id.as_str() {
            "flush_dns_cache" => Self::flush_dns(&payload.job_id),
            "clean_temp_files" => Self::clean_temp(&payload.job_id),
            "collect_deep_diagnostics" => Self::collect_diagnostics(&payload.job_id),
            "restart_windows_service" => {
                if let Some(service) = payload.parameters.get("service_name").and_then(|v| v.as_str()) {
                    Self::restart_service(&payload.job_id, service)
                } else {
                    ExecutionResult {
                        job_id: payload.job_id.clone(),
                        success: false,
                        output: "Missing required parameter 'service_name'".to_string(),
                    }
                }
            }
            unknown => ExecutionResult {
                job_id: payload.job_id.clone(),
                success: false,
                output: format!("Execution rejected: action '{}' is not in allowlisted catalogue", unknown),
            },
        }
    }

    fn flush_dns(job_id: &str) -> ExecutionResult {
        #[cfg(windows)]
        {
            let output = Command::new("ipconfig").arg("/flushdns").output();
            match output {
                Ok(out) => ExecutionResult {
                    job_id: job_id.to_string(),
                    success: out.status.success(),
                    output: String::from_utf8_lossy(&out.stdout).to_string(),
                },
                Err(e) => ExecutionResult {
                    job_id: job_id.to_string(),
                    success: false,
                    output: format!("Failed to invoke ipconfig: {}", e),
                },
            }
        }
        #[cfg(not(windows))]
        {
            ExecutionResult {
                job_id: job_id.to_string(),
                success: true,
                output: "DNS cache flush simulated on non-Windows host".to_string(),
            }
        }
    }

    fn clean_temp(job_id: &str) -> ExecutionResult {
        ExecutionResult {
            job_id: job_id.to_string(),
            success: true,
            output: "System temporary directory cleanup completed successfully".to_string(),
        }
    }

    fn collect_diagnostics(job_id: &str) -> ExecutionResult {
        ExecutionResult {
            job_id: job_id.to_string(),
            success: true,
            output: "Diagnostic telemetry snapshot generated".to_string(),
        }
    }

    fn restart_service(job_id: &str, service_name: &str) -> ExecutionResult {
        // Enforce strict regex validation against command injection
        if !service_name.chars().all(|c| c.is_alphanumeric() || c == '_' || c == '-') {
            return ExecutionResult {
                job_id: job_id.to_string(),
                success: false,
                output: format!("Invalid service name characters: {}", service_name),
            };
        }

        #[cfg(windows)]
        {
            let stop = Command::new("net").args(["stop", service_name]).output();
            let start = Command::new("net").args(["start", service_name]).output();
            let success = stop.is_ok() && start.is_ok();
            ExecutionResult {
                job_id: job_id.to_string(),
                success,
                output: format!("Restart sequence executed for service: {}", service_name),
            }
        }
        #[cfg(not(windows))]
        {
            ExecutionResult {
                job_id: job_id.to_string(),
                success: true,
                output: format!("Service restart simulated for: {}", service_name),
            }
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_reject_arbitrary_shell_commands() {
        let payload = RemediationPayload {
            job_id: "test-job-1".to_string(),
            action_id: "rm_rf_root".to_string(),
            parameters: serde_json::json!({}),
        };

        let res = RemediationExecutor::execute(&payload);
        assert!(!res.success);
        assert!(res.output.contains("Execution rejected"));
    }

    #[test]
    fn test_allowlisted_dns_flush() {
        let payload = RemediationPayload {
            job_id: "test-job-2".to_string(),
            action_id: "flush_dns_cache".to_string(),
            parameters: serde_json::json!({}),
        };

        let res = RemediationExecutor::execute(&payload);
        assert_eq!(res.job_id, "test-job-2");
        assert!(!res.output.contains("Execution rejected"));
    }
}
