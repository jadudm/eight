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
)

func main() {
	env.InitGlobalEnv()

	this, err := env.Env.GetUserService("serve")
	s, _ := env.Env.GetUserService("serve")
	external_port := s.GetParamInt64("external_port")
	static_files_path := s.GetParamString("static_files_path")

	log.Println("environment initialized")

	ch := make(chan *serve.ServeRequest)

	r := api.BaseMux()
	extended_api := ServeApi(r, ch)

	r.Route("/search", func(r chi.Router) {
		r.Get("/{host}", func(rw http.ResponseWriter, r *http.Request) {
			host := chi.URLParam(r, "host")
			rw.Header().Set("x-search-host", host)
			data, err := os.ReadFile("index.html")
			data = bytes.ReplaceAll(data, []byte("{HOST}"), []byte(host))
			data = bytes.ReplaceAll(data, []byte("{PORT}"), []byte(fmt.Sprintf("%d", external_port)))

			if err != nil {
				log.Fatal(err)
			}
			rw.Write(data)
		})
	})

	// Serve up the search page
	fs := http.FileServer(http.Dir(static_files_path))
	r.Handle("/static/*", http.StripPrefix("/static/", fs))

	go serve.Serve(ch)

	if err != nil {
		log.Fatal(err)
	}
	http.ListenAndServe(fmt.Sprintf(":%d", this.Credentials.Port), extended_api)

}
