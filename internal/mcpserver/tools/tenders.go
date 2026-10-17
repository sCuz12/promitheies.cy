package tools

import (
	"context"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/georgehadjisavvas/promitheies-cy/internal/web"
)

type searchTendersIn struct {
	Query         string `json:"query,omitempty" jsonschema:"free-text search over tender titles (Greek or English)"`
	AuthoritySlug string `json:"authority_slug,omitempty" jsonschema:"restrict to one contracting authority; get slugs from list_authorities"`
	CPVDivision   string `json:"cpv_division,omitempty" jsonschema:"2-digit CPV division, e.g. 45 for construction"`
	Limit         int    `json:"limit,omitempty" jsonschema:"max results (default 20, max 50)"`
}

type openTendersIn struct {
	Query       string `json:"query,omitempty" jsonschema:"free-text search over tender titles (Greek or English)"`
	CPVDivision string `json:"cpv_division,omitempty" jsonschema:"2-digit CPV division, e.g. 72 for IT services"`
	Limit       int    `json:"limit,omitempty" jsonschema:"max results (default 20, max 50)"`
}

type getTenderIn struct {
	ID int64 `json:"id" jsonschema:"tender id, as returned by search_tenders or list_open_tenders"`
}

type tendersOut struct {
	Tenders []web.TenderRow `json:"tenders"`
}

type tenderOut struct {
	Tender *web.Tender `json:"tender"`
}

func registerTenders(s *mcp.Server, pool *pgxpool.Pool) {
	readOnly := &mcp.ToolAnnotations{ReadOnlyHint: true}

	mcp.AddTool(s, &mcp.Tool{
		Name:        "search_tenders",
		Description: "Search all Cyprus public tenders (open and historical) by title, authority and CPV division. Newest first.",
		Annotations: readOnly,
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in searchTendersIn) (*mcp.CallToolResult, tendersOut, error) {
		rows, err := web.SearchTenders(ctx, pool, in.Query, in.AuthoritySlug, in.CPVDivision, clampLimit(in.Limit))
		return nil, tendersOut{Tenders: rows}, err
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_open_tenders",
		Description: "List currently open tender opportunities (deadline not yet passed) that businesses can still bid on. Newest first.",
		Annotations: readOnly,
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in openTendersIn) (*mcp.CallToolResult, tendersOut, error) {
		rows, err := web.ListOpenTenders(ctx, pool, in.Query, in.CPVDivision, clampLimit(in.Limit))
		return nil, tendersOut{Tenders: rows}, err
	})

	mcp.AddTool(s, &mcp.Tool{
		Name:        "get_tender",
		Description: "Get full details of one tender: authority, value, status, procedure type, deadline and source.",
		Annotations: readOnly,
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in getTenderIn) (*mcp.CallToolResult, tenderOut, error) {
		t, err := web.GetTender(ctx, pool, in.ID)
		if err != nil {
			return nil, tenderOut{}, notFound(err, "tender", strconv.FormatInt(in.ID, 10))
		}
		return nil, tenderOut{Tender: t}, nil
	})
}
