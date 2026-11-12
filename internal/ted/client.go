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
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

// MaxPageLimit is the TED Search API's documented maximum for the "limit"
// request parameter.
const MaxPageLimit = 250

// userAgent identifies the importer to TED's CDN/WAF; Go's default agent
// is more likely to be challenged.
const userAgent = "promitheies.cy-ingest/1.0 (+https://promitheies.cy)"

type Client struct {
	baseURL string
	http    *http.Client
	// backoff returns the wait before retry number attempt (1-based).
	backoff func(attempt int) time.Duration
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		http:    &http.Client{Timeout: 30 * time.Second},
		backoff: func(attempt int) time.Duration { return time.Duration(attempt*attempt) * 5 * time.Second },
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
// been observed to occasionally return a gateway HTML error page, and its
// CDN may answer 403/429 when throttling or blocking an egress IP).
const searchRetries = 4

// httpError is a non-success HTTP response, with enough context
// (status, CDN headers, body snippet) to diagnose blocks without a shell.
type httpError struct {
	Status  int
	Headers string
	Body    string
}

func (e *httpError) Error() string {
	return fmt.Sprintf("ted api http %d (%s): %q", e.Status, e.Headers, e.Body)
}

// retryable reports whether err is worth retrying.
func retryable(err error) bool {
	var he *httpError
	if errors.As(err, &he) {
		return he.Status == http.StatusForbidden || he.Status == http.StatusTooManyRequests || he.Status >= 500
	}
	return !errors.Is(err, context.Canceled)
}

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
		if !retryable(err) || attempt == searchRetries {
			break
		}
		log.Printf("ted search page %d attempt %d/%d failed: %v", page, attempt, searchRetries, err)
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(c.backoff(attempt)):
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
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", userAgent)

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(io.LimitReader(resp.Body, 64<<20))
	if err != nil {
		return nil, fmt.Errorf("read response (status %d): %w", resp.StatusCode, err)
	}

	var result struct {
		SearchResponse
		Message string `json:"message"`
	}
	jsonErr := json.Unmarshal(raw, &result)

	if resp.StatusCode != http.StatusOK && jsonErr != nil {
		return nil, &httpError{Status: resp.StatusCode, Headers: diagHeaders(resp.Header), Body: snippet(raw)}
	}
	if jsonErr != nil {
		return nil, fmt.Errorf("decode response (status %d): %w (headers: %s, body: %q)",
			resp.StatusCode, jsonErr, diagHeaders(resp.Header), snippet(raw))
	}
	if result.Message != "" {
		return nil, &httpError{Status: resp.StatusCode, Headers: diagHeaders(resp.Header), Body: result.Message}
	}
	if resp.StatusCode != http.StatusOK {
		return nil, &httpError{Status: resp.StatusCode, Headers: diagHeaders(resp.Header), Body: snippet(raw)}
	}

	return &result.SearchResponse, nil
}

// diagHeaders summarises response headers useful for identifying who
// answered (CloudFront request id, server, content type).
func diagHeaders(h http.Header) string {
	var parts []string
	for _, k := range []string{"Content-Type", "Server", "Via", "X-Cache", "X-Amz-Cf-Id", "Retry-After"} {
		if v := h.Get(k); v != "" {
			parts = append(parts, k+"="+v)
		}
	}
	return strings.Join(parts, "; ")
}

// snippet returns the first 300 bytes of body with whitespace collapsed.
func snippet(b []byte) string {
	if len(b) > 300 {
		b = b[:300]
	}
	return strings.Join(strings.Fields(string(b)), " ")
}
