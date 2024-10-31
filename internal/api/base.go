package api

import (
	"net/http"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
)

type HttpHandlerFunc func(next http.Handler) http.Handler

func BaseMux() *chi.Mux {
	r := chi.NewMux()

	r.Use(middleware.Logger)
	r.Use(middleware.Heartbeat("/heartbeat"))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("HELO"))
	})

	MemInfo(r)

	return r
}
