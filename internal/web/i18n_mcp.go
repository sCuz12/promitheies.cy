package web

import "maps"

// Translations for the public MCP documentation page. Kept in their own file
// so i18n.go stays focused on the shared UI strings; merged in at startup.
func init() {
	maps.Copy(strings_, mcpStrings)
}

var mcpStrings = map[string]map[Lang]string{
	"nav.footer.mcp":  {LangEL: "Σύνδεση με AI (MCP)", LangEN: "Connect your AI (MCP)"},
	"title.mcp":       {LangEL: "Σύνδεση με AI", LangEN: "Connect your AI"},
	"breadcrumb.docs": {LangEL: "Τεκμηρίωση", LangEN: "Docs"},

	"mcp.eyebrow": {LangEL: "MCP server · ανοιχτό · δωρεάν", LangEN: "MCP server · open · free"},
	"mcp.h1": {
		LangEL: "Ρωτήστε το AI σας για τις δημόσιες συμβάσεις της Κύπρου",
		LangEN: "Ask your AI about Cyprus public procurement",
	},
	"mcp.sub": {
		LangEL: "Συνδέστε το Claude, το Cursor ή οποιονδήποτε MCP client στο Symvaseis.cy και κάντε ερωτήσεις σε φυσική γλώσσα για διαγωνισμούς, αναθέσεις, αναθέτουσες αρχές και αναδόχους.",
		LangEN: "Connect Claude, Cursor or any MCP client to Symvaseis.cy and ask plain-language questions about tenders, awards, contracting authorities and contractors.",
	},

	"mcp.what.h": {LangEL: "Τι είναι", LangEN: "What it is"},
	"mcp.what.body": {
		LangEL: "Το Model Context Protocol (MCP) είναι ένα ανοιχτό πρότυπο που επιτρέπει σε βοηθούς AI να χρησιμοποιούν εξωτερικά δεδομένα. Ο server μας είναι μόνο για ανάγνωση, δεν απαιτεί λογαριασμό ή κλειδί και χρησιμοποιεί το Streamable HTTP transport.",
		LangEN: "The Model Context Protocol (MCP) is an open standard that lets AI assistants use external data. Our server is read-only, needs no account or API key, and uses the Streamable HTTP transport.",
	},

	"mcp.connect.h": {LangEL: "Σύνδεση", LangEN: "Connect"},
	"mcp.url.label": {LangEL: "Διεύθυνση server", LangEN: "Server URL"},
	"mcp.copy":      {LangEL: "Αντιγραφή", LangEN: "Copy"},
	"mcp.copied":    {LangEL: "Αντιγράφηκε", LangEN: "Copied"},
	"mcp.claude.h":  {LangEL: "Claude (claude.ai και Desktop)", LangEN: "Claude (claude.ai and Desktop)"},
	"mcp.claude.body": {
		LangEL: "Ανοίξτε Settings → Connectors → Add custom connector, δώστε ένα όνομα και επικολλήστε τη διεύθυνση του server.",
		LangEN: "Open Settings → Connectors → Add custom connector, give it a name and paste the server URL.",
	},
	"mcp.code.h": {LangEL: "Claude Code", LangEN: "Claude Code"},
	"mcp.code.body": {
		LangEL: "Εκτελέστε στο τερματικό:",
		LangEN: "Run in your terminal:",
	},
	"mcp.cursor.h": {LangEL: "Cursor και άλλοι clients", LangEN: "Cursor and other clients"},
	"mcp.cursor.body": {
		LangEL: "Προσθέστε τον server στο αρχείο ρυθμίσεων MCP του client (π.χ. .cursor/mcp.json):",
		LangEN: "Add the server to your client's MCP config file (e.g. .cursor/mcp.json):",
	},
	"mcp.inspector.h": {LangEL: "Δοκιμή με MCP Inspector", LangEN: "Try it with MCP Inspector"},
	"mcp.inspector.body": {
		LangEL: "Εκτελέστε την εντολή, επιλέξτε Transport Type: Streamable HTTP και επικολλήστε τη διεύθυνση.",
		LangEN: "Run the command, choose Transport Type: Streamable HTTP and paste the URL.",
	},

	"mcp.examples.h": {LangEL: "Παραδείγματα ερωτήσεων", LangEN: "Example questions"},
	"mcp.example.1": {
		LangEL: "Ποιοι ανοιχτοί διαγωνισμοί πληροφορικής στην Κύπρο λήγουν αυτόν τον μήνα;",
		LangEN: "Which open IT tenders in Cyprus close this month?",
	},
	"mcp.example.2": {
		LangEL: "Πόσα έχει ξοδέψει ο Δήμος Λευκωσίας σε έργα κατασκευών;",
		LangEN: "How much has the Nicosia municipality spent on construction contracts?",
	},
	"mcp.example.3": {
		LangEL: "Ποιοι ανάδοχοι κερδίζουν τις περισσότερες συμβάσεις από αυτή την αρχή;",
		LangEN: "Which contractors win the most awards from this authority?",
	},
	"mcp.example.4": {
		LangEL: "Βρες διαγωνισμούς για ασφαλτόστρωση και σύγκρινε τις εκτιμώμενες αξίες.",
		LangEN: "Find tenders for road resurfacing and compare their estimated values.",
	},

	"mcp.tools.h": {LangEL: "Διαθέσιμα εργαλεία", LangEN: "Available tools"},
	"mcp.tools.intro": {
		LangEL: "Όλα τα εργαλεία είναι μόνο για ανάγνωση. Οι περιγραφές είναι στα αγγλικά στον ίδιο τον server, ώστε να τις διαβάζουν σωστά τα μοντέλα.",
		LangEN: "All tools are read-only. Descriptions are in English on the server itself so models read them reliably.",
	},
	"mcp.tools.params":   {LangEL: "Παράμετροι", LangEN: "Parameters"},
	"mcp.tools.none":     {LangEL: "Καμία", LangEN: "None"},
	"mcp.tools.required": {LangEL: "υποχρεωτική", LangEN: "required"},

	"mcp.tool.search_tenders": {
		LangEL: "Αναζήτηση σε όλους τους δημόσιους διαγωνισμούς (ανοιχτούς και παλαιότερους) με βάση τον τίτλο, την αρχή και τον τομέα CPV. Τα νεότερα πρώτα.",
		LangEN: "Search all Cyprus public tenders (open and historical) by title, authority and CPV division. Newest first.",
	},
	"mcp.tool.list_open_tenders": {
		LangEL: "Λίστα διαγωνισμών που είναι ανοιχτοί για προσφορές αυτή τη στιγμή. Τα νεότερα πρώτα.",
		LangEN: "List currently open tender opportunities (deadline not yet passed) that businesses can still bid on. Newest first.",
	},
	"mcp.tool.get_tender": {
		LangEL: "Πλήρη στοιχεία ενός διαγωνισμού: αρχή, αξία, κατάσταση, διαδικασία, προθεσμία και πηγή.",
		LangEN: "Full details of one tender: authority, value, status, procedure type, deadline and source.",
	},
	"mcp.tool.list_authorities": {
		LangEL: "Αναθέτουσες αρχές (υπουργεία, δήμοι, οργανισμοί) ταξινομημένες κατά συνολική αξία αναθέσεων.",
		LangEN: "Contracting authorities (ministries, municipalities, agencies) ranked by total awarded contract value.",
	},
	"mcp.tool.get_authority": {
		LangEL: "Προφίλ δαπανών αρχής: σύνολα, κορυφαίοι ανάδοχοι, κατανομή ανά τομέα CPV και πρόσφατες αναθέσεις.",
		LangEN: "A contracting authority's spending profile: totals, top contractors, spend by CPV sector and recent awards.",
	},
	"mcp.tool.list_contractors": {
		LangEL: "Ανάδοχοι που κέρδισαν δημόσιες συμβάσεις, ταξινομημένοι κατά συνολική αξία.",
		LangEN: "Contractors that won public contracts, ranked by total awarded value.",
	},
	"mcp.tool.get_contractor": {
		LangEL: "Προφίλ αναδόχου: συνολική αξία, ποιες αρχές του αναθέτουν συμβάσεις και πρόσφατες αναθέσεις.",
		LangEN: "A contractor's profile: total won, which authorities award them contracts, and recent awards.",
	},

	"mcp.param.query": {
		LangEL: "Ελεύθερο κείμενο στον τίτλο (ελληνικά ή αγγλικά).",
		LangEN: "Free text matched against tender titles (Greek or English).",
	},
	"mcp.param.authority_slug": {
		LangEL: "Περιορισμός σε μία αρχή· τα slugs βρίσκονται με το list_authorities.",
		LangEN: "Restrict to one authority; get slugs from list_authorities.",
	},
	"mcp.param.cpv_division": {
		LangEL: "Διαίρεση CPV 2 ψηφίων, π.χ. 45 για κατασκευές, 72 για υπηρεσίες IT.",
		LangEN: "2-digit CPV division, e.g. 45 for construction, 72 for IT services.",
	},
	"mcp.param.limit": {
		LangEL: "Μέγιστος αριθμός αποτελεσμάτων (προεπιλογή 20, μέγιστο 50).",
		LangEN: "Maximum results (default 20, max 50).",
	},
	"mcp.param.id": {
		LangEL: "Αναγνωριστικό διαγωνισμού, από το search_tenders ή το list_open_tenders.",
		LangEN: "Tender id, as returned by search_tenders or list_open_tenders.",
	},
	"mcp.param.slug": {
		LangEL: "Το slug της αρχής ή του αναδόχου, όπως επιστρέφεται από τις λίστες.",
		LangEN: "The authority or contractor slug, as returned by the list tools.",
	},

	"mcp.notes.h": {LangEL: "Καλό είναι να γνωρίζετε", LangEN: "Good to know"},
	"mcp.notes.readonly": {
		LangEL: "Μόνο για ανάγνωση: δεν απαιτείται λογαριασμός ή κλειδί API και τίποτα δεν τροποποιείται.",
		LangEN: "Read-only: no account or API key required, and nothing is ever modified.",
	},
	"mcp.notes.rate": {
		LangEL: "Όριο χρήσης: περίπου 5 αιτήματα το δευτερόλεπτο ανά διεύθυνση IP, με ριπές έως 20. Αν το ξεπεράσετε, ο server απαντά με HTTP 429 και μπορείτε να ξαναδοκιμάσετε σε λίγο.",
		LangEN: "Rate limit: about 5 requests per second per IP address, with bursts of up to 20. Beyond that the server answers HTTP 429 and you can retry shortly.",
	},
	"mcp.notes.data": {
		LangEL: "Τα δεδομένα προέρχονται από το data.gov.cy και το TED. Οι τίτλοι είναι κυρίως στα ελληνικά και οι αξίες σε ευρώ.",
		LangEN: "Data comes from data.gov.cy and TED. Titles are mostly Greek and values are in euros.",
	},
	"mcp.notes.disclaimer": {
		LangEL: "Δεν αποτελεί επίσημη κρατική υπηρεσία. Επιβεβαιώστε τα στοιχεία στην αρχική προκήρυξη πριν υποβάλετε προσφορά.",
		LangEN: "Not an official government service. Confirm details on the original notice before bidding.",
	},
}
