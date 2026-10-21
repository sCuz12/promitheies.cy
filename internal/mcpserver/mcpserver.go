// Package mcpserver exposes the procurement dataset to LLM clients over the
// Model Context Protocol (Streamable HTTP transport). The tool handlers live
// in the tools subpackage; this file only wires them into a server.
package mcpserver

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/georgehadjisavvas/promitheies-cy/internal/mcpserver/tools"
)

const version = "v0.1.0"

// Handler returns an http.Handler serving the MCP endpoint. Mount it at /mcp.
func Handler(pool *pgxpool.Pool) http.Handler {
	s := mcp.NewServer(&mcp.Implementation{
		Name:    "promitheies-cy",
		Version: version,
	}, &mcp.ServerOptions{
		Instructions: "Read-only access to Cyprus public procurement data: tenders, awards, " +
			"contracting authorities and contractors. Tender titles are mostly Greek. " +
			"CPV divisions are 2-digit codes (e.g. 45 = construction, 72 = IT services). " +
			"Monetary values are in EUR.",
	})
	tools.Register(s, pool)

	return mcp.NewStreamableHTTPHandler(
		func(*http.Request) *mcp.Server { return s },
		&mcp.StreamableHTTPOptions{
			Stateless:    true,
			JSONResponse: true,
		},
	)
}
