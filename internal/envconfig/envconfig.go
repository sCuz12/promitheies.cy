// Package envconfig loads local environment configuration for commands.
package envconfig

import (
	"errors"
	"os"

	"github.com/joho/godotenv"
)

// Load reads a local .env file when it exists. Values already supplied by the
// process environment take precedence, so deployment-provided credentials are
// never replaced by local development values.
func Load() error {
	err := godotenv.Load()
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}
