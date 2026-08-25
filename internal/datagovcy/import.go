package datagovcy

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/georgehadjisavvas/promitheies-cy/internal/cpv"
	"github.com/georgehadjisavvas/promitheies-cy/internal/entities"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Source identifies this importer's rows in tenders.source, awards.source,
// and the *_aliases.source columns for provenance.
const Source = "datagovcy"

// placeholderContractorNames are EONAME values observed in the source data
// that are generic procurement-process labels or test records rather than
// an actual winning company (e.g. "ΟΙΚΟΝΟΜΙΚΟΣ ΦΟΡΕΑΣ" = "Economic Operator",
// a placeholder that alone would otherwise rank as the #1 contractor by
// value). Treated the same as a blank contractor name: the tender is still
// imported, just without an award row.
var placeholderContractorNames = map[string]bool{
	"ΟΙΚΟΝΟΜΙΚΟΣ ΦΟΡΕΑΣ": true,
	"Αγορές μέσω Πρεσβείας της Κυπριακής Δημοκρατίας": true,
	"Γενικό Λογιστήριο της Δημοκρατίας (Παρατηρητής/Test)": true,
}

// Stats summarizes one import run for logging into ingest_runs.
type Stats struct {
	Processed int
	Inserted  int
	Updated   int
	Skipped   int
}

// Import upserts every award in awards into the database: resolving (or
// creating) the authority and contractor, then upserting the tender and its
// award row keyed on the source's CFTID. Re-running Import with the same
// data is a no-op past the first run (idempotent on external_ids->>'cftid').
func Import(ctx context.Context, pool *pgxpool.Pool, awards []Award) (Stats, error) {
	var stats Stats

	for _, a := range awards {
		stats.Processed++

		if a.AuthorityName == "" {
			stats.Skipped++
			continue
		}

		authorityID, err := entities.ResolveAuthority(ctx, pool, a.AuthorityName, Source)
		if err != nil {
			return stats, fmt.Errorf("resolve authority for CFTID %s: %w", a.CFTID, err)
		}

		var contractorID *int64
		if a.ContractorName != "" && !placeholderContractorNames[a.ContractorName] {
			cid, err := entities.ResolveContractor(ctx, pool, a.ContractorName, Source)
			if err != nil {
				return stats, fmt.Errorf("resolve contractor for CFTID %s: %w", a.CFTID, err)
			}
			contractorID = &cid
		}

		var primaryCPV, cpvDivision *string
		if len(a.CPVCodes) > 0 {
			primaryCPV = &a.CPVCodes[0]
			div := cpv.DivisionCode(a.CPVCodes[0]) + "000000"
			cpvDivision = &div
		}

		status := "unknown"
		switch {
		case a.AwardDate != nil && contractorID != nil:
			status = "awarded"
		case a.Deadline != nil:
			status = "open"
		}

		rawPayload, err := json.Marshal(a)
		if err != nil {
			return stats, fmt.Errorf("marshal raw payload for CFTID %s: %w", a.CFTID, err)
		}

		var tenderID int64
		var inserted bool
		err = pool.QueryRow(ctx, `
			INSERT INTO tenders (
				external_ids, title_el, authority_id, cpv_code, cpv_division,
				estimated_value, deadline, status, procedure_type, source,
				published_at, raw_payload
			) VALUES (
				jsonb_build_object('cftid', $1::text), $2, $3, $4, $5,
				$6, $7, $8, $9, $10,
				$11, $12
			)
			ON CONFLICT (source, (external_ids->>'cftid')) WHERE external_ids ? 'cftid'
			DO UPDATE SET
				title_el = EXCLUDED.title_el,
				authority_id = EXCLUDED.authority_id,
				cpv_code = EXCLUDED.cpv_code,
				cpv_division = EXCLUDED.cpv_division,
				estimated_value = EXCLUDED.estimated_value,
				deadline = EXCLUDED.deadline,
				status = EXCLUDED.status,
				procedure_type = EXCLUDED.procedure_type,
				published_at = EXCLUDED.published_at,
				raw_payload = EXCLUDED.raw_payload,
				updated_at = now()
			RETURNING id, (xmax = 0)
		`,
			a.CFTID, nullableStr(a.Title), authorityID, primaryCPV, cpvDivision,
			a.EstimatedValue, a.Deadline, status, nullableStr(a.ProcedureType), Source,
			a.PublishedAt, rawPayload,
		).Scan(&tenderID, &inserted)
		if err != nil {
			return stats, fmt.Errorf("upsert tender for CFTID %s: %w", a.CFTID, err)
		}
		if inserted {
			stats.Inserted++
		} else {
			stats.Updated++
		}

		if contractorID != nil && a.AwardedValue != nil {
			_, err = pool.Exec(ctx, `
				INSERT INTO awards (tender_id, contractor_id, value, award_date, source)
				VALUES ($1, $2, $3, $4, $5)
				ON CONFLICT (tender_id, source) DO UPDATE SET
					contractor_id = EXCLUDED.contractor_id,
					value = EXCLUDED.value,
					award_date = EXCLUDED.award_date
			`, tenderID, *contractorID, *a.AwardedValue, a.AwardDate, Source)
			if err != nil {
				return stats, fmt.Errorf("upsert award for CFTID %s: %w", a.CFTID, err)
			}
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
