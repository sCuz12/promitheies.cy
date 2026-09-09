package datagovcy

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// dateLayout matches the portal's "DD-MM-YYYY HH:MM:SS" timestamps. Some
// columns (e.g. deadlines on very old rows) may carry only the date part;
// callers fall back to dateOnlyLayout when this fails.
const (
	dateLayout     = "02-01-2006 15:04:05"
	dateOnlyLayout = "02-01-2006"
)

// Award is one row of the awarded-contracts CSV, mapped from the source's
// column names (which vary in order, and occasionally in exact set, across
// yearly files — see Parse).
type Award struct {
	CFTID           string
	AuthorityName   string
	AuthorityTypeEN string
	Title           string
	PublishedAt     *time.Time
	Deadline        *time.Time
	ProcedureType   string
	EstimatedValue  *float64
	AwardedValue    *float64
	AwardDate       *time.Time
	ContractorName  string
	CPVCodes        []string
	RawRow          map[string]string
}

// Fetch downloads and parses the CSV at url, returning one Award per row.
// Rows with a blank CFTID (the source's stable per-record key) are skipped.
func Fetch(url string) ([]Award, error) {
	client := &http.Client{Timeout: 2 * time.Minute}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; SymvaseisCyBot/1.0)")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("download csv: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download csv: unexpected status %d", resp.StatusCode)
	}

	return Parse(resp.Body)
}

// Parse reads a data.gov.cy awarded-contracts CSV and maps each row by
// header name rather than column position, since column order (and the
// presence of the odd duplicate "Date Published_1" column) has been
// observed to vary between yearly files.
func Parse(r io.Reader) ([]Award, error) {
	reader := csv.NewReader(bomReader(r))
	reader.LazyQuotes = true
	reader.FieldsPerRecord = -1

	header, err := reader.Read()
	if err != nil {
		return nil, fmt.Errorf("read header: %w", err)
	}
	col := make(map[string]int, len(header))
	for i, name := range header {
		col[strings.TrimSpace(name)] = i
	}

	get := func(row []string, name string) string {
		i, ok := col[name]
		if !ok || i >= len(row) {
			return ""
		}
		return strings.TrimSpace(row[i])
	}

	var awards []Award
	for {
		row, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read row: %w", err)
		}

		cftid := get(row, "CFTID")
		if cftid == "" {
			continue
		}

		a := Award{
			CFTID:           cftid,
			AuthorityName:   get(row, "ORGANIZATIONNAME"),
			AuthorityTypeEN: get(row, "CA_TYPE"),
			Title:           get(row, "CFTTITLE"),
			PublishedAt:     parseDate(get(row, "Date Published")),
			Deadline:        parseDate(get(row, "Tenders Submission Deadline")),
			ProcedureType:   get(row, "Procedure El"),
			EstimatedValue:  parseAmount(get(row, "Estimated Value")),
			AwardedValue:    parseAmount(get(row, "AWARDEDCONTRVALUE")),
			AwardDate:       parseDate(get(row, "AWARDDATE")),
			ContractorName:  get(row, "EONAME"),
			CPVCodes:        parseCPVCodes(get(row, "CPVCODES")),
		}
		awards = append(awards, a)
	}
	return awards, nil
}

func parseDate(s string) *time.Time {
	if s == "" {
		return nil
	}
	if t, err := time.Parse(dateLayout, s); err == nil {
		return &t
	}
	if t, err := time.Parse(dateOnlyLayout, s); err == nil {
		return &t
	}
	return nil
}

func parseAmount(s string) *float64 {
	s = strings.ReplaceAll(s, ",", "")
	if s == "" {
		return nil
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return nil
	}
	return &v
}

func parseCPVCodes(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	codes := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			codes = append(codes, p)
		}
	}
	return codes
}

// bomReader strips a leading UTF-8 byte-order mark, which every observed
// data.gov.cy CSV export begins with.
func bomReader(r io.Reader) io.Reader {
	br := bufio.NewReader(r)
	bom, err := br.Peek(3)
	if err == nil && bom[0] == 0xEF && bom[1] == 0xBB && bom[2] == 0xBF {
		br.Discard(3)
	}
	return br
}
