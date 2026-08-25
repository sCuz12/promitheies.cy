// Command jobs runs one-off and scheduled maintenance tasks (seeding,
// stats recomputation, alert dispatch, weekly summaries) as subcommands,
// invoked directly or via cron/systemd timers.
package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/georgehadjisavvas/promitheies-cy/internal/cpv"
	"github.com/georgehadjisavvas/promitheies-cy/internal/db"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: jobs <seed-cpv>")
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
	default:
		log.Fatalf("unknown command %q", os.Args[1])
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
