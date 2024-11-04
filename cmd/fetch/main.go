package main

import (
	"log"
	"net/http"
	"time"

	common "github.com/jadudm/eight/internal/common"
	"github.com/jadudm/eight/internal/env"
	"github.com/patrickmn/go-cache"
	"go.uber.org/zap"
)

var recently_visited_cache *cache.Cache
var polite_sleep_milliseconds int64

func main() {
	env.InitGlobalEnv()
	InitializeQueues()
	InitializeStorage()

	engine := common.InitializeAPI()
	ExtendApi(engine)

	log.Println("environment initialized")

	// Init a cache for the workers
	service, _ := env.Env.GetUserService("fetch")

	// Pre-compute/lookup the sleep duration for backoff
	millis := service.GetParamInt64("polite_sleep_milliseconds")
	polite_sleep_milliseconds = millis

	recently_visited_cache = cache.New(
		time.Duration(service.GetParamInt64("polite_cache_default_expiration_minutes"))*time.Minute,
		time.Duration(service.GetParamInt64("polite_cache_cleanup_interval_minutes"))*time.Minute)

	zap.L().Info("listening to the music of the spheres",
		zap.String("port", env.Env.Port))
	// Local and Cloud should both get this from the environment.
	http.ListenAndServe(":"+env.Env.Port, engine)

}
