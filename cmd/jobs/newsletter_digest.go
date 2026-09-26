package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/georgehadjisavvas/promitheies-cy/internal/digest"
	"github.com/georgehadjisavvas/promitheies-cy/internal/email"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mailtrap/mailtrap-go"
)

const (
	newsletterDigestSource = "newsletter-digest"
	newsletterWindowDays   = 100
	newsletterTenderLimit  = 200
)

// newMailerFromEnv builds the email.Mailer selected by EMAIL_PROVIDER
// (default "resend"), reading that provider's own credential env vars. To
// add another provider (Brevo, ...), give it its own client type in
// internal/email implementing email.Mailer, then add a case for it here.
// Missing credentials or an unrecognized provider are a hard failure: this
// job's entire purpose is sending email, so a silent no-op every week would
// be a worse failure mode than a loud cron failure.
func newMailerFromEnv(from string) email.Mailer {
	provider := os.Getenv("EMAIL_PROVIDER")
	if provider == "" {
		provider = "resend"
	}

	switch provider {
	case "resend":
		apiKey := os.Getenv("RESEND_API_KEY")
		if apiKey == "" {
			log.Fatal("RESEND_API_KEY is not set")
		}
		return email.NewResendClient(apiKey, from)
	case "mailtrap":
		apiKey := os.Getenv("MAILTRAP_API_TOKEN")
		if apiKey == "" {
			log.Fatal("MAILTRAP_API_TOKEN is not set")
		}
		mailer, err := email.NewMailtrapClient(apiKey, from,
			mailtrap.WithSandbox(true),
			mailtrap.WithSandboxID(923293))
		if err != nil {
			log.Fatalf("build mailtrap client: %v", err)
		}
		return mailer
	default:
		log.Fatalf("unsupported EMAIL_PROVIDER %q", provider)
		return nil
	}
}

// runNewsletterDigest sends each newsletter subscriber a weekly email
// digest of new open tenders matching their saved sector/budget filters.
// A subscriber with no matches gets no email.
func runNewsletterDigest(ctx context.Context, pool *pgxpool.Pool) error {
	from := os.Getenv("NEWSLETTER_FROM_EMAIL")
	if from == "" {
		log.Fatal("NEWSLETTER_FROM_EMAIL is not set")
	}
	publicBaseURL := os.Getenv("PUBLIC_BASE_URL")
	mailer := newMailerFromEnv(from)

	runID, err := startIngestRun(ctx, pool, newsletterDigestSource)
	if err != nil {
		return fmt.Errorf("start ingest run: %w", err)
	}

	subscribers, err := digest.LoadSubscribers(ctx, pool)
	if err != nil {
		finishIngestRun(ctx, pool, runID, 0, 0, "failed", err.Error())
		return fmt.Errorf("load subscribers: %w", err)
	}

	since := time.Now().AddDate(0, 0, -newsletterWindowDays)

	var processed, sent, failed int
	for _, sub := range subscribers {
		processed++

		tenders, err := digest.MatchingTenders(ctx, pool, sub, since, newsletterTenderLimit)
		if err != nil {
			log.Printf("warning: subscriber %d: match tenders: %v", sub.ID, err)
			failed++
			continue
		}
		if len(tenders) == 0 {
			continue
		}

		subject, htmlBody, textBody := digest.RenderDigest(sub.Language, tenders, publicBaseURL)
		if err := mailer.Send(ctx, sub.Email, subject, htmlBody, textBody); err != nil {
			log.Printf("warning: subscriber %d: send email: %v", sub.ID, err)
			failed++
			continue
		}

		ids := make([]int64, len(tenders))
		for i, t := range tenders {
			ids[i] = t.ID
		}
		if err := digest.RecordSent(ctx, pool, sub.ID, ids); err != nil {
			log.Printf("error: subscriber %d: sent but failed to record — risk of duplicate next run: %v", sub.ID, err)
			failed++
			continue
		}
		sent++
	}

	status := "success"
	errMsg := ""
	if failed > 0 {
		errMsg = fmt.Sprintf("%d of %d subscribers failed", failed, processed)
	}
	finishIngestRun(ctx, pool, runID, processed, sent, status, errMsg)
	log.Printf("newsletter-digest done: processed=%d sent=%d failed=%d", processed, sent, failed)
	return nil
}
