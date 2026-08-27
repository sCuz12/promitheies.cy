package web

import (
	"log"
	"net/http"

	"github.com/georgehadjisavvas/promitheies-cy/internal/cpv"
)

func (s *Server) handleHome(w http.ResponseWriter, r *http.Request) {
	stats, err := GetStats(r.Context(), s.pool)
	if err != nil {
		s.serverError(w, err)
		return
	}
	s.render(w, r, "home", stats)
}

func (s *Server) handleAuthorities(w http.ResponseWriter, r *http.Request) {
	list, err := ListAuthorities(r.Context(), s.pool)
	if err != nil {
		s.serverError(w, err)
		return
	}
	s.render(w, r, "authorities", list)
}

func (s *Server) handleAuthority(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	a, err := GetAuthority(r.Context(), s.pool, slug)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	s.render(w, r, "authority", a)
}

func (s *Server) handleContractors(w http.ResponseWriter, r *http.Request) {
	list, err := ListContractors(r.Context(), s.pool, 500)
	if err != nil {
		s.serverError(w, err)
		return
	}
	s.render(w, r, "contractors", list)
}

func (s *Server) handleContractor(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	c, err := GetContractor(r.Context(), s.pool, slug)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	s.render(w, r, "contractor", c)
}

type searchPageData struct {
	Query       string
	CPVDivision string
	Divisions   []cpv.Division
	Results     []TenderRow
}

func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	cpvDivision := r.URL.Query().Get("cpv")

	data := searchPageData{Query: q, CPVDivision: cpvDivision, Divisions: cpv.Divisions}

	if q != "" || cpvDivision != "" {
		results, err := SearchTenders(r.Context(), s.pool, q, "", cpvDivision, 100)
		if err != nil {
			s.serverError(w, err)
			return
		}
		data.Results = results
	}

	s.render(w, r, "search", data)
}

func (s *Server) serverError(w http.ResponseWriter, err error) {
	log.Printf("server error: %v", err)
	http.Error(w, "internal server error", http.StatusInternalServerError)
}
