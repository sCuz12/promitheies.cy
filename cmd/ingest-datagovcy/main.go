// Command ingest-datagovcy imports Cyprus's awarded-contracts dataset from
// data.gov.cy: one CSV resource per year, discovered from the dataset's
// HTML page (the portal has no public API). Safe to re-run — imports are
// idempotent per CFTID.
package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/georgehadjisavvas/promitheies-cy/internal/datagovcy"
	"github.com/georgehadjisavvas/promitheies-cy/internal/db"
	"github.com/georgehadjisavvas/promitheies-cy/internal/envconfig"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	if err := envconfig.Load(); err != nil {
		log.Fatalf("load .env: %v", err)
	}

	ctx := context.Background()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is not set")
	}
	baseURL := os.Getenv("DATAGOVCY_BASE_URL")
	datasetURL := os.Getenv("DATAGOVCY_DATASET_URL")
	if baseURL == "" || datasetURL == "" {
		log.Fatal("DATAGOVCY_BASE_URL and DATAGOVCY_DATASET_URL must be set")
	}

	pool, err := db.NewPool(ctx, dbURL)
	if err != nil {
		log.Fatalf("connect to database: %v", err)
	}
	defer pool.Close()

	resources, err := datagovcy.DiscoverResources(baseURL, datasetURL)
	if err != nil {
		log.Fatalf("discover resources: %v", err)
	}
	log.Printf("found %d resources", len(resources))

	var totalProcessed, totalInserted, totalUpdated, totalSkipped int
	for _, r := range resources {
		log.Printf("importing %q (%s)", r.Title, r.DownloadURL)

		runID, err := startIngestRun(ctx, pool, datagovcy.Source)
		if err != nil {
			log.Fatalf("start ingest run: %v", err)
		}

		awards, err := datagovcy.Fetch(r.DownloadURL)
		if err != nil {
			finishIngestRun(ctx, pool, runID, 0, 0, "failed", err.Error())
			log.Fatalf("fetch %q: %v", r.Title, err)
		}

		stats, err := datagovcy.Import(ctx, pool, awards)
		if err != nil {
			finishIngestRun(ctx, pool, runID, stats.Processed, stats.Inserted, "failed", err.Error())
			log.Fatalf("import %q: %v", r.Title, err)
		}

		finishIngestRun(ctx, pool, runID, stats.Processed, stats.Inserted, "success", "")
		log.Printf("  processed=%d inserted=%d updated=%d skipped=%d",
			stats.Processed, stats.Inserted, stats.Updated, stats.Skipped)

		totalProcessed += stats.Processed
		totalInserted += stats.Inserted
		totalUpdated += stats.Updated
		totalSkipped += stats.Skipped
	}

	log.Printf("done: processed=%d inserted=%d updated=%d skipped=%d",
		totalProcessed, totalInserted, totalUpdated, totalSkipped)
}

func startIngestRun(ctx context.Context, pool *pgxpool.Pool, source string) (int64, error) {
	var id int64
	err := pool.QueryRow(ctx,
		`INSERT INTO ingest_runs (source, started_at, status) VALUES ($1, now(), 'running') RETURNING id`,
		source,
	).Scan(&id)
	return id, err
}

func finishIngestRun(ctx context.Context, pool *pgxpool.Pool, id int64, processed, inserted int, status, errMsg string) {
	_, err := pool.Exec(ctx, `
		UPDATE ingest_runs
		SET finished_at = $2, records_processed = $3, records_inserted = $4, status = $5, error = NULLIF($6, '')
		WHERE id = $1
	`, id, time.Now(), processed, inserted, status, errMsg)
	if err != nil {
		log.Printf("warning: failed to record ingest_run %d: %v", id, err)
	}
}
