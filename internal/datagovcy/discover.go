// Package datagovcy imports Cyprus's awarded-contracts dataset from
// data.gov.cy. The portal (EKAN, not CKAN) has no public API — resources
// are periodic per-year CSV files linked from the dataset's HTML page, so
// discovery parses that page rather than calling an API.
package datagovcy

import (
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// Resource is one periodic CSV file linked from the dataset page, e.g. the
// "2025" or "2009-2012" award-contracts file.
type Resource struct {
	ID          string
	Title       string
	DownloadURL string
}

var (
	// Parse anchors in two steps so href and download attributes can appear in
	// either order and the visible title may contain small bits of markup.
	anchorPattern     = regexp.MustCompile(`(?is)<a\b([^>]*)>(.*?)</a>`)
	hrefPattern       = regexp.MustCompile(`(?is)\bhref\s*=\s*["']([^"']+)["']`)
	downloadPattern   = regexp.MustCompile(`(?i)(?:^|\s)download(?:\s|=|$)`)
	resourceIDPattern = regexp.MustCompile(`(?i)/(?:en|el)/resource/(\d+)(?:[/?#]|$)`)
	htmlTagPattern    = regexp.MustCompile(`(?is)<[^>]+>`)
)

// yearPattern filters resource anchors down to actual periodic data files
// (their titles always contain a year), excluding unrelated resource links.
var yearPattern = regexp.MustCompile(`(19|20)\d{2}`)

// DiscoverResources fetches the dataset page at datasetURL and returns every
// periodic CSV resource it links to, with DownloadURL rooted at baseURL.
func DiscoverResources(baseURL, datasetURL string) ([]Resource, error) {
	client := &http.Client{Timeout: 30 * time.Second}

	req, err := http.NewRequest(http.MethodGet, datasetURL, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; SymvasisCyBot/1.0)")

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

	resources, err := parseResources(baseURL, string(body))
	if err != nil {
		return nil, err
	}

	if len(resources) == 0 {
		return nil, fmt.Errorf("no dataset resources found at %s — page structure may have changed", datasetURL)
	}
	return resources, nil
}

// parseResources supports both portal layouts seen in production. The older
// layout linked /en/resource/{id} and required a synthesized /download/file
// URL. Drupal 11 links /index.php/en/resource/{id} and exposes a direct file
// link in the following anchor.
func parseResources(baseURL, body string) ([]Resource, error) {
	base, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("parse base URL: %w", err)
	}

	var resources []Resource
	seen := map[string]bool{}
	current := -1
	for _, match := range anchorPattern.FindAllStringSubmatch(body, -1) {
		attrs, content := match[1], match[2]
		hrefMatch := hrefPattern.FindStringSubmatch(attrs)
		if hrefMatch == nil {
			continue
		}
		href := html.UnescapeString(hrefMatch[1])
		title := cleanLinkText(content)

		if idMatch := resourceIDPattern.FindStringSubmatch(href); idMatch != nil && yearPattern.MatchString(title) {
			id := idMatch[1]
			if seen[id] {
				current = -1
				continue
			}
			seen[id] = true
			resources = append(resources, Resource{
				ID:    href,
				Title: title,
			})
			current = len(resources) - 1
			continue
		}

		if current >= 0 && resources[current].DownloadURL == "" && downloadPattern.MatchString(attrs) {
			resources[current].DownloadURL = resolveURL(base, href)
			current = -1
		}
	}

	for i := range resources {
		if resources[i].DownloadURL == "" {
			resources[i].DownloadURL = resolveURL(base, resources[i].ID+"/download/file")
		}
	}
	return resources, nil
}

func resolveURL(base *url.URL, ref string) string {
	reference, err := url.Parse(ref)
	if err != nil {
		return ref
	}
	return base.ResolveReference(reference).String()
}

func cleanLinkText(s string) string {
	s = htmlTagPattern.ReplaceAllString(s, " ")
	s = html.UnescapeString(s)
	return strings.Join(strings.Fields(s), " ")
}
