// Package entities resolves authority and contractor names from ingested
// sources into canonical database records, since the same entity appears
// with different spellings/languages across data.gov.cy, TED, and OpenTender.
package entities

import (
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

// contractorSuffixes are legal-form tokens stripped when normalizing
// contractor names for matching, so "Acme Ltd" and "Acme Limited" resolve
// to the same normalized form.
var contractorSuffixes = []string{
	" ltd", " limited", " λτδ", " plc", " llc", " inc", " co", " company",
	" ε.π.ε", " επε", " α.ε", " αε", " ο.ε", " οε",
}

// NormalizeForMatching lowercases, strips diacritics (including Greek
// tonos marks), collapses whitespace, and trims. Two spellings that differ
// only by accent or case normalize to the same string.
func NormalizeForMatching(s string) string {
	s = norm.NFD.String(s)
	var b strings.Builder
	for _, r := range s {
		if unicode.Is(unicode.Mn, r) { // skip combining marks (accents/tonos)
			continue
		}
		b.WriteRune(unicode.ToLower(r))
	}
	return collapseWhitespace(strings.TrimSpace(b.String()))
}

// NormalizeContractorName applies NormalizeForMatching and additionally
// strips common legal-form suffixes so "Acme Ltd" and "Acme Limited" match.
func NormalizeContractorName(s string) string {
	n := NormalizeForMatching(s)
	for _, suffix := range contractorSuffixes {
		n = strings.TrimSuffix(n, suffix)
	}
	return strings.TrimSpace(n)
}

func collapseWhitespace(s string) string {
	fields := strings.Fields(s)
	return strings.Join(fields, " ")
}

var greekToLatin = map[rune]string{
	'α': "a", 'β': "b", 'γ': "g", 'δ': "d", 'ε': "e", 'ζ': "z", 'η': "i",
	'θ': "th", 'ι': "i", 'κ': "k", 'λ': "l", 'μ': "m", 'ν': "n", 'ξ': "x",
	'ο': "o", 'π': "p", 'ρ': "r", 'σ': "s", 'ς': "s", 'τ': "t", 'υ': "y",
	'φ': "f", 'χ': "ch", 'ψ': "ps", 'ω': "o",
}

// Slugify produces a URL-safe, ASCII slug from a name, transliterating
// Greek characters so authority/contractor URLs stay human-readable even
// when the only available name is in Greek.
func Slugify(s string) string {
	s = NormalizeForMatching(s)
	var b strings.Builder
	for _, r := range s {
		if latin, ok := greekToLatin[r]; ok {
			b.WriteString(latin)
			continue
		}
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
		default:
			b.WriteRune('-')
		}
	}
	slug := b.String()
	for strings.Contains(slug, "--") {
		slug = strings.ReplaceAll(slug, "--", "-")
	}
	return strings.Trim(slug, "-")
}
