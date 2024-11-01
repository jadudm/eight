package main

import (
	"log"
	"net/http"

	"github.com/jadudm/eight/internal/api"
	"github.com/jadudm/eight/internal/env"
	"github.com/jadudm/eight/pkg/pack"
)

// For using _modernc.org
// type sqliteDriver struct {
// 	*sqlite.Driver
// }

// func init() {
// 	sql.Register("sqlite3", sqliteDriver{Driver: &sqlite.Driver{}})
// }

func main() {
	env.InitGlobalEnv()

	log.Println("environment initialized")

	ch := make(chan *pack.PackRequest)

	r := api.BaseMux()
	extended_api := PackApi(r, ch)

	go pack.Pack(ch)

	// Local and Cloud should both get this from the environment.
	http.ListenAndServe(":"+env.Env.Port, extended_api)
}
