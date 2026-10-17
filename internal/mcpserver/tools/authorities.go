package tools

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/georgehadjisavvas/promitheies-cy/internal/web"
)

type listAuthoritiesIn struct {
	Limit int `json:"limit,omitempty" jsonschema:"max results (default 20, max 50); ordered by total awarded value"`
}

type getAuthorityIn struct {
	Slug string `json:"slug" jsonschema:"authority slug, as returned by list_authorities or any tender"`
}

type authoritiesOut struct {
	Authorities []web.EntitySummary `json:"authorities"`
}

type authorityOut struct {
	Authority *web.Authority `json:"authority"`
}

func registerAuthorities(s *mcp.Server, pool *pgxpool.Pool) {
	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_authorities",
		Title:       "List Authorities",
		Description: "List contracting authorities (ministries, municipalities, agencies) ranked by total awarded contract value.",
		Annotations: readOnly("List Authorities"),
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in listAuthoritiesIn) (*mcp.CallToolResult, authoritiesOut, error) {
		rows, err := web.ListAuthorities(ctx, pool, clampLimit(in.Limit))
		return nil, authoritiesOut{Authorities: rows}, err
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_authority",
		Title:       "Get Authority",
		Description: "Get a contracting authority's spending profile: totals, top contractors, spend by CPV sector and recent awards.",
		Annotations: readOnly("Get Authority"),
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in getAuthorityIn) (*mcp.CallToolResult, authorityOut, error) {
		a, err := web.GetAuthority(ctx, pool, in.Slug)
		if err != nil {
			return nil, authorityOut{}, notFound(err, "authority", in.Slug)
		}
		return nil, authorityOut{Authority: a}, nil
	})
}
