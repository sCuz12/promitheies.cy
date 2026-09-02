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
	req := DraftRequest{
		Cadence:  CadenceDaily,
		Platform: PlatformX,
		Window:   Window{Start: time.Date(2026, 9, 2, 0, 0, 0, 0, time.UTC), End: time.Date(2026, 9, 3, 0, 0, 0, 0, time.UTC)},
		Tenders: []Tender{{
			ID:            42,
			Title:         "Road works",
			AuthorityName: "Ministry of Transport",
			ExternalID:    "123456-2026",
		}},
	}

	md := MarkdownDraft(req, []string{
		"First tender post",
		"Second tender post",
		"Third tender post",
	})

	assertContains(t, md, "## Copy options")
	assertContains(t, md, "### Option 1")
	assertContains(t, md, "```text\nFirst tender post\n```")
	assertContains(t, md, "### Option 3")
	assertContains(t, md, "#42 TED 123456-2026: Road works")
	assertContains(t, md, "Source: TED")
}

func TestBuildPromptIncludesOnlyTEDTenderFacts(t *testing.T) {
	value := 12500.0
	deadline := time.Date(2026, 9, 20, 0, 0, 0, 0, time.UTC)
	cpv := "45000000"
	req := DraftRequest{
		Cadence:       CadenceDaily,
		Platform:      PlatformX,
		Window:        Window{Start: time.Date(2026, 9, 2, 0, 0, 0, 0, time.UTC), End: time.Date(2026, 9, 3, 0, 0, 0, 0, time.UTC)},
		PublicBaseURL: "https://promitheies.cy/",
		Tenders: []Tender{{
			ID:                42,
			Title:             "Road works",
			AuthorityName:     "Ministry of Transport",
			CPVDivision:       &cpv,
			EstimatedValueEUR: &value,
			Deadline:          &deadline,
		}},
	}

	prompt := BuildPrompt(req)

	assertContains(t, prompt, "Write exactly 3 alternative X posts")
	assertContains(t, prompt, "helps Cyprus businesses discover public tender opportunities")
	assertContains(t, prompt, "The goal of this post is awareness")
	assertContains(t, prompt, "choose the strongest factual angle")
	assertContains(t, prompt, "New open TED tenders: 1")
	assertContains(t, prompt, "Open tenders URL: https://promitheies.cy/diagonismoi")
	assertContains(t, prompt, "title: Road works")
	assertContains(t, prompt, "estimated value EUR 12500.00")
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

	posts, err := GenerateDraft(context.Background(), llm, req)
	if err != nil {
		t.Fatalf("GenerateDraft returned error: %v", err)
	}
	if llm.called {
		t.Fatal("LLM was called for no-tender draft")
	}
	if got, want := len(posts), 1; got != want {
		t.Fatalf("got %d fallback posts, want %d", got, want)
	}
	assertContains(t, posts[0], "No new open TED tenders")
}

func TestGenerateDraftUsesLLMAndCleansOutput(t *testing.T) {
	llm := &fakeLLM{post: strings.Join([]string{
		"1. New tenders in Cyprus today. Check deadlines: /diagonismoi",
		"2. Cyprus SMEs: fresh public-sector opportunities are live. View: /diagonismoi",
		"3. New TED tender leads for Cyprus are ready to review. View: /diagonismoi",
	}, "\n")}
	req := DraftRequest{
		Cadence:  CadenceDaily,
		Platform: PlatformX,
		Window:   Window{Start: time.Date(2026, 9, 2, 0, 0, 0, 0, time.UTC), End: time.Date(2026, 9, 3, 0, 0, 0, 0, time.UTC)},
		Tenders:  []Tender{{ID: 1, Title: "Tender"}},
	}

	posts, err := GenerateDraft(context.Background(), llm, req)
	if err != nil {
		t.Fatalf("GenerateDraft returned error: %v", err)
	}
	if !llm.called {
		t.Fatal("LLM was not called")
	}
	if got, want := len(posts), DefaultOptionCount; got != want {
		t.Fatalf("got %d posts, want %d", got, want)
	}
	if strings.Contains(posts[0], "1.") {
		t.Fatalf("post still contains numbering: %q", posts[0])
	}
}

func TestParsePostOptionsRequiresThreeOptions(t *testing.T) {
	_, err := ParsePostOptions("1. One\n2. Two")
	if err == nil {
		t.Fatal("expected missing option to fail")
	}
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
