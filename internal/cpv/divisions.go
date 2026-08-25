// Package cpv holds the CPV (Common Procurement Vocabulary) reference data.
package cpv

// Division is a top-level (2-digit) CPV category, stored as an 8-digit code
// ending in "000000" per the EU CPV vocabulary convention.
type Division struct {
	Code          string
	DescriptionEN string
}

// Divisions is the standard EU CPV main-division list. Sub-codes referenced
// by ingested tenders that fall under one of these divisions are matched by
// their leading two digits at query time; individual sub-codes are not
// pre-seeded to keep this table small and low-maintenance.
var Divisions = []Division{
	{"03000000", "Agricultural, farming, fishing, forestry and related products"},
	{"09000000", "Petroleum products, fuel, electricity and other sources of energy"},
	{"14000000", "Mining, basic metals and related products"},
	{"15000000", "Food, beverages, tobacco and related products"},
	{"16000000", "Agricultural machinery"},
	{"18000000", "Clothing, footwear, luggage articles and accessories"},
	{"19000000", "Leather and textile fabrics, plastic and rubber materials"},
	{"22000000", "Printed matter and related products"},
	{"24000000", "Chemical products"},
	{"30000000", "Office and computing machinery, equipment and supplies except furniture and software packages"},
	{"31000000", "Electrical machinery, apparatus, equipment and consumables; lighting"},
	{"32000000", "Radio, television, communication, telecommunication and related equipment"},
	{"33000000", "Medical equipments, pharmaceuticals and personal care products"},
	{"34000000", "Transport equipment and auxiliary products to transportation"},
	{"35000000", "Security, fire-fighting, police and defence equipment"},
	{"37000000", "Musical instruments, sport goods, games, toys, handicraft, art materials and accessories"},
	{"38000000", "Laboratory, optical and precision equipments (excl. glasses)"},
	{"39000000", "Furniture (incl. office furniture), furnishings, domestic appliances (excl. lighting) and cleaning products"},
	{"41000000", "Collected and purified water"},
	{"42000000", "Industrial machinery"},
	{"43000000", "Machinery for mining, quarrying, construction equipment"},
	{"44000000", "Construction structures and materials; auxiliary products to construction (except electric apparatus)"},
	{"45000000", "Construction work"},
	{"48000000", "Software package and information systems"},
	{"50000000", "Repair and maintenance services"},
	{"51000000", "Installation services (except software)"},
	{"55000000", "Hotel, restaurant and retail trade services"},
	{"60000000", "Transport services (excl. waste transport)"},
	{"63000000", "Supporting and auxiliary transport services; travel agencies services"},
	{"64000000", "Postal and telecommunications services"},
	{"65000000", "Public utilities"},
	{"66000000", "Financial and insurance services"},
	{"70000000", "Real estate services"},
	{"71000000", "Architectural, construction, engineering and inspection services"},
	{"72000000", "IT services: consulting, software development, Internet and support"},
	{"73000000", "Research and development services and related consultancy services"},
	{"75000000", "Administration, defence and social security services"},
	{"76000000", "Services related to the oil and gas industry"},
	{"77000000", "Agricultural, forestry, horticultural, aquacultural and apicultural services"},
	{"79000000", "Business services: law, marketing, consulting, recruitment, printing and security"},
	{"80000000", "Education and training services"},
	{"85000000", "Health and social work services"},
	{"90000000", "Sewage, refuse, cleaning and environmental services"},
	{"92000000", "Recreational, cultural and sporting services"},
	{"98000000", "Other community, social and personal services"},
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
