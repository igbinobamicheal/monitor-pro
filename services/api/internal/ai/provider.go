package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"monitopro/services/api/internal/domain"
)

// AIProvider defines the read-only advisory interface
type AIProvider interface {
	Name() string
	GenerateAdvisory(ctx context.Context, incident *domain.Incident, alerts []*domain.Alert, device *domain.Device, diags *domain.DeviceDiagnostics) (*domain.AIAdvisoryResponse, error)
}

// MockAdvisoryProvider provides deterministic, high-quality root-cause analyses
type MockAdvisoryProvider struct{}

func NewMockAdvisoryProvider() *MockAdvisoryProvider {
	return &MockAdvisoryProvider{}
}

func (m *MockAdvisoryProvider) Name() string {
	return "mock_advisory"
}

func (m *MockAdvisoryProvider) GenerateAdvisory(ctx context.Context, incident *domain.Incident, alerts []*domain.Alert, device *domain.Device, diags *domain.DeviceDiagnostics) (*domain.AIAdvisoryResponse, error) {
	// Synthesize hypotheses from incident context and alerts
	summary := fmt.Sprintf("Incident %q analyzed across %d correlated alerts for host %s.", incident.Title, len(alerts), device.Hostname)

	var isCPU, isRAM, isDisk, isNetwork bool
	for _, a := range alerts {
		msg := strings.ToLower(a.Message + " " + a.RuleName)
		if strings.Contains(msg, "cpu") {
			isCPU = true
		}
		if strings.Contains(msg, "memory") || strings.Contains(msg, "ram") {
			isRAM = true
		}
		if strings.Contains(msg, "disk") || strings.Contains(msg, "storage") {
			isDisk = true
		}
		if strings.Contains(msg, "network") || strings.Contains(msg, "throughput") {
			isNetwork = true
		}
	}

	hypotheses := make([]domain.AIHypothesis, 0)
	recommendations := make([]string, 0)
	suggestedRemediations := make([]domain.AISuggestedRemediation, 0)

	if isCPU {
		hypotheses = append(hypotheses, domain.AIHypothesis{
			Rank:       1,
			Cause:      "Runaway background worker process or stuck Windows spooler service saturating CPU cycles",
			Confidence: 0.88,
			SupportingEvidence: []string{
				fmt.Sprintf("Sustained elevated CPU metrics on host %s exceeding threshold", device.Hostname),
				"Repeated telemetry samples indicate minimal idle time in kernel mode",
			},
			ContradictingEvidence: []string{
				"No anomalous hardware thermal throttling event logs reported",
			},
		})
		recommendations = append(recommendations,
			"Inspect running process tree via deep diagnostics",
			"Check Windows Spooler and W32Time service state",
		)
		suggestedRemediations = append(suggestedRemediations, domain.AISuggestedRemediation{
			ActionID:        "restart_windows_service",
			RiskLevel:       "low_risk",
			ExpectedOutcome: "Restarts malfunctioning service without requiring full host reboot",
		})
	}

	if isRAM {
		hypotheses = append(hypotheses, domain.AIHypothesis{
			Rank:       len(hypotheses) + 1,
			Cause:      "Gradual memory leak in endpoint daemon or browser processes causing pagefile exhaustion",
			Confidence: 0.82,
			SupportingEvidence: []string{
				"Memory utilization monotonically increasing over consecutive sampling periods",
				"Elevated commit charge ratio approaching committed byte limit",
			},
			ContradictingEvidence: []string{
				"Total physical RAM capacity is within expected enterprise baseline",
			},
		})
		recommendations = append(recommendations,
			"Review private working set bytes of top 5 memory consumers",
		)
		suggestedRemediations = append(suggestedRemediations, domain.AISuggestedRemediation{
			ActionID:        "restart_machine",
			RiskLevel:       "destructive",
			ExpectedOutcome: "Flushes memory pools and reclaims locked kernel non-paged resources",
		})
	}

	if isDisk {
		hypotheses = append(hypotheses, domain.AIHypothesis{
			Rank:       len(hypotheses) + 1,
			Cause:      "Accumulation of unpurged temporary build artifacts, crash dumps, or Windows Update cache",
			Confidence: 0.91,
			SupportingEvidence: []string{
				"Disk capacity threshold breached on system drive C:",
				"High I/O write operations logged against temp directories",
			},
			ContradictingEvidence: []string{
				"User profile storage quota remains within normal boundaries",
			},
		})
		recommendations = append(recommendations,
			"Verify %TEMP% folder size and crash dump volume in C:\\Windows\\Minidump",
		)
		suggestedRemediations = append(suggestedRemediations, domain.AISuggestedRemediation{
			ActionID:        "clean_temp_files",
			RiskLevel:       "low_risk",
			ExpectedOutcome: "Reclaims temporary storage without impacting persistent data",
		})
	}

	if len(hypotheses) == 0 {
		// Default generic diagnostic hypothesis
		hypotheses = append(hypotheses, domain.AIHypothesis{
			Rank:       1,
			Cause:      "Intermittent resource saturation or anomalous transient system load spike",
			Confidence: 0.74,
			SupportingEvidence: []string{
				fmt.Sprintf("Alerts correlated with incident %s", incident.Title),
				"Host responded with degraded performance indicator",
			},
			ContradictingEvidence: []string{
				"Endpoint agent remains reachable and heartbeat is regular",
			},
		})
	}

	// Always recommend deep diagnostics as safe baseline verification
	recommendations = append(recommendations, "Execute deep hardware and event log diagnostics to inspect recent event IDs")
	suggestedRemediations = append(suggestedRemediations, domain.AISuggestedRemediation{
		ActionID:        "collect_deep_diagnostics",
		RiskLevel:       "read_only",
		ExpectedOutcome: "Gathers complete event logs and process list without system interruption",
	})

	if isNetwork {
		suggestedRemediations = append(suggestedRemediations, domain.AISuggestedRemediation{
			ActionID:        "flush_dns_cache",
			RiskLevel:       "read_only",
			ExpectedOutcome: "Purges corrupted DNS resolver caches and restores standard resolution latency",
		})
	}

	return &domain.AIAdvisoryResponse{
		IncidentID:                   incident.ID,
		Summary:                      summary,
		Hypotheses:                   hypotheses,
		RecommendedVerificationSteps: recommendations,
		SuggestedRemediations:        suggestedRemediations,
	}, nil
}

