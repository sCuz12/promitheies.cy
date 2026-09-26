package digest

import (
	"fmt"
	"html/template"
	"strconv"
	"strings"
	"time"

	"github.com/georgehadjisavvas/promitheies-cy/internal/cpv"
)

type tenderView struct {
	URL           string
	Title         string
	AuthorityName string
	CategoryName  string
	ValueLabel    string
	DeadlineLabel string
}

type digestView struct {
	Count   int
	Tenders []tenderView
}

var (
	htmlTemplates = map[string]*template.Template{
		"el": template.Must(template.New("digest_el").Parse(htmlTemplateEL)),
		"en": template.Must(template.New("digest_en").Parse(htmlTemplateEN)),
	}
)

const htmlTemplateEL = `<!DOCTYPE html>
<html>
<body style="font-family: sans-serif; color: #1a1a1a;">
  <p>{{.Count}} νέοι διαγωνισμοί που ταιριάζουν με τις προτιμήσεις σας:</p>
  {{range .Tenders}}
  <div style="margin-bottom: 16px; padding-bottom: 16px; border-bottom: 1px solid #ddd;">
    <p><a href="{{.URL}}"><strong>{{.Title}}</strong></a></p>
    {{if .AuthorityName}}<p>Αναθέτουσα αρχή: {{.AuthorityName}}</p>{{end}}
    {{if .CategoryName}}<p>Κατηγορία: {{.CategoryName}}</p>{{end}}
    {{if .ValueLabel}}<p>Εκτιμώμενη αξία: {{.ValueLabel}}</p>{{end}}
    {{if .DeadlineLabel}}<p>Προθεσμία: {{.DeadlineLabel}}</p>{{end}}
  </div>
  {{end}}
</body>
</html>`

const htmlTemplateEN = `<!DOCTYPE html>
<html>
<body style="font-family: sans-serif; color: #1a1a1a;">
  <p>{{.Count}} new tenders matching your preferences:</p>
  {{range .Tenders}}
  <div style="margin-bottom: 16px; padding-bottom: 16px; border-bottom: 1px solid #ddd;">
    <p><a href="{{.URL}}"><strong>{{.Title}}</strong></a></p>
    {{if .AuthorityName}}<p>Authority: {{.AuthorityName}}</p>{{end}}
    {{if .CategoryName}}<p>Category: {{.CategoryName}}</p>{{end}}
    {{if .ValueLabel}}<p>Estimated value: {{.ValueLabel}}</p>{{end}}
    {{if .DeadlineLabel}}<p>Deadline: {{.DeadlineLabel}}</p>{{end}}
  </div>
  {{end}}
</body>
</html>`

// RenderDigest builds the subject, HTML body, and plaintext body for a
// digest email in the given language ("el" or "en"), listing tenders in
// the order given.
func RenderDigest(lang string, tenders []Tender, publicBaseURL string) (subject, htmlBody, textBody string) {
	views := make([]tenderView, 0, len(tenders))
	for _, t := range tenders {
		views = append(views, tenderView{
			URL:           tenderURL(publicBaseURL, t.ID),
			Title:         tenderTitle(t.Title),
			AuthorityName: t.AuthorityName,
			CategoryName:  categoryName(lang, t.CPVDivision),
			ValueLabel:    formatValue(t.EstimatedValueEUR),
			DeadlineLabel: formatDeadline(t.Deadline),
		})
	}
	view := digestView{Count: len(views), Tenders: views}

	subject = subjectFor(lang, len(views))

	tmpl := htmlTemplates["en"]
	if lang == "el" {
		tmpl = htmlTemplates["el"]
	}
	var b strings.Builder
	if err := tmpl.Execute(&b, view); err != nil {
		panic(fmt.Sprintf("render digest html: %v", err))
	}
	htmlBody = b.String()

	textBody = renderText(lang, view)
	return subject, htmlBody, textBody
}

func subjectFor(lang string, count int) string {
	if lang == "el" {
		return fmt.Sprintf("%d νέοι διαγωνισμοί αυτή την εβδομάδα", count)
	}
	return fmt.Sprintf("%d new tenders this week", count)
}

func renderText(lang string, view digestView) string {
	var b strings.Builder
	if lang == "el" {
		fmt.Fprintf(&b, "%d νέοι διαγωνισμοί που ταιριάζουν με τις προτιμήσεις σας:\n\n", view.Count)
	} else {
		fmt.Fprintf(&b, "%d new tenders matching your preferences:\n\n", view.Count)
	}
	for _, t := range view.Tenders {
		fmt.Fprintf(&b, "%s\n%s\n", t.Title, t.URL)
		if t.AuthorityName != "" {
			fmt.Fprintf(&b, "%s\n", t.AuthorityName)
		}
		if t.CategoryName != "" {
			fmt.Fprintf(&b, "%s\n", t.CategoryName)
		}
		if t.ValueLabel != "" {
			fmt.Fprintf(&b, "%s\n", t.ValueLabel)
		}
		if t.DeadlineLabel != "" {
			fmt.Fprintf(&b, "%s\n", t.DeadlineLabel)
		}
		b.WriteString("\n")
	}
	return b.String()
}

func tenderTitle(title string) string {
	title = strings.TrimSpace(title)
	if title == "" {
		return "Untitled tender"
	}
	return title
}

func categoryName(lang string, division *string) string {
	if division == nil || *division == "" {
		return ""
	}
	return cpv.NameLang(lang, *division)
}

func formatValue(value *float64) string {
	if value == nil {
		return ""
	}
	return fmt.Sprintf("EUR %.2f", *value)
}

func formatDeadline(deadline *time.Time) string {
	if deadline == nil {
		return ""
	}
	return deadline.Format("2006-01-02")
}

func tenderURL(publicBaseURL string, id int64) string {
	base := strings.TrimRight(publicBaseURL, "/")
	if base == "" {
		return "/tender/" + strconv.FormatInt(id, 10)
	}
	return base + "/tender/" + strconv.FormatInt(id, 10)
}
