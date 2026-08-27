package web

import (
	"net/http"
	"net/url"
)

// Lang is a supported UI language. Greek is the site default — this is a
// Cyprus public-records tool built for a Greek-speaking audience first.
type Lang string

const (
	LangEL Lang = "el"
	LangEN Lang = "en"

	langCookieName = "lang"
)

// resolveLang determines the active language for a request. An explicit
// ?lang= query parameter wins and is persisted to a cookie so it sticks on
// the next visit; otherwise a previously-set cookie is used; otherwise the
// site defaults to Greek.
func (s *Server) resolveLang(w http.ResponseWriter, r *http.Request) Lang {
	if q := r.URL.Query().Get("lang"); q == string(LangEL) || q == string(LangEN) {
		http.SetCookie(w, &http.Cookie{
			Name:     langCookieName,
			Value:    q,
			Path:     "/",
			MaxAge:   3600 * 24 * 365,
			SameSite: http.SameSiteLaxMode,
		})
		return Lang(q)
	}
	if c, err := r.Cookie(langCookieName); err == nil {
		if c.Value == string(LangEL) || c.Value == string(LangEN) {
			return Lang(c.Value)
		}
	}
	return LangEL
}

// Page wraps a page's own data together with the resolved request language,
// so both the shared layout (nav, footer, language switch) and the page
// content block can render translated text via .T / .Tip, while page-specific
// fields stay reachable under .Data.
type Page struct {
	Lang     Lang
	Data     any
	path     string
	rawQuery string
}

func newPage(lang Lang, r *http.Request, data any) Page {
	return Page{Lang: lang, Data: data, path: r.URL.Path, rawQuery: r.URL.RawQuery}
}

// T returns the UI string for key in the page's language.
func (p Page) T(key string) string { return T(p.Lang, key) }

// Tip returns the soft-tooltip explanation for key in the page's language,
// or "" if none is defined (callers skip rendering the tip icon in that case).
func (p Page) Tip(key string) string { return Tip(p.Lang, key) }

// IsEL reports whether the page is currently rendering in Greek.
func (p Page) IsEL() bool { return p.Lang == LangEL }

// OtherLang returns the language the switcher should offer.
func (p Page) OtherLang() Lang {
	if p.Lang == LangEL {
		return LangEN
	}
	return LangEL
}

// SwitchURL returns the URL for viewing the current page in lang ("el" or
// "en"), preserving any other query parameters (an active search, a
// selected sector, …). Takes a plain string so it can be called directly
// with a literal from templates.
func (p Page) SwitchURL(lang string) string {
	v, _ := url.ParseQuery(p.rawQuery)
	v.Set("lang", lang)
	return p.path + "?" + v.Encode()
}

// T looks up a UI string by key in the given language, falling back to Greek
// (the site default) and finally to the key itself if nothing is defined.
func T(lang Lang, key string) string {
	if m, ok := strings_[key]; ok {
		if s, ok := m[lang]; ok {
			return s
		}
		if s, ok := m[LangEL]; ok {
			return s
		}
	}
	return key
}

// Tip looks up a tooltip explanation by key, the same way T does, returning
// "" (not the key) when no tip is defined for that key.
func Tip(lang Lang, key string) string {
	if m, ok := tips[key]; ok {
		if s, ok := m[lang]; ok {
			return s
		}
		if s, ok := m[LangEL]; ok {
			return s
		}
	}
	return ""
}