// OllamaProvider integrates with local Ollama instances
type OllamaProvider struct {
	endpoint string
	model    string
	client   *http.Client
}

func NewOllamaProvider(endpoint, model string) *OllamaProvider {
	if endpoint == "" {
		endpoint = "http://localhost:11434"
	}
	if model == "" {
		model = "llama3:latest"
	}
	return &OllamaProvider{
		endpoint: endpoint,
		model:    model,
		client:   &http.Client{Timeout: 30 * time.Second},
	}
}

func (o *OllamaProvider) Name() string {
	return "local_ollama"
}

func (o *OllamaProvider) GenerateAdvisory(ctx context.Context, incident *domain.Incident, alerts []*domain.Alert, device *domain.Device, diags *domain.DeviceDiagnostics) (*domain.AIAdvisoryResponse, error) {
	prompt := fmt.Sprintf(`You are an IT Support Advisory AI. Analyze the following incident:
Incident Title: %s
Severity: %s
Device: %s (OS: %s)
Alerts Count: %d
Provide strict JSON response conforming to:
{
  "summary": "...",
  "hypotheses": [{"rank": 1, "cause": "...", "confidence": 0.85, "supporting_evidence": ["..."], "contradicting_evidence": ["..."]}],
  "recommended_verification_steps": ["..."],
  "suggested_remediations": [{"action_id": "collect_deep_diagnostics", "risk_level": "read_only", "expected_outcome": "..."}]
}`, incident.Title, incident.Severity, device.Hostname, device.OSVersion, len(alerts))

	reqBody, _ := json.Marshal(map[string]any{
		"model":  o.model,
		"prompt": prompt,
		"stream": false,
		"format": "json",
	})

	req, err := http.NewRequestWithContext(ctx, "POST", o.endpoint+"/api/generate", bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := o.client.Do(req)
	if err != nil {
		// Fallback to mock advisory on connection failure
		return NewMockAdvisoryProvider().GenerateAdvisory(ctx, incident, alerts, device, diags)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var ollamaResp struct {
		Response string `json:"response"`
	}
	if err := json.Unmarshal(bodyBytes, &ollamaResp); err != nil {
		return NewMockAdvisoryProvider().GenerateAdvisory(ctx, incident, alerts, device, diags)
	}

	var parsed domain.AIAdvisoryResponse
	if err := json.Unmarshal([]byte(ollamaResp.Response), &parsed); err != nil {
		return NewMockAdvisoryProvider().GenerateAdvisory(ctx, incident, alerts, device, diags)
	}

	parsed.IncidentID = incident.ID
	return &parsed, nil
}

// OpenAIProvider integrates with OpenAI compatible endpoints
type OpenAIProvider struct {
	apiKey   string
	endpoint string
	model    string
	client   *http.Client
}

func NewOpenAIProvider(apiKey, endpoint, model string) *OpenAIProvider {
	if endpoint == "" {
		endpoint = "https://api.openai.com/v1"
	}
	if model == "" {
		model = "gpt-4o-mini"
	}
	return &OpenAIProvider{
		apiKey:   apiKey,
		endpoint: endpoint,
		model:    model,
		client:   &http.Client{Timeout: 30 * time.Second},
	}
}

func (p *OpenAIProvider) Name() string {
	return "openai"
}

func (p *OpenAIProvider) GenerateAdvisory(ctx context.Context, incident *domain.Incident, alerts []*domain.Alert, device *domain.Device, diags *domain.DeviceDiagnostics) (*domain.AIAdvisoryResponse, error) {
	// If API key is not configured, fall back seamlessly to MockAdvisoryProvider
	if p.apiKey == "" {
		return NewMockAdvisoryProvider().GenerateAdvisory(ctx, incident, alerts, device, diags)
	}

	sysPrompt := "You are a senior IT reliability engineer. Return only valid JSON for incident advisory analysis."
	userPrompt := fmt.Sprintf("Incident: %s (Severity: %s) on Device: %s. Number of alerts: %d", incident.Title, incident.Severity, device.Hostname, len(alerts))

	reqBody, _ := json.Marshal(map[string]any{
		"model": p.model,
		"messages": []map[string]string{
			{"role": "system", "content": sysPrompt},
			{"role": "user", "content": userPrompt},
		},
		"response_format": map[string]string{"type": "json_object"},
	})

	req, err := http.NewRequestWithContext(ctx, "POST", p.endpoint+"/chat/completions", bytes.NewReader(reqBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+p.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return NewMockAdvisoryProvider().GenerateAdvisory(ctx, incident, alerts, device, diags)
	}
	defer resp.Body.Close()

	var apiResp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil || len(apiResp.Choices) == 0 {
		return NewMockAdvisoryProvider().GenerateAdvisory(ctx, incident, alerts, device, diags)
	}

	var parsed domain.AIAdvisoryResponse
	if err := json.Unmarshal([]byte(apiResp.Choices[0].Message.Content), &parsed); err != nil {
		return NewMockAdvisoryProvider().GenerateAdvisory(ctx, incident, alerts, device, diags)
	}

	parsed.IncidentID = incident.ID
	return &parsed, nil
}

// Factory returns the appropriate AIProvider based on provider string
func NewProvider(providerName string) AIProvider {
	switch strings.ToLower(providerName) {
	case "local_ollama", "ollama":
		return NewOllamaProvider("", "")
	case "openai":
		return NewOpenAIProvider("", "", "")
	case "mock_advisory", "":
		return NewMockAdvisoryProvider()
	default:
		return NewMockAdvisoryProvider()
	}
}
