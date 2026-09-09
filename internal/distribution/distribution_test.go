package distribution

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestWindowForDaily(t *testing.T) {
	loc := time.FixedZone("EET", 2*60*60)
	selected := time.Date(2026, 9, 2, 15, 4, 0, 0, loc)

	w, err := WindowFor(selected, CadenceDaily)
	if err != nil {
		t.Fatalf("WindowFor returned error: %v", err)
	}
	if got, want := w.Start.Format(time.RFC3339), "2026-09-02T00:00:00+02:00"; got != want {
		t.Fatalf("start = %s, want %s", got, want)
	}
	if got, want := w.End.Format(time.RFC3339), "2026-09-03T00:00:00+02:00"; got != want {
		t.Fatalf("end = %s, want %s", got, want)
	}
}

func TestWindowForWeeklyStartsMonday(t *testing.T) {
	loc := time.UTC
	selected := time.Date(2026, 9, 2, 15, 4, 0, 0, loc)

	w, err := WindowFor(selected, CadenceWeekly)
	if err != nil {
		t.Fatalf("WindowFor returned error: %v", err)
	}
	if got, want := w.Start.Format("2006-01-02"), "2026-08-31"; got != want {
		t.Fatalf("start = %s, want %s", got, want)
	}
	if got, want := w.End.Format("2006-01-02"), "2026-09-07"; got != want {
		t.Fatalf("end = %s, want %s", got, want)
	}
}

func TestWindowForWeeklySundayStartsPreviousMonday(t *testing.T) {
	selected := time.Date(2026, 9, 6, 15, 4, 0, 0, time.UTC) // Sunday

	w, err := WindowFor(selected, CadenceWeekly)
	if err != nil {
		t.Fatalf("WindowFor returned error: %v", err)
	}
	if got, want := w.Start.Format("2006-01-02"), "2026-08-31"; got != want {
		t.Fatalf("start = %s, want %s", got, want)
	}
	if got, want := w.End.Format("2006-01-02"), "2026-09-07"; got != want {
		t.Fatalf("end = %s, want %s", got, want)
	}
}

func TestOutputFilenameIsUnderPosts(t *testing.T) {
	w := Window{Start: time.Date(2026, 9, 2, 0, 0, 0, 0, time.UTC)}

	got, err := OutputFilename(PlatformX, CadenceDaily, w)
	if err != nil {
		t.Fatalf("OutputFilename returned error: %v", err)
	}
	if want := "posts/x-daily-2026-09-02.md"; got != want {
		t.Fatalf("filename = %q, want %q", got, want)
	}
}

func TestMarkdownDraftIncludesCopyBlockAndSources(t *testing.T) {
	category := "Κατασκευαστικές εργασίες"
	req := DraftRequest{
		Cadence:  CadenceDaily,
		Platform: PlatformX,
		Window:   Window{Start: time.Date(2026, 9, 2, 0, 0, 0, 0, time.UTC), End: time.Date(2026, 9, 3, 0, 0, 0, 0, time.UTC)},
		Tenders: []Tender{{
			ID:              42,
			Title:           "Road works",
			AuthorityName:   "Ministry of Transport",
			ExternalID:      "123456-2026",
			CPVCategoryName: &category,
		}},
	}

	md := MarkdownDraft(req, "Οι τελευταίοι διαγωνισμοί είναι διαθέσιμοι.")

	assertContains(t, md, "## Copy post")
	assertContains(t, md, "```text\nΟι τελευταίοι διαγωνισμοί είναι διαθέσιμοι.\n```")
	assertContains(t, md, "#42 TED 123456-2026: Road works")
	assertContains(t, md, "(Κατασκευαστικές εργασίες)")
	assertContains(t, md, "Source: TED")
}

func TestBuildPromptIncludesOnlyTEDTenderFacts(t *testing.T) {
	value := 12500.0
	deadline := time.Date(2026, 9, 20, 0, 0, 0, 0, time.UTC)
	cpv := "45000000"
	category := "Κατασκευαστικές εργασίες"
	req := DraftRequest{
		Cadence:       CadenceDaily,
		Platform:      PlatformX,
		Window:        Window{Start: time.Date(2026, 9, 2, 0, 0, 0, 0, time.UTC), End: time.Date(2026, 9, 3, 0, 0, 0, 0, time.UTC)},
		PublicBaseURL: "https://symvaseis.cy/",
		Tenders: []Tender{{
			ID:                42,
			Title:             "Road works",
			AuthorityName:     "Ministry of Transport",
			CPVDivision:       &cpv,
			CPVCategoryName:   &category,
			EstimatedValueEUR: &value,
			Deadline:          &deadline,
		}},
	}

	prompt := BuildPrompt(req)

	assertContains(t, prompt, "Write exactly one X post for Symvaseis.CY in Greek")
	assertContains(t, prompt, "Output language: Greek")
	assertContains(t, prompt, "helps Cyprus businesses discover public tender opportunities")
	assertContains(t, prompt, "The goal of this post is awareness")
	assertContains(t, prompt, "latest tenders retrieved by Symvaseis.CY")
	assertContains(t, prompt, "Mention two or three of these recent tenders")
	assertContains(t, prompt, "include the authority and category name")
	assertContains(t, prompt, "one clean paragraph")
	assertContains(t, prompt, "Choose the strongest factual angle")
	assertContains(t, prompt, "New open TED tenders: 1")
	assertContains(t, prompt, "Open tenders URL: https://symvaseis.cy/diagonismoi")
	assertContains(t, prompt, "title: Road works")
	assertContains(t, prompt, "category: Κατασκευαστικές εργασίες")
	assertContains(t, prompt, "estimated value EUR 12500.00")
}

