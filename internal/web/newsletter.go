package web

import (
	"context"
	"fmt"
	"net/http"
	"net/mail"
	"strings"

	"github.com/georgehadjisavvas/promitheies-cy/internal/cpv"
	"github.com/jackc/pgx/v5/pgxpool"
)

const maxNewsletterFormBytes = 8 << 10

// allowedMinValues is the fixed set of options the wizard's budget-range
// select offers. Validated as an allow-list rather than parsed as a free
// float, since the value arrives as untrusted form input.
var allowedMinValues = map[string]float64{
	"25000":  25000,
	"140000": 140000,
	"500000": 500000,
}

// sectorKeysFromForm returns the sector keys the wizard submitted that are
// recognized CPV sectors, dropping anything else (unknown values, injected
// garbage) silently.
func sectorKeysFromForm(raw []string) []string {
	var keys []string
	for _, k := range raw {
		for _, sec := range cpv.Sectors {
			if sec.Key == k {
				keys = append(keys, k)
				break
			}
		}
	}
	return keys
}

// parseMinValue reads the wizard's budget-range field, returning nil for
// "any" or an unrecognized value.
func parseMinValue(raw string) *float64 {
	v, ok := allowedMinValues[strings.TrimSpace(raw)]
	if !ok {
		return nil
	}
	return &v
}

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

// StoreNewsletterSubscriber inserts an address along with the onboarding
// wizard's sector/budget answers. Signing up again with the same address
// updates those preferences in place (the wizard is how a subscriber changes
// their mind about what to be notified on) but always returns the same
// success response either way, which avoids disclosing whether an address
// was already registered.
func StoreNewsletterSubscriber(ctx context.Context, pool *pgxpool.Pool, email string, lang Lang, cpvDivisions []string, minValue *float64) error {
	if lang != LangEN {
		lang = LangEL
	}
	_, err := pool.Exec(ctx, `
		INSERT INTO newsletter_subscribers (email, language, cpv_divisions, min_value)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (email) DO UPDATE SET
			language      = EXCLUDED.language,
			cpv_divisions = EXCLUDED.cpv_divisions,
			min_value     = EXCLUDED.min_value
	`, email, lang, cpvDivisions, minValue)
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

	cpvDivisions := cpv.DivisionsForSectorKeys(sectorKeysFromForm(r.Form["sector"]))
	minValue := parseMinValue(r.FormValue("min_value"))

	if err := s.storeNewsletterSub(r.Context(), email, s.resolveLang(w, r), cpvDivisions, minValue); err != nil {
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
