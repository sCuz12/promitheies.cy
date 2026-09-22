package web

import (
	"context"
	"fmt"
	"net/http"
	"net/mail"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

const maxNewsletterFormBytes = 8 << 10

// normalizeSubscriberEmail accepts a single mailbox, strips harmless outer
// whitespace, and stores it in lowercase so repeat signups are idempotent.
func normalizeSubscriberEmail(raw string) (string, bool) {
	raw = strings.TrimSpace(raw)
	if raw == "" || len(raw) > 254 || strings.ContainsAny(raw, " \t\r\n") {
		return "", false
	}
	parsed, err := mail.ParseAddress(raw)
	if err != nil || !strings.EqualFold(parsed.Address, raw) {
		return "", false
	}
	email := strings.ToLower(parsed.Address)
	at := strings.LastIndexByte(email, '@')
	if at < 1 || at == len(email)-1 || !strings.Contains(email[at+1:], ".") {
		return "", false
	}
	return email, true
}

// StoreNewsletterSubscriber inserts an address once. Signing up again is a
// successful no-op, which avoids disclosing whether an address is registered.
func StoreNewsletterSubscriber(ctx context.Context, pool *pgxpool.Pool, email string, lang Lang) error {
	if lang != LangEN {
		lang = LangEL
	}
	_, err := pool.Exec(ctx, `
		INSERT INTO newsletter_subscribers (email, language)
		VALUES ($1, $2)
		ON CONFLICT (email) DO NOTHING
	`, email, lang)
	if err != nil {
		return fmt.Errorf("store newsletter subscriber: %w", err)
	}
	return nil
}

func (s *Server) handleNewsletterSubscribe(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxNewsletterFormBytes)
	if err := r.ParseForm(); err != nil {
		redirectNewsletter(w, r, false)
		return
	}

	// Bots tend to fill this visually-hidden field. Return the same success
	// response without writing anything so the field cannot be used as an
	// oracle for bypass attempts.
	if strings.TrimSpace(r.FormValue("website")) != "" {
		redirectNewsletter(w, r, true)
		return
	}

	email, ok := normalizeSubscriberEmail(r.FormValue("email"))
	if !ok {
		redirectNewsletter(w, r, false)
		return
	}
	if s.storeNewsletterSub == nil {
		s.serverError(w, fmt.Errorf("newsletter subscriber store is not configured"))
		return
	}
	if err := s.storeNewsletterSub(r.Context(), email, s.resolveLang(w, r)); err != nil {
		s.serverError(w, err)
		return
	}
	redirectNewsletter(w, r, true)
}

func redirectNewsletter(w http.ResponseWriter, r *http.Request, success bool) {
	result := "invalid"
	if success {
		result = "joined"
	}
	http.Redirect(w, r, "/?newsletter="+result+"#weekly-digest", http.StatusSeeOther)
}
