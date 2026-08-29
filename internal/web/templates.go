package web

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"github.com/georgehadjisavvas/promitheies-cy/internal/cpv"
)

var funcMap = template.FuncMap{
	"commas": func(n int64) string { return groupThousands(strconv.FormatInt(n, 10)) },
	"euro": func(v float64) string {
		return groupThousands(strconv.FormatFloat(v, 'f', 0, 64))
	},
	"inc":     func(i int) int { return i + 1 },
	"cpvName": func(lang Lang, code string) string { return cpv.NameLang(string(lang), code) },
}

// groupThousands inserts "," every three digits from the right of an
// unsigned integer string.
func groupThousands(s string) string {
	neg := strings.HasPrefix(s, "-")
	if neg {
		s = s[1:]
	}
	n := len(s)
	if n <= 3 {
		if neg {
			return "-" + s
		}
		return s
	}
	var b strings.Builder
	lead := n % 3
	if lead > 0 {
		b.WriteString(s[:lead])
	}
	for i := lead; i < n; i += 3 {
		if b.Len() > 0 {
			b.WriteByte(',')
		}
		b.WriteString(s[i : i+3])
	}
	out := b.String()
	if neg {
		return "-" + out
	}
	return out
}

func loadTemplates(dir string) (map[string]*template.Template, error) {
	pages := []string{"home", "authorities", "authority", "contractors", "contractor", "open_tenders", "search"}
	templates := make(map[string]*template.Template, len(pages))

	for _, page := range pages {
		t, err := template.New("layout").Funcs(funcMap).ParseFiles(
			dir+"/layout.html",
			dir+"/"+page+".html",
		)
		if err != nil {
			return nil, fmt.Errorf("parse templates for %s: %w", page, err)
		}
		templates[page] = t
	}
	return templates, nil
}

func (s *Server) render(w http.ResponseWriter, r *http.Request, page string, data any) {
	t, ok := s.templates[page]
	if !ok {
		http.Error(w, "template not found: "+page, http.StatusInternalServerError)
		return
	}
	lang := s.resolveLang(w, r)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := t.ExecuteTemplate(w, "layout", newPage(lang, r, data)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
