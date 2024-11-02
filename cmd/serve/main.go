package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/jadudm/eight/internal/api"
	"github.com/jadudm/eight/internal/env"
	"github.com/jadudm/eight/pkg/serve"
	"go.uber.org/zap"
)

func main() {
	env.InitGlobalEnv()

	s, _ := env.Env.GetUserService("serve")
	static_files_path := s.GetParamString("static_files_path")
	external_scheme := s.GetParamString("external_scheme")
	external_host := s.GetParamString("external_host")
	external_port := s.GetParamInt64("external_port")

	log.Println("environment initialized")
	zap.L().Info("serve environment",
		zap.String("static_files_path", static_files_path),
		zap.String("external_host", external_host),
		zap.Int64("external_port", external_port),
	)

	ch := make(chan *serve.ServeRequest)

	r := api.BaseMux()
	extended_api := ServeApi(r, ch)

	r.Route("/search", func(r chi.Router) {
		r.Get("/{host}", func(rw http.ResponseWriter, r *http.Request) {
			host := chi.URLParam(r, "host")
			rw.Header().Set("x-search-host", host)
			data, err := os.ReadFile("index.html")
			if err != nil {
				log.Println("SERVE could not read index.html")
				log.Fatal(err)
			}
			data = bytes.ReplaceAll(data, []byte("{SCHEME}"), []byte(external_scheme))
			data = bytes.ReplaceAll(data, []byte("{HOST}"), []byte(external_host))
			data = bytes.ReplaceAll(data, []byte("{SEARCH_HOST}"), []byte(host))

			data = bytes.ReplaceAll(data, []byte("{PORT}"), []byte(fmt.Sprintf("%d", external_port)))

			rw.Write(data)
		})
	})

	// Serve up the search page
	fs := http.FileServer(http.Dir(static_files_path))
	r.Handle("/static/*", http.StripPrefix("/static/", fs))

	go serve.Serve(ch)

	// Local and Cloud should both get this from the environment.
	http.ListenAndServe(":"+env.Env.Port, extended_api)

}
