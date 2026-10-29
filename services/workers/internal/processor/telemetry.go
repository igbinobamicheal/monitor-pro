package processor

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
)

type TelemetryEvent struct {
	EventID        string    `json:"event_id"`
	OrganizationID string    `json:"organization_id"`
	DeviceID       string    `json:"device_id"`
	CPUUsage       float64   `json:"cpu_usage"`
	RAMUsage       float64   `json:"ram_usage"`
	DiskUsage      float64   `json:"disk_usage"`
	NetworkIn      float64   `json:"network_in"`
	NetworkOut     float64   `json:"network_out"`
	Uptime         int64     `json:"uptime"`
	Timestamp      time.Time `json:"timestamp"`
}

type TelemetryProcessor struct {
	nc     *nats.Conn
	js     nats.JetStreamContext
	dbPool *pgxpool.Pool
}

func NewTelemetryProcessor(nc *nats.Conn, js nats.JetStreamContext, pool *pgxpool.Pool) *TelemetryProcessor {
	return &TelemetryProcessor{
		nc:     nc,
		js:     js,
		dbPool: pool,
	}
}

// Start begins consuming telemetry events with a durable consumer
func (p *TelemetryProcessor) Start(ctx context.Context) error {
	if p.js == nil {
		log.Println("[TelemetryProcessor] JetStream not available, running in standalone poll mode")
		return nil
	}

	// Create durable pull or push subscription
	sub, err := p.js.Subscribe("events.telemetry.*", func(msg *nats.Msg) {
		if err := p.handleMessage(ctx, msg); err != nil {
			log.Printf("[TelemetryProcessor] Error handling telemetry message: %v", err)
			_ = msg.Nak()
			return
		}
		_ = msg.Ack()
	}, nats.Durable("worker-telemetry-evaluator"), nats.ManualAck())

	if err != nil {
		return fmt.Errorf("failed to subscribe to telemetry events: %w", err)
	}

	log.Printf("[TelemetryProcessor] Subscribed to events.telemetry.* with durable consumer worker-telemetry-evaluator")

	<-ctx.Done()
	_ = sub.Unsubscribe()
	return nil
}

func (p *TelemetryProcessor) handleMessage(ctx context.Context, msg *nats.Msg) error {
	var ev TelemetryEvent
	if err := json.Unmarshal(msg.Data, &ev); err != nil {
		return fmt.Errorf("invalid json telemetry event: %w", err)
	}

	log.Printf("[TelemetryProcessor] Processing telemetry for device %s (org %s): CPU=%.1f%% RAM=%.1f%% Disk=%.1f%%",
		ev.DeviceID, ev.OrganizationID, ev.CPUUsage, ev.RAMUsage, ev.DiskUsage)

	if p.dbPool == nil {
		return nil
	}

	// 1. Evaluate rules for this organization
	// Query active rules
	rows, err := p.dbPool.Query(ctx,
		`SELECT id, metric_type, operator, threshold, severity 
		 FROM rules 
		 WHERE organization_id = $1 AND enabled = true`,
		ev.OrganizationID)
	if err != nil {
		return fmt.Errorf("failed to query rules: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var ruleID, metricType, op, severity string
		var threshold float64
		if err := rows.Scan(&ruleID, &metricType, &op, &threshold, &severity); err != nil {
			continue
		}

		var metricVal float64
		switch metricType {
		case "cpu":
			metricVal = ev.CPUUsage
		case "ram":
			metricVal = ev.RAMUsage
		case "disk":
			metricVal = ev.DiskUsage
		default:
			continue
		}

		breached := false
		switch op {
		case ">":
			breached = metricVal > threshold
		case ">=":
			breached = metricVal >= threshold
		case "<":
			breached = metricVal < threshold
		case "<=":
			breached = metricVal <= threshold
		}

		if breached {
			// Alert condition met! Record alert in database
			msgText := fmt.Sprintf("%s usage reached %.1f%% (threshold: %s %.1f%%)",
				metricType, metricVal, op, threshold)

			_, err := p.dbPool.Exec(ctx,
				`INSERT INTO alerts (id, organization_id, device_id, rule_id, message, severity, resolved, created_at)
				 VALUES ($1, $2, $3, $4, $5, $6, false, $7)`,
				uuid.New().String(), ev.OrganizationID, ev.DeviceID, ruleID, msgText, severity, time.Now())
			if err != nil {
				log.Printf("[TelemetryProcessor] Warning: failed to record alert: %v", err)
			} else {
				log.Printf("[TelemetryProcessor] ALERT TRIGGERED: %s for device %s", msgText, ev.DeviceID)
			}
		}
	}

	return nil
}
