package ted

import (
	"encoding/json"
	"os"
	"testing"
)

func TestParseNotice_RealFixture(t *testing.T) {
	data, err := os.ReadFile("testdata/search_response.json")
	if err != nil {
		t.Fatal(err)
	}

	var resp SearchResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		t.Fatal(err)
	}
	if len(resp.Notices) != 2 {
		t.Fatalf("got %d notices, want 2", len(resp.Notices))
	}

	n, err := ParseNotice(resp.Notices[0])
	if err != nil {
		t.Fatal(err)
	}

	if n.PublicationNumber != "292871-2016" {
		t.Errorf("PublicationNumber = %q", n.PublicationNumber)
	}
	if n.AuthorityName != "Υπουργείο Υγείας" {
		t.Errorf("AuthorityName = %q", n.AuthorityName)
	}
	if n.TitleEN != "Cyprus-Nicosia: Medical consumables" {
		t.Errorf("TitleEN = %q", n.TitleEN)
	}
	if len(n.CPVCodes) != 1 || n.CPVCodes[0] != "33140000" {
		t.Errorf("CPVCodes = %v", n.CPVCodes)
	}
	if n.Status != "open" {
		t.Errorf("Status = %q, want open (form-type=competition)", n.Status)
	}
	if n.PublishedAt == nil || n.PublishedAt.Format("2006-01-02") != "2016-08-24" {
		t.Errorf("PublishedAt = %v", n.PublishedAt)
	}
}

func TestDedupe(t *testing.T) {
	got := dedupe([]string{"45000000", "50340000", "45000000", "", "50340000"})
	want := []string{"45000000", "50340000"}
	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("got[%d]=%q want %q", i, got[i], want[i])
		}
	}
}

func TestFirstBuyerName_PrefersGreek(t *testing.T) {
	got := firstBuyerName(map[string][]string{
		"eng": {"Ministry of Health"},
		"ell": {"Υπουργείο Υγείας"},
	})
	if got != "Υπουργείο Υγείας" {
		t.Errorf("got %q, want Greek name preferred", got)
	}
}