// strings_ holds every translated UI string, keyed by a dotted identifier.
// Named with a trailing underscore to avoid shadowing the "strings" package.
var strings_ = map[string]map[Lang]string{
	// nav
	"nav.authorities": {LangEL: "Αρχές", LangEN: "Authorities"},
	"nav.contractors": {LangEL: "Ανάδοχοι", LangEN: "Contractors"},
	"nav.search":      {LangEL: "Αναζήτηση", LangEN: "Search"},

	// language switch
	"lang.label.el": {LangEL: "ΕΛ", LangEN: "ΕΛ"},
	"lang.label.en": {LangEL: "EN", LangEN: "EN"},

	// footer
	"footer.attribution": {
		LangEL: "Τα δεδομένα προέρχονται από το data.gov.cy και το TED (Tenders Electronic Daily). Δεν αποτελεί επίσημη κρατική υπηρεσία. πρόκειται για ανεξάρτητο δημόσιο μητρώο δημοσιοποιημένων στοιχείων συμβάσεων.",
		LangEN: "Data from data.gov.cy and TED (Tenders Electronic Daily). Not an official government service. an independent public register of disclosed procurement records.",
	},

	// page titles
	"title.home":        {LangEL: "Επισκόπηση", LangEN: "Dashboard"},
	"title.authorities": {LangEL: "Αναθέτουσες Αρχές", LangEN: "Authorities"},
	"title.contractors": {LangEL: "Ανάδοχοι", LangEN: "Contractors"},
	"title.search":      {LangEL: "Αναζήτηση", LangEN: "Search"},

	// breadcrumbs
	"breadcrumb.registry":    {LangEL: "Μητρώο", LangEN: "Registry"},
	"breadcrumb.authorities": {LangEL: "Αναθέτουσες Αρχές", LangEN: "Authorities"},
	"breadcrumb.contractors": {LangEL: "Ανάδοχοι", LangEN: "Contractors"},
	"breadcrumb.search":      {LangEL: "Αναζήτηση", LangEN: "Search"},

	// home hero
	"home.eyebrow":   {LangEL: "Δημόσιο μητρώο · data.gov.cy + TED", LangEN: "Public register · data.gov.cy + TED"},
	"home.h1":        {LangEL: "Οι δημόσιες συμβάσεις της Κύπρου, σε ένα σημείο", LangEN: "Cyprus public procurement, in one place"},
	"home.sub":       {LangEL: "Κάθε διαγωνισμός, ανάθεση και ανάδοχος που δημοσιοποιεί το κράτος. τεκμηριωμένα και δωρεάν.", LangEN: "Every tender, award and contractor the state discloses — searchable, sourced, and free."},
	"stat.tenders":   {LangEL: "διαγωνισμοί", LangEN: "tenders tracked"},
	"stat.awards":    {LangEL: "αναθέσεις", LangEN: "awards recorded"},
	"stat.value":     {LangEL: "συνολική αξία αναθέσεων", LangEN: "total awarded value"},
	"section.recent": {LangEL: "Πρόσφατες αναθέσεις", LangEN: "Recent awards"},

	// table headers
	"th.date":        {LangEL: "Ημερομηνία", LangEN: "Date"},
	"th.tender":      {LangEL: "Διαγωνισμός", LangEN: "Tender"},
	"th.title":       {LangEL: "Τίτλος", LangEN: "Title"},
	"th.authority":   {LangEL: "Αρχή", LangEN: "Authority"},
	"th.contractor":  {LangEL: "Ανάδοχος", LangEN: "Contractor"},
	"th.value":       {LangEL: "Αξία", LangEN: "Value"},
	"th.estvalue":    {LangEL: "Εκτ. αξία", LangEN: "Est. value"},
	"th.awards":      {LangEL: "Αναθέσεις", LangEN: "Awards"},
	"th.totalvalue":  {LangEL: "Συνολική αξία", LangEN: "Total value"},
	"th.rownum":      {LangEL: "Α/Α", LangEN: "#"},
	"th.sector":      {LangEL: "Τομέας", LangEN: "Sector"},
	"th.status":      {LangEL: "Κατάσταση", LangEN: "Status"},
	"th.cpvdivision": {LangEL: "Τομέας CPV", LangEN: "CPV division"},

	// authorities list
	"authorities.subtitle": {
		LangEL: "%d αναθέτουσες αρχές με ανατεθείσες συμβάσεις, σε κατάταξη κατά συνολική αξία.",
		LangEN: "%d contracting authorities with awarded contracts, ranked by total value.",
	},

	// contractors list
	"contractors.subtitle": {
		LangEL: "Κορυφαίοι %d ανάδοχοι κατά συνολική αξία ανάθεσης.",
		LangEN: "Top %d contractors by total awarded value.",
	},

	// authority profile
	"authority.file":            {LangEL: "Φάκελος αρχής αρ.", LangEN: "Authority file №"},
	"authority.stat.count":      {LangEL: "αναθέσεις", LangEN: "awards"},
	"authority.stat.total":      {LangEL: "συνολική αξία", LangEN: "total value"},
	"authority.topcontractors":  {LangEL: "Κορυφαίοι ανάδοχοι", LangEN: "Top contractors"},
	"authority.sectorbreakdown": {LangEL: "Δαπάνες ανά τομέα", LangEN: "Spend by sector"},

	// contractor profile
	"contractor.file":        {LangEL: "Φάκελος αναδόχου αρ.", LangEN: "Contractor file №"},
	"contractor.stat.count":  {LangEL: "κερδισμένες αναθέσεις", LangEN: "awards won"},
	"contractor.stat.total":  {LangEL: "συνολική αξία", LangEN: "total value"},
	"contractor.authorities": {LangEL: "Συνεργασία με αρχές", LangEN: "Authorities worked with"},

	// search page
	"search.h1":              {LangEL: "Αναζήτηση στο μητρώο διαγωνισμών", LangEN: "Search the tender register"},
	"search.keywords":        {LangEL: "Λέξεις-κλειδιά", LangEN: "Keywords"},
	"search.keywords.ph":     {LangEL: "π.χ. ασφαλτόστρωση δρόμου, ιατρικός εξοπλισμός…", LangEN: "e.g. road resurfacing, medical equipment…"},
	"search.sector":          {LangEL: "Τομέας", LangEN: "Sector"},
	"search.allsectors":      {LangEL: "Όλοι οι τομείς", LangEN: "All sectors"},
	"search.submit":          {LangEL: "Αναζήτηση", LangEN: "Search"},
	"search.empty.noresults": {LangEL: "Δεν βρέθηκαν διαγωνισμοί.", LangEN: "No tenders matched."},
	"search.empty.tryfewer":  {LangEL: "Δοκιμάστε λιγότερες λέξεις-κλειδιά ή καθαρίστε το φίλτρο τομέα.", LangEN: "Try fewer keywords, or clear the sector filter."},
	"search.empty.prompt":    {LangEL: "Αναζητήστε στο δημόσιο μητρώο διαγωνισμών.", LangEN: "Search the public tender register."},
	"search.empty.hint":      {LangEL: "Δοκιμάστε μια λέξη-κλειδί όπως «ασφαλτόστρωση δρόμου», ή φιλτράρετε ανά τομέα παραπάνω.", LangEN: "Try a keyword like “road resurfacing”, or filter by sector above."},

	// tender status badges
	"status.open":    {LangEL: "Ανοιχτός", LangEN: "Open"},
	"status.awarded": {LangEL: "Ανατέθηκε", LangEN: "Awarded"},
	"status.unknown": {LangEL: "Άγνωστο", LangEN: "Unknown"},
}

