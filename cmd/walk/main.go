package main

import (
	"fmt"
	"log"
	"net/http"

	"search.eight/internal/api"
	"search.eight/internal/env"
	"search.eight/pkg/walk"
)

func main() {
	env.InitGlobalEnv()

	log.Println("environment initialized")

	ch := make(chan *walk.WalkRequest)

	r := api.BaseMux()
	extended_api := WalkApi(r, ch)

	go walk.Walk(ch)

	this, err := env.Env.GetServiceByName("user-provided", "walk")
	if err != nil {
		log.Fatal(err)
	}
	http.ListenAndServe(fmt.Sprintf(":%d", this.Credentials.Port), extended_api)

}
