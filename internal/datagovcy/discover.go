// Package datagovcy imports Cyprus's awarded-contracts dataset from
// data.gov.cy. The portal (EKAN, not CKAN) has no public API — resources
// are periodic per-year CSV files linked from the dataset's HTML page, so
// discovery parses that page rather than calling an API.
package datagovcy

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"time"
)

// Resource is one periodic CSV file linked from the dataset page, e.g. the
// "2025" or "2009-2012" award-contracts file.
type Resource struct {
	ID          string
	Title       string
	DownloadURL string
}

// resourceLinkPattern matches the dataset page's resource list rows, e.g.
// `<a href="/en/resource/4304">Δημόσιες Συμβάσεις που Κατακυρώθηκαν 2025</a>`.
var resourceLinkPattern = regexp.MustCompile(`<a href="(/en/resource/\d+)">([^<]*)</a>`)

// yearPattern filters resourceLinkPattern matches down to actual periodic
// data files (their titles always contain a year), excluding unrelated
// links elsewhere on the page that happen to match the href shape.
var yearPattern = regexp.MustCompile(`(19|20)\d{2}`)

// DiscoverResources fetches the dataset page at datasetURL and returns every
// periodic CSV resource it links to, with DownloadURL rooted at baseURL.
func DiscoverResources(baseURL, datasetURL string) ([]Resource, error) {
	client := &http.Client{Timeout: 30 * time.Second}

	req, err := http.NewRequest(http.MethodGet, datasetURL, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; PromitheiesCyBot/1.0)")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch dataset page: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch dataset page: unexpected status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read dataset page: %w", err)
	}

	var resources []Resource
	seen := map[string]bool{}
	for _, m := range resourceLinkPattern.FindAllStringSubmatch(string(body), -1) {
		path, title := m[1], m[2]
		if !yearPattern.MatchString(title) || seen[path] {
			continue
		}
		seen[path] = true
		resources = append(resources, Resource{
			ID:          path,
			Title:       title,
			DownloadURL: baseURL + path + "/download/file",
		})
	}

	if len(resources) == 0 {
		return nil, fmt.Errorf("no dataset resources found at %s — page structure may have changed", datasetURL)
	}
	return resources, nil
}
