package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jadudm/eight/internal/env"
	search_db "github.com/jadudm/eight/internal/sqlite/schemas"
	"go.uber.org/zap"
)

type LDBRequestInput struct{}

type LDBRequestReturn func(ctx context.Context, input *LDBRequestInput) (*struct{}, error)

type LDBResponse struct {
	Databases []string `json:"databases"`
}

type LDBResponseBody struct {
	Body *LDBResponse
}

func listHostedDomains() []string {

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

// Opening and closing the DB for this is slow.
// For the MVP, just cache it.
var cached_counts sync.Map

func countPages(domain string) int64 {
	s, _ := env.Env.GetUserService("serve")
	database_files_path := s.GetParamString("database_files_path")

	ctx := context.Background()
	sqlite_filename := database_files_path + "/" + domain + ".sqlite"

	if cached, ok := cached_counts.Load(domain); ok {
		return cached.(int64)
	} else {
		new_db, err := sql.Open("sqlite3", sqlite_filename)
		if err != nil {
			zap.L().Panic("cannot open database", zap.String("sqlite_filename", sqlite_filename))
		}
		defer new_db.Close()
		queries := search_db.New(new_db)
		pages, err := queries.CountSiteIndex(ctx)
		if err != nil {
			zap.L().Panic("could not get pages in database", zap.String("sqlite_filename", sqlite_filename))
		}
		cached_counts.Store(domain, pages)
		return pages
	}

}

func DatabasesHandler(c *gin.Context) {
	start := time.Now()

	duration := time.Since(start)
	dbs := listHostedDomains()

	c.IndentedJSON(http.StatusOK, gin.H{
		"databases": dbs,
		"elapsed":   duration,
	})
}
