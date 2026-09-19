package ted

import (
	"encoding/json"
	"strconv"
	"strings"
	"time"
)

// Fields is the set of search-result fields this importer requests. Keep
// in sync with rawNotice's json tags.
var Fields = []string{
	"publication-number",
	"form-type",
	"notice-type",
	"procedure-type",
	"classification-cpv",
	"buyer-name",
	"buyer-country",
	"notice-title",
	"publication-date",
	"deadline",
	"deadline-receipt-tender-date-lot",
	"estimated-value-proc",
	"estimated-value-cur-proc",
}

type rawNotice struct {
	PublicationNumber     string              `json:"publication-number"`
	FormType              string              `json:"form-type"`
	NoticeType            string              `json:"notice-type"`
	ProcedureType         string              `json:"procedure-type"`
	ClassificationCPV     []string            `json:"classification-cpv"`
	BuyerName             map[string][]string `json:"buyer-name"`
	BuyerCountry          []string            `json:"buyer-country"`
	NoticeTitle           map[string]string   `json:"notice-title"`
	PublicationDate       string              `json:"publication-date"`
	Deadline              []string            `json:"deadline"`
	TenderDeadlineDates   []string            `json:"deadline-receipt-tender-date-lot"`
	EstimatedValueProc    string              `json:"estimated-value-proc"`
	EstimatedValueCurProc string              `json:"estimated-value-cur-proc"`
}

// Notice is a TED notice mapped into the shape our importer works with.
type Notice struct {
	PublicationNumber string
	FormType          string
	ProcedureType     string
	CPVCodes          []string
	AuthorityName     string
	TitleEN           string
	TitleEL           string
	PublishedAt       *time.Time
	Deadline          *time.Time
	EstimatedValue    *float64
	Currency          string
	Status            string
	Raw               json.RawMessage
}

// openFormTypes are notices describing a tender still accepting (or about
// to accept) submissions.
var openFormTypes = map[string]bool{
	"competition": true,
	"planning":    true,
}

// awardedFormTypes are notices describing a completed award. See the
// package doc: award winner/value aren't reliably present at this level.
var awardedFormTypes = map[string]bool{
	"result":       true,
	"dir-awa-pre":  true,
	"can-standard": true,
}

// ParseNotice maps one raw search-result notice into a Notice.
func ParseNotice(raw json.RawMessage) (Notice, error) {
	var rn rawNotice
	if err := json.Unmarshal(raw, &rn); err != nil {
		return Notice{}, err
	}

	n := Notice{
		PublicationNumber: rn.PublicationNumber,
		FormType:          rn.FormType,
		ProcedureType:     rn.ProcedureType,
		CPVCodes:          dedupe(rn.ClassificationCPV),
		AuthorityName:     firstBuyerName(rn.BuyerName),
		TitleEN:           rn.NoticeTitle["eng"],
		TitleEL:           rn.NoticeTitle["ell"],
		PublishedAt:       parsePublicationDate(rn.PublicationDate),
		Deadline:          parseDeadline(rn.Deadline, rn.TenderDeadlineDates),
		EstimatedValue:    parseAmount(rn.EstimatedValueProc),
		Currency:          rn.EstimatedValueCurProc,
		Raw:               raw,
	}

	switch {
	case openFormTypes[rn.FormType]:
		n.Status = "open"
	case awardedFormTypes[rn.FormType]:
		n.Status = "awarded"
	default:
		n.Status = "unknown"
	}
	if n.Currency == "" {
		n.Currency = "EUR"
	}

	return n, nil
}

// firstBuyerName prefers the Greek name, then English, then whatever's
// first in the map (TED buyer-name is keyed by ISO 639-2 language code,
// each value a list — multiple buyers on a joint procurement).
func firstBuyerName(m map[string][]string) string {
	for _, lang := range []string{"ell", "eng"} {
		if names := m[lang]; len(names) > 0 && names[0] != "" {
			return names[0]
		}
	}
	for _, names := range m {
		if len(names) > 0 && names[0] != "" {
			return names[0]
		}
	}
	return ""
}

func dedupe(codes []string) []string {
	seen := make(map[string]bool, len(codes))
	out := make([]string, 0, len(codes))
	for _, c := range codes {
		if c != "" && !seen[c] {
			seen[c] = true
			out = append(out, c)
		}
	}
	return out
}

// parsePublicationDate handles TED's "YYYY-MM-DD+HH:MM" format by taking
// just the date portion.
func parsePublicationDate(s string) *time.Time {
	if len(s) < 10 {
		return nil
	}
	t, err := time.Parse("2006-01-02", s[:10])
	if err != nil {
		return nil
	}
	return &t
}

// parseDeadline prefers TED's generic deadline field because it matches the
// deadline surfaced by TED search (and can represent a request-to-participate
// deadline in staged procedures). The tender-receipt dates are a fallback.
// When a notice exposes multiple lot deadlines, keep the earliest actionable
// date for the site's countdown.
func parseDeadline(groups ...[]string) *time.Time {
	for _, values := range groups {
		var earliest *time.Time
		for _, value := range values {
			parsed := parsePublicationDate(value)
			if parsed != nil && (earliest == nil || parsed.Before(*earliest)) {
				earliest = parsed
			}
		}
		if earliest != nil {
			return earliest
		}
	}
	return nil
}

func parseAmount(s string) *float64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return nil
	}
	return &v
}
