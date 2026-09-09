package distribution

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/georgehadjisavvas/promitheies-cy/internal/cpv"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	PlatformX      = "x"
	PlatformReddit = "reddit"

	CadenceDaily  = "daily"
	CadenceWeekly = "weekly"

	MaxXChars          = 280
	MaxRedditChars     = 10_000
	DefaultTenderLimit = 3
)

type Tender struct {
	ID                int64
	Title             string
	AuthorityName     string
	ExternalID        string
	CPVDivision       *string
	CPVCategoryName   *string
	EstimatedValueEUR *float64
	PublishedAt       *time.Time
	Deadline          *time.Time
}

type Window struct {
	Start time.Time
	End   time.Time
}

type DraftRequest struct {
	Cadence       string
	Platform      string
	Window        Window
	Tenders       []Tender
	PublicBaseURL string
}

type LLMClient interface {
	GeneratePost(ctx context.Context, prompt string) (string, error)
}

func WindowFor(selected time.Time, cadence string) (Window, error) {
	day := time.Date(selected.Year(), selected.Month(), selected.Day(), 0, 0, 0, 0, selected.Location())
	switch cadence {
	case CadenceDaily:
		return Window{Start: day, End: day.AddDate(0, 0, 1)}, nil
	case CadenceWeekly:
		// offset := (int(day.Weekday()) + 6) % 7
		start := day.AddDate(0, 0, -6)
		return Window{Start: start, End: start.AddDate(0, 0, 7)}, nil
	default:
		return Window{}, fmt.Errorf("unsupported cadence %q", cadence)
	}
}

func LoadNewTEDTenders(ctx context.Context, pool *pgxpool.Pool, w Window, limit int) ([]Tender, error) {
	if limit <= 0 {
		limit = DefaultTenderLimit
	}
	rows, err := pool.Query(ctx, `
		SELECT
			t.id,
			coalesce(t.title_el, t.title_en, ''),
			coalesce(auth.canonical_name_el, auth.canonical_name_en, ''),
			t.external_ids->>'ted',
			t.cpv_division,
			coalesce(cpv.description_el, cpv.description_en, ''),
			t.estimated_value,
			t.published_at,
			t.deadline
		FROM tenders t
		JOIN authorities auth ON auth.id = t.authority_id
		LEFT JOIN cpv_categories cpv ON cpv.code = t.cpv_division
		WHERE t.source = 'ted'
		  AND t.status = 'open'
		  AND (t.deadline IS NULL OR t.deadline >= CURRENT_DATE)
		  AND t.created_at >= $1
		  AND t.created_at < $2
		ORDER BY t.created_at DESC, t.estimated_value DESC NULLS LAST, t.deadline ASC NULLS LAST, t.id DESC
		LIMIT $3
	`, w.Start, w.End, limit)
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
			&t.ExternalID,
			&t.CPVDivision,
			&t.CPVCategoryName,
			&t.EstimatedValueEUR,
			&t.PublishedAt,
			&t.Deadline,
		); err != nil {
			return nil, err
		}
		tenders = append(tenders, t)
	}
	return tenders, rows.Err()
}

func GenerateDraft(ctx context.Context, llm LLMClient, req DraftRequest) (string, error) {
	if !SupportedPlatform(req.Platform) {
		return "", fmt.Errorf("unsupported platform %q", req.Platform)
	}
	if len(req.Tenders) == 0 {
		post := NoNewTendersPost(req)
		if err := ValidatePost(req.Platform, post); err != nil {
			return "", err
		}
		return post, nil
	}
	if llm == nil {
		return "", errors.New("llm client is required")
	}

	raw, err := llm.GeneratePost(ctx, BuildPrompt(req))
	if err != nil {
		return "", err
	}
	post := CleanPost(raw)
	if err := ValidatePost(req.Platform, post); err != nil {
		return "", err
	}
	return post, nil
}

func NoNewTendersPost(req DraftRequest) string {
	if req.Platform == PlatformReddit {
		return fmt.Sprintf("## Cyprus public tender update — %s\n\nNo new open TED tenders for Cyprus were retrieved for this period. You can still browse currently open opportunities on Symvaseis.CY: %s",
			DateRangeLabel(req.Window), OpenTendersURL(req.PublicBaseURL))
	}
	return fmt.Sprintf("Δεν εντοπίστηκαν νέοι ανοικτοί διαγωνισμοί TED για την Κύπρο από το Symvaseis.CY για %s. Δείτε τους διαθέσιμους: %s",
		DateRangeLabel(req.Window), OpenTendersURL(req.PublicBaseURL))
}

