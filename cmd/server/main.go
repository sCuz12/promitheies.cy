// Command server runs the free transparency dashboard: an HTTP server
// serving read-only pages over the tenders/awards data ingested by the
// import commands.
package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/georgehadjisavvas/promitheies-cy/internal/db"
	"github.com/georgehadjisavvas/promitheies-cy/internal/web"
)

func main() {
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

	log.Printf("listening on %s", addr)
	if err := http.ListenAndServe(addr, srv); err != nil {
		log.Fatal(err)
	}
}
