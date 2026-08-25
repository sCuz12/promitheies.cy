// Command migrate applies or rolls back database schema migrations.
package main

import (
	"errors"
	"flag"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	direction := flag.String("direction", "up", "up or down")
	steps := flag.Int("steps", 0, "number of steps (0 = all)")
	flag.Parse()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	m, err := migrate.New("file://migrations", dbURL)
	if err != nil {
		log.Fatalf("init migrate: %v", err)
	}
	defer m.Close()

	switch {
	case *steps != 0:
		if *direction == "down" {
			*steps = -*steps
		}
		err = m.Steps(*steps)
	case *direction == "down":
		err = m.Down()
	default:
		err = m.Up()
	}

	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatalf("migrate: %v", err)
	}
	log.Println("migrations applied successfully")
}
