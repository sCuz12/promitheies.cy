// Package tools holds the MCP tool handlers, one file per domain area.
// Each handler is a thin adapter over the query functions in internal/web.
package tools

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	defaultLimit = 20
	maxLimit     = 50
)

// Register adds every tool to s.
func Register(s *mcp.Server, pool *pgxpool.Pool) {
	registerTenders(s, pool)
	registerAuthorities(s, pool)
	registerContractors(s, pool)
}

// clampLimit applies the default and upper bound to a client-supplied limit.
func clampLimit(n int) int {
	if n <= 0 {
		return defaultLimit
	}
	return min(n, maxLimit)
}

// notFound maps pgx.ErrNoRows to a friendly error; the SDK surfaces handler
// errors to the model as tool errors it can react to.
func notFound(err error, what, id string) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("%s %q not found", what, id)
	}
	return err
}

// readOnly returns the annotations shared by every tool: they only read data.
// The title is required for the Claude connectors directory listing.
func readOnly(title string) *mcp.ToolAnnotations {
	return &mcp.ToolAnnotations{Title: title, ReadOnlyHint: true}
}
