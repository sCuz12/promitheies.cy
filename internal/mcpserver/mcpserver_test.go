package mcpserver_test

import (
	"context"
	"net/http/httptest"
	"slices"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/georgehadjisavvas/promitheies-cy/internal/web"

	"github.com/georgehadjisavvas/promitheies-cy/internal/mcpserver"
)

// TestListTools checks the endpoint speaks Streamable HTTP and advertises
// every tool. It needs no database: listing tools never touches the pool.
func TestListTools(t *testing.T) {
	ts := httptest.NewServer(mcpserver.Handler(nil))
	defer ts.Close()

	ctx := context.Background()
	client := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "0"}, nil)
	session, err := client.Connect(ctx, &mcp.StreamableClientTransport{Endpoint: ts.URL}, nil)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	defer session.Close()

	res, err := session.ListTools(ctx, nil)
	if err != nil {
		t.Fatalf("list tools: %v", err)
	}
	var got []string
	for _, tool := range res.Tools {
		got = append(got, tool.Name)
		// The Claude connectors directory rejects tools without these.
		if tool.Annotations == nil || tool.Annotations.Title == "" || !tool.Annotations.ReadOnlyHint {
			t.Errorf("tool %s: want annotations.title and readOnlyHint set, got %+v", tool.Name, tool.Annotations)
		}
	}
	slices.Sort(got)
	// The public docs page (internal/web/mcp_docs.go) lists tools by hand;
	// fail if it drifts from what the server really registers.
	docs := web.MCPToolNames()
	slices.Sort(docs)
	if !slices.Equal(got, docs) {
		t.Fatalf("docs page tools = %v, server tools = %v", docs, got)
	}
	want := []string{"get_authority", "get_contractor", "get_tender", "list_authorities", "list_contractors", "list_open_tenders", "search_tenders"}
	if !slices.Equal(got, want) {
		t.Fatalf("tools = %v, want %v", got, want)
	}
}
