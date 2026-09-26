package digest

import (
	"strings"
	"testing"
	"time"
)

func TestRenderDigestEscapesHTMLAndAddsLink(t *testing.T) {
	value := 12500.5
	deadline := time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC)
	cpvDivision := "45000000"

	subject, htmlBody, textBody := RenderDigest("en", []Tender{
		{
			ID:                42,
			Title:             "Works & <services>",
			AuthorityName:     "Ministry <Health>",
			CPVDivision:       &cpvDivision,
			EstimatedValueEUR: &value,
			Deadline:          &deadline,
		},
	}, "https://symvaseis.cy/")

	assertContains(t, subject, "1 new tenders this week")
	assertContains(t, htmlBody, "Works &amp; &lt;services&gt;")
	assertContains(t, htmlBody, "https://symvaseis.cy/tender/42")
	assertContains(t, htmlBody, "Ministry &lt;Health&gt;")
	assertContains(t, htmlBody, "EUR 12500.50")
	assertContains(t, htmlBody, "2026-08-31")

	assertContains(t, textBody, "Works & <services>")
	assertContains(t, textBody, "https://symvaseis.cy/tender/42")
}

func TestRenderDigestSubjectCountBothLanguages(t *testing.T) {
	tenders := []Tender{{ID: 1, Title: "One"}, {ID: 2, Title: "Two"}}

	subjectEN, _, _ := RenderDigest("en", tenders, "https://symvaseis.cy")
	assertContains(t, subjectEN, "2 new tenders this week")

	subjectEL, _, _ := RenderDigest("el", tenders, "https://symvaseis.cy")
	assertContains(t, subjectEL, "2 νέοι διαγωνισμοί αυτή την εβδομάδα")
}

func TestRenderDigestBuildsURLFromPublicBaseURL(t *testing.T) {
	_, htmlBody, textBody := RenderDigest("en", []Tender{{ID: 7, Title: "Tender"}}, "https://example.com")

	assertContains(t, htmlBody, "https://example.com/tender/7")
	assertContains(t, textBody, "https://example.com/tender/7")
}

func assertContains(t *testing.T, s, substr string) {
	t.Helper()
	if !strings.Contains(s, substr) {
		t.Fatalf("expected %q to contain %q", s, substr)
	}
}
