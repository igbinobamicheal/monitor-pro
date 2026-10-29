#[derive(Debug, Clone, PartialEq)]
pub enum UserConsentDecision {
    Approved,
    Rejected,
    TimedOut,
}

pub struct EndpointUI;

impl EndpointUI {
    pub const MANDATORY_CONSENT_TEXT: &'static str = 
        "If you approve this request, your screen will be shared with the authorized IT operator, and remote screen control will be activated. The operator will be able to view your screen and, while control is enabled, interact with your computer using the mouse and keyboard. You can reject the request or end the session at any time.";

    #[cfg(windows)]
    pub fn prompt_user_consent(operator_name: &str, reason: &str) -> UserConsentDecision {
        use windows_sys::Win32::UI::WindowsAndMessaging::{
            MessageBoxW, MB_ICONWARNING, MB_YESNO, IDYES,
        };

        let title: Vec<u16> = format!("Remote Support Request from {}\0", operator_name).encode_utf16().collect();
        let message: Vec<u16> = format!(
            "An authorized IT operator ({}) is requesting remote screen sharing and control.\nReason: {}\n\n{}\n\nDo you authorize this session?",
            operator_name, reason, Self::MANDATORY_CONSENT_TEXT
        ).encode_utf16().collect();

        let result = unsafe {
            MessageBoxW(
                0,
                message.as_ptr(),
                title.as_ptr(),
                MB_YESNO | MB_ICONWARNING,
            )
        };

        if result == IDYES {
            UserConsentDecision::Approved
        } else {
            UserConsentDecision::Rejected
        }
    }

    #[cfg(not(windows))]
    pub fn prompt_user_consent(_operator_name: &str, _reason: &str) -> UserConsentDecision {
        UserConsentDecision::Approved
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_mandatory_consent_wording() {
        assert!(EndpointUI::MANDATORY_CONSENT_TEXT.contains("If you approve this request"));
        assert!(EndpointUI::MANDATORY_CONSENT_TEXT.contains("remote screen control will be activated"));
        assert!(EndpointUI::MANDATORY_CONSENT_TEXT.contains("end the session at any time"));
    }
}
