package main

import (
	"context"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	common "github.com/jadudm/eight/internal/common"
	"github.com/riverqueue/river"
	"go.uber.org/zap"
)

var FETCH_API_VERSION = "1.0.0"

type FetchRequestInput struct {
	Scheme string `json:"scheme" maxLength:"10" doc:"Resource scheme"`
	Host   string `json:"host" maxLength:"500" doc:"Host of resource"`
	Path   string `json:"path" maxLength:"1500" doc:"Path to resource"`
	ApiKey string `json:"api-key"`
}

func Heartbeat(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}

func InitializeAPI() *gin.Engine {
	router := gin.Default()
	router.GET("/heartbeat")
	router.PUT("fetch", FetchRequestHandler)
	return router
}

// https://dev.to/kashifsoofi/rest-api-with-go-chi-and-inmemory-store-43ag
func FetchRequestHandler(c *gin.Context) {
	var fri FetchRequestInput
	if err := c.BindJSON(&fri); err != nil {
		return
	}
	//zap.L().Debug("api checking key", zap.String("api-key", fri.ApiKey))

	if fri.ApiKey == os.Getenv("API_KEY") {
		zap.L().Debug("api enqueue", zap.String("host", fri.Host), zap.String("path", fri.Path))

		if fetchClient == nil {
			zap.L().Error("fetchClient IS NIL")
		}

		fetchClient.Insert(context.Background(), common.FetchArgs{
			Scheme: fri.Scheme,
			Host:   fri.Host,
			Path:   fri.Path,
		}, &river.InsertOpts{
			Queue: "fetch",
		})

		c.IndentedJSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	}
}
