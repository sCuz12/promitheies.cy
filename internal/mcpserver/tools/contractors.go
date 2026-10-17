package tools

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/georgehadjisavvas/promitheies-cy/internal/web"
)

type listContractorsIn struct {
	Limit int `json:"limit,omitempty" jsonschema:"max results (default 20, max 50); ordered by total awarded value"`
}

type getContractorIn struct {
	Slug string `json:"slug" jsonschema:"contractor slug, as returned by list_contractors or an award row"`
}

type contractorsOut struct {
	Contractors []web.EntitySummary `json:"contractors"`
}

type contractorOut struct {
	Contractor *web.Contractor `json:"contractor"`
}

func registerContractors(s *mcp.Server, pool *pgxpool.Pool) {
	readOnly := &mcp.ToolAnnotations{ReadOnlyHint: true}

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_contractors",
		Description: "List contractors that won public contracts, ranked by total awarded value.",
		Annotations: readOnly,
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in listContractorsIn) (*mcp.CallToolResult, contractorsOut, error) {
		rows, err := web.ListContractors(ctx, pool, clampLimit(in.Limit))
		return nil, contractorsOut{Contractors: rows}, err
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_contractor",
		Description: "Get a contractor's profile: total won, which authorities award them contracts, and recent awards.",
		Annotations: readOnly,
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in getContractorIn) (*mcp.CallToolResult, contractorOut, error) {
		c, err := web.GetContractor(ctx, pool, in.Slug)
		if err != nil {
			return nil, contractorOut{}, notFound(err, "contractor", in.Slug)
		}
		return nil, contractorOut{Contractor: c}, nil
	})
}
