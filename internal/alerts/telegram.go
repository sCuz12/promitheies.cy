package alerts

import (
	"context"
	"fmt"
	"html"
	"strconv"
	"strings"
	"time"
)

type Sender interface {
	SendMessage(ctx context.Context, text string) error
}

type Tender struct {
	ID                int64
	Title             string
	AuthorityName     string
	Source            string
	ExternalID        string
	CPVDivision       *string
	EstimatedValueEUR *float64
	PublishedAt       *time.Time
	Deadline          *time.Time
}

func SendNewTenderMessages(ctx context.Context, sender Sender, tenders []Tender, max int, publicBaseURL string) error {
	if max <= 0 || max > len(tenders) {
		max = len(tenders)
	}
	for i := 0; i < max; i++ {
		if err := sender.SendMessage(ctx, FormatNewTenderMessage(tenders[i], publicBaseURL)); err != nil {
			return fmt.Errorf("send tender %d: %w", tenders[i].ID, err)
		}
	}
	if omitted := len(tenders) - max; omitted > 0 {
		msg := fmt.Sprintf("<b>%d more new tenders imported.</b>\nView them: %s", omitted, html.EscapeString(openTendersURL(publicBaseURL)))
		if err := sender.SendMessage(ctx, msg); err != nil {
			return fmt.Errorf("send omitted summary: %w", err)
		}
	}
	return nil
}

func FormatNewTenderMessage(t Tender, publicBaseURL string) string {
	title := strings.TrimSpace(t.Title)
	if title == "" {
		title = "Untitled tender"
	}

	var lines []string
	lines = append(lines, "<b>New open tender</b>")
	lines = append(lines, html.EscapeString(title))

	if t.AuthorityName != "" {
		lines = append(lines, "Authority: "+html.EscapeString(t.AuthorityName))
	}
	if t.EstimatedValueEUR != nil {
		lines = append(lines, fmt.Sprintf("Estimated value: EUR %.2f", *t.EstimatedValueEUR))
	}
	if t.Deadline != nil {
		lines = append(lines, "Deadline: "+t.Deadline.Format("2006-01-02"))
	}
	if t.PublishedAt != nil {
		lines = append(lines, "Published: "+t.PublishedAt.Format("2006-01-02"))
	}
	if t.CPVDivision != nil && *t.CPVDivision != "" {
		lines = append(lines, "CPV division: "+html.EscapeString(*t.CPVDivision))
	}
	if t.Source != "" && t.ExternalID != "" {
		lines = append(lines, fmt.Sprintf("Source: %s <code>%s</code>", html.EscapeString(strings.ToUpper(t.Source)), html.EscapeString(t.ExternalID)))
	}

	if publicBaseURL != "" {
		lines = append(lines, "View: "+html.EscapeString(tenderURL(publicBaseURL, t.ID)))
	}

	return strings.Join(lines, "\n")
}

func tenderURL(publicBaseURL string, id int64) string {
	base := strings.TrimRight(publicBaseURL, "/")
	if base == "" {
		return "/tender/" + strconv.FormatInt(id, 10)
	}
	return base + "/tender/" + strconv.FormatInt(id, 10)
}

func openTendersURL(publicBaseURL string) string {
	base := strings.TrimRight(publicBaseURL, "/")
	if base == "" {
		return "/diagonismoi"
	}
	return base + "/diagonismoi"
}
