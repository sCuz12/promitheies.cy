package datagovcy

import (
	"reflect"
	"testing"
)

func TestParseResources_CurrentDrupalLayout(t *testing.T) {
	body := `
		<div class="dataset-resource views-row">
			<a href="/index.php/el/resource/awards-2026" class="resource-link"><span>csv</span></a>
			<a href="/index.php/en/resource/5230"><span>Awarded contracts</span> 2026</a>
			<a class="fa-download" href="/sites/default/files/awards%202026.csv" download>Download</a>
		</div>
		<div class="dataset-resource views-row">
			<a href="/index.php/en/resource/4304">Awarded contracts 2025</a>
			<a download="download" href="https://files.example/awards-2025.csv">Download</a>
		</div>`

	got, err := parseResources("https://www.data.gov.cy", body)
	if err != nil {
		t.Fatal(err)
	}
	want := []Resource{
		{ID: "/index.php/en/resource/5230", Title: "Awarded contracts 2026", DownloadURL: "https://www.data.gov.cy/sites/default/files/awards%202026.csv"},
		{ID: "/index.php/en/resource/4304", Title: "Awarded contracts 2025", DownloadURL: "https://files.example/awards-2025.csv"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("resources mismatch\n got: %#v\nwant: %#v", got, want)
	}
}

func TestParseResources_LegacyLayout(t *testing.T) {
	body := `<a class="resource" href="/en/resource/123"><strong>Awards</strong> 2024</a>`

	got, err := parseResources("https://www.data.gov.cy/", body)
	if err != nil {
		t.Fatal(err)
	}
	want := []Resource{{
		ID:          "/en/resource/123",
		Title:       "Awards 2024",
		DownloadURL: "https://www.data.gov.cy/en/resource/123/download/file",
	}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("resources mismatch\n got: %#v\nwant: %#v", got, want)
	}
}

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
