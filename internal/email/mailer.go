package email

import "context"

// Mailer delivers one HTML email with a plaintext fallback to a single
// recipient. Every provider client in this package (ResendClient,
// MailtrapClient, ...) implements it, so callers can depend on Mailer
// instead of a concrete provider and switch providers without changing
// anything downstream of construction.
type Mailer interface {
	Send(ctx context.Context, to, subject, html, text string) error
}