func SupportedPlatform(platform string) bool {
	return platform == PlatformX || platform == PlatformReddit
}

func ValidatePost(platform, post string) error {
	switch platform {
	case PlatformX:
		return ValidateXPost(post)
	case PlatformReddit:
		post = strings.TrimSpace(post)
		if post == "" {
			return errors.New("Reddit post is empty")
		}
		if n := utf8.RuneCountInString(post); n > MaxRedditChars {
			return fmt.Errorf("Reddit post is %d characters; max is %d", n, MaxRedditChars)
		}
		return nil
	default:
		return fmt.Errorf("unsupported platform %q", platform)
	}
}

func ValidateXPost(post string) error {
	post = strings.TrimSpace(post)
	if post == "" {
		return errors.New("post is empty")
	}
	if n := utf8.RuneCountInString(post); n > MaxXChars {
		return fmt.Errorf("X post is %d characters; max is %d", n, MaxXChars)
	}
	return nil
}

func CleanPost(post string) string {
	post = strings.TrimSpace(post)
	post = strings.TrimPrefix(post, "```text")
	post = strings.TrimPrefix(post, "```")
	post = strings.TrimSuffix(post, "```")
	post = strings.TrimSpace(post)
	post = strings.Trim(post, "\"")
	return strings.TrimSpace(post)
}

func MarkdownDraft(req DraftRequest, post string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# Symvaseis.CY %s %s draft\n\n", platformLabel(req.Platform), req.Cadence)
	fmt.Fprintf(&b, "Date range: %s\n", DateRangeLabel(req.Window))
	fmt.Fprintf(&b, "Platform: %s\n", platformLabel(req.Platform))
	fmt.Fprintf(&b, "Source: TED\n\n")
	b.WriteString("## Copy post\n\n")
	b.WriteString("```text\n")
	b.WriteString(strings.TrimSpace(post))
	b.WriteString("\n```\n\n")
	b.WriteString("## Source tenders\n\n")
	if len(req.Tenders) == 0 {
		b.WriteString("- No new open TED tenders found.\n")
		return b.String()
	}
	for _, t := range req.Tenders {
		fmt.Fprintf(&b, "- #%d", t.ID)
		if t.ExternalID != "" {
			fmt.Fprintf(&b, " TED %s", t.ExternalID)
		}
		if t.Title != "" {
			fmt.Fprintf(&b, ": %s", t.Title)
		}
		if t.AuthorityName != "" {
			fmt.Fprintf(&b, " - %s", t.AuthorityName)
		}
		if category := t.CategoryName(); category != "" {
			fmt.Fprintf(&b, " (%s)", category)
		}
		b.WriteString("\n")
	}
	return b.String()
}

func OutputFilename(platform, cadence string, w Window) (string, error) {
	if !SupportedPlatform(platform) {
		return "", fmt.Errorf("unsupported platform %q", platform)
	}
	switch cadence {
	case CadenceDaily, CadenceWeekly:
	default:
		return "", fmt.Errorf("unsupported cadence %q", cadence)
	}
	return fmt.Sprintf("posts/%s-%s-%s.md", platform, cadence, w.Start.Format("2006-01-02")), nil
}

