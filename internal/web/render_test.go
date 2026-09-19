package web

import (
	"bytes"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/georgehadjisavvas/promitheies-cy/internal/cpv"
)

// fakePages builds one representative data value per page template, used to
// verify every template renders in both supported languages without a live
// database.
func fakePages() map[string]any {
	val := 125000.50
	date := "2026-03-14"
	cpvDiv := "45000000"

	award := AwardRow{
		TenderTitle: "Προμήθεια ιατρικού εξοπλισμού", AuthoritySlug: "dimos-lemesou",
		AuthorityName: "Δήμος Λεμεσού", ContractorSlug: "acme-ltd", ContractorName: "ACME LTD",
		Value: &val, AwardDate: &date, CPVDivision: &cpvDiv,
	}
	entity := EntitySummary{ID: 1, Slug: "dimos-lemesou", Name: "Δήμος Λεμεσού", Total: 4200000, Count: 37}

	return map[string]any{
		"home": Stats{
			TotalTenders: 72742, TotalAwards: 14343, TotalValue: 980000000,
			OpenTenderCount:      42,
			BySource:             []SourceCount{{Source: "data.gov.cy", Count: 58000}, {Source: "TED", Count: 14343}},
			OpenTenderCategories: []OpenTenderCategory{{Code: cpvDiv, Count: 12}},
			LatestOpenTenders: []TenderRow{{
				ID: 99, Title: "Προμήθεια φαρμάκων", AuthoritySlug: "dimos-lemesou", AuthorityName: "Δήμος Λεμεσού",
				CPVDivision: &cpvDiv, EstimatedVal: &val, Status: "open", PublishedAt: &date, Deadline: &date, Source: "ted",
			}},
			RecentAwards: []AwardRow{award},
		},
		"authorities": []EntitySummary{entity},
		"authority": &Authority{
			EntitySummary: entity, Type: "Δήμος", Region: "Λεμεσός",
			TopContractors:  []EntitySummary{{ID: 2, Slug: "acme-ltd", Name: "ACME LTD", Total: 900000, Count: 5}},
			SectorBreakdown: []SectorBreakdown{{CPVDivision: cpvDiv, Total: 1500000, Count: 8}},
			RecentAwards:    []AwardRow{award},
		},
		"contractors": []EntitySummary{entity},
		"contractor": &Contractor{
			EntitySummary: entity,
			Authorities:   []EntitySummary{entity},
			RecentAwards:  []AwardRow{award},
		},
		"tender": &Tender{
			ID: 99, Title: "Προμήθεια φαρμάκων", AuthoritySlug: "dimos-lemesou", AuthorityName: "Δήμος Λεμεσού",
			CPVDivision: &cpvDiv, EstimatedVal: &val, Status: "open", PublishedAt: &date, Source: "ted",
		},
		"open_tenders": OpenTendersPageData{
			Query: "φάρμακα", CPVDivision: cpvDiv, Divisions: cpv.Divisions, LastUpdated: &date,
			Results: []TenderRow{{
				ID: 99, Title: "Προμήθεια φαρμάκων", AuthoritySlug: "dimos-lemesou", AuthorityName: "Δήμος Λεμεσού",
				CPVDivision: &cpvDiv, EstimatedVal: &val, Status: "open", PublishedAt: &date, Source: "ted",
			}},
		},
		"search": searchPageData{
			Query: "ασφαλτόστρωση", CPVDivision: cpvDiv, Divisions: cpv.Divisions,
			Results: []TenderRow{{
				ID: 100, Title: "Ασφαλτόστρωση δρόμου", AuthoritySlug: "dimos-lemesou", AuthorityName: "Δήμος Λεμεσού",
				CPVDivision: &cpvDiv, EstimatedVal: &val, Status: "open", PublishedAt: &date,
			}},
		},
	}
}

func TestPageTemplatesRenderInBothLanguages(t *testing.T) {
	templates, err := loadTemplates("../../templates")
	if err != nil {
		t.Fatalf("loadTemplates: %v", err)
	}

	for page, data := range fakePages() {
		for _, lang := range []Lang{LangEL, LangEN} {
			t.Run(page+"_"+string(lang), func(t *testing.T) {
				tmpl, ok := templates[page]
				if !ok {
					t.Fatalf("no template registered for page %q", page)
				}
				r := httptest.NewRequest("GET", "/"+page+"?lang="+string(lang), nil)
				var buf bytes.Buffer
				if err := tmpl.ExecuteTemplate(&buf, "layout", newPage(lang, r, data)); err != nil {
					t.Fatalf("execute %s (%s): %v", page, lang, err)
				}
				out := buf.String()
				if !strings.Contains(out, "<html") {
					t.Fatalf("%s (%s): output doesn't look like HTML", page, lang)
				}
				// Every page must carry the lang switcher and both nav labels.
				if !strings.Contains(out, `lang-switch`) {
					t.Fatalf("%s (%s): missing language switcher", page, lang)
				}
			})
		}
	}
}

func TestHomeRendersOpenTenderNavigation(t *testing.T) {
	templates, err := loadTemplates("../../templates")
	if err != nil {
		t.Fatalf("loadTemplates: %v", err)
	}

	r := httptest.NewRequest("GET", "/?lang=en", nil)
	var buf bytes.Buffer
	if err := templates["home"].ExecuteTemplate(&buf, "layout", newPage(LangEN, r, fakePages()["home"])); err != nil {
		t.Fatalf("execute home: %v", err)
	}

	out := buf.String()
	for _, want := range []string{
		"Latest open tenders",
		`href="/diagonismoi?cpv=45000000"`,
		`href="/tender/99"`,
		"View all open tenders (42)",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("home output missing %q", want)
		}
	}
}
