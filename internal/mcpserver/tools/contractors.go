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
	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_contractors",
		Title:       "List Contractors",
		Description: "List contractors that won public contracts, ranked by total awarded value.",
		Annotations: readOnly("List Contractors"),
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in listContractorsIn) (*mcp.CallToolResult, contractorsOut, error) {
		rows, err := web.ListContractors(ctx, pool, clampLimit(in.Limit))
		return nil, contractorsOut{Contractors: rows}, err
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_contractor",
		Title:       "Get Contractor",
		Description: "Get a contractor's profile: total won, which authorities award them contracts, and recent awards.",
		Annotations: readOnly("Get Contractor"),
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in getContractorIn) (*mcp.CallToolResult, contractorOut, error) {
		c, err := web.GetContractor(ctx, pool, in.Slug)
		if err != nil {
			return nil, contractorOut{}, notFound(err, "contractor", in.Slug)
		}
		return nil, contractorOut{Contractor: c}, nil
	})
}
