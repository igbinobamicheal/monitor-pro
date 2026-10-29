package ai

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"monitopro/services/api/internal/domain"
)

func TestRedactPII(t *testing.T) {
	input := "Crash observed on server at 192.168.1.100 using Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.xyz"
	redacted := RedactPII(input)

	if redacted == input {
		t.Error("Expected text to be redacted")
	}
	if !containsString(redacted, "[REDACTED_IP]") {
		t.Error("Expected IP address to be replaced with [REDACTED_IP]")
	}
}

func TestMockAdvisoryProvider(t *testing.T) {
	provider := NewMockAdvisoryProvider()
	orgID := uuid.New()
	devID := uuid.New()
	inc := &domain.Incident{
		ID:             uuid.New(),
		OrganizationID: orgID,
		Title:          "Host Disk Saturated",
		Severity:       domain.SeverityCritical,
		Status:         domain.IncidentStatusOpen,
		CreatedAt:      time.Now(),
	}

	dev := &domain.Device{
		ID:             devID,
		OrganizationID: orgID,
		Hostname:       "srv-db-01",
		Status:         domain.DeviceStatusOnline,
	}

	alerts := []*domain.Alert{
		{
			ID:             uuid.New(),
			OrganizationID: orgID,
			DeviceID:       devID,
			Message:        "Root volume disk storage exceeds 92%",
			RuleName:       "High Disk Utilization",
			Severity:       "critical",
		},
	}

	report, err := provider.GenerateAdvisory(context.Background(), inc, alerts, dev, nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if report.IncidentID != inc.ID {
		t.Errorf("Expected incident ID %v, got %v", inc.ID, report.IncidentID)
	}
	if len(report.Hypotheses) == 0 {
		t.Error("Expected at least one hypothesis")
	}
	if len(report.SuggestedRemediations) == 0 {
		t.Error("Expected at least one recommendation")
	}
}

func containsString(haystack, needle string) bool {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}
