// Package digest matches newly-ingested open tenders against each
// newsletter subscriber's saved sector/budget filters, and tracks which
// tenders have already been sent to which subscriber so the weekly digest
// job never repeats itself.
package digest

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Subscriber is one row of newsletter_subscribers, including the sector and
// budget filters captured by the onboarding wizard.
type Subscriber struct {
	ID           int64
	Email        string
	Language     string // "el" | "en"
	CPVDivisions []string
	MinValue     *float64
}

// Tender is the subset of a tenders row needed to match against a
// subscriber's filters and render a digest email.
type Tender struct {
	ID                int64
	Title             string
	AuthorityName     string
	CPVDivision       *string
	EstimatedValueEUR *float64
	Deadline          *time.Time
	ExternalID        string // TED publication number
}

// LoadSubscribers returns every newsletter subscriber, in a stable order.
func LoadSubscribers(ctx context.Context, pool *pgxpool.Pool) ([]Subscriber, error) {
	rows, err := pool.Query(ctx, `
		SELECT id, email, language, cpv_divisions, min_value
		FROM newsletter_subscribers
		ORDER BY id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subs []Subscriber
	for rows.Next() {
		var s Subscriber
		if err := rows.Scan(&s.ID, &s.Email, &s.Language, &s.CPVDivisions, &s.MinValue); err != nil {
			return nil, err
		}
		subs = append(subs, s)
	}
	return subs, rows.Err()
}

// MatchingTenders returns open TED tenders created on or after since that
// match sub's saved CPV divisions and minimum value, excluding any tender
// already recorded as sent to sub (see RecordSent). A tender with an
// unknown estimated value passes the min-value filter rather than being
// hidden.
func MatchingTenders(ctx context.Context, pool *pgxpool.Pool, sub Subscriber, since time.Time, limit int) ([]Tender, error) {
	rows, err := pool.Query(ctx, `
		SELECT
			t.id,
			coalesce(t.title_el, t.title_en, ''),
			coalesce(auth.canonical_name_el, auth.canonical_name_en, ''),
			t.cpv_division,
			t.estimated_value,
			t.deadline,
			t.external_ids->>'ted'
		FROM tenders t
		JOIN authorities auth ON auth.id = t.authority_id
		WHERE t.source = 'ted'
		  AND t.status = 'open'
		  AND (t.deadline IS NULL OR t.deadline >= CURRENT_DATE)
		  AND t.created_at >= $1
		  AND (cardinality($2::text[]) = 0 OR t.cpv_division = ANY($2::text[]))
		  AND ($3::numeric IS NULL OR t.estimated_value IS NULL OR t.estimated_value >= $3)
		  AND NOT EXISTS (
		      SELECT 1 FROM newsletter_sends ns
		      WHERE ns.subscriber_id = $4 AND ns.tender_id = t.id
		  )
		ORDER BY t.created_at DESC, t.id DESC
		LIMIT $5
	`, since, sub.CPVDivisions, sub.MinValue, sub.ID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tenders []Tender
	for rows.Next() {
		var t Tender
		if err := rows.Scan(
			&t.ID,
			&t.Title,
			&t.AuthorityName,
			&t.CPVDivision,
			&t.EstimatedValueEUR,
			&t.Deadline,
			&t.ExternalID,
		); err != nil {
			return nil, err
		}
		tenders = append(tenders, t)
	}
	return tenders, rows.Err()
}

// RecordSent marks tenderIDs as sent to subscriberID, so a future run's
// NOT EXISTS anti-join in MatchingTenders excludes them. Safe to call more
// than once for the same pair (idempotent via ON CONFLICT DO NOTHING).
func RecordSent(ctx context.Context, pool *pgxpool.Pool, subscriberID int64, tenderIDs []int64) error {
	if len(tenderIDs) == 0 {
		return nil
	}
	_, err := pool.Exec(ctx, `
		INSERT INTO newsletter_sends (subscriber_id, tender_id)
		SELECT $1, unnest($2::bigint[])
		ON CONFLICT (subscriber_id, tender_id) DO NOTHING
	`, subscriberID, tenderIDs)
	return err
}
