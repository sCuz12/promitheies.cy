package datagovcy

import "testing"

func TestDiscoverResources_Live(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping live network test in -short mode")
	}
	resources, err := DiscoverResources("https://www.data.gov.cy", "https://www.data.gov.cy/en/dataset/780")
	if err != nil {
		t.Fatal(err)
	}
	if len(resources) < 10 {
		t.Errorf("got %d resources, want at least 10 (one per year since 2009)", len(resources))
	}
	for _, r := range resources {
		if r.DownloadURL == "" || r.Title == "" {
			t.Errorf("incomplete resource: %+v", r)
		}
	}
}
