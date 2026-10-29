package remediation

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"sync"

	"monitopro/services/api/internal/domain"
)

// Catalogue wraps the loaded allowlisted remediation actions
type Catalogue struct {
	mu      sync.RWMutex
	actions map[string]domain.RemediationAction
}

// RawCatalogueFormat matches shared/contracts/remediation_catalogue.json
type RawCatalogueFormat struct {
	Title   string                     `json:"title"`
	Version string                     `json:"version"`
	Actions []domain.RemediationAction `json:"actions"`
}

var serviceNameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

// DefaultCatalogueJSON embedded as fallback
const DefaultCatalogueJSON = `{
  "title": "Hardened Remediation Action Catalogue",
  "version": "2.0.0",
  "actions": [
    {
      "id": "collect_deep_diagnostics",
      "title": "Trigger Deep Hardware & Event Log Diagnostics",
      "risk_level": "read_only",
      "requires_approval": false,
      "timeout_seconds": 60,
      "parameter_schema": {
        "type": "object",
        "properties": {
          "log_window_minutes": { "type": "integer", "default": 60, "maximum": 1440 }
        }
      }
    },
    {
      "id": "restart_windows_service",
      "title": "Restart Named Windows Service",
      "risk_level": "low_risk",
      "requires_approval": true,
      "timeout_seconds": 30,
      "parameter_schema": {
        "type": "object",
        "required": ["service_name"],
        "properties": {
          "service_name": {
            "type": "string",
            "pattern": "^[a-zA-Z0-9_-]+$",
            "description": "Exact service identifier (e.g. Spooler, W32Time)"
          }
        }
      }
    },
    {
      "id": "clean_temp_files",
      "title": "Clean System Temp Files",
      "risk_level": "low_risk",
      "requires_approval": true,
      "timeout_seconds": 120,
      "parameter_schema": {
        "type": "object",
        "properties": {
          "older_than_days": { "type": "integer", "default": 7, "minimum": 1 }
        }
      }
    },
    {
      "id": "flush_dns_cache",
      "title": "Flush DNS Resolver Cache",
      "risk_level": "read_only",
      "requires_approval": false,
      "timeout_seconds": 15,
      "parameter_schema": { "type": "object" }
    },
    {
      "id": "restart_machine",
      "title": "Graceful System Reboot",
      "risk_level": "destructive",
      "requires_approval": true,
      "timeout_seconds": 60,
      "parameter_schema": {
        "type": "object",
        "required": ["grace_period_seconds", "user_notification_message"],
        "properties": {
          "grace_period_seconds": { "type": "integer", "default": 60, "minimum": 10, "maximum": 600 },
          "user_notification_message": { "type": "string", "maxLength": 200 }
        }
      }
    }
  ]
}`

// NewCatalogue returns catalogue initialized with default embedded rules
func NewCatalogue() *Catalogue {
	cat, _ := LoadCatalogue("")
	return cat
}

// LoadCatalogue loads and initializes the actions catalogue from file path or fallback
func LoadCatalogue(path string) (*Catalogue, error) {
	data := []byte(DefaultCatalogueJSON)
	if path != "" {
		fileBytes, err := os.ReadFile(path)
		if err == nil && len(fileBytes) > 0 {
			data = fileBytes
		}
	}

	var parsed RawCatalogueFormat
	if err := json.Unmarshal(data, &parsed); err != nil {
		return nil, fmt.Errorf("failed to parse remediation catalogue: %w", err)
	}

	cat := &Catalogue{
		actions: make(map[string]domain.RemediationAction),
	}

	for _, a := range parsed.Actions {
		cat.actions[a.ID] = a
	}

	return cat, nil
}

// GetAction returns an action by ID
func (c *Catalogue) GetAction(id string) (domain.RemediationAction, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	action, found := c.actions[id]
	return action, found
}

// ListActions returns all allowlisted actions
func (c *Catalogue) ListActions() []domain.RemediationAction {
	c.mu.RLock()
	defer c.mu.RUnlock()
	list := make([]domain.RemediationAction, 0, len(c.actions))
	for _, a := range c.actions {
		list = append(list, a)
	}
	return list
}

// ValidateParameters strictly checks parameters against the action's schema
func (c *Catalogue) ValidateParameters(actionID string, rawParams json.RawMessage) (map[string]any, error) {
	action, found := c.GetAction(actionID)
	if !found {
		return nil, fmt.Errorf("action %q is not in the allowlisted remediation catalogue", actionID)
	}

	var params map[string]any
	if len(rawParams) > 0 && string(rawParams) != "null" {
		if err := json.Unmarshal(rawParams, &params); err != nil {
			return nil, fmt.Errorf("invalid json parameters: %w", err)
		}
	}
	if params == nil {
		params = make(map[string]any)
	}

	switch action.ID {
	case "collect_deep_diagnostics":
		if val, ok := params["log_window_minutes"]; ok {
			num, ok := val.(float64)
			if !ok || num < 1 || num > 1440 {
				return nil, fmt.Errorf("log_window_minutes must be an integer between 1 and 1440")
			}
		} else {
			params["log_window_minutes"] = 60
		}

	case "restart_windows_service":
		val, ok := params["service_name"]
		if !ok {
			return nil, fmt.Errorf("service_name is required")
		}
		serviceName, ok := val.(string)
		if !ok || serviceName == "" {
			return nil, fmt.Errorf("service_name must be a non-empty string")
		}
		if !serviceNameRegex.MatchString(serviceName) {
			return nil, fmt.Errorf("service_name %q contains disallowed characters; must match ^[a-zA-Z0-9_-]+$", serviceName)
		}

	case "clean_temp_files":
		if val, ok := params["older_than_days"]; ok {
			num, ok := val.(float64)
			if !ok || num < 1 {
				return nil, fmt.Errorf("older_than_days must be an integer >= 1")
			}
		} else {
			params["older_than_days"] = 7
		}

	case "flush_dns_cache":
		// No mandatory parameters

	case "restart_machine":
		val, ok := params["grace_period_seconds"]
		if !ok {
			return nil, fmt.Errorf("grace_period_seconds is required")
		}
		gracePeriod, ok := val.(float64)
		if !ok || gracePeriod < 10 || gracePeriod > 600 {
			return nil, fmt.Errorf("grace_period_seconds must be between 10 and 600 seconds")
		}

		valMsg, ok := params["user_notification_message"]
		if !ok {
			return nil, fmt.Errorf("user_notification_message is required")
		}
		msg, ok := valMsg.(string)
		if !ok || len(msg) > 200 {
			return nil, fmt.Errorf("user_notification_message must be a string up to 200 characters")
		}

	default:
		return nil, fmt.Errorf("unrecognized action: %s", actionID)
	}

	return params, nil
}
