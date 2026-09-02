package distribution

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	PlatformX = "x"

	CadenceDaily  = "daily"
	CadenceWeekly = "weekly"

	MaxXChars = 280

	DefaultOptionCount = 3
)

type Tender struct {
	ID                int64
	Title             string
	AuthorityName     string
	ExternalID        string
	CPVDivision       *string
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
		offset := (int(day.Weekday()) + 6) % 7
		start := day.AddDate(0, 0, -offset)
		return Window{Start: start, End: start.AddDate(0, 0, 7)}, nil
	default:
		return Window{}, fmt.Errorf("unsupported cadence %q", cadence)
	}
}

func LoadNewTEDTenders(ctx context.Context, pool *pgxpool.Pool, w Window, limit int) ([]Tender, error) {
	if limit <= 0 {
		limit = 20
	}
	rows, err := pool.Query(ctx, `
		SELECT
			t.id,
			coalesce(t.title_el, t.title_en, ''),
			coalesce(auth.canonical_name_el, auth.canonical_name_en, ''),
			t.external_ids->>'ted',
			t.cpv_division,
			t.estimated_value,
			t.published_at,
			t.deadline
		FROM tenders t
		JOIN authorities auth ON auth.id = t.authority_id
		WHERE t.source = 'ted'
		  AND t.status = 'open'
		  AND (t.deadline IS NULL OR t.deadline >= CURRENT_DATE)
		  AND t.created_at >= $1
		  AND t.created_at < $2
		ORDER BY t.estimated_value DESC NULLS LAST, t.deadline ASC NULLS LAST, t.id DESC
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

func GenerateDraft(ctx context.Context, llm LLMClient, req DraftRequest) ([]string, error) {
	if req.Platform != PlatformX {
		return nil, fmt.Errorf("unsupported platform %q", req.Platform)
	}
	if len(req.Tenders) == 0 {
		post := NoNewTendersPost(req)
		if err := ValidateXPost(post); err != nil {
			return nil, err
		}
		return []string{post}, nil
	}
	if llm == nil {
		return nil, errors.New("llm client is required")
	}

	raw, err := llm.GeneratePost(ctx, BuildPrompt(req))
	if err != nil {
		return nil, err
	}
	posts, err := ParsePostOptions(raw)
	if err != nil {
		return nil, err
	}
	return posts, nil
}

func NoNewTendersPost(req DraftRequest) string {
	return fmt.Sprintf("No new open TED tenders for Cyprus found by Promitheies for %s. View current opportunities: %s",
		DateRangeLabel(req.Window), OpenTendersURL(req.PublicBaseURL))
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

func ParsePostOptions(raw string) ([]string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, errors.New("post options are empty")
	}

	lines := strings.Split(raw, "\n")
	var posts []string
	re := regexp.MustCompile(`^\s*(?:option\s*)?\d+[\).\:-]\s*`)
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		line = strings.TrimPrefix(line, "- ")
		line = strings.TrimPrefix(line, "* ")
		line = re.ReplaceAllString(line, "")
		line = CleanPost(line)
		if line == "" {
			continue
		}
		if err := ValidateXPost(line); err != nil {
			return nil, err
		}
		posts = append(posts, line)
	}
	if len(posts) != DefaultOptionCount {
		return nil, fmt.Errorf("expected %d post options, got %d", DefaultOptionCount, len(posts))
	}
	return posts, nil
}

func MarkdownDraft(req DraftRequest, posts []string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# Promitheies X %s draft\n\n", req.Cadence)
	fmt.Fprintf(&b, "Date range: %s\n", DateRangeLabel(req.Window))
	fmt.Fprintf(&b, "Platform: X\n")
	fmt.Fprintf(&b, "Source: TED\n\n")
	b.WriteString("## Copy options\n\n")
	for i, post := range posts {
		fmt.Fprintf(&b, "### Option %d\n\n", i+1)
		b.WriteString("```text\n")
		b.WriteString(strings.TrimSpace(post))
		b.WriteString("\n```\n\n")
	}
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
		b.WriteString("\n")
	}
	return b.String()
}

func OutputFilename(platform, cadence string, w Window) (string, error) {
	if platform != PlatformX {
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
	stats := summarize(req.Tenders)
	var b strings.Builder
	fmt.Fprintf(&b, "Write exactly %d alternative X posts for Promitheies.cy.\n", DefaultOptionCount)
	fmt.Fprintf(&b, "Promitheies.cy helps Cyprus businesses discover public tender opportunities without manually checking fragmented procurement portals.\n")
	fmt.Fprintf(&b, "The audience is Cyprus SMEs, contractors, suppliers, consultants, and operators who can bid on government work.\n")
	fmt.Fprintf(&b, "The goal of this post is awareness: make people understand that Promitheies tracks new public tenders and is worth checking regularly.\n")
	fmt.Fprintf(&b, "Each option should let the LLM choose the strongest factual angle from the tender data: urgency, opportunity size, sector relevance, number of new tenders, or deadline pressure.\n")
	fmt.Fprintf(&b, "Each post must be under %d characters, professional, factual, and copy-ready.\n", MaxXChars)
	fmt.Fprintf(&b, "Use a strong first sentence, but avoid hype, clickbait, emojis, hashtags, markdown, and vague claims.\n")
	fmt.Fprintf(&b, "Return each option on its own line, numbered 1-3, and nothing else.\n")
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
