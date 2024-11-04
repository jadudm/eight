package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jadudm/eight/internal/env"
)

type LDBRequestInput struct{}

type LDBRequestReturn func(ctx context.Context, input *LDBRequestInput) (*struct{}, error)

type LDBResponse struct {
	Databases []string `json:"databases"`
}

type LDBResponseBody struct {
	Body *LDBResponse
}

func listHostedDOmains() []string {

	s, _ := env.Env.GetUserService("serve")
	database_files_path := s.GetParamString("database_files_path")

	files, err := os.ReadDir(database_files_path)
	if err != nil {
		log.Println("SERVE could not get directory listing")
		log.Fatal(err)
	}

	dbs := make([]string, 0)
	suffix := ".sqlite"
	for _, file := range files {
		if strings.HasSuffix(file.Name(), suffix) {
			dbs = append(dbs, strings.TrimSuffix(file.Name(), suffix))
		}
	}
	return dbs
}

func DatabasesHandler(c *gin.Context) {
	start := time.Now()

	duration := time.Since(start)
	dbs := listHostedDOmains()

	c.IndentedJSON(http.StatusOK, gin.H{
		"databases": dbs,
		"elapsed":   duration,
	})
}
