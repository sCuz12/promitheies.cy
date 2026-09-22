package web

import (
	"context"
	"html/template"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Server struct {
	pool               *pgxpool.Pool
	templates          map[string]*template.Template
	mux                *http.ServeMux
	storeNewsletterSub func(context.Context, string, Lang) error
}

// NewServer loads templates from templatesDir and wires up all routes,
// serving static assets from staticDir at /static/.
func NewServer(pool *pgxpool.Pool, templatesDir, staticDir string) (*Server, error) {
	templates, err := loadTemplates(templatesDir)
	if err != nil {
		return nil, err
	}
	s := &Server{pool: pool, templates: templates}
	s.storeNewsletterSub = func(ctx context.Context, email string, lang Lang) error {
		return StoreNewsletterSubscriber(ctx, pool, email, lang)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", s.handleHome)
	mux.HandleFunc("GET /authorities", s.handleAuthorities)
	mux.HandleFunc("GET /authority/{slug}", s.handleAuthority)
	mux.HandleFunc("GET /contractors", s.handleContractors)
	mux.HandleFunc("GET /contractor/{slug}", s.handleContractor)
	mux.HandleFunc("GET /tender/{id}", s.handleTender)
	mux.HandleFunc("GET /diagonismoi", s.handleOpenTenders)
	mux.HandleFunc("GET /search", s.handleSearch)
	mux.HandleFunc("POST /newsletter/subscribe", s.handleNewsletterSubscribe)
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir))))

	s.mux = mux
	return s, nil
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}
