// Command ingest-ted imports Cyprus notices from the EU TED Search API
// (all form types: open competitions, planning, results). Safe to re-run —
// imports are idempotent per TED publication-number.
package main

import (
	"context"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/georgehadjisavvas/promitheies-cy/internal/alerts"
	"github.com/georgehadjisavvas/promitheies-cy/internal/db"
	"github.com/georgehadjisavvas/promitheies-cy/internal/ted"
	"github.com/georgehadjisavvas/promitheies-cy/internal/telegram"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx := context.Background()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is not set")
	}
	baseURL := os.Getenv("TED_API_BASE_URL")
	if baseURL == "" {
		log.Fatal("TED_API_BASE_URL is not set")
	}

	pool, err := db.NewPool(ctx, dbURL)
	if err != nil {
		log.Fatalf("connect to database: %v", err)
	}
	defer pool.Close()

	client := ted.NewClient(baseURL)
	telegramClient := newTelegramClientFromEnv()
	telegramMaxMessages := envInt("TELEGRAM_MAX_NEW_TENDER_MESSAGES", 20)
	publicBaseURL := os.Getenv("PUBLIC_BASE_URL")

	runID, err := startIngestRun(ctx, pool, ted.Source)
	if err != nil {
		log.Fatalf("start ingest run: %v", err)
	}

	var totalProcessed, totalInserted, totalUpdated, totalSkipped int
	var newTenders []alerts.Tender
	page := 1
	for {
		resp, err := client.Search(ctx, ted.CyprusQuery, ted.Fields, page, ted.MaxPageLimit)
		if err != nil {
			finishIngestRun(ctx, pool, runID, totalProcessed, totalInserted, "failed", err.Error())
			log.Fatalf("search page %d: %v", page, err)
		}
		if len(resp.Notices) == 0 {
			break
		}

		notices := make([]ted.Notice, 0, len(resp.Notices))
		for _, raw := range resp.Notices {
			n, err := ted.ParseNotice(raw)
			if err != nil {
				log.Printf("warning: skipping unparseable notice: %v", err)
				continue
			}
			notices = append(notices, n)
		}

		stats, err := ted.Import(ctx, pool, notices)
		if err != nil {
			finishIngestRun(ctx, pool, runID, totalProcessed, totalInserted, "failed", err.Error())
			log.Fatalf("import page %d: %v", page, err)
		}

		totalProcessed += stats.Processed
		totalInserted += stats.Inserted
		totalUpdated += stats.Updated
		totalSkipped += stats.Skipped
		newTenders = append(newTenders, stats.NewTenders...)

		log.Printf("page %d/%d: processed=%d inserted=%d updated=%d skipped=%d (total so far: %d)",
			page, (resp.TotalNoticeCount+ted.MaxPageLimit-1)/ted.MaxPageLimit,
			stats.Processed, stats.Inserted, stats.Updated, stats.Skipped, totalProcessed)

		if totalProcessed >= resp.TotalNoticeCount {
			break
		}
		page++
	}

	finishIngestRun(ctx, pool, runID, totalProcessed, totalInserted, "success", "")
	if telegramClient != nil && len(newTenders) > 0 {
		if err := alerts.SendNewTenderMessages(ctx, telegramClient, newTenders, telegramMaxMessages, publicBaseURL); err != nil {
			log.Printf("warning: failed to send telegram new-tender notifications: %v", err)
		} else {
			log.Printf("sent telegram notifications for %d new open tenders", len(newTenders))
		}
	}
	log.Printf("done: processed=%d inserted=%d updated=%d skipped=%d",
		totalProcessed, totalInserted, totalUpdated, totalSkipped)
}

func newTelegramClientFromEnv() *telegram.Client {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	chatID := os.Getenv("TELEGRAM_CHAT_ID")
	if token == "" || chatID == "" {
		return nil
	}
	return telegram.NewClient(token, chatID)
}

func envInt(name string, fallback int) int {
	raw := os.Getenv(name)
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 0 {
		return fallback
	}
	return value
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
