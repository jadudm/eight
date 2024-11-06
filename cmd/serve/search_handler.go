package main

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jadudm/eight/internal/common"
	"github.com/jadudm/eight/internal/env"
	"github.com/jadudm/eight/internal/sqlite/schemas"
	"github.com/kljensen/snowball"
	"go.uber.org/zap"
)

// FIXME This becomes the API search interface
type ServeRequestInput struct {
	Host  string `json:"host"`
	Terms string `json:"terms"`
}

var statmap sync.Map

func SearchHandler(c *gin.Context) {
	start := time.Now()
	var sri ServeRequestInput
	if err := c.BindJSON(&sri); err != nil {
		return
	}

	s, _ := env.Env.GetUserService("serve")
	database_files_path := s.GetParamString("database_files_path")
	results_per_query := s.GetParamInt64("results_per_query")

	sqlite_file := database_files_path + "/" + sri.Host + ".sqlite"
	if _, err := os.Stat(sqlite_file); errors.Is(err, os.ErrNotExist) {
		duration := time.Since(start)
		c.IndentedJSON(http.StatusOK, gin.H{
			"result":  "err",
			"elapsed": duration,
			"results": nil,
		})
		return
	}

	db, err := sql.Open("sqlite3", sqlite_file)
	if err != nil {
		log.Fatal("SERVCE cannot open SQLite file", sqlite_file)
	}

	// Stem the terms and add wildcards.
	improved_terms := make([]string, 0)
	for _, t := range strings.Split(sri.Terms, " ") {
		stemmed, err := snowball.Stem(t, "english", true)
		if err != nil {
			// Pass. Keep the value as-is
			improved_terms = append(improved_terms, t)
		} else {
			improved_terms = append(improved_terms, stemmed+"*")
		}
	}
	improved_terms_string := strings.Join(improved_terms, " ")

	zap.L().Info("search string",
		zap.String("original", sri.Terms),
		zap.String("improved", improved_terms_string))
	queries := schemas.New(db)
	res, err := queries.SearchSiteIndexSnippets(context.Background(), schemas.SearchSiteIndexSnippetsParams{
		Text:  improved_terms_string, //sri.Terms,
		Limit: results_per_query,
	})

	duration := time.Since(start)

	// Search accounting
	totalStats := common.NewBaseStats("serve")
	totalStats.Increment("queries")
	totalStats.Sum("total_query_time", duration.Nanoseconds())
	if totalStats.HasKey("total_query_time") && totalStats.HasKey("queries") {
		totalStats.Set("average_query_time", int64(totalStats.Get("total_query_time")/totalStats.Get("queries")))
	}

	var stats *common.BaseStats
	if m, ok := statmap.Load(sri.Host); ok {
		stats = m.(*common.BaseStats)
	} else {
		stats = common.NewBaseStats(sri.Host)
		statmap.Store(sri.Host, stats)
	}

	stats.Increment("queries")
	// stats.Increment("_" + sri.Host)
	stats.Sum("total_query_time", duration.Nanoseconds())
	if stats.HasKey("total_query_time") && stats.HasKey("queries") {
		stats.Set("average_query_time", int64(stats.Get("total_query_time")/stats.Get("queries")))
	}

	// Count all the search terms? Why not!
	for _, t := range strings.Split(sri.Terms, " ") {
		stats.Increment("term:" + t)
	}

	if err != nil {
		c.IndentedJSON(http.StatusOK, gin.H{
			"result":  "err",
			"elapsed": duration,
			"results": nil,
		})
		return
	} else {
		c.IndentedJSON(http.StatusOK, gin.H{
			"result":  "ok",
			"elapsed": duration,
			"results": res,
		})
		return
	}
}
