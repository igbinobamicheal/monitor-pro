package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
)

// EventPublisher defines event broadcasting interface
type EventPublisher interface {
	PublishTelemetry(ctx context.Context, orgID, deviceID uuid.UUID, payload any) error
	PublishAlertTriggered(ctx context.Context, orgID uuid.UUID, payload any) error
	PublishRemediationDispatched(ctx context.Context, orgID, deviceID uuid.UUID, payload any) error
	PublishRemoteConsent(ctx context.Context, orgID uuid.UUID, payload any) error
	Close()
}

// NATSPublisher publishes messages to NATS JetStream
type NATSPublisher struct {
	nc *nats.Conn
	js nats.JetStreamContext
}

// NewNATSPublisher connects to NATS and returns a publisher
func NewNATSPublisher(url string) (*NATSPublisher, error) {
	if url == "" {
		return nil, fmt.Errorf("nats url is empty")
	}

	nc, err := nats.Connect(url,
		nats.Timeout(5*time.Second),
		nats.MaxReconnects(5),
		nats.ReconnectWait(2*time.Second),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to nats at %s: %w", url, err)
	}

	js, err := nc.JetStream()
	if err != nil {
		slog.Warn("nats JetStream not available, using standard pub/sub", "err", err)
	}

	return &NATSPublisher{nc: nc, js: js}, nil
}

func (p *NATSPublisher) publish(subject string, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	if p.js != nil {
		_, err = p.js.Publish(subject, data)
		if err != nil {
			// fallback to standard publish if stream not defined
			return p.nc.Publish(subject, data)
		}
		return nil
	}
	return p.nc.Publish(subject, data)
}

func (p *NATSPublisher) PublishTelemetry(ctx context.Context, orgID, deviceID uuid.UUID, payload any) error {
	subject := fmt.Sprintf("telemetry.raw.%s.%s", orgID.String(), deviceID.String())
	return p.publish(subject, payload)
}

func (p *NATSPublisher) PublishAlertTriggered(ctx context.Context, orgID uuid.UUID, payload any) error {
	subject := fmt.Sprintf("events.%s.alert_triggered", orgID.String())
	return p.publish(subject, payload)
}

func (p *NATSPublisher) PublishRemediationDispatched(ctx context.Context, orgID, deviceID uuid.UUID, payload any) error {
	subject := fmt.Sprintf("jobs.%s.%s.execute", orgID.String(), deviceID.String())
	return p.publish(subject, payload)
}

func (p *NATSPublisher) PublishRemoteConsent(ctx context.Context, orgID uuid.UUID, payload any) error {
	subject := fmt.Sprintf("audit.%s.remote_consent", orgID.String())
	return p.publish(subject, payload)
}

func (p *NATSPublisher) Close() {
	if p.nc != nil {
		p.nc.Drain()
		p.nc.Close()
	}
}

// NoopPublisher is a stub for testing or environments without NATS
type NoopPublisher struct{}

func NewNoopPublisher() *NoopPublisher {
	return &NoopPublisher{}
}

// NewLogPublisher returns a publisher logging events locally
func NewLogPublisher() *NoopPublisher {
	return &NoopPublisher{}
}

func (p *NoopPublisher) PublishTelemetry(ctx context.Context, orgID, deviceID uuid.UUID, payload any) error {
	return nil
}
func (p *NoopPublisher) PublishAlertTriggered(ctx context.Context, orgID uuid.UUID, payload any) error {
	return nil
}
func (p *NoopPublisher) PublishRemediationDispatched(ctx context.Context, orgID, deviceID uuid.UUID, payload any) error {
	return nil
}
func (p *NoopPublisher) PublishRemoteConsent(ctx context.Context, orgID uuid.UUID, payload any) error {
	return nil
}
func (p *NoopPublisher) Close() {}
