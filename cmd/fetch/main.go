package main

import (
	"log"
	"net/http"

	"github.com/jadudm/eight/internal/api"
	"github.com/jadudm/eight/internal/env"
	"github.com/jadudm/eight/pkg/fetch"
)

func main() {
	env.InitGlobalEnv()
	log.Println("environment initialized")

	ch := make(chan *fetch.FetchRequest)

	r := api.BaseMux()
	extended_api := FetchApi(r, ch)

	go fetch.Fetch(ch)

	log.Println("Listening to the music of the spheres on", env.Env.Port)
	// Local and Cloud should both get this from the environment.
	http.ListenAndServe(":"+env.Env.Port, extended_api)

}
