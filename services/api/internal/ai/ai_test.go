package ai_test

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
	"monitopro/services/api/internal/ai"
	"monitopro/services/api/internal/domain"
)

func TestPIIRedaction(t *testing.T) {
	input := "Alert on host workstation-01 at 192.168.1.105 with MAC 00:1A:2B:3C:4D:5E. User: michele@corp.monitopro.com path: C:\\Users\\john_doe\\AppData\\Local\\Temp and Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9"

	redacted := ai.RedactPII(input)

	if strings.Contains(redacted, "192.168.1.105") {
		t.Errorf("IPv4 was not redacted: %s", redacted)
	}
	if strings.Contains(redacted, "00:1A:2B:3C:4D:5E") {
		t.Errorf("MAC was not redacted: %s", redacted)
	}
	if strings.Contains(redacted, "michele@corp.monitopro.com") {
		t.Errorf("Email was not redacted: %s", redacted)
	}
	if strings.Contains(redacted, "john_doe") {
		t.Errorf("Windows username was not redacted: %s", redacted)
	}
	if strings.Contains(redacted, "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9") {
		t.Errorf("Bearer token was not redacted: %s", redacted)
	}
}

func TestMockAdvisoryProvider(t *testing.T) {
	provider := ai.NewMockAdvisoryProvider()
	orgID := uuid.New()
	incidentID := uuid.New()
	deviceID := uuid.New()

	incident := &domain.Incident{
		ID:             incidentID,
		OrganizationID: orgID,
		Title:          "Critical CPU Saturation and Disk Shortage",
		Severity:       domain.SeverityCritical,
		Status:         domain.IncidentStatusInvestigating,
	}

	alerts := []*domain.Alert{
		{
			ID:             uuid.New(),
			OrganizationID: orgID,
			DeviceID:       deviceID,
			RuleName:       "High CPU Utilization",
			Message:        "CPU utilization exceeded 92%",
			Severity:       "critical",
		},
		{
			ID:             uuid.New(),
			OrganizationID: orgID,
			DeviceID:       deviceID,
			RuleName:       "High Disk Utilization",
			Message:        "Disk usage exceeded 91%",
			Severity:       "warning",
		},
	}

	device := &domain.Device{
		ID:             deviceID,
		OrganizationID: orgID,
		Hostname:       "srv-db-01",
		Status:         domain.DeviceStatusDegraded,
		OSVersion:      "Windows Server 2022",
	}

	res, err := provider.GenerateAdvisory(context.Background(), incident, alerts, device, nil)
	if err != nil {
		t.Fatalf("Failed to generate advisory: %v", err)
	}

	if res.IncidentID != incidentID {
		t.Errorf("IncidentID mismatch: expected %v, got %v", incidentID, res.IncidentID)
	}
	if len(res.Hypotheses) == 0 {
		t.Fatal("Expected at least one hypothesis")
	}
	if len(res.RecommendedVerificationSteps) == 0 {
		t.Fatal("Expected recommended verification steps")
	}
	if len(res.SuggestedRemediations) == 0 {
		t.Fatal("Expected suggested remediations")
	}

	for _, h := range res.Hypotheses {
		if h.Confidence <= 0 || h.Confidence > 1.0 {
			t.Errorf("Confidence out of range: %f", h.Confidence)
		}
		if len(h.SupportingEvidence) == 0 {
			t.Errorf("Hypothesis %d missing supporting evidence", h.Rank)
		}
	}
}
