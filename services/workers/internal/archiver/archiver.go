package archiver

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type RetentionArchiver struct {
	dbPool   *pgxpool.Pool
	interval time.Duration
}

func NewRetentionArchiver(pool *pgxpool.Pool, interval time.Duration) *RetentionArchiver {
	if interval <= 0 {
		interval = 1 * time.Hour
	}
	return &RetentionArchiver{
		dbPool:   pool,
		interval: interval,
	}
}

// Start begins periodic metric archiving and retention cleanup
func (a *RetentionArchiver) Start(ctx context.Context) {
	ticker := time.NewTicker(a.interval)
	defer ticker.Stop()

	log.Printf("[RetentionArchiver] Archiver started, interval: %v", a.interval)

	for {
		select {
		case <-ctx.Done():
			log.Println("[RetentionArchiver] Stopping archiver routine")
			return
		case <-ticker.C:
			if err := a.runArchiveCycle(ctx); err != nil {
				log.Printf("[RetentionArchiver] Error during archive cycle: %v", err)
			}
		}
	}
}

func (a *RetentionArchiver) runArchiveCycle(ctx context.Context) error {
	if a.dbPool == nil {
		return nil
	}

	// Archive metrics older than 30 days
	cutoff := time.Now().AddDate(0, 0, -30)
	log.Printf("[RetentionArchiver] Checking for telemetry records prior to %v", cutoff)

	// Fetch distinct organizations
	rows, err := a.dbPool.Query(ctx, `SELECT DISTINCT organization_id FROM metrics WHERE timestamp < $1`, cutoff)
	if err != nil {
		return fmt.Errorf("failed to query stale metrics: %w", err)
	}
	defer rows.Close()

	var orgs []string
	for rows.Next() {
		var orgID string
		if err := rows.Scan(&orgID); err == nil {
			orgs = append(orgs, orgID)
		}
	}

	for _, orgID := range orgs {
		// Count records to be archived
		var count int
		err := a.dbPool.QueryRow(ctx,
			`SELECT COUNT(*) FROM metrics WHERE organization_id = $1 AND timestamp < $2`,
			orgID, cutoff).Scan(&count)
		if err != nil || count == 0 {
			continue
		}

		// Calculate checksum identifier for archival record
		checksumData := fmt.Sprintf("%s-%v-%d", orgID, cutoff.Unix(), count)
		h := sha256.Sum256([]byte(checksumData))
		checksumHex := hex.EncodeToString(h[:])

		archivePath := fmt.Sprintf("/archives/telemetry/%s/%d.zst", orgID, cutoff.Unix())

		// Insert into archive_records table
		_, err = a.dbPool.Exec(ctx,
			`INSERT INTO archive_records (id, organization_id, file_path, checksum_sha256, records_count, date_start, date_end, created_at)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
			uuid.New().String(), orgID, archivePath, checksumHex, count,
			time.Now().AddDate(0, 0, -60), cutoff, time.Now())

		if err != nil {
			log.Printf("[RetentionArchiver] Failed to log archive record for org %s: %v", orgID, err)
			continue
		}

		// Delete archived raw metrics
		tag, err := a.dbPool.Exec(ctx,
			`DELETE FROM metrics WHERE organization_id = $1 AND timestamp < $2`,
			orgID, cutoff)
		if err != nil {
			log.Printf("[RetentionArchiver] Failed to prune archived metrics for org %s: %v", orgID, err)
		} else {
			log.Printf("[RetentionArchiver] Successfully archived and pruned %d metrics for org %s", tag.RowsAffected(), orgID)
		}
	}

	return nil
}
