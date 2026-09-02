// Command distribute generates copy-ready distribution drafts from newly
// imported tender data.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/georgehadjisavvas/promitheies-cy/internal/db"
	"github.com/georgehadjisavvas/promitheies-cy/internal/distribution"
)

func main() {
	var (
		cadence  = flag.String("cadence", distribution.CadenceDaily, "distribution cadence: daily or weekly")
		platform = flag.String("platform", distribution.PlatformX, "distribution platform: x")
		date     = flag.String("date", "", "selected date in YYYY-MM-DD format; defaults to today")
		limit    = flag.Int("limit", 20, "maximum TED tenders to include in the LLM prompt")
	)
	flag.Parse()

	if *platform != distribution.PlatformX {
		log.Fatalf("unsupported platform %q; only x is supported", *platform)
	}

	ctx := context.Background()
	loc := time.Local
	selected, err := selectedDate(*date, loc)
	if err != nil {
		log.Fatal(err)
	}
	window, err := distribution.WindowFor(selected, *cadence)
	if err != nil {
		log.Fatal(err)
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is not set")
	}
	pool, err := db.NewPool(ctx, dbURL)
	if err != nil {
		log.Fatalf("connect to database: %v", err)
	}
	defer pool.Close()

	tenders, err := distribution.LoadNewTEDTenders(ctx, pool, window, *limit)
	if err != nil {
		log.Fatalf("load TED tenders: %v", err)
	}

	req := distribution.DraftRequest{
		Cadence:       *cadence,
		Platform:      *platform,
		Window:        window,
		Tenders:       tenders,
		PublicBaseURL: os.Getenv("PUBLIC_BASE_URL"),
	}

	llm := distribution.NewOpenAIClient(os.Getenv("OPENAI_API_KEY"), openAIModel())
	posts, err := distribution.GenerateDraft(ctx, llm, req)
	if err != nil {
		log.Fatalf("generate draft: %v", err)
	}

	filename, err := distribution.OutputFilename(*platform, *cadence, window)
	if err != nil {
		log.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(filename), 0o755); err != nil {
		log.Fatalf("create posts directory: %v", err)
	}
	if err := os.WriteFile(filename, []byte(distribution.MarkdownDraft(req, posts)), 0o644); err != nil {
		log.Fatalf("write draft: %v", err)
	}

	fmt.Printf("wrote %s\n\n", filename)
	for i, post := range posts {
		fmt.Printf("Option %d:\n%s\n\n", i+1, post)
	}
}

func selectedDate(raw string, loc *time.Location) (time.Time, error) {
	if raw == "" {
		return time.Now().In(loc), nil
	}
	t, err := time.ParseInLocation("2006-01-02", raw, loc)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse -date: %w", err)
	}
	return t, nil
}

func openAIModel() string {
	if model := os.Getenv("OPENAI_MODEL"); model != "" {
		return model
	}
	return "gpt-5.6-luna"
}
