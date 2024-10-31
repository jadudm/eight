package main

import (
	"fmt"
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

	this, err := env.Env.GetServiceByName("user-provided", "extract")
	if err != nil {
		log.Fatal(err)
	}
	http.ListenAndServe(fmt.Sprintf(":%d", this.Credentials.Port), extended_api)

}
