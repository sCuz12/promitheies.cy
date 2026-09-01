package alerts

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestFormatNewTenderMessageEscapesHTMLAndAddsLink(t *testing.T) {
	value := 12500.5
	published := time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC)
	cpvDivision := "45000000"

	msg := FormatNewTenderMessage(Tender{
		Title:             "Works & <services>",
		AuthorityName:     "Ministry <Health>",
		Source:            "ted",
		ExternalID:        "123",
		CPVDivision:       &cpvDivision,
		EstimatedValueEUR: &value,
		PublishedAt:       &published,
		ID:                42,
	}, "https://promitheies.cy/")

	assertContains(t, msg, "<b>New open tender</b>")
	assertContains(t, msg, "Works &amp; &lt;services&gt;")
	assertContains(t, msg, "Authority: Ministry &lt;Health&gt;")
	assertContains(t, msg, "Estimated value: EUR 12500.50")
	assertContains(t, msg, "Published: 2026-08-31")
	assertContains(t, msg, "CPV division: 45000000")
	assertContains(t, msg, "Source: TED <code>123</code>")
	assertContains(t, msg, "https://promitheies.cy/tender/42")
}

func TestSendNewTenderMessagesCapsAndSummarizes(t *testing.T) {
	sender := &recordingSender{}
	tenders := []Tender{
		{ID: 1, Title: "One"},
		{ID: 2, Title: "Two"},
		{ID: 3, Title: "Three"},
	}

	err := SendNewTenderMessages(context.Background(), sender, tenders, 2, "https://promitheies.cy")
	if err != nil {
		t.Fatalf("SendNewTenderMessages returned error: %v", err)
	}
	if got, want := len(sender.messages), 3; got != want {
		t.Fatalf("sent %d messages, want %d", got, want)
	}
	assertContains(t, sender.messages[0], "One")
	assertContains(t, sender.messages[1], "Two")
	assertContains(t, sender.messages[2], "1 more new tenders imported.")
}

type recordingSender struct {
	messages []string
}

func (s *recordingSender) SendMessage(_ context.Context, text string) error {
	s.messages = append(s.messages, text)
	return nil
}

func assertContains(t *testing.T, s, substr string) {
	t.Helper()
	if !strings.Contains(s, substr) {
		t.Fatalf("expected %q to contain %q", s, substr)
	}
}
