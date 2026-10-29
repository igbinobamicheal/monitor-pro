package remediation_test

import (
	"encoding/json"
	"testing"

	"monitopro/services/api/internal/remediation"
)

func TestCatalogueLoading(t *testing.T) {
	cat := remediation.NewCatalogue()
	actions := cat.ListActions()
	if len(actions) < 5 {
		t.Fatalf("Expected at least 5 catalogue actions, got %d", len(actions))
	}

	act, found := cat.GetAction("restart_windows_service")
	if !found {
		t.Fatal("Action restart_windows_service not found")
	}
	if !act.RequiresApproval {
		t.Fatal("restart_windows_service should require approval")
	}
	if act.RiskLevel != "low_risk" {
		t.Fatalf("Expected risk low_risk, got %s", act.RiskLevel)
	}
}

func TestParameterValidation(t *testing.T) {
	cat := remediation.NewCatalogue()

	// 1. Valid restart_windows_service
	validServiceParams, _ := json.Marshal(map[string]any{"service_name": "Spooler"})
	params, err := cat.ValidateParameters("restart_windows_service", validServiceParams)
	if err != nil {
		t.Fatalf("Expected valid service_name to pass: %v", err)
	}
	if params["service_name"] != "Spooler" {
		t.Errorf("Expected Spooler, got %v", params["service_name"])
	}

	// 2. Malicious injection in service_name should be rejected
	maliciousParams, _ := json.Marshal(map[string]any{"service_name": "Spooler; rm -rf /"})
	if _, err := cat.ValidateParameters("restart_windows_service", maliciousParams); err == nil {
		t.Fatal("Expected error on invalid characters in service_name, got nil")
	}

	// 3. Missing service_name should be rejected
	emptyParams, _ := json.Marshal(map[string]any{})
	if _, err := cat.ValidateParameters("restart_windows_service", emptyParams); err == nil {
		t.Fatal("Expected error on missing service_name, got nil")
	}

	// 4. Valid restart_machine
	validRebootParams, _ := json.Marshal(map[string]any{
		"grace_period_seconds":      30,
		"user_notification_message": "Maintenance reboot in 30 seconds",
	})
	if _, err := cat.ValidateParameters("restart_machine", validRebootParams); err != nil {
		t.Fatalf("Expected valid reboot params to pass: %v", err)
	}

	// 5. Out of range grace_period_seconds (< 10)
	invalidRebootParams, _ := json.Marshal(map[string]any{
		"grace_period_seconds":      2,
		"user_notification_message": "Reboot now",
	})
	if _, err := cat.ValidateParameters("restart_machine", invalidRebootParams); err == nil {
		t.Fatal("Expected error on grace_period_seconds < 10, got nil")
	}

	// 6. Unknown action rejected
	if _, err := cat.ValidateParameters("unapproved_script_execution", validServiceParams); err == nil {
		t.Fatal("Expected error for non-allowlisted action ID, got nil")
	}
}
