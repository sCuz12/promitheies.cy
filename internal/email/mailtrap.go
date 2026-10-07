package email

import (
	"context"
	"fmt"

	"github.com/mailtrap/mailtrap-go"
)

var _ Mailer = (*MailtrapClient)(nil)

// MailtrapClient sends transactional emails via the Mailtrap Email Sending
// API (github.com/mailtrap/mailtrap-go) from a single pre-configured "from"
// address.
type MailtrapClient struct {
	client *mailtrap.Client
	from   string
}

// NewMailtrapClient builds a MailtrapClient for the given Mailtrap API token
// and "from" address. opts are passed through to mailtrap.NewClient, e.g.
// mailtrap.WithSandbox(true) plus mailtrap.WithSandboxID(id) to send into a
// sandbox inbox instead of delivering real email.
func NewMailtrapClient(apiKey, from string, opts ...mailtrap.Option) (*MailtrapClient, error) {
	client, err := mailtrap.NewClient(apiKey, opts...)
	if err != nil {
		return nil, fmt.Errorf("build mailtrap client: %w", err)
	}
	return &MailtrapClient{client: client, from: from}, nil
}

// Send delivers one HTML email with a plaintext fallback to one recipient.
func (c *MailtrapClient) Send(ctx context.Context, to, subject, html, text string) error {
	out, _, err := c.client.Send(ctx, &mailtrap.SendRequest{
		From:    mailtrap.Address{Email: c.from},
		To:      []mailtrap.Address{{Email: to}},
		Subject: subject,
		HTML:    html,
		Text:    text,
	})
	if err != nil {
		return fmt.Errorf("mailtrap: send: %w", err)
	}
	if !out.Success {
		return fmt.Errorf("mailtrap: send unsuccessful")
	}
	return nil
}
