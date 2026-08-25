package entities

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// isGreek reports whether s contains Greek script characters, used to
// decide whether a freshly-seen name belongs in the _el or _en canonical
// column when a record is first created.
func isGreek(s string) bool {
	for _, r := range s {
		if (r >= 0x0370 && r <= 0x03FF) || (r >= 0x1F00 && r <= 0x1FFF) {
			return true
		}
	}
	return false
}

// ResolveAuthority finds the canonical authority ID for name, creating a
// new authority (and alias) if no existing alias matches. source records
// which importer first observed this specific spelling.
func ResolveAuthority(ctx context.Context, pool *pgxpool.Pool, name, source string) (int64, error) {
	normalized := NormalizeForMatching(name)
	if normalized == "" {
		return 0, fmt.Errorf("authority name %q normalizes to empty string", name)
	}

	var id int64
	err := pool.QueryRow(ctx,
		`SELECT authority_id FROM authority_aliases WHERE normalized_alias = $1`,
		normalized,
	).Scan(&id)
	if err == nil {
		return id, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return 0, fmt.Errorf("lookup authority alias: %w", err)
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// Re-check inside the transaction in case of a race with a concurrent importer.
	err = tx.QueryRow(ctx,
		`SELECT authority_id FROM authority_aliases WHERE normalized_alias = $1`,
		normalized,
	).Scan(&id)
	if err == nil {
		return id, tx.Commit(ctx)
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return 0, fmt.Errorf("lookup authority alias in tx: %w", err)
	}

	slug, err := uniqueSlug(ctx, tx, "authorities", name)
	if err != nil {
		return 0, err
	}

	var nameEN, nameEL *string
	if isGreek(name) {
		nameEL = &name
	} else {
		nameEN = &name
	}

	err = tx.QueryRow(ctx,
		`INSERT INTO authorities (canonical_name_en, canonical_name_el, slug) VALUES ($1, $2, $3) RETURNING id`,
		nameEN, nameEL, slug,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("insert authority: %w", err)
	}

	_, err = tx.Exec(ctx,
		`INSERT INTO authority_aliases (authority_id, alias, normalized_alias, source) VALUES ($1, $2, $3, $4)`,
		id, name, normalized, source,
	)
	if err != nil {
		return 0, fmt.Errorf("insert authority alias: %w", err)
	}

	return id, tx.Commit(ctx)
}

// ResolveContractor finds the canonical contractor ID for name, creating a
// new contractor (and alias) if no existing alias matches.
func ResolveContractor(ctx context.Context, pool *pgxpool.Pool, name, source string) (int64, error) {
	normalized := NormalizeContractorName(name)
	if normalized == "" {
		return 0, fmt.Errorf("contractor name %q normalizes to empty string", name)
	}

	var id int64
	err := pool.QueryRow(ctx,
		`SELECT contractor_id FROM contractor_aliases WHERE normalized_alias = $1`,
		normalized,
	).Scan(&id)
	if err == nil {
		return id, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return 0, fmt.Errorf("lookup contractor alias: %w", err)
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	err = tx.QueryRow(ctx,
		`SELECT contractor_id FROM contractor_aliases WHERE normalized_alias = $1`,
		normalized,
	).Scan(&id)
	if err == nil {
		return id, tx.Commit(ctx)
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return 0, fmt.Errorf("lookup contractor alias in tx: %w", err)
	}

	slug, err := uniqueSlug(ctx, tx, "contractors", name)
	if err != nil {
		return 0, err
	}

	err = tx.QueryRow(ctx,
		`INSERT INTO contractors (canonical_name, normalized_name, slug) VALUES ($1, $2, $3) RETURNING id`,
		name, normalized, slug,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("insert contractor: %w", err)
	}

	_, err = tx.Exec(ctx,
		`INSERT INTO contractor_aliases (contractor_id, alias, normalized_alias, source) VALUES ($1, $2, $3, $4)`,
		id, name, normalized, source,
	)
	if err != nil {
		return 0, fmt.Errorf("insert contractor alias: %w", err)
	}

	return id, tx.Commit(ctx)
}

// uniqueSlug generates a slug for name and appends "-2", "-3", ... until it
// finds one not already used in table.
func uniqueSlug(ctx context.Context, tx pgx.Tx, table, name string) (string, error) {
	base := Slugify(name)
	if base == "" {
		base = "entry"
	}
	slug := base
	for attempt := 2; ; attempt++ {
		var exists bool
		err := tx.QueryRow(ctx,
			fmt.Sprintf(`SELECT EXISTS(SELECT 1 FROM %s WHERE slug = $1)`, table),
			slug,
		).Scan(&exists)
		if err != nil {
			return "", fmt.Errorf("check slug uniqueness: %w", err)
		}
		if !exists {
			return slug, nil
		}
		slug = fmt.Sprintf("%s-%d", base, attempt)
	}
}
