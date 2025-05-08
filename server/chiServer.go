package server

import (
	"net/http"

	"github.com/eslami200117/ala_unlimited/handler"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
)

type ChiServer struct {
	route *chi.Mux
}

func NewChiServer() *ChiServer {
	return &ChiServer{
		route: chi.NewRouter(),
	}
}

func (s *ChiServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.route.ServeHTTP(w, r)
}

func (s *ChiServer) Initialize(api *handler.Api) {

	go cronJobs()

	s.route.Route("/api", func(r chi.Router) {
		r.Use(middleware.Logger)
		r.Use(middleware.Recoverer)
		r.Get("/start", api.StartCore)
		r.Get("/quit", api.QuitCore)
		r.Post("/update_seller", api.UpdateSeller)
	})

	// s.route.Get("/dkp", func(w http.ResponseWriter, r *http.Request) {
	// 	w.Write([]byte("DKP"))
	// })
	// s.route.Get("/dkp/{dkp}", func(w http.ResponseWriter, r *http.Request) {
	// 	dkp := chi.URLParam(r, "dkp")
	// 	w.Write([]byte(dkp))
	// })
}

func cronJobs() {
}
