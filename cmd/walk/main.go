package main

import (
	"log"
	"net/http"
	"time"

	expirable "github.com/go-pkgz/expirable-cache/v3"
	"github.com/patrickmn/go-cache"

	"github.com/jadudm/eight/internal/common"
	"github.com/jadudm/eight/internal/env"
	"go.uber.org/zap"
)

var expirable_cache expirable.Cache[string, int]
var recently_visited_cache *cache.Cache
var polite_sleep_milliseconds time.Duration

func get_ttl() int64 {
	ws, err := env.Env.GetUserService("walk")
	if err != nil {
		log.Println("WALK no service")
	}
	minutes := ws.GetParamInt64("cache-ttl-minutes")
	seconds := ws.GetParamInt64("cache-ttl-seconds")
	return (minutes * 60) + seconds
}

func main() {
	env.InitGlobalEnv()
	InitializeQueues()
	InitializeStorage()
	log.Println("environment initialized")
	service, _ := env.Env.GetUserService("walk")

	engine := common.InitializeAPI()

	ttl := get_ttl()
	expirable_cache = expirable.NewCache[string, int]().WithTTL(time.Second * time.Duration(ttl))

	recently_visited_cache = cache.New(
		time.Duration(service.GetParamInt64("polite_cache_default_expiration_minutes"))*time.Minute,
		time.Duration(service.GetParamInt64("polite_cache_cleanup_interval_minutes"))*time.Minute)

	log.Println("WALK listening on", env.Env.Port)

	zap.L().Info("listening to the music of the spheres",
		zap.String("port", env.Env.Port))
	// Local and Cloud should both get this from the environment.
	http.ListenAndServe(":"+env.Env.Port, engine)
}
