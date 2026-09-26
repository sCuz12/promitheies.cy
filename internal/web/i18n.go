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
	Lang             Lang
	Data             any
	path             string
	rawQuery         string
	newsletterResult string
}

func newPage(lang Lang, r *http.Request, data any) Page {
	return Page{
		Lang: lang, Data: data, path: r.URL.Path, rawQuery: r.URL.RawQuery,
		newsletterResult: r.URL.Query().Get("newsletter"),
	}
}

func (p Page) NewsletterJoined() bool  { return p.newsletterResult == "joined" }
func (p Page) NewsletterInvalid() bool { return p.newsletterResult == "invalid" }

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
	"nav.authorities":  {LangEL: "Αρχές", LangEN: "Authorities"},
	"nav.contractors":  {LangEL: "Ανάδοχοι", LangEN: "Contractors"},
	"nav.open_tenders": {LangEL: "Διαγωνισμοί", LangEN: "Open tenders"},
	"nav.search":       {LangEL: "Αναζήτηση", LangEN: "Search"},
	"nav.toggle":       {LangEL: "Μενού", LangEN: "Menu"},

	// language switch
	"lang.label.el": {LangEL: "ΕΛ", LangEN: "ΕΛ"},
	"lang.label.en": {LangEL: "EN", LangEN: "EN"},

	// footer
	"footer.attribution": {
		LangEL: "Τα δεδομένα προέρχονται από το data.gov.cy και το TED (Tenders Electronic Daily). Δεν αποτελεί επίσημη κρατική υπηρεσία. πρόκειται για ανεξάρτητο δημόσιο μητρώο δημοσιοποιημένων στοιχείων συμβάσεων.",
		LangEN: "Data from data.gov.cy and TED (Tenders Electronic Daily). Not an official government service. an independent public register of disclosed procurement records.",
	},
	"footer.follow": {LangEL: "@sCuz123 στο X", LangEN: "@sCuz123 on X"},

	// page titles
	"title.home":         {LangEL: "Επισκόπηση", LangEN: "Dashboard"},
	"title.authorities":  {LangEL: "Αναθέτουσες Αρχές", LangEN: "Authorities"},
	"title.contractors":  {LangEL: "Ανάδοχοι", LangEN: "Contractors"},
	"title.open_tenders": {LangEL: "Ανοιχτοί Διαγωνισμοί", LangEN: "Open Tenders"},
	"title.search":       {LangEL: "Αναζήτηση", LangEN: "Search"},

	// breadcrumbs
	"breadcrumb.registry":     {LangEL: "Μητρώο", LangEN: "Registry"},
	"breadcrumb.authorities":  {LangEL: "Αναθέτουσες Αρχές", LangEN: "Authorities"},
	"breadcrumb.contractors":  {LangEL: "Ανάδοχοι", LangEN: "Contractors"},
	"breadcrumb.open_tenders": {LangEL: "Ανοιχτοί Διαγωνισμοί", LangEN: "Open Tenders"},
	"breadcrumb.search":       {LangEL: "Αναζήτηση", LangEN: "Search"},

	// home hero
	"home.eyebrow": {LangEL: "Δημόσιο μητρώο · data.gov.cy + TED", LangEN: "Public register · data.gov.cy + TED"},
	"home.h1":      {LangEL: "Οι δημόσιες συμβάσεις της Κύπρου, σε ένα σημείο", LangEN: "Cyprus public procurement, in one place"},
	"home.sub":     {LangEL: "Κάθε διαγωνισμός, ανάθεση και ανάδοχος που δημοσιοποιεί το κράτος. τεκμηριωμένα και δωρεάν.", LangEN: "Every tender, award and contractor the state discloses. searchable, sourced, and free."},
	"home.telegram_cta": {
		LangEL: "Λάβετε alerts στο Telegram",
		LangEN: "Get alerts on Telegram",
	},
	"newsletter.kicker": {
		LangEL: "Δωρεάν εβδομαδιαία ενημέρωση",
		LangEN: "Free weekly digest",
	},
	"newsletter.title": {
		LangEL: "Όλοι οι νέοι διαγωνισμοί. Ένα email. Κάθε εβδομάδα.",
		LangEN: "Every new tender. One email. Every week.",
	},
	"newsletter.body": {
		LangEL: "Λάβετε τους νέους διαγωνισμούς που εντοπίζει το Symvaseis.cy, με προθεσμίες, αξίες και συνδέσμους στις επίσημες προκηρύξεις.",
		LangEN: "Get the new tenders tracked by Symvaseis.cy, with deadlines, values and links to the official notices.",
	},
	"newsletter.email":       {LangEL: "1. Email", LangEN: "1. Email"},
	"newsletter.placeholder": {LangEL: "name@example.com", LangEN: "name@example.com"},
	"newsletter.sectors_label": {
		LangEL: "2. Ποιοι τομείς σας ενδιαφέρουν; (προαιρετικό)",
		LangEN: "2. Which sectors interest you? (optional)",
	},
	"newsletter.sectors_hint": {
		LangEL: "Θα φιλτράρουμε αυτόματα τους διαγωνισμούς σε αυτούς τους τομείς (CPV). Αφήστε το κενό για όλους τους τομείς.",
		LangEN: "We'll automatically filter tenders to these sectors (CPV). Leave blank for every sector.",
	},
	"newsletter.min_value_label": {
		LangEL: "3. Ελάχιστη αξία σύμβασης",
		LangEN: "3. Minimum contract value",
	},
	"newsletter.min_value.any":          {LangEL: "Οποιαδήποτε αξία", LangEN: "Any value"},
	"newsletter.min_value.eu_threshold": {LangEL: "όριο ΕΕ", LangEN: "EU threshold"},
	"newsletter.submit":                 {LangEL: "Εγγραφή δωρεάν", LangEN: "Join free"},
	"newsletter.consent": {
		LangEL: "Με την εγγραφή συμφωνείτε να λαμβάνετε την εβδομαδιαία ενημέρωση. Μπορείτε να διαγραφείτε οποτεδήποτε.",
		LangEN: "By joining, you agree to receive the weekly digest. You can unsubscribe at any time.",
	},
	"newsletter.success": {
		LangEL: "Είστε στη λίστα. Θα σας ενημερώσουμε μόλις ξεκινήσει η εβδομαδιαία αποστολή.",
		LangEN: "You're on the list. We'll let you know when the weekly digest launches.",
	},
	"newsletter.invalid": {
		LangEL: "Εισάγετε μια έγκυρη διεύθυνση email.",
		LangEN: "Enter a valid email address.",
	},
	"newsletter.website":         {LangEL: "Ιστότοπος", LangEN: "Website"},
	"home.open.kicker":           {LangEL: "Ευκαιρίες τώρα", LangEN: "Opportunities now"},
	"home.open.title":            {LangEL: "Νέοι ανοιχτοί διαγωνισμοί", LangEN: "Latest open tenders"},
	"home.open.intro":            {LangEL: "Βρείτε ενεργές ευκαιρίες και μεταβείτε απευθείας στον τομέα που σας ενδιαφέρει.", LangEN: "Find active opportunities and jump straight to the sector relevant to you."},
	"home.open.viewall":          {LangEL: "Όλοι οι ανοιχτοί διαγωνισμοί", LangEN: "View all open tenders"},
	"home.open.viewall_count":    {LangEL: "Δείτε όλους τους ανοιχτούς διαγωνισμούς", LangEN: "Browse all open tenders"},
	"home.open.categories_label": {LangEL: "Κατηγορίες ανοιχτών διαγωνισμών", LangEN: "Open tender categories"},
	"home.open.tender_one":       {LangEL: "διαγωνισμός", LangEN: "tender"},
	"home.open.tender_many":      {LangEL: "διαγωνισμοί", LangEN: "tenders"},
	"home.open.published":        {LangEL: "Δημοσίευση", LangEN: "Published"},
	"home.open.official":         {LangEL: "Επίσημη προκήρυξη", LangEN: "Official notice"},
	"tender.official_notice":     {LangEL: "Προβολή επίσημης προκήρυξης", LangEN: "View official notice"},
	"deadline.days":              {LangEL: "ημέρες απομένουν", LangEN: "days remaining"},
	"deadline.one_day":           {LangEL: "1 ημέρα απομένει", LangEN: "1 day remaining"},
	"deadline.today":             {LangEL: "Λήγει σήμερα", LangEN: "Closes today"},
	"deadline.passed":            {LangEL: "Η προθεσμία έληξε", LangEN: "Deadline passed"},
	"deadline.closing_soon":      {LangEL: "Λήγει σύντομα", LangEN: "Closing soon"},
	"stat.tenders":               {LangEL: "διαγωνισμοί", LangEN: "tenders tracked"},
	"stat.awards":                {LangEL: "αναθέσεις", LangEN: "awards recorded"},
	"stat.value":                 {LangEL: "συνολική αξία αναθέσεων", LangEN: "total awarded value"},
	"section.recent":             {LangEL: "Πρόσφατες αναθέσεις", LangEN: "Recent awards"},

	// table headers
	"th.date":        {LangEL: "Ημερομηνία", LangEN: "Date"},
	"th.published":   {LangEL: "Δημοσίευση", LangEN: "Published"},
	"th.tender":      {LangEL: "Διαγωνισμός", LangEN: "Tender"},
	"th.title":       {LangEL: "Τίτλος", LangEN: "Title"},
	"th.authority":   {LangEL: "Αρχή", LangEN: "Authority"},
	"th.contractor":  {LangEL: "Ανάδοχος", LangEN: "Contractor"},
	"th.deadline":    {LangEL: "Προθεσμία", LangEN: "Deadline"},
	"th.value":       {LangEL: "Αξία", LangEN: "Value"},
	"th.estvalue":    {LangEL: "Εκτ. αξία", LangEN: "Est. value"},
	"th.source":      {LangEL: "Πηγή", LangEN: "Source"},
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

	// open tenders page
	"open_tenders.h1": {LangEL: "Ανοιχτοί διαγωνισμοί", LangEN: "Open tenders"},
	"open_tenders.subtitle": {
		LangEL: "Ανοιχτοί διαγωνισμοί που σχετίζονται με την Κύπρο και δημοσιεύονται στο TED. Εθνικοί διαγωνισμοί κάτω των ευρωπαϊκών ορίων ενδέχεται να μην περιλαμβάνονται.",
		LangEN: "Open Cyprus-related tenders published on TED. National below-threshold tenders may not be included.",
	},
	"open_tenders.updated":     {LangEL: "Τελευταία ενημέρωση TED:", LangEN: "Last TED update:"},
	"open_tenders.keywords.ph": {LangEL: "π.χ. φάρμακα, κατασκευές, υπηρεσίες καθαρισμού…", LangEN: "e.g. medicine, construction, cleaning services…"},
	"open_tenders.submit":      {LangEL: "Φιλτράρισμα", LangEN: "Filter"},
	"open_tenders.empty.noresults": {
		LangEL: "Δεν βρέθηκαν ανοιχτοί διαγωνισμοί.",
		LangEN: "No open tenders matched.",
	},
	"open_tenders.empty.tryfewer": {
		LangEL: "Δοκιμάστε λιγότερες λέξεις-κλειδιά ή καθαρίστε το φίλτρο τομέα.",
		LangEN: "Try fewer keywords, or clear the sector filter.",
	},

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
		LangEL: "CPV: το κοινό λεξιλόγιο της ΕΕ για δημόσιες συμβάσεις. ταξινομεί κάθε διαγωνισμό ανά αντικείμενο (π.χ. κατασκευές, ιατρικός εξοπλισμός).",
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
