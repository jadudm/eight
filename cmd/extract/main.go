package main

import (
	"log"
	"net/http"

	"github.com/jadudm/eight/internal/common"
	"github.com/jadudm/eight/internal/env"
	"go.uber.org/zap"
)

func main() {
	env.InitGlobalEnv()
	InitializeQueues()
	InitializeStorage()

	log.Println("environment initialized")
	routers := common.InitializeAPI()

	zap.L().Info("listening to the music of the spheres",
		zap.String("port", env.Env.Port))
	// Local and Cloud should both get this from the environment.
	http.ListenAndServe(":"+env.Env.Port, routers)

}