func TestBuildRedditPromptPrioritizesCommunityValue(t *testing.T) {
	category := "Construction work"
	req := DraftRequest{
		Cadence:       CadenceWeekly,
		Platform:      PlatformReddit,
		Window:        Window{Start: time.Date(2026, 8, 31, 0, 0, 0, 0, time.UTC), End: time.Date(2026, 9, 7, 0, 0, 0, 0, time.UTC)},
		PublicBaseURL: "https://symvaseis.cy",
		Tenders: []Tender{{
			ID:              42,
			Title:           "Road works",
			AuthorityName:   "Ministry of Transport",
			CPVCategoryName: &category,
		}},
	}

	prompt := BuildPrompt(req)
	assertContains(t, prompt, "useful Reddit post for a Cyprus community")
	assertContains(t, prompt, "community update, not an advertisement")
	assertContains(t, prompt, "What opened")
	assertContains(t, prompt, "Why it may be useful")
	assertContains(t, prompt, "verify requirements in the official notice")
	assertContains(t, prompt, "title: Road works")
	assertContains(t, prompt, "category: Construction work")
}

func TestValidateXPostRejectsOverLimit(t *testing.T) {
	err := ValidateXPost(strings.Repeat("a", MaxXChars+1))
	if err == nil {
		t.Fatal("expected over-limit post to fail validation")
	}
}

func TestGenerateDraftSkipsLLMWhenNoTenders(t *testing.T) {
	llm := &fakeLLM{post: "should not be used"}
	req := DraftRequest{
		Cadence:  CadenceDaily,
		Platform: PlatformX,
		Window:   Window{Start: time.Date(2026, 9, 2, 0, 0, 0, 0, time.UTC), End: time.Date(2026, 9, 3, 0, 0, 0, 0, time.UTC)},
	}

	post, err := GenerateDraft(context.Background(), llm, req)
	if err != nil {
		t.Fatalf("GenerateDraft returned error: %v", err)
	}
	if llm.called {
		t.Fatal("LLM was called for no-tender draft")
	}
	assertContains(t, post, "Δεν εντοπίστηκαν νέοι ανοικτοί διαγωνισμοί TED")
}

func TestGenerateDraftUsesLLMAndCleansOutput(t *testing.T) {
	llm := &fakeLLM{post: "```text\nΟι τελευταίοι διαγωνισμοί που ανέκτησε το Symvaseis.CY: οδικά έργα, υπηρεσίες καθαρισμού και προμήθειες εξοπλισμού. Δείτε προθεσμίες: /diagonismoi\n```"}
	req := DraftRequest{
		Cadence:  CadenceDaily,
		Platform: PlatformX,
		Window:   Window{Start: time.Date(2026, 9, 2, 0, 0, 0, 0, time.UTC), End: time.Date(2026, 9, 3, 0, 0, 0, 0, time.UTC)},
		Tenders:  []Tender{{ID: 1, Title: "Tender"}},
	}

	post, err := GenerateDraft(context.Background(), llm, req)
	if err != nil {
		t.Fatalf("GenerateDraft returned error: %v", err)
	}
	if !llm.called {
		t.Fatal("LLM was not called")
	}
	if strings.Contains(post, "```") {
		t.Fatalf("post still contains code fence: %q", post)
	}
	assertContains(t, post, "Οι τελευταίοι διαγωνισμοί")
}

type fakeLLM struct {
	post   string
	called bool
}

func (f *fakeLLM) GeneratePost(_ context.Context, _ string) (string, error) {
	f.called = true
	return f.post, nil
}

func assertContains(t *testing.T, s, substr string) {
	t.Helper()
	if !strings.Contains(s, substr) {
		t.Fatalf("expected %q to contain %q", s, substr)
	}
}
