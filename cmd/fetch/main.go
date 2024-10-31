package main

import (
	"fmt"
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

	this, err := env.Env.GetServiceByName("user-provided", "fetch")
	if err != nil {
		log.Fatal(err)
	}
	http.ListenAndServe(fmt.Sprintf(":%d", this.Credentials.Port), extended_api)

}