// tips holds short, plain-language explanations shown in a soft hover/focus
// bubble next to terms that aren't self-explanatory — procurement jargon,
// EU classification codes, or things about the data itself worth flagging.
var tips = map[string]map[Lang]string{
	"cpv": {
		LangEL: "CPV: το κοινό λεξιλόγιο της ΕΕ για δημόσιες συμβάσεις — ταξινομεί κάθε διαγωνισμό ανά αντικείμενο (π.χ. κατασκευές, ιατρικός εξοπλισμός).",
		LangEN: "CPV: the EU's Common Procurement Vocabulary. Classifies each tender by what it's actually for (e.g. construction, medical equipment).",
	},
	"authority.file": {
		LangEL: "Αριθμός φακέλου που αποδίδει αυτόματα αυτός ο ιστότοπος για εύκολη αναφορά.δεν είναι επίσημος κρατικός αριθμός μητρώου.",
		LangEN: "A reference number this site assigns automatically for easy citation. not an official government registry number.",
	},
	"contractor.file": {
		LangEL: "Αριθμός φακέλου που αποδίδει αυτόματα αυτός ο ιστότοπος για εύκολη αναφορά. δεν είναι επίσημος κρατικός αριθμός μητρώου.",
		LangEN: "A reference number this site assigns automatically for easy citation. not an official government registry number.",
	},
	"status.open": {
		LangEL: "Ο διαγωνισμός δέχεται ακόμη προσφορές. δεν έχει ανατεθεί.",
		LangEN: "The tender is still accepting bids. no contract has been awarded yet.",
	},
	"status.awarded": {
		LangEL: "Η σύμβαση έχει ήδη ανατεθεί σε ανάδοχο.",
		LangEN: "The contract has already been awarded to a contractor.",
	},
	"status.unknown": {
		LangEL: "Η πηγή δεδομένων δεν δηλώνει την τρέχουσα κατάσταση αυτού του διαγωνισμού.",
		LangEN: "The source data doesn't state this tender's current status.",
	},
	"stat.tenders": {
		LangEL: "Πλήθος διαγωνισμών που έχουν δημοσιευθεί σε μία τουλάχιστον από τις πηγές μας, είτε έχουν ανατεθεί είτε όχι.",
		LangEN: "Count of tenders published in at least one of our sources, whether or not they've been awarded yet.",
	},
	"stat.awards": {
		LangEL: "Πλήθος συμβάσεων που έχουν επίσημα ανατεθεί σε ανάδοχο, σύμφωνα με τα δημοσιευμένα στοιχεία.",
		LangEN: "Count of contracts formally awarded to a contractor, per the published records.",
	},
	"stat.value": {
		LangEL: "Άθροισμα της δηλωμένης αξίας ανάθεσης όπου είναι διαθέσιμη. δεν καλύπτει συμβάσεις χωρίς δημοσιευμένη αξία.",
		LangEN: "Sum of the disclosed award value where available. contracts published without a value aren't counted here.",
	},
	"authority.sectorbreakdown": {
		LangEL: "Κατανομή της συνολικής αξίας των αναθέσεων αυτής της αρχής ανά κατηγορία CPV.",
		LangEN: "How this authority's total award value splits across CPV categories.",
	},
}
