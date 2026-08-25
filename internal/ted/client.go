// Package ted imports Cyprus notices from the EU's TED (Tenders Electronic
// Daily) Search API — a public, unauthenticated REST/JSON API, used in
// preference to the FTP/XML bulk feed for simplicity. Note: the search
// endpoint's flat field projection does not reliably surface award winner
// name or award value for "result" notices (verified against live data —
// zero of 250 sampled Cyprus result notices had either populated), so this
// importer captures tender data (title, buyer, CPV, estimated value,
// status) but does not create awards rows from TED; award data comes from
// the datagovcy importer, which does expose it.
package ted

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// MaxPageLimit is the TED Search API's documented maximum for the "limit"
// request parameter.
const MaxPageLimit = 250

type Client struct {
	baseURL string
	http    *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		http:    &http.Client{Timeout: 30 * time.Second},
	}
}

type searchRequest struct {
	Query  string   `json:"query"`
	Fields []string `json:"fields"`
	Page   int      `json:"page"`
	Limit  int      `json:"limit"`
}

// SearchResponse is the subset of the API's response we use.
type SearchResponse struct {
	Notices          []json.RawMessage `json:"notices"`
	TotalNoticeCount int               `json:"totalNoticeCount"`
}

// searchRetries bounds retries against transient failures (the API has
// been observed to occasionally return a gateway HTML error page instead
// of JSON).
const searchRetries = 3

// Search runs one page of a notice search. query uses TED's expert search
// syntax (e.g. "place-of-performance=CYP AND form-type=competition").
func (c *Client) Search(ctx context.Context, query string, fields []string, page, limit int) (*SearchResponse, error) {
	if limit > MaxPageLimit {
		limit = MaxPageLimit
	}

	var lastErr error
	for attempt := 1; attempt <= searchRetries; attempt++ {
		resp, err := c.doSearch(ctx, query, fields, page, limit)
		if err == nil {
			return resp, nil
		}
		lastErr = err
		if attempt < searchRetries {
			time.Sleep(time.Duration(attempt) * 2 * time.Second)
		}
	}
	return nil, lastErr
}

func (c *Client) doSearch(ctx context.Context, query string, fields []string, page, limit int) (*SearchResponse, error) {
	body, err := json.Marshal(searchRequest{Query: query, Fields: fields, Page: page, Limit: limit})
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/notices/search", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		SearchResponse
		Message string `json:"message"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response (status %d): %w", resp.StatusCode, err)
	}
	if result.Message != "" {
		return nil, fmt.Errorf("ted api error: %s", result.Message)
	}

	return &result.SearchResponse, nil
}
