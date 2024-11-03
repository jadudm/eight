package common

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func Heartbeat(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}

func InitializeAPI() *gin.Engine {
	router := gin.Default()
	router.GET("/heartbeat")
	return router
}
