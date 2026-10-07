// Package email holds one client type per transactional-email provider
// (Resend today; Mailtrap, Brevo, etc. can be added the same way). Every
// client exposes the same Send(ctx, to, subject, html, text) error method,
// so callers can depend on that method set instead of a concrete provider
// and switch providers by swapping which client they construct.
package email

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

var _ Mailer = (*ResendClient)(nil)

// ResendClient sends transactional emails via the Resend API from a single
// pre-configured "from" address.
type ResendClient struct {
	baseURL string // overridable in tests; defaults to the real Resend API
	apiKey  string
	from    string
	http    *http.Client
}

// NewResendClient builds a ResendClient for the given Resend API key and
// "from" address.
func NewResendClient(apiKey, from string) *ResendClient {
	return &ResendClient{
		baseURL: "https://api.resend.com",
		apiKey:  apiKey,
		from:    from,
		http:    &http.Client{Timeout: 10 * time.Second},
	}
}

type resendSendRequest struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	HTML    string   `json:"html"`
	Text    string   `json:"text"`
}

type resendErrorResponse struct {
	Message string `json:"message"`
	Name    string `json:"name"`
}

// Send delivers one HTML email with a plaintext fallback to one recipient.
func (c *ResendClient) Send(ctx context.Context, to, subject, html, text string) error {
	body, err := json.Marshal(resendSendRequest{
		From:    c.from,
		To:      []string{to},
		Subject: subject,
		HTML:    html,
		Text:    text,
	})
	if err != nil {
		return fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/emails", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var out resendErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
			return fmt.Errorf("resend API error: status %d", resp.StatusCode)
		}
		return fmt.Errorf("resend API error: %s: %s", out.Name, out.Message)
	}
	return nil
}
