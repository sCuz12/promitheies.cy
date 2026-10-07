// Command jobs runs one-off and scheduled maintenance tasks (seeding,
// stats recomputation, alert dispatch, weekly summaries) as subcommands,
// invoked directly or via cron/systemd timers.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/georgehadjisavvas/promitheies-cy/internal/cpv"
	"github.com/georgehadjisavvas/promitheies-cy/internal/db"
	"github.com/georgehadjisavvas/promitheies-cy/internal/envconfig"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	if err := envconfig.Load(); err != nil {
		log.Fatalf("load .env: %v", err)
	}

	if len(os.Args) < 2 {
		log.Fatal("usage: jobs <seed-cpv|newsletter-digest>")
	}

	ctx := context.Background()
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	pool, err := db.NewPool(ctx, dbURL)
	if err != nil {
		log.Fatalf("connect to database: %v", err)
	}
	defer pool.Close()

	switch os.Args[1] {
	case "seed-cpv":
		if err := seedCPV(ctx, pool); err != nil {
			log.Fatalf("seed-cpv: %v", err)
		}
	case "newsletter-digest":
		if err := runNewsletterDigest(ctx, pool); err != nil {
			log.Fatalf("newsletter-digest: %v", err)
		}
	default:
		log.Fatalf("unknown command %q", os.Args[1])
	}
}

// startIngestRun and finishIngestRun mirror the identical helpers in
// cmd/ingest-ted/main.go; copied rather than shared since these are two
// separate `main` packages and it's not worth extracting a package for two
// call sites. They record job runs in the existing ingest_runs table,
// repurposed here for non-ingest jobs: for newsletter-digest,
// records_processed = subscribers processed and records_inserted =
// successful sends.
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

func seedCPV(ctx context.Context, pool *pgxpool.Pool) error {
	for _, d := range cpv.Divisions {
		_, err := pool.Exec(ctx, `
			INSERT INTO cpv_categories (code, description_en)
			VALUES ($1, $2)
			ON CONFLICT (code) DO UPDATE SET description_en = EXCLUDED.description_en
		`, d.Code, d.DescriptionEN)
		if err != nil {
			return fmt.Errorf("upsert cpv %s: %w", d.Code, err)
		}
	}
	log.Printf("seeded %d CPV divisions", len(cpv.Divisions))
	return nil
}
