package main

import (
	"fmt"
	"log"
	"net/http"

	"search.eight/internal/api"
	"search.eight/internal/env"
	"search.eight/pkg/serve"
)

func main() {
	env.InitGlobalEnv()

	log.Println("environment initialized")

	ch := make(chan *serve.ServeRequest)

	r := api.BaseMux()
	extended_api := ServeApi(r, ch)

	go serve.Serve(ch)

	this, err := env.Env.GetServiceByName("user-provided", "serve")
	if err != nil {
		log.Fatal(err)
	}
	http.ListenAndServe(fmt.Sprintf(":%d", this.Credentials.Port), extended_api)

}
