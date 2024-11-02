package main

import (
	"log"
	"net/http"

	"github.com/jadudm/eight/internal/api"
	"github.com/jadudm/eight/internal/env"
	"github.com/jadudm/eight/pkg/walk"
)

func main() {
	env.InitGlobalEnv()

	log.Println("environment initialized")

	ch := make(chan *walk.WalkRequest)

	r := api.BaseMux()
	extended_api := WalkApi(r, ch)

	go walk.Walk(ch)

	log.Println("WALK listening on", env.Env.Port)
	// Local and Cloud should both get this from the environment.
	http.ListenAndServe(":"+env.Env.Port, extended_api)

}