func BuildPrompt(req DraftRequest) string {
	if req.Platform == PlatformReddit {
		return BuildRedditPrompt(req)
	}
	stats := summarize(req.Tenders)
	var b strings.Builder
	fmt.Fprintf(&b, "Write exactly one X post for Symvaseis.CY in Greek.\n")
	fmt.Fprintf(&b, "Output language: Greek. Use natural Cyprus/Greek business language. Keep the brand name Symvaseis.CY unchanged.\n")
	fmt.Fprintf(&b, "Symvaseis.CY helps Cyprus businesses discover public tender opportunities without manually checking fragmented procurement portals.\n")
	fmt.Fprintf(&b, "The audience is Cyprus SMEs, contractors, suppliers, consultants, and operators who can bid on government work.\n")
	fmt.Fprintf(&b, "The goal of this post is awareness: make people understand that Symvaseis.CY retrieved the latest public tenders and is worth checking regularly.\n")
	fmt.Fprintf(&b, "The post must clearly communicate that these are the latest tenders retrieved by Symvaseis.CY for the selected period.\n")
	fmt.Fprintf(&b, "Base the post on the tender facts below. These are the latest retrieved TED tenders selected from the database.\n")
	fmt.Fprintf(&b, "Mention two or three of these recent tenders when possible, using short readable summaries instead of full titles if needed.\n")
	fmt.Fprintf(&b, "For each tender you mention, include the authority and category name when available.\n")
	fmt.Fprintf(&b, "Choose the strongest factual angle from the tender data: urgency, opportunity size, sector relevance, number of new tenders, or deadline pressure.\n")
	fmt.Fprintf(&b, "The post must be under %d characters, professional, factual, and copy-ready.\n", MaxXChars)
	fmt.Fprintf(&b, "Format for copy/paste: one clean paragraph, no bullets, no numbering, no labels like 'Post:', no markdown, and no surrounding quotes.\n")
	fmt.Fprintf(&b, "Use a strong first sentence, but avoid hype, clickbait, emojis, hashtags, and vague claims.\n")
	fmt.Fprintf(&b, "Return only the final post text and nothing else.\n")
	fmt.Fprintf(&b, "Do not invent tender counts, values, authorities, sectors, deadlines, or URLs.\n")
	fmt.Fprintf(&b, "Cadence: %s\n", req.Cadence)
	fmt.Fprintf(&b, "Date range: %s\n", DateRangeLabel(req.Window))
	fmt.Fprintf(&b, "New open TED tenders: %d\n", len(req.Tenders))
	if len(stats.CPVDivisions) > 0 {
		fmt.Fprintf(&b, "Top CPV divisions: %s\n", strings.Join(stats.CPVDivisions, ", "))
	}
	if stats.EarliestDeadline != "" {
		fmt.Fprintf(&b, "Earliest deadline: %s\n", stats.EarliestDeadline)
	}
	fmt.Fprintf(&b, "Open tenders URL: %s\n\n", OpenTendersURL(req.PublicBaseURL))
	b.WriteString("Tender facts:\n")
	for _, t := range req.Tenders {
		fmt.Fprintf(&b, "- ID %d", t.ID)
		if t.Title != "" {
			fmt.Fprintf(&b, "; title: %s", t.Title)
		}
		if t.AuthorityName != "" {
			fmt.Fprintf(&b, "; authority: %s", t.AuthorityName)
		}
		if category := t.CategoryName(); category != "" {
			fmt.Fprintf(&b, "; category: %s", category)
		}
		if t.CPVDivision != nil && *t.CPVDivision != "" {
			fmt.Fprintf(&b, "; CPV division: %s", *t.CPVDivision)
		}
		if t.EstimatedValueEUR != nil {
			fmt.Fprintf(&b, "; estimated value EUR %.2f", *t.EstimatedValueEUR)
		}
		if t.Deadline != nil {
			fmt.Fprintf(&b, "; deadline: %s", t.Deadline.Format("2006-01-02"))
		}
		b.WriteString("\n")
	}
	return b.String()
}

