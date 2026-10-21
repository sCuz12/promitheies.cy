package web

import (
	"net"
	"net/http"
)

// mcpParam documents one input of an MCP tool on the public docs page.
type mcpParam struct {
	Name     string
	Type     string
	Required bool
}

// mcpTool documents one MCP tool. The description is looked up in the
// translation table under "mcp.tool.<Name>", parameter text under
// "mcp.param.<param name>".
type mcpTool struct {
	Name   string
	Params []mcpParam
}

type mcpDocsData struct {
	URL      string
	Tools    []mcpTool
	Examples []string
}

// mcpToolDocs mirrors the tools registered in internal/mcpserver/tools; a test
// in internal/mcpserver fails if the two drift apart.
var mcpToolDocs = []mcpTool{
	{Name: "search_tenders", Params: []mcpParam{
		{"query", "string", false}, {"authority_slug", "string", false},
		{"cpv_division", "string", false}, {"limit", "integer", false},
	}},
	{Name: "list_open_tenders", Params: []mcpParam{
		{"query", "string", false}, {"cpv_division", "string", false}, {"limit", "integer", false},
	}},
	{Name: "get_tender", Params: []mcpParam{{"id", "integer", true}}},
	{Name: "list_authorities", Params: []mcpParam{{"limit", "integer", false}}},
	{Name: "get_authority", Params: []mcpParam{{"slug", "string", true}}},
	{Name: "list_contractors", Params: []mcpParam{{"limit", "integer", false}}},
	{Name: "get_contractor", Params: []mcpParam{{"slug", "string", true}}},
}

// MCPToolNames returns the documented tool names.
func MCPToolNames() []string {
	names := make([]string, len(mcpToolDocs))
	for i, t := range mcpToolDocs {
		names[i] = t.Name
	}
	return names
}

func (s *Server) handlePrivacy(w http.ResponseWriter, r *http.Request) {
	s.render(w, r, "privacy", nil)
}

func (s *Server) handleSupport(w http.ResponseWriter, r *http.Request) {
	s.render(w, r, "support", nil)
}

func (s *Server) handleTerms(w http.ResponseWriter, r *http.Request) {
	s.render(w, r, "terms", nil)
}

func (s *Server) handleMCPDocs(w http.ResponseWriter, r *http.Request) {
	s.render(w, r, "mcp_docs", mcpDocsData{
		URL:      requestBaseURL(r) + "/mcp",
		Tools:    mcpToolDocs,
		Examples: []string{"mcp.example.1", "mcp.example.2", "mcp.example.3", "mcp.example.4"},
	})
}

// requestBaseURL reconstructs scheme://host for display in connection
// snippets. Fly's proxy sets X-Forwarded-Proto; the value is only echoed back
// as documentation text, never used for any decision.
func requestBaseURL(r *http.Request) string {
	scheme := "https"
	if proto := r.Header.Get("X-Forwarded-Proto"); proto != "" {
		scheme = proto
	} else if r.TLS == nil {
		host, _, err := net.SplitHostPort(r.Host)
		if err != nil {
			host = r.Host
		}
		if host == "localhost" {
			scheme = "http"
		}
	}
	return scheme + "://" + r.Host
}
