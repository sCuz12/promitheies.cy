package email

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

var _ Mailer = (*BrevoClient)(nil)

// BrevoClient sends transactional emails via the Brevo (formerly Sendinblue)
// API from a single pre-configured "from" address.
type BrevoClient struct {
	baseURL string // overridable in tests; defaults to the real Brevo API
	apiKey  string
	from    string
	http    *http.Client
}

// NewBrevoClient builds a BrevoClient for the given Brevo API key and "from"
// address.
func NewBrevoClient(apiKey, from string) *BrevoClient {
	return &BrevoClient{
		baseURL: "https://api.brevo.com/v3",
		apiKey:  apiKey,
		from:    from,
		http:    &http.Client{Timeout: 10 * time.Second},
	}
}

type brevoAddress struct {
	Email string `json:"email"`
}

type brevoSendRequest struct {
	Sender      brevoAddress   `json:"sender"`
	To          []brevoAddress `json:"to"`
	Subject     string         `json:"subject"`
	HTMLContent string         `json:"htmlContent"`
	TextContent string         `json:"textContent"`
}

type brevoErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Send delivers one HTML email with a plaintext fallback to one recipient.
func (c *BrevoClient) Send(ctx context.Context, to, subject, html, text string) error {
	body, err := json.Marshal(brevoSendRequest{
		Sender:      brevoAddress{Email: c.from},
		To:          []brevoAddress{{Email: to}},
		Subject:     subject,
		HTMLContent: html,
		TextContent: text,
	})
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/smtp/email", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("api-key", c.apiKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var out brevoErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
			return fmt.Errorf("brevo API error: status %d", resp.StatusCode)
		}
		return fmt.Errorf("brevo API error: %s: %s", out.Code, out.Message)
	}
	return nil
}
