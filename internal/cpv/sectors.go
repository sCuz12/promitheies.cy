package cpv

// Sector is a business-facing grouping of CPV divisions. It lets a
// non-technical bidder pick "my industry" during onboarding instead of
// memorizing CPV codes; the onboarding flow resolves the chosen sectors to
// their Divisions to build a CPV filter automatically.
type Sector struct {
	Key       string
	NameEN    string
	NameEL    string
	Divisions []string
}

// NameLang returns the sector name in the given language ("el" for Greek,
// anything else falls back to English).
func (s Sector) NameLang(lang string) string {
	if lang == "el" && s.NameEL != "" {
		return s.NameEL
	}
	return s.NameEN
}

// Sectors groups every CPV division (see Divisions) into a small set of
// industries a Cyprus SME bidder would recognize. Every division code
// appears in exactly one sector.
var Sectors = []Sector{
	{"construction", "Construction & Engineering", "Κατασκευές & Μηχανική",
		[]string{"45000000", "44000000", "43000000", "71000000"}},
	{"it_telecom", "IT, Telecoms & Electronics", "Πληροφορική, Τηλεπικοινωνίες & Ηλεκτρονικά",
		[]string{"48000000", "72000000", "32000000", "30000000", "64000000"}},
	{"healthcare", "Healthcare & Social Care", "Υγεία & Κοινωνική Φροντίδα",
		[]string{"33000000", "85000000"}},
	{"transport", "Transport & Logistics", "Μεταφορές & Εφοδιαστική",
		[]string{"34000000", "60000000", "63000000"}},
	{"energy_utilities", "Energy, Utilities & Environment", "Ενέργεια, Κοινή Ωφέλεια & Περιβάλλον",
		[]string{"09000000", "65000000", "41000000", "90000000"}},
	{"security_facilities", "Security, Facilities & Maintenance", "Ασφάλεια, Εγκαταστάσεις & Συντήρηση",
		[]string{"35000000", "50000000", "51000000"}},
	{"professional_services", "Professional & Business Services", "Επαγγελματικές & Επιχειρηματικές Υπηρεσίες",
		[]string{"79000000", "73000000", "66000000", "70000000", "75000000"}},
	{"food_hospitality", "Food, Hospitality & Consumer Goods", "Τρόφιμα, Φιλοξενία & Καταναλωτικά Αγαθά",
		[]string{"15000000", "55000000", "18000000", "39000000", "37000000"}},
	{"industrial_manufacturing", "Industrial, Raw Materials & Manufacturing", "Βιομηχανία, Πρώτες Ύλες & Μεταποίηση",
		[]string{"03000000", "14000000", "16000000", "19000000", "22000000", "24000000", "31000000", "38000000", "42000000", "76000000", "77000000"}},
	{"education_culture", "Education, Culture & Public Administration", "Εκπαίδευση, Πολιτισμός & Δημόσια Διοίκηση",
		[]string{"80000000", "92000000", "98000000"}},
}

var sectorByKey = func() map[string]Sector {
	m := make(map[string]Sector, len(Sectors))
	for _, sec := range Sectors {
		m[sec.Key] = sec
	}
	return m
}()

// DivisionsForSectorKeys resolves onboarding-wizard sector selections to
// their union of CPV division codes, silently dropping unrecognized keys.
// The result is deduplicated but not sorted.
func DivisionsForSectorKeys(keys []string) []string {
	seen := make(map[string]bool)
	var out []string
	for _, k := range keys {
		sec, ok := sectorByKey[k]
		if !ok {
			continue
		}
		for _, d := range sec.Divisions {
			if !seen[d] {
				seen[d] = true
				out = append(out, d)
			}
		}
	}
	return out
}