// BuildRedditPrompt creates a useful community update rather than an advert:
// it explains what opened, which local businesses may care, and where to
// inspect the primary tender details.
func BuildRedditPrompt(req DraftRequest) string {
	stats := summarize(req.Tenders)
	var b strings.Builder
	fmt.Fprintf(&b, "Write one useful Reddit post for a Cyprus community about public tender opportunities that opened in the selected period.\n")
	fmt.Fprintf(&b, "Output language: clear, natural English for Cyprus residents and businesses. Preserve official Greek tender or authority names when they are supplied. Keep Symvaseis.CY unchanged.\n")
	fmt.Fprintf(&b, "Symvaseis.CY helps people find Cyprus public procurement opportunities without manually checking fragmented portals.\n")
	fmt.Fprintf(&b, "This is a community update, not an advertisement. Lead with the practical value: what opened, the sectors involved, who may find it relevant, and any deadline people should notice.\n")
	fmt.Fprintf(&b, "Use the tender facts only. Do not infer eligibility, contract scope, economic impact, values, deadlines, counts, or links that are not stated. Do not provide legal, procurement, or bidding advice.\n")
	fmt.Fprintf(&b, "Format for copy/paste in Reddit Markdown:\n")
	fmt.Fprintf(&b, "- First line: a specific, neutral title under 120 characters; do not prefix it with 'Title:'.\n")
	fmt.Fprintf(&b, "- Then a short 1–2 sentence summary of the selected day or week and why it may matter to Cyprus businesses or interested residents.\n")
	fmt.Fprintf(&b, "- Add a 'What opened' section with up to five concise bullets. For each tender mentioned, include its plain-language subject, contracting authority, category, estimated value, and deadline only when available.\n")
	fmt.Fprintf(&b, "- Add a brief 'Why it may be useful' section that connects only the stated sectors to likely interested local audiences (for example contractors, suppliers, consultants, or service providers), without claiming they qualify.\n")
	fmt.Fprintf(&b, "- End with a low-pressure invitation to explore the open-tenders link and verify requirements in the official notice.\n")
	fmt.Fprintf(&b, "Avoid marketing language, hype, emojis, hashtags, unsupported opinions, and calls to upvote. Be factual, skimmable, and genuinely helpful.\n")
	fmt.Fprintf(&b, "Return only the final Reddit post; no explanation, code fence, or surrounding quotes.\n")
	fmt.Fprintf(&b, "Cadence: %s\n", req.Cadence)
	fmt.Fprintf(&b, "Date range: %s\n", DateRangeLabel(req.Window))
	fmt.Fprintf(&b, "New open TED tenders: %d\n", len(req.Tenders))
	if len(stats.CPVDivisions) > 0 {
		fmt.Fprintf(&b, "Top CPV divisions: %s\n", strings.Join(stats.CPVDivisions, ", "))
	}
	if stats.EarliestDeadline != "" {
		fmt.Fprintf(&b, "Earliest deadline: %s\n", stats.EarliestDeadline)
	}
	fmt.Fprintf(&b, "Open tenders URL: %s\n\n", OpenTendersURL(req.PublicBaseURL))
	b.WriteString("Tender facts:\n")
	for _, t := range req.Tenders {
		fmt.Fprintf(&b, "- ID %d", t.ID)
		if t.Title != "" {
			fmt.Fprintf(&b, "; title: %s", t.Title)
		}
		if t.AuthorityName != "" {
			fmt.Fprintf(&b, "; authority: %s", t.AuthorityName)
		}
		if category := t.CategoryName(); category != "" {
			fmt.Fprintf(&b, "; category: %s", category)
		}
		if t.CPVDivision != nil && *t.CPVDivision != "" {
			fmt.Fprintf(&b, "; CPV division: %s", *t.CPVDivision)
		}
		if t.EstimatedValueEUR != nil {
			fmt.Fprintf(&b, "; estimated value EUR %.2f", *t.EstimatedValueEUR)
		}
		if t.Deadline != nil {
			fmt.Fprintf(&b, "; deadline: %s", t.Deadline.Format("2006-01-02"))
		}
		b.WriteString("\n")
	}
	return b.String()
}

func platformLabel(platform string) string {
	if platform == PlatformReddit {
		return "Reddit"
	}
	return "X"
}

func (t Tender) CategoryName() string {
	if t.CPVCategoryName != nil && strings.TrimSpace(*t.CPVCategoryName) != "" {
		return strings.TrimSpace(*t.CPVCategoryName)
	}
	if t.CPVDivision != nil && strings.TrimSpace(*t.CPVDivision) != "" {
		return cpv.NameLang("el", strings.TrimSpace(*t.CPVDivision))
	}
	return ""
}

func DateRangeLabel(w Window) string {
	if w.End.Sub(w.Start) <= 24*time.Hour {
		return w.Start.Format("2006-01-02")
	}
	return fmt.Sprintf("%s to %s", w.Start.Format("2006-01-02"), w.End.AddDate(0, 0, -1).Format("2006-01-02"))
}

func OpenTendersURL(publicBaseURL string) string {
	base := strings.TrimRight(publicBaseURL, "/")
	if base == "" {
		return "/diagonismoi"
	}
	return base + "/diagonismoi"
}

type summary struct {
	CPVDivisions     []string
	EarliestDeadline string
}

func summarize(tenders []Tender) summary {
	counts := map[string]int{}
	var earliest *time.Time
	for _, t := range tenders {
		if t.CPVDivision != nil && *t.CPVDivision != "" {
			counts[*t.CPVDivision]++
		}
		if t.Deadline != nil && (earliest == nil || t.Deadline.Before(*earliest)) {
			d := *t.Deadline
			earliest = &d
		}
	}
	divisions := make([]string, 0, len(counts))
	for div := range counts {
		divisions = append(divisions, div)
	}
	sort.Slice(divisions, func(i, j int) bool {
		if counts[divisions[i]] == counts[divisions[j]] {
			return divisions[i] < divisions[j]
		}
		return counts[divisions[i]] > counts[divisions[j]]
	})
	if len(divisions) > 3 {
		divisions = divisions[:3]
	}
	out := summary{CPVDivisions: divisions}
	if earliest != nil {
		out.EarliestDeadline = earliest.Format("2006-01-02")
	}
	return out
}
