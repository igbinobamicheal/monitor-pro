use serde::{Deserialize, Serialize};

#[derive(Debug, Serialize, Deserialize)]
pub enum RemoteControlInput {
    MouseMove { x: i32, y: i32 },
    MouseDown { button: u32 },
    MouseUp { button: u32 },
    KeyPress { key_code: u16 },
    KeyRelease { key_code: u16 },
}

pub struct RemoteSupportEngine {
    session_active: bool,
    control_granted: bool,
}

impl RemoteSupportEngine {
    pub fn new() -> Self {
        Self {
            session_active: false,
            control_granted: false,
        }
    }

    pub fn start_session(&mut self, control_permission: bool) {
        self.session_active = true;
        self.control_granted = control_permission;
    }

    pub fn stop_session(&mut self) {
        self.session_active = false;
        self.control_granted = false;
    }

    pub fn is_active(&self) -> bool {
        self.session_active
    }

    pub fn can_control(&self) -> bool {
        self.session_active && self.control_granted
    }

    pub fn handle_input(&self, input: RemoteControlInput) -> bool {
        if !self.can_control() {
            // Strictly reject input if remote control permission was not granted
            return false;
        }

        #[cfg(windows)]
        {
            use windows_sys::Win32::UI::Input::KeyboardAndMouse::{
                SendInput, INPUT, INPUT_MOUSE, MOUSEINPUT, MOUSEEVENTF_MOVE, MOUSEEVENTF_ABSOLUTE,
            };

            match input {
                RemoteControlInput::MouseMove { x, y } => {
                    let mut input_record = INPUT {
                        r#type: INPUT_MOUSE,
                        Anonymous: windows_sys::Win32::UI::Input::KeyboardAndMouse::INPUT_0 {
                            mi: MOUSEINPUT {
                                dx: x,
                                dy: y,
                                mouseData: 0,
                                dwFlags: MOUSEEVENTF_MOVE | MOUSEEVENTF_ABSOLUTE,
                                time: 0,
                                dwExtraInfo: 0,
                            },
                        },
                    };
                    unsafe {
                        SendInput(1, &mut input_record, std::mem::size_of::<INPUT>() as i32);
                    }
                    true
                }
                _ => true,
            }
        }
        #[cfg(not(windows))]
        {
            let _ = input;
            true
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_input_rejected_without_control_permission() {
        let mut engine = RemoteSupportEngine::new();
        engine.start_session(false); // Viewing only, no control
        assert!(!engine.handle_input(RemoteControlInput::MouseMove { x: 100, y: 100 }));
    }

    #[test]
    fn test_input_accepted_when_control_granted() {
        let mut engine = RemoteSupportEngine::new();
        engine.start_session(true); // Viewing & Control granted
        assert!(engine.handle_input(RemoteControlInput::MouseMove { x: 100, y: 100 }));
    }
}
