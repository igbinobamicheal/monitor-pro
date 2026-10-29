package processor

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
)

func TestTelemetryEventDeserialization(t *testing.T) {
	raw := `{
		"event_id": "` + uuid.New().String() + `",
		"organization_id": "` + uuid.New().String() + `",
		"device_id": "` + uuid.New().String() + `",
		"cpu_usage": 92.5,
		"ram_usage": 84.1,
		"disk_usage": 76.0,
		"network_in": 1048576.0,
		"network_out": 524288.0,
		"uptime": 86400,
		"timestamp": "` + time.Now().Format(time.RFC3339) + `"
	}`

	var ev TelemetryEvent
	if err := json.Unmarshal([]byte(raw), &ev); err != nil {
		t.Fatalf("Failed to unmarshal telemetry event: %v", err)
	}

	if ev.CPUUsage != 92.5 {
		t.Errorf("Expected CPU 92.5, got %v", ev.CPUUsage)
	}
	if ev.RAMUsage != 84.1 {
		t.Errorf("Expected RAM 84.1, got %v", ev.RAMUsage)
	}
	if ev.DiskUsage != 76.0 {
		t.Errorf("Expected Disk 76.0, got %v", ev.DiskUsage)
	}
}

func TestTelemetryProcessorNilSafety(t *testing.T) {
	// Processor without DB pool or NATS should handle message gracefully without panic
	proc := NewTelemetryProcessor(nil, nil, nil)
	ctx := context.Background()

	msg := &nats.Msg{
		Data: []byte(`{"cpu_usage": 50.0, "ram_usage": 50.0, "disk_usage": 50.0}`),
	}

	err := proc.handleMessage(ctx, msg)
	if err != nil {
		t.Errorf("Expected nil error for safe nil DB handling, got %v", err)
	}
}
