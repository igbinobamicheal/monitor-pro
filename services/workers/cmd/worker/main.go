package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"

	"monitopro/services/workers/internal/archiver"
	"monitopro/services/workers/internal/processor"
)

func main() {
	log.Println("[WorkerDaemon] Starting MonitoPro Background Durable Worker...")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 1. PostgreSQL connection
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		databaseURL = "postgres://monitopro:monitopro_secure_pass@localhost:5432/monitopro?sslmode=disable"
	}

	var pool *pgxpool.Pool
	var err error
	pool, err = pgxpool.New(ctx, databaseURL)
	if err != nil {
		log.Printf("[WorkerDaemon] Warning: Could not connect to PostgreSQL: %v", err)
	} else {
		defer pool.Close()
		log.Println("[WorkerDaemon] Connected to PostgreSQL")
	}

	// 2. NATS connection
	natsURL := os.Getenv("NATS_URL")
	if natsURL == "" {
		natsURL = "nats://localhost:4222"
	}

	nc, err := nats.Connect(natsURL, nats.Timeout(5*time.Second))
	var js nats.JetStreamContext
	if err != nil {
		log.Printf("[WorkerDaemon] Warning: Could not connect to NATS at %s: %v", natsURL, err)
	} else {
		defer nc.Close()
		log.Printf("[WorkerDaemon] Connected to NATS broker at %s", natsURL)

		js, err = nc.JetStream()
		if err != nil {
			log.Printf("[WorkerDaemon] Warning: Could not initialize JetStream context: %v", err)
		}
	}

	// 3. Initialize components
	telemetryWorker := processor.NewTelemetryProcessor(nc, js, pool)
	retentionWorker := archiver.NewRetentionArchiver(pool, 1*time.Hour)

	// 4. Start workers
	go func() {
		if err := telemetryWorker.Start(ctx); err != nil {
			log.Printf("[WorkerDaemon] Telemetry processor exited: %v", err)
		}
	}()

	go retentionWorker.Start(ctx)

	log.Println("[WorkerDaemon] All background workers initialized. Ready for events.")

	// 5. Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigChan
	log.Printf("[WorkerDaemon] Received signal %v. Initiating graceful shutdown...", sig)
	cancel()

	// Drain period
	time.Sleep(1 * time.Second)
	log.Println("[WorkerDaemon] Graceful shutdown complete.")
}
