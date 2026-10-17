// Command server runs the free transparency dashboard: an HTTP server
// serving read-only pages over the tenders/awards data ingested by the
// import commands.
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/georgehadjisavvas/promitheies-cy/internal/db"
	"github.com/georgehadjisavvas/promitheies-cy/internal/envconfig"
	"github.com/georgehadjisavvas/promitheies-cy/internal/mcpserver"
	"github.com/georgehadjisavvas/promitheies-cy/internal/ratelimit"
	"github.com/georgehadjisavvas/promitheies-cy/internal/web"
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
	addr := os.Getenv("HTTP_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	pool, err := db.NewPool(ctx, dbURL)
	if err != nil {
		log.Fatalf("connect to database: %v", err)
	}
	defer pool.Close()

	srv, err := web.NewServer(pool, "templates", "static")
	if err != nil {
		log.Fatalf("init server: %v", err)
	}

	root := http.NewServeMux()
	rl := ratelimit.New(ctx, envFloat("MCP_RATE_LIMIT_RPS", 5), int(envFloat("MCP_RATE_LIMIT_BURST", 20)))
	root.Handle("/mcp", rl.Middleware(mcpserver.Handler(pool)))
	root.Handle("/", srv)

	log.Printf("listening on %s", addr)
	if err := http.ListenAndServe(addr, root); err != nil {
		log.Fatal(err)
	}
}

func envFloat(name string, fallback float64) float64 {
	v, err := strconv.ParseFloat(os.Getenv(name), 64)
	if err != nil || v <= 0 {
		return fallback
	}
	return v
}
