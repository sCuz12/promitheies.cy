package ted

import (
	"context"
	"fmt"

	"github.com/georgehadjisavvas/promitheies-cy/internal/cpv"
	"github.com/georgehadjisavvas/promitheies-cy/internal/entities"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Source identifies this importer's rows in tenders.source and the
// *_aliases.source columns for provenance.
const Source = "ted"

// CyprusQuery selects every TED notice with a place of performance in
// Cyprus, across all form types (competition, planning, result, ...).
const CyprusQuery = "place-of-performance=CYP"

type Stats struct {
	Processed int
	Inserted  int
	Updated   int
	Skipped   int
}

// Import upserts every notice into tenders, keyed on the source's
// publication-number so re-running Import is idempotent.
func Import(ctx context.Context, pool *pgxpool.Pool, notices []Notice) (Stats, error) {
	var stats Stats

	for _, n := range notices {
		stats.Processed++

		if n.AuthorityName == "" {
			stats.Skipped++
			continue
		}

		authorityID, err := entities.ResolveAuthority(ctx, pool, n.AuthorityName, Source)
		if err != nil {
			return stats, fmt.Errorf("resolve authority for %s: %w", n.PublicationNumber, err)
		}

		var primaryCPV, cpvDivision *string
		if len(n.CPVCodes) > 0 {
			primaryCPV = &n.CPVCodes[0]
			div := cpv.DivisionCode(n.CPVCodes[0]) + "000000"
			cpvDivision = &div
		}

		var inserted bool
		err = pool.QueryRow(ctx, `
			INSERT INTO tenders (
				external_ids, title_en, title_el, authority_id, cpv_code, cpv_division,
				estimated_value, currency, status, procedure_type, source, published_at, raw_payload
			) VALUES (
				jsonb_build_object('ted', $1::text), $2, $3, $4, $5, $6,
				$7, $8, $9, $10, $11, $12, $13
			)
			ON CONFLICT (source, (external_ids->>'ted')) WHERE external_ids ? 'ted'
			DO UPDATE SET
				title_en = EXCLUDED.title_en,
				title_el = EXCLUDED.title_el,
				authority_id = EXCLUDED.authority_id,
				cpv_code = EXCLUDED.cpv_code,
				cpv_division = EXCLUDED.cpv_division,
				estimated_value = EXCLUDED.estimated_value,
				currency = EXCLUDED.currency,
				status = EXCLUDED.status,
				procedure_type = EXCLUDED.procedure_type,
				published_at = EXCLUDED.published_at,
				raw_payload = EXCLUDED.raw_payload,
				updated_at = now()
			RETURNING (xmax = 0)
		`,
			n.PublicationNumber, nullableStr(n.TitleEN), nullableStr(n.TitleEL), authorityID, primaryCPV, cpvDivision,
			n.EstimatedValue, n.Currency, n.Status, nullableStr(n.ProcedureType), Source, n.PublishedAt, n.Raw,
		).Scan(&inserted)
		if err != nil {
			return stats, fmt.Errorf("upsert tender for %s: %w", n.PublicationNumber, err)
		}
		if inserted {
			stats.Inserted++
		} else {
			stats.Updated++
		}
	}

	return stats, nil
}

func nullableStr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
