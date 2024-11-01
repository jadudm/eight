package main

import (
	"log"
	"net/http"

	"github.com/jadudm/eight/internal/api"
	"github.com/jadudm/eight/internal/env"
	"github.com/jadudm/eight/pkg/extract"
)

func main() {
	env.InitGlobalEnv()

	log.Println("environment initialized")

	ch := make(chan *extract.ExtractRequest)

	r := api.BaseMux()
	extended_api := ExtractApi(r, ch)

	go extract.Extract(ch)

	// Local and Cloud should both get this from the environment.
	http.ListenAndServe(":"+env.Env.Port, extended_api)

}
