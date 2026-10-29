# Native Remote Support & Session Protocol Specification

**Subsystem:** Native Screen Sharing & Remote Input Control  
**Implementation:** Rust (Windows Desktop Duplication + Win32 Input) & WebRTC / WSS  
**Requirement:** Explicit Endpoint User Consent, Persistent Banner, Mutual Revocation  

---

## 1. Session Negotiation Lifecycle

```mermaid
sequenceDiagram
    autonumber
    actor Operator as IT Operator (Browser)
    participant API as Go Control Plane
    participant Agent as Rust Agent (Windows)
    actor User as Endpoint User

    Operator->>API: POST /api/v1/remote-support/sessions (device_id, perms, reason)
    API->>Agent: SessionRequest (session_id, operator_name, perms, reason)
    Agent->>User: Display Native Win32 Consent Window
    Note over User,Agent: Mandatory consent text displayed with Approve / Reject buttons
    alt User Rejects or 60s Timeout
        User->>Agent: Click "Reject"
        Agent->>API: ConsentDecision (rejected)
        API->>Operator: Session Failed: Consent Denied
    else User Approves
        User->>Agent: Click "Approve"
        Agent->>Agent: Display Persistent Non-Closable Banner (Top of Screen)
        Agent->>API: ConsentDecision (approved, granted_permissions)
        API->>Operator: Session Approved (WebRTC / WSS Signaling Credentials)
        Operator<-->Agent: Establish Encrypted Peer Connection / WebSocket Stream
        loop Live Session (Max 60 min)
            Agent->>Operator: Desktop Duplication Video Frames (H.264 / VP8)
            alt "remote_control" Granted
                Operator->>Agent: MouseMove, MouseClick, KeyPress Events
                Agent->>Agent: Verify Permission & Inject via SendInput()
            end
            Agent<-->Operator: Heartbeat (every 2s; 5s loss terminates)
        end
        opt Session Termination
            User->>Agent: Click "End Session" on Persistent Banner
            Agent->>Operator: SessionTerminatedEvent
            Agent->>Agent: Remove Banner, Release Desktop Duplication, Disable Input
            Agent->>API: POST /sessions/{id}/terminate (reason: user_local_termination)
        end
    end
```

---

## 2. Windows Endpoint Native UI Requirements

1. **Consent Window (Win32 API):**
   - Must be created on the active console session (`WTSGetActiveConsoleSessionId`).
   - Must explicitly state:
     > *"If you approve this request, your screen will be shared with the authorized IT operator, and remote screen control will be activated. The operator will be able to view your screen and, while control is enabled, interact with your computer using the mouse and keyboard. You can reject the request or end the session at any time."*
   - Buttons: `Approve` and `Decline`.
2. **Persistent Active Session Indicator:**
   - Always-on-top window (`WS_EX_TOPMOST` | `WS_EX_TOOLWINDOW`) anchored at the top-center of the screen.
   - Distinctive amber/green styling clearly indicating:
     - "Remote Support Active"
     - Operator identity
     - Active permissions: `[Viewing Only]` or `[Viewing & Control]`
     - Large, red **"End Session"** button.
   - If the agent process or connection crashes, the banner automatically closes and all input hooks are immediately released.
