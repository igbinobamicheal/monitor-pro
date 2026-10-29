package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"monitopro/services/api/internal/ai"
	"monitopro/services/api/internal/auth"
	"monitopro/services/api/internal/config"
	"monitopro/services/api/internal/events"
	"monitopro/services/api/internal/handlers"
	"monitopro/services/api/internal/remediation"
	"monitopro/services/api/internal/storage"
)

func main() {
	log.Println("[ControlPlane] Starting Proactive IT Support Dashboard V2 API...")

	// 1. Load Configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("[ControlPlane] Failed to load config: %v", err)
	}

	// 2. Initialize Database Storage
	var store storage.Storage
	var healthHandler *handlers.HealthHandler

	pgStore, err := storage.NewPostgresStorageFromConnString(context.Background(), cfg.DatabaseURL())
	if err != nil {
		log.Printf("[ControlPlane] Notice: PostgreSQL offline (%v); falling back to in-memory store for local startup.", err)
		store = storage.NewMemoryStorage()
		healthHandler = handlers.NewHealthHandler(cfg.ReplicaID, nil)
	} else {
		defer pgStore.Close()
		store = pgStore
		healthHandler = handlers.NewHealthHandler(cfg.ReplicaID, pgStore.Pool())
		log.Println("[ControlPlane] PostgreSQL connection pool active.")
	}

	// 3. Components
	tokenManager := auth.NewTokenManager(cfg.JWTSecret, 24*time.Hour)
	var publisher events.EventPublisher
	if natsPub, err := events.NewNATSPublisher(cfg.NATSURL); err == nil {
		publisher = natsPub
		defer natsPub.Close()
		log.Printf("[ControlPlane] NATS JetStream connected at %s", cfg.NATSURL)
	} else {
		log.Printf("[ControlPlane] NATS unavailable (%v); using local event publisher", err)
		publisher = events.NewLogPublisher()
	}

	catalogue, err := remediation.LoadCatalogue("shared/contracts/remediation_catalogue.json")
	if err != nil {
		catalogue, err = remediation.LoadCatalogue("../../shared/contracts/remediation_catalogue.json")
		if err != nil {
			catalogue = remediation.NewCatalogue()
		}
	}
	aiProvider := ai.NewMockAdvisoryProvider()

	// 4. Construct HTTP Router
	router := handlers.NewRouter(&handlers.RouterConfig{
		Store:         store,
		TokenManager:  tokenManager,
		Publisher:     publisher,
		Catalogue:     catalogue,
		AIProvider:    aiProvider,
		ReplicaID:     cfg.ReplicaID,
		HealthHandler: healthHandler,
	})

	// 5. Server with Graceful Draining
	server := &http.Server{
		Addr:         fmt.Sprintf("0.0.0.0:%d", cfg.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("[ControlPlane] Replica %s listening on :%d", cfg.ReplicaID, cfg.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[ControlPlane] Server failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("[ControlPlane] Initiating graceful shutdown...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("[ControlPlane] Server forced shutdown: %v", err)
	}

	log.Println("[ControlPlane] Server gracefully exited.")
}
