package main

import (
	"fmt"
	"log"
	"net/http"

	"search.eight/internal/api"
	"search.eight/internal/env"
	"search.eight/pkg/pack"
)

func main() {
	env.InitGlobalEnv()

	log.Println("environment initialized")

	ch := make(chan *pack.PackRequest)

	r := api.BaseMux()
	extended_api := PackApi(r, ch)

	go pack.Pack(ch)

	this, err := env.Env.GetServiceByName("user-provided", "pack")
	if err != nil {
		log.Fatal(err)
	}
	http.ListenAndServe(fmt.Sprintf(":%d", this.Credentials.Port), extended_api)

}
