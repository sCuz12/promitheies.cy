package datagovcy

import (
	"os"
	"testing"
)

func TestParse_CurrentColumnOrder(t *testing.T) {
	f, err := os.Open("testdata/sample_current.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	awards, err := Parse(f)
	if err != nil {
		t.Fatal(err)
	}
	if len(awards) != 5 {
		t.Fatalf("got %d awards, want 5", len(awards))
	}

	a := awards[0]
	if a.CFTID != "7854195" {
		t.Errorf("CFTID = %q, want 7854195", a.CFTID)
	}
	if a.ContractorName != "GP Information Business Solutions Cyprus iBSC LTD" {
		t.Errorf("ContractorName = %q", a.ContractorName)
	}
	if a.AwardedValue == nil || *a.AwardedValue != 5891 {
		t.Errorf("AwardedValue = %v, want 5891", a.AwardedValue)
	}
	if a.EstimatedValue == nil || *a.EstimatedValue != 6000 {
		t.Errorf("EstimatedValue = %v, want 6000", a.EstimatedValue)
	}
	if len(a.CPVCodes) != 1 || a.CPVCodes[0] != "72611000" {
		t.Errorf("CPVCodes = %v, want [72611000]", a.CPVCodes)
	}
	if a.AwardDate == nil || a.AwardDate.Format("2006-01-02") != "2025-01-01" {
		t.Errorf("AwardDate = %v, want 2025-01-01", a.AwardDate)
	}
	// "Tenders Submission Deadline" is the literal string "N/A" on some rows.
	if a.Deadline != nil {
		t.Errorf("Deadline = %v, want nil for N/A", a.Deadline)
	}

	// Second row exercises a quoted field containing commas.
	b := awards[1]
	if b.Title == "" || b.ContractorName != "Man.ifesto enterprises LTD" {
		t.Errorf("row 2 not parsed correctly: %+v", b)
	}
}

func TestParse_AlternateColumnOrder(t *testing.T) {
	// This fixture has a different column order (AWARDDATE before CPVCODES
	// is swapped relative to sample_current.csv, plus an extra duplicate
	// "Date Published_1" column) — Parse must map by header name, not
	// position, or this test will pull values from the wrong columns.
	f, err := os.Open("testdata/sample_altorder.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	awards, err := Parse(f)
	if err != nil {
		t.Fatal(err)
	}
	if len(awards) != 3 {
		t.Fatalf("got %d awards, want 3", len(awards))
	}

	a := awards[0]
	if a.CFTID != "4655693" {
		t.Errorf("CFTID = %q, want 4655693", a.CFTID)
	}
	if len(a.CPVCodes) != 6 {
		t.Errorf("CPVCodes = %v, want 6 codes", a.CPVCodes)
	}
	if a.CPVCodes[0] != "45000000" {
		t.Errorf("first CPV code = %q, want 45000000", a.CPVCodes[0])
	}
	if a.AwardDate == nil || a.AwardDate.Format("2006-01-02") != "2022-01-03" {
		t.Errorf("AwardDate = %v, want 2022-01-03", a.AwardDate)
	}
}

func TestParseCPVCodes(t *testing.T) {
	got := parseCPVCodes("45000000, 45212351, 71500000")
	want := []string{"45000000", "45212351", "71500000"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("got[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}
