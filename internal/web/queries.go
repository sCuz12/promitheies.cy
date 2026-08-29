// Package web serves the free transparency dashboard: server-rendered
// html/template pages backed by direct SQL queries over the tenders/awards
// data ingested by the datagovcy and ted importers.
package web

import (
	"context"
	"fmt"

	"github.com/georgehadjisavvas/promitheies-cy/internal/cpv"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Stats struct {
	TotalTenders int64
	TotalAwards  int64
	TotalValue   float64
	BySource     []SourceCount
	RecentAwards []AwardRow
}

type SourceCount struct {
	Source string
	Count  int64
}

// AwardRow is one row in an award listing (dashboard, authority page,
// contractor page): a tender title paired with who won it, for how much.
type AwardRow struct {
	TenderID       int64
	TenderTitle    string
	AuthorityID    int64
	AuthoritySlug  string
	AuthorityName  string
	ContractorID   int64
	ContractorSlug string
	ContractorName string
	Value          *float64
	AwardDate      *string
	CPVDivision    *string
}

type EntitySummary struct {
	ID    int64
	Slug  string
	Name  string
	Total float64
	Count int64
}

func GetStats(ctx context.Context, pool *pgxpool.Pool) (Stats, error) {
	var s Stats

	if err := pool.QueryRow(ctx, `SELECT count(*) FROM tenders`).Scan(&s.TotalTenders); err != nil {
		return s, fmt.Errorf("count tenders: %w", err)
	}
	if err := pool.QueryRow(ctx, `SELECT count(*), coalesce(sum(value), 0) FROM awards`).Scan(&s.TotalAwards, &s.TotalValue); err != nil {
		return s, fmt.Errorf("count awards: %w", err)
	}

	rows, err := pool.Query(ctx, `SELECT source, count(*) FROM tenders GROUP BY source ORDER BY source`)
	if err != nil {
		return s, fmt.Errorf("count by source: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var sc SourceCount
		if err := rows.Scan(&sc.Source, &sc.Count); err != nil {
			return s, err
		}
		s.BySource = append(s.BySource, sc)
	}

	s.RecentAwards, err = recentAwards(ctx, pool, 20, nil, nil)
	if err != nil {
		return s, fmt.Errorf("recent awards: %w", err)
	}

	return s, nil
}

// recentAwards lists the most recent awards, optionally filtered to one
// authority or one contractor (mutually exclusive; pass nil for both to
// filter neither).
func recentAwards(ctx context.Context, pool *pgxpool.Pool, limit int, authorityID, contractorID *int64) ([]AwardRow, error) {
	query := `
		SELECT
			t.id, coalesce(t.title_el, t.title_en, ''), t.cpv_division,
			auth.id, auth.slug, coalesce(auth.canonical_name_el, auth.canonical_name_en, ''),
			c.id, c.slug, c.canonical_name,
			a.value, to_char(a.award_date, 'YYYY-MM-DD')
		FROM awards a
		JOIN tenders t ON t.id = a.tender_id
		JOIN authorities auth ON auth.id = t.authority_id
		JOIN contractors c ON c.id = a.contractor_id
		WHERE ($2::bigint IS NULL OR auth.id = $2)
		  AND ($3::bigint IS NULL OR c.id = $3)
		ORDER BY a.award_date DESC NULLS LAST, a.id DESC
		LIMIT $1
	`
	rows, err := pool.Query(ctx, query, limit, authorityID, contractorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []AwardRow
	for rows.Next() {
		var r AwardRow
		if err := rows.Scan(
			&r.TenderID, &r.TenderTitle, &r.CPVDivision,
			&r.AuthorityID, &r.AuthoritySlug, &r.AuthorityName,
			&r.ContractorID, &r.ContractorSlug, &r.ContractorName,
			&r.Value, &r.AwardDate,
		); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// ListAuthorities returns every authority with at least one award, ordered
// by total award value descending.
func ListAuthorities(ctx context.Context, pool *pgxpool.Pool) ([]EntitySummary, error) {
	rows, err := pool.Query(ctx, `
		SELECT auth.id, auth.slug, coalesce(auth.canonical_name_el, auth.canonical_name_en, ''),
		       coalesce(sum(a.value), 0), count(a.id)
		FROM authorities auth
		JOIN tenders t ON t.authority_id = auth.id
		JOIN awards a ON a.tender_id = t.id
		GROUP BY auth.id
		ORDER BY sum(a.value) DESC NULLS LAST
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanSummaries(rows)
}

// ListContractors returns every contractor with at least one award, ordered
// by total award value descending, capped at limit (the full list can be
// large — this backs both the top-N dashboard section and the paginated
// listing page).
func ListContractors(ctx context.Context, pool *pgxpool.Pool, limit int) ([]EntitySummary, error) {
	rows, err := pool.Query(ctx, `
		SELECT c.id, c.slug, c.canonical_name, coalesce(sum(a.value), 0), count(a.id)
		FROM contractors c
		JOIN awards a ON a.contractor_id = c.id
		GROUP BY c.id
		ORDER BY sum(a.value) DESC NULLS LAST
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanSummaries(rows)
}

func scanSummaries(rows interface {
	Next() bool
	Scan(...any) error
	Err() error
}) ([]EntitySummary, error) {
	var out []EntitySummary
	for rows.Next() {
		var e EntitySummary
		if err := rows.Scan(&e.ID, &e.Slug, &e.Name, &e.Total, &e.Count); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

type SectorBreakdown struct {
	CPVDivision string
	Total       float64
	Count       int64
}

type Authority struct {
	EntitySummary
	Type            string
	Region          string
	TopContractors  []EntitySummary
	SectorBreakdown []SectorBreakdown
	RecentAwards    []AwardRow
}

func GetAuthority(ctx context.Context, pool *pgxpool.Pool, slug string) (*Authority, error) {
	var a Authority
	err := pool.QueryRow(ctx, `
		SELECT auth.id, auth.slug, coalesce(auth.canonical_name_el, auth.canonical_name_en, ''), auth.type, auth.region,
		       coalesce(sum(a.value), 0), count(a.id)
		FROM authorities auth
		LEFT JOIN tenders t ON t.authority_id = auth.id
		LEFT JOIN awards a ON a.tender_id = t.id
		WHERE auth.slug = $1
		GROUP BY auth.id
	`, slug).Scan(&a.ID, &a.Slug, &a.Name, &a.Type, &a.Region, &a.Total, &a.Count)
	if err != nil {
		return nil, err
	}

	contractorRows, err := pool.Query(ctx, `
		SELECT c.id, c.slug, c.canonical_name, sum(a.value), count(a.id)
		FROM awards a
		JOIN tenders t ON t.id = a.tender_id
		JOIN contractors c ON c.id = a.contractor_id
		WHERE t.authority_id = $1
		GROUP BY c.id
		ORDER BY sum(a.value) DESC NULLS LAST
		LIMIT 10
	`, a.ID)
	if err != nil {
		return nil, fmt.Errorf("top contractors: %w", err)
	}
	a.TopContractors, err = scanSummaries(contractorRows)
	contractorRows.Close()
	if err != nil {
		return nil, err
	}

	sectorRows, err := pool.Query(ctx, `
		SELECT coalesce(t.cpv_division, 'unknown'), sum(a.value), count(a.id)
		FROM awards a
		JOIN tenders t ON t.id = a.tender_id
		WHERE t.authority_id = $1
		GROUP BY t.cpv_division
		ORDER BY sum(a.value) DESC NULLS LAST
	`, a.ID)
	if err != nil {
		return nil, fmt.Errorf("sector breakdown: %w", err)
	}
	defer sectorRows.Close()
	for sectorRows.Next() {
		var sb SectorBreakdown
		if err := sectorRows.Scan(&sb.CPVDivision, &sb.Total, &sb.Count); err != nil {
			return nil, err
		}
		a.SectorBreakdown = append(a.SectorBreakdown, sb)
	}

	a.RecentAwards, err = recentAwards(ctx, pool, 30, &a.ID, nil)
	if err != nil {
		return nil, fmt.Errorf("recent awards: %w", err)
	}

	return &a, nil
}

type Contractor struct {
	EntitySummary
	Authorities  []EntitySummary
	RecentAwards []AwardRow
}

func GetContractor(ctx context.Context, pool *pgxpool.Pool, slug string) (*Contractor, error) {
	var c Contractor
	err := pool.QueryRow(ctx, `
		SELECT c.id, c.slug, c.canonical_name, coalesce(sum(a.value), 0), count(a.id)
		FROM contractors c
		LEFT JOIN awards a ON a.contractor_id = c.id
		WHERE c.slug = $1
		GROUP BY c.id
	`, slug).Scan(&c.ID, &c.Slug, &c.Name, &c.Total, &c.Count)
	if err != nil {
		return nil, err
	}

	authRows, err := pool.Query(ctx, `
		SELECT auth.id, auth.slug, coalesce(auth.canonical_name_el, auth.canonical_name_en, ''), sum(a.value), count(a.id)
		FROM awards a
		JOIN tenders t ON t.id = a.tender_id
		JOIN authorities auth ON auth.id = t.authority_id
		WHERE a.contractor_id = $1
		GROUP BY auth.id
		ORDER BY sum(a.value) DESC NULLS LAST
	`, c.ID)
	if err != nil {
		return nil, fmt.Errorf("authorities: %w", err)
	}
	c.Authorities, err = scanSummaries(authRows)
	authRows.Close()
	if err != nil {
		return nil, err
	}

	c.RecentAwards, err = recentAwards(ctx, pool, 30, nil, &c.ID)
	if err != nil {
		return nil, fmt.Errorf("recent awards: %w", err)
	}

	return &c, nil
}

type SearchResult struct {
	Query   string
	Tenders []TenderRow
}

type OpenTendersPageData struct {
	Query       string
	CPVDivision string
	Divisions   []cpv.Division
	LastUpdated *string
	Results     []TenderRow
}

type TenderRow struct {
	ID            int64
	Slug          string
	Title         string
	AuthoritySlug string
	AuthorityName string
	CPVDivision   *string
	EstimatedVal  *float64
	Status        string
	PublishedAt   *string
	Deadline      *string
	Source        string
	ExternalID    *string
}

// SearchTenders does a full-text search (matching the datagovcy/ted title
// FTS index) over tender titles, optionally narrowed by authority slug and
// CPV division prefix.
func SearchTenders(ctx context.Context, pool *pgxpool.Pool, q, authoritySlug, cpvDivision string, limit int) ([]TenderRow, error) {
	rows, err := pool.Query(ctx, `
		SELECT t.id, coalesce(t.title_el, t.title_en, ''), auth.slug, coalesce(auth.canonical_name_el, auth.canonical_name_en, ''),
		       t.cpv_division, t.estimated_value, t.status, to_char(t.published_at, 'YYYY-MM-DD'), to_char(t.deadline, 'YYYY-MM-DD'), t.source, NULL::text
		FROM tenders t
		JOIN authorities auth ON auth.id = t.authority_id
		WHERE ($1 = '' OR to_tsvector('simple', coalesce(t.title_en,'') || ' ' || coalesce(t.title_el,'')) @@ plainto_tsquery('simple', $1))
		  AND ($2 = '' OR auth.slug = $2)
		  AND ($3 = '' OR t.cpv_division = $3)
		ORDER BY t.published_at DESC NULLS LAST
		LIMIT $4
	`, q, authoritySlug, cpvDivision, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []TenderRow
	for rows.Next() {
		var r TenderRow
		var authoritySlug string
		if err := rows.Scan(&r.ID, &r.Title, &authoritySlug, &r.AuthorityName, &r.CPVDivision, &r.EstimatedVal, &r.Status, &r.PublishedAt, &r.Deadline, &r.Source, &r.ExternalID); err != nil {
			return nil, err
		}
		r.AuthoritySlug = authoritySlug
		out = append(out, r)
	}
	return out, rows.Err()
}

// ListOpenTenders returns the TED-backed open tender feed shown to users.
// data.gov.cy is intentionally excluded here: it is an awarded-contracts
// dataset, while TED is our current source for live opportunity notices.
func ListOpenTenders(ctx context.Context, pool *pgxpool.Pool, q, cpvDivision string, limit int) ([]TenderRow, error) {
	rows, err := pool.Query(ctx, `
		SELECT t.id, coalesce(t.title_el, t.title_en, ''), auth.slug, coalesce(auth.canonical_name_el, auth.canonical_name_en, ''),
		       t.cpv_division, t.estimated_value, t.status, to_char(t.published_at, 'YYYY-MM-DD'), to_char(t.deadline, 'YYYY-MM-DD'),
		       t.source, t.external_ids->>'ted'
		FROM tenders t
		JOIN authorities auth ON auth.id = t.authority_id
		WHERE t.source = 'ted'
		  AND t.status = 'open'
		  AND (t.deadline IS NULL OR t.deadline >= CURRENT_DATE)
		  AND ($1 = '' OR to_tsvector('simple', coalesce(t.title_en,'') || ' ' || coalesce(t.title_el,'')) @@ plainto_tsquery('simple', $1))
		  AND ($2 = '' OR t.cpv_division = $2)
		ORDER BY t.published_at DESC NULLS LAST, t.id DESC
		LIMIT $3
	`, q, cpvDivision, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []TenderRow
	for rows.Next() {
		var r TenderRow
		if err := rows.Scan(
			&r.ID, &r.Title, &r.AuthoritySlug, &r.AuthorityName,
			&r.CPVDivision, &r.EstimatedVal, &r.Status, &r.PublishedAt,
			&r.Deadline, &r.Source, &r.ExternalID,
		); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func LastSuccessfulIngest(ctx context.Context, pool *pgxpool.Pool, source string) (*string, error) {
	var finishedAt *string
	err := pool.QueryRow(ctx, `
		SELECT to_char(max(finished_at), 'YYYY-MM-DD HH24:MI')
		FROM ingest_runs
		WHERE source = $1
		  AND status = 'success'
	`, source).Scan(&finishedAt)
	if err != nil {
		return nil, err
	}
	return finishedAt, nil
}
