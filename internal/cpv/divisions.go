// Package cpv holds the CPV (Common Procurement Vocabulary) reference data.
package cpv

// Division is a top-level (2-digit) CPV category, stored as an 8-digit code
// ending in "000000" per the EU CPV vocabulary convention.
type Division struct {
	Code          string
	DescriptionEN string
	DescriptionEL string
}

// Divisions is the standard EU CPV main-division list. Sub-codes referenced
// by ingested tenders that fall under one of these divisions are matched by
// their leading two digits at query time; individual sub-codes are not
// pre-seeded to keep this table small and low-maintenance.
var Divisions = []Division{
	{"03000000", "Agricultural, farming, fishing, forestry and related products", "Προϊόντα γεωργίας, κτηνοτροφίας, αλιείας, δασοκομίας και συναφή προϊόντα"},
	{"09000000", "Petroleum products, fuel, electricity and other sources of energy", "Πετρελαϊκά προϊόντα, καύσιμα, ηλεκτρισμός και άλλες πηγές ενέργειας"},
	{"14000000", "Mining, basic metals and related products", "Εξορυκτικά, βασικά μέταλλα και συναφή προϊόντα"},
	{"15000000", "Food, beverages, tobacco and related products", "Τρόφιμα, ποτά, καπνός και συναφή προϊόντα"},
	{"16000000", "Agricultural machinery", "Γεωργικά μηχανήματα"},
	{"18000000", "Clothing, footwear, luggage articles and accessories", "Ενδύματα, υποδήματα, είδη αποσκευών και εξαρτήματα"},
	{"19000000", "Leather and textile fabrics, plastic and rubber materials", "Δέρμα και υφαντουργικά υφάσματα, πλαστικά και ελαστικά υλικά"},
	{"22000000", "Printed matter and related products", "Έντυπο υλικό και συναφή προϊόντα"},
	{"24000000", "Chemical products", "Χημικά προϊόντα"},
	{"30000000", "Office and computing machinery, equipment and supplies except furniture and software packages", "Μηχανήματα γραφείου και υπολογιστές, εξοπλισμός και προμήθειες εκτός επίπλων και πακέτων λογισμικού"},
	{"31000000", "Electrical machinery, apparatus, equipment and consumables; lighting", "Ηλεκτρολογικός εξοπλισμός, μηχανήματα, συσκευές και αναλώσιμα· φωτισμός"},
	{"32000000", "Radio, television, communication, telecommunication and related equipment", "Ραδιοτηλεοπτικός, επικοινωνιακός και τηλεπικοινωνιακός εξοπλισμός και συναφή είδη"},
	{"33000000", "Medical equipments, pharmaceuticals and personal care products", "Ιατρικός εξοπλισμός, φαρμακευτικά προϊόντα και προϊόντα προσωπικής φροντίδας"},
	{"34000000", "Transport equipment and auxiliary products to transportation", "Εξοπλισμός μεταφορών και βοηθητικά προϊόντα μεταφορών"},
	{"35000000", "Security, fire-fighting, police and defence equipment", "Εξοπλισμός ασφάλειας, πυρόσβεσης, αστυνόμευσης και άμυνας"},
	{"37000000", "Musical instruments, sport goods, games, toys, handicraft, art materials and accessories", "Μουσικά όργανα, αθλητικά είδη, παιχνίδια, είδη χειροτεχνίας, καλλιτεχνικά υλικά και εξαρτήματα"},
	{"38000000", "Laboratory, optical and precision equipments (excl. glasses)", "Εργαστηριακός, οπτικός εξοπλισμός και εξοπλισμός ακριβείας (εκτός γυαλιών)"},
	{"39000000", "Furniture (incl. office furniture), furnishings, domestic appliances (excl. lighting) and cleaning products", "Έπιπλα (και έπιπλα γραφείου), επιπλώσεις, οικιακές συσκευές (εκτός φωτισμού) και προϊόντα καθαρισμού"},
	{"41000000", "Collected and purified water", "Συλλεγόμενο και καθαρισμένο νερό"},
	{"42000000", "Industrial machinery", "Βιομηχανικά μηχανήματα"},
	{"43000000", "Machinery for mining, quarrying, construction equipment", "Μηχανήματα εξόρυξης, λατόμευσης και εξοπλισμός κατασκευών"},
	{"44000000", "Construction structures and materials; auxiliary products to construction (except electric apparatus)", "Κατασκευαστικές δομές και υλικά· βοηθητικά προϊόντα κατασκευών (εκτός ηλεκτρικών συσκευών)"},
	{"45000000", "Construction work", "Κατασκευαστικές εργασίες"},
	{"48000000", "Software package and information systems", "Πακέτα λογισμικού και συστήματα πληροφορικής"},
	{"50000000", "Repair and maintenance services", "Υπηρεσίες επισκευής και συντήρησης"},
	{"51000000", "Installation services (except software)", "Υπηρεσίες εγκατάστασης (εκτός λογισμικού)"},
	{"55000000", "Hotel, restaurant and retail trade services", "Υπηρεσίες ξενοδοχείων, εστιατορίων και λιανικού εμπορίου"},
	{"60000000", "Transport services (excl. waste transport)", "Υπηρεσίες μεταφορών (εκτός μεταφοράς αποβλήτων)"},
	{"63000000", "Supporting and auxiliary transport services; travel agencies services", "Βοηθητικές και επικουρικές υπηρεσίες μεταφορών· ταξιδιωτικά πρακτορεία"},
	{"64000000", "Postal and telecommunications services", "Ταχυδρομικές και τηλεπικοινωνιακές υπηρεσίες"},
	{"65000000", "Public utilities", "Υπηρεσίες κοινής ωφέλειας"},
	{"66000000", "Financial and insurance services", "Χρηματοπιστωτικές και ασφαλιστικές υπηρεσίες"},
	{"70000000", "Real estate services", "Υπηρεσίες ακίνητης περιουσίας"},
	{"71000000", "Architectural, construction, engineering and inspection services", "Αρχιτεκτονικές, κατασκευαστικές, μηχανικές υπηρεσίες και υπηρεσίες επιθεώρησης"},
	{"72000000", "IT services: consulting, software development, Internet and support", "Υπηρεσίες πληροφορικής: παροχή συμβουλών, ανάπτυξη λογισμικού, διαδίκτυο και υποστήριξη"},
	{"73000000", "Research and development services and related consultancy services", "Υπηρεσίες έρευνας και ανάπτυξης και συναφείς συμβουλευτικές υπηρεσίες"},
	{"75000000", "Administration, defence and social security services", "Υπηρεσίες δημόσιας διοίκησης, άμυνας και κοινωνικής ασφάλισης"},
	{"76000000", "Services related to the oil and gas industry", "Υπηρεσίες σχετικές με τη βιομηχανία πετρελαίου και φυσικού αερίου"},
	{"77000000", "Agricultural, forestry, horticultural, aquacultural and apicultural services", "Γεωργικές, δασοκομικές, κηπουρικές, υδατοκαλλιεργητικές και μελισσοκομικές υπηρεσίες"},
	{"79000000", "Business services: law, marketing, consulting, recruitment, printing and security", "Επιχειρηματικές υπηρεσίες: νομικά, μάρκετινγκ, παροχή συμβουλών, προσλήψεις, εκτυπώσεις και ασφάλεια"},
	{"80000000", "Education and training services", "Υπηρεσίες εκπαίδευσης και κατάρτισης"},
	{"85000000", "Health and social work services", "Υπηρεσίες υγείας και κοινωνικής μέριμνας"},
	{"90000000", "Sewage, refuse, cleaning and environmental services", "Υπηρεσίες αποχέτευσης, απορριμμάτων, καθαρισμού και περιβάλλοντος"},
	{"92000000", "Recreational, cultural and sporting services", "Ψυχαγωγικές, πολιτιστικές και αθλητικές υπηρεσίες"},
	{"98000000", "Other community, social and personal services", "Άλλες κοινοτικές, κοινωνικές και ατομικές υπηρεσίες"},
}

// DivisionCode returns the 2-digit division code (e.g. "45") a full CPV code
// belongs to, used to bucket a tender's specific CPV code into its division
// when the tender doesn't carry a division-level code itself.
func DivisionCode(fullCode string) string {
	if len(fullCode) < 2 {
		return fullCode
	}
	return fullCode[:2]
}

var byCode = func() map[string]Division {
	m := make(map[string]Division, len(Divisions))
	for _, d := range Divisions {
		m[d.Code] = d
	}
	return m
}()

// Name returns the English description for a division code (e.g.
// "45000000" -> "Construction work"), or the code itself if unrecognized.
func Name(code string) string {
	if d, ok := byCode[code]; ok {
		return d.DescriptionEN
	}
	return code
}

// NameLang returns the division description in the given language ("el" for
// Greek, anything else falls back to English), or the code itself if
// unrecognized.
func NameLang(lang, code string) string {
	d, ok := byCode[code]
	if !ok {
		return code
	}
	if lang == "el" && d.DescriptionEL != "" {
		return d.DescriptionEL
	}
	return d.DescriptionEN
}
